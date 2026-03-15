package kgclient

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cortexa-llc/ai-pack/internal/constants"
	"github.com/cortexa-llc/ai-pack/internal/knowledge"
	"github.com/cortexa-llc/ai-pack/internal/mcp"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
)

// toolLine matches log lines like:
//
//	[13:07:21]       🔧 Tool: Read(internal/server/foo.go)
//	[13:07:21]       🔧 Tool: Bash(cd /foo && go test)
//	[13:07:21]       🔧 Tool: Write(internal/kgclient/log_indexer.go)
var toolLine = regexp.MustCompile(`🔧 Tool: ([A-Za-z_][A-Za-z0-9_]*)`)

// turnLine matches log lines like:
//
//	[13:07:21]    Turn 5 (inactive: 0)...
var turnLine = regexp.MustCompile(`Turn (\d+) \(inactive:`)

// cumulativeTokens matches log lines like:
//
//	[13:07:21]       API: 8474ms | in:17867 out:81 | cumulative:17948
var cumulativeTokens = regexp.MustCompile(`cumulative:(\d+)`)

// errorLine matches lines that indicate a tool returned an error.
// We look for the "✗" marker (logged by processOneTurn on tool errors)
// as well as the "❌" emoji used elsewhere in the execution log.
var errorIndicator = regexp.MustCompile(`(?i)\b(error|fail|panic|❌|✗)\b`)

// ParsedLog holds the structured data extracted from an execution.log file.
type ParsedLog struct {
	// Turns is the highest turn number observed in the log.
	Turns int
	// ToolCounts maps each tool name to the number of times it was invoked.
	ToolCounts map[string]int
	// FilesTouched contains deduplicated file paths extracted from Write/Edit/Read
	// tool invocations in the log.
	FilesTouched []string
	// HasErrors is true if any error indicator was found in the log.
	HasErrors bool
	// TotalTokens is the last "cumulative" token count seen in the log,
	// which represents the total tokens used across all API calls.
	TotalTokens int64
	// KgPreflightBytes is the byte length of the KG context block injected into
	// the system prompt. Zero means KG was absent or returned an empty block.
	KgPreflightBytes int
}

