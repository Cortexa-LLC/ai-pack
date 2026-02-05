package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/auth"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/beads"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/claude"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/config"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/execution_log"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/protocol"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/proxy"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/tools"
	"gopkg.in/yaml.v3"
)

const (
	Version = "2.1.0"

	// Token limits and thresholds
	MaxContextTokens      = 200000 // Claude API limit
	SafeTokenThreshold    = 180000 // Start aggressive truncation at 90%
	EmergencyThreshold    = 190000 // Emergency summarization at 95%
	TokensPerChar         = 0.25   // Rough estimate: 4 chars per token

	// Content block types
	ContentTypeText    = "text"
	ContentTypeToolUse = "tool_use"

	// Message event types
	MessageEventStart       = "message_start"
	MessageEventDelta       = "message_delta"
	MessageEventStop        = "message_stop"
	ContentBlockStart       = "content_block_start"
	ContentBlockDelta       = "content_block_delta"
	ContentBlockStop        = "content_block_stop"
	MessageDeltaUsage       = "message_delta"
	PingEvent               = "ping"
	ErrorEvent              = "error"

	// File names
	MetadataFileName        = "00-metadata.json"
	ProjectRegistryFileName = "project-registry.json"

	// Directory names
	BeadsDir = ".beads"

	// Project cleanup threshold (days)
	ProjectInactiveDays = 30
)

type AgentServer struct {
	rootDir          string
	anthropicKey     string
	client           anthropic.Client // SDK v1.19+ returns Client by value
	beadsClient      *beads.Client
	claudeSettings   *claude.Settings // Claude Code settings (deny patterns, etc.)
	executionLog     *execution_log.ExecutionLog // Persistent agent execution log
	maxConcurrent    int              // Maximum concurrent agents (configurable)
	maxTokens        int              // Maximum tokens per API call
	model            string           // Anthropic model to use
	maxInactiveTurns int              // Stop agent after N turns without progress
	config           *config.Config   // Server configuration

	// Concurrent execution tracking
	mu           sync.RWMutex
	activeTasks  map[string]*TaskExecution
	taskQueue    chan *TaskExecution
	workerPool   chan struct{}          // Semaphore for max concurrent agents
	projectRoots map[string]time.Time   // Registry of known project roots with last access time
}

type TaskExecution struct {
	TaskID      string
	Role        string
	Task        string
	Config      *AgentConfig
	StartTime   time.Time
	Status      string // "queued", "in_progress", "completed", "failed"
	Result      string
	Error       string
	ProjectRoot string // Project root directory where task metadata is stored

	// Beads integration
	metadata map[string]string

	// Streaming
	streamChan chan *protocol.StreamEvent
	streamOpen bool

	// Cancellation
	cancel context.CancelFunc
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
	Tools            []string               `yaml:"tools"`
	SuccessCriteria  []string               `yaml:"success_criteria"`
	Metadata         map[string]interface{} `yaml:"metadata"`          // Changed from map[string]string to support arrays
	ExtendedThinking bool                   `yaml:"extended_thinking"` // Enable extended thinking mode
	MaxTurns         int                    `yaml:"max_turns"`         // Deprecated: No turn limit, only max_inactive stops execution
}

// GetProjectRoots returns all known project roots
func (s *AgentServer) GetProjectRoots() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	roots := make([]string, 0, len(s.projectRoots)+1)
	roots = append(roots, s.rootDir) // Always include server root

	for root := range s.projectRoots {
		if root != s.rootDir {
			roots = append(roots, root)
		}
	}

	return roots
}

