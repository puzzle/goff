package argocd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/argoproj/argo-cd/v2/applicationset/generators"
	"github.com/argoproj/argo-cd/v2/applicationset/utils"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/ghodss/yaml"
)

func RenderApplicationSet(appSetFile, outDir string) {

	appSet := &v1alpha1.ApplicationSet{}

	data, err := os.ReadFile(appSetFile)
	if err != nil {
		panic(err)
	}

	data, err = yaml.YAMLToJSON(data)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(data, appSet)
	if err != nil {
		panic(err)
	}

	listGen := generators.NewListGenerator()
	supportedGens := make(map[string]generators.Generator)
	supportedGens["List"] = listGen

	apps, _, err := generateApplications(*appSet, supportedGens)

	err = writeApplications(apps, outDir)
	if err != nil {
		panic(err)
	}

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

			fmt.Println("error generating application from params")
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
					fmt.Println("error generating application from params")

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
