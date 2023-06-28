package argocd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/argoproj/argo-cd/v2/applicationset/generators"
	"github.com/argoproj/argo-cd/v2/applicationset/utils"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
)

func RenderApplicationSets(inputDir, outDir string) error {
	inputDir, err := filepath.Abs(inputDir)
	if err != nil {
		return err
	}

	files := make([]string, 0)
	filepath.WalkDir(inputDir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	var hasError bool
	for i := range files {
		file := files[i]
		err = RenderApplicationSet(file, outDir)
		if err != nil {
			log.Errorf("could not process application set '%s': %s", file, err.Error())
			hasError = true
		}
	}

	if hasError {
		return fmt.Errorf("failed to process application set")
	}
	return nil
}

func RenderApplicationSet(appSetFile, outDir string) error {

	appSet := &v1alpha1.ApplicationSet{}

	data, err := os.ReadFile(appSetFile)
	if err != nil {
		return fmt.Errorf("could not read applicationSet: %w", err)
	}

	data, err = yaml.YAMLToJSON(data)
	if err != nil {
		return fmt.Errorf("could not convert ApplicationSet to YAML: %w", err)
	}

	err = yaml.Unmarshal(data, appSet)
	if err != nil {
		return fmt.Errorf("could not unmarshall ApplicationSet: %w", err)
	}

	listGen := generators.NewListGenerator()
	supportedGens := make(map[string]generators.Generator)
	supportedGens["List"] = listGen

	apps, _, err := generateApplications(*appSet, supportedGens)
	if err != nil {
		return fmt.Errorf("could not generate applications: %w", err)
	}

	outDir = filepath.Join(outDir, appSet.Namespace, appSet.Name)

	err = writeApplications(apps, outDir)
	if err != nil {
		return fmt.Errorf("could not write applications: %w", err)
	}
	return nil
}

func writeApplications(apps []v1alpha1.Application, ouputDir string) error {

	err := os.MkdirAll(ouputDir, 0777)
	if err != nil {
		return err
	}

	for i := range apps {
		app := apps[i]
		data, err := yaml.Marshal(app)
		if err != nil {
			return err
		}

		fileName := fmt.Sprintf("application-%d.yaml", i)
		fileName = filepath.Join(ouputDir, fileName)
		err = os.WriteFile(fileName, data, 0777)
		if err != nil {
			return err
		}
	}

	return nil
}

func generateApplications(applicationSetInfo v1alpha1.ApplicationSet, supportedGenerators map[string]generators.Generator) ([]v1alpha1.Application, v1alpha1.ApplicationSetReasonType, error) {
	var res []v1alpha1.Application

	var firstError error
	var applicationSetReason v1alpha1.ApplicationSetReasonType

	renderer := utils.Render{}

	for _, requestedGenerator := range applicationSetInfo.Spec.Generators {
		t, err := generators.Transform(requestedGenerator, supportedGenerators, applicationSetInfo.Spec.Template, &applicationSetInfo, map[string]interface{}{})
		if err != nil {

			log.Errorf("error generating application from params")
			if firstError == nil {
				firstError = err
				applicationSetReason = v1alpha1.ApplicationSetReasonApplicationParamsGenerationError
			}
			continue
		}

		for _, a := range t {
			tmplApplication := getTempApplication(a.Template)

			for _, p := range a.Params {
				app, err := renderer.RenderTemplateParams(tmplApplication, applicationSetInfo.Spec.SyncPolicy, p, applicationSetInfo.Spec.GoTemplate)

				if err != nil {
					log.Errorf("error generating application from params")

					if firstError == nil {
						firstError = err
						applicationSetReason = v1alpha1.ApplicationSetReasonRenderTemplateParamsError
					}
					continue
				}
				res = append(res, *app)
			}
		}

		//log.WithField("generator", requestedGenerator).Infof("generated %d applications", len(res))
		//log.WithField("generator", requestedGenerator).Debugf("apps from generator: %+v", res)
	}

	return res, applicationSetReason, firstError
}

func getTempApplication(applicationSetTemplate v1alpha1.ApplicationSetTemplate) *v1alpha1.Application {
	var tmplApplication v1alpha1.Application
	tmplApplication.Annotations = applicationSetTemplate.Annotations
	tmplApplication.Labels = applicationSetTemplate.Labels
	tmplApplication.Namespace = applicationSetTemplate.Namespace
	tmplApplication.Name = applicationSetTemplate.Name
	tmplApplication.Spec = applicationSetTemplate.Spec
	tmplApplication.Finalizers = applicationSetTemplate.Finalizers

	return &tmplApplication
}
