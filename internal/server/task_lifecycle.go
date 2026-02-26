package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/beads"

	"github.com/cortexa-llc/ai-pack/internal/constants"
	"github.com/cortexa-llc/ai-pack/internal/kgclient"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/internal/protocol"
)

func (s *AgentServer) markExecutionAsSuperseded(taskID, projectRoot, reason string) error {
	// Find the execution metadata file
	var metadataPath string

	if projectRoot == "" {
		// Search all project roots
		for _, root := range s.GetProjectRoots() {
			path := filepath.Join(root, constants.BeadsDir, "tasks", taskID, constants.MetadataFileName)
			if _, err := os.Stat(path); err == nil {
				metadataPath = path
				break
			}
		}
	} else {
		metadataPath = filepath.Join(projectRoot, constants.BeadsDir, "tasks", taskID, constants.MetadataFileName)
	}

	if metadataPath == "" {
		return fmt.Errorf("metadata not found for task %s", taskID)
	}

	// Read existing metadata
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to read metadata: %w", err)
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		return fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Mark as superseded
	metadata["superseded"] = true
	metadata["superseded_at"] = time.Now().Format(time.RFC3339)
	metadata["superseded_reason"] = reason
	metadata["updated_at"] = time.Now().Format(time.RFC3339)

	// Write back
	updatedData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	monitoring.Logger.Info("marked_execution_superseded",
		"task_id", taskID,
		"reason", reason)

	return nil
}

// findMostRecentExecutionFolderInRoot finds the most recent timestamped execution folder for a Beads task ID in the server root

func (s *AgentServer) updateTaskStatus(taskID, projectRoot, status, errorMsg string) error {
	metadataPath := filepath.Join(projectRoot, constants.BeadsDir, "tasks", taskID, constants.MetadataFileName)
	monitoring.Logger.Info("updating_task_status", "task_id", taskID, "status", status, "path", metadataPath)

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		monitoring.Logger.Error("metadata_read_error", "task_id", taskID, "path", metadataPath, "error", err)
		return fmt.Errorf("failed to read metadata: %w", err)
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		monitoring.Logger.Error("metadata_unmarshal_error", "task_id", taskID, "error", err)
		return fmt.Errorf("failed to parse metadata: %w", err)
	}

	metadata["status"] = status
	metadata["updated_at"] = time.Now().Format(time.RFC3339)
	if errorMsg != "" {
		metadata["error"] = errorMsg
	}

	updatedData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		monitoring.Logger.Error("metadata_marshal_error", "task_id", taskID, "error", err)
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, updatedData, 0644); err != nil {
		monitoring.Logger.Error("metadata_write_error", "task_id", taskID, "path", metadataPath, "error", err)
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	monitoring.Logger.Info("task_status_updated", "task_id", taskID, "status", status)
	return nil
}

// buildSystemPrompt returns the system prompt string for the agentic loop.
// loadSharedPolicy loads the shared agent-policy.md that applies to all agents.

