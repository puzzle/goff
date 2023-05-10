package kustomize

import (
	"bytes"
	"goff/kustomize/kustomizationfile"
	"goff/util"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/build"
)

func BuildAll(sourceDir, targetDir string) {

	dirs, err := kustomizationfile.New().GetDirectories(sourceDir)
	if err != nil {
		panic(err)
	}

	fSys := filesys.MakeFsOnDisk()
	for _, dir := range dirs {

		buffy := new(bytes.Buffer)
		cmd := build.NewCmdBuild(fSys, build.MakeHelp("foo", "bar"), buffy)
		if err := cmd.RunE(cmd, []string{dir}); err != nil {
			panic(err)
		}

		if buffy.Len() == 0 {
			continue
		}

		base := strings.TrimPrefix(dir, sourceDir)
		outPath := filepath.Join(targetDir, base)

		err = os.MkdirAll(outPath, 0777)
		if err != nil {
			panic(err)
		}

		outFiles, err := splitYAML(buffy.Bytes())
		if err != nil {
			panic(err)
		}

		for _, f := range outFiles {
			content := string(f)
			fileName, err := util.FileNameFromManifest(content)
			if err != nil {
				panic(err)
			}

			outFile := filepath.Join(outPath, fileName)

			err = os.WriteFile(outFile, buffy.Bytes(), 0777)
			if err != nil {
				panic(err)
			}
		}
	}
}

func splitYAML(resources []byte) ([][]byte, error) {

	dec := yaml.NewDecoder(bytes.NewReader(resources))

	var res [][]byte
	for {
		var value interface{}
		err := dec.Decode(&value)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		valueBytes, err := yaml.Marshal(value)
		if err != nil {
			return nil, err
		}
		res = append(res, valueBytes)
	}
	return res, nil
}
