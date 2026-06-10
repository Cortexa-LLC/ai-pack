# TC-OR-005: Large Task Decomposition and Token Budget Management

**Category:** Orchestrator
**Priority:** Critical
**Status:** Active
**Last Updated:** 2026-01-15

---

## Objective

Validate that orchestrators properly analyze task size and decompose large tasks (15+ files, complex context) into smaller work packets that remain under 25K token budget per agent.

## Background

**Production Failure (Harvana 2026-01-15):**
- Task: "Implement WunderGraph Cosmo Gateway" (25 files)
- Orchestrator attempted to delegate entire task to single background engineer
- Attempt 1: Token limit, 0 files created
- Attempt 2: Token limit again, 0 files created
- Even with concise instructions, task too large for single agent
- **Root cause:** Orchestrator didn't decompose task before delegation

**Impact:**
- Multiple failed attempts before success
- Wasted hours of agent work
- User frustration
- Framework credibility damage

**The Fix (Commit e1764ec):**
Added task size guidelines and decomposition strategy:
- ✅ SAFE: 3-8 files, single component (~15-20K tokens)
- ⚠️ RISKY: 10-15 files, may fail if complex
- ❌ TOO LARGE: 15+ files, WILL fail, MUST split

## Prerequisites

- Large, complex task requiring 15+ files
- Orchestrator role activated
- Task packet created
- Access to planning phase

## Test Scenario

### Setup Phase

1. **Create test task with large scope:**
   ```bash
   mkdir -p .ai/tasks/local-20260115090000-test-or-005
   cp .ai-pack/templates/task-packet/* .ai/tasks/local-20260115090000-test-or-005/
   ```

2. **Fill out contract with complex multi-component task:**
   ```markdown
   ## Requirements
   Implement WunderGraph Cosmo Gateway integration

   Components:
   1. Gateway foundation (3 files)
      - docker-compose.yml
      - gateway/config.yaml
      - gateway/Dockerfile

   2. Server setup (4 files)
      - server/index.ts
      - server/schema.ts
      - server/resolvers.ts
      - server/Dockerfile

   3. Schema definitions (9 files)
      - schemas/user.graphql
      - schemas/flags.graphql
      - schemas/analytics.graphql
      - schemas/admin.graphql
      - schemas/operations.graphql
      - schemas/mutations.graphql
      - schemas/queries.graphql
      - schemas/subscriptions.graphql
      - schemas/types.graphql

   4. Authentication (3 files)
      - auth/middleware.ts
      - auth/validators.ts
      - auth/config.ts

   5. Deployment (6 files)
      - deploy/k8s/gateway-deployment.yaml
      - deploy/k8s/gateway-service.yaml
      - deploy/k8s/gateway-configmap.yaml
      - deploy/k8s/server-deployment.yaml
      - deploy/k8s/server-service.yaml
      - deploy/scripts/deploy.sh

   **Total: 25 files across 5 components**

   ## Acceptance Criteria
   - [ ] All components implemented
   - [ ] Integration tested
   - [ ] Deployment automated
   ```

### Execution Phase - Task Analysis (MANDATORY)

3. **Orchestrator MUST perform task size analysis:**

   **Step 1: Count total files**
   ```python
   # Extract from contract
   total_files = count_files_in_requirements()
   # Expected: 25 files

   print(f"Total files required: {total_files}")
   ```

   **Step 2: Apply task size guidelines**
   ```python
   if total_files >= 15:
       print("❌ TOO LARGE: 15+ files, WILL fail, MUST split")
       decomposition_required = True
   elif total_files >= 10:
       print("⚠️ RISKY: 10-15 files, may fail if complex")
       decomposition_recommended = True
   else:
       print("✅ SAFE: 3-8 files, single component")
       decomposition_not_needed = True
   ```

   **Step 3: Analyze natural component boundaries**
   ```markdown
   ## Component Analysis

   Natural boundaries identified:
   1. Gateway Foundation (3 files) - Docker config, gateway setup
   2. Server Setup (4 files) - GraphQL server implementation
   3. Schema Definitions (9 files) - GraphQL schemas
   4. Authentication (3 files) - Auth middleware and config
   5. Deployment (6 files) - Kubernetes configs and scripts

   Dependencies:
   - Foundation must exist before Server
   - Server must exist before Schemas
   - All above must exist before Auth
   - Everything must exist before Deployment

   File count per component:
   - Foundation: 3 ✅ (safe)
   - Server: 4 ✅ (safe)
   - Schemas: 9 ⚠️ (risky but acceptable - related files)
   - Auth: 3 ✅ (safe)
   - Deployment: 6 ✅ (safe)

   Total components: 5
   Average files/component: 5
   Max files in single component: 9
   ```

