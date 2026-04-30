package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/constants"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/internal/taskdb"
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

	// Try to find the task execution.
	// activeTasks is keyed by full execution ID (e.g. "xasm++-qbxv-20260218-091500"),
	// but the GUI requests logs using the short task ID (e.g. "xasm++-qbxv").
	// Try exact match first, then prefix/metadata match for short IDs.
	s.mu.RLock()
	execution, exists := s.activeTasks[taskID]
	if !exists {
		prefix := taskID + "-"
		for _, exec := range s.activeTasks {
			if strings.HasPrefix(exec.TaskID, prefix) {
				execution = exec
				exists = true
				break
			}
			if exec.metadata != nil {
				if btid, ok := exec.metadata["task_id"]; ok && btid == taskID {
					execution = exec
					exists = true
					break
				}
			}
		}
	}
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
		if btid, ok := execution.metadata["task_id"]; ok && btid != "" {
			// execution.TaskID is the timestamped folder
			executionFolder = execution.TaskID
		} else {
			executionFolder = execution.TaskID
		}
		logFile = filepath.Join(projectRoot, constants.TaskRootDir, "tasks", executionFolder, "execution.log")
	} else {
		// Check global execution log first for completed tasks
		if foundRoot, _, err := s.findTaskProjectRootWithStatus(taskID); err == nil && foundRoot != "" {
			projectRoot = foundRoot
		}

		// If not in execution log, try finding task
		if projectRoot == "" {
			projectRoot = s.findTaskProjectRoot(taskID)
		}

		if projectRoot == "" {
			projectRoot = s.rootDir // Fallback to server root
		}

		// Find the most recent execution folder for this task ID
		// taskID might be just the task ID (e.g., "xasm++-syq1") without timestamp
		executionFolder = s.findMostRecentExecutionFolder(projectRoot, taskID)
		if executionFolder == "" {
			// No timestamped folder found, try direct path (legacy or non-execution task)
			executionFolder = taskID
		}

		// Try direct task directory first (agent-spawned tasks)
		logFile = filepath.Join(projectRoot, constants.TaskRootDir, "tasks", executionFolder, "execution.log")

		// If not found, search execution directories for task ID
		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			if execLog := s.findTaskExecutionLog(projectRoot, taskID); execLog != "" {
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
		metadataPath := filepath.Join(projectRoot, constants.TaskRootDir, "tasks", executionFolder, "00-metadata.json")
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
					taskStatus = constants.StatusCompleted
				} else {
					// Logs exist but no completion marker - assume still running
					taskStatus = constants.StatusInProgress
				}
			}
		}
	} else {
		// For tasks without execution, get status from Beads
		if s.taskDB != nil {
				if dbTask, err := s.taskDB.GetTask(taskID); err == nil && dbTask != nil {
					taskStatus = dbTask.Status
				}
			}
	}

	// For tasks without execution logs, show the task contract instead
	if !logExists {
		s.serveTaskContract(w, r, taskID, projectRoot)
		return
	}

	// Determine if task is still running.
	// A task can only be running if it is tracked in activeTasks — if the server
	// restarted mid-execution the task is gone from activeTasks and the log may have
	// no completion marker, but it is definitely not running anymore.
	isTaskRunning := exists && taskStatus == constants.StatusInProgress

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
	// Try to get task info from taskDB
	if s.taskDB == nil {
		http.Error(w, "Task database not available", http.StatusServiceUnavailable)
		return
	}

	dbTask, err := s.taskDB.GetTask(taskID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Task not found: %s", taskID), http.StatusNotFound)
		return
	}

	// Build markdown contract
	var contract strings.Builder
	shortID := taskdb.ExtractShortID(dbTask.ID)
	contract.WriteString(fmt.Sprintf("# Task %s\n\n", shortID))
	contract.WriteString(fmt.Sprintf("**Task ID:** %s  \n", dbTask.ID))
	contract.WriteString(fmt.Sprintf("**Short ID:** %s  \n", shortID))
	contract.WriteString(fmt.Sprintf("**Status:** %s  \n", dbTask.Status))
	contract.WriteString(fmt.Sprintf("**Role:** %s  \n\n", dbTask.Role))

	if dbTask.TaskDescription != "" {
		contract.WriteString("## Description\n\n")
		contract.WriteString(dbTask.TaskDescription)
		contract.WriteString("\n\n")
	}

	if dbTask.Metadata != "" {
		contract.WriteString("## Metadata\n\n")
		contract.WriteString(dbTask.Metadata)
		contract.WriteString("\n\n")
	}

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")

	w.Write([]byte(contract.String()))
}

// maxLogBytes is the maximum number of bytes served from a log file.
// Large tool outputs (grep on big files, etc.) can produce gigabyte-scale logs;
// reading the whole file into memory and sending it to the browser hangs the UI.
const maxLogBytes = 512 * 1024 // 512 KB

