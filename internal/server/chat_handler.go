package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/cortexa-llc/ai-pack/internal/constants"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/internal/streaming"
)

// convertAnthropicToolsToStreaming converts Anthropic tool format to provider-agnostic streaming format
func convertAnthropicToolsToStreaming(anthropicTools []anthropic.ToolUnionParam) []streaming.Tool {
	tools := make([]streaming.Tool, 0, len(anthropicTools))

	for _, toolUnion := range anthropicTools {
		// Extract the tool from the union type
		tool := toolUnion.OfTool
		if tool == nil {
			continue
		}

		// Convert to streaming.Tool format
		desc := ""
		if d := toolUnion.GetDescription(); d != nil {
			desc = *d
		}
		streamTool := streaming.Tool{
			Name:        tool.Name,
			Description: desc,
			InputSchema: map[string]interface{}{
				"type":       tool.InputSchema.Type,
				"properties": tool.InputSchema.Properties,
				"required":   tool.InputSchema.Required,
			},
		}
		tools = append(tools, streamTool)
	}

	return tools
}

// sendToolUseEvent sends a tool_use SSE event to the client
func sendToolUseEvent(w http.ResponseWriter, flusher http.Flusher, toolCall streaming.ToolUse) {
	data, _ := json.Marshal(map[string]interface{}{
		"tool":  toolCall.Name,
		"id":    toolCall.ID,
		"input": toolCall.Input,
	})
	fmt.Fprintf(w, "event: tool_use\n")
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

// sendToolResultEvent sends a tool_result SSE event to the client
func sendToolResultEvent(w http.ResponseWriter, flusher http.Flusher, toolCall streaming.ToolUse, result string, isError bool) {
	data, _ := json.Marshal(map[string]interface{}{
		"tool":     toolCall.Name,
		"id":       toolCall.ID,
		"result":   result,
		"is_error": isError,
	})
	fmt.Fprintf(w, "event: tool_result\n")
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

// continueWithToolResults continues the conversation after tool execution, supporting
// multiple rounds of tool use until the model produces a final text response.
func (s *AgentServer) continueWithToolResults(
	ctx context.Context,
	w http.ResponseWriter,
	flusher http.Flusher,
	req *ChatRequest,
	previousReq streaming.StreamRequest,
	previousMessage *streaming.CompletedMessage,
	toolUses []streaming.ToolUse,
	toolResults []streaming.ToolResult,
) error {
	// Build initial continuation request
	continuationReq := streaming.StreamRequest{
		Messages: append(previousReq.Messages,
			streaming.Message{
				Role:     "assistant",
				Content:  previousMessage.Content,
				ToolUses: toolUses,
			},
			streaming.Message{
				Role:        "user",
				ToolResults: toolResults,
			},
		),
		SystemPrompt: previousReq.SystemPrompt,
		MaxTokens:    previousReq.MaxTokens,
		Tools:        previousReq.Tools,
		Model:        previousReq.Model,
	}

	const maxToolRounds = 10
	for round := 0; round < maxToolRounds; round++ {
		stream, err := s.streamingService.CreateStream(ctx, req.Role, continuationReq)
		if err != nil {
			return fmt.Errorf("failed to create continuation stream: %w", err)
		}

		roundToolCalls := []streaming.ToolUse{}
		var roundContent string

		for stream.Next() {
			event := stream.Current()
			if event.ToolUse != nil {
				roundToolCalls = append(roundToolCalls, *event.ToolUse)
				monitoring.Logger.Info("tool_use_in_continuation", "tool", event.ToolUse.Name, "round", round+1)
				continue
			}
			if event.Delta != nil && event.Delta.Text != "" {
				roundContent += event.Delta.Text
				textData, _ := json.Marshal(map[string]interface{}{"text": event.Delta.Text})
				fmt.Fprintf(w, "event: delta\n")
				fmt.Fprintf(w, "data: %s\n\n", textData)
				flusher.Flush()
			}
		}

		streamErr := stream.Err()
		finalMessage := stream.GetMessage()
		stream.Close() //nolint:errcheck

		if streamErr != nil {
			return fmt.Errorf("continuation stream error: %w", streamErr)
		}

		if len(roundToolCalls) == 0 {
			// No more tool calls — send completion and return
			completeData, _ := json.Marshal(map[string]interface{}{
				"status":        "complete",
				"input_tokens":  finalMessage.InputTokens,
				"output_tokens": finalMessage.OutputTokens,
				"model":         finalMessage.Model,
			})
			fmt.Fprintf(w, "event: complete\n")
			fmt.Fprintf(w, "data: %s\n\n", completeData)
			flusher.Flush()
			return nil
		}

		// Execute tools and notify the client
		monitoring.Logger.Info("executing_tools_continuation", "count", len(roundToolCalls), "round", round+1)
		roundToolResults := []streaming.ToolResult{}
		for _, toolCall := range roundToolCalls {
			sendToolUseEvent(w, flusher, toolCall)
			result, execErr := s.ExecuteTool(toolCall.Name, toolCall.Input)
			if execErr != nil {
				monitoring.Logger.Error("tool_execution_failed", "tool", toolCall.Name, "round", round+1, "error", execErr)
				errStr := fmt.Sprintf("Error: %v", execErr)
				sendToolResultEvent(w, flusher, toolCall, errStr, true)
				roundToolResults = append(roundToolResults, streaming.ToolResult{
					ToolUseID: toolCall.ID,
					ToolName:  toolCall.Name,
					Content:   errStr,
					IsError:   true,
				})
			} else {
				sendToolResultEvent(w, flusher, toolCall, result, false)
				roundToolResults = append(roundToolResults, streaming.ToolResult{
					ToolUseID: toolCall.ID,
					ToolName:  toolCall.Name,
					Content:   result,
					IsError:   false,
				})
			}
		}

		// If this round produced text before tool calls, send a paragraph break
		// so the next round's reply starts on a new line.
		if roundContent != "" {
			sepData, _ := json.Marshal(map[string]interface{}{"text": "\n\n"})
			fmt.Fprintf(w, "event: delta\n")
			fmt.Fprintf(w, "data: %s\n\n", sepData)
			flusher.Flush()
		}

		// Append this round to the conversation history for the next turn
		continuationReq.Messages = append(continuationReq.Messages,
			streaming.Message{
				Role:     "assistant",
				Content:  roundContent,
				ToolUses: roundToolCalls,
			},
			streaming.Message{
				Role:        "user",
				ToolResults: roundToolResults,
			},
		)
	}

	return fmt.Errorf("exceeded maximum tool call rounds (%d)", maxToolRounds)
}

// ChatRequest represents an incoming chat message
type ChatRequest struct {
	Message           string        `json:"message"`
	Messages          []ChatMessage `json:"messages,omitempty"`            // Full conversation history
	Role              string        `json:"role,omitempty"`                // Agent role (orchestrator, engineer, etc.)
	Model             string        `json:"model,omitempty"`               // Override model (e.g. "Qwen3.5-35B-A3B-Q6_K.gguf")
	Mode              string        `json:"mode,omitempty"`                // "chat" or "agent"
	ProjectRoot       string        `json:"project_root,omitempty"`        // Working directory for agent mode
	UseProjectContext bool          `json:"use_project_context,omitempty"` // Whether to load project context files
}

// ChatMessage represents a message in the conversation
type ChatMessage struct {
	Role    string `json:"role"` // "user" or "assistant"
	Content string `json:"content"`
}

// ChatResponse represents the response from a chat request
type ChatResponse struct {
	Status        string `json:"status"`            // "streaming", "complete", "agent_spawned"
	TaskID        string `json:"task_id,omitempty"` // For agent mode
	Text          string `json:"text,omitempty"`
	ContextLoaded string `json:"context_loaded,omitempty"` // Name of context file loaded (e.g., "CLAUDE.md")
}

// HandleChat handles chat requests with streaming responses
func (s *AgentServer) HandleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}

	// Default to chat mode if not specified
	if req.Mode == "" {
		req.Mode = "chat"
	}

	// Handle agent mode - spawn a task instead of chatting
	if req.Mode == "agent" {
		s.handleAgentMode(w, r, &req)
		return
	}

	// Continue with chat mode (using clean streaming architecture)
	s.handleChatMode(w, r, &req)
}

