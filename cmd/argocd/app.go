/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package argocd

import (
	"goff/argocd"
	pCmd "goff/cmd"

	"github.com/spf13/cobra"
)

var repoServerUrl *string

// argocdCmd represents the argocd command
var argocdAppCmd = &cobra.Command{
	Use:   "app [sourceDir]",
	Short: "Render manifests from ArgoCD Applications",
	Args:  cobra.ExactArgs(1),
	Long:  `Render manifests from ArgoCD Applications`,
	Run: func(cmd *cobra.Command, args []string) {
		argocd.Render(args[0], *repoServerUrl, *pCmd.ArgoOutputDir)
	},
}

func init() {
	pCmd.ArgocdCmd.AddCommand(argocdAppCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// argocdCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	repoServerUrl = argocdAppCmd.Flags().String("repoServer", "localhost:8081", "URL to argoCD repo server")
}
