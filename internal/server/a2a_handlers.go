package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cortexa-llc/ai-pack/internal/constants"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/internal/protocol"
)

// Error message constants
const (
	errMethodNotAllowed = "Method not allowed"
	errParseError       = "Parse error"
	errInvalidRequest   = "Invalid request"
	errMethodNotFound   = "Method not found"
	errInvalidParams    = "Invalid params"
)

// A2A Protocol Handlers
// Implements A2A protocol endpoints using JSON-RPC 2.0

// decodeAndValidateJSONRPC is a helper that decodes a JSON-RPC request from the
// given reader, validates it, and checks that the method matches one of the
// provided allowed methods. On success it returns the decoded request and true.
// On failure it writes the appropriate JSON-RPC error response and returns false.
func (s *AgentServer) decodeAndValidateJSONRPC(w http.ResponseWriter, body io.Reader, allowedMethods ...string) (protocol.JSONRPCRequest, bool) {
	var req protocol.JSONRPCRequest
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		response := protocol.NewJSONRPCError(nil, protocol.ParseError, errParseError, err.Error())
		s.sendJSONRPCResponse(w, response)
		return req, false
	}

	if err := protocol.ValidateRequest(&req); err != nil {
		response := protocol.NewJSONRPCError(req.ID, protocol.InvalidRequest, errInvalidRequest, err.Error())
		s.sendJSONRPCResponse(w, response)
		return req, false
	}

	for _, m := range allowedMethods {
		if req.Method == m {
			return req, true
		}
	}

	response := protocol.NewJSONRPCError(req.ID, protocol.MethodNotFound, errMethodNotFound, req.Method)
	s.sendJSONRPCResponse(w, response)
	return req, false
}

