package taskdb

// Schema defines the SQLite database schema for task tracking.
// This is the source of truth for all task state across projects.
const Schema = `
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    project_root TEXT NOT NULL,
    role TEXT NOT NULL,
    task_description TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('queued', 'in_progress', 'paused', 'completed', 'failed', 'cancelled', 'blocked')),

    -- Ownership/locking for multi-agent coordination
    owner_agent_id TEXT,
    claimed_at TIMESTAMP,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Execution output
    result TEXT,
    error TEXT,
    metadata TEXT,

    -- Points to the most recent task_runs row (NULL if never executed)
    latest_run_id TEXT
);

-- One row per execution attempt of a logical task.
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
);

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project_root);
CREATE INDEX IF NOT EXISTS idx_tasks_updated ON tasks(updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_runs_task_id ON task_runs(task_id);
CREATE INDEX IF NOT EXISTS idx_runs_status ON task_runs(status);
CREATE INDEX IF NOT EXISTS idx_runs_created ON task_runs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_runs_task_created ON task_runs(task_id, created_at DESC);
`

// InitializeSchema applies the schema to the database.
// It is idempotent (uses CREATE TABLE IF NOT EXISTS).
func InitializeSchema(db *DB) error {
	_, err := db.db.Exec(Schema)
	return err
}