// handleAgentMode spawns an agent task
func (s *AgentServer) handleAgentMode(w http.ResponseWriter, r *http.Request, req *ChatRequest) {
	// Role is required for agent mode
	role := req.Role
	if role == "" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"status":  "error",
			"message": "Role is required for agent mode.",
		})
		return
	}

	// Require project root for agent mode - agents must work in a specific project context
	projectRoot := req.ProjectRoot
	if projectRoot == "" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"status":  "error",
			"message": "Project root is required for agent mode. Please select a project directory first.",
		})
		monitoring.Logger.Warn("chat_agent_mode_no_project_root")
		return
	}

	// Validate project root exists
	if _, err := os.Stat(projectRoot); os.IsNotExist(err) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"status":  "error",
			"message": fmt.Sprintf("Project root does not exist: %s", projectRoot),
		})
		monitoring.Logger.Warn("chat_agent_mode_invalid_project_root", "path", projectRoot)
		return
	}

	// Create a task from the message first using bd create command
	cmd := exec.Command("bd", "create", req.Message)
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"status":  "error",
			"message": fmt.Sprintf("Failed to create task: %v (output: %s)", err, string(output)),
		})
		return
	}

	// Parse task ID from output - look for "Created issue: <task-id>"
	outputStr := string(output)
	taskID := ""
	for _, line := range strings.Split(outputStr, "\n") {
		if strings.Contains(line, "Created issue:") {
			// Extract task ID from line like "✓ Created issue: xasm++-h0y2"
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "issue:" && i+1 < len(parts) {
					taskID = parts[i+1]
					break
				}
			}
			break
		}
	}

	if taskID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"status":  "error",
			"message": fmt.Sprintf("Failed to parse task ID from output: %s", outputStr),
		})
		return
	}

	monitoring.Logger.Info("chat_agent_mode_task_created", "task_id", taskID, "role", role, "project", projectRoot)

	// Spawn agent task with the task ID
	response, err := s.spawnAgentTask(role, taskID, projectRoot)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"status":  "error",
			"message": fmt.Sprintf("Failed to spawn agent: %v", err),
		})
		return
	}

	// Return task info as JSON with clear project context
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get relative project name for display
	projectName := projectRoot
	if parts := strings.Split(projectRoot, "/"); len(parts) > 0 {
		projectName = parts[len(parts)-1]
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "agent_spawned",
		"task_id":       taskID,
		"project_root":  projectRoot,
		"project_name":  projectName,
		"role":          role,
		"message":       fmt.Sprintf("Agent task spawned with ID: %s (task: %s)", response.TaskID, taskID),
	})
}

