package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/constants"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/internal/protocol"
)

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
	// Search across all registered project roots (not just server root).
	// Deduplicate by canonical beads root to avoid searching a2a-agent/ when
	// the real .beads/beads.db is at the parent project root.
	rawRoots := s.GetProjectRoots()
	seen := make(map[string]bool)
	var projectRoots []string
	for _, r := range rawRoots {
		canonical := resolveBeadsRoot(r)
		if !seen[canonical] {
			seen[canonical] = true
			projectRoots = append(projectRoots, canonical)
		}
	}

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

	var createdAt time.Time
	if spawnedAt, ok := metadata["spawned_at"].(string); ok && spawnedAt != "" {
		createdAt, _ = time.Parse(time.RFC3339, spawnedAt)
	}
	if createdAt.IsZero() {
		// Fall back to parsing the timestamp from the execution folder name.
		// Format: {beads-id}-YYYYMMDD-HHMMSS  e.g. xasm++-qbxv-20260218-084509
		parts := strings.Split(executionFolder, "-")
		if len(parts) >= 2 {
			lastPart := parts[len(parts)-1]
			secondLastPart := parts[len(parts)-2]
			if len(lastPart) == 6 && len(secondLastPart) == 8 {
				createdAt, _ = time.Parse("20060102-150405", secondLastPart+"-"+lastPart)
			}
		}
	}

	var updatedAt time.Time
	if updatedAtStr, ok := metadata["updated_at"].(string); ok && updatedAtStr != "" {
		updatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
	}
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}

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

	// Status reconciliation is no longer needed - taskDB is the source of truth

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

func (s *AgentServer) buildPrompt(role, task, roleContext string, config *AgentConfig, taskPacketPath, workingDir string) string {
	// If a contract exists, extract its key sections so we can place them prominently.
	contractSections := ""
	if taskPacketPath != "" {
		contractPath := filepath.Join(workingDir, taskPacketPath, "00-contract.md")
		if contractData, err := os.ReadFile(contractPath); err == nil {
			contractSections = extractContractSections(string(contractData))
		}
	}

	// Build the prompt with contract sections FIRST so the agent sees explicit steps
	// before the general task description. This prevents the agent from defaulting to
	// exploration mode when explicit commands are provided.
	var prompt string
	if contractSections != "" {
		prompt = fmt.Sprintf(`You are a %s agent.

**Working Directory:** %s
**Task Packet:** %s

All file operations must use paths relative to the working directory above.

---

%s

---

**Task:** %s`,
			role,
			workingDir,
			taskPacketPath,
			contractSections,
			task)
	} else {
		prompt = fmt.Sprintf(`You are a %s agent.

**Your Task:**
%s

**Working Directory:**
%s

All file operations (Read, Write, Edit, Glob, Grep, Bash) must be performed relative to the working directory above.`,
			role,
			task,
			workingDir)

		if taskPacketPath != "" {
			prompt += fmt.Sprintf("\n\n**Task Packet:** %s\n\nThe task packet contains: 00-contract.md, 10-plan.md, 20-work-log.md, 30-review.md, 40-acceptance.md.", taskPacketPath)
		}
	}

	prompt += fmt.Sprintf(`

**Configuration:**
- Timeout: %s
- Tools: %v

Execute the task according to your role definition.`,
		config.Delegation.Timeout,
		config.Tools,
	)

	return prompt
}

func (s *AgentServer) buildSystemPrompt(roleContext string) string {
	return s.buildSystemPromptForProject(roleContext, "")
}

func (s *AgentServer) buildSystemPromptForProject(roleContext string, projectRoot string) string {
	policy := s.loadSharedPolicy(projectRoot)
	if policy == "" {
		return roleContext
	}
	return policy + "\n\n---\n\n" + roleContext
}

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

	// Update taskDB
	if s.taskDB != nil {
		if err := s.taskDB.UpdateTaskStatus(execution.TaskID, constants.StatusInProgress, ""); err != nil {
			monitoring.Logger.Warn("failed_to_update_taskdb_start", "task_id", execution.TaskID, "error", err.Error())
		}
	}

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
	logMsg(fmt.Sprintf("📖 Loading role context from: %s", execution.Config.Context.RoleFile))

	// Role content is already loaded in config from .md file
	roleContext := execution.Config.Context.RoleContent
	if roleContext == "" {
		err := fmt.Errorf("role content is empty")
		monitoring.Logger.Error("role_context_empty", "task_id", execution.TaskID, "file", execution.Config.Context.RoleFile)
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
			// Only pass the task packet path to the agent if the directory actually exists.
			// If it doesn't exist, the agent would waste dozens of turns trying to read it.
			fullPath := path
			if !filepath.IsAbs(path) {
				projectRoot := execution.metadata["project_root"]
				if projectRoot == "" {
					projectRoot = s.rootDir
				}
				fullPath = filepath.Join(projectRoot, path)
			}
			if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
				taskPacketPath = path
				logMsg(fmt.Sprintf("📦 Task packet: %s", taskPacketPath))
			} else {
				logMsg(fmt.Sprintf("⚠️  Task packet path not found, skipping: %s", path))
			}
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
