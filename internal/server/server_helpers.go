package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/constants"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/internal/protocol"
	"github.com/cortexa-llc/ai-pack/internal/taskdb"
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

// findTaskInDB looks up a task by short ID in the database
func (s *AgentServer) findTaskInDB(shortID, projectRoot string) *taskdb.Task {
	tasks, err := s.taskDB.ListTasks(taskdb.TaskFilter{
		ProjectRoot: projectRoot,
		Limit:       1000,
	})
	if err != nil {
		return nil
	}

	// Find task with matching short ID
	for _, task := range tasks {
		if taskdb.ExtractShortID(task.ID) == shortID {
			return task
		}
	}
	return nil
}

// findTaskPacketPath scans the .ai/tasks directory to find a task packet for the given short ID
func (s *AgentServer) findTaskPacketPath(shortID, projectRoot string) string {
	tasksDir := filepath.Join(projectRoot, constants.TaskRootDir, "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return ""
	}

	// Look for directories starting with the short ID
	prefix := shortID + "-"
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			// Verify it has a 00-contract.md file
			contractPath := filepath.Join(tasksDir, entry.Name(), "00-contract.md")
			if _, err := os.Stat(contractPath); err == nil {
				return filepath.Join(constants.TaskRootDir, "tasks", entry.Name())
			}
		}
	}
	return ""
}

// readTaskDescriptionFromContract reads the task description from 00-contract.md
func (s *AgentServer) readTaskDescriptionFromContract(taskPacketPath, projectRoot string) string {
	contractPath := filepath.Join(projectRoot, taskPacketPath, "00-contract.md")
	data, err := os.ReadFile(contractPath)
	if err != nil {
		return ""
	}

	// Parse the "## Task Description" section
	lines := strings.Split(string(data), "\n")
	inTaskDescription := false
	var descLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Start capturing when we hit "## Task Description"
		if strings.HasPrefix(trimmed, "## Task Description") {
			inTaskDescription = true
			continue
		}

		// Stop at next header or end
		if inTaskDescription {
			if strings.HasPrefix(trimmed, "#") {
				break
			}
			// Skip empty lines at the start
			if len(descLines) == 0 && trimmed == "" {
				continue
			}
			descLines = append(descLines, line)
		}
	}

	// Join and trim
	desc := strings.TrimSpace(strings.Join(descLines, "\n"))

	// Skip template placeholder text
	if strings.Contains(desc, "[Clear, concise description") ||
	   strings.Contains(desc, "[Describe what needs to be done]") {
		return ""
	}

	// Return first 200 chars for display
	if len(desc) > 200 {
		return desc[:200] + "..."
	}
	return desc
}

// applyTaskContractOverrides reads 00-contract.md from the task packet directory
// and applies any Model, Timeout, MaxBudgetTokens, or MaxTurns overrides found there,
// allowing individual tasks to redirect to a different model or adjust limits
// without permanently modifying the role.
//
// Override fields use the same **Key:** value header format as role files.
// Missing fields leave the role value unchanged.
// taskPacketPath is relative to projectRoot (e.g. ".ai/tasks/foo/").
func applyTaskContractOverrides(config *AgentConfig, taskPacketPath, projectRoot string) {
	if taskPacketPath == "" || projectRoot == "" {
		return
	}
	contractPath := filepath.Join(projectRoot, taskPacketPath, "00-contract.md")
	data, err := os.ReadFile(contractPath)
	if err != nil {
		return // contract absent or unreadable — not an error
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, markdownFieldStart) || !strings.Contains(line, markdownFieldEnd) {
			continue
		}
		parts := strings.SplitN(line, markdownFieldEnd, 2)
		if len(parts) != 2 {
			continue
		}
		field := strings.TrimPrefix(parts[0], markdownFieldStart)
		value := strings.TrimSpace(parts[1])

		switch field {
		case configFieldModel:
			if value != "" {
				monitoring.Logger.Info("task_contract_override",
					"field", "Model",
					"role_value", config.Model,
					"contract_value", value,
					"contract", contractPath,
				)
				config.Model = value
			}
		case configFieldTimeout:
			if value != "" {
				monitoring.Logger.Info("task_contract_override",
					"field", "Timeout",
					"role_value", config.Delegation.Timeout,
					"contract_value", value,
					"contract", contractPath,
				)
				config.Delegation.Timeout = value
			}
		case configFieldMaxBudgetTokens:
			var v int
			if n, _ := fmt.Sscanf(value, "%d", &v); n == 1 {
				monitoring.Logger.Info("task_contract_override",
					"field", "MaxBudgetTokens",
					"role_value", config.Delegation.MaxBudgetTokens,
					"contract_value", v,
					"contract", contractPath,
				)
				config.Delegation.MaxBudgetTokens = v
			}
		case configFieldMaxTurns:
			var v int
			if n, _ := fmt.Sscanf(value, "%d", &v); n == 1 {
				monitoring.Logger.Info("task_contract_override",
					"field", "MaxTurns",
					"role_value", config.Delegation.MaxTurns,
					"contract_value", v,
					"contract", contractPath,
				)
				config.Delegation.MaxTurns = v
			}
		}
	}
}

