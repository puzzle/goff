package argocd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/puzzle/goff/util"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v2/reposerver/apiclient"
	"github.com/argoproj/argo-cd/v2/util/argo"
	dbmocks "github.com/argoproj/argo-cd/v2/util/db/mocks"
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
)

type RepoCredentails struct {
	Username string
	Password string
	KeyFile  string
}

func Render(dir, repoServerUrl, outputDir string, creds RepoCredentails) error {
	conn := apiclient.NewRepoServerClientset(repoServerUrl, 600, apiclient.TLSConfiguration{StrictValidation: false})
	r, client, err := conn.NewRepoServerClient()
	defer r.Close()

	if err != nil {
		return errors.Wrap(err, "could not connect to repo server")
	}

	files, err := findArgoApps(dir)

	if err != nil {
		return errors.Wrap(err, "could not find argo apps")
	}

	var lastErr error
	for _, file := range files {
		log.Debugf("processing ArgoCD Application at: %s", file)
		err = renderFile(file, repoServerUrl, outputDir, client, creds)
		if err != nil {
			log.Errorf("could not render argoCD Application: %v", err)
			lastErr = err
		}
	}
	return lastErr
}

func renderFile(file, repoServerUrl, outputDir string, client apiclient.RepoServerServiceClient, creds RepoCredentails) error {

	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	data, err = yaml.YAMLToJSON(data)
	if err != nil {
		return err
	}

	app := &v1alpha1.Application{}

	err = json.Unmarshal(data, app)
	if err != nil {
		return err
	}

	repoDB := &dbmocks.ArgoDB{}
	source := v1alpha1.ApplicationSource{}

	var privateKey string
	if creds.KeyFile != "" {
		data, err := os.ReadFile(creds.KeyFile)
		if err != nil {
			return err
		}
		privateKey = string(data)
	}

	if app.Spec.Source != nil {
		repoDB.On("GetRepository", context.Background(), app.Spec.Source.RepoURL).Return(&v1alpha1.Repository{
			Repo:               app.Spec.Source.RepoURL,
			SSHPrivateKey:      privateKey,
			Username:           creds.Username,
			Password:           creds.Password,
			ForceHttpBasicAuth: true,
		}, nil)
		source = *app.Spec.Source
	}

	if app.Spec.Sources != nil {
		for i := range app.Spec.Sources {
			source = app.Spec.Sources[i]
			repo := app.Spec.Sources[i].RepoURL
			if repo != "" {
				repoDB.On("GetRepository", context.Background(), repo).Return(&v1alpha1.Repository{
					Repo:               repo,
					SSHPrivateKey:      privateKey,
					Username:           creds.Username,
					Password:           creds.Password,
					ForceHttpBasicAuth: true,
				}, nil)
			}
		}
	}

	refSources, err := argo.GetRefSources(context.Background(), app.Spec, repoDB)
	req := &apiclient.ManifestRequest{
		ApplicationSource:  &source,
		AppName:            "goff-test",
		NoCache:            true,
		RefSources:         refSources,
		HasMultipleSources: true,
		Revision:           source.TargetRevision,
		KustomizeOptions: &v1alpha1.KustomizeOptions{
			BuildOptions: "--enable-helm",
		},
		Repo: &v1alpha1.Repository{
			Repo:               source.RepoURL,
			SSHPrivateKey:      privateKey,
			Username:           creds.Username,
			Password:           creds.Password,
			ForceHttpBasicAuth: true,
		},
	}

	resp, err := client.GenerateManifest(context.Background(), req)
	if err != nil {
		return fmt.Errorf("could not process application '%s': %w", app.Name, err)
	}

	err = os.MkdirAll(outputDir, 0777)
	if err != nil {
		return err
	}

	for _, manifest := range resp.Manifests {

		fileName, err := util.FileNameFromManifest(manifest)
		if err != nil {
			return err
		}

		outputFile := filepath.Join(outputDir, fileName)

		yamlManifest, err := yaml.JSONToYAML([]byte(manifest))
		if err != nil {
			return err
		}

		err = os.WriteFile(outputFile, yamlManifest, 0777)
		if err != nil {
			return err
		}

	}
	return nil
}

func findArgoApps(rootDir string) ([]string, error) {
	var argoAppFiles []string
	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if strings.HasSuffix(info.Name(), ".yml") || strings.HasSuffix(info.Name(), ".yaml") {

			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			res := &util.Ressource{}
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
	log.Debugf("Found %d ArgoCD Applications to process", len(argoAppFiles))
	return argoAppFiles, err
}