func NewAgentServer(rootDir string, maxConcurrent int, maxTokens int, model string, cfg *config.Config) (*AgentServer, error) {
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
		clientOpts = append(clientOpts, option.WithHTTPClient(proxy.NewBearerTokenClient(apiKey, &cfg.API)))
		monitoring.Logger.Info("api_configuration", "auth_type", "bearer_token")
	} else {
		// Standard API key - let SDK set x-api-key header
		clientOpts = append(clientOpts, option.WithAPIKey(apiKey))

		// Add custom HTTP client for proxy support (URL rewriting only)
		if httpClient := proxy.NewHTTPClient(&cfg.API); httpClient != nil {
			clientOpts = append(clientOpts, option.WithHTTPClient(httpClient))
		}
		monitoring.Logger.Info("api_configuration", "auth_type", "api_key")
	}

	// Set request timeout to avoid streaming enforcement
	// SDK requires streaming if calculated timeout > 10 minutes
	// By setting explicit timeout, we bypass that check
	if cfg.API.TimeoutSeconds > 0 {
		timeout := time.Duration(cfg.API.TimeoutSeconds) * time.Second
		clientOpts = append(clientOpts, option.WithRequestTimeout(timeout))
		monitoring.Logger.Info("api_configuration", "request_timeout", fmt.Sprintf("%v", timeout))
	}

	// Log proxy mode
	monitoring.Logger.Info("api_configuration", "proxy_mode", proxy.LogProxyMode(&cfg.API))

	client := anthropic.NewClient(clientOpts...)

	// Load Claude Code settings (global + project-specific)
	claudeSettings, err := claude.LoadSettings(rootDir)
	if err != nil {
		monitoring.Logger.Warn("failed_to_load_claude_settings", "error", err.Error())
		claudeSettings = &claude.Settings{} // Empty settings (allow all)
	} else {
		monitoring.Logger.Info("claude_settings_loaded",
			"deny_patterns", len(claudeSettings.Permissions.Deny),
			"ask_patterns", len(claudeSettings.Permissions.Ask))
	}

	// Get max inactive turns from config
	maxInactiveTurns := cfg.Agent.MaxInactiveTurns
	if maxInactiveTurns == 0 {
		maxInactiveTurns = 10 // Default fallback
	}

	// Initialize execution log
	execLog, err := execution_log.NewExecutionLog(rootDir)
	if err != nil {
		monitoring.Logger.Warn("failed_to_create_execution_log", "error", err.Error())
	}

	server := &AgentServer{
		rootDir:          rootDir,
		anthropicKey:     apiKey,
		client:           client,
		beadsClient:      beads.NewClient(),
		claudeSettings:   claudeSettings,
		executionLog:     execLog,
		maxConcurrent:    maxConcurrent,
		maxTokens:        maxTokens,
		model:            model,
		maxInactiveTurns: maxInactiveTurns,
		config:           cfg,
		activeTasks:      make(map[string]*TaskExecution),
		taskQueue:        make(chan *TaskExecution, 100),
		workerPool:       make(chan struct{}, maxConcurrent),
		projectRoots:     make(map[string]time.Time),
	}

	// Start worker pool
	go server.startWorkerPool()

	// Log Beads availability
	if beads.IsInstalled() {
		monitoring.Logger.Info("beads_available", "installed", true)
	} else {
		monitoring.Logger.Warn("beads_not_installed", "message", "Install Beads for better task tracking: https://github.com/steveyegge/beads")
	}

	// Load registered project roots from disk
	if err := server.loadProjectRoots(); err != nil {
		monitoring.Logger.Warn("failed_to_load_project_registry", "error", err.Error())
	}

	// Clean up inactive projects
	server.cleanupInactiveProjects()

	// Archive old completed tasks if enabled
	if cfg.TaskCleanup.Enabled && beads.IsInstalled() {
		server.archiveOldTasks()
	}

	// Handle orphaned in-progress beads tasks from previous server runs
	if beads.IsInstalled() {
		go server.handleOrphanedTasks()
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

func (s *AgentServer) spawnAgentTask(role, taskInput string, projectRoot string) (*protocol.ExecuteTaskResponse, error) {
	// Validate and get task description
	var beadsTaskID string
	var isBeadsTask bool

	// If no project root specified, use server's root directory
	if projectRoot == "" {
		projectRoot = s.rootDir
	}

	// Check if it looks like a Beads task ID
	if beads.IsBeadsTaskID(taskInput) {
		// Validate the task exists in Beads (use project root for bd commands)
		if err := s.beadsClient.ValidateTaskIDFromDir(taskInput, projectRoot); err != nil {
			return nil, fmt.Errorf("invalid Beads task: %w", err)
		}

		beadsTaskID = taskInput
		isBeadsTask = true
		monitoring.Logger.Info("spawning_with_beads_task", "task_id", beadsTaskID, "project_root", projectRoot)

		// Check dependencies
		if beads.IsInstalled() {
			depsOK, unmetDeps, err := s.beadsClient.CheckDependenciesFromDir(beadsTaskID, projectRoot)
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
	taskDescription, taskPacketPath, workingDir, _, err := s.beadsClient.GetTaskDescriptionFromDir(taskInput, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get task description: %w", err)
	}

	// If working directory not specified in task, use project root
	if workingDir == "" {
		workingDir = projectRoot
	}

	// Use Beads task ID as primary ID when available (Beads is source of truth)
	// Only generate internal task ID for non-Beads tasks
	var taskID string
	if isBeadsTask {
		taskID = beadsTaskID
	} else {
		taskID = fmt.Sprintf("task-%s-%s", role, time.Now().Format("20060102-150405-000000"))
	}

	// Load agent configuration
	config, err := s.loadAgentConfig(role)
	if err != nil {
		return nil, fmt.Errorf("failed to load agent config: %w", err)
	}

	// Create task packet in project's .beads/tasks/ directory
	if err := s.createTaskPacketInProject(taskID, role, taskDescription, config, projectRoot); err != nil {
		return nil, fmt.Errorf("failed to create task packet: %w", err)
	}

	// Store task packet path, working directory, and project root in metadata if available
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
	if projectRoot != "" {
		metadata["project_root"] = projectRoot
	}

	// If Beads task, mark as started
	if isBeadsTask && beads.IsInstalled() {
		if err := s.beadsClient.StartTaskFromDir(beadsTaskID, projectRoot); err != nil {
			monitoring.Logger.Warn("failed_to_start_beads_task", "error", err.Error())
		} else {
			monitoring.Logger.Info("beads_task_started", "task_id", beadsTaskID)
		}
	}

	// Create task execution
	execution := &TaskExecution{
		TaskID:      taskID,
		Role:        role,
		Task:        taskDescription,
		Config:      config,
		StartTime:   time.Now(),
		Status:      "queued",
		ProjectRoot: projectRoot,
		streamChan:  make(chan *protocol.StreamEvent, 100),
		streamOpen:  true,
		metadata:    metadata,
	}

	// Update task packet metadata with Beads task ID and project root
	if err := s.updateTaskPacketMetadataInProject(taskID, metadata, projectRoot); err != nil {
		monitoring.Logger.Warn("failed_to_update_task_metadata", "error", err.Error())
	}

	// Register task and project root
	s.mu.Lock()
	s.activeTasks[taskID] = execution
	if projectRoot != "" && projectRoot != s.rootDir {
		s.projectRoots[projectRoot] = time.Now()
		s.mu.Unlock()
		// Persist project registry
		if err := s.saveProjectRoots(); err != nil {
			monitoring.Logger.Warn("failed_to_save_project_registry", "error", err.Error())
		}
	} else {
		s.mu.Unlock()
	}

	// Queue for execution
	s.taskQueue <- execution

	// Log spawned event
	if s.executionLog != nil {
		metadataMap := make(map[string]interface{})
		for k, v := range metadata {
			metadataMap[k] = v
		}
		if err := s.executionLog.LogSpawned(taskID, role, taskDescription, metadataMap); err != nil {
			monitoring.Logger.Warn("failed_to_log_spawned_event", "error", err.Error())
		}
	}

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
	return s.createTaskPacketInProject(taskID, role, task, config, s.rootDir)
}

func (s *AgentServer) createTaskPacketInProject(taskID, role, task string, config *AgentConfig, projectRoot string) error {
	taskDir := filepath.Join(projectRoot, BeadsDir, "tasks", taskID)

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

	if err := os.WriteFile(filepath.Join(taskDir, MetadataFileName), metadataJSON, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	// Create plan file
	planContent := fmt.Sprintf("# Task Plan: %s\n\n**Role**: %s\n**Task**: %s\n**Created**: %s\n\n## Execution Plan\n\n(Agent will populate during execution)\n",
		taskID, role, task, time.Now().Format(time.RFC3339))

	if err := os.WriteFile(filepath.Join(taskDir, "10-plan.md"), []byte(planContent), 0644); err != nil {
		return fmt.Errorf("failed to write plan: %w", err)
	}

	monitoring.Logger.Info("task_packet_created", "task_id", taskID, "path", taskDir, "project", projectRoot)
	return nil
}

// updateTaskPacketMetadata updates the task metadata with Beads task ID and other runtime metadata
func (s *AgentServer) updateTaskPacketMetadata(taskID string, runtimeMetadata map[string]string) error {
	return s.updateTaskPacketMetadataInProject(taskID, runtimeMetadata, s.rootDir)
}

// updateTaskPacketMetadataInProject updates the task metadata in the project's directory
func (s *AgentServer) updateTaskPacketMetadataInProject(taskID string, runtimeMetadata map[string]string, projectRoot string) error {
	metadataPath := filepath.Join(projectRoot, BeadsDir, "tasks", taskID, MetadataFileName)

	// Read existing metadata
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to read metadata: %w", err)
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		return fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Add runtime metadata
	metadata["metadata"] = runtimeMetadata

	// Write updated metadata
	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, metadataJSON, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

func (s *AgentServer) loadTaskStatusFromDisk(taskID string) (*protocol.TaskStatusResponse, error) {
	metadataPath := filepath.Join(s.rootDir, BeadsDir, "tasks", taskID, MetadataFileName)

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
	resultsPath := filepath.Join(s.rootDir, BeadsDir, "tasks", taskID, "30-results.md")
	if resultData, err := os.ReadFile(resultsPath); err == nil {
		result = string(resultData)
	}

	createdAt, _ := time.Parse(time.RFC3339, metadata["spawned_at"].(string))
	updatedAt, _ := time.Parse(time.RFC3339, metadata["updated_at"].(string))

	// Extract status with fallback for older tasks
	status := "unknown"
	if s, ok := metadata["status"].(string); ok {
		status = s
	}

	// Extract role with fallback
	role := "unknown"
	if r, ok := metadata["role"].(string); ok {
		role = r
	}

	// Extract description with fallback
	description := ""
	if d, ok := metadata["description"].(string); ok {
		description = d
	}

	response := &protocol.TaskStatusResponse{
		TaskID:    taskID,
		Role:      role,
		Task:      description,
		Status:    status,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Result:    result,
		Metadata:  metadata,
	}

	if errorMsg, ok := metadata["error"].(string); ok {
		response.Error = errorMsg
	}

	return response, nil
}

func (s *AgentServer) buildPrompt(role, task, roleContext string, config *AgentConfig, taskPacketPath, workingDir string) string {
	// Note: roleContext is now passed separately as a system message for prompt caching
	// This prompt only contains the task-specific information
	prompt := fmt.Sprintf(`You are a %s agent.

**Your Task:**
%s

**Working Directory:**
%s

All file operations (Read, Write, Edit, Glob, Grep, Bash) must be performed relative to the working directory above.`,
		role,
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

func (s *AgentServer) updateTaskStatus(taskID, projectRoot, status, errorMsg string) {
	metadataPath := filepath.Join(projectRoot, BeadsDir, "tasks", taskID, MetadataFileName)

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

// buildSystemPrompt creates system messages with prompt caching for role context
func (s *AgentServer) buildSystemPrompt(roleContext string) []anthropic.TextBlockParam {
	// Cache the role context with 5-minute TTL (auto-refreshed on each turn)
	// This provides massive speedup across agentic loop turns while automatically
	// expiring between tasks
	return []anthropic.TextBlockParam{
		{
			Text:         roleContext,
			Type:         ContentTypeText,
			CacheControl: anthropic.NewCacheControlEphemeralParam(),
		},
	}
}

// convertToToolUnionParams converts []ToolParam to []ToolUnionParam for SDK v1.19+
func convertToToolUnionParams(tools []anthropic.ToolParam) []anthropic.ToolUnionParam {
	result := make([]anthropic.ToolUnionParam, len(tools))
	for i, tool := range tools {
		result[i] = anthropic.ToolUnionParam{
			OfTool: &tool,
		}
	}
	return result
}

// estimateTokenCount provides a rough estimate of token count for context management
// Uses character-based estimation (roughly 4 chars per token for English text)
func estimateTokenCount(text string) int {
	return int(float64(len(text)) * TokensPerChar)
}

// estimateMessagesTokenCount estimates total tokens for a slice of messages
func estimateMessagesTokenCount(messages []anthropic.MessageParam, systemPrompt []anthropic.TextBlockParam) int {
	total := 0

	// Count system prompt
	for _, block := range systemPrompt {
		total += estimateTokenCount(block.Text)
	}

	// Count messages
	for _, msg := range messages {
		// Estimate based on message content
		// Note: This is a rough estimate; actual tokens may vary
		msgJSON, _ := json.Marshal(msg)
		total += estimateTokenCount(string(msgJSON))
	}

	return total
}

// executeAgenticLoop runs the agentic loop with tool support
func (s *AgentServer) executeAgenticLoop(ctx context.Context, taskID string, initialPrompt string, roleContext string, workingDir string, config *AgentConfig, logMsg func(string)) (string, error) {
	// Define tools (matches Claude Code's exact tool names)
	toolDefs := tools.DefineTools()

	// Build system prompt with caching for role context
	systemPrompt := s.buildSystemPrompt(roleContext)

	// Build messages starting with user prompt (task-specific)
	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(initialPrompt)),
	}

	// Estimate initial context size and warn if large
	initialTokens := estimateMessagesTokenCount(messages, systemPrompt)
	logMsg(fmt.Sprintf("🔄 Starting agentic loop (max_inactive: %d, caching: enabled, extended_thinking: %v)",
		s.maxInactiveTurns, config.ExtendedThinking))
	logMsg(fmt.Sprintf("   Initial context: ~%d tokens (limit: %d)", initialTokens, MaxContextTokens))

	if initialTokens > SafeTokenThreshold {
		logMsg(fmt.Sprintf("   ⚠️  WARNING: Initial context is large (%d tokens > %d threshold)",
			initialTokens, SafeTokenThreshold))
		logMsg("   💡 Consider using a more concise role context or task description")
	}

	var finalResult strings.Builder
	totalInputTokens := int64(0)
	totalOutputTokens := int64(0)

	// Progress tracking for turn counter reset
	turn := 1
	inactiveTurns := 0
	lastTextLength := 0
	lastToolSignature := "" // Tracks tool names + input hash for better progress detection

	for {
		logMsg(fmt.Sprintf("   Turn %d (inactive: %d)...", turn, inactiveTurns))

		// Estimate current token count to prevent API failures
		estimatedTokens := estimateMessagesTokenCount(messages, systemPrompt)

		// Determine truncation strategy based on token count
		var maxHistoryTurns int
		const messagesPerTurn = 2 // Each turn: assistant message + user tool results

		if estimatedTokens > EmergencyThreshold {
			// EMERGENCY: Very aggressive truncation - keep only last 5 turns
			maxHistoryTurns = 5
			logMsg(fmt.Sprintf("      ⚠️  EMERGENCY: Estimated %d tokens (>%d) - aggressive truncation to last %d turns",
				estimatedTokens, EmergencyThreshold, maxHistoryTurns))
		} else if estimatedTokens > SafeTokenThreshold {
			// WARNING: Moderate truncation - keep last 10 turns
			maxHistoryTurns = 10
			logMsg(fmt.Sprintf("      ⚠️  WARNING: Estimated %d tokens (>%d) - reducing history to last %d turns",
				estimatedTokens, SafeTokenThreshold, maxHistoryTurns))
		} else {
			// NORMAL: Standard truncation - keep last 25 turns
			maxHistoryTurns = 25
		}

		// Truncate conversation history to prevent token overflow
		// Keep first message (initial prompt) + last N turns based on token count
		truncatedMessages := messages
		if len(messages) > 1+maxHistoryTurns*messagesPerTurn {
			firstMsg := messages[0] // Keep initial prompt with task context
			recentMsgs := messages[len(messages)-(maxHistoryTurns*messagesPerTurn):]
			truncatedMessages = append([]anthropic.MessageParam{firstMsg}, recentMsgs...)

			// Log truncation for visibility
			if turn%10 == 0 || estimatedTokens > SafeTokenThreshold { // Log more frequently when near limit
				logMsg(fmt.Sprintf("      📉 Truncated history: %d → %d messages (keeping last %d turns, est. %d tokens)",
					len(messages), len(truncatedMessages), maxHistoryTurns, estimatedTokens))
			}
		}

		// Final safety check: If still over limit after truncation, fail fast with clear error
		finalEstimate := estimateMessagesTokenCount(truncatedMessages, systemPrompt)
		if finalEstimate > MaxContextTokens {
			errMsg := fmt.Sprintf("context too large: estimated %d tokens exceeds %d token limit even after aggressive truncation. This task requires more summarization or breaking into smaller subtasks.",
				finalEstimate, MaxContextTokens)
			logMsg(fmt.Sprintf("❌ %s", errMsg))
			return "", fmt.Errorf(errMsg)
		}

		// Prepare API params with system prompt (SDK v1.19+ uses direct values, not F() wrappers)
		params := anthropic.MessageNewParams{
			Model:     anthropic.Model(s.model),
			MaxTokens: int64(s.maxTokens),
			Messages:  truncatedMessages, // Use truncated history to reduce input tokens
			Tools:     convertToToolUnionParams(toolDefs),
			System:    systemPrompt,
		}

		// Note: Extended thinking support requires newer SDK version
		// TODO: Add when SDK updated
		if config.ExtendedThinking {
			logMsg("      ⚠️  Extended thinking requested but not yet supported in SDK")
		}

		// Make API call with streaming
		apiStart := time.Now()

		stream := s.client.Messages.NewStreaming(ctx, params)

		// Use SDK's Accumulate method to build the message
		var message anthropic.Message
		eventCount := 0
		for stream.Next() {
			event := stream.Current()
			eventCount++
			if err := message.Accumulate(event); err != nil {
				// Check if we hit max_tokens limit
				stopReason := string(message.StopReason)
				if stopReason == "max_tokens" {
					monitoring.Logger.Warn("max_tokens_limit_reached",
						"task_id", taskID,
						"turn", turn,
						"output_tokens", int(message.Usage.OutputTokens),
					)
					logMsg(fmt.Sprintf("      ⚠️  Max tokens reached (%d). Completing turn.", message.Usage.OutputTokens))
					break
				}

				// Log stream error with essential context
				monitoring.Logger.Error("stream_error",
					"task_id", taskID,
					"turn", turn,
					"event", eventCount,
					"stop_reason", stopReason,
					"error", err.Error(),
				)

				return "", fmt.Errorf("stream error on turn %d: %w", turn, err)
			}
		}

		if err := stream.Err(); err != nil {
			errMsg := err.Error()

			// Check for token limit errors and provide actionable guidance
			if strings.Contains(errMsg, "prompt is too long") || strings.Contains(errMsg, "maximum") {
				logMsg(fmt.Sprintf("❌ Context size exceeded API limit on turn %d", turn))
				logMsg(fmt.Sprintf("   Estimated tokens before call: %d", finalEstimate))
				logMsg(fmt.Sprintf("   💡 Recommendation: Break this task into smaller subtasks"))

				return "", fmt.Errorf("API token limit exceeded on turn %d (estimated %d tokens). "+
					"This task is too complex for a single agent execution. "+
					"Please break it into smaller subtasks: %w", turn, finalEstimate, err)
			}

			return "", fmt.Errorf("API call failed on turn %d: %w", turn, err)
		}

		apiDuration := time.Since(apiStart).Milliseconds()
		totalInputTokens += message.Usage.InputTokens
		totalOutputTokens += message.Usage.OutputTokens

		// Log token usage with percentage of limit
		tokenUsagePercent := float64(totalInputTokens+totalOutputTokens) / float64(MaxContextTokens) * 100
		logMsg(fmt.Sprintf("      API: %dms | in:%d out:%d | total:%d (%.1f%% of limit)",
			apiDuration, message.Usage.InputTokens, message.Usage.OutputTokens,
			totalInputTokens+totalOutputTokens, tokenUsagePercent))

		// Warn if cumulative usage is high
		if totalInputTokens+totalOutputTokens > SafeTokenThreshold && turn%5 == 0 {
			logMsg(fmt.Sprintf("      ⚠️  High cumulative token usage: %d tokens (%.1f%% of limit)",
				totalInputTokens+totalOutputTokens, tokenUsagePercent))
		}

		// Record per-turn token metrics
		monitoring.GlobalMetrics.RecordTurnTokens(taskID, turn, int64(message.Usage.InputTokens), int64(message.Usage.OutputTokens), apiDuration)

		// Process response blocks
		var toolUses []anthropic.ToolUseBlock
		hasText := false

		for _, block := range message.Content {
			switch block.Type {
			case ContentTypeText:
				finalResult.WriteString(block.Text)
				finalResult.WriteString("\n")
				hasText = true
				logMsg(fmt.Sprintf("      💬 Text: %d chars", len(block.Text)))

			case ContentTypeToolUse:
				// Access fields directly from union instead of using AsToolUse()
				// since we manually constructed these blocks without JSON.raw
				toolUse := anthropic.ToolUseBlock{
					ID:    block.ID,
					Name:  block.Name,
					Input: block.Input,
				}
				toolUses = append(toolUses, toolUse)
				logMsg(fmt.Sprintf("      🔧 Tool: %s", toolUse.Name))
			}
		}

		// If no tool uses and we have text, we're done
		if len(toolUses) == 0 {
			if hasText {
				logMsg(fmt.Sprintf("✅ Agent completed in %d turns", turn))
				logMsg(fmt.Sprintf("   Total tokens: %d (in:%d out:%d)", totalInputTokens+totalOutputTokens, totalInputTokens, totalOutputTokens))
				break
			}
			return "", fmt.Errorf("no output from agent on turn %d", turn)
		}

		// Execute tools and build tool results
		var toolResultBlocks []anthropic.ContentBlockParamUnion
		for _, toolUse := range toolUses {
			// Parse tool input from JSON
			var inputMap map[string]interface{}
			if err := json.Unmarshal(toolUse.Input, &inputMap); err != nil {
				logMsg(fmt.Sprintf("         ❌ Failed to parse tool input for %s: %v", toolUse.Name, err))
				continue
			}

			// Execute tool with Claude settings
			result, err := tools.ExecuteTool(toolUse.Name, inputMap, workingDir, s.claudeSettings)
			if err != nil {
				logMsg(fmt.Sprintf("         ❌ Tool execution failed: %v", err))
				result = fmt.Sprintf("Error: %v", err)
			} else {
				// Truncate long results for logging
				displayResult := result
				if len(displayResult) > 100 {
					displayResult = displayResult[:100] + "..."
				}
				logMsg(fmt.Sprintf("         ✓ %s", displayResult))
			}

			// Add tool result to blocks
			toolResultBlocks = append(toolResultBlocks, anthropic.NewToolResultBlock(toolUse.ID, result, false))
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

		// Build tool signature including input details for better progress detection
		var toolSignatures []string
		for _, toolUse := range toolUses {
			// Create a signature that includes tool name and key input parameters
			var inputMap map[string]interface{}
			if err := json.Unmarshal(toolUse.Input, &inputMap); err == nil {
				// Extract a simple signature from the input (first 100 chars of JSON)
				inputJSON := string(toolUse.Input)
				if len(inputJSON) > 100 {
					inputJSON = inputJSON[:100]
				}
				toolSignatures = append(toolSignatures, fmt.Sprintf("%s:%s", toolUse.Name, inputJSON))
			} else {
				toolSignatures = append(toolSignatures, toolUse.Name)
			}
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

		// Add assistant message with tool uses (SDK v1.19+ uses Type field)
		var assistantContent []anthropic.ContentBlockParamUnion
		for _, block := range message.Content {
			switch block.Type {
			case ContentTypeText:
				textBlock := block.AsText()
				assistantContent = append(assistantContent, anthropic.NewTextBlock(textBlock.Text))
			case ContentTypeToolUse:
				toolBlock := block.AsToolUse()
				assistantContent = append(assistantContent, anthropic.NewToolUseBlock(toolBlock.ID, toolBlock.Input, toolBlock.Name))
			}
		}
		messages = append(messages, anthropic.NewAssistantMessage(assistantContent...))

		// Add tool results as user message
		messages = append(messages, anthropic.NewUserMessage(toolResultBlocks...))

		// Increment turn counter
		turn++
	}

	monitoring.GlobalMetrics.IncrementAPICallsSuccess()
	monitoring.LogAPICall(ctx, taskID, s.model, int(totalInputTokens+totalOutputTokens))

	// Record token usage for this session
	monitoring.GlobalMetrics.RecordTokenUsage(taskID, totalInputTokens, totalOutputTokens, int64(turn-1))

	return finalResult.String(), nil
}

func (s *AgentServer) executeAgentTask(execution *TaskExecution) {
	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

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

	// Initialize task execution
	s.initializeTaskExecution(execution, logMsg)

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

// setupExecutionLogger creates the execution log file and returns a logging function
func (s *AgentServer) setupExecutionLogger(execution *TaskExecution) func(string) {
	logPath := filepath.Join(execution.ProjectRoot, BeadsDir, "tasks", execution.TaskID, "execution.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		monitoring.Logger.Error("log_create_error", "task_id", execution.TaskID, "error", err)
	}

	if logFile != nil {
		// Use runtime.SetFinalizer to ensure closure, but also provide explicit close
		runtime.SetFinalizer(logFile, func(f *os.File) { f.Close() })
	}

	return func(msg string) {
		timestamp := time.Now().Format("15:04:05")
		line := fmt.Sprintf("[%s] %s\n", timestamp, msg)
		if logFile != nil {
			if _, err := logFile.WriteString(line); err != nil {
				monitoring.Logger.Error("log_write_error", "task_id", execution.TaskID, "error", err)
			}
			if err := logFile.Sync(); err != nil {
				monitoring.Logger.Error("log_sync_error", "task_id", execution.TaskID, "error", err)
			}
		}
		monitoring.Logger.Info("agent_log", "task_id", execution.TaskID, "message", msg)
	}
}

// initializeTaskExecution sets initial status and progress
func (s *AgentServer) initializeTaskExecution(execution *TaskExecution, logMsg func(string)) {
	s.mu.Lock()
	execution.Status = "in_progress"
	s.mu.Unlock()

	s.updateTaskStatus(execution.TaskID, execution.ProjectRoot, "in_progress", "")
	logMsg("📝 Status updated: in_progress")

	s.sendStreamEvent(execution, "status_update", map[string]interface{}{
		"status": "in_progress",
	})

	// Log started event
	if s.executionLog != nil {
		if err := s.executionLog.LogStarted(execution.TaskID); err != nil {
			monitoring.Logger.Warn("failed_to_log_started_event", "error", err.Error())
		}
	}
}

// loadAndLogRoleContext loads role context with logging and error handling
func (s *AgentServer) loadAndLogRoleContext(execution *TaskExecution, logMsg func(string)) (string, error) {
	logMsg(fmt.Sprintf("📖 Loading role context: %s", execution.Config.Context.RoleFile))
	roleContext, err := s.loadRoleContext(execution.Config.Context.RoleFile)
	if err != nil {
		monitoring.Logger.Error("role_context_load_error", "task_id", execution.TaskID, "error", err)
		logMsg(fmt.Sprintf("❌ Failed to load role context: %v", err))
		s.failTask(execution, fmt.Sprintf("Failed to load role context: %v", err))
		return "", err
	}
	logMsg("✅ Role context loaded")
	return roleContext, nil
}

// extractTaskMetadata extracts task packet path and working directory from metadata
func (s *AgentServer) extractTaskMetadata(execution *TaskExecution, logMsg func(string)) (string, string) {
	taskPacketPath := ""
	workingDir := s.rootDir

	if execution.metadata != nil {
		if path := execution.metadata["task_packet_path"]; path != "" {
			taskPacketPath = path
			logMsg(fmt.Sprintf("📦 Task packet: %s", taskPacketPath))
		}
		if dir := execution.metadata["working_directory"]; dir != "" {
			workingDir = dir
			logMsg(fmt.Sprintf("📂 Working directory: %s", workingDir))
		}
	}

	return taskPacketPath, workingDir
}

// buildAndSavePrompt builds the agent prompt and saves it to disk
func (s *AgentServer) buildAndSavePrompt(execution *TaskExecution, roleContext, taskPacketPath, workingDir string, logMsg func(string)) string {
	logMsg("🔨 Building agent prompt...")
	prompt := s.buildPrompt(execution.Role, execution.Task, roleContext, execution.Config, taskPacketPath, workingDir)
	logMsg(fmt.Sprintf("✅ Prompt built (%d chars)", len(prompt)))

	promptPath := filepath.Join(execution.ProjectRoot, BeadsDir, "tasks", execution.TaskID, "agent-prompt.txt")
	if err := os.WriteFile(promptPath, []byte(prompt), 0644); err != nil {
		monitoring.Logger.Warn("prompt_save_error", "task_id", execution.TaskID, "error", err)
	} else {
		logMsg(fmt.Sprintf("💾 Prompt saved: %s", promptPath))
	}

	return prompt
}

// executeAgentWorkflow runs the agentic loop with context support for cancellation.
// The context can be cancelled via CancelTask() or timeout via DeadlineExceeded.
// Cancellation flow: CancelTask() -> execution.cancel() -> ctx.Err() checked here
func (s *AgentServer) executeAgentWorkflow(ctx context.Context, execution *TaskExecution, prompt, roleContext, workingDir string, logMsg func(string)) (string, error) {
	s.sendStreamEvent(execution, "api_call_start", map[string]interface{}{})

	result, err := s.executeAgenticLoop(ctx, execution.TaskID, prompt, roleContext, workingDir, execution.Config, logMsg)
	if err != nil {
		// Check if error is due to cancellation or timeout.
		// context.Canceled: User-initiated cancellation via CancelTask()
		// context.DeadlineExceeded: Task exceeded configured timeout
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

// saveAndCompleteTask saves results, updates status, and marks task complete
func (s *AgentServer) saveAndCompleteTask(ctx context.Context, execution *TaskExecution, result string, startTime time.Time, logMsg func(string)) {
	// Save results
	s.saveTaskResults(execution, result, logMsg)

	// Update task status
	beadsTaskID, projectRoot := s.updateTaskCompletion(execution, result)

	// Complete Beads task if applicable
	if beadsTaskID != "" {
		s.completeBeadsTask(beadsTaskID, projectRoot, logMsg)
	}

	// Finalize task
	durationMs := time.Since(startTime).Milliseconds()
	s.updateTaskStatus(execution.TaskID, execution.ProjectRoot, "completed", "")
	s.sendStreamEvent(execution, "completed", map[string]interface{}{
		"result": result,
	})

	s.closeStream(execution)

	// Log completed event
	if s.executionLog != nil {
		resultSummary := result
		if len(resultSummary) > 500 {
			resultSummary = resultSummary[:500] + "..."
		}
		if err := s.executionLog.LogCompleted(execution.TaskID, durationMs, resultSummary); err != nil {
			monitoring.Logger.Warn("failed_to_log_completed_event", "error", err.Error())
		}
	}

	logMsg(fmt.Sprintf("🎉 Task completed successfully (duration: %dms)", durationMs))
	logMsg("=" + strings.Repeat("=", 70))

	monitoring.LogTaskCompleted(ctx, execution.TaskID, execution.Role, durationMs)
	monitoring.GlobalMetrics.IncrementTasksCompleted(durationMs)
}

// saveTaskResults saves the task results to disk
func (s *AgentServer) saveTaskResults(execution *TaskExecution, result string, logMsg func(string)) {
	logMsg("💾 Saving results...")
	resultsPath := filepath.Join(execution.ProjectRoot, BeadsDir, "tasks", execution.TaskID, "30-results.md")
	resultsContent := fmt.Sprintf("# Task Results: %s\n\n**Role**: %s\n**Task**: %s\n**Completed**: %s\n\n## Agent Output\n\n%s\n",
		execution.TaskID, execution.Role, execution.Task, time.Now().Format(time.RFC3339), result)

	if err := os.WriteFile(resultsPath, []byte(resultsContent), 0644); err != nil {
		monitoring.Logger.Warn("results_save_error", "task_id", execution.TaskID, "error", err)
		logMsg(fmt.Sprintf("⚠️  Failed to save results: %v", err))
	} else {
		logMsg(fmt.Sprintf("✅ Results saved: %s", resultsPath))
		logMsg(fmt.Sprintf("   Output length: %d chars", len(result)))
	}
}

// updateTaskCompletion updates execution status and extracts Beads task ID and project root
func (s *AgentServer) updateTaskCompletion(execution *TaskExecution, result string) (string, string) {
	s.mu.Lock()
	execution.Status = "completed"
	execution.Result = result
	beadsTaskID := ""
	projectRoot := ""
	if execution.metadata != nil {
		beadsTaskID = execution.metadata["beads_task_id"]
		projectRoot = execution.metadata["project_root"]
	}
	// Remove from active tasks map since task is now completed
	delete(s.activeTasks, execution.TaskID)
	s.mu.Unlock()

	return beadsTaskID, projectRoot
}

// completeBeadsTask marks the corresponding Beads task as complete
func (s *AgentServer) completeBeadsTask(beadsTaskID string, projectRoot string, logMsg func(string)) {
	if !beads.IsInstalled() {
		return
	}

	logMsg(fmt.Sprintf("🔗 Marking Beads task complete: %s", beadsTaskID))
	if err := s.beadsClient.CompleteTaskFromDir(beadsTaskID, projectRoot); err != nil {
		monitoring.Logger.Warn("failed_to_complete_beads_task", "task_id", beadsTaskID, "error", err.Error())
		logMsg(fmt.Sprintf("⚠️  Failed to complete Beads task: %v", err))
	} else {
		monitoring.Logger.Info("beads_task_completed", "task_id", beadsTaskID)
		logMsg("✅ Beads task marked complete")
	}
}

func (s *AgentServer) failTask(execution *TaskExecution, errorMsg string) {
	ctx := context.Background()
	durationMs := time.Since(execution.StartTime).Milliseconds()

	// Log failure to execution log
	logPath := filepath.Join(execution.ProjectRoot, BeadsDir, "tasks", execution.TaskID, "execution.log")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		timestamp := time.Now().Format("15:04:05")
		_, _ = logFile.WriteString(fmt.Sprintf("[%s] ❌ Task failed: %s\n", timestamp, errorMsg))
		_, _ = logFile.WriteString(fmt.Sprintf("[%s]    Duration: %dms\n", timestamp, durationMs))
		_, _ = logFile.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, strings.Repeat("=", 70)))
		_ = logFile.Close()
	}

	s.mu.Lock()
	execution.Status = "failed"
	execution.Error = errorMsg
	// Remove from active tasks map since task is now failed
	delete(s.activeTasks, execution.TaskID)
	s.mu.Unlock()

	s.updateTaskStatus(execution.TaskID, execution.ProjectRoot, "failed", errorMsg)
	s.sendStreamEvent(execution, "failed", map[string]interface{}{
		"error": errorMsg,
	})
	s.closeStream(execution)

	// Log failed event
	if s.executionLog != nil {
		if err := s.executionLog.LogFailed(execution.TaskID, durationMs, errorMsg); err != nil {
			monitoring.Logger.Warn("failed_to_log_failed_event", "error", err.Error())
		}
	}

	monitoring.LogTaskFailed(ctx, execution.TaskID, execution.Role, errorMsg, durationMs)
	monitoring.GlobalMetrics.IncrementTasksFailed(durationMs)
}

// CancelTask cancels a running task by calling its context cancel function.
// This triggers context cancellation which is detected in executeAgentWorkflow,
// causing the task to be marked as cancelled via cancelTaskExecution.
// Can be called from CLI, GUI, or GraphQL API.
func (s *AgentServer) CancelTask(taskID string) error {
	s.mu.Lock()
	execution, exists := s.activeTasks[taskID]
	s.mu.Unlock()

	if !exists {
		return fmt.Errorf("task not found or not active: %s", taskID)
	}

	// Call the cancel function to trigger context cancellation
	// This will cause ctx.Err() to return context.Canceled in the execution loop
	if execution.cancel != nil {
		execution.cancel()

		// Log cancellation to execution log
		logPath := filepath.Join(execution.ProjectRoot, BeadsDir, "tasks", execution.TaskID, "execution.log")
		logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			timestamp := time.Now().Format("15:04:05")
			_, _ = logFile.WriteString(fmt.Sprintf("[%s] 🛑 Task cancelled by user\n", timestamp))
			_, _ = logFile.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, strings.Repeat("=", 70)))
			_ = logFile.Close()
		}

		return nil
	}

	return fmt.Errorf("task cannot be cancelled: %s", taskID)
}

// cancelTaskExecution marks a task as cancelled (called after context cancellation)
// This is invoked when a task is cancelled via CancelTask() or when the context
// is cancelled/times out during execution.
func (s *AgentServer) cancelTaskExecution(execution *TaskExecution, message string) {
	ctx := context.Background()
	durationMs := time.Since(execution.StartTime).Milliseconds()

	s.mu.Lock()
	execution.Status = "cancelled"
	execution.Error = message
	// Remove from active tasks map since task is now cancelled
	delete(s.activeTasks, execution.TaskID)
	s.mu.Unlock()

	s.updateTaskStatus(execution.TaskID, execution.ProjectRoot, "cancelled", message)
	s.sendStreamEvent(execution, "cancelled", map[string]interface{}{
		"message": message,
	})
	s.closeStream(execution)

	// Log cancelled event
	if s.executionLog != nil {
		if err := s.executionLog.LogCancelled(execution.TaskID, durationMs); err != nil {
			monitoring.Logger.Warn("failed_to_log_cancelled_event", "error", err.Error())
		}
	}

	monitoring.LogTaskFailed(ctx, execution.TaskID, execution.Role, message, durationMs)
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

	// Send to channel for live streaming
	select {
	case execution.streamChan <- event:
	default:
		// Channel full, skip event (but still write to file)
	}

	// Also write to per-task log file for historical access
	s.writeStreamEventToFile(execution, event)
}

// writeStreamEventToFile appends a stream event to the per-task log file
func (s *AgentServer) writeStreamEventToFile(execution *TaskExecution, event *protocol.StreamEvent) {
	// Build path to task log directory
	logDir := filepath.Join(execution.ProjectRoot, BeadsDir, "tasks", execution.TaskID)
	logPath := filepath.Join(logDir, "execution.log")

	// Ensure directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		monitoring.Logger.Warn("failed_to_create_log_dir", "task_id", execution.TaskID, "error", err.Error())
		return
	}

	// Marshal event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		monitoring.Logger.Warn("failed_to_marshal_stream_event", "task_id", execution.TaskID, "error", err.Error())
		return
	}

	// Append to log file
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		monitoring.Logger.Warn("failed_to_open_task_log", "task_id", execution.TaskID, "error", err.Error())
		return
	}
	defer f.Close()

	// Write JSON line
	if _, err := f.Write(append(data, '\n')); err != nil {
		monitoring.Logger.Warn("failed_to_write_stream_event", "task_id", execution.TaskID, "error", err.Error())
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

// loadProjectRoots loads the project registry from config
func (s *AgentServer) loadProjectRoots() error {
	if s.config.Projects == nil {
		s.config.Projects = make(map[string]string)
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for path, lastAccessedStr := range s.config.Projects {
		lastAccessed, err := time.Parse(time.RFC3339, lastAccessedStr)
		if err != nil {
			continue // Skip invalid entries
		}
		s.projectRoots[path] = lastAccessed
	}

	monitoring.Logger.Info("project_registry_loaded", "count", len(s.projectRoots))
	return nil
}

// saveProjectRoots persists the project registry to config file
func (s *AgentServer) saveProjectRoots() error {
	s.mu.RLock()
	registry := make(map[string]string)
	for path, lastAccessed := range s.projectRoots {
		registry[path] = lastAccessed.Format(time.RFC3339)
	}
	s.mu.RUnlock()

	// Update config
	s.config.Projects = registry

	// Save to config file (~/.claude/agent-server.json)
	configPath := resolveConfigPath()
	if configPath == "" {
		// No config file to save to
		return nil
	}

	return config.SaveConfig(s.config, configPath)
}

// resolveConfigPath returns the path to the active config file
func resolveConfigPath() string {
	// Check for explicit config path from environment
	if envPath := os.Getenv("AGENT_SERVER_CONFIG"); envPath != "" {
		return envPath
	}

	// Default to ~/.claude/agent-server.json
	if homeDir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(homeDir, ".claude", "agent-server.json")
	}

	return ""
}

// cleanupInactiveProjects removes projects that haven't been accessed in a while
func (s *AgentServer) cleanupInactiveProjects() {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -ProjectInactiveDays)
	removed := 0

	for path, lastAccessed := range s.projectRoots {
		if lastAccessed.Before(cutoff) {
			delete(s.projectRoots, path)
			removed++
		}
	}

	if removed > 0 {
		monitoring.Logger.Info("cleaned_inactive_projects", "count", removed)
		// Save after cleanup
		go func() {
			if err := s.saveProjectRoots(); err != nil {
				monitoring.Logger.Warn("failed_to_save_after_cleanup", "error", err.Error())
			}
		}()
	}
}

// handleOrphanedTasks checks for beads tasks that are in_progress but have no active execution
// This happens when the server restarts while tasks were running
// Also cleans up stale entries in activeTasks map where Beads shows task as completed/failed
func (s *AgentServer) handleOrphanedTasks() {
	// Wait a bit for server to fully initialize
	time.Sleep(2 * time.Second)

	projectRoots := s.GetProjectRoots()
	orphanedCount := 0
	staleCount := 0

	// Build a map of all Beads tasks by ID for quick lookup
	beadsTasksByID := make(map[string]*beads.Task)
	for _, projectRoot := range projectRoots {
		beadsTasks, err := s.beadsClient.ListAllTasksFromDir(projectRoot)
		if err != nil {
			continue
		}

		for _, beadsTask := range beadsTasks {
			beadsTasksByID[beadsTask.ID] = &beadsTask

			// Check for orphaned tasks (marked in_progress in Beads but not running)
			if beadsTask.Status == "in_progress" {
				s.mu.RLock()
				_, hasActiveTask := s.activeTasks[beadsTask.ID]
				s.mu.RUnlock()

				if !hasActiveTask {
					// This is an orphaned task - it's marked in_progress but not running
					// Mark it as "open" so it can be retried/restarted
					monitoring.Logger.Warn("orphaned_task_detected",
						"task_id", beadsTask.ID,
						"title", beadsTask.Title,
						"project", projectRoot,
						"action", "resetting_to_open",
					)

					// Reset the task to "open" status in Beads
					// This allows it to be retried without creating a new task
					if beads.IsInstalled() {
						// Use bd update to reset status
						cmd := exec.Command("bd", "update", "--status", "open", beadsTask.ID)
						cmd.Dir = projectRoot
						if err := cmd.Run(); err != nil {
							monitoring.Logger.Error("failed_to_reset_orphaned_task",
								"task_id", beadsTask.ID,
								"error", err.Error())
						} else {
							monitoring.Logger.Info("orphaned_task_reset",
								"task_id", beadsTask.ID,
								"new_status", "open")
						}
					}

					orphanedCount++
				}
			}
		}
	}

	// Check for stale entries in activeTasks - tasks that are no longer in_progress in Beads
	s.mu.Lock()
	for taskID, execution := range s.activeTasks {
		beadsTask, exists := beadsTasksByID[taskID]

		// If task doesn't exist in Beads or is no longer in_progress, remove from activeTasks
		if !exists || (beadsTask.Status != "in_progress") {
			monitoring.Logger.Warn("stale_active_task_removed",
				"task_id", taskID,
				"beads_exists", exists,
				"beads_status", func() string {
					if beadsTask != nil {
						return beadsTask.Status
					}
					return "not_found"
				}(),
			)
			delete(s.activeTasks, taskID)

			// Close the stream if it exists
			if execution.streamChan != nil {
				s.closeStream(execution)
			}
			staleCount++
		}
	}
	s.mu.Unlock()

	if orphanedCount > 0 || staleCount > 0 {
		monitoring.Logger.Info("task_cleanup_summary",
			"orphaned", orphanedCount,
			"stale_removed", staleCount,
		)
	}
}

// archiveOldTasks archives completed/closed tasks older than configured threshold
func (s *AgentServer) archiveOldTasks() {
	if !s.config.TaskCleanup.Enabled {
		return
	}

	archiveDays := s.config.TaskCleanup.ArchiveAfterDays
	if archiveDays <= 0 {
		monitoring.Logger.Warn("invalid_archive_days", "days", archiveDays)
		return
	}

	cutoffTime := time.Now().AddDate(0, 0, -archiveDays)
	projectRoots := s.GetProjectRoots()
	totalArchived := 0

	monitoring.Logger.Info("starting_task_archival",
		"archive_after_days", archiveDays,
		"cutoff_date", cutoffTime.Format("2006-01-02"),
		"project_count", len(projectRoots))

	for _, projectRoot := range projectRoots {
		// List all tasks from this project
		beadsTasks, err := s.beadsClient.ListAllTasksFromDir(projectRoot)
		if err != nil {
			monitoring.Logger.Warn("failed_to_list_tasks_for_archival",
				"project", projectRoot,
				"error", err.Error())
			continue
		}

		archivedInProject := 0

		for _, task := range beadsTasks {
			// Only archive closed/completed tasks
			if task.Status != "closed" && task.Status != "done" && task.Status != "completed" {
				continue
			}

			// Check if task is old enough to archive
			var taskDate time.Time
			if task.ClosedAt != "" {
				taskDate, err = time.Parse(time.RFC3339, task.ClosedAt)
				if err != nil {
					// Try alternate format
					taskDate, err = time.Parse("2006-01-02T15:04:05", task.ClosedAt)
					if err != nil {
						monitoring.Logger.Warn("failed_to_parse_closed_at",
							"task_id", task.ID,
							"closed_at", task.ClosedAt)
						continue
					}
				}
			} else if task.UpdatedAt != "" {
				taskDate, err = time.Parse(time.RFC3339, task.UpdatedAt)
				if err != nil {
					taskDate, err = time.Parse("2006-01-02T15:04:05", task.UpdatedAt)
					if err != nil {
						continue
					}
				}
			} else {
				continue
			}

			// Skip if not old enough
			if taskDate.After(cutoffTime) {
				continue
			}

			// Archive the task
			if err := s.archiveTask(projectRoot, &task); err != nil {
				monitoring.Logger.Warn("failed_to_archive_task",
					"task_id", task.ID,
					"project", projectRoot,
					"error", err.Error())
			} else {
				archivedInProject++
				totalArchived++
			}
		}

		if archivedInProject > 0 {
			monitoring.Logger.Info("archived_tasks_in_project",
				"project", projectRoot,
				"count", archivedInProject)
		}
	}

	if totalArchived > 0 {
		monitoring.Logger.Info("task_archival_complete",
			"total_archived", totalArchived,
			"archive_after_days", archiveDays)
	} else {
		monitoring.Logger.Info("no_tasks_to_archive",
			"archive_after_days", archiveDays)
	}
}

// archiveTask moves a task's data to the archive directory
func (s *AgentServer) archiveTask(projectRoot string, task *beads.Task) error {
	// Source: .beads/tasks/<task-id>/
	taskDir := filepath.Join(projectRoot, BeadsDir, "tasks", task.ID)

	// Check if task directory exists
	if _, err := os.Stat(taskDir); os.IsNotExist(err) {
		// Task directory doesn't exist, nothing to archive
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to stat task directory: %w", err)
	}

	// Destination: .beads/archive/<YYYY-MM>/<task-id>/
	now := time.Now()
	archiveDir := filepath.Join(projectRoot, BeadsDir, "archive", now.Format("2006-01"))

	// Create archive directory if it doesn't exist
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}

	// Move task directory to archive
	destDir := filepath.Join(archiveDir, task.ID)
	if err := os.Rename(taskDir, destDir); err != nil {
		return fmt.Errorf("failed to move task to archive: %w", err)
	}

	monitoring.Logger.Info("task_archived",
		"task_id", task.ID,
		"title", task.Title,
		"archive_path", destDir)

	return nil
}

// GetActiveTaskCount returns the number of currently running tasks
func (s *AgentServer) GetActiveTaskCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.activeTasks)
}

