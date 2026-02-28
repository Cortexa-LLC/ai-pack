package kgclient

import (
	"bufio"
	"context"
	"os"
	"strings"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/mcp"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
)

// ParseRelatedProjects scans a 00-contract.md file for lines matching
// "Related Projects: /path/to/project" and returns the list of paths.
// Lines that don't start with "Related Projects:" are ignored.
// If the file cannot be read, an empty slice is returned silently.
func ParseRelatedProjects(contractPath string) []string {
	f, err := os.Open(contractPath)
	if err != nil {
		return nil
	}
	defer f.Close()

	var paths []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		const prefix = "Related Projects:"
		if strings.HasPrefix(line, prefix) {
			rest := strings.TrimSpace(line[len(prefix):])
			if rest != "" {
				paths = append(paths, rest)
			}
		}
	}
	return paths
}

// PreflightContext calls the knowledge MCP server's get_preflight_context tool.
// Returns the markdown block (may be empty if tool returns no relevant context, or on error).
func PreflightContext(ctx context.Context, mcpManager *mcp.Manager, taskDescription string, projectRoot string) string {
	if mcpManager == nil {
		monitoring.Logger.Info("preflight_skipped", "reason", "no MCP manager defined")
		return ""
	}
	callCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	params := map[string]interface{}{
		"task": taskDescription,
	}
	result, err := mcpManager.CallToolForProject(callCtx, projectRoot, "get_preflight_context", params)
	if err != nil {
		monitoring.Logger.Info("preflight_skipped", "reason", err.Error())
		return ""
	}
	if result == nil || len(result.Content) == 0 {
		monitoring.Logger.Info("preflight_skipped", "reason", "empty content")
		return ""
	}
	// Prepend all text blocks in order.
	md := ""
	for _, cb := range result.Content {
		if cb.Type == "text" && cb.Text != "" {
			md += cb.Text + "\n"
		}
	}
	if md == "" {
		monitoring.Logger.Info("preflight_skipped", "reason", "no text blocks")
		return ""
	}
	monitoring.Logger.Info("preflight_context_injected", "chars", len(md))
	return md
}
