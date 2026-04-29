package taskdb

import "time"

// Task represents a task in the database.
type Task struct {
	ID              string     `json:"id"`
	BeadsID         string     `json:"beads_id,omitempty"` // Deprecated: legacy field for backward compatibility
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
	Metadata        string     `json:"metadata,omitempty"` // JSON string
}

// Task statuses
const (
	StatusQueued     = "queued"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
	StatusCancelled  = "cancelled"
)
