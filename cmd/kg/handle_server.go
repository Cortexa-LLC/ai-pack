package main

import (
	"fmt"
	"os"

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

	store, projectID, err := openStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "kg server: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	if err := knowledge.RunMCPServer(store, projectID); err != nil {
		fmt.Fprintf(os.Stderr, "kg server: MCP server error: %v\n", err)
		os.Exit(2)
	}
}
