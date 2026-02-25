package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/beads"

	"github.com/cortexa-llc/ai-pack/internal/constants"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/internal/protocol"
)

func parseRoleTimeout(s string) time.Duration {
	if s == "" {
		return defaultRoleTimeout
	}
	// Try standard Go format first ("10m", "1h", "30s", etc.).
	if d, err := time.ParseDuration(s); err == nil && d > 0 {
		return d
	}
	// Normalise common long-form suffixes.
	normalised := strings.NewReplacer("min", "m", "sec", "s").Replace(s)
	if d, err := time.ParseDuration(normalised); err == nil && d > 0 {
		return d
	}
	return defaultRoleTimeout
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

	// Sanitize projectRoot: reject relative paths (including traversal attempts).
	if projectRoot != "" {
		cleanedRoot := filepath.Clean(projectRoot)
		if !filepath.IsAbs(cleanedRoot) {
			return nil, fmt.Errorf("project_root must be an absolute path, got: %q", projectRoot)
		}
		projectRoot = cleanedRoot
	}

	// Sanitize role: reject names containing path separators to prevent traversal.
	if strings.ContainsAny(role, "/\\") {
		return nil, fmt.Errorf("role must not contain path separators, got: %q", role)
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

	// v2 structural risk assessment (all roles)
	riskAssessment := s.complexityRiskAnalyzer.ComputeComplexityRisk(role, taskDescription)
	monitoring.Logger.Info("complexity_risk_assessed",
		"role", role,
		"risk_level", string(riskAssessment.RiskLevel),
		"adjusted_score", riskAssessment.AdjustedScore,
		"base_score", riskAssessment.BaseScore,
	)
	if riskAssessment.ShouldWarn {
		monitoring.Logger.Warn("complexity_risk_high",
			"role", role,
			"risk_level", string(riskAssessment.RiskLevel),
			"recommendation", riskAssessment.Recommendation,
		)
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

	// Queue for execution — non-blocking: return immediately if queue is full
	select {
	case s.taskQueue <- execution:
	default:
		s.mu.Lock()
		delete(s.activeTasks, taskID)
		s.mu.Unlock()
		return nil, ErrTaskQueueFull
	}

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

	response := &protocol.ExecuteTaskResponse{
		TaskID:    taskID,
		Status:    constants.StatusQueued,
		Message:   fmt.Sprintf("Agent %s queued for execution. Task ID: %s", role, taskID),
		StreamURL: streamURL,
		CreatedAt: time.Now(),
	}

	// Attach complexity warning to the response when the pre-spawn gate triggered or v2 risk is high.
	if riskAssessment.ShouldWarn {
		warning := &protocol.ComplexityWarning{
			RiskLevel: string(riskAssessment.RiskLevel),
			Components: &protocol.RiskComponents{
				ScopeScore:       riskAssessment.Components.ScopeScore,
				MultiStepScore:   riskAssessment.Components.MultiStepScore,
				UncertaintyScore: riskAssessment.Components.UncertaintyScore,
				StructuralScore:  riskAssessment.Components.StructuralScore,
				HistoricalScore:  riskAssessment.Components.HistoricalScore,
				RoleMultiplier:   riskAssessment.Components.RoleMultiplier,
			},
		}
		if riskAssessment.Recommendation != "" {
			warning.Recommendation = riskAssessment.Recommendation
		}
		response.ComplexityWarning = warning
	}

	return response, nil
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

// parseMarkdownConfig extracts configuration from structured markdown headers
// Expected format at top of file:
// # Role Name
// **Agent:** engineer
// **Description:** Implementation specialist
// **Model:** gpt-4o-mini
// **Tier:** minimal
// **Timeout:** 10min
// **Tools:** read, write, edit, bash
// **Gates:** tdd-enforcement
// **Delegation:** delegate
// **MaxContext:** 32000
// ---
//
// Returns: config, roleContent, error

func parseMarkdownConfig(data []byte, roleName string) (*AgentConfig, string, error) {
	content := string(data)
	lines := strings.Split(content, "\n")

	config := &AgentConfig{
		Name: roleName,
		// Apply defaults
		Tier:  defaultTier,
		Model: defaultModel,
		Delegation: struct {
			Mode            string
			Timeout         string
			MaxContext      int
			MaxBudgetTokens int // 0 = unlimited
			MaxTurns        int // 0 = unlimited
		}{
			Mode:            defaultDelegation,
			Timeout:         defaultTimeout,
			MaxContext:      defaultMaxContext,
			MaxBudgetTokens: defaultMaxBudgetTokens,
		},
	}

	// Find the separator (---)
	separatorIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == configSeparator {
			separatorIdx = i
			break
		}
	}

	if separatorIdx == -1 {
		return nil, "", fmt.Errorf("missing %s separator after config header (required format: see role file documentation)", configSeparator)
	}

	// Parse header section (before ---)
	for i := 0; i < separatorIdx; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and markdown headers
		}

		// Parse **Field:** value format
		if strings.HasPrefix(line, markdownFieldStart) && strings.Contains(line, markdownFieldEnd) {
			parts := strings.SplitN(line, markdownFieldEnd, 2)
			if len(parts) != 2 {
				continue
			}

			field := strings.TrimPrefix(parts[0], markdownFieldStart)
			value := strings.TrimSpace(parts[1])

			switch field {
			case configFieldAgent:
				config.Name = value
			case configFieldDescription:
				config.Description = value
			case configFieldModel:
				config.Model = value
			case configFieldTier:
				config.Tier = value
				// Register tier with global model selector so grade-based selection
				// starts at the role-specified tier rather than the TierLow default.
				if tier := monitoring.ParseTierString(value); tier > 0 && monitoring.GlobalModelSelector != nil {
					monitoring.GlobalModelSelector.SetRoleDefaultTier(roleName, tier)
				}
			case configFieldClass:
				config.Class = value
				// Register class filter so only models of this class are eligible.
				if class := monitoring.ParseClassString(value); class != "" && monitoring.GlobalModelSelector != nil {
					monitoring.GlobalModelSelector.SetRoleRequiredClass(roleName, class)
				}
			case configFieldTimeout:
				config.Delegation.Timeout = value
			case configFieldMaxContext:
				fmt.Sscanf(value, "%d", &config.Delegation.MaxContext)
			case configFieldMaxBudgetTokens:
				fmt.Sscanf(value, "%d", &config.Delegation.MaxBudgetTokens)
			case configFieldMaxTurns:
				fmt.Sscanf(value, "%d", &config.Delegation.MaxTurns)
			case configFieldDelegation:
				config.Delegation.Mode = value
			case configFieldTools:
				// Parse comma-separated list
				tools := strings.Split(value, ",")
				for _, tool := range tools {
					tool = strings.TrimSpace(tool)
					if tool != "" {
						config.Tools = append(config.Tools, tool)
					}
				}
			case configFieldGates:
				// Parse comma-separated list
				gates := strings.Split(value, ",")
				for _, gate := range gates {
					gate = strings.TrimSpace(gate)
					if gate != "" {
						config.Context.Gates = append(config.Context.Gates, gate)
					}
				}
			case configFieldChatTools:
				config.ChatTools = strings.EqualFold(value, "true")
			}
		}
	}

	// Extract role content (after ---)
	roleContent := strings.Join(lines[separatorIdx+1:], "\n")
	roleContent = strings.TrimSpace(roleContent)

	// Validation - required fields
	if config.Name == "" {
		return nil, "", fmt.Errorf("missing required field: %s%s%s", markdownFieldStart, configFieldAgent, markdownFieldEnd)
	}
	if config.Description == "" {
		return nil, "", fmt.Errorf("missing required field: %s%s%s", markdownFieldStart, configFieldDescription, markdownFieldEnd)
	}

	return config, roleContent, nil
}

