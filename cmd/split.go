/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/puzzle/goff/util"

	"github.com/spf13/cobra"
)

var outputSplitDir *string

// diffCmd represents the diff command
var splitCmd = &cobra.Command{
	Use:   "split",
	Short: "Split manifests [manifestFile]",
	Args:  cobra.ExactArgs(1),
	Long:  `Split multi document yaml into single files`,
	Run: func(cmd *cobra.Command, args []string) {
		err := util.SplitManifests(args[0], *outputSplitDir)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(splitCmd)
	outputSplitDir = splitCmd.Flags().StringP("output-dir", "o", ".", "Output directory")
}
