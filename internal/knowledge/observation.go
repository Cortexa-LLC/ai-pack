package knowledge

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateObservation adds a new observation to an entity
func (s *Store) CreateObservation(entityID, content, projectID string) (*Observation, error) {
	// Verify entity exists and belongs to this project
	_, err := s.GetEntity(entityID, projectID)
	if err != nil {
		return nil, fmt.Errorf("entity: %w", err)
	}

	obs := &Observation{
		ID:        uuid.New().String(),
		EntityID:  entityID,
		Content:   content,
		CreatedAt: time.Now().UTC(),
	}

	// Create observation node and relationship in a single query
	query := fmt.Sprintf(`
		MATCH (e:Entity)
		WHERE e.id = '%s'
		CREATE (o:Observation {
			id: '%s',
			entity_id: '%s',
			content: '%s',
			created_at: timestamp('%s')
		})
		CREATE (e)-[:HAS_OBSERVATION]->(o)
	`,
		escapeCypher(entityID), obs.ID, escapeCypher(entityID), escapeCypher(content),
		obs.CreatedAt.Format(time.RFC3339))

	result, err := s.query(query)
	if err != nil {
		return nil, fmt.Errorf("create observation node and relationship: %w", err)
	}
	defer result.Close()

	return obs, nil
}