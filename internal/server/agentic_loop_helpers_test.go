package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cortexa-llc/ai-pack/internal/streaming"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

func nopLog(string) {}

func logLines(lines *[]string) func(string) {
	return func(s string) { *lines = append(*lines, s) }
}

// ─── applyTokenBudgetCheck ───────────────────────────────────────────────────

func TestApplyTokenBudgetCheck_NoLimit(t *testing.T) {
	s := &AgentServer{}
	cfg := &AgentConfig{}
	cfg.Delegation.MaxBudgetTokens = 0

	result, err := s.applyTokenBudgetCheck(
		"task1", t.TempDir(), "engineer", cfg,
		1, 5000, 3000, 0, 0, 0, "", nil, "", nopLog,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != tokenBudgetContinue {
		t.Errorf("expected tokenBudgetContinue, got %d", result)
	}
}

func TestApplyTokenBudgetCheck_BelowLimit(t *testing.T) {
	s := &AgentServer{}
	cfg := &AgentConfig{}
	cfg.Delegation.MaxBudgetTokens = 100_000

	result, err := s.applyTokenBudgetCheck(
		"task1", t.TempDir(), "engineer", cfg,
		1, 5000, 3000, 0, 0, 0, "", nil, "", nopLog,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != tokenBudgetContinue {
		t.Errorf("expected tokenBudgetContinue, got %d", result)
	}
}

func TestApplyTokenBudgetCheck_Warn80Pct(t *testing.T) {
	s := &AgentServer{}
	cfg := &AgentConfig{}
	cfg.Delegation.MaxBudgetTokens = 10_000

	var logs []string
	result, err := s.applyTokenBudgetCheck(
		"task1", t.TempDir(), "engineer", cfg,
		1, 8_500, 100, 0, 0, 0, "", nil, "", logLines(&logs),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != tokenBudgetContinue {
		t.Errorf("expected tokenBudgetContinue, got %d", result)
	}
	// Should have logged a warning.
	warned := false
	for _, l := range logs {
		if len(l) > 0 {
			warned = true
			break
		}
	}
	if !warned {
		t.Error("expected a warning log but got none")
	}
}

func TestApplyTokenBudgetCheck_BudgetExhausted(t *testing.T) {
	projectRoot := t.TempDir()
	// writeCheckpoint needs the .beads/tasks/<taskID>/ directory.
	taskID := "task-budget-test"
	if err := os.MkdirAll(filepath.Join(projectRoot, ".beads", "tasks", taskID), 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	s := &AgentServer{}
	cfg := &AgentConfig{Model: "claude-opus-4"}
	cfg.Delegation.MaxBudgetTokens = 1_000

	result, err := s.applyTokenBudgetCheck(
		taskID, projectRoot, "engineer", cfg,
		5, 800, 300, // used = 1100 > 1000
		2, 0, 42, "sig-xyz",
		[]streaming.Message{{Role: "user", Content: "hello"}},
		"partial result",
		nopLog,
	)

	if !errors.Is(err, ErrTaskPaused) {
		t.Fatalf("expected ErrTaskPaused, got %v", err)
	}
	if result != tokenBudgetPaused {
		t.Errorf("expected tokenBudgetPaused, got %d", result)
	}

	// Verify checkpoint was written.
	cp, loadErr := loadCheckpoint(projectRoot, taskID)
	if loadErr != nil {
		t.Fatalf("loadCheckpoint: %v", loadErr)
	}
	if cp.Turn != 5 {
		t.Errorf("checkpoint.Turn: want 5, got %d", cp.Turn)
	}
	if cp.BudgetUsed != 1100 {
		t.Errorf("checkpoint.BudgetUsed: want 1100, got %d", cp.BudgetUsed)
	}
	if cp.BudgetLimit != 1000 {
		t.Errorf("checkpoint.BudgetLimit: want 1000, got %d", cp.BudgetLimit)
	}
	if cp.PartialResult != "partial result" {
		t.Errorf("checkpoint.PartialResult: want 'partial result', got %q", cp.PartialResult)
	}
}

func TestApplyTokenBudgetCheck_ExactlyAtLimit(t *testing.T) {
	projectRoot := t.TempDir()
	taskID := "task-exact"
	if err := os.MkdirAll(filepath.Join(projectRoot, ".beads", "tasks", taskID), 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	s := &AgentServer{}
	cfg := &AgentConfig{}
	cfg.Delegation.MaxBudgetTokens = 1_000

	// used == limit (exactly at boundary)
	result, err := s.applyTokenBudgetCheck(
		taskID, projectRoot, "engineer", cfg,
		1, 500, 500, 0, 0, 0, "", nil, "", nopLog,
	)
	if !errors.Is(err, ErrTaskPaused) {
		t.Fatalf("expected ErrTaskPaused at exactly limit, got %v", err)
	}
	if result != tokenBudgetPaused {
		t.Errorf("expected tokenBudgetPaused, got %d", result)
	}
}

// ─── processOneTurn ──────────────────────────────────────────────────────────

// fakeServer is a minimal AgentServer with an overridable executeTool.
type fakeServer struct {
	AgentServer
	toolFunc func(ctx context.Context, name string, input map[string]interface{}, dir string) (string, error)
}

// Override executeTool for the fakeServer by embedding and dispatching through a hook.
// Because executeTool is a method we can't easily mock it — instead we exercise
// processOneTurn through the real AgentServer with a nil MCP client, which
// causes executeTool to fall through to its native dispatch. For TaskComplete we
// don't even reach executeTool, so we can test that path unconditionally.

func makeTestServer() *AgentServer {
	return &AgentServer{maxInactiveTurns: 5}
}

func TestProcessOneTurn_TaskComplete(t *testing.T) {
	s := makeTestServer()

	toolUses := []streaming.ToolUse{
		{ID: "tu1", Name: "TaskComplete", Input: map[string]interface{}{"summary": "all done"}},
	}

	res := s.processOneTurn(context.Background(), toolUses, "/tmp", nopLog)

	if res.CompletionSummary != "all done" {
		t.Errorf("CompletionSummary: want 'all done', got %q", res.CompletionSummary)
	}
	if len(res.ToolResults) != 1 {
		t.Fatalf("expected 1 ToolResult, got %d", len(res.ToolResults))
	}
	if res.ToolResults[0].IsError {
		t.Error("TaskComplete result should not be flagged as error")
	}
}

func TestProcessOneTurn_TaskComplete_NoSummary(t *testing.T) {
	s := makeTestServer()

	toolUses := []streaming.ToolUse{
		{ID: "tu1", Name: "TaskComplete", Input: map[string]interface{}{}},
	}

	res := s.processOneTurn(context.Background(), toolUses, "/tmp", nopLog)

	if res.CompletionSummary != "(no summary provided)" {
		t.Errorf("expected default summary, got %q", res.CompletionSummary)
	}
}

func TestProcessOneTurn_NoTools(t *testing.T) {
	s := makeTestServer()

	res := s.processOneTurn(context.Background(), nil, "/tmp", nopLog)

	if res.CompletionSummary != "" {
		t.Errorf("expected empty CompletionSummary, got %q", res.CompletionSummary)
	}
	if len(res.ToolResults) != 0 {
		t.Errorf("expected no ToolResults, got %d", len(res.ToolResults))
	}
}

func TestProcessOneTurn_ToolError(t *testing.T) {
	s := makeTestServer()
	// Use an unknown tool name so executeTool returns an error immediately.
	toolUses := []streaming.ToolUse{
		{ID: "tu1", Name: "NonExistentTool", Input: map[string]interface{}{}},
	}

	res := s.processOneTurn(context.Background(), toolUses, "/tmp", nopLog)

	if len(res.ToolResults) != 1 {
		t.Fatalf("expected 1 ToolResult, got %d", len(res.ToolResults))
	}
	if !res.ToolResults[0].IsError {
		t.Error("expected ToolResult.IsError = true for unknown tool")
	}
	if res.CompletionSummary != "" {
		t.Errorf("expected no completion summary on error, got %q", res.CompletionSummary)
	}
}

// ─── checkStallConditions ────────────────────────────────────────────────────

func makeToolResult(isErr bool) streaming.ToolResult {
	return streaming.ToolResult{IsError: isErr}
}

func makeToolUse(name, input string) streaming.ToolUse {
	return streaming.ToolUse{Name: name, Input: map[string]interface{}{"cmd": input}}
}

func TestCheckStallConditions_ProgressByText(t *testing.T) {
	toolResults := []streaming.ToolResult{makeToolResult(false)}
	toolUses := []streaming.ToolUse{makeToolUse("Read", "file.go")}

	newInactive, newConsecErr, newTextLen, newSig, sr := checkStallConditions(
		toolResults, toolUses,
		100, // currentTextLength (grew)
		50,  // lastTextLength
		"",  // lastToolSignature (same tool signature)
		2, 0, 5, nopLog,
	)

	if sr != stallNone {
		t.Errorf("expected stallNone, got %d", sr)
	}
	if newInactive != 0 {
		t.Errorf("expected inactive counter reset to 0, got %d", newInactive)
	}
	if newConsecErr != 0 {
		t.Errorf("expected consecutiveErrorTurns=0, got %d", newConsecErr)
	}
	if newTextLen != 100 {
		t.Errorf("expected newTextLen=100, got %d", newTextLen)
	}
	_ = newSig
}

func TestCheckStallConditions_ProgressByDifferentTool(t *testing.T) {
	toolResults := []streaming.ToolResult{makeToolResult(false)}
	toolUses := []streaming.ToolUse{makeToolUse("Write", "file.go")}

	_, _, _, _, sr := checkStallConditions(
		toolResults, toolUses,
		50, 50, // text did NOT grow
		"Read:{\"cmd\":\"file.go\"}", // different from Write
		0, 0, 5, nopLog,
	)

	if sr != stallNone {
		t.Errorf("expected stallNone, got %d", sr)
	}
}

func TestCheckStallConditions_NoProgress_BelowThreshold(t *testing.T) {
	toolResults := []streaming.ToolResult{makeToolResult(false)}
	toolUses := []streaming.ToolUse{makeToolUse("Read", "file.go")}
	sig := fmt.Sprintf("Read:{\"cmd\":\"file.go\"}")

	newInactive, _, _, _, sr := checkStallConditions(
		toolResults, toolUses,
		50, 50, // no text growth
		sig,    // same signature
		1, 0, 5, nopLog,
	)

	if sr != stallNone {
		t.Errorf("expected stallNone below threshold, got %d", sr)
	}
	if newInactive != 2 {
		t.Errorf("expected newInactive=2, got %d", newInactive)
	}
}

func TestCheckStallConditions_NoProgress_AtThreshold(t *testing.T) {
	toolResults := []streaming.ToolResult{makeToolResult(false)}
	toolUses := []streaming.ToolUse{makeToolUse("Read", "file.go")}
	sig := fmt.Sprintf("Read:{\"cmd\":\"file.go\"}")

	_, _, _, _, sr := checkStallConditions(
		toolResults, toolUses,
		50, 50, sig,
		4, // inactiveTurns — next increment brings it to 5 == maxInactiveTurns
		0, 5, nopLog,
	)

	if sr != stallAbort {
		t.Errorf("expected stallAbort, got %d", sr)
	}
}

func TestCheckStallConditions_AllToolsError_BelowThreshold(t *testing.T) {
	// consecErrors=0 → increments to 1, still below maxInactiveTurns(5) → no abort from error path.
	// Text did not grow and tool signature is the same → progress check increments inactiveTurns.
	// inactiveTurns was 0 → becomes 1, still < 5 → stallNone.
	toolResults := []streaming.ToolResult{makeToolResult(true)}
	toolUses := []streaming.ToolUse{makeToolUse("Read", "x")}

	newInactive, newConsecErr, _, _, sr := checkStallConditions(
		toolResults, toolUses,
		50, 50, "Read:{\"cmd\":\"x\"}",
		0, 0, 5, nopLog,
	)

	if sr != stallNone {
		t.Errorf("expected stallNone below threshold, got %d", sr)
	}
	if newConsecErr != 1 {
		t.Errorf("expected newConsecErr=1, got %d", newConsecErr)
	}
	if newInactive != 1 {
		t.Errorf("expected newInactive=1, got %d", newInactive)
	}
}

func TestCheckStallConditions_AllToolsError_AtThreshold(t *testing.T) {
	toolResults := []streaming.ToolResult{makeToolResult(true)}
	toolUses := []streaming.ToolUse{makeToolUse("Read", "x")}

	_, _, _, _, sr := checkStallConditions(
		toolResults, toolUses,
		50, 50, "Read:{\"cmd\":\"x\"}",
		0, 4, // consecutiveErrorTurns = 4; next increment brings to 5 == max
		5, nopLog,
	)

	if sr != stallAbort {
		t.Errorf("expected stallAbort for consecutive errors, got %d", sr)
	}
}

func TestCheckStallConditions_MixedErrors_ResetsCounter(t *testing.T) {
	// One success among tool results — should reset consecutive error counter.
	toolResults := []streaming.ToolResult{makeToolResult(true), makeToolResult(false)}
	toolUses := []streaming.ToolUse{makeToolUse("Read", "x"), makeToolUse("Write", "y")}

	_, newConsecErr, _, _, _ := checkStallConditions(
		toolResults, toolUses,
		100, 50, "", // text grew => progress
		0, 3, 5, nopLog,
	)

	if newConsecErr != 0 {
		t.Errorf("expected consecutiveErrorTurns reset to 0, got %d", newConsecErr)
	}
}

// ─── flushCheckpoint ─────────────────────────────────────────────────────────

func TestFlushCheckpoint_WritesFile(t *testing.T) {
	projectRoot := t.TempDir()
	taskID := "task-flush"
	if err := os.MkdirAll(filepath.Join(projectRoot, ".beads", "tasks", taskID), 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	msgs := []streaming.Message{{Role: "user", Content: "hello"}}

	err := flushCheckpoint(
		projectRoot, taskID, "engineer", "claude-opus-4",
		7, 1200, 800,
		3, 1,
		42, "sig-abc",
		1000, 2000,
		msgs, "partial",
		nopLog,
	)
	if err != nil {
		t.Fatalf("flushCheckpoint: %v", err)
	}

	cp, loadErr := loadCheckpoint(projectRoot, taskID)
	if loadErr != nil {
		t.Fatalf("loadCheckpoint: %v", loadErr)
	}

	if cp.TaskID != taskID {
		t.Errorf("TaskID: want %q, got %q", taskID, cp.TaskID)
	}
	if cp.Turn != 7 {
		t.Errorf("Turn: want 7, got %d", cp.Turn)
	}
	if cp.TotalInputTokens != 1200 {
		t.Errorf("TotalInputTokens: want 1200, got %d", cp.TotalInputTokens)
	}
	if cp.TotalOutputTokens != 800 {
		t.Errorf("TotalOutputTokens: want 800, got %d", cp.TotalOutputTokens)
	}
	if cp.InactiveTurns != 3 {
		t.Errorf("InactiveTurns: want 3, got %d", cp.InactiveTurns)
	}
	if cp.ConsecutiveErrorTurns != 1 {
		t.Errorf("ConsecutiveErrorTurns: want 1, got %d", cp.ConsecutiveErrorTurns)
	}
	if cp.LastTextLength != 42 {
		t.Errorf("LastTextLength: want 42, got %d", cp.LastTextLength)
	}
	if cp.LastToolSignature != "sig-abc" {
		t.Errorf("LastToolSignature: want 'sig-abc', got %q", cp.LastToolSignature)
	}
	if cp.BudgetLimit != 1000 {
		t.Errorf("BudgetLimit: want 1000, got %d", cp.BudgetLimit)
	}
	if cp.BudgetUsed != 2000 {
		t.Errorf("BudgetUsed: want 2000, got %d", cp.BudgetUsed)
	}
	if cp.PartialResult != "partial" {
		t.Errorf("PartialResult: want 'partial', got %q", cp.PartialResult)
	}
	if cp.Role != "engineer" {
		t.Errorf("Role: want 'engineer', got %q", cp.Role)
	}
	if cp.Model != "claude-opus-4" {
		t.Errorf("Model: want 'claude-opus-4', got %q", cp.Model)
	}
}

func TestFlushCheckpoint_ErrorOnBadPath(t *testing.T) {
	// projectRoot does not exist => writeCheckpoint should fail.
	err := flushCheckpoint(
		"/does/not/exist", "task1", "engineer", "model",
		1, 0, 0, 0, 0, 0, "",
		0, 0, nil, "", nopLog,
	)
	if err == nil {
		t.Error("expected error for non-existent projectRoot, got nil")
	}
}
