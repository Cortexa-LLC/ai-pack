# ADR-003: Knowledge Graph Storage and Build Strategy

**Date:** 2026-02-24
**Status:** Accepted
**Deciders:** Bryan Woodruff

---

## Context

Agents lose all context between sessions. We are adding a per-project knowledge graph
to persist findings, code relationships, and architectural knowledge. Key decisions:

1. What database to use as the graph store
2. How to handle the CGO/C++ dependency across platforms
3. Where Kuzu lives in the component topology

---

## Decision 1: Kuzu over SQLite + Cypher Translator

**Chosen: Kuzu**

### Options Considered

**Option A — SQLite (modernc.org/sqlite, pure Go) + hand-built Cypher translator**
- Build a Cypher parser from the openCypher ANTLR4 grammar
- Translate Cypher AST to SQLite recursive CTEs
- FTS5 built-in for keyword search
- Pure Go, no CGO

**Option B — Kuzu (embedded property graph DB)**
- Native openCypher support, no translator
- Columnar storage, optimised for graph traversals
- Requires CGO + C++ static library

### Why Kuzu

Building a Cypher-to-SQL translator is 3–4 weeks of work that adds nothing to the
product — it is infrastructure to support other infrastructure. Kuzu gives us native
Cypher, better multi-hop traversal performance, shortest-path queries, bulk loading
via `COPY FROM`, and the Kuzu Explorer web UI — all without writing a query engine.

The CGO cost (below) is contained and manageable.

**SQLite is not eliminated**: FTS5 keyword search on the knowledge graph may be added
as a lightweight secondary index later if Kuzu's `CONTAINS` predicate proves insufficient
for recall. The two are not mutually exclusive.

---

## Decision 2: Static Linking over Shared Library

**Chosen: Static linking (`libkuzu.a`)**

### Why

A shared library (`.so` / `.dylib`) requires the file to be present at the path expected
at runtime. For a developer tool this creates a fragile deployment: users must install
Kuzu separately or we must bundle the `.so` alongside the binary.

Static linking embeds `libkuzu.a` directly into the `kg` binary. The result is a single
self-contained executable with no runtime library dependency. Binary size increases by
~60–80MB, which is acceptable for a developer CLI.

### Platform Libraries

Pre-built `libkuzu.a` files are downloaded per platform via `make setup-kuzu` /
`scripts/download-kuzu.sh`, sourced from Kuzu's official GitHub releases. Libraries are
stored in `lib/kuzu/<platform>/` and committed via git-lfs or downloaded on demand.

---

## Decision 3: Kuzu Isolated to cmd/kg via MCP Boundary

**Chosen: MCP as the isolation boundary**

### Options Considered

**Option A — Kuzu embedded in cmd/server**
- Server directly queries the graph for pre-flight injection and MCP tools
- CGO dependency spreads to cmd/server

**Option B — Kuzu isolated in cmd/kg, exposed via MCP**
- cmd/kg runs as an MCP child process (`kg server --stdio`)
- cmd/server communicates via MCP protocol — same mechanism as all other MCP tools
- cmd/server remains CGO-free and pure Go

### Why Option B

Option A would require cmd/server to carry CGO + Kuzu, expanding the CGO surface to
the central server binary. Every build of cmd/server would require a C++ toolchain and
the Kuzu static library.

Option B uses the MCP boundary we already have. cmd/server spawns `kg server` as a
child process via `mcpManager` (identical to how other MCP servers are registered today)
and calls `get_preflight_context`, `search_knowledge`, etc. over the protocol. No Kuzu
code in cmd/server, no CGO in cmd/server.

The cost is one MCP round-trip for pre-flight context (~10ms). This is negligible
relative to a task that runs for minutes.

---

## Decision 4: Native CI Runners over Cross-Compilation

**Chosen: Native runners per platform**

Cross-compiling CGO to a different target (e.g., building `linux/arm64` on `darwin/arm64`)
requires the target platform's C++ toolchain and target-specific Kuzu static library.
This is error-prone and poorly supported by standard Go cross-compilation tooling.

Instead, GitHub Actions provides native runners for all four targets:
- `ubuntu-24.04` → linux/amd64
- `ubuntu-24.04-arm` → linux/arm64
- `macos-15` → darwin/arm64
- `macos-13` → darwin/amd64

cmd/agent (pure Go, CGO-free) additionally cross-compiles to `windows/amd64`.

---

## Consequences

- `cmd/kg` requires CGO and a C++ compiler at build time
- `cmd/server` and `cmd/agent` remain pure Go
- `lib/kuzu/` must be populated via `make setup-kuzu` before building `cmd/kg`
- CI matrix uses four native runners; goreleaser manages the release pipeline
- `kg server --stdio` must be on PATH (or configured with full path) for the
  agent-server's mcpManager to spawn it