// handleChatMode handles conversational chat using the clean streaming architecture
func (s *AgentServer) handleChatMode(w http.ResponseWriter, r *http.Request, req *ChatRequest) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Build generic streaming request
	model := s.model
	if req.Model != "" {
		model = req.Model
	}
	streamReq := streaming.StreamRequest{
		Messages:  make([]streaming.Message, 0),
		MaxTokens: s.maxTokens,
		Model:     model,
	}

	// Add conversation history
	for _, msg := range req.Messages {
		streamReq.Messages = append(streamReq.Messages, streaming.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Add current message
	streamReq.Messages = append(streamReq.Messages, streaming.Message{
		Role:    "user",
		Content: req.Message,
	})

	// Load role context and config as system prompt
	if req.Role != "" {
		roleConfig, err := s.loadAgentConfigForRole(req.Role)
		if err != nil {
			monitoring.Logger.Warn("chat_role_load_failed", "role", req.Role, "error", err)
		} else {
			streamReq.SystemPrompt = roleConfig.Context.RoleContent

			// Inject chat tools if the role config declares ChatTools: true
			if roleConfig.ChatTools {
				anthropicTools := GetOrchestratorTools()
				streamReq.Tools = convertAnthropicToolsToStreaming(anthropicTools)
				monitoring.Logger.Info("chat_tools_enabled", "role", req.Role, "tool_count", len(streamReq.Tools))
			}
		}
	}

	// Create stream using the service (handles model selection automatically)
	ctx := context.Background()
	stream, err := s.streamingService.CreateStream(ctx, req.Role, streamReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create stream: %v", err), http.StatusInternalServerError)
		return
	}
	defer stream.Close()

	// Send connected event
	fmt.Fprintf(w, "event: connected\n")
	fmt.Fprintf(w, "data: {\"status\":\"connected\"}\n\n")
	flusher.Flush()

	// Stream events and handle tool calls
	toolCalls := []streaming.ToolUse{}

	for stream.Next() {
		event := stream.Current()

		// Collect tool use events
		if event.ToolUse != nil {
			toolCalls = append(toolCalls, *event.ToolUse)
			monitoring.Logger.Info("tool_use_detected", "tool", event.ToolUse.Name, "id", event.ToolUse.ID)
			continue
		}

		// Send delta events
		if event.Delta != nil && event.Delta.Text != "" {
			textData, _ := json.Marshal(map[string]interface{}{
				"text": event.Delta.Text,
			})
			fmt.Fprintf(w, "event: delta\n")
			fmt.Fprintf(w, "data: %s\n\n", textData)
			flusher.Flush()
		}
	}

	// Check for errors
	if err := stream.Err(); err != nil {
		monitoring.Logger.Error("chat_stream_error", "error", err)
		errorData, _ := json.Marshal(map[string]interface{}{
			"error": err.Error(),
		})
		fmt.Fprintf(w, "event: error\n")
		fmt.Fprintf(w, "data: %s\n\n", errorData)
		flusher.Flush()
		return
	}

	// Get final message with token usage
	message := stream.GetMessage()

	// Handle tool calls if any were detected
	if len(toolCalls) > 0 {
		monitoring.Logger.Info("executing_tools", "count", len(toolCalls))

		// Execute each tool and collect results
		toolResults := []streaming.ToolResult{}
		for _, toolCall := range toolCalls {
			sendToolUseEvent(w, flusher, toolCall)
			result, err := s.ExecuteTool(toolCall.Name, toolCall.Input)
			if err != nil {
				monitoring.Logger.Error("tool_execution_failed", "tool", toolCall.Name, "error", err)
				errStr := fmt.Sprintf("Error: %v", err)
				sendToolResultEvent(w, flusher, toolCall, errStr, true)
				toolResults = append(toolResults, streaming.ToolResult{
					ToolUseID: toolCall.ID,
					ToolName:  toolCall.Name,
					Content:   errStr,
					IsError:   true,
				})
			} else {
				sendToolResultEvent(w, flusher, toolCall, result, false)
				toolResults = append(toolResults, streaming.ToolResult{
					ToolUseID: toolCall.ID,
					ToolName:  toolCall.Name,
					Content:   result,
					IsError:   false,
				})
			}
		}

		// If the model produced text before the tool calls, send a paragraph break
		// so the continuation reply starts on a new line rather than running together.
		if message.Content != "" {
			sepData, _ := json.Marshal(map[string]interface{}{"text": "\n\n"})
			fmt.Fprintf(w, "event: delta\n")
			fmt.Fprintf(w, "data: %s\n\n", sepData)
			flusher.Flush()
		}

		// Continue conversation with tool results
		if err := s.continueWithToolResults(ctx, w, flusher, req, streamReq, message, toolCalls, toolResults); err != nil {
			monitoring.Logger.Error("tool_continuation_failed", "error", err)
			errorData, _ := json.Marshal(map[string]interface{}{
				"error": fmt.Sprintf("Tool continuation failed: %v", err),
			})
			fmt.Fprintf(w, "event: error\n")
			fmt.Fprintf(w, "data: %s\n\n", errorData)
			flusher.Flush()
		}
		return
	}

	// Send completion event
	completeData, _ := json.Marshal(map[string]interface{}{
		"status":        "complete",
		"input_tokens":  message.InputTokens,
		"output_tokens": message.OutputTokens,
		"model":         message.Model,
	})
	fmt.Fprintf(w, "event: complete\n")
	fmt.Fprintf(w, "data: %s\n\n", completeData)
	flusher.Flush()

	monitoring.Logger.Info("chat_complete",
		"role", req.Role,
		"model", message.Model,
		"input_tokens", message.InputTokens,
		"output_tokens", message.OutputTokens)

	// Record API call success
	monitoring.GlobalMetrics.IncrementAPICallsSuccess()

	// Record provider usage (provider comes from streaming layer)
	monitoring.GlobalMetrics.RecordProviderUsage(message.Provider, message.Model, int64(message.InputTokens), int64(message.OutputTokens))

	// Record per-project persistent daily usage (use server root for chat)
	if pm, err := s.getOrCreateProjectMetrics(s.rootDir); err == nil {
		if err := pm.RecordUsage(message.Provider, message.Model, int64(message.InputTokens), int64(message.OutputTokens)); err != nil {
			monitoring.Logger.Warn("failed_to_record_chat_metrics", "error", err.Error())
		}
	}
}

// generateFollowUpSuggestion asks Claude for a relevant follow-up question
func (s *AgentServer) generateFollowUpSuggestion(ctx context.Context, assistantResponse string) string {
	// Create a prompt asking for a single follow-up suggestion
	prompt := fmt.Sprintf(`Based on this response I just gave:

%s

Suggest ONE brief, natural follow-up question the user might want to ask. Return ONLY the question text itself, no explanation, no quotes, no preamble. Keep it under 60 characters if possible.`, assistantResponse)

	// Make a non-streaming request to Claude
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(s.model),
		MaxTokens: 100,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	}

	response, err := s.client.Messages.New(ctx, params)
	if err != nil {
		monitoring.Logger.Warn("suggestion_generation_failed", "error", err)
		return "" // Return empty string on error
	}

	// Extract suggestion text
	if len(response.Content) > 0 && response.Content[0].Type == constants.ContentTypeText {
		suggestion := strings.TrimSpace(response.Content[0].Text)
		// Remove quotes if present
		suggestion = strings.Trim(suggestion, "\"'")
		return suggestion
	}

	return ""
}

