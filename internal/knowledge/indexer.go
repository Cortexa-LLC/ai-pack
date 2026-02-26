package knowledge

import (
	"encoding/csv"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	gitignore "github.com/sabhiram/go-gitignore"
)

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
	relationTypes := []string{
		"CALLS", "IMPORTS", "FIXES", "SUPERSEDES", "CAUSED_BY",
		"DEPENDS_ON", "IMPLEMENTS", "RELATES_TO", "TESTS", "DOCUMENTS",
	}

	for _, relType := range relationTypes {
		query := fmt.Sprintf(`
			MATCH (from:Entity {project_id: '%s'})-[r:%s]->(to:Entity {project_id: '%s'})
			DELETE r
		`, idx.projectID, relType, idx.projectID)

		result, err := idx.store.query(query)
		if err != nil {
			return fmt.Errorf("delete %s relations: %w", relType, err)
		}
		result.Close()
	}

	// Then delete entities
	query := fmt.Sprintf(`
		MATCH (e:Entity {project_id: '%s'})
		DELETE e
	`, idx.projectID)

	result, err := idx.store.query(query)
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

	// Create temporary CSV files
	entitiesPath := filepath.Join(os.TempDir(), fmt.Sprintf("kg-entities-%d.csv", time.Now().Unix()))
	defer os.Remove(entitiesPath)

	entitiesFile, err := os.Create(entitiesPath)
	if err != nil {
		return nil, fmt.Errorf("create entities CSV: %w", err)
	}
	defer entitiesFile.Close()

	entityWriter := csv.NewWriter(entitiesFile)
	defer entityWriter.Flush()

	// Write CSV headers (excluding embedding column)
	entityWriter.Write([]string{"id", "name", "type", "project_id", "created_at", "updated_at"})

	// Track seen entities to avoid duplicates
	seenEntities := make(map[string]bool)

	// Collect relations in memory (will batch insert via Cypher)
	var relations []relationRecord

	// Walk the project directory
	err = filepath.Walk(idx.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(idx.root, path)
		if err != nil {
			return err
		}

		// Check if ignored
		if idx.ignorer.MatchesPath(relPath) {
			return nil
		}

		// Process based on file type
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".go":
			if err := idx.processGoFile(path, relPath, entityWriter, seenEntities, &relations, stats); err != nil {
				fmt.Printf("Warning: Failed to process %s: %v\n", relPath, err)
				stats.Errors++
			}
			stats.FilesScanned++

		case ".ts", ".tsx", ".js", ".jsx":
			if err := idx.processJSFile(path, relPath, "", entityWriter, seenEntities, &relations, stats); err != nil {
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

	// Flush CSV writer
	entityWriter.Flush()
	if err := entityWriter.Error(); err != nil {
		return nil, fmt.Errorf("flush entities CSV: %w", err)
	}
	entitiesFile.Close()

	// Bulk load entities
	if err := idx.bulkLoadEntities(entitiesPath); err != nil {
		return nil, fmt.Errorf("bulk load entities: %w", err)
	}

	// Batch create relations
	if err := idx.batchCreateRelations(relations, stats); err != nil {
		return nil, fmt.Errorf("batch create relations: %w", err)
	}

	return stats, nil
}

// writeEntity writes an entity to CSV if not already seen
func writeEntity(writer *csv.Writer, seen map[string]bool, id, name, typ, projectID, createdAt, updatedAt string) bool {
	if seen[id] {
		return false // Already seen
	}
	seen[id] = true // Mark as seen BEFORE writing
	writer.Write([]string{id, name, typ, projectID, createdAt, updatedAt})
	return true
}

// processGoFile extracts structural information from Go source files
func (idx *Indexer) processGoFile(absPath, relPath string, entityWriter *csv.Writer, seenEntities map[string]bool, relations *[]relationRecord, stats *IndexStats) error {
	now := time.Now().UTC().Format(time.RFC3339)

	// Parse the Go file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, absPath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse file: %w", err)
	}

	// Create file entity
	fileID := fmt.Sprintf("file:%s", relPath)
	if writeEntity(entityWriter, seenEntities, fileID, relPath, "file", idx.projectID, now, now) {
		stats.EntitiesCreated++
	}

	// Create package entity
	pkgName := node.Name.Name
	pkgID := fmt.Sprintf("package:%s", pkgName)
	if writeEntity(entityWriter, seenEntities, pkgID, pkgName, "package", idx.projectID, now, now) {
		stats.EntitiesCreated++
	}

	// File belongs to package
	*relations = append(*relations, relationRecord{
		FromID: fileID,
		ToID:   pkgID,
		Type:   "BELONGS_TO",
	})
	stats.RelationsCreated++

	// Extract imports
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")
		importID := fmt.Sprintf("package:%s", importPath)

		// Create import entity if not exists
		if writeEntity(entityWriter, seenEntities, importID, importPath, "package", idx.projectID, now, now) {
			stats.EntitiesCreated++
		}

		// File imports package
		*relations = append(*relations, relationRecord{
			FromID: fileID,
			ToID:   importID,
			Type:   "IMPORTS",
		})
		stats.RelationsCreated++
	}

	// Extract declarations
	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			// Function declaration
			if d.Name != nil && d.Name.IsExported() {
				funcName := d.Name.Name
				// Use file path in function ID to ensure uniqueness across files
				funcID := fmt.Sprintf("function:%s:%s", relPath, funcName)

				if writeEntity(entityWriter, seenEntities, funcID, funcName, "function", idx.projectID, now, now) {
					stats.EntitiesCreated++
				}

				// File contains function
				*relations = append(*relations, relationRecord{
					FromID: fileID,
					ToID:   funcID,
					Type:   "CONTAINS",
				})
				stats.RelationsCreated++
			}

		case *ast.GenDecl:
			// Type or variable declaration
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name.IsExported() {
						typeName := s.Name.Name
						typeID := fmt.Sprintf("type:%s.%s", pkgName, typeName)

						writeEntity(entityWriter, seenEntities, typeID, typeName, "type", idx.projectID, now, now)
						stats.EntitiesCreated++

						// File contains type
						*relations = append(*relations, relationRecord{
							FromID: fileID,
							ToID:   typeID,
							Type:   "CONTAINS",
						})
						stats.RelationsCreated++
					}
				}
			}
		}
	}

	return nil
}

