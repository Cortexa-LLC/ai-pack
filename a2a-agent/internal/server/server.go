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
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/streaming"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/execution_log"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/protocol"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/proxy"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/tools"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/constants"
	"github.com/sashabaranov/go-openai"
	"gopkg.in/yaml.v3"
)

var Version = "2.1.0"

type AgentServer struct {
	rootDir          string
	anthropicKey     string
	client           anthropic.Client // SDK v1.19+ returns Client by value
	beadsClient      *beads.Client
	claudeSettings   *claude.Settings // Claude Code settings (deny patterns, etc.)
	executionLog     *execution_log.ExecutionLog // Persistent agent execution log
	maxConcurrent    int              // Maximum concurrent agents (configurable)
	maxTokens        int              // Maximum tokens per API call
	model            string           // Default Anthropic model to use
	maxInactiveTurns int              // Stop agent after N turns without progress
	config           *config.Config   // Server configuration

	// Multi-provider LLM support
	openaiKey          string
	openaiClient       *openai.Client
	anthropicProvider  *AnthropicProvider
	openaiProvider     *OpenAIProvider
	modelSelector      *ModelSelector
	streamingService   *streaming.Service // Clean streaming abstraction
	projectMetrics     map[string]*monitoring.PersistentMetrics // Per-project persistent metrics
	providerCosts      map[string][2]float64 // Provider cost configuration

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
	Model       string `yaml:"model"` // LLM model to use (e.g., "gpt-4o-mini", "claude-sonnet-3-5-20241022")
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

// GetProjectCostsData returns cost data for all projects
func (s *AgentServer) GetProjectCostsData() ([]map[string]interface{}, error) {
	projects := s.GetProjectRoots()
	result := make([]map[string]interface{}, 0, len(projects))

	for _, projectRoot := range projects {
		// Try to get daily usage from the project's metrics directory
		metricsFile := filepath.Join(projectRoot, ".claude", "metrics", "daily", time.Now().Format("2006-01-02")+".json")
		data, err := os.ReadFile(metricsFile)
		if err != nil {
			continue // Skip projects without today's metrics
		}

		var dailyUsage struct {
			TotalInputTokens  int64 `json:"total_input_tokens"`
			TotalOutputTokens int64 `json:"total_output_tokens"`
			ProviderBreakdown map[string]struct {
				Provider     string  `json:"provider"`
				Model        string  `json:"model"`
				Calls        int64   `json:"calls"`
				InputTokens  int64   `json:"input_tokens"`
				OutputTokens int64   `json:"output_tokens"`
				Cost         float64 `json:"cost"`
			} `json:"provider_breakdown"`
		}

		if err := json.Unmarshal(data, &dailyUsage); err != nil {
			continue
		}

		// Calculate total cost
		totalCost := 0.0
		providers := make([]map[string]interface{}, 0)

		for _, usage := range dailyUsage.ProviderBreakdown {
			totalCost += usage.Cost
			providers = append(providers, map[string]interface{}{
				"provider":      usage.Provider,
				"model":         usage.Model,
				"calls":         usage.Calls,
				"inputTokens":   usage.InputTokens,
				"outputTokens":  usage.OutputTokens,
				"cost":          usage.Cost,
			})
		}

		projectName := filepath.Base(projectRoot)
		result = append(result, map[string]interface{}{
			"projectRoot":       projectRoot,
			"projectName":       projectName,
			"totalCost":         totalCost,
			"totalInputTokens":  dailyUsage.TotalInputTokens,
			"totalOutputTokens": dailyUsage.TotalOutputTokens,
			"providerBreakdown": providers,
		})
	}

	return result, nil
}

// getOrCreateProjectMetrics returns or creates persistent metrics for a project
func (s *AgentServer) getOrCreateProjectMetrics(projectRoot string) (*monitoring.PersistentMetrics, error) {
	s.mu.RLock()
	pm, exists := s.projectMetrics[projectRoot]
	s.mu.RUnlock()

	if exists {
		return pm, nil
	}

	// Create new metrics for this project
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check again in case another goroutine created it
	if pm, exists := s.projectMetrics[projectRoot]; exists {
		return pm, nil
	}

	pm, err := monitoring.NewPersistentMetrics(projectRoot, s.providerCosts)
	if err != nil {
		return nil, err
	}

	s.projectMetrics[projectRoot] = pm
	return pm, nil
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

	// Initialize OpenAI client for multi-provider support
	var openaiClient *openai.Client
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey != "" {
		openaiClient = openai.NewClient(openaiKey)
		monitoring.Logger.Info("openai_client_initialized", "provider", "openai")
	} else {
		monitoring.Logger.Warn("openai_api_key_not_set", "message", "Set OPENAI_API_KEY to enable GPT models for cost savings")
	}

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

	// Build cost map from config for per-project metrics
	costs := make(map[string][2]float64)
	for modelKey, modelCost := range cfg.ProviderCosts.Models {
		costs[modelKey] = [2]float64{modelCost.InputCost, modelCost.OutputCost}
	}

	// Initialize performance grading system
	gradesDir := filepath.Join(rootDir, ".claude", "performance_grades")
	if err := monitoring.InitGradeManager(gradesDir, &cfg.GradingCriteria); err != nil {
		monitoring.Logger.Warn("failed_to_init_grade_manager", "error", err.Error())
	} else {
		monitoring.Logger.Info("performance_grading_initialized", "grades_dir", gradesDir)
	}

	// Initialize complexity analyzer
	monitoring.InitComplexityAnalyzer()

	// Initialize model selector
	if monitoring.GlobalGradeManager != nil && monitoring.GlobalComplexityAnalyzer != nil {
		monitoring.InitModelSelector(monitoring.GlobalGradeManager, monitoring.GlobalComplexityAnalyzer)
		monitoring.Logger.Info("adaptive_model_selection_enabled")
	}

	// Create LLM providers for multi-provider support
	anthropicProvider := NewAnthropicProvider(client, model, maxTokens)
	var openaiProvider *OpenAIProvider
	if openaiClient != nil {
		openaiProvider = NewOpenAIProvider(openaiClient, "gpt-4o-mini", maxTokens)
	}

	// Create server instance first (needed for model selector)
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

		// Multi-provider LLM support
		openaiKey:          openaiKey,
		openaiClient:       openaiClient,
		anthropicProvider:  anthropicProvider,
		openaiProvider:     openaiProvider,
		projectMetrics:     make(map[string]*monitoring.PersistentMetrics),
		providerCosts:      costs,

		activeTasks:      make(map[string]*TaskExecution),
		taskQueue:        make(chan *TaskExecution, 100),
		workerPool:       make(chan struct{}, maxConcurrent),
		projectRoots:     make(map[string]time.Time),
	}

	// Initialize model selector
	server.modelSelector = NewModelSelector(server)

	// Initialize streaming service with performance-grade-based model selection
	modelSelector := streaming.NewPerformanceGradeModelSelector(
		rootDir,
		model,
		openaiClient != nil,
	)
	streamingService := streaming.NewService(modelSelector, "anthropic")

	// Register provider factories
	streamingService.RegisterProvider(streaming.NewAnthropicFactory(apiKey, maxTokens, clientOpts...))
	if openaiClient != nil {
		streamingService.RegisterProvider(streaming.NewOpenAIFactory(openaiClient, maxTokens))
	}

	server.streamingService = streamingService
	monitoring.Logger.Info("streaming_service_initialized",
		"default", "anthropic")

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
	// Validate Beads is installed
	if !beads.IsInstalled() {
		return nil, fmt.Errorf("Beads is not installed. All tasks must be Beads tasks. Install with: brew install beads")
	}

	// Validate taskInput is a Beads task ID
	if !beads.IsBeadsTaskID(taskInput) {
		return nil, fmt.Errorf("invalid task ID format. All tasks must be Beads tasks. Create with: bd create '<description>'")
	}

	// If no project root specified, use server's root directory
	if projectRoot == "" {
		projectRoot = s.rootDir
	}

	// Validate the task exists in Beads (use project root for bd commands)
	if err := s.beadsClient.ValidateTaskIDFromDir(taskInput, projectRoot); err != nil {
		return nil, fmt.Errorf("invalid Beads task: %w", err)
	}

	beadsTaskID := taskInput
	monitoring.Logger.Info("spawning_with_beads_task", "task_id", beadsTaskID, "project_root", projectRoot)

	// Check dependencies
	depsOK, unmetDeps, err := s.beadsClient.CheckDependenciesFromDir(beadsTaskID, projectRoot)
	if err != nil {
		monitoring.Logger.Warn("dependency_check_failed", "error", err.Error())
	} else if !depsOK {
		return nil, fmt.Errorf("task %s has unmet dependencies: %v\nPlease complete these tasks first", beadsTaskID, unmetDeps)
	}

	// Get task description, task packet path, and working directory
	taskDescription, taskPacketPath, workingDir, _, err := s.beadsClient.GetTaskDescriptionFromDir(taskInput, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get task description: %w", err)
	}

	// If working directory not specified in task, use project root
	if workingDir == "" {
		workingDir = projectRoot
	}

	// Use Beads task ID with timestamp as the task ID (single source of truth)
	// Format: {beads-id}-{YYYYMMDD}-{HHMMSS}
	timestamp := time.Now().Format("20060102-150405")
	taskID := fmt.Sprintf("%s-%s", beadsTaskID, timestamp)

	// Load agent configuration
	config, err := s.loadAgentConfig(role, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load agent config: %w", err)
	}

	// Create task packet in project's .beads/tasks/ directory
	if err := s.createTaskPacketInProject(taskID, role, taskDescription, config, projectRoot); err != nil {
		return nil, fmt.Errorf("failed to create task packet: %w", err)
	}

	// Store task packet path, working directory, and project root in metadata if available
	metadata := map[string]string{}
	// CRITICAL: Store the Beads task ID so retry/logs can reference it
	metadata["beads_task_id"] = beadsTaskID
	if taskPacketPath != "" {
		metadata["task_packet_path"] = taskPacketPath
	}
	if workingDir != "" {
		metadata["working_directory"] = workingDir
	}
	if projectRoot != "" {
		metadata["project_root"] = projectRoot
	}

	// Mark Beads task as started
	if err := s.beadsClient.StartTaskFromDir(beadsTaskID, projectRoot); err != nil {
		monitoring.Logger.Warn("failed_to_start_beads_task", "error", err.Error())
	} else {
		monitoring.Logger.Info("beads_task_started", "task_id", beadsTaskID)
	}

	// Create task execution
	execution := &TaskExecution{
		TaskID:      taskID,
		Role:        role,
		Task:        taskDescription,
		Config:      config,
		StartTime:   time.Now(),
		Status:      constants.StatusQueued,
		ProjectRoot: projectRoot,
		streamChan:  make(chan *protocol.StreamEvent, 100),
		streamOpen:  true,
		metadata:    metadata,
	}

	// Update task packet metadata with Beads task ID and project root
	if err := s.updateTaskPacketMetadataInProject(taskID, metadata, projectRoot); err != nil {
		monitoring.Logger.Warn("failed_to_update_task_metadata", "error", err.Error())
	}

	// Check for legacy folders before registering new project root
	if projectRoot != "" && projectRoot != s.rootDir {
		s.mu.RLock()
		_, alreadyRegistered := s.projectRoots[projectRoot]
		s.mu.RUnlock()

		if !alreadyRegistered {
			// New project - check for legacy folders
			hasLegacy, legacyFolders := DetectLegacyTaskFoldersInProject(projectRoot)
			if hasLegacy {
				monitoring.Logger.Error("legacy_folders_detected_in_new_project",
					"project_root", projectRoot,
					"legacy_count", len(legacyFolders))
				return nil, fmt.Errorf("project %s has %d legacy task folders that must be migrated first. Run: agent-server --migrate-tasks", projectRoot, len(legacyFolders))
			}
		}
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
		Status:    constants.StatusQueued,
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

	if execution.Status == constants.StatusCompleted || execution.Status == constants.StatusFailed {
		completedAt := time.Now()
		response.CompletedAt = &completedAt
	}

	return response, nil
}

func (s *AgentServer) loadAgentConfig(role string, projectRoot string) (*AgentConfig, error) {
	// Use provided projectRoot or fall back to server root
	if projectRoot == "" {
		projectRoot = s.rootDir
	}

	// Try paths in order:
	// 1. Project override (.ai/agents/)
	// 2. Framework (.ai-pack/agents/) - production
	// 3. Development (../agents/) - when running from a2a-agent dir
	// 4. Development (agents/) - when running from repo root
	candidatePaths := []struct {
		path   string
		source string
	}{
		{filepath.Join(projectRoot, ".ai", "agents", role+".yml"), "project_override"},
		{filepath.Join(projectRoot, ".ai-pack", "agents", role+".yml"), "framework"},
		{filepath.Join(projectRoot, "..", "agents", role+".yml"), "dev_parent"},
		{filepath.Join(projectRoot, "agents", role+".yml"), "dev_root"},
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

func (s *AgentServer) loadRoleContext(roleFile string, projectRoot string) (string, error) {
	// Use provided projectRoot or fall back to server root
	if projectRoot == "" {
		projectRoot = s.rootDir
	}

	// Support override pattern: try .ai/ first, then .ai-pack/
	var rolePath string

	// If roleFile starts with .ai/, try project override first
	if strings.HasPrefix(roleFile, ".ai/") {
		projectPath := filepath.Join(projectRoot, roleFile)
		if _, err := os.Stat(projectPath); err == nil {
			rolePath = projectPath
		} else {
			// Fallback to framework path
			frameworkPath := strings.Replace(roleFile, ".ai/", ".ai-pack/", 1)
			rolePath = filepath.Join(projectRoot, frameworkPath)
		}
	} else {
		// Direct path specified
		rolePath = filepath.Join(projectRoot, roleFile)
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
	taskDir := filepath.Join(projectRoot, constants.BeadsDir, "tasks", taskID)

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
		"status":      constants.StatusQueued,
		"config":      config,
		"updated_at":  time.Now().Format(time.RFC3339),
	}

	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(filepath.Join(taskDir, constants.MetadataFileName), metadataJSON, 0644); err != nil {
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
	metadataPath := filepath.Join(projectRoot, constants.BeadsDir, "tasks", taskID, constants.MetadataFileName)

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
	// Search across all registered project roots (not just server root)
	projectRoots := s.GetProjectRoots()

	var metadataPath string
	var data []byte
	var err error
	var foundProjectRoot string

	for _, projectRoot := range projectRoots {
		// Try direct path first
		metadataPath = filepath.Join(projectRoot, constants.BeadsDir, "tasks", taskID, constants.MetadataFileName)
		data, err = os.ReadFile(metadataPath)

		if err != nil {
			// If direct path doesn't exist, try finding most recent execution folder
			// This handles the case where taskID is just the Beads ID without timestamp
			executionFolder := s.findMostRecentExecutionInProject(projectRoot, taskID)
			if executionFolder != "" {
				metadataPath = filepath.Join(projectRoot, constants.BeadsDir, "tasks", executionFolder, constants.MetadataFileName)
				data, err = os.ReadFile(metadataPath)
			}
		}

		if err == nil {
			foundProjectRoot = projectRoot
			break
		}
	}

	if err != nil || foundProjectRoot == "" {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Read results if available
	var result string
	// Extract the execution folder from the metadataPath we found
	executionFolder := filepath.Base(filepath.Dir(metadataPath))
	resultsPath := filepath.Join(foundProjectRoot, constants.BeadsDir, "tasks", executionFolder, "30-results.md")
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

	// Reconcile execution metadata with Beads task status to fix stale data
	// This handles cases where tasks were blocked/failed but later completed

	// Get or infer the Beads task ID
	beadsTaskID := ""
	if btid, ok := metadata["beads_task_id"].(string); ok && btid != "" {
		beadsTaskID = btid
	} else {
		// For older executions without beads_task_id, infer from execution folder name
		// Format: {beads-id}-YYYYMMDD-HHMMSS → extract {beads-id}
		// Example: xasm++-5tu1-20260214-085757 → xasm++-5tu1
		// The timestamp is always the last 2 parts: 8-digit date and 6-digit time
		parts := strings.Split(executionFolder, "-")
		if len(parts) >= 3 {
			// Check if last 2 parts look like timestamps (8 digits + 6 digits)
			lastPart := parts[len(parts)-1]
			secondLastPart := parts[len(parts)-2]
			if len(lastPart) == 6 && len(secondLastPart) == 8 {
				// Last 2 parts are date-time, everything before is the Beads ID
				beadsTaskID = strings.Join(parts[:len(parts)-2], "-")
				monitoring.Logger.Debug("inferred_beads_task_id",
					"execution_folder", executionFolder,
					"inferred_id", beadsTaskID)
			}
		}
	}

	if beadsTaskID != "" {
		if beadsTask, err := s.beadsClient.GetTaskFromDir(beadsTaskID, foundProjectRoot); err == nil {
			beadsStatus := strings.ToLower(beadsTask.Status)

			// If Beads shows closed/completed but execution shows blocked/failed, reconcile
			if (beadsStatus == constants.StatusClosed || beadsStatus == constants.StatusCompleted) &&
			   (status == "blocked" || status == "failed") {
				monitoring.Logger.Info("reconciling_stale_execution_metadata",
					"task_id", beadsTaskID,
					"old_status", status,
					"beads_status", beadsStatus,
					"execution_folder", executionFolder)

				// Update metadata
				metadata["status"] = constants.StatusCompleted
				metadata["error"] = nil
				metadata["updated_at"] = time.Now().Format(time.RFC3339)
				metadata["reconciled"] = true
				metadata["reconciled_at"] = time.Now().Format(time.RFC3339)

				// Store the Beads task ID if it was missing
				if _, ok := metadata["beads_task_id"]; !ok || metadata["beads_task_id"] == nil {
					metadata["beads_task_id"] = beadsTaskID
				}

				// Write back to disk
				if updatedData, err := json.MarshalIndent(metadata, "", "  "); err == nil {
					if err := os.WriteFile(metadataPath, updatedData, 0644); err == nil {
						status = constants.StatusCompleted
						updatedAt = time.Now()
						monitoring.Logger.Info("reconciled_execution_metadata",
							"task_id", beadsTaskID,
							"execution_folder", executionFolder)
					}
				}
			}
		}
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

// markExecutionAsSuperseded marks an execution as superseded (replaced by retry/rerun)
// This prevents stale execution metadata from showing in the GUI after retries
func (s *AgentServer) markExecutionAsSuperseded(taskID, projectRoot, reason string) error {
	// Find the execution metadata file
	var metadataPath string

	if projectRoot == "" {
		// Search all project roots
		for _, root := range s.GetProjectRoots() {
			path := filepath.Join(root, constants.BeadsDir, "tasks", taskID, constants.MetadataFileName)
			if _, err := os.Stat(path); err == nil {
				metadataPath = path
				break
			}
		}
	} else {
		metadataPath = filepath.Join(projectRoot, constants.BeadsDir, "tasks", taskID, constants.MetadataFileName)
	}

	if metadataPath == "" {
		return fmt.Errorf("metadata not found for task %s", taskID)
	}

	// Read existing metadata
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to read metadata: %w", err)
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		return fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Mark as superseded
	metadata["superseded"] = true
	metadata["superseded_at"] = time.Now().Format(time.RFC3339)
	metadata["superseded_reason"] = reason
	metadata["updated_at"] = time.Now().Format(time.RFC3339)

	// Write back
	updatedData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	monitoring.Logger.Info("marked_execution_superseded",
		"task_id", taskID,
		"reason", reason)

	return nil
}

// findMostRecentExecutionFolderInRoot finds the most recent timestamped execution folder for a Beads task ID in the server root
// Returns the folder name (e.g., "xasm++-syq1-20260211-080508") or empty string if not found
func (s *AgentServer) findMostRecentExecutionFolderInRoot(beadsTaskID string) string {
	return s.findMostRecentExecutionInProject(s.rootDir, beadsTaskID)
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

func (s *AgentServer) updateTaskStatus(taskID, projectRoot, status, errorMsg string) error {
	metadataPath := filepath.Join(projectRoot, constants.BeadsDir, "tasks", taskID, constants.MetadataFileName)
	monitoring.Logger.Info("updating_task_status", "task_id", taskID, "status", status, "path", metadataPath)

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		monitoring.Logger.Error("metadata_read_error", "task_id", taskID, "path", metadataPath, "error", err)
		return fmt.Errorf("failed to read metadata: %w", err)
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		monitoring.Logger.Error("metadata_unmarshal_error", "task_id", taskID, "error", err)
		return fmt.Errorf("failed to parse metadata: %w", err)
	}

	metadata["status"] = status
	metadata["updated_at"] = time.Now().Format(time.RFC3339)
	if errorMsg != "" {
		metadata["error"] = errorMsg
	}

	updatedData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		monitoring.Logger.Error("metadata_marshal_error", "task_id", taskID, "error", err)
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, updatedData, 0644); err != nil {
		monitoring.Logger.Error("metadata_write_error", "task_id", taskID, "path", metadataPath, "error", err)
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	monitoring.Logger.Info("task_status_updated", "task_id", taskID, "status", status)
	return nil
}

// buildSystemPrompt creates system messages with prompt caching for role context
func (s *AgentServer) buildSystemPrompt(roleContext string) []anthropic.TextBlockParam {
	// Cache the role context with 5-minute TTL (auto-refreshed on each turn)
	// This provides massive speedup across agentic loop turns while automatically
	// expiring between tasks
	return []anthropic.TextBlockParam{
		{
			Text:         roleContext,
			Type:         constants.ContentTypeText,
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
// executeAgenticLoop runs the agentic loop with tool support
func (s *AgentServer) executeAgenticLoop(ctx context.Context, taskID string, role string, initialPrompt string, roleContext string, workingDir string, config *AgentConfig, logMsg func(string)) (string, error) {
	// Define tools (matches Claude Code's exact tool names)
	toolDefs := tools.DefineTools()

	// Build system prompt with caching for role context
	systemPrompt := s.buildSystemPrompt(roleContext)

	// Build messages starting with user prompt (task-specific)
	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(initialPrompt)),
	}

	logMsg(fmt.Sprintf("🔄 Starting agentic loop (max_inactive: %d, caching: enabled, extended_thinking: %v)",
		s.maxInactiveTurns, config.ExtendedThinking))

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

		// Simple message-count-based truncation to keep context manageable
		// Keep first message (initial prompt) + last 50 messages
		const maxHistoryMessages = 50
		truncatedMessages := messages
		if len(messages) > 1+maxHistoryMessages {
			firstMsg := messages[0] // Keep initial prompt with task context
			recentMsgs := messages[len(messages)-maxHistoryMessages:]
			truncatedMessages = append([]anthropic.MessageParam{firstMsg}, recentMsgs...)

			// Only log truncation occasionally to avoid spam
			if len(messages)%10 == 0 {
				logMsg(fmt.Sprintf("      📉 Truncated: %d → %d messages", len(messages), len(truncatedMessages)))
			}
		}

		// Convert messages to provider-agnostic format
		streamMessages := make([]streaming.Message, 0, len(truncatedMessages))
		for _, msg := range truncatedMessages {
			// Extract text from Anthropic message format
			var content string
			for _, block := range msg.Content {
				if textBlock := block.OfText; textBlock != nil {
					content += textBlock.Text
				}
			}

			streamMessages = append(streamMessages, streaming.Message{
				Role:    string(msg.Role),
				Content: content,
			})
		}

		// Convert tools to provider-agnostic format
		streamTools := make([]streaming.Tool, 0, len(toolDefs))
		for _, tool := range toolDefs {
			streamTools = append(streamTools, streaming.Tool{
				Name:        tool.Name,
				Description: "", // TODO: Extract from param.Opt
				InputSchema: map[string]interface{}{
					"type":       tool.InputSchema.Type,
					"properties": tool.InputSchema.Properties,
					"required":   tool.InputSchema.Required,
				},
			})
		}

		// Build system prompt string from blocks
		var systemPromptStr string
		for _, block := range systemPrompt {
			systemPromptStr += block.Text
		}

		// Prepare streaming request
		streamReq := streaming.StreamRequest{
			Messages:     streamMessages,
			SystemPrompt: systemPromptStr,
			MaxTokens:    s.maxTokens,
			Model:        s.model, // Default model - will be overridden by role-based selection
			Tools:        streamTools,
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

		// Record per-turn token metrics
		monitoring.GlobalMetrics.RecordTurnTokens(taskID, turn, inputTokens, outputTokens, apiDuration)

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

		// Log tool uses
		for _, toolUse := range toolUses {
			logMsg(fmt.Sprintf("      🔧 Tool: %s", toolUse.Name))
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
			// Execute tool with Claude settings
			result, err := tools.ExecuteTool(toolUse.Name, toolUse.Input, workingDir, s.claudeSettings)
			if err != nil {
				logMsg(fmt.Sprintf("         ❌ Tool execution failed: %v", err))
				result = fmt.Sprintf("Error: %v", err)
			} else {
				// Log full tool result without truncation
				logMsg(fmt.Sprintf("         ✓ %s", result))
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
			inputJSON, _ := json.Marshal(toolUse.Input)
			if len(inputJSON) > 100 {
				inputJSON = inputJSON[:100]
			}
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
		var assistantContent []anthropic.ContentBlockParamUnion
		if hasText {
			assistantContent = append(assistantContent, anthropic.NewTextBlock(responseTextStr))
		}
		for _, toolUse := range toolUses {
			// Convert tool input to json.RawMessage
			inputJSON, _ := json.Marshal(toolUse.Input)
			assistantContent = append(assistantContent, anthropic.NewToolUseBlock(toolUse.ID, json.RawMessage(inputJSON), toolUse.Name))
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

// setupExecutionLogger creates the execution log file and returns a logging function
func (s *AgentServer) setupExecutionLogger(execution *TaskExecution) func(string) {
	logPath := filepath.Join(execution.ProjectRoot, constants.BeadsDir, "tasks", execution.TaskID, "execution.log")
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
func (s *AgentServer) initializeTaskExecution(execution *TaskExecution, logMsg func(string)) error {
	s.mu.Lock()
	execution.Status = constants.StatusInProgress
	s.mu.Unlock()

	if err := s.updateTaskStatus(execution.TaskID, execution.ProjectRoot, constants.StatusInProgress, ""); err != nil {
		logMsg(fmt.Sprintf("❌ Failed to update status: %v", err))
		return fmt.Errorf("failed to update task status: %w", err)
	}
	logMsg("📝 Status updated: in_progress")

	s.sendStreamEvent(execution, "status_update", map[string]interface{}{
		"status": constants.StatusInProgress,
	})

	// Log started event
	if s.executionLog != nil {
		if err := s.executionLog.LogStarted(execution.TaskID); err != nil {
			monitoring.Logger.Warn("failed_to_log_started_event", "error", err.Error())
		}
	}

	return nil
}

// loadAndLogRoleContext loads role context with logging and error handling
func (s *AgentServer) loadAndLogRoleContext(execution *TaskExecution, logMsg func(string)) (string, error) {
	logMsg(fmt.Sprintf("📖 Loading role context: %s", execution.Config.Context.RoleFile))
	roleContext, err := s.loadRoleContext(execution.Config.Context.RoleFile, execution.ProjectRoot)
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

	promptPath := filepath.Join(execution.ProjectRoot, constants.BeadsDir, "tasks", execution.TaskID, "agent-prompt.txt")
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

	result, err := s.executeAgenticLoop(ctx, execution.TaskID, execution.Role, prompt, roleContext, workingDir, execution.Config, logMsg)
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

	// Detect if the agent stopped due to blockers (missing task packets, etc.)
	// Check for common blocker patterns in the result text
	resultLower := strings.ToLower(result)
	isBlocked := strings.Contains(resultLower, "task packet missing") ||
		strings.Contains(resultLower, "blocking issue") ||
		strings.Contains(resultLower, "stop - task packet") ||
		strings.Contains(resultLower, "cannot proceed") ||
		strings.Contains(resultLower, "stopping - task packet required")

	var finalStatus string
	var statusMessage string

	if isBlocked {
		finalStatus = "blocked"
		statusMessage = "Task blocked: Agent stopped due to missing prerequisites (task packet)"
		logMsg("⚠️  Task marked as BLOCKED - agent identified missing prerequisites")
	} else {
		finalStatus = constants.StatusCompleted
		statusMessage = ""
	}

	// Update task status
	beadsTaskID, projectRoot := s.updateTaskCompletion(execution, result)

	// Complete Beads task only if truly completed (not blocked)
	var errorMsg string
	if beadsTaskID != "" && !isBlocked {
		if err := s.completeBeadsTask(beadsTaskID, projectRoot, logMsg); err != nil {
			errorMsg = fmt.Sprintf("Warning: %v", err)
			monitoring.Logger.Warn("beads_update_failed_but_task_completed", "task_id", execution.TaskID, "error", err.Error())
		}
	} else if isBlocked {
		errorMsg = statusMessage
	}

	// Finalize task
	durationMs := time.Since(startTime).Milliseconds()
	if err := s.updateTaskStatus(execution.TaskID, execution.ProjectRoot, finalStatus, errorMsg); err != nil {
		monitoring.Logger.Error("failed_to_update_status", "task_id", execution.TaskID, "status", finalStatus, "error", err)
		logMsg(fmt.Sprintf("⚠️  Warning: Failed to update status: %v", err))
	}
	s.sendStreamEvent(execution, finalStatus, map[string]interface{}{
		"result": result,
	})

	s.closeStream(execution)

	// Log completion event
	if s.executionLog != nil {
		resultSummary := result
		if len(resultSummary) > 500 {
			resultSummary = resultSummary[:500] + "..."
		}
		if err := s.executionLog.LogCompleted(execution.TaskID, durationMs, resultSummary); err != nil {
			monitoring.Logger.Warn("failed_to_log_completed_event", "error", err.Error())
		}
	}

	// Log appropriate message based on final status
	if finalStatus == "blocked" {
		logMsg(fmt.Sprintf("⚠️  Task blocked - prerequisites missing (duration: %dms)", durationMs))
		monitoring.Logger.Warn("task_blocked", "task_id", execution.TaskID, "duration_ms", durationMs)
	} else {
		logMsg(fmt.Sprintf("🎉 Task completed successfully (duration: %dms)", durationMs))
		monitoring.LogTaskCompleted(ctx, execution.TaskID, execution.Role, durationMs)
		monitoring.GlobalMetrics.IncrementTasksCompleted(durationMs)

		// Record performance grade for adaptive model selection
		if monitoring.GlobalGradeManager != nil && execution.metadata != nil {
			modelID := s.model // Default model
			if executionModel, ok := execution.metadata["model"]; ok {
				modelID = executionModel
			}

			tokensUsed := int64(0)
			if tokensStr, ok := execution.metadata["total_tokens"]; ok {
				fmt.Sscanf(tokensStr, "%d", &tokensUsed)
			}

			// Record successful completion
			if err := monitoring.GlobalGradeManager.RecordTaskCompletion(
				execution.TaskID,
				modelID,
				execution.Role,
				projectRoot,
				true, // success
				0,    // retries (we don't track this yet)
				tokensUsed,
				durationMs,
				false, // wasEscalated (not tracked yet)
				false, // wasDowngraded (not tracked yet)
			); err != nil {
				monitoring.Logger.Warn("failed_to_record_performance_grade", "error", err.Error())
			} else {
				monitoring.Logger.Debug("performance_grade_recorded", "task_id", execution.TaskID, "model", modelID, "role", execution.Role)
			}
		}
	}
	logMsg("=" + strings.Repeat("=", 70))
}

// saveTaskResults saves the task results to disk
func (s *AgentServer) saveTaskResults(execution *TaskExecution, result string, logMsg func(string)) {
	logMsg("💾 Saving results...")
	resultsPath := filepath.Join(execution.ProjectRoot, constants.BeadsDir, "tasks", execution.TaskID, "30-results.md")
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
	execution.Status = constants.StatusCompleted
	execution.Result = result
	beadsTaskID := execution.TaskID // TaskID is the Beads task ID
	projectRoot := ""
	if execution.metadata != nil {
		projectRoot = execution.metadata["project_root"]
	}
	// Remove from active tasks map since task is now completed
	delete(s.activeTasks, execution.TaskID)
	s.mu.Unlock()

	return beadsTaskID, projectRoot
}

// completeBeadsTask marks the corresponding Beads task as complete
// Returns error if the Beads update failed
func (s *AgentServer) completeBeadsTask(beadsTaskID string, projectRoot string, logMsg func(string)) error {
	if !beads.IsInstalled() {
		return nil
	}

	logMsg(fmt.Sprintf("🔗 Marking Beads task complete: %s", beadsTaskID))
	if err := s.beadsClient.CompleteTaskFromDir(beadsTaskID, projectRoot); err != nil {
		monitoring.Logger.Warn("failed_to_complete_beads_task", "task_id", beadsTaskID, "error", err.Error())
		logMsg(fmt.Sprintf("⚠️  Failed to complete Beads task: %v", err))
		return fmt.Errorf("beads update failed: %w", err)
	}

	monitoring.Logger.Info("beads_task_completed", "task_id", beadsTaskID)
	logMsg("✅ Beads task marked complete")
	return nil
}

func (s *AgentServer) failTask(execution *TaskExecution, errorMsg string) {
	ctx := context.Background()
	durationMs := time.Since(execution.StartTime).Milliseconds()

	// Log failure to execution log
	logPath := filepath.Join(execution.ProjectRoot, constants.BeadsDir, "tasks", execution.TaskID, "execution.log")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		timestamp := time.Now().Format("15:04:05")
		_, _ = logFile.WriteString(fmt.Sprintf("[%s] ❌ Task failed: %s\n", timestamp, errorMsg))
		_, _ = logFile.WriteString(fmt.Sprintf("[%s]    Duration: %dms\n", timestamp, durationMs))
		_, _ = logFile.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, strings.Repeat("=", 70)))
		_ = logFile.Close()
	}

	s.mu.Lock()
	execution.Status = constants.StatusFailed
	execution.Error = errorMsg
	// Remove from active tasks map since task is now failed
	delete(s.activeTasks, execution.TaskID)
	s.mu.Unlock()

	if err := s.updateTaskStatus(execution.TaskID, execution.ProjectRoot, constants.StatusFailed, errorMsg); err != nil {
		monitoring.Logger.Error("failed_to_update_failed_status", "task_id", execution.TaskID, "error", err)
	}
	s.sendStreamEvent(execution, constants.StatusFailed, map[string]interface{}{
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

	// Record performance grade for adaptive model selection
	if monitoring.GlobalGradeManager != nil && execution.metadata != nil {
		modelID := s.model // Default model
		if executionModel, ok := execution.metadata["model"]; ok {
			modelID = executionModel
		}

		projectRoot := ""
		if pr, ok := execution.metadata["project_root"]; ok {
			projectRoot = pr
		}

		tokensUsed := int64(0)
		if tokensStr, ok := execution.metadata["total_tokens"]; ok {
			fmt.Sscanf(tokensStr, "%d", &tokensUsed)
		}

		// Record failed completion
		if err := monitoring.GlobalGradeManager.RecordTaskCompletion(
			execution.TaskID,
			modelID,
			execution.Role,
			projectRoot,
			false, // success = false
			0,     // retries (we don't track this yet)
			tokensUsed,
			durationMs,
			false, // wasEscalated (not tracked yet)
			false, // wasDowngraded (not tracked yet)
		); err != nil {
			monitoring.Logger.Warn("failed_to_record_performance_grade", "error", err.Error())
		}
	}
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
		logPath := filepath.Join(execution.ProjectRoot, constants.BeadsDir, "tasks", execution.TaskID, "execution.log")
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

	if err := s.updateTaskStatus(execution.TaskID, execution.ProjectRoot, "cancelled", message); err != nil {
		monitoring.Logger.Error("failed_to_update_cancelled_status", "task_id", execution.TaskID, "error", err)
	}
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
	logDir := filepath.Join(execution.ProjectRoot, constants.BeadsDir, "tasks", execution.TaskID)
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

	cutoff := time.Now().AddDate(0, 0, -constants.ProjectInactiveDays)
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
			if beadsTask.Status == constants.StatusInProgress {
				s.mu.RLock()
				_, hasActiveTask := s.activeTasks[beadsTask.ID]
				s.mu.RUnlock()

				if !hasActiveTask {
					// This is an orphaned task - it's marked in_progress but not running
					// Mark it as "queued" so it can be automatically resumed
					monitoring.Logger.Warn("orphaned_task_detected",
						"task_id", beadsTask.ID,
						"title", beadsTask.Title,
						"project", projectRoot,
						"action", "resetting_to_queued",
					)

					// Reset the task to "queued" status in Beads
					// This allows it to be automatically picked up and resumed
					if beads.IsInstalled() {
						// Use bd update to reset status
						cmd := exec.Command("bd", "update", "--status", "queued", beadsTask.ID)
						cmd.Dir = projectRoot
						if err := cmd.Run(); err != nil {
							monitoring.Logger.Error("failed_to_reset_orphaned_task",
								"task_id", beadsTask.ID,
								"error", err.Error())
						} else {
							monitoring.Logger.Info("orphaned_task_reset",
								"task_id", beadsTask.ID,
								"new_status", "queued")
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
	taskDir := filepath.Join(projectRoot, constants.BeadsDir, "tasks", task.ID)

	// Check if task directory exists
	if _, err := os.Stat(taskDir); os.IsNotExist(err) {
		// Task directory doesn't exist, nothing to archive
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to stat task directory: %w", err)
	}

	// Destination: .beads/archive/<YYYY-MM>/<task-id>/
	now := time.Now()
	archiveDir := filepath.Join(projectRoot, constants.BeadsDir, "archive", now.Format("2006-01"))

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

// updateTaskMetadata updates task metadata with model/provider/tier information
func (s *AgentServer) updateTaskMetadata(projectRoot, taskID, model, provider string, tier int) error {
	// Find task directory
	taskDir := filepath.Join(projectRoot, ".beads", "tasks", taskID)
	metadataPath := filepath.Join(taskDir, constants.MetadataFileName)

	// Read existing metadata
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to read metadata: %w", err)
	}

	// Parse metadata
	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		return fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Update model/provider/tier fields
	if config, ok := metadata["config"].(map[string]interface{}); ok {
		config["Model"] = model
		config["Provider"] = provider
		if tier > 0 {
			config["Tier"] = fmt.Sprintf("tier%d", tier)
		}
	}
	metadata["model"] = model
	metadata["provider"] = provider
	if tier > 0 {
		metadata["tier"] = fmt.Sprintf("tier%d", tier)
	}
	metadata["updated_at"] = time.Now().Format(time.RFC3339)

	// Write back
	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, metadataJSON, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	monitoring.Logger.Info("task_metadata_updated",
		"task_id", taskID,
		"model", model,
		"provider", provider,
		"tier", tier)

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
