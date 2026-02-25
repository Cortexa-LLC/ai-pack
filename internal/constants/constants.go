package constants

import "time"

// Task and Execution Status
const (
	StatusPending    = "pending"
	StatusQueued     = "queued"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
	StatusPaused     = "paused"
	StatusOpen       = "open"
	StatusClosed     = "closed"
	StatusDone       = "done"
)

// LLM Providers
const (
	ProviderAnthropic = "anthropic"
	ProviderOpenAI    = "openai"
	ProviderGemini    = "gemini"
)

// Message Roles (for LLM conversations)
const (
	MessageRoleUser      = "user"
	MessageRoleAssistant = "assistant"
	MessageRoleSystem    = "system"
)

// Buffer Sizes
const (
	// Channel buffer sizes for concurrent operations
	DefaultChannelBuffer = 100

	// Log streaming buffer
	LogStreamBuffer = 100

	// Task queue buffer
	TaskQueueBuffer = 100

	// SSE stream event buffer
	SSEEventBuffer = 100
)

// Timeouts and Intervals
const (
	// Polling interval for task status checks
	TaskStatusPollInterval = 1 * time.Second

	// Orchestrator update interval
	OrchestratorUpdateInterval = 10 * time.Second

	// Log streaming ticker interval
	LogStreamInterval = 1 * time.Second

	// Health check grace period
	HealthCheckGracePeriod = 2 * time.Second

	// Default HTTP client timeout
	DefaultHTTPTimeout = 5 * time.Second
)

// Metrics Configuration
const (
	// Maximum number of task durations to keep for statistics
	MaxTaskDurations = 1000

	// Maximum number of task token usage records to keep
	MaxTokenUsageRecords = 100

	// Maximum number of turn token data records to keep
	MaxTurnDataRecords = 1000

	// Default interval for updating provider costs (days)
	DefaultCostUpdateInterval = 30
)

// HTTP and Content Types
const (
	ContentTypeJSON        = "application/json"
	ContentTypeTextPlain   = "text/plain"
	ContentTypeEventStream = "text/event-stream"
)

// Event Types
const (
	EventTypeTaskStarted   = "task_started"
	EventTypeTaskCompleted = "task_completed"
	EventTypeTaskFailed    = "task_failed"
	EventTypeProgress      = "progress"
	EventTypeError         = "error"
)

// Content Block Types
const (
	ContentTypeText       = "text"
	ContentTypeToolUse    = "tool_use"
	ContentTypeToolResult = "tool_result"
)

// Message Event Types (Streaming)
const (
	MessageEventStart = "message_start"
	MessageEventDelta = "message_delta"
	MessageEventStop  = "message_stop"
	ContentBlockStart = "content_block_start"
	ContentBlockDelta = "content_block_delta"
	ContentBlockStop  = "content_block_stop"
	MessageDeltaUsage = "message_delta"
	EventTypePing     = "ping"
)

// A2A Protocol Methods
const (
	MethodExecute      = "execute"
	MethodStream       = "stream"
	MethodA2AExecute   = "a2a.execute"
	MethodStatus       = "status"
	MethodA2AStatus    = "a2a.status"
	MethodDiscovery    = "discovery"
	MethodA2ADiscovery = "a2a.discovery"
)

// Beads Actions
const (
	BeadsActionCreate = "create"
	BeadsActionUpdate = "update"
	BeadsActionList   = "list"
	BeadsActionGet    = "get"
	BeadsActionDelete = "delete"
)

// Tool Names
const (
	ToolBash   = "bash"
	ToolRead   = "read"
	ToolWrite  = "write"
	ToolEdit   = "edit"
	ToolGrep   = "grep"
	ToolGlob   = "glob"
	ToolSearch = "search"
	ToolBrowse = "browse"
)

// File Names
const (
	MetadataFileName        = "00-metadata.json"
	ProjectRegistryFileName = "project-registry.json"
	ConfigFileName          = "agent-server.json"
)

// Directory Names
const (
	BeadsDir     = ".beads"
	ClaudeDir    = ".claude"
	AIDir        = ".ai"
	AIPackDir    = ".ai-pack"
	RolesDir     = "roles"
	AgentsDir    = "agents"
	TasksDir     = "tasks"
	LogsDir      = "logs"
	TemplatesDir = "templates"
)

// Token Limits
const (
	MaxContextTokens = 200000 // Claude API limit
)

// Project Settings
const (
	// Project cleanup threshold (days)
	ProjectInactiveDays = 30
)

// Severity Levels
const (
	SeverityError   = "error"
	SeverityWarning = "warning"
	SeverityInfo    = "info"
)

// Response Status
const (
	ResponseStatusOK    = "ok"
	ResponseStatusError = "error"
)

// Default Values
const (
	// Default log level
	DefaultLogLevel = "info"

	// Default server host
	DefaultServerHost = "localhost"

	// Default server port
	DefaultServerPort = 8080

	// Default max concurrent agents
	DefaultMaxConcurrent = 10

	// Default max tokens per API call
	DefaultMaxTokens = 24000

	// Default max inactive turns before stopping agent
	DefaultMaxInactiveTurns = 10

	// Default task cleanup archive after days
	DefaultArchiveAfterDays = 15
)
