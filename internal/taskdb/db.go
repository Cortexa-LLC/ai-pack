package taskdb

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the SQLite database connection for task tracking.
type DB struct {
	db *sql.DB
}

// Open opens or creates the SQLite task database at the given path.
// Enables WAL mode for concurrent reads.
func Open(path string) (*DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open with WAL mode and busy timeout
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	// SQLite supports multiple readers but only one writer
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	taskDB := &DB{db: db}

	// Initialize schema
	if err := InitializeSchema(taskDB); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	// Run migrations
	if err := RunMigrations(taskDB); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return taskDB, nil
}

// Exec executes a raw SQL statement (for testing)
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.db.Exec(query, args...)
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.db.Close()
}

// CreateTask inserts a new logical task with status 'queued'.
func (db *DB) CreateTask(task *Task) error {
	query := `
		INSERT INTO tasks (
			id, project_root, role, task_description, status,
			created_at, updated_at, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now
	task.Status = StatusQueued

	_, err := db.db.Exec(query,
		task.ID, task.ProjectRoot, task.Role,
		task.TaskDescription, task.Status,
		task.CreatedAt, task.UpdatedAt, task.Metadata,
	)

	return err
}

// CreateTaskRun inserts a new execution attempt into task_runs and updates
// the parent task's latest_run_id and status to in_progress.
func (db *DB) CreateTaskRun(run *TaskRun) error {
	now := time.Now()
	run.CreatedAt = now
	run.UpdatedAt = now
	run.Status = StatusInProgress

	// Insert the run
	_, err := db.db.Exec(`
		INSERT INTO task_runs (
			run_id, task_id, project_root, role, status,
			created_at, updated_at, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		run.RunID, run.TaskID, run.ProjectRoot, run.Role, run.Status,
		run.CreatedAt, run.UpdatedAt, run.Metadata,
	)
	if err != nil {
		return fmt.Errorf("insert task_run: %w", err)
	}

	// Update parent task: set latest_run_id and status to in_progress
	_, err = db.db.Exec(`
		UPDATE tasks
		SET latest_run_id = ?,
		    status = ?,
		    started_at = COALESCE(started_at, ?),
		    updated_at = ?
		WHERE id = ?
	`, run.RunID, StatusInProgress, now, now, run.TaskID)
	if err != nil {
		return fmt.Errorf("update parent task: %w", err)
	}

	return nil
}