func (s *AgentServer) spawnAgentTask(role, taskInput string, projectRoot string) (*protocol.ExecuteTaskResponse, error) {
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

	// taskInput is the task ID (short form like "ai-pack-abc123")
	taskID := taskInput
	monitoring.Logger.Info("spawning_task", "beads_id", taskID, "project_root", projectRoot)

	// Try to get task description from existing task in database
	taskDescription := fmt.Sprintf("Task %s", taskID)
	taskPacketPath := ""
	workingDir := projectRoot

	// Look up existing task to get metadata (task packet path)
	if s.taskDB != nil {
		if existingTask := s.findTaskInDB(taskID, projectRoot); existingTask != nil {
			// Extract task packet path from metadata
			if existingTask.Metadata != "" {
				var metadata map[string]string
				if err := json.Unmarshal([]byte(existingTask.Metadata), &metadata); err == nil {
					if path := metadata["task_packet_path"]; path != "" {
						taskPacketPath = path
					}
				}
			}
		}
	}

	// If no task packet path in metadata, scan filesystem
	if taskPacketPath == "" {
		taskPacketPath = s.findTaskPacketPath(taskID, projectRoot)
	}

	// Read description from contract if found
	if taskPacketPath != "" {
		if desc := s.readTaskDescriptionFromContract(taskPacketPath, projectRoot); desc != "" {
			taskDescription = desc
		}
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
	// Generate unique task ID
	taskID = taskdb.GenerateTaskID(taskID)

	// Supersede any previous execution for this task so the GUI
	// shows the new run rather than the old failed one.
	if prevFolder := s.findMostRecentExecutionInProject(projectRoot, taskID); prevFolder != "" {
		monitoring.Logger.Info("superseding_previous_execution",
			"task_id", taskID,
			"previous_execution", prevFolder,
			"new_task_id", taskID)
		if err := s.markExecutionAsSuperseded(prevFolder, projectRoot, "rerun"); err != nil {
			monitoring.Logger.Warn("failed_to_supersede_previous_execution",
				"folder", prevFolder, "error", err)
		}
	}

	// Load agent configuration
	config, err := s.loadAgentConfig(role, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load agent config: %w", err)
	}

	// Apply task-level overrides from the contract file (MaxBudgetTokens, MaxTurns).
	// These take precedence over role defaults, allowing per-task budget tuning
	// without changing the role file.
	applyTaskContractOverrides(config, taskPacketPath, projectRoot)

	// Create task packet in project's .beads/tasks/ directory
	if err := s.createTaskPacketInProject(taskID, role, taskDescription, config, projectRoot); err != nil {
		return nil, fmt.Errorf("failed to create task packet: %w", err)
	}

	// Store task packet path, working directory, and project root in metadata if available
	metadata := map[string]string{}
	// CRITICAL: Store the task ID so retry/logs can reference it
	metadata["task_id"] = taskID
	if taskPacketPath != "" {
		metadata["task_packet_path"] = taskPacketPath
	}
	if workingDir != "" {
		metadata["working_directory"] = workingDir
	}
	if projectRoot != "" {
		metadata["project_root"] = projectRoot
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

	// Create task in SQLite database for persistent tracking
	if s.taskDB != nil {
		dbTask := &taskdb.Task{
			ID:              taskID,
			ProjectRoot:     projectRoot,
			Role:            role,
			TaskDescription: taskDescription,
		}
		if err := s.taskDB.CreateTask(dbTask); err != nil {
			monitoring.Logger.Warn("failed_to_create_taskdb_entry", "error", err.Error(), "task_id", taskID)
		}
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

	// Register in activeTasks BEFORE writing metadata to disk.
	// handleOrphanedTasks runs concurrently at startup and checks activeTasks to
	// decide whether a disk-visible in_progress entry is stale. Writing metadata
	// first creates a race window where the orphan scanner can find the file,
	// see no active entry, and incorrectly mark the task as failed.
	s.mu.Lock()
	s.activeTasks[taskID] = execution
	isNewProject := false
	if projectRoot != "" && projectRoot != s.rootDir {
		isNewProject = true
		s.projectRoots[projectRoot] = time.Now()
		s.mu.Unlock()
		// Persist project registry
		if err := s.saveProjectRoots(); err != nil {
			monitoring.Logger.Warn("failed_to_save_project_registry", "error", err.Error())
		}
	} else {
		s.mu.Unlock()
	}

	// Start KG server for newly registered projects now (not lazily in executeAgentWorkflow)
	// so it warms up while the task waits in queue.
	// Also initialize Beads on the shared Dolt server so the project is
	// immediately usable for task tracking.
	if isNewProject {
		go s.ensureKGForProject(projectRoot)
		go s.ensureTaskDirForProject(projectRoot)
	}

	// Write metadata to disk after the task is in activeTasks so that
	// handleOrphanedTasks (running concurrently at startup) cannot observe
	// an in_progress file with no corresponding active entry.
	if err := s.updateTaskPacketMetadataInProject(taskID, metadata, projectRoot); err != nil {
		monitoring.Logger.Warn("failed_to_update_task_metadata", "error", err.Error())
	}

	// Queue for execution using a non-blocking send. If the buffered channel is
	// full the task is rejected immediately and ErrTaskQueueFull is returned to
	// the caller, which translates it into an HTTP 429 response. This avoids
	// blocking the HTTP handler goroutine while the server is at capacity.
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
	// First try taskDB (persistent storage)
	if s.taskDB != nil {
		dbTask, err := s.taskDB.GetTask(taskID)
		if err == nil && dbTask != nil {
			response := &protocol.TaskStatusResponse{
				TaskID:      dbTask.ID,
				Role:        dbTask.Role,
				Task:        dbTask.TaskDescription,
				Status:      dbTask.Status,
				CreatedAt:   dbTask.CreatedAt,
				UpdatedAt:   dbTask.UpdatedAt,
				Result:      dbTask.Result,
				Error:       dbTask.Error,
				CompletedAt: dbTask.CompletedAt,
			}
			return response, nil
		}
	}

	// Fall back to in-memory activeTasks
	s.mu.RLock()
	execution, exists := s.activeTasks[taskID]
	s.mu.RUnlock()

	if !exists {
		// Try loading from disk (legacy)
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

	if response.Status == constants.StatusCompleted || response.Status == constants.StatusFailed {
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
		ExplicitFields: make(map[string]bool),
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

			// Record that this field was explicitly present in the file header.
			config.ExplicitFields[field] = true

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
			case configFieldExtends:
				config.Extends = value
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
			case configFieldSkills:
				// Parse comma-separated skill list
				skillNames := strings.Split(value, ",")
				for _, sn := range skillNames {
					sn = strings.TrimSpace(sn)
					if sn != "" {
						config.Skills = append(config.Skills, sn)
					}
				}
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

// mergeRoleContent merges free-text role content by section headings.
//
// Sections in the project content replace matching sections in the base content.
// Sections absent in the project content are inherited from the base content.
// Sections in the project content that have no matching base section are appended.
//
// A "section" is delimited by a markdown heading line (starting with one or more '#').
func mergeRoleContent(baseContent, projectContent string) string {
	parseSections := func(text string) ([]string, map[string]string) {
		lines := strings.Split(text, "\n")
		var order []string
		sections := make(map[string]string)
		var currentHeading string
		var buf strings.Builder

		for _, line := range lines {
			if strings.HasPrefix(line, "#") {
				// Save previous section
				if currentHeading != "" || buf.Len() > 0 {
					sections[currentHeading] = buf.String()
					if currentHeading != "" {
						order = append(order, currentHeading)
					}
				}
				currentHeading = strings.TrimSpace(line)
				buf.Reset()
				buf.WriteString(line + "\n")
			} else {
				buf.WriteString(line + "\n")
			}
		}
		// Save last section
		if currentHeading != "" || buf.Len() > 0 {
			sections[currentHeading] = buf.String()
			if currentHeading != "" {
				order = append(order, currentHeading)
			}
		}
		return order, sections
	}

	baseOrder, baseSections := parseSections(baseContent)
	projectOrder, projectSections := parseSections(projectContent)

	// Start with base sections, replace with project overrides.
	merged := make(map[string]string)
	for k, v := range baseSections {
		merged[k] = v
	}
	for k, v := range projectSections {
		merged[k] = v
	}

	// Build output: base order first, then project-only sections at the end.
	var out strings.Builder
	seen := make(map[string]bool)
	for _, heading := range baseOrder {
		out.WriteString(merged[heading])
		seen[heading] = true
	}
	for _, heading := range projectOrder {
		if !seen[heading] {
			out.WriteString(merged[heading])
		}
	}

	return strings.TrimRight(out.String(), "\n")
}

// mergeRoleConfigs merges a project-override config (Tier 3b) on top of a base config.
//
// Rules (ADR 006):
//   - Tier: is locked to the base value; project file must not set it.
//   - All scalar fields (Model, Description, Class, Delegation.*) use the project
//     value when the field was explicitly set in the project file; otherwise the
//     base value is kept.
//   - Slice fields (Tools, Gates, Skills, SuccessCriteria) use the project slice
//     when it is non-empty; otherwise the base slice is kept.
//   - Bool fields (ExtendedThinking, ChatTools) use the project value when the
//     field was explicitly set in the project file; otherwise the base value is kept.
//   - Free-text content is merged at the section level via mergeRoleContent.
func mergeRoleConfigs(base, project *AgentConfig) (*AgentConfig, error) {
	// Tier is locked — project file must not set it.
	if project.ExplicitFields[configFieldTier] {
		return nil, fmt.Errorf("role extension error: %q sets Tier: which is locked to the base role; remove Tier: from the project role file", project.Name)
	}

	merged := *base // shallow copy; slices replaced below

	// Scalar fields: use project value when explicitly set.
	if project.ExplicitFields[configFieldDescription] {
		merged.Description = project.Description
	}
	if project.ExplicitFields[configFieldModel] {
		merged.Model = project.Model
	}
	if project.ExplicitFields[configFieldClass] {
		merged.Class = project.Class
	}
	if project.ExplicitFields[configFieldDelegation] {
		merged.Delegation.Mode = project.Delegation.Mode
	}
	if project.ExplicitFields[configFieldTimeout] {
		merged.Delegation.Timeout = project.Delegation.Timeout
	}
	if project.ExplicitFields[configFieldMaxContext] {
		merged.Delegation.MaxContext = project.Delegation.MaxContext
	}
	if project.ExplicitFields[configFieldMaxBudgetTokens] {
		merged.Delegation.MaxBudgetTokens = project.Delegation.MaxBudgetTokens
	}
	if project.ExplicitFields[configFieldMaxTurns] {
		merged.Delegation.MaxTurns = project.Delegation.MaxTurns
	}

	// Bool fields: use project value when explicitly set.
	if project.ExplicitFields[configFieldChatTools] {
		merged.ChatTools = project.ChatTools
	}

	// Slice fields: use project slice when non-empty.
	if len(project.Tools) > 0 {
		merged.Tools = project.Tools
	}
	if len(project.Context.Gates) > 0 {
		merged.Context.Gates = project.Context.Gates
	}
	if len(project.Skills) > 0 {
		merged.Skills = project.Skills
	}
	if len(project.SuccessCriteria) > 0 {
		merged.SuccessCriteria = project.SuccessCriteria
	}

	// Free-text content: section-level merge.
	merged.Context.RoleContent = mergeRoleContent(base.Context.RoleContent, project.Context.RoleContent)

	// Use project role file path for reference.
	merged.Context.RoleFile = project.Context.RoleFile

	// Clear internal-only fields so they are not forwarded to the agent.
	merged.Extends = ""
	merged.ExplicitFields = nil

	return &merged, nil
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

	// Tier 3b: Extends inheritance (ADR 006).
	// Only project-override files (.ai/roles/) may use Extends:.
	if config.Extends != "" {
		if source != "project_override" {
			return nil, fmt.Errorf("role extension error: base role %q contains Extends: which is not permitted (only 1-level deep inheritance is allowed)", role)
		}

		baseName := config.Extends

		// Load base role from non-project paths only (prevents chaining and infinite loops).
		baseData, basePath, err := s.readBaseRoleFile(baseName, projectRoot)
		if err != nil {
			return nil, fmt.Errorf("role extension error: cannot load base role %q for %q: %w", baseName, role, err)
		}

		baseConfig, baseRoleContent, err := parseMarkdownConfig(baseData, baseName)
		if err != nil {
			return nil, fmt.Errorf("role extension error: failed to parse base role %q at %s: %w", baseName, basePath, err)
		}

		// Chain detection: base role must not itself use Extends:.
		if baseConfig.Extends != "" {
			return nil, fmt.Errorf("role extension error: base role %q also has Extends: — only 1-level deep inheritance is allowed", baseName)
		}

		baseConfig.Context.RoleContent = baseRoleContent
		baseConfig.Context.RoleFile = basePath

		merged, err := mergeRoleConfigs(baseConfig, config)
		if err != nil {
			return nil, err
		}
		config = merged
	}

	// Compose skills into the config (ADR 004 Phase 1)
	if err := composeSkills(config, projectRoot); err != nil {
		return nil, fmt.Errorf("skill composition failed for role %s: %w", role, err)
	}

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

func (s *AgentServer) findMostRecentExecutionFolderInRoot(taskID string) string {
	return s.findMostRecentExecutionInProject(s.rootDir, taskID)
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

// readBaseRoleFile loads a base role file from non-project paths only (ADR 006).
// Project override paths (.ai/roles/) are intentionally excluded to prevent chaining.
func (s *AgentServer) readBaseRoleFile(baseName, projectRoot string) ([]byte, string, error) {
	// Search base role in framework and development paths only — never in .ai/roles/.
	candidates := []string{
		filepath.Join(projectRoot, ".ai-pack", "roles", baseName+".md"),
		filepath.Join(s.rootDir, ".ai-pack", "roles", baseName+".md"),
		filepath.Join(s.rootDir, "..", "roles", baseName+".md"),
		filepath.Join(s.rootDir, "roles", baseName+".md"),
	}

	for _, p := range candidates {
		data, err := os.ReadFile(p)
		if err == nil {
			return data, p, nil
		}
	}

	return nil, "", fmt.Errorf("base role %q not found in any framework role path", baseName)
}

// loadAgentConfigForRole loads the agent config from the role .md file
func (s *AgentServer) loadAgentConfigForRole(role string) (*AgentConfig, error) {
	// Check multiple locations for role files
	locations := []string{
		filepath.Join(s.rootDir, "roles", fmt.Sprintf("%s.md", role)),
	}

	// Convention: check for a {role}-chat.md variant first (any role may have one).
	chatLocations := []string{
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
