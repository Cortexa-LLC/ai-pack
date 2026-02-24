# Knowledge Graph Architecture

**Last updated:** 2026-02-24

Per-project knowledge graph that persists findings, code relationships, architectural
decisions, and incident history across agent sessions. Agents read from it via MCP tools;
the agent-server injects relevant context into the system prompt before each task starts.

For storage and build-tooling rationale see `docs/adr/003-knowledge-graph.md`.

---

## Table of Contents

1. [Motivation](#motivation)
2. [Component Overview](#component-overview)
3. [Isolation Boundary](#isolation-boundary)
4. [Storage: Kuzu](#storage-kuzu)
5. [Schema](#schema)
6. [Search Layer](#search-layer)
7. [Embeddings](#embeddings)
8. [CLI (cmd/kg)](#cli-cmdkg)
9. [MCP Server Mode](#mcp-server-mode)
10. [Pre-Flight Context Injection](#pre-flight-context-injection)
11. [Structural Indexer](#structural-indexer)
12. [Cross-Platform Builds](#cross-platform-builds)
13. [Module Restructure](#module-restructure)
14. [Execution Plan](#execution-plan)

---

## Motivation

Every agent session starts cold. The first 10–20 turns are spent re-reading files,
re-tracing call chains, and re-discovering facts already established in prior sessions.

The knowledge graph persists structured findings per project. Before each task, the
server queries it and injects relevant context into the system prompt — eliminating
investigation turns for known problem spaces.

---

## Component Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                        ai-pack project                           │
│                                                                  │
│   cmd/agent          cmd/server              cmd/kg              │
│  ┌──────────┐       ┌─────────────────┐     ┌──────────────┐    │
│  │ agent    │─HTTP─▶│  agent-server   │     │  kg CLI      │    │
│  │ CLI      │       │                 │     │              │    │
│  └──────────┘       │ ┌─────────────┐ │     │  search      │    │
│   CGO: none         │ │ pre-flight  │ │     │  show        │    │
│                     │ │ injection   │ │     │  add         │    │
│                     │ └──────┬──────┘ │     │  link        │    │
│                     │        │ MCP    │     │  query       │    │
│                     │ ┌──────▼──────┐ │     │  index       │    │
│                     │ │ mcp manager │ │     │  embed       │    │
│                     │ │ (existing)  │ │     │  server ─────┼──┐ │
│                     │ └──────┬──────┘ │     └──────────────┘  │ │
│                     └────────┼────────┘      CGO: Kuzu        │ │
│   CGO: none                  │ stdio MCP                      │ │
│                              └────────────────────────────────┘ │
│                                                                  │
│                         .ai/knowledge.db  (Kuzu, per-project)   │
└──────────────────────────────────────────────────────────────────┘
```

---

## Isolation Boundary

**Kuzu lives only in `cmd/kg`.** The agent-server has no direct dependency on Kuzu or
CGO. It communicates with `kg` via the MCP protocol — the same mechanism used for all
other MCP tools. This means:

- `cmd/agent` — CGO-free, pure Go
- `cmd/server` — CGO-free, pure Go (MCP client to kg)
- `cmd/kg` — CGO + Kuzu static library (owns the graph)

The agent-server spawns `kg server` as a child MCP process at startup (via the existing
`mcpManager`), exactly like any other MCP server. Agents receive `search_knowledge`,
`add_entity`, and other tools transparently.

---

## Storage: Kuzu

[Kuzu](https://kuzudb.com) is an embedded property graph database — the graph equivalent
of SQLite. It runs in-process, stores data in `.ai/knowledge.db`, and requires no server.

Kuzu was chosen over SQLite + a hand-built Cypher translator because:
- Native openCypher support — no translator to write or maintain
- Columnar storage optimised for multi-hop graph traversals
- Shortest-path and variable-length path queries are first-class
- Schema-enforced node and relationship tables
- `COPY FROM CSV` for bulk structural indexing
- Kuzu Explorer (web UI) ships free for graph visualization

See `docs/adr/003-knowledge-graph.md` for the full SQLite vs Kuzu decision record.

### Go Binding

`github.com/kuzudb/go-kuzu` — official CGO binding. A platform-specific static library
(`libkuzu.a`) is downloaded by `make setup-kuzu` and linked at build time. No shared
library needed at runtime.

---

## Schema

Node tables and relationship tables are defined once at database creation. Kuzu enforces
types and primary keys.

```cypher
-- Node tables
CREATE NODE TABLE Entity (
    id         STRING,
    project_id STRING,
    type       STRING,   -- file | function | module | concept | incident | decision | task
    name       STRING,
    summary    STRING,
    embedding  FLOAT[1536],   -- NULL until kg embed runs
    created_at INT64,
    updated_at INT64,
    source_task STRING,
    PRIMARY KEY (id)
);

-- Relationship tables
CREATE REL TABLE CALLS       (FROM Entity TO Entity, weight DOUBLE, source_task STRING);
CREATE REL TABLE IMPORTS     (FROM Entity TO Entity, source_task STRING);
CREATE REL TABLE FIXES       (FROM Entity TO Entity, source_task STRING);
CREATE REL TABLE SUPERSEDES  (FROM Entity TO Entity, source_task STRING);
CREATE REL TABLE CAUSED_BY   (FROM Entity TO Entity, source_task STRING);
CREATE REL TABLE DEPENDS_ON  (FROM Entity TO Entity, weight DOUBLE, source_task STRING);
CREATE REL TABLE IMPLEMENTS  (FROM Entity TO Entity, source_task STRING);
CREATE REL TABLE RELATED_TO  (FROM Entity TO Entity, source_task STRING);

-- Observations: evidence attached to entities
CREATE NODE TABLE Observation (
    id          STRING,
    entity_id   STRING,
    content     STRING,
    embedding   FLOAT[1536],
    created_at  INT64,
    source_task STRING,
    PRIMARY KEY (id)
);
CREATE REL TABLE HAS_OBSERVATION (FROM Entity TO Observation);
```

### Cypher Examples

```cypher
-- Find all functions called by a file within 3 hops
MATCH (f:Entity {type: "file", name: "server_core.go"})-[:CALLS*1..3]->(fn:Entity)
RETURN fn.name, fn.summary

-- What incidents are related to a concept?
MATCH (c:Entity {type: "concept"})-[:RELATED_TO]-(i:Entity {type: "incident"})
WHERE c.name CONTAINS "grade"
RETURN c.name, i.name, i.summary

-- Shortest path between two entities
MATCH p = shortestPath(
    (a:Entity {id: $from})-[*]-(b:Entity {id: $to})
) RETURN p

-- Recent observations for a file
MATCH (f:Entity {id: $file_id})-[:HAS_OBSERVATION]->(o:Observation)
RETURN o.content, o.created_at
ORDER BY o.created_at DESC LIMIT 10
```

---

## Search Layer

Hybrid search combines keyword matching and vector similarity.

### Keyword Search

Kuzu's `CONTAINS` predicate on `name` and `summary` fields. For stemmed full-text search,
a lightweight secondary index may be added later — `CONTAINS` is sufficient for v1.

```cypher
MATCH (n:Entity {project_id: $project})
WHERE n.name CONTAINS $term OR n.summary CONTAINS $term
RETURN n ORDER BY n.updated_at DESC LIMIT 20
```

### Vector Search

Embeddings stored as `FLOAT[1536]` on Entity and Observation nodes. Cosine similarity
computed in Go over the candidate set returned by keyword search:

```go
func cosineSim(a, b []float32) float32 {
    var dot, na, nb float32
    for i := range a {
        dot += a[i] * b[i]; na += a[i]*a[i]; nb += b[i]*b[i]
    }
    return dot / (sqrt32(na) * sqrt32(nb))
}
```

### Hybrid Ranking

```
score = α * keyword_match + (1-α) * cosine_similarity + β * recency_boost
```

Default α=0.4, β=0.1. Tunable per role via `knowledge.toml`.

---

## Embeddings

### Interface

```go
type Embedder interface {
    Embed(ctx context.Context, texts []string) ([][]float32, error)
    Dims() int
    ModelID() string
}
```

### Implementations

| Backend | Model | Dims | Cost |
|---------|-------|------|------|
| OpenAI (default) | `text-embedding-3-small` | 1536 | ~$0.02/1M tokens |
| Ollama (local) | `nomic-embed-text` | 768 | Free |

Configured via `KNOWLEDGE_EMBED_MODEL` env var or `.ai/knowledge.toml`.

### When Generated

**Lazy** (default): entities written without embeddings; `kg embed` batch-processes all
un-embedded nodes. Avoids blocking writes on API latency.

**Eager** (opt-in `--embed-eager`): embedding generated inline at write time.

### What Gets Embedded

| Content | Vectorized |
|---------|-----------|
| Entity name + summary | ✅ |
| Observation content | ✅ |
| Task outcome summaries | ✅ |
| Relations | ❌ structural, not semantic |
| Raw file content | ❌ use code tools |

---

## CLI (cmd/kg)

Binary: `kg`. Built by `make build-kg`. Requires Kuzu static library.

```
kg search <query>                   hybrid search (keyword + vector)
kg show <entity-id>                 entity + relations + observations
kg add entity --type <t> --name <n> [--summary <s>] [--embed-eager]
kg add observation <entity-id> <text>
kg link <from-id> --rel <relation> <to-id>
kg query "<cypher>"                 raw Cypher via Kuzu
kg graph <entity-id> [--depth N]    neighborhood as ASCII tree or mermaid
kg index [--path <dir>]             scan codebase → structural entities
kg embed [--batch N]                batch-generate missing embeddings
kg server [--port N] [--stdio]      start MCP server (stdio or SSE)
kg gc [--dry-run]                   prune orphaned/stale nodes
kg stats                            counts: nodes, edges, observations
kg export [--format json|cypher]    dump for backup/migration
```

### Config: `.ai/knowledge.toml`

```toml
[store]
db_path = ".ai/knowledge.db"

[embeddings]
model      = "text-embedding-3-small"
eager      = false
batch_size = 100

[search]
alpha = 0.4   # keyword weight
beta  = 0.1   # recency weight

[mcp]
stdio = true  # default: stdio transport for agent-server child process
```

---

## MCP Server Mode

`kg server --stdio` starts `kg` as an MCP server over stdin/stdout. The agent-server
spawns it as a child process via `mcpManager` — identical to how other MCP tools are
registered today.

### MCP Tools Exposed

| Tool | Description |
|------|-------------|
| `search_knowledge` | Hybrid search. Returns top-N entities with observations. |
| `add_entity` | Create or upsert a node. |
| `add_observation` | Attach evidence to an entity. |
| `link_entities` | Create a relation between two entities. |
| `get_file_context` | All entities + observations for a file path. |
| `query_graph` | Execute raw Cypher. |
| `get_preflight_context` | Assemble pre-flight block for a task description. |

### Agent-Server Registration

```json
// agent-server.json — existing mcp_servers section
{
  "mcp_servers": [
    {
      "name": "knowledge",
      "command": "kg",
      "args": ["server", "--stdio"],
      "env": { "KUZU_DB": ".ai/knowledge.db" }
    }
  ]
}
```

Roles opt in by adding `knowledge` to their `**Tools:**` list:
```
**Tools:** read, grep, glob, bash, write, knowledge
```

---

## Pre-Flight Context Injection

Before the agentic loop starts, the agent-server calls `get_preflight_context` via the
knowledge MCP server and prepends the result to the system prompt.

### Flow

```
task description
      │
      ▼
server calls get_preflight_context(task_description, project_id, limit=15)
      │  (MCP call to kg child process)
      ▼
kg: extract key terms → hybrid search → rank → format block
      │
      ▼
"## Project Knowledge\n..." string returned to server
      │
      ▼
server prepends block to system prompt
      │
      ▼
agentic loop starts with context already loaded
```

### Injected Block Format

```markdown
## Project Knowledge (2026-02-24 · ai-pack)

**[concept] performance-grade-selection**
Grade-based model selection using LiveBench coding scores.
- recalculateGrade had no source awareness — 1 run overwrote LiveBench Grade D with A.
  Fixed with minSamplesForRuntimeGrade=5. (task: ai-pack-49c6)
- defaultModel="gpt-4o-mini" in server_core.go:57 caused grade selector bypass for all
  unpinned roles. Fixed 2026-02-24. (task: ai-pack-asy)

**[file] internal/server/server_core.go**
AgentServer struct, configuration defaults, defaultModel constant.
- Contains grade selector initialization and MCP manager setup.

**[incident] ai-pack-1g9-gpt4o-default-model**
Spelunker selected gpt-4o-mini despite Grade D LiveBench seed.
Root cause: defaultModel constant at server_core.go:57.
```

---

## Structural Indexer

`kg index` populates Tier 1 (structural) entities from static analysis. No LLM calls.

### Go AST Scanner

Uses `go/ast` + `go/parser` from stdlib:
- **Entities**: packages, files, exported functions, types, variables
- **Relations**: `IMPORTS`, `CALLS` (best-effort), `IMPLEMENTS`

### File Graph Scanner

Language-agnostic for non-Go files:
- TypeScript/JS: regex import extraction → `IMPORTS` relations
- All files: directory hierarchy → `CONTAINS` relations

### Bulk Load

Structural scan outputs CSVs, loaded via Kuzu's `COPY FROM`:
```cypher
COPY Entity FROM 'entities.csv'
COPY IMPORTS FROM 'imports.csv'
```
Scanning 1000+ files takes seconds with bulk load vs. thousands of individual inserts.

---

## Cross-Platform Builds

CGO is required only for `cmd/kg`. `cmd/server` and `cmd/agent` are pure Go.

### Platform Matrix

| Platform | Supported |
|----------|-----------|
| `darwin/arm64` | ✅ |
| `darwin/amd64` | ✅ |
| `linux/amd64` | ✅ |
| `linux/arm64` | ✅ |
| `windows/amd64` | Deferred |

### Static Library Setup

```bash
make setup-kuzu          # downloads libkuzu.a for host platform
make build               # builds all three binaries
```

`scripts/download-kuzu.sh` fetches the pre-built static library from Kuzu's GitHub
releases into `lib/kuzu/<platform>/`. No CMake, no building Kuzu from source.

### CI Strategy

Native runners per platform — no cross-compilation, no QEMU:

```
ubuntu-24.04      → linux/amd64
ubuntu-24.04-arm  → linux/arm64
macos-15          → darwin/arm64
macos-13          → darwin/amd64
```

`cmd/agent` (pure Go, CGO-free) additionally builds for `windows/amd64`.

---

## Module Restructure

Prerequisite for all knowledge graph work. The Go module moves from `a2a-agent/` to the
project root. Source directories reorganize; **compiled binary names do not change**.

```
BEFORE                               AFTER
──────────────────────────────────   ────────────────────────────────────
ai-pack/                             ai-pack/
  a2a-agent/                           go.mod  (github.com/cortexa-llc/ai-pack)
    go.mod  ← real module              cmd/
    cmd/                                 agent/        (binary: agent)
      agent/                             server/       (binary: agent-server)
      agent-server/                      kg/           (binary: kg)  ← new
    internal/                          internal/
      server/                            server/
      streaming/                         streaming/
      monitoring/                        monitoring/
      ...                                knowledge/    ← new
  go.mod  ← stale, removed            lib/
                                         kuzu/         ← platform static libs
```

Import path: `github.com/cortexa-llc/ai-pack/a2a-agent/internal/X`
→ `github.com/cortexa-llc/ai-pack/internal/X`

---

## Execution Plan

### Phase Dependencies

```
Phase 0: Module Restructure   (ai-pack-asy)
    └── blocks all other phases

Phase 1: Kuzu Store + Schema  (ai-pack-8ze)
    └── depends on Phase 0
    └── blocks Phase 2, 3, 4, 5, 7

Phase 2: Search Layer         (ai-pack-gwe)   ┐
Phase 3: Embeddings           (ai-pack-bjj)   ├── parallel, all depend on Phase 1
Phase 7: Structural Indexer   (ai-pack-g2e)   ┘

Phase 4: CLI                  (ai-pack-21j)   ┐
Phase 5: MCP Server           (ai-pack-iyx)   ┘ parallel, depend on Phase 1+2+3

Phase 6: Pre-Flight Injection (ai-pack-r1g)
    └── depends on Phase 5
```

### Agent Assignments

| Phase | Beads ID | Description | Parallel with |
|-------|----------|-------------|---------------|
| 0 | ai-pack-asy | Module restructure | — |
| 1 | ai-pack-8ze | Kuzu store + schema | — |
| 2 | ai-pack-gwe | Search layer | 3, 7 |
| 3 | ai-pack-bjj | Embeddings | 2, 7 |
| 7 | ai-pack-g2e | Structural indexer | 2, 3 |
| 4 | ai-pack-21j | CLI | 5 |
| 5 | ai-pack-iyx | MCP server | 4 |
| 6 | ai-pack-r1g | Pre-flight injection | — |

After Phase 0, 1, and 4+5: spawn `reviewer` before proceeding to the next gate.
