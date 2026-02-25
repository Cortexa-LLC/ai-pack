package main

import (
	"fmt"
	"strings"

	"github.com/cortexa-llc/ai-pack/internal/knowledge"
	"github.com/spf13/cobra"
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
		defer store.Close()

		results, err := store.Execute(args[0])
		if err != nil {
			return err
		}
		defer results.Close()

		for results.HasNext() {
			tuple, err := results.Next()
			if err != nil {
				return fmt.Errorf("failed to get next row: %w", err)
			}
			row, err := tuple.GetAsSlice()
			if err != nil {
				return fmt.Errorf("failed to read row: %w", err)
			}
			var values []string
			for _, val := range row {
				values = append(values, fmt.Sprintf("%v", val))
			}
			fmt.Println(strings.Join(values, "\t"))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(queryCmd)
}
