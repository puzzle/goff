package kustomizationfile

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

// KustomizationFile represents a kustomization yaml file
type KustomizationFile struct {
	Bases                 []string `yaml:"bases"`
	Resources             []string `yaml:"resources"`
	Patches               []string `yaml:"patches"`
	Components            []string `yaml:"components"`
	PatchesStrategicMerge []string `yaml:"patchesStrategicMerge"`
}

// KustomizationFileNames represents a list of allowed filenames that
// kustomize searches for
var KustomizationFileNames = []string{
	"kustomization.yaml",
	"kustomization.yml",
	"Kustomization",
}

type kustomizationFileContext struct {
	fileSystem afero.Fs
}

// New returns a new context to interact with kustomization files
func New() *kustomizationFileContext {
	defaultFileSystem := afero.NewOsFs()

	return NewFromFileSystem(defaultFileSystem)
}

// NewFromFileSystem creates a context to interact with kustomization files from a provided file system
func NewFromFileSystem(fileSystem afero.Fs) *kustomizationFileContext {
	return &kustomizationFileContext{
		fileSystem: fileSystem,
	}
}

// GetDirectories attempts to read a kustomization.yaml file from the given directory
func (k *kustomizationFileContext) GetDirectories(directoryRootPath string) ([]string, error) {
	dir, err := os.Stat(directoryRootPath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not state file %v", directoryRootPath)
	}

	if !dir.IsDir() {
		return nil, nil
	}

	directories := make([]string, 0)
	err = filepath.Walk(directoryRootPath, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if contains(KustomizationFileNames, info.Name()) {
			p := strings.TrimSuffix(path, info.Name())
			directories = append(directories, p)
		}

		return nil
	})

	return directories, err
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

// GetFromDirectory attempts to read a kustomization.yaml file from the given directory
func (k *kustomizationFileContext) GetFromDirectory(directoryPath string) (*KustomizationFile, error) {

	var kustomizationFile KustomizationFile

	fileUtility := &afero.Afero{Fs: k.fileSystem}

	exists, err := fileUtility.DirExists(directoryPath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not state file %v", directoryPath)
	}

	if !exists {
		return nil, nil
	}

	fileFoundCount := 0
	kustomizationFilePath := ""
	for _, kustomizationFile := range KustomizationFileNames {
		currentPath := path.Join(directoryPath, kustomizationFile)

		exists, err := fileUtility.Exists(currentPath)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not check if file %v exists", currentPath)
		}

		if exists {
			kustomizationFilePath = currentPath
			fileFoundCount++
		}
	}

	if kustomizationFilePath == "" {
		return nil, errors.Wrapf(errors.New("Missing kustomization file"), "Error in directory %v", directoryPath)
	}

	if fileFoundCount > 1 {
		return nil, errors.Wrapf(errors.New("Too many kustomization files"), "Error in directory %v", directoryPath)
	}

	kustomizationFileBytes, err := fileUtility.ReadFile(kustomizationFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not read file %s", kustomizationFilePath)
	}

	err = yaml.Unmarshal(kustomizationFileBytes, &kustomizationFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not unmarshal yaml file %s", kustomizationFilePath)
	}

	return &kustomizationFile, nil
}

// GetMissingResources returns a collection of resources that exist in the directory
// but are not defined in the given kustomization file
func (k *kustomizationFileContext) GetMissingResources(directoryPath string, kustomizationFile *KustomizationFile) ([]string, error) {
	definedResources := []string{}
	definedResources = append(definedResources, kustomizationFile.Resources...)
	definedResources = append(definedResources, kustomizationFile.Patches...)
	definedResources = append(definedResources, kustomizationFile.PatchesStrategicMerge...)

	fileUtility := &afero.Afero{Fs: k.fileSystem}
	directoryInfo, err := fileUtility.ReadDir(directoryPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not read directory %s", directoryPath)
	}

	missingResources := []string{}
	for _, info := range directoryInfo {
		if info.IsDir() {
			continue
		}

		// Only consider the resource missing if it is a yaml file
		if filepath.Ext(info.Name()) != ".yaml" {
			continue
		}

		// Ignore the kustomization files
		if existsInSlice(KustomizationFileNames, info.Name()) {
			continue
		}

		if !existsInSlice(definedResources, info.Name()) {
			missingResources = append(missingResources, info.Name())
		}
	}

	return missingResources, nil
}

func existsInSlice(slice []string, element string) bool {
	for _, current := range slice {
		if current == element {
			return true
		}
	}
	return false
}
