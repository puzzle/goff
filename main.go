/*
Copyright © 2023 Ch. Schlatter schlatter@puzzle.ch

*/
package main

import (
	"goff/cmd"
	_ "goff/cmd/argocd"
)

func main() {
	cmd.Execute()
}
