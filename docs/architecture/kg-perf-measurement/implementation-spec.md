# Implementation Spec: KG Performance Measurement Framework

**Feature:** Knowledge Graph Preflight Performance Instrumentation  
**ADR:** [ADR-007](../../adr/007-kg-performance-measurement.md)  
**Status:** Ready for Implementation  
**Date:** 2026-03-01

---

## Overview

This spec describes the three engineering tasks needed to instrument KG preflight impact
measurement. Each task is independently mergeable and maps to a filed Beads engineer task.

---

## 1. Metrics Model and Capture (`ExecutionMetrics`)

### Goal

Produce a `metrics.json` file in `.beads/tasks/<taskID>/` at the end of every agent execution,
regardless of whether KG preflight was used.

### Struct Definition

**File:** `internal/monitoring/execution_metrics.go` (new file)

```go
package monitoring

import (
    "encoding/json"
    "os"
    "path/filepath"
    "time"
)

// ExecutionMetrics captures per-task performance data for later analysis.
// Written to .beads/tasks/<taskID>/metrics.json at task completion.
type ExecutionMetrics struct {
    TaskID            string    `json:"task_id"`
    Role              string    `json:"role"`
    StartTime         time.Time `json:"start_time"`
    EndTime           time.Time `json:"end_time"`
    DurationMs        int64     `json:"duration_ms"`

    // Turn and token counts (from ParsedLog)
    Turns             int       `json:"turns"`
    TotalTokens       int       `json:"total_tokens"`

    // Tool usage
    ToolCounts        map[string]int `json:"tool_counts"`
    ExplorationRatio  float64   `json:"exploration_ratio"` // exploration_calls / total_calls

    // Error signal
    HasErrors         bool      `json:"has_errors"`

    // KG treatment: 0 = KG absent or returned empty; >0 = bytes injected
    KgPreflightBytes  int       `json:"kg_preflight_bytes"`

    // Derived
    KgEnabled         bool      `json:"kg_enabled"` // true iff KgPreflightBytes > 0
}

// explorationTools is the set of tool names that indicate orienting/exploration.
var explorationTools = map[string]bool{
    "Read": true, "Grep": true, "Glob": true,
    "Bash": true, "search_knowledge": true, "open_nodes": true,
}

// ComputeExplorationRatio returns exploration_calls / total_calls, or 0 if no calls.
func ComputeExplorationRatio(toolCounts map[string]int) float64 {
    var total, exploration int
    for name, count := range toolCounts {
        total += count
        if explorationTools[name] {
            exploration += count
        }
    }
    if total == 0 {
        return 0
    }
    return float64(exploration) / float64(total)
}

// WriteMetrics serialises m to <taskDir>/metrics.json, creating the directory if needed.
func WriteMetrics(taskDir string, m *ExecutionMetrics) error {
    if err := os.MkdirAll(taskDir, 0o755); err != nil {
        return err
    }
    data, err := json.MarshalIndent(m, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(filepath.Join(taskDir, "metrics.json"), data, 0o644)
}
```

### Instrumentation Points

**File:** `internal/kgclient/log_indexer.go`

1. Add `KgPreflightBytes int` to `ParsedLog` struct.
2. In `buildLogObservations`, append `"kg_preflight_bytes: N"` observation.

**File:** `internal/server/task_execution.go`

1. After `PreflightContext(...)` call, measure `len(kgBlock)` → store in execution context.
2. At end of `executeAgenticLoop` (before or after `IndexExecutionLog` call), call:

```go
metrics := &monitoring.ExecutionMetrics{
    TaskID:           taskID,
    Role:             role,
    StartTime:        startTime,
    EndTime:          time.Now(),
    DurationMs:       time.Since(startTime).Milliseconds(),
    Turns:            parsed.Turns,
    TotalTokens:      parsed.TotalTokens,
    ToolCounts:       parsed.ToolCounts,
    ExplorationRatio: monitoring.ComputeExplorationRatio(parsed.ToolCounts),
    HasErrors:        parsed.HasErrors,
    KgPreflightBytes: kgPreflightBytes, // captured earlier
    KgEnabled:        kgPreflightBytes > 0,
}
taskDir := filepath.Join(projectRoot, constants.BeadsDir, "tasks", taskID)
_ = monitoring.WriteMetrics(taskDir, metrics)
```

