package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/auth"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/config"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/protocol"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/proxy"
	"gopkg.in/yaml.v3"
)

const (
	Version = "2.0.0-phase2"
)

type AgentServer struct {
	rootDir       string
	anthropicKey  string
	client        *anthropic.Client
	maxConcurrent int // Maximum concurrent agents (configurable)
	maxTokens     int // Maximum tokens per API call
	model         string // Anthropic model to use

	// Concurrent execution tracking
	mu            sync.RWMutex
	activeTasks   map[string]*TaskExecution
	taskQueue     chan *TaskExecution
	workerPool    chan struct{} // Semaphore for max concurrent agents
}

type TaskExecution struct {
	TaskID    string
	Role      string
	Task      string
	Config    *AgentConfig
	StartTime time.Time
	Status    string // "queued", "in_progress", "completed", "failed"
	Progress  float64
	Result    string
	Error     string

	// Streaming
	streamChan chan *protocol.StreamEvent
	streamOpen bool
}

type AgentConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Tier        string `yaml:"tier"`
	Context     struct {
		RoleFile               string   `yaml:"role_file"`
		Gates                  []string `yaml:"gates"`
		AdditionalInstructions string   `yaml:"additional_instructions"`
	} `yaml:"context"`
	Delegation struct {
		Mode       string `yaml:"mode"`
		Timeout    string `yaml:"timeout"`
		MaxContext int    `yaml:"max_context"`
	} `yaml:"delegation"`
	Tools           []string          `yaml:"tools"`
	SuccessCriteria []string          `yaml:"success_criteria"`
	Metadata        map[string]string `yaml:"metadata"`
}

func NewAgentServer(rootDir string, maxConcurrent int, maxTokens int, model string, cfg *config.APIConfig) (*AgentServer, error) {
	// Try to get API key from multiple sources (env var or Claude Code helper)
	apiKey, err := auth.GetAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	// Build client options
	clientOpts := []option.RequestOption{option.WithAPIKey(apiKey)}

	// Add custom HTTP client for proxy support
	if httpClient := proxy.NewHTTPClient(cfg); httpClient != nil {
		clientOpts = append(clientOpts, option.WithHTTPClient(httpClient))
	}

	// Log proxy mode
	monitoring.Logger.Info("api_configuration", "proxy_mode", proxy.LogProxyMode(cfg))

	client := anthropic.NewClient(clientOpts...)

	server := &AgentServer{
		rootDir:       rootDir,
		anthropicKey:  apiKey,
		client:        client,
		maxConcurrent: maxConcurrent,
		maxTokens:     maxTokens,
		model:         model,
		activeTasks:   make(map[string]*TaskExecution),
		taskQueue:     make(chan *TaskExecution, 100),
		workerPool:    make(chan struct{}, maxConcurrent),
	}

	// Start worker pool
	go server.startWorkerPool()

	return server, nil
}

func (s *AgentServer) startWorkerPool() {
	for i := 0; i < s.maxConcurrent; i++ {
		go s.worker()
	}
}

func (s *AgentServer) worker() {
	for task := range s.taskQueue {
		// Acquire semaphore slot
		s.workerPool <- struct{}{}

		// Execute task
		s.executeAgentTask(task)

		// Release semaphore slot
		<-s.workerPool
	}
}

