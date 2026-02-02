package server

import (
	"encoding/json"
	"fmt"
	"os"
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
	for _, projectRoot := range projectRoots {
		beadsTasks, err := beadsClient.ListAllTasksFromDir(projectRoot)
		if err != nil {
			continue // Skip if can't list tasks from this project
		}

		// Convert beads tasks to TaskInfo
		for _, beadsTask := range beadsTasks {
			taskID := beadsTask.ID
			// Skip if already in tasks map (active tasks take precedence)
			if _, exists := tasks[taskID]; exists {
				continue
			}

			// Convert beads.Task to graphql.TaskInfo
			taskInfo := convertBeadsTaskToTaskInfo(beadsTask, projectRoot)
			tasks[taskID] = taskInfo
		}
	}

	return tasks
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
	taskInfo := &graphql.TaskInfo{
		TaskID:    status.TaskID,
		Role:      status.Role,
		Task:      status.Task,
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
	if beadsID, ok := status.Metadata["beads_task_id"]; ok {
		taskInfo.BeadsTaskID = &beadsID
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

	// Count active tasks
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

	if beadsID, ok := execution.metadata["beads_task_id"]; ok {
		taskInfo.BeadsTaskID = &beadsID
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
		CreatedAt:   time.Now().Format(time.RFC3339), // Beads doesn't expose creation time in Task struct
		UpdatedAt:   time.Now().Format(time.RFC3339),
		Metadata:    make(map[string]string),
		BeadsTaskID: &beadsTask.ID,
		ProjectRoot: &projectRoot,
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

		// Check metadata for matching beads_task_id
		metadataPath := filepath.Join(tasksDir, entry.Name(), "00-metadata.json")
		data, err := os.ReadFile(metadataPath)
		if err != nil {
			continue
		}

		var metadata struct {
			Metadata map[string]string `json:"metadata"`
		}
		if err := json.Unmarshal(data, &metadata); err != nil {
			continue
		}

		if metadata.Metadata["beads_task_id"] != beadsTaskID {
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
