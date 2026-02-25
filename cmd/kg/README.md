# kg - Knowledge Graph CLI

Command-line interface for the ai-pack knowledge graph system.

## Installation

```bash
go build -o kg ./cmd/kg
```

## Usage

### Index Codebase

Scan the codebase and populate the knowledge graph with structural data:

```bash
./kg index
```

Output:
```
🔍 Indexing codebase at /path/to/repo...
✅ Indexing complete!
   Files scanned:     191
   Entities created:  1113
   Relations created: 1517
   Duration:          5.156s
```

**What gets indexed:**
- Go packages, files, functions, types, imports
- File structure and organization
- Import relationships (IMPORTS relations)

**What is NOT indexed:**
- Files matching .gitignore patterns
- Files matching .claudeignore patterns
- Binary files
- Generated files

### Query the Graph

Execute Cypher queries against the knowledge graph:

```bash
./kg query "MATCH (f:Entity {type:'file'}) RETURN count(f)"
```

**Sample Queries:**

Count all files:
```bash
./kg query "MATCH (f:Entity {type:'file'}) RETURN count(f)"
```

List imports for a specific file:
```bash
./kg query "MATCH (f:Entity {name:'cmd/kg/main.go'})-[:IMPORTS]->(p:Entity) RETURN p.name"
```

Entity type distribution:
```bash
./kg query "MATCH (e:Entity) RETURN e.type, count(*) ORDER BY count(*) DESC"
```

Find all files that import a package:
```bash
./kg query "MATCH (f:Entity)-[:IMPORTS]->(p:Entity {name:'fmt'}) RETURN f.name LIMIT 10"
```

### Help

```bash
./kg --help
```

## Knowledge Graph Schema

### Entity Types

- **file** - Source code files (e.g., `cmd/kg/main.go`)
- **package** - Go packages (e.g., `main`)
- **module** - Go modules (e.g., `github.com/cortexa-llc/ai-pack`)
- **function** - Functions and methods (e.g., `main.main`)
- **type** - Type declarations (e.g., `Store`)

### Relation Types

- **IMPORTS** - Import relationships between files/packages
- **CALLS** - Function call relationships (future)
- **CONTAINS** - Containment relationships (future)
- **IMPLEMENTS** - Interface implementation (future)
- **EXTENDS** - Type extension relationships (future)

### Entity Properties

- `name` - Primary identifier (file path, package name, etc.)
- `type` - Entity type (file, package, function, type, module)
- `path` - File system path (for file entities)
- `project` - Project name (currently "ai-pack")

## Performance

Indexing the ai-pack repository (191 files):
- **Scan time:** ~3 seconds
- **Load time:** ~2 seconds
- **Total:** ~5 seconds

The indexer uses CSV bulk loading via Kuzu's COPY FROM for optimal performance.

## Database Location

The knowledge graph database is stored in:
```
.kuzu/
```

This directory is automatically created on first index and should be added to .gitignore.

## Development

To add new entity or relation types:

1. Update `internal/knowledge/store.go` schema
2. Add extraction logic in `internal/knowledge/indexer.go`
3. Write CSV rows via `writeEntity()` or `writeRelation()`
4. Rebuild and test: `go build ./cmd/kg && ./kg index`

## Troubleshooting

**Error: "database already open"**
- Another kg process is running
- Close other instances and retry

**Error: "permission denied"**
- Check file permissions on .kuzu/ directory
- Ensure write access to working directory

**Empty results after indexing**
- Check .gitignore and .claudeignore patterns
- Verify files are not excluded from indexing
- Run with verbose logging (future feature)

**Import relations missing**
- Ensure Go files have valid syntax
- Check AST parsing errors in logs
- Verify import statements are standard format

## Future Enhancements

- [ ] Progress bar for large repositories
- [ ] Verbose logging mode
- [ ] Incremental indexing (only changed files)
- [ ] Function call graph extraction
- [ ] Type relationship extraction
- [ ] Documentation comment extraction
- [ ] Test coverage integration
