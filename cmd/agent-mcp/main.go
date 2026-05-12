// cmd/mcp-agent/main.go
// MCP stdio server that exposes agent CLI operations as MCP tools.
// CGO_ENABLED=0 — no SQLite dependency; all work is delegated to the agent CLI.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/cortexa-llc/ai-pack/internal/mcp"
)

func main() {
	tools := []mcp.Tool{
		{
			Name:        "create_task",
			Description: "Create a new agent task in the ai-pack task system. Returns the task_id to pass to spawn_agent.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"description": map[string]any{
						"type":        "string",
						"description": "Human-readable task title/description",
					},
					"role": map[string]any{
						"type":        "string",
						"enum":        []string{"engineer", "architect", "reviewer", "spelunker"},
						"description": "Agent role to assign",
					},
					"priority": map[string]any{
						"type":        "string",
						"enum":        []string{"P0", "P1", "P2", "P3", "P4"},
						"description": "Task priority (P0=critical, P4=low)",
					},
				},
				"required": []string{"description", "role"},
			},
		},
		{
			Name:        "spawn_agent",
			Description: "Spawn an agent to work on a task. Use stream=true for sequential tasks (blocks until done). Use stream=false for parallel tasks (returns immediately).",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"task_id": map[string]any{
						"type":        "string",
						"description": "Task ID returned by create_task (e.g. ai-pack-f32)",
					},
					"role": map[string]any{
						"type":        "string",
						"enum":        []string{"engineer", "architect", "reviewer", "spelunker"},
					},
					"stream": map[string]any{
						"type":        "boolean",
						"description": "Block until complete and stream output (default: true)",
					},
				},
				"required": []string{"task_id", "role"},
			},
		},
		{
			Name:        "list_tasks",
			Description: "List agent tasks. Filter by status to find in-progress or completed work.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"status": map[string]any{
						"type":        "string",
						"enum":        []string{"all", "running", "completed", "failed", "open"},
						"description": "Filter by status (default: all)",
					},
				},
			},
		},
		{
			Name:        "get_task_status",
			Description: "Get current status and summary for a specific task.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"task_id": map[string]any{
						"type": "string",
					},
				},
				"required": []string{"task_id"},
			},
		},
		{
			Name:        "get_task_logs",
			Description: "Get recent execution log output for a task.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"task_id": map[string]any{
						"type": "string",
					},
					"lines": map[string]any{
						"type":        "integer",
						"description": "Number of tail lines to return (default: 50)",
					},
				},
				"required": []string{"task_id"},
			},
		},
	}

	handlers := map[string]mcp.ToolHandler{
		"create_task":     handleCreateTask,
		"spawn_agent":     handleSpawnAgent,
		"list_tasks":      handleListTasks,
		"get_task_status": handleGetTaskStatus,
		"get_task_logs":   handleGetTaskLogs,
	}

	server := mcp.NewServer(tools, handlers, bufio.NewReader(os.Stdin), os.Stdout)
	if err := server.Serve(); err != nil {
		fmt.Fprintf(os.Stderr, "mcp-agent: server error: %v\n", err)
		os.Exit(1)
	}
}

// agentBin returns the path to the agent binary, checking PATH first.
func agentBin() string {
	if path, err := exec.LookPath("agent"); err == nil {
		return path
	}
	return "/usr/local/bin/agent"
}

