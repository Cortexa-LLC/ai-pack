package server

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/config"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
)

// TestHandleLogsStream_SSEHeaders tests that the log stream endpoint sets correct SSE headers
func TestHandleLogsStream_SSEHeaders(t *testing.T) {
	// Setup
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	cfg := config.DefaultConfig()
	cfg.API.Mode = "direct"
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create test HTTP request with cancellable context
	req := httptest.NewRequest(http.MethodGet, "/logs/stream", nil)
	w := httptest.NewRecorder()

	// Execute in goroutine since it's a long-running stream
	done := make(chan bool, 1)
	go func() {
		server.HandleLogsStream(w, req)
		done <- true
	}()

	// Give it time to write headers and start streaming
	time.Sleep(100 * time.Millisecond)

	// Get response (headers should be written by now)
	resp := w.Result()

	// Verify SSE headers
	if contentType := resp.Header.Get("Content-Type"); contentType != "text/event-stream" {
		t.Errorf("Expected Content-Type 'text/event-stream', got '%s'", contentType)
	}

	if cacheControl := resp.Header.Get("Cache-Control"); cacheControl != "no-cache" {
		t.Errorf("Expected Cache-Control 'no-cache', got '%s'", cacheControl)
	}

	if connection := resp.Header.Get("Connection"); connection != "keep-alive" {
		t.Errorf("Expected Connection 'keep-alive', got '%s'", connection)
	}

	if cors := resp.Header.Get("Access-Control-Allow-Origin"); cors != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin '*', got '%s'", cors)
	}

	// Close response body to trigger context cancellation
	resp.Body.Close()
}

// TestHandleLogsStream_ConnectedEvent tests that stream sends initial connected event
func TestHandleLogsStream_ConnectedEvent(t *testing.T) {
	// Setup
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	cfg := config.DefaultConfig()
	cfg.API.Mode = "direct"
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create test HTTP request
	req := httptest.NewRequest(http.MethodGet, "/logs/stream", nil)
	w := httptest.NewRecorder()

	// Execute in goroutine
	done := make(chan bool)
	go func() {
		server.HandleLogsStream(w, req)
		done <- true
	}()

	// Give it time to send connected event
	time.Sleep(100 * time.Millisecond)

	resp := w.Result()

	// Read initial events
	scanner := bufio.NewScanner(resp.Body)
	var events []string
	var currentEvent string

	// Read first event
	timeout := time.After(500 * time.Millisecond)
	eventChan := make(chan string, 1)

	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				if currentEvent != "" {
					eventChan <- currentEvent
					return
				}
			} else {
				if currentEvent != "" {
					currentEvent += "\n"
				}
				currentEvent += line
			}
		}
	}()

	select {
	case event := <-eventChan:
		events = append(events, event)
	case <-timeout:
		t.Fatal("Timeout waiting for connected event")
	}

	resp.Body.Close()

	// Verify connected event
	if len(events) == 0 {
		t.Fatal("Expected to receive connected event")
	}

	connectedEvent := events[0]
	if !strings.Contains(connectedEvent, "event: connected") {
		t.Errorf("Expected 'event: connected', got: %s", connectedEvent)
	}

	if !strings.Contains(connectedEvent, "Log stream connected") {
		t.Errorf("Expected 'Log stream connected' message, got: %s", connectedEvent)
	}
}

// TestHandleLogsStream_LogEvents tests that stream sends log events
func TestHandleLogsStream_LogEvents(t *testing.T) {
	// Setup
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	cfg := config.DefaultConfig()
	cfg.API.Mode = "direct"
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create test HTTP request
	req := httptest.NewRequest(http.MethodGet, "/logs/stream", nil)
	w := httptest.NewRecorder()

	// Execute in goroutine
	go func() {
		server.HandleLogsStream(w, req)
	}()

	// Give stream time to start
	time.Sleep(50 * time.Millisecond)

	// Generate some log entries
	logBuffer := monitoring.GetLogBuffer()
	logBuffer.Add(monitoring.LogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Message:   "test_message",
		Attrs: map[string]interface{}{
			"task_id": "test-123",
		},
	})

	// Give time for event to be sent
	time.Sleep(100 * time.Millisecond)

	resp := w.Result()
	defer resp.Body.Close()

	// Read events
	scanner := bufio.NewScanner(resp.Body)
	foundLogEvent := false
	foundTestMessage := false

	// Read with timeout
	timeout := time.After(500 * time.Millisecond)
	lineChan := make(chan string, 10)

	go func() {
		for scanner.Scan() {
			lineChan <- scanner.Text()
		}
		close(lineChan)
	}()

	for {
		select {
		case line, ok := <-lineChan:
			if !ok {
				goto done
			}
			if strings.Contains(line, "event: log") {
				foundLogEvent = true
			}
			if strings.Contains(line, "test_message") {
				foundTestMessage = true
			}
		case <-timeout:
			goto done
		}
	}

done:
	if !foundLogEvent {
		t.Error("Expected to find 'event: log'")
	}
	if !foundTestMessage {
		t.Error("Expected to find 'test_message' in log data")
	}
}

