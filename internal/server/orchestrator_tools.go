package server

import (
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
						"description": "Optional: Task priority (low/medium/high)",
						"enum":        []string{"low", "medium", "high"},
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
						"description": "The Beads task ID to work on (e.g. xasm++-abc123)",
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
		priority = "medium"
	}

	// Create Beads task using bd create command
	cmd := exec.Command("bd", "create", description, "--priority", priority, "--json")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to create Beads task: %v (output: %s)", err, string(output))
	}

	// Parse task ID from JSON output
	var createResult struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(output, &createResult); err != nil {
		return "", fmt.Errorf("failed to parse Beads task creation output: %w", err)
	}

	if createResult.ID == "" {
		return "", fmt.Errorf("Beads task created but no ID returned")
	}

	monitoring.Logger.Info("orchestrator_task_created", "task_id", createResult.ID, "project", projectRoot)

	result := map[string]interface{}{
		"success":  true,
		"task_id":  createResult.ID,
		"message":  fmt.Sprintf("Task created: %s", createResult.ID),
		"priority": priority,
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
		"success":   true,
		"task_id":   response.TaskID,
		"status":    response.Status,
		"message":   fmt.Sprintf("Agent %s spawned for task %s", role, taskID),
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

	resultJSON, _ := json.Marshal(result.Data.Task)
	return string(resultJSON), nil
}

func (s *AgentServer) executeUpdateTaskStatus(input map[string]interface{}) (string, error) {
	taskID, _ := input["task_id"].(string)
	status, _ := input["status"].(string)
	reason, _ := input["reason"].(string)

	if taskID == "" || status == "" {
		return "", fmt.Errorf("task_id and status are required")
	}

	// Find the task execution metadata and update it
	// The taskID could be either a Beads task ID or a timestamped execution folder
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
// - A Beads task ID (e.g., "xasm++-8hyz")
// - A timestamped execution folder (e.g., "xasm++-8hyz-20260211-111548")
func (s *AgentServer) findTaskExecutionMetadata(taskID string) (projectRoot, executionFolder string, err error) {
	// Get all registered project roots
	projectRoots := s.GetProjectRoots()

	// Try each project root
	for _, pr := range projectRoots {
		// Try direct path first (in case taskID is the execution folder)
		metadataPath := filepath.Join(pr, ".beads", "tasks", taskID, "00-metadata.json")
		if _, statErr := os.Stat(metadataPath); statErr == nil {
			return pr, taskID, nil
		}

		// If direct path doesn't exist, try finding most recent execution
		// This handles the case where taskID is just the Beads task ID
		execFolder := s.findMostRecentExecutionInProject(pr, taskID)
		if execFolder != "" {
			metadataPath = filepath.Join(pr, ".beads", "tasks", execFolder, "00-metadata.json")
			if _, statErr := os.Stat(metadataPath); statErr == nil {
				return pr, execFolder, nil
			}
		}
	}

	return "", "", fmt.Errorf("no execution metadata found for task %s in any project", taskID)
}

// findMostRecentExecutionInProject finds the most recent timestamped execution folder for a Beads task ID
func (s *AgentServer) findMostRecentExecutionInProject(projectRoot, beadsTaskID string) string {
	tasksDir := filepath.Join(projectRoot, ".beads", "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return ""
	}

	var mostRecentFolder string
	var mostRecentTime time.Time

	// Find all folders matching {beads-id}-{timestamp} pattern
	prefix := beadsTaskID + "-"
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		folderName := entry.Name()
		// Check if folder matches pattern: {beads-id}-{timestamp}
		if strings.HasPrefix(folderName, prefix) {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if mostRecentFolder == "" || info.ModTime().After(mostRecentTime) {
				mostRecentFolder = folderName
				mostRecentTime = info.ModTime()
			}
		}
	}

	return mostRecentFolder
}

// updateTaskExecutionStatus updates the status field in a task execution metadata file
func (s *AgentServer) updateTaskExecutionStatus(projectRoot, executionFolder, status, reason string) error {
	metadataPath := filepath.Join(projectRoot, ".beads", "tasks", executionFolder, "00-metadata.json")

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
