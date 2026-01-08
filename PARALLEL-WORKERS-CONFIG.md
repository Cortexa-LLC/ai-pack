# Parallel Workers Configuration

**Version:** 1.0.0
**Last Updated:** 2026-01-07
**Status:** DEFAULT CONFIGURATION

---

## Overview

The ai-pack framework is **configured by default** to use **multiple parallel workers** for resolving work packages. This configuration optimizes delivery speed and resource utilization while maintaining quality and coordination.

---

## Default Configuration

### Automatic Parallelization

**ENABLED BY DEFAULT** for work packages meeting these criteria:

```
✅ AUTO-PARALLEL when:
- 3 or more independent subtasks identified
- Subtasks modify different files/modules
- No cross-subtask dependencies
- Each subtask has isolated acceptance criteria
```

**Maximum Parallel Workers:** 4 concurrent agents

**Launch Pattern:** Single message block (true parallelism)

---

## Configuration Rules

### When Parallel Workers Are Used (DEFAULT)

```
Work Package Type                    | Strategy          | Workers
-------------------------------------|-------------------|----------
3+ independent API endpoints         | Parallel (DEFAULT)| 2-4
3+ independent modules/components    | Parallel (DEFAULT)| 2-4
3+ similar features (no dependencies)| Parallel (DEFAULT)| 2-4
Multiple bug fixes (different areas) | Parallel (DEFAULT)| 2-4
Independent refactoring tasks        | Parallel (DEFAULT)| 2-4
```

### When Sequential Execution Is Used

```
Work Package Type                    | Strategy          | Workers
-------------------------------------|-------------------|----------
1-2 subtasks                         | Sequential        | 1
Tightly coupled changes              | Sequential        | 1
Sequential database migrations       | Sequential        | 1
Same file modified multiple times    | Sequential        | 1
Subtasks with execution dependencies | Sequential        | 1
```

### Hybrid Approach

```
When work package has BOTH independent and dependent subtasks:
1. Sequence the dependent chain
2. Parallelize independent groups
3. Coordinate integration points

Example: Database migration + 3 new API endpoints
→ Sequence: DB migration first
→ Then parallel: 3 workers for 3 API endpoints
```

---

## Orchestrator Behavior

### Default Delegation Pattern

The Orchestrator role automatically applies parallel execution for eligible work packages:

```pseudocode
FUNCTION delegate_work_package(work_package):
  subtasks = analyze_and_decompose(work_package)

  IF subtasks.count >= 3 AND are_independent(subtasks) THEN
    // DEFAULT: Spawn parallel workers
    workers = []
    FOR each subtask in subtasks.take(4):
      workers.append(spawn_worker(subtask))
    END FOR

    // Launch all workers in single message block
    launch_parallel(workers)
    monitor_parallel_progress(workers)

  ELSE IF has_dependencies(subtasks) THEN
    // Hybrid: sequence + parallel
    dependent_chain = extract_dependencies(subtasks)
    independent_group = extract_independent(subtasks)

    execute_sequential(dependent_chain)
    launch_parallel(independent_group)

  ELSE
    // Sequential for small/coupled tasks
    execute_sequential(subtasks)
  END IF
END FUNCTION
```

---

## Worker Coordination

### Parallel Worker Responsibilities

When multiple workers execute in parallel:

```
Each Worker Must:
✓ Own isolated, non-overlapping deliverable
✓ Avoid modifying files owned by other workers
✓ Work independently without blocking others
✓ Complete with own tests and verification
✓ Flag integration concerns early
✓ Communicate dependency needs proactively
```

### Orchestrator Coordination Responsibilities

```
Orchestrator Must:
✓ Spawn workers in single message block (true parallel)
✓ Track progress independently per worker
✓ Monitor for conflicts or integration issues
✓ Coordinate handoff points
✓ Verify cross-cutting concerns
✓ Ensure overall system coherence
✓ Merge results and verify integration
```

---

## Integration Points

### Task Packet Structure for Parallel Work

Each parallel worker gets isolated task packet:

```
.ai/tasks/
├── 2026-01-07_work-package/
│   ├── 00-contract.md              # Overall work package contract
│   ├── 10-plan.md                  # Master plan with parallel strategy
│   ├── worker-1/                   # Worker 1 isolated context
│   │   ├── 00-contract.md
│   │   ├── 20-work-log.md
│   │   └── 40-acceptance.md
│   ├── worker-2/                   # Worker 2 isolated context
│   │   ├── 00-contract.md
│   │   ├── 20-work-log.md
│   │   └── 40-acceptance.md
│   ├── worker-3/                   # Worker 3 isolated context
│   │   ├── 00-contract.md
│   │   ├── 20-work-log.md
│   │   └── 40-acceptance.md
│   ├── 30-review.md                # Integrated review
│   └── 40-acceptance.md            # Overall acceptance
```

---

## Examples

### Example 1: Add 3 New API Endpoints (DEFAULT PARALLEL)

**Work Package:** Implement user profile, notifications, and settings APIs

**Default Configuration Applied:**
```
✓ 3 independent subtasks identified
✓ Each modifies different files
✓ No dependencies between endpoints
✓ Each has isolated acceptance criteria

→ AUTOMATIC: Spawn 3 parallel workers (up to 4 if more endpoints)

Worker 1: Implement user profile API + tests
Worker 2: Implement notifications API + tests
Worker 3: Implement settings API + tests

Launch: Single message with 3 Task() calls (or 4 if 4 endpoints)
Timeline: All complete in parallel
Integration: Orchestrator verifies API consistency
```

