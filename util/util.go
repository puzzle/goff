package util

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
)

//Multi doc yaml, split and save
func SplitManifests(manifestFile, outDir string) error {

	data, err := os.ReadFile(manifestFile)
	if err != nil {
		return err
	}

	err = os.MkdirAll(outDir, 0777)
	if err != nil {
		return err
	}

	splitted := bytes.Split(data, []byte("---"))

	for i := range splitted {
		if len(splitted[i]) == 0 {
			continue
		}

		res := &Ressource{}
		err := yaml.Unmarshal(splitted[i], res)
		if err != nil {
			return err
		}

		filename := fmt.Sprintf("%s-%s.yaml", res.Kind, res.Metadata.Name)
		filename = filepath.Join(outDir, filename)

		err = os.WriteFile(filename, []byte(splitted[i]), 0777)
		if err != nil {
			return err
		}
		fmt.Println("wrote file at: " + filename)
	}

	return nil

}

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
