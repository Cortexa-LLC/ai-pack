package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cortexa-llc/ai-pack/internal/config"
	"github.com/cortexa-llc/ai-pack/internal/taskdb"
)

// TestHandleTasksList_Empty tests task list with no active tasks
func TestHandleTasksListEmpty(t *testing.T) {
	// Setup
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	// Create test-isolated task DB
	dbPath := filepath.Join(tmpDir, "tasks.db")
	db, err := taskdb.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open taskDB: %v", err)
	}
	defer db.Close()

	cfg := config.DefaultConfig()
	cfg.API.Mode = "direct"
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	// Override with test-isolated DB
	server.taskDB = db

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

	// Open taskDB
	dbPath := filepath.Join(tmpDir, "tasks.db")
	db, err := taskdb.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open taskDB: %v", err)
	}
	defer db.Close()

	cfg := config.DefaultConfig()
	cfg.API.Mode = "direct"
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	server.taskDB = db

	// Add tasks to taskDB
	task1 := &taskdb.Task{
		ID:              "task-1",
		
		ProjectRoot:     tmpDir,
		Role:            "engineer",
		TaskDescription: "Test task 1",
	}
	if err := db.CreateTask(task1); err != nil {
		t.Fatalf("Failed to create task-1: %v", err)
	}

	task2 := &taskdb.Task{
		ID:              "task-2",
		
		ProjectRoot:     tmpDir,
		Role:            "tester",
		TaskDescription: "Test task 2",
	}
	if err := db.CreateTask(task2); err != nil {
		t.Fatalf("Failed to create task-2: %v", err)
	}

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
	firstTask, ok := tasks[0].(map[string]interface{})
	if !ok {
		t.Fatal("Expected task to be an object")
	}

	// Check required fields
	if taskID, ok := firstTask["task_id"].(string); !ok || (taskID != "task-1" && taskID != "task-2") {
		t.Errorf("Expected valid task_id, got %v", firstTask["task_id"])
	}

	if _, ok := firstTask["status"].(string); !ok {
		t.Errorf("Expected status to be a string, got %T", firstTask["status"])
	}

	if _, ok := firstTask["role"].(string); !ok {
		t.Errorf("Expected role to be a string, got %T", firstTask["role"])
	}

	if _, ok := firstTask["description"].(string); !ok {
		t.Errorf("Expected description to be a string, got %T", firstTask["description"])
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

// TestHandleA2AExecute_PathTraversal_ProjectRoot tests that path traversal attempts
// via project_root are rejected before any filesystem operations occur.
func TestHandleA2AExecutePathTraversalProjectRoot(t *testing.T) {
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_KEY", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	cfg := config.DefaultConfig()
	cfg.API.Mode = "direct"
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	cases := []struct {
		name        string
		projectRoot string
	}{
		{"relative path", "../../etc"},
		{"relative with dot", "./relative/path"},
		{"traversal only", "../.."},
		{"mixed traversal", "/valid/../../../etc"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := makeA2AExecuteBody(t, map[string]interface{}{
				"role":         "engineer",
				"task":         "ai-pack-abc-20260101000000-test",
				"project_root": tc.projectRoot,
			})

			req := httptest.NewRequest(http.MethodPost, "/a2a/execute", body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.handleA2AExecute(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			// Must return an error response (not 200 OK triggering filesystem ops)
			var rpcResp map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if _, hasError := rpcResp["error"]; !hasError {
				t.Errorf("expected JSON-RPC error for project_root=%q, got: %v", tc.projectRoot, rpcResp)
			}
		})
	}
}

// TestHandleA2AExecute_PathTraversal_Role tests that path traversal attempts
// via role are rejected before any filesystem operations occur.
func TestHandleA2AExecutePathTraversalRole(t *testing.T) {
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_KEY", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	cfg := config.DefaultConfig()
	cfg.API.Mode = "direct"
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	cases := []struct {
		name string
		role string
	}{
		{"unix traversal", "../../malicious-role"},
		{"windows traversal", `..\..\malicious-role`},
		{"leading slash", "/etc/malicious"},
		{"embedded slash", "roles/../../malicious"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := makeA2AExecuteBody(t, map[string]interface{}{
				"role":         tc.role,
				"task":         "ai-pack-abc-20260101000000-test",
				"project_root": tmpDir,
			})

			req := httptest.NewRequest(http.MethodPost, "/a2a/execute", body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.handleA2AExecute(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			var rpcResp map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if _, hasError := rpcResp["error"]; !hasError {
				t.Errorf("expected JSON-RPC error for role=%q, got: %v", tc.role, rpcResp)
			}
		})
	}
}

// makeA2AExecuteBody creates a JSON-RPC "tasks/execute" request body for testing.
func makeA2AExecuteBody(t *testing.T, params map[string]interface{}) *strings.Reader {
	t.Helper()
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "tasks/execute",
		"id":      "test-id",
		"params":  params,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}
	return strings.NewReader(string(data))
}