// handleAgentCard serves GET /.well-known/agent.json (A2A AgentCard).
// This endpoint is publicly accessible (exempt from API key middleware).
func (s *AgentServer) handleAgentCard(w http.ResponseWriter, r *http.Request) {
	discovery := s.getDiscoveryResponse()
	result := struct {
		*protocol.DiscoveryResponse
		Authentication protocol.AgentAuthentication `json:"authentication"`
	}{
		DiscoveryResponse: discovery,
		Authentication: protocol.AgentAuthentication{
			Schemes: []string{"bearer", "api-key"},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleA2ADiscovery handles the /a2a/discovery endpoint
func (s *AgentServer) handleA2ADiscovery(w http.ResponseWriter, r *http.Request) {
	// Discovery can be GET or POST (JSON-RPC)
	if r.Method == http.MethodGet {
		s.handleDiscoveryGET(w, r)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	s.handleDiscoveryRPC(w, r)
}

// handleDiscoveryGET handles GET /a2a/discovery (simple JSON response)
func (s *AgentServer) handleDiscoveryGET(w http.ResponseWriter, r *http.Request) {
	discovery := s.getDiscoveryResponse()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(discovery)
}

// handleDiscoveryRPC handles POST /a2a/discovery (JSON-RPC)
func (s *AgentServer) handleDiscoveryRPC(w http.ResponseWriter, r *http.Request) {
	req, ok := s.decodeAndValidateJSONRPC(w, r.Body, constants.MethodDiscovery, constants.MethodA2ADiscovery)
	if !ok {
		return
	}

	discovery := s.getDiscoveryResponse()
	response := protocol.NewJSONRPCResponse(req.ID, discovery)
	s.sendJSONRPCResponse(w, response)
}

// handleA2AExecute handles the /a2a/execute endpoint (JSON-RPC)
func (s *AgentServer) handleA2AExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	req, ok := s.decodeAndValidateJSONRPC(w, r.Body, constants.MethodExecute, constants.MethodA2AExecute)
	if !ok {
		return
	}

	// Parse execute task request
	execReq, err := protocol.ParseExecuteTaskRequest(req.Params)
	if err != nil {
		response := protocol.NewJSONRPCError(req.ID, protocol.InvalidParams, errInvalidParams, err.Error())
		s.sendJSONRPCResponse(w, response)
		return
	}

	ctx := context.Background()
	// Validate projectRoot: must be non-empty, absolute, and free of traversal sequences.
	if execReq.ProjectRoot == "" {
		response := protocol.NewJSONRPCError(req.ID, protocol.InvalidParams, errInvalidParams, "project_root is required")
		s.sendJSONRPCResponse(w, response)
		return
	}
	cleanedRoot := filepath.Clean(execReq.ProjectRoot)
	if !filepath.IsAbs(cleanedRoot) {
		response := protocol.NewJSONRPCError(req.ID, protocol.InvalidParams, errInvalidParams, "project_root must be an absolute path")
		s.sendJSONRPCResponse(w, response)
		return
	}
	execReq.ProjectRoot = cleanedRoot

	// Validate role: must not contain path separators to prevent path traversal.
	if strings.ContainsAny(execReq.Role, "/\\") {
		response := protocol.NewJSONRPCError(req.ID, protocol.InvalidParams, errInvalidParams, "role must not contain path separators")
		s.sendJSONRPCResponse(w, response)
		return
	}

	monitoring.Logger.Info("a2a_execute_request", "role", execReq.Role, "task", execReq.Task, "project_root", execReq.ProjectRoot)
	monitoring.LogTaskSpawned(ctx, "", execReq.Role, execReq.Task)

	// Execute task (reuse existing spawn logic)
	result, err := s.spawnAgentTask(execReq.Role, execReq.Task, execReq.ProjectRoot)
	if err != nil {
		monitoring.Logger.Error("a2a_execute_failed", "role", execReq.Role, "error", err)
		if errors.Is(err, ErrTaskQueueFull) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "5")
			w.WriteHeader(http.StatusTooManyRequests)
			response := protocol.NewJSONRPCError(req.ID, protocol.InternalError, "Too Many Requests", err.Error())
			s.sendJSONRPCResponse(w, response)
			return
		}
		response := protocol.NewJSONRPCError(req.ID, protocol.InternalError, "Execution failed", err.Error())
		s.sendJSONRPCResponse(w, response)
		return
	}

	// Return JSON-RPC success response
	response := protocol.NewJSONRPCResponse(req.ID, result)
	s.sendJSONRPCResponse(w, response)
}

// handleA2AStatus handles the /a2a/status endpoint (supports both GET and POST)
func (s *AgentServer) handleA2AStatus(w http.ResponseWriter, r *http.Request) {
	// GET /a2a/status/:task_id - simple JSON response
	if r.Method == http.MethodGet {
		s.handleStatusGET(w, r)
		return
	}

	// POST /a2a/status - JSON-RPC format
	if r.Method != http.MethodPost {
		http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	req, ok := s.decodeAndValidateJSONRPC(w, r.Body, constants.MethodStatus, constants.MethodA2AStatus)
	if !ok {
		return
	}

	// Parse status request
	statusReq, err := protocol.ParseTaskStatusRequest(req.Params)
	if err != nil {
		response := protocol.NewJSONRPCError(req.ID, protocol.InvalidParams, errInvalidParams, err.Error())
		s.sendJSONRPCResponse(w, response)
		return
	}

	// Get task status
	status, err := s.getTaskStatus(statusReq.TaskID)
	if err != nil {
		response := protocol.NewJSONRPCError(req.ID, protocol.InternalError, "Failed to get status", err.Error())
		s.sendJSONRPCResponse(w, response)
		return
	}

	response := protocol.NewJSONRPCResponse(req.ID, status)
	s.sendJSONRPCResponse(w, response)
}

// getDiscoveryResponse builds the discovery response
func (s *AgentServer) getDiscoveryResponse() *protocol.DiscoveryResponse {
	// Dynamically discover available agents from filesystem
	agents := s.discoverAvailableAgents()

	return &protocol.DiscoveryResponse{
		Name:        "ai-pack-agent-server",
		Version:     Version,
		Description: "AI-Pack Agent Server - Multi-agent task execution with A2A protocol support",
		Capabilities: protocol.AgentCapabilities{
			Streaming:       true,
			Parallel:        true,
			MaxConcurrent:   s.maxConcurrent,
			SupportedModels: []string{s.model},
		},
		Agents: agents,
		Endpoints: map[string]protocol.Endpoint{
			"discovery": {
				Path:        "/a2a/discovery",
				Method:      "GET or POST",
				Description: "Get agent server capabilities and available agents",
			},
			constants.MethodExecute: {
				Path:        "/a2a/execute",
				Method:      "POST",
				Description: "Execute a task with specified agent role (JSON-RPC 2.0)",
			},
			"status": {
				Path:        "/a2a/status",
				Method:      "POST",
				Description: "Get task execution status (JSON-RPC 2.0)",
			},
			constants.MethodStream: {
				Path:        "/stream/:task_id",
				Method:      "GET",
				Description: "Stream real-time task execution progress (SSE)",
			},
		},
	}
}

// discoverAvailableAgents scans the agents directory and builds descriptions
func (s *AgentServer) discoverAvailableAgents() []protocol.AgentDescription {
	var agents []protocol.AgentDescription

	// Build set of all agent roles (combining framework and project overrides)
	agentRoles := make(map[string]bool)

	frameworkAgentsDirs := []string{
		s.rootDir + "/roles",
	}

	// Project overrides
	projectAgentsDir := s.rootDir + "/.ai/agents"

	// Check framework agents (try all possible locations)
	for _, frameworkAgentsDir := range frameworkAgentsDirs {
		if entries, err := os.ReadDir(frameworkAgentsDir); err == nil {
			monitoring.Logger.Info("found_agents_directory", "path", frameworkAgentsDir)
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}

				name := entry.Name()
				var role string

				if strings.HasSuffix(name, ".yml") {
					role = strings.TrimSuffix(name, ".yml")
				} else if strings.HasSuffix(name, ".yaml") {
					role = strings.TrimSuffix(name, ".yaml")
				} else {
					continue
				}

				if role != "" {
					monitoring.Logger.Info("discovered_agent", "role", role, "file", name)
					agentRoles[role] = true
				}
			}
		} else {
			monitoring.Logger.Debug("agents_dir_not_found", "path", frameworkAgentsDir, "error", err)
		}
	}

	// Check project overrides
	if entries, err := os.ReadDir(projectAgentsDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			var role string

			if strings.HasSuffix(name, ".md") {
				role = strings.TrimSuffix(name, ".md")
			} else {
				continue
			}

			if role != "" {
				agentRoles[role] = true
			}
		}
	}

	// Load each agent config and build description
	for role := range agentRoles {
		config, err := s.loadAgentConfig(role, s.rootDir)
		if err != nil {
			monitoring.Logger.Warn("failed_to_load_agent_config_for_discovery", "role", role, "error", err)
			continue
		}

		// Provide basic agent description
		// Full metadata is in the role .md file which is loaded during execution
		agents = append(agents, protocol.AgentDescription{
			Role:        role,
			Description: fmt.Sprintf("%s agent (model: %s, tier: %s)", role, config.Model, config.Tier),
			Tools:       []string{"bash", "read", "write", "edit", "grep", "glob"},
			Timeout:     "30m",
		})
	}

	return agents
}

