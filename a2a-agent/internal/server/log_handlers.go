package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/beads"
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

// HandleTaskLogs returns logs for a specific task
func (s *AgentServer) HandleTaskLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	// Extract task ID from path: /a2a/tasks/:task_id/logs
	path := strings.TrimPrefix(r.URL.Path, "/a2a/tasks/")
	path = strings.TrimSuffix(path, "/logs")
	taskID := path

	if taskID == "" {
		http.Error(w, "Task ID required", http.StatusBadRequest)
		return
	}

	// Try to find the task execution
	s.mu.RLock()
	execution, exists := s.activeTasks[taskID]
	s.mu.RUnlock()

	// Determine the project root and log file path
	var logFile string
	var projectRoot string
	var executionFolder string

	if exists {
		// Active task - use execution info
		projectRoot = execution.ProjectRoot
		if projectRoot == "" {
			projectRoot = s.rootDir
		}
		// For active tasks, check metadata for the actual execution folder (timestamped)
		if btid, ok := execution.metadata["beads_task_id"]; ok && btid != "" {
			// execution.TaskID is the timestamped folder
			executionFolder = execution.TaskID
		} else {
			executionFolder = execution.TaskID
		}
		logFile = filepath.Join(projectRoot, ".beads", "tasks", executionFolder, "execution.log")
	} else {
		// Check global execution log first for completed tasks
		if foundRoot, _, err := s.findTaskProjectRoot(taskID); err == nil && foundRoot != "" {
			projectRoot = foundRoot
		}

		// If not in execution log, try finding beads task
		if projectRoot == "" {
			projectRoot = s.findBeadsTaskProjectRoot(taskID)
		}

		if projectRoot == "" {
			projectRoot = s.rootDir // Fallback to server root
		}

		// Find the most recent execution folder for this Beads task ID
		// taskID might be just the Beads ID (e.g., "xasm++-syq1") without timestamp
		executionFolder = s.findMostRecentExecutionFolder(projectRoot, taskID)
		if executionFolder == "" {
			// No timestamped folder found, try direct path (legacy or non-execution task)
			executionFolder = taskID
		}

		// Try direct task directory first (agent-spawned tasks)
		logFile = filepath.Join(projectRoot, ".beads", "tasks", executionFolder, "execution.log")

		// If not found, search execution directories for beads task ID
		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			if execLog := s.findBeadsTaskExecutionLog(projectRoot, taskID); execLog != "" {
				logFile = execLog
			}
		}
	}

	// Check if log file exists first - if it exists, serve logs regardless of Beads status
	logExists := false
	if _, err := os.Stat(logFile); err == nil {
		logExists = true
	}

	// Determine task status (we need this for streaming decision)
	var taskStatus string
	if exists {
		taskStatus = execution.Status
	} else if logExists {
		// Try to get status from execution metadata first
		metadataPath := filepath.Join(projectRoot, ".beads", "tasks", executionFolder, "00-metadata.json")
		if metadataData, err := os.ReadFile(metadataPath); err == nil {
			var metadata map[string]interface{}
			if json.Unmarshal(metadataData, &metadata) == nil {
				if status, ok := metadata["status"].(string); ok {
					taskStatus = status
				}
			}
		}

		// If still no status, check log content for completion markers
		if taskStatus == "" {
			if content, err := os.ReadFile(logFile); err == nil {
				contentStr := string(content)
				if strings.Contains(contentStr, "✅ Agent completed") ||
					strings.Contains(contentStr, "🎉 Task completed successfully") ||
					strings.Contains(contentStr, "❌ Task failed") {
					taskStatus = "completed"
				} else {
					// Logs exist but no completion marker - assume still running
					taskStatus = "in_progress"
				}
			}
		}
	} else {
		// For beads tasks without execution, get status from Beads
		if beadsTask, err := s.beadsClient.GetTaskFromDir(taskID, projectRoot); err == nil {
			taskStatus = beadsTask.Status
		}
	}

	// For tasks without execution logs, show the task contract instead
	if !logExists {
		s.serveTaskContract(w, r, taskID, projectRoot)
		return
	}

	// Determine if task is still running
	isTaskRunning := taskStatus == "in_progress"

	// Check if streaming is requested and task is still running
	// Completed tasks should never stream (follow mode)
	stream := r.URL.Query().Get("stream") == "true" && isTaskRunning

	if stream {
		s.streamTaskLogs(w, r, logFile, taskID)
	} else {
		s.serveTaskLogs(w, r, logFile)
	}
}

