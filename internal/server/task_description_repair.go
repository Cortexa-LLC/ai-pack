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
		// Skip if description is not generic
		if !strings.HasPrefix(task.TaskDescription, "Task ") {
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
