package main

import (
	"github.com/spf13/cobra"
	"os"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a kg server (optionally with --stdio for AI pack integration)",
	Run: func(cmd *cobra.Command, args []string) {
		handleServer(os.Args[2:])
	},
}

func init() {
}
