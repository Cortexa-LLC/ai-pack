package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/protocol"
)

// A2A Protocol Handlers
// Implements A2A protocol endpoints using JSON-RPC 2.0

// handleA2ADiscovery handles the /a2a/discovery endpoint
func (s *AgentServer) handleA2ADiscovery(w http.ResponseWriter, r *http.Request) {
	// Discovery can be GET or POST (JSON-RPC)
	if r.Method == http.MethodGet {
		s.handleDiscoveryGET(w, r)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
	var req protocol.JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := protocol.NewJSONRPCError(nil, protocol.ParseError, "Parse error", err.Error())
		s.sendJSONRPCResponse(w, response)
		return
	}

	if err := protocol.ValidateRequest(&req); err != nil {
		response := protocol.NewJSONRPCError(req.ID, protocol.InvalidRequest, "Invalid request", err.Error())
		s.sendJSONRPCResponse(w, response)
		return
	}

	if req.Method != "discovery" && req.Method != "a2a.discovery" {
		response := protocol.NewJSONRPCError(req.ID, protocol.MethodNotFound, "Method not found", req.Method)
		s.sendJSONRPCResponse(w, response)
		return
	}

	discovery := s.getDiscoveryResponse()
	response := protocol.NewJSONRPCResponse(req.ID, discovery)
	s.sendJSONRPCResponse(w, response)
}

// handleA2AExecute handles the /a2a/execute endpoint (JSON-RPC)
func (s *AgentServer) handleA2AExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := protocol.NewJSONRPCError(nil, protocol.ParseError, "Parse error", err.Error())
		s.sendJSONRPCResponse(w, response)
		return
	}

	if err := protocol.ValidateRequest(&req); err != nil {
		response := protocol.NewJSONRPCError(req.ID, protocol.InvalidRequest, "Invalid request", err.Error())
		s.sendJSONRPCResponse(w, response)
		return
	}

	if req.Method != "execute" && req.Method != "a2a.execute" {
		response := protocol.NewJSONRPCError(req.ID, protocol.MethodNotFound, "Method not found", req.Method)
		s.sendJSONRPCResponse(w, response)
		return
	}

	// Parse execute task request
	execReq, err := protocol.ParseExecuteTaskRequest(req.Params)
	if err != nil {
		response := protocol.NewJSONRPCError(req.ID, protocol.InvalidParams, "Invalid params", err.Error())
		s.sendJSONRPCResponse(w, response)
		return
	}

	ctx := context.Background()
	monitoring.Logger.Info("a2a_execute_request", "role", execReq.Role, "task", execReq.Task)
	monitoring.LogTaskSpawned(ctx, "", execReq.Role, execReq.Task)

	// Execute task (reuse existing spawn logic)
	result, err := s.spawnAgentTask(execReq.Role, execReq.Task)
	if err != nil {
		monitoring.Logger.Error("a2a_execute_failed", "role", execReq.Role, "error", err)
		response := protocol.NewJSONRPCError(req.ID, protocol.InternalError, "Execution failed", err.Error())
		s.sendJSONRPCResponse(w, response)
		return
	}

	// Return JSON-RPC success response
	response := protocol.NewJSONRPCResponse(req.ID, result)
	s.sendJSONRPCResponse(w, response)
}

// handleA2AStatus handles the /a2a/status endpoint (JSON-RPC)
func (s *AgentServer) handleA2AStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := protocol.NewJSONRPCError(nil, protocol.ParseError, "Parse error", err.Error())
		s.sendJSONRPCResponse(w, response)
		return
	}

	if err := protocol.ValidateRequest(&req); err != nil {
		response := protocol.NewJSONRPCError(req.ID, protocol.InvalidRequest, "Invalid request", err.Error())
		s.sendJSONRPCResponse(w, response)
		return
	}

	if req.Method != "status" && req.Method != "a2a.status" {
		response := protocol.NewJSONRPCError(req.ID, protocol.MethodNotFound, "Method not found", req.Method)
		s.sendJSONRPCResponse(w, response)
		return
	}

	// Parse status request
	statusReq, err := protocol.ParseTaskStatusRequest(req.Params)
	if err != nil {
		response := protocol.NewJSONRPCError(req.ID, protocol.InvalidParams, "Invalid params", err.Error())
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
		Agents: []protocol.AgentDescription{
			{
				Role:        "engineer",
				Description: "Implementation specialist following TDD practices",
				Tools:       []string{"read", "write", "edit", "bash", "grep", "glob"},
				Timeout:     "10min",
			},
			{
				Role:        "tester",
				Description: "Testing specialist targeting >80% coverage",
				Tools:       []string{"read", "write", "edit", "bash", "grep", "glob"},
				Timeout:     "10min",
			},
			{
				Role:        "reviewer",
				Description: "Code review specialist for quality and security",
				Tools:       []string{"read", "grep", "glob", "bash"},
				Timeout:     "8min",
			},
		},
		Endpoints: map[string]protocol.Endpoint{
			"discovery": {
				Path:        "/a2a/discovery",
				Method:      "GET or POST",
				Description: "Get agent server capabilities and available agents",
			},
			"execute": {
				Path:        "/a2a/execute",
				Method:      "POST",
				Description: "Execute a task with specified agent role (JSON-RPC 2.0)",
			},
			"status": {
				Path:        "/a2a/status",
				Method:      "POST",
				Description: "Get task execution status (JSON-RPC 2.0)",
			},
			"stream": {
				Path:        "/stream/:task_id",
				Method:      "GET",
				Description: "Stream real-time task execution progress (SSE)",
			},
		},
	}
}

// sendJSONRPCResponse sends a JSON-RPC response
func (s *AgentServer) sendJSONRPCResponse(w http.ResponseWriter, response *protocol.JSONRPCResponse) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		monitoring.Logger.Error("jsonrpc_response_encode_error", "error", err)
	}
}
