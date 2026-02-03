package server

import (
	"testing"
	"time"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/config"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
)

// TestGetMetrics verifies the GraphQL adapter returns correct metrics
func TestGetMetrics(t *testing.T) {
	// Setup: Initialize monitoring
	monitoring.InitMetrics()

	// Create a test server
	cfg := &config.Config{
		API: config.APIConfig{
			MaxTokens:      4096,
			AnthropicModel: "claude-3-5-sonnet-20241022",
		},
	}

	server, err := NewAgentServer("/tmp", 1, 4096, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create adapter
	adapter := NewGraphQLAdapter(server)

	// Record some metrics
	monitoring.GlobalMetrics.IncrementTasksSpawned()
	monitoring.GlobalMetrics.IncrementTasksCompleted(1000)
	monitoring.GlobalMetrics.IncrementAPICallsSuccess()
	monitoring.GlobalMetrics.RecordTokenUsage("test-task", 100, 200, 1)

	// Act: Get metrics through adapter
	metrics := adapter.GetMetrics()

	// Assert: Verify metrics are returned correctly
	if metrics == nil {
		t.Fatal("GetMetrics returned nil")
	}

	// Verify task counts
	if metrics.TasksSpawned < 1 {
		t.Errorf("Expected TasksSpawned >= 1, got %d", metrics.TasksSpawned)
	}

	if metrics.TasksCompleted < 1 {
		t.Errorf("Expected TasksCompleted >= 1, got %d", metrics.TasksCompleted)
	}

	// Verify API calls
	if metrics.APICalls < 1 {
		t.Errorf("Expected APICalls >= 1, got %d", metrics.APICalls)
	}

	// Verify token metrics
	if metrics.TotalTokens < 300 {
		t.Errorf("Expected TotalTokens >= 300, got %d", metrics.TotalTokens)
	}
}

// TestGetActiveTasksEmpty verifies empty task list
func TestGetActiveTasksEmpty(t *testing.T) {
	cfg := &config.Config{
		API: config.APIConfig{
			MaxTokens:      4096,
			AnthropicModel: "claude-3-5-sonnet-20241022",
		},
	}

	server, err := NewAgentServer("/tmp", 1, 4096, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	adapter := NewGraphQLAdapter(server)
	tasks := adapter.GetActiveTasks()

	if tasks == nil {
		t.Fatal("GetActiveTasks returned nil")
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(tasks))
	}
}

// TestGetActiveTasksWithTask verifies task tracking
func TestGetActiveTasksWithTask(t *testing.T) {
	cfg := &config.Config{
		API: config.APIConfig{
			MaxTokens:      4096,
			AnthropicModel: "claude-3-5-sonnet-20241022",
		},
	}

	server, err := NewAgentServer("/tmp", 1, 4096, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Add a task to the server
	taskID := "test-task-123"
	execution := &TaskExecution{
		TaskID:    taskID,
		Role:      "engineer",
		Task:      "test task",
		Status:    "in_progress",
		StartTime: time.Now(),
		Result:    "",
		Error:     "",
		metadata:  make(map[string]string),
	}

	server.mu.Lock()
	server.activeTasks[taskID] = execution
	server.mu.Unlock()

	adapter := NewGraphQLAdapter(server)
	tasks := adapter.GetActiveTasks()

	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}

	task, exists := tasks[taskID]
	if !exists {
		t.Fatal("Task not found in results")
	}

	if task.TaskID != taskID {
		t.Errorf("Expected TaskID %s, got %s", taskID, task.TaskID)
	}

	if task.Role != "engineer" {
		t.Errorf("Expected Role 'engineer', got %s", task.Role)
	}

	if task.Status != "in_progress" {
		t.Errorf("Expected Status 'in_progress', got %s", task.Status)
	}
}
