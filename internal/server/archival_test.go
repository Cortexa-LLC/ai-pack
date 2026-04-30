package server

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/config"
	"github.com/cortexa-llc/ai-pack/internal/constants"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/internal/taskdb"
)

// TestArchiveOldTasks_CompletedTasks verifies old completed tasks are archived
func TestArchiveOldTasks_CompletedTasks(t *testing.T) {
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
		TaskCleanup: config.TaskCleanupConfig{
			Enabled:          true,
			ArchiveAfterDays: 7,
		},
	}

	server, err := NewAgentServer(tmpDir, 1, 4096, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	server.taskDB = db

	// Create an old completed task (15 days ago)
	oldTask := &taskdb.Task{
		ID:              "old-task-1",
		ProjectRoot:     tmpDir,
		Role:            "engineer",
		TaskDescription: "old completed task",
	}
	if err := db.CreateTask(oldTask); err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Complete it
	if err := db.CompleteTask("old-task-1", "done"); err != nil {
		t.Fatalf("Failed to complete task: %v", err)
	}

	// Manually set CompletedAt to 15 days ago (can't use SQL UPDATE in test without direct DB access)
	// Instead, create the execution folder and run archival
	taskDir := filepath.Join(tmpDir, constants.TaskRootDir, "tasks", "old-task-1")
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.Fatalf("Failed to create task dir: %v", err)
	}

	// Create a dummy file to verify archival moved the directory
	dummyFile := filepath.Join(taskDir, "execution.log")
	if err := os.WriteFile(dummyFile, []byte("test log"), 0644); err != nil {
		t.Fatalf("Failed to write dummy file: %v", err)
	}

	// Update task's CompletedAt in database to be old
	// We need to do this via raw SQL since the API doesn't expose it
	oldTime := time.Now().AddDate(0, 0, -15)
	_, execErr := db.Exec("UPDATE tasks SET completed_at = ?, updated_at = ? WHERE id = ?",
		oldTime, oldTime, "old-task-1")
	if execErr != nil {
		t.Fatalf("Failed to update task timestamps: %v", execErr)
	}

	// Run archival
	server.archiveOldTasks()

	// Verify task directory was moved to archive
	archiveMonth := oldTime.Format("2006-01")
	archivedDir := filepath.Join(tmpDir, constants.TaskRootDir, "archive", archiveMonth, "old-task-1")

	if _, err := os.Stat(archivedDir); os.IsNotExist(err) {
		t.Errorf("Task directory should be archived at %s", archivedDir)
	}

	// Verify original directory no longer exists
	if _, err := os.Stat(taskDir); !os.IsNotExist(err) {
		t.Error("Original task directory should be deleted after archival")
	}

	// Verify archived file exists
	archivedFile := filepath.Join(archivedDir, "execution.log")
	if _, err := os.Stat(archivedFile); os.IsNotExist(err) {
		t.Error("Execution log should exist in archive")
	}
}

// TestArchiveOldTasks_RecentTasksNotArchived verifies recent tasks are not archived
func TestArchiveOldTasks_RecentTasksNotArchived(t *testing.T) {
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
		TaskCleanup: config.TaskCleanupConfig{
			Enabled:          true,
			ArchiveAfterDays: 7,
		},
	}

	server, err := NewAgentServer(tmpDir, 1, 4096, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	server.taskDB = db

	// Create a recent completed task (2 days ago)
	recentTask := &taskdb.Task{
		ID:              "recent-task-1",
		ProjectRoot:     tmpDir,
		Role:            "engineer",
		TaskDescription: "recent completed task",
	}
	if err := db.CreateTask(recentTask); err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	if err := db.CompleteTask("recent-task-1", "done"); err != nil {
		t.Fatalf("Failed to complete task: %v", err)
	}

	// Create execution folder
	taskDir := filepath.Join(tmpDir, constants.TaskRootDir, "tasks", "recent-task-1")
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.Fatalf("Failed to create task dir: %v", err)
	}

	dummyFile := filepath.Join(taskDir, "execution.log")
	if err := os.WriteFile(dummyFile, []byte("test log"), 0644); err != nil {
		t.Fatalf("Failed to write dummy file: %v", err)
	}

	// Update CompletedAt to 2 days ago (within 7 day threshold)
	recentTime := time.Now().AddDate(0, 0, -2)
	_, execErr := db.Exec("UPDATE tasks SET completed_at = ?, updated_at = ? WHERE id = ?",
		recentTime, recentTime, "recent-task-1")
	if execErr != nil {
		t.Fatalf("Failed to update task timestamps: %v", execErr)
	}

	// Run archival
	server.archiveOldTasks()

	// Verify task directory was NOT moved (still in tasks/)
	if _, err := os.Stat(taskDir); os.IsNotExist(err) {
		t.Error("Recent task directory should NOT be archived")
	}

	// Verify it's not in archive
	archiveMonth := recentTime.Format("2006-01")
	archivedDir := filepath.Join(tmpDir, constants.TaskRootDir, "archive", archiveMonth, "recent-task-1")
	if _, err := os.Stat(archivedDir); !os.IsNotExist(err) {
		t.Error("Recent task should NOT be in archive")
	}
}

