# ADR 007: KG Preflight Performance Measurement Framework

**Status:** Accepted  
**Date:** 2026-03-01  
**Deciders:** AI-Pack Core Team  
**Related:** [ADR-003 Knowledge Graph](003-knowledge-graph.md), [KG Perf Spec](../architecture/kg-perf-measurement/implementation-spec.md)

---

## Context and Problem Statement

The KG preflight context feature (ADR-003) injects relevant knowledge graph observations into
every agent's system prompt before execution. The hypothesis is that this reduces the number
of turns an agent needs to orient itself, reduces redundant exploration tool calls (Read/Grep/Glob),
lowers total token spend, and decreases error rate.

No measurement infrastructure exists to validate or falsify this hypothesis. Without baseline
data and a controlled comparison methodology, we cannot know whether KG preflight is improving
performance, having no effect, or even harming it (by bloating the system prompt).

We need a measurement framework that:
1. Captures per-execution metrics at well-defined instrumentation points
2. Establishes a reproducible baseline (before/without KG preflight)
3. Enables A/B comparison between executions with and without preflight context
4. Produces human-readable reports and machine-queryable data

---

## Decision

We will add a **lightweight, opt-in execution metrics system** that:

1. **Extends `ParsedLog`** with a `KgPreflightBytes` field, set to the byte length of the
   preflight context block injected into the system prompt (0 = KG was absent or returned empty).

2. **Extends `buildLogObservations`** to record `kg_preflight_bytes: N` in the KG entity for
   every execution, making it queryable via KG search.

3. **Adds an `ExecutionMetrics` struct** in `internal/monitoring` that captures
   `Turns`, `TotalTokens`, `ExplorationRatio`, `HasErrors`, `DurationMs`, and `KgPreflightBytes`
   in a single JSON record persisted to `.beads/tasks/<taskID>/metrics.json`.

4. **Adds a `kg perf` subcommand** to the `kg` CLI that reads `metrics.json` files across all
   tasks in a project, bins them by `kg_preflight=true/false`, and prints a comparison report.

5. **Uses `AIPACK_KG_DISABLED=1` env flag** in the existing `task_execution.go` to allow
   baseline runs with KG explicitly suppressed, without code changes to the hot path.

---

## Rationale

### Why extend `ParsedLog` rather than a separate system?

`ParsedLog` is already produced for every execution via `IndexExecutionLog`. Adding
`KgPreflightBytes` here is a one-line change and guarantees the field is set in the same
code path where every other execution metric originates. There is no need for a separate
pipeline.

### Why `metrics.json` per task rather than a central DB?

The project already stores per-task data in `.beads/tasks/<taskID>/`. A per-task `metrics.json`
is consistent with this convention, requires no shared state or locking, is trivially readable
for debugging, and can be aggregated by the `kg perf` command at query time. A central DB would
add schema migration burden for a local developer tool.

### Why `AIPACK_KG_DISABLED=1` rather than a separate KG-off build?

A/B testing requires the same binary, same role, same task description. An environment flag
is the least invasive way to suppress preflight for a baseline run without forking the codebase.
It is already idiomatic: `mcpManager == nil` guards in `PreflightContext` and `IndexExecutionLog`
handle the no-KG path; we only need to pass `nil` there when the flag is set.

### Why `ExplorationRatio` as a metric?

Exploration tool calls (Read, Grep, Glob, Bash) that precede first productive output (Write,
Edit, TaskComplete) indicate an agent spending turns orienting itself. A lower ratio after
KG preflight would validate that preflight context reduces cold-start exploration. This ratio
is calculable from `ToolCounts` which `ParsedLog` already populates.

---

## Consequences

**Positive:**
- Gives the team empirical data to decide whether KG preflight is worth the token overhead.
- `metrics.json` files accumulate silently; no opt-in needed once deployed.
- `kg perf` report is runnable by any developer at any time with zero configuration.
- Small code footprint: ~3 files changed, ~2 files added.

**Negative:**
- `KgPreflightBytes` is a proxy metric; it does not capture the semantic quality of the
  preflight content, only its presence and size.
- A/B comparison requires intentional baseline runs (`AIPACK_KG_DISABLED=1`); the framework
  does not auto-randomize treatment assignment.
- `metrics.json` files are not cleaned up automatically; they accumulate with every execution.

**Neutral:**
- `kg perf` output is printed to stdout; no dashboard or persistent report file is produced
  (engineers can redirect to a file if needed).

---

## Alternatives Considered

**A. OpenTelemetry traces per execution**  
Pros: industry-standard, rich span model.  
Cons: requires OTEL collector, adds runtime dependency, heavy for a local CLI tool. Rejected.

**B. Modify KG entity observations to embed all metrics**  
Pros: metrics queryable via `kg search`.  
Cons: metrics would be lost when KG is reset; data lives in Neo4j/sqlite, not in the task
directory; harder to aggregate cross-task. Rejected.

**C. CSV append to a project-level `perf.csv`**  
Pros: trivially importable into spreadsheets.  
Cons: concurrent writes require locking; CSV schema is fragile. Rejected in favour of
per-task JSON.

---

## Related Decisions

- [ADR-003](003-knowledge-graph.md): Establishes the KG and preflight context feature this ADR measures.
- [ADR-005](005-grade-seeding-redesign.md): Established per-task grade files as the pattern for per-task data; `metrics.json` follows the same convention.
