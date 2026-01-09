# Execution Strategy Gate

**Version:** 1.0.0
**Last Updated:** 2026-01-08
**Type:** Mandatory Enforcement Gate

## Overview

The Execution Strategy Gate is a **mandatory checkpoint** that enforces explicit analysis of parallelization opportunities before work delegation. This gate transforms parallel execution from optional best practice to required analysis.

**Purpose:** Ensure orchestrator automatically determines optimal execution strategy (parallel/sequential/hybrid) and documents the decision before delegating work.

---

## Trigger Conditions

This gate activates when:
```
IF orchestrator has 2+ subtasks ready for delegation THEN
  TRIGGER execution strategy gate
  REQUIRE explicit analysis
  BLOCK delegation until analysis complete
END IF
```

**Specifically triggers when:**
- Work package decomposed into 2+ subtasks
- Orchestrator preparing to delegate work
- Before spawning any worker agents

---

## Mandatory Analysis Procedure

The orchestrator MUST complete this analysis before delegating work:

### Step 1: Count and Classify Subtasks

```
PROCEDURE count_subtasks:
  subtasks = identify_all_subtasks(work_package)
  independent_count = count_independent_subtasks(subtasks)
  dependent_chains = identify_dependency_chains(subtasks)

  RETURN (subtasks, independent_count, dependent_chains)
END PROCEDURE
```

**Independence Criteria:**
```
Subtask is INDEPENDENT when ALL of:
✓ Modifies different files/modules from other subtasks
✓ No shared state or resources with other subtasks
✓ Can be tested independently
✓ Has isolated acceptance criteria
✓ No execution order dependency on other subtasks
✓ Compatible with shared context constraints (see below)
```

**Shared Context Constraints (CRITICAL):**

Parallel workers operate in a SHARED context - they cannot create isolated environments. This imposes critical constraints:

```
✅ SHARED resources (all workers access same):
   - Source repository (single git working directory)
   - Build folders (shared build tree)
   - Test databases (coordinate access or use different schemas)
   - Coverage data files (must merge, not overwrite)
   - Configuration files
   - Dependency caches

❌ FORBIDDEN during parallel execution:
   - Deleting/recreating build folders
   - Removing coverage data files
   - Creating per-worker git branches
   - git reset --hard or other destructive operations
   - Operations that invalidate shared context
   - Modifying shared configuration in incompatible ways

⚠️ REQUIRES COORDINATION:
   - Build operations (may conflict on same targets)
   - Coverage report generation (merge results properly)
   - Database migrations (must sequence these)
   - Git commits (coordinate timing, no conflicts)
   - Shared file modifications (different workers, different files)

💡 DESIGN FOR SHARED CONTEXT:
   - Assign different files to different workers
   - Use different database schemas if parallel DB work needed
   - Plan for coverage merging at orchestrator level
   - Coordinate build targets if possible
   - Document shared resource usage in worker assignments
```

---

### Step 2: Assess Dependency Structure

```
PROCEDURE assess_dependencies:
  FOR each subtask pair (A, B):
    IF B requires output from A THEN
      mark dependency: A → B
    END IF
    IF A and B modify same files THEN
      mark conflict: A ⚡ B
    END IF
  END FOR

  RETURN dependency_graph
END PROCEDURE
```

**Dependency Types:**
- **Execution dependency:** B needs A's results
- **File conflict:** Both modify same file
- **Resource conflict:** Both need exclusive access to resource
- **Ordering constraint:** Semantic requirement for sequence

---

### Step 3: Determine Execution Strategy

```
PROCEDURE determine_strategy:
  IF independent_count >= 3 AND no_conflicts THEN
    RETURN "PARALLEL" with justification
  ELSE IF independent_count >= 2 AND has_some_dependencies THEN
    RETURN "HYBRID" with justification
  ELSE IF strong_dependencies OR independent_count < 3 THEN
    RETURN "SEQUENTIAL" with justification
  END IF
END PROCEDURE
```

**Strategy Decision Matrix:**
```
Subtasks | Independence | Dependencies | Strategy    | Justification Required
---------|--------------|--------------|-------------|----------------------
3+       | All          | None         | PARALLEL    | No (automatic)
3+       | Mixed        | Some         | HYBRID      | Yes (explain split)
3+       | None         | Strong       | SEQUENTIAL  | Yes (explain deps)
1-2      | Any          | Any          | SEQUENTIAL  | No (too few tasks)
```

---

### Step 4: Document Decision

The orchestrator MUST document the execution strategy decision:

**Required Documentation:**
```markdown
## Execution Strategy Analysis

**Subtask Count:** [N]
**Independent Subtasks:** [M]
**Dependencies:** [List or "None"]

**Strategy:** PARALLEL | SEQUENTIAL | HYBRID

**Justification:**
[Explain why this strategy was chosen]

**Parallel Workers (if applicable):**
- Worker 1: [Subtask description]
- Worker 2: [Subtask description]
- Worker 3: [Subtask description]
- Worker 4: [Subtask description]

**Execution Plan:**
[How workers will be spawned and coordinated]
```

