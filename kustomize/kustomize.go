package kustomize

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/puzzle/goff/util"

	"github.com/puzzle/goff/kustomize/kustomizationfile"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/build"
)

func BuildAll(sourceDir, targetDir string) error {

	dirs, err := kustomizationfile.New().GetDirectories(sourceDir)
	if err != nil {
		return err
	}

	fSys := filesys.MakeFsOnDisk()
	for _, dir := range dirs {

		buffy := new(bytes.Buffer)
		cmd := build.NewCmdBuild(fSys, build.MakeHelp("foo", "bar"), buffy)
		if err := cmd.RunE(cmd, []string{dir}); err != nil {
			return err
		}

		if buffy.Len() == 0 {
			continue
		}

		ad, _ := filepath.Abs(dir)
		asd, _ := filepath.Abs(sourceDir)
		base := strings.TrimPrefix(ad, asd)
		outPath := filepath.Join(targetDir, base)

		err = os.MkdirAll(outPath, 0777)
		if err != nil {
			return err
		}

		outFiles := bytes.Split(buffy.Bytes(), []byte("---"))

		for _, f := range outFiles {
			content := string(f)
			fileName, err := util.FileNameFromManifest(content)
			if err != nil {
				return err
			}

			outFile := filepath.Join(outPath, fileName)

			err = os.WriteFile(outFile, f, 0777)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
