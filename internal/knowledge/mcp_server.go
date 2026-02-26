package knowledge

import (
	"bufio"
	"fmt"
	"os"

	"github.com/cortexa-llc/ai-pack/internal/mcp"
)

// RunMCPServer exposes Store APIs as MCP tools over stdio (MCP protocol)
func RunMCPServer(store *Store, projectID string) error {
	tools := []mcp.Tool{
		{
			Name:        "get_preflight_context",
			Description: "Returns formatted context block for a given task description",
			InputSchema: map[string]interface{}{"task": "string"},
		},
		{
			Name:        "search_knowledge",
			Description: "Hybrid search for entities/observations in the knowledge graph",
			InputSchema: map[string]interface{}{"query": "string", "limit": "integer"},
		},
		{
			Name:        "add_entity",
			Description: "Upsert an entity in the graph",
			InputSchema: map[string]interface{}{"name": "string", "type": "string"},
		},
		{
			Name:        "add_observation",
			Description: "Attach an observation to an entity",
			InputSchema: map[string]interface{}{"entity_id": "string", "content": "string"},
		},
		{
			Name:        "link_entities",
			Description: "Create a relation (edge) between entities",
			InputSchema: map[string]interface{}{"from_id": "string", "relation": "string", "to_id": "string"},
		},
		{
			Name:        "get_file_context",
			Description: "Return all entities associated with a file path",
			InputSchema: map[string]interface{}{"file": "string"},
		},
		{
			Name:        "query_graph",
			Description: "Run a raw Cypher query",
			InputSchema: map[string]interface{}{"cypher": "string"},
		},
	}

	handlers := map[string]mcp.ToolHandler{
		"search_knowledge": func(req *mcp.ToolCallRequest) (any, error) {
			q, _ := req.Arguments["query"].(string)
			lim, _ := req.Arguments["limit"].(float64)
			if lim == 0 {
				lim = 12
			}
			return store.KeywordSearch(projectID, q, int(lim))
		},
		"add_entity": func(req *mcp.ToolCallRequest) (any, error) {
			name, _ := req.Arguments["name"].(string)
			typeStr, _ := req.Arguments["type"].(string)
			return store.CreateEntity(name, typeStr, projectID)
		},
		"add_observation": func(req *mcp.ToolCallRequest) (any, error) {
			entityID, _ := req.Arguments["entity_id"].(string)
			content, _ := req.Arguments["content"].(string)
			return store.CreateObservation(entityID, content, projectID)
		},
		"link_entities": func(req *mcp.ToolCallRequest) (any, error) {
			from, _ := req.Arguments["from_id"].(string)
			rel, _ := req.Arguments["relation"].(string)
			to, _ := req.Arguments["to_id"].(string)
			return nil, store.CreateRelation(from, to, rel, projectID)
		},
		"get_file_context": func(req *mcp.ToolCallRequest) (any, error) {
			file, _ := req.Arguments["file"].(string)
			return store.ListEntities(projectID, file)
		},
		"query_graph": func(req *mcp.ToolCallRequest) (any, error) {
			cypher, _ := req.Arguments["cypher"].(string)
			if err := isReadOnlyCypher(cypher); err != nil {
				return nil, err
			}
			result, err := store.query(cypher)
			if err != nil {
				return nil, fmt.Errorf("query: %w", err)
			}
			defer result.Close()
			var rows [][]any
			for result.HasNext() {
				tuple, err := result.Next()
				if err != nil {
					return nil, err
				}
				cols, err := tuple.GetAsSlice()
				tuple.Close()
				if err != nil {
					return nil, err
				}
				rows = append(rows, cols)
			}
			return rows, nil
		},		"get_preflight_context": func(req *mcp.ToolCallRequest) (any, error) {
			task, _ := req.Arguments["task"].(string)
			// Naive: Just return entities matching the task string as context block
			entities, err := store.KeywordSearch(projectID, task, 16)
			if err != nil {
				return nil, err
			}
			// Format as a context block
			res := "---\nRelevant Knowledge Entities for Task\n---\n"
			for _, e := range entities {
				if e.Entity != nil {
					res += "- " + e.Entity.Name + " (" + e.Entity.Type + ")\n"
				}
			}
			return res, nil
		},
	}

	server := mcp.NewServer(tools, handlers, bufio.NewReader(os.Stdin), os.Stdout)
	return server.Serve()
}
