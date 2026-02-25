package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/constants"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/internal/protocol"
	"github.com/cortexa-llc/ai-pack/internal/streaming"
)

func (s *AgentServer) executeAgenticLoop(ctx context.Context, taskID string, role string, initialPrompt string, roleContext string, workingDir string, projectRoot string, config *AgentConfig, logMsg func(string)) (string, error) {
	// Define tools (native + MCP tools) in provider-agnostic format
	toolDefs := s.getAllTools()

	// roleContext already has the shared policy prepended by the caller
	systemPrompt := roleContext

	// Build messages starting with user prompt (task-specific)
	messages := []streaming.Message{
		{Role: "user", Content: initialPrompt},
	}

	logMsg(fmt.Sprintf("🔄 Starting agentic loop (max_inactive: %d, caching: enabled, extended_thinking: %v)",
		s.maxInactiveTurns, config.ExtendedThinking))

	var finalResult strings.Builder
	totalInputTokens := int64(0)
	totalOutputTokens := int64(0)

	// Progress tracking for turn counter reset
	turn := 1
	inactiveTurns := 0
	consecutiveErrorTurns := 0 // Turns where EVERY tool call returned an error (different-path looping)
	totalToolCalls := 0        // Cumulative tool calls across all turns; used to detect acknowledgement-only responses
	lastTextLength := 0
	lastToolSignature := "" // Tracks tool names + input hash for better progress detection

	for {
		logMsg(fmt.Sprintf("   Turn %d (inactive: %d)...", turn, inactiveTurns))

		// Simple message-count-based truncation to keep context manageable
		// Keep first message (initial prompt) + last 50 messages
		const maxHistoryMessages = 50
		truncatedMessages := messages
		if len(messages) > 1+maxHistoryMessages {
			firstMsg := messages[0] // Keep initial prompt with task context
			recentMsgs := messages[len(messages)-maxHistoryMessages:]
			truncatedMessages = append([]streaming.Message{firstMsg}, recentMsgs...)

			// Only log truncation occasionally to avoid spam
			if len(messages)%10 == 0 {
				logMsg(fmt.Sprintf("      📉 Truncated: %d → %d messages", len(messages), len(truncatedMessages)))
			}
		}

		// Truncate tool result content in older messages to prevent context blowup.
		// Full results are only needed for the most recent turns — older results
		// just need to convey what happened, not the full content.
		const maxOldResultChars = 500
		const recentTurnsToKeepFull = 3
		if len(truncatedMessages) > recentTurnsToKeepFull*2+1 {
			cutoff := len(truncatedMessages) - recentTurnsToKeepFull*2
			compacted := make([]streaming.Message, len(truncatedMessages))
			copy(compacted, truncatedMessages)
			for i := 1; i < cutoff; i++ { // skip index 0 (initial prompt)
				if len(compacted[i].ToolResults) > 0 {
					newResults := make([]streaming.ToolResult, len(compacted[i].ToolResults))
					for j, tr := range compacted[i].ToolResults {
						newResults[j] = tr
						if len(tr.Content) > maxOldResultChars {
							newResults[j].Content = tr.Content[:maxOldResultChars] + fmt.Sprintf(" …[%d chars]", len(tr.Content))
						}
					}
					compacted[i].ToolResults = newResults
				}
			}
			truncatedMessages = compacted
		}

		// Prepare streaming request — messages already in provider-agnostic format.
		// Pass the role-configured model as the explicit override only when it is set
		// in the role config file. When no model is pinned, pass "" so the
		// performance-grade model selector picks the most cost-effective option.
		requestModel := ""
		if config.Model != "" {
			requestModel = config.Model
		}
		streamReq := streaming.StreamRequest{
			Messages:         truncatedMessages,
			SystemPrompt:     systemPrompt,
			MaxTokens:        s.maxTokens,
			Model:            requestModel,
			Tools:            toolDefs,
			MinContextTokens: config.Delegation.MaxContext,
		}

		// Note: Extended thinking support requires newer SDK version
		if config.ExtendedThinking {
			logMsg("      ⚠️  Extended thinking requested but not yet supported")
		}

		// Make API call with streaming (uses performance-grade model selection)
		apiStart := time.Now()

		stream, err := s.streamingService.CreateStream(ctx, role, streamReq)
		if err != nil {
			return "", fmt.Errorf("failed to create stream on turn %d: %w", turn, err)
		}
		defer stream.Close()

		// Update task metadata with actual model used on first turn
		if turn == 1 {
			selectedModel := stream.GetModel()
			selectedProvider := stream.GetProvider()

			logMsg(fmt.Sprintf("   📊 Model selected: %s (%s)", selectedModel, selectedProvider))

			// Extract project root from working directory
			projectRoot := workingDir
			for projectRoot != "" && projectRoot != "/" {
				if _, err := os.Stat(filepath.Join(projectRoot, ".beads")); err == nil {
					break
				}
				projectRoot = filepath.Dir(projectRoot)
			}

			if projectRoot != "" && projectRoot != "/" {
				// Determine tier from model
				tier := 3 // Default to TierMedium (Sonnet)
				if strings.Contains(strings.ToLower(selectedModel), "haiku") || strings.Contains(strings.ToLower(selectedModel), "mini") {
					tier = 1 // TierMinimal
				} else if strings.Contains(strings.ToLower(selectedModel), "gpt-4o") || strings.Contains(strings.ToLower(selectedModel), "gpt-5") {
					tier = 2 // TierLow
				} else if strings.Contains(strings.ToLower(selectedModel), "opus") {
					tier = 4 // TierHigh
				}

				if err := s.updateTaskMetadata(projectRoot, taskID, selectedModel, selectedProvider, tier); err != nil {
					monitoring.Logger.Warn("failed_to_update_task_metadata",
						"task_id", taskID,
						"error", err.Error())
				}
				// Sync selected model into in-memory metadata so grade recording
				// in saveAndCompleteTask uses the actual model, not the server default.
				s.mu.Lock()
				if exec, ok := s.activeTasks[taskID]; ok {
					exec.metadata["model"] = selectedModel
					exec.metadata["provider"] = selectedProvider
				}
				s.mu.Unlock()
			}
		}

		// Accumulate response from streaming events
		var responseText strings.Builder
		var toolUses []streaming.ToolUse
		eventCount := 0

		for stream.Next() {
			event := stream.Current()
			eventCount++

			// Accumulate text
			if event.Delta != nil && event.Delta.Text != "" {
				responseText.WriteString(event.Delta.Text)
			}

			// Collect tool uses
			if event.ToolUse != nil {
				toolUses = append(toolUses, *event.ToolUse)
			}
		}

		// Get final accumulated message with stop reason and token usage
		finalMessage := stream.GetMessage()
		stopReason := ""
		inputTokens := int64(0)
		outputTokens := int64(0)
		if finalMessage != nil {
			stopReason = finalMessage.StopReason
			inputTokens = int64(finalMessage.InputTokens)
			outputTokens = int64(finalMessage.OutputTokens)
		}

		// Check for streaming errors
		if err := stream.Err(); err != nil {
			monitoring.GlobalMetrics.IncrementAPICallsFailed()
			errMsg := err.Error()

			// Check for token limit errors and provide actionable guidance
			if strings.Contains(errMsg, "prompt is too long") || strings.Contains(errMsg, "maximum") {
				logMsg(fmt.Sprintf("❌ Context size exceeded API limit on turn %d", turn))
				logMsg(fmt.Sprintf("   💡 Recommendation: Break this task into smaller subtasks"))

				return "", fmt.Errorf("API token limit exceeded on turn %d. "+
					"This task is too complex for a single agent execution. "+
					"Please break it into smaller subtasks: %w", turn, err)
			}

			return "", fmt.Errorf("API call failed on turn %d: %w", turn, err)
		}

		// Check for max_tokens limit
		if stopReason == "max_tokens" {
			monitoring.Logger.Warn("max_tokens_limit_reached",
				"task_id", taskID,
				"turn", turn,
				"output_tokens", outputTokens,
			)
			logMsg(fmt.Sprintf("      ⚠️  Max tokens reached (%d). Completing turn.", outputTokens))
		}

		apiDuration := time.Since(apiStart).Milliseconds()
		totalInputTokens += inputTokens
		totalOutputTokens += outputTokens

		// Log token usage (cumulative total is informational, not a limit)
		logMsg(fmt.Sprintf("      API: %dms | in:%d out:%d | cumulative:%d",
			apiDuration, inputTokens, outputTokens,
			totalInputTokens+totalOutputTokens))

		// Enforce token budget if set in role config
		if config.Delegation.MaxBudgetTokens > 0 {
			used := totalInputTokens + totalOutputTokens
			limit := int64(config.Delegation.MaxBudgetTokens)
			if used >= limit {
				logMsg(fmt.Sprintf("⏸️  Token budget exhausted: %d/%d tokens used — pausing", used, limit))
				cp := &AgentCheckpoint{
					TaskID: taskID, CreatedAt: time.Now(),
					Turn: turn, TotalInputTokens: totalInputTokens, TotalOutputTokens: totalOutputTokens,
					InactiveTurns: inactiveTurns, ConsecutiveErrorTurns: consecutiveErrorTurns,
					LastTextLength: lastTextLength, LastToolSignature: lastToolSignature,
					BudgetLimit: limit, BudgetUsed: used,
					Messages: messages, PartialResult: finalResult.String(),
					Role: role, ProjectRoot: projectRoot, Model: config.Model,
				}
				if err := writeCheckpoint(projectRoot, taskID, cp); err != nil {
					logMsg(fmt.Sprintf("⚠️  Failed to write checkpoint: %v", err))
					return finalResult.String(), fmt.Errorf("token budget exceeded and checkpoint failed: %w", err)
				}
				return finalResult.String(), ErrTaskPaused
			}
			// Warn at 80%
			if used >= limit*8/10 {
				logMsg(fmt.Sprintf("⚠️  Token budget warning: %d/%d tokens used (%.0f%%)", used, limit, float64(used)/float64(limit)*100))
			}
		}

		// Record per-turn token metrics and API call count
		monitoring.GlobalMetrics.RecordTurnTokens(taskID, turn, inputTokens, outputTokens, apiDuration)
		monitoring.GlobalMetrics.IncrementAPICallsSuccess()

		// Get model and provider from stream
		selectedModel := stream.GetModel()
		selectedProvider := stream.GetProvider()

		// Record provider-specific usage
		monitoring.GlobalMetrics.RecordProviderUsage(selectedProvider, selectedModel, inputTokens, outputTokens)

		// Record persistent daily usage
		if pm, err := s.getOrCreateProjectMetrics(s.rootDir); err == nil {
			if err := pm.RecordUsage(selectedProvider, selectedModel, inputTokens, outputTokens); err != nil {
				monitoring.Logger.Warn("failed_to_record_persistent_metrics", "error", err.Error())
			}
		}

		// Process response content
		responseTextStr := responseText.String()
		hasText := len(responseTextStr) > 0

		if hasText {
			finalResult.WriteString(responseTextStr)
			finalResult.WriteString("\n")
			logMsg(fmt.Sprintf("      💬 Text: %d chars", len(responseTextStr)))
		}

		// Log tool uses with key parameter
		for _, toolUse := range toolUses {
			param := toolParamPreview(toolUse.Name, toolUse.Input)
			if param != "" {
				logMsg(fmt.Sprintf("      🔧 Tool: %s(%s)", toolUse.Name, param))
			} else {
				logMsg(fmt.Sprintf("      🔧 Tool: %s", toolUse.Name))
			}
		}

		// If no tool uses and we have text, the agent is signalling completion.
		// Require at least one prior tool call so a model that simply acknowledges
		// the task on turn 1 (without doing any work) is not accepted as done.
		if len(toolUses) == 0 {
			if hasText {
				if turn == 1 && totalToolCalls == 0 {
					// Model acknowledged but did nothing — treat as a missed start,
					// return an error so the task can be retried or escalated.
					return "", fmt.Errorf("agent produced no tool calls on turn 1 (acknowledgement without work)")
				}
				logMsg(fmt.Sprintf("✅ Agent completed in %d turns", turn))
				logMsg(fmt.Sprintf("   Total tokens: %d (in:%d out:%d)", totalInputTokens+totalOutputTokens, totalInputTokens, totalOutputTokens))
				break
			}
			return "", fmt.Errorf("no output from agent on turn %d", turn)
		}

		// Execute tools and accumulate results
		totalToolCalls += len(toolUses)
		var toolResults []streaming.ToolResult
		for _, toolUse := range toolUses {
			// Execute tool (native or MCP)
			result, err := s.executeTool(ctx, toolUse.Name, toolUse.Input, workingDir)
			isError := err != nil
			if err != nil {
				logMsg(fmt.Sprintf("         ❌ Tool execution failed: %v", err))
				result = fmt.Sprintf("Error: %v", err)
			} else {
				// Log a truncated preview of the tool result to keep execution logs manageable
				preview := result
				const maxLogPreview = 500
				if len(preview) > maxLogPreview {
					preview = preview[:maxLogPreview] + fmt.Sprintf("… (%d chars total)", len(result))
				}
				logMsg(fmt.Sprintf("         ✓ %s", preview))
			}

			// Cap tool result size. Large results are compacted in history on subsequent turns,
			// but we still cap at initial storage to avoid a single result eating the context.
			const maxToolResultChars = 8000
			if len(result) > maxToolResultChars {
				result = result[:maxToolResultChars] + fmt.Sprintf("\n\n[Output truncated: %d chars total, showing first %d]", len(result), maxToolResultChars)
			}
			toolResults = append(toolResults, streaming.ToolResult{
				ToolUseID: toolUse.ID,
				Content:   result,
				IsError:   isError,
			})
		}

		// Track whether every tool call this turn returned an error
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
			maxConsecutiveErrors := s.maxInactiveTurns
			if consecutiveErrorTurns >= maxConsecutiveErrors {
				logMsg(fmt.Sprintf("❌ Agent stuck: %d consecutive turns with all tools failing", maxConsecutiveErrors))
				return "", fmt.Errorf("agent stuck: %d consecutive turns where every tool call returned an error", maxConsecutiveErrors)
			}
		} else {
			consecutiveErrorTurns = 0
		}

		// Progress detection: check if agent is making progress
		currentTextLength := finalResult.Len()
		textGrew := currentTextLength > lastTextLength

		// Build tool pattern for this turn (just names for logging)
		var toolNames []string
		for _, toolUse := range toolUses {
			toolNames = append(toolNames, toolUse.Name)
		}
		currentToolPattern := strings.Join(toolNames, ",")

		// Build tool signature including input details for better progress detection.
		// Use the full input JSON (no truncation) so that commands sharing a long common
		// prefix (e.g. `go doc pkg.TypeA` vs `go doc pkg.TypeB`) are correctly treated
		// as distinct tool calls.
		var toolSignatures []string
		for _, toolUse := range toolUses {
			inputJSON, _ := json.Marshal(toolUse.Input)
			toolSignatures = append(toolSignatures, fmt.Sprintf("%s:%s", toolUse.Name, string(inputJSON)))
		}
		currentToolSignature := strings.Join(toolSignatures, "|")

		// Check if agent is making progress
		// Progress = text grew OR different tools OR same tools with different inputs
		madeProgress := textGrew || (currentToolSignature != lastToolSignature)

		if madeProgress {
			// Agent is making progress - reset inactive counter
			if inactiveTurns > 0 {
				logMsg(fmt.Sprintf("      ✓ Progress detected - resetting inactive counter (was %d)", inactiveTurns))
			}
			inactiveTurns = 0
			lastTextLength = currentTextLength
			lastToolSignature = currentToolSignature
		} else {
			// No progress - increment inactive counter
			inactiveTurns++
			logMsg(fmt.Sprintf("      ⚠️  No progress (%d/%d inactive turns)", inactiveTurns, s.maxInactiveTurns))

			if inactiveTurns >= s.maxInactiveTurns {
				logMsg(fmt.Sprintf("❌ Agent stuck after %d turns without progress", s.maxInactiveTurns))
				logMsg(fmt.Sprintf("   Last tool pattern: %s", currentToolPattern))
				return "", fmt.Errorf("agent stuck after %d turns without progress - repeating: %s", s.maxInactiveTurns, currentToolPattern)
			}
		}

		// Add assistant message with text and tool uses
		messages = append(messages, streaming.Message{
			Role:     "assistant",
			Content:  responseTextStr,
			ToolUses: toolUses,
		})

		// Add tool results as user message
		messages = append(messages, streaming.Message{
			Role:        "user",
			ToolResults: toolResults,
		})

		// Increment turn counter
		turn++

		// Enforce turn budget
		if config.Delegation.MaxTurns > 0 && turn > config.Delegation.MaxTurns {
			logMsg(fmt.Sprintf("❌ Turn budget exhausted: %d turns reached (limit: %d)", turn-1, config.Delegation.MaxTurns))
			return finalResult.String(), fmt.Errorf("turn budget exceeded: %d turns (limit: %d)", turn-1, config.Delegation.MaxTurns)
		}
	}

	monitoring.LogAPICall(ctx, taskID, s.model, int(totalInputTokens+totalOutputTokens))

	// Record token usage for this session
	monitoring.GlobalMetrics.RecordTokenUsage(taskID, totalInputTokens, totalOutputTokens, int64(turn-1))

	return finalResult.String(), nil
}

