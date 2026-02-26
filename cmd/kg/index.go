package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/cortexa-llc/ai-pack/internal/knowledge"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Index the codebase and populate the knowledge graph with structural data",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		store, err := knowledge.OpenStore(".ai/knowledge.db")
		if err != nil {
			return fmt.Errorf("open Kuzu store: %w", err)
		}
		projectID := filepath.Base(cwd)
		indexer, err := knowledge.NewIndexer(store, projectID, cwd)
		if err != nil {
			return err
		}
		fmt.Printf("🔍 Indexing codebase at %s...\n", cwd)
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
