package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	anthropic_option "github.com/anthropics/anthropic-sdk-go/option"
	"github.com/cortexa-llc/ai-pack/internal/beads"

	"github.com/cortexa-llc/ai-pack/internal/claude"
	"github.com/cortexa-llc/ai-pack/internal/config"
	"github.com/cortexa-llc/ai-pack/internal/constants"
	"github.com/cortexa-llc/ai-pack/internal/execution_log"
	"github.com/cortexa-llc/ai-pack/internal/mcp"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/internal/protocol"
	"github.com/cortexa-llc/ai-pack/internal/streaming"
)

var Version = "dev"
var Commit = "unknown"
var BuildTime = "unknown"

var ErrTaskPaused = errors.New("task paused: token budget exhausted")
var ErrTaskQueueFull = errors.New("task queue full: server is at capacity")

// Configuration field names (for markdown header parsing)
const (
	configFieldAgent           = "Agent"
	configFieldDescription     = "Description"
	configFieldModel           = "Model"
	configFieldTier            = "Tier"
	configFieldTimeout         = "Timeout"
	configFieldMaxContext      = "MaxContext"
	configFieldMaxBudgetTokens = "MaxBudgetTokens"
	configFieldMaxTurns        = "MaxTurns"
	configFieldDelegation      = "Delegation"
	configFieldTools           = "Tools"
	configFieldGates           = "Gates"
	configFieldChatTools       = "ChatTools"
	configFieldSkills          = "Skills"
	configFieldClass           = "Class"
	configFieldExtends         = "Extends"
)

// Configuration defaults
const (
	defaultTier            = "minimal"
	defaultModel           = "" // No model default: grade selector chooses per-role
	defaultDelegation      = "delegate"
	defaultTimeout         = "10min"
	defaultMaxContext      = 32000
	defaultMaxBudgetTokens = 0 // 0 = unlimited
	configSeparator        = "---"
	markdownFieldStart     = "**"
	markdownFieldEnd       = ":**"
)

// defaultRoleTimeout is the fallback used when a role's Timeout field is absent
// or cannot be parsed.
const defaultRoleTimeout = 10 * time.Minute

// taskQueueMultiplier is the ratio of task-queue buffer depth to max concurrency.
// A multiplier of 4 provides burst headroom: if all workers are busy, up to
// 3×maxConcurrent additional submissions can queue without blocking the HTTP
// handler goroutine. Increase if monitoring shows handler goroutines blocking
// under sustained burst load.
const taskQueueMultiplier = 4

// parseRoleTimeout converts a human-friendly timeout string (e.g. "10min", "1h",
// "30sec") to a time.Duration. It first attempts time.ParseDuration (which handles
// standard Go suffixes such as "10m", "1h30m"). If that fails it retries after
// mapping the common long-form suffixes "min" → "m" and "sec" → "s". If parsing
// still fails, defaultRoleTimeout is returned.