// GetActiveTaskIDs returns a list of active task IDs and their roles
func (s *AgentServer) GetActiveTaskIDs() []map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]map[string]string, 0, len(s.activeTasks))
	for taskID, execution := range s.activeTasks {
		tasks = append(tasks, map[string]string{
			"task_id": taskID,
			"role":    execution.Role,
			"status":  execution.Status,
		})
	}
	return tasks
}

// Shutdown performs graceful shutdown of the server
// Returns true if shutdown should proceed, false if user cancelled
func (s *AgentServer) Shutdown(ctx context.Context) error {
	monitoring.Logger.Info("shutdown_initiated")

	// Check for active tasks
	activeCount := s.GetActiveTaskCount()
	if activeCount > 0 {
		activeTasks := s.GetActiveTaskIDs()
		monitoring.Logger.Warn("shutdown_with_active_tasks",
			"count", activeCount,
			"tasks", activeTasks)

		// Log details about active tasks
		for _, task := range activeTasks {
			monitoring.Logger.Warn("active_task_during_shutdown",
				"task_id", task["task_id"],
				"role", task["role"],
				"status", task["status"])
		}

		return fmt.Errorf("cannot shutdown: %d active task(s) running. Wait for completion or cancel tasks first", activeCount)
	}

	monitoring.Logger.Info("shutdown_complete", "active_tasks", 0)
	return nil
}
