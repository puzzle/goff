package argocd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v2/reposerver/apiclient"
	"github.com/argoproj/argo-cd/v2/util/argo"
	dbmocks "github.com/argoproj/argo-cd/v2/util/db/mocks"
	"github.com/ghodss/yaml"
)

func Render(dir, repoServerUrl, outputDir string) {
	conn := apiclient.NewRepoServerClientset(repoServerUrl, 30, apiclient.TLSConfiguration{StrictValidation: false})
	r, b, err := conn.NewRepoServerClient()
	defer r.Close()

	if err != nil {
		panic(err)
	}

	files, err := findArgoApps(dir)

	if err != nil {
		panic(err)
	}

	for _, file := range files {
		renderFile(file, repoServerUrl, outputDir, b)
	}

}

func renderFile(file, repoServerUrl, outputDir string, client apiclient.RepoServerServiceClient) {

	data, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}

	data, err = yaml.YAMLToJSON(data)
	if err != nil {
		panic(err)
	}

	app := &v1alpha1.Application{}

	err = json.Unmarshal(data, app)
	if err != nil {
		panic(err)
	}

	repoDB := &dbmocks.ArgoDB{}
	repoDB.On("GetRepository", context.Background(), "https://github.com/schlapzz/goff-examples.git").Return(&v1alpha1.Repository{
		Repo: "https://github.com/schlapzz/goff-examples.git",
	}, nil)

	refSources, err := argo.GetRefSources(context.Background(), app.Spec, repoDB)
	req := &apiclient.ManifestRequest{
		ApplicationSource:  &app.Spec.Sources[0],
		AppName:            "goff-test",
		NoCache:            true,
		RefSources:         refSources,
		HasMultipleSources: true,
		Revision:           app.Spec.Sources[0].TargetRevision,
		Repo: &v1alpha1.Repository{
			Repo: app.Spec.Sources[0].RepoURL,
		},
	}

	resp, err := client.GenerateManifest(context.Background(), req)
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(outputDir, 0777)
	if err != nil {
		panic(err)
	}

	for _, manifest := range resp.Manifests {

		fileName, err := fileNameFromManifest(manifest)
		if err != nil {
			panic(err)
		}

		outputFile := filepath.Join(outputDir, fileName)

		yamlManifest, err := yaml.JSONToYAML([]byte(manifest))
		if err != nil {
			panic(err)
		}

		err = os.WriteFile(outputFile, yamlManifest, 0777)
		if err != nil {
			panic(err)
		}

	}

}

func fileNameFromManifest(manifest string) (string, error) {
	res := &Ressource{}
	err := yaml.Unmarshal([]byte(manifest), res)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%s.yaml", res.Kind, res.Metadata.Name), nil
}

type Ressource struct {
	Metadata   metadata `yaml:"metadata"`
	ApiVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
}

type metadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

func findArgoApps(rootDir string) ([]string, error) {
	var argoAppFiles []string
	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if strings.HasSuffix(info.Name(), ".yml") || strings.HasSuffix(info.Name(), ".yaml") {

			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			res := &Ressource{}
			err = yaml.Unmarshal(data, res)
			if err != nil {
				return err
			}

			if res.Kind == "Application" && res.ApiVersion == "argoproj.io/v1alpha1" {
				argoAppFiles = append(argoAppFiles, path)
			}

		}
		return nil
	})

	return argoAppFiles, err
}
