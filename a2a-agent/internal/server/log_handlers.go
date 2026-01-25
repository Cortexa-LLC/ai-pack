package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
)

// HandleLogsStream streams logs via SSE
func (s *AgentServer) HandleLogsStream(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get log buffer
	logBuffer := monitoring.GetLogBuffer()
	if logBuffer == nil {
		http.Error(w, "Log buffer not initialized", http.StatusInternalServerError)
		return
	}

	// Subscribe to log stream
	logChan := logBuffer.Subscribe()
	defer logBuffer.Unsubscribe(logChan)

	// Get flusher for SSE
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Send initial connection event
	fmt.Fprintf(w, "event: connected\n")
	fmt.Fprintf(w, "data: {\"message\":\"Log stream connected\"}\n\n")
	flusher.Flush()

	// Stream logs
	for {
		select {
		case <-r.Context().Done():
			// Client disconnected
			return
		case entry, ok := <-logChan:
			if !ok {
				// Channel closed
				return
			}

			// Send log entry as SSE event
			data, err := json.Marshal(entry)
			if err != nil {
				continue
			}

			fmt.Fprintf(w, "event: log\n")
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

// HandleLogsRecent returns recent log entries
func (s *AgentServer) HandleLogsRecent(w http.ResponseWriter, r *http.Request) {
	// Get limit from query param (default: 100)
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
			if limit > 1000 {
				limit = 1000 // Cap at 1000
			}
		}
	}

	// Get log buffer
	logBuffer := monitoring.GetLogBuffer()
	if logBuffer == nil {
		http.Error(w, "Log buffer not initialized", http.StatusInternalServerError)
		return
	}

	// Get recent entries
	entries := logBuffer.GetRecent(limit)

	// Return as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs":  entries,
		"count": len(entries),
		"limit": limit,
	})
}
