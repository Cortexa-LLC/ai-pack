# ADR-009: KG Health — Timestamp Fix, Server-Era Curation, Health Report

**Date:** 2026-08-23
**Status:** Accepted
**Deciders:** Bryan Woodruff

---

## Context

Every ai-pack role reads the knowledge graph first (`kg server --stdio`, standalone Go
binary from the `mcp` submodule, source `mcp/src/kg/` + shared library `mcp/src/kglib/`).
PRD "Framework Strengthening" Epic 1 (P0) found two defects that make the KG a liability:

1. **US-103:** `search_knowledge` returns `created_at: 0001-01-01T00:00:00Z` for every
   observation and entity, even ones written seconds earlier — staleness is unassessable.
2. **US-101:** the graph contains server-era guidance (retired `agent` CLI, agent-mcp
   tools, port 8082, Beads/`bd`, performance-grade seeding, launchd) that KG-first agents
   faithfully retrieve as if current.

US-102 asks for a health report so graph condition is visible without hand-written Cypher.

This repo's DB at investigation time: `.ai/knowledge.db` (KuzuDB), 8,301 entities,
~9k relations, 1,780 observations, single `project_id` `ai-pack`.

---

## Decision 1 (US-103): Fix the timestamp bug in the kglib row scan — retrieval-side only

### Root cause (verified in source and empirically)

**Storage is correct. Every read path silently discards the timestamp.**

- **Write path is sound.** `CreateObservation` (`mcp/src/kglib/observation.go:21–47`)
  binds `CreatedAt: time.Now().UTC()` as a Cypher parameter; go-kuzu encodes `time.Time`
  params as native Kuzu timestamps (`go-kuzu@v0.11.3/value_helper.go:611–616`). Entity
  writes (`kglib/entity.go:11–43`) are equivalent. Schema columns are `TIMESTAMP`
  (`kglib/schema.go:49–50,59`).
- **Why the write appears to work:** `add_observation` returns the in-memory struct
  (`observation.go:21–26,53`) — it echoes `time.Now()`, never re-reading the DB. That is
  why writes show real timestamps while every read shows zero.
- **The bug:** every read scans the timestamp column with an `int64` type assertion,
  e.g. `if ts, ok := row[3].(int64); ok { obs.CreatedAt = time.UnixMicro(ts).UTC() }`
  under the comment "Kuzu returns timestamps as int64 microseconds". That was never true
  for go-kuzu v0.11.3: TIMESTAMP values decode to `time.Time`
  (`go-kuzu@v0.11.3/value_helper.go:367–374`, `return time.Unix(...)`). The assertion
  fails, `ok` is false, `CreatedAt` stays the zero value, and the MCP layer serializes
  `0001-01-01T00:00:00Z`.
- **All affected sites (in the mcp repo):**
  - `src/kglib/search.go:109,112` (KeywordSearch entity scan), `:184` (GetTopObservations),
    `:252` (batchGetObservations — this feeds `search_knowledge` results)
  - `src/kglib/observation.go:93` (GetObservations)
  - `src/kglib/entity.go:81,84,126,129,178,181` (GetEntity / GetEntityByName / ListEntities)
  - `src/kglib/relation.go:173,176`
  - `src/kglib/hnsw_index.go:108,111`
