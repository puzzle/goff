/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/puzzle/goff/argocd"

	"github.com/spf13/cobra"
)

var repoServerUrl *string

var argoOutputDir *string
var repoUsername *string
var repoPassword *string
var repoSshKey *string

// argocdCmd represents the argocd command
var argocdCmd = &cobra.Command{
	Use:   "argocd [rootDir]",
	Short: "Render manifests from ArgoCD Application",
	Args:  cobra.ExactArgs(1),
	Long:  `Render manifests from ArgoCD Application`,
	Run: func(cmd *cobra.Command, args []string) {
		argocd.Render(args[0], *repoServerUrl, *argoOutputDir, argocd.RepoCredentails{
			Username: *repoUsername,
			Password: *repoPassword,
			KeyFile:  *repoSshKey,
		})
	},
}

func init() {
	rootCmd.AddCommand(argocdCmd)

	repoServerUrl = argocdCmd.Flags().String("repoServer", "localhost:8081", "URL to argoCD repo server")
	argoOutputDir = argocdCmd.Flags().StringP("output-dir", "o", ".", "Output directory")
	repoUsername = argocdCmd.Flags().StringP("username", "u", "", "Repo username")
	repoPassword = argocdCmd.Flags().StringP("password", "p", "", "Repo password")
	repoSshKey = argocdCmd.Flags().StringP("ssh-key", "i", "", "Repo SSH Key")
}
