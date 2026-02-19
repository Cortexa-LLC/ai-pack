package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/constants"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
)

// handleStream handles SSE streaming for task progress
// Uses hybrid approach: channel-based for active tasks, file-based for completed tasks
func (s *AgentServer) handleStream(w http.ResponseWriter, r *http.Request) {
	// Extract task ID from path: /stream/:task_id
	path := strings.TrimPrefix(r.URL.Path, "/stream/")
	taskID := path

	if taskID == "" {
		http.Error(w, "Task ID required", http.StatusBadRequest)
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

	ctx := r.Context()
	monitoring.LogStreamOpened(context.Background(), taskID)
	monitoring.GlobalMetrics.IncrementStreamsOpened()

	// Check if task is active (use channel-based streaming)
	s.mu.RLock()
	execution, isActive := s.activeTasks[taskID]
	s.mu.RUnlock()

	if isActive {
		// Active task - stream from channel with real-time updates
		s.streamActiveTaskFromChannel(ctx, w, flusher, execution)
	} else {
		// Completed task - stream from log file
		projectRoot, status, err := s.findTaskProjectRoot(taskID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Task not found: %s - %v", taskID, err), http.StatusNotFound)
			return
		}

		logPath := filepath.Join(projectRoot, constants.BeadsDir, "tasks", taskID, "execution.log")

		// Send initial connection event
		s.sendSSEEvent(w, flusher, "connected", map[string]interface{}{
			"task_id": taskID,
			"status":  status,
			"message": "Stream connected (historical)",
		})

		// Check if per-task log exists
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			// No per-task log - send summary from global log
			s.streamFromGlobalLog(w, flusher, taskID)
		} else {
			// Per-task log exists - stream it
			s.streamCompletedTaskLog(w, flusher, logPath, taskID)
		}

		// Stream closed
		s.sendSSEEvent(w, flusher, "stream_closed", map[string]interface{}{
			"task_id": taskID,
			"message": "Stream closed",
		})
	}

	monitoring.LogStreamClosed(context.Background(), taskID)
	monitoring.GlobalMetrics.IncrementStreamsClosed()
}

// streamActiveTaskFromChannel streams events from an active task's channel
func (s *AgentServer) streamActiveTaskFromChannel(ctx context.Context, w http.ResponseWriter, flusher http.Flusher, execution *TaskExecution) {
	taskID := execution.TaskID

	// Send initial connection event
	s.sendSSEEvent(w, flusher, "connected", map[string]interface{}{
		"task_id": taskID,
		"message": "Stream connected (live)",
	})

	// Heartbeat keeps the connection alive through proxies and prevents client inactivity timeouts.
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	// Stream events from channel
	for {
		select {
		case <-ctx.Done():
			// Client disconnected
			return
		case <-heartbeat.C:
			// SSE comment — ignored by clients but resets TCP/proxy idle timers
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		case event, ok := <-execution.streamChan:
			if !ok {
				// Channel closed, task completed
				s.sendSSEEvent(w, flusher, "stream_closed", map[string]interface{}{
					"task_id": taskID,
					"message": "Stream closed",
				})
				return
			}

			// Marshal the entire event structure
			eventData, err := json.Marshal(event)
			if err != nil {
				monitoring.Logger.Error("event_marshal_error", "task_id", taskID, "error", err)
				continue
			}

			// Send SSE event
			fmt.Fprintf(w, "data: %s\n\n", eventData)
			flusher.Flush()
		}
	}
}

// findTaskProjectRoot finds a task's project root by checking activeTasks or global execution log
func (s *AgentServer) findTaskProjectRoot(taskID string) (string, string, error) {
	// First check active tasks (fastest)
	s.mu.RLock()
	if execution, exists := s.activeTasks[taskID]; exists {
		projectRoot := execution.ProjectRoot
		status := execution.Status
		s.mu.RUnlock()
		return projectRoot, status, nil
	}
	s.mu.RUnlock()

	// Check global execution log
	if s.executionLog != nil {
		events, err := s.executionLog.GetEventsByTaskID(taskID)
		if err != nil {
			return "", "", fmt.Errorf("failed to query execution log: %w", err)
		}

		if len(events) == 0 {
			return "", "", fmt.Errorf("task not found in execution log")
		}

		// Get the most recent event for status
		latestEvent := events[len(events)-1]

		// Extract project_root from metadata (first event that has it)
		for _, event := range events {
			if event.Metadata != nil {
				if projectRoot, ok := event.Metadata["project_root"].(string); ok && projectRoot != "" {
					return projectRoot, latestEvent.Status, nil
				}
			}
		}

		return "", "", fmt.Errorf("no project_root found in task metadata")
	}

	return "", "", fmt.Errorf("task not found: %s", taskID)
}

// streamFromGlobalLog streams basic lifecycle events from global execution log
// Used when per-task log doesn't exist (e.g., old tasks)
func (s *AgentServer) streamFromGlobalLog(w http.ResponseWriter, flusher http.Flusher, taskID string) {
	if s.executionLog == nil {
		s.sendSSEEvent(w, flusher, "error", map[string]interface{}{
			"message": "No execution log available",
		})
		return
	}

	events, err := s.executionLog.GetEventsByTaskID(taskID)
	if err != nil {
		s.sendSSEEvent(w, flusher, "error", map[string]interface{}{
			"message": fmt.Sprintf("Failed to load task history: %v", err),
		})
		return
	}

	if len(events) == 0 {
		s.sendSSEEvent(w, flusher, "info", map[string]interface{}{
			"message": "No events recorded for this task",
		})
		return
	}

	// Convert ExecutionEvents to StreamEvent format and send
	for _, execEvent := range events {
		streamEvent := map[string]interface{}{
			"type":      execEvent.EventType,
			"task_id":   execEvent.TaskID,
			"timestamp": execEvent.Timestamp,
			"data": map[string]interface{}{
				"status":   execEvent.Status,
				"role":     execEvent.Role,
				"task":     execEvent.Task,
				"error":    execEvent.Error,
				"duration": execEvent.DurationMs,
				"result":   execEvent.Result,
			},
		}

		eventData, err := json.Marshal(streamEvent)
		if err != nil {
			monitoring.Logger.Error("event_marshal_error", "task_id", taskID, "error", err)
			continue
		}

		fmt.Fprintf(w, "data: %s\n\n", eventData)
		flusher.Flush()
		time.Sleep(10 * time.Millisecond)
	}
}

// streamCompletedTaskLog serves the entire log file for a completed task
// Reads JSONL file where each line is a StreamEvent
func (s *AgentServer) streamCompletedTaskLog(w http.ResponseWriter, flusher http.Flusher, logPath, taskID string) {
	file, err := os.Open(logPath)
	if err != nil {
		monitoring.Logger.Error("log_open_error", "task_id", taskID, "error", err)
		s.sendSSEEvent(w, flusher, "error", map[string]interface{}{
			"message": "Failed to open log file",
		})
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		// Each line is a JSON StreamEvent - send it directly as SSE
		fmt.Fprintf(w, "data: %s\n\n", line)
		flusher.Flush()

		// Small delay to avoid overwhelming client
		time.Sleep(10 * time.Millisecond)
	}

	if err := scanner.Err(); err != nil {
		monitoring.Logger.Error("log_scan_error", "task_id", taskID, "error", err)
	}
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
