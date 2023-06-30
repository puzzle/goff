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
	Use:   "appset [sourceDir]",
	Short: "Render ArgoCD Applications manifests from ApplicationSets",
	Args:  cobra.ExactArgs(1),
	Long:  `Render ArgoCD Applications manifests from ArgoCD ApplicationSets`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return argocd.RenderApplicationSets(args[0], *ArgoOutputDir)
	},
}
