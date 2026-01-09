# Parallel Engineers Configuration

**Version:** 1.1.0
**Last Updated:** 2026-01-08
**Status:** ENFORCED CONFIGURATION

---

## Overview

The ai-pack framework **automatically enforces** the use of **multiple parallel engineers** for resolving work packages when criteria are met. This enforcement ensures optimal delivery speed and resource utilization while maintaining quality and coordination.

**Enforcement Mechanism:** [Execution Strategy Gate](gates/25-execution-strategy.md)

---

## Enforced Configuration

### Automatic Parallelization Analysis

**MANDATORY ANALYSIS** for work packages with 2+ subtasks. **ENFORCED PARALLEL EXECUTION** when criteria are met:

```
✅ MANDATORY PARALLEL EXECUTION when ALL of:
- 3 or more independent subtasks identified
- Subtasks modify different files/modules
- No cross-subtask dependencies
- Each subtask has isolated acceptance criteria
- Shared context constraints can be managed

IF criteria met AND orchestrator uses sequential THEN
  GATE VIOLATION (Execution Strategy Gate)
  REQUIRE documented justification
END IF
```

**Maximum Parallel Engineers:** 5 concurrent agents

**Launch Pattern:** Single message block (true parallelism) - REQUIRED, not optional

**Shared Context:** All engineers operate in same source repository, build folders, and working directory

---

## Configuration Rules

### When Parallel Engineers Are Used (DEFAULT)

```
Work Package Type                    | Strategy          | Workers
-------------------------------------|-------------------|----------
3+ independent API endpoints         | Parallel (DEFAULT)| 2-5
3+ independent modules/components    | Parallel (DEFAULT)| 2-5
3+ similar features (no dependencies)| Parallel (DEFAULT)| 2-5
Multiple bug fixes (different areas) | Parallel (DEFAULT)| 2-5
Independent refactoring tasks        | Parallel (DEFAULT)| 2-5
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

### Enforced Delegation Pattern

The Orchestrator role MUST analyze and apply appropriate execution strategy. This is enforced by the Execution Strategy Gate:

```pseudocode
FUNCTION delegate_work_package_ENFORCED(work_package):
  // MANDATORY GATE: Execution Strategy Analysis
  subtasks = analyze_and_decompose(work_package)

  // GATE REQUIREMENT: Must complete analysis for 2+ subtasks
  IF subtasks.count >= 2 THEN
    ENFORCE execution_strategy_gate_analysis(subtasks)
    // Gate blocks until analysis documented
  END IF

  // Analysis includes shared context evaluation
  shared_context_conflicts = assess_shared_context(subtasks)

  IF subtasks.count >= 3 AND are_independent(subtasks) AND no_context_conflicts THEN
    // ENFORCED: Must spawn parallel engineers (not optional)
    ASSERT parallel_execution_required("3+ independent subtasks with manageable context")

    workers = []
    FOR each subtask in subtasks.take(5):  // Max 5 workers
      workers.append(spawn_worker(subtask))
    END FOR

    // REQUIRED: Launch all workers in single message block
    launch_in_single_message(workers)  // True parallelism
    monitor_parallel_progress(workers)
    coordinate_shared_context(workers)

  ELSE IF has_dependencies(subtasks) OR has_context_conflicts THEN
    // ENFORCED: Hybrid approach for mixed scenarios
    dependent_chain = extract_dependencies(subtasks)
    independent_group = extract_independent(subtasks)

    // Sequence dependent tasks first
    execute_sequential(dependent_chain)

    // THEN parallelize independent tasks
    IF independent_group.count >= 3 THEN
      ENFORCE parallel_execution(independent_group)
    END IF

  ELSE
    // Sequential acceptable for 1-2 subtasks or strong coupling
    DOCUMENT_RATIONALE("Sequential justified: " + reason)
    execute_sequential(subtasks)
  END IF

  // Gate verification
  VERIFY_STRATEGY_MATCHES_DOCUMENTED()
END FUNCTION
```

**Key Enforcement Points:**
- ✅ Analysis MANDATORY for 2+ subtasks
- ✅ Parallel execution REQUIRED for 3+ independent subtasks
- ✅ Single message launch ENFORCED (not sequential spawning)
- ✅ Shared context constraints EVALUATED
- ✅ Strategy/execution match VERIFIED

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
✓ Respect shared context constraints (no build deletion, etc.)
✓ Coordinate shared resource access
```

**Shared Context Constraints (CRITICAL):**
```
All parallel engineers share:
- Source repository (single git working directory)
- Build folders (no deletion/recreation allowed)
- Coverage data files (merge, don't overwrite)
- Test databases (coordinate access)
- Configuration files

Workers MUST NOT:
- Delete or recreate build folders
- Remove coverage data
- Create per-worker git branches
- Perform destructive git operations
- Invalidate shared context for other workers

Workers MUST coordinate:
- Build operations (avoid conflicts)
- Coverage generation (merge results)
- Database access (use different schemas if needed)
- Git commits (timing and conflict resolution)
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

→ AUTOMATIC: Spawn 3 parallel engineers (up to 4 if more endpoints)

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

### Expected Improvements with Parallel Engineers

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

## Quality Assurance with Parallel Engineers

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
- Set parallel engineers as default for 3+ independent subtasks
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
