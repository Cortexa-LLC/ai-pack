package knowledge

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// SearchConfig defines search behavior and ranking weights
type SearchConfig struct {
	KeywordWeight float64 // α — keyword match weight, default 0.4
	RecencyWeight float64 // β — recency boost weight, default 0.1
	Limit         int     // maximum results to return, default 20
}

// DefaultSearchConfig returns the default search configuration
func DefaultSearchConfig() SearchConfig {
	return SearchConfig{
		KeywordWeight: 0.4,
		RecencyWeight: 0.1,
		Limit:         20,
	}
}

// SearchResult represents a single search result with score and metadata
type SearchResult struct {
	Entity       *Entity
	Observations []*Observation // top 3 observations
	Score        float64
	MatchType    string // "keyword" | "vector" | "hybrid"
}

// KeywordSearch performs full-text search on entity names and observation content
// using Cypher's CONTAINS operator for case-insensitive substring matching.
func (s *Store) KeywordSearch(projectID, query string, limit int) ([]*SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}
	if limit <= 0 {
		limit = 20
	}

	// Normalize query for case-insensitive matching
	normalizedQuery := strings.ToLower(query)

	// Cypher query to find entities where name contains the search term
	cypherQuery := fmt.Sprintf(`
		MATCH (e:Entity)
		WHERE e.project_id = '%s' AND lower(e.name) CONTAINS '%s'
		RETURN e.id, e.name, e.type, e.project_id, e.created_at, e.updated_at
		LIMIT %d
	`, projectID, normalizedQuery, limit)

	result, err := s.conn.Query(cypherQuery)
	if err != nil {
		return nil, fmt.Errorf("execute keyword search: %w", err)
	}
	defer result.Close()

	var results []*SearchResult
	for result.HasNext() {
		tuple, err := result.Next()
		if err != nil {
			return nil, fmt.Errorf("get next: %w", err)
		}

		row, err := tuple.GetAsSlice()
		if err != nil {
			return nil, fmt.Errorf("get row as slice: %w", err)
		}
		tuple.Close()

		entity := &Entity{
			ID:        row[0].(string),
			Name:      row[1].(string),
			Type:      row[2].(string),
			ProjectID: row[3].(string),
		}

		// Parse timestamps (Kuzu returns timestamps as int64 microseconds)
		if ts, ok := row[4].(int64); ok {
			entity.CreatedAt = time.UnixMicro(ts).UTC()
		}
		if ts, ok := row[5].(int64); ok {
			entity.UpdatedAt = time.UnixMicro(ts).UTC()
		}

		// Fetch top 3 observations for this entity
		obs, err := s.GetTopObservations(entity.ID, projectID, 3)
		if err != nil {
			// Don't fail the entire search if observations fail
			obs = []*Observation{}
		}

		// Calculate keyword match score (simple: 1.0 for match)
		score := 1.0

		results = append(results, &SearchResult{
			Entity:       entity,
			Observations: obs,
			Score:        score,
			MatchType:    "keyword",
		})
	}

	return results, nil
}

// GetTopObservations retrieves the most recent observations for an entity
func (s *Store) GetTopObservations(entityID, projectID string, limit int) ([]*Observation, error) {
	query := fmt.Sprintf(`
		MATCH (o:Observation)
		WHERE o.entity_id = '%s'
		RETURN o.id, o.entity_id, o.content, o.created_at
		ORDER BY o.created_at DESC
		LIMIT %d
	`, entityID, limit)

	result, err := s.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query observations: %w", err)
	}
	defer result.Close()

	var observations []*Observation
	for result.HasNext() {
		tuple, err := result.Next()
		if err != nil {
			return nil, fmt.Errorf("get next: %w", err)
		}

		row, err := tuple.GetAsSlice()
		if err != nil {
			return nil, fmt.Errorf("get row as slice: %w", err)
		}
		tuple.Close()

		obs := &Observation{
			ID:       row[0].(string),
			EntityID: row[1].(string),
			Content:  row[2].(string),
		}

		// Parse timestamp (Kuzu returns timestamps as int64 microseconds)
		if ts, ok := row[3].(int64); ok {
			obs.CreatedAt = time.UnixMicro(ts).UTC()
		}

		observations = append(observations, obs)
	}

	return observations, nil
}

