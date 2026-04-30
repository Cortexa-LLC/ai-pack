package server

import (
	"github.com/cortexa-llc/ai-pack/internal/constants"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
)

// GetOrchestratorTools returns the tool definitions for the orchestrator
func GetOrchestratorTools() []anthropic.ToolUnionParam {
	return []anthropic.ToolUnionParam{
		anthropic.ToolUnionParamOfTool(
			anthropic.ToolInputSchemaParam{
				Type: "object",
				Properties: map[string]interface{}{
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Description of the task to create",
					},
					"project_root": map[string]interface{}{
						"type":        "string",
						"description": "The project root directory for the task",
					},
					"priority": map[string]interface{}{
						"type":        "string",
						"description": "Optional: Task priority (P0=critical, P1=urgent, P2=normal, P3=low, P4=backlog)",
						"enum":        []string{"P0", "P1", "P2", "P3", "P4"},
					},
				},
				Required: []string{"description", "project_root"},
			},
			"create_task",
		),
		anthropic.ToolUnionParamOfTool(
			anthropic.ToolInputSchemaParam{
				Type: "object",
				Properties: map[string]interface{}{
					"role": map[string]interface{}{
						"type":        "string",
						"description": "The agent role to spawn (must match a role file in the project roles/ directory)",
					},
					"task_id": map[string]interface{}{
						"type":        "string",
						"description": "The task ID to work on (e.g. xasm++-abc123)",
					},
					"project_root": map[string]interface{}{
						"type":        "string",
						"description": "The project root directory where the agent should execute",
					},
				},
				Required: []string{"role", "task_id", "project_root"},
			},
			"spawn_agent",
		),
		anthropic.ToolUnionParamOfTool(
			anthropic.ToolInputSchemaParam{
				Type: "object",
				Properties: map[string]interface{}{
					"status_filter": map[string]interface{}{
						"type":        "string",
						"description": "Optional: Filter tasks by status (queued, in_progress, completed, failed, blocked)",
					},
				},
				Required: []string{},
			},
			"query_tasks",
		),
		anthropic.ToolUnionParamOfTool(
			anthropic.ToolInputSchemaParam{
				Type: "object",
				Properties: map[string]interface{}{
					"task_id": map[string]interface{}{
						"type":        "string",
						"description": "The task ID to get details for",
					},
				},
				Required: []string{"task_id"},
			},
			"get_task_details",
		),
		anthropic.ToolUnionParamOfTool(
			anthropic.ToolInputSchemaParam{
				Type: "object",
				Properties: map[string]interface{}{
					"task_id": map[string]interface{}{
						"type":        "string",
						"description": "The task ID to update",
					},
					"status": map[string]interface{}{
						"type":        "string",
						"description": "The new status",
						"enum":        []string{"queued", "in_progress", "blocked", "completed", "failed"},
					},
					"reason": map[string]interface{}{
						"type":        "string",
						"description": "Optional: Reason for the status change",
					},
				},
				Required: []string{"task_id", "status"},
			},
			"update_task_status",
		),
	}
}

// ExecuteTool executes an orchestrator tool and returns the result
func (s *AgentServer) ExecuteTool(toolName string, toolInput map[string]interface{}) (string, error) {
	monitoring.Logger.Info("orchestrator_tool_execution", "tool", toolName, "input", toolInput)

	switch toolName {
	case "create_task":
		return s.executeCreateTask(toolInput)
	case "spawn_agent":
		return s.executeSpawnAgent(toolInput)
	case "query_tasks":
		return s.executeQueryTasks(toolInput)
	case "get_task_details":
		return s.executeGetTaskDetails(toolInput)
	case "update_task_status":
		return s.executeUpdateTaskStatus(toolInput)
	default:
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}
}

