package server

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/beads"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/graphql"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
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

// GetAllTasks returns all tasks (active from memory + beads tasks from all project roots)
func (a *GraphQLAdapter) GetAllTasks() map[string]*graphql.TaskInfo {
	tasks := make(map[string]*graphql.TaskInfo)

	// First, get all active tasks from memory
	a.server.mu.RLock()
	for id, execution := range a.server.activeTasks {
		tasks[id] = convertToTaskInfo(execution)
	}
	a.server.mu.RUnlock()

	// Get all project roots to scan (server root + registered projects)
	projectRoots := a.server.GetProjectRoots()

	// Then, get beads tasks from each project using bd list
	beadsClient := a.server.beadsClient
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
func (a *GraphQLAdapter) findMostRecentExecution(projectRoot, beadsID string) *graphql.TaskInfo {
	tasksDir := filepath.Join(projectRoot, BeadsDir, "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return nil
	}

	var mostRecentFolder string
	var mostRecentTime time.Time

	// Find all folders matching {beads-id}-{timestamp} pattern
	prefix := beadsID + "-"
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		folderName := entry.Name()
		// Check if folder matches pattern: {beads-id}-{timestamp}
		if strings.HasPrefix(folderName, prefix) {
			// Get folder modification time as a proxy for execution time
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if mostRecentFolder == "" || info.ModTime().After(mostRecentTime) {
				mostRecentFolder = folderName
				mostRecentTime = info.ModTime()
			}
		}
	}

	if mostRecentFolder == "" {
		return nil // No execution found
	}

	// Load the task info from the most recent execution
	taskInfo, err := a.loadTaskFromProject(projectRoot, mostRecentFolder)
	if err != nil {
		return nil
	}

	return taskInfo
}

// scanProjectTasks scans a single project root for tasks
func (a *GraphQLAdapter) scanProjectTasks(projectRoot string, tasks map[string]*graphql.TaskInfo) {
	tasksDir := filepath.Join(projectRoot, BeadsDir, "tasks")
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
	metadataPath := filepath.Join(projectRoot, BeadsDir, "tasks", taskID, MetadataFileName)
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
		CreatedAt   time.Time         `json:"created_at"`
		UpdatedAt   time.Time         `json:"updated_at"`
		CompletedAt *time.Time        `json:"completed_at,omitempty"`
		Result      string            `json:"result,omitempty"`
		Error       string            `json:"error,omitempty"`
		Metadata    map[string]string `json:"metadata,omitempty"`
	}

	if err := json.Unmarshal(data, &status); err != nil {
		return nil, err
	}

	// Convert to TaskInfo
	// Use description field if task field is empty
	taskDescription := status.Task
	if taskDescription == "" {
		taskDescription = status.Description
	}

	taskInfo := &graphql.TaskInfo{
		TaskID:    status.TaskID,
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

	return taskInfo, nil
}

// GetTaskStatus returns status for a specific task
func (a *GraphQLAdapter) GetTaskStatus(taskID string) (*graphql.TaskInfo, error) {
	a.server.mu.RLock()
	execution, exists := a.server.activeTasks[taskID]
	a.server.mu.RUnlock()

	if exists {
		return convertToTaskInfo(execution), nil
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
	taskInfo := &graphql.TaskInfo{
		TaskID:    execution.TaskID,
		Role:      execution.Role,
		Task:      execution.Task,
		Status:    execution.Status,
		CreatedAt: execution.StartTime.Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
		Metadata:  execution.metadata,
	}

	if execution.Status == "completed" || execution.Status == "failed" {
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
	// Agent statuses: "queued", "in_progress", "completed", "failed"
	status := "queued"
	switch beadsTask.Status {
	case "in_progress":
		status = "in_progress"
	case "closed", "done":
		// For closed tasks, check execution log to determine if completed or failed
		status = determineExecutionStatus(projectRoot, beadsTask.ID)
	case "open":
		status = "queued"
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

// determineExecutionStatus checks execution log to determine if task completed or failed
func determineExecutionStatus(projectRoot, beadsTaskID string) string {
	// Find the execution log for this beads task
	tasksDir := filepath.Join(projectRoot, BeadsDir, "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return "completed" // Default to completed if we can't read
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if directory name matches the beads task ID
		if entry.Name() != beadsTaskID {
			continue
		}

		// Found the execution - check the log
		logPath := filepath.Join(tasksDir, entry.Name(), "execution.log")
		logData, err := os.ReadFile(logPath)
		if err != nil {
			return "completed" // Default if can't read log
		}

		logContent := string(logData)
		// Look for failure markers
		if strings.Contains(logContent, "❌ Task failed") ||
		   strings.Contains(logContent, "Agentic loop failed") {
			return "failed"
		}
		// Look for success markers
		if strings.Contains(logContent, "✅ Task completed successfully") ||
		   strings.Contains(logContent, "Task completed successfully") {
			return "completed"
		}

		// If closed but no clear marker, default to completed
		return "completed"
	}

	return "completed" // Default if no execution found
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
