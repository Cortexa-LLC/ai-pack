package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/streaming"
)

// tokenBudgetResult is the return value of applyTokenBudgetCheck.
// It tells the caller whether to continue looping or stop.
type tokenBudgetResult int

const (
	tokenBudgetContinue tokenBudgetResult = iota // budget not yet reached — keep looping
	tokenBudgetPaused                            // budget exhausted — checkpoint written, loop must stop
)

// ─── applyTokenBudgetCheck ───────────────────────────────────────────────────

// applyTokenBudgetCheck enforces the per-role token budget defined in config.
// It is called once per turn, after token counters have been updated.
//
// When the budget is exhausted it:
//  1. logs a pause message,
//  2. writes a checkpoint so the task can be resumed later,
//  3. returns tokenBudgetPaused and ErrTaskPaused.
//
// When the budget is between 80 % and 100 % it logs a warning and returns
// tokenBudgetContinue, nil.
//
// When there is no budget (MaxBudgetTokens == 0) it returns
// tokenBudgetContinue, nil immediately.
func (s *AgentServer) applyTokenBudgetCheck(
	taskID string,
	projectRoot string,
	role string,
	config *AgentConfig,
	turn int,
	totalInputTokens int64,
	totalOutputTokens int64,
	inactiveTurns int,
	consecutiveErrorTurns int,
	lastTextLength int,
	lastToolSignature string,
	messages []streaming.Message,
	partialResult string,
	logMsg func(string),
) (tokenBudgetResult, error) {
	if config.Delegation.MaxBudgetTokens <= 0 {
		return tokenBudgetContinue, nil
	}

	used := totalInputTokens + totalOutputTokens
	limit := int64(config.Delegation.MaxBudgetTokens)

	if used >= limit {
		logMsg(fmt.Sprintf("⏸️  Token budget exhausted: %d/%d tokens used — pausing", used, limit))

		cp := &AgentCheckpoint{
			TaskID:                taskID,
			CreatedAt:             time.Now(),
			Turn:                  turn,
			TotalInputTokens:      totalInputTokens,
			TotalOutputTokens:     totalOutputTokens,
			InactiveTurns:         inactiveTurns,
			ConsecutiveErrorTurns: consecutiveErrorTurns,
			LastTextLength:        lastTextLength,
			LastToolSignature:     lastToolSignature,
			BudgetLimit:           limit,
			BudgetUsed:            used,
			Messages:              messages,
			PartialResult:         partialResult,
			Role:                  role,
			ProjectRoot:           projectRoot,
			Model:                 config.Model,
		}

		if err := writeCheckpoint(projectRoot, taskID, cp); err != nil {
			logMsg(fmt.Sprintf("⚠️  Failed to write checkpoint: %v", err))
			return tokenBudgetPaused, fmt.Errorf("token budget exceeded and checkpoint failed: %w", err)
		}

		return tokenBudgetPaused, ErrTaskPaused
	}

	// Warn at 80 % budget consumed.
	if used >= limit*8/10 {
		logMsg(fmt.Sprintf("⚠️  Token budget warning: %d/%d tokens used (%.0f%%)", used, limit, float64(used)/float64(limit)*100))
	}

	return tokenBudgetContinue, nil
}

// ─── processOneTurn ──────────────────────────────────────────────────────────

// processOneTurnResult carries everything the caller needs after executing the
// tool uses for one agentic turn.
type processOneTurnResult struct {
	// ToolResults is the list of results to feed back as the next "user" message.
	ToolResults []streaming.ToolResult

	// CompletionSummary is non-empty when the agent called TaskComplete.
	// When it is set the caller should exit the loop with success.
	CompletionSummary string
}

