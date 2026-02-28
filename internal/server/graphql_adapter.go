package server

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/beads"
	"github.com/cortexa-llc/ai-pack/internal/constants"
	"github.com/cortexa-llc/ai-pack/internal/graphql"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
)

// beadsTaskGetter is a narrow interface used by applyBeadsStatusOverride so that
// tests can supply a fake without spawning a real `bd` process.
type beadsTaskGetter interface {
	GetTask(taskID string) (*beads.Task, error)
}

// GraphQLAdapter adapts AgentServer to implement graphql.ServerInterface
// This allows GraphQL resolvers to access server functionality without circular dependencies
type GraphQLAdapter struct {
	server     *AgentServer
	taskGetter beadsTaskGetter // defaults to server.beadsClient; overridable in tests
}

// NewGraphQLAdapter creates a new GraphQL adapter for the agent server
func NewGraphQLAdapter(server *AgentServer) *GraphQLAdapter {
	return &GraphQLAdapter{server: server, taskGetter: server.beadsClient}
}

// GetActiveTasks returns all currently active tasks
func (a *GraphQLAdapter) GetActiveTasks() map[string]*graphql.TaskInfo {
	a.server.mu.RLock()
	defer a.server.mu.RUnlock()

	tasks := make(map[string]*graphql.TaskInfo)
	for id, execution := range a.server.activeTasks {
		tasks[id] = convertToTaskInfo(execution)
	}
	return tasks
}