- **Empirical confirmation (raw read-only Kuzu query against a copy of this repo's DB):**
  0 of 1,780 observations and 0 of 8,301 entities have a zero `created_at`; stored values
  range 2026-03-01 → 2026-08-23, including the exact observation whose MCP write returned
  `2026-08-23T21:36:04Z`. The data was never lost.

### Fix (lands in the kg repo — mcp/src/kglib, not here)

1. Add one helper in kglib:
   ```go
   // scanTime converts a Kuzu-returned timestamp cell to time.Time.
   // go-kuzu returns time.Time for TIMESTAMP columns; int64 (micros) kept defensively.
   func scanTime(v any) time.Time {
       switch t := v.(type) {
       case time.Time:
           return t.UTC()
       case int64:
           return time.UnixMicro(t).UTC()
       }
       return time.Time{}
   }
   ```
2. Replace all 15 `.(int64)` timestamp assertions listed above with
   `obs.CreatedAt = scanTime(row[3])` (etc.). Do not touch `store.go` count assertions
   (`count(*)` genuinely is int64).
3. Regression test: create entity + observation, retrieve via `KeywordSearch` and
   `GetObservations`, assert `CreatedAt` within the last minute and non-zero. (Existing
   tests only exercised the write-path echo, which is why this shipped.)

### Migration / legacy rows

**None needed.** Because storage was always correct, fixing the scan retroactively
restores real timestamps for all existing rows. No backfill, no schema change, no data
migration. The health report (Decision 3) still counts zero-timestamp rows defensively —
a DB written by some other/older writer could contain real zeros — and must label them
"legacy, age unknown", but for this repo's DB the expected count is 0 after the fix.

### Alternatives rejected

- **Schema change / rewrite timestamps:** unnecessary — data is intact; a rewrite risks
  destroying the very provenance we want.
- **Fix at the MCP serialization layer only:** would leave CLI (`kg show`, `kg search`)
  and recency logic (`calculateRecencyScore`, ORDER BY consumers in Go) still broken.

---

## Decision 2 (US-101): Curate server-era knowledge — delete operational noise, mark historical records, plus an agent-side disregard rule

### Baseline measurement (the US-101 acceptance benchmark)

Method: replicated the exact agent-facing search (`search_knowledge` → `KeywordSearch`,
`src/kglib/search.go:39–145`: tokenized OR `CONTAINS` over entity name + observation
content, `ORDER BY e.updated_at DESC LIMIT 10`, top-3 observations per entity by
`created_at DESC`) against a copy of `.ai/knowledge.db`, scanning surfaced content
(entity name + top-3 observations) for retired-component patterns
(`agent-mcp|mcp__agent|agent create|agent CLI|port 8082|8082|beads|bd init|bd create|performance.grade|seed-grades|launchd|agent[- ]server`, case-insensitive).

| Query | Top-10 results | Flagged |
|---|---|---|
| task tracking | 10 | 6 |
| agent spawn | 10 | 6 |
| orchestrate | 6 | 5 |
| agent create | 10 | 6 |
| agent-mcp | 5 | 5 |
| beads | 10 | 9 |
| performance grade | 10 | 2 |
| **Total** | **61** | **39 (64%)** |

Hand classification of the flagged set:

- **Guidance-shaped server-era records (the actual problem):** task-completion topics
  such as "Added agent-mcp MCP server … documented shared task DB usage", "Created MCP
  stdio server exposing agent CLI as MCP tools", dead-code entities like
  `ensureBeadsForProject` ("Runs 'bd init --server-host …'"), and `exec:*` /
  `task_outcome` execution records referencing Beads task flows. An agent reading these
  has no signal they describe a retired system.
- **Legitimate history (must survive, identifiable as history):** the agent-server
  deprecation inventory, the PRD investigation finding.
- **Current-era records that merely mention retired components** (e.g. "Remove stale
  Beads references" completion): acceptable; mentions of removal are not directives.

The acceptance metric is therefore **directive hits**, not raw mentions: a result counts
as a failure if it surfaces retired-component instructions *without* an `[OBSOLETE]`
marker. Baseline directive count is the dominant share of the 39; target is **0**.

### Why not supersede-only (option b) — checked against the ranking code

Search recency behavior, from source: entities rank by `e.updated_at DESC`
(`search.go:77`); observations within an entity rank by `created_at DESC`, top 3 shown
(`search.go:124,213–218`). But `CreateObservation` never bumps the parent entity's
`updated_at`, and MCP `add_entity` always `CREATE`s a new entity
(`kg/internal/knowledge/mcp_server.go:206` → `CreateEntity`) rather than upserting. So
appending a correction (i) does not demote the polluted entity in result order, and
(ii) at best occupies one of its three observation slots while the other two still show
obsolete guidance. **Supersede-only is ineffective with the current ranking; rejected as
the sole mechanism.**

### Chosen approach: (a) + (c) combination

Mutation surface reality (from source): kglib has `DeleteObservation`
(`observation.go:104`), `DeleteEntity` (`entity.go:192`), `DeleteRelation` — but **no
update function**, and none of these are exposed via MCP tools or the CLI (`kg gc` is an
unimplemented stub, `kg/gc.go`). Curation therefore runs as a one-off script with direct
DB access; the repeatable *check* is a separate read-only script.

1. **Archive first:** `kg export` full JSONL, stored under `.ai/archives/` (plus
   `cp -r .ai/knowledge.db .ai/knowledge.db.bak-<date>`). Nothing is destroyed
   unrecoverably.
2. **Delete** (option a) pure server-era operational noise — no guidance or historical
   value beyond the archive: `task_outcome` entities and `exec:*` topics from the server
   era, server-era task-completion topics whose content directs use of retired
   components, and dead-code entities (`function`/`file` entities whose source no longer
   exists in the tree, e.g. `ensureBeadsForProject`). Use `DETACH DELETE` semantics
   (delete observations + edges + entity).
3. **Mark** (option c) genuinely historical records — the deprecation inventory,
   architecture investigations that explain *why* the pivot happened — by prefixing each
   observation's content:
   `[OBSOLETE — server era, retired 2026-08-22, see tag v2.0-server-final] `.
   Applied via raw Cypher `SET o.content = ...` (there is no update API; delete+re-add
   would falsify `created_at`). Idempotent: skip rows already carrying the prefix.
4. **Keep unmarked:** the deprecation *decision* record itself and all current-era
   knowledge.

### Concrete procedure (engineer-executable against this repo's DB)

- `scripts/curate-kg-server-era.py` — Python + `kuzu==0.11.3` (pinned; matches the kg
  binary's embedded Kuzu — read/write compatibility verified during this investigation).
  Contains the explicit delete/mark rule lists (entity-name patterns + content
  patterns). Defaults to `--dry-run` printing every planned mutation; `--apply` executes.
  **Must run with no `kg server` process open on this project** (KuzuDB single-writer);
  backup step is built in and non-optional.
- `scripts/kg-benchmark.py` — read-only verification: runs the 7-query benchmark above
  (same Cypher as `KeywordSearch`), fails non-zero if any top-10 surfaced content matches
  a directive pattern without the `[OBSOLETE]` marker. This is the repeatable US-101
  check; re-run after curation (target 0) and periodically thereafter.
- **Agent-side contract change (flagged, not made here):** one instruction line —
  "knowledge prefixed `[OBSOLETE — …]` is history, never guidance; do not act on it" —
  added to the KG-first section of **all 7 agent files** in `plugin/agents/` (and the
  orchestrate skill if it cites KG results). This changes agent behavior contracts and
  is escalated to the orchestrator as its own task.

### Hazard noted during investigation

`kg index` **wipes all project data first** — including agent-written topics and
observations (`kg/internal/knowledge/indexer.go:141–178`, `clearProjectData`, called
unconditionally at `indexer.go:185`). Re-indexing must never be used as a "cleanup"
mechanism, and the curation runbook must say so. (Separate defect worth its own issue in
the kg repo: index should only clear code-derived entity types.)

### Alternatives rejected

- **Delete everything flagged:** violates the PRD constraint that true history stays
  retrievable as identifiable history (and the PRD risk "curation deletes knowledge that
  was still useful").
- **Supersede-only (b):** ineffective — see ranking analysis above.
- **Mark-only (c):** markers on `task_outcome`/`exec:*` noise would still consume top-10
  slots and observation slots, degrading recall for current knowledge; noise has no
  historical value once archived.

---

## Decision 3 (US-102): `kg health` command in the kg repo

### Where it lives

Extend the `kg` CLI with a new `kg health [--json]` subcommand (mcp repo,
`src/kg/health.go`), reusing the scope/read-only plumbing already in `kg/stats.go`
(`OpenStoreReadOnly`, scope resolution). Rejected: a Python script in `scripts/` — a
*recurring* tool must not depend on a version-pinned Python Kuzu client tracking the
binary's storage format (acceptable for the one-off curation script, fragile as an
ongoing interface). Rejected: extending `kg stats` in place — stats is a quick count
tool; health carries state (growth) and policy (legacy labeling).

### Report contents (single command, current project/scope)

- **Counts:** entities (total + by type), relations, observations.
- **Growth since last run:** delta of each count vs the previous snapshot, persisted at
  `.ai/kg-health.json` (written on every run; first run reports "no previous snapshot").
- **Staleness:** share of observations with zero `created_at`, labeled
  **"legacy, age unknown"** (see Decision 1 — expected 0 in this repo after the scan
  fix, kept for foreign/older DBs); plus newest/oldest/median observation age.
- **Orphaned entities:** count of entities with no `HAS_OBSERVATION` edge and no typed
  relation in either direction (single Cypher with `NOT EXISTS` on both patterns).
- **Curation status:** count of observations carrying the `[OBSOLETE` prefix
  (`STARTS WITH`), per Decision 2.
- **Output:** human summary to stdout by default; `--json` emits one machine-readable
  object (all metrics + snapshot timestamps) for CI or trending.

Exit code 0 always (report, not gate); the US-101 gate is `scripts/kg-benchmark.py`.

---

## Consequences

- Spec 1 and Spec 3 are changes to the **kg repo** (github.com/Cortexa-LLC/mcp:
  `src/kglib` scan fix + tests, `src/kg/health.go`); ai-pack then bumps the `mcp`
  submodule pin and rebuilds/installs `kg`. Both specs above are self-contained for an
  engineer working in that repo.
- Spec 2 is executed **in this repo** (`scripts/curate-kg-server-era.py`,
  `scripts/kg-benchmark.py`) against `.ai/knowledge.db`, after the Spec 1 fix is
  installed (so verification output shows real timestamps).
- Ordering: Spec 1 → Spec 2 → Spec 3 is ideal but only Spec 2's *verification* truly
  benefits from Spec 1; the three can proceed in parallel by different engineers.
- The agent-file `[OBSOLETE]`-disregard instruction touches all 7 role definitions —
  orchestrator decision, out of this ADR's scope.
- Until Spec 1 ships, recency-based ranking refinements are pointless (every timestamp
  reads as zero in Go); do not build ranking work on top of the broken scan.