// processJSFile extracts imports from TypeScript/JavaScript files using regex
func (idx *Indexer) processJSFile(absPath, relPath, fileID string, entityWriter *csv.Writer, seenEntities map[string]bool, relations *[]relationRecord, stats *IndexStats) error {
	now := time.Now().UTC().Format(time.RFC3339)

	// Read file content
	content, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	// Create file entity
	if fileID == "" {
		fileID = fmt.Sprintf("file:%s", relPath)
		if writeEntity(entityWriter, seenEntities, fileID, relPath, "file", idx.projectID, now, now) {
			stats.EntitiesCreated++
		}
	}

	// Extract imports using regex
	// Matches: import ... from "path"  or  import("path")
	importRegex := regexp.MustCompile(`import\s+(?:[\w\s{},*]*\s+from\s+)?["']([^"']+)["']|import\s*\(\s*["']([^"']+)["']\s*\)`)
	matches := importRegex.FindAllSubmatch(content, -1)

	seenImports := make(map[string]bool)
	for _, match := range matches {
		var importPath string
		if len(match) > 1 && len(match[1]) > 0 {
			importPath = string(match[1])
		} else if len(match) > 2 && len(match[2]) > 0 {
			importPath = string(match[2])
		}

		if importPath == "" || seenImports[importPath] {
			continue
		}
		seenImports[importPath] = true

		// Create module entity
		importID := fmt.Sprintf("module:%s", importPath)
		writeEntity(entityWriter, seenEntities, importID, importPath, "module", idx.projectID, now, now)
		stats.EntitiesCreated++

		*relations = append(*relations, relationRecord{
			FromID: fileID,
			ToID:   importID,
			Type:   "IMPORTS",
		})
		stats.RelationsCreated++
	}

	return nil
}

// bulkLoadEntities loads entities from CSV file into Kuzu using COPY FROM
func (idx *Indexer) bulkLoadEntities(entitiesPath string) error {
	// Remove the embedding column from header since we'll only load the first 6 columns
	query := fmt.Sprintf(`
		COPY Entity(id, name, type, project_id, created_at, updated_at) FROM '%s' (HEADER=true)
	`, entitiesPath)

	result, err := idx.store.query(query)
	if err != nil {
		return fmt.Errorf("load entities: %w", err)
	}
	result.Close()

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
