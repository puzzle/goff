/*
Copyright © 2023 Ch. Schlatter schlatter@puzzle.ch

*/
package main

import (
	"goff/argocd"
)

func main() {
	//cmd.Execute()
	argocd.Render("./testdata/argocd", "localhost:8081", ".")
}
