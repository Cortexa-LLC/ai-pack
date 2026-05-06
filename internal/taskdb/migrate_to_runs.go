package taskdb

import (
	"database/sql"
	"fmt"
	"strings"
)

// MigrateToTaskRunsModel migrates the database from the flat tasks model to the
// tasks + task_runs model.
//
// The old model stored every execution attempt as a separate row in tasks with
// a timestamped ID like "ai-pack-aa0-20260505-170335-b83db8". This caused
// orphaned 'queued' entries and confused short-ID lookups.
//
// The new model:
//   - tasks table: one row per logical task (short ID "ai-pack-aa0"), with
//     latest_run_id pointing to the most recent execution
//   - task_runs table: one row per execution attempt (timestamped ID)
//
// This migration:
//  1. Adds latest_run_id column to tasks table (if missing)
//  2. Ensures task_runs table exists
//  3. Identifies old timestamped rows in tasks and moves them to task_runs
//  4. Creates/updates the parent logical-task row with the latest status
//
// It is safe to run multiple times (idempotent).
func MigrateToTaskRunsModel(db *DB) error {
	// Step 1: Add latest_run_id column to tasks if missing
	if err := addLatestRunIDColumn(db); err != nil {
		return fmt.Errorf("add latest_run_id column: %w", err)
	}

	// Step 2: Ensure task_runs table exists (schema may not have been applied yet
	// if the DB was created before this migration was added)
	if err := ensureTaskRunsTable(db); err != nil {
		return fmt.Errorf("ensure task_runs table: %w", err)
	}

	// Step 3: Move legacy timestamped rows into task_runs and consolidate parents
	if err := migrateTimestampedRows(db); err != nil {
		return fmt.Errorf("migrate timestamped rows: %w", err)
	}

	return nil
}

