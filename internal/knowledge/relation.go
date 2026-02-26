package knowledge

import (
	"fmt"
	"time"
)

// CreateRelation creates a directed relationship between two entities
func (s *Store) CreateRelation(fromID, toID, relType, projectID string) error {
	// Verify both entities exist and belong to this project
	_, err := s.GetEntity(fromID, projectID)
	if err != nil {
		return fmt.Errorf("source entity: %w", err)
	}

	_, err = s.GetEntity(toID, projectID)
	if err != nil {
		return fmt.Errorf("target entity: %w", err)
	}

	if err := validateRelType(relType); err != nil {
		return err
	}

	query := fmt.Sprintf(`
		MATCH (from:Entity), (to:Entity)
		WHERE from.id = '%s' AND to.id = '%s'
		CREATE (from)-[:%s]->(to)
	`, escapeCypher(fromID), escapeCypher(toID), relType)

	result, err := s.conn.Query(query)
	if err != nil {
		return fmt.Errorf("create relation: %w", err)
	}
	defer result.Close()

	return nil
}

// GetRelations retrieves all outgoing relations from an entity
func (s *Store) GetRelations(entityID, projectID string) ([]*Relation, error) {
	// Verify entity exists and belongs to this project
	_, err := s.GetEntity(entityID, projectID)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		MATCH (from:Entity)-[r]->(to:Entity)
		WHERE from.id = '%s' AND from.project_id = '%s'
		RETURN from.id, to.id, label(r)
	`, escapeCypher(entityID), escapeCypher(projectID))

	result, err := s.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query relations: %w", err)
	}
	defer result.Close()

	var relations []*Relation
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

		relation := &Relation{
			FromID: row[0].(string),
			ToID:   row[1].(string),
			Type:   row[2].(string),
		}

		relations = append(relations, relation)
	}

	return relations, nil
}

// DeleteRelation removes a specific relation between two entities
func (s *Store) DeleteRelation(fromID, toID, relType, projectID string) error {
	if err := validateRelType(relType); err != nil {
		return err
	}

	// Verify both entities exist and belong to this project
	_, err := s.GetEntity(fromID, projectID)
	if err != nil {
		return fmt.Errorf("source entity: %w", err)
	}

	_, err = s.GetEntity(toID, projectID)
	if err != nil {
		return fmt.Errorf("target entity: %w", err)
	}

	query := fmt.Sprintf(`
		MATCH (from:Entity)-[r:%s]->(to:Entity)
		WHERE from.id = '%s' AND to.id = '%s' AND from.project_id = '%s'
		DELETE r
	`, relType, escapeCypher(fromID), escapeCypher(toID), escapeCypher(projectID))

	result, err := s.conn.Query(query)
	if err != nil {
		return fmt.Errorf("delete relation: %w", err)
	}
	defer result.Close()

	return nil
}

// TraverseRelations follows a relation type from an entity and returns connected entities
func (s *Store) TraverseRelations(entityID, relType, projectID string) ([]*Entity, error) {
	if err := validateRelType(relType); err != nil {
		return nil, err
	}

	// Verify source entity exists and belongs to this project
	_, err := s.GetEntity(entityID, projectID)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		MATCH (from:Entity)-[:%s]->(to:Entity)
		WHERE from.id = '%s' AND from.project_id = '%s'
		RETURN to.id, to.name, to.type, to.project_id, to.created_at, to.updated_at
	`, relType, escapeCypher(entityID), escapeCypher(projectID))

	result, err := s.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("traverse relations: %w", err)
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