> The `_ =` assignment follows project style for best-effort metrics writes (same pattern
> as `IndexExecutionLog`).

### `AIPACK_KG_DISABLED` Flag

**File:** `internal/server/task_execution.go`

```go
// Near the top of executeAgenticLoop, before PreflightContext call:
mcpForPreflight := s.mcpManager
if os.Getenv("AIPACK_KG_DISABLED") == "1" {
    mcpForPreflight = nil // suppress KG preflight for baseline runs
}
kgBlock := kgclient.PreflightContext(ctx, mcpForPreflight, execution.Task, execution.ProjectRoot)
kgPreflightBytes := len(kgBlock)
```

---

## 2. Exploration Ratio Definition

Exploration tools are calls that orient the agent but produce no output artifact:

| Tool | Category |
|---|---|
| `Read` | exploration |
| `Grep` | exploration |
| `Glob` | exploration |
| `Bash` | exploration |
| `search_knowledge` | exploration |
| `open_nodes` | exploration |
| `query_graph` | exploration |
| `get_file_context` | exploration |
| `Write` | productive |
| `Edit` | productive |
| `MultiEdit` | productive |
| `TaskComplete` | productive |
| `add_entity` / `create_entities` | productive |
| `add_observation` / `add_observations` | productive |
| `link_entities` / `create_relations` | productive |
| `sequentialthinking` | neutral (omitted from ratio) |

**Formula:**  
`exploration_ratio = Σ(exploration_tool_calls) / Σ(exploration + productive tool_calls)`

Neutral tools (thinking tools) are excluded from both numerator and denominator to avoid
penalising agents that reason explicitly.

**Expected KG effect:** A well-seeded KG should reduce `exploration_ratio` because the agent
already has key file paths, function signatures, and design decisions in its context window
from the preflight block, reducing the need for orienting Read/Grep/Glob calls.

---

## 3. `kg perf` CLI Command

### Goal

Aggregate `metrics.json` files across all tasks, split by `kg_enabled`, and print a comparison.

### Command

```
kg perf [--project <root>] [--role <role>] [--min-samples <n>] [--json]
```

**File:** `cmd/kg/perf.go` (new file)

### Output Format (human-readable, default)

```
KG Preflight Performance Report
================================
Project: /Users/bryanw/Projects/Vibe/ai-pack
Period:  all time (127 executions)

                    KG OFF (n=52)    KG ON (n=75)    Delta
─────────────────────────────────────────────────────────────
Avg Turns           18.3             12.1            -33.9%  ✓
Avg Tokens          42,100           38,400          -8.8%   ✓
Avg Exploration%    61%              44%             -17pp   ✓
Error Rate          19%              11%             -8pp    ✓
Avg Duration (s)    142              138             -2.8%   ~

✓ = improvement  ~ = negligible  ✗ = regression

Roles sampled: architect(31), engineer(58), pm(22), reviewer(16)

Caveat: tasks are NOT randomly assigned; KG=OFF runs use AIPACK_KG_DISABLED=1.
```

### Output Format (JSON, `--json` flag)

```json
{
  "project": "/path/to/project",
  "generated_at": "2026-03-01T12:00:00Z",
  "kg_off": {
    "n": 52,
    "avg_turns": 18.3,
    "avg_tokens": 42100,
    "avg_exploration_ratio": 0.61,
    "error_rate": 0.19,
    "avg_duration_ms": 142000
  },
  "kg_on": {
    "n": 75,
    "avg_turns": 12.1,
    "avg_tokens": 38400,
    "avg_exploration_ratio": 0.44,
    "error_rate": 0.11,
    "avg_duration_ms": 138000
  },
  "deltas": {
    "turns_pct": -33.9,
    "tokens_pct": -8.8,
    "exploration_ratio_pp": -17.0,
    "error_rate_pp": -8.0,
    "duration_pct": -2.8
  }
}
```

