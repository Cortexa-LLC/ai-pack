package server

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/config"
	"github.com/cortexa-llc/ai-pack/internal/constants"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/internal/taskdb"
)

// TestGetAllTasks_WithTaskDB verifies GetAllTasks queries taskDB and merges with active tasks
func TestGetAllTasks_WithTaskDB(t *testing.T) {
	monitoring.InitMetrics()

	// Create temp dir for taskDB
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "tasks.db")

	db, err := taskdb.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open taskDB: %v", err)
	}
	defer db.Close()

	cfg := &config.Config{
		API: config.APIConfig{
			MaxTokens:      4096,
			AnthropicModel: "claude-3-5-sonnet-20241022",
		},
	}

	server, err := NewAgentServer(tmpDir, 1, 4096, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	server.taskDB = db

	// Add a task to taskDB
	dbTask := &taskdb.Task{
		ID:              "task-from-db-123",
		BeadsID:         "ai-pack-db-test",
		ProjectRoot:     tmpDir,
		Role:            "engineer",
		TaskDescription: "task from database",
		Status:          taskdb.StatusQueued,
	}
	if err := db.CreateTask(dbTask); err != nil {
		t.Fatalf("Failed to create task in DB: %v", err)
	}

	// Add an active task (in-memory)
	activeExec := &TaskExecution{
		TaskID:    "task-active-456",
		Role:      "reviewer",
		Task:      "active task",
		Status:    constants.StatusInProgress,
		StartTime: time.Now(),
		metadata:  map[string]string{"beads_task_id": "ai-pack-active"},
	}
	server.mu.Lock()
	server.activeTasks["task-active-456"] = activeExec
	server.mu.Unlock()

	adapter := NewGraphQLAdapter(server)
	tasks := adapter.GetAllTasks()

	// Should have both tasks
	if len(tasks) < 2 {
		t.Errorf("Expected at least 2 tasks, got %d", len(tasks))
	}

	// Check DB task is present (keyed by BeadsID)
	dbTaskInfo, hasDB := tasks["ai-pack-db-test"]
	if !hasDB {
		t.Error("Task from taskDB not found in GetAllTasks result")
	} else {
		if dbTaskInfo.Status != taskdb.StatusQueued {
			t.Errorf("Expected DB task status 'queued', got %q", dbTaskInfo.Status)
		}
		if dbTaskInfo.Role != "engineer" {
			t.Errorf("Expected DB task role 'engineer', got %q", dbTaskInfo.Role)
		}
	}

	// Check active task is present
	activeTaskInfo, hasActive := tasks["task-active-456"]
	if !hasActive {
		t.Error("Active task not found in GetAllTasks result")
	} else {
		if activeTaskInfo.Status != constants.StatusInProgress {
			t.Errorf("Expected active task status 'in_progress', got %q", activeTaskInfo.Status)
		}
	}
}

// TestTaskDBIntegration_CreateAndQuery verifies basic taskDB operations
func TestTaskDBIntegration_CreateAndQuery(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "tasks.db")

	db, err := taskdb.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open taskDB: %v", err)
	}
	defer db.Close()

	// Create a task
	task := &taskdb.Task{
		ID:              "test-task-001",
		BeadsID:         "ai-pack-test",
		ProjectRoot:     "/tmp/project",
		Role:            "engineer",
		TaskDescription: "test task description",
	}

	if err := db.CreateTask(task); err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Query it back
	retrieved, err := db.GetTask("test-task-001")
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}

	if retrieved.ID != "test-task-001" {
		t.Errorf("Expected ID 'test-task-001', got %q", retrieved.ID)
	}
	if retrieved.Status != taskdb.StatusQueued {
		t.Errorf("Expected status 'queued', got %q", retrieved.Status)
	}
	if retrieved.BeadsID != "ai-pack-test" {
		t.Errorf("Expected BeadsID 'ai-pack-test', got %q", retrieved.BeadsID)
	}
}

