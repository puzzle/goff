package cmd

import (
	"fmt"
	"os/exec"

	"github.com/puzzle/goff/cmd/kustomize"
	"github.com/puzzle/goff/kustomize/kustomizationgraph"

	"github.com/spf13/cobra"
)

var outputDotDir *string
var version *bool

var kustomizeCmd = &cobra.Command{
	Use:   "kustomize [rootDir]",
	Short: "Generate a DOT file to visualize the dependencies between your kustomize components",
	Args: func(cmd *cobra.Command, args []string) error {
		if *version {
			return nil
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	Long: `Generate a DOT file to visualize the dependencies between your kustomize components`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if *version {
			kustomizeCmd := exec.CommandContext(cmd.Context(), "kustomize", "version")
			kustomizeCmd.Stdout = cmd.OutOrStdout()
			kustomizeCmd.Stderr = cmd.OutOrStderr()
			if err := kustomizeCmd.Run(); err != nil {
				return fmt.Errorf("unable to run kustomize: %w", err)
			}
			return nil
		}

		kustomizationgraph.Graph(args[0], *outputDotDir)
		return nil
	},
}

func init() {
	kustomizeCmd.AddCommand(kustomize.KustomizeBuildCmd)
	rootCmd.AddCommand(kustomizeCmd)

	version = kustomizeCmd.Flags().BoolP("version", "v", false, "Display version of kustomize")
	outputDotDir = kustomizeCmd.Flags().StringP("output-dir", "o", ".", "Output directory")
}
