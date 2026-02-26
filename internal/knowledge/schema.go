package knowledge

import (
	"errors"
)

// AllowedRelTypes is the centralized whitelist for relation types.
var AllowedRelTypes = []string{
	"parent",
	"child",
	"sibling",
	"friend",
	"colleague",
}

// initSchema initializes the schema with allowed relation types and other setups.
func initSchema() {
	// Schema initialization logic...
}

// validateRelType checks if a relation type is allowed.
func validateRelType(relType string) error {
	for _, allowed := range AllowedRelTypes {
		if relType == allowed {
			return nil
		}
	}
	return errors.New("invalid relType: " + relType)
}

// migrateEmbeddings migrates embeddings associated with the knowledge graph schema.
func migrateEmbeddings() {
	// Migration logic...
}