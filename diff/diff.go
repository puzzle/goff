package diff

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/go-godo/godo/glob"

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

// Returns true if any chenges detected
func Diff(title, templateName, sourceDir, tragetDir, glob, outputDir string, exitCode int) (bool, error) {

	template, err := getTemplate(templateName)
	if err != nil {
		return false, err
	}

	diffs := CreateDiffs(tragetDir, sourceDir, glob)

	d := diffMd{
		Title: title,
		Files: diffs,
	}

	path := filepath.Join(outputDir, "diff.md")

	f, err := os.Create(path)
	if err != nil {
		return false, err
	}

	if len(diffs) < 1 {
		os.WriteFile(path, []byte("### ⚠️ No changes detected!"), 0777)
		return false, nil
	}

	err = template.Execute(f, d)
	if err != nil {
		return true, err
	}

	return true, nil
}

func CreateDiffs(tragetDir, sourceDir, glob string) []fileDiff {
	target, _ := findAsMap(tragetDir, glob)
	source, _ := findAsMap(sourceDir, glob)
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
	return diffs
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

func findAsMap(root, globPattern string) (map[string]string, error) {
	var f map[string]string
	f = make(map[string]string)

	globPattern = path.Join(path.Clean(root), globPattern)

	files, _, err := glob.Glob([]string{globPattern})
	if err != nil {
		return nil, err
	}

	for i := range files {
		file := files[i].Path
		relPath := strings.TrimPrefix(file, path.Clean(root))

		content, _ := os.ReadFile(file)
		f[relPath] = string(content)
	}

	return f, nil
}