// resolveBeadsRoot walks up from dir until it finds a .beads/ directory
// containing an actual beads database (beads.db or issues.jsonl).
// A .beads/ with only a tasks/ subdirectory (agent execution scratch space)
// is not a real database root and is skipped.
// Falls back to the original dir if no database is found.
func resolveBeadsRoot(dir string) string {
	current := dir
	for {
		beadsDir := filepath.Join(current, constants.BeadsDir)
		// Check for a real beads database file
		if _, err := os.Stat(filepath.Join(beadsDir, "beads.db")); err == nil {
			return current
		}
		if _, err := os.Stat(filepath.Join(beadsDir, "issues.jsonl")); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return dir
}

// GetAllTasks returns all tasks (active from memory + beads tasks from all project roots)
func (a *GraphQLAdapter) GetAllTasks() map[string]*graphql.TaskInfo {
	tasks := make(map[string]*graphql.TaskInfo)

	// First, get all active tasks from memory
	a.server.mu.RLock()
	for id, execution := range a.server.activeTasks {
		info := convertToTaskInfo(execution)
		a.applyBeadsStatusOverride(info, execution)
		tasks[id] = info
	}
	a.server.mu.RUnlock()

	// Get all project roots to scan (server root + registered projects).
	// Resolve each to its canonical beads root (walk up to find .beads/tasks)
	// and deduplicate — prevents a subdir (e.g. a2a-agent/) and its parent
	// (ai-pack/) from both being scanned when they share the same .beads/.
	rawRoots := a.server.GetProjectRoots()
	seen := make(map[string]bool)
	var projectRoots []string
	for _, r := range rawRoots {
		canonical := resolveBeadsRoot(r)
		if !seen[canonical] {
			seen[canonical] = true
			projectRoots = append(projectRoots, canonical)
		}
	}

	// Then, get beads tasks from each project using bd list
	a.server.mu.RLock()
	beadsClient := a.server.beadsClient
	a.server.mu.RUnlock()
	monitoring.Logger.Info("fetching_tasks_from_projects", "project_count", len(projectRoots))
	for _, projectRoot := range projectRoots {
		beadsTasks, err := beadsClient.ListAllTasksFromDir(projectRoot)
		if err != nil {
			monitoring.Logger.Warn("failed_to_list_tasks_from_project", "project_root", projectRoot, "error", err.Error())
			continue // Skip if can't list tasks from this project
		}
		monitoring.Logger.Info("listed_tasks_from_project", "project_root", projectRoot, "task_count", len(beadsTasks))

		// Convert beads tasks to TaskInfo
		convertedCount := 0
		skippedDuplicate := 0
		for _, beadsTask := range beadsTasks {
			beadsID := beadsTask.ID

			// Skip if an active task with this Beads ID exists
			// Check for both exact match and timestamped format: {beads-id}-{timestamp}
			alreadyActive := false
			for activeTaskID := range tasks {
				// Check if active task matches this Beads ID exactly or starts with it
				if activeTaskID == beadsID || strings.HasPrefix(activeTaskID, beadsID+"-") {
					alreadyActive = true
					break
				}
			}
			if alreadyActive {
				skippedDuplicate++
				continue
			}

			// Check if there's a recent execution on disk (failed or completed)
			// If so, show that execution instead of the Beads task
			recentExecution := a.findMostRecentExecution(projectRoot, beadsID)
			if recentExecution != nil {
				// Use the execution's actual status (could be "failed", "completed", etc.)
				tasks[recentExecution.TaskID] = recentExecution
				convertedCount++
				continue
			}

			// No recent execution found, show the Beads task as-is
			taskInfo := convertBeadsTaskToTaskInfo(beadsTask, projectRoot)
			tasks[beadsID] = taskInfo
			convertedCount++
		}
		monitoring.Logger.Info("converted_tasks_from_project", "project_root", projectRoot, "converted", convertedCount, "skipped", skippedDuplicate)
	}

	monitoring.Logger.Info("get_all_tasks_complete", "total_tasks", len(tasks))
	return tasks
}

// findMostRecentExecution finds the most recent execution for a Beads task ID in a project root
// Returns nil if no execution found, otherwise returns the TaskInfo for the most recent execution
// Skips executions marked as superseded (from retries)
func (a *GraphQLAdapter) findMostRecentExecution(projectRoot, beadsID string) *graphql.TaskInfo {
	tasksDir := filepath.Join(projectRoot, constants.BeadsDir, "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return nil
	}

	// Build list of all matching executions sorted by time (most recent first)
	type executionEntry struct {
		folderName string
		modTime    time.Time
	}
	var executions []executionEntry

	// Find all folders matching {beads-id}-{timestamp} pattern
	prefix := beadsID + "-"
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		folderName := entry.Name()
		// Check if folder matches pattern: {beads-id}-{timestamp}
		if strings.HasPrefix(folderName, prefix) {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			executions = append(executions, executionEntry{
				folderName: folderName,
				modTime:    info.ModTime(),
			})
		}
	}

	if len(executions) == 0 {
		return nil // No execution found
	}

	// Sort by time (most recent first)
	sort.Slice(executions, func(i, j int) bool {
		return executions[i].modTime.After(executions[j].modTime)
	})

	// Find first non-superseded execution
	for _, exec := range executions {
		// Check if this execution is marked as superseded
		metadataPath := filepath.Join(projectRoot, constants.BeadsDir, "tasks", exec.folderName, constants.MetadataFileName)
		if data, err := os.ReadFile(metadataPath); err == nil {
			var metadata map[string]interface{}
			if json.Unmarshal(data, &metadata) == nil {
				if superseded, ok := metadata["superseded"].(bool); ok && superseded {
					monitoring.Logger.Debug("skipping_superseded_execution",
						"folder", exec.folderName,
						"beads_id", beadsID)
					continue // Skip superseded executions
				}
			}
		}

		// Load the task info from this execution
		taskInfo, err := a.loadTaskFromProject(projectRoot, exec.folderName)
		if err != nil {
			continue // Try next execution if this one fails to load
		}

		return taskInfo
	}

	return nil // All executions were superseded or failed to load
}

// scanProjectTasks scans a single project root for tasks
func (a *GraphQLAdapter) scanProjectTasks(projectRoot string, tasks map[string]*graphql.TaskInfo) {
	tasksDir := filepath.Join(projectRoot, constants.BeadsDir, "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return // Skip if can't read directory
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		taskID := entry.Name()
		// Skip if already in tasks map
		if _, exists := tasks[taskID]; exists {
			continue
		}

		// Load from disk
		taskInfo, err := a.loadTaskFromProject(projectRoot, taskID)
		if err == nil {
			tasks[taskID] = taskInfo
		}
	}
}

