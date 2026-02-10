package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
)

// ChatRequest represents an incoming chat message
type ChatRequest struct {
	Message           string          `json:"message"`
	Messages          []ChatMessage   `json:"messages,omitempty"` // Full conversation history
	Role              string          `json:"role,omitempty"`     // Agent role (orchestrator, engineer, etc.)
	Mode              string          `json:"mode,omitempty"`     // "chat" or "agent"
	ProjectRoot       string          `json:"project_root,omitempty"` // Working directory for agent mode
	UseProjectContext bool            `json:"use_project_context,omitempty"` // Whether to load project context files
}

// ChatMessage represents a message in the conversation
type ChatMessage struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"`
}

// ChatResponse represents the response from a chat request
type ChatResponse struct {
	Status        string `json:"status"` // "streaming", "complete", "agent_spawned"
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

	// Continue with chat mode
	s.handleChatMode(w, r, &req)
}

// handleAgentMode spawns an agent task
func (s *AgentServer) handleAgentMode(w http.ResponseWriter, r *http.Request, req *ChatRequest) {
	// Default role if not specified
	role := req.Role
	if role == "" {
		role = "engineer"
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

	// Create a Beads task from the message first using bd create command
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
			"message": fmt.Sprintf("Failed to create Beads task: %v (output: %s)", err, string(output)),
		})
		return
	}

	// Parse task ID from output - look for "Created issue: <task-id>"
	outputStr := string(output)
	beadsTaskID := ""
	for _, line := range strings.Split(outputStr, "\n") {
		if strings.Contains(line, "Created issue:") {
			// Extract task ID from line like "✓ Created issue: xasm++-h0y2"
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "issue:" && i+1 < len(parts) {
					beadsTaskID = parts[i+1]
					break
				}
			}
			break
		}
	}

	if beadsTaskID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"status":  "error",
			"message": fmt.Sprintf("Failed to parse Beads task ID from output: %s", outputStr),
		})
		return
	}

	monitoring.Logger.Info("chat_agent_mode_task_created", "task_id", beadsTaskID, "role", role, "project", projectRoot)

	// Spawn agent task with the Beads task ID
	response, err := s.spawnAgentTask(role, beadsTaskID, projectRoot)
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
		"status":       "agent_spawned",
		"task_id":      response.TaskID,
		"beads_task_id": beadsTaskID,
		"project_root": projectRoot,
		"project_name": projectName,
		"role":         role,
		"message":      fmt.Sprintf("Agent task spawned with ID: %s (Beads task: %s)", response.TaskID, beadsTaskID),
	})
}

