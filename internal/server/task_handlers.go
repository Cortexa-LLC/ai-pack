package server

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/internal/taskdb"
)

// HandleTaskCreate handles POST /a2a/tasks/create
func (s *AgentServer) HandleTaskCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		Description string `json:"description"`
		ProjectRoot string `json:"project_root"`
		Role        string `json:"role"`
		Metadata    string `json:"metadata"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.Description == "" {
		http.Error(w, "Description is required", http.StatusBadRequest)
		return
	}

	if req.ProjectRoot == "" {
		http.Error(w, "Project root is required", http.StatusBadRequest)
		return
	}

	// Generate task ID
	taskID := generateTaskID(req.ProjectRoot)

	// Create task in database
	task := &taskdb.Task{
		ID:              taskID,
		ProjectRoot:     req.ProjectRoot,
		Role:            req.Role,
		TaskDescription: req.Description,
		Metadata:        req.Metadata,
	}

	if s.taskDB == nil {
		http.Error(w, "Task database not initialized", http.StatusInternalServerError)
		return
	}

	if err := s.taskDB.CreateTask(task); err != nil {
		monitoring.Logger.Error("failed_to_create_task", "error", err.Error())
		http.Error(w, fmt.Sprintf("Failed to create task: %v", err), http.StatusInternalServerError)
		return
	}

	monitoring.Logger.Info("task_created", "task_id", taskID, "role", req.Role)

	// Return task ID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"task_id":      taskID,
		"status":       "queued",
		"project_root": req.ProjectRoot,
	})
}

// HandleTaskGet handles GET /a2a/tasks/{taskID}
func (s *AgentServer) HandleTaskGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract task ID from path: /a2a/tasks/{taskID}
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}
	taskID := pathParts[2]
	if taskID == "" {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	if s.taskDB == nil {
		http.Error(w, "Task database not initialized", http.StatusInternalServerError)
		return
	}

	// Query database
	task, err := s.taskDB.GetTask(taskID)
	if err != nil {
		monitoring.Logger.Error("failed_to_get_task", "task_id", taskID, "error", err.Error())
		http.Error(w, fmt.Sprintf("Failed to get task: %v", err), http.StatusInternalServerError)
		return
	}

	if task == nil {
		http.Error(w, fmt.Sprintf("Task not found: %s", taskID), http.StatusNotFound)
		return
	}

	// Build response
	response := map[string]interface{}{
		"id":          task.ID,
		"status":      task.Status,
		"role":        task.Role,
		"description": task.TaskDescription,
		"created_at":  task.CreatedAt.Format(time.RFC3339),
		"updated_at":  task.UpdatedAt.Format(time.RFC3339),
	}

	if task.ProjectRoot != "" {
		response["project_root"] = task.ProjectRoot
	}
	if task.StartedAt != nil {
		response["started_at"] = task.StartedAt.Format(time.RFC3339)
	}
	if task.CompletedAt != nil {
		response["completed_at"] = task.CompletedAt.Format(time.RFC3339)
	}
	if task.Result != "" {
		response["result"] = task.Result
	}
	if task.Error != "" {
		response["error"] = task.Error
	}
	if task.Metadata != "" {
		response["metadata"] = task.Metadata
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleTaskUpdate handles PUT /a2a/tasks/{taskID}
func (s *AgentServer) HandleTaskUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract task ID from path: /a2a/tasks/{taskID} or /a2a/tasks/{taskID}/...
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}
	taskID := pathParts[2]
	if taskID == "" {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req struct {
		Status   string `json:"status"`
		Result   string `json:"result"`
		Metadata string `json:"metadata"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read body: %v", err), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if s.taskDB == nil {
		http.Error(w, "Task database not initialized", http.StatusInternalServerError)
		return
	}

	// Verify task exists
	task, err := s.taskDB.GetTask(taskID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get task: %v", err), http.StatusInternalServerError)
		return
	}
	if task == nil {
		http.Error(w, fmt.Sprintf("Task not found: %s", taskID), http.StatusNotFound)
		return
	}

	// Update fields
	if req.Status != "" {
		if err := s.taskDB.UpdateTaskStatus(taskID, req.Status, ""); err != nil {
			http.Error(w, fmt.Sprintf("Failed to update status: %v", err), http.StatusInternalServerError)
			return
		}
	}

	if req.Result != "" {
		if err := s.taskDB.UpdateTaskResult(taskID, req.Result); err != nil {
			http.Error(w, fmt.Sprintf("Failed to update result: %v", err), http.StatusInternalServerError)
			return
		}
	}

	monitoring.Logger.Info("task_updated", "task_id", taskID, "status", req.Status)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"task_id": taskID,
		"status":  "updated",
	})
}

// HandleTaskClose handles PUT /a2a/tasks/{taskID}/close
func (s *AgentServer) HandleTaskClose(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract task ID from path: /a2a/tasks/{taskID}/close
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 || pathParts[3] != "close" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	taskID := pathParts[2]
	if taskID == "" {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req struct {
		Result string `json:"result"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if s.taskDB == nil {
		http.Error(w, "Task database not initialized", http.StatusInternalServerError)
		return
	}

	// Verify task exists
	task, err := s.taskDB.GetTask(taskID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get task: %v", err), http.StatusInternalServerError)
		return
	}
	if task == nil {
		http.Error(w, fmt.Sprintf("Task not found: %s", taskID), http.StatusNotFound)
		return
	}

	// Complete task
	if err := s.taskDB.CompleteTask(taskID, req.Result); err != nil {
		http.Error(w, fmt.Sprintf("Failed to complete task: %v", err), http.StatusInternalServerError)
		return
	}

	monitoring.Logger.Info("task_closed", "task_id", taskID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"task_id": taskID,
		"status":  "completed",
	})
}

// HandleTaskDelete handles DELETE /a2a/tasks/{taskID}
func (s *AgentServer) HandleTaskDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract task ID from path: /a2a/tasks/{taskID}
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}
	taskID := pathParts[2]
	if taskID == "" {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	if s.taskDB == nil {
		http.Error(w, "Task database not initialized", http.StatusInternalServerError)
		return
	}

	// Verify task exists
	task, err := s.taskDB.GetTask(taskID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get task: %v", err), http.StatusInternalServerError)
		return
	}
	if task == nil {
		http.Error(w, fmt.Sprintf("Task not found: %s", taskID), http.StatusNotFound)
		return
	}

	// Delete task
	if err := s.taskDB.DeleteTask(taskID); err != nil {
		monitoring.Logger.Error("failed_to_delete_task", "task_id", taskID, "error", err.Error())
		http.Error(w, fmt.Sprintf("Failed to delete task: %v", err), http.StatusInternalServerError)
		return
	}

	monitoring.Logger.Info("task_deleted", "task_id", taskID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"task_id": taskID,
		"status":  "deleted",
	})
}

// generateTaskID creates a unique task ID based on project name and hash
func generateTaskID(projectRoot string) string {
	projectName := filepath.Base(projectRoot)

	// Create hash from project + timestamp
	data := fmt.Sprintf("%s-%d", projectRoot, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))

	// Convert to base36 and take first 3 characters
	hashStr := fmt.Sprintf("%x", hash[:4])
	shortID := hashStr[:3]

	return fmt.Sprintf("%s-%s", projectName, shortID)
}