// GetTaskRun retrieves a single task_run by its run_id.
func (db *DB) GetTaskRun(runID string) (*TaskRun, error) {
	query := `
		SELECT run_id, task_id, project_root, role, status,
		       owner_agent_id, claimed_at, created_at, started_at, completed_at,
		       updated_at, result, error, metadata
		FROM task_runs WHERE run_id = ?
	`

	run := &TaskRun{}
	var ownerAgentID, result, errorMsg, metadata sql.NullString
	var claimedAt, startedAt, completedAt sql.NullTime

	err := db.db.QueryRow(query, runID).Scan(
		&run.RunID, &run.TaskID, &run.ProjectRoot, &run.Role, &run.Status,
		&ownerAgentID, &claimedAt, &run.CreatedAt, &startedAt, &completedAt,
		&run.UpdatedAt, &result, &errorMsg, &metadata,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if ownerAgentID.Valid {
		run.OwnerAgentID = ownerAgentID.String
	}
	if claimedAt.Valid {
		run.ClaimedAt = &claimedAt.Time
	}
	if startedAt.Valid {
		run.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		run.CompletedAt = &completedAt.Time
	}
	if result.Valid {
		run.Result = result.String
	}
	if errorMsg.Valid {
		run.Error = errorMsg.String
	}
	if metadata.Valid {
		run.Metadata = metadata.String
	}

	return run, nil
}

// isRunID returns true if the given ID looks like a timestamped run ID
// (e.g. "ai-pack-aa0-20260505-170335-b83db8") rather than a short task ID.
func isRunID(id string) bool {
	// Run IDs have the form <prefix>-<YYYYMMDD>-<HHMMSS>-<hex>
	// We detect them by checking if a task_run with this ID exists.
	return strings.Count(id, "-") >= 5
}

// resolveToTaskID resolves either a run_id or task_id to a task_id.
// If runID is provided and exists in task_runs, it returns the parent task_id.
// Otherwise returns id unchanged (treating it as a task ID).
func (db *DB) resolveToTaskID(id string) (taskID string, runID string, err error) {
	if isRunID(id) {
		// Try to find this as a run_id first
		var tID sql.NullString
		err = db.db.QueryRow(`SELECT task_id FROM task_runs WHERE run_id = ?`, id).Scan(&tID)
		if err == nil && tID.Valid {
			return tID.String, id, nil
		}
		if err != nil && err != sql.ErrNoRows {
			return "", "", err
		}
	}
	// Treat id as a task_id
	return id, "", nil
}

// GetTask retrieves a task by ID (task_id). Returns nil if not found.
func (db *DB) GetTask(id string) (*Task, error) {
	query := `
		SELECT id, project_root, role, task_description, status,
		       owner_agent_id, claimed_at, created_at, started_at, completed_at,
		       updated_at, result, error, metadata, latest_run_id
		FROM tasks WHERE id = ?
	`

	task := &Task{}
	var ownerAgentID, result, errorMsg, metadata, latestRunID sql.NullString
	var claimedAt, startedAt, completedAt sql.NullTime

	err := db.db.QueryRow(query, id).Scan(
		&task.ID, &task.ProjectRoot, &task.Role,
		&task.TaskDescription, &task.Status,
		&ownerAgentID, &claimedAt, &task.CreatedAt, &startedAt, &completedAt,
		&task.UpdatedAt, &result, &errorMsg, &metadata, &latestRunID,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if ownerAgentID.Valid {
		task.OwnerAgentID = ownerAgentID.String
	}
	if claimedAt.Valid {
		task.ClaimedAt = &claimedAt.Time
	}
	if startedAt.Valid {
		task.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}
	if result.Valid {
		task.Result = result.String
	}
	if errorMsg.Valid {
		task.Error = errorMsg.String
	}
	if metadata.Valid {
		task.Metadata = metadata.String
	}
	if latestRunID.Valid {
		task.LatestRunID = latestRunID.String
	}

	return task, nil
}

// UpdateTaskStatus updates the status of a task (or run if given a run_id).
func (db *DB) UpdateTaskStatus(id, status, errorMsg string) error {
	taskID, runID, err := db.resolveToTaskID(id)
	if err != nil {
		return err
	}

	if runID != "" {
		// Update run status
		_, err = db.db.Exec(`
			UPDATE task_runs SET status = ?, error = ?, updated_at = CURRENT_TIMESTAMP
			WHERE run_id = ?
		`, status, errorMsg, runID)
		if err != nil {
			return err
		}
		id = taskID
	}

	_, err = db.db.Exec(`
		UPDATE tasks SET status = ?, error = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, status, errorMsg, id)
	return err
}

// UpdateTaskResult updates the result of a completed task.
func (db *DB) UpdateTaskResult(id, result string) error {
	taskID, runID, err := db.resolveToTaskID(id)
	if err != nil {
		return err
	}

	if runID != "" {
		_, err = db.db.Exec(`
			UPDATE task_runs SET result = ?, updated_at = CURRENT_TIMESTAMP WHERE run_id = ?
		`, result, runID)
		if err != nil {
			return err
		}
		id = taskID
	}

	_, err = db.db.Exec(`
		UPDATE tasks SET result = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, result, id)
	return err
}

// CompleteTask marks a task (or run) as completed.
// If id is a run_id, the run is marked completed and the status propagates to the parent task.
func (db *DB) CompleteTask(id, result string) error {
	taskID, runID, err := db.resolveToTaskID(id)
	if err != nil {
		return err
	}

	now := time.Now()

	if runID != "" {
		// Mark the run as completed
		_, err = db.db.Exec(`
			UPDATE task_runs
			SET status = ?, result = ?, completed_at = ?, updated_at = ?
			WHERE run_id = ?
		`, StatusCompleted, result, now, now, runID)
		if err != nil {
			return fmt.Errorf("complete run: %w", err)
		}
		// Propagate to parent task
		_, err = db.db.Exec(`
			UPDATE tasks
			SET status = ?, result = ?, completed_at = ?, updated_at = ?
			WHERE id = ?
		`, StatusCompleted, result, now, now, taskID)
		return err
	}

	// Direct task completion
	_, err = db.db.Exec(`
		UPDATE tasks
		SET status = ?, result = ?, completed_at = ?, updated_at = ?
		WHERE id = ?
	`, StatusCompleted, result, now, now, id)
	return err
}

// FailTask marks a task (or run) as failed.
// If id is a run_id, the run is marked failed and the status propagates to the parent task.
func (db *DB) FailTask(id, errorMsg string) error {
	taskID, runID, err := db.resolveToTaskID(id)
	if err != nil {
		return err
	}

	now := time.Now()

	if runID != "" {
		// Mark the run as failed
		_, err = db.db.Exec(`
			UPDATE task_runs
			SET status = ?, error = ?, completed_at = ?, updated_at = ?
			WHERE run_id = ?
		`, StatusFailed, errorMsg, now, now, runID)
		if err != nil {
			return fmt.Errorf("fail run: %w", err)
		}
		// Propagate to parent task
		_, err = db.db.Exec(`
			UPDATE tasks
			SET status = ?, error = ?, completed_at = ?, updated_at = ?
			WHERE id = ?
		`, StatusFailed, errorMsg, now, now, taskID)
		return err
	}

	// Direct task failure
	_, err = db.db.Exec(`
		UPDATE tasks
		SET status = ?, error = ?, completed_at = ?, updated_at = ?
		WHERE id = ?
	`, StatusFailed, errorMsg, now, now, id)
	return err
}

// CancelTask marks a task as cancelled.
func (db *DB) CancelTask(id string) error {
	query := `
		UPDATE tasks
		SET status = ?, completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := db.db.Exec(query, StatusCancelled, id)
	return err
}

// ListTasks returns all tasks matching the given filter.
func (db *DB) ListTasks(filter TaskFilter) ([]*Task, error) {
	query := "SELECT id, project_root, role, task_description, status, " +
		"owner_agent_id, claimed_at, created_at, started_at, completed_at, " +
		"updated_at, result, error, metadata, latest_run_id FROM tasks WHERE 1=1"

	args := []interface{}{}

	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}
	if filter.ProjectRoot != "" {
		query += " AND project_root = ?"
		args = append(args, filter.ProjectRoot)
	}
	if filter.Role != "" {
		query += " AND role = ?"
		args = append(args, filter.Role)
	}
	if filter.OwnerAgentID != "" {
		query += " AND owner_agent_id = ?"
		args = append(args, filter.OwnerAgentID)
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	rows, err := db.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		var ownerAgentID, result, errorMsg, metadata, latestRunID sql.NullString
		var claimedAt, startedAt, completedAt sql.NullTime

		err := rows.Scan(
			&task.ID, &task.ProjectRoot, &task.Role,
			&task.TaskDescription, &task.Status,
			&ownerAgentID, &claimedAt, &task.CreatedAt, &startedAt, &completedAt,
			&task.UpdatedAt, &result, &errorMsg, &metadata, &latestRunID,
		)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if ownerAgentID.Valid {
			task.OwnerAgentID = ownerAgentID.String
		}
		if claimedAt.Valid {
			task.ClaimedAt = &claimedAt.Time
		}
		if startedAt.Valid {
			task.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}
		if result.Valid {
			task.Result = result.String
		}
		if errorMsg.Valid {
			task.Error = errorMsg.String
		}
		if metadata.Valid {
			task.Metadata = metadata.String
		}
		if latestRunID.Valid {
			task.LatestRunID = latestRunID.String
		}

		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

// DeleteTask deletes a task from the database.
func (db *DB) DeleteTask(id string) error {
	query := `DELETE FROM tasks WHERE id = ?`
	result, err := db.db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("task not found: %s", id)
	}

	return nil
}

// TaskFilter defines filtering criteria for listing tasks.
type TaskFilter struct {
	Status       string
	ProjectRoot  string
	Role         string
	OwnerAgentID string
	Limit        int
}