// sendJSONRPCResponse sends a JSON-RPC response
func (s *AgentServer) sendJSONRPCResponse(w http.ResponseWriter, response *protocol.JSONRPCResponse) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		monitoring.Logger.Error("jsonrpc_response_encode_error", "error", err)
	}
}

// handleStatusGET handles GET /a2a/status/:task_id (simple REST endpoint)
func (s *AgentServer) handleStatusGET(w http.ResponseWriter, r *http.Request) {
	// Extract task ID from path: /a2a/status/:task_id
	path := strings.TrimPrefix(r.URL.Path, "/a2a/status/")
	taskID := path

	if taskID == "" || taskID == "/a2a/status" {
		http.Error(w, "Task ID required", http.StatusBadRequest)
		return
	}

	// Get task status
	status, err := s.getTaskStatus(taskID)
	if err != nil {
		// Return JSON error response instead of plain text
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "task_not_found",
			"message": fmt.Sprintf("Task not found: %s", taskID),
			"task_id": taskID,
		})
		return
	}

	// Return simple JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// HandleTasksList returns all active tasks machine-wide
func (s *AgentServer) HandleTasksList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	// Build list of all tasks
	type TaskInfo struct {
		TaskID      string `json:"task_id"`
		BeadsTaskID string `json:"beads_task_id,omitempty"`
		Status      string `json:"status"`
		Role        string `json:"role"`
		Description string `json:"description"`
		ProjectRoot string `json:"project_root,omitempty"`
		Error       string `json:"error,omitempty"`
	}

	tasksMap := make(map[string]TaskInfo)

	// First, get all active tasks from memory
	s.mu.RLock()
	for _, execution := range s.activeTasks {
		task := TaskInfo{
			TaskID:      execution.TaskID,
			Status:      execution.Status,
			Role:        execution.Role,
			Description: execution.Task, // Task field contains the description
			Error:       execution.Error,
			ProjectRoot: execution.ProjectRoot,
		}
		if execution.metadata != nil {
			task.BeadsTaskID = execution.metadata["beads_task_id"]
		}

		tasksMap[task.TaskID] = task
	}
	s.mu.RUnlock()

	// Get all project roots to scan (server root + registered projects)
	projectRoots := s.GetProjectRoots()

	// Build set of beads IDs already covered by active in-memory tasks
	activeBeadsIDs := make(map[string]bool)
	for _, task := range tasksMap {
		if task.BeadsTaskID != "" {
			activeBeadsIDs[task.BeadsTaskID] = true
		}
	}

	// Then, get beads tasks from each project using bd list
	beadsClient := s.beadsClient
	for _, projectRoot := range projectRoots {
		beadsTasks, err := beadsClient.ListAllTasksFromDir(projectRoot)
		if err != nil {
			continue // Skip if can't list tasks from this project
		}

		// Convert beads tasks to TaskInfo
		for _, beadsTask := range beadsTasks {
			taskID := beadsTask.ID
			// Skip if already in tasks map by full ID or by beads ID
			if _, exists := tasksMap[taskID]; exists || activeBeadsIDs[taskID] {
				continue
			}

			// Map beads status to agent status
			status := "queued"
			switch beadsTask.Status {
			case "in_progress":
				status = "in_progress"
			case "closed", "done":
				// A task closed in Beads is always completed, regardless of execution outcome
				status = constants.StatusCompleted
			case "open":
				status = "queued"
			}

			task := TaskInfo{
				TaskID:      taskID,
				BeadsTaskID: taskID,
				Status:      status,
				Role:        "beads-task",
				Description: beadsTask.Title,
				ProjectRoot: projectRoot,
			}

			tasksMap[taskID] = task
		}
	}

	// Convert map to slice
	var tasks []TaskInfo
	for _, task := range tasksMap {
		tasks = append(tasks, task)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tasks": tasks,
		"count": len(tasks),
	})
}

