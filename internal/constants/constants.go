package constants

import (
	"os"
	"path/filepath"
	"time"
)

// ResolveExecutionLogPath returns the path to an execution log file.
// Runs are written to <projectRoot>/.ai/tasks/<runID>/execution.log.
// When the exact path doesn't exist, scans for the most recently modified
// directory whose name starts with runID (handles the case where
// latest_run_id is not populated in the DB and the stored ID is the
// short task ID rather than the full timestamped run directory name).
func ResolveExecutionLogPath(projectRoot, runID string) string {
	primary := filepath.Join(projectRoot, TaskRootDir, TasksDir, runID, "execution.log")
	if _, err := os.Stat(primary); err == nil {
		return primary
	}
	// Prefix scan: find the most recent run directory that starts with runID.
	if best := latestRunDirWithPrefix(projectRoot, runID); best != "" {
		return filepath.Join(best, "execution.log")
	}
	// Return the primary path even if it doesn't exist — callers handle missing files.
	return primary
}

// ResolveTaskDir returns the directory for a task execution run.
// Falls back to a prefix scan when the exact path doesn't exist.
func ResolveTaskDir(projectRoot, runID string) string {
	primary := filepath.Join(projectRoot, TaskRootDir, TasksDir, runID)
	if _, err := os.Stat(primary); err == nil {
		return primary
	}
	if best := latestRunDirWithPrefix(projectRoot, runID); best != "" {
		return best
	}
	return primary
}

// latestRunDirWithPrefix scans .ai/tasks/ for directories whose names begin
// with prefix and returns the one with the most recent modification time.
// Returns "" if none found.
func latestRunDirWithPrefix(projectRoot, prefix string) string {
	roots := []string{
		filepath.Join(projectRoot, TaskRootDir, TasksDir),
	}
	var bestPath string
	var bestTime int64
	for _, root := range roots {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() || len(e.Name()) <= len(prefix) {
				continue
			}
			// Match: name must start with prefix followed by a non-alphanumeric char
			// (e.g. "ai-pack-f32-20260512-..." matches prefix "ai-pack-f32").
			if e.Name()[:len(prefix)] != prefix {
				continue
			}
			next := e.Name()[len(prefix)]
			if next != '-' && next != '_' && next != '.' {
				continue
			}
			full := filepath.Join(root, e.Name())
			// Only consider directories that contain an execution.log — this
			// distinguishes run directories from task-packet directories which
			// share the same prefix but never have an execution log.
			if _, err := os.Stat(filepath.Join(full, "execution.log")); err != nil {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			if info.ModTime().UnixNano() > bestTime {
				bestTime = info.ModTime().UnixNano()
				bestPath = full
			}
		}
	}
	return bestPath
}

// Task and Execution Status
const (
	StatusPending    = "pending"
	StatusQueued     = "queued"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
	StatusPaused     = "paused"
	StatusCancelled  = "cancelled"
	StatusOpen       = "open"
	StatusClosed     = "closed"
	StatusDone       = "done"
)

// LLM Providers
const (
	ProviderAnthropic = "anthropic"
	ProviderOpenAI    = "openai"
	ProviderGemini    = "gemini"
	ProviderQwen      = "qwen"
)

// Local provider endpoints
const (
	QwenLocalBaseURL = "http://localhost:9090/v1"
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
	TaskRootDir  = ".ai"   // Root directory for task execution data
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

// Tool Result Limits
const (
	// MaxToolResultChars is the maximum number of characters kept from a single
	// tool result before it is truncated. Keeping results bounded prevents
	// runaway context growth when tools emit large outputs.
	MaxToolResultChars = 8000
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