// processOneTurn handles the tool-execution phase of a single agentic turn.
//
// It:
//   - intercepts TaskComplete and captures its summary,
//   - dispatches every other tool to s.executeTool,
//   - caps individual tool results at maxToolResultChars,
//   - logs progress via logMsg.
//
// Returns a processOneTurnResult. If CompletionSummary != "" the loop must exit.
func (s *AgentServer) processOneTurn(
	ctx context.Context,
	toolUses []streaming.ToolUse,
	workingDir string,
	logMsg func(string),
) processOneTurnResult {
	const maxToolResultChars = 8000

	var toolResults []streaming.ToolResult
	var completionSummary string

	for _, toolUse := range toolUses {
		if strings.EqualFold(toolUse.Name, "TaskComplete") {
			// Capture summary; add synthetic result; do NOT dispatch to executeTool.
			summary, _ := toolUse.Input["summary"].(string)
			if summary == "" {
				summary = "(no summary provided)"
			}
			completionSummary = summary
			logMsg(fmt.Sprintf("      🏁 TaskComplete: %s", summary[:min(80, len(summary))]))
			toolResults = append(toolResults, streaming.ToolResult{
				ToolUseID: toolUse.ID,
				ToolName:  toolUse.Name,
				Content:   "Task marked complete.",
				IsError:   false,
			})
			continue
		}

		// Execute tool (native or MCP)
		result, err := s.executeTool(ctx, toolUse.Name, toolUse.Input, workingDir)
		isError := err != nil
		if err != nil {
			logMsg(fmt.Sprintf("         ❌ Tool execution failed: %v", err))
			result = fmt.Sprintf("Error: %v", err)
		} else {
			// Log a truncated preview of the tool result to keep execution logs manageable.
			preview := result
			const maxLogPreview = 500
			if len(preview) > maxLogPreview {
				preview = preview[:maxLogPreview] + fmt.Sprintf("… (%d chars total)", len(result))
			}
			logMsg(fmt.Sprintf("         ✓ %s", preview))
		}

		// Cap tool result size.
		if len(result) > maxToolResultChars {
			result = result[:maxToolResultChars] + fmt.Sprintf(
				"\n\n[Output truncated: %d chars total, showing first %d]",
				len(result), maxToolResultChars)
		}
		toolResults = append(toolResults, streaming.ToolResult{
			ToolUseID: toolUse.ID,
			ToolName:  toolUse.Name,
			Content:   result,
			IsError:   isError,
		})
	}

	return processOneTurnResult{
		ToolResults:       toolResults,
		CompletionSummary: completionSummary,
	}
}

// ─── checkStallConditions ────────────────────────────────────────────────────

// stallResult indicates whether the agentic loop should abort due to stalling.
type stallResult int

const (
	stallNone   stallResult = iota // progress detected or below threshold – continue
	stallAbort                     // stuck for too long – return error
)

