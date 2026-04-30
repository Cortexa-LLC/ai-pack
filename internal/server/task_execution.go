package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/constants"
	"github.com/cortexa-llc/ai-pack/internal/kgclient"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/internal/protocol"
	"github.com/cortexa-llc/ai-pack/internal/streaming"
)

func (s *AgentServer) executeAgenticLoop(ctx context.Context, roleTimeout time.Duration, taskID string, role string, initialPrompt string, roleContext string, workingDir string, projectRoot string, config *AgentConfig, logMsg func(string)) (string, error) {
	// Define tools (native + MCP tools) in provider-agnostic format
	toolDefs := s.getAllTools(projectRoot)

	// roleContext already has the shared policy prepended by the caller
	systemPrompt := roleContext

	// Build messages starting with user prompt (task-specific)
	messages := []streaming.Message{
		{Role: "user", Content: initialPrompt},
	}

	logMsg(fmt.Sprintf("🔄 Starting agentic loop (max_inactive: %d, max_consec_errors: %d, caching: enabled, extended_thinking: %v)",
		s.maxInactiveTurns, s.maxConsecutiveErrorTurns, config.ExtendedThinking))

	var finalResult strings.Builder
	totalInputTokens := int64(0)
	totalOutputTokens := int64(0)

	// Progress tracking for turn counter reset
	turn := 1
	inactiveTurns := 0
	consecutiveErrorTurns := 0    // Turns where EVERY tool call returned an error (different-path looping)
	consecutiveTextOnlyTurns := 0 // Consecutive turns with text and no tool calls; triggers failure after limit
	totalToolCalls := 0           // Cumulative tool calls across all turns
	lastTextLength := 0
	lastToolSignature := "" // Tracks tool names + input hash for better progress detection

	for {
		logMsg(fmt.Sprintf("   Turn %d (inactive: %d)...", turn, inactiveTurns))

		// Track retries: each turn after the first is a retry attempt.
		if turn > 1 {
			s.mu.Lock()
			if exec, ok := s.activeTasks[taskID]; ok {
				exec.RetryCount++
			}
			s.mu.Unlock()
		}

		// Resolve the model override once per turn so the truncation budget can
		// use the same value that will be passed to the stream request.
		// requestModel == "" signals the provider layer to use the grade-selected
		// model (selectedModel above). A non-empty override bypasses grade-based
		// selection entirely and pins the request to the specified model.
		requestModel := ""
		if config.Model != "" {
			requestModel = config.Model
		}

		// Token-budget-based truncation: keep messages[0] (initial prompt) and
		// fill backwards from the most-recent turn until the estimated token budget
		// is exhausted.  This replaces the old hard-coded 50-message limit.
		// This truncation runs on every turn, including the first turn after a
		// checkpoint restore — the checkpointed history may be arbitrarily long,
		// so we must re-validate it against the current context window before
		// sending it to the model.
		budget := contextWindowTokens(requestModel)
		truncatedMessages := truncateMessagesByTokenBudget(messages, budget, logMsg)

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
		// (requestModel was already set above for the token-budget computation.)
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

			// Determine tier from model using the canonical registry.
			// We need this for both metadata storage and escalation/downgrade detection.
			selectedTier, found := monitoring.GetModelTier(selectedModel)
			if !found {
				selectedTier = monitoring.TierMedium // default when model not in registry
			}

			// Extract project root from working directory
			projectRoot := workingDir
			for projectRoot != "" && projectRoot != "/" {
				if _, err := os.Stat(filepath.Join(projectRoot, ".beads")); err == nil {
					break
				}
				projectRoot = filepath.Dir(projectRoot)
			}

			if projectRoot != "" && projectRoot != "/" {
				tier := int(selectedTier)

				if err := s.updateTaskMetadata(projectRoot, taskID, selectedModel, selectedProvider, tier); err != nil {
					monitoring.Logger.Warn("failed_to_update_task_metadata",
						"task_id", taskID,
						"error", err.Error())
				}
			}

			// Sync selected model into in-memory metadata and detect escalation/downgrade.
			s.mu.Lock()
			if exec, ok := s.activeTasks[taskID]; ok {
				exec.metadata["model"] = selectedModel
				exec.metadata["provider"] = selectedProvider

				// Detect escalation / downgrade relative to the role's default tier.
				defaultTier := monitoring.GlobalModelSelector.GetRoleDefaultTier(role)
				exec.WasEscalated = selectedTier > defaultTier
				exec.WasDowngraded = selectedTier < defaultTier
			}
			s.mu.Unlock()

			// Extend context deadline for free/local providers — they are slower than
			// cloud APIs so the default timeout is insufficient.
			isFreeProvider := false
			for _, fp := range s.freeProviders {
				if fp == selectedProvider {
					isFreeProvider = true
					break
				}
			}
			if isFreeProvider {
				extendedTimeout := roleTimeout * 2
				newCtx, newCancel := context.WithTimeout(s.ctx, extendedTimeout)
				ctx = newCtx
				s.mu.Lock()
				if exec, ok := s.activeTasks[taskID]; ok {
					exec.cancel = newCancel
				}
				s.mu.Unlock()
				logMsg(fmt.Sprintf("   ⏱️ Local provider %q — timeout extended to %v (2× multiplier)", selectedProvider, extendedTimeout))
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

			// LM Studio KV-cache diff failure: the server's cached prompt state has
			// stale tool_call counts (e.g. from a previous run that used OpenAI format
			// while the current run uses XML format). Retry the turn once — LM Studio
			// will process the full prompt fresh since the cache is now invalidated.
			if strings.Contains(errMsg, "Invalid diff") || strings.Contains(errMsg, "now finding less tool calls") {
				logMsg(fmt.Sprintf("⚠️  LM Studio KV-cache mismatch on turn %d — retrying turn once", turn))
				stream.Close()
				retryStream, retryErr := s.streamingService.CreateStream(ctx, role, streamReq)
				if retryErr != nil {
					// Check if it's a timeout error
					if strings.Contains(retryErr.Error(), "context deadline exceeded") ||
					   strings.Contains(retryErr.Error(), "deadline") ||
					   strings.Contains(retryErr.Error(), "timeout") {
						return "", fmt.Errorf("API request timeout on turn %d (retry): %w", turn, retryErr)
					}
					return "", fmt.Errorf("API call failed on turn %d (retry): %w", turn, retryErr)
				}
				responseText.Reset()
				toolUses = nil
				for retryStream.Next() {
					ev := retryStream.Current()
					if ev.Delta != nil && ev.Delta.Text != "" {
						responseText.WriteString(ev.Delta.Text)
					}
					if ev.ToolUse != nil {
						toolUses = append(toolUses, *ev.ToolUse)
					}
				}
				if retryErr2 := retryStream.Err(); retryErr2 != nil {
					retryStream.Close()
					// Check if it's a timeout error
					if strings.Contains(retryErr2.Error(), "context deadline exceeded") ||
					   strings.Contains(retryErr2.Error(), "deadline") ||
					   strings.Contains(retryErr2.Error(), "timeout") {
						return "", fmt.Errorf("API request timeout on turn %d (retry): %w", turn, retryErr2)
					}
					return "", fmt.Errorf("API call failed on turn %d (retry): %w", turn, retryErr2)
				}
				if retryFinal := retryStream.GetMessage(); retryFinal != nil {
					finalMessage = retryFinal
					stopReason = retryFinal.StopReason
					inputTokens = int64(retryFinal.InputTokens)
					outputTokens = int64(retryFinal.OutputTokens)
				}
				retryStream.Close()
				goto afterStreamErr
			}

			// Check for token limit errors and provide actionable guidance
			if strings.Contains(errMsg, "prompt is too long") || strings.Contains(errMsg, "maximum") {
				logMsg(fmt.Sprintf("❌ Context size exceeded API limit on turn %d", turn))
				logMsg(fmt.Sprintf("   💡 Recommendation: Break this task into smaller subtasks"))

				return "", fmt.Errorf("API token limit exceeded on turn %d. "+
					"This task is too complex for a single agent execution. "+
					"Please break it into smaller subtasks: %w", turn, err)
			}

			// Check if it's a timeout error and provide clearer message
			if strings.Contains(errMsg, "context deadline exceeded") ||
			   strings.Contains(errMsg, "deadline") ||
			   strings.Contains(errMsg, "timeout") {
				return "", fmt.Errorf("API request timeout on turn %d: %w", turn, err)
			}

			return "", fmt.Errorf("API call failed on turn %d: %w", turn, err)
		}
	afterStreamErr:

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
		if budgetResult, err := s.applyTokenBudgetCheck(
			taskID, projectRoot, role, config,
			turn, totalInputTokens, totalOutputTokens,
			inactiveTurns, consecutiveErrorTurns,
			lastTextLength, lastToolSignature,
			messages, finalResult.String(), logMsg,
		); budgetResult == tokenBudgetPaused {
			return finalResult.String(), err
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

		// If no tool uses: text-only response. Before giving up, attempt a fallback
		// XML parse for Qwen — thinking tokens can confuse the streaming-layer parser,
		// causing valid <tool_call> blocks to be missed there. We re-parse here from
		// the clean responseText (which never contains thinking tokens from the loop).
		if len(toolUses) == 0 && hasText {
			if xmlUses, cleanText := streaming.ParseQwenXMLToolCalls(responseTextStr); len(xmlUses) > 0 {
				logMsg(fmt.Sprintf("   🔄 Fallback XML parse recovered %d tool call(s) from text", len(xmlUses)))
				toolUses = xmlUses
				responseTextStr = cleanText
			}
		}

		// If no tool uses: text-only response. Completion is ONLY signalled by calling
		// TaskComplete — never by text alone. Nudge, but fail after too many in a row.
		if len(toolUses) == 0 {
			if hasText {
				consecutiveTextOnlyTurns++
				logMsg(fmt.Sprintf("⚠️  Text-only response (turn %d, %d consecutive) — nudging to use tools:", turn, consecutiveTextOnlyTurns))
				logMsg(responseTextStr)
				// Capture the reasoning to the KG so retries inherit the agent's findings.
				kgclient.WriteAgentReasoning(ctx, s.mcpManager, projectRoot, shortTaskIDFromFullID(taskID), turn, responseTextStr)
				const maxTextOnlyTurns = 5
				if consecutiveTextOnlyTurns >= maxTextOnlyTurns {
					return "", fmt.Errorf("agent produced %d consecutive text-only responses without making a tool call", maxTextOnlyTurns)
				}
				messages = append(messages, streaming.Message{
					Role:    "assistant",
					Content: responseTextStr,
				})
				messages = append(messages, streaming.Message{
					Role:    "user",
					Content: "Start with a tool call. If your work is fully complete, call TaskComplete with a summary. If not, use Read/Grep/Glob/Bash/Write/Edit to continue working.",
				})
				turn++
				continue
			}
			return "", fmt.Errorf("no output from agent on turn %d", turn)
		}
		consecutiveTextOnlyTurns = 0 // reset on any turn with tool calls

		// Execute tools and accumulate results.
		// TaskComplete is intercepted here — other tools run normally first,
		// then if TaskComplete was called the loop exits with the summary.
		totalToolCalls += len(toolUses)
		turnResult := s.processOneTurn(ctx, toolUses, workingDir, projectRoot, logMsg)
		toolResults := turnResult.ToolResults
		completionSummary := turnResult.CompletionSummary

		// If TaskComplete was called this turn, exit cleanly — but reject premature
		// completions (no real work done yet) by bouncing the agent back.
		if completionSummary != "" {
			// Count non-TaskComplete tool calls made so far (excluding this turn's TaskComplete).
			realToolCalls := totalToolCalls - len(toolUses)
			if realToolCalls == 0 {
				logMsg("⚠️  TaskComplete called on turn 1 with no prior tool calls — rejecting premature completion")
				messages = append(messages, streaming.Message{
					Role:    "assistant",
					Content: completionSummary,
				})
				messages = append(messages, streaming.Message{
					Role:    "user",
					Content: "You have not done any work yet. Do NOT call TaskComplete until you have actually investigated and made changes. Start by using Read, Grep, Glob, or Bash to investigate the task.",
				})
				turn++
				continue
			}
			logMsg(fmt.Sprintf("✅ Agent called TaskComplete in %d turns", turn))
			logMsg(fmt.Sprintf("   Total tokens: %d (in:%d out:%d)", totalInputTokens+totalOutputTokens, totalInputTokens, totalOutputTokens))
			monitoring.LogAPICall(ctx, taskID, s.model, int(totalInputTokens+totalOutputTokens))
			monitoring.GlobalMetrics.RecordTokenUsage(taskID, totalInputTokens, totalOutputTokens, int64(turn-1))
			if hasText && responseTextStr != "" {
				return responseTextStr + "\n\n" + completionSummary, nil
			}
			return completionSummary, nil
		}

		// Progress detection: check if agent is making progress
		currentTextLength := finalResult.Len()
		{
			newInactive, newConsecErr, newTextLen, newSig, sr := checkStallConditions(
				toolResults, toolUses,
				currentTextLength, lastTextLength, lastToolSignature,
				inactiveTurns, consecutiveErrorTurns, s.maxInactiveTurns,
				s.maxConsecutiveErrorTurns,
				logMsg,
			)
			inactiveTurns = newInactive
			consecutiveErrorTurns = newConsecErr
			lastTextLength = newTextLen
			lastToolSignature = newSig
			if sr == stallAbort {
				return "", fmt.Errorf("agent stuck after %d turns without progress", s.maxInactiveTurns)
			}

			// Popcorn bidding: KG checkpoint written this turn — reset the deadline to
			// a fresh roleTimeout window. KG writes are the canonical signal that the
			// agent has crystallised a finding, not just spun through reads/greps.
			// An agent that keeps checkpointing progress will never be interrupted
			// mid-investigation; only agents that stop producing KG findings hit the wall.
			if hasKGProgress(toolUses) {
				newCtx, newCancel := context.WithTimeout(s.ctx, roleTimeout)
				ctx = newCtx
				s.mu.Lock()
				if exec, ok := s.activeTasks[taskID]; ok {
					if exec.cancel != nil {
						exec.cancel() // free the old timer goroutine
					}
					exec.cancel = newCancel
				}
				s.mu.Unlock()
				logMsg(fmt.Sprintf("   📚 KG checkpoint written — deadline reset (%v window)", roleTimeout))
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
}

func (s *AgentServer) executeAgentTask(execution *TaskExecution) {
	// Create context with deadline derived from the role's Timeout config field.
	// parseRoleTimeout converts human-friendly values like "10min" or "1h" to a
	// time.Duration; it falls back to defaultRoleTimeout when the value is missing
	// or unparseable.
	// s.ctx is the server-level context cancelled by Shutdown(), so task
	// cancellation propagates automatically on server shutdown.
	roleTimeout := parseRoleTimeout(execution.Config.Delegation.Timeout)

	ctx, cancel := context.WithTimeout(s.ctx, roleTimeout)

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
	result, err := s.executeAgentWorkflow(ctx, roleTimeout, execution, prompt, roleContext, workingDir, logMsg)
	if err != nil {
		return
	}

	// Save and complete task
	s.saveAndCompleteTask(ctx, execution, result, startTime, logMsg)
}

func (s *AgentServer) executeAgentWorkflow(ctx context.Context, roleTimeout time.Duration, execution *TaskExecution, prompt, roleContext, workingDir string, logMsg func(string)) (string, error) {
	s.sendStreamEvent(execution, "api_call_start", map[string]interface{}{})

	// Lazy safety net: ensure the KG server is running for this project's root.
	s.ensureKGForProject(execution.ProjectRoot)

	// Pre-flight: inject knowledge-graph context into system prompt (best-effort, 2 s timeout).
	// Set AIPACK_KG_DISABLED=1 to suppress preflight for baseline/benchmarking runs.
	systemPrompt := s.buildSystemPromptForProject(roleContext, execution.ProjectRoot)
	kgPreflightBytes := 0
	if os.Getenv("AIPACK_KG_DISABLED") != "1" {
		if kgBlock := kgclient.PreflightContext(ctx, s.mcpManager, execution.Task, execution.ProjectRoot); kgBlock != "" {
			kgPreflightBytes = len(kgBlock)
			systemPrompt = kgBlock + "\n---\n\n" + systemPrompt
			s.mu.Lock()
			s.kgPreflightHits[execution.ProjectRoot]++
			s.mu.Unlock()
		}
	}

	// Pre-flight for related projects declared in the task contract.
	// Reads "Related Projects: /path" lines from 00-contract.md and merges each
	// project's KG context into the system prompt.
	if os.Getenv("AIPACK_KG_DISABLED") != "1" {
		if taskPacketPath := execution.metadata["task_packet_path"]; taskPacketPath != "" {
			contractPath := filepath.Join(workingDir, taskPacketPath, "00-contract.md")
			if relatedPaths := kgclient.ParseRelatedProjects(contractPath); len(relatedPaths) > 0 {
				for _, relProject := range relatedPaths {
					s.ensureKGForProject(relProject)
					if relBlock := kgclient.PreflightContext(ctx, s.mcpManager, execution.Task, relProject); relBlock != "" {
						kgPreflightBytes += len(relBlock)
						logMsg(fmt.Sprintf("   🔗 Merged KG context from related project: %s", relProject))
						systemPrompt = relBlock + "\n---\n\n" + systemPrompt
					}
				}
			}
		}
	}

	// Store kg_preflight_bytes in metadata so saveAndCompleteTask can include it in metrics.
	s.mu.Lock()
	if exec, ok := s.activeTasks[execution.TaskID]; ok {
		exec.metadata["kg_preflight_bytes"] = fmt.Sprintf("%d", kgPreflightBytes)
	}
	s.mu.Unlock()

	result, err := s.executeAgenticLoop(ctx, roleTimeout, execution.TaskID, execution.Role, prompt, systemPrompt, workingDir, execution.ProjectRoot, execution.Config, logMsg)
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

			// Write a minimal checkpoint so the task can be resumed
			// We don't have the full loop state here, but we can save what's available
			cp := &AgentCheckpoint{
				TaskID:        execution.TaskID,
				CreatedAt:     time.Now(),
				Turn:          0, // Unknown - agent will restart from beginning with more time
				PartialResult: result,
				ResumeReason:  "timeout",
				Role:          execution.Role,
				ProjectRoot:   execution.ProjectRoot,
				Model:         execution.Config.Model,
			}
			if err := writeCheckpoint(execution.ProjectRoot, execution.TaskID, cp); err != nil {
				logMsg(fmt.Sprintf("⚠️  Failed to write timeout checkpoint: %v", err))
			} else {
				logMsg(fmt.Sprintf("   💾 Checkpoint written for resume"))
			}

			// Mark task as failed with TIMEOUT prefix so resume can identify it
			s.cancelTaskExecution(execution, "TIMEOUT: Task exceeded time deadline")
			return "", fmt.Errorf("task timeout")
		}
		logMsg(fmt.Sprintf("❌ Agentic loop failed: %v", err))
		s.failTask(execution, fmt.Sprintf("Execution error: %v", err))
		return "", err
	}

	s.sendStreamEvent(execution, "api_call_complete", map[string]interface{}{})

	return result, nil
}

func (s *AgentServer) resumeFromCheckpoint(taskID, projectRoot string, cp *AgentCheckpoint, extendBudget int64, extendDuration time.Duration) {
	// Load config for the role (picks up any role-file changes since the pause)
	config, err := s.loadAgentConfig(cp.Role, projectRoot)
	if err != nil {
		monitoring.Logger.Error("resume_load_config_failed", "task_id", taskID, "error", err)
		return
	}

	// Calculate budget: either full reset or extended by specified amount
	roleBudget := int64(config.Delegation.MaxBudgetTokens)
	if extendBudget > 0 {
		// Extend by adding to existing budget used
		config.Delegation.MaxBudgetTokens = int(cp.BudgetUsed + extendBudget)
	} else {
		// Default behavior (extendBudget == 0): reset to full role budget
		config.Delegation.MaxBudgetTokens = int(roleBudget)
	}

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

	// Calculate timeout: either full reset or extended by specified duration
	roleTimeout := parseRoleTimeout(config.Delegation.Timeout)
	if extendDuration > 0 {
		// Extend by the specified duration instead of resetting
		roleTimeout = roleTimeout + extendDuration
	}
	// Default behavior (extendDuration == 0): use full roleTimeout (reset)

	// Derive the task context from the server context so that Shutdown()
	// pre-empts in-flight delegated tasks.
	ctx, cancel := context.WithTimeout(s.ctx, roleTimeout)
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

	// Log resume details
	var resumeDetails []string
	if extendDuration > 0 {
		resumeDetails = append(resumeDetails, fmt.Sprintf("timeout extended by %v (total: %v)", extendDuration, roleTimeout))
	} else {
		resumeDetails = append(resumeDetails, fmt.Sprintf("timeout reset to %v", roleTimeout))
	}

	if extendBudget > 0 {
		newTotal := cp.BudgetUsed + extendBudget
		resumeDetails = append(resumeDetails, fmt.Sprintf("budget extended by %d (total: %d)", extendBudget, newTotal))
	} else {
		resumeDetails = append(resumeDetails, fmt.Sprintf("budget reset to %d", roleBudget))
	}

	logMsg(fmt.Sprintf("▶️  Resuming task %s from checkpoint (%s)", taskID, strings.Join(resumeDetails, ", ")))

	// Update disk status to in_progress
	if err := s.updateTaskStatus(taskID, projectRoot, constants.StatusInProgress, ""); err != nil {
		logMsg(fmt.Sprintf("⚠️  Failed to update status: %v", err))
	}

	// Build a resume prompt that injects prior partial result as context
	var resumePrompt string
	if cp.ResumeReason == "timeout" {
		// Resumed from a timeout - inform the agent so it can adjust strategy
		// Use roleTimeout which reflects any extension applied
		originalTimeout := parseRoleTimeout(config.Delegation.Timeout)
		if cp.PartialResult != "" {
			if extendDuration > 0 {
				resumePrompt = fmt.Sprintf(
					"You are resuming work on task %s after a TIMEOUT.\n\n"+
						"== CONTEXT ==\n"+
						"Your previous attempt timed out after exceeding the %v time limit.\n"+
						"The timeout has been EXTENDED by %v (total: %v) to complete this work.\n\n"+
						"== PRIOR PARTIAL RESULT ==\n%s\n\n"+
						"== INSTRUCTIONS ==\n"+
						"Adjust your approach to work within the extended time limit. Consider:\n"+
						"- Breaking the work into smaller steps\n"+
						"- Prioritizing the most critical parts first\n"+
						"- Using more efficient tool calls\n"+
						"Continue from where you left off and complete the remaining work.",
					taskID, originalTimeout, extendDuration, roleTimeout, cp.PartialResult)
			} else {
				resumePrompt = fmt.Sprintf(
					"You are resuming work on task %s after a TIMEOUT.\n\n"+
						"== CONTEXT ==\n"+
						"Your previous attempt timed out after exceeding the %v time limit.\n"+
						"You now have a fresh %v time budget to complete this work.\n\n"+
						"== PRIOR PARTIAL RESULT ==\n%s\n\n"+
						"== INSTRUCTIONS ==\n"+
						"Adjust your approach to work within the time limit. Consider:\n"+
						"- Breaking the work into smaller steps\n"+
						"- Prioritizing the most critical parts first\n"+
						"- Using more efficient tool calls\n"+
						"Continue from where you left off and complete the remaining work.",
					taskID, originalTimeout, roleTimeout, cp.PartialResult)
			}
		} else {
			if extendDuration > 0 {
				resumePrompt = fmt.Sprintf(
					"You are resuming work on task %s after a TIMEOUT.\n\n"+
						"== CONTEXT ==\n"+
						"Your previous attempt timed out after exceeding the %v time limit.\n"+
						"The timeout has been EXTENDED by %v (total: %v) to complete this work.\n\n"+
						"== INSTRUCTIONS ==\n"+
						"Adjust your approach to work within the extended time limit. Start fresh and complete the task efficiently.",
					taskID, originalTimeout, extendDuration, roleTimeout)
			} else {
				resumePrompt = fmt.Sprintf(
					"You are resuming work on task %s after a TIMEOUT.\n\n"+
						"== CONTEXT ==\n"+
						"Your previous attempt timed out after exceeding the %v time limit.\n"+
						"You now have a fresh %v time budget to complete this work.\n\n"+
						"== INSTRUCTIONS ==\n"+
						"Adjust your approach to work within the time limit. Start fresh and complete the task efficiently.",
					taskID, originalTimeout, roleTimeout)
			}
		}
	} else {
		// Resumed from token budget pause
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
	}

	roleContext, err := s.loadAndLogRoleContext(execution, logMsg)
	if err != nil {
		logMsg(fmt.Sprintf("⚠️  Could not load role context: %v", err))
		roleContext = ""
	}

	_, workingDir := s.extractTaskMetadata(execution, logMsg)

	// Run the agentic loop
	result, loopErr := s.executeAgentWorkflow(ctx, roleTimeout, execution, resumePrompt, roleContext, workingDir, logMsg)

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

// shortTaskIDFromFullID strips the YYYYMMDD-HHMMSS timestamp suffix from a task ID
// to recover the bare task ID. E.g. "xasm++-fz5t-20260306-074606" → "xasm++-fz5t".
// If the ID has no timestamp suffix it is returned unchanged.
func shortTaskIDFromFullID(taskID string) string {
	parts := strings.Split(taskID, "-")
	n := len(parts)
	if n >= 2 {
		last := parts[n-1]
		secondLast := parts[n-2]
		if len(last) == 6 && len(secondLast) == 8 {
			allDigits := func(s string) bool {
				for _, c := range s {
					if c < '0' || c > '9' {
						return false
					}
				}
				return true
			}
			if allDigits(last) && allDigits(secondLast) {
				return strings.Join(parts[:n-2], "-")
			}
		}
	}
	return taskID
}
