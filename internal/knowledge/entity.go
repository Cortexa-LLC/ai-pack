package knowledge

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateEntity adds a new entity to the knowledge graph
func (s *Store) CreateEntity(name, entityType, projectID string) (*Entity, error) {
	entity := &Entity{
		ID:        uuid.New().String(),
		Name:      name,
		Type:      entityType,
		ProjectID: projectID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	query := fmt.Sprintf(`
		CREATE (e:Entity {
			id: '%s',
			name: '%s',
			type: '%s',
			project_id: '%s',
			created_at: timestamp('%s'),
			updated_at: timestamp('%s')
		})
	`, entity.ID, escapeCypher(name), escapeCypher(entityType), 
	   escapeCypher(projectID), entity.CreatedAt.Format(time.RFC3339), 
	   entity.UpdatedAt.Format(time.RFC3339))

	result, err := s.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("create entity: %w", err)
	}
	defer result.Close()

	return entity, nil
}

// GetEntity retrieves an entity by ID for a specific project
func (s *Store) GetEntity(id, projectID string) (*Entity, error) {
	query := fmt.Sprintf(`
		MATCH (e:Entity)
		WHERE e.id = '%s' AND e.project_id = '%s'
		RETURN e.id, e.name, e.type, e.project_id, e.created_at, e.updated_at
	`, escapeCypher(id), escapeCypher(projectID))

	result, err := s.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query entity: %w", err)
	}
	defer result.Close()

	if !result.HasNext() {
		return nil, fmt.Errorf("entity not found: %s", id)
	}

	tuple, err := result.Next()
	if err != nil {
		return nil, fmt.Errorf("get next: %w", err)
	}
	defer tuple.Close()

	row, err := tuple.GetAsSlice()
	if err != nil {
		return nil, fmt.Errorf("get row: %w", err)
	}

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

	return entity, nil
}

// ListEntities retrieves all entities for a project, optionally filtered by type
func (s *Store) ListEntities(projectID, entityType string) ([]*Entity, error) {
	var query string
	if entityType == "" {
		query = fmt.Sprintf(`
			MATCH (e:Entity)
			WHERE e.project_id = '%s'
			RETURN e.id, e.name, e.type, e.project_id, e.created_at, e.updated_at
		`, escapeCypher(projectID))
	} else {
		query = fmt.Sprintf(`
			MATCH (e:Entity)
			WHERE e.project_id = '%s' AND e.type = '%s'
			RETURN e.id, e.name, e.type, e.project_id, e.created_at, e.updated_at
		`, escapeCypher(projectID), escapeCypher(entityType))
	}

	result, err := s.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query entities: %w", err)
	}
	defer result.Close()

	var entities []*Entity
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

		entities = append(entities, entity)
	}

	return entities, nil
}

// DeleteEntity removes an entity and all its relations
func (s *Store) DeleteEntity(id, projectID string) error {
	// First verify the entity exists and belongs to this project
	_, err := s.GetEntity(id, projectID)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`
		MATCH (e:Entity)
		WHERE e.id = '%s' AND e.project_id = '%s'
		DETACH DELETE e
	`, escapeCypher(id), escapeCypher(projectID))

	result, err := s.conn.Query(query)
	if err != nil {
		return fmt.Errorf("delete entity: %w", err)
	}
	defer result.Close()

	return nil
}

// escapeCypher escapes single quotes in strings for Cypher queries
func escapeCypher(s string) string {
	// Simple escape for single quotes - in production, use parameterized queries
	result := ""
	for _, ch := range s {
		if ch == '\'' {
			result += "\\'"
		} else {
			result += string(ch)
		}
	}
	return result
}
