# Tier 2 Stress Test Plan

**Purpose:** Exercise Claude Code workflows under real-world complex system stress

**Status:** Planning Phase
**Priority:** CRITICAL - Validates production readiness

---

## Gap Analysis: What's Missing from Current Tests

### Current Tier 2 Tests (Simulation Mode)
- ✅ Simple agent spawning
- ✅ File persistence
- ✅ Tool usage (Read, Write, Edit)
- ✅ Multi-agent coordination (3 agents)
- ✅ WIP limit enforcement
- ✅ Permissions validation

### Missing: Real-World Stress Scenarios

#### 1. Token Limit Stress (CRITICAL)
**Real-world scenario:** Large feature with 20+ files
- ❌ Agent creating 15-25 files (token limit risk)
- ❌ Agent output approaching 25K-32K token limit
- ❌ Truncated output mid-file
- ❌ Task decomposition required but not applied
- ❌ Recovery from token limit failure

#### 2. Parallel Agent Chaos (HIGH)
**Real-world scenario:** Complex system with multiple independent subtasks
- ❌ 5+ parallel background agents (exceeds WIP limit)
- ❌ Agents competing for shared resources
- ❌ Agent dependencies creating deadlocks
- ❌ Race conditions in file creation
- ❌ Incomplete agent output verification

#### 3. Silent Failure Scenarios (CRITICAL)
**Real-world scenario:** Agent claims success but artifacts missing
- ❌ Agent hits token limit, appears successful
- ❌ Partial file creation (5 of 10 files created)
- ❌ Files created in wrong directory
- ❌ Files created in sandbox instead of repository
- ❌ Agent crash without error reporting

#### 4. Complex Workflow Stress (HIGH)
**Real-world scenario:** Full feature workflow with specialists
- ❌ Cartographer → Architect → Designer → Engineer → Tester → Reviewer
- ❌ 6 sequential agents with file dependencies
- ❌ Each agent producing multiple files
- ❌ Cross-references between documents
- ❌ Verification at each gate

#### 5. File I/O Stress (MEDIUM)
**Real-world scenario:** Large codebase modifications
- ❌ Creating 20+ files in single agent
- ❌ Editing 10+ existing files
- ❌ Large file content (>5KB per file)
- ❌ Directory structure creation (nested)
- ❌ File path length issues

#### 6. Error Recovery Stress (HIGH)
**Real-world scenario:** Agent failures and retries
- ❌ Agent fails mid-execution
- ❌ Retry mechanism validation
- ❌ Partial completion handling
- ❌ Rollback on failure
- ❌ Error propagation to orchestrator

#### 7. Real Background Agent Behavior (CRITICAL)
**Real-world scenario:** Actual Task tool with run_in_background=True
- ❌ Real agent spawning (not simulation)
- ❌ Actual API calls
- ❌ Real output capture
- ❌ Completion detection
- ❌ Output file reading

---

## Proposed Stress Test Suite

### Test Category 1: Token Limit Stress Tests

**File:** `test_tier2_token_limits.py`

#### Test 1.1: Large File Count (15 files)
```python
def test_agent_creates_15_files():
    """
    Stress Test: Agent creates 15 files (approaching token limit)

    Scenario: Feature implementation requiring many files
    Expected: Agent completes OR decomposes task if limit hit
    """
    # Spawn agent to create 15 Python files
    # Each file ~500-1000 tokens
    # Total: ~15,000 tokens (within limit)
    # Verify all 15 files created
```

#### Test 1.2: Token Limit Exceeded (25 files)
```python
def test_agent_hits_token_limit_25_files():
    """
    Stress Test: Agent attempts 25 files (EXCEEDS token limit)

    Scenario: Task not properly decomposed
    Expected: Agent fails gracefully OR auto-decomposes
    Failure Mode: Truncated output, partial files
    """
    # Spawn agent to create 25 files
    # Total: ~25,000+ tokens (exceeds limit)
    # Verify detection of token limit
    # Verify partial completion flagged
    # Verify no "false success" reported
```

#### Test 1.3: Large Individual Files
```python
def test_agent_creates_large_files():
    """
    Stress Test: Agent creates files with large content

    Scenario: Generating comprehensive documentation
    Expected: Files complete, not truncated
    """
    # Create 5 files, each 3-5KB
    # Verify complete content (not truncated)
    # Check for mid-sentence cutoffs
```

#### Test 1.4: Task Decomposition Prevention
```python
def test_orchestrator_decomposes_large_task():
    """
    Stress Test: Orchestrator decomposes 30-file task

    Scenario: Large feature requiring decomposition
    Expected: Orchestrator creates 3 subtasks (10 files each)
    """
    # Initial task: 30 files
    # Orchestrator analyzes
    # Creates 3 task packets (10 files each)
    # Spawns 3 sequential agents
    # Verifies all 30 files created
```

