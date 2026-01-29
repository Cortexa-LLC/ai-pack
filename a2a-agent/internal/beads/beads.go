package beads

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// Task represents a Beads task
type Task struct {
	ID           string                 `json:"id"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	Status       string                 `json:"status"`
	Dependencies []interface{}          `json:"dependencies,omitempty"` // Can be strings or objects
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Client handles Beads task operations
type Client struct {
	// Future: could add configuration options here
}

// NewClient creates a new Beads client
func NewClient() *Client {
	return &Client{}
}

// IsBeadsTaskID checks if a string is a Beads task ID
// Supports both default "bd-" prefix and custom prefixes like "a2a-agent-"
func IsBeadsTaskID(input string) bool {
	// Check if it looks like a Beads task ID: <prefix>-<hash>
	// Must contain at least one dash and have characters after it
	if !strings.Contains(input, "-") {
		return false
	}

	// Split on last dash to get hash part
	parts := strings.Split(input, "-")
	if len(parts) < 2 {
		return false
	}

	// Hash part should be alphanumeric and reasonably short (2-10 chars typically)
	hashPart := parts[len(parts)-1]
	if len(hashPart) < 2 || len(hashPart) > 10 {
		return false
	}

	// Check if hash part is alphanumeric
	for _, r := range hashPart {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '+') {
			return false
		}
	}

	return true
}

// GetTask retrieves a task from Beads
func (c *Client) GetTask(taskID string) (*Task, error) {
	return c.GetTaskFromDir(taskID, "")
}

// GetTaskFromDir retrieves a task from Beads using a specific working directory
func (c *Client) GetTaskFromDir(taskID string, workingDir string) (*Task, error) {
	if !IsBeadsTaskID(taskID) {
		return nil, fmt.Errorf("invalid Beads task ID: %s (expected format: <prefix>-<hash>)", taskID)
	}

	// Run: bd show <task-id> --json
	cmd := exec.Command("bd", "show", taskID, "--json")
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	output, err := cmd.Output()
	if err != nil {
		// Check if bd command exists
		if _, checkErr := exec.LookPath("bd"); checkErr != nil {
			return nil, fmt.Errorf("bd command not found - install Beads first: https://github.com/steveyegge/beads")
		}
		return nil, fmt.Errorf("failed to get Beads task %s: %w", taskID, err)
	}

	// bd show --json returns an array with one task
	var tasks []Task
	if err := json.Unmarshal(output, &tasks); err != nil {
		return nil, fmt.Errorf("failed to parse Beads task: %w", err)
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	return &tasks[0], nil
}

// StartTask marks a task as started in Beads
func (c *Client) StartTask(taskID string) error {
	return c.StartTaskFromDir(taskID, "")
}

// StartTaskFromDir marks a task as started in Beads using a specific working directory
func (c *Client) StartTaskFromDir(taskID string, workingDir string) error {
	if !IsBeadsTaskID(taskID) {
		return fmt.Errorf("invalid Beads task ID: %s", taskID)
	}

	// Run: bd update --claim <task-id>
	// This atomically claims the task (sets assignee and status to in_progress)
	cmd := exec.Command("bd", "update", "--claim", taskID)
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to claim task %s: %w", taskID, err)
	}

	return nil
}

// CompleteTask marks a task as complete in Beads
func (c *Client) CompleteTask(taskID string) error {
	return c.CompleteTaskFromDir(taskID, "")
}

// CompleteTaskFromDir marks a task as complete in Beads using a specific working directory
func (c *Client) CompleteTaskFromDir(taskID string, workingDir string) error {
	if !IsBeadsTaskID(taskID) {
		return fmt.Errorf("invalid Beads task ID: %s", taskID)
	}

	// Run: bd close <task-id>
	cmd := exec.Command("bd", "close", taskID)
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to complete task %s: %w", taskID, err)
	}

	return nil
}

// CheckDependencies verifies that all task dependencies are complete
func (c *Client) CheckDependencies(taskID string) (bool, []string, error) {
	return c.CheckDependenciesFromDir(taskID, "")
}

// CheckDependenciesFromDir verifies that all task dependencies are complete using a specific working directory
func (c *Client) CheckDependenciesFromDir(taskID string, workingDir string) (bool, []string, error) {
	task, err := c.GetTaskFromDir(taskID, workingDir)
	if err != nil {
		return false, nil, err
	}

	if len(task.Dependencies) == 0 {
		return true, nil, nil
	}

	var unmetDeps []string
	for _, dep := range task.Dependencies {
		// Dependencies can be either strings or objects with an "id" field
		var depID string
		switch v := dep.(type) {
		case string:
			depID = v
		case map[string]interface{}:
			if id, ok := v["id"].(string); ok {
				depID = id
			} else {
				continue // Skip if no valid ID
			}
		default:
			continue // Skip unknown types
		}

		depTask, err := c.GetTaskFromDir(depID, workingDir)
		if err != nil {
			return false, nil, fmt.Errorf("failed to check dependency %s: %w", depID, err)
		}

		if depTask.Status != "closed" && depTask.Status != "done" {
			unmetDeps = append(unmetDeps, depID)
		}
	}

	return len(unmetDeps) == 0, unmetDeps, nil
}

// IsInstalled checks if Beads (bd command) is installed
func IsInstalled() bool {
	_, err := exec.LookPath("bd")
	return err == nil
}

// ValidateTaskID validates that a Beads task ID exists and is accessible
func (c *Client) ValidateTaskID(taskID string) error {
	return c.ValidateTaskIDFromDir(taskID, "")
}

// ValidateTaskIDFromDir validates that a Beads task ID exists and is accessible using a specific working directory
func (c *Client) ValidateTaskIDFromDir(taskID string, workingDir string) error {
	if !IsBeadsTaskID(taskID) {
		return fmt.Errorf("invalid Beads task ID format: %s (expected format: <prefix>-<hash>)", taskID)
	}

	// Verify Beads is installed
	if !IsInstalled() {
		return fmt.Errorf("Beads (bd) is not installed - cannot validate task ID %s", taskID)
	}

	// Verify task exists by trying to get it
	_, err := c.GetTaskFromDir(taskID, workingDir)
	if err != nil {
		return fmt.Errorf("Beads task %s does not exist or is not accessible: %w", taskID, err)
	}

	return nil
}

func buildTaskDescription(task *Task) string {
	description := task.Title
	if task.Description != "" && task.Description != task.Title {
		description = fmt.Sprintf("%s\n\n%s", task.Title, task.Description)
	}
	return description
}

func getTaskPacketPath(task *Task) string {
	// Check metadata first
	if task.Metadata != nil {
		if path, ok := task.Metadata["task_packet"].(string); ok {
			return path
		}
	}

	// Try parsing from description
	if task.Description != "" {
		return extractTaskPacketPath(task.Description)
	}

	return ""
}

// GetTaskDescription extracts the task description from either a Beads task ID or free-form text
// Returns: (description, taskPacketPath, workingDirectory, isBeadsTask, error)
func (c *Client) GetTaskDescription(input string) (string, string, string, bool, error) {
	return c.GetTaskDescriptionFromDir(input, "")
}

// GetTaskDescriptionFromDir extracts the task description from either a Beads task ID or free-form text using a specific working directory
// Returns: (description, taskPacketPath, workingDirectory, isBeadsTask, error)
func (c *Client) GetTaskDescriptionFromDir(input string, projectRoot string) (string, string, string, bool, error) {
	if !IsBeadsTaskID(input) {
		return input, "", "", false, nil
	}

	if err := c.ValidateTaskIDFromDir(input, projectRoot); err != nil {
		return "", "", "", true, err
	}

	task, err := c.GetTaskFromDir(input, projectRoot)
	if err != nil {
		return "", "", "", true, fmt.Errorf("failed to get Beads task: %w", err)
	}

	description := buildTaskDescription(task)
	taskPacketPath := getTaskPacketPath(task)
	workingDirectory := ""
	if task.Description != "" {
		workingDirectory = extractWorkingDirectory(task.Description)
	}

	return description, taskPacketPath, workingDirectory, true, nil
}

// extractTaskPacketPath extracts task packet path from description
// Looks for pattern: "Task packet: <path>" or "task packet: <path>"
func extractTaskPacketPath(description string) string {
	// Pattern matches "Task packet: .ai/tasks/..." or "task packet: .ai/tasks/..."
	re := regexp.MustCompile(`(?i)Task packet:\s*([^\s\n]+)`)
	matches := re.FindStringSubmatch(description)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// extractWorkingDirectory extracts working directory from description
// Looks for pattern: "Working directory: <path>" or "working directory: <path>"
func extractWorkingDirectory(description string) string {
	// Pattern matches "Working directory: /path/..." or "working directory: /path/..."
	re := regexp.MustCompile(`(?i)Working directory:\s*([^\n]+)`)
	matches := re.FindStringSubmatch(description)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}