// loadTaskFromProject loads a task from a specific project root
func (a *GraphQLAdapter) loadTaskFromProject(projectRoot, taskID string) (*graphql.TaskInfo, error) {
	metadataPath := filepath.Join(projectRoot, constants.BeadsDir, "tasks", taskID, constants.MetadataFileName)
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, err
	}

	var status struct {
		TaskID      string            `json:"task_id"`
		Role        string            `json:"role"`
		Task        string            `json:"task"`
		Description string            `json:"description"`
		Status      string            `json:"status"`
		CreatedAt   time.Time         `json:"spawned_at"`
		UpdatedAt   time.Time         `json:"updated_at"`
		CompletedAt *time.Time        `json:"completed_at,omitempty"`
		Result      string            `json:"result,omitempty"`
		Error       string            `json:"error,omitempty"`
		Metadata    map[string]string `json:"metadata,omitempty"`
		Model       string            `json:"model,omitempty"`
		Provider    string            `json:"provider,omitempty"`
	}

	if err := json.Unmarshal(data, &status); err != nil {
		return nil, err
	}

	// Fall back to parsing timestamp from execution folder name when spawned_at is missing.
	// Folder format: {beads-id}-YYYYMMDD-HHMMSS  e.g. xasm++-qbxv-20260218-084509
	if status.CreatedAt.IsZero() {
		parts := strings.Split(taskID, "-")
		if len(parts) >= 2 {
			lastPart := parts[len(parts)-1]
			secondLastPart := parts[len(parts)-2]
			if len(lastPart) == 6 && len(secondLastPart) == 8 {
				status.CreatedAt, _ = time.Parse("20060102-150405", secondLastPart+"-"+lastPart)
			}
		}
	}
	if status.UpdatedAt.IsZero() {
		status.UpdatedAt = status.CreatedAt
	}

	// Convert to TaskInfo
	// Use description field if task field is empty
	taskDescription := status.Task
	if taskDescription == "" {
		taskDescription = status.Description
	}

	// Use Beads task ID as the primary identifier if available
	// This ensures retry/logs use the task ID rather than timestamped execution folder
	primaryTaskID := status.TaskID
	if beadsID, ok := status.Metadata["beads_task_id"]; ok && beadsID != "" {
		primaryTaskID = beadsID
	}

	taskInfo := &graphql.TaskInfo{
		TaskID:    primaryTaskID,
		Role:      status.Role,
		Task:      taskDescription,
		Status:    status.Status,
		CreatedAt: status.CreatedAt.Format(time.RFC3339),
		UpdatedAt: status.UpdatedAt.Format(time.RFC3339),
		Metadata:  make(map[string]string),
	}

	// Copy metadata
	for k, v := range status.Metadata {
		taskInfo.Metadata[k] = v
	}

	// Promote top-level model/provider into metadata so the GUI can display them
	if status.Model != "" {
		taskInfo.Metadata["model"] = status.Model
	}
	if status.Provider != "" {
		taskInfo.Metadata["provider"] = status.Provider
	}

	// Add project root to metadata if not already there
	if _, exists := taskInfo.Metadata["project_root"]; !exists {
		taskInfo.Metadata["project_root"] = projectRoot
	}

	// Set project root
	taskInfo.ProjectRoot = &projectRoot

	if status.CompletedAt != nil {
		completedAt := status.CompletedAt.Format(time.RFC3339)
		taskInfo.CompletedAt = &completedAt
	}
	if status.Result != "" {
		taskInfo.Result = &status.Result
	}
	if status.Error != "" {
		taskInfo.Error = &status.Error
	}

	// Reconcile stale terminal statuses against live Beads task status.
	// A task closed in Beads is always completed regardless of how the execution ended.
	if taskInfo.Status == constants.StatusCancelled ||
		taskInfo.Status == constants.StatusFailed ||
		taskInfo.Status == "blocked" {
		beadsTaskID := taskInfo.Metadata["beads_task_id"]
		if beadsTaskID == "" {
			// Infer from execution folder name: {beads-id}-YYYYMMDD-HHMMSS
			parts := strings.Split(taskID, "-")
			if len(parts) >= 3 {
				lastPart := parts[len(parts)-1]
				secondLastPart := parts[len(parts)-2]
				if len(lastPart) == 6 && len(secondLastPart) == 8 {
					beadsTaskID = strings.Join(parts[:len(parts)-2], "-")
				}
			}
		}
		if beadsTaskID != "" {
			if beadsTask, err := a.server.beadsClient.GetTaskFromDir(beadsTaskID, projectRoot); err == nil {
				beadsStatus := strings.ToLower(beadsTask.Status)
				if beadsStatus == constants.StatusClosed || beadsStatus == constants.StatusDone {
					monitoring.Logger.Info("reconciling_stale_execution_metadata",
						"task_id", beadsTaskID,
						"old_status", taskInfo.Status,
						"beads_status", beadsStatus,
						"execution_folder", taskID)
					taskInfo.Status = constants.StatusCompleted
					taskInfo.Error = nil
					// Write back to disk to avoid repeating this lookup on every render
					var rawMetadata map[string]interface{}
					if json.Unmarshal(data, &rawMetadata) == nil {
						rawMetadata["status"] = constants.StatusCompleted
						rawMetadata["error"] = nil
						rawMetadata["updated_at"] = time.Now().Format(time.RFC3339)
						rawMetadata["reconciled"] = true
						rawMetadata["reconciled_at"] = time.Now().Format(time.RFC3339)
						if updatedData, merr := json.MarshalIndent(rawMetadata, "", "  "); merr == nil {
							_ = os.WriteFile(metadataPath, updatedData, 0644)
						}
					}
				}
			}
		}
	}

	return taskInfo, nil
}

