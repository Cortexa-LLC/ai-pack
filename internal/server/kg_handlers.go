package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/monitoring"
)

// KGProjectStats holds knowledge graph statistics for a single project.
type KGProjectStats struct {
	ProjectRoot    string           `json:"project_root"`
	ProjectName    string           `json:"project_name"`
	EntityCount    int64            `json:"entity_count"`
	RelationCount  int64            `json:"relation_count"`
	EntityByType   map[string]int64 `json:"entity_by_type"`
	RelationByType map[string]int64 `json:"relation_by_type"`
	PreflightHits  int64            `json:"preflight_hits"`
	Available      bool             `json:"available"`
	Error          string           `json:"error,omitempty"`
}

// KGStatsResponse is the response payload for GET /api/kg/stats.
type KGStatsResponse struct {
	Projects        []KGProjectStats `json:"projects"`
	TotalEntities   int64            `json:"total_entities"`
	TotalRelations  int64            `json:"total_relations"`
	IndexedProjects int              `json:"indexed_projects"`
	GeneratedAt     time.Time        `json:"generated_at"`
}

// HandleKGStats returns knowledge-graph statistics for all registered projects.
func (s *AgentServer) HandleKGStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	projectRoots := s.GetProjectRoots()
	projects := make([]KGProjectStats, 0, len(projectRoots))
	var totalEntities, totalRelations int64
	indexedProjects := 0

	for _, root := range projectRoots {
		stats := s.queryKGStatsForProject(root)
		projects = append(projects, stats)
		if stats.Available {
			totalEntities += stats.EntityCount
			totalRelations += stats.RelationCount
			indexedProjects++
		}
	}

	resp := KGStatsResponse{
		Projects:        projects,
		TotalEntities:   totalEntities,
		TotalRelations:  totalRelations,
		IndexedProjects: indexedProjects,
		GeneratedAt:     time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

func (s *AgentServer) queryKGStatsForProject(projectRoot string) KGProjectStats {
	stats := KGProjectStats{
		ProjectRoot:    projectRoot,
		ProjectName:    filepath.Base(projectRoot),
		EntityByType:   make(map[string]int64),
		RelationByType: make(map[string]int64),
	}

	// Read preflight hit counter.
	s.mu.RLock()
	stats.PreflightHits = s.kgPreflightHits[projectRoot]
	s.mu.RUnlock()

	if s.mcpManager == nil {
		stats.Error = "MCP manager not initialized"
		return stats
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Total entity count.
	if rows, err := s.kgQuery(ctx, projectRoot, "MATCH (n) RETURN count(n) AS count"); err == nil {
		if len(rows) > 0 && len(rows[0]) > 0 {
			stats.EntityCount = toInt64(rows[0][0])
		}
	} else {
		stats.Error = err.Error()
		return stats
	}

	// Entity count by type.
	if rows, err := s.kgQuery(ctx, projectRoot,
		"MATCH (n) WHERE n.type IS NOT NULL RETURN n.type AS type, count(n) AS count ORDER BY count DESC",
	); err == nil {
		for _, row := range rows {
			if len(row) >= 2 {
				stats.EntityByType[fmt.Sprintf("%v", row[0])] = toInt64(row[1])
			}
		}
	}

	// Total relation count.
	if rows, err := s.kgQuery(ctx, projectRoot, "MATCH ()-[r]->() RETURN count(r) AS count"); err == nil {
		if len(rows) > 0 && len(rows[0]) > 0 {
			stats.RelationCount = toInt64(rows[0][0])
		}
	}

	// Relation count by type (top 10).
	if rows, err := s.kgQuery(ctx, projectRoot,
		"MATCH ()-[r]->() RETURN type(r) AS type, count(r) AS count ORDER BY count DESC LIMIT 10",
	); err == nil {
		for _, row := range rows {
			if len(row) >= 2 {
				stats.RelationByType[fmt.Sprintf("%v", row[0])] = toInt64(row[1])
			}
		}
	}

	stats.Available = true
	monitoring.Logger.Debug("kg_stats_queried", "project", stats.ProjectName, "entities", stats.EntityCount)
	return stats
}

// kgQuery runs a Cypher query against a project's KG via MCP and returns the rows.
func (s *AgentServer) kgQuery(ctx context.Context, projectRoot, cypher string) ([][]interface{}, error) {
	result, err := s.mcpManager.CallToolForProject(ctx, projectRoot, "query_graph", map[string]interface{}{
		"cypher": cypher,
	})
	if err != nil {
		return nil, err
	}
	if result == nil || len(result.Content) == 0 {
		return nil, fmt.Errorf("empty result")
	}
	var rows [][]interface{}
	if err := json.Unmarshal([]byte(result.Content[0].Text), &rows); err != nil {
		return nil, fmt.Errorf("parse result: %w", err)
	}
	return rows, nil
}

// toInt64 safely converts common numeric types from JSON to int64.
func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case float64:
		return int64(val)
	case int64:
		return val
	case int:
		return int64(val)
	case json.Number:
		n, _ := val.Int64()
		return n
	}
	return 0
}
