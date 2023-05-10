package util

import (
	"fmt"

	"github.com/ghodss/yaml"
)

func FileNameFromManifest(manifest string) (string, error) {
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