func (s *AgentServer) saveAndCompleteTask(ctx context.Context, execution *TaskExecution, result string, startTime time.Time, logMsg func(string)) {
	// Save results
	s.saveTaskResults(execution, result, logMsg)

	// Persist task outcome to the knowledge graph (best-effort, non-blocking).
	// A snapshot of the variables is taken so the goroutine doesn't race on
	// fields that may be mutated after saveAndCompleteTask returns.
	go kgclient.WriteBack(
		context.Background(),
		s.mcpManager,
		execution.ProjectRoot,
		execution.Role,
		execution.Task,
		result,
		startTime,
	)

	// Detect if the agent stopped due to blockers (missing task packets, etc.)
	// Check for common blocker patterns in the result text
	resultLower := strings.ToLower(result)
	isBlocked := strings.Contains(resultLower, "task packet missing") ||
		strings.Contains(resultLower, "blocking issue") ||
		strings.Contains(resultLower, "stop - task packet") ||
		strings.Contains(resultLower, "cannot proceed") ||
		strings.Contains(resultLower, "stopping - task packet required")

	var finalStatus string
	var statusMessage string

	if isBlocked {
		finalStatus = "blocked"
		statusMessage = "Task blocked: Agent stopped due to missing prerequisites (task packet)"
		logMsg("⚠️  Task marked as BLOCKED - agent identified missing prerequisites")
	} else {
		finalStatus = constants.StatusCompleted
		statusMessage = ""
	}

	// Update task status
	beadsTaskID, projectRoot := s.updateTaskCompletion(execution, result)

	// Complete Beads task only if truly completed (not blocked)
	var errorMsg string
	if beadsTaskID != "" && !isBlocked {
		if err := s.completeBeadsTask(beadsTaskID, projectRoot, logMsg); err != nil {
			errorMsg = fmt.Sprintf("Warning: %v", err)
			monitoring.Logger.Warn("beads_update_failed_but_task_completed", "task_id", execution.TaskID, "error", err.Error())
		}
	} else if isBlocked {
		errorMsg = statusMessage
	}

	// Finalize task
	durationMs := time.Since(startTime).Milliseconds()
	if err := s.updateTaskStatus(execution.TaskID, execution.ProjectRoot, finalStatus, errorMsg); err != nil {
		monitoring.Logger.Error("failed_to_update_status", "task_id", execution.TaskID, "status", finalStatus, "error", err)
		logMsg(fmt.Sprintf("⚠️  Warning: Failed to update status: %v", err))
	}
	s.sendStreamEvent(execution, finalStatus, map[string]interface{}{
		"result": result,
	})

	s.closeStream(execution)

	// Log completion event
	if s.executionLog != nil {
		resultSummary := result
		if len(resultSummary) > 500 {
			resultSummary = resultSummary[:500] + "..."
		}
		if err := s.executionLog.LogCompleted(execution.TaskID, durationMs, resultSummary); err != nil {
			monitoring.Logger.Warn("failed_to_log_completed_event", "error", err.Error())
		}
	}

	// Log appropriate message based on final status
	if finalStatus == "blocked" {
		logMsg(fmt.Sprintf("⚠️  Task blocked - prerequisites missing (duration: %dms)", durationMs))
		monitoring.Logger.Warn("task_blocked", "task_id", execution.TaskID, "duration_ms", durationMs)
	} else {
		logMsg(fmt.Sprintf("🎉 Task completed successfully (duration: %dms)", durationMs))
		monitoring.LogTaskCompleted(ctx, execution.TaskID, execution.Role, durationMs)
		monitoring.GlobalMetrics.IncrementTasksCompleted(durationMs)

		// Record performance grade for adaptive model selection
		if monitoring.GlobalGradeManager != nil && execution.metadata != nil {
			modelID := s.model // Default model
			if executionModel, ok := execution.metadata["model"]; ok {
				modelID = executionModel
			}

			tokensUsed := int64(0)
			if tokensStr, ok := execution.metadata["total_tokens"]; ok {
				fmt.Sscanf(tokensStr, "%d", &tokensUsed)
			}

			// Opt-in out-of-scope file detection via git diff.
			// workingDir narrows the expected write scope; any git-tracked
			// changes outside that directory are considered catastrophic
			// (e.g. the model overwrote entity.go while working in /tmp/task/).
			workingDir := ""
			if execution.metadata != nil {
				workingDir = execution.metadata["working_directory"]
			}
			outOfScope, _ := detectOutOfScopeChanges(projectRoot, workingDir)
			if len(outOfScope) > 0 {
				reason := fmt.Sprintf("agent modified %d file(s) outside working directory: %s",
					len(outOfScope), strings.Join(outOfScope, ", "))
				logMsg(fmt.Sprintf("🚨 CATASTROPHIC: %s", reason))
				monitoring.Logger.Error("catastrophic_out_of_scope_writes",
					"task_id", execution.TaskID,
					"model", modelID,
					"files", outOfScope,
				)
				if err := monitoring.GlobalGradeManager.RecordCatastrophicFailure(
					execution.TaskID,
					modelID,
					execution.Role,
					projectRoot,
					reason,
				); err != nil {
					monitoring.Logger.Warn("failed_to_record_catastrophic_failure", "error", err.Error())
				}
			} else {
				// Record successful completion
				if err := monitoring.GlobalGradeManager.RecordTaskCompletion(
					execution.TaskID,
					modelID,
					execution.Role,
					projectRoot,
					true, // success
					0,    // retries (we don't track this yet)
					tokensUsed,
					durationMs,
					false, // wasEscalated (not tracked yet)
					false, // wasDowngraded (not tracked yet)
				); err != nil {
					monitoring.Logger.Warn("failed_to_record_performance_grade", "error", err.Error())
				} else {
					monitoring.Logger.Debug("performance_grade_recorded", "task_id", execution.TaskID, "model", modelID, "role", execution.Role)
				}
			}
		}
	}
	logMsg("=" + strings.Repeat("=", 70))
}

