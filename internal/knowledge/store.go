package knowledge

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	kuzu "github.com/kuzudb/go-kuzu"
)

// Store manages the Kuzu knowledge graph database.
//
// Concurrency model: KuzuDB connections are not goroutine-safe.
// mu serialises all calls that touch s.conn (Query, Prepare, Execute).
// QueryResult iteration (HasNext/Next) operates on a materialised C struct
// and does not call back into the connection, so it is safe after mu is
// released.
type Store struct {
	db      *kuzu.Database
	conn    *kuzu.Connection
	mu      sync.Mutex // guards all s.conn calls
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

// Close closes the database connection.
// It acquires mu to ensure no query is in flight while closing.
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn != nil {
		s.conn.Close()
	}
	if s.db != nil {
		s.db.Close()
	}
	return nil
}

// query runs a raw Cypher statement and returns the Kuzu result handle.
// Use only for schema DDL and other statements that contain no user-supplied values.
// For queries containing user input, use queryParams instead.
// mu is held for the duration of the call; result iteration is safe after release.
func (s *Store) query(stmt string) (*kuzu.QueryResult, error) {
	s.mu.Lock()
	result, err := s.conn.Query(stmt)
	s.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("execute query: %w", err)
	}
	return result, nil
}

// queryParams prepares a Cypher statement and executes it with bound parameters,
// preventing Cypher injection from user-supplied string values.
// Use $paramName placeholders in stmt and provide matching keys in params.
// mu is held for the prepare→execute sequence; result iteration is safe after release.
func (s *Store) queryParams(stmt string, params map[string]any) (*kuzu.QueryResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	prepared, err := s.conn.Prepare(stmt)
	if err != nil {
		return nil, fmt.Errorf("prepare query: %w", err)
	}
	defer prepared.Close()
	result, err := s.conn.Execute(prepared, params)
	if err != nil {
		return nil, fmt.Errorf("execute query: %w", err)
	}
	return result, nil
}
