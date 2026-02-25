package mcp

import (
	"bufio"
	"encoding/json"
	"io"
)

// ToolHandler is a function that takes a tool call (name, arguments) and returns a result or error.
type ToolHandler func(req *ToolCallRequest) (any, error)

type Server struct {
	tools    []Tool
	handlers map[string]ToolHandler
	in       *bufio.Reader
	out      io.Writer
}

func NewServer(tools []Tool, handlers map[string]ToolHandler, in *bufio.Reader, out io.Writer) *Server {
	return &Server{
		tools:    tools,
		handlers: handlers,
		in:       in,
		out:      out,
	}
}

// Serve runs the main MCP server loop: reads requests over stdin, dispatches to handlers, writes responses.
func (s *Server) Serve() error {
	// Announce tools as MCP capability
	caps := map[string]any{
		"jsonrpc": "2.0",
		"method":  "capabilities",
		"params": map[string]any{
			"tools": s.tools,
		},
	}
	json.NewEncoder(s.out).Encode(caps)

	for {
		line, err := s.in.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		var req struct {
			JSONRPC string                 `json:"jsonrpc"`
			ID      interface{}            `json:"id"`
			Method  string                 `json:"method"`
			Params  map[string]interface{} `json:"params"`
		}
		if err := json.Unmarshal(line, &req); err != nil {
			json.NewEncoder(s.out).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      nil,
				"error":   map[string]any{"code": -32600, "message": "Malformed JSON"},
			})
			continue
		}
		if req.Method != "callTool" {
			json.NewEncoder(s.out).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"error":   map[string]any{"code": -32601, "message": "Unknown method"},
			})
			continue
		}
		name, _ := req.Params["name"].(string)
		args, _ := req.Params["arguments"].(map[string]interface{})
		handler, ok := s.handlers[name]
		if !ok {
			json.NewEncoder(s.out).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"error":   map[string]any{"code": -32601, "message": "Unknown tool: " + name},
			})
			continue
		}
		resp, err := handler(&ToolCallRequest{Name: name, Arguments: args})
		if err != nil {
			json.NewEncoder(s.out).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"error":   map[string]any{"code": -32000, "message": err.Error()},
			})
			continue
		}
		json.NewEncoder(s.out).Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"result":  resp,
		})
	}
}
