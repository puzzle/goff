/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"goff/kustomize/kustomizationgraph"

	"github.com/spf13/cobra"
)

// kustomizeCmd represents the kustomize command
var kustomizeCmd = &cobra.Command{
	Use:   "kustomize",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

func run() {
	graph, err := kustomizationgraph.New("main").Generate("/home/schlatter/puzzle/goff/goff/testdata/")
	if err != nil {
		panic(err)
	}

	fmt.Print(graph)
}

func init() {
	rootCmd.AddCommand(kustomizeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// kustomizeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// kustomizeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