func (s *AgentServer) executeCreateTask(input map[string]interface{}) (string, error) {
	description, _ := input["description"].(string)
	projectRoot, _ := input["project_root"].(string)
	priority, _ := input["priority"].(string)

	if description == "" || projectRoot == "" {
		return "", fmt.Errorf("missing required parameters: description, project_root")
	}

	// Default priority
	if priority == "" {
		priority = "P2"
	}

	// Create task using bd create command
	cmd := exec.Command("bd", "create", description, "--priority", priority, "--json")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to create task: %v (output: %s)", err, string(output))
	}

	// Parse task ID from JSON output
	var createResult struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(output, &createResult); err != nil {
		return "", fmt.Errorf("failed to parse task creation output: %w", err)
	}

	if createResult.ID == "" {
		return "", fmt.Errorf("task created but no ID returned")
	}

	// Build task packet directory slug: <shortTaskID>-<timestamp>-<short-desc>
	timestamp := time.Now().Format("20060102150405")
	firstLine := description
	if idx := strings.Index(description, "\n"); idx != -1 {
		firstLine = description[:idx]
	}
	// Build slug from first 5 words of description, lowercase, hyphens
	words := strings.Fields(firstLine)
	if len(words) > 5 {
		words = words[:5]
	}
	shortDesc := strings.ToLower(strings.Join(words, "-"))
	// Remove characters unsafe for directory names
	var slugBuf bytes.Buffer
	for _, r := range shortDesc {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			slugBuf.WriteRune(r)
		} else if r == ' ' {
			slugBuf.WriteRune('-')
		}
	}
	slug := fmt.Sprintf("%s-%s-%s", createResult.ID, timestamp, slugBuf.String())

	// Create task packet directory under .ai/tasks/<slug>/
	taskPacketDir := filepath.Join(projectRoot, ".ai", "tasks", slug)
	if err := os.MkdirAll(taskPacketDir, 0755); err != nil {
		monitoring.Logger.Warn("task_packet_dir_failed", "dir", taskPacketDir, "error", err)
	} else {
		// Copy template files from templates/task-packet/ into the task packet dir
		templatesDir := filepath.Join(s.rootDir, "templates", "task-packet")
		entries, readErr := os.ReadDir(templatesDir)
		if readErr != nil {
			monitoring.Logger.Warn("task_packet_templates_unreadable", "dir", templatesDir, "error", readErr)
		} else {
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				srcPath := filepath.Join(templatesDir, entry.Name())
				dstPath := filepath.Join(taskPacketDir, entry.Name())
				srcData, readFileErr := os.ReadFile(srcPath)
				if readFileErr != nil {
					monitoring.Logger.Warn("task_packet_template_read_failed", "file", srcPath, "error", readFileErr)
					continue
				}
				if writeErr := os.WriteFile(dstPath, srcData, 0644); writeErr != nil {
					monitoring.Logger.Warn("task_packet_template_write_failed", "file", dstPath, "error", writeErr)
				}
			}
		}

		// Update the Beads description to include Working directory and Task packet lines
		taskPacketRelPath := filepath.Join(".ai", "tasks", slug)
		updatedDescription := fmt.Sprintf("%s\n\nWorking directory: %s\nTask packet: %s/", description, projectRoot, taskPacketRelPath)
		updateCmd := exec.Command("bd", "update", createResult.ID, "-d", updatedDescription)
		updateCmd.Dir = projectRoot
		if updateOut, updateErr := updateCmd.CombinedOutput(); updateErr != nil {
			monitoring.Logger.Warn("task_packet_description_update_failed", "task_id", createResult.ID, "error", updateErr, "output", string(updateOut))
		}
	}

	monitoring.Logger.Info("orchestrator_task_created", "task_id", createResult.ID, "project", projectRoot, "packet_dir", taskPacketDir)

	result := map[string]interface{}{
		"success":    true,
		"task_id":    createResult.ID,
		"message":    fmt.Sprintf("Task created: %s", createResult.ID),
		"priority":   priority,
		"packet_dir": filepath.Join(".ai", "tasks", slug),
	}

	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

