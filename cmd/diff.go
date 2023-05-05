/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"goff/diff"

	"github.com/spf13/cobra"
)

var markdown *string
var title *string
var outputDir *string

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Diff files",
	Args:  cobra.ExactArgs(2),
	Long:  `Generate diff between directories`,
	Run: func(cmd *cobra.Command, args []string) {
		diff.Diff(*title, *markdown, args[0], args[1], *outputDir)
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// diffCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	markdown = diffCmd.Flags().StringP("markdown", "m", "markdown", "Markdown template")
	title = diffCmd.Flags().StringP("title", "t", "title", "Title for markdown")
	outputDir = diffCmd.Flags().StringP("output-dir", "o", ".", "Output directory")
}
