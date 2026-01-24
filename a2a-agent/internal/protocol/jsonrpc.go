package protocol

import (
	"encoding/json"
	"fmt"
)

// JSON-RPC 2.0 implementation
// Spec: https://www.jsonrpc.org/specification

const (
	JSONRPC2Version = "2.0"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

// RPCError represents a JSON-RPC 2.0 error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Standard JSON-RPC error codes
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// NewJSONRPCResponse creates a successful response
func NewJSONRPCResponse(id interface{}, result interface{}) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: JSONRPC2Version,
		Result:  result,
		ID:      id,
	}
}

// NewJSONRPCError creates an error response
func NewJSONRPCError(id interface{}, code int, message string, data interface{}) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: JSONRPC2Version,
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
		ID: id,
	}
}

// ValidateRequest validates a JSON-RPC 2.0 request
func ValidateRequest(req *JSONRPCRequest) error {
	if req.JSONRPC != JSONRPC2Version {
		return fmt.Errorf("invalid jsonrpc version: %s (expected 2.0)", req.JSONRPC)
	}
	if req.Method == "" {
		return fmt.Errorf("method is required")
	}
	if req.ID == nil {
		return fmt.Errorf("id is required")
	}
	return nil
}
