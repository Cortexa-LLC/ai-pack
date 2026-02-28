package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cortexa-llc/ai-pack/internal/knowledge"
)

func handleServer(args []string) {
	// Parse --stdio flag (all other flags obsolete now that project root is auto-detected)
	useStdio := false
	for _, a := range args {
		if a == "--stdio" {
			useStdio = true
		}
	}

	if !useStdio {
		fmt.Fprintln(os.Stderr, "kg server: --stdio flag required (MCP mode)")
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "kg server: get cwd: %v\n", err)
		os.Exit(1)
	}
	projectRoot := findProjectRoot(cwd)
	dbPath := filepath.Join(projectRoot, ".ai", "knowledge.db")
	projectID := filepath.Base(projectRoot)

	// Each MCP tool call opens the DB, operates, and closes it — no lock is
	// held between calls, so `kg index` and other CLI commands can run freely.
	if err := knowledge.RunMCPServer(dbPath, projectID, projectRoot); err != nil {
		fmt.Fprintf(os.Stderr, "kg server: MCP server error: %v\n", err)
		os.Exit(2)
	}
}