// loadProjectContext loads project context from CLAUDE.md or README.md
func (s *AgentServer) loadProjectContext(projectRoot string) (string, string, error) {
	// Try CLAUDE.md first (AI-specific context)
	contextFiles := []string{"CLAUDE.md", "README.md", ".ai/context.md"}

	for _, filename := range contextFiles {
		filePath := fmt.Sprintf("%s/%s", projectRoot, filename)
		content, err := os.ReadFile(filePath)
		if err == nil {
			return string(content), filename, nil
		}
	}

	return "", "", fmt.Errorf("no context file found")
}

// HandleChatOptions handles CORS preflight for chat endpoint
func (s *AgentServer) HandleChatOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(http.StatusOK)
}

// ModelInfo describes an available model for the chat UI.
type ModelInfo struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
}

// HandleModels returns available models: local Qwen models + known cloud models.
func (s *AgentServer) HandleModels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", constants.ContentTypeJSON)
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var models []ModelInfo

	// Fetch local Qwen models from llama-server
	qwenBaseURL := os.Getenv("QWEN_BASE_URL")
	if qwenBaseURL == "" {
		qwenBaseURL = constants.QwenLocalBaseURL
	}
	client := &http.Client{Timeout: 2 * time.Second}
	if resp, err := client.Get(qwenBaseURL + "/models"); err == nil {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var result struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		if json.Unmarshal(body, &result) == nil {
			for _, m := range result.Data {
				models = append(models, ModelInfo{ID: m.ID, Provider: constants.ProviderQwen})
			}
		}
	}

	// Always include cloud fallbacks
	cloudModels := []ModelInfo{
		{ID: "claude-sonnet-4-6", Provider: constants.ProviderAnthropic},
		{ID: "claude-haiku-4-5-20251001", Provider: constants.ProviderAnthropic},
	}
	if s.openaiKey != "" {
		cloudModels = append(cloudModels,
			ModelInfo{ID: "gpt-4.1-mini", Provider: constants.ProviderOpenAI},
			ModelInfo{ID: "gpt-4.1", Provider: constants.ProviderOpenAI},
		)
	}
	models = append(models, cloudModels...)

	json.NewEncoder(w).Encode(map[string]interface{}{"models": models})
}
