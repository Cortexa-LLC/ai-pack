package protocol

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// Test constants
const (
	testTaskID     = "task-123"
	testA2AExecute = "a2a.execute"
	testJSONRPCVer = "2.0"

	// Error message templates
	errExpectedTaskID  = "Expected task_id 'task-123', got '%s'"
	errExpectedNoError = "Expected no error, got: %v"
	errExpectedJSONRPC = "Expected jsonrpc '2.0', got '%s'"
)

// JSON-RPC Tests

func TestNewJSONRPCResponse(t *testing.T) {
	result := map[string]string{"status": "ok"}
	resp := NewJSONRPCResponse(1, result)

	if resp.JSONRPC != testJSONRPCVer {
		t.Errorf(errExpectedJSONRPC, resp.JSONRPC)
	}
	if resp.ID != 1 {
		t.Errorf("Expected id 1, got %v", resp.ID)
	}
	if resp.Result == nil {
		t.Error("Expected result to be set")
	}
	if resp.Error != nil {
		t.Error("Expected error to be nil")
	}
}

func TestNewJSONRPCError(t *testing.T) {
	resp := NewJSONRPCError(1, InvalidRequest, "Invalid request", nil)

	if resp.JSONRPC != testJSONRPCVer {
		t.Errorf(errExpectedJSONRPC, resp.JSONRPC)
	}
	if resp.ID != 1 {
		t.Errorf("Expected id 1, got %v", resp.ID)
	}
	if resp.Result != nil {
		t.Error("Expected result to be nil")
	}
	if resp.Error == nil {
		t.Fatal("Expected error to be set")
	}
	if resp.Error.Code != InvalidRequest {
		t.Errorf("Expected error code %d, got %d", InvalidRequest, resp.Error.Code)
	}
	if resp.Error.Message != "Invalid request" {
		t.Errorf("Expected error message 'Invalid request', got '%s'", resp.Error.Message)
	}
}

func TestValidateRequestValid(t *testing.T) {
	req := &JSONRPCRequest{
		JSONRPC: testJSONRPCVer,
		Method:  testA2AExecute,
		ID:      1,
	}

	err := ValidateRequest(req)
	if err != nil {
		t.Errorf("Expected no error for valid request, got: %v", err)
	}
}

func TestValidateRequestInvalidVersion(t *testing.T) {
	req := &JSONRPCRequest{
		JSONRPC: "1.0",
		Method:  testA2AExecute,
		ID:      1,
	}

	err := ValidateRequest(req)
	if err == nil {
		t.Error("Expected error for invalid version")
	}
}

func TestValidateRequestMissingMethod(t *testing.T) {
	req := &JSONRPCRequest{
		JSONRPC: testJSONRPCVer,
		Method:  "",
		ID:      1,
	}

	err := ValidateRequest(req)
	if err == nil {
		t.Error("Expected error for missing method")
	}
}

func TestValidateRequestMissingID(t *testing.T) {
	req := &JSONRPCRequest{
		JSONRPC: testJSONRPCVer,
		Method:  testA2AExecute,
		ID:      nil,
	}

	err := ValidateRequest(req)
	if err == nil {
		t.Error("Expected error for missing ID")
	}
}

func TestErrorCodes(t *testing.T) {
	tests := []struct {
		name string
		code int
	}{
		{"ParseError", ParseError},
		{"InvalidRequest", InvalidRequest},
		{"MethodNotFound", MethodNotFound},
		{"InvalidParams", InvalidParams},
		{"InternalError", InternalError},
	}

	expectedCodes := map[string]int{
		"ParseError":     -32700,
		"InvalidRequest": -32600,
		"MethodNotFound": -32601,
		"InvalidParams":  -32602,
		"InternalError":  -32603,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := expectedCodes[tt.name]
			if tt.code != expected {
				t.Errorf("Expected %s code %d, got %d", tt.name, expected, tt.code)
			}
		})
	}
}

// A2A Protocol Tests

func TestParseExecuteTaskRequestValid(t *testing.T) {
	params := json.RawMessage(`{"role": "engineer", "task": "Create a function"}`)

	req, err := ParseExecuteTaskRequest(params)

	if err != nil {
		t.Fatalf(errExpectedNoError, err)
	}
	if req.Role != "engineer" {
		t.Errorf("Expected role 'engineer', got '%s'", req.Role)
	}
	if req.Task != "Create a function" {
		t.Errorf("Expected task 'Create a function', got '%s'", req.Task)
	}
}

func TestParseExecuteTaskRequestWithOptions(t *testing.T) {
	params := json.RawMessage(`{
		"role": "tester",
		"task": "Test the code",
		"options": {"async": true}
	}`)

	req, err := ParseExecuteTaskRequest(params)

	if err != nil {
		t.Fatalf(errExpectedNoError, err)
	}
	if req.Role != "tester" {
		t.Errorf("Expected role 'tester', got '%s'", req.Role)
	}
	if req.Options == nil {
		t.Fatal("Expected options to be set")
	}
	if async, ok := req.Options["async"].(bool); !ok || !async {
		t.Error("Expected options to contain async: true")
	}
}