**Documentation Location:**
- In task packet `10-plan.md` (preferred)
- OR in orchestrator's analysis output before delegation
- OR in work package contract

---

## Enforcement Rules

### Rule 1: Analysis Cannot Be Skipped

```
ENFORCEMENT:
  IF orchestrator has 2+ subtasks THEN
    REQUIRE execution_strategy_analysis()
    BLOCK delegation until analysis documented

    IF orchestrator proceeds without analysis THEN
      GATE VIOLATION
      HALT execution
      REPORT to user
    END IF
  END IF
```

---

### Rule 2: Parallel Execution is Mandatory When Criteria Met

```
ENFORCEMENT:
  IF independent_count >= 3 AND no_conflicts THEN
    REQUIRE parallel execution
    REQUIRE worker spawning in single message

    IF orchestrator uses sequential execution THEN
      GATE VIOLATION (unless justified)
      REQUIRE documented justification
    END IF
  END IF
```

**Acceptable justifications for sequential when parallel possible:**
- Hidden dependencies discovered during analysis
- Resource constraints (though should still try up to 4 workers)
- Experimental/exploratory work with unclear boundaries
- User explicitly requested sequential

---

### Rule 3: Strategy Must Match Documented Decision

```
ENFORCEMENT:
  documented_strategy = read_documented_strategy()
  actual_execution = observe_execution()

  IF actual_execution != documented_strategy THEN
    GATE VIOLATION
    REPORT mismatch to user
    REQUIRE explanation
  END IF
```

---

## Execution Strategy Types

### PARALLEL Strategy

**When to Use:**
- 3+ independent subtasks
- No cross-subtask dependencies
- Different files/modules modified
- Isolated acceptance criteria

**Implementation:**
```
PARALLEL EXECUTION:
  workers = []
  FOR each independent subtask (up to 5):
    workers.append(Task(worker, subtask_description))
  END FOR

  // CRITICAL: Launch all workers in SINGLE message
  launch_all(workers)  // True parallelism

  monitor_progress(workers)
  coordinate_integration_points()
```

**Benefits:**
- 3-4x faster delivery
- Independent verification
- Clear ownership
- Reduced bottlenecks

---

### SEQUENTIAL Strategy

**When to Use:**
- 1-2 subtasks
- Strong execution dependencies
- Same files modified sequentially
- Tightly coupled changes

**Implementation:**
```
SEQUENTIAL EXECUTION:
  worker = Task(worker, combined_subtasks)
  wait_for_completion(worker)
  verify_results()
```

**Benefits:**
- Simpler coordination
- No conflict resolution needed
- Appropriate for small tasks

---

### HYBRID Strategy

**When to Use:**
- Mix of dependent and independent subtasks
- Some must execute in sequence
- Others can run in parallel

**Implementation:**
```
HYBRID EXECUTION:
  // Phase 1: Execute dependent chain
  FOR each dependent_subtask in order:
    worker = Task(worker, dependent_subtask)
    wait_for_completion(worker)
  END FOR

  // Phase 2: Parallel execution of independent group
  workers = []
  FOR each independent_subtask (up to 5):
    workers.append(Task(worker, independent_subtask))
  END FOR
  launch_all(workers)

  // Phase 3: Integration
  verify_integration()
```

**Benefits:**
- Optimizes both sequence and parallelism
- Respects dependencies
- Maximizes parallelization where possible

---

## Analysis Template

Use this template when performing execution strategy analysis:

```markdown
# Execution Strategy Analysis

## Subtask Inventory
1. [Subtask 1 description]
   - Files: [list]
   - Dependencies: [none | depends on N]
   - Independent: [yes | no]

2. [Subtask 2 description]
   - Files: [list]
   - Dependencies: [none | depends on N]
   - Independent: [yes | no]

3. [Subtask 3 description]
   - Files: [list]
   - Dependencies: [none | depends on N]
   - Independent: [yes | no]

## Independence Assessment
- **Total subtasks:** [N]
- **Independent subtasks:** [M]
- **Dependent chains:** [describe or "none"]
- **File conflicts:** [list or "none"]

## Strategy Decision
**Chosen Strategy:** PARALLEL | SEQUENTIAL | HYBRID

**Rationale:**
[Explain why this strategy is appropriate based on analysis above]

## Implementation Plan
**Parallel Workers (if applicable):**
- Worker 1: [subtask + files]
- Worker 2: [subtask + files]
- Worker 3: [subtask + files]
- Worker 4: [subtask + files]

**Sequencing (if applicable):**
1. [Phase/subtask 1]
2. [Phase/subtask 2]
...

**Launch Pattern:**
[Single message with N Task calls | Sequential execution | Hybrid phases]

## Coordination Plan
[How integration points will be handled]
[How conflicts will be resolved if they arise]
```

---

## Integration with Other Gates