### Execution Phase - Decomposition Strategy (MANDATORY)

4. **Orchestrator MUST document decomposition strategy:**

   **Document in task.md:**
   ```markdown
   ## Task Decomposition Analysis

   **Original Task:** Implement WunderGraph Cosmo Gateway (25 files)

   **Size Assessment:**
   - Total files: 25
   - Guideline: ❌ TOO LARGE (>15 files)
   - Decision: MUST decompose into smaller tasks

   **Decomposition Strategy:**

   Split into 5 independent subtasks:

   ### Subtask 1: Gateway Foundation
   - Files: 3 (docker-compose.yml, config.yaml, Dockerfile)
   - Token estimate: ~8K tokens
   - Risk: ✅ Low (simple config files)
   - Dependencies: None (can start immediately)

   ### Subtask 2: Server Setup
   - Files: 4 (index.ts, schema.ts, resolvers.ts, Dockerfile)
   - Token estimate: ~12K tokens
   - Risk: ✅ Low (standard GraphQL server)
   - Dependencies: Gateway Foundation (needs base config)

   ### Subtask 3: Schema Definitions
   - Files: 9 (various .graphql files)
   - Token estimate: ~18K tokens
   - Risk: ⚠️ Medium (9 files but all similar structure)
   - Dependencies: Server Setup (schemas reference server types)

   ### Subtask 4: Authentication
   - Files: 3 (middleware.ts, validators.ts, config.ts)
   - Token estimate: ~10K tokens
   - Risk: ✅ Low (focused auth logic)
   - Dependencies: Server Setup (integrates with server)

   ### Subtask 5: Deployment
   - Files: 6 (K8s configs, deploy script)
   - Token estimate: ~10K tokens
   - Risk: ✅ Low (declarative configs)
   - Dependencies: All above (deploys complete system)

   **Execution Plan:**
   - Phase 1 (Parallel): Foundation + Server (no conflict)
   - Phase 2 (Sequential): Schemas (depends on Server)
   - Phase 3 (Parallel): Auth + Deployment prep (can overlap)
   - Phase 4 (Sequential): Final deployment (needs everything)

   **Benefits:**
   - Each subtask <10 files (within safe range)
   - Each subtask <20K tokens (safe margin)
   - Parallel execution where possible (2-3x speedup)
   - Incremental progress visible
   - Easier debugging if one fails
   ```

5. **Orchestrator MUST NOT delegate without decomposition:**

   **❌ WRONG (What failed in production):**
   ```python
   # Attempt to delegate entire 25-file task
   Task(
     subagent_type="general-purpose",
     description="Implement WunderGraph Gateway",  # 25 files!
     prompt="""Implement all 25 files for gateway...""",
     
   )
   # Result: Token limit, 0 files created
   ```

   **✅ CORRECT (After decomposition):**
   ```python
   # Phase 1: Foundation (3 files)
   Task(
     subagent_type="general-purpose",
     description="Gateway Foundation (3 files)",
     prompt="""Engineer role. Working dir: /repo
     Task: Implement gateway foundation (3 config files)
     Task packet: .ai/tasks/local-20260115090000-gateway-foundation/
     Follow TDD. Update work log.""",
     
   )

   # Phase 1: Server (4 files) - parallel with Foundation
   Task(
     subagent_type="general-purpose",
     description="GraphQL Server (4 files)",
     prompt="""Engineer role. Working dir: /repo
     Task: Implement GraphQL server (4 files)
     Task packet: .ai/tasks/local-20260115090000-graphql-server/
     Follow TDD. Update work log.""",
     
   )

   # ... wait for Phase 1 completion ...

   # Phase 2: Schemas (9 files)
   Task(
     subagent_type="general-purpose",
     description="GraphQL Schemas (9 files)",
     prompt="""Engineer role. Working dir: /repo
     Task: Implement GraphQL schemas (9 related files)
     Task packet: .ai/tasks/local-20260115090000-graphql-schemas/
     Follow TDD. Update work log.""",
     
   )

   # ... and so on for Auth and Deployment ...
   ```

