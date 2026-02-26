package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}
package knowledge

import (
"errors"
)

type Store struct {
query func(string) (interface{}, error)
}

func (s *Store) Execute(query string) (interface{}, error) {
if s.query == nil {
return nil, errors.New("query function not set")
}
return s.query(query)
}package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}package knowledge

import (
	"errors"
)

// Store struct for the knowledge base
type Store struct {
	query func(string) (interface{}, error)
}

// Execute runs a query against the knowledge graph and ensures the query function is valid
func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

// Execute runs a query against the knowledge graph and ensures the query function is valid
func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

// Execute runs a query against the knowledge graph and ensures the query function is valid
func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

// Execute runs a query against the knowledge graph and ensures the query function is valid
func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

// Execute runs a query against the knowledge graph and ensures the query function is valid
func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}  package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

// Execute runs a query against the knowledge graph and ensures the query function is valid
func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}
package knowledge

import (
	"fmt"
	"strings"
	"time"
	"github.com/google/uuid"
)

// Store provides a way to prepare and execute queries on the knowledge graph.
type Store struct {
	query func(string) (interface{}, error)
}

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

	result, err := s.query(query)
	if err != nil {
		return nil, fmt.Errorf("create entity: %w", err)
	}
	defer result.Close()

	return entity, nil
}

// escapeCypher escapes strings for safe interpolation into Cypher queries.
func escapeCypher(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	return s
}

// Store struct for the knowledgebase
type Store struct {
	query func(string) (interface{}, error)
}

// Execute runs a query against the knowledge graph and ensures the query function is valid
func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}package knowledge

import (
	"errors"
)

// Store struct for the knowledgebase

type Store struct {
	query func(string) (interface{}, error)
}

// Execute runs a query against the knowledge graph and ensures the query function is valid
func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}package knowledge

import (
	"errors"
)

// Store struct for the knowledgebase

type Store struct {
	query func(string) (interface{}, error)
}

// Execute runs a query against the knowledge graph and ensures the query function is valid
func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}package knowledge

import (
	"errors"
)

// Store struct for the knowledgebase

type Store struct {
	query func(string) (interface{}, error)
}

// Execute runs a query against the knowledge graph and ensures the query function is valid
func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

// Execute runs a query against the knowledge graph and ensures the query function is valid
func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

// Execute runs a query against the knowledge graph and ensures the query function is valid
func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}
package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

// Execute runs a query against the knowledge graph and ensures the query function is valid
func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}
package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

// Execute runs a query against the knowledge graph and ensures the query function is valid
func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}
package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}
package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

// Execute runs a query against the knowledge graph
func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}
package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

// Execute runs a query against the knowledge graph and ensures the query function is valid
func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}
package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

// Execute runs a query against the knowledge graph and ensures the query function is valid
func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

// Execute runs a query against the knowledge graph and ensures the query function is valid
func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}
package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

// Execute runs a query against the knowledge graph
func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}package knowledge

import (
	"fmt"
	"strings"
	"time"
	"github.com/google/uuid"
)

// Store provides a way to prepare and execute queries on the knowledge graph.
type Store struct {
	query func(string) (interface{}, error)
}

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

	result, err := s.query(query)
	if err != nil {
		return nil, fmt.Errorf("create entity: %w", err)
	}
	defer result.Close()

	return entity, nil
}

// escapeCypher escapes strings for safe interpolation into Cypher queries.
func escapeCypher(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	return s
}

type Store struct {
	query func(string) (interface{}, error)
}

// Execute runs a query against the knowledge graph and ensures the query function is valid
func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

// Execute runs a query against the knowledge graph
func (s *Store) Execute(query string) (interface{}, error) {
	if s.query == nil {
		return nil, errors.New("query function not set")
	}
	return s.query(query)
}package knowledge

import (
	"errors"
)

type Store struct {
	query func(string) (interface{}, error)
}

// Execute runs a query against the knowledge graph
func (s *Store) Execute(query string) (interface{}, error) {
	return s.query(query)
}package knowledge

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

package knowledge

import (
	"fmt"
	"strings"
	"time"
	"github.com/google/uuid"
)

// Store provides a way to prepare and execute queries on the knowledge graph.
type Store struct {
	// Fields go here
}

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

	result, err := s.query(query)
	if err != nil {
		return nil, fmt.Errorf("create entity: %w", err)
	}
	defer result.Close()

	return entity, nil
}

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

	result, err := s.query(query)
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

	result, err := s.query(query)
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

	result, err := s.query(query)
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

	result, err := s.query(query)
	if err != nil {
		return fmt.Errorf("delete entity: %w", err)
	}
	defer result.Close()

	return nil
}

// escapeCypher escapes strings for safe interpolation into Cypher queries.
// It first escapes backslashes, then single quotes, preventing backslash-injection
// attacks where a crafted input like foo\' could re-enable quote injection.
func escapeCypher(s string) string {
	// Must escape backslashes BEFORE single quotes; otherwise a trailing
	// backslash in the input would escape the quote we add next.
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	return s
}
