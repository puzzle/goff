package argocd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v2/reposerver/apiclient"
	"github.com/argoproj/argo-cd/v2/util/argo"
	dbmocks "github.com/argoproj/argo-cd/v2/util/db/mocks"
	"github.com/ghodss/yaml"
)

func Render() {

	conn := apiclient.NewRepoServerClientset("0.0.0.0:8081", 30, apiclient.TLSConfiguration{StrictValidation: false})
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

	fmt.Printf("%+v", resp.Manifests)

}
