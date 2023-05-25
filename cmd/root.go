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
	Long:  `GitOps Diff Tool`,
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

	logLevel = rootCmd.PersistentFlags().StringP("logLevel", "l", "error", "Set loglevel [debug, info, error]")
}
