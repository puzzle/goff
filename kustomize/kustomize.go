package kustomize

import (
	"bytes"
	"goff/kustomize/kustomizationfile"
	"os"
	"path/filepath"
	"strings"

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

		outFile := filepath.Join(outPath, "out.yaml")

		err = os.WriteFile(outFile, buffy.Bytes(), 0777)
		if err != nil {
			panic(err)
		}
	}
}
