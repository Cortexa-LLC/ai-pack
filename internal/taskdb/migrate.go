package taskdb

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MigrateFromLegacy imports tasks from .beads/tasks directories into the SQLite database.
// It scans all task directories and creates corresponding database entries.
func (db *DB) MigrateFromLegacy(projectRoot string) (int, error) {
	tasksDir := filepath.Join(projectRoot, ".beads", "tasks")

	// Check if tasks directory exists
	if _, err := os.Stat(tasksDir); os.IsNotExist(err) {
		return 0, nil // No tasks to migrate
	}

	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read tasks directory: %w", err)
	}

	migrated := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		taskID := entry.Name()
		metadataPath := filepath.Join(tasksDir, taskID, "00-metadata.json")

		// Read metadata file
		data, err := os.ReadFile(metadataPath)
		if err != nil {
			// Skip tasks without metadata
			continue
		}

		var metadata map[string]interface{}
		if err := json.Unmarshal(data, &metadata); err != nil {
			continue
		}

		// Check if task already exists in database
		existing, _ := db.GetTask(taskID)
		if existing != nil {
			continue // Skip already migrated tasks
		}

		// Map metadata to Task struct
		task := &Task{
			ID:              taskID,
			ProjectRoot:     projectRoot,
			Role:            getString(metadata, "role"),
			TaskDescription: getString(metadata, "task"),
			Status:          mapStatus(getString(metadata, "status")),
		}

		// Parse timestamps
		if createdStr := getString(metadata, "spawned_at"); createdStr != "" {
			if t, err := time.Parse(time.RFC3339, createdStr); err == nil {
				task.CreatedAt = t
			}
		}
		if task.CreatedAt.IsZero() {
			task.CreatedAt = time.Now()
		}

		if updatedStr := getString(metadata, "updated_at"); updatedStr != "" {
			if t, err := time.Parse(time.RFC3339, updatedStr); err == nil {
				task.UpdatedAt = t
			}
		}
		if task.UpdatedAt.IsZero() {
			task.UpdatedAt = task.CreatedAt
		}

		// Extract error message if present
		if errorMsg := getString(metadata, "error"); errorMsg != "" {
			task.Error = errorMsg
		}

		// Read result if task is completed
		if task.Status == StatusCompleted || task.Status == StatusFailed {
			resultPath := filepath.Join(tasksDir, taskID, "result.txt")
			if resultData, err := os.ReadFile(resultPath); err == nil {
				task.Result = string(resultData)
			}

			completedTime := task.UpdatedAt
			task.CompletedAt = &completedTime
		}

		// Insert task (bypass CreateTask to preserve original timestamps)
		query := `
			INSERT INTO tasks (
				id, project_root, role, task_description, status,
				created_at, updated_at, completed_at, result, error
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		_, err = db.db.Exec(query,
			task.ID, task.ProjectRoot, task.Role,
			task.TaskDescription, task.Status,
			task.CreatedAt, task.UpdatedAt, task.CompletedAt,
			task.Result, task.Error,
		)

		if err != nil {
			return migrated, fmt.Errorf("failed to insert task %s: %w", taskID, err)
		}

		migrated++
	}

	return migrated, nil
}

// extractShortTaskID extracts the short task ID from a full task ID.
// Example: "listingsgql-5b6-20260424-162611" -> "listingsgql-5b6"
func extractShortTaskID(taskID string) string {
	// Split by hyphen and take everything before the timestamp
	parts := strings.Split(taskID, "-")
	if len(parts) < 3 {
		return taskID
	}

	// Find the first part that looks like a timestamp (YYYYMMDD)
	for i := 2; i < len(parts); i++ {
		if len(parts[i]) == 8 && isNumeric(parts[i]) {
			// Everything before this is the task ID
			return strings.Join(parts[:i], "-")
		}
	}

	return taskID
}

// mapStatus maps metadata status to database status constants.
func mapStatus(metaStatus string) string {
	switch metaStatus {
	case "in_progress", "working", "running":
		return StatusInProgress
	case "completed", "done", "success":
		return StatusCompleted
	case "failed", "error":
		return StatusFailed
	case "cancelled", "canceled":
		return StatusCancelled
	default:
		return StatusQueued
	}
}

// getString safely extracts a string from a map.
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// isNumeric checks if a string contains only digits.
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
