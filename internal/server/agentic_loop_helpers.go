package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/constants"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
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
//   - caps individual tool results at constants.MaxToolResultChars,
//   - logs progress via logMsg.
//
// Returns a processOneTurnResult. If CompletionSummary != "" the loop must exit.
func (s *AgentServer) processOneTurn(
	ctx context.Context,
	toolUses []streaming.ToolUse,
	workingDir string,
	projectRoot string,
	logMsg func(string),
) processOneTurnResult {
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
		result, err := s.executeTool(ctx, toolUse.Name, toolUse.Input, workingDir, projectRoot)
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
		if len(result) > constants.MaxToolResultChars {
			result = result[:constants.MaxToolResultChars] + fmt.Sprintf(
				"\n\n[Output truncated: %d chars total, showing first %d]",
				len(result), constants.MaxToolResultChars)
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
	maxConsecutiveErrorTurns int,
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
		if consecutiveErrorTurns >= maxConsecutiveErrorTurns {
			logMsg(fmt.Sprintf("❌ Agent stuck: %d consecutive turns with all tools failing", maxConsecutiveErrorTurns))
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

// ─── hasKGProgress ───────────────────────────────────────────────────────────

// hasKGProgress returns true if any tool call this turn wrote to the knowledge
// graph. KG writes are the canonical signal for "forward progress": the agent
// has crystallised a finding into a durable checkpoint, not merely searched or
// read files in a loop.
//
// These tool names match the kg__ prefix registered by buildMCPStreamTool and
// represent the write-side of the KG API. Read-only calls (search_knowledge,
// get_file_context, get_preflight_context) are intentionally excluded — they
// may be repeated without indicating convergence.
func hasKGProgress(toolUses []streaming.ToolUse) bool {
	for _, t := range toolUses {
		switch t.Name {
		case "kg__add_entity", "kg__add_observation", "kg__link_entities", "kg__index_project":
			return true
		}
	}
	return false
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

// ─── Token-budget message truncation ─────────────────────────────────────────

// charsPerToken is a conservative character-to-token ratio used to estimate
// message sizes without a real tokeniser. English prose averages ~4 chars/token;
// code and JSON average ~3–3.5 chars/token. Using 4 intentionally under-estimates
// token counts by 10–30% for code-heavy content, leaving extra headroom in the
// truncation budget and avoiding accidental context-window overrun.
//
// If this heuristic becomes a bottleneck (e.g. excessive truncation for prose),
// replace with a real tokeniser such as tiktoken.
const charsPerToken = 4

// estimateMessageTokens returns a conservative (under-)estimate of the token count
// for a single message by summing the character lengths of every text field and
// dividing by charsPerToken. The estimate may be 10–30% below the true count for
// code-heavy content; this bias is intentional — see charsPerToken.
func estimateMessageTokens(m streaming.Message) int {
	chars := len(m.Content)
	for _, tu := range m.ToolUses {
		chars += len(tu.Name)
		if raw, err := json.Marshal(tu.Input); err == nil {
			chars += len(raw)
		}
	}
	for _, tr := range m.ToolResults {
		chars += len(tr.Content)
	}
	tokens := chars / charsPerToken
	if tokens == 0 {
		tokens = 1 // every message costs at least 1 token
	}
	return tokens
}

// truncateMessagesByTokenBudget keeps messages[0] (the initial user prompt)
// unconditionally and then fills in as many of the most-recent messages as fit
// within the given token budget (expressed in tokens, not characters).
//
// If messages[0] alone already exceeds the budget the function still returns it
// alone so the caller always has at least one message to send.
//
// Truncation is logged via logMsg when at least one message is dropped.
func truncateMessagesByTokenBudget(
	messages []streaming.Message,
	budgetTokens int,
	logMsg func(string),
) []streaming.Message {
	if len(messages) == 0 {
		return messages
	}

	// Always preserve the first message (initial user prompt / task context).
	firstMsg := messages[0]
	firstTokens := estimateMessageTokens(firstMsg)
	remaining := messages[1:]

	if len(remaining) == 0 {
		return messages
	}

	// Available budget for the history after reserving space for the first message.
	available := budgetTokens - firstTokens
	if available <= 0 {
		// First message alone exceeds budget – return it anyway.
		return []streaming.Message{firstMsg}
	}

	// Walk backwards through the remaining messages, accumulating until we run
	// out of budget.
	kept := 0
	tokensUsed := 0
	for i := len(remaining) - 1; i >= 0; i-- {
		t := estimateMessageTokens(remaining[i])
		if tokensUsed+t > available {
			break
		}
		tokensUsed += t
		kept++
	}

	dropped := len(remaining) - kept
	if dropped == 0 {
		return messages
	}

	logMsg(fmt.Sprintf(
		"      📉 Context truncated: dropped %d older message(s) to stay within ~%d-token budget (kept first + %d recent, ~%d tokens used)",
		dropped, budgetTokens, kept, firstTokens+tokensUsed,
	))

	result := make([]streaming.Message, 0, 1+kept)
	result = append(result, firstMsg)
	result = append(result, remaining[len(remaining)-kept:]...)
	return result
}

// contextWindowTokens returns the effective token budget for message history.
// It reads the model's known context window from ModelInfo when available, then
// reserves a response buffer (constants.DefaultMaxTokens) to ensure the model
// always has room to generate its reply.  When the model is unknown it falls
// back to constants.MaxContextTokens.
func contextWindowTokens(modelID string) int {
	budget := constants.MaxContextTokens
	if info, ok := monitoring.GetModelInfo(modelID); ok && info.ContextWindow > 0 {
		budget = info.ContextWindow
	}
	// Reserve space for the model's own output so we don't leave it zero headroom.
	budget -= constants.DefaultMaxTokens
	if budget <= 0 {
		budget = constants.MaxContextTokens
	}
	return budget
}
