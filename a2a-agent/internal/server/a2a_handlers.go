package server

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/protocol"
)

// Error message constants
const (
	errMethodNotAllowed = "Method not allowed"
	errParseError       = "Parse error"
	errInvalidRequest   = "Invalid request"
	errMethodNotFound   = "Method not found"
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
	var req protocol.JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := protocol.NewJSONRPCError(nil, protocol.ParseError, errParseError, err.Error())
		s.sendJSONRPCResponse(w, response)
		return
	}

	if err := protocol.ValidateRequest(&req); err != nil {
		response := protocol.NewJSONRPCError(req.ID, protocol.InvalidRequest, errInvalidRequest, err.Error())
		s.sendJSONRPCResponse(w, response)
		return
	}

	if req.Method != "discovery" && req.Method != "a2a.discovery" {
		response := protocol.NewJSONRPCError(req.ID, protocol.MethodNotFound, errMethodNotFound, req.Method)
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
		http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	var req protocol.JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := protocol.NewJSONRPCError(nil, protocol.ParseError, "errParseError", err.Error())
		s.sendJSONRPCResponse(w, response)
		return
	}

	if err := protocol.ValidateRequest(&req); err != nil {
		response := protocol.NewJSONRPCError(req.ID, protocol.InvalidRequest, "errInvalidRequest", err.Error())
		s.sendJSONRPCResponse(w, response)
		return
	}

	if req.Method != "execute" && req.Method != "a2a.execute" {
		response := protocol.NewJSONRPCError(req.ID, protocol.MethodNotFound, "errMethodNotFound", req.Method)
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
		http.Error(w, "errMethodNotAllowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := protocol.NewJSONRPCError(nil, protocol.ParseError, "errParseError", err.Error())
		s.sendJSONRPCResponse(w, response)
		return
	}

	if err := protocol.ValidateRequest(&req); err != nil {
		response := protocol.NewJSONRPCError(req.ID, protocol.InvalidRequest, "errInvalidRequest", err.Error())
		s.sendJSONRPCResponse(w, response)
		return
	}

	if req.Method != "status" && req.Method != "a2a.status" {
		response := protocol.NewJSONRPCError(req.ID, protocol.MethodNotFound, "errMethodNotFound", req.Method)
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

// discoverAvailableAgents scans the agents directory and builds descriptions
func (s *AgentServer) discoverAvailableAgents() []protocol.AgentDescription {
	var agents []protocol.AgentDescription

	// Build set of all agent roles (combining framework and project overrides)
	agentRoles := make(map[string]bool)

	// Try multiple locations for agent configs:
	// 1. Production: .ai-pack/agents/ (when deployed as submodule)
	// 2. Development: ../agents/ (when running from a2a-agent dir)
	// 3. Development: agents/ (when running from repo root)
	frameworkAgentsDirs := []string{
		s.rootDir + "/.ai-pack/agents",
		s.rootDir + "/../agents",
		s.rootDir + "/agents",
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

			if strings.HasSuffix(name, ".yml") {
				role = strings.TrimSuffix(name, ".yml")
			} else if strings.HasSuffix(name, ".yaml") {
				role = strings.TrimSuffix(name, ".yaml")
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
		config, err := s.loadAgentConfig(role)
		if err != nil {
			monitoring.Logger.Warn("failed_to_load_agent_config_for_discovery", "role", role, "error", err)
			continue
		}

		agents = append(agents, protocol.AgentDescription{
			Role:        config.Name,
			Description: config.Description,
			Tools:       config.Tools,
			Timeout:     config.Delegation.Timeout,
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
