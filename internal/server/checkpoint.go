package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/streaming"
)

// AgentCheckpoint captures the full mutable state of the agentic loop
// at the moment the token budget was exhausted or timeout occurred.
// All fields are required for a successful resume.
type AgentCheckpoint struct {
	// Identity
	TaskID    string    `json:"task_id"`
	CreatedAt time.Time `json:"created_at"`

	// Loop counters (mirrors local variables in runAgentLoop)
	Turn                  int    `json:"turn"`
	TotalInputTokens      int64  `json:"total_input_tokens"`
	TotalOutputTokens     int64  `json:"total_output_tokens"`
	InactiveTurns         int    `json:"inactive_turns"`
	ConsecutiveErrorTurns int    `json:"consecutive_error_turns"` // number of turns where every tool call returned an error
	LastTextLength        int    `json:"last_text_length"`
	LastToolSignature     string `json:"last_tool_signature"`

	// Token budget snapshot
	BudgetLimit int64 `json:"budget_limit"` // original MaxBudgetTokens
	BudgetUsed  int64 `json:"budget_used"`  // TotalInputTokens + TotalOutputTokens at pause

	// Conversation history
	Messages []streaming.Message `json:"messages"`

	// Accumulated result text (finalResult.String() at pause point)
	PartialResult string `json:"partial_result"`

	// Pause reason: "token_budget" or "timeout"
	ResumeReason string `json:"resume_reason"`

	// Config snapshot — enough to recreate the agent; the full AgentConfig
	// is re-loaded from disk on resume, but MaxBudgetTokens is overridden
	// by the resume call's additionalBudget parameter.
	Role        string `json:"role"`
	ProjectRoot string `json:"project_root"`
	Model       string `json:"model"` // model name at pause time; may differ from current config if the model was changed between runs
}

const checkpointFileName = "checkpoint.json"

// writeCheckpoint serialises cp to {projectRoot}/.beads/tasks/{taskID}/checkpoint.json.
func writeCheckpoint(projectRoot, taskID string, cp *AgentCheckpoint) error {
	dir := filepath.Join(projectRoot, ".beads", "tasks", taskID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("checkpoint mkdir: %w", err)
	}
	data, err := json.MarshalIndent(cp, "", "  ")
	if err != nil {
		return fmt.Errorf("checkpoint marshal: %w", err)
	}
	path := filepath.Join(dir, checkpointFileName)
	// Write atomically: temp file + rename
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("checkpoint write: %w", err)
	}
	return os.Rename(tmp, path)
}

// loadCheckpoint reads and deserialises the checkpoint for taskID.
func loadCheckpoint(projectRoot, taskID string) (*AgentCheckpoint, error) {
	path := filepath.Join(projectRoot, ".beads", "tasks", taskID, checkpointFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("checkpoint read: %w", err)
	}
	var cp AgentCheckpoint
	if err := json.Unmarshal(data, &cp); err != nil {
		return nil, fmt.Errorf("checkpoint unmarshal: %w", err)
	}
	return &cp, nil
}

// checkpointPath returns the canonical filesystem path to the checkpoint file
// for the given task. Use this when only the path is needed (e.g. to include
// in an SSE event payload or to delete the file); use loadCheckpoint when
// the checkpoint data itself is required.
func checkpointPath(projectRoot, taskID string) string {
	return filepath.Join(projectRoot, ".beads", "tasks", taskID, checkpointFileName)
}