// TestArchiveOldTasks_FailedTasks verifies old failed tasks are also archived
func TestArchiveOldTasks_FailedTasks(t *testing.T) {
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
		TaskCleanup: config.TaskCleanupConfig{
			Enabled:          true,
			ArchiveAfterDays: 7,
		},
	}

	server, err := NewAgentServer(tmpDir, 1, 4096, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	server.taskDB = db

	// Create an old failed task
	failedTask := &taskdb.Task{
		ID:              "failed-task-1",
		ProjectRoot:     tmpDir,
		Role:            "engineer",
		TaskDescription: "old failed task",
	}
	if err := db.CreateTask(failedTask); err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	if err := db.FailTask("failed-task-1", "task failed"); err != nil {
		t.Fatalf("Failed to fail task: %v", err)
	}

	// Create execution folder
	taskDir := filepath.Join(tmpDir, constants.TaskRootDir, "tasks", "failed-task-1")
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.Fatalf("Failed to create task dir: %v", err)
	}

	// Update CompletedAt to be old
	oldTime := time.Now().AddDate(0, 0, -15)
	_, execErr := db.Exec("UPDATE tasks SET completed_at = ?, updated_at = ? WHERE id = ?",
		oldTime, oldTime, "failed-task-1")
	if execErr != nil {
		t.Fatalf("Failed to update task timestamps: %v", execErr)
	}

	// Run archival
	server.archiveOldTasks()

	// Verify failed task was archived
	archiveMonth := oldTime.Format("2006-01")
	archivedDir := filepath.Join(tmpDir, constants.TaskRootDir, "archive", archiveMonth, "failed-task-1")

	if _, err := os.Stat(archivedDir); os.IsNotExist(err) {
		t.Error("Failed task should be archived")
	}
}

// TestArchiveOldTasks_DisabledNoArchival verifies archival doesn't run when disabled
func TestArchiveOldTasks_DisabledNoArchival(t *testing.T) {
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
		TaskCleanup: config.TaskCleanupConfig{
			Enabled:          false, // Disabled
			ArchiveAfterDays: 7,
		},
	}

	server, err := NewAgentServer(tmpDir, 1, 4096, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	server.taskDB = db

	// Create an old task
	oldTask := &taskdb.Task{
		ID:              "old-disabled-1",
		ProjectRoot:     tmpDir,
		Role:            "engineer",
		TaskDescription: "old task with archival disabled",
	}
	if err := db.CreateTask(oldTask); err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	if err := db.CompleteTask("old-disabled-1", "done"); err != nil {
		t.Fatalf("Failed to complete task: %v", err)
	}

	taskDir := filepath.Join(tmpDir, constants.TaskRootDir, "tasks", "old-disabled-1")
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.Fatalf("Failed to create task dir: %v", err)
	}

	oldTime := time.Now().AddDate(0, 0, -15)
	_, execErr := db.Exec("UPDATE tasks SET completed_at = ?, updated_at = ? WHERE id = ?",
		oldTime, oldTime, "old-disabled-1")
	if execErr != nil {
		t.Fatalf("Failed to update task timestamps: %v", execErr)
	}

	// Run archival (should be no-op)
	server.archiveOldTasks()

	// Verify task was NOT archived
	if _, err := os.Stat(taskDir); os.IsNotExist(err) {
		t.Error("Task should NOT be archived when archival is disabled")
	}
}

// TestArchiveOldTasks_NoExecutionFolderSkipped verifies tasks without execution folders are skipped gracefully
func TestArchiveOldTasks_NoExecutionFolderSkipped(t *testing.T) {
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
		TaskCleanup: config.TaskCleanupConfig{
			Enabled:          true,
			ArchiveAfterDays: 7,
		},
	}

	server, err := NewAgentServer(tmpDir, 1, 4096, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	server.taskDB = db

	// Create an old completed task WITHOUT execution folder
	oldTask := &taskdb.Task{
		ID:              "no-folder-1",
		ProjectRoot:     tmpDir,
		Role:            "engineer",
		TaskDescription: "task without execution folder",
	}
	if err := db.CreateTask(oldTask); err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	if err := db.CompleteTask("no-folder-1", "done"); err != nil {
		t.Fatalf("Failed to complete task: %v", err)
	}

	oldTime := time.Now().AddDate(0, 0, -15)
	_, execErr := db.Exec("UPDATE tasks SET completed_at = ?, updated_at = ? WHERE id = ?",
		oldTime, oldTime, "no-folder-1")
	if execErr != nil {
		t.Fatalf("Failed to update task timestamps: %v", execErr)
	}

	// Run archival (should not error, should skip gracefully)
	server.archiveOldTasks()

	// Verify archive directory was not created
	archiveMonth := oldTime.Format("2006-01")
	archivedDir := filepath.Join(tmpDir, constants.TaskRootDir, "archive", archiveMonth, "no-folder-1")
	if _, err := os.Stat(archivedDir); !os.IsNotExist(err) {
		t.Error("Should not create archive for non-existent execution folder")
	}
}