func (s *AgentServer) spawnAgentTask(role, task string) (*protocol.ExecuteTaskResponse, error) {
	// Generate task ID
	taskID := fmt.Sprintf("task-%s-%s", role, time.Now().Format("20060102-150405-000000"))

	// Load agent configuration
	config, err := s.loadAgentConfig(role)
	if err != nil {
		return nil, fmt.Errorf("failed to load agent config: %w", err)
	}

	// Create task packet
	if err := s.createTaskPacket(taskID, role, task, config); err != nil {
		return nil, fmt.Errorf("failed to create task packet: %w", err)
	}

	// Create task execution
	execution := &TaskExecution{
		TaskID:     taskID,
		Role:       role,
		Task:       task,
		Config:     config,
		StartTime:  time.Now(),
		Status:     "queued",
		Progress:   0.0,
		streamChan: make(chan *protocol.StreamEvent, 100),
		streamOpen: true,
	}

	// Register task
	s.mu.Lock()
	s.activeTasks[taskID] = execution
	s.mu.Unlock()

	// Queue for execution
	s.taskQueue <- execution

	// Build stream URL
	streamURL := fmt.Sprintf("/stream/%s", taskID)

	return &protocol.ExecuteTaskResponse{
		TaskID:    taskID,
		Status:    "queued",
		Message:   fmt.Sprintf("Agent %s queued for execution. Task ID: %s", role, taskID),
		StreamURL: streamURL,
		CreatedAt: time.Now(),
	}, nil
}

func (s *AgentServer) getTaskStatus(taskID string) (*protocol.TaskStatusResponse, error) {
	s.mu.RLock()
	execution, exists := s.activeTasks[taskID]
	s.mu.RUnlock()

	if !exists {
		// Try loading from disk
		return s.loadTaskStatusFromDisk(taskID)
	}

	response := &protocol.TaskStatusResponse{
		TaskID:    execution.TaskID,
		Role:      execution.Role,
		Task:      execution.Task,
		Status:    execution.Status,
		Progress:  execution.Progress,
		CreatedAt: execution.StartTime,
		UpdatedAt: time.Now(),
		Result:    execution.Result,
		Error:     execution.Error,
	}

	if execution.Status == "completed" || execution.Status == "failed" {
		completedAt := time.Now()
		response.CompletedAt = &completedAt
	}

	return response, nil
}

