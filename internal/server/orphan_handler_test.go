package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/config"
	"github.com/cortexa-llc/ai-pack/internal/constants"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/internal/taskdb"
)

// TestOrphanDetection_TaskDBInProgress verifies orphaned tasks are detected from taskDB
func TestOrphanDetection_TaskDBInProgress(t *testing.T) {
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

	// Create a task marked as in_progress in taskDB
	orphanTask := &taskdb.Task{
		ID:              "orphan-task-1",
		ProjectRoot:     tmpDir,
		Role:            "engineer",
		TaskDescription: "orphaned task",
	}
	if err := db.CreateTask(orphanTask); err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Mark it as in_progress (simulating it was running but server crashed)
	if err := db.UpdateTaskStatus("orphan-task-1", taskdb.StatusInProgress, ""); err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	// Verify it's in_progress before orphan detection
	beforeTask, err := db.GetTask("orphan-task-1")
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}
	if beforeTask.Status != taskdb.StatusInProgress {
		t.Fatalf("Expected status in_progress, got %q", beforeTask.Status)
	}

	// Run orphan detection (it's not in activeTasks, so it's orphaned)
	server.handleOrphanedTasks()

	// Verify it was marked as failed
	afterTask, err := db.GetTask("orphan-task-1")
	if err != nil {
		t.Fatalf("Failed to get task after orphan detection: %v", err)
	}
	if afterTask.Status != taskdb.StatusFailed {
		t.Errorf("Expected status failed after orphan detection, got %q", afterTask.Status)
	}
	if afterTask.Error == "" {
		t.Error("Expected error message to be set")
	}
	if afterTask.Error != "Task orphaned - server restarted or crashed" {
		t.Errorf("Expected specific error message, got %q", afterTask.Error)
	}
}

// TestOrphanDetection_ActiveTaskNotMarkedOrphaned verifies active tasks are not marked orphaned
func TestOrphanDetection_ActiveTaskNotMarkedOrphaned(t *testing.T) {
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

	// Create a task in taskDB
	activeTask := &taskdb.Task{
		ID:              "active-task-1",
		ProjectRoot:     tmpDir,
		Role:            "engineer",
		TaskDescription: "active task",
	}
	if err := db.CreateTask(activeTask); err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Mark it as in_progress
	if err := db.UpdateTaskStatus("active-task-1", taskdb.StatusInProgress, ""); err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	// Add it to activeTasks (simulating it's actually running)
	execution := &TaskExecution{
		TaskID:      "active-task-1",
		Role:        "engineer",
		Task:        "active task",
		Status:      constants.StatusInProgress,
		StartTime:   time.Now(),
		ProjectRoot: tmpDir,
	}
	server.mu.Lock()
	server.activeTasks["active-task-1"] = execution
	server.mu.Unlock()

	// Run orphan detection
	server.handleOrphanedTasks()

	// Verify it was NOT marked as failed (still in_progress)
	afterTask, err := db.GetTask("active-task-1")
	if err != nil {
		t.Fatalf("Failed to get task after orphan detection: %v", err)
	}
	if afterTask.Status != taskdb.StatusInProgress {
		t.Errorf("Expected active task to remain in_progress, got %q", afterTask.Status)
	}
	if afterTask.Error != "" {
		t.Errorf("Expected no error for active task, got %q", afterTask.Error)
	}
}

// TestOrphanDetection_StaleMetadata verifies stale metadata files are detected
func TestOrphanDetection_StaleMetadata(t *testing.T) {
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

	// Create a stale execution folder with in_progress metadata
	taskID := "stale-execution-1"
	taskDir := filepath.Join(tmpDir, constants.TaskRootDir, "tasks", taskID)
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.Fatalf("Failed to create task dir: %v", err)
	}

	// Write metadata with status=in_progress and old spawned_at
	oldTime := time.Now().Add(-5 * time.Minute)
	metadata := map[string]interface{}{
		"task_id":    taskID,
		"role":       "engineer",
		"task":       "stale task",
		"status":     "in_progress",
		"spawned_at": oldTime.Format(time.RFC3339),
		"updated_at": oldTime.Format(time.RFC3339),
	}
	metadataJSON, _ := json.Marshal(metadata)
	metadataPath := filepath.Join(taskDir, constants.MetadataFileName)
	if err := os.WriteFile(metadataPath, metadataJSON, 0644); err != nil {
		t.Fatalf("Failed to write metadata: %v", err)
	}

	// Run orphan detection
	server.handleOrphanedTasks()

	// Read metadata back and verify it was marked failed
	updatedMetadata, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatalf("Failed to read updated metadata: %v", err)
	}

	var meta map[string]interface{}
	if err := json.Unmarshal(updatedMetadata, &meta); err != nil {
		t.Fatalf("Failed to unmarshal metadata: %v", err)
	}

	if meta["status"] != constants.StatusFailed {
		t.Errorf("Expected status 'failed', got %v", meta["status"])
	}
	if meta["error"] == nil {
		t.Error("Expected error field to be set")
	}
}