// checkStallConditions examines the turn results to determine whether the agent
// is making progress and, if not, whether it has been stuck long enough to abort.
//
// Progress is measured by two independent signals:
//  1. textGrew      — the accumulated result text grew since the previous turn.
//  2. toolsDiffer   — the tool-name+input signature differs from the previous turn.
//
// If either signal fires, the inactive counter resets; otherwise it increments.
// A separate consecutiveErrorTurns counter tracks turns where every tool failed.
//
// Returns the updated counters and a stallResult indicating whether the loop
// should abort or continue normally.
func checkStallConditions(
	toolResults []streaming.ToolResult,
	toolUses []streaming.ToolUse,
	currentTextLength int,
	lastTextLength int,
	lastToolSignature string,
	inactiveTurns int,
	consecutiveErrorTurns int,
	maxInactiveTurns int,
	logMsg func(string),
) (newInactiveTurns int, newConsecutiveErrorTurns int, newLastTextLength int, newLastToolSignature string, sr stallResult) {
	// ── consecutive-error tracking ──────────────────────────────────────────
	allToolsErrored := len(toolResults) > 0
	for _, tr := range toolResults {
		if !tr.IsError {
			allToolsErrored = false
			break
		}
	}

	if allToolsErrored {
		consecutiveErrorTurns++
		logMsg(fmt.Sprintf("      ⚠️  All tools errored this turn (%d consecutive error turns)", consecutiveErrorTurns))
		if consecutiveErrorTurns >= maxInactiveTurns {
			logMsg(fmt.Sprintf("❌ Agent stuck: %d consecutive turns with all tools failing", maxInactiveTurns))
			return inactiveTurns, consecutiveErrorTurns, currentTextLength, lastToolSignature, stallAbort
		}
	} else {
		consecutiveErrorTurns = 0
	}

	// ── progress detection ──────────────────────────────────────────────────
	textGrew := currentTextLength > lastTextLength

	// Build tool pattern for logging (names only).
	var toolNames []string
	for _, toolUse := range toolUses {
		toolNames = append(toolNames, toolUse.Name)
	}
	currentToolPattern := strings.Join(toolNames, ",")

	// Build tool signature including input details for better progress detection.
	// Use the full input JSON so that commands sharing a long common prefix
	// (e.g. `go doc pkg.TypeA` vs `go doc pkg.TypeB`) are correctly treated as
	// distinct tool calls.
	var toolSignatures []string
	for _, toolUse := range toolUses {
		inputJSON, _ := json.Marshal(toolUse.Input)
		toolSignatures = append(toolSignatures, fmt.Sprintf("%s:%s", toolUse.Name, string(inputJSON)))
	}
	currentToolSignature := strings.Join(toolSignatures, "|")

	madeProgress := textGrew || (currentToolSignature != lastToolSignature)

	if madeProgress {
		if inactiveTurns > 0 {
			logMsg(fmt.Sprintf("      ✓ Progress detected - resetting inactive counter (was %d)", inactiveTurns))
		}
		return 0, consecutiveErrorTurns, currentTextLength, currentToolSignature, stallNone
	}

	// No progress.
	inactiveTurns++
	logMsg(fmt.Sprintf("      ⚠️  No progress (%d/%d inactive turns)", inactiveTurns, maxInactiveTurns))

	if inactiveTurns >= maxInactiveTurns {
		logMsg(fmt.Sprintf("❌ Agent stuck after %d turns without progress", maxInactiveTurns))
		logMsg(fmt.Sprintf("   Last tool pattern: %s", currentToolPattern))
		return inactiveTurns, consecutiveErrorTurns, currentTextLength, currentToolSignature, stallAbort
	}

	return inactiveTurns, consecutiveErrorTurns, currentTextLength, currentToolSignature, stallNone
}

// ─── flushCheckpoint ─────────────────────────────────────────────────────────

// flushCheckpoint serialises the current loop state into an AgentCheckpoint and
// writes it to disk via writeCheckpoint. It is a thin, independently-testable
// wrapper around writeCheckpoint that keeps checkpoint-creation logic out of
// executeAgenticLoop.
//
// Any write error is returned so the caller can decide how to handle it.
func flushCheckpoint(
	projectRoot string,
	taskID string,
	role string,
	model string,
	turn int,
	totalInputTokens int64,
	totalOutputTokens int64,
	inactiveTurns int,
	consecutiveErrorTurns int,
	lastTextLength int,
	lastToolSignature string,
	budgetLimit int64,
	budgetUsed int64,
	messages []streaming.Message,
	partialResult string,
	logMsg func(string),
) error {
	cp := &AgentCheckpoint{
		TaskID:                taskID,
		CreatedAt:             time.Now(),
		Turn:                  turn,
		TotalInputTokens:      totalInputTokens,
		TotalOutputTokens:     totalOutputTokens,
		InactiveTurns:         inactiveTurns,
		ConsecutiveErrorTurns: consecutiveErrorTurns,
		LastTextLength:        lastTextLength,
		LastToolSignature:     lastToolSignature,
		BudgetLimit:           budgetLimit,
		BudgetUsed:            budgetUsed,
		Messages:              messages,
		PartialResult:         partialResult,
		Role:                  role,
		ProjectRoot:           projectRoot,
		Model:                 model,
	}

	if err := writeCheckpoint(projectRoot, taskID, cp); err != nil {
		logMsg(fmt.Sprintf("⚠️  Failed to write checkpoint: %v", err))
		return err
	}
	return nil
}
