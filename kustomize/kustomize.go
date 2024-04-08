package kustomize

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/puzzle/goff/kustomize/kustomizationfile"
	"github.com/puzzle/goff/util"
)

func BuildAll(sourceDir, targetDir string) error {

	dirs, err := kustomizationfile.New().GetDirectories(sourceDir)
	if err != nil {
		return err
	}

	absoluteSourceDirPath, _ := filepath.Abs(sourceDir)

	for _, dir := range dirs {
		absoluteKustomizationPath, _ := filepath.Abs(dir)

		var stdout strings.Builder

		// TODO: Make customizable
		cmd := exec.Command("kustomize", "build", absoluteKustomizationPath)
		cmd.Stdout = &stdout
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error running kustomize: %w", err)
		}

		if stdout.Len() == 0 {
			// TODO: May log warning that there was no output for the directory.
			continue
		}

		// Remove the source directory to the found directory of the `Kustomization`.
		//
		// Example: If source is `/dir/to/kustomization` and the found kustomization is `/dir/to/kustomization/overlays/x`,
		// the basePath will be `overlays/x`.
		basePath := strings.TrimPrefix(absoluteKustomizationPath, absoluteSourceDirPath)
		outputPath := filepath.Join(targetDir, basePath)
		log.Println(absoluteKustomizationPath, " -> ", outputPath)

		if err := os.MkdirAll(outputPath, 0777); err != nil {
			return fmt.Errorf("unable to create target direcories: %w", err)
		}

		manifests := strings.Split(stdout.String(), "---")
		for _, manifest := range manifests {
			fileName, err := util.FileNameFromManifest(manifest)
			if err != nil {
				return fmt.Errorf("cannot get name of manifest: %w", err)
			}

			if err := os.WriteFile(filepath.Join(outputPath, fileName), []byte(manifest), 0777); err != nil {
				return fmt.Errorf("unable to write manifest to file: %w", err)
			}
		}
	}

	return nil
}
