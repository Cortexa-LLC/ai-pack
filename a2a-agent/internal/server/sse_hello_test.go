package server

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/config"
)

// Note: Helper functions setupTestDir and clearAuthEnvVars are defined in server_test.go

// TestSSEHelloWorld tests a simple SSE "Hello World" endpoint
// This verifies that SSE streaming works with a basic message
func TestSSEHelloWorld(t *testing.T) {
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
	req := httptest.NewRequest(http.MethodGet, "/sse/hello", nil)
	w := httptest.NewRecorder()

	// Execute
	server.handleSSEHello(w, req)

	// Verify
	resp := w.Result()
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	// Check SSE headers
	if contentType := resp.Header.Get("Content-Type"); contentType != "text/event-stream" {
		t.Errorf("Expected Content-Type 'text/event-stream', got '%s'", contentType)
	}

	if cacheControl := resp.Header.Get("Cache-Control"); cacheControl != "no-cache" {
		t.Errorf("Expected Cache-Control 'no-cache', got '%s'", cacheControl)
	}

	if connection := resp.Header.Get("Connection"); connection != "keep-alive" {
		t.Errorf("Expected Connection 'keep-alive', got '%s'", connection)
	}

	// Read response body and verify SSE format
	scanner := bufio.NewScanner(resp.Body)
	var events []string
	var currentEvent string

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			// Empty line marks end of event
			if currentEvent != "" {
				events = append(events, currentEvent)
				currentEvent = ""
			}
		} else {
			if currentEvent != "" {
				currentEvent += "\n"
			}
			currentEvent += line
		}
	}

	// Add last event if any
	if currentEvent != "" {
		events = append(events, currentEvent)
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading response: %v", err)
	}

	// Should have at least one event
	if len(events) == 0 {
		t.Fatal("Expected at least one SSE event")
	}

	// First event should be "hello" event with "Hello, World!" message
	firstEvent := events[0]
	if !strings.Contains(firstEvent, "event: hello") {
		t.Errorf("Expected event type 'hello', got: %s", firstEvent)
	}

	if !strings.Contains(firstEvent, "Hello, World!") {
		t.Errorf("Expected data to contain 'Hello, World!', got: %s", firstEvent)
	}
}

// TestSSEHelloWorldMultipleEvents tests SSE streaming with multiple events
func TestSSEHelloWorldMultipleEvents(t *testing.T) {
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
	req := httptest.NewRequest(http.MethodGet, "/sse/hello?count=3", nil)
	w := httptest.NewRecorder()

	// Execute
	server.handleSSEHello(w, req)

	// Verify
	resp := w.Result()
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	// Read response body and parse events
	scanner := bufio.NewScanner(resp.Body)
	eventCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "event: hello") {
			eventCount++
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading response: %v", err)
	}

	// Should have received 3 events
	if eventCount != 3 {
		t.Errorf("Expected 3 events, got %d", eventCount)
	}
}

// TestSSEHelloWorldTiming tests that SSE events are sent with proper timing
func TestSSEHelloWorldTiming(t *testing.T) {
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

	// Create test HTTP server to measure timing
	handler := http.HandlerFunc(server.handleSSEHello)
	testServer := httptest.NewServer(handler)
	defer testServer.Close()

	// Make request
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(testServer.URL + "/sse/hello")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Verify we can read events in real-time
	scanner := bufio.NewScanner(resp.Body)
	startTime := time.Now()

	if scanner.Scan() {
		// First event should arrive quickly
		elapsed := time.Since(startTime)
		if elapsed > 100*time.Millisecond {
			t.Errorf("First event took too long: %v", elapsed)
		}
	} else {
		t.Error("Expected to receive at least one event")
	}
}
