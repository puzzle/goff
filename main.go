/*
Copyright © 2023 Ch. Schlatter schlatter@puzzle.ch

*/
package main

import (
	"goff/kustomize"
)

func main() {
	//cmd.Execute()
	kustomize.Build("testdata/kustomize/source/kustomize/envs/integration-gpu", "")
}
