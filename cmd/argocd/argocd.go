/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package argocd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ArgoOutputDir *string
var RepoUsername *string
var RepoPassword *string
var RepoSshKey *string

// argocdCmd represents the argocd command
var ArgocdCmd = &cobra.Command{
	Use:   "argocd",
	Short: "Render manifests from ArgoCD resources",
	Long:  `Render manifests from ArgoCD resources`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("specify one subcommand [app|appset]")
	},
}

func init() {
	ArgocdCmd.AddCommand(ArgocdAppCmd)
	ArgocdCmd.AddCommand(ArgocdAppSetCmd)

	ArgoOutputDir = ArgocdCmd.PersistentFlags().StringP("output-dir", "o", ".", "Output directory")
	RepoUsername = ArgocdCmd.PersistentFlags().StringP("username", "u", "", "Repo username")
	RepoPassword = ArgocdCmd.PersistentFlags().StringP("password", "p", "", "Repo password")
	RepoSshKey = ArgocdCmd.PersistentFlags().StringP("ssh-key", "i", "", "Repo SSH Key")
}
