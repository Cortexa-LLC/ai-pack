package main

import (
	"flag"
	"fmt"
	"github.com/cortexa-llc/ai-pack/internal/knowledge"
	"os"
)

func handleServer(args []string) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	useStdio := fs.Bool("stdio", false, "Serve over stdio (MCP mode)")
	projectID := fs.String("project", "default", "Project ID (default: 'default')")
	_ = fs.Parse(args)
	if *useStdio {
		store, err := knowledge.OpenStore(".ai/knowledge.db")
		if err != nil {
			fmt.Fprintf(os.Stderr, "kg server: could not open store: %v\n", err)
			os.Exit(1)
		}
		err = knowledge.RunMCPServer(store, *projectID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "kg server: MCP server error: %v\n", err)
			os.Exit(2)
		}
		return
	}
	fmt.Println("kg server: --stdio flag required (MCP mode)")
	os.Exit(1)
}