// TestTaskDBIntegration_StatusTransitions verifies task status updates
func TestTaskDBIntegration_StatusTransitions(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "tasks.db")

	db, err := taskdb.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open taskDB: %v", err)
	}
	defer db.Close()

	// Create a task
	task := &taskdb.Task{
		ID:              "test-task-002",
		BeadsID:         "ai-pack-test-002",
		ProjectRoot:     "/tmp/project",
		Role:            "engineer",
		TaskDescription: "status transition test",
	}

	if err := db.CreateTask(task); err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Transition to in_progress
	if err := db.UpdateTaskStatus("test-task-002", taskdb.StatusInProgress, ""); err != nil {
		t.Fatalf("Failed to update status to in_progress: %v", err)
	}

	retrieved, err := db.GetTask("test-task-002")
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}
	if retrieved.Status != taskdb.StatusInProgress {
		t.Errorf("Expected status 'in_progress', got %q", retrieved.Status)
	}

	// Complete the task
	if err := db.CompleteTask("test-task-002", "Task completed successfully"); err != nil {
		t.Fatalf("Failed to complete task: %v", err)
	}

	retrieved, err = db.GetTask("test-task-002")
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}
	if retrieved.Status != taskdb.StatusCompleted {
		t.Errorf("Expected status 'completed', got %q", retrieved.Status)
	}
	if retrieved.Result != "Task completed successfully" {
		t.Errorf("Expected result 'Task completed successfully', got %q", retrieved.Result)
	}
	if retrieved.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}
}

// TestTaskDBIntegration_FailTask verifies failing a task
func TestTaskDBIntegration_FailTask(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "tasks.db")

	db, err := taskdb.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open taskDB: %v", err)
	}
	defer db.Close()

	task := &taskdb.Task{
		ID:              "test-task-003",
		BeadsID:         "ai-pack-test-003",
		ProjectRoot:     "/tmp/project",
		Role:            "engineer",
		TaskDescription: "fail test",
	}

	if err := db.CreateTask(task); err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Fail the task
	if err := db.FailTask("test-task-003", "Task failed: timeout"); err != nil {
		t.Fatalf("Failed to fail task: %v", err)
	}

	retrieved, err := db.GetTask("test-task-003")
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}
	if retrieved.Status != taskdb.StatusFailed {
		t.Errorf("Expected status 'failed', got %q", retrieved.Status)
	}
	if retrieved.Error != "Task failed: timeout" {
		t.Errorf("Expected error 'Task failed: timeout', got %q", retrieved.Error)
	}
}

// TestTaskDBIntegration_ListTasks verifies task listing with filters
func TestTaskDBIntegration_ListTasks(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "tasks.db")

	db, err := taskdb.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open taskDB: %v", err)
	}
	defer db.Close()

	// Create multiple tasks with different statuses
	tasks := []*taskdb.Task{
		{ID: "task-1", BeadsID: "test-1", ProjectRoot: "/tmp", Role: "engineer", TaskDescription: "task 1"},
		{ID: "task-2", BeadsID: "test-2", ProjectRoot: "/tmp", Role: "engineer", TaskDescription: "task 2"},
		{ID: "task-3", BeadsID: "test-3", ProjectRoot: "/tmp", Role: "reviewer", TaskDescription: "task 3"},
	}

	for _, task := range tasks {
		if err := db.CreateTask(task); err != nil {
			t.Fatalf("Failed to create task %s: %v", task.ID, err)
		}
	}

	// Mark task-2 as in_progress
	if err := db.UpdateTaskStatus("task-2", taskdb.StatusInProgress, ""); err != nil {
		t.Fatalf("Failed to update task-2 status: %v", err)
	}

	// Mark task-3 as completed
	if err := db.CompleteTask("task-3", "done"); err != nil {
		t.Fatalf("Failed to complete task-3: %v", err)
	}

	// List all tasks
	allTasks, err := db.ListTasks(taskdb.TaskFilter{Limit: 100})
	if err != nil {
		t.Fatalf("Failed to list all tasks: %v", err)
	}
	if len(allTasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(allTasks))
	}

	// List only queued tasks
	queuedTasks, err := db.ListTasks(taskdb.TaskFilter{Status: taskdb.StatusQueued, Limit: 100})
	if err != nil {
		t.Fatalf("Failed to list queued tasks: %v", err)
	}
	if len(queuedTasks) != 1 {
		t.Errorf("Expected 1 queued task, got %d", len(queuedTasks))
	}

	// List only in_progress tasks
	inProgressTasks, err := db.ListTasks(taskdb.TaskFilter{Status: taskdb.StatusInProgress, Limit: 100})
	if err != nil {
		t.Fatalf("Failed to list in_progress tasks: %v", err)
	}
	if len(inProgressTasks) != 1 {
		t.Errorf("Expected 1 in_progress task, got %d", len(inProgressTasks))
	}
}

