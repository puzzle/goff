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
var exitCode *int

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Diff files [sourceDir] [targetDir]",
	Args:  cobra.ExactArgs(2),
	Long:  `Generate diff between two directories`,
	Run: func(cmd *cobra.Command, args []string) {
		diff.Diff(*title, *markdown, args[0], args[1], *outputDir, *exitCode)
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)

	markdown = diffCmd.Flags().StringP("markdown", "m", "markdown", "Markdown template")
	title = diffCmd.Flags().StringP("title", "t", "Preview", "Title for markdown")
	outputDir = diffCmd.Flags().StringP("output-dir", "o", ".", "Output directory")
	exitCode = diffCmd.Flags().IntP("exit-code", "x", 0, "Exit code if no diff is found")
}
