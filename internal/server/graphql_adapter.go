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

	"github.com/cortexa-llc/ai-pack/internal/constants"
	"github.com/cortexa-llc/ai-pack/internal/graphql"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/internal/taskdb"
)

// GraphQLAdapter adapts AgentServer to implement graphql.ServerInterface
// This allows GraphQL resolvers to access server functionality without circular dependencies
type GraphQLAdapter struct {
	server *AgentServer
}

// NewGraphQLAdapter creates a new GraphQL adapter for the agent server
func NewGraphQLAdapter(server *AgentServer) *GraphQLAdapter {
	return &GraphQLAdapter{server: server}
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

// resolveTaskRoot walks up from dir until it finds a .beads/ directory
// containing an actual task database (beads.db or issues.jsonl).
// A .beads/ with only a tasks/ subdirectory (agent execution scratch space)
// is not a real database root and is skipped.
// Falls back to the original dir if no database is found.
func resolveTaskRoot(dir string) string {
	current := dir
	for {
		taskRootDir := filepath.Join(current, constants.TaskRootDir)
		// Check for a real task database file
		if _, err := os.Stat(filepath.Join(taskRootDir, "beads.db")); err == nil {
			return current
		}
		if _, err := os.Stat(filepath.Join(taskRootDir, "issues.jsonl")); err == nil {
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

// GetAllTasks returns all tasks (active from memory + taskDB + disk scan for orphans)
func (a *GraphQLAdapter) GetAllTasks() map[string]*graphql.TaskInfo {
	tasks := make(map[string]*graphql.TaskInfo)

	// First, get all active tasks from memory
	a.server.mu.RLock()
	for id, execution := range a.server.activeTasks {
		info := convertToTaskInfo(execution)
		tasks[id] = info
	}
	a.server.mu.RUnlock()

	// Second, get all tasks from taskDB
	if a.server.taskDB != nil {
		dbTasks, err := a.server.taskDB.ListTasks(taskdb.TaskFilter{
			Limit: 10000, // Get all tasks
		})
		if err != nil {
			monitoring.Logger.Warn("failed_to_list_tasks_from_taskdb", "error", err.Error())
		} else {
			monitoring.Logger.Info("loaded_tasks_from_taskdb", "count", len(dbTasks))
			for _, dbTask := range dbTasks {
				// Use short task ID as the key (like "ai-pack-abc123")
				taskKey := taskdb.ExtractShortID(dbTask.ID)

				// Skip if already present from activeTasks (use most recent status)
				if _, exists := tasks[taskKey]; exists {
					continue
				}

				// Also check for timestamped variants
				alreadyActive := false
				for activeID := range tasks {
					if strings.HasPrefix(activeID, taskKey+"-") {
						alreadyActive = true
						break
					}
				}
				if alreadyActive {
					continue
				}

				// Convert taskDB task to TaskInfo
				taskInfo := &graphql.TaskInfo{
					TaskID:    taskKey,
					Role:      dbTask.Role,
					Task:      dbTask.TaskDescription,
					Status:    dbTask.Status,
					CreatedAt: dbTask.CreatedAt.Format(time.RFC3339),
					UpdatedAt: dbTask.UpdatedAt.Format(time.RFC3339),
					Metadata:  make(map[string]string),
				}

				if dbTask.ProjectRoot != "" {
					taskInfo.ProjectRoot = &dbTask.ProjectRoot
				}

				if dbTask.CompletedAt != nil && !dbTask.CompletedAt.IsZero() {
					completedAt := dbTask.CompletedAt.Format(time.RFC3339)
					taskInfo.CompletedAt = &completedAt
				}

				if dbTask.Result != "" {
					taskInfo.Result = &dbTask.Result
				}

				if dbTask.Error != "" {
					taskInfo.Error = &dbTask.Error
				}

				// Parse metadata JSON into map
				if dbTask.Metadata != "" {
					var metaMap map[string]string
					if json.Unmarshal([]byte(dbTask.Metadata), &metaMap) == nil {
						taskInfo.Metadata = metaMap
					}
				}

				tasks[taskKey] = taskInfo
			}
		}
	}

	// Get all project roots to scan (server root + registered projects).
	// Resolve each to its canonical beads root (walk up to find .beads/tasks)
	// and deduplicate — prevents a subdir (e.g. a2a-agent/) and its parent
	// (ai-pack/) from both being scanned when they share the same .beads/.
	rawRoots := a.server.GetProjectRoots()
	seen := make(map[string]bool)
	var projectRoots []string
	for _, r := range rawRoots {
		canonical := resolveTaskRoot(r)
		if !seen[canonical] {
			seen[canonical] = true
			projectRoots = append(projectRoots, canonical)
		}
	}

	// Finally, scan project directories for execution artifacts not in taskDB
	// This catches orphaned executions or tasks from before migration
	monitoring.Logger.Info("scanning_project_tasks", "project_count", len(projectRoots))
	for _, projectRoot := range projectRoots {
		a.scanProjectTasks(projectRoot, tasks)
	}

	monitoring.Logger.Info("get_all_tasks_complete", "total_tasks", len(tasks))
	return tasks
}

// parseFolderTimestamp extracts and parses the timestamp suffix from an execution folder name.
// Folder format: {short-task-id}-YYYYMMDD-HHMMSS (e.g. "ai-pack-x008-20260228-130230").
// The prefix (e.g. "ai-pack-x008-") must be supplied so we know where the timestamp starts.
// Returns the zero time if parsing fails so callers can fall back to other sorting strategies.
func parseFolderTimestamp(folderName, prefix string) time.Time {
	suffix := strings.TrimPrefix(folderName, prefix)
	// Timestamp format produced by time.Now().Format("20060102-150405")
	t, err := time.Parse("20060102-150405", suffix)
	if err != nil {
		return time.Time{}
	}
	return t
}

// findMostRecentExecution finds the most recent execution for a task ID in a project root
// Returns nil if no execution found, otherwise returns the TaskInfo for the most recent execution
// Skips executions marked as superseded (from retries)
func (a *GraphQLAdapter) findMostRecentExecution(projectRoot, shortTaskID string) *graphql.TaskInfo {
	tasksDir := filepath.Join(projectRoot, constants.TaskRootDir, "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return nil
	}

	// Build list of all matching executions sorted by time (most recent first)
	type executionEntry struct {
		folderName string
		folderTime time.Time // parsed from folder name, not filesystem mtime
	}
	var executions []executionEntry

	// Find all folders matching {short-task-id}-{timestamp} pattern
	prefix := shortTaskID + "-"
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		folderName := entry.Name()
		// Check if folder matches pattern: {short-task-id}-{timestamp}
		if strings.HasPrefix(folderName, prefix) {
			executions = append(executions, executionEntry{
				folderName: folderName,
				// Sort by the timestamp embedded in the folder name, not by filesystem
				// mtime. mtime is unreliable: marking an old execution as superseded
				// updates its mtime AFTER the new execution folder was created, causing
				// the old (failed) folder to appear newer than the running one.
				folderTime: parseFolderTimestamp(folderName, prefix),
			})
		}
	}

	if len(executions) == 0 {
		return nil // No execution found
	}

	// Sort by folder name timestamp (most recent first)
	sort.Slice(executions, func(i, j int) bool {
		return executions[i].folderTime.After(executions[j].folderTime)
	})

	// Find first non-superseded execution
	for _, exec := range executions {
		// Check if this execution is marked as superseded
		metadataPath := filepath.Join(projectRoot, constants.TaskRootDir, "tasks", exec.folderName, constants.MetadataFileName)
		if data, err := os.ReadFile(metadataPath); err == nil {
			var metadata map[string]interface{}
			if json.Unmarshal(data, &metadata) == nil {
				if superseded, ok := metadata["superseded"].(bool); ok && superseded {
					monitoring.Logger.Debug("skipping_superseded_execution",
						"folder", exec.folderName,
						"beads_id", shortTaskID)
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

// inferShortTaskID extracts the task ID from an execution folder name.
// Folder format: {short-task-id}-YYYYMMDD-HHMMSS (e.g. "xasm++-01qi-20260210-120445").
// Returns the folder name unchanged when it doesn't match the expected format.
func inferShortTaskID(folderName string) string {
	parts := strings.Split(folderName, "-")
	if len(parts) >= 3 && len(parts[len(parts)-1]) == 6 && len(parts[len(parts)-2]) == 8 {
		return strings.Join(parts[:len(parts)-2], "-")
	}
	return folderName
}

// scanProjectTasks scans a single project root for tasks.
// Groups execution folders by inferred task ID and picks the most recent
// non-superseded execution for each task, mirroring findMostRecentExecution logic.
func (a *GraphQLAdapter) scanProjectTasks(projectRoot string, tasks map[string]*graphql.TaskInfo) {
	tasksDir := filepath.Join(projectRoot, constants.TaskRootDir, "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return // Skip if can't read directory
	}

	type execEntry struct {
		folderName string
		folderTime time.Time
	}

	// Group execution folders by inferred task ID
	byShortTaskID := make(map[string][]execEntry)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		folderName := entry.Name()
		shortTaskID := inferShortTaskID(folderName)
		prefix := shortTaskID + "-"
		byShortTaskID[shortTaskID] = append(byShortTaskID[shortTaskID], execEntry{
			folderName: folderName,
			folderTime: parseFolderTimestamp(folderName, prefix),
		})
	}

	for shortTaskID, executions := range byShortTaskID {
		// Skip if already present (active task or main-project bd list result)
		if _, exists := tasks[shortTaskID]; exists {
			continue
		}

		// Sort most-recent first by embedded timestamp
		sort.Slice(executions, func(i, j int) bool {
			return executions[i].folderTime.After(executions[j].folderTime)
		})

		// Pick the most recent non-superseded execution
		for _, exec := range executions {
			metadataPath := filepath.Join(projectRoot, constants.TaskRootDir, "tasks", exec.folderName, constants.MetadataFileName)
			if data, readErr := os.ReadFile(metadataPath); readErr == nil {
				var meta map[string]interface{}
				if json.Unmarshal(data, &meta) == nil {
					if superseded, ok := meta["superseded"].(bool); ok && superseded {
						continue // skip superseded retries
					}
				}
			}

			taskInfo, loadErr := a.loadTaskFromProject(projectRoot, exec.folderName)
			if loadErr != nil {
				continue
			}
			// Ensure the task is keyed by task ID, not folder name
			if taskInfo.TaskID == "" {
				taskInfo.TaskID = shortTaskID
			}
			tasks[taskInfo.TaskID] = taskInfo
			break
		}
	}
}

// loadTaskFromProject loads a task from a specific project root
func (a *GraphQLAdapter) loadTaskFromProject(projectRoot, taskID string) (*graphql.TaskInfo, error) {
	metadataPath := filepath.Join(projectRoot, constants.TaskRootDir, "tasks", taskID, constants.MetadataFileName)
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
	// Folder format: {short-task-id}-YYYYMMDD-HHMMSS  e.g. xasm++-qbxv-20260218-084509
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

	// Use task ID as the primary identifier if available
	// This ensures retry/logs use the task ID rather than timestamped execution folder
	primaryTaskID := status.TaskID
	if shortTaskID, ok := status.Metadata["task_id"]; ok && shortTaskID != "" {
		primaryTaskID = shortTaskID
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

	// Reconcile stale terminal statuses against live task status.
	// A task closed in Beads is always completed regardless of how the execution ended.
	if taskInfo.Status == constants.StatusCancelled ||
		taskInfo.Status == constants.StatusFailed ||
		taskInfo.Status == "blocked" {
		taskID := taskInfo.Metadata["task_id"]
		if taskID == "" {
			// Infer from execution folder name: {short-task-id}-YYYYMMDD-HHMMSS
			parts := strings.Split(taskID, "-")
			if len(parts) >= 3 {
				lastPart := parts[len(parts)-1]
				secondLastPart := parts[len(parts)-2]
				if len(lastPart) == 6 && len(secondLastPart) == 8 {
					taskID = strings.Join(parts[:len(parts)-2], "-")
				}
			}
		}
		// Status reconciliation is no longer needed - taskDB is the source of truth
	}

	return taskInfo, nil
}

// GetTaskStatus returns status for a specific task
func (a *GraphQLAdapter) GetTaskStatus(taskID string) (*graphql.TaskInfo, error) {
	// Search active tasks: match by exact execution ID OR by task ID prefix.
	// convertToTaskInfo rewrites TaskID to the short task ID, so the UI often
	// calls back with the short ID (e.g. "ai-pack-x008") while activeTasks is
	// keyed by the full timestamped ID (e.g. "ai-pack-x008-20260228-130230").
	a.server.mu.RLock()
	execution, exists := a.server.activeTasks[taskID]
	if !exists {
		// Try prefix match: active key starts with "<shortTaskID>-"
		prefix := taskID + "-"
		for activeID, exec := range a.server.activeTasks {
			if strings.HasPrefix(activeID, prefix) {
				execution = exec
				exists = true
				break
			}
		}
	}
	a.server.mu.RUnlock()

	if exists {
		taskInfo := convertToTaskInfo(execution)
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
	// Use task ID as the primary identifier if available
	// This ensures retry/logs use the task ID rather than timestamped execution folder
	taskID := execution.TaskID
	if shortTaskID, ok := execution.metadata["task_id"]; ok && shortTaskID != "" {
		taskID = shortTaskID
	}

	status := execution.Status
	// "closed" is the internal marker set by CloseTask when cancelling; surface it
	// as "cancelled" so the GUI swimlane routes it correctly.
	if status == constants.StatusClosed {
		status = constants.StatusCancelled
	}

	taskInfo := &graphql.TaskInfo{
		TaskID:    taskID,
		Role:      execution.Role,
		Task:      execution.Task,
		Status:    status,
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

	// First try direct lookup (for new tasks using task ID as primary ID)
	for projectRoot := range a.server.projectRoots {
		metadataPath := filepath.Join(projectRoot, ".beads", "tasks", taskID, "metadata.json")
		if _, err := os.Stat(metadataPath); err == nil {
			// Found the task, update it
			monitoring.Logger.Info("found_task_direct", "project_root", projectRoot, "task_id", taskID)
			return a.updateTaskMetadataOnDisk(projectRoot, taskID, "closed")
		}
	}

	// If not found, search all task directories for one with matching beads_task_id
	// This handles legacy tasks created before standardizing on task IDs
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

			// Check if directory name matches the task ID (taskID is task ID)
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
	// Ignore failure: the task may already be deleted while execution
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

	// Remove from in-memory active tasks (both short task ID and any timestamped executions)
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