---

### Test Category 2: Parallel Agent Chaos Tests

**File:** `test_tier2_parallel_agents.py`

#### Test 2.1: Maximum WIP (3 agents)
```python
def test_3_parallel_agents_at_wip_limit():
    """
    Stress Test: 3 parallel agents (at WIP limit)

    Scenario: Feature with 3 independent components
    Expected: All 3 agents complete successfully
    """
    # Spawn 3 background agents
    # Each creates 5 files
    # Verify all 15 files created
    # Verify no interference
```

#### Test 2.2: Exceeding WIP Limit (5 agents)
```python
def test_5_agents_exceeds_wip_limit():
    """
    Stress Test: Attempt to spawn 5 agents (EXCEEDS WIP)

    Scenario: Orchestrator tries too many parallel tasks
    Expected: Only 3 spawn, 2 queued
    """
    # Attempt to spawn 5 agents
    # Verify only 3 active
    # Verify 2 waiting
    # Verify sequential completion
```

#### Test 2.3: Agent Dependencies (Sequential)
```python
def test_sequential_agents_with_dependencies():
    """
    Stress Test: 5 agents with dependencies

    Scenario: Agent B depends on Agent A output
    Expected: Sequential execution, no deadlocks
    """
    # Agent A creates file
    # Agent B reads A's file, creates new file
    # Agent C reads B's file, creates new file
    # etc.
    # Verify correct sequence
```

#### Test 2.4: Race Condition Handling
```python
def test_agents_avoid_race_conditions():
    """
    Stress Test: 3 agents modifying same directory

    Scenario: Multiple agents creating files in same dir
    Expected: No file corruption, all files present
    """
    # 3 agents create files in same directory
    # Verify no file overwrites
    # Verify all files present
```

---

### Test Category 3: Silent Failure Detection Tests

**File:** `test_tier2_silent_failures.py`

#### Test 3.1: Partial File Creation
```python
def test_detect_partial_file_creation():
    """
    Stress Test: Agent creates only 5 of 10 files

    Scenario: Agent hits token limit mid-execution
    Expected: Orchestrator detects missing 5 files
    """
    # Agent task: Create 10 files
    # Agent only creates 5 (simulated failure)
    # Agent reports "success"
    # Orchestrator verification detects 5 missing
    # Status: FAILED (not success)
```

#### Test 3.2: Files in Wrong Directory
```python
def test_detect_files_in_wrong_directory():
    """
    Stress Test: Agent creates files in CWD instead of target

    Scenario: Relative paths used instead of absolute
    Expected: Orchestrator detects files missing from target
    """
    # Expected: .ai/tasks/*/files/
    # Actual: files/ (in CWD)
    # Orchestrator checks expected location
    # Detects files missing
```

#### Test 3.3: Sandbox vs Repository
```python
def test_detect_sandbox_files():
    """
    Stress Test: Agent creates files in sandbox (ephemeral)

    Scenario: Files don't persist after agent completes
    Expected: Orchestrator detects missing files
    """
    # Agent creates files (in sandbox)
    # Agent completes
    # Files disappear
    # Orchestrator verification: FAILED
```

#### Test 3.4: Truncated File Content
```python
def test_detect_truncated_files():
    """
    Stress Test: Token limit causes file truncation

    Scenario: Last file cut off mid-sentence
    Expected: Detection of incomplete file
    """
    # Agent creates files
    # Last file truncated (no closing brace, incomplete)
    # Verification detects truncation
```

---

### Test Category 4: Complex Workflow Stress Tests

**File:** `test_tier2_full_workflow.py`

#### Test 4.1: Complete Feature Workflow
```python
def test_complete_feature_workflow_with_specialists():
    """
    Stress Test: Full feature workflow (6 roles)

    Scenario: Large feature requiring all specialists
    Workflow: Cartographer → Architect → Designer → Engineer → Tester → Reviewer
    Expected: All deliverables created, all gates passed
    """
    # Phase 0: Planning
    # - Cartographer creates PRD (docs/product/)
    # - Architect creates design (docs/architecture/, docs/adr/)
    # - Designer creates wireframes (docs/design/)

    # Phase 1: Contract
    # - Orchestrator creates task packet

    # Phase 2: Implementation
    # - Engineer creates code + tests (10 files)

    # Phase 3: Review
    # - Tester validates (30-review.md)
    # - Reviewer validates (30-review.md)

    # Phase 4: Acceptance
    # - Orchestrator signs off (40-acceptance.md)

    # Verify:
    # - All 20+ files created
    # - All cross-references correct
    # - All gates passed
```