func (s *AgentServer) executeAgentTask(execution *TaskExecution) {
	// Create context with deadline derived from the role's Timeout config field.
	// parseRoleTimeout converts human-friendly values like "10min" or "1h" to a
	// time.Duration; it falls back to defaultRoleTimeout when the value is missing
	// or unparseable.
	roleTimeout := parseRoleTimeout(execution.Config.Delegation.Timeout)
	ctx, cancel := context.WithTimeout(context.Background(), roleTimeout)

	// Store cancel function so task can be cancelled later
	s.mu.Lock()
	execution.cancel = cancel
	s.mu.Unlock()

	// Ensure cancel is called on exit
	defer cancel()

	taskID := execution.TaskID
	startTime := time.Now()

	// Create execution log and logger
	logMsg := s.setupExecutionLogger(execution)

	logMsg("🚀 Agent execution started")
	logMsg(fmt.Sprintf("   Role: %s", execution.Role))
	logMsg(fmt.Sprintf("   Task: %s", execution.Task))

	monitoring.LogTaskStarted(ctx, taskID, execution.Role)
	monitoring.GlobalMetrics.IncrementTasksSpawned()

	// Initialize task execution (sets status to in_progress)
	if err := s.initializeTaskExecution(execution, logMsg); err != nil {
		s.failTask(execution, fmt.Sprintf("Failed to initialize: %v", err))
		return
	}

	// Load role context
	roleContext, err := s.loadAndLogRoleContext(execution, logMsg)
	if err != nil {
		return
	}

	// Get task metadata (paths, working directory)
	taskPacketPath, workingDir := s.extractTaskMetadata(execution, logMsg)

	// Build and save prompt
	prompt := s.buildAndSavePrompt(execution, roleContext, taskPacketPath, workingDir, logMsg)

	// Execute the agent's work
	result, err := s.executeAgentWorkflow(ctx, execution, prompt, roleContext, workingDir, logMsg)
	if err != nil {
		return
	}

	// Save and complete task
	s.saveAndCompleteTask(ctx, execution, result, startTime, logMsg)
}

