package argocd

import (
	"os"
	"testing"
)

func TestRenderApplicationSet(t *testing.T) {

	outPath := "../out/app-set/"
	err := RenderApplicationSets("../testdata/appSet/input", outPath)
	if err != nil {
		t.Error(err)
	}
	generated, err := os.ReadDir(outPath)
	if err != nil {
		t.Error(err)
	}

	expected, err := os.ReadDir("../testdata/appSet/expected/")
	if err != nil {
		t.Error(err)
	}

OUTER:
	for _, e := range expected {
		expectedContent, _ := os.ReadFile(e.Name())
		for _, g := range generated {
			generatedContent, _ := os.ReadFile(g.Name())
			if string(expectedContent) == string(generatedContent) {
				continue OUTER
			}
		}
		t.Errorf("no matching file found for:  %s", e.Name())
	}
}