func (s *AgentServer) loadAgentConfig(role string) (*AgentConfig, error) {
	configPath := filepath.Join(s.rootDir, ".ai-pack", "agents", "lightweight", role+".yml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config AgentConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

func (s *AgentServer) loadRoleContext(roleFile string) (string, error) {
	rolePath := filepath.Join(s.rootDir, roleFile)

	data, err := os.ReadFile(rolePath)
	if err != nil {
		return "", fmt.Errorf("failed to read role file: %w", err)
	}

	return string(data), nil
}

func (s *AgentServer) createTaskPacket(taskID, role, task string, config *AgentConfig) error {
	taskDir := filepath.Join(s.rootDir, ".beads", "tasks", taskID)

	if err := os.MkdirAll(taskDir, 0755); err != nil {
		return fmt.Errorf("failed to create task directory: %w", err)
	}

	// Create metadata
	metadata := map[string]interface{}{
		"task_id":     taskID,
		"role":        role,
		"description": task,
		"tier":        config.Tier,
		"spawned_by":  "go-agent-server-v2",
		"spawned_at":  time.Now().Format(time.RFC3339),
		"status":      "queued",
		"config":      config,
		"updated_at":  time.Now().Format(time.RFC3339),
	}

	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(filepath.Join(taskDir, "00-metadata.json"), metadataJSON, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	// Create plan file
	planContent := fmt.Sprintf("# Task Plan: %s\n\n**Role**: %s\n**Task**: %s\n**Created**: %s\n\n## Execution Plan\n\n(Agent will populate during execution)\n",
		taskID, role, task, time.Now().Format(time.RFC3339))

	if err := os.WriteFile(filepath.Join(taskDir, "10-plan.md"), []byte(planContent), 0644); err != nil {
		return fmt.Errorf("failed to write plan: %w", err)
	}

	monitoring.Logger.Info("task_packet_created", "task_id", taskID, "path", taskDir)
	return nil
}

func (s *AgentServer) loadTaskStatusFromDisk(taskID string) (*protocol.TaskStatusResponse, error) {
	metadataPath := filepath.Join(s.rootDir, ".beads", "tasks", taskID, "00-metadata.json")

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Read results if available
	var result string
	resultsPath := filepath.Join(s.rootDir, ".beads", "tasks", taskID, "30-results.md")
	if resultData, err := os.ReadFile(resultsPath); err == nil {
		result = string(resultData)
	}

	createdAt, _ := time.Parse(time.RFC3339, metadata["spawned_at"].(string))
	updatedAt, _ := time.Parse(time.RFC3339, metadata["updated_at"].(string))

	response := &protocol.TaskStatusResponse{
		TaskID:    taskID,
		Role:      metadata["role"].(string),
		Task:      metadata["description"].(string),
		Status:    metadata["status"].(string),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Result:    result,
		Metadata:  metadata,
	}

	if metadata["error"] != nil {
		response.Error = metadata["error"].(string)
	}

	return response, nil
}

func (s *AgentServer) buildPrompt(role, task, roleContext string, config *AgentConfig) string {
	return fmt.Sprintf(`You are a %s agent.

**Your Role Context:**
%s

**Your Task:**
%s

**Configuration:**
- Timeout: %s
- Tools: %v
- Success Criteria: %v

**Instructions:**
1. Execute the task according to your role
2. Follow all quality gates and standards
3. Produce working, tested code where applicable
4. Document your work clearly

Please complete this task now.`,
		role,
		roleContext,
		task,
		config.Delegation.Timeout,
		config.Tools,
		config.SuccessCriteria,
	)
}

func (s *AgentServer) updateTaskStatus(taskID, status, errorMsg string) {
	metadataPath := filepath.Join(s.rootDir, ".beads", "tasks", taskID, "00-metadata.json")

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		monitoring.Logger.Error("metadata_read_error", "task_id", taskID, "error", err)
		return
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		monitoring.Logger.Error("metadata_unmarshal_error", "task_id", taskID, "error", err)
		return
	}

	metadata["status"] = status
	metadata["updated_at"] = time.Now().Format(time.RFC3339)
	if errorMsg != "" {
		metadata["error"] = errorMsg
	}

	updatedData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		monitoring.Logger.Error("metadata_marshal_error", "task_id", taskID, "error", err)
		return
	}

	if err := os.WriteFile(metadataPath, updatedData, 0644); err != nil {
		monitoring.Logger.Error("metadata_write_error", "task_id", taskID, "error", err)
	}
}

func (s *AgentServer) executeAgentTask(execution *TaskExecution) {
	ctx := context.Background()
	taskID := execution.TaskID
	startTime := time.Now()

	monitoring.LogTaskStarted(ctx, taskID, execution.Role)
	monitoring.GlobalMetrics.IncrementTasksSpawned()

	// Update status
	s.mu.Lock()
	execution.Status = "in_progress"
	execution.Progress = 0.1
	s.mu.Unlock()
	s.updateTaskStatus(taskID, "in_progress", "")

	// Send stream event
	s.sendStreamEvent(execution, "status_update", map[string]interface{}{
		"status":   "in_progress",
		"progress": 0.1,
	})

	// Load role context
	roleContext, err := s.loadRoleContext(execution.Config.Context.RoleFile)
	if err != nil {
		monitoring.Logger.Error("role_context_load_error", "task_id", taskID, "error", err)
		s.failTask(execution, fmt.Sprintf("Failed to load role context: %v", err))
		return
	}

	// Build prompt
	prompt := s.buildPrompt(execution.Role, execution.Task, roleContext, execution.Config)

	// Save agent prompt
	promptPath := filepath.Join(s.rootDir, ".beads", "tasks", taskID, "agent-prompt.txt")
	if err := os.WriteFile(promptPath, []byte(prompt), 0644); err != nil {
		monitoring.Logger.Warn("prompt_save_error", "task_id", taskID, "error", err)
	}

	// Update progress
	s.mu.Lock()
	execution.Progress = 0.3
	s.mu.Unlock()
	s.sendStreamEvent(execution, "api_call_start", map[string]interface{}{
		"progress": 0.3,
	})

	// Execute via Anthropic API
	apiStartTime := time.Now()
	monitoring.Logger.Info("api_call_start", "task_id", taskID, "model", "claude-3-5-sonnet-20241022")

	message, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(s.model),
		MaxTokens: anthropic.F(int64(s.maxTokens)),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		}),
	})

	apiDurationMs := time.Since(apiStartTime).Milliseconds()

	if err != nil {
		monitoring.Logger.Error("api_call_failed", "task_id", taskID, "error", err, "duration_ms", apiDurationMs)
		monitoring.GlobalMetrics.IncrementAPICallsFailed()
		s.failTask(execution, fmt.Sprintf("API error: %v", err))
		return
	}

	monitoring.GlobalMetrics.IncrementAPICallsSuccess()
	monitoring.LogAPICall(ctx, taskID, "claude-3-5-sonnet-20241022", int(message.Usage.InputTokens+message.Usage.OutputTokens))

	// Extract response
	var result string
	for _, block := range message.Content {
		if textBlock, ok := block.AsUnion().(anthropic.TextBlock); ok {
			result += textBlock.Text
		}
	}

	// Update progress
	s.mu.Lock()
	execution.Progress = 0.9
	s.mu.Unlock()
	s.sendStreamEvent(execution, "api_call_complete", map[string]interface{}{
		"progress": 0.9,
	})

	// Save results
	resultsPath := filepath.Join(s.rootDir, ".beads", "tasks", taskID, "30-results.md")
	resultsContent := fmt.Sprintf("# Task Results: %s\n\n**Role**: %s\n**Task**: %s\n**Completed**: %s\n\n## Agent Output\n\n%s\n",
		taskID, execution.Role, execution.Task, time.Now().Format(time.RFC3339), result)

	if err := os.WriteFile(resultsPath, []byte(resultsContent), 0644); err != nil {
		monitoring.Logger.Warn("results_save_error", "task_id", taskID, "error", err)
	}

	// Complete task
	s.mu.Lock()
	execution.Status = "completed"
	execution.Progress = 1.0
	execution.Result = result
	s.mu.Unlock()

	durationMs := time.Since(startTime).Milliseconds()
	s.updateTaskStatus(taskID, "completed", "")
	s.sendStreamEvent(execution, "completed", map[string]interface{}{
		"progress": 1.0,
		"result":   result,
	})

	// Close stream
	s.closeStream(execution)

	monitoring.LogTaskCompleted(ctx, taskID, execution.Role, durationMs)
	monitoring.GlobalMetrics.IncrementTasksCompleted(durationMs)
}