// HandleCancelTask cancels a running task
func (s *AgentServer) HandleCancelTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	// Extract task ID from URL path: /a2a/cancel/{taskID}
	taskID := strings.TrimPrefix(r.URL.Path, "/a2a/cancel/")
	if taskID == "" {
		http.Error(w, "task ID required", http.StatusBadRequest)
		return
	}

	// Cancel the task
	err := s.CancelTask(taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Task cancelled",
		"task_id": taskID,
	})
}

// HandleRetryTask retries a failed task
func (s *AgentServer) HandleRetryTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	// Extract task ID from URL path: /a2a/retry/{taskID}
	taskID := strings.TrimPrefix(r.URL.Path, "/a2a/retry/")
	if taskID == "" {
		http.Error(w, "task ID required", http.StatusBadRequest)
		return
	}

	// Get the task info
	taskInfo, err := s.loadTaskStatusFromDisk(taskID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Task not found: %v", err), http.StatusNotFound)
		return
	}

	// Extract Beads task ID from metadata (this is the correct ID to use for retry)
	// The taskID parameter might be a timestamped execution folder, but we need the Beads ID
	beadsTaskID := taskID
	if btid, ok := taskInfo.Metadata["beads_task_id"].(string); ok && btid != "" {
		beadsTaskID = btid
	}

	// Extract project root from metadata
	projectRoot := ""
	if pr, ok := taskInfo.Metadata["project_root"].(string); ok {
		projectRoot = pr
	}

	// Mark the old execution as superseded before retrying
	// This helps track execution chains and prevents stale status display
	if err := s.markExecutionAsSuperseded(taskID, projectRoot, "retry"); err != nil {
		monitoring.Logger.Warn("failed_to_mark_execution_superseded",
			"task_id", taskID,
			"error", err)
		// Don't fail the retry - this is just metadata cleanup
	}

	// Spawn a new agent task with the Beads task ID (not the timestamped execution folder)
	newTaskResponse, err := s.spawnAgentTask(taskInfo.Role, beadsTaskID, projectRoot)
	if err != nil {
		if errors.Is(err, ErrTaskQueueFull) {
			w.Header().Set("Retry-After", "5")
			http.Error(w, "Too Many Requests: "+err.Error(), http.StatusTooManyRequests)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to retry task: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"message":     "Task retried successfully",
		"task_id":     taskID,
		"new_task_id": newTaskResponse.TaskID,
	})
}

