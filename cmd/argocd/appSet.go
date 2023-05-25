/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package argocd

import (
	"github.com/puzzle/goff/argocd"

	"github.com/spf13/cobra"
)

// argocdCmd represents the argocd command
var ArgocdAppSetCmd = &cobra.Command{
	Use:   "appSet [sourceDir]",
	Short: "Render manifests from ArgoCD ApplicationSets",
	Args:  cobra.ExactArgs(1),
	Long:  `Render manifests from ArgoCD ApplicationSets`,
	Run: func(cmd *cobra.Command, args []string) {
		argocd.RenderApplicationSet(args[0], *ArgoOutputDir)
	},
}