// TestHandleLogsRecent_DefaultLimit tests recent logs with default limit
func TestHandleLogsRecent_DefaultLimit(t *testing.T) {
	// Setup
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	cfg := config.DefaultConfig()
	cfg.API.Mode = "direct"
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Clear any existing logs from previous tests
	logBuffer := monitoring.GetLogBuffer()
	// Note: GetRecent doesn't modify buffer, so we just add fresh entries

	// Add exactly 10 test log entries
	for i := 0; i < 10; i++ {
		logBuffer.Add(monitoring.LogEntry{
			Timestamp: time.Now(),
			Level:     "INFO",
			Message:   "test_log_default_limit",
			Attrs: map[string]interface{}{
				"index": i,
			},
		})
	}

	// Create test HTTP request
	req := httptest.NewRequest(http.MethodGet, "/logs/recent", nil)
	w := httptest.NewRecorder()

	// Execute
	server.HandleLogsRecent(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Verify status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	// Verify Content-Type
	if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify structure
	if _, ok := result["logs"]; !ok {
		t.Error("Expected 'logs' field in response")
	}

	if limit, ok := result["limit"].(float64); !ok || limit != 100 {
		t.Errorf("Expected limit 100, got %v", result["limit"])
	}

	if count, ok := result["count"].(float64); !ok || count != 10 {
		t.Errorf("Expected count 10, got %v", result["count"])
	}
}

// TestHandleLogsRecent_CustomLimit tests recent logs with custom limit
func TestHandleLogsRecent_CustomLimit(t *testing.T) {
	// Setup
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	cfg := config.DefaultConfig()
	cfg.API.Mode = "direct"
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Add test log entries
	logBuffer := monitoring.GetLogBuffer()
	for i := 0; i < 20; i++ {
		logBuffer.Add(monitoring.LogEntry{
			Timestamp: time.Now(),
			Level:     "INFO",
			Message:   "test_log",
			Attrs: map[string]interface{}{
				"index": i,
			},
		})
	}

	// Create test HTTP request with limit=5
	req := httptest.NewRequest(http.MethodGet, "/logs/recent?limit=5", nil)
	w := httptest.NewRecorder()

	// Execute
	server.HandleLogsRecent(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify limit is respected
	if limit, ok := result["limit"].(float64); !ok || limit != 5 {
		t.Errorf("Expected limit 5, got %v", result["limit"])
	}

	logs, ok := result["logs"].([]interface{})
	if !ok {
		t.Fatal("Expected logs to be an array")
	}

	if len(logs) != 5 {
		t.Errorf("Expected 5 log entries, got %d", len(logs))
	}
}

// TestHandleLogsRecent_MaxLimit tests that limit is capped at 1000
func TestHandleLogsRecent_MaxLimit(t *testing.T) {
	// Setup
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	cfg := config.DefaultConfig()
	cfg.API.Mode = "direct"
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create test HTTP request with limit > 1000
	req := httptest.NewRequest(http.MethodGet, "/logs/recent?limit=5000", nil)
	w := httptest.NewRecorder()

	// Execute
	server.HandleLogsRecent(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify limit is capped at 1000
	if limit, ok := result["limit"].(float64); !ok || limit != 1000 {
		t.Errorf("Expected limit capped at 1000, got %v", result["limit"])
	}
}

// TestHandleLogsRecent_InvalidLimit tests handling of invalid limit parameter
func TestHandleLogsRecent_InvalidLimit(t *testing.T) {
	// Setup
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	cfg := config.DefaultConfig()
	cfg.API.Mode = "direct"
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tests := []struct {
		name          string
		queryParam    string
		expectedLimit float64
	}{
		{"invalid string", "limit=invalid", 100},
		{"negative number", "limit=-5", 100},
		{"zero", "limit=0", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/logs/recent?"+tt.queryParam, nil)
			w := httptest.NewRecorder()

			server.HandleLogsRecent(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			// Should fall back to default limit of 100
			if limit, ok := result["limit"].(float64); !ok || limit != tt.expectedLimit {
				t.Errorf("Expected limit %v for %s, got %v", tt.expectedLimit, tt.name, result["limit"])
			}
		})
	}
}

// TestHandleLogsRecent_LogEntryFormat tests the format of returned log entries
func TestHandleLogsRecent_LogEntryFormat(t *testing.T) {
	// Setup
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	cfg := config.DefaultConfig()
	cfg.API.Mode = "direct"
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Add a test log entry with known structure
	logBuffer := monitoring.GetLogBuffer()
	testTime := time.Now()
	logBuffer.Add(monitoring.LogEntry{
		Timestamp: testTime,
		Level:     "ERROR",
		Message:   "test_error_message",
		Attrs: map[string]interface{}{
			"task_id": "task-456",
			"error":   "something went wrong",
		},
	})

	// Create test HTTP request
	req := httptest.NewRequest(http.MethodGet, "/logs/recent?limit=1", nil)
	w := httptest.NewRecorder()

	// Execute
	server.HandleLogsRecent(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	logs, ok := result["logs"].([]interface{})
	if !ok || len(logs) == 0 {
		t.Fatal("Expected at least one log entry")
	}

	logEntry, ok := logs[0].(map[string]interface{})
	if !ok {
		t.Fatal("Expected log entry to be an object")
	}

	// Verify all required fields
	if _, ok := logEntry["timestamp"].(string); !ok {
		t.Errorf("Expected timestamp to be a string, got %T", logEntry["timestamp"])
	}

	if level, ok := logEntry["level"].(string); !ok || level != "ERROR" {
		t.Errorf("Expected level 'ERROR', got %v", logEntry["level"])
	}

	if message, ok := logEntry["message"].(string); !ok || message != "test_error_message" {
		t.Errorf("Expected message 'test_error_message', got %v", logEntry["message"])
	}

	if attrs, ok := logEntry["attrs"].(map[string]interface{}); !ok {
		t.Error("Expected attrs to be an object")
	} else {
		if taskID, ok := attrs["task_id"].(string); !ok || taskID != "task-456" {
			t.Errorf("Expected task_id 'task-456', got %v", attrs["task_id"])
		}
		if errMsg, ok := attrs["error"].(string); !ok || errMsg != "something went wrong" {
			t.Errorf("Expected error message, got %v", attrs["error"])
		}
	}
}
