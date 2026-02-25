package knowledge

import (
	"context"
	"fmt"
	"strings"
)

// BatchEmbed processes all un-embedded entities and observations in a project
func (s *Store) BatchEmbed(ctx context.Context, projectID string, embedder Embedder) error {
	// Get un-embedded entities
	entities, err := s.GetUnembeddedEntities(projectID)
	if err != nil {
		return fmt.Errorf("get un-embedded entities: %w", err)
	}

	if len(entities) > 0 {
		// Prepare texts for batch embedding
		texts := make([]string, len(entities))
		for i, entity := range entities {
			texts[i] = fmt.Sprintf("%s: %s", entity.Type, entity.Name)
		}

		// Generate embeddings
		embeddings, err := embedder.Embed(ctx, texts)
		if err != nil {
			return fmt.Errorf("generate embeddings: %w", err)
		}

		// Store embeddings
		for i, entity := range entities {
			if err := s.SetEmbedding(entity.ID, embeddings[i]); err != nil {
				return fmt.Errorf("set embedding for entity %s: %w", entity.ID, err)
			}
		}
	}

	// Get un-embedded observations
	observations, err := s.GetUnembeddedObservations(projectID)
	if err != nil {
		return fmt.Errorf("get un-embedded observations: %w", err)
	}

	if len(observations) == 0 {
		return nil
	}

	// Prepare observation texts
	obsTexts := make([]string, len(observations))
	for i, obs := range observations {
		obsTexts[i] = obs.Content
	}

	// Generate observation embeddings
	obsEmbeddings, err := embedder.Embed(ctx, obsTexts)
	if err != nil {
		return fmt.Errorf("generate observation embeddings: %w", err)
	}

	// Store observation embeddings
	for i, obs := range observations {
		if err := s.SetObservationEmbedding(obs.ID, obsEmbeddings[i]); err != nil {
			return fmt.Errorf("set embedding for observation %s: %w", obs.ID, err)
		}
	}

	return nil
}

// GetUnembeddedEntities returns all entities without embeddings
func (s *Store) GetUnembeddedEntities(projectID string) ([]Entity, error) {
	query := fmt.Sprintf(`
		MATCH (e:Entity)
		WHERE e.project_id = '%s' AND e.embedding IS NULL
		RETURN e.id, e.name, e.type, e.project_id, e.created_at, e.updated_at
	`, escapeCypher(projectID))

	result, err := s.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query un-embedded entities: %w", err)
	}
	defer result.Close()

	var entities []Entity
	for result.HasNext() {
		row, err := result.Next()
		if err != nil {
			return nil, fmt.Errorf("get next row: %w", err)
		}

		id, _ := row.GetValue(0)
		name, _ := row.GetValue(1)
		typ, _ := row.GetValue(2)
		projID, _ := row.GetValue(3)

		entity := Entity{
			ID:        id.(string),
			Name:      name.(string),
			Type:      typ.(string),
			ProjectID: projID.(string),
		}
		entities = append(entities, entity)
	}

	return entities, nil
}

// GetUnembeddedObservations returns all observations without embeddings
func (s *Store) GetUnembeddedObservations(projectID string) ([]Observation, error) {
	query := fmt.Sprintf(`
		MATCH (e:Entity)-[:HAS_OBSERVATION]->(o:Observation)
		WHERE e.project_id = '%s' AND o.embedding IS NULL
		RETURN o.id, o.entity_id, o.content, o.created_at
	`, escapeCypher(projectID))

	result, err := s.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query un-embedded observations: %w", err)
	}
	defer result.Close()

	var observations []Observation
	for result.HasNext() {
		row, err := result.Next()
		if err != nil {
			return nil, fmt.Errorf("get next row: %w", err)
		}

		id, _ := row.GetValue(0)
		entityID, _ := row.GetValue(1)
		content, _ := row.GetValue(2)

		obs := Observation{
			ID:       id.(string),
			EntityID: entityID.(string),
			Content:  content.(string),
		}
		observations = append(observations, obs)
	}

	return observations, nil
}

// SetEmbedding stores an embedding vector for an entity
func (s *Store) SetEmbedding(entityID string, embedding []float32) error {
	query := fmt.Sprintf(`
		MATCH (e:Entity {id: '%s'})
		SET e.embedding = %s
	`, escapeCypher(entityID), formatFloatArray(embedding))

	result, err := s.conn.Query(query)
	if err != nil {
		return fmt.Errorf("set embedding: %w", err)
	}
	defer result.Close()

	return nil
}

// SetObservationEmbedding stores an embedding vector for an observation
func (s *Store) SetObservationEmbedding(observationID string, embedding []float32) error {
	query := fmt.Sprintf(`
		MATCH (o:Observation {id: '%s'})
		SET o.embedding = %s
	`, escapeCypher(observationID), formatFloatArray(embedding))

	result, err := s.conn.Query(query)
	if err != nil {
		return fmt.Errorf("set observation embedding: %w", err)
	}
	defer result.Close()

	return nil
}

// formatFloatArray formats a []float32 as a Kuzu Cypher FLOAT array literal,
// e.g. [0.1, 0.2, 0.3]
func formatFloatArray(v []float32) string {
	if len(v) == 0 {
		return "[]"
	}
	parts := make([]string, len(v))
	for i, f := range v {
		parts[i] = fmt.Sprintf("%v", f)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
