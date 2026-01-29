package protocol

import (
	"encoding/json"
	"time"
)

// A2A Protocol Types
// Based on A2A protocol specification

// DiscoveryResponse represents agent capabilities
type DiscoveryResponse struct {
	Name         string              `json:"name"`
	Version      string              `json:"version"`
	Description  string              `json:"description"`
	Capabilities AgentCapabilities   `json:"capabilities"`
	Agents       []AgentDescription  `json:"agents"`
	Endpoints    map[string]Endpoint `json:"endpoints"`
}

// AgentCapabilities defines what the server can do
type AgentCapabilities struct {
	Streaming       bool     `json:"streaming"`
	Parallel        bool     `json:"parallel"`
	MaxConcurrent   int      `json:"max_concurrent"`
	SupportedModels []string `json:"supported_models"`
}

// AgentDescription describes an available agent type
type AgentDescription struct {
	Role        string   `json:"role"`
	Description string   `json:"description"`
	Tools       []string `json:"tools"`
	Timeout     string   `json:"timeout"`
}

// Endpoint describes an API endpoint
type Endpoint struct {
	Path        string `json:"path"`
	Method      string `json:"method"`
	Description string `json:"description"`
}

// ExecuteTaskRequest represents a task execution request
type ExecuteTaskRequest struct {
	Role        string                 `json:"role"`
	Task        string                 `json:"task"`
	ProjectRoot string                 `json:"project_root,omitempty"` // Project root directory for Beads integration
	Options     map[string]interface{} `json:"options,omitempty"`
}

// ExecuteTaskResponse represents a task execution response
type ExecuteTaskResponse struct {
	TaskID    string    `json:"task_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	StreamURL string    `json:"stream_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// TaskStatusRequest represents a task status query
type TaskStatusRequest struct {
	TaskID string `json:"task_id"`
}

// TaskStatusResponse represents task status information
type TaskStatusResponse struct {
	TaskID      string                 `json:"task_id"`
	Role        string                 `json:"role"`
	Task        string                 `json:"task"`
	Status      string                 `json:"status"`
	Progress    float64                `json:"progress"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Result      string                 `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// StreamEvent represents a real-time progress event
type StreamEvent struct {
	Type      string                 `json:"type"`
	TaskID    string                 `json:"task_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// ParseExecuteTaskRequest parses JSON-RPC params into ExecuteTaskRequest
func ParseExecuteTaskRequest(params json.RawMessage) (*ExecuteTaskRequest, error) {
	var req ExecuteTaskRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

// ParseTaskStatusRequest parses JSON-RPC params into TaskStatusRequest
func ParseTaskStatusRequest(params json.RawMessage) (*TaskStatusRequest, error) {
	var req TaskStatusRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}
	return &req, nil
}
