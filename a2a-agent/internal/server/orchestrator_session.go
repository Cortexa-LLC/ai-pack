package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
)

// OrchestratorSession manages a persistent orchestrator chat session
type OrchestratorSession struct {
	sessionID       string
	projectRoot     string
	monitoring      bool
	lastTaskCheck   time.Time
	knownTasks      map[string]string // task_id -> status
	updateChan      chan OrchestratorUpdate
	stopChan        chan bool
	mu              sync.RWMutex
	server          *AgentServer
}

// OrchestratorUpdate represents a proactive update from the orchestrator
type OrchestratorUpdate struct {
	Type      string                 `json:"type"`      // "task_complete", "task_blocked", "agent_spawned", "status"
	Message   string                 `json:"message"`   // Human-readable message
	TaskID    string                 `json:"task_id,omitempty"`
	Status    string                 `json:"status,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

var (
	sessions   = make(map[string]*OrchestratorSession)
	sessionsMu sync.RWMutex
)

// GetOrCreateSession gets or creates an orchestrator session for a project
func GetOrCreateSession(server *AgentServer, projectRoot string) *OrchestratorSession {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	sessionID := fmt.Sprintf("orch_%s_%d", projectRoot, time.Now().Unix())

	// Check if session exists for this project
	for _, session := range sessions {
		if session.projectRoot == projectRoot && session.monitoring {
			return session
		}
	}

	// Create new session
	session := &OrchestratorSession{
		sessionID:     sessionID,
		projectRoot:   projectRoot,
		monitoring:    true,
		lastTaskCheck: time.Now(),
		knownTasks:    make(map[string]string),
		updateChan:    make(chan OrchestratorUpdate, 100),
		stopChan:      make(chan bool),
		server:        server,
	}

	sessions[sessionID] = session

	// Start monitoring loop
	go session.monitorLoop()

	monitoring.Logger.Info("orchestrator_session_created", "session_id", sessionID, "project", projectRoot)

	return session
}

// monitorLoop runs in background and monitors task system
func (s *OrchestratorSession) monitorLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	monitoring.Logger.Info("orchestrator_monitor_started", "session_id", s.sessionID, "project", s.projectRoot)

	for {
		select {
		case <-s.stopChan:
			monitoring.Logger.Info("orchestrator_monitor_stopped", "session_id", s.sessionID)
			return
		case <-ticker.C:
			s.checkForUpdates()
		}
	}
}

// checkForUpdates queries the task system and detects changes
func (s *OrchestratorSession) checkForUpdates() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Query tasks via GraphQL
	tasks, err := s.queryTasks()
	if err != nil {
		monitoring.Logger.Error("orchestrator_query_failed", "error", err)
		return
	}

	// Check for task status changes
	for _, task := range tasks {
		oldStatus, known := s.knownTasks[task.ID]

		if !known {
			// New task discovered
			s.knownTasks[task.ID] = task.Status
			monitoring.Logger.Debug("orchestrator_new_task", "task_id", task.ID, "status", task.Status)
			continue
		}

		// Task status changed
		if oldStatus != task.Status {
			s.handleTaskStatusChange(task.ID, oldStatus, task.Status, &task)
			s.knownTasks[task.ID] = task.Status
		}
	}

	// Check for newly unblocked tasks
	s.checkForReadyTasks(tasks)
}

// handleTaskStatusChange handles when a task changes status
func (s *OrchestratorSession) handleTaskStatusChange(taskID, oldStatus, newStatus string, task *TaskInfo) {
	monitoring.Logger.Info("orchestrator_status_change", "task_id", taskID, "old", oldStatus, "new", newStatus)

	var message string
	var updateType string

	switch newStatus {
	case "completed":
		message = fmt.Sprintf("✅ Task %s completed", taskID)
		updateType = "task_complete"
	case "failed":
		message = fmt.Sprintf("❌ Task %s failed", taskID)
		updateType = "task_failed"
	case "blocked":
		message = fmt.Sprintf("🚧 Task %s blocked", taskID)
		updateType = "task_blocked"
	case "in_progress":
		message = fmt.Sprintf("⏳ Task %s in progress", taskID)
		updateType = "task_started"
	default:
		return
	}

	s.sendUpdate(OrchestratorUpdate{
		Type:      updateType,
		Message:   message,
		TaskID:    taskID,
		Status:    newStatus,
		Timestamp: time.Now(),
	})
}

// checkForReadyTasks looks for tasks that are ready to be executed
func (s *OrchestratorSession) checkForReadyTasks(tasks []TaskInfo) {
	for _, task := range tasks {
		// Check if task is in "queued" status and not blocked
		if task.Status == "queued" {
			// TODO: Check dependencies via GraphQL
			// For now, auto-spawn if queued
			monitoring.Logger.Info("orchestrator_ready_task", "task_id", task.ID)

			// Don't auto-spawn in Phase 2 - just notify
			// Phase 3 will add auto-spawning via tool use
			s.sendUpdate(OrchestratorUpdate{
				Type:      "task_ready",
				Message:   fmt.Sprintf("📋 Task %s is ready to be assigned", task.ID),
				TaskID:    task.ID,
				Status:    "queued",
				Timestamp: time.Now(),
			})
		}
	}
}

// TaskInfo represents task data from GraphQL
type TaskInfo struct {
	ID          string
	Status      string
	Title       string
	Description string
}

// queryTasks queries the GraphQL endpoint for tasks
func (s *OrchestratorSession) queryTasks() ([]TaskInfo, error) {
	// Query via HTTP GraphQL endpoint
	query := `
		query {
			tasks {
				taskID
				status
				task
				description
			}
		}
	`

	requestBody, err := json.Marshal(map[string]string{
		"query": query,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	resp, err := http.Post("http://localhost:8080/graphql", "application/json", bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to query GraphQL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GraphQL returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var result struct {
		Data struct {
			Tasks []struct {
				ID          string `json:"taskID"`
				Status      string `json:"status"`
				Task        string `json:"task"`
				Description string `json:"description"`
			} `json:"tasks"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse GraphQL response: %w", err)
	}

	tasks := make([]TaskInfo, len(result.Data.Tasks))
	for i, t := range result.Data.Tasks {
		tasks[i] = TaskInfo{
			ID:          t.ID,
			Status:      t.Status,
			Title:       t.Task,
			Description: t.Description,
		}
	}

	return tasks, nil
}