// addLatestRunIDColumn adds the latest_run_id column to the tasks table if it doesn't exist.
func addLatestRunIDColumn(db *DB) error {
	var hasColumn bool
	err := db.db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('tasks')
		WHERE name = 'latest_run_id'
	`).Scan(&hasColumn)
	if err != nil {
		return err
	}
	if hasColumn {
		return nil
	}

	// Add the column with no default (NULL means no run yet)
	_, err = db.db.Exec(`ALTER TABLE tasks ADD COLUMN latest_run_id TEXT`)
	return err
}

// ensureTaskRunsTable creates the task_runs table if it does not exist.
func ensureTaskRunsTable(db *DB) error {
	_, err := db.db.Exec(`
		CREATE TABLE IF NOT EXISTS task_runs (
			run_id TEXT PRIMARY KEY,
			task_id TEXT NOT NULL,
			project_root TEXT NOT NULL,
			role TEXT NOT NULL,
			status TEXT NOT NULL CHECK(status IN ('queued', 'in_progress', 'paused', 'completed', 'failed', 'cancelled', 'blocked')),
			owner_agent_id TEXT,
			claimed_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			started_at TIMESTAMP,
			completed_at TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			result TEXT,
			error TEXT,
			metadata TEXT,
			FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Create indexes
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_runs_task_id ON task_runs(task_id)`,
		`CREATE INDEX IF NOT EXISTS idx_runs_status ON task_runs(status)`,
		`CREATE INDEX IF NOT EXISTS idx_runs_created ON task_runs(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_runs_task_created ON task_runs(task_id, created_at DESC)`,
	}
	for _, idx := range indexes {
		if _, err := db.db.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}

// migrateTimestampedRows identifies legacy timestamped task rows in the tasks
// table and migrates them into the task_runs model.
//
// A legacy timestamped row has an ID like "ai-pack-aa0-20260505-170335-b83db8".
// The short/parent ID is "ai-pack-aa0".
//
// For each such row we:
//  1. Ensure a parent tasks row exists for the short ID
//  2. Insert the timestamped row into task_runs (if not already present)
//  3. Update the parent task's latest_run_id to the most recent run
//  4. Delete the timestamped row from tasks
func migrateTimestampedRows(db *DB) error {
	// Find all rows in tasks whose ID looks like a timestamped run
	// Pattern: ends with "-YYYYMMDD-HHMMSS-xxxxxx" (6 extra segments)
	rows, err := db.db.Query(`
		SELECT id, project_root, role, task_description, status,
		       owner_agent_id, claimed_at, created_at, started_at, completed_at,
		       updated_at, result, error, metadata
		FROM tasks
		WHERE id GLOB '*-2[0-9][0-9][0-9][0-9][0-9][0-9][0-9]-[0-9][0-9][0-9][0-9][0-9][0-9]-*'
		ORDER BY created_at ASC
	`)
	if err != nil {
		return err
	}

	type legacyRow struct {
		id, projectRoot, role, taskDesc, status string
		ownerAgentID, result, errorMsg, metadata sql.NullString
		claimedAt, startedAt, completedAt        sql.NullTime
		createdAt, updatedAt                     string
	}

	var legacy []legacyRow
	for rows.Next() {
		var r legacyRow
		if err := rows.Scan(
			&r.id, &r.projectRoot, &r.role, &r.taskDesc, &r.status,
			&r.ownerAgentID, &r.claimedAt, &r.createdAt, &r.startedAt, &r.completedAt,
			&r.updatedAt, &r.result, &r.errorMsg, &r.metadata,
		); err != nil {
			rows.Close()
			return err
		}
		legacy = append(legacy, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	if len(legacy) == 0 {
		return nil // nothing to migrate
	}

	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, r := range legacy {
		shortID := extractShortIDFromTimestamped(r.id)
		if shortID == "" {
			continue // skip if we can't determine the parent
		}

		// Ensure parent task row exists
		var exists bool
		if err := tx.QueryRow(`SELECT 1 FROM tasks WHERE id = ?`, shortID).Scan(&exists); err == sql.ErrNoRows {
			// Create the parent task row from the data in this run
			_, err = tx.Exec(`
				INSERT INTO tasks (
					id, project_root, role, task_description, status,
					created_at, updated_at, metadata
				) VALUES (?, ?, ?, ?, 'queued', ?, ?, ?)`,
				shortID, r.projectRoot, r.role, r.taskDesc,
				r.createdAt, r.updatedAt, r.metadata,
			)
			if err != nil {
				return fmt.Errorf("create parent task %s: %w", shortID, err)
			}
		} else if err != nil {
			return err
		}

		// Insert into task_runs (ignore if already present from a previous migration)
		_, err = tx.Exec(`
			INSERT OR IGNORE INTO task_runs (
				run_id, task_id, project_root, role, status,
				owner_agent_id, claimed_at, created_at, started_at, completed_at,
				updated_at, result, error, metadata
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			r.id, shortID, r.projectRoot, r.role, r.status,
			r.ownerAgentID, r.claimedAt, r.createdAt, r.startedAt, r.completedAt,
			r.updatedAt, r.result, r.errorMsg, r.metadata,
		)
		if err != nil {
			return fmt.Errorf("insert task_run %s: %w", r.id, err)
		}

		// Delete the timestamped row from tasks
		if _, err := tx.Exec(`DELETE FROM tasks WHERE id = ?`, r.id); err != nil {
			return fmt.Errorf("delete legacy row %s: %w", r.id, err)
		}
	}

	// Update each parent task's latest_run_id and status to reflect the most recent run
	if _, err := tx.Exec(`
		UPDATE tasks
		SET latest_run_id = (
			SELECT run_id FROM task_runs
			WHERE task_id = tasks.id
			ORDER BY created_at DESC
			LIMIT 1
		),
		status = (
			SELECT status FROM task_runs
			WHERE task_id = tasks.id
			ORDER BY created_at DESC
			LIMIT 1
		)
		WHERE EXISTS (
			SELECT 1 FROM task_runs WHERE task_id = tasks.id
		)
	`); err != nil {
		return fmt.Errorf("update parent task statuses: %w", err)
	}

	return tx.Commit()
}

// extractShortIDFromTimestamped extracts the base short ID from a timestamped run ID.
// e.g. "ai-pack-aa0-20260505-170335-b83db8" → "ai-pack-aa0"
// Returns empty string if the ID doesn't look like a timestamped run.
func extractShortIDFromTimestamped(id string) string {
	// Timestamped IDs end with "-YYYYMMDD-HHMMSS-xxxxxx"
	// Split from right: last 3 dash-separated segments are timestamp components
	parts := strings.Split(id, "-")
	if len(parts) < 4 {
		return ""
	}
	// Check that parts[-3] looks like a date (8 digits) and parts[-2] looks like time (6 digits)
	dateIdx := len(parts) - 3
	timeIdx := len(parts) - 2
	if len(parts[dateIdx]) == 8 && isAllDigits(parts[dateIdx]) &&
		len(parts[timeIdx]) == 6 && isAllDigits(parts[timeIdx]) {
		return strings.Join(parts[:dateIdx], "-")
	}
	return ""
}

func isAllDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}