This gate integrates with:

- **[Global Gates](00-global-gates.md)** - Incremental progress, quality baseline
- **[Tool Policy](20-tool-policy.md)** - Agent spawning policy enforcement
- **[Verification Gates](30-verification.md)** - Post-execution verification

**Relationship:**
```
Global Gates → Set baseline requirements
     ↓
Execution Strategy Gate → Enforce parallelization analysis
     ↓
Tool Policy → Define agent spawning patterns
     ↓
Verification Gates → Validate results
```

---

## Examples

### Example 1: Three Independent API Endpoints (PARALLEL)

**Scenario:** Implement user profile, notifications, and settings APIs

**Analysis:**
```
Subtasks:
1. User profile API (src/api/profile.js + tests)
2. Notifications API (src/api/notifications.js + tests)
3. Settings API (src/api/settings.js + tests)

Independence: ✓ All independent
- Different files
- No shared state
- Isolated acceptance criteria

Strategy: PARALLEL (mandatory)
Workers: 3 (one per endpoint)
```

**Implementation:**
```
Single message with 3 Task calls:
- Task(worker, "Implement user profile API + tests")
- Task(worker, "Implement notifications API + tests")
- Task(worker, "Implement settings API + tests")
```

---

### Example 2: Database Migration + UI Updates (HYBRID)

**Scenario:** Update user schema, then add 3 UI components

**Analysis:**
```
Subtasks:
1. Database migration (must run first)
2. UI component 1 (depends on migration)
3. UI component 2 (depends on migration)
4. UI component 3 (depends on migration)

Dependencies:
- Components 2-4 depend on component 1
- Components 2-4 are independent of each other

Strategy: HYBRID
- Phase 1: Sequential (migration)
- Phase 2: Parallel (3 UI components)
```

**Implementation:**
```
Phase 1 (sequential):
  Task(worker, "Database migration + tests")
  wait_for_completion()

Phase 2 (parallel, single message):
  Task(worker, "UI component 1 + tests")
  Task(worker, "UI component 2 + tests")
  Task(worker, "UI component 3 + tests")
```

---

### Example 3: Tightly Coupled Refactoring (SEQUENTIAL)

**Scenario:** Refactor authentication flow across 5 interconnected files

**Analysis:**
```
Subtasks:
1. Update auth module
2. Update middleware (depends on auth module)
3. Update controllers (depends on middleware)
4. Update tests (depends on all above)

Dependencies: Strong sequential chain
File conflicts: Multiple subtasks touch same files

Strategy: SEQUENTIAL (justified)
Workers: 1 (handles entire refactoring)
```

**Implementation:**
```
Single worker:
  Task(worker, "Complete authentication refactoring per plan")
```

---

## Gate Compliance Checklist

Before delegating work, verify:

```
□ Subtask count determined
□ Independence assessed for each subtask
□ Dependencies identified and documented
□ Execution strategy determined (PARALLEL/SEQUENTIAL/HYBRID)
□ Strategy justification documented
□ Worker spawning plan created
□ Launch pattern defined (single message for parallel)

IF all checked THEN
  GATE PASSED - proceed with delegation
ELSE
  GATE BLOCKED - complete missing analysis
END IF
```

---

## Violation Handling

### Violation Types

**Type 1: Analysis Skipped**
```
Orchestrator delegates work without strategy analysis

Response:
- HALT execution
- REPORT: "Execution Strategy Gate violation - analysis required"
- REQUIRE: Complete analysis before proceeding
```

**Type 2: Parallel Execution Not Used When Required**
```
3+ independent subtasks exist but orchestrator uses sequential

Response:
- REPORT: "Parallel execution required for 3+ independent subtasks"
- REQUIRE: Justification for sequential approach OR
- REQUIRE: Switch to parallel execution
```

**Type 3: Strategy Mismatch**
```
Documented strategy differs from actual execution

Response:
- REPORT: "Strategy mismatch detected"
- REQUIRE: Update documentation OR
- REQUIRE: Adjust execution to match strategy
```

---

## Success Criteria

This gate is successful when:

- ✓ Orchestrator automatically analyzes parallelization for every work package
- ✓ Strategy determination is explicit and documented
- ✓ Parallel workers spawn automatically when criteria met
- ✓ No user reminders needed for parallelization
- ✓ Execution matches documented strategy

---

## References

- **[Orchestrator Role](../roles/orchestrator.md)** - Section 2.5: Mandatory Parallel Execution Analysis
- **[Tool Policy Gate](20-tool-policy.md)** - Section 7: Agent Spawning Policy
- **[Standard Workflow](../workflows/standard.md)** - Phase 2.4: Execution Strategy Determination
- **[Parallel Workers Config](../PARALLEL-ENGINEERS-CONFIG.md)** - Enforced Configuration Details

---

**Last reviewed:** 2026-01-08
**Next review:** When execution patterns evolve or enforcement needs adjustment
**Status:** ACTIVE - Mandatory enforcement gate