type AgentServer struct {
	rootDir          string
	anthropicKey     string
	client           anthropic.Client // SDK v1.19+ returns Client by value, store value to match
	beadsClient      *beads.Client
	claudeSettings   *claude.Settings            // Claude Code settings (deny patterns, etc.)
	executionLog     *execution_log.ExecutionLog // Persistent agent execution log
	maxConcurrent    int                         // Maximum concurrent agents (configurable)
	maxTokens        int                         // Maximum tokens per API call
	model            string                      // Default Anthropic model to use
	maxInactiveTurns         int // Stop agent after N turns without progress
	maxConsecutiveErrorTurns int // Stop agent after N consecutive all-error turns
	config           *config.Config              // Server configuration

	// Multi-provider LLM support
	openaiKey              string
	geminiKey              string
	streamingService       *streaming.Service                       // Clean streaming abstraction
	projectMetrics         map[string]*monitoring.PersistentMetrics // Per-project persistent metrics
	providerCosts          map[string][2]float64                    // Provider cost configuration
	freeProviders          []string                                 // Free/local provider names (savings tracking)
	referenceModel         string                                   // Reference model for savings calculations
	mcpManager             *mcp.Manager                             // MCP server manager
	complexityRiskAnalyzer *monitoring.ComplexityRiskAnalyzer       // v2 structural risk scorer

	// Concurrent execution tracking
	mu               sync.RWMutex
	activeTasks      map[string]*TaskExecution
	taskQueue        chan *TaskExecution
	workerPool       chan struct{}             // Semaphore for max concurrent agents
	projectRoots     map[string]time.Time     // Registry of known project roots with last access time
	kgPreflightHits  map[string]int64         // KG preflight context injections per project root

	// Server-level context — cancelled by Shutdown() to pre-empt in-flight tasks.
	ctx    context.Context
	cancel context.CancelFunc
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

	// Performance grading fields (populated during executeAgenticLoop)
	RetryCount    int  // number of agentic-loop turns after the first
	WasEscalated  bool // model tier was raised above the role default
	WasDowngraded bool // model tier was lowered below the role default
}

type AgentConfig struct {
	Name        string
	Description string
	Tier        string
	Class       string // Model class filter: "agentic", "completion", "reasoning" (empty = no filter)
	Model       string // LLM model to use (e.g., "gpt-4o-mini", "claude-sonnet-3-5-20241022")
	Context     struct {
		RoleFile               string
		RoleContent            string // Loaded from .md file content
		Gates                  []string
		AdditionalInstructions string
	}
	Delegation struct {
		Mode            string
		Timeout         string
		MaxContext      int
		MaxBudgetTokens int // 0 = unlimited
		MaxTurns        int // 0 = unlimited
	}
	Tools            []string
	SuccessCriteria  []string
	Metadata         map[string]interface{}
	ExtendedThinking bool
	ChatTools        bool     // If true, inject chat-mode tools (spawn_agent, query_tasks, etc.)
	Skills           []string // Skill names declared in the role file (ADR 004)
	SkillsLoaded     []string // Skill names successfully composed at spawn (ADR 004)

	// Extends: name of base role this project-role file inherits from (ADR 006, Tier 3b).
	// Empty means full substitution (Tier 3a). Consumed during loadAgentConfig; not forwarded.
	Extends string

	// ExplicitFields tracks which config header fields were explicitly present in the source
	// file. Used during Extends merge to distinguish "not set" from "set to zero/false/empty".
	// This field is internal and is not forwarded to the agent.
	ExplicitFields map[string]bool
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
				"provider":     usage.Provider,
				"model":        usage.Model,
				"calls":        usage.Calls,
				"inputTokens":  usage.InputTokens,
				"outputTokens": usage.OutputTokens,
				"cost":         usage.Cost,
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

	pm, err := monitoring.NewPersistentMetrics(projectRoot, s.providerCosts, s.freeProviders, s.referenceModel)
	if err != nil {
		return nil, err
	}

	s.projectMetrics[projectRoot] = pm
	return pm, nil
}

