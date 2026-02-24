package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// handleSSEHello handles a simple SSE "Hello World" endpoint.
//
// This endpoint demonstrates basic Server-Sent Events (SSE) functionality.
// It sends one or more "hello" events followed by a completion event.
//
// Query Parameters:
//   - count: Number of hello events to send (default: 1, must be positive)
//
// Response Format:
//   - Content-Type: text/event-stream
//   - Events:
//   - "hello" events with JSON data containing message, count, total, and timestamp
//   - "complete" event with JSON data containing completion message and total count
//
// Example Usage:
//
//	GET /sse/hello          -> Sends 1 hello event
//	GET /sse/hello?count=3  -> Sends 3 hello events
//
// Error Handling:
//   - Returns 500 if streaming is not supported by the ResponseWriter
//   - Skips events with JSON marshaling errors (logs internally)
func (s *AgentServer) handleSSEHello(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get flusher for streaming
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Parse count parameter (default to 1)
	count := 1
	if countStr := r.URL.Query().Get("count"); countStr != "" {
		if parsedCount, err := strconv.Atoi(countStr); err == nil && parsedCount > 0 {
			count = parsedCount
		}
	}

	// Send hello events
	for i := 0; i < count; i++ {
		// Create event data
		eventData := map[string]interface{}{
			"message": "Hello, World!",
			"count":   i + 1,
			"total":   count,
			"time":    time.Now().Format(time.RFC3339),
		}

		// Send the event
		if err := sendSSEEvent(w, flusher, "hello", eventData); err != nil {
			// Skip this event on error (already logged)
			continue
		}

		// Add small delay between events (except for last one)
		// This simulates real-world streaming behavior
		if i < count-1 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	// Send completion event
	completionData := map[string]interface{}{
		"message": "SSE stream complete",
		"total":   count,
	}

	// Send completion event (ignore errors at this point)
	_ = sendSSEEvent(w, flusher, "complete", completionData)
}

// sendSSEEvent sends a single SSE event with proper formatting.
//
// Parameters:
//   - w: HTTP ResponseWriter to write the event to
//   - flusher: HTTP Flusher to flush the response buffer
//   - eventType: Type of the event (e.g., "hello", "complete")
//   - data: Event data to be JSON-encoded
//
// Returns:
//   - error if JSON marshaling fails
//
// SSE Format:
//
//	event: <eventType>
//	data: <json-encoded-data>
//	<blank line>
func sendSSEEvent(w http.ResponseWriter, flusher http.Flusher, eventType string, data map[string]interface{}) error {
	// Marshal data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal SSE event data: %w", err)
	}

	// Write event in SSE format
	fmt.Fprintf(w, "event: %s\n", eventType)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)

	// Flush the response buffer to send immediately
	flusher.Flush()

	return nil
}
