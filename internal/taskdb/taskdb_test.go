package taskdb_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cortexa-llc/ai-pack/internal/taskdb"
)

func openTestDB(t *testing.T) (*taskdb.DB, func()) {
	t.Helper()
	dir := t.TempDir()
	db, err := taskdb.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return db, func() {
		db.Close()
		os.RemoveAll(dir)
	}
}

// TestCreateAndGetTask verifies the basic task lifecycle.
func TestCreateAndGetTask(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()

	task := &taskdb.Task{
		ID:              "ai-pack-aa0",
		ProjectRoot:     "/tmp/project",
		Role:            "engineer",
		TaskDescription: "Do something",
	}
	if err := db.CreateTask(task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	got, err := db.GetTask("ai-pack-aa0")
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if got == nil {
		t.Fatal("GetTask returned nil")
	}
	if got.ID != "ai-pack-aa0" {
		t.Errorf("id = %q, want %q", got.ID, "ai-pack-aa0")
	}
	if got.Status != taskdb.StatusQueued {
		t.Errorf("status = %q, want %q", got.Status, taskdb.StatusQueued)
	}
	if got.LatestRunID != "" {
		t.Errorf("LatestRunID = %q, want empty", got.LatestRunID)
	}
}

// TestCreateTaskRun verifies that a task_run is linked to the parent task
// and that the parent's latest_run_id and status are updated.
func TestCreateTaskRun(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()

	// Create parent
	parent := &taskdb.Task{
		ID:          "ai-pack-aa0",
		ProjectRoot: "/tmp/project",
		Role:        "engineer",
	}
	if err := db.CreateTask(parent); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	run := &taskdb.TaskRun{
		RunID:       "ai-pack-aa0-20260505-170335-b83db8",
		TaskID:      "ai-pack-aa0",
		ProjectRoot: "/tmp/project",
		Role:        "engineer",
	}
	if err := db.CreateTaskRun(run); err != nil {
		t.Fatalf("CreateTaskRun: %v", err)
	}

	// Parent should now reflect the run
	got, err := db.GetTask("ai-pack-aa0")
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if got.LatestRunID != run.RunID {
		t.Errorf("LatestRunID = %q, want %q", got.LatestRunID, run.RunID)
	}

	// Retrieve the run directly
	gotRun, err := db.GetTaskRun(run.RunID)
	if err != nil {
		t.Fatalf("GetTaskRun: %v", err)
	}
	if gotRun == nil {
		t.Fatal("GetTaskRun returned nil")
	}
	if gotRun.TaskID != "ai-pack-aa0" {
		t.Errorf("run.TaskID = %q, want %q", gotRun.TaskID, "ai-pack-aa0")
	}
}

// TestCompleteTaskUpdatesParent verifies that completing a run propagates status
// to the parent task.
func TestCompleteTaskUpdatesParent(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()

	if err := db.CreateTask(&taskdb.Task{
		ID:          "ai-pack-aa0",
		ProjectRoot: "/tmp/project",
		Role:        "engineer",
	}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	if err := db.CreateTaskRun(&taskdb.TaskRun{
		RunID:       "ai-pack-aa0-20260505-170335-b83db8",
		TaskID:      "ai-pack-aa0",
		ProjectRoot: "/tmp/project",
		Role:        "engineer",
	}); err != nil {
		t.Fatalf("CreateTaskRun: %v", err)
	}

	// Complete via the run_id
	if err := db.CompleteTask("ai-pack-aa0-20260505-170335-b83db8", "all done"); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}

	parent, err := db.GetTask("ai-pack-aa0")
	if err != nil || parent == nil {
		t.Fatalf("GetTask: %v", err)
	}
	if parent.Status != taskdb.StatusCompleted {
		t.Errorf("parent status = %q, want completed", parent.Status)
	}
	if parent.Result != "all done" {
		t.Errorf("parent result = %q, want %q", parent.Result, "all done")
	}
}

// TestMultipleRunsLatestStatus verifies that the second run supersedes the first.
func TestMultipleRunsLatestStatus(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()

	if err := db.CreateTask(&taskdb.Task{
		ID:          "ai-pack-aa0",
		ProjectRoot: "/tmp/project",
		Role:        "engineer",
	}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	// First run – fails
	run1ID := "ai-pack-aa0-20260505-170335-b83db8"
	if err := db.CreateTaskRun(&taskdb.TaskRun{
		RunID:       run1ID,
		TaskID:      "ai-pack-aa0",
		ProjectRoot: "/tmp/project",
		Role:        "engineer",
	}); err != nil {
		t.Fatalf("CreateTaskRun 1: %v", err)
	}
	if err := db.FailTask(run1ID, "timeout"); err != nil {
		t.Fatalf("FailTask: %v", err)
	}

	// Second run – succeeds
	run2ID := "ai-pack-aa0-20260506-120000-aabbcc"
	if err := db.CreateTaskRun(&taskdb.TaskRun{
		RunID:       run2ID,
		TaskID:      "ai-pack-aa0",
		ProjectRoot: "/tmp/project",
		Role:        "engineer",
	}); err != nil {
		t.Fatalf("CreateTaskRun 2: %v", err)
	}
	if err := db.CompleteTask(run2ID, "success"); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}

	parent, err := db.GetTask("ai-pack-aa0")
	if err != nil || parent == nil {
		t.Fatalf("GetTask: %v", err)
	}
	if parent.LatestRunID != run2ID {
		t.Errorf("LatestRunID = %q, want %q", parent.LatestRunID, run2ID)
	}
	if parent.Status != taskdb.StatusCompleted {
		t.Errorf("parent status = %q, want completed", parent.Status)
	}

	// Verify first run shows failed, second shows completed
	run1, _ := db.GetTaskRun(run1ID)
	if run1 == nil || run1.Status != taskdb.StatusFailed {
		t.Errorf("run1 status = %q, want failed", run1.Status)
	}
	run2, _ := db.GetTaskRun(run2ID)
	if run2 == nil || run2.Status != taskdb.StatusCompleted {
		t.Errorf("run2 status = %q, want completed", run2.Status)
	}
}

// TestListTasksReturnsLatestStatus verifies ListTasks returns one row per logical task.
func TestListTasksReturnsLatestStatus(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()

	// Create two tasks
	for _, id := range []string{"ai-pack-aa0", "ai-pack-bb1"} {
		if err := db.CreateTask(&taskdb.Task{
			ID:          id,
			ProjectRoot: "/tmp/project",
			Role:        "engineer",
		}); err != nil {
			t.Fatalf("CreateTask %s: %v", id, err)
		}
	}

	// Run and complete task aa0
	if err := db.CreateTaskRun(&taskdb.TaskRun{
		RunID:   "ai-pack-aa0-20260505-170335-b83db8",
		TaskID:  "ai-pack-aa0",
		Role:    "engineer",
	}); err != nil {
		t.Fatalf("CreateTaskRun: %v", err)
	}
	if err := db.CompleteTask("ai-pack-aa0-20260505-170335-b83db8", "done"); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}

	tasks, err := db.ListTasks(taskdb.TaskFilter{})
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("ListTasks returned %d tasks, want 2", len(tasks))
	}

	statusByID := make(map[string]string)
	for _, task := range tasks {
		statusByID[task.ID] = task.Status
	}
	if statusByID["ai-pack-aa0"] != taskdb.StatusCompleted {
		t.Errorf("aa0 status = %q, want completed", statusByID["ai-pack-aa0"])
	}
	if statusByID["ai-pack-bb1"] != taskdb.StatusQueued {
		t.Errorf("bb1 status = %q, want queued", statusByID["ai-pack-bb1"])
	}
}

