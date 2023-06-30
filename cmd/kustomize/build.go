package kustomize

import (
	"github.com/puzzle/goff/kustomize"

	"github.com/spf13/cobra"
)

/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/

var outputBuildDir *string

// kustomizeBuildCmd represents the kustomize command
var KustomizeBuildCmd = &cobra.Command{
	Use:   "build [rootDir]",
	Short: "Build all kustomize file within parent directory",
	Args:  cobra.ExactArgs(1),
	Long:  `Build all kustomize file within parent directory`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return kustomize.BuildAll(args[0], *outputBuildDir)
	},
}

func init() {

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// kustomizeBuildCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// kustomizeBuildCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	outputBuildDir = KustomizeBuildCmd.Flags().StringP("output-dir", "o", ".", "Output directory")
}
