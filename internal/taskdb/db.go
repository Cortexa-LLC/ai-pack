package taskdb

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
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

// CreateTask inserts a new task with status 'queued'.
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

// GetTask retrieves a task by ID.
func (db *DB) GetTask(id string) (*Task, error) {
	query := `
		SELECT id, project_root, role, task_description, status,
		       owner_agent_id, claimed_at, created_at, started_at, completed_at,
		       updated_at, result, error, metadata
		FROM tasks WHERE id = ?
	`

	task := &Task{}
	var ownerAgentID, result, errorMsg, metadata sql.NullString
	var claimedAt, startedAt, completedAt sql.NullTime

	err := db.db.QueryRow(query, id).Scan(
		&task.ID, &task.ProjectRoot, &task.Role,
		&task.TaskDescription, &task.Status,
		&ownerAgentID, &claimedAt, &task.CreatedAt, &startedAt, &completedAt,
		&task.UpdatedAt, &result, &errorMsg, &metadata,
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

	return task, nil
}


// UpdateTaskStatus updates the status of a task.
func (db *DB) UpdateTaskStatus(id, status, errorMsg string) error {
	query := `
		UPDATE tasks
		SET status = ?, error = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := db.db.Exec(query, status, errorMsg, id)
	return err
}

// UpdateTaskResult updates the result of a completed task.
func (db *DB) UpdateTaskResult(id, result string) error {
	query := `
		UPDATE tasks
		SET result = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := db.db.Exec(query, result, id)
	return err
}

// CompleteTask marks a task as completed.
func (db *DB) CompleteTask(id, result string) error {
	query := `
		UPDATE tasks
		SET status = ?, result = ?, completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := db.db.Exec(query, StatusCompleted, result, id)
	return err
}

// FailTask marks a task as failed.
func (db *DB) FailTask(id, errorMsg string) error {
	query := `
		UPDATE tasks
		SET status = ?, error = ?, completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := db.db.Exec(query, StatusFailed, errorMsg, id)
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
		"updated_at, result, error, metadata FROM tasks WHERE 1=1"

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
		var ownerAgentID, result, errorMsg, metadata sql.NullString
		var claimedAt, startedAt, completedAt sql.NullTime

		err := rows.Scan(
			&task.ID, &task.ProjectRoot, &task.Role,
			&task.TaskDescription, &task.Status,
			&ownerAgentID, &claimedAt, &task.CreatedAt, &startedAt, &completedAt,
			&task.UpdatedAt, &result, &errorMsg, &metadata,
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