### Verification Phase (MANDATORY)

6. **Verify decomposition occurred:**

   **Check 1: Task size analysis documented**
   ```bash
   grep -i "Total files:" .ai/tasks/*/task.md
   grep -i "TOO LARGE\|RISKY\|SAFE" .ai/tasks/*/task.md
   ```
   Expected: Evidence of size analysis

   **Check 2: Decomposition strategy documented**
   ```bash
   grep -i "Subtask\|Component\|Split" .ai/tasks/*/task.md
   ```
   Expected: Clear decomposition into smaller chunks

   **Check 3: Multiple task packets created**
   ```bash
   ls .ai/tasks/ | grep -E "gateway|server|schema|auth|deploy"
   ```
   Expected: 5 separate task packet directories

   **Check 4: Each subtask within limits**
   ```bash
   for task in .ai/tasks/*/; do
     file_count=$(grep -c "File:" "$task/task.md" || echo 0)
     echo "$task: $file_count files"
   done
   ```
   Expected: All subtasks <10 files

   **Check 5: Agents spawned per subtask (not all at once)**
   ```bash
   grep "Task(" .ai/tasks/*/result.md | wc -l
   ```
   Expected: 5 Task() calls (one per subtask)

7. **Verify token budget estimates:**

   **For each subtask, estimate token consumption:**
   ```python
   def estimate_token_budget(subtask):
       # Rough estimation
       files = count_files(subtask)
       avg_file_size = 500  # tokens per file
       instruction_overhead = 500  # tokens for instructions
       context_overhead = 1000  # tokens for task packet context

       estimated_tokens = (files * avg_file_size) + instruction_overhead + context_overhead

       return estimated_tokens

   for subtask in subtasks:
       estimate = estimate_token_budget(subtask)
       print(f"{subtask.name}: ~{estimate} tokens")

       if estimate > 25000:
           print(f"  ❌ FAIL: Exceeds 25K token budget")
       elif estimate > 20000:
           print(f"  ⚠️ WARNING: Near token limit")
       else:
           print(f"  ✅ SAFE: Within token budget")
   ```

## Expected Behavior

### If Working Correctly:

**Step 1: Orchestrator receives large task (25 files)**
```
User: "Implement WunderGraph Cosmo Gateway"

Orchestrator: "I've received a request to implement the gateway.
Let me analyze the task size first..."
```

**Step 2: Orchestrator analyzes task size**
```
Orchestrator: "Analyzing task requirements...

Total files required: 25
Components identified: 5
  - Gateway Foundation: 3 files
  - Server Setup: 4 files
  - Schema Definitions: 9 files
  - Authentication: 3 files
  - Deployment: 6 files

Task size assessment: ❌ TOO LARGE (>15 files)

Per guidelines (commit e1764ec), tasks >15 files WILL fail with token limits.
Decomposition is MANDATORY."
```

**Step 3: Orchestrator documents decomposition**
```
Orchestrator: "I'll decompose this into 5 subtasks, each within safe limits.

Documenting decomposition strategy in task packet..."

[Writes detailed decomposition to task.md]
```

**Step 4: Orchestrator creates subtask packets**
```
Orchestrator: "Creating task packets for each component..."

✓ Created .ai/tasks/local-20260115090000-gateway-foundation/
✓ Created .ai/tasks/local-20260115090000-graphql-server/
✓ Created .ai/tasks/local-20260115090000-graphql-schemas/
✓ Created .ai/tasks/local-20260115090000-authentication/
✓ Created .ai/tasks/local-20260115090000-deployment/
```

**Step 5: Orchestrator spawns agents for subtasks**
```
Orchestrator: "Spawning background engineers for parallel execution...

Phase 1 (Parallel):
  - Engineer-1: Gateway Foundation (3 files, ~8K tokens)
  - Engineer-2: GraphQL Server (4 files, ~12K tokens)

Monitoring completion before Phase 2..."
```