// sendUpdate sends an update to the SSE stream
func (s *OrchestratorSession) sendUpdate(update OrchestratorUpdate) {
	select {
	case s.updateChan <- update:
		monitoring.Logger.Debug("orchestrator_update_sent", "type", update.Type, "message", update.Message)
	default:
		monitoring.Logger.Warn("orchestrator_update_dropped", "type", update.Type)
	}
}

// Stop stops the monitoring loop
func (s *OrchestratorSession) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.monitoring {
		s.monitoring = false
		close(s.stopChan)
		monitoring.Logger.Info("orchestrator_session_stopped", "session_id", s.sessionID)
	}
}

// HandleOrchestratorSSE handles SSE connections for orchestrator updates
func (s *AgentServer) HandleOrchestratorSSE(w http.ResponseWriter, r *http.Request) {
	projectRoot := r.URL.Query().Get("project_root")
	if projectRoot == "" {
		http.Error(w, "project_root query parameter required", http.StatusBadRequest)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Get or create session
	session := GetOrCreateSession(s, projectRoot)

	// Send connected event
	fmt.Fprintf(w, "event: connected\n")
	fmt.Fprintf(w, "data: {\"status\":\"connected\",\"session_id\":\"%s\"}\n\n", session.sessionID)
	flusher.Flush()

	monitoring.Logger.Info("orchestrator_sse_connected", "session_id", session.sessionID, "project", projectRoot)

	// Stream updates
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			monitoring.Logger.Info("orchestrator_sse_disconnected", "session_id", session.sessionID)
			return
		case update := <-session.updateChan:
			data, _ := json.Marshal(update)
			fmt.Fprintf(w, "event: update\n")
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}
