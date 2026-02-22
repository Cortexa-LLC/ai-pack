package graphql

import (
	"context"
	"testing"
	"time"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockServerInterface implements ServerInterface for testing
type MockServerInterface struct {
	tasks   map[string]*TaskInfo
	metrics *MetricsInfo
}

func (m *MockServerInterface) GetActiveTasks() map[string]*TaskInfo {
	return m.tasks
}

func (m *MockServerInterface) GetTaskStatus(taskID string) (*TaskInfo, error) {
	if task, ok := m.tasks[taskID]; ok {
		return task, nil
	}
	return nil, assert.AnError
}

func (m *MockServerInterface) SpawnAgent(role, task, projectRoot string) (*TaskInfo, error) {
	taskID := "test-task-1"
	now := time.Now().Format(time.RFC3339)
	taskInfo := &TaskInfo{
		TaskID:    taskID,
		Role:      role,
		Task:      task,
		Status:    "in_progress",
		Progress:  0.1,
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  map[string]string{},
	}
	m.tasks[taskID] = taskInfo
	return taskInfo, nil
}

func (m *MockServerInterface) GetAllTasks() map[string]*TaskInfo {
	return m.tasks
}

func (m *MockServerInterface) CancelTask(taskID string) error   { return nil }
func (m *MockServerInterface) CloseTask(taskID string) error    { return nil }
func (m *MockServerInterface) DeleteTask(taskID string) error   { return nil }
func (m *MockServerInterface) GetMetrics() *MetricsInfo         { return m.metrics }
func (m *MockServerInterface) GetProjectCostsData() ([]map[string]interface{}, error) {
	return nil, nil
}

func TestHealthQuery(t *testing.T) {
	// Setup
	mockServer := &MockServerInterface{
		tasks:   make(map[string]*TaskInfo),
		metrics: &MetricsInfo{},
	}
	// Use the global metrics instance
	monitoring.InitMetrics()
	resolver := NewResolver(mockServer, monitoring.GlobalMetrics)
	ctx := context.Background()

	// Execute
	health, err := resolver.Query().Health(ctx)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "healthy", health.Status)
	assert.Equal(t, "2.1.0", health.Version)
	assert.Equal(t, "a2a-agent", health.Server)
	assert.NotNil(t, health.Features)
	assert.True(t, health.Features.A2aProtocol)
	assert.True(t, health.Features.Monitoring)
	assert.True(t, health.Features.ParallelExecution)
	assert.True(t, health.Features.SseStreaming)
}

func TestSpawnAgentMutation(t *testing.T) {
	// Setup
	mockServer := &MockServerInterface{
		tasks:   make(map[string]*TaskInfo),
		metrics: &MetricsInfo{},
	}
	monitoring.InitMetrics()
	resolver := NewResolver(mockServer, monitoring.GlobalMetrics)
	ctx := context.Background()

	// Execute
	result, err := resolver.Mutation().SpawnAgent(ctx, "engineer", "task-123", nil)

	// Assert
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "test-task-1", result.TaskID)
	assert.NotNil(t, result.Message)
}

func TestTasksQuery(t *testing.T) {
	// Setup
	now := time.Now().Format(time.RFC3339)
	mockServer := &MockServerInterface{
		tasks: map[string]*TaskInfo{
			"task-1": {
				TaskID:    "task-1",
				Role:      "engineer",
				Task:      "Test task",
				Status:    "in_progress",
				Progress:  0.5,
				CreatedAt: now,
				UpdatedAt: now,
				Metadata:  map[string]string{},
			},
			"task-2": {
				TaskID:    "task-2",
				Role:      "tester",
				Task:      "Test task 2",
				Status:    "completed",
				Progress:  1.0,
				CreatedAt: now,
				UpdatedAt: now,
				Metadata:  map[string]string{},
			},
		},
		metrics: &MetricsInfo{},
	}
	monitoring.InitMetrics()
	resolver := NewResolver(mockServer, monitoring.GlobalMetrics)
	ctx := context.Background()

	// Execute - get all tasks
	tasks, err := resolver.Query().Tasks(ctx, nil)

	// Assert
	require.NoError(t, err)
	assert.Len(t, tasks, 2)

	// Execute - filter by status
	status := TaskStatusInProgress
	filteredTasks, err := resolver.Query().Tasks(ctx, &status)

	// Assert
	require.NoError(t, err)
	assert.Len(t, filteredTasks, 1)
	assert.Equal(t, "task-1", filteredTasks[0].TaskID)
}

func TestTaskQuery(t *testing.T) {
	// Setup
	now := time.Now().Format(time.RFC3339)
	mockServer := &MockServerInterface{
		tasks: map[string]*TaskInfo{
			"task-1": {
				TaskID:    "task-1",
				Role:      "engineer",
				Task:      "Test task",
				Status:    "in_progress",
				Progress:  0.5,
				CreatedAt: now,
				UpdatedAt: now,
				Metadata:  map[string]string{},
			},
		},
		metrics: &MetricsInfo{},
	}
	monitoring.InitMetrics()
	resolver := NewResolver(mockServer, monitoring.GlobalMetrics)
	ctx := context.Background()

	// Execute
	task, err := resolver.Query().Task(ctx, "task-1")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "task-1", task.TaskID)
	assert.Equal(t, "engineer", task.Role)
	assert.Equal(t, TaskStatusInProgress, task.Status)
}

