package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/auth"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/beads"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/config"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/protocol"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/proxy"
	"gopkg.in/yaml.v3"
)

const (
	Version = "2.1.0"
)

type AgentServer struct {
	rootDir       string
	anthropicKey  string
	client        *anthropic.Client
	beadsClient   *beads.Client
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

	// Beads integration
	metadata map[string]string

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
	Tools           []string               `yaml:"tools"`
	SuccessCriteria []string               `yaml:"success_criteria"`
	Metadata        map[string]interface{} `yaml:"metadata"` // Changed from map[string]string to support arrays
}

func NewAgentServer(rootDir string, maxConcurrent int, maxTokens int, model string, cfg *config.APIConfig) (*AgentServer, error) {
	// Get authentication credentials
	// ANTHROPIC_API_TOKEN = Bearer token (for corporate proxies)
	// ANTHROPIC_API_KEY = API key (standard x-api-key header)
	apiKey, isBearerToken, err := auth.GetAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	// Build client options
	var clientOpts []option.RequestOption

	if isBearerToken {
		// For Bearer tokens, we need to use a custom HTTP client that sets Authorization header
		clientOpts = append(clientOpts, option.WithHTTPClient(proxy.NewBearerTokenClient(apiKey, cfg)))
		monitoring.Logger.Info("api_configuration", "auth_type", "bearer_token")
	} else {
		// Standard API key - let SDK set x-api-key header
		clientOpts = append(clientOpts, option.WithAPIKey(apiKey))

		// Add custom HTTP client for proxy support (URL rewriting only)
		if httpClient := proxy.NewHTTPClient(cfg); httpClient != nil {
			clientOpts = append(clientOpts, option.WithHTTPClient(httpClient))
		}
		monitoring.Logger.Info("api_configuration", "auth_type", "api_key")
	}

	// Log proxy mode
	monitoring.Logger.Info("api_configuration", "proxy_mode", proxy.LogProxyMode(cfg))

	client := anthropic.NewClient(clientOpts...)

	server := &AgentServer{
		rootDir:       rootDir,
		anthropicKey:  apiKey,
		client:        client,
		beadsClient:   beads.NewClient(),
		maxConcurrent: maxConcurrent,
		maxTokens:     maxTokens,
		model:         model,
		activeTasks:   make(map[string]*TaskExecution),
		taskQueue:     make(chan *TaskExecution, 100),
		workerPool:    make(chan struct{}, maxConcurrent),
	}

	// Start worker pool
	go server.startWorkerPool()

	// Log Beads availability
	if beads.IsInstalled() {
		monitoring.Logger.Info("beads_available", "installed", true)
	} else {
		monitoring.Logger.Warn("beads_not_installed", "message", "Install Beads for better task tracking: https://github.com/steveyegge/beads")
	}

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

func (s *AgentServer) spawnAgentTask(role, taskInput string) (*protocol.ExecuteTaskResponse, error) {
	// Validate and get task description
	var beadsTaskID string
	var isBeadsTask bool

	// Check if it looks like a Beads task ID
	if beads.IsBeadsTaskID(taskInput) {
		// Validate the task exists in Beads
		if err := s.beadsClient.ValidateTaskID(taskInput); err != nil {
			return nil, fmt.Errorf("invalid Beads task: %w", err)
		}

		beadsTaskID = taskInput
		isBeadsTask = true
		monitoring.Logger.Info("spawning_with_beads_task", "task_id", beadsTaskID)

		// Check dependencies
		if beads.IsInstalled() {
			depsOK, unmetDeps, err := s.beadsClient.CheckDependencies(beadsTaskID)
			if err != nil {
				monitoring.Logger.Warn("dependency_check_failed", "error", err.Error())
			} else if !depsOK {
				return nil, fmt.Errorf("task %s has unmet dependencies: %v\nPlease complete these tasks first", beadsTaskID, unmetDeps)
			}
		}
	} else {
		monitoring.Logger.Info("spawning_with_free_form_description", "description", taskInput)
	}

	// Get task description, task packet path, and working directory (from Beads or use free-form input)
	taskDescription, taskPacketPath, workingDir, _, err := s.beadsClient.GetTaskDescription(taskInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get task description: %w", err)
	}

	// If working directory not specified in task, use server's root directory
	if workingDir == "" {
		workingDir = s.rootDir
	}

	// Generate internal task ID
	taskID := fmt.Sprintf("task-%s-%s", role, time.Now().Format("20060102-150405-000000"))

	// Load agent configuration
	config, err := s.loadAgentConfig(role)
	if err != nil {
		return nil, fmt.Errorf("failed to load agent config: %w", err)
	}

	// Create task packet
	if err := s.createTaskPacket(taskID, role, taskDescription, config); err != nil {
		return nil, fmt.Errorf("failed to create task packet: %w", err)
	}

	// Store task packet path and working directory in metadata if available
	metadata := map[string]string{}
	if beadsTaskID != "" {
		metadata["beads_task_id"] = beadsTaskID
	}
	if taskPacketPath != "" {
		metadata["task_packet_path"] = taskPacketPath
	}
	if workingDir != "" {
		metadata["working_directory"] = workingDir
	}

	// If Beads task, mark as started
	if isBeadsTask && beads.IsInstalled() {
		if err := s.beadsClient.StartTask(beadsTaskID); err != nil {
			monitoring.Logger.Warn("failed_to_start_beads_task", "error", err.Error())
		} else {
			monitoring.Logger.Info("beads_task_started", "task_id", beadsTaskID)
		}
	}

	// Create task execution
	execution := &TaskExecution{
		TaskID:     taskID,
		Role:       role,
		Task:       taskDescription,
		Config:     config,
		StartTime:  time.Now(),
		Status:     "queued",
		Progress:   0.0,
		streamChan: make(chan *protocol.StreamEvent, 100),
		streamOpen: true,
		metadata:   metadata,
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
	// Try paths in order:
	// 1. Project override (.ai/agents/)
	// 2. Framework (.ai-pack/agents/) - production
	// 3. Development (../agents/) - when running from a2a-agent dir
	// 4. Development (agents/) - when running from repo root
	candidatePaths := []struct {
		path   string
		source string
	}{
		{filepath.Join(s.rootDir, ".ai", "agents", role+".yml"), "project_override"},
		{filepath.Join(s.rootDir, ".ai-pack", "agents", role+".yml"), "framework"},
		{filepath.Join(s.rootDir, "..", "agents", role+".yml"), "dev_parent"},
		{filepath.Join(s.rootDir, "agents", role+".yml"), "dev_root"},
	}

	var configPath string
	var source string

	for _, candidate := range candidatePaths {
		if _, err := os.Stat(candidate.path); err == nil {
			configPath = candidate.path
			source = candidate.source
			break
		}
	}

	if configPath == "" {
		return nil, fmt.Errorf("no config found for role %s (tried: .ai/agents, .ai-pack/agents, ../agents, agents)", role)
	}

	monitoring.Logger.Info("loading_agent_config", "role", role, "source", source, "path", configPath)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config AgentConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Normalize role_file path based on config source
	// Role file paths in configs are relative to the ai-pack root
	if source == "dev_parent" {
		// Config loaded from ../agents/, role files are in ../roles/
		if !filepath.IsAbs(config.Context.RoleFile) && !strings.HasPrefix(config.Context.RoleFile, ".ai") {
			config.Context.RoleFile = filepath.Join("..", config.Context.RoleFile)
		}
	} else if source == "dev_root" {
		// Config loaded from agents/, role files are in roles/
		// Path is already correct as-is
	} else if source == "framework" {
		// Config loaded from .ai-pack/agents/, role files are in .ai-pack/roles/
		if !filepath.IsAbs(config.Context.RoleFile) && !strings.HasPrefix(config.Context.RoleFile, ".ai") {
			config.Context.RoleFile = filepath.Join(".ai-pack", config.Context.RoleFile)
		}
	}

	return &config, nil
}

func (s *AgentServer) loadRoleContext(roleFile string) (string, error) {
	// Support override pattern: try .ai/ first, then .ai-pack/
	var rolePath string

	// If roleFile starts with .ai/, try project override first
	if strings.HasPrefix(roleFile, ".ai/") {
		projectPath := filepath.Join(s.rootDir, roleFile)
		if _, err := os.Stat(projectPath); err == nil {
			rolePath = projectPath
		} else {
			// Fallback to framework path
			frameworkPath := strings.Replace(roleFile, ".ai/", ".ai-pack/", 1)
			rolePath = filepath.Join(s.rootDir, frameworkPath)
		}
	} else {
		// Direct path specified
		rolePath = filepath.Join(s.rootDir, roleFile)
	}

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

func (s *AgentServer) buildPrompt(role, task, roleContext string, config *AgentConfig, taskPacketPath, workingDir string) string {
	prompt := fmt.Sprintf(`You are a %s agent.

**Your Role Context:**
%s

**Your Task:**
%s

**Working Directory:**
%s

All file operations (Read, Write, Edit, Glob, Grep, Bash) must be performed relative to the working directory above.`,
		role,
		roleContext,
		task,
		workingDir)

	// Add task packet location if available (the role file defines how to use it)
	if taskPacketPath != "" {
		prompt += fmt.Sprintf(`

**Task Packet Location:**
%s

The task packet contains:
- 00-contract.md - Requirements and acceptance criteria
- 10-plan.md - Implementation plan
- 20-work-log.md - Progress tracking
- 30-review.md - Review notes
- 40-acceptance.md - Completion checklist

Your role definition specifies how to use the task packet.`,
			taskPacketPath)
	}

	prompt += fmt.Sprintf(`

**Configuration:**
- Timeout: %s
- Tools: %v
- Success Criteria: %v

Execute the task according to your role definition.`,
		config.Delegation.Timeout,
		config.Tools,
		config.SuccessCriteria,
	)

	return prompt
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

	// Get task packet path and working directory from metadata
	taskPacketPath := ""
	workingDir := s.rootDir // Default to server's root directory
	if execution.metadata != nil {
		if path := execution.metadata["task_packet_path"]; path != "" {
			taskPacketPath = path
		}
		if dir := execution.metadata["working_directory"]; dir != "" {
			workingDir = dir
		}
	}

	// Build prompt
	prompt := s.buildPrompt(execution.Role, execution.Task, roleContext, execution.Config, taskPacketPath, workingDir)

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
	monitoring.Logger.Info("api_call_start", "task_id", taskID, "model", s.model)

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
	monitoring.LogAPICall(ctx, taskID, s.model, int(message.Usage.InputTokens+message.Usage.OutputTokens))

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
	beadsTaskID := ""
	if execution.metadata != nil {
		beadsTaskID = execution.metadata["beads_task_id"]
	}
	s.mu.Unlock()

	// If this was a Beads task, mark it complete in Beads
	if beadsTaskID != "" && beads.IsInstalled() {
		if err := s.beadsClient.CompleteTask(beadsTaskID); err != nil {
			monitoring.Logger.Warn("failed_to_complete_beads_task", "task_id", beadsTaskID, "error", err.Error())
		} else {
			monitoring.Logger.Info("beads_task_completed", "task_id", beadsTaskID)
		}
	}

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