// saveTaskResults saves the task results to disk
func (s *AgentServer) saveTaskResults(execution *TaskExecution, result string, logMsg func(string)) {
	logMsg("💾 Saving results...")
	resultsPath := filepath.Join(execution.ProjectRoot, constants.BeadsDir, "tasks", execution.TaskID, "30-results.md")
	resultsContent := fmt.Sprintf("# Task Results: %s\n\n**Role**: %s\n**Task**: %s\n**Completed**: %s\n\n## Agent Output\n\n%s\n",
		execution.TaskID, execution.Role, execution.Task, time.Now().Format(time.RFC3339), result)

	if err := os.WriteFile(resultsPath, []byte(resultsContent), 0644); err != nil {
		monitoring.Logger.Warn("results_save_error", "task_id", execution.TaskID, "error", err)
		logMsg(fmt.Sprintf("⚠️  Failed to save results: %v", err))
	} else {
		logMsg(fmt.Sprintf("✅ Results saved: %s", resultsPath))
		logMsg(fmt.Sprintf("   Output length: %d chars", len(result)))
	}
}

// updateTaskCompletion updates execution status and extracts Beads task ID and project root
func (s *AgentServer) updateTaskCompletion(execution *TaskExecution, result string) (string, string) {
	s.mu.Lock()
	execution.Status = constants.StatusCompleted
	execution.Result = result
	// Use the short Beads task ID (e.g. "xasm++-qbxv") stored in metadata,
	// not the full execution TaskID (e.g. "xasm++-qbxv-20260218-084509").
	beadsTaskID := execution.TaskID // fallback to full ID if metadata missing
	projectRoot := ""
	if execution.metadata != nil {
		if id, ok := execution.metadata["beads_task_id"]; ok && id != "" {
			beadsTaskID = id
		}
		projectRoot = execution.metadata["project_root"]
	}
	// Remove from active tasks map since task is now completed
	delete(s.activeTasks, execution.TaskID)
	s.mu.Unlock()

	return beadsTaskID, projectRoot
}

// completeBeadsTask marks the corresponding Beads task as complete
// Returns error if the Beads update failed
func (s *AgentServer) completeBeadsTask(beadsTaskID string, projectRoot string, logMsg func(string)) error {
	if !beads.IsInstalled() {
		return nil
	}

	logMsg(fmt.Sprintf("🔗 Marking Beads task complete: %s", beadsTaskID))
	if err := s.beadsClient.CompleteTaskFromDir(beadsTaskID, projectRoot); err != nil {
		monitoring.Logger.Warn("failed_to_complete_beads_task", "task_id", beadsTaskID, "error", err.Error())
		logMsg(fmt.Sprintf("⚠️  Failed to complete Beads task: %v", err))
		return fmt.Errorf("beads update failed: %w", err)
	}

	monitoring.Logger.Info("beads_task_completed", "task_id", beadsTaskID)
	logMsg("✅ Beads task marked complete")
	return nil
}

// resetBeadsTask resets an in-progress beads task back to "open" when execution
// fails or is cancelled, so it no longer appears as RUNNING in the GUI after restart.
func (s *AgentServer) resetBeadsTask(execution *TaskExecution) {
	if !beads.IsInstalled() {
		return
	}
	beadsTaskID := execution.TaskID
	projectRoot := execution.ProjectRoot
	if execution.metadata != nil {
		if id, ok := execution.metadata["beads_task_id"]; ok && id != "" {
			beadsTaskID = id
		}
		if pr := execution.metadata["project_root"]; pr != "" {
			projectRoot = pr
		}
	}
	if !beads.IsBeadsTaskID(beadsTaskID) {
		return
	}
	cmd := exec.Command("bd", "update", "-s", "open", beadsTaskID)
	if projectRoot != "" {
		cmd.Dir = projectRoot
	}
	if err := cmd.Run(); err != nil {
		monitoring.Logger.Warn("failed_to_reset_beads_task", "task_id", beadsTaskID, "error", err.Error())
	} else {
		monitoring.Logger.Info("beads_task_reset_to_open", "task_id", beadsTaskID)
	}
}

