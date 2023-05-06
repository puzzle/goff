package diff

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
)

var (
	//go:embed templates/*
	files     embed.FS
	templates map[string]*template.Template
)

type diffMd struct {
	Title string
	Files []fileDiff
}

type fileDiff struct {
	Filename string
	Diff     string
}

func Diff(title, templateName, sourceDir, tragetDir, outputDir string) {

	target, _ := findAsMap(tragetDir)
	source, _ := findAsMap(sourceDir)
	for file, _ := range source {
		if _, ok := target[file]; !ok {
			target[file] = ""
		}
	}

	for file, _ := range target {
		if _, ok := source[file]; !ok {
			source[file] = ""
		}
	}

	template, err := getTemplate(templateName)
	if err != nil {
		panic(err)
	}

	diffs := make([]fileDiff, 0)

	for file, contentSrc := range source {

		contentTarget := target[file]
		edits := myers.ComputeEdits(span.URIFromPath(file), contentTarget, contentSrc)
		diff := fmt.Sprint(gotextdiff.ToUnified(file, file, contentTarget, edits))

		if diff == "" {
			continue
		}

		diffFile := fileDiff{
			Filename: file,
			Diff:     diff,
		}

		diffs = append(diffs, diffFile)

	}

	d := diffMd{
		Title: title,
		Files: diffs,
	}

	path := filepath.Join(outputDir, "diff.md")

	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}

	err = template.Execute(f, d)
	if err != nil {
		panic(err)
	}

}

func getTemplate(templateName string) (*template.Template, error) {

	var file string

	switch templateName {
	case "gitlab":
		file = "gitlab.md"
	case "markdown":
		file = "markdown.md"
	default:
		return nil, errors.New("unsupported template")
	}

	templateFile := filepath.Join("templates/", file)

	return template.ParseFS(files, templateFile)

}

func findAsMap(root string) (map[string]string, error) {
	var f map[string]string
	f = make(map[string]string)

	err := filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == ".yaml" {
			content, _ := os.ReadFile(s)
			f[d.Name()] = string(content)
		}
		return nil
	})
	return f, err
}