func NewAgentServer(rootDir string, maxConcurrent int, maxTokens int, model string, cfg *config.Config) (*AgentServer, error) {
	anthropicKey := os.Getenv(cfg.ProviderAPIKeyEnv("anthropic", "ANTHROPIC_API_KEY"))
	openaiKey := os.Getenv(cfg.ProviderAPIKeyEnv("openai", "OPENAI_API_KEY"))
	geminiKey := os.Getenv(cfg.ProviderAPIKeyEnv("gemini", "GEMINI_API_KEY"))

	// Qwen local provider — endpoint configurable via providers.qwen.endpoint or
	// QWEN_BASE_URL env var, defaulting to localhost:9000. API key env var
	// configurable via providers.qwen.api_key_env, defaulting to LLAMA_API_KEY.
	qwenBaseURL := cfg.ProviderEndpoint("qwen")
	if qwenBaseURL == "" {
		qwenBaseURL = os.Getenv("QWEN_BASE_URL")
	}
	if qwenBaseURL == "" {
		qwenBaseURL = constants.QwenLocalBaseURL
	}
	qwenAPIKey := os.Getenv(cfg.ProviderAPIKeyEnv("qwen", "LLAMA_API_KEY")) // LM Studio / llama.cpp compatible key

	anthropicOpts := []anthropic_option.RequestOption{
		anthropic_option.WithRequestTimeout(10 * time.Second),
	}

	if anthropicKey != "" {
		anthropicOpts = append(anthropicOpts, anthropic_option.WithAPIKey(anthropicKey))
	}

	anthropicClient := anthropic.NewClient(anthropicOpts...)

	serverCtx, serverCancel := context.WithCancel(context.Background())

	server := &AgentServer{
		rootDir:        rootDir,
		maxConcurrent:  maxConcurrent,
		maxTokens:      maxTokens,
		model:          model,
		config:         cfg,
		client:         anthropicClient,
		projectMetrics: make(map[string]*monitoring.PersistentMetrics),
		providerCosts:  make(map[string][2]float64),
		activeTasks:    make(map[string]*TaskExecution),
		taskQueue:      make(chan *TaskExecution, maxConcurrent*taskQueueMultiplier),
		workerPool:     make(chan struct{}, maxConcurrent),
		projectRoots:    make(map[string]time.Time),
		kgPreflightHits: make(map[string]int64),
		ctx:            serverCtx,
		cancel:         serverCancel,
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
	maxInactiveTurns := 0
	if cfg != nil {
		maxInactiveTurns = cfg.Agent.MaxInactiveTurns
	}
	if maxInactiveTurns == 0 {
		maxInactiveTurns = 10 // Default fallback
	}

	// Get max consecutive error turns from config; fall back to maxInactiveTurns.
	maxConsecutiveErrorTurns := 0
	if cfg != nil {
		maxConsecutiveErrorTurns = cfg.Agent.MaxConsecutiveErrorTurns
	}
	if maxConsecutiveErrorTurns == 0 {
		maxConsecutiveErrorTurns = maxInactiveTurns
	}

	// Initialize execution log
	execLog, err := execution_log.NewExecutionLog(rootDir)
	if err != nil {
		monitoring.Logger.Warn("failed_to_create_execution_log", "error", err.Error())
	}

	// Build cost map from config for per-project metrics
	costs := make(map[string][2]float64)
	if cfg != nil {
		for modelKey, modelCost := range cfg.ProviderCosts.Models {
			costs[modelKey] = [2]float64{modelCost.InputCost, modelCost.OutputCost}
		}
	}

	// Initialize performance grading system.
	// Grades are stored in DataDir (home-relative) so the path is identical
	// whether the binary runs from /usr/local/bin or a dev checkout.
	if cfg != nil {
		dataDir, err := config.DataDir()
		if err != nil {
			monitoring.Logger.Warn("failed_to_resolve_data_dir", "error", err.Error())
		} else {
			gradesDir := filepath.Join(dataDir, "performance_grades")
			if err := monitoring.InitGradeManager(gradesDir, &cfg.GradingCriteria); err != nil {
				monitoring.Logger.Warn("failed_to_init_grade_manager", "error", err.Error())
			} else {
				monitoring.Logger.Info("performance_grading_initialized", "grades_dir", gradesDir)
			}
		}
	}

	// Initialize complexity analyzer
	monitoring.InitComplexityAnalyzer()

	// Initialize v2 structural risk analyzer
	if cfg != nil {
		gcfg := cfg.ComplexityGate
		server.complexityRiskAnalyzer = monitoring.NewComplexityRiskAnalyzer(
			gcfg.Enabled,
			gcfg.WarnThreshold,
			gcfg.CriticalThreshold,
			gcfg.Weights.Scope,
			gcfg.Weights.MultiStep,
			gcfg.Weights.Uncertainty,
			gcfg.Weights.Structural,
			gcfg.RoleMultipliers,
			monitoring.GlobalGradeManager,
		)
	} else {
		// Fallback: enabled with defaults
		server.complexityRiskAnalyzer = monitoring.NewComplexityRiskAnalyzer(
			true, 0.50, 0.75, 0.30, 0.25, 0.25, 0.20,
			map[string]float64{},
			nil,
		)
	}

	// Initialize model selector
	if monitoring.GlobalGradeManager != nil && monitoring.GlobalComplexityAnalyzer != nil {
		monitoring.InitModelSelector(monitoring.GlobalGradeManager, monitoring.GlobalComplexityAnalyzer, openaiKey != "", geminiKey != "")
		adaptiveEnabled := cfg == nil || cfg.API.AdaptiveModelSelection
		if !adaptiveEnabled {
			monitoring.GlobalModelSelector.SetEnabled(false)
			monitoring.Logger.Info("adaptive_model_selection_disabled", "reason", "config adaptive_model_selection=false")
		} else {
			monitoring.Logger.Info("adaptive_model_selection_enabled")
		}
	}

	// Populate server fields that require pre-computed values
	server.anthropicKey = anthropicKey
	server.openaiKey = openaiKey
	server.geminiKey = geminiKey
	server.beadsClient = beads.NewClient()
	server.claudeSettings = claudeSettings
	server.executionLog = execLog
	server.maxInactiveTurns = maxInactiveTurns
	server.maxConsecutiveErrorTurns = maxConsecutiveErrorTurns
	server.providerCosts = costs
	if cfg != nil {
		server.freeProviders = cfg.ProviderCosts.FreeProviders
		server.referenceModel = cfg.ProviderCosts.ReferenceModel
	}
	if len(server.freeProviders) == 0 {
		server.freeProviders = []string{"qwen"}
	}
	if server.referenceModel == "" {
		server.referenceModel = "claude-sonnet-4-6"
	}

	// Initialize streaming service with performance-grade-based model selection
	streamingSelector := streaming.NewPerformanceGradeModelSelector(
		rootDir,
		model,
		openaiKey != "",
		geminiKey != "",
	)
	var modelTranslations map[string]map[string]string
	if cfg != nil {
		modelTranslations = cfg.ModelTranslations
	}
	streamingService := streaming.NewService(streamingSelector, constants.ProviderAnthropic, modelTranslations)

	// Register provider factories
	streamingService.RegisterProvider(streaming.NewAnthropicFactory(anthropicKey, maxTokens))
	if openaiKey != "" {
		streamingService.RegisterProvider(streaming.NewOpenAIFactory(openaiKey))
	}
	if geminiKey != "" {
		streamingService.RegisterProvider(streaming.NewGeminiFactory(geminiKey, maxTokens))
	}
	streamingService.RegisterProvider(streaming.NewQwenFactory(qwenBaseURL, qwenAPIKey))

	server.streamingService = streamingService
	monitoring.Logger.Info("streaming_service_initialized",
		"default", constants.ProviderAnthropic)

	// Initialize MCP manager if enabled
	server.mcpManager = mcp.NewManager()
	if cfg != nil && cfg.MCP.Enabled {
		// Load MCP servers from user/project configs
		mcpServers, err := config.LoadMCPServers(&cfg.MCP, rootDir)
		if err != nil {
			monitoring.Logger.Warn("failed_to_load_mcp_servers", "error", err.Error())
		} else {
			// Start MCP servers
			for name, serverCfg := range mcpServers {
				mcpCfg := mcp.ServerConfig{
					Command: serverCfg.Command,
					Args:    serverCfg.Args,
					Env:     serverCfg.Env,
				}

				if err := server.mcpManager.StartServer(context.Background(), name, mcpCfg); err != nil {
					monitoring.Logger.Warn("failed_to_start_mcp_server",
						"server", name,
						"error", err.Error())
				} else {
					monitoring.Logger.Info("mcp_server_started",
						"server", name,
						"command", serverCfg.Command)
				}
			}

			// Log active servers and available tools
			activeServers := server.mcpManager.GetActiveServers()
			if len(activeServers) > 0 {
				toolCount := 0
				for _, tools := range server.mcpManager.GetAllTools() {
					toolCount += len(tools)
				}
				monitoring.Logger.Info("mcp_integration_enabled",
					"active_servers", len(activeServers),
					"total_tools", toolCount,
					"servers", strings.Join(activeServers, ", "))
			}
		}
	} else {
		monitoring.Logger.Info("mcp_integration_disabled")
	}

	// Start KG server for the server root directory
	server.ensureKGForProject(rootDir)

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
	if cfg != nil && cfg.TaskCleanup.Enabled && beads.IsInstalled() {
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

func (s *AgentServer) loadProjectRoots() error {
	if s.config.Projects == nil {
		s.config.Projects = make(map[string]string)
		return nil
	}

	var loadedPaths []string

	s.mu.Lock()
	for path, lastAccessedStr := range s.config.Projects {
		lastAccessed, err := time.Parse(time.RFC3339, lastAccessedStr)
		if err != nil {
			continue // Skip invalid entries
		}
		s.projectRoots[path] = lastAccessed
		loadedPaths = append(loadedPaths, path)
	}
	s.mu.Unlock()

	monitoring.Logger.Info("project_registry_loaded", "count", len(loadedPaths))

	// Start KG servers for all known project roots (outside lock to avoid holding
	// s.mu while spawning subprocesses).
	for _, path := range loadedPaths {
		s.ensureKGForProject(path)
	}

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
			if beadsTask.Status == constants.StatusPaused {
				continue
			}
			if beadsTask.Status == constants.StatusInProgress {
				s.mu.RLock()
				_, hasActiveTask := s.activeTasks[beadsTask.ID]
				s.mu.RUnlock()

				if !hasActiveTask {
					// This is an orphaned task - marked in_progress in beads but no active execution.
					// Reset to "open" so it shows as queued (not stuck as running) in the GUI.
					monitoring.Logger.Warn("orphaned_task_detected",
						"task_id", beadsTask.ID,
						"title", beadsTask.Title,
						"project", projectRoot,
						"action", "resetting_to_open",
					)

					// "bd update -s open" is the correct reset; "queued" is not a valid beads status.
					// Reset Beads status first — if this fails, leave checkpoint intact so
					// the task remains in a recoverable state (checkpoint present, status
					// still in_progress → next sweep will retry).
					if beads.IsInstalled() {
						cmd := exec.Command("bd", "update", "-s", "open", beadsTask.ID)
						cmd.Dir = projectRoot
						if err := cmd.Run(); err != nil {
							monitoring.Logger.Error("failed_to_reset_orphaned_task",
								"task_id", beadsTask.ID,
								"error", err.Error())
							continue // leave checkpoint intact; retry on next sweep
						}
						monitoring.Logger.Info("orphaned_task_reset",
							"task_id", beadsTask.ID,
							"new_status", "open")
					}

					// Delete any checkpoint only after a successful Beads reset, so the
					// task restarts from scratch instead of resuming stale state from the
					// previous (aborted) run. If we left the checkpoint in place, the next
					// resume would replay old messages into a freshly-opened task and
					// produce incorrect behaviour.
					cpPath := checkpointPath(projectRoot, beadsTask.ID)
					if removeErr := os.Remove(cpPath); removeErr == nil {
						monitoring.Logger.Info("orphaned_task_checkpoint_deleted",
							"task_id", beadsTask.ID,
							"checkpoint", cpPath)
					} else if !os.IsNotExist(removeErr) {
						monitoring.Logger.Warn("failed_to_delete_orphaned_checkpoint",
							"task_id", beadsTask.ID,
							"checkpoint", cpPath,
							"error", removeErr.Error())
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

	// Second pass: scan execution folders for stale in_progress metadata.
	// This catches tasks that were interrupted mid-run (e.g. server killed) whose
	// Beads status was already reset to "open" by bd reopen, but whose execution
	// metadata file still says "in_progress".
	staleMeta := 0
	for _, projectRoot := range projectRoots {
		tasksDir := filepath.Join(projectRoot, constants.BeadsDir, "tasks")
		entries, err := os.ReadDir(tasksDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			folderName := entry.Name()
			metadataPath := filepath.Join(tasksDir, folderName, constants.MetadataFileName)
			data, err := os.ReadFile(metadataPath)
			if err != nil {
				continue
			}
			var meta map[string]interface{}
			if json.Unmarshal(data, &meta) != nil {
				continue
			}
			if meta["status"] != "in_progress" {
				continue
			}
			// Skip folders spawned very recently — they may belong to a task that was
			// running in the previous server process which hasn't fully exited yet.
			// launchctl starts the new process immediately on restart, so there is a
			// race window (up to the 5-second graceful-shutdown timeout) where both the
			// old and new processes are alive simultaneously.  Marking a task failed here
			// while the old process is still executing it causes the stale-failed UX bug.
			if spawnedStr, ok := meta["spawned_at"].(string); ok {
				if spawned, err := time.Parse(time.RFC3339, spawnedStr); err == nil {
					if time.Since(spawned) < 30*time.Second {
						monitoring.Logger.Info("skipping_recently_spawned_execution",
							"folder", folderName,
							"spawned_at", spawnedStr,
							"age_seconds", int(time.Since(spawned).Seconds()),
						)
						continue
					}
				}
			}
			// Not active — mark it failed so the GUI doesn't show it as running
			s.mu.RLock()
			_, isActive := s.activeTasks[folderName]
			s.mu.RUnlock()
			if isActive {
				continue
			}
			meta["status"] = constants.StatusFailed
			meta["error"] = "interrupted: server restarted while task was running"
			meta["updated_at"] = time.Now().Format(time.RFC3339)
			if updated, err := json.MarshalIndent(meta, "", "  "); err == nil {
				if writeErr := os.WriteFile(metadataPath, updated, 0644); writeErr == nil {
					monitoring.Logger.Info("stale_execution_marked_failed",
						"folder", folderName,
						"project", projectRoot,
					)
					staleMeta++
				}
			}
		}
	}

	if orphanedCount > 0 || staleCount > 0 || staleMeta > 0 {
		monitoring.Logger.Info("task_cleanup_summary",
			"orphaned", orphanedCount,
			"stale_removed", staleCount,
			"stale_metadata_fixed", staleMeta,
		)
	}
}

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

// ensureKGForProject starts a KG MCP server for the given project root if one
// is not already running. It is idempotent and safe to call concurrently.
func (s *AgentServer) ensureKGForProject(projectRoot string) {
	if s.mcpManager == nil {
		return
	}
	cfg := mcp.ServerConfig{
		Command: "kg",
		Args:    []string{"server", "--stdio"},
		Dir:     projectRoot,
	}
	if err := s.mcpManager.EnsureProjectServer(context.Background(), projectRoot, cfg); err != nil {
		monitoring.Logger.Warn("kg_server_start_failed", "project", projectRoot, "err", err.Error())
	} else {
		monitoring.Logger.Info("kg_server_ready", "project", projectRoot)
	}
}

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

func (s *AgentServer) Shutdown(ctx context.Context) error {
	monitoring.Logger.Info("shutdown_initiated")

	// Cancel the server context to pre-empt any in-flight tasks.
	s.cancel()

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

	// Shutdown MCP servers
	if s.mcpManager != nil {
		if err := s.mcpManager.Close(); err != nil {
			monitoring.Logger.Warn("mcp_shutdown_error", "error", err.Error())
		} else {
			monitoring.Logger.Info("mcp_servers_closed")
		}
	}

	monitoring.Logger.Info("shutdown_complete", "active_tasks", 0)
	return nil
}