func (s *AgentServer) failTask(execution *TaskExecution, errorMsg string) {
	ctx := context.Background()
	durationMs := time.Since(execution.StartTime).Milliseconds()

	// Log failure to execution log
	logPath := filepath.Join(execution.ProjectRoot, constants.BeadsDir, "tasks", execution.TaskID, "execution.log")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		timestamp := time.Now().Format("15:04:05")
		_, _ = logFile.WriteString(fmt.Sprintf("[%s] ❌ Task failed: %s\n", timestamp, errorMsg))
		_, _ = logFile.WriteString(fmt.Sprintf("[%s]    Duration: %dms\n", timestamp, durationMs))
		_, _ = logFile.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, strings.Repeat("=", 70)))
		_ = logFile.Close()
	}

	s.mu.Lock()
	execution.Status = constants.StatusFailed
	execution.Error = errorMsg
	// Remove from active tasks map since task is now failed
	delete(s.activeTasks, execution.TaskID)
	s.mu.Unlock()

	if err := s.updateTaskStatus(execution.TaskID, execution.ProjectRoot, constants.StatusFailed, errorMsg); err != nil {
		monitoring.Logger.Error("failed_to_update_failed_status", "task_id", execution.TaskID, "error", err)
	}
	// Reset beads task to open so it does not appear as RUNNING after a server restart.
	s.resetBeadsTask(execution)
	s.sendStreamEvent(execution, constants.StatusFailed, map[string]interface{}{
		"error": errorMsg,
	})
	s.closeStream(execution)

	// Log failed event
	if s.executionLog != nil {
		if err := s.executionLog.LogFailed(execution.TaskID, durationMs, errorMsg); err != nil {
			monitoring.Logger.Warn("failed_to_log_failed_event", "error", err.Error())
		}
	}

	monitoring.LogTaskFailed(ctx, execution.TaskID, execution.Role, errorMsg, durationMs)
	monitoring.GlobalMetrics.IncrementTasksFailed(durationMs)

	// Record performance grade for adaptive model selection
	if monitoring.GlobalGradeManager != nil && execution.metadata != nil {
		modelID := s.model // Default model
		if executionModel, ok := execution.metadata["model"]; ok {
			modelID = executionModel
		}

		projectRoot := ""
		if pr, ok := execution.metadata["project_root"]; ok {
			projectRoot = pr
		}

		tokensUsed := int64(0)
		if tokensStr, ok := execution.metadata["total_tokens"]; ok {
			fmt.Sscanf(tokensStr, "%d", &tokensUsed)
		}

		// Record failed completion
		if err := monitoring.GlobalGradeManager.RecordTaskCompletion(
			execution.TaskID,
			modelID,
			execution.Role,
			projectRoot,
			false, // success = false
			0,     // retries (we don't track this yet)
			tokensUsed,
			durationMs,
			false, // wasEscalated (not tracked yet)
			false, // wasDowngraded (not tracked yet)
		); err != nil {
			monitoring.Logger.Warn("failed_to_record_performance_grade", "error", err.Error())
		}
	}
}

func (s *AgentServer) CancelTask(taskID string) error {
	s.mu.Lock()
	execution, exists := s.activeTasks[taskID]
	if !exists {
		// Try prefix match for short Beads IDs (e.g. "xasm++-qbxv" → "xasm++-qbxv-20260218-101958")
		prefix := taskID + "-"
		for _, exec := range s.activeTasks {
			if strings.HasPrefix(exec.TaskID, prefix) {
				execution = exec
				exists = true
				break
			}
			if exec.metadata != nil {
				if btid, ok := exec.metadata["beads_task_id"]; ok && btid == taskID {
					execution = exec
					exists = true
					break
				}
			}
		}
	}
	s.mu.Unlock()

	if !exists {
		return fmt.Errorf("task not found or not active: %s", taskID)
	}

	// Call the cancel function to trigger context cancellation
	// This will cause ctx.Err() to return context.Canceled in the execution loop
	if execution.cancel != nil {
		execution.cancel()

		// Log cancellation to execution log
		logPath := filepath.Join(execution.ProjectRoot, constants.BeadsDir, "tasks", execution.TaskID, "execution.log")
		logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			timestamp := time.Now().Format("15:04:05")
			_, _ = logFile.WriteString(fmt.Sprintf("[%s] 🛑 Task cancelled by user\n", timestamp))
			_, _ = logFile.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, strings.Repeat("=", 70)))
			_ = logFile.Close()
		}

		return nil
	}

	return fmt.Errorf("task cannot be cancelled: %s", taskID)
}

