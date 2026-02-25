package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/cortexa-llc/ai-pack/internal/knowledge"
)

var fromID string
var toID string
var relType string

var linkCmd = &cobra.Command{
	Use:   "link <from-id> --rel <RELTYPE> <to-id>",
	Short: "Create a relation between entities",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := knowledge.OpenStore("./data/kg.kuzu")
		if err != nil {
			return err
		}
		fromID := args[0]
		toID := args[1]
		if relType == "" {
			return fmt.Errorf("missing --rel flag")
		}
		if err := store.CreateRelation(fromID, toID, relType, ""); err != nil {
			return err
		}
		fmt.Printf("Linked %s --%s--> %s\n", fromID, relType, toID)
		return nil
	},
}

func init() {
	linkCmd.Flags().StringVar(&relType, "rel", "", "Relation type")
	linkCmd.MarkFlagRequired("rel")
	rootCmd.AddCommand(linkCmd)
}