func TestParseExecuteTaskRequestInvalidJSON(t *testing.T) {
	params := json.RawMessage(`{invalid json}`)

	_, err := ParseExecuteTaskRequest(params)

	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestParseTaskStatusRequestValid(t *testing.T) {
	params := json.RawMessage(fmt.Sprintf(`{"task_id": "%s"}`, testTaskID))

	req, err := ParseTaskStatusRequest(params)

	if err != nil {
		t.Fatalf(errExpectedNoError, err)
	}
	if req.TaskID != testTaskID {
		t.Errorf(errExpectedTaskID, req.TaskID)
	}
}

func TestParseTaskStatusRequestInvalidJSON(t *testing.T) {
	params := json.RawMessage(`{invalid}`)

	_, err := ParseTaskStatusRequest(params)

	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestExecuteTaskResponseStructure(t *testing.T) {
	resp := ExecuteTaskResponse{
		TaskID:    testTaskID,
		Status:    "queued",
		Message:   "Task queued",
		StreamURL: "/stream/task-123",
		CreatedAt: time.Now(),
	}

	if resp.TaskID != testTaskID {
		t.Errorf("Expected task_id 'task-123', got '%s'", resp.TaskID)
	}
	if resp.Status != "queued" {
		t.Errorf("Expected status 'queued', got '%s'", resp.Status)
	}
	if resp.StreamURL != "/stream/task-123" {
		t.Errorf("Expected stream_url '/stream/task-123', got '%s'", resp.StreamURL)
	}
}

func TestTaskStatusResponseStructure(t *testing.T) {
	now := time.Now()
	resp := TaskStatusResponse{
		TaskID:    testTaskID,
		Role:      "engineer",
		Task:      "Test task",
		Status:    "completed",
		Progress:  1.0,
		CreatedAt: now,
		UpdatedAt: now,
		Result:    "Task completed successfully",
	}

	if resp.TaskID != testTaskID {
		t.Errorf("Expected task_id 'task-123', got '%s'", resp.TaskID)
	}
	if resp.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", resp.Status)
	}
	if resp.Progress != 1.0 {
		t.Errorf("Expected progress 1.0, got %f", resp.Progress)
	}
}

func TestStreamEventStructure(t *testing.T) {
	event := StreamEvent{
		Type:      "status_update",
		TaskID:    testTaskID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"status":   "in_progress",
			"progress": 0.5,
		},
	}

	if event.Type != "status_update" {
		t.Errorf("Expected type 'status_update', got '%s'", event.Type)
	}
	if event.TaskID != testTaskID {
		t.Errorf("Expected task_id 'task-123', got '%s'", event.TaskID)
	}
	if event.Data == nil {
		t.Fatal("Expected data to be set")
	}
}

func TestDiscoveryResponseStructure(t *testing.T) {
	resp := DiscoveryResponse{
		Name:        "AI-Pack Agent Server",
		Version:     "2.0.0",
		Description: "Multi-agent orchestration server",
		Capabilities: AgentCapabilities{
			Streaming:     true,
			Parallel:      true,
			MaxConcurrent: 5,
			SupportedModels: []string{
				"claude-3-5-sonnet-20241022",
			},
		},
		Agents: []AgentDescription{
			{
				Role:        "engineer",
				Description: "Software engineer",
				Tools:       []string{"read", "write", "bash"},
				Timeout:     "10min",
			},
		},
		Endpoints: map[string]Endpoint{
			"execute": {
				Path:        "/a2a/execute",
				Method:      "POST",
				Description: "Execute agent task",
			},
		},
	}

	if resp.Name != "AI-Pack Agent Server" {
		t.Errorf("Expected name 'AI-Pack Agent Server', got '%s'", resp.Name)
	}
	if !resp.Capabilities.Streaming {
		t.Error("Expected streaming capability to be true")
	}
	if len(resp.Agents) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(resp.Agents))
	}
	if len(resp.Endpoints) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(resp.Endpoints))
	}
}

func TestJSONRPCRequestMarshaling(t *testing.T) {
	req := JSONRPCRequest{
		JSONRPC: testJSONRPCVer,
		Method:  testA2AExecute,
		Params:  json.RawMessage(`{"role":"engineer"}`),
		ID:      1,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	var decoded JSONRPCRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if decoded.Method != req.Method {
		t.Errorf("Expected method '%s', got '%s'", req.Method, decoded.Method)
	}
}

func TestJSONRPCResponseMarshaling(t *testing.T) {
	resp := NewJSONRPCResponse(1, map[string]string{"status": "ok"})

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	var decoded JSONRPCResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if decoded.JSONRPC != testJSONRPCVer {
		t.Errorf("Expected jsonrpc '2.0', got '%s'", decoded.JSONRPC)
	}
}