// serveTaskContract serves the task description/contract for queued tasks
func (s *AgentServer) serveTaskContract(w http.ResponseWriter, r *http.Request, taskID, projectRoot string) {
	// Try to get beads task info
	var beadsTask *beads.Task
	var err error

	if projectRoot != "" {
		beadsTask, err = s.beadsClient.GetTaskFromDir(taskID, projectRoot)
	} else {
		// Try all project roots
		for _, root := range s.GetProjectRoots() {
			beadsTask, err = s.beadsClient.GetTaskFromDir(taskID, root)
			if err == nil {
				break
			}
		}
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Task not found: %s", taskID), http.StatusNotFound)
		return
	}

	// Build markdown contract
	var contract strings.Builder
	contract.WriteString(fmt.Sprintf("# %s\n\n", beadsTask.Title))
	contract.WriteString(fmt.Sprintf("**Task ID:** %s  \n", beadsTask.ID))
	contract.WriteString(fmt.Sprintf("**Status:** %s  \n\n", beadsTask.Status))

	if beadsTask.Description != "" && beadsTask.Description != beadsTask.Title {
		contract.WriteString("## Description\n\n")
		contract.WriteString(beadsTask.Description)
		contract.WriteString("\n\n")
	}

	if len(beadsTask.Dependencies) > 0 {
		contract.WriteString("## Dependencies\n\n")
		for _, dep := range beadsTask.Dependencies {
			if depID, ok := dep.(string); ok {
				contract.WriteString(fmt.Sprintf("- %s\n", depID))
			} else if depMap, ok := dep.(map[string]interface{}); ok {
				if id, ok := depMap["id"].(string); ok {
					contract.WriteString(fmt.Sprintf("- %s\n", id))
				}
			}
		}
		contract.WriteString("\n")
	}

	if beadsTask.Metadata != nil && len(beadsTask.Metadata) > 0 {
		contract.WriteString("## Metadata\n\n")
		for k, v := range beadsTask.Metadata {
			contract.WriteString(fmt.Sprintf("- **%s:** %v\n", k, v))
		}
	}

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")

	w.Write([]byte(contract.String()))
}

// serveTaskLogs serves the complete log file as plain text
func (s *AgentServer) serveTaskLogs(w http.ResponseWriter, r *http.Request, logFile string) {
	// Read log file
	data, err := os.ReadFile(logFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read logs: %v", err), http.StatusInternalServerError)
		return
	}

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Return logs as plain text
	w.Write(data)
}

