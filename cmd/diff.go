/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/puzzle/goff/diff"

	"github.com/spf13/cobra"
)

var markdown *string
var title *string
var outputDir *string
var glob *string
var exitCode *int

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Diff files [sourceDir] [targetDir]",
	Args:  cobra.ExactArgs(2),
	Long:  `Generate diff between two directories. You can use the --include optoin to include or exclude certain files with a glob pattern`,
	RunE: func(cmd *cobra.Command, args []string) error {
		found, err := diff.Diff(*title, *markdown, args[0], args[1], *glob, *outputDir, *exitCode)
		if err != nil {
			return err
		}

		if !found && *exitCode != 0 {
			os.Exit(*exitCode)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)

	markdown = diffCmd.Flags().StringP("markdown", "m", "markdown", "Markdown template")
	title = diffCmd.Flags().StringP("title", "t", "Preview", "Title for markdown")
	outputDir = diffCmd.Flags().StringP("output-dir", "o", ".", "Output directory")
	exitCode = diffCmd.Flags().IntP("exit-code", "x", 0, "Exit code if no diff is found")
	glob = diffCmd.Flags().String("include", "*.yaml", "Define glob pattern to include files")
}
