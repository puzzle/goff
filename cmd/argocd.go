/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/spf13/cobra"
)

var repoServerUrl *string

var ArgoOutputDir *string

// argocdCmd represents the argocd command
var ArgocdCmd = &cobra.Command{
	Use:   "argocd [command]",
	Short: "Render manifests from ArgoCD",
	Long:  `Render manifests from ArgoCD`,
}

func init() {
	rootCmd.AddCommand(ArgocdCmd)
	ArgoOutputDir = ArgocdCmd.PersistentFlags().StringP("output-dir", "o", ".", "Output directory")
}
