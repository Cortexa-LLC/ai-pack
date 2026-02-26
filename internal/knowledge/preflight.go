package knowledge

import (
	"context"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/internal/mcp"
)

// PreflightContext calls the knowledge MCP server's get_preflight_context tool.
// Returns the markdown block (may be empty if tool returns no relevant context, or on error).
func PreflightContext(ctx context.Context, mcpManager *mcp.Manager, taskDescription string, projectID string) string {
	if mcpManager == nil {
		monitoring.Logger.Info("preflight_skipped", "reason", "no MCP manager defined")
		return ""
	}
	callCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	params := map[string]interface{}{
		"task_description": taskDescription,
		"project_id": projectID,
	}
	result, err := mcpManager.CallTool(callCtx, "get_preflight_context", params)
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