// cancelTaskExecution marks a task as cancelled (called after context cancellation)
// This is invoked when a task is cancelled via CancelTask() or when the context
// is cancelled/times out during execution.
func (s *AgentServer) cancelTaskExecution(execution *TaskExecution, message string) {
	ctx := context.Background()
	durationMs := time.Since(execution.StartTime).Milliseconds()

	s.mu.Lock()
	execution.Status = "cancelled"
	execution.Error = message
	// Remove from active tasks map since task is now cancelled
	delete(s.activeTasks, execution.TaskID)
	s.mu.Unlock()

	if err := s.updateTaskStatus(execution.TaskID, execution.ProjectRoot, "cancelled", message); err != nil {
		monitoring.Logger.Error("failed_to_update_cancelled_status", "task_id", execution.TaskID, "error", err)
	}
	// Reset beads task to open so it does not appear as RUNNING after a server restart.
	s.resetBeadsTask(execution)
	s.sendStreamEvent(execution, "cancelled", map[string]interface{}{
		"message": message,
	})
	s.closeStream(execution)

	// Log cancelled event
	if s.executionLog != nil {
		if err := s.executionLog.LogCancelled(execution.TaskID, durationMs); err != nil {
			monitoring.Logger.Warn("failed_to_log_cancelled_event", "error", err.Error())
		}
	}

	monitoring.LogTaskFailed(ctx, execution.TaskID, execution.Role, message, durationMs)
	monitoring.GlobalMetrics.IncrementTasksFailed(durationMs)
}

func (s *AgentServer) sendStreamEvent(execution *TaskExecution, eventType string, data map[string]interface{}) {
	if !execution.streamOpen {
		return
	}

	event := &protocol.StreamEvent{
		Type:      eventType,
		TaskID:    execution.TaskID,
		Timestamp: time.Now(),
		Data:      data,
	}

	// Send to channel for live streaming
	select {
	case execution.streamChan <- event:
	default:
		// Channel full, skip event (but still write to file)
	}

	// Also write to per-task log file for historical access
	s.writeStreamEventToFile(execution, event)
}

// writeStreamEventToFile appends a stream event to the per-task log file
func (s *AgentServer) writeStreamEventToFile(execution *TaskExecution, event *protocol.StreamEvent) {
	// Build path to task log directory
	logDir := filepath.Join(execution.ProjectRoot, constants.BeadsDir, "tasks", execution.TaskID)
	logPath := filepath.Join(logDir, "execution.log")

	// Ensure directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		monitoring.Logger.Warn("failed_to_create_log_dir", "task_id", execution.TaskID, "error", err.Error())
		return
	}

	// Marshal event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		monitoring.Logger.Warn("failed_to_marshal_stream_event", "task_id", execution.TaskID, "error", err.Error())
		return
	}

	// Append to log file
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		monitoring.Logger.Warn("failed_to_open_task_log", "task_id", execution.TaskID, "error", err.Error())
		return
	}
	defer f.Close()

	// Write JSON line
	if _, err := f.Write(append(data, '\n')); err != nil {
		monitoring.Logger.Warn("failed_to_write_stream_event", "task_id", execution.TaskID, "error", err.Error())
	}
}

func (s *AgentServer) closeStream(execution *TaskExecution) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if execution.streamOpen {
		execution.streamOpen = false
		close(execution.streamChan)
	}
}

// detectOutOfScopeChanges runs "git diff --name-only HEAD" in projectRoot and
// returns paths of files that were modified outside workingDir.
//
// Both projectRoot and workingDir must be non-empty and workingDir must be an
// ancestor-or-equal directory of files we care about.  The function is a
// best-effort check: if git is unavailable or the directory is not a git
// repository it returns nil, nil so callers can treat it as "no violations".
func detectOutOfScopeChanges(projectRoot, workingDir string) ([]string, error) {
	if projectRoot == "" || workingDir == "" {
		return nil, nil
	}

	// Normalise to absolute paths before comparison.
	absProject, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil, nil
	}
	absWorking, err := filepath.Abs(workingDir)
	if err != nil {
		return nil, nil
	}

	// If workingDir IS the project root, every file is in-scope.
	if absProject == absWorking {
		return nil, nil
	}

	// Ensure workingDir is actually inside projectRoot (guard against bad config).
	rel, err := filepath.Rel(absProject, absWorking)
	if err != nil || strings.HasPrefix(rel, "..") {
		return nil, nil
	}

	cmd := exec.Command("git", "-C", absProject, "diff", "--name-only", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		// Not a git repo, git not installed, or no commits yet – silently skip.
		return nil, nil
	}

	var outOfScope []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		// git diff --name-only returns paths relative to the repo root.
		absFile := filepath.Join(absProject, line)
		// A file is "in scope" if it lives inside workingDir.
		fileRel, err := filepath.Rel(absWorking, absFile)
		if err != nil || strings.HasPrefix(fileRel, "..") {
			outOfScope = append(outOfScope, line)
		}
	}

	return outOfScope, nil
}
