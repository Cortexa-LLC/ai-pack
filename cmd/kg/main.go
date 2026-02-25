package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/knowledge"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "index":
		handleIndex(os.Args[2:])
	case "query":
		handleQuery(os.Args[2:])
	case "init":
		handleInit(os.Args[2:])
	case "help", "--help", "-h":
		usage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("Usage: kg <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  index    Scan codebase and populate knowledge graph")
	fmt.Println("  query    Execute Cypher query on knowledge graph")
	fmt.Println("  init     Initialize knowledge graph database")
	fmt.Println("  help     Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  kg init")
	fmt.Println("  kg index --root .")
	fmt.Println("  kg query \"MATCH (f:Entity {type:'file'}) RETURN count(f)\"")
}

func handleInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	dbPath := fs.String("db", ".kuzu/kg.db", "Database path")
	fs.Parse(args)

	store, err := knowledge.OpenStore(*dbPath)
	if err != nil {
		fmt.Printf("❌ Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	fmt.Printf("✅ Initialized knowledge graph at %s\n", *dbPath)
}

func handleIndex(args []string) {
	fs := flag.NewFlagSet("index", flag.ExitOnError)
	dbPath := fs.String("db", ".kuzu/kg.db", "Database path")
	root := fs.String("root", ".", "Root directory to scan")
	projectID := fs.String("project", "default", "Project ID")
	fs.Parse(args)

	// Open store
	store, err := knowledge.OpenStore(*dbPath)
	if err != nil {
		fmt.Printf("❌ Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	// Create indexer
	indexer, err := knowledge.NewIndexer(store, *projectID, *root)
	if err != nil {
		fmt.Printf("❌ Failed to create indexer: %v\n", err)
		os.Exit(1)
	}

	// Run indexing
	fmt.Printf("🔍 Indexing codebase at %s...\n", *root)
	startTime := time.Now()

	stats, err := indexer.Index()
	if err != nil {
		fmt.Printf("❌ Indexing failed: %v\n", err)
		os.Exit(1)
	}

	// Display results
	fmt.Println()
	fmt.Println("✅ Indexing complete!")
	fmt.Printf("   Files scanned:     %d\n", stats.FilesScanned)
	fmt.Printf("   Entities created:  %d\n", stats.EntitiesCreated)
	fmt.Printf("   Relations created: %d\n", stats.RelationsCreated)
	fmt.Printf("   Duration:          %v\n", time.Since(startTime).Round(time.Millisecond))
}

func handleQuery(args []string) {
	fs := flag.NewFlagSet("query", flag.ExitOnError)
	dbPath := fs.String("db", ".kuzu/kg.db", "Database path")
	fs.Parse(args)

	if len(fs.Args()) < 1 {
		fmt.Println("Usage: kg query <cypher-query>")
		os.Exit(1)
	}

	query := fs.Args()[0]

	// Open store
	store, err := knowledge.OpenStore(*dbPath)
	if err != nil {
		fmt.Printf("❌ Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	// Execute query
	result, err := store.Execute(query)
	if err != nil {
		fmt.Printf("❌ Query failed: %v\n", err)
		os.Exit(1)
	}

	// Display results
	if result.HasNext() {
		fmt.Println("Results:")
		for result.HasNext() {
			row, err := result.Next()
			if err != nil {
				fmt.Printf("❌ Error reading result: %v\n", err)
				break
			}
			fmt.Println(row)
		}
	} else {
		fmt.Println("(no results)")
	}
}

func getProjectRoot() string {
	// Try to find .git directory
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "."
}
