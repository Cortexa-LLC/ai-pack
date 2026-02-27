package kgclient

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/knowledge"
	"github.com/cortexa-llc/ai-pack/internal/mcp"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
)

// filePath matches common file path patterns in agent output:
//   - absolute:  /some/path/to/file.go
//   - relative:  src/main.go  or  ./src/main.go
//   - modified:  "modified: foo/bar.go" lines from git diff
//
// The pattern requires an extension so we don't capture bare directory names.
var filePathRe = regexp.MustCompile(`(?m)(?:^|\s|["'` + "`" + `])(\./|/)?([A-Za-z0-9_\-][A-Za-z0-9_\-./]*\.[A-Za-z]{1,10})`)

// ExtractFilePaths scans text for file-path-like tokens and returns a
// deduplicated, sorted list. Only paths that contain at least one "/" are
// included so that lone words like "foo.go" in normal prose are not matched
// unless they look like paths.
func ExtractFilePaths(text string) []string {
	matches := filePathRe.FindAllStringSubmatch(text, -1)
	seen := make(map[string]struct{})
	var paths []string
	for _, m := range matches {
		// m[0] = full match, m[1] = optional ./ or /, m[2] = path body
		raw := strings.TrimSpace(m[0])
		// Trim surrounding punctuation that may have been caught
		raw = strings.Trim(raw, `"' `+"`")
		if raw == "" {
			continue
		}
		// Skip paths that are just an extension (e.g. ".go")
		if strings.HasPrefix(raw, ".") && !strings.Contains(raw, "/") {
			continue
		}
		if _, dup := seen[raw]; dup {
			continue
		}
		seen[raw] = struct{}{}
		paths = append(paths, raw)
	}
	return paths
}

// WriteBack persists a task-completion record to the knowledge graph via the
// knowledge MCP server. It runs best-effort: errors are logged but never
// returned (so callers are never blocked on KG failures).
//
// The entity name follows the convention "task:<role>@<RFC3339-timestamp>" so
// future preflight searches can find recent work by role or time.
func WriteBack(
	ctx context.Context,
	mcpManager *mcp.Manager,
	projectRoot string,
	role string,
	taskDescription string,
	fullOutput string,
	startTime time.Time,
) {
	if mcpManager == nil {
		monitoring.Logger.Debug("kg_writeback_skip", "reason", "no_mcp_manager", "project", projectRoot)
		return
	}

	// Use a short timeout so a slow KG server cannot stall task teardown.
	wbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	now := time.Now()
	entityName := fmt.Sprintf("task:%s@%s", role, now.UTC().Format(time.RFC3339))

	// 1. Create the entity
	var entity knowledge.Entity
	err := mcpManager.CallToolInto(wbCtx, "add_entity", map[string]interface{}{
		"name": entityName,
		"type": "task_outcome",
	}, &entity)
	if err != nil {
		monitoring.Logger.Warn("kg_writeback_create_entity_failed",
			"project", projectRoot,
			"role", role,
			"error", err.Error(),
		)
		return
	}

	entityID := entity.ID
	if entityID == "" {
		monitoring.Logger.Warn("kg_writeback_empty_entity_id",
			"project", projectRoot,
			"role", role,
		)
		return
	}

	// 2. Build observations
	observations := buildObservations(role, taskDescription, fullOutput, startTime, now)
	for _, obs := range observations {
		if err := mcpManager.CallToolInto(wbCtx, "add_observation", map[string]interface{}{
			"entity_id": entityID,
			"content":   obs,
		}, nil); err != nil {
			// Non-fatal: log and continue with remaining observations
			monitoring.Logger.Warn("kg_writeback_add_observation_failed",
				"entity_id", entityID,
				"error", err.Error(),
			)
		}
	}

	monitoring.Logger.Info("kg_writeback_complete",
		"entity_id", entityID,
		"entity_name", entityName,
		"project", projectRoot,
		"role", role,
		"observations", len(observations),
	)
}

// buildObservations constructs the set of observation strings for a task outcome.
func buildObservations(
	role string,
	taskDescription string,
	fullOutput string,
	startTime time.Time,
	endTime time.Time,
) []string {
	duration := endTime.Sub(startTime).Round(time.Second)

	var obs []string

	// Role
	obs = append(obs, fmt.Sprintf("role: %s", role))

	// Duration
	obs = append(obs, fmt.Sprintf("duration: %s", duration))

	// Summary: first 300 chars of the task description
	summary := strings.TrimSpace(taskDescription)
	if len(summary) > 300 {
		summary = summary[:300] + "…"
	}
	obs = append(obs, fmt.Sprintf("task: %s", summary))

	// Changed files extracted from output
	changed := ExtractFilePaths(fullOutput)
	if len(changed) > 0 {
		obs = append(obs, fmt.Sprintf("changed_files: %s", strings.Join(changed, ", ")))
	}

	// Truncated output tail (last 500 chars) as context for future agents
	outputTail := strings.TrimSpace(fullOutput)
	if len(outputTail) > 500 {
		outputTail = "…" + outputTail[len(outputTail)-500:]
	}
	if outputTail != "" {
		obs = append(obs, fmt.Sprintf("output_tail: %s", outputTail))
	}

	return obs
}
