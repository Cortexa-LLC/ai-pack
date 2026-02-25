package main

import (
	"os"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kg",
	Short: "Knowledge graph CLI for Kuzu store",
	Long:  "The kg CLI provides commands to interact with a Kuzu-backed knowledge graph.",
}

func init() {
	rootCmd.AddCommand(addEntityCmd)
	addCmd.AddCommand(addObservationCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(linkCmd)
	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(indexCmd)
rootCmd.AddCommand(serverCmd)
rootCmd.AddCommand(exportCmd)
rootCmd.AddCommand(gcCmd)
rootCmd.AddCommand(embedCmd)
rootCmd.AddCommand(graphCmd)
	rootCmd.AddCommand(statsCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
