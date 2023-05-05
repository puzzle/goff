/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"goff/kustomize/kustomizationgraph"

	"github.com/spf13/cobra"
)

var outputDotDir *string

// kustomizeCmd represents the kustomize command
var kustomizeCmd = &cobra.Command{
	Use:   "kustomize [rootDir]",
	Short: "Generate a DOT file to visualize the dependencies betw",
	Args:  cobra.ExactArgs(1),
	Long:  `Generate a DOT file to visualize the dependencies between your kustomize components`,
	Run: func(cmd *cobra.Command, args []string) {
		kustomizationgraph.Graph(args[0], *outputDotDir)
	},
}

func init() {
	rootCmd.AddCommand(kustomizeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// kustomizeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// kustomizeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	outputDotDir = kustomizeCmd.Flags().StringP("output-dir", "o", ".", "Output directory")
}