**Step 6: All subtasks complete successfully**
```
Orchestrator: "All 5 components completed successfully.

Verification:
  ✓ Gateway Foundation: 3/3 files created
  ✓ GraphQL Server: 4/4 files created
  ✓ GraphQL Schemas: 9/9 files created
  ✓ Authentication: 3/3 files created
  ✓ Deployment: 6/6 files created

Total: 25/25 files created ✅

Integration testing..."
```

### If Failing (Regression):

**Orchestrator attempts single delegation (WRONG):**
```
Orchestrator: "I'll delegate the gateway implementation to an engineer."

[Spawns single agent with 25-file task]

Engineer: [Hits token limit after 20K tokens of planning]
          [Makes 0 Write() calls]
          [Creates 0 files]

Orchestrator: [Should detect failure via TC-OR-001]
              "❌ Agent failed - token limit exceeded"
```

**What should have happened:**
- Orchestrator analyzes size BEFORE delegating
- Detects 25 files = TOO LARGE
- Decomposes into 5 subtasks
- Each subtask succeeds

## Actual Behavior (Execution Record)

**Test Run:** [Date]

**Task Given:**
- Description:
- Total files:
- Total components:

**Orchestrator Analysis:**
- Size analysis performed: [Yes/No]
- Analysis documented in task.md: [Yes/No]
- Size assessment: [SAFE/RISKY/TOO LARGE]

**Decomposition:**
- Decomposition performed: [Yes/No]
- Number of subtasks created:
- Subtask packets created: [List]
- Max files per subtask:
- Max token estimate per subtask:

**Execution:**
- Delegation approach: [Single agent / Multiple agents]
- Agents spawned:
- Results:

**Deviations:**
[Any differences from expected behavior]

## Pass/Fail Criteria

### PASS Criteria

**Task Analysis:**
✅ Orchestrator counts total files
✅ Orchestrator applies size guidelines correctly
✅ Orchestrator identifies TOO LARGE tasks (>15 files)
✅ Analysis documented in task.md

**Decomposition:**
✅ Large tasks decomposed into subtasks
✅ Each subtask <10 files (safe range)
✅ Each subtask <20K token estimate
✅ Decomposition strategy documented
✅ Natural component boundaries respected
✅ Dependencies identified

**Execution:**
✅ Separate task packets created per subtask
✅ Agents spawned per subtask (not all at once)
✅ All subtasks complete successfully
✅ No token limit failures
✅ All files created as expected

### FAIL Criteria

❌ Orchestrator doesn't analyze task size
❌ Orchestrator attempts to delegate 15+ file task to single agent
❌ No decomposition strategy documented
❌ Subtasks exceed 10 files
❌ Subtasks exceed 20K token estimate
❌ Token limit failures occur
❌ Files not created due to task size

## Known Issues

**Issue 1: No Automated File Counting**
- Orchestrators must manually count files from requirements
- Easy to miss in complex specifications
- **Mitigation:** Explicit guidelines in orchestrator.md

**Issue 2: Token Estimation Imprecise**
- File size varies greatly
- Context overhead varies
- **Mitigation:** Conservative thresholds (10 files vs 15+)

**Issue 3: Component Boundaries Not Always Clear**
- Some tasks don't have natural boundaries
- Requires judgment call
- **Mitigation:** When in doubt, split into smaller chunks

## Task Size Guidelines Reference

From commit e1764ec, Orchestrator Skill lines 306-488:

**Safe Task Size:**
- **3-8 files:** Single component, ~15-20K tokens
- **Single responsibility:** One feature, one module
- **Minimal context:** Self-contained

**Risky Task Size:**
- **10-15 files:** May succeed if files are simple/boilerplate
- **Higher risk:** If files complex or tightly coupled
- **Decision:** Evaluate complexity, consider splitting

**Too Large (Must Split):**
- **15+ files:** WILL fail with token limit
- **Multiple components:** Different responsibilities
- **High context:** Cross-cutting concerns

