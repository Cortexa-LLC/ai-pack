package knowledge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	gitignore "github.com/sabhiram/go-gitignore"
)

// alwaysSkipDirs is the set of directory base-names that are never worth
// indexing, regardless of what .gitignore says. Pruning these early avoids
// walking into directories that can have thousands of files (node_modules) or
// that contain binary/runtime data rather than source (.git, .claude).
var alwaysSkipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	".claude":      true,
	".beads":       true,
	"dist":         true,
	"build":        true,
	".build":       true,
	"__pycache__":  true,
	".mypy_cache":  true,
	".pytest_cache": true,
	".next":        true,
	".nuxt":        true,
	"target":       true, // Rust/Maven build output
	"coverage":     true,
}

// Indexer scans source files and populates the knowledge graph
type Indexer struct {
	store     *Store
	projectID string
	root      string
	ignorer   *gitignore.GitIgnore
}

// IndexStats tracks indexing progress
type IndexStats struct {
	FilesScanned     int
	EntitiesCreated  int
	RelationsCreated int
	Errors           int
}

// entityRecord holds entity data before batch insert
type entityRecord struct {
	ID        string
	Name      string
	Type      string
	ProjectID string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// relationRecord holds relation data before batch insert
type relationRecord struct {
	FromID string
	ToID   string
	Type   string
}

// NewIndexer creates a new indexer
func NewIndexer(store *Store, projectID, root string) (*Indexer, error) {
	// Load ignore patterns from .gitignore and .claudeignore
	var ignorer *gitignore.GitIgnore
	gitignorePath := filepath.Join(root, ".gitignore")
	claudeignorePath := filepath.Join(root, ".claudeignore")

	// Try .gitignore first
	if _, err := os.Stat(gitignorePath); err == nil {
		ignorer, err = gitignore.CompileIgnoreFile(gitignorePath)
		if err != nil {
			return nil, fmt.Errorf("load .gitignore: %w", err)
		}
	}

	// Merge with .claudeignore if it exists
	if _, err := os.Stat(claudeignorePath); err == nil {
		if ignorer != nil {
			// Both files exist - merge them
			combined, err := gitignore.CompileIgnoreFileAndLines(gitignorePath, readLinesFromFile(claudeignorePath)...)
			if err != nil {
				return nil, fmt.Errorf("load .claudeignore: %w", err)
			}
			ignorer = combined
		} else {
			// Only .claudeignore exists
			ignorer, err = gitignore.CompileIgnoreFile(claudeignorePath)
			if err != nil {
				return nil, fmt.Errorf("load .claudeignore: %w", err)
			}
		}
	}

	// If no ignore files, create empty ignorer
	if ignorer == nil {
		ignorer = gitignore.CompileIgnoreLines()
	}

	return &Indexer{
		store:     store,
		projectID: projectID,
		root:      root,
		ignorer:   ignorer,
	}, nil
}

// readLinesFromFile reads all lines from a file
func readLinesFromFile(path string) []string {
	content, err := os.ReadFile(path)
	if err != nil {
		return []string{}
	}
	lines := strings.Split(string(content), "\n")
	// Filter out empty lines and comments
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			result = append(result, line)
		}
	}
	return result
}

// clearProjectData removes all entities and relations for this project
func (idx *Indexer) clearProjectData() error {
	// Delete all relation types for this project
	relationTypes := AllowedRelTypes

	for _, relType := range relationTypes {
		// relType is from a hardcoded list above, not user input.
		query := fmt.Sprintf(`
			MATCH (from:Entity {project_id: $project_id})-[r:%s]->(to:Entity {project_id: $project_id})
			DELETE r
		`, relType)

		result, err := idx.store.queryParams(query, map[string]any{"project_id": idx.projectID})
		if err != nil {
			return fmt.Errorf("delete %s relations: %w", relType, err)
		}
		result.Close()
	}

	// Then delete entities
	result, err := idx.store.queryParams(`
		MATCH (e:Entity {project_id: $project_id})
		DELETE e
	`, map[string]any{"project_id": idx.projectID})
	if err != nil {
		return fmt.Errorf("delete entities: %w", err)
	}
	result.Close()

	return nil
}

