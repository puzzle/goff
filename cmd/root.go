/*
Copyright © 2023 Ch. Schlatter schlatter@puzzle.ch

*/
package cmd

import (
	"os"

	"github.com/puzzle/goff/cmd/argocd"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var logLevel *string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "goff",
	Short: "GitOps Diff Tool",
	Long:  `Helper tool to show changes between .....`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		panic(err)
	}

	log.SetLevel(level)
	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

	rootCmd.AddCommand(argocd.ArgocdCmd)

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	logLevel = rootCmd.PersistentFlags().StringP("logLevel", "l", "error", "Set loglevel [debug, info, error]")
}
