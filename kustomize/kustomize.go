package kustomize

import (
	"bytes"
	"fmt"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/build"
)

func Build(sourceDir, targetDir string) {
	fSys := filesys.MakeFsOnDisk()

	buffy := new(bytes.Buffer)
	cmd := build.NewCmdBuild(fSys, build.MakeHelp("foo", "bar"), buffy)
	if err := cmd.RunE(cmd, []string{sourceDir}); err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(buffy.String())
}
