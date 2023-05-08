package argocd

import (
	"context"
	"encoding/json"
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

	data, err := os.ReadFile("testdata/app_helm.yaml")
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

	resp, err := b.GenerateManifest(context.Background(), req)
	if err != nil {
		panic(err)
	}

	outputFile := filepath.Join(outputDir, "argo-rendered.yaml")

	os.WriteFile(outputFile, []byte(resp.Manifests[0]), 0777)

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
			f := filepath.Join(path, info.Name())
			data, err := os.ReadFile(f)
			if err != nil {
				return err
			}
			res := &Ressource{}
			err = yaml.Unmarshal(data, res)
			if err != nil {
				return err
			}

			if res.Kind == "Application" && res.ApiVersion == "argoproj.io/v1alpha1" {
				argoAppFiles = append(argoAppFiles, f)
			}

		}
		return nil
	})

	return argoAppFiles, err
}