// Index scans the project and populates the knowledge graph
func (idx *Indexer) Index() (*IndexStats, error) {
	stats := &IndexStats{}

	// Clear existing data for this project (rebuild from scratch)
	if err := idx.clearProjectData(); err != nil {
		return nil, fmt.Errorf("clear existing data: %w", err)
	}

	// Collect entities and relations in memory; insert via parameterized Cypher
	// to avoid all CSV quoting/delimiter issues with complex entity names.
	var entities []entityRecord
	seenEntities := make(map[string]bool)
	var relations []relationRecord

	// Walk the project directory
	err := filepath.Walk(idx.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path (needed for both directory and file decisions)
		relPath, err := filepath.Rel(idx.root, path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip the root itself
			if relPath == "." {
				return nil
			}
			// Always-skip directories that are never useful to index
			base := info.Name()
			if alwaysSkipDirs[base] {
				return filepath.SkipDir
			}
			// Apply gitignore / claudeignore to directories so we prune entire subtrees
			// (MatchesPath on a dir path skips the whole subtree via SkipDir)
			if idx.ignorer.MatchesPath(relPath) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if ignored
		if idx.ignorer.MatchesPath(relPath) {
			return nil
		}

		// Process based on file type
		ext := strings.ToLower(filepath.Ext(path))
		if cfg, ok := langRegistry[ext]; ok {
			if err := idx.processWithTreeSitter(path, relPath, cfg, &entities, seenEntities, &relations, stats); err != nil {
				fmt.Printf("Warning: Failed to process %s: %v\n", relPath, err)
				stats.Errors++
			}
			stats.FilesScanned++
		} else if asmMatchesPath(path) {
			if err := idx.processAsmFile(path, relPath, &entities, seenEntities, &relations, stats); err != nil {
				fmt.Printf("Warning: Failed to process %s: %v\n", relPath, err)
				stats.Errors++
			}
			stats.FilesScanned++
		} else if ext == ".md" {
			if err := idx.processMarkdownFile(path, relPath, &entities, seenEntities, &relations, stats); err != nil {
				fmt.Printf("Warning: Failed to process %s: %v\n", relPath, err)
				stats.Errors++
			}
			stats.FilesScanned++
		} else if ext == ".yaml" || ext == ".yml" {
			if err := idx.processYAMLFile(path, relPath, &entities, seenEntities, &relations, stats); err != nil {
				fmt.Printf("Warning: Failed to process %s: %v\n", relPath, err)
				stats.Errors++
			}
			stats.FilesScanned++
		} else if ext == ".html" || ext == ".htm" {
			if err := idx.processHTMLFile(path, relPath, &entities, seenEntities, &relations, stats); err != nil {
				fmt.Printf("Warning: Failed to process %s: %v\n", relPath, err)
				stats.Errors++
			}
			stats.FilesScanned++
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walk directory: %w", err)
	}

	// Batch create entities via parameterized Cypher (immune to special characters in names)
	if err := idx.batchCreateEntities(entities); err != nil {
		return nil, fmt.Errorf("batch create entities: %w", err)
	}

	// Batch create relations
	if err := idx.batchCreateRelations(relations, stats); err != nil {
		return nil, fmt.Errorf("batch create relations: %w", err)
	}

	return stats, nil
}

// writeEntity appends an entity to the slice if not already seen.
// ts is the timestamp to use for both created_at and updated_at.
func writeEntity(entities *[]entityRecord, seen map[string]bool, id, name, typ, projectID string, ts time.Time) bool {
	if seen[id] {
		return false
	}
	seen[id] = true
	*entities = append(*entities, entityRecord{
		ID:        id,
		Name:      name,
		Type:      typ,
		ProjectID: projectID,
		CreatedAt: ts,
		UpdatedAt: ts,
	})
	return true
}

// batchCreateEntities inserts entities in batches using parameterized Cypher.
// This avoids all CSV quoting/delimiter issues with entity names that contain
// commas, quotes, or other special characters.
// Note: EntitiesCreated is counted by callers of writeEntity, not here.
func (idx *Indexer) batchCreateEntities(entities []entityRecord) error {
	batchSize := 100
	for i := 0; i < len(entities); i += batchSize {
		end := i + batchSize
		if end > len(entities) {
			end = len(entities)
		}
		if err := idx.createEntityBatch(entities[i:end]); err != nil {
			return fmt.Errorf("entity batch %d-%d: %w", i, end, err)
		}
	}
	return nil
}

// createEntityBatch inserts a single batch of entities
func (idx *Indexer) createEntityBatch(batch []entityRecord) error {
	for _, ent := range batch {
		result, err := idx.store.queryParams(`
			CREATE (e:Entity {
				id: $id,
				name: $name,
				type: $type,
				project_id: $project_id,
				created_at: $created_at,
				updated_at: $updated_at
			})
		`, map[string]any{
			"id":         ent.ID,
			"name":       ent.Name,
			"type":       ent.Type,
			"project_id": ent.ProjectID,
			"created_at": ent.CreatedAt,
			"updated_at": ent.UpdatedAt,
		})
		if err != nil {
			// Continue on duplicate key; log unexpected errors
			fmt.Printf("Warning: insert entity %s: %v\n", ent.ID, err)
			continue
		}
		result.Close()
	}
	return nil
}

// batchCreateRelations creates relations in batches using Cypher statements
func (idx *Indexer) batchCreateRelations(relations []relationRecord, stats *IndexStats) error {
	if len(relations) == 0 {
		return nil
	}

	// Process in batches of 100
	batchSize := 100
	for i := 0; i < len(relations); i += batchSize {
		end := i + batchSize
		if end > len(relations) {
			end = len(relations)
		}

		batch := relations[i:end]
		if err := idx.createRelationBatch(batch); err != nil {
			return fmt.Errorf("batch %d-%d: %w", i, end, err)
		}
	}

	return nil
}

// createRelationBatch creates a single batch of relations
func (idx *Indexer) createRelationBatch(batch []relationRecord) error {
	for _, rel := range batch {
		// Use the Store's CreateRelation method which handles relation types properly
		if err := idx.store.CreateRelation(rel.FromID, rel.ToID, rel.Type, idx.projectID); err != nil {
			// Log but don't fail - entity might not exist yet or relation might be duplicate
			// TODO: Better error handling
			continue
		}
	}
	return nil
}