// HandleStartTask starts an agent for a Beads task
func (s *AgentServer) HandleStartTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	// Extract task ID from URL path: /a2a/start/{taskID}
	taskID := strings.TrimPrefix(r.URL.Path, "/a2a/start/")
	if taskID == "" {
		http.Error(w, "task ID required", http.StatusBadRequest)
		return
	}

	// Parse request body for role and project_root (optional)
	var req struct {
		Role        string `json:"role"`
		ProjectRoot string `json:"project_root"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.ProjectRoot = ""
	}

	// Role is required
	if req.Role == "" {
		http.Error(w, "role is required", http.StatusBadRequest)
		return
	}

	// Use server root if no project root specified
	if req.ProjectRoot == "" {
		req.ProjectRoot = s.rootDir
	}

	// Check if task is already active
	s.mu.RLock()
	for activeTaskID := range s.activeTasks {
		// Check for exact match or timestamped variant
		if activeTaskID == taskID || strings.HasPrefix(activeTaskID, taskID+"-") {
			s.mu.RUnlock()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("Task %s already has an active agent", taskID),
			})
			return
		}
	}
	s.mu.RUnlock()

	// Check if task packet exists
	_, taskPacketPath, _, _, err := s.beadsClient.GetTaskDescriptionFromDir(taskID, req.ProjectRoot)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to get task info: %v", err),
		})
		return
	}

	// Require a task packet — without one the agent has no contract to work from.
	if taskPacketPath == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("No task packet found for task %s. Create a task packet before starting an agent.", taskID),
		})
		return
	}

	roleToSpawn := req.Role
	monitoring.Logger.Info("start_task_with_packet", "task_id", taskID, "spawning", roleToSpawn, "packet_path", taskPacketPath)

	// Spawn agent for this Beads task
	response, err := s.spawnAgentTask(roleToSpawn, taskID, req.ProjectRoot)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, ErrTaskQueueFull) {
			w.Header().Set("Retry-After", "5")
			w.WriteHeader(http.StatusTooManyRequests)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to start agent: %v", err),
		})
		return
	}

	message := fmt.Sprintf("Agent (%s) started for task %s", roleToSpawn, taskID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": message,
		"task_id": response.TaskID,
		"role":    roleToSpawn,
		"status":  response.Status,
	})
}

// HandleResumeTask resumes a paused task from its checkpoint with a new token budget.
func (s *AgentServer) HandleResumeTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	// Extract task ID from URL path: /a2a/resume/{taskID}
	taskID := strings.TrimPrefix(r.URL.Path, "/a2a/resume/")
	if taskID == "" {
		http.Error(w, "task ID required", http.StatusBadRequest)
		return
	}

	// Parse optional body for new budget
	var req struct {
		NewBudget int64 `json:"new_budget"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Default: no additional budget specified, use 0 (unlimited)
		req.NewBudget = 0
	}

	// Load task info to find project root
	taskInfo, err := s.loadTaskStatusFromDisk(taskID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Task not found: %v", err), http.StatusNotFound)
		return
	}

	// Verify task is paused
	if taskInfo.Status != constants.StatusPaused {
		http.Error(w, fmt.Sprintf("Task is not paused (status: %s)", taskInfo.Status), http.StatusBadRequest)
		return
	}

	// Determine project root
	projectRoot := s.rootDir
	if pr, ok := taskInfo.Metadata["project_root"].(string); ok && pr != "" {
		projectRoot = pr
	}

	// Load checkpoint
	cp, err := loadCheckpoint(projectRoot, taskID)
	if err != nil {
		http.Error(w, fmt.Sprintf("No checkpoint found for task: %v", err), http.StatusNotFound)
		return
	}

	// Launch resume in background
	go s.resumeFromCheckpoint(taskID, projectRoot, cp, req.NewBudget)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"message":    "Task resuming from checkpoint",
		"task_id":    taskID,
		"budget_was": cp.BudgetUsed,
		"new_budget": req.NewBudget,
	})
}
