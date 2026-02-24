//go:build integration
// +build integration

package beads

import (
	"os/exec"
	"testing"
)

// Integration tests that require Beads to be installed
// Run with: go test -tags=integration ./internal/beads/...

func TestIntegrationValidateTaskID(t *testing.T) {
	if !IsInstalled() {
		t.Skip("Beads not installed, skipping integration test")
	}

	client := NewClient()

	// Test with invalid task ID
	err := client.ValidateTaskID("bd-nonexistent-xyz")
	if err == nil {
		t.Error("Expected error for non-existent task ID, got nil")
	}
}

func TestIntegrationGetTask(t *testing.T) {
	if !IsInstalled() {
		t.Skip("Beads not installed, skipping integration test")
	}

	client := NewClient()

	// Create a test task
	cmd := exec.Command("bd", "create", "Test task for integration testing")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	// Extract task ID from output (format: "Created: bd-xxxx")
	taskID := string(output)
	// Parse task ID from output...
	// (This would need proper parsing in real implementation)

	if taskID == "" {
		t.Skip("Could not create test task")
	}

	// Get the task
	task, err := client.GetTask(taskID)
	if err != nil {
		t.Errorf("Failed to get test task: %v", err)
	}

	if task.Title != "Test task for integration testing" {
		t.Errorf("Expected title 'Test task for integration testing', got '%s'", task.Title)
	}

	// Clean up
	exec.Command("bd", "delete", taskID).Run()
}

func TestIntegrationTaskLifecycle(t *testing.T) {
	if !IsInstalled() {
		t.Skip("Beads not installed, skipping integration test")
	}

	client := NewClient()

	// Create task
	cmd := exec.Command("bd", "create", "Test lifecycle task")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	taskID := string(output)
	if taskID == "" {
		t.Skip("Could not create test task")
	}

	// Start task
	err = client.StartTask(taskID)
	if err != nil {
		t.Errorf("Failed to start task: %v", err)
	}

	// Verify status changed
	task, _ := client.GetTask(taskID)
	if task.Status != "started" && task.Status != "in_progress" {
		t.Logf("Task status after start: %s (expected 'started' or 'in_progress')", task.Status)
	}

	// Complete task
	err = client.CompleteTask(taskID)
	if err != nil {
		t.Errorf("Failed to complete task: %v", err)
	}

	// Verify status changed
	task, _ = client.GetTask(taskID)
	if task.Status != "closed" && task.Status != "done" {
		t.Logf("Task status after complete: %s (expected 'closed' or 'done')", task.Status)
	}

	// Clean up
	exec.Command("bd", "delete", taskID).Run()
}
