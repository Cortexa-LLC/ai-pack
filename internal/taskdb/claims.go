package taskdb

import (
	"database/sql"
	"time"
)

// ClaimNextTask atomically finds and claims the next queued task for the given role.
// Returns nil if no tasks are available.
// This prevents race conditions when multiple agents try to claim the same task.
func (db *DB) ClaimNextTask(role, agentID string) (*Task, error) {
	tx, err := db.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Find next queued task for this role
	query := `
		SELECT id FROM tasks
		WHERE status = ? AND role = ?
		ORDER BY created_at ASC
		LIMIT 1
	`

	var taskID string
	err = tx.QueryRow(query, StatusQueued, role).Scan(&taskID)
	if err == sql.ErrNoRows {
		return nil, nil // No tasks available
	}
	if err != nil {
		return nil, err
	}

	// Atomically claim it (double-check it's still queued)
	update := `
		UPDATE tasks
		SET status = ?,
		    owner_agent_id = ?,
		    claimed_at = ?,
		    started_at = ?,
		    updated_at = ?
		WHERE id = ? AND status = ?
	`

	now := time.Now()
	result, err := tx.Exec(update,
		StatusInProgress, agentID, now, now, now,
		taskID, StatusQueued,
	)
	if err != nil {
		return nil, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rows == 0 {
		// Someone else claimed it between SELECT and UPDATE
		return nil, nil
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Return the claimed task
	return db.GetTask(taskID)
}

// ReleaseTask releases a claimed task back to the queue.
// This is useful if an agent crashes or decides not to execute the task.
func (db *DB) ReleaseTask(id string) error {
	query := `
		UPDATE tasks
		SET status = ?,
		    owner_agent_id = NULL,
		    claimed_at = NULL,
		    started_at = NULL,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = ?
	`

	_, err := db.db.Exec(query, StatusQueued, id, StatusInProgress)
	return err
}

// CleanupStaleTasks resets tasks that have been claimed for longer than the timeout.
// This handles cases where an agent crashes without releasing its tasks.
// Returns the number of tasks reset.
func (db *DB) CleanupStaleTasks(timeout time.Duration) (int, error) {
	cutoff := time.Now().Add(-timeout)

	query := `
		UPDATE tasks
		SET status = ?,
		    owner_agent_id = NULL,
		    claimed_at = NULL,
		    started_at = NULL,
		    updated_at = CURRENT_TIMESTAMP
		WHERE status = ? AND claimed_at < ?
	`

	result, err := db.db.Exec(query, StatusQueued, StatusInProgress, cutoff)
	if err != nil {
		return 0, err
	}

	rows, err := result.RowsAffected()
	return int(rows), err
}

// IsTaskClaimed checks if a task is currently claimed by an agent.
func (db *DB) IsTaskClaimed(id string) (bool, string, error) {
	query := `
		SELECT owner_agent_id FROM tasks
		WHERE id = ? AND status = ?
	`

	var ownerAgentID sql.NullString
	err := db.db.QueryRow(query, id, StatusInProgress).Scan(&ownerAgentID)

	if err == sql.ErrNoRows {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}

	if ownerAgentID.Valid {
		return true, ownerAgentID.String, nil
	}

	return false, "", nil
}
