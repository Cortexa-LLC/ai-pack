package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/config"
	"github.com/cortexa-llc/ai-pack/internal/constants"
	"github.com/cortexa-llc/ai-pack/internal/streaming"
)

// createTaskWithCheckpoint is a test helper that creates a task directory, checkpoint, and metadata
func createTaskWithCheckpoint(t *testing.T, tmpDir, taskID string, cp *AgentCheckpoint, status, errorMsg string) {
	t.Helper()

	// Create task directory
	taskDir := filepath.Join(tmpDir, constants.BeadsDir, "tasks", taskID)
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.Fatalf("Failed to create task dir: %v", err)
	}

	// Write checkpoint
	if err := writeCheckpoint(tmpDir, taskID, cp); err != nil {
		t.Fatalf("Failed to write checkpoint: %v", err)
	}

	// Create task metadata
	metadata := map[string]interface{}{
		"task_id":      taskID,
		"status":       status,
		"project_root": tmpDir,
		"role":         cp.Role,
		"spawned_at":   cp.CreatedAt.Format(time.RFC3339),
		"updated_at":   time.Now().Format(time.RFC3339),
	}
	if errorMsg != "" {
		metadata["error"] = errorMsg
	}

	metaBytes, _ := json.MarshalIndent(metadata, "", "  ")
	metaPath := filepath.Join(taskDir, constants.MetadataFileName)
	if err := os.WriteFile(metaPath, metaBytes, 0644); err != nil {
		t.Fatalf("Failed to write metadata: %v", err)
	}
}

// TestResumeTimeoutTask tests resuming a task that failed due to timeout
func TestResumeTimeoutTask(t *testing.T) {
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

	// Create a checkpoint with timeout reason
	taskID := "test-task-timeout"
	cp := &AgentCheckpoint{
		TaskID:        taskID,
		CreatedAt:     time.Now(),
		Turn:          5,
		PartialResult: "Partial work done before timeout",
		ResumeReason:  "timeout",
		Role:          "engineer",
		ProjectRoot:   tmpDir,
		Model:         "claude-3-5-sonnet-20241022",
		BudgetUsed:    50000,
		BudgetLimit:   100000,
		Messages:      []streaming.Message{},
	}

	createTaskWithCheckpoint(t, tmpDir, taskID, cp, constants.StatusFailed, "TIMEOUT: Task exceeded time deadline")

	// Test resume endpoint
	reqBody := `{"extend_timeout":"15m"}`
	req := httptest.NewRequest(http.MethodPost, "/a2a/resume/"+taskID, strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.HandleResumeTask(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		t.Errorf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body[:n]))
		return
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !result["success"].(bool) {
		t.Error("Expected success=true")
	}
	if result["task_id"].(string) != taskID {
		t.Errorf("Expected task_id=%s, got %s", taskID, result["task_id"])
	}
}

// TestResumeTokenBudgetTask tests resuming a task that was paused due to token budget
func TestResumeTokenBudgetTask(t *testing.T) {
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

	// Create a checkpoint with token_budget reason
	taskID := "test-task-budget"
	cp := &AgentCheckpoint{
		TaskID:        taskID,
		CreatedAt:     time.Now(),
		Turn:          10,
		PartialResult: "Work done before budget exhaustion",
		ResumeReason:  "token_budget",
		Role:          "engineer",
		ProjectRoot:   tmpDir,
		Model:         "claude-3-5-sonnet-20241022",
		BudgetUsed:    100000,
		BudgetLimit:   100000,
		Messages:      []streaming.Message{},
	}

	createTaskWithCheckpoint(t, tmpDir, taskID, cp, constants.StatusPaused, "")

	// Test resume with budget extension
	reqBody := `{"extend_budget":50000}`
	req := httptest.NewRequest(http.MethodPost, "/a2a/resume/"+taskID, strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.HandleResumeTask(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		t.Errorf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body[:n]))
		return
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !result["success"].(bool) {
		t.Error("Expected success=true")
	}
	if int64(result["extend_budget"].(float64)) != 50000 {
		t.Errorf("Expected extend_budget=50000, got %v", result["extend_budget"])
	}
}

