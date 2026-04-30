package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/constants"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/internal/taskdb"
)

func (s *AgentServer) handleOrphanedTasks() {
	// Wait a bit for server to fully initialize
	time.Sleep(2 * time.Second)

	orphanedCount := 0
	staleMeta := 0

	// First pass: Detect orphaned tasks using taskDB
	// Tasks marked in_progress in database but not in activeTasks map
	if s.taskDB != nil {
		inProgressTasks, err := s.taskDB.ListTasks(taskdb.TaskFilter{
			Status: taskdb.StatusInProgress,
			Limit:  1000,
		})
		if err != nil {
			monitoring.Logger.Error("failed_to_list_in_progress_tasks", "error", err.Error())
			return
		}

		for _, dbTask := range inProgressTasks {
			// Check if task is actively running
			s.mu.RLock()
			_, hasActiveTask := s.activeTasks[dbTask.ID]
			s.mu.RUnlock()

			if !hasActiveTask {
				// Orphaned task detected - reset to failed
				monitoring.Logger.Warn("orphaned_task_detected",
					"task_id", dbTask.ID,
					"short_id", taskdb.ExtractShortID(dbTask.ID),
					"role", dbTask.Role,
					"project", dbTask.ProjectRoot,
					"action", "marking_failed",
				)

				// Mark as failed in database
				if err := s.taskDB.FailTask(dbTask.ID, "Task orphaned - server restarted or crashed"); err != nil {
					monitoring.Logger.Error("failed_to_mark_orphaned_task_failed",
						"task_id", dbTask.ID,
						"error", err.Error())
					continue
				}

				// Delete checkpoint if present
				cpPath := checkpointPath(dbTask.ProjectRoot, dbTask.ID)
				if removeErr := os.Remove(cpPath); removeErr == nil {
					monitoring.Logger.Info("orphaned_task_checkpoint_deleted",
						"task_id", dbTask.ID,
						"checkpoint", cpPath)
				} else if !os.IsNotExist(removeErr) {
					monitoring.Logger.Warn("failed_to_delete_orphaned_task_checkpoint",
						"task_id", dbTask.ID,
						"checkpoint", cpPath,
						"error", removeErr.Error())
				}

				orphanedCount++
			}
		}
	}

	// Second pass: scan execution folders for stale in_progress metadata.
	// This catches tasks that were interrupted mid-run (e.g. server killed) whose
	// metadata file still says "in_progress" but aren't in activeTasks or taskDB.
	projectRoots := s.GetProjectRoots()
	for _, projectRoot := range projectRoots {
		tasksDir := filepath.Join(projectRoot, constants.TaskRootDir, "tasks")
		entries, err := os.ReadDir(tasksDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			folderName := entry.Name()
			metadataPath := filepath.Join(tasksDir, folderName, constants.MetadataFileName)
			data, err := os.ReadFile(metadataPath)
			if err != nil {
				continue
			}
			var meta map[string]interface{}
			if json.Unmarshal(data, &meta) != nil {
				continue
			}
			if meta["status"] != "in_progress" {
				continue
			}
			// Skip folders spawned very recently — they may belong to a task that was
			// running in the previous server process which hasn't fully exited yet.
			if spawnedStr, ok := meta["spawned_at"].(string); ok {
				if spawned, err := time.Parse(time.RFC3339, spawnedStr); err == nil {
					if time.Since(spawned) < 30*time.Second {
						monitoring.Logger.Info("skipping_recently_spawned_execution",
							"folder", folderName,
							"spawned_at", spawnedStr,
							"age_seconds", int(time.Since(spawned).Seconds()),
						)
						continue
					}
				}
			}
			// Not active — mark it failed so the GUI doesn't show it as running
			s.mu.RLock()
			_, isActive := s.activeTasks[folderName]
			s.mu.RUnlock()
			if isActive {
				continue
			}
			meta["status"] = constants.StatusFailed
			meta["error"] = "interrupted: server restarted while task was running"
			meta["updated_at"] = time.Now().Format(time.RFC3339)
			if updated, err := json.MarshalIndent(meta, "", "  "); err == nil {
				if writeErr := os.WriteFile(metadataPath, updated, 0644); writeErr == nil {
					monitoring.Logger.Info("stale_execution_marked_failed",
						"folder", folderName,
						"project", projectRoot,
					)
					staleMeta++
				}
			}
		}
	}

	if orphanedCount > 0 || staleMeta > 0 {
		monitoring.Logger.Info("task_cleanup_summary",
			"orphaned", orphanedCount,
			"stale_metadata", staleMeta)
	}
}
