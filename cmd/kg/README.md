# kg — Knowledge Graph CLI

The `kg` binary provides a CLI and MCP server for the ai-pack knowledge graph, backed
by [KuzuDB](https://kuzudb.com). It runs as a subprocess managed by the ai-pack MCP
manager, and can also be used directly from the command line.

**Database location:** `.ai/knowledge.db` relative to the detected project root.
The project root is determined by walking up from the current directory to find a
`.ai/` directory, a git root, or a common project marker (`go.mod`, `package.json`, etc.).

## Build

```bash
make build-kg          # build to bin/kg
make install-kg        # build + install to /usr/local/bin/kg
```

Or from the source directory:
```bash
CGO_ENABLED=1 go build -o kg ./cmd/kg
```

## First-Time Setup: `kg index`

Run once per project to scan the codebase and populate the graph:

```bash
kg index
```

```
🔍 Indexing codebase at /path/to/your-project...
✅ Indexing complete!
   Files scanned:     191
   Entities created:  1113
   Relations created: 1517
   Duration:          5.2s
```

Re-run after large structural changes or significant refactors. The MCP tool
`kg__index_project` triggers the same operation from within a Claude session.

**Skipped automatically:** `.git`, `node_modules`, `vendor`, `dist`, `build`, and any
path matching `.gitignore` or `.claudeignore` patterns.

## Commands

```
kg index                              # scan codebase → .ai/knowledge.db
kg search <query>                     # keyword search across entities + observations
kg stats                              # count of entities, relations, observations
kg show <entity-id>                   # show entity with relations + observations
kg add entity --name <n> --type <t>   # add an entity manually
kg add observation <id> <content>     # attach an observation to an entity
kg link <from> --rel <TYPE> <to>      # create a directed relation
kg export                             # export to GraphML/JSON
kg graph                              # write GraphML to stdout
kg gc                                 # remove orphaned nodes + observations
kg server --stdio                     # start MCP server (used by ai-pack, not manually)
kg version                            # print version, commit, build time
```

## MCP Server Mode

The MCP manager in `internal/mcp/` spawns `kg server --stdio` as a subprocess per
project. Tool calls arrive over stdin as JSON-RPC, results are written to stdout.
The server uses an open-use-close pattern — each tool call opens the database,
executes, and closes it, so concurrent `kg index` or CLI commands never block.

## Knowledge Graph Schema

### Entity Types

| Type | Description |
|------|-------------|
| `file` | Source file (e.g. `internal/auth/token.go`) |
| `package` | Go package or language module |
| `module` | Go module (e.g. `github.com/cortexa-llc/ai-pack`) |
| `function` | Function or method |
| `type` | Type declaration (struct, interface, enum, class) |
| `topic` | Conceptual entity — architecture decision, investigation topic |
| `import` | Import path |
| `concept` | Free-form concept (manually added) |

### Relation Types

| Relation | Direction | Meaning |
|----------|-----------|---------|
| `CONTAINS` | file → function/type | File declares this entity |
| `IMPORTS` | file/package → import | File imports this dependency |
| `CALLS` | function → function | Function calls another |
| `IMPLEMENTS` | type → type | Type implements interface |
| `BELONGS_TO` | entity → package | Entity belongs to package |
| `DEPENDS_ON` | entity → entity | Architectural dependency |
| `RELATES_TO` | entity → entity | Free-form association |

### Observation Prefixes (convention)

| Prefix | Use for |
|--------|---------|
| `[INVESTIGATION]` | Findings from debugging or exploration |
| `[DECISION]` | Architectural choices and rationale |
| `[CAVEAT]` | Known limitations, edge cases, gotchas |
| `[PERFORMANCE]` | Measured characteristics or bottlenecks |

## Indexed Languages

Go, Python, TypeScript, JavaScript, Rust, Java, Kotlin, C, C++, C#, Swift, Ruby,
Bash, Groovy, CSS, HTML, YAML, Markdown, GraphQL, JSON Schema, PDF, Assembly, Makefile.

## Troubleshooting

**"database already open"** — another `kg` process has the database locked.
Close other instances and retry. The MCP server uses open-use-close per call to
avoid holding the lock between tool calls.

**Empty results after indexing** — check `.gitignore` and `.claudeignore` patterns.
Verify the files you expect are not excluded.

**Import relations missing** — ensure Go files have valid syntax. The tree-sitter
parser skips files with parse errors.
