package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/beads"
	"github.com/cortexa-llc/ai-pack/internal/config"
	"github.com/cortexa-llc/ai-pack/internal/constants"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
)

// mockBeadsTaskGetter is a test double for beadsTaskGetter.
type mockBeadsTaskGetter struct {
	tasks  map[string]*beads.Task
	errors map[string]error
}

func (m *mockBeadsTaskGetter) GetTask(taskID string) (*beads.Task, error) {
	if err, ok := m.errors[taskID]; ok {
		return nil, err
	}
	if t, ok := m.tasks[taskID]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("task %q not found in mock", taskID)
}

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
	monitoring.GlobalMetrics.RecordTurnTokens("test-task", 1, 100, 200, 50)

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

// TestGetAllTasks_CancelledBeadsTaskShowsCancelled verifies that when CloseTask
// sets execution.Status = "closed" (the actual cancel sequence), the task is
// reported as "cancelled" even though beads shows "open".
// It also verifies that a genuinely running task (Status "in_progress") whose
// beads entry was incorrectly reset to "open" by orphan cleanup continues to
// show as "in_progress" — not downgraded to "cancelled".
func TestGetAllTasks_CancelledBeadsTaskShowsCancelled(t *testing.T) {
	monitoring.InitMetrics()

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

	beadsID := "ai-pack-test-cancelled"

	// CloseTask sets execution.Status = "closed" before removing from activeTasks.
	// This is the in-flight window a GetAllTasks call would observe for a cancelled task.
	execution := &TaskExecution{
		TaskID:    "task-cancelled-test",
		Role:      "engineer",
		Task:      "some cancelled task",
		Status:    "closed",
		StartTime: time.Now(),
		metadata:  map[string]string{"beads_task_id": beadsID},
	}

	// Also add a legitimately running task whose beads was incorrectly reset to "open"
	// by orphan cleanup from a second server instance. It must stay "in_progress".
	beadsIDRunning := "ai-pack-test-running"
	executionRunning := &TaskExecution{
		TaskID:    "task-running-beads-open",
		Role:      "engineer",
		Task:      "running task with orphaned beads",
		Status:    "in_progress",
		StartTime: time.Now(),
		metadata:  map[string]string{"beads_task_id": beadsIDRunning},
	}

	server.mu.Lock()
	server.activeTasks["task-cancelled-test"] = execution
	server.activeTasks["task-running-beads-open"] = executionRunning
	server.mu.Unlock()

	adapter := NewGraphQLAdapter(server)
	adapter.taskGetter = &mockBeadsTaskGetter{
		tasks: map[string]*beads.Task{
			beadsID:        {ID: beadsID, Status: "open"},
			beadsIDRunning: {ID: beadsIDRunning, Status: "open"},
		},
	}

	tasks := adapter.GetAllTasks()

	got, ok := tasks["task-cancelled-test"]
	if !ok {
		t.Fatal("Expected task-cancelled-test in GetAllTasks result, but it was missing")
	}
	if got.Status != "cancelled" {
		t.Errorf("Expected status 'cancelled' for closed execution, got %q", got.Status)
	}

	gotRunning, ok := tasks["task-running-beads-open"]
	if !ok {
		t.Fatal("Expected task-running-beads-open in GetAllTasks result, but it was missing")
	}
	if gotRunning.Status != "in_progress" {
		t.Errorf("Expected status 'in_progress' for running task with orphaned beads, got %q — must not downgrade running task", gotRunning.Status)
	}
}

// TestParseFolderTimestamp validates the helper that extracts the timestamp
// from an execution folder name so that executions can be sorted by their
// embedded creation time rather than by filesystem mtime.
func TestParseFolderTimestamp(t *testing.T) {
	cases := []struct {
		name       string
		folder     string
		prefix     string
		wantZero   bool
		wantParsed time.Time
	}{
		{
			name:       "valid folder",
			folder:     "ai-pack-x008-20260228-130230",
			prefix:     "ai-pack-x008-",
			wantParsed: time.Date(2026, 2, 28, 13, 2, 30, 0, time.UTC),
		},
		{
			name:     "malformed suffix returns zero",
			folder:   "ai-pack-x008-not-a-timestamp",
			prefix:   "ai-pack-x008-",
			wantZero: true,
		},
		{
			name:     "prefix not stripped",
			folder:   "other-20260228-130230",
			prefix:   "ai-pack-x008-",
			wantZero: true, // suffix after TrimPrefix is the full folder name
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseFolderTimestamp(tc.folder, tc.prefix)
			if tc.wantZero {
				if !got.IsZero() {
					t.Errorf("expected zero time, got %v", got)
				}
			} else {
				if !got.Equal(tc.wantParsed) {
					t.Errorf("expected %v, got %v", tc.wantParsed, got)
				}
			}
		})
	}
}

