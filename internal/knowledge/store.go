package knowledge

import (
	"fmt"
	"os"
	"path/filepath"

	kuzu "github.com/kuzudb/go-kuzu"
)

// Store manages the Kuzu knowledge graph database
type Store struct {
	db      *kuzu.Database
	conn    *kuzu.Connection
	path    string
	hnswIdx *vectorIndexCache // per-project lazy HNSW index
}

// OpenStore opens or creates a Kuzu database at the given path
func OpenStore(dbPath string) (*Store, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	// Open database
	db, err := kuzu.OpenDatabase(dbPath, kuzu.DefaultSystemConfig())
	if err != nil {
		return nil, fmt.Errorf("open kuzu database: %w", err)
	}

	// Create connection
	conn, err := kuzu.OpenConnection(db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create connection: %w", err)
	}

	store := &Store{
		db:      db,
		conn:    conn,
		path:    dbPath,
		hnswIdx: newVectorIndexCache(),
	}

	// Initialize schema
	if err := store.initSchema(); err != nil {
		store.Close()
		return nil, fmt.Errorf("initialize schema: %w", err)
	}

	return store, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.db != nil {
		s.db.Close()
	}
	return nil
}

// query runs a raw Cypher statement and returns the result.
// It is the internal counterpart of the public Execute method.
func (s *Store) query(stmt string) (*kuzu.QueryResult, error) {
	result, err := s.conn.Query(stmt)
	if err != nil {
		return nil, fmt.Errorf("execute query: %w", err)
	}
	return result, nil
}

// Execute runs a raw Cypher query and returns the result
func (s *Store) Execute(query string) (*kuzu.QueryResult, error) {
	return s.query(query)
}
