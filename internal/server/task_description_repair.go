package server

import (
	"encoding/json"
	"strings"

	"github.com/cortexa-llc/ai-pack/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/internal/taskdb"
)

// RepairTaskDescriptions scans the database for tasks with generic "Task {id}" descriptions
// and updates them with real descriptions from contract files.
func (s *AgentServer) RepairTaskDescriptions() error {
	if s.taskDB == nil {
		return nil
	}

	monitoring.Logger.Info("task_description_repair_started")

	// Get all tasks
	tasks, err := s.taskDB.ListTasks(taskdb.TaskFilter{
		Limit: 10000, // Get all tasks
	})
	if err != nil {
		return err
	}

	repairedCount := 0
	for _, task := range tasks {
		// Skip if description is already meaningful (not generic "Task X" and not empty)
		if task.TaskDescription != "" && !strings.HasPrefix(task.TaskDescription, "Task ") {
			continue
		}

		shortID := taskdb.ExtractShortID(task.ID)
		taskPacketPath := ""

		// First try to get from metadata
		if task.Metadata != "" {
			var metadata map[string]string
			if err := json.Unmarshal([]byte(task.Metadata), &metadata); err == nil {
				taskPacketPath = metadata["task_packet_path"]
			}
		}

		// If no metadata, try to find task packet by scanning filesystem
		if taskPacketPath == "" && task.ProjectRoot != "" {
			taskPacketPath = s.findTaskPacketPath(shortID, task.ProjectRoot)
		}

		// Read description from contract if found
		if taskPacketPath != "" {
			description := s.readTaskDescriptionFromContract(taskPacketPath, task.ProjectRoot)
			if description != "" && description != task.TaskDescription {
				// Update the task description in database
				_, err := s.taskDB.Exec(`UPDATE tasks SET task_description = ? WHERE id = ?`, description, task.ID)
				if err != nil {
					monitoring.Logger.Warn("failed_to_update_task_description",
						"task_id", task.ID,
						"error", err.Error())
				} else {
					monitoring.Logger.Debug("task_description_repaired",
						"task_id", shortID,
						"old_description", task.TaskDescription,
						"new_description", description)
					repairedCount++
				}
			}
		}
	}

	monitoring.Logger.Info("task_description_repair_complete",
		"total_tasks", len(tasks),
		"repaired_count", repairedCount)

	return nil
}

// CleanupOrphanedTasks removes tasks from the database that have no description
// and no task packet directory (orphaned/incomplete tasks).
func (s *AgentServer) CleanupOrphanedTasks() error {
	if s.taskDB == nil {
		return nil
	}

	monitoring.Logger.Info("orphaned_task_cleanup_started")

	// Get all tasks
	tasks, err := s.taskDB.ListTasks(taskdb.TaskFilter{
		Limit: 10000,
	})
	if err != nil {
		return err
	}

	deletedCount := 0
	for _, task := range tasks {
		// Only consider tasks with no meaningful description
		hasDescription := task.TaskDescription != "" && !strings.HasPrefix(task.TaskDescription, "Task ")
		if hasDescription {
			continue
		}

		shortID := taskdb.ExtractShortID(task.ID)

		// Check if task packet exists
		taskPacketPath := ""
		if task.Metadata != "" {
			var metadata map[string]string
			if err := json.Unmarshal([]byte(task.Metadata), &metadata); err == nil {
				taskPacketPath = metadata["task_packet_path"]
			}
		}

		// Try to find task packet by scanning filesystem
		if taskPacketPath == "" && task.ProjectRoot != "" {
			taskPacketPath = s.findTaskPacketPath(shortID, task.ProjectRoot)
		}

		// Delete if no task packet found
		if taskPacketPath == "" {
			if err := s.taskDB.DeleteTask(task.ID); err != nil {
				monitoring.Logger.Warn("failed_to_delete_orphaned_task",
					"task_id", shortID,
					"error", err.Error())
			} else {
				monitoring.Logger.Info("orphaned_task_deleted",
					"task_id", shortID,
					"full_id", task.ID,
					"project_root", task.ProjectRoot)
				deletedCount++
			}
		}
	}

	monitoring.Logger.Info("orphaned_task_cleanup_complete",
		"total_tasks", len(tasks),
		"deleted_count", deletedCount)

	return nil
}
