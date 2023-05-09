/*
Copyright © 2023 Ch. Schlatter schlatter@puzzle.ch

*/
package main

import (
	"goff/cmd"
	"goff/kustomize"
)

func main() {
	cmd.Execute()
	kustomize.BuildAll("testdata/kustomize/source/kustomize", "./out")
}