func (s *AgentServer) executeAgentWorkflow(ctx context.Context, execution *TaskExecution, prompt, roleContext, workingDir string, logMsg func(string)) (string, error) {
	s.sendStreamEvent(execution, "api_call_start", map[string]interface{}{})

	result, err := s.executeAgenticLoop(ctx, execution.TaskID, execution.Role, prompt, s.buildSystemPromptForProject(roleContext, execution.ProjectRoot), workingDir, execution.ProjectRoot, execution.Config, logMsg)
	if err != nil {
		// Check if error is due to cancellation or timeout.
		// context.Canceled: User-initiated cancellation via CancelTask()
		// context.DeadlineExceeded: Task exceeded configured timeout
		if errors.Is(err, ErrTaskPaused) {
			logMsg(fmt.Sprintf("⏸️  Task %s paused at checkpoint", execution.TaskID))
			s.sendStreamEvent(execution, "budget_paused", map[string]interface{}{
				"used":            0,
				"turn":            0,
				"checkpoint_path": checkpointPath(execution.ProjectRoot, execution.TaskID),
			})
			if err2 := s.updateTaskStatus(execution.TaskID, execution.ProjectRoot, constants.StatusPaused, ""); err2 != nil {
				monitoring.Logger.Error("failed_to_update_paused_status", "task_id", execution.TaskID, "error", err2)
			}
			s.closeStream(execution)
			return result, ErrTaskPaused
		}
		if ctx.Err() == context.Canceled {
			logMsg("🛑 Task cancelled")
			s.cancelTaskExecution(execution, "Task cancelled by user")
			return "", fmt.Errorf("task cancelled")
		}
		if ctx.Err() == context.DeadlineExceeded {
			logMsg("⏱️ Task timeout exceeded")
			s.cancelTaskExecution(execution, "Task exceeded deadline")
			return "", fmt.Errorf("task timeout")
		}
		logMsg(fmt.Sprintf("❌ Agentic loop failed: %v", err))
		s.failTask(execution, fmt.Sprintf("Execution error: %v", err))
		return "", err
	}

	s.sendStreamEvent(execution, "api_call_complete", map[string]interface{}{})

	return result, nil
}

