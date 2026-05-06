package taskdb

import "time"

// Task represents a logical task in the database.
// A task may have many execution attempts (TaskRuns). The task's status
// reflects the most recent run's outcome via latest_run_id.
type Task struct {
	ID              string     `json:"id"`
	ProjectRoot     string     `json:"project_root"`
	Role            string     `json:"role"`
	TaskDescription string     `json:"task_description"`
	Status          string     `json:"status"`
	OwnerAgentID    string     `json:"owner_agent_id,omitempty"`
	ClaimedAt       *time.Time `json:"claimed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	UpdatedAt       time.Time  `json:"updated_at"`
	Result          string     `json:"result,omitempty"`
	Error           string     `json:"error,omitempty"`
	Metadata        string     `json:"metadata,omitempty"`
	DependsOn       []string   `json:"depends_on,omitempty"`

	// LatestRunID points to the most recent task_runs row for this task.
	// Empty when the task has never been executed.
	LatestRunID string `json:"latest_run_id,omitempty"`
}

// TaskRun represents a single execution attempt of a logical Task.
// Each time a task is retried or re-executed, a new TaskRun is created.
type TaskRun struct {
	RunID        string     `json:"run_id"`
	TaskID       string     `json:"task_id"`
	ProjectRoot  string     `json:"project_root"`
	Role         string     `json:"role"`
	Status       string     `json:"status"`
	OwnerAgentID string     `json:"owner_agent_id,omitempty"`
	ClaimedAt    *time.Time `json:"claimed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Result       string     `json:"result,omitempty"`
	Error        string     `json:"error,omitempty"`
	Metadata     string     `json:"metadata,omitempty"`
}

// Task status constants.
const (
	StatusQueued     = "queued"
	StatusInProgress = "in_progress"
	StatusPaused     = "paused"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
	StatusCancelled  = "cancelled"
	StatusBlocked    = "blocked"
)
