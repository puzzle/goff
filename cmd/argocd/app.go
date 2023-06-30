/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package argocd

import (
	"github.com/puzzle/goff/argocd"
	"github.com/spf13/cobra"
)

var repoServerUrl *string

// argocdCmd represents the argocd command
var ArgocdAppCmd = &cobra.Command{
	Use:   "app [rootDir]",
	Short: "Render manifests from ArgoCD Application",
	Args:  cobra.ExactArgs(1),
	Long:  `Render manifests from ArgoCD Application`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return argocd.Render(args[0], *repoServerUrl, *ArgoOutputDir, argocd.RepoCredentails{
			Username: *RepoUsername,
			Password: *RepoPassword,
			KeyFile:  *RepoSshKey,
		})
	},
}

func init() {

	repoServerUrl = ArgocdAppCmd.Flags().String("repo-server", "localhost:8081", "URL to argoCD repo server")

}