// TestMigrationFromLegacyTimestampedRows verifies that existing databases
// with timestamped task IDs in the tasks table are migrated correctly.
func TestMigrationFromLegacyTimestampedRows(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// Manually create a DB without the task_runs table to simulate legacy state
	db, err := taskdb.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	// Insert a legacy-style timestamped row directly into tasks
	// This simulates what the old model did: create "ai-pack-aa0-TIMESTAMP" as a task row
	_, err = db.Exec(`
		INSERT INTO tasks (id, project_root, role, task_description, status, created_at, updated_at)
		VALUES ('ai-pack-aa0-20260505-170335-b83db8', '/tmp', 'engineer', 'legacy task', 'completed', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		t.Fatalf("insert legacy row: %v", err)
	}

	db.Close()

	// Re-open: migration should run and convert the timestamped row
	db2, err := taskdb.Open(dbPath)
	if err != nil {
		t.Fatalf("reopen db: %v", err)
	}
	defer db2.Close()

	// Parent task should exist now
	parent, err := db2.GetTask("ai-pack-aa0")
	if err != nil {
		t.Fatalf("GetTask after migration: %v", err)
	}
	if parent == nil {
		t.Fatal("parent task not created by migration")
	}
	if parent.Status != taskdb.StatusCompleted {
		t.Errorf("parent status = %q, want completed", parent.Status)
	}
	if parent.LatestRunID != "ai-pack-aa0-20260505-170335-b83db8" {
		t.Errorf("parent.LatestRunID = %q, want timestamped ID", parent.LatestRunID)
	}

	// Timestamped row should now be in task_runs
	run, err := db2.GetTaskRun("ai-pack-aa0-20260505-170335-b83db8")
	if err != nil {
		t.Fatalf("GetTaskRun: %v", err)
	}
	if run == nil {
		t.Fatal("task_run not created by migration")
	}

	// Timestamped row should NOT be in tasks table anymore
	ghost, err := db2.GetTask("ai-pack-aa0-20260505-170335-b83db8")
	if err != nil {
		t.Fatalf("GetTask ghost check: %v", err)
	}
	if ghost != nil {
		t.Error("legacy timestamped row still present in tasks table after migration")
	}
}
