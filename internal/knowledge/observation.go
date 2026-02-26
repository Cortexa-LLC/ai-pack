package knowledge

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateObservation adds a new observation to an entity.
// The Observation node and its HAS_OBSERVATION edge are created in a single
// Cypher statement so the write is atomic – there is no window where an
// orphaned Observation node can exist if the relationship step fails.
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

	// Single atomic statement: match the parent Entity, create the Observation
	// node, and link them in one go.  Because everything happens in one Cypher
	// statement there is no risk of an orphaned Observation node if the edge
	// creation were to fail.
	query := fmt.Sprintf(`
		MATCH (e:Entity {id: '%s', project_id: '%s'})
		CREATE (o:Observation {
			id: '%s',
			entity_id: '%s',
			content: '%s',
			created_at: timestamp('%s')
		})
		CREATE (e)-[:HAS_OBSERVATION]->(o)
	`,
		escapeCypher(entityID), escapeCypher(projectID),
		obs.ID, escapeCypher(entityID), escapeCypher(content),
		obs.CreatedAt.Format(time.RFC3339),
	)

	result, err := s.query(query)
	if err != nil {
		return nil, fmt.Errorf("create observation: %w", err)
	}
	result.Close()

	return obs, nil
}

// GetObservations retrieves all observations for an entity
func (s *Store) GetObservations(entityID, projectID string) ([]*Observation, error) {
	// Verify entity exists and belongs to this project
	_, err := s.GetEntity(entityID, projectID)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		MATCH (e:Entity)-[:HAS_OBSERVATION]->(o:Observation)
		WHERE e.id = '%s' AND e.project_id = '%s'
		RETURN o.id, o.entity_id, o.content, o.created_at
	`, escapeCypher(entityID), escapeCypher(projectID))

	result, err := s.query(query)
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
		tuple.Close()
		if err != nil {
			return nil, fmt.Errorf("get row: %w", err)
		}

		obs := &Observation{
			ID:       row[0].(string),
			EntityID: row[1].(string),
			Content:  row[2].(string),
		}

		if ts, ok := row[3].(int64); ok {
			obs.CreatedAt = time.UnixMicro(ts).UTC()
		}

		observations = append(observations, obs)
	}

	return observations, nil
}

// DeleteObservation removes an observation
func (s *Store) DeleteObservation(obsID, entityID, projectID string) error {
	// Verify entity exists and belongs to this project
	_, err := s.GetEntity(entityID, projectID)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`
		MATCH (e:Entity)-[:HAS_OBSERVATION]->(o:Observation)
		WHERE o.id = '%s' AND e.id = '%s' AND e.project_id = '%s'
		DETACH DELETE o
	`, escapeCypher(obsID), escapeCypher(entityID), escapeCypher(projectID))

	result, err := s.query(query)
	if err != nil {
		return fmt.Errorf("delete observation: %w", err)
	}
	defer result.Close()

	return nil
}
