package beads

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Task represents a Beads task
type Task struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Status       string   `json:"status"`
	Dependencies []string `json:"dependencies,omitempty"`
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
func IsBeadsTaskID(input string) bool {
	return strings.HasPrefix(input, "bd-") && len(input) > 3
}

// GetTask retrieves a task from Beads
func (c *Client) GetTask(taskID string) (*Task, error) {
	if !IsBeadsTaskID(taskID) {
		return nil, fmt.Errorf("invalid Beads task ID: %s (expected format: bd-xxxx)", taskID)
	}

	// Run: bd show <task-id> --json
	cmd := exec.Command("bd", "show", taskID, "--json")
	output, err := cmd.Output()
	if err != nil {
		// Check if bd command exists
		if _, checkErr := exec.LookPath("bd"); checkErr != nil {
			return nil, fmt.Errorf("bd command not found - install Beads first: https://github.com/steveyegge/beads")
		}
		return nil, fmt.Errorf("failed to get Beads task %s: %w", taskID, err)
	}

	var task Task
	if err := json.Unmarshal(output, &task); err != nil {
		return nil, fmt.Errorf("failed to parse Beads task: %w", err)
	}

	return &task, nil
}

// StartTask marks a task as started in Beads
func (c *Client) StartTask(taskID string) error {
	if !IsBeadsTaskID(taskID) {
		return fmt.Errorf("invalid Beads task ID: %s", taskID)
	}

	// Run: bd start <task-id>
	cmd := exec.Command("bd", "start", taskID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start task %s: %w", taskID, err)
	}

	return nil
}

// CompleteTask marks a task as complete in Beads
func (c *Client) CompleteTask(taskID string) error {
	if !IsBeadsTaskID(taskID) {
		return fmt.Errorf("invalid Beads task ID: %s", taskID)
	}

	// Run: bd close <task-id>
	cmd := exec.Command("bd", "close", taskID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to complete task %s: %w", taskID, err)
	}

	return nil
}

// CheckDependencies verifies that all task dependencies are complete
func (c *Client) CheckDependencies(taskID string) (bool, []string, error) {
	task, err := c.GetTask(taskID)
	if err != nil {
		return false, nil, err
	}

	if len(task.Dependencies) == 0 {
		return true, nil, nil
	}

	var unmetDeps []string
	for _, depID := range task.Dependencies {
		dep, err := c.GetTask(depID)
		if err != nil {
			return false, nil, fmt.Errorf("failed to check dependency %s: %w", depID, err)
		}

		if dep.Status != "closed" && dep.Status != "done" {
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
	if !IsBeadsTaskID(taskID) {
		return fmt.Errorf("invalid Beads task ID format: %s (expected format: bd-xxxx)", taskID)
	}

	// Verify Beads is installed
	if !IsInstalled() {
		return fmt.Errorf("Beads (bd) is not installed - cannot validate task ID %s", taskID)
	}

	// Verify task exists by trying to get it
	_, err := c.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("Beads task %s does not exist or is not accessible: %w", taskID, err)
	}

	return nil
}

// GetTaskDescription extracts the task description from either a Beads task ID or free-form text
func (c *Client) GetTaskDescription(input string) (string, bool, error) {
	if IsBeadsTaskID(input) {
		// Validate the task exists first
		if err := c.ValidateTaskID(input); err != nil {
			return "", true, err
		}

		// Get the task (we know it exists now)
		task, err := c.GetTask(input)
		if err != nil {
			return "", true, fmt.Errorf("failed to get Beads task: %w", err)
		}

		// Use title and description
		description := task.Title
		if task.Description != "" && task.Description != task.Title {
			description = fmt.Sprintf("%s\n\n%s", task.Title, task.Description)
		}

		return description, true, nil
	}

	// It's free-form text
	return input, false, nil
}
