package graphql

import (
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// ServerInterface defines the interface for accessing server functionality
// This allows GraphQL resolvers to interact with the agent server without circular dependencies
type ServerInterface interface {
	GetActiveTasks() map[string]*TaskInfo
	GetAllTasks() map[string]*TaskInfo
	GetTaskStatus(taskID string) (*TaskInfo, error)
	SpawnAgent(role, task, projectRoot string) (*TaskInfo, error)
	CancelTask(taskID string) error
	CloseTask(taskID string) error
	GetMetrics() *MetricsInfo
}

// TaskInfo represents task information from the server
type TaskInfo struct {
	TaskID      string
	Role        string
	Task        string
	Status      string
	CreatedAt   string
	UpdatedAt   string
	CompletedAt *string
	Result      *string
	Error       *string
	Metadata    map[string]string
	ProjectRoot *string
}

// MetricsInfo represents system metrics
type MetricsInfo struct {
	TasksSpawned         int
	TasksCompleted       int
	TasksFailed          int
	TasksActive          int
	AverageDurationMs    float64
	TotalTokens          int64
	InputTokens          int64
	OutputTokens         int64
	APICalls             int64
	APISuccess           int64
	APIFailed            int64
	AverageTokensPerTask int64
	Uptime               string
	// Detailed metrics
	TotalTurns          int64
	AvgInputPerTurn     int64
	AvgOutputPerTurn    int64
	RecentTurns         []monitoring.TurnTokenData
	RecentSessions      []monitoring.TaskTokenUsage
	StreamsOpened       int64
	StreamsClosed       int64
	StreamsActive       int64
	HTTPRequestsTotal   int64
	HTTPErrors          int64
	RateLimitViolations int64
}

// Resolver holds dependencies for GraphQL resolvers
type Resolver struct {
	server  ServerInterface
	monitor *monitoring.Metrics
}

// NewResolver creates a new GraphQL resolver with dependencies
func NewResolver(server ServerInterface, monitor *monitoring.Metrics) *Resolver {
	return &Resolver{
		server:  server,
		monitor: monitor,
	}
}
