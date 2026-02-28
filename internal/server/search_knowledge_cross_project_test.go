package server

import (
	"context"
	"testing"
)

// TestExecuteSearchKnowledgeInProjectMissingArgs verifies that the handler
// returns an error when required arguments are absent.
func TestExecuteSearchKnowledgeInProjectMissingArgs(t *testing.T) {
	s := &AgentServer{}

	t.Run("missing project_path", func(t *testing.T) {
		_, err := s.executeSearchKnowledgeInProject(context.Background(), map[string]interface{}{
			"query": "some query",
		})
		if err == nil {
			t.Fatal("expected error for missing project_path, got nil")
		}
	})

	t.Run("missing query", func(t *testing.T) {
		_, err := s.executeSearchKnowledgeInProject(context.Background(), map[string]interface{}{
			"project_path": "/some/path",
		})
		if err == nil {
			t.Fatal("expected error for missing query, got nil")
		}
	})

	t.Run("nil mcpManager", func(t *testing.T) {
		// With valid args but no MCP manager, should return a descriptive error.
		_, err := s.executeSearchKnowledgeInProject(context.Background(), map[string]interface{}{
			"project_path": "/some/path",
			"query":        "some query",
		})
		if err == nil {
			t.Fatal("expected error when MCP manager is nil, got nil")
		}
	})
}

// TestSearchKnowledgeInProjectToolDefined verifies that the search_knowledge_in_project
// tool is present in the tool list returned by getAllTools.
func TestSearchKnowledgeInProjectToolDefined(t *testing.T) {
	s := &AgentServer{}
	tools := s.getAllTools("")

	found := false
	for _, tool := range tools {
		if tool.Name == "search_knowledge_in_project" {
			found = true
			// Verify required properties are declared.
			if tool.InputSchema == nil {
				t.Error("search_knowledge_in_project: InputSchema is nil")
			}
			break
		}
	}
	if !found {
		t.Error("search_knowledge_in_project tool not found in getAllTools output")
	}
}
