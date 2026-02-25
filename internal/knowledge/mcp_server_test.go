package knowledge

import (
	"testing"
	"bytes"
	"encoding/json"
	"os"
)

func TestGetPreflightContext(t *testing.T) {
	// Assumes in-memory store setup
	store, cleanup := testStore(t)
	defer cleanup()
	projectID := "testproj"
	store.CreateEntity("TestEntity", "file", projectID)

	in := new(bytes.Buffer)
	out := new(bytes.Buffer)

	tools := []mcp.Tool{
		{
			Name:        "get_preflight_context",
			Description: "",
			InputSchema: map[string]interface{}{"task": "string"},
		},
	}
	handlers := map[string]mcp.ToolHandler{
		"get_preflight_context": func(req *mcp.ToolCallRequest) (any, error) {
			task, _ := req.Arguments["task"].(string)
			entities, err := store.KeywordSearch(projectID, task, 16)
			if err != nil {
				return nil, err
			}
			res := "---\nRelevant Knowledge Entities for Task\n---\n"
			for _, e := range entities {
				if e.Entity != nil {
					res += "- " + e.Entity.Name + " (" + e.Entity.Type + ")\n"
				}
			}
			return res, nil
		},
	}
	server := mcp.NewServer(tools, handlers, in, out)

	// Emulate MCP preflight query
	msg := map[string]interface{}{
		"tool": "get_preflight_context",
		"arguments": map[string]interface{}{"task": "TestEntity"},
	}
	_ = json.NewEncoder(in).Encode(msg)
	os.Stdin = in
	os.Stdout = out
	
	go server.Serve()

	// Parse output and check contents
	var response map[string]interface{}
	_ = json.NewDecoder(out).Decode(&response)
	result, ok := response["result"].(string)
	if !ok || result == "" {
		t.Fatalf("expected non-empty result, got %v", response)
	}
	if !(bytes.Contains([]byte(result), []byte("TestEntity"))) {
		t.Fatalf("expected context to include entity name, got: %s", result)
	}
}