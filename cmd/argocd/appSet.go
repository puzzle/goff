/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package argocd

import (
	"fmt"
	"goff/cmd"

	"github.com/spf13/cobra"
)

// argocdCmd represents the argocd command
var argocdAppSetCmd = &cobra.Command{
	Use:   "appSet [sourceDir]",
	Short: "Render manifests from ArgoCD ApplicationSets",
	Args:  cobra.ExactArgs(1),
	Long:  `Render manifests from ArgoCD ApplicationSets`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("called applicationSets")
	},
}

func init() {
	cmd.ArgocdCmd.AddCommand(argocdAppSetCmd)
}
