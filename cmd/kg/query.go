package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/cortexa-llc/ai-pack/internal/knowledge"
)

var queryCmd = &cobra.Command{
	Use:   "query [cypher-query]",
	Short: "Run a raw Cypher query in the Kuzu graph store",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := knowledge.OpenStore("./data/kg.kuzu")
		if err != nil {
			return err
		}
		results, err := store.Execute(args[0])
		if err != nil {
			return err
		}
		fmt.Println(results)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(queryCmd)
}