#### Test 4.2: Multi-Task Feature
```python
def test_feature_decomposed_into_3_tasks():
    """
    Stress Test: Large feature decomposed into 3 subtasks

    Scenario: 30-file feature split into 3 tasks (10 files each)
    Expected: All 3 tasks complete, all files created
    """
    # Task 1: Backend API (10 files)
    # Task 2: Frontend UI (10 files)
    # Task 3: Integration tests (10 files)

    # Verify all 30 files
    # Verify dependencies respected
```

---

### Test Category 5: Real Background Agent Tests

**File:** `test_tier2_real_agents.py`

**CRITICAL:** These use actual Task tool

#### Test 5.1: Single Real Agent
```python
def test_spawn_real_background_agent():
    """
    REAL Agent Test: Spawn actual background agent

    Uses: Task tool with run_in_background=True
    Expected: Agent spawns, executes, creates files
    """
    from claude_code import Task

    task = Task(
        subagent_type="general-purpose",
        description="Create test files",
        prompt=f"""
            Create 3 test files at:
            - {absolute_path}/file1.py
            - {absolute_path}/file2.py
            - {absolute_path}/file3.py

            Use Write tool with ABSOLUTE PATHS.
            Verify each file exists after creation.
        """,
        run_in_background=True
    )

    # Wait for completion
    # Verify 3 files exist
    # Verify in correct location
```

#### Test 5.2: Real Agent with Large Output
```python
def test_real_agent_creates_15_files():
    """
    REAL Agent Test: Agent creates 15 files (stress)

    Uses: Actual Task tool
    Expected: All 15 files created, no token limit hit
    """
    # Real agent task: Create 15 Python files
    # Verify all created
    # Check for truncation
```

#### Test 5.3: Real Parallel Agents
```python
def test_3_real_parallel_agents():
    """
    REAL Agent Test: 3 parallel background agents

    Uses: 3 Task tools simultaneously
    Expected: All complete, no interference
    """
    # Spawn 3 real agents in parallel
    # Each creates different files
    # Verify all deliverables
```

---

## Execution Strategy

### Phase 1: Create Stress Tests
1. `test_tier2_token_limits.py` (4 tests)
2. `test_tier2_parallel_agents.py` (4 tests)
3. `test_tier2_silent_failures.py` (4 tests)
4. `test_tier2_full_workflow.py` (2 tests)
5. `test_tier2_real_agents.py` (3 tests)

**Total:** 17 new stress tests

### Phase 2: Run in Simulation Mode
- Validate test structure
- Verify assertions correct
- Test cleanup working

### Phase 3: Run with Real Agents (CRITICAL)
- Execute in actual Claude Code environment
- Use real Task tool
- Monitor for failures
- Document issues found

### Phase 4: Iterate Based on Findings
- Fix issues discovered
- Add tests for new failure modes
- Re-run until stable

---

## Success Criteria

### Tier 2 Stress Tests Must Validate:

1. **Token Limits**
   - ✅ Agent completes 15 files successfully
   - ✅ Agent fails gracefully at 25 files
   - ✅ Orchestrator decomposes large tasks
   - ✅ No truncated output reported as success

2. **Parallel Agents**
   - ✅ 3 parallel agents complete successfully
   - ✅ WIP limit prevents 4th agent
   - ✅ Sequential dependencies respected
   - ✅ No race conditions

3. **Silent Failures**
   - ✅ Partial creation detected
   - ✅ Wrong directory detected
   - ✅ Sandbox files detected
   - ✅ Truncated files detected

4. **Complex Workflows**
   - ✅ Full feature workflow completes
   - ✅ Multi-task decomposition works
   - ✅ All cross-references correct

5. **Real Agents**
   - ✅ Real agent spawns and executes
   - ✅ Real agent creates persistent files
   - ✅ Multiple real agents coordinate

---

## Risk Assessment

### High Risk Scenarios (Must Test)
1. **Token limit with false success** - Agent appears successful but truncated
2. **Partial file creation** - Some files missing, agent claims complete
3. **Sandbox file creation** - Files don't persist to repository
4. **Agent deadlock** - Circular dependencies freeze execution
5. **Race conditions** - Parallel agents corrupt shared resources

### Medium Risk Scenarios
1. **Slow agent execution** - Timeout issues
2. **Large file content** - Memory issues
3. **Deep directory nesting** - Path length limits
4. **Permission errors** - Write failures

### Low Risk Scenarios
1. **Simple file creation** - Already validated in Tier 1
2. **Single agent execution** - Already validated in Tier 1
3. **Tool usage** - Already validated in Tier 1

---

## Next Steps

1. **Create stress test files** (5 files)
2. **Run in simulation mode** - Validate structure
3. **Prepare Claude Code environment** - Ensure Task tool available
4. **Execute real agent tests** - Monitor closely
5. **Document findings** - Issues discovered
6. **Iterate** - Fix and re-test

---

**Priority:** START with token limit tests (highest risk)
**Timeline:** Phase 1 creation: 1-2 hours
**Execution:** Phase 3 real agents: 30-60 minutes

