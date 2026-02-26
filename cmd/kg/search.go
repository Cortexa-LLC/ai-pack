package main

import (
	"fmt"
	"github.com/cortexa-llc/ai-pack/internal/knowledge"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search the knowledge graph",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		store, err := knowledge.OpenStore(".ai/knowledge.db")
		if err != nil {
			return fmt.Errorf("open store: %w", err)
		}
		results, err := store.HybridSearch("", query, nil, knowledge.DefaultSearchConfig())
		if err != nil {
			return err
		}
		for _, res := range results {
			fmt.Printf("%s\t%s\t%s\n", res.Entity.ID, res.Entity.Type, res.Entity.Name)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