// VectorSearch performs semantic search using cosine similarity over stored entity embeddings.
// Entities must have been previously embedded via SetEmbedding / BatchEmbed.
// Results are sorted by descending cosine similarity and capped at limit.
func (s *Store) VectorSearch(projectID string, queryEmbedding []float64, limit int) ([]*SearchResult, error) {
	if len(queryEmbedding) == 0 {
		return nil, fmt.Errorf("query embedding cannot be empty")
	}
	if limit <= 0 {
		limit = 20
	}

	// Retrieve all entities for this project that have a stored embedding.
	// We fetch them in Go and compute cosine similarity here because Kuzu does
	// not yet expose a native vector-similarity function.
	cypherQuery := fmt.Sprintf(`
		MATCH (e:Entity)
		WHERE e.project_id = '%s' AND e.embedding IS NOT NULL
		RETURN e.id, e.name, e.type, e.project_id, e.created_at, e.updated_at, e.embedding
	`, escapeCypher(projectID))

	result, err := s.conn.Query(cypherQuery)
	if err != nil {
		return nil, fmt.Errorf("query entities with embeddings: %w", err)
	}
	defer result.Close()

	type candidate struct {
		entity    *Entity
		embedding []float32
	}
	var candidates []candidate

	for result.HasNext() {
		tuple, err := result.Next()
		if err != nil {
			return nil, fmt.Errorf("get next: %w", err)
		}

		row, err := tuple.GetAsSlice()
		if err != nil {
			return nil, fmt.Errorf("get row as slice: %w", err)
		}
		tuple.Close()

		entity := &Entity{
			ID:        row[0].(string),
			Name:      row[1].(string),
			Type:      row[2].(string),
			ProjectID: row[3].(string),
		}

		if ts, ok := row[4].(int64); ok {
			entity.CreatedAt = time.UnixMicro(ts).UTC()
		}
		if ts, ok := row[5].(int64); ok {
			entity.UpdatedAt = time.UnixMicro(ts).UTC()
		}

		// row[6] is the embedding returned as []any containing float32 values
		rawEmb, ok := row[6].([]any)
		if !ok || len(rawEmb) == 0 {
			continue
		}
		emb := make([]float32, len(rawEmb))
		for i, v := range rawEmb {
			if f, ok := v.(float32); ok {
				emb[i] = f
			}
		}

		candidates = append(candidates, candidate{entity: entity, embedding: emb})
	}

	if len(candidates) == 0 {
		return []*SearchResult{}, nil
	}

	// Compute cosine similarity for each candidate
	type scored struct {
		entity *Entity
		score  float64
	}
	scores := make([]scored, 0, len(candidates))
	for _, c := range candidates {
		sim := cosineSimilarity(queryEmbedding, c.embedding)
		scores = append(scores, scored{entity: c.entity, score: sim})
	}

	// Sort descending by score
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Truncate to limit
	if len(scores) > limit {
		scores = scores[:limit]
	}

	// Build SearchResult slice with top observations
	results := make([]*SearchResult, 0, len(scores))
	for _, s2 := range scores {
		obs, err := s.GetTopObservations(s2.entity.ID, projectID, 3)
		if err != nil {
			obs = []*Observation{}
		}
		results = append(results, &SearchResult{
			Entity:       s2.entity,
			Observations: obs,
			Score:        s2.score,
			MatchType:    "vector",
		})
	}

	return results, nil
}

// cosineSimilarity computes the cosine similarity between a float64 query
// vector and a float32 stored vector.  Returns 0 if either vector is zero.
func cosineSimilarity(query []float64, stored []float32) float64 {
	// Use the shorter length to avoid index out-of-range if dimensions differ.
	n := len(query)
	if len(stored) < n {
		n = len(stored)
	}
	if n == 0 {
		return 0
	}

	var dot, normQ, normS float64
	for i := 0; i < n; i++ {
		q := query[i]
		s := float64(stored[i])
		dot += q * s
		normQ += q * q
		normS += s * s
	}
	if normQ == 0 || normS == 0 {
		return 0
	}
	return dot / (math.Sqrt(normQ) * math.Sqrt(normS))
}

// HybridSearch combines keyword and vector search with configurable weights
func (s *Store) HybridSearch(projectID, query string, queryEmbedding []float64, config SearchConfig) ([]*SearchResult, error) {
	if query == "" && len(queryEmbedding) == 0 {
		return nil, fmt.Errorf("either query or embedding must be provided")
	}

	// Use default config if not provided
	if config.Limit == 0 {
		config = DefaultSearchConfig()
	}

	// Collect results from both search methods
	var allResults []*SearchResult

	// Keyword search results
	if query != "" {
		keywordResults, err := s.KeywordSearch(projectID, query, config.Limit*2)
		if err != nil {
			return nil, fmt.Errorf("keyword search: %w", err)
		}
		allResults = append(allResults, keywordResults...)
	}

	// Vector search results (when implemented)
	if len(queryEmbedding) > 0 {
		vectorResults, err := s.VectorSearch(projectID, queryEmbedding, config.Limit*2)
		if err != nil {
			return nil, fmt.Errorf("vector search: %w", err)
		}
		allResults = append(allResults, vectorResults...)
	}

	// Deduplicate by entity ID and combine scores
	entityScores := make(map[string]*SearchResult)
	for _, result := range allResults {
		entityID := result.Entity.ID
		if existing, found := entityScores[entityID]; found {
			// Combine scores: weighted sum of keyword and semantic scores
			existing.Score += result.Score
			existing.MatchType = "hybrid"
		} else {
			entityScores[entityID] = result
		}
	}

	// Convert map back to slice
	var hybridResults []*SearchResult
	for _, result := range entityScores {
		// Apply recency boost
		recencyScore := calculateRecencyScore(result.Entity.UpdatedAt)
		result.Score = result.Score + config.RecencyWeight*recencyScore
		result.MatchType = "hybrid"
		hybridResults = append(hybridResults, result)
	}

	// Sort by score descending
	sort.Slice(hybridResults, func(i, j int) bool {
		return hybridResults[i].Score > hybridResults[j].Score
	})

	// Limit results
	if len(hybridResults) > config.Limit {
		hybridResults = hybridResults[:config.Limit]
	}

	return hybridResults, nil
}

// calculateRecencyScore computes a recency boost based on entity update time
// Returns a score between 0.0 (very old) and 1.0 (very recent)
func calculateRecencyScore(updatedAt time.Time) float64 {
	if updatedAt.IsZero() {
		return 0.0
	}

	// Calculate age in days
	now := time.Now().UTC()
	age := now.Sub(updatedAt)
	ageDays := age.Hours() / 24.0

	// Exponential decay: score = e^(-age/30)
	// Half-life of ~21 days
	score := math.Exp(-ageDays / 30.0)
	return math.Max(0.0, math.Min(1.0, score))
}
