package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
)

// handleStream handles SSE streaming for task progress
func (s *AgentServer) handleStream(w http.ResponseWriter, r *http.Request) {
	// Extract task ID from path: /stream/:task_id
	path := strings.TrimPrefix(r.URL.Path, "/stream/")
	taskID := path

	if taskID == "" {
		http.Error(w, "Task ID required", http.StatusBadRequest)
		return
	}

	// Get task execution
	s.mu.RLock()
	execution, exists := s.activeTasks[taskID]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, fmt.Sprintf("Task not found: %s", taskID), http.StatusNotFound)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	monitoring.LogStreamOpened(ctx, taskID)
	monitoring.GlobalMetrics.IncrementStreamsOpened()

	// Send initial connection event
	s.sendSSEEvent(w, flusher, "connected", map[string]interface{}{
		"task_id": taskID,
		"message": "Stream connected",
	})

	// Stream events from channel
	for event := range execution.streamChan {
		eventData, err := json.Marshal(event.Data)
		if err != nil {
			monitoring.Logger.Error("event_marshal_error", "task_id", taskID, "error", err)
			continue
		}

		// Send SSE event
		fmt.Fprintf(w, "event: %s\n", event.Type)
		fmt.Fprintf(w, "data: %s\n\n", eventData)
		flusher.Flush()
	}

	// Stream closed
	s.sendSSEEvent(w, flusher, "stream_closed", map[string]interface{}{
		"task_id": taskID,
		"message": "Stream closed",
	})

	monitoring.LogStreamClosed(ctx, taskID)
	monitoring.GlobalMetrics.IncrementStreamsClosed()
}

func (s *AgentServer) sendSSEEvent(w http.ResponseWriter, flusher http.Flusher, eventType string, data map[string]interface{}) {
	eventData, err := json.Marshal(data)
	if err != nil {
		monitoring.Logger.Error("sse_event_marshal_error", "event_type", eventType, "error", err)
		return
	}

	fmt.Fprintf(w, "event: %s\n", eventType)
	fmt.Fprintf(w, "data: %s\n\n", eventData)
	flusher.Flush()
}
