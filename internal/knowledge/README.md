# Knowledge Graph Store

This package provides a Kuzu-backed knowledge graph store for ai-pack's persistent memory system.

## Overview

The knowledge graph stores entities (files, functions, modules, etc.), relations between them (CALLS, IMPORTS, FIXES, etc.), and observations (notes, insights, learnings) attached to entities.

## Architecture

- **Database**: Kuzu embedded graph database
- **Schema**: Node tables (Entity, Observation) + Relationship tables
- **Isolation**: Per-project via `project_id` field
- **Queries**: Cypher via prepared statements

## Data Model

### Entity Node
```cypher
CREATE NODE TABLE Entity(
    id STRING,
    name STRING,
    type STRING,
    project_id STRING,
    created_at INT64,
    updated_at INT64,
    PRIMARY KEY(id)
)
```

### Observation Node
```cypher
CREATE NODE TABLE Observation(
    id STRING,
    content STRING,
    created_at INT64,
    PRIMARY KEY(id)
)
```

### Relationships
- `CALLS` - Function/method invocation
- `IMPORTS` - Module dependency
- `FIXES` - Bug fix relationship
- `SUPERSEDES` - Replacement/deprecation
- `CAUSED_BY` - Root cause linking
- `DEPENDS_ON` - General dependency
- `IMPLEMENTS` - Interface/contract implementation
- `RELATES_TO` - General association
- `TESTS` - Test coverage
- `DOCUMENTS` - Documentation relationship
- `HAS_OBSERVATION` - Entity → Observation link

## Usage

### Opening a Store

```go
store, err := knowledge.OpenStore("/path/to/db")
if err != nil {
    log.Fatal(err)
}
defer store.Close()
```

### Creating Entities

```go
entity, err := store.CreateEntity("main.go", "file", "my-project")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Created entity: %s\n", entity.ID)
```

### Creating Relations

```go
err := store.CreateRelation(funcID, targetID, "CALLS", "my-project")
if err != nil {
    log.Fatal(err)
}
```

### Traversing Relations

```go
targets, err := store.TraverseRelations(funcID, "CALLS", "my-project")
if err != nil {
    log.Fatal(err)
}
for _, target := range targets {
    fmt.Printf("Calls: %s\n", target.Name)
}
```

### Adding Observations

```go
obs, err := store.CreateObservation(entityID, "Performance critical", "my-project")
if err != nil {
    log.Fatal(err)
}
```

## Per-Project Isolation

All operations require a `project_id` parameter. Entities, relations, and observations are isolated by project:

```go
// These are completely separate
entityA, _ := store.CreateEntity("test.go", "file", "project-a")
entityB, _ := store.CreateEntity("test.go", "file", "project-b")

// Cannot access across projects
_, err := store.GetEntity(entityA.ID, "project-b") // Returns error
```

## Testing

Run tests with CGO enabled (required for Kuzu):

```bash
CGO_ENABLED=1 \
CGO_CFLAGS="-Ilib/kuzu/darwin-arm64/include" \
CGO_LDFLAGS="-Llib/kuzu/darwin-arm64 -lkuzu -lstdc++ -lm" \
go test ./internal/knowledge/...
```

Or use the Makefile:

```bash
make test-knowledge
```

## Implementation Notes

### Timestamp Handling

Kuzu stores timestamps as INT64 microseconds since Unix epoch. The package handles conversion automatically:

```go
// Go time.Time ↔ Kuzu INT64 microseconds
createdAt := time.Now()
micros := createdAt.UnixMicro() // Store in Kuzu
retrieved := time.UnixMicro(micros) // Retrieve from Kuzu
```

### Query Safety

All queries use prepared statements with parameter binding to prevent injection:

```go
query := "MATCH (e:Entity) WHERE e.id = $id RETURN e"
stmt, _ := store.conn.Prepare(query)
result, _ := store.conn.Execute(stmt, map[string]any{"id": entityID})
```

### Resource Management

Always close resources:

```go
result, err := store.Execute(query)
if err != nil {
    return err
}
defer result.Close() // Always close result sets

if result.HasNext() {
    tuple, _ := result.Next()
    defer tuple.Close() // Always close tuples
    // ... process tuple
}
```

## Future Enhancements

- Vector embeddings for semantic search (phase 2)
- Full-text search on entity names and observations (phase 3)
- Graph visualization export (phase 4)
- Backup/restore utilities (phase 5)

## References

- [Kuzu Documentation](https://docs.kuzudb.com/)
- [Architecture Decision Record](../../docs/adr/003-knowledge-graph.md)
- [Knowledge Graph Design](../../docs/architecture/knowledge-graph.md)
