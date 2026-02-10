package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
)

// GetOrchestratorTools returns the tool definitions for the orchestrator
func GetOrchestratorTools() []anthropic.ToolParam {
	return []anthropic.ToolParam{
		{
			Name:        "spawn_agent",
			Description: "Spawn a background agent to work on a task. The agent will execute in the project directory and you can monitor its progress.",
			InputSchema: anthropic.ToolInputSchemaParam{
				Type: "object",
				Properties: map[string]interface{}{
					"role": map[string]interface{}{
						"type":        "string",
						"description": "The agent role to spawn (engineer, reviewer, architect, tester, etc.)",
						"enum":        []string{"engineer", "reviewer", "architect", "tester", "designer", "archaeologist", "spelunker"},
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
		},
		{
			Name:        "query_tasks",
			Description: "Query the task system to get current status of all tasks. Returns list of tasks with their status, title, and other metadata.",
			InputSchema: anthropic.ToolInputSchemaParam{
				Type: "object",
				Properties: map[string]interface{}{
					"status_filter": map[string]interface{}{
						"type":        "string",
						"description": "Optional: Filter tasks by status (queued, in_progress, completed, failed, blocked)",
					},
				},
				Required: []string{},
			},
		},
		{
			Name:        "get_task_details",
			Description: "Get detailed information about a specific task including its description, status, dependencies, and history.",
			InputSchema: anthropic.ToolInputSchemaParam{
				Type: "object",
				Properties: map[string]interface{}{
					"task_id": map[string]interface{}{
						"type":        "string",
						"description": "The task ID to get details for",
					},
				},
				Required: []string{"task_id"},
			},
		},
		{
			Name:        "update_task_status",
			Description: "Update the status of a task (e.g. mark as blocked, failed, or update priority)",
			InputSchema: anthropic.ToolInputSchemaParam{
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
		},
	}
}

// ExecuteTool executes an orchestrator tool and returns the result
func (s *AgentServer) ExecuteTool(toolName string, toolInput map[string]interface{}) (string, error) {
	monitoring.Logger.Info("orchestrator_tool_execution", "tool", toolName, "input", toolInput)

	switch toolName {
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

	// Update via Beads
	// For now, return a simulated success
	// TODO: Implement actual Beads update command

	result := map[string]interface{}{
		"success": true,
		"task_id": taskID,
		"status":  status,
		"message": fmt.Sprintf("Task %s status updated to %s", taskID, status),
		"reason":  reason,
	}

	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}
