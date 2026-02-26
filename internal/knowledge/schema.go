package knowledge

import (
	"errors"
)

// AllowedRelTypes is the canonical list of relation types in the system.
var AllowedRelTypes = []string{
	"parent",
	"child",
	"sibling",
	"friend",
}

// initSchema initializes the schema for the knowledge graph.
func initSchema() {
	// Implementation of schema initialization goes here
}

// validateRelType checks if the provided relation type is valid.
func validateRelType(relType string) error {
	for _, validType := range AllowedRelTypes {
		if relType == validType {
			return nil
		}
	}
	return errors.New("invalid relation type")
}