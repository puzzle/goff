/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/puzzle/goff/cmd/kustomize"
	"github.com/puzzle/goff/kustomize/kustomizationgraph"

	"github.com/spf13/cobra"
)

var outputDotDir *string

// kustomizeCmd represents the kustomize command
var kustomizeCmd = &cobra.Command{
	Use:   "kustomize [rootDir]",
	Short: "Generate a DOT file to visualize the dependencies between your kustomize components",
	Args:  cobra.ExactArgs(1),
	Long:  `Generate a DOT file to visualize the dependencies between your kustomize components`,
	Run: func(cmd *cobra.Command, args []string) {
		kustomizationgraph.Graph(args[0], *outputDotDir)
	},
}

func init() {
	kustomizeCmd.AddCommand(kustomize.KustomizeBuildCmd)
	rootCmd.AddCommand(kustomizeCmd)

	outputDotDir = kustomizeCmd.Flags().StringP("output-dir", "o", ".", "Output directory")
}