### Implementation Notes

- Walk `.beads/tasks/*/metrics.json` relative to `--project` (defaults to `git rev-parse --show-toplevel`).
- Filter by `--role` if provided.
- Skip files that cannot be parsed (log warning, continue).
- Require `--min-samples` (default 5) on each side before printing; warn otherwise.
- Deltas use signed percentages for means, signed percentage-points for rates.

---

## 4. Baseline Methodology

### How to Establish a Baseline

1. **Select a task corpus**: Choose 20–50 representative tasks from project history that
   already have execution logs. These become your "before" dataset.

2. **Re-run with KG disabled**: For each baseline task, create a fresh task with the same
   role and description, then execute with `AIPACK_KG_DISABLED=1`. This populates
   `metrics.json` with `kg_enabled: false` records.

3. **Normal runs accumulate KG-on data**: Everyday execution with KG enabled automatically
   produces `kg_enabled: true` records. Over time, the corpus grows without manual effort.

4. **Minimum corpus for significance**: 30 executions per treatment arm per role is the
   recommended minimum for a stable mean. With `--min-samples 30`, the CLI will warn when
   below this threshold.

5. **Control for confounders**: Filter by role (`--role engineer`) before comparing, since
   different roles have inherently different turn counts and tool usage patterns.

### Confounders and Caveats

| Confounder | Mitigation |
|---|---|
| Task difficulty varies | Filter by role; normalize by `TotalTokens` as proxy for task size |
| KG quality degrades over time (stale data) | Record `kg_preflight_bytes` to detect empty/tiny preflight blocks |
| Model version changes | Record in `metrics.json` as future extension |
| Same task not re-run identically | Acknowledge in report; compare distributions not individual pairs |

---

## 5. Files Changed / Created

| File | Change |
|---|---|
| `internal/monitoring/execution_metrics.go` | **New** — `ExecutionMetrics` struct, `WriteMetrics`, `ComputeExplorationRatio` |
| `internal/kgclient/log_indexer.go` | **Modify** — add `KgPreflightBytes` to `ParsedLog`; record in observations |
| `internal/server/task_execution.go` | **Modify** — capture `kgPreflightBytes`, write metrics, support `AIPACK_KG_DISABLED` |
| `cmd/kg/perf.go` | **New** — `kg perf` subcommand |
| `cmd/kg/root.go` | **Modify** — register `perf` command |
| `internal/monitoring/execution_metrics_test.go` | **New** — unit tests for ratio computation and JSON serialisation |
| `internal/server/task_execution_metrics_test.go` | **New** — integration test for `AIPACK_KG_DISABLED` flag |

---

## 6. Acceptance Criteria

- [ ] `metrics.json` is written for every task completion (success and failure).
- [ ] `kg_preflight_bytes: 0` when KG is absent or `AIPACK_KG_DISABLED=1`.
- [ ] `kg_preflight_bytes: N > 0` when preflight context was injected.
- [ ] `kg perf` prints report with correct delta signs.
- [ ] `kg perf --json` produces parseable JSON.
- [ ] `ExplorationRatio` matches manual count from a known execution log.
- [ ] `AIPACK_KG_DISABLED=1` suppresses preflight without changing any other behaviour.
- [ ] All new code covered by unit tests; existing tests pass.

---

## Related Documents

- ADR: [docs/adr/007-kg-performance-measurement.md](../../adr/007-kg-performance-measurement.md)
- ADR-003 (KG foundation): [docs/adr/003-knowledge-graph.md](../../adr/003-knowledge-graph.md)
- ADR-005 (grade file pattern): [docs/adr/005-grade-seeding-redesign.md](../../adr/005-grade-seeding-redesign.md)