// TestOrphanDetection_RecentlySpawnedSkipped verifies recently spawned tasks are skipped
func TestOrphanDetection_RecentlySpawnedSkipped(t *testing.T) {
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

	// Create an execution folder with in_progress metadata spawned very recently
	taskID := "recent-execution-1"
	taskDir := filepath.Join(tmpDir, constants.TaskRootDir, "tasks", taskID)
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.Fatalf("Failed to create task dir: %v", err)
	}

	// Write metadata with status=in_progress and RECENT spawned_at
	recentTime := time.Now().Add(-10 * time.Second) // Less than 30 second threshold
	metadata := map[string]interface{}{
		"task_id":    taskID,
		"role":       "engineer",
		"task":       "recent task",
		"status":     "in_progress",
		"spawned_at": recentTime.Format(time.RFC3339),
		"updated_at": recentTime.Format(time.RFC3339),
	}
	metadataJSON, _ := json.Marshal(metadata)
	metadataPath := filepath.Join(taskDir, constants.MetadataFileName)
	if err := os.WriteFile(metadataPath, metadataJSON, 0644); err != nil {
		t.Fatalf("Failed to write metadata: %v", err)
	}

	// Run orphan detection
	server.handleOrphanedTasks()

	// Read metadata back and verify it was NOT marked failed (still in_progress)
	updatedMetadata, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatalf("Failed to read updated metadata: %v", err)
	}

	var meta map[string]interface{}
	if err := json.Unmarshal(updatedMetadata, &meta); err != nil {
		t.Fatalf("Failed to unmarshal metadata: %v", err)
	}

	if meta["status"] != "in_progress" {
		t.Errorf("Expected recently spawned task to remain in_progress, got %v", meta["status"])
	}
	if meta["error"] != nil {
		t.Errorf("Expected no error for recently spawned task, got %v", meta["error"])
	}
}

// TestOrphanDetection_CheckpointDeletion verifies checkpoints are deleted for orphaned tasks
func TestOrphanDetection_CheckpointDeletion(t *testing.T) {
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

	// Create an orphaned task
	orphanTask := &taskdb.Task{
		ID:              "orphan-with-checkpoint",
		ProjectRoot:     tmpDir,
		Role:            "engineer",
		TaskDescription: "orphaned task with checkpoint",
	}
	if err := db.CreateTask(orphanTask); err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	if err := db.UpdateTaskStatus("orphan-with-checkpoint", taskdb.StatusInProgress, ""); err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	// Create a checkpoint file
	cpPath := checkpointPath(tmpDir, "orphan-with-checkpoint")
	cpDir := filepath.Dir(cpPath)
	if err := os.MkdirAll(cpDir, 0755); err != nil {
		t.Fatalf("Failed to create checkpoint dir: %v", err)
	}
	if err := os.WriteFile(cpPath, []byte("checkpoint data"), 0644); err != nil {
		t.Fatalf("Failed to write checkpoint: %v", err)
	}

	// Verify checkpoint exists
	if _, err := os.Stat(cpPath); err != nil {
		t.Fatalf("Checkpoint should exist before orphan detection: %v", err)
	}

	// Run orphan detection
	server.handleOrphanedTasks()

	// Verify checkpoint was deleted
	if _, err := os.Stat(cpPath); !os.IsNotExist(err) {
		t.Error("Checkpoint should be deleted after orphan detection")
	}
}
