package taskdb

// Schema defines the SQLite database schema for task tracking.
// This is the source of truth for all task state across projects.
const Schema = `
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    project_root TEXT NOT NULL,
    role TEXT NOT NULL,
    task_description TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('queued', 'in_progress', 'completed', 'failed', 'cancelled')),

    -- Ownership/locking for multi-agent coordination
    owner_agent_id TEXT,
    claimed_at TIMESTAMP,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Results
    result TEXT,
    error TEXT,

    -- Flexible metadata as JSON
    metadata TEXT
);

CREATE INDEX IF NOT EXISTS idx_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_project ON tasks(project_root);
CREATE INDEX IF NOT EXISTS idx_owner ON tasks(owner_agent_id);
CREATE INDEX IF NOT EXISTS idx_created ON tasks(created_at);
CREATE INDEX IF NOT EXISTS idx_role ON tasks(role);

-- Compound index for querying available tasks by role
CREATE INDEX IF NOT EXISTS idx_queued_role ON tasks(status, role, created_at) WHERE status = 'queued';
`

// InitializeSchema creates tables and indexes if they don't exist.
func InitializeSchema(db *DB) error {
	_, err := db.db.Exec(Schema)
	return err
}