func (s *AgentServer) failTask(execution *TaskExecution, errorMsg string) {
	ctx := context.Background()
	durationMs := time.Since(execution.StartTime).Milliseconds()

	s.mu.Lock()
	execution.Status = "failed"
	execution.Error = errorMsg
	s.mu.Unlock()

	s.updateTaskStatus(execution.TaskID, "failed", errorMsg)
	s.sendStreamEvent(execution, "failed", map[string]interface{}{
		"error": errorMsg,
	})
	s.closeStream(execution)

	monitoring.LogTaskFailed(ctx, execution.TaskID, execution.Role, errorMsg, durationMs)
	monitoring.GlobalMetrics.IncrementTasksFailed(durationMs)
}

func (s *AgentServer) sendStreamEvent(execution *TaskExecution, eventType string, data map[string]interface{}) {
	if !execution.streamOpen {
		return
	}

	event := &protocol.StreamEvent{
		Type:      eventType,
		TaskID:    execution.TaskID,
		Timestamp: time.Now(),
		Data:      data,
	}

	select {
	case execution.streamChan <- event:
	default:
		// Channel full, skip event
	}
}

func (s *AgentServer) closeStream(execution *TaskExecution) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if execution.streamOpen {
		execution.streamOpen = false
		close(execution.streamChan)
	}
}
