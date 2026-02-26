package knowledge

import (
	"errors"
)

// AllowedRelTypes defines the valid relation types for the knowledge graph.
var AllowedRelTypes = map[string]struct{}{
	"friend": {},
	"colleague": {},
	"family": {},
	// Add other relation types
}

// initSchema initializes the knowledge graph schema.
func initSchema(conn *kuzu.Connection) error {
	// Implementation here
	return nil
}

// migrateEmbeddings handles the migration of embeddings.
func migrateEmbeddings(conn *kuzu.Connection) error {
	// Implementation here
	return nil
}

// validateRelType checks if the given relation type is valid.
func validateRelType(relType string) error {
	if _, ok := AllowedRelTypes[relType]; !ok {
		return errors.New("invalid relation type")
	}
	return nil
}package knowledge

import (
	"errors"
)

// AllowedRelTypes defines the valid relation types for the knowledge graph.
var AllowedRelTypes = map[string]struct{}{
	"friend": {},
	"colleague": {},
	"family": {},
	// Add other relation types
}

// initSchema initializes the knowledge graph schema.
func initSchema(conn *kuzu.Connection) error {
	// Implementation here
	return nil
}

// migrateEmbeddings handles the migration of embeddings.
func migrateEmbeddings(conn *kuzu.Connection) error {
	// Implementation here
	return nil
}

// validateRelType checks if the given relation type is valid.
func validateRelType(relType string) error {
	if _, ok := AllowedRelTypes[relType]; !ok {
		return errors.New("invalid relation type")
	}
	return nil
}