func (s *AgentServer) resumeFromCheckpoint(taskID, projectRoot string, cp *AgentCheckpoint, newBudget int64) {
	// Load config for the role (picks up any role-file changes since the pause)
	config, err := s.loadAgentConfig(cp.Role, projectRoot)
	if err != nil {
		monitoring.Logger.Error("resume_load_config_failed", "task_id", taskID, "error", err)
		return
	}

	// Override budget with the new allocation (0 = unlimited)
	config.Delegation.MaxBudgetTokens = int(newBudget)

	// Rebuild a TaskExecution that mirrors how executeAgentTask works
	execution := &TaskExecution{
		TaskID:      taskID,
		Role:        cp.Role,
		Task:        cp.PartialResult, // used as context; real prompt built below
		Config:      config,
		StartTime:   time.Now(),
		Status:      constants.StatusInProgress,
		ProjectRoot: projectRoot,
		metadata:    map[string]string{"project_root": projectRoot},
		streamChan:  make(chan *protocol.StreamEvent, 100),
		streamOpen:  true,
	}

	roleTimeout := parseRoleTimeout(config.Delegation.Timeout)
	ctx, cancel := context.WithTimeout(context.Background(), roleTimeout)
	execution.cancel = cancel
	defer cancel()

	// Register as active
	s.mu.Lock()
	s.activeTasks[taskID] = execution
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.activeTasks, taskID)
		s.mu.Unlock()
	}()

	logMsg := s.setupExecutionLogger(execution)
	logMsg(fmt.Sprintf("▶️  Resuming task %s from checkpoint (turn %d, budget %d→%d)",
		taskID, cp.Turn, cp.BudgetUsed, newBudget))

	// Update disk status to in_progress
	if err := s.updateTaskStatus(taskID, projectRoot, constants.StatusInProgress, ""); err != nil {
		logMsg(fmt.Sprintf("⚠️  Failed to update status: %v", err))
	}

	// Build a resume prompt that injects prior partial result as context
	var resumePrompt string
	if cp.PartialResult != "" {
		resumePrompt = fmt.Sprintf(
			"You are resuming work on task %s after a token-budget pause.\n\n"+
				"== PRIOR PARTIAL RESULT (from %d completed turns) ==\n%s\n\n"+
				"== INSTRUCTIONS ==\n"+
				"Continue from where you left off. Do not repeat work already done.\n"+
				"Pick up where you left off and complete the remaining work.",
			taskID, cp.Turn, cp.PartialResult)
	} else {
		resumePrompt = fmt.Sprintf(
			"You are resuming work on task %s after a token-budget pause at turn %d. "+
				"No partial result was saved. Start fresh and complete the task.",
			taskID, cp.Turn)
	}

	roleContext, err := s.loadAndLogRoleContext(execution, logMsg)
	if err != nil {
		logMsg(fmt.Sprintf("⚠️  Could not load role context: %v", err))
		roleContext = ""
	}

	_, workingDir := s.extractTaskMetadata(execution, logMsg)

	// Run the agentic loop
	result, loopErr := s.executeAgentWorkflow(ctx, execution, resumePrompt, roleContext, workingDir, logMsg)

	if loopErr != nil {
		if errors.Is(loopErr, ErrTaskPaused) {
			// Already handled inside executeAgentWorkflow (checkpoint written, status set)
			logMsg("⏸️  Task paused again at new budget limit")
			return
		}
		logMsg(fmt.Sprintf("❌ Resume failed: %v", loopErr))
		s.failTask(execution, loopErr.Error())
		return
	}

	s.saveAndCompleteTask(ctx, execution, result, execution.StartTime, logMsg)
}
