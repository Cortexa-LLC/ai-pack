package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/config"
)

// TestHandleTasksList_Empty tests task list with no active tasks
func TestHandleTasksListEmpty(t *testing.T) {
	// Setup
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	cfg := config.DefaultConfig()
	cfg.API.Mode = "direct"
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create test HTTP request
	req := httptest.NewRequest(http.MethodGet, "/a2a/tasks", nil)
	w := httptest.NewRecorder()

	// Execute
	server.HandleTasksList(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Verify status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	// Verify Content-Type
	if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify structure
	if count, ok := result["count"].(float64); !ok || count != 0 {
		t.Errorf("Expected count 0 for empty server, got %v", result["count"])
	}
}

// TestHandleTasksList_WithActiveTasks tests task list with active tasks
func TestHandleTasksListWithActiveTasks(t *testing.T) {
	// Setup
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	cfg := config.DefaultConfig()
	cfg.API.Mode = "direct"
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Add mock tasks to server's active tasks
	server.mu.Lock()
	server.activeTasks["task-1"] = &TaskExecution{
		TaskID:      "task-1",
		Role:        "engineer",
		Task:        "Test task 1",
		Status:      "in_progress",
		ProjectRoot: "/test/project",
		metadata: map[string]string{
			"beads_task_id": "test-123",
		},
	}
	server.activeTasks["task-2"] = &TaskExecution{
		TaskID: "task-2",
		Role:   "tester",
		Task:   "Test task 2",
		Status: "queued",
	}
	server.mu.Unlock()

	// Create test HTTP request
	req := httptest.NewRequest(http.MethodGet, "/a2a/tasks", nil)
	w := httptest.NewRecorder()

	// Execute
	server.HandleTasksList(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify count
	if count, ok := result["count"].(float64); !ok || count != 2 {
		t.Errorf("Expected count 2, got %v", result["count"])
	}

	// Verify tasks array
	tasks, ok := result["tasks"].([]interface{})
	if !ok {
		t.Fatal("Expected tasks to be an array")
	}

	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
	}

	// Verify first task structure
	task1, ok := tasks[0].(map[string]interface{})
	if !ok {
		t.Fatal("Expected task to be an object")
	}

	// Check required fields
	if taskID, ok := task1["task_id"].(string); !ok || (taskID != "task-1" && taskID != "task-2") {
		t.Errorf("Expected valid task_id, got %v", task1["task_id"])
	}

	if _, ok := task1["status"].(string); !ok {
		t.Errorf("Expected status to be a string, got %T", task1["status"])
	}

	if _, ok := task1["role"].(string); !ok {
		t.Errorf("Expected role to be a string, got %T", task1["role"])
	}

	if _, ok := task1["description"].(string); !ok {
		t.Errorf("Expected description to be a string, got %T", task1["description"])
	}
}

// TestHandleTasksList_MethodNotAllowed tests that non-GET requests are rejected
func TestHandleTasksListMethodNotAllowed(t *testing.T) {
	// Setup
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	cfg := config.DefaultConfig()
	cfg.API.Mode = "direct"
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test POST (should be rejected)
	req := httptest.NewRequest(http.MethodPost, "/a2a/tasks", nil)
	w := httptest.NewRecorder()

	server.HandleTasksList(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code 405 for POST, got %d", resp.StatusCode)
	}
}

// TestHandleStatusGET_Success tests GET status endpoint with existing task
func TestHandleStatusGETSuccess(t *testing.T) {
	// Setup
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	cfg := config.DefaultConfig()
	cfg.API.Mode = "direct"
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Add mock task
	server.mu.Lock()
	server.activeTasks["test-task-123"] = &TaskExecution{
		TaskID: "test-task-123",
		Role:   "engineer",
		Task:   "Test task",
		Status: "in_progress",
	}
	server.mu.Unlock()

	// Create test HTTP request
	req := httptest.NewRequest(http.MethodGet, "/a2a/status/test-task-123", nil)
	w := httptest.NewRecorder()

	// Execute
	server.handleStatusGET(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Verify status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify fields
	if taskID, ok := result["task_id"].(string); !ok || taskID != "test-task-123" {
		t.Errorf("Expected task_id 'test-task-123', got %v", result["task_id"])
	}

	if status, ok := result["status"].(string); !ok || status != "in_progress" {
		t.Errorf("Expected status 'in_progress', got %v", result["status"])
	}
}

// TestHandleStatusGET_NotFound tests GET status endpoint with non-existent task
func TestHandleStatusGETNotFound(t *testing.T) {
	// Setup
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	cfg := config.DefaultConfig()
	cfg.API.Mode = "direct"
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create test HTTP request for non-existent task
	req := httptest.NewRequest(http.MethodGet, "/a2a/status/nonexistent-task", nil)
	w := httptest.NewRecorder()

	// Execute
	server.handleStatusGET(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Verify 404
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code 404, got %d", resp.StatusCode)
	}
}

// TestHandleStatusGET_MissingTaskID tests GET status endpoint without task ID
func TestHandleStatusGETMissingTaskID(t *testing.T) {
	// Setup
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	cfg := config.DefaultConfig()
	cfg.API.Mode = "direct"
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create test HTTP request without task ID
	req := httptest.NewRequest(http.MethodGet, "/a2a/status/", nil)
	w := httptest.NewRecorder()

	// Execute
	server.handleStatusGET(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Verify 400
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", resp.StatusCode)
	}
}

// TestHandleA2AStatus_GET tests that GET requests are routed to handleStatusGET
func TestHandleA2AStatusGET(t *testing.T) {
	// Setup
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	cfg := config.DefaultConfig()
	cfg.API.Mode = "direct"
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Add mock task
	server.mu.Lock()
	server.activeTasks["test-123"] = &TaskExecution{
		TaskID: "test-123",
		Role:   "engineer",
		Status: "completed",
	}
	server.mu.Unlock()

	// Create GET request
	req := httptest.NewRequest(http.MethodGet, "/a2a/status/test-123", nil)
	w := httptest.NewRecorder()

	// Execute
	server.HandleA2AStatus(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Should succeed (routed to handleStatusGET)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	// Verify it's JSON
	if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}
}