// streamTaskLogs streams log file updates via SSE
func (s *AgentServer) streamTaskLogs(w http.ResponseWriter, r *http.Request, logFile string, taskID string) {
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

	// Send initial connection event
	fmt.Fprintf(w, "event: connected\n")
	fmt.Fprintf(w, "data: {\"message\":\"Log stream connected\",\"task_id\":\"%s\"}\n\n", taskID)
	flusher.Flush()

	// Read initial content
	initialData, err := os.ReadFile(logFile)
	if err != nil {
		fmt.Fprintf(w, "event: error\n")
		fmt.Fprintf(w, "data: {\"message\":\"Failed to read log file\"}\n\n")
		flusher.Flush()
		return
	}

	// Send initial log lines
	scanner := bufio.NewScanner(strings.NewReader(string(initialData)))
	for scanner.Scan() {
		line := scanner.Text()
		// Escape the line for JSON
		escapedLine := strings.ReplaceAll(line, "\"", "\\\"")
		escapedLine = strings.ReplaceAll(escapedLine, "\n", "\\n")
		fmt.Fprintf(w, "event: log\n")
		fmt.Fprintf(w, "data: {\"line\":\"%s\"}\n\n", escapedLine)
		flusher.Flush()
	}

	// Check if task is already completed
	content := string(initialData)
	if strings.Contains(content, "✅ Agent completed") ||
		strings.Contains(content, "🎉 Task completed successfully") ||
		strings.Contains(content, "❌ Task failed") {
		// Task is done, close the stream
		fmt.Fprintf(w, "event: complete\n")
		fmt.Fprintf(w, "data: {\"message\":\"Task completed\"}\n\n")
		flusher.Flush()
		return
	}

	// Follow new lines
	lastSize := int64(len(initialData))
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			// Client disconnected
			return
		case <-ticker.C:
			// Check for new content
			stat, err := os.Stat(logFile)
			if err != nil {
				continue
			}

			if stat.Size() > lastSize {
				file, err := os.Open(logFile)
				if err != nil {
					continue
				}

				// Seek to last read position
				file.Seek(lastSize, 0)
				newData, _ := io.ReadAll(file)
				file.Close()

				// Send new lines
				scanner := bufio.NewScanner(strings.NewReader(string(newData)))
				for scanner.Scan() {
					line := scanner.Text()
					// Escape the line for JSON
					escapedLine := strings.ReplaceAll(line, "\"", "\\\"")
					escapedLine = strings.ReplaceAll(escapedLine, "\n", "\\n")
					fmt.Fprintf(w, "event: log\n")
					fmt.Fprintf(w, "data: {\"line\":\"%s\"}\n\n", escapedLine)
					flusher.Flush()

					// Check for completion markers
					if strings.Contains(line, "✅ Agent completed") ||
						strings.Contains(line, "🎉 Task completed successfully") ||
						strings.Contains(line, "❌ Task failed") {
						fmt.Fprintf(w, "event: complete\n")
						fmt.Fprintf(w, "data: {\"message\":\"Task completed\"}\n\n")
						flusher.Flush()
						return
					}
				}

				lastSize = stat.Size()
			}
		}
	}
}

// findBeadsTaskProjectRoot finds which project root contains the given beads task
// Returns the project root path, or empty string if not found
func (s *AgentServer) findBeadsTaskProjectRoot(taskID string) string {
	// Try each registered project root
	for _, root := range s.GetProjectRoots() {
		if _, err := s.beadsClient.GetTaskFromDir(taskID, root); err == nil {
			return root
		}
	}
	return ""
}

// findBeadsTaskExecutionLog searches for an execution log by beads task ID
// Beads stores execution logs in timestamped directories like task-engineer-20260202-122957-000000
// This function scans those directories to find which one has the matching beads_task_id in metadata
func (s *AgentServer) findBeadsTaskExecutionLog(projectRoot, beadsTaskID string) string {
	tasksDir := filepath.Join(projectRoot, ".beads", "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if directory name matches the beads task ID
		if entry.Name() == beadsTaskID {
			logPath := filepath.Join(tasksDir, entry.Name(), "execution.log")
			if _, err := os.Stat(logPath); err == nil {
				return logPath
			}
		}
	}

	return ""
}

// findMostRecentExecutionFolder finds the most recent timestamped execution folder for a Beads task ID
// Returns the folder name (e.g., "xasm++-syq1-20260211-080508") or empty string if not found
func (s *AgentServer) findMostRecentExecutionFolder(projectRoot, beadsTaskID string) string {
	tasksDir := filepath.Join(projectRoot, ".beads", "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return ""
	}

	var mostRecentFolder string
	var mostRecentTime time.Time

	// Find all folders matching {beads-id}-{timestamp} pattern
	prefix := beadsTaskID + "-"
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		folderName := entry.Name()
		// Check if folder matches pattern: {beads-id}-{timestamp}
		// or exact match (in case taskID is already the full folder name)
		if folderName == beadsTaskID || strings.HasPrefix(folderName, prefix) {
			// Get folder modification time as a proxy for execution time
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if mostRecentFolder == "" || info.ModTime().After(mostRecentTime) {
				mostRecentFolder = folderName
				mostRecentTime = info.ModTime()
			}
		}
	}

	return mostRecentFolder
}
