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
			Name: "query_graph",
			// Restrict query_graph to read-only Cypher: allowed commands are only MATCH and RETURN, any mutations will be rejected. Cypher: MATCH/RETURN only; reject mutations (CREATE, DELETE, MERGE, SET).,
			InputSchema: map[string]interface{}{"cypher": "string"},
		},
	}

	handlers := map[string]mcp.ToolHandler{
		"search_knowledge": func(req *mcp.ToolCallRequest) (any, error) {
			q, ok := req.Arguments["query"].(string)
			if !ok || q == "" {
				return nil, fmt.Errorf("search_knowledge: missing or invalid 'query' argument")
			}
			lim, _ := req.Arguments["limit"].(float64)
			if lim == 0 {
				lim = 12
			}
			results, err := store.KeywordSearch(projectID, q, int(lim))
			if err != nil {
				return nil, fmt.Errorf("search_knowledge: %w", err)
			}
			return results, nil
		},
		"add_entity": func(req *mcp.ToolCallRequest) (any, error) {
			name, ok := req.Arguments["name"].(string)
			if !ok || name == "" {
				return nil, fmt.Errorf("add_entity: missing or invalid 'name' argument")
			}
			typeStr, ok := req.Arguments["type"].(string)
			if !ok || typeStr == "" {
				return nil, fmt.Errorf("add_entity: missing or invalid 'type' argument")
			}
			entity, err := store.CreateEntity(name, typeStr, projectID)
			if err != nil {
				return nil, fmt.Errorf("add_entity: %w", err)
			}
			return entity, nil
		},
		"add_observation": func(req *mcp.ToolCallRequest) (any, error) {
			entityID, ok := req.Arguments["entity_id"].(string)
			if !ok || entityID == "" {
				return nil, fmt.Errorf("add_observation: missing or invalid 'entity_id' argument")
			}
			content, ok := req.Arguments["content"].(string)
			if !ok || content == "" {
				return nil, fmt.Errorf("add_observation: missing or invalid 'content' argument")
			}
			obs, err := store.CreateObservation(entityID, content, projectID)
			if err != nil {
				return nil, fmt.Errorf("add_observation: %w", err)
			}
			return obs, nil
		},
		"link_entities": func(req *mcp.ToolCallRequest) (any, error) {
			from, ok := req.Arguments["from_id"].(string)
			if !ok || from == "" {
				return nil, fmt.Errorf("link_entities: missing or invalid 'from_id' argument")
			}
			rel, ok := req.Arguments["relation"].(string)
			if !ok || rel == "" {
				return nil, fmt.Errorf("link_entities: missing or invalid 'relation' argument")
			}
			to, ok := req.Arguments["to_id"].(string)
			if !ok || to == "" {
				return nil, fmt.Errorf("link_entities: missing or invalid 'to_id' argument")
			}
			if err := store.CreateRelation(from, to, rel, projectID); err != nil {
				return nil, fmt.Errorf("link_entities: %w", err)
			}
			return nil, nil
		},
		"get_file_context": func(req *mcp.ToolCallRequest) (any, error) {
			file, ok := req.Arguments["file"].(string)
			if !ok || file == "" {
				return nil, fmt.Errorf("get_file_context: missing or invalid 'file' argument")
			}
			entities, err := store.ListEntities(projectID, file)
			if err != nil {
				return nil, fmt.Errorf("get_file_context: %w", err)
			}
			return entities, nil
		},
		"query_graph": func(req *mcp.ToolCallRequest) (any, error) {
			cypher, ok := req.Arguments["cypher"].(string)
			if !ok || cypher == "" {
				return nil, fmt.Errorf("query_graph: missing or invalid 'cypher' argument")
			}
			result, err := store.Execute(cypher)
			if err != nil {
				return nil, fmt.Errorf("query_graph: %w", err)
			}
			defer result.Close()
			var rows [][]any
			for result.HasNext() {
				tuple, err := result.Next()
				if err != nil {
					return nil, fmt.Errorf("query_graph: %w", err)
				}
				cols, err := tuple.GetAsSlice()
				tuple.Close()
				if err != nil {
					return nil, fmt.Errorf("query_graph: %w", err)
				}
				rows = append(rows, cols)
			}
			return rows, nil
		},
		"get_preflight_context": func(req *mcp.ToolCallRequest) (any, error) {
			task, ok := req.Arguments["task"].(string)
			if !ok || task == "" {
				return nil, fmt.Errorf("get_preflight_context: missing or invalid 'task' argument")
			}
			// Naive: Just return entities matching the task string as context block
			entities, err := store.KeywordSearch(projectID, task, 16)
			if err != nil {
				return nil, fmt.Errorf("get_preflight_context: %w", err)
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
