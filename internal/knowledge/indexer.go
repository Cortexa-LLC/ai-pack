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

	"github.com/cortexa-llc/ai-pack/pkg/gitignore"
)

// Indexer scans source files and populates the knowledge graph
type Indexer struct {
	store      *Store
	projectID  string
	root       string
	ignorer    *gitignore.Ignorer
}

// IndexStats tracks indexing progress
type IndexStats struct {
	FilesScanned      int
	EntitiesCreated   int
	RelationsCreated  int
	Errors            int
}

// relationRecord holds relation data before batch insert
type relationRecord struct {
	FromID string
	ToID   string
	Type   string
}

// NewIndexer creates a new indexer
func NewIndexer(store *Store, projectID, root string) (*Indexer, error) {
	// Load ignore patterns
	ignorer, err := gitignore.NewIgnorer(root)
	if err != nil {
		return nil, fmt.Errorf("load ignore patterns: %w", err)
	}

	return &Indexer{
		store:     store,
		projectID: projectID,
		root:      root,
		ignorer:   ignorer,
	}, nil
}

// Index scans the project and populates the knowledge graph
func (idx *Indexer) Index() (*IndexStats, error) {
	stats := &IndexStats{}

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

	// Write CSV headers
	entityWriter.Write([]string{"id", "name", "type", "project_id", "created_at", "updated_at"})

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
		if idx.ignorer.Match(relPath) {
			return nil
		}

		// Process based on file type
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".go":
			if err := idx.processGoFile(path, relPath, entityWriter, &relations, stats); err != nil {
				fmt.Printf("Warning: Failed to process %s: %v\n", relPath, err)
				stats.Errors++
			}
			stats.FilesScanned++

		case ".ts", ".tsx", ".js", ".jsx":
			if err := idx.processJSFile(path, relPath, "", entityWriter, &relations, stats); err != nil {
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

// processGoFile extracts structural information from Go source files
func (idx *Indexer) processGoFile(absPath, relPath string, entityWriter *csv.Writer, relations *[]relationRecord, stats *IndexStats) error {
	now := time.Now().UTC().Format(time.RFC3339)

	// Parse the Go file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, absPath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse file: %w", err)
	}

	// Create file entity
	fileID := fmt.Sprintf("file:%s", relPath)
	entityWriter.Write([]string{fileID, relPath, "file", idx.projectID, now, now})
	stats.EntitiesCreated++

	// Create package entity
	pkgName := node.Name.Name
	pkgID := fmt.Sprintf("package:%s", pkgName)
	entityWriter.Write([]string{pkgID, pkgName, "package", idx.projectID, now, now})
	stats.EntitiesCreated++

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
		entityWriter.Write([]string{importID, importPath, "package", idx.projectID, now, now})
		stats.EntitiesCreated++

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
				funcID := fmt.Sprintf("function:%s.%s", pkgName, funcName)

				entityWriter.Write([]string{funcID, funcName, "function", idx.projectID, now, now})
				stats.EntitiesCreated++

				// Function defined in file
				*relations = append(*relations, relationRecord{
					FromID: funcID,
					ToID:   fileID,
					Type:   "DEFINED_IN",
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

						entityWriter.Write([]string{typeID, typeName, "type", idx.projectID, now, now})
						stats.EntitiesCreated++

						// Type defined in file
						*relations = append(*relations, relationRecord{
							FromID: typeID,
							ToID:   fileID,
							Type:   "DEFINED_IN",
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
func (idx *Indexer) processJSFile(absPath, relPath, fileID string, entityWriter *csv.Writer, relations *[]relationRecord, stats *IndexStats) error {
	now := time.Now().UTC().Format(time.RFC3339)

	// Read file content
	content, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	// Create file entity
	if fileID == "" {
		fileID = fmt.Sprintf("file:%s", relPath)
		entityWriter.Write([]string{fileID, relPath, "file", idx.projectID, now, now})
		stats.EntitiesCreated++
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
		entityWriter.Write([]string{importID, importPath, "module", idx.projectID, now, now})
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
	query := fmt.Sprintf(`
		COPY Entity FROM '%s' (HEADER=true)
	`, entitiesPath)

	result, err := idx.store.Execute(query)
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