### Example 2: Database Migration + UI Changes (HYBRID)

**Work Package:** Update user schema and add 4 new UI components

**Configuration Applied:**
```
✓ Database migration must complete first (dependency)
✓ 4 UI components are independent after DB change

→ HYBRID: Sequence DB, then parallel UI workers

Phase 1: Worker 1 - Database migration + tests
Phase 2 (parallel):
  Worker 2: UI component 1 + tests
  Worker 3: UI component 2 + tests
  Worker 4: UI component 3 + tests
  Worker 5: UI component 4 + tests

Launch: Sequential for phase 1, parallel for phase 2 (4 workers)
```

### Example 3: Tightly Coupled Feature (SEQUENTIAL)

**Work Package:** Refactor authentication flow across 5 files

**Configuration Applied:**
```
✗ Changes tightly coupled
✗ Same files modified multiple times
✗ Strong execution dependencies

→ SEQUENTIAL: Single worker

Worker 1: Complete entire refactoring sequentially
  Step 1: Update auth module
  Step 2: Update middleware
  Step 3: Update tests
  Step 4: Update integration points
  Step 5: End-to-end verification
```

---

## Performance Benefits

### Expected Improvements with Parallel Workers

```
Metric                        | Sequential | Parallel (4 workers) | Improvement
------------------------------|------------|----------------------|-------------
Work package delivery time    | T          | ~T/4                 | 4x faster
Resource utilization          | 25%        | 100%                 | 4x better
Verification isolation        | Cumulative | Per-worker           | Clearer
Integration issues detected   | Late       | Early                | Faster fix
Developer productivity        | Serial     | Parallel             | Higher
```

### Actual Timeline Example

```
Task: Implement 4 new API endpoints (each ~2 hours)

Sequential Approach:
├─ Worker starts endpoint 1 → 2 hours
├─ Worker starts endpoint 2 → 2 hours
├─ Worker starts endpoint 3 → 2 hours
└─ Worker starts endpoint 4 → 2 hours
Total: 8 hours

Parallel Approach (DEFAULT):
├─ Worker 1: endpoint 1 ──────► 2 hours
├─ Worker 2: endpoint 2 ──────► 2 hours
├─ Worker 3: endpoint 3 ──────► 2 hours
└─ Worker 4: endpoint 4 ──────► 2 hours
Total: 2 hours + integration (20 min) = 2.33 hours

Improvement: 3.4x faster delivery
```

---

## Conflict Resolution

### Handling Parallel Worker Conflicts

**File Conflicts:**
```
IF two workers modify same file THEN
  Orchestrator detects conflict
  Coordinate merge strategy
  One worker waits or adjusts scope
  Test integrated result
END IF
```

**Resource Conflicts:**
```
IF workers need same resource THEN
  Orchestrator prioritizes access
  Sequence resource usage
  Parallelize non-conflicting work
END IF
```

**Dependency Discoveries:**
```
IF unexpected dependency discovered THEN
  Worker flags issue immediately
  Orchestrator re-sequences work
  Adjust parallel strategy dynamically
END IF
```

---

## Quality Assurance with Parallel Workers

### Testing Strategy

```
Per-Worker Testing:
- Each worker runs own unit tests
- Each worker verifies own acceptance criteria
- Each worker achieves 80-90% coverage on own code

Integration Testing (Orchestrator):
- Orchestrator runs integration tests
- Verify cross-worker interactions
- End-to-end system verification
- Performance testing across all changes
```

### Review Process

```
Option 1: Independent Reviews
- Each worker's output reviewed separately
- Faster parallel review
- Focus on isolated changes

Option 2: Integrated Review
- Review all changes together
- Verify overall coherence
- Catch integration issues

Recommended: Hybrid
- Independent reviews first
- Integrated review for cross-cutting concerns
```

---

## Configuration Override

### When to Override Default Parallel Configuration

You may choose sequential execution even with 3+ subtasks when:

```
⚠️ Override to Sequential if:
- Team unfamiliar with parallel workflows
- High coordination overhead expected
- Subtasks appear independent but have hidden dependencies
- Experimental/exploratory work with unclear boundaries
- Insufficient Orchestrator capacity for monitoring
```

### How to Override

In the work package contract or plan:

```markdown
## Execution Strategy

**Override Default:** Sequential execution chosen

**Rationale:**
[Explain why parallel execution not appropriate]

**Approach:**
Single worker will complete subtasks sequentially:
1. Subtask 1
2. Subtask 2
3. Subtask 3
```

---

## References

- **[roles/orchestrator.md](roles/orchestrator.md)** - Section 2.5: Parallel Worker Configuration
- **[gates/20-tool-policy.md](gates/20-tool-policy.md)** - Section 7: Agent Spawning Policy
- **[workflows/standard.md](workflows/standard.md)** - Phase 2: Parallel Execution Strategy
- **[workflows/standard.md](workflows/standard.md)** - Role Assignments with Parallel Coordination

---

## Version History

```
v1.0.0 - 2026-01-07
- Initial configuration
- Set parallel workers as default for 3+ independent subtasks
- Maximum 3 concurrent workers
- Documented decision criteria and examples
```

---

**Configuration Status:** ACTIVE
**Default Behavior:** PARALLEL for eligible work packages
**Override Available:** Yes (document rationale required)

---

**Made with ❤️ by Cortexa LLC**

*Faster delivery through intelligent parallelization*
