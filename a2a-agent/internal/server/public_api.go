package server

import (
	"context"
	"net/http"
	"time"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/protocol"
)

// Public API methods for the server

// SpawnAgentTask spawns an agent task (public method for legacy endpoint)
func (s *AgentServer) SpawnAgentTask(role, task string) (*protocol.ExecuteTaskResponse, error) {
	return s.spawnAgentTask(role, task)
}

// ExecuteTaskSync executes a task synchronously and waits for completion
// Used for protocol handler mode (agent:// URLs)
func (s *AgentServer) ExecuteTaskSync(role, task string) (*protocol.ExecuteTaskResponse, error) {
	// Spawn the task
	response, err := s.spawnAgentTask(role, task)
	if err != nil {
		return nil, err
	}

	taskID := response.TaskID

	// Wait for completion by polling status
	for {
		time.Sleep(1 * time.Second)

		status, err := s.getTaskStatus(taskID)
		if err != nil {
			return nil, err
		}

		if status.Status == "completed" {
			response.Status = "completed"
			response.Message = "Task completed successfully"
			return response, nil
		}

		if status.Status == "failed" {
			response.Status = "failed"
			response.Message = status.Error
			return response, nil
		}

		// Still in progress, continue polling
	}
}

// HandleA2ADiscovery is the public handler for A2A discovery
func (s *AgentServer) HandleA2ADiscovery(w http.ResponseWriter, r *http.Request) {
	s.handleA2ADiscovery(w, r)
}

// HandleA2AExecute is the public handler for A2A execute
func (s *AgentServer) HandleA2AExecute(w http.ResponseWriter, r *http.Request) {
	s.handleA2AExecute(w, r)
}

// HandleA2AStatus is the public handler for A2A status
func (s *AgentServer) HandleA2AStatus(w http.ResponseWriter, r *http.Request) {
	s.handleA2AStatus(w, r)
}

// HandleStream is the public handler for SSE streaming
func (s *AgentServer) HandleStream(w http.ResponseWriter, r *http.Request) {
	s.handleStream(w, r)
}

// GetMetricsSnapshot returns a snapshot of current server metrics
func (s *AgentServer) GetMetricsSnapshot() monitoring.MetricsSnapshot {
	return monitoring.GlobalMetrics.GetSnapshot()
}

// GetMaxConcurrent returns the maximum concurrent agents setting
func (s *AgentServer) GetMaxConcurrent() int {
	return s.maxConcurrent
}

// LoggingMiddleware creates HTTP middleware for request logging
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := context.Background()

		monitoring.GlobalMetrics.IncrementHTTPRequests()

		// Wrap response writer to capture status code
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call next handler
		next.ServeHTTP(lrw, r)

		// Log request
		durationMs := time.Since(start).Milliseconds()
		monitoring.LogHTTPRequest(ctx, r.Method, r.URL.Path, lrw.statusCode, durationMs)

		if lrw.statusCode >= 400 {
			monitoring.GlobalMetrics.IncrementHTTPErrors()
		}
	})
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Flush implements http.Flusher for SSE streaming support
func (lrw *loggingResponseWriter) Flush() {
	if flusher, ok := lrw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
