package monitoring

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"
)

// Global logger instance
var Logger *slog.Logger

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Attrs     map[string]interface{} `json:"attrs"`
}

// LogBuffer stores recent log entries
type LogBuffer struct {
	mu      sync.RWMutex
	entries []LogEntry
	maxSize int
	subs    []chan LogEntry // Subscribers for streaming
}

var globalLogBuffer *LogBuffer

// NewLogBuffer creates a new log buffer
func NewLogBuffer(maxSize int) *LogBuffer {
	return &LogBuffer{
		entries: make([]LogEntry, 0, maxSize),
		maxSize: maxSize,
		subs:    make([]chan LogEntry, 0),
	}
}

// Add adds a log entry to the buffer
func (lb *LogBuffer) Add(entry LogEntry) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// Add to buffer
	lb.entries = append(lb.entries, entry)
	if len(lb.entries) > lb.maxSize {
		// Keep only the most recent entries
		lb.entries = lb.entries[len(lb.entries)-lb.maxSize:]
	}

	// Broadcast to subscribers (non-blocking)
	for _, sub := range lb.subs {
		select {
		case sub <- entry:
		default:
			// Skip if channel is full
		}
	}
}

// Clear removes all log entries from the buffer
func (lb *LogBuffer) Clear() {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.entries = nil
}

// GetRecent returns the N most recent log entries
func (lb *LogBuffer) GetRecent(n int) []LogEntry {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	if n > len(lb.entries) {
		n = len(lb.entries)
	}

	// Return the last N entries
	start := len(lb.entries) - n
	result := make([]LogEntry, n)
	copy(result, lb.entries[start:])
	return result
}

// Subscribe creates a new subscriber channel for log streaming
func (lb *LogBuffer) Subscribe() chan LogEntry {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	ch := make(chan LogEntry, 100)
	lb.subs = append(lb.subs, ch)
	return ch
}

// Unsubscribe removes a subscriber channel
func (lb *LogBuffer) Unsubscribe(ch chan LogEntry) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for i, sub := range lb.subs {
		if sub == ch {
			lb.subs = append(lb.subs[:i], lb.subs[i+1:]...)
			close(ch)
			break
		}
	}
}

// BufferedHandler wraps slog.Handler to also write to log buffer
type BufferedHandler struct {
	handler slog.Handler
	buffer  *LogBuffer
}

func (h *BufferedHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *BufferedHandler) Handle(ctx context.Context, r slog.Record) error {
	// Create log entry for buffer
	attrs := make(map[string]interface{})
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})

	entry := LogEntry{
		Timestamp: r.Time,
		Level:     r.Level.String(),
		Message:   r.Message,
		Attrs:     attrs,
	}

	// Add to buffer
	h.buffer.Add(entry)

	// Pass through to underlying handler
	return h.handler.Handle(ctx, r)
}

func (h *BufferedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &BufferedHandler{
		handler: h.handler.WithAttrs(attrs),
		buffer:  h.buffer,
	}
}

func (h *BufferedHandler) WithGroup(name string) slog.Handler {
	return &BufferedHandler{
		handler: h.handler.WithGroup(name),
		buffer:  h.buffer,
	}
}

// InitLogger initializes structured logging with buffering
func InitLogger(level slog.Level) {
	// Create log buffer (store last 1000 entries)
	globalLogBuffer = NewLogBuffer(1000)

	opts := &slog.HandlerOptions{
		Level: level,
	}

	// Use TextHandler for terminal output
	textHandler := slog.NewTextHandler(os.Stderr, opts)
	bufferedHandler := &BufferedHandler{
		handler: textHandler,
		buffer:  globalLogBuffer,
	}

	Logger = slog.New(bufferedHandler)
}

// GetLogBuffer returns the global log buffer
func GetLogBuffer() *LogBuffer {
	return globalLogBuffer
}

// LogTaskSpawned logs when a task is spawned
func LogTaskSpawned(ctx context.Context, taskID, role, task string) {
	Logger.InfoContext(ctx, "task_spawned",
		slog.String("task_id", taskID),
		slog.String("role", role),
		slog.String("task", task),
		slog.String("status", "queued"),
	)
}

// LogTaskStarted logs when a task starts execution
func LogTaskStarted(ctx context.Context, taskID, role string) {
	Logger.InfoContext(ctx, "task_started",
		slog.String("task_id", taskID),
		slog.String("role", role),
		slog.String("status", "in_progress"),
	)
}

// LogTaskCompleted logs when a task completes successfully
func LogTaskCompleted(ctx context.Context, taskID, role string, durationMs int64) {
	Logger.InfoContext(ctx, "task_completed",
		slog.String("task_id", taskID),
		slog.String("role", role),
		slog.String("status", "completed"),
		slog.Int64("duration_ms", durationMs),
	)
}

// LogTaskFailed logs when a task fails
func LogTaskFailed(ctx context.Context, taskID, role, errorMsg string, durationMs int64) {
	Logger.ErrorContext(ctx, "task_failed",
		slog.String("task_id", taskID),
		slog.String("role", role),
		slog.String("status", "failed"),
		slog.String("error", errorMsg),
		slog.Int64("duration_ms", durationMs),
	)
}

// LogAPICall logs Anthropic API calls
func LogAPICall(ctx context.Context, taskID, model string, tokenCount int) {
	Logger.InfoContext(ctx, "api_call",
		slog.String("task_id", taskID),
		slog.String("model", model),
		slog.Int("tokens", tokenCount),
	)
}

// LogStreamOpened logs when an SSE stream is opened
func LogStreamOpened(ctx context.Context, taskID string) {
	Logger.InfoContext(ctx, "stream_opened",
		slog.String("task_id", taskID),
	)
}

// LogStreamClosed logs when an SSE stream is closed
func LogStreamClosed(ctx context.Context, taskID string) {
	Logger.InfoContext(ctx, "stream_closed",
		slog.String("task_id", taskID),
	)
}

// LogHTTPRequest logs HTTP requests
func LogHTTPRequest(ctx context.Context, method, path string, statusCode int, durationMs int64) {
	Logger.InfoContext(ctx, "http_request",
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status_code", statusCode),
		slog.Int64("duration_ms", durationMs),
	)
}

// LogRateLimitExceeded logs rate limit violations
func LogRateLimitExceeded(ctx context.Context, identifier string) {
	Logger.WarnContext(ctx, "rate_limit_exceeded",
		slog.String("identifier", identifier),
	)
}