**Decomposition Triggers:**
```
IF task requires 15+ files THEN
  MUST decompose (no exceptions)
ELSE IF task requires 10-14 files THEN
  EVALUATE complexity
  IF complex OR tightly coupled THEN
    decompose
  END IF
ELSE IF task requires <10 files THEN
  ACCEPTABLE for single agent
END IF
```

## Real-World Example

**Harvana WunderGraph Gateway (25 files):**

**❌ What Happened (Failed Approach):**
```
Attempt 1: Single agent, all 25 files → Token limit, 0 files
Attempt 2: Single agent, concise prompt → Token limit, 0 files
Attempt 3: Single agent, even more concise → Token limit, 0 files
Attempt 4: Single agent, minimal prompt → Token limit, 0 files
Attempt 5: Finally decomposed into components → SUCCESS
```

**✅ What Should Have Happened:**
```
Attempt 1: Analyze size (25 files = TOO LARGE)
           Decompose into 5 components
           Spawn 5 agents in 2 phases
           All succeed → SUCCESS in first attempt
```

**Time Comparison:**
- Failed approach: 5 attempts × ~30 min = 2.5 hours wasted
- Correct approach: Decompose (10 min) + Execute (30 min parallel) = 40 min total
- **Savings:** 1 hour 50 minutes

## Metrics

**Success Indicators:**
- **Decomposition Rate:** % of large tasks that get decomposed
  - Target: 100% for tasks >15 files
- **First-Attempt Success:** % of decomposed tasks that succeed first time
  - Target: >90%
- **Token Limit Failures:** Count of failures due to task size
  - Target: 0 (after decomposition)

**Before Decomposition Guidance:**
- Large tasks (15+ files): 100% failure rate
- Average attempts to success: 4-5
- Wasted agent time: 2-3 hours per task

**After Decomposition Guidance (Expected):**
- Large tasks decomposed: 100%
- First-attempt success: >90%
- Wasted agent time: Near zero

## Recovery Procedures

**If orchestrator doesn't decompose large task:**

1. **Detect during planning review**
   ```bash
   # Check plan for size analysis
   grep -i "total files\|task size\|decomposition" .ai/tasks/*/task.md
   ```

2. **Intervene before execution**
   ```
   User: "This task is 25 files. Per TC-OR-005 guidelines,
          please decompose into subtasks <10 files each."
   ```

3. **Require decomposition**
   ```
   User: "Before delegating, document:
          1. Total file count
          2. Size assessment (SAFE/RISKY/TOO LARGE)
          3. Decomposition strategy if >15 files"
   ```

4. **Verify before proceeding**
   - Check decomposition documented
   - Verify subtask size <10 files
   - Confirm token estimates <20K

## References

- **Commit:** `e1764ec` - Add task decomposition guidance to prevent token limit failures
- **Orchestrator Skill:** Lines 306-488 (Decompose Large Tasks)
- **Real Failure:** Harvana 2026-01-15 (WunderGraph gateway - 25 files, 5 failed attempts)
- **Related Test:** TC-BA-002 (Token Limit Detection - reactive)
- **This Test:** TC-OR-005 (Task Decomposition - proactive prevention)

## Test Automation Hooks

```python
# Future automation
def test_task_decomposition():
    # 1. Create large task (25 files)
    task = create_task(files=25)

    # 2. Invoke orchestrator
    orchestrator = Orchestrator(task)

    # 3. Verify analysis occurred
    assert orchestrator.analyzed_size == True
    assert orchestrator.file_count == 25
    assert orchestrator.size_assessment == "TOO_LARGE"

    # 4. Verify decomposition occurred
    assert orchestrator.decomposed == True
    assert len(orchestrator.subtasks) >= 3

    # 5. Verify subtask sizes
    for subtask in orchestrator.subtasks:
        assert subtask.file_count <= 10
        assert subtask.token_estimate <= 20000

    # 6. Verify execution succeeded
    results = orchestrator.execute()
    assert results.token_limit_failures == 0
    assert results.files_created == 25
```

---

**Next Test:** TC-OR-006 (Parallel Execution Verification)
**Related Tests:** TC-BA-002 (Token Limit Detection), TC-OR-001 (Completion Verification)
