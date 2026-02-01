package server

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

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

// GetAllTasks returns all tasks (active + completed/failed from disk)
func (a *GraphQLAdapter) GetAllTasks() map[string]*graphql.TaskInfo {
	tasks := make(map[string]*graphql.TaskInfo)

	// First, get all active tasks
	a.server.mu.RLock()
	for id, execution := range a.server.activeTasks {
		tasks[id] = convertToTaskInfo(execution)
	}
	a.server.mu.RUnlock()

	// Then, scan disk for completed/failed tasks
	tasksDir := filepath.Join(a.server.rootDir, BeadsDir, "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return tasks // Return active tasks if can't read directory
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		taskID := entry.Name()
		// Skip if already in active tasks
		if _, exists := tasks[taskID]; exists {
			continue
		}

		// Load from disk
		taskInfo, err := a.GetTaskStatus(taskID)
		if err == nil {
			tasks[taskID] = taskInfo
		}
	}

	return tasks
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
		MemoryUsageMB:        0.0, // Removed from display
		Goroutines:           0,   // Removed from display
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

	return taskInfo
}