func (s *AgentServer) executeSpawnAgent(input map[string]interface{}) (string, error) {
	role, _ := input["role"].(string)
	taskID, _ := input["task_id"].(string)
	projectRoot, _ := input["project_root"].(string)

	if role == "" || taskID == "" || projectRoot == "" {
		return "", fmt.Errorf("missing required parameters: role, task_id, project_root")
	}

	// Spawn agent via /a2a/start endpoint
	response, err := s.spawnAgentTask(role, taskID, projectRoot)
	if err != nil {
		return "", fmt.Errorf("failed to spawn agent: %w", err)
	}

	result := map[string]interface{}{
		"success":    true,
		"task_id":    response.TaskID,
		"status":     response.Status,
		"message":    fmt.Sprintf("Agent %s spawned for task %s", role, taskID),
		"stream_url": fmt.Sprintf("/stream/%s", response.TaskID),
	}

	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

func (s *AgentServer) executeQueryTasks(input map[string]interface{}) (string, error) {
	statusFilter, _ := input["status_filter"].(string)

	// Query via GraphQL
	query := `
		query {
			tasks {
				id
				status
				task
				description
				projectRoot
			}
		}
	`

	requestBody, _ := json.Marshal(map[string]string{"query": query})
	resp, err := http.Post("http://localhost:8080/graphql", "application/json", bytes.NewReader(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to query tasks: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Tasks []map[string]interface{} `json:"tasks"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	tasks := result.Data.Tasks

	// Apply status filter if provided
	if statusFilter != "" {
		filtered := []map[string]interface{}{}
		for _, task := range tasks {
			if task["status"] == statusFilter {
				filtered = append(filtered, task)
			}
		}
		tasks = filtered
	}

	resultJSON, _ := json.Marshal(map[string]interface{}{
		"tasks": tasks,
		"count": len(tasks),
	})

	return string(resultJSON), nil
}

func (s *AgentServer) executeGetTaskDetails(input map[string]interface{}) (string, error) {
	taskID, _ := input["task_id"].(string)
	if taskID == "" {
		return "", fmt.Errorf("task_id is required")
	}

	// Query specific task via GraphQL
	query := fmt.Sprintf(`
		query {
			task(id: "%s") {
				id
				status
				task
				description
				projectRoot
				createdAt
				updatedAt
			}
		}
	`, taskID)

	requestBody, _ := json.Marshal(map[string]string{"query": query})
	resp, err := http.Post("http://localhost:8080/graphql", "application/json", bytes.NewReader(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to query task: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Task map[string]interface{} `json:"task"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Data.Task == nil {
		return "", fmt.Errorf("task not found: %s", taskID)
	}

	task := result.Data.Task

	// Augment with last execution metadata (status, error, turn count) so the
	// model knows if a previous run was interrupted, failed, or completed.
	projectRoot, _ := task["projectRoot"].(string)
	if projectRoot != "" {
		if execStatus := s.getLastExecutionStatus(taskID, projectRoot); execStatus != nil {
			task["last_execution"] = execStatus
		}
	}

	resultJSON, _ := json.Marshal(task)
	return string(resultJSON), nil
}

// getLastExecutionStatus reads the most recent execution metadata for a task and
// returns a summary suitable for the orchestrator model.
func (s *AgentServer) getLastExecutionStatus(taskID, projectRoot string) map[string]interface{} {
	tasksDir := filepath.Join(projectRoot, constants.TaskRootDir, "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return nil
	}

	prefix := taskID + "-"
	type execEntry struct {
		name string
		t    time.Time
	}
	var found []execEntry
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), prefix) {
			continue
		}
		suffix := strings.TrimPrefix(e.Name(), prefix)
		t, err := time.Parse("20060102-150405", suffix)
		if err != nil {
			continue
		}
		found = append(found, execEntry{e.Name(), t})
	}
	if len(found) == 0 {
		return nil
	}
	// Sort most-recent first
	for i := 0; i < len(found)-1; i++ {
		for j := i + 1; j < len(found); j++ {
			if found[j].t.After(found[i].t) {
				found[i], found[j] = found[j], found[i]
			}
		}
	}

	metadataPath := filepath.Join(tasksDir, found[0].name, "00-metadata.json")
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil
	}
	var meta map[string]interface{}
	if json.Unmarshal(data, &meta) != nil {
		return nil
	}

	summary := map[string]interface{}{
		"execution_folder": found[0].name,
		"started_at":       meta["started_at"],
		"updated_at":       meta["updated_at"],
		"status":           meta["status"],
	}
	if r, ok := meta["error_reason"].(string); ok && r != "" {
		summary["error_reason"] = r
	}
	if n, ok := meta["turn_count"]; ok {
		summary["turn_count"] = n
	}
	return summary
}

