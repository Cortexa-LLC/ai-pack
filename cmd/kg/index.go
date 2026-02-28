package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/knowledge"
	"github.com/spf13/cobra"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Index the codebase and populate the knowledge graph with structural data",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		root := findProjectRoot(cwd)
		dbPath := root + "/.ai/knowledge.db"
		store, err := knowledge.OpenStore(dbPath)
		if err != nil {
			return fmt.Errorf("open Kuzu store: %w", err)
		}
		defer store.Close()

		projectID := projectIDFromCwd(cwd)
		indexer, err := knowledge.NewIndexer(store, projectID, root)
		if err != nil {
			return err
		}
		fmt.Printf("🔍 Indexing codebase at %s...\n", root)
		start := time.Now()
		stats, err := indexer.Index()
		if err != nil {
			return err
		}
		dur := time.Since(start)
		fmt.Println("✅ Indexing complete!")
		fmt.Printf("   Files scanned:     %d\n", stats.FilesScanned)
		fmt.Printf("   Entities created:  %d\n", stats.EntitiesCreated)
		fmt.Printf("   Relations created: %d\n", stats.RelationsCreated)
		fmt.Printf("   Duration:          %.3fs\n", dur.Seconds())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(indexCmd)
}