// TestFindMostRecentExecution_UsesNameTimestampNotMtime verifies that
// findMostRecentExecution picks the execution folder whose NAME contains the
// newest timestamp, not the one with the newest filesystem mtime.
//
// Regression test for: running tasks showing as "failed" after a retry.
//
// Setup:
//   - Old folder (20260228-120000) with status=failed, marked superseded.
//     Its mtime is updated AFTER the new folder is created when the superseded
//     marker is written — making it appear newer by mtime.
//   - New folder (20260228-130000) with status=in_progress.
//
// Expected: new folder is selected.
func TestFindMostRecentExecution_UsesNameTimestampNotMtime(t *testing.T) {
	monitoring.InitMetrics()

	projectRoot := t.TempDir()
	beadsID := "ai-pack-x008"
	tasksDir := filepath.Join(projectRoot, constants.BeadsDir, "tasks")

	newFolderName := beadsID + "-20260228-130000"
	oldFolderName := beadsID + "-20260228-120000"

	// Write metadata for old (failed, superseded) execution
	oldDir := filepath.Join(tasksDir, oldFolderName)
	if err := os.MkdirAll(oldDir, 0o755); err != nil {
		t.Fatal(err)
	}
	oldMeta := map[string]interface{}{
		"task_id":    oldFolderName,
		"status":     "failed",
		"role":       "engineer",
		"task":       "old task",
		"superseded": true,
		"metadata":   map[string]string{"beads_task_id": beadsID},
	}
	writeMetaJSON(t, oldDir, oldMeta)

	// Write metadata for new (in_progress) execution
	newDir := filepath.Join(tasksDir, newFolderName)
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		t.Fatal(err)
	}
	newMeta := map[string]interface{}{
		"task_id":  newFolderName,
		"status":   "in_progress",
		"role":     "engineer",
		"task":     "new task",
		"metadata": map[string]string{"beads_task_id": beadsID},
	}
	writeMetaJSON(t, newDir, newMeta)

	// Simulate superseded-marker write updating old folder's mtime to NOW
	// (i.e. newer than the new execution folder).
	// Sleep briefly then touch old folder to force its mtime to be newer.
	time.Sleep(5 * time.Millisecond)
	oldMetaPath := filepath.Join(oldDir, constants.MetadataFileName)
	if err := os.Chtimes(oldMetaPath, time.Now(), time.Now()); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		API: config.APIConfig{
			MaxTokens:      4096,
			AnthropicModel: "claude-3-5-sonnet-20241022",
		},
	}
	server, err := NewAgentServer(projectRoot, 1, 4096, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatal(err)
	}
	adapter := NewGraphQLAdapter(server)
	adapter.taskGetter = &mockBeadsTaskGetter{
		tasks: map[string]*beads.Task{
			beadsID: {ID: beadsID, Status: "in_progress"},
		},
	}

	result := adapter.findMostRecentExecution(projectRoot, beadsID)
	if result == nil {
		t.Fatal("findMostRecentExecution returned nil")
	}

	if result.TaskID != newFolderName && result.TaskID != beadsID {
		t.Errorf("expected task from new execution folder, got TaskID=%q", result.TaskID)
	}

	if result.Status != "in_progress" {
		t.Errorf("expected status 'in_progress' from new execution, got %q — running task shown as failed!", result.Status)
	}
}

// TestGetTaskStatus_FindsActiveTaskByBeadsIDPrefix verifies that GetTaskStatus
// correctly locates an actively-running task when called with the short Beads
// task ID (e.g. "ai-pack-x008") even though the activeTasks map is keyed by
// the full timestamped execution ID (e.g. "ai-pack-x008-20260228-130000").
func TestGetTaskStatus_FindsActiveTaskByBeadsIDPrefix(t *testing.T) {
	monitoring.InitMetrics()

	cfg := &config.Config{
		API: config.APIConfig{
			MaxTokens:      4096,
			AnthropicModel: "claude-3-5-sonnet-20241022",
		},
	}
	server, err := NewAgentServer("/tmp", 1, 4096, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatal(err)
	}

	beadsID := "ai-pack-x008"
	execID := beadsID + "-20260228-130000"

	execution := &TaskExecution{
		TaskID:    execID,
		Role:      "engineer",
		Task:      "running task",
		Status:    "in_progress",
		StartTime: time.Now(),
		metadata:  map[string]string{"beads_task_id": beadsID},
	}
	server.mu.Lock()
	server.activeTasks[execID] = execution
	server.mu.Unlock()

	adapter := NewGraphQLAdapter(server)
	adapter.taskGetter = &mockBeadsTaskGetter{
		tasks: map[string]*beads.Task{
			beadsID: {ID: beadsID, Status: "in_progress"},
		},
	}

	// Query with the SHORT beads ID — should find the active execution
	taskInfo, err := adapter.GetTaskStatus(beadsID)
	if err != nil {
		t.Fatalf("GetTaskStatus(%q) returned error: %v", beadsID, err)
	}
	if taskInfo == nil {
		t.Fatal("GetTaskStatus returned nil")
	}
	if taskInfo.Status != "in_progress" {
		t.Errorf("expected status 'in_progress', got %q", taskInfo.Status)
	}
}

// writeMetaJSON is a test helper that writes metadata as JSON to the
// constants.MetadataFileName file inside dir.
func writeMetaJSON(t *testing.T, dir string, meta map[string]interface{}) {
	t.Helper()
	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, constants.MetadataFileName), data, 0o644); err != nil {
		t.Fatal(err)
	}
}