// handleCreateTask creates a new task and returns the task_id.
func handleCreateTask(req *mcp.ToolCallRequest) (any, error) {
	description, _ := req.Arguments["description"].(string)
	if description == "" {
		return nil, fmt.Errorf("description is required")
	}
	role, _ := req.Arguments["role"].(string)
	if role == "" {
		return nil, fmt.Errorf("role is required")
	}
	priority, _ := req.Arguments["priority"].(string)

	args := []string{"create", description, "--role", role, "--json"}
	if priority != "" {
		args = append(args, "--priority", priority)
	}

	cmd := exec.Command(agentBin(), args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("agent create failed: %w\noutput: %s", err, string(out))
	}

	// Parse JSON output to extract task_id
	var result map[string]any
	if jsonErr := json.Unmarshal(out, &result); jsonErr != nil {
		// Return raw output if not JSON
		return strings.TrimSpace(string(out)), nil
	}
	// Try common task_id keys
	for _, key := range []string{"task_id", "id", "taskId"} {
		if v, ok := result[key].(string); ok && v != "" {
			return map[string]any{"task_id": v, "raw": result}, nil
		}
	}
	return result, nil
}

// handleSpawnAgent runs `agent <role> <task_id>`.
// stream=true (default): blocks until done, returns combined output.
// stream=false: fires and forgets, returns immediately.
func handleSpawnAgent(req *mcp.ToolCallRequest) (any, error) {
	taskID, _ := req.Arguments["task_id"].(string)
	if taskID == "" {
		return nil, fmt.Errorf("task_id is required")
	}
	role, _ := req.Arguments["role"].(string)
	if role == "" {
		return nil, fmt.Errorf("role is required")
	}

	// Default stream=true
	stream := true
	if sv, ok := req.Arguments["stream"].(bool); ok {
		stream = sv
	}

	args := []string{role, taskID}

	cmd := exec.Command(agentBin(), args...)

	if stream {
		out, err := cmd.CombinedOutput()
		result := strings.TrimSpace(string(out))
		if err != nil {
			return fmt.Sprintf("Agent exited with error: %v\n\nOutput:\n%s", err, result), nil
		}
		return result, nil
	}

	// Non-blocking: start and detach
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to spawn agent: %w", err)
	}
	// Detach from parent — let it run independently
	go func() { _ = cmd.Wait() }()
	return fmt.Sprintf("Agent spawned for task %s (role: %s). Use get_task_status or get_task_logs to monitor.", taskID, role), nil
}

// handleListTasks runs `agent list` with an optional status filter.
func handleListTasks(req *mcp.ToolCallRequest) (any, error) {
	status, _ := req.Arguments["status"].(string)

	args := []string{"list"}
	switch status {
	case "running":
		args = append(args, "--status", "in_progress")
	case "completed":
		args = append(args, "--status", "done")
	case "failed":
		args = append(args, "--status", "failed")
	case "open":
		args = append(args, "--status", "open")
	case "all", "":
		// no filter — list everything
	default:
		args = append(args, "--status", status)
	}

	cmd := exec.Command(agentBin(), args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("agent list failed: %w\noutput: %s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

// handleGetTaskStatus runs `agent show <task_id>`.
func handleGetTaskStatus(req *mcp.ToolCallRequest) (any, error) {
	taskID, _ := req.Arguments["task_id"].(string)
	if taskID == "" {
		return nil, fmt.Errorf("task_id is required")
	}

	cmd := exec.Command(agentBin(), "show", taskID)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("agent show failed: %w\noutput: %s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

// handleGetTaskLogs runs `agent logs <task_id>` and tails the last N lines.
func handleGetTaskLogs(req *mcp.ToolCallRequest) (any, error) {
	taskID, _ := req.Arguments["task_id"].(string)
	if taskID == "" {
		return nil, fmt.Errorf("task_id is required")
	}

	lines := 50
	switch v := req.Arguments["lines"].(type) {
	case float64:
		lines = int(v)
	case int:
		lines = v
	case string:
		if n, err := strconv.Atoi(v); err == nil {
			lines = n
		}
	}

	cmd := exec.Command(agentBin(), "logs", taskID)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("agent logs failed: %w\noutput: %s", err, string(out))
	}

	text := strings.TrimSpace(string(out))
	allLines := strings.Split(text, "\n")
	if len(allLines) > lines {
		allLines = allLines[len(allLines)-lines:]
	}
	return strings.Join(allLines, "\n"), nil
}
