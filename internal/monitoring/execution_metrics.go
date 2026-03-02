package monitoring

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// ExecutionMetrics captures a single agent task execution's performance snapshot.
// It is written as metrics.json inside the task packet directory so that
// analysis tools can compute aggregate statistics without re-parsing logs.
type ExecutionMetrics struct {
	// TaskID is the Beads task identifier (e.g. "ai-pack-abc1").
	TaskID string `json:"task_id"`

	// Role is the agent role that ran (e.g. "engineer", "architect").
	Role string `json:"role"`

	// StartTime is when the task execution began (UTC).
	StartTime time.Time `json:"start_time"`

	// EndTime is when the task execution completed (UTC).
	EndTime time.Time `json:"end_time"`

	// DurationMs is the wall-clock duration of the execution in milliseconds.
	DurationMs int64 `json:"duration_ms"`

	// Turns is the number of agent turns (LLM round-trips) taken.
	Turns int `json:"turns"`

	// TotalTokens is the total input + output token count for the execution.
	TotalTokens int `json:"total_tokens"`

	// ToolCallsTotal is the total number of tool invocations across all turns.
	ToolCallsTotal int `json:"tool_calls_total"`

	// KgPreflightBytes is the byte length of the KG context block injected into
	// the system prompt before the first turn. Zero means KG was absent or empty.
	KgPreflightBytes int `json:"kg_preflight_bytes"`

	// ExplorationRatio is the fraction of tool calls that were read-only
	// exploration calls (Read, Grep, Glob, search_nodes, open_nodes, etc.)
	// relative to total tool calls. Ranges [0, 1]; -1 when tool_calls_total == 0.
	ExplorationRatio float64 `json:"exploration_ratio"`

	// HasErrors indicates whether any error patterns were detected in the log.
	HasErrors bool `json:"has_errors"`
}

// explorationToolNames is the set of tool names that are classified as
// read-only / exploratory for the purposes of ExplorationRatio.
var explorationToolNames = map[string]bool{
	"Read":           true,
	"Grep":           true,
	"Glob":           true,
	"search_nodes":   true,
	"open_nodes":     true,
	"search_knowledge": true,
	"query_graph":    true,
	"get_file_context": true,
}

// ComputeExplorationRatio returns the fraction of tool calls in toolCounts that
// are classified as exploratory. Returns -1 if totalCalls is zero.
func ComputeExplorationRatio(toolCounts map[string]int, totalCalls int) float64 {
	if totalCalls == 0 {
		return -1
	}
	exploratoryCount := 0
	for name, count := range toolCounts {
		if explorationToolNames[name] {
			exploratoryCount += count
		}
	}
	return float64(exploratoryCount) / float64(totalCalls)
}

// WriteMetrics serialises m as indented JSON and writes it to
// <taskPacketDir>/metrics.json, creating the file if it does not exist and
// overwriting it if it does. The directory must already exist.
// Errors are returned to the caller; callers that want best-effort behaviour
// should log and discard the error.
func WriteMetrics(taskPacketDir string, m ExecutionMetrics) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	dest := filepath.Join(taskPacketDir, "metrics.json")
	return os.WriteFile(dest, data, 0o644)
}