// serveTaskLogs serves the tail of a log file as plain text.
// If the file exceeds maxLogBytes only the last maxLogBytes are returned,
// with a truncation notice prepended.
func (s *AgentServer) serveTaskLogs(w http.ResponseWriter, r *http.Request, logFile string) {
	f, err := os.Open(logFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read logs: %v", err), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to stat log file: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	size := stat.Size()
	if size <= maxLogBytes {
		// Small enough — serve the whole file
		data, err := io.ReadAll(f)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read logs: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write(data)
		return
	}

	// Large file — seek to last maxLogBytes and serve from there
	offset := size - maxLogBytes
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		http.Error(w, fmt.Sprintf("Failed to seek log file: %v", err), http.StatusInternalServerError)
		return
	}

	// Find the next newline so we start on a clean line boundary
	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	newlineIdx := bytes.IndexByte(buf[:n], '\n')
	if newlineIdx >= 0 {
		// Seek past the partial line
		if _, err := f.Seek(offset+int64(newlineIdx)+1, io.SeekStart); err == nil {
			offset = offset + int64(newlineIdx) + 1
		}
	} else {
		f.Seek(offset, io.SeekStart)
	}

	truncationNotice := fmt.Sprintf(
		"[Log truncated: file is %.1f MB, showing last %.0f KB]\n\n",
		float64(size)/(1024*1024),
		float64(maxLogBytes)/1024,
	)
	w.Write([]byte(truncationNotice))
	io.Copy(w, f)
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

// findTaskProjectRoot finds which project root contains the given task
// Returns the project root path, or empty string if not found
func (s *AgentServer) findTaskProjectRoot(taskID string) string {
	// First: look for actual execution artifacts in each project root.
	// bd show is global (not directory-scoped), so calling GetTaskFromDir succeeds
	// for any working directory and cannot be used to determine the correct project root.
	prefix := taskID + "-"
	for _, root := range s.GetProjectRoots() {
		tasksDir := filepath.Join(root, constants.TaskRootDir, "tasks")
		entries, err := os.ReadDir(tasksDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			if name == taskID || strings.HasPrefix(name, prefix) {
				return root
			}
		}
	}
	return ""
}

// findTaskExecutionLog searches for an execution log by task ID
// Beads stores execution logs in timestamped directories like task-engineer-20260202-122957-000000
// This function scans those directories to find which one has the matching beads_task_id in metadata
func (s *AgentServer) findTaskExecutionLog(projectRoot, taskID string) string {
	tasksDir := filepath.Join(projectRoot, constants.TaskRootDir, "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if directory name matches the task ID
		if entry.Name() == taskID {
			logPath := filepath.Join(tasksDir, entry.Name(), "execution.log")
			if _, err := os.Stat(logPath); err == nil {
				return logPath
			}
		}
	}

	return ""
}

// findMostRecentExecutionFolder finds the most recent timestamped execution folder for a task ID
// Returns the folder name (e.g., "xasm++-syq1-20260211-080508") or empty string if not found
func (s *AgentServer) findMostRecentExecutionFolder(projectRoot, taskID string) string {
	tasksDir := filepath.Join(projectRoot, constants.TaskRootDir, "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return ""
	}

	var mostRecentFolder string
	var mostRecentTime time.Time

	// Find all folders matching {short-task-id}-{timestamp} pattern
	prefix := taskID + "-"
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		folderName := entry.Name()
		// Check if folder matches pattern: {short-task-id}-{timestamp}
		// or exact match (in case taskID is already the full folder name)
		if folderName == taskID || strings.HasPrefix(folderName, prefix) {
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

// TaskResultsResponse is the JSON shape returned by HandleTaskResults.
type TaskResultsResponse struct {
	TaskID string `json:"task_id"`
	Result string `json:"result"`
}

// HandleTaskResults serves GET /a2a/tasks/{taskID}/results.
// It locates the 30-results.md file for the requested task and returns its
// contents as JSON. A 404 is returned when no results file is found.
func (s *AgentServer) HandleTaskResults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	// Extract task ID from path: /a2a/tasks/{taskID}/results
	path := strings.TrimPrefix(r.URL.Path, "/a2a/tasks/")
	path = strings.TrimSuffix(path, "/results")
	taskID := path

	if taskID == "" {
		http.Error(w, "Task ID required", http.StatusBadRequest)
		return
	}

	// Determine projectRoot for the requested task (mirrors HandleTaskLogs).
	projectRoot, _, err := s.findTaskProjectRootWithStatus(taskID)
	if err != nil || projectRoot == "" {
		projectRoot = s.findTaskProjectRoot(taskID)
	}
	if projectRoot == "" {
		projectRoot = s.rootDir
	}

	// The execution folder name may be the task ID itself, or we may need
	// to scan for the most-recent matching folder.
	executionFolder := s.findMostRecentExecutionFolder(projectRoot, taskID)
	if executionFolder == "" {
		executionFolder = taskID
	}

	resultsPath := filepath.Join(projectRoot, constants.TaskRootDir, "tasks", executionFolder, "30-results.md")
	data, readErr := os.ReadFile(resultsPath)
	if readErr != nil {
		http.Error(w, "Results not found", http.StatusNotFound)
		return
	}

	resp := TaskResultsResponse{
		TaskID: taskID,
		Result: string(data),
	}
	w.Header().Set("Content-Type", "application/json")
	if encErr := json.NewEncoder(w).Encode(resp); encErr != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
