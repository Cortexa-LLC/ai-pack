package knowledge

import (
	"bufio"
	"fmt"
	"os"

	"github.com/cortexa-llc/ai-pack/internal/mcp"
)

// RunMCPServer exposes Store APIs as MCP tools over stdio (MCP protocol).
//
// Open-use-close pattern: each tool handler opens the database, executes its
// operation, and closes the database before returning.  No connection is held
// between calls, so `kg index` (and any other CLI command) can run at any time
// without hitting a "database is locked" error.
//
// Reads open in read-only mode (multiple concurrent readers allowed).
// Writes open in write mode (exclusive, but released immediately after the call).
func RunMCPServer(dbPath, projectID, projectRoot string) error {
	tools := []mcp.Tool{
		{
			Name:        "get_preflight_context",
			Description: "Returns formatted context block of relevant knowledge entities for a given task description",
			InputSchema: map[string]interface{}{"task": "string"},
		},
		{
			Name:        "search_knowledge",
			Description: "Hybrid search for entities/observations in the knowledge graph",
			InputSchema: map[string]interface{}{"query": "string", "limit": "integer"},
		},
		{
			Name:        "add_entity",
			Description: "Upsert an entity in the graph (name and type required; returns the entity ID)",
			InputSchema: map[string]interface{}{"name": "string", "type": "string"},
		},
		{
			Name:        "add_observation",
			Description: "Attach a text observation/note to an existing entity",
			InputSchema: map[string]interface{}{"entity_id": "string", "content": "string"},
		},
		{
			Name:        "link_entities",
			Description: "Create a directed relation (edge) between two entities",
			InputSchema: map[string]interface{}{"from_id": "string", "relation": "string", "to_id": "string"},
		},
		{
			Name:        "get_file_context",
			Description: "Return all entities associated with a file path",
			InputSchema: map[string]interface{}{"file": "string"},
		},
		{
			Name:        "query_graph",
			Description: "Run a read-only Cypher query against the knowledge graph",
			InputSchema: map[string]interface{}{"cypher": "string"},
		},
		{
			Name:        "index_project",
			Description: "Re-index the project codebase into the knowledge graph (scans all source files and updates entities/relations). Call this after significant code changes.",
			InputSchema: map[string]interface{}{},
		},
	}

	// withRO opens the DB in read-only mode, runs fn, then closes.
	withRO := func(fn func(*Store) (any, error)) (any, error) {
		s, err := OpenStoreReadOnly(dbPath)
		if err != nil {
			return nil, fmt.Errorf("open store: %w", err)
		}
		defer s.Close()
		return fn(s)
	}

	// withRW opens the DB in read-write mode, runs fn, then closes.
	withRW := func(fn func(*Store) (any, error)) (any, error) {
		s, err := OpenStore(dbPath)
		if err != nil {
			return nil, fmt.Errorf("open store: %w", err)
		}
		defer s.Close()
		return fn(s)
	}

	handlers := map[string]mcp.ToolHandler{
		"get_preflight_context": func(req *mcp.ToolCallRequest) (any, error) {
			return withRO(func(s *Store) (any, error) {
				task, _ := req.Arguments["task"].(string)
				entities, err := s.KeywordSearch(projectID, task, 16)
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
			})
		},

		"search_knowledge": func(req *mcp.ToolCallRequest) (any, error) {
			return withRO(func(s *Store) (any, error) {
				q, _ := req.Arguments["query"].(string)
				lim, _ := req.Arguments["limit"].(float64)
				if lim == 0 {
					lim = 12
				}
				return s.KeywordSearch(projectID, q, int(lim))
			})
		},

		"get_file_context": func(req *mcp.ToolCallRequest) (any, error) {
			return withRO(func(s *Store) (any, error) {
				file, _ := req.Arguments["file"].(string)
				return s.ListEntities(projectID, file)
			})
		},

		"query_graph": func(req *mcp.ToolCallRequest) (any, error) {
			return withRO(func(s *Store) (any, error) {
				cypher, _ := req.Arguments["cypher"].(string)
				if err := isReadOnlyCypher(cypher); err != nil {
					return nil, err
				}
				result, err := s.query(cypher)
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
			})
		},

		"add_entity": func(req *mcp.ToolCallRequest) (any, error) {
			return withRW(func(s *Store) (any, error) {
				name, _ := req.Arguments["name"].(string)
				typeStr, _ := req.Arguments["type"].(string)
				return s.CreateEntity(name, typeStr, projectID)
			})
		},

		"add_observation": func(req *mcp.ToolCallRequest) (any, error) {
			return withRW(func(s *Store) (any, error) {
				entityID, _ := req.Arguments["entity_id"].(string)
				content, _ := req.Arguments["content"].(string)
				return s.CreateObservation(entityID, content, projectID)
			})
		},

		"link_entities": func(req *mcp.ToolCallRequest) (any, error) {
			return withRW(func(s *Store) (any, error) {
				from, _ := req.Arguments["from_id"].(string)
				rel, _ := req.Arguments["relation"].(string)
				to, _ := req.Arguments["to_id"].(string)
				return nil, s.CreateRelation(from, to, rel, projectID)
			})
		},

		"index_project": func(req *mcp.ToolCallRequest) (any, error) {
			return withRW(func(s *Store) (any, error) {
				indexer, err := NewIndexer(s, projectID, projectRoot)
				if err != nil {
					return nil, fmt.Errorf("create indexer: %w", err)
				}
				stats, err := indexer.Index()
				if err != nil {
					return nil, fmt.Errorf("index project: %w", err)
				}
				return fmt.Sprintf("Indexed %d files, created %d entities and %d relations in project '%s'",
					stats.FilesScanned, stats.EntitiesCreated, stats.RelationsCreated, projectID), nil
			})
		},
	}

	server := mcp.NewServer(tools, handlers, bufio.NewReader(os.Stdin), os.Stdout)
	return server.Serve()
}