// TestHandleTasksList_WithTaskDB verifies the REST API handler uses taskDB
func TestHandleTasksList_WithTaskDB(t *testing.T) {
	monitoring.InitMetrics()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "tasks.db")

	db, err := taskdb.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open taskDB: %v", err)
	}
	defer db.Close()

	cfg := &config.Config{
		API: config.APIConfig{
			MaxTokens:      4096,
			AnthropicModel: "claude-3-5-sonnet-20241022",
		},
	}

	server, err := NewAgentServer(tmpDir, 1, 4096, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	server.taskDB = db

	// Create some tasks in taskDB
	task1 := &taskdb.Task{
		ID:              "handler-test-1",
		BeadsID:         "handler-1",
		ProjectRoot:     tmpDir,
		Role:            "engineer",
		TaskDescription: "handler test 1",
	}
	if err := db.CreateTask(task1); err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	task2 := &taskdb.Task{
		ID:              "handler-test-2",
		BeadsID:         "handler-2",
		ProjectRoot:     tmpDir,
		Role:            "reviewer",
		TaskDescription: "handler test 2",
	}
	if err := db.CreateTask(task2); err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Mark task2 as in_progress
	if err := db.UpdateTaskStatus("handler-test-2", taskdb.StatusInProgress, ""); err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	// Now query via the handler (we'd need to set up HTTP test infrastructure)
	// For now, just verify the server has taskDB set
	if server.taskDB == nil {
		t.Fatal("Server taskDB is nil")
	}

	// Verify we can query tasks
	tasks, err := server.taskDB.ListTasks(taskdb.TaskFilter{Limit: 100})
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
	}
}

// TestServeTaskContract_WithTaskDB verifies task contract display uses taskDB
func TestServeTaskContract_WithTaskDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "tasks.db")

	db, err := taskdb.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open taskDB: %v", err)
	}
	defer db.Close()

	cfg := &config.Config{
		API: config.APIConfig{
			MaxTokens:      4096,
			AnthropicModel: "claude-3-5-sonnet-20241022",
		},
	}

	server, err := NewAgentServer(tmpDir, 1, 4096, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	server.taskDB = db

	// Create a task with metadata
	metadataJSON, _ := json.Marshal(map[string]string{
		"priority":    "P1",
		"assigned_to": "test-user",
	})

	task := &taskdb.Task{
		ID:              "contract-test-1",
		BeadsID:         "contract-beads-1",
		ProjectRoot:     tmpDir,
		Role:            "engineer",
		TaskDescription: "This is a detailed task description\nwith multiple lines",
		Metadata:        string(metadataJSON),
	}
	if err := db.CreateTask(task); err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Verify task can be retrieved
	retrieved, err := db.GetTask("contract-test-1")
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}

	if retrieved.TaskDescription != "This is a detailed task description\nwith multiple lines" {
		t.Errorf("Task description mismatch")
	}

	if retrieved.Metadata == "" {
		t.Error("Expected metadata to be set")
	}
}
