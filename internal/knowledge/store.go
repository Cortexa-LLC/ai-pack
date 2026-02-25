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

// initSchema creates node and relationship tables if they don't exist
func (s *Store) initSchema() error {
	schema := []string{
		// Entity node table
		`CREATE NODE TABLE IF NOT EXISTS Entity(
			id STRING PRIMARY KEY,
			name STRING,
			type STRING,
			project_id STRING,
			created_at TIMESTAMP,
			updated_at TIMESTAMP,
			embedding FLOAT[1536]
		)`,

		// Observation node table
		`CREATE NODE TABLE IF NOT EXISTS Observation(
			id STRING PRIMARY KEY,
			entity_id STRING,
			content STRING,
			created_at TIMESTAMP,
			embedding FLOAT[1536]
		)`,

		// Relationship tables
		`CREATE REL TABLE IF NOT EXISTS CALLS(FROM Entity TO Entity)`,
		`CREATE REL TABLE IF NOT EXISTS IMPORTS(FROM Entity TO Entity)`,
		`CREATE REL TABLE IF NOT EXISTS CONTAINS(FROM Entity TO Entity)`,
		`CREATE REL TABLE IF NOT EXISTS FIXES(FROM Entity TO Entity)`,
		`CREATE REL TABLE IF NOT EXISTS SUPERSEDES(FROM Entity TO Entity)`,
		`CREATE REL TABLE IF NOT EXISTS CAUSED_BY(FROM Entity TO Entity)`,
		`CREATE REL TABLE IF NOT EXISTS DEPENDS_ON(FROM Entity TO Entity)`,
		`CREATE REL TABLE IF NOT EXISTS IMPLEMENTS(FROM Entity TO Entity)`,
		`CREATE REL TABLE IF NOT EXISTS RELATES_TO(FROM Entity TO Entity)`,
		`CREATE REL TABLE IF NOT EXISTS TESTS(FROM Entity TO Entity)`,
		`CREATE REL TABLE IF NOT EXISTS DOCUMENTS(FROM Entity TO Entity)`,
		`CREATE REL TABLE IF NOT EXISTS HAS_OBSERVATION(FROM Entity TO Observation)`,
	}

	for _, stmt := range schema {
		result, err := s.conn.Query(stmt)
		if err != nil {
			return fmt.Errorf("execute schema statement: %w", err)
		}
		result.Close()
	}

	// Migrate existing tables to add embedding column if missing
	if err := s.migrateEmbeddings(); err != nil {
		return fmt.Errorf("migrate embeddings: %w", err)
	}

	return nil
}

// migrateEmbeddings adds embedding columns to existing tables if they don't exist
func (s *Store) migrateEmbeddings() error {
	migrations := []string{
		`ALTER TABLE Entity ADD embedding FLOAT[1536]`,
		`ALTER TABLE Observation ADD embedding FLOAT[1536]`,
	}

	for _, stmt := range migrations {
		result, err := s.conn.Query(stmt)
		if err != nil {
			// Ignore errors if column already exists
			// Kuzu will return an error if we try to add a column that exists
			continue
		}
		result.Close()
	}

	return nil
}

// Execute runs a raw Cypher query and returns the result
func (s *Store) Execute(query string) (*kuzu.QueryResult, error) {
	result, err := s.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("execute query: %w", err)
	}
	return result, nil
}
