package execution_log

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ExecutionEvent represents a single agent lifecycle event
type ExecutionEvent struct {
	EventType   string                 `json:"event_type"`   // "spawned", "started", "progress", "completed", "failed", "cancelled"
	TaskID      string                 `json:"task_id"`
	Role        string                 `json:"role"`
	Task        string                 `json:"task"`
	Timestamp   time.Time              `json:"timestamp"`
	Status      string                 `json:"status,omitempty"`      // "queued", "in_progress", "completed", "failed", "cancelled"
	Error       string                 `json:"error,omitempty"`
	DurationMs  int64                  `json:"duration_ms,omitempty"` // For completed/failed events
	Result      string                 `json:"result,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ExecutionLog manages the persistent agent execution log
type ExecutionLog struct {
	logPath string
	mu      sync.Mutex
}

const (
	EventSpawned   = "spawned"
	EventStarted   = "started"
	EventProgress  = "progress"
	EventCompleted = "completed"
	EventFailed    = "failed"
	EventCancelled = "cancelled"
)

// NewExecutionLog creates a new execution log in global directory
func NewExecutionLog(rootDir string) (*ExecutionLog, error) {
	// Use global directory for machine-wide task visibility
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	logDir := filepath.Join(homeDir, ".ai-pack")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath := filepath.Join(logDir, "agent-execution.log")

	return &ExecutionLog{
		logPath: logPath,
	}, nil
}

// LogEvent appends an event to the execution log
func (el *ExecutionLog) LogEvent(event *ExecutionEvent) error {
	el.mu.Lock()
	defer el.mu.Unlock()

	// Open file in append mode
	f, err := os.OpenFile(el.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open execution log: %w", err)
	}
	defer f.Close()

	// Ensure timestamp is set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Marshal to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Write JSON line
	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	return nil
}

// LogSpawned logs when a task is spawned
func (el *ExecutionLog) LogSpawned(taskID, role, task string, metadata map[string]interface{}) error {
	return el.LogEvent(&ExecutionEvent{
		EventType: EventSpawned,
		TaskID:    taskID,
		Role:      role,
		Task:      task,
		Status:    "queued",
		Metadata:  metadata,
	})
}

// LogStarted logs when a task starts executing
func (el *ExecutionLog) LogStarted(taskID string) error {
	return el.LogEvent(&ExecutionEvent{
		EventType: EventStarted,
		TaskID:    taskID,
		Status:    "in_progress",
	})
}

// LogCompleted logs when a task completes successfully
func (el *ExecutionLog) LogCompleted(taskID string, durationMs int64, result string) error {
	return el.LogEvent(&ExecutionEvent{
		EventType:  EventCompleted,
		TaskID:     taskID,
		Status:     "completed",
		DurationMs: durationMs,
		Result:     result,
	})
}

// LogFailed logs when a task fails
func (el *ExecutionLog) LogFailed(taskID string, durationMs int64, errorMsg string) error {
	return el.LogEvent(&ExecutionEvent{
		EventType:  EventFailed,
		TaskID:     taskID,
		Status:     "failed",
		DurationMs: durationMs,
		Error:      errorMsg,
	})
}

// LogCancelled logs when a task is cancelled
func (el *ExecutionLog) LogCancelled(taskID string, durationMs int64) error {
	return el.LogEvent(&ExecutionEvent{
		EventType:  EventCancelled,
		TaskID:     taskID,
		Status:     "cancelled",
		DurationMs: durationMs,
	})
}

// GetRecentEvents returns recent events from the log
func (el *ExecutionLog) GetRecentEvents(limit int) ([]*ExecutionEvent, error) {
	el.mu.Lock()
	defer el.mu.Unlock()

	// Check if log file exists
	if _, err := os.Stat(el.logPath); os.IsNotExist(err) {
		return []*ExecutionEvent{}, nil
	}

	// Read entire file
	data, err := os.ReadFile(el.logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read execution log: %w", err)
	}

	// Parse JSON lines
	lines := splitLines(data)
	events := make([]*ExecutionEvent, 0, len(lines))

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		var event ExecutionEvent
		if err := json.Unmarshal(line, &event); err != nil {
			// Skip malformed lines
			continue
		}

		events = append(events, &event)
	}

	// Return last N events
	if limit > 0 && len(events) > limit {
		return events[len(events)-limit:], nil
	}

	return events, nil
}

// GetEventsSince returns events since a given timestamp
func (el *ExecutionLog) GetEventsSince(since time.Time) ([]*ExecutionEvent, error) {
	allEvents, err := el.GetRecentEvents(0) // Get all
	if err != nil {
		return nil, err
	}

	filtered := make([]*ExecutionEvent, 0)
	for _, event := range allEvents {
		if event.Timestamp.After(since) {
			filtered = append(filtered, event)
		}
	}

	return filtered, nil
}

// GetEventsByTaskID returns all events for a specific task
func (el *ExecutionLog) GetEventsByTaskID(taskID string) ([]*ExecutionEvent, error) {
	allEvents, err := el.GetRecentEvents(0) // Get all
	if err != nil {
		return nil, err
	}

	filtered := make([]*ExecutionEvent, 0)
	for _, event := range allEvents {
		if event.TaskID == taskID {
			filtered = append(filtered, event)
		}
	}

	return filtered, nil
}

// GetTaskHistory returns a summary of tasks with their final status
func (el *ExecutionLog) GetTaskHistory(limit int) ([]*TaskSummary, error) {
	events, err := el.GetRecentEvents(0) // Get all events
	if err != nil {
		return nil, err
	}

	// Build task map
	taskMap := make(map[string]*TaskSummary)

	for _, event := range events {
		summary, exists := taskMap[event.TaskID]
		if !exists {
			summary = &TaskSummary{
				TaskID:  event.TaskID,
				Role:    event.Role,
				Task:    event.Task,
				Status:  event.Status,
				Created: event.Timestamp,
				Updated: event.Timestamp,
			}
			taskMap[event.TaskID] = summary
		}

		// Update with latest info
		if event.Timestamp.After(summary.Updated) {
			summary.Updated = event.Timestamp
			summary.Status = event.Status

			if event.Error != "" {
				summary.Error = event.Error
			}
			if event.DurationMs > 0 {
				summary.DurationMs = event.DurationMs
			}
			if event.Result != "" {
				summary.Result = event.Result
			}
		}

		// Track event types
		summary.Events = append(summary.Events, event.EventType)
	}

	// Convert to slice
	summaries := make([]*TaskSummary, 0, len(taskMap))
	for _, summary := range taskMap {
		summaries = append(summaries, summary)
	}

	// Sort by updated time (most recent first)
	sortByUpdatedDesc(summaries)

	// Apply limit
	if limit > 0 && len(summaries) > limit {
		return summaries[:limit], nil
	}

	return summaries, nil
}

// TaskSummary represents a summary of a task's lifecycle
type TaskSummary struct {
	TaskID     string    `json:"task_id"`
	Role       string    `json:"role"`
	Task       string    `json:"task"`
	Status     string    `json:"status"`
	Created    time.Time `json:"created"`
	Updated    time.Time `json:"updated"`
	DurationMs int64     `json:"duration_ms"`
	Error      string    `json:"error,omitempty"`
	Result     string    `json:"result,omitempty"`
	Events     []string  `json:"events"` // Event types seen
}

// Helper functions

func splitLines(data []byte) [][]byte {
	lines := make([][]byte, 0)
	start := 0
	for i, b := range data {
		if b == '\n' {
			if i > start {
				lines = append(lines, data[start:i])
			}
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}

func sortByUpdatedDesc(summaries []*TaskSummary) {
	// Simple bubble sort (good enough for small lists)
	n := len(summaries)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if summaries[j].Updated.Before(summaries[j+1].Updated) {
				summaries[j], summaries[j+1] = summaries[j+1], summaries[j]
			}
		}
	}
}
