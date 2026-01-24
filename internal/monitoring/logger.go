package monitoring

import (
	"context"
	"log/slog"
	"os"
)

// Global logger instance
var Logger *slog.Logger

// InitLogger initializes structured logging
func InitLogger(level slog.Level) {
	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	Logger = slog.New(handler)
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
