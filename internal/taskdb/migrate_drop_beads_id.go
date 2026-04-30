package taskdb

import (
	"fmt"
)

// MigrateDropBeadsID removes the legacy beads_id column and index from the tasks table.
// This migration is safe to run multiple times (idempotent).
func MigrateDropBeadsID(db *DB) error {
	// Check if beads_id column exists
	var hasBeadsID bool
	err := db.db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('tasks')
		WHERE name = 'beads_id'
	`).Scan(&hasBeadsID)
	if err != nil {
		return fmt.Errorf("failed to check for beads_id column: %w", err)
	}

	if !hasBeadsID {
		// Column already removed, nothing to do
		return nil
	}

	// SQLite doesn't support DROP COLUMN directly before version 3.35.0
	// We need to recreate the table without the beads_id column

	tx, err := db.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create new table without beads_id
	_, err = tx.Exec(`
		CREATE TABLE tasks_new (
			id TEXT PRIMARY KEY,
			project_root TEXT NOT NULL,
			role TEXT NOT NULL,
			task_description TEXT NOT NULL,
			status TEXT NOT NULL CHECK(status IN ('queued', 'in_progress', 'completed', 'failed', 'cancelled')),
			owner_agent_id TEXT,
			claimed_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			started_at TIMESTAMP,
			completed_at TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			result TEXT,
			error TEXT,
			metadata TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create new tasks table: %w", err)
	}

	// Copy data from old table (excluding beads_id)
	_, err = tx.Exec(`
		INSERT INTO tasks_new
		SELECT id, project_root, role, task_description, status,
		       owner_agent_id, claimed_at, created_at, started_at,
		       completed_at, updated_at, result, error, metadata
		FROM tasks
	`)
	if err != nil {
		return fmt.Errorf("failed to copy data to new table: %w", err)
	}

	// Drop old table
	_, err = tx.Exec(`DROP TABLE tasks`)
	if err != nil {
		return fmt.Errorf("failed to drop old table: %w", err)
	}

	// Rename new table
	_, err = tx.Exec(`ALTER TABLE tasks_new RENAME TO tasks`)
	if err != nil {
		return fmt.Errorf("failed to rename new table: %w", err)
	}

	// Recreate indexes (without idx_beads_id)
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_status ON tasks(status)`,
		`CREATE INDEX IF NOT EXISTS idx_project ON tasks(project_root)`,
		`CREATE INDEX IF NOT EXISTS idx_owner ON tasks(owner_agent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_created ON tasks(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_role ON tasks(role)`,
		`CREATE INDEX IF NOT EXISTS idx_queued_role ON tasks(status, role, created_at) WHERE status = 'queued'`,
	}

	for _, indexSQL := range indexes {
		if _, err := tx.Exec(indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// RunMigrations runs all pending database migrations.
func RunMigrations(db *DB) error {
	// Run migration to drop beads_id column
	if err := MigrateDropBeadsID(db); err != nil {
		return fmt.Errorf("beads_id migration failed: %w", err)
	}

	return nil
}