func (s *AgentServer) loadAgentConfig(role string, projectRoot string) (*AgentConfig, error) {
	// Use provided projectRoot or fall back to server root
	if projectRoot == "" {
		projectRoot = s.rootDir
	}

	// Try paths in order:
	// 1. Project override (.ai/roles/)
	// 2. Framework (.ai-pack/roles/) - production
	// 3. Development (../roles/) - when running from a2a-agent dir
	// 4. Development (roles/) - when running from repo root
	// Note: Using .md files with YAML frontmatter (single source of truth)
	candidatePaths := []struct {
		path   string
		source string
	}{
		{filepath.Join(projectRoot, ".ai", "roles", role+".md"), "project_override"},
		{filepath.Join(projectRoot, ".ai-pack", "roles", role+".md"), "framework"},
		{filepath.Join(projectRoot, "..", "roles", role+".md"), "dev_parent"},
		{filepath.Join(projectRoot, "roles", role+".md"), "dev_root"},
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
		return nil, fmt.Errorf("no config found for role %s (tried: .ai/roles, .ai-pack/roles, ../roles, roles)", role)
	}

	monitoring.Logger.Info("loading_agent_config", "role", role, "source", source, "path", configPath)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Parse pure markdown configuration
	config, roleContent, err := parseMarkdownConfig(data, role)
	if err != nil {
		return nil, fmt.Errorf("failed to parse role config %s: %w (ensure file has required markdown header format)", configPath, err)
	}

	// Store role content and path in config for later use
	config.Context.RoleContent = roleContent
	config.Context.RoleFile = configPath // Store actual path for reference

	return config, nil
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

func (s *AgentServer) findMostRecentExecutionFolderInRoot(beadsTaskID string) string {
	return s.findMostRecentExecutionInProject(s.rootDir, beadsTaskID)
}

func extractContractSections(content string) string {
	// Sections we want to include
	keep := []string{
		"## Task Description",
		"## Success Criteria",
		"## Background",
	}

	lines := strings.Split(content, "\n")
	var result []string
	inKeptSection := false
	sectionHasContent := false
	var sectionLines []string

	flushSection := func() {
		if inKeptSection && sectionHasContent {
			result = append(result, sectionLines...)
		}
		sectionLines = nil
		sectionHasContent = false
	}

	for _, line := range lines {
		// Check if this line starts a new section heading
		if strings.HasPrefix(line, "## ") {
			flushSection()
			inKeptSection = false
			for _, k := range keep {
				if strings.HasPrefix(line, k) {
					inKeptSection = true
					break
				}
			}
			if inKeptSection {
				sectionLines = []string{line}
			}
			continue
		}

		if !inKeptSection {
			continue
		}

		// Skip lines that are pure boilerplate placeholders
		trimmed := strings.TrimSpace(line)
		isPlaceholder := trimmed == "" ||
			(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) ||
			strings.HasPrefix(trimmed, "□ [") ||
			strings.HasPrefix(trimmed, "✗ [") ||
			strings.HasPrefix(trimmed, "✓ [") ||
			trimmed == "```" ||
			trimmed == "---"

		sectionLines = append(sectionLines, line)
		if !isPlaceholder {
			sectionHasContent = true
		}
	}
	flushSection()

	return strings.TrimSpace(strings.Join(result, "\n"))
}

func (s *AgentServer) loadSharedPolicy(projectRoot string) string {
	if projectRoot == "" {
		projectRoot = s.rootDir
	}
	candidates := []string{
		filepath.Join(projectRoot, ".ai", "roles", "shared", "agent-policy.md"),
		filepath.Join(projectRoot, ".ai-pack", "roles", "shared", "agent-policy.md"),
		filepath.Join(projectRoot, "..", "roles", "shared", "agent-policy.md"),
		filepath.Join(projectRoot, "roles", "shared", "agent-policy.md"),
	}
	for _, path := range candidates {
		if data, err := os.ReadFile(path); err == nil {
			return string(data)
		}
	}
	return ""
}

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

// GetModelForRole returns the model to use for a given role
// Loads the agent config and returns the model field, or default if not specified
func (s *AgentServer) GetModelForRole(role string) string {
	if role == "" {
		return s.model // Default model
	}

	// Try to load agent config
	cfg, err := s.loadAgentConfigForRole(role)
	if err != nil || cfg.Model == "" {
		// No model specified in config, use default
		return s.model
	}

	return cfg.Model
}

// loadAgentConfigForRole loads the agent config from the role .md file
func (s *AgentServer) loadAgentConfigForRole(role string) (*AgentConfig, error) {
	// Check multiple locations for role files
	locations := []string{
		filepath.Join(s.rootDir, ".ai-pack", "agents", fmt.Sprintf("%s.md", role)),
		filepath.Join(s.rootDir, "agents", fmt.Sprintf("%s.md", role)),
		filepath.Join(s.rootDir, "roles", fmt.Sprintf("%s.md", role)),
	}

	// Convention: check for a {role}-chat.md variant first (any role may have one).
	chatLocations := []string{
		filepath.Join(s.rootDir, ".ai-pack", "agents", fmt.Sprintf("%s-chat.md", role)),
		filepath.Join(s.rootDir, "roles", fmt.Sprintf("%s-chat.md", role)),
	}
	locations = append(chatLocations, locations...)

	for _, path := range locations {
		if _, err := os.Stat(path); err == nil {
			return s.parseAgentConfig(path)
		}
	}

	return nil, fmt.Errorf("agent config not found for role: %s", role)
}

// parseAgentConfig parses the markdown configuration from an agent .md file
func (s *AgentServer) parseAgentConfig(path string) (*AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Extract role name from filename
	roleName := filepath.Base(path)
	roleName = roleName[:len(roleName)-3] // Remove .md extension

	// Parse using markdown config parser
	cfg, roleContent, err := parseMarkdownConfig(data, roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse markdown config from %s: %w", path, err)
	}

	cfg.Context.RoleContent = roleContent
	cfg.Context.RoleFile = path

	return cfg, nil
}