func TestMetricsQuery(t *testing.T) {
	// Setup
	mockServer := &MockServerInterface{
		tasks: make(map[string]*TaskInfo),
		metrics: &MetricsInfo{
			TasksSpawned:    10,
			TasksCompleted:  8,
			TasksFailed:     2,
			TasksActive:     3,
			AverageDurationMs: 1500.0,
			TotalTokens:     10000,
			InputTokens:     6000,
			OutputTokens:    4000,
			APICalls:        50,
			APISuccess:      48,
			APIFailed:       2,
		},
	}
	monitoring.InitMetrics()
	resolver := NewResolver(mockServer, monitoring.GlobalMetrics)
	ctx := context.Background()

	// Execute
	metrics, err := resolver.Query().Metrics(ctx)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 10, metrics.TasksSpawned)
	assert.Equal(t, 8, metrics.TasksCompleted)
	assert.Equal(t, 2, metrics.TasksFailed)
	assert.Equal(t, 3, metrics.TasksInProgress)
	assert.Equal(t, 1500.0, metrics.AvgDurationMs)
	assert.Equal(t, 6000, metrics.TotalInputTokens)
	assert.Equal(t, 4000, metrics.TotalOutputTokens)
	assert.Equal(t, 50, metrics.APICallsTotal)
	assert.Equal(t, 48, metrics.APICallsSuccess)
	assert.Equal(t, 2, metrics.APICallsFailed)
}

func TestPerformanceQuery(t *testing.T) {
	// Setup
	mockServer := &MockServerInterface{
		tasks:   make(map[string]*TaskInfo),
		metrics: &MetricsInfo{},
	}
	monitoring.InitMetrics()
	
	// Add some test data to the monitor
	monitoring.GlobalMetrics.RecordTokenUsage("task-1", 1000, 500, 3)
	monitoring.GlobalMetrics.RecordTurnTokens("task-1", 1, 300, 150, 500)
	
	resolver := NewResolver(mockServer, monitoring.GlobalMetrics)
	ctx := context.Background()

	// Execute
	performance, err := resolver.Query().Performance(ctx)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, performance)
	assert.NotNil(t, performance.TaskTokenUsage)
	assert.NotNil(t, performance.RecentTurns)
}

func TestConvertTaskInfoToGraphQL(t *testing.T) {
	// Setup
	now := time.Now().Format(time.RFC3339)
	errorMsg := "Test error"
	taskInfo := &TaskInfo{
		TaskID:      "task-1",
		Role:        "engineer",
		Task:        "Test task",
		Status:      "completed",
		Progress:    1.0,
		CreatedAt:   now,
		UpdatedAt:   now,
		Error:       &errorMsg,
		Metadata: map[string]string{
			"project_root": "/test/path",
		},
	}

	// Execute
	graphqlTask := convertTaskInfoToGraphQL(taskInfo)

	// Assert
	assert.Equal(t, "task-1", graphqlTask.TaskID)
	assert.Equal(t, "engineer", graphqlTask.Role)
	assert.Equal(t, "Test task", graphqlTask.Description)
	assert.Equal(t, TaskStatusCompleted, graphqlTask.Status)
	assert.Equal(t, 1.0, graphqlTask.Progress)
	assert.NotNil(t, graphqlTask.Error)
	assert.Equal(t, "Test error", *graphqlTask.Error)
	assert.NotNil(t, graphqlTask.ProjectRoot)
	assert.Equal(t, "/test/path", *graphqlTask.ProjectRoot)
}

func TestConvertTaskStatusMapping(t *testing.T) {
	tests := []struct {
		name           string
		inputStatus    string
		expectedStatus TaskStatus
	}{
		{"in_progress maps to IN_PROGRESS", "in_progress", TaskStatusInProgress},
		{"queued maps to IN_PROGRESS", "queued", TaskStatusInProgress},
		{"completed maps to COMPLETED", "completed", TaskStatusCompleted},
		{"failed maps to FAILED", "failed", TaskStatusFailed},
		{"unknown defaults to IN_PROGRESS", "unknown", TaskStatusInProgress},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now().Format(time.RFC3339)
			taskInfo := &TaskInfo{
				TaskID:    "task-1",
				Role:      "test",
				Task:      "test",
				Status:    tt.inputStatus,
				Progress:  0.5,
				CreatedAt: now,
				UpdatedAt: now,
				Metadata:  map[string]string{},
			}

			graphqlTask := convertTaskInfoToGraphQL(taskInfo)
			assert.Equal(t, tt.expectedStatus, graphqlTask.Status)
		})
	}
}