// TestResumeWithBothExtensions tests resuming with both timeout and budget extensions
func TestResumeWithBothExtensions(t *testing.T) {
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

	// Create a checkpoint
	taskID := "test-task-both"
	cp := &AgentCheckpoint{
		TaskID:        taskID,
		CreatedAt:     time.Now(),
		Turn:          3,
		PartialResult: "Some work done",
		ResumeReason:  "timeout",
		Role:          "engineer",
		ProjectRoot:   tmpDir,
		Model:         "claude-3-5-sonnet-20241022",
		BudgetUsed:    75000,
		BudgetLimit:   100000,
		Messages:      []streaming.Message{},
	}

	createTaskWithCheckpoint(t, tmpDir, taskID, cp, constants.StatusFailed, "TIMEOUT: Task exceeded time deadline")

	// Test resume with both extensions
	reqBody := `{"extend_timeout":"30m","extend_budget":25000}`
	req := httptest.NewRequest(http.MethodPost, "/a2a/resume/"+taskID, strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.HandleResumeTask(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		t.Errorf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body[:n]))
		return
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !result["success"].(bool) {
		t.Error("Expected success=true")
	}
	if result["extend_timeout"].(string) != "30m" {
		t.Errorf("Expected extend_timeout=30m, got %s", result["extend_timeout"])
	}
	if int64(result["extend_budget"].(float64)) != 25000 {
		t.Errorf("Expected extend_budget=25000, got %v", result["extend_budget"])
	}
}

// TestResumeDefaultReset tests that resume without flags resets to defaults
func TestResumeDefaultReset(t *testing.T) {
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

	// Create a checkpoint
	taskID := "test-task-default"
	cp := &AgentCheckpoint{
		TaskID:        taskID,
		CreatedAt:     time.Now(),
		Turn:          5,
		PartialResult: "Work done",
		ResumeReason:  "token_budget",
		Role:          "engineer",
		ProjectRoot:   tmpDir,
		Model:         "claude-3-5-sonnet-20241022",
		BudgetUsed:    100000,
		BudgetLimit:   100000,
		Messages:      []streaming.Message{},
	}

	createTaskWithCheckpoint(t, tmpDir, taskID, cp, constants.StatusPaused, "")

	// Test resume without any extensions (should reset to defaults)
	req := httptest.NewRequest(http.MethodPost, "/a2a/resume/"+taskID, bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.HandleResumeTask(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		t.Errorf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body[:n]))
		return
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !result["success"].(bool) {
		t.Error("Expected success=true")
	}

	// Verify no extensions
	extendBudget := result["extend_budget"]
	if extendBudget != nil && extendBudget != float64(0) {
		t.Errorf("Expected extend_budget=0, got %v", extendBudget)
	}
}

// TestResumeRejectsNonPausedTask tests that resume rejects tasks not in paused or timeout-failed state
func TestResumeRejectsNonPausedTask(t *testing.T) {
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

	// Create a checkpoint
	taskID := "test-task-completed"
	cp := &AgentCheckpoint{
		TaskID:        taskID,
		CreatedAt:     time.Now(),
		Turn:          5,
		PartialResult: "Work done",
		ResumeReason:  "token_budget",
		Role:          "engineer",
		ProjectRoot:   tmpDir,
		Model:         "claude-3-5-sonnet-20241022",
		Messages:      []streaming.Message{},
	}

	// Create task metadata with COMPLETED status (not resumable)
	createTaskWithCheckpoint(t, tmpDir, taskID, cp, constants.StatusCompleted, "")

	// Test resume - should fail
	req := httptest.NewRequest(http.MethodPost, "/a2a/resume/"+taskID, bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.HandleResumeTask(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

// Note: Helper functions setupTestDir and clearAuthEnvVars are defined in test_helpers.go
