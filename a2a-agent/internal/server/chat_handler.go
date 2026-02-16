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
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/streaming"
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

	// Continue with chat mode (using clean streaming architecture)
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
	streamReq := streaming.StreamRequest{
		Messages:  make([]streaming.Message, 0),
		MaxTokens: s.maxTokens,
		Model:     s.model, // Default model, will be overridden by role config
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

	// Load role context as system prompt
	if req.Role != "" {
		roleFile := fmt.Sprintf("../.ai-pack/agents/%s.md", req.Role)
		if req.Role == "orchestrator" {
			roleFile = "../.ai-pack/agents/orchestrator-chat.md"
		}

		roleContext, err := s.loadRoleContext(roleFile)
		if err != nil {
			monitoring.Logger.Warn("chat_role_load_failed", "role", req.Role, "error", err)
		} else {
			streamReq.SystemPrompt = roleContext
		}

		// TODO: Add tools for orchestrator
		// Tools need to be defined in provider-agnostic format first
		// For now, orchestrator will work without tools until we implement
		// proper tool abstraction in the streaming layer
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

	// Stream events
	for stream.Next() {
		event := stream.Current()

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