// handleChatMode handles conversational chat with optional role context
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

	// Track what context was loaded
	contextLoaded := ""

	// Load role context if specified
	var systemPrompt []anthropic.TextBlockParam
	if req.Role != "" {
		roleFile := fmt.Sprintf(".ai-pack/agents/%s.md", req.Role)
		roleContext, err := s.loadRoleContext(roleFile)
		if err != nil {
			monitoring.Logger.Warn("chat_role_load_failed", "role", req.Role, "error", err)
			// Continue without role context rather than failing
		} else {
			systemPrompt = []anthropic.TextBlockParam{
				{
					Text: roleContext,
					Type: ContentTypeText,
				},
			}
		}
	}

	// Load project context if enabled and project root is provided
	if req.UseProjectContext && req.ProjectRoot != "" {
		projectContext, filename, err := s.loadProjectContext(req.ProjectRoot)
		if err == nil && projectContext != "" {
			contextLoaded = filename
			// Prepend project context to system prompt
			contextBlock := anthropic.TextBlockParam{
				Text: fmt.Sprintf("# Project Context\n\n%s", projectContext),
				Type: ContentTypeText,
			}
			if len(systemPrompt) > 0 {
				systemPrompt = append([]anthropic.TextBlockParam{contextBlock}, systemPrompt...)
			} else {
				systemPrompt = []anthropic.TextBlockParam{contextBlock}
			}
			monitoring.Logger.Info("chat_project_context_loaded", "file", filename, "project", req.ProjectRoot)
		}
	}

	// Build messages array for Claude
	var messages []anthropic.MessageParam

	// Add conversation history if provided
	for _, msg := range req.Messages {
		if msg.Role == "user" {
			messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
		} else if msg.Role == "assistant" {
			messages = append(messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content)))
		}
	}

	// Add current message
	messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(req.Message)))

	// Create streaming request
	ctx := context.Background()
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(s.model),
		MaxTokens: int64(s.maxTokens),
		Messages:  messages,
	}

	// Add system prompt if we have role context
	if len(systemPrompt) > 0 {
		params.System = systemPrompt
	}

	// Add tools for orchestrator role
	if req.Role == "orchestrator" {
		tools := GetOrchestratorTools()
		params.Tools = tools
		monitoring.Logger.Info("orchestrator_tools_enabled", "count", len(tools))
	}

	stream := s.client.Messages.NewStreaming(ctx, params)

	// Send connected event
	fmt.Fprintf(w, "event: connected\n")
	fmt.Fprintf(w, "data: {\"status\":\"connected\"}\n\n")
	flusher.Flush()

	// Stream deltas and accumulate the message
	var message anthropic.Message
	for stream.Next() {
		event := stream.Current()

		// Log event type for debugging
		monitoring.Logger.Debug("chat_stream_event", "type", event.Type)

		// Send delta events for content blocks
		if event.Type == ContentBlockDelta {
			// Extract text from delta
			deltaJSON, _ := json.Marshal(event)
			var deltaData map[string]interface{}
			json.Unmarshal(deltaJSON, &deltaData)

			monitoring.Logger.Debug("chat_delta_data", "data", deltaData)

			if delta, ok := deltaData["delta"].(map[string]interface{}); ok {
				if text, ok := delta["text"].(string); ok {
					// Send delta to client
					textData, _ := json.Marshal(map[string]interface{}{
						"text": text,
					})
					fmt.Fprintf(w, "event: delta\n")
					fmt.Fprintf(w, "data: %s\n\n", textData)
					flusher.Flush()
				}
			}
		}

		if err := message.Accumulate(event); err != nil {
			monitoring.Logger.Error("chat_accumulate_error", "error", err)
			continue
		}
	}

	if err := stream.Err(); err != nil {
		monitoring.Logger.Error("chat_stream_error", "error", err)
		fmt.Fprintf(w, "event: error\n")
		errMsg := strings.ReplaceAll(err.Error(), "\"", "\\\"")
		fmt.Fprintf(w, "data: {\"error\":\"%s\"}\n\n", errMsg)
		flusher.Flush()
		return
	}

	// Get full text from message
	fullText := ""
	for _, block := range message.Content {
		if block.Type == "text" {
			fullText += block.Text
		}
	}

	// Generate a follow-up suggestion from Claude
	suggestion := s.generateFollowUpSuggestion(ctx, fullText)

	// Send completion event with full message, suggestion, and context info
	completionData, _ := json.Marshal(map[string]interface{}{
		"status":         "complete",
		"text":           fullText,
		"suggestion":     suggestion,
		"context_loaded": contextLoaded,
		"usage": map[string]int{
			"input_tokens":  int(message.Usage.InputTokens),
			"output_tokens": int(message.Usage.OutputTokens),
		},
	})

	fmt.Fprintf(w, "event: complete\n")
	fmt.Fprintf(w, "data: %s\n\n", completionData)
	flusher.Flush()

	// Log metrics
	monitoring.Logger.Info("chat_completed",
		"input_tokens", message.Usage.InputTokens,
		"output_tokens", message.Usage.OutputTokens)
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
	if len(response.Content) > 0 && response.Content[0].Type == "text" {
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
