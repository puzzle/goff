/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"goff/argocd"

	"github.com/spf13/cobra"
)

var repoServerUrl *string

var argoOutputDir *string

// argocdCmd represents the argocd command
var argocdCmd = &cobra.Command{
	Use:   "argocd [rrotDir]",
	Short: "Render manifests from ArgoCD Application",
	Args:  cobra.ExactArgs(1),
	Long:  `Render manifests from ArgoCD Application`,
	Run: func(cmd *cobra.Command, args []string) {
		argocd.Render(args[0], *repoServerUrl, *argoOutputDir)
	},
}

func init() {
	rootCmd.AddCommand(argocdCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// argocdCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	repoServerUrl = argocdCmd.Flags().String("repoServer", "localhost:8081", "URL to argoCD repo server")
	argoOutputDir = diffCmd.Flags().StringP("output-dir", "o", ".", "Output directory")
}