// parseExecutionLog reads and parses an execution.log file, extracting
// structured data that can be indexed into the knowledge graph.
func parseExecutionLog(logPath string) (*ParsedLog, error) {
	f, err := os.Open(logPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	parsed := &ParsedLog{
		ToolCounts: make(map[string]int),
	}

	filesSeen := make(map[string]struct{})
	scanner := bufio.NewScanner(f)

	// Use a 2 MB buffer so long tool result lines don't truncate.
	const maxScanBuf = 2 * 1024 * 1024
	buf := make([]byte, maxScanBuf)
	scanner.Buffer(buf, maxScanBuf)

	for scanner.Scan() {
		line := scanner.Text()

		// Tool invocation
		if m := toolLine.FindStringSubmatch(line); m != nil {
			name := m[1]
			parsed.ToolCounts[name]++

			// Extract file paths from Read/Write/Edit tool lines.
			// The format is "Tool: Read(path)" or "Tool: Write(path)".
			if name == "Read" || name == "Write" || name == "Edit" || name == "MultiEdit" {
				// Grab the argument portion after "Tool: <name>("
				prefix := fmt.Sprintf("🔧 Tool: %s(", name)
				idx := strings.Index(line, prefix)
				if idx >= 0 {
					arg := line[idx+len(prefix):]
					// The argument runs up to the last ')'. For paths there is
					// no comma in the path itself, so just take what's before any
					// closing paren or comma.
					if end := strings.IndexAny(arg, ",)"); end > 0 {
						arg = strings.TrimSpace(arg[:end])
					}
					// Only record if it looks like a file path (has a dot extension)
					if strings.ContainsRune(arg, '.') && !strings.ContainsRune(arg, ' ') {
						if _, seen := filesSeen[arg]; !seen {
							filesSeen[arg] = struct{}{}
							parsed.FilesTouched = append(parsed.FilesTouched, arg)
						}
					}
				}
			}
		}

		// Turn counter — record the maximum turn number seen.
		if m := turnLine.FindStringSubmatch(line); m != nil {
			if n, err := strconv.Atoi(m[1]); err == nil && n > parsed.Turns {
				parsed.Turns = n
			}
		}

		// Cumulative token usage — keep the last (highest) value.
		if m := cumulativeTokens.FindStringSubmatch(line); m != nil {
			if n, err := strconv.ParseInt(m[1], 10, 64); err == nil {
				parsed.TotalTokens = n
			}
		}

		// Error indicator — mark if any error-related content appears.
		if !parsed.HasErrors && errorIndicator.MatchString(line) {
			// Exclude common false positives: lines that talk about error handling
			// in code excerpts (e.g. "if err != nil"), rather than actual failures.
			lowered := strings.ToLower(line)
			if strings.Contains(lowered, "err != nil") ||
				strings.Contains(lowered, "return err") ||
				strings.Contains(lowered, "error handling") ||
				strings.Contains(lowered, "\"error\"") ||
				strings.Contains(lowered, "`error`") {
				// Likely code context, not an actual runtime error.
			} else {
				parsed.HasErrors = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning execution log: %w", err)
	}

	// Sort files for deterministic output.
	sort.Strings(parsed.FilesTouched)

	return parsed, nil
}

// buildLogObservations builds the list of KG observation strings from a
// parsed execution log.  The result is deterministic so tests can assert on it.
func buildLogObservations(parsed *ParsedLog, role, taskDesc string, startTime time.Time, endTime time.Time) []string {
	var obs []string

	// Role
	obs = append(obs, "role: "+role)

	// Task description (truncated to 300 chars)
	const maxDesc = 300
	desc := taskDesc
	if utf8.RuneCountInString(desc) > maxDesc {
		runes := []rune(desc)
		desc = string(runes[:maxDesc]) + "…"
	}
	obs = append(obs, "task: "+desc)

	// Duration
	dur := endTime.Sub(startTime).Truncate(time.Second)
	obs = append(obs, fmt.Sprintf("duration: %s", dur))

	// Turns
	obs = append(obs, fmt.Sprintf("turns: %d", parsed.Turns))

	// Total tokens
	if parsed.TotalTokens > 0 {
		obs = append(obs, fmt.Sprintf("total_tokens: %d", parsed.TotalTokens))
	}

	// Tool call summary — sorted by name for determinism.
	if len(parsed.ToolCounts) > 0 {
		names := make([]string, 0, len(parsed.ToolCounts))
		for name := range parsed.ToolCounts {
			names = append(names, name)
		}
		sort.Strings(names)
		parts := make([]string, 0, len(names))
		for _, name := range names {
			parts = append(parts, fmt.Sprintf("%s:%d", name, parsed.ToolCounts[name]))
		}
		obs = append(obs, "tool_calls: "+strings.Join(parts, ", "))
	}

	// Files touched
	if len(parsed.FilesTouched) > 0 {
		obs = append(obs, "files_touched: "+strings.Join(parsed.FilesTouched, ", "))
	}

	// KG preflight context size
	if parsed.KgPreflightBytes > 0 {
		obs = append(obs, fmt.Sprintf("kg_preflight_bytes: %d", parsed.KgPreflightBytes))
	}

	// Errors
	if parsed.HasErrors {
		obs = append(obs, "had_errors: true")
	}

	return obs
}

// IndexExecutionLog reads the execution log for taskID and indexes it into the
// knowledge graph as a structured entity so future agents can learn from this run.
//
// The entity name is "exec:<taskID>" so it can be found via preflight context
// searches.  This function is best-effort: errors are logged but never returned.
//
// Call signature is designed to mirror WriteBack so both can be invoked from
// saveAndCompleteTask.
func IndexExecutionLog(
	ctx context.Context,
	mcpManager *mcp.Manager,
	projectRoot string,
	taskID string,
	role string,
	taskDesc string,
	startTime time.Time,
) {
	if mcpManager == nil {
		return // no KG client configured — silently skip
	}

	logPath := filepath.Join(projectRoot, constants.BeadsDir, "tasks", taskID, "execution.log")
	parsed, err := parseExecutionLog(logPath)
	if err != nil {
		monitoring.Logger.Warn("kg_log_indexer_parse_failed",
			"task_id", taskID,
			"log_path", logPath,
			"error", err.Error())
		return
	}

	endTime := time.Now()
	entityName := "exec:" + taskID
	observations := buildLogObservations(parsed, role, taskDesc, startTime, endTime)

	// Create the entity in the KG.  Type "topic" is the lightweight container used
	// for freeform knowledge entries.
	indexCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var entity knowledge.Entity
	err = mcpManager.CallToolIntoForProject(indexCtx, projectRoot, "add_entity", map[string]interface{}{
		"name": entityName,
		"type": "topic",
	}, &entity)
	if err != nil {
		monitoring.Logger.Warn("kg_log_indexer_add_entity_failed",
			"task_id", taskID,
			"error", err.Error())
		return
	}
	if entity.ID == "" {
		monitoring.Logger.Warn("kg_log_indexer_empty_entity_id", "task_id", taskID)
		return
	}

	// Attach each observation separately — same pattern as WriteBack.
	for _, o := range observations {
		if err := mcpManager.CallToolIntoForProject(indexCtx, projectRoot, "add_observation", map[string]interface{}{
			"entity_id": entity.ID,
			"content":   o,
		}, nil); err != nil {
			monitoring.Logger.Warn("kg_log_indexer_add_observation_failed",
				"task_id", taskID,
				"observation", o,
				"error", err.Error())
		}
	}

	monitoring.Logger.Info("kg_log_indexed",
		"task_id", taskID,
		"entity", entityName,
		"observations", len(observations))
}