func (s *AgentServer) executeUpdateTaskStatus(input map[string]interface{}) (string, error) {
	taskID, _ := input["task_id"].(string)
	status, _ := input["status"].(string)
	reason, _ := input["reason"].(string)

	if taskID == "" || status == "" {
		return "", fmt.Errorf("task_id and status are required")
	}

	// Find the task execution metadata and update it
	// The taskID could be either a task ID or a timestamped execution folder
	projectRoot, executionFolder, err := s.findTaskExecutionMetadata(taskID)
	if err != nil {
		return "", fmt.Errorf("task not found: %w", err)
	}

	// Update the metadata file
	if err := s.updateTaskExecutionStatus(projectRoot, executionFolder, status, reason); err != nil {
		return "", fmt.Errorf("failed to update task status: %w", err)
	}

	monitoring.Logger.Info("orchestrator_task_status_updated",
		"task_id", taskID,
		"execution_folder", executionFolder,
		"status", status,
		"project_root", projectRoot)

	result := map[string]interface{}{
		"success":          true,
		"task_id":          taskID,
		"execution_folder": executionFolder,
		"status":           status,
		"message":          fmt.Sprintf("Task %s status updated to %s", taskID, status),
		"reason":           reason,
		"project_root":     projectRoot,
	}

	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// findTaskExecutionMetadata finds the project root and execution folder for a task ID
// The taskID can be either:
// - A task ID (e.g., "xasm++-8hyz")
// - A timestamped execution folder (e.g., "xasm++-8hyz-20260211-111548")
func (s *AgentServer) findTaskExecutionMetadata(taskID string) (projectRoot, executionFolder string, err error) {
	// Get all registered project roots
	projectRoots := s.GetProjectRoots()

	// Try each project root
	for _, pr := range projectRoots {
		// Try direct path first (in case taskID is the execution folder)
		metadataPath := filepath.Join(pr, constants.TaskRootDir, "tasks", taskID, "00-metadata.json")
		if _, statErr := os.Stat(metadataPath); statErr == nil {
			return pr, taskID, nil
		}

		// If direct path doesn't exist, try finding most recent execution
		// This handles the case where taskID is just the task ID
		execFolder := s.findMostRecentExecutionInProject(pr, taskID)
		if execFolder != "" {
			metadataPath = filepath.Join(pr, constants.TaskRootDir, "tasks", execFolder, "00-metadata.json")
			if _, statErr := os.Stat(metadataPath); statErr == nil {
				return pr, execFolder, nil
			}
		}
	}

	return "", "", fmt.Errorf("no execution metadata found for task %s in any project", taskID)
}

// findMostRecentExecutionInProject finds the most recent timestamped execution folder for a task ID
func (s *AgentServer) findMostRecentExecutionInProject(projectRoot, taskID string) string {
	tasksDir := filepath.Join(projectRoot, constants.TaskRootDir, "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return ""
	}

	var mostRecentFolder string
	var mostRecentFolderTime time.Time

	// Find all folders matching {short-task-id}-{timestamp} pattern
	prefix := taskID + "-"
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		folderName := entry.Name()
		// Check if folder matches pattern: {short-task-id}-{timestamp}
		if strings.HasPrefix(folderName, prefix) {
			// Sort by the timestamp embedded in the folder name, not by filesystem
			// mtime. mtime is unreliable: marking an old execution as superseded
			// updates its mtime AFTER the new execution folder was created, causing
			// the old (failed) folder to appear newer than the running one.
			suffix := strings.TrimPrefix(folderName, prefix)
			folderTime, err := time.Parse("20060102-150405", suffix)
			if err != nil {
				continue
			}
			if mostRecentFolder == "" || folderTime.After(mostRecentFolderTime) {
				mostRecentFolder = folderName
				mostRecentFolderTime = folderTime
			}
		}
	}

	return mostRecentFolder
}

// updateTaskExecutionStatus updates the status field in a task execution metadata file
func (s *AgentServer) updateTaskExecutionStatus(projectRoot, executionFolder, status, reason string) error {
	metadataPath := filepath.Join(projectRoot, constants.TaskRootDir, "tasks", executionFolder, "00-metadata.json")

	// Read current metadata
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to read metadata: %w", err)
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		return fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Update status and timestamp
	metadata["status"] = status
	metadata["updated_at"] = time.Now().Format(time.RFC3339)

	// Add reason if provided
	if reason != "" {
		metadata["status_reason"] = reason
	}

	// Write updated metadata
	updatedData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}