// applyBeadsStatusOverride updates taskInfo.Status based on the live Beads task status.
// This is the source of truth for tasks that have been closed, cancelled, or completed.
//
// Mapping:
//   - Beads "closed" / "done"               → constants.StatusCompleted
//   - Beads "open" with stale in-progress   → constants.StatusCancelled
//     (task was reset to open after cancel)
//   - anything else                         → leave taskInfo.Status unchanged
func (a *GraphQLAdapter) applyBeadsStatusOverride(taskInfo *graphql.TaskInfo, execution *TaskExecution) {
	beadsTaskID, ok := execution.metadata["beads_task_id"]
	if !ok || beadsTaskID == "" {
		return
	}
	beadsTask, err := a.taskGetter.GetTask(beadsTaskID)
	if err != nil {
		return
	}
	switch beadsTask.Status {
	case constants.StatusClosed, constants.StatusDone:
		taskInfo.Status = constants.StatusCompleted
	case "open":
		// Task was reset to open (e.g. after cancel) — show as cancelled
		if taskInfo.Status == "in_progress" || taskInfo.Status == constants.StatusCancelled {
			taskInfo.Status = constants.StatusCancelled
		}
	}
}

// GetTaskStatus returns status for a specific task
func (a *GraphQLAdapter) GetTaskStatus(taskID string) (*graphql.TaskInfo, error) {
	a.server.mu.RLock()
	execution, exists := a.server.activeTasks[taskID]
	a.server.mu.RUnlock()

	if exists {
		taskInfo := convertToTaskInfo(execution)
		a.applyBeadsStatusOverride(taskInfo, execution)
		return taskInfo, nil
	}

	// Try loading from disk
	status, err := a.server.loadTaskStatusFromDisk(taskID)
	if err != nil {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	// Convert protocol.TaskStatusResponse to graphql.TaskInfo
	taskInfo := &graphql.TaskInfo{
		TaskID:    status.TaskID,
		Role:      status.Role,
		Task:      status.Task,
		Status:    status.Status,
		CreatedAt: status.CreatedAt.Format(time.RFC3339),
		UpdatedAt: status.UpdatedAt.Format(time.RFC3339),
	}

	if status.CompletedAt != nil {
		completedAt := status.CompletedAt.Format(time.RFC3339)
		taskInfo.CompletedAt = &completedAt
	}
	if status.Result != "" {
		taskInfo.Result = &status.Result
	}
	if status.Error != "" {
		taskInfo.Error = &status.Error
	}

	return taskInfo, nil
}

// SpawnAgent spawns a new agent task
func (a *GraphQLAdapter) SpawnAgent(role, task, projectRoot string) (*graphql.TaskInfo, error) {
	response, err := a.server.spawnAgentTask(role, task, projectRoot)
	if err != nil {
		return nil, err
	}

	// Get the task execution
	a.server.mu.RLock()
	execution := a.server.activeTasks[response.TaskID]
	a.server.mu.RUnlock()

	if execution == nil {
		return nil, fmt.Errorf("task spawned but not found: %s", response.TaskID)
	}

	return convertToTaskInfo(execution), nil
}

// CancelTask cancels a running task
func (a *GraphQLAdapter) CancelTask(taskID string) error {
	return a.server.CancelTask(taskID)
}

// GetMetrics returns current system metrics
func (a *GraphQLAdapter) GetMetrics() *graphql.MetricsInfo {
	// Get metrics snapshot
	snapshot := monitoring.GlobalMetrics.GetSnapshot()

	// Count active tasks from server's activeTasks map
	// This is the source of truth for tasks currently executing
	a.server.mu.RLock()
	activeCount := len(a.server.activeTasks)
	a.server.mu.RUnlock()

	// Calculate total tokens
	totalTokens := snapshot.TotalInputTokens + snapshot.TotalOutputTokens

	// Convert turn token data
	recentTurns := make([]monitoring.TurnTokenData, 0)
	if len(snapshot.TurnTokenData) > 0 {
		// Get last 10 turns
		start := 0
		if len(snapshot.TurnTokenData) > 10 {
			start = len(snapshot.TurnTokenData) - 10
		}
		recentTurns = snapshot.TurnTokenData[start:]
	}

	// Convert to GraphQL metrics info
	return &graphql.MetricsInfo{
		TasksSpawned:         int(snapshot.TasksSpawned),
		TasksCompleted:       int(snapshot.TasksCompleted),
		TasksFailed:          int(snapshot.TasksFailed),
		TasksActive:          activeCount,
		AverageDurationMs:    float64(snapshot.AvgDurationMs),
		TotalTokens:          totalTokens,
		InputTokens:          snapshot.TotalInputTokens,
		OutputTokens:         snapshot.TotalOutputTokens,
		APICalls:             snapshot.APICallsTotal,
		APISuccess:           snapshot.APICallsSuccess,
		APIFailed:            snapshot.APICallsFailed,
		AverageTokensPerTask: snapshot.AverageTokensPerTask,
		Uptime:               formatUptime(snapshot.Uptime),
		// New detailed metrics
		TotalTurns:          snapshot.TotalTurns,
		AvgInputPerTurn:     snapshot.AvgInputPerTurn,
		AvgOutputPerTurn:    snapshot.AvgOutputPerTurn,
		RecentTurns:         recentTurns,
		RecentSessions:      snapshot.TaskTokenUsage,
		StreamsOpened:       snapshot.StreamsOpened,
		StreamsClosed:       snapshot.StreamsClosed,
		StreamsActive:       snapshot.StreamsActive,
		HTTPRequestsTotal:   snapshot.HTTPRequestsTotal,
		HTTPErrors:          snapshot.HTTPErrors,
		RateLimitViolations: snapshot.RateLimitViolations,
		ProviderBreakdown:   snapshot.ProviderBreakdown,
	}
}

// formatUptime formats a duration into a human-readable string
func formatUptime(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if hours > 24 {
		days := hours / 24
		hours = hours % 24
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// convertToTaskInfo converts TaskExecution to graphql.TaskInfo
func convertToTaskInfo(execution *TaskExecution) *graphql.TaskInfo {
	// Use Beads task ID as the primary identifier if available
	// This ensures retry/logs use the task ID rather than timestamped execution folder
	taskID := execution.TaskID
	if beadsID, ok := execution.metadata["beads_task_id"]; ok && beadsID != "" {
		taskID = beadsID
	}

	taskInfo := &graphql.TaskInfo{
		TaskID:    taskID,
		Role:      execution.Role,
		Task:      execution.Task,
		Status:    execution.Status,
		CreatedAt: execution.StartTime.Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
		Metadata:  execution.metadata,
	}

	if execution.Status == constants.StatusCompleted || execution.Status == constants.StatusFailed {
		completedAt := time.Now().Format(time.RFC3339)
		taskInfo.CompletedAt = &completedAt
	}

	if execution.Result != "" {
		taskInfo.Result = &execution.Result
	}
	if execution.Error != "" {
		taskInfo.Error = &execution.Error
	}

	if execution.ProjectRoot != "" {
		taskInfo.ProjectRoot = &execution.ProjectRoot
	}

	return taskInfo
}

// convertBeadsTaskToTaskInfo converts beads.Task to graphql.TaskInfo
func convertBeadsTaskToTaskInfo(beadsTask beads.Task, projectRoot string) *graphql.TaskInfo {
	// Map beads status to agent status
	// Beads statuses: "open", "in_progress", "closed", "done"
	// Agent statuses: "open", "queued", "in_progress", "completed", "failed"
	status := "open"
	switch beadsTask.Status {
	case "in_progress":
		status = "in_progress"
	case "closed", "done":
		// A task closed in Beads is always completed, regardless of execution outcome
		status = constants.StatusCompleted
	case "open":
		status = "open" // Keep as "open" to distinguish from queued agents
	}

	taskInfo := &graphql.TaskInfo{
		TaskID:      beadsTask.ID,
		Role:        "beads-task",
		Task:        beadsTask.Title,
		Status:      status,
		CreatedAt:   beadsTask.CreatedAt,
		UpdatedAt:   beadsTask.UpdatedAt,
		Metadata:    make(map[string]string),
		ProjectRoot: &projectRoot,
	}

	// Add Beads status to metadata so GUI can filter closed tasks
	taskInfo.Metadata["beads_status"] = beadsTask.Status

	// Add completion timestamp if available
	if beadsTask.ClosedAt != "" {
		taskInfo.CompletedAt = &beadsTask.ClosedAt
	}

	// Add description as result if available
	if beadsTask.Description != "" && beadsTask.Description != beadsTask.Title {
		taskInfo.Result = &beadsTask.Description
	}

	return taskInfo
}

// CloseTask marks a task as closed so it won't appear in the GUI
// Tasks are stored in each project's .beads/tasks/<taskID>/metadata.json
func (a *GraphQLAdapter) CloseTask(taskID string) error {
	a.server.mu.Lock()
	defer a.server.mu.Unlock()

	// Check if task exists in active tasks - has projectRoot in memory
	execution, exists := a.server.activeTasks[taskID]
	if exists {
		execution.Status = "closed"
		// Also update on disk if projectRoot is known
		if execution.ProjectRoot != "" {
			if err := a.updateTaskMetadataOnDisk(execution.ProjectRoot, taskID, "closed"); err != nil {
				// Log but don't fail - in-memory update succeeded
				monitoring.Logger.Error("failed_to_update_task_on_disk", "error", err.Error())
			}
		}
		return nil
	}

	// If not in active tasks, search through known project roots
	monitoring.Logger.Info("searching_for_task_to_close", "task_id", taskID, "known_projects", len(a.server.projectRoots))

	// First try direct lookup (for new tasks using Beads ID as primary ID)
	for projectRoot := range a.server.projectRoots {
		metadataPath := filepath.Join(projectRoot, ".beads", "tasks", taskID, "metadata.json")
		if _, err := os.Stat(metadataPath); err == nil {
			// Found the task, update it
			monitoring.Logger.Info("found_task_direct", "project_root", projectRoot, "task_id", taskID)
			return a.updateTaskMetadataOnDisk(projectRoot, taskID, "closed")
		}
	}

	// If not found, search all task directories for one with matching beads_task_id
	// This handles legacy tasks created before standardizing on Beads IDs
	monitoring.Logger.Info("searching_task_directories_for_beads_id", "task_id", taskID)
	for projectRoot := range a.server.projectRoots {
		tasksDir := filepath.Join(projectRoot, ".beads", "tasks")
		entries, err := os.ReadDir(tasksDir)
		if err != nil {
			monitoring.Logger.Info("skipping_project", "project_root", projectRoot, "error", err.Error())
			continue // Project might not have a .beads/tasks directory yet
		}

		monitoring.Logger.Info("scanning_tasks_in_project", "project_root", projectRoot, "task_count", len(entries))
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			// Check if directory name matches the task ID (taskID is Beads ID)
			if entry.Name() == taskID {
				return a.updateTaskMetadataOnDisk(projectRoot, entry.Name(), "closed")
			}
		}
	}

	// If still not found in known projects, try using bd CLI to find the task
	// Beads can search across the filesystem to find tasks
	monitoring.Logger.Info("task_not_in_known_projects_trying_bd_show", "task_id", taskID)

	// Use bd show to verify the task exists and get its project location
	cmd := exec.Command("bd", "show", taskID)
	output, err := cmd.CombinedOutput()
	if err == nil && len(output) > 0 {
		// Task exists in Beads, even if not in our registry
		// Mark as closed by updating metadata through bd or directly
		monitoring.Logger.Info("task_found_via_bd_not_updating_agent_server", "task_id", taskID)
		// Return success since Beads knows about it and close succeeded
		return nil
	}

	return fmt.Errorf("task not found in any known project: %s", taskID)
}

// DeleteTask permanently removes a task via `bd delete` and cleans up local state.
func (a *GraphQLAdapter) DeleteTask(taskID string) error {
	// Find the project root for this task (needed to set working directory for bd)
	projectRoot := a.findProjectRootForTask(taskID)

	// Run bd delete — this removes the task from the Beads database entirely.
	// Ignore failure: the Beads task may already be deleted while execution
	// artifacts remain on disk (orphaned folders after a server crash, etc.).
	cmd := exec.Command("bd", "delete", taskID)
	if projectRoot != "" {
		cmd.Dir = projectRoot
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		monitoring.Logger.Warn("bd_delete_failed",
			"task_id", taskID,
			"error", err.Error(),
			"output", string(out))
	} else {
		monitoring.Logger.Info("bd_delete_succeeded", "task_id", taskID)
	}

	// Remove from in-memory active tasks (both short Beads ID and any timestamped executions)
	a.server.mu.Lock()
	delete(a.server.activeTasks, taskID)
	for id := range a.server.activeTasks {
		if strings.HasPrefix(id, taskID+"-") {
			delete(a.server.activeTasks, id)
		}
	}
	a.server.mu.Unlock()

	// Remove all execution folders on disk: both exact match and timestamped variants
	// (e.g. "xasm++-sduc" and "xasm++-sduc-20260218-170420")
	if projectRoot != "" {
		tasksDir := filepath.Join(projectRoot, ".beads", "tasks")
		prefix := taskID + "-"
		entries, err := os.ReadDir(tasksDir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				name := entry.Name()
				if name == taskID || strings.HasPrefix(name, prefix) {
					dir := filepath.Join(tasksDir, name)
					if err := os.RemoveAll(dir); err != nil {
						monitoring.Logger.Warn("failed_to_remove_task_dir",
							"task_id", taskID, "path", dir, "error", err.Error())
					} else {
						monitoring.Logger.Info("removed_task_dir", "path", dir)
					}
				}
			}
		}
	}

	return nil
}

// findProjectRootForTask returns the project root for a task, or "" if not found.
func (a *GraphQLAdapter) findProjectRootForTask(taskID string) string {
	a.server.mu.RLock()
	if execution, exists := a.server.activeTasks[taskID]; exists {
		root := execution.ProjectRoot
		a.server.mu.RUnlock()
		return root
	}
	// Also check timestamped execution IDs
	prefix := taskID + "-"
	for id, execution := range a.server.activeTasks {
		if strings.HasPrefix(id, prefix) {
			root := execution.ProjectRoot
			a.server.mu.RUnlock()
			return root
		}
	}
	a.server.mu.RUnlock()

	// Scan registered project roots for execution artifacts on disk.
	// (bd show is global and cannot be used to determine the correct project root.)
	for projectRoot := range a.server.projectRoots {
		tasksDir := filepath.Join(projectRoot, ".beads", "tasks")
		entries, err := os.ReadDir(tasksDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			if name == taskID || strings.HasPrefix(name, taskID+"-") {
				return projectRoot
			}
		}
	}
	return ""
}

// GetProjectCostsData returns cost data for all projects
func (a *GraphQLAdapter) GetProjectCostsData() ([]map[string]interface{}, error) {
	return a.server.GetProjectCostsData()
}

// GetProjectRoots returns all registered project roots
func (a *GraphQLAdapter) GetProjectRoots() []string {
	return a.server.GetProjectRoots()
}

// updateTaskMetadataOnDisk updates a task's status in its project's .beads directory
func (a *GraphQLAdapter) updateTaskMetadataOnDisk(projectRoot, taskID, status string) error {
	metadataPath := filepath.Join(projectRoot, ".beads", "tasks", taskID, "metadata.json")

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to read task metadata: %w", err)
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		return fmt.Errorf("failed to parse task metadata: %w", err)
	}

	// Update status to closed
	metadata["status"] = status
	metadata["updated_at"] = time.Now().Format(time.RFC3339)

	// Write back to disk
	updatedData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal updated metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write updated metadata: %w", err)
	}

	return nil
}
