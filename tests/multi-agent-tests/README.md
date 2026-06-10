# Multi-Agent Testing Guide

**Purpose:** Validate AI-Pack framework with real Claude Code spawned agents

---

## Overview

Multi-Agent tests use **real Claude Code spawned agents** (not simulation) to validate:
- Task execution with actual API calls
- File persistence to repository
- Token limit handling
- Parallel and sequential workflows
- Accurate status reporting (no silent failures)

**Difference from Tier 1:**
- **Tier 1:** Python unittest with simulated agents
- **Multi-Agent:** Real Claude Code agents spawned via Task tool

---

## Test Categories

### 1. Sequential Execution Tests (Completed ✅)
**File:** `test_multi-agent_real_execution.py`

Tests single agents creating multiple files:
- 3-file baseline
- 15-file moderate stress
- 20-file high stress
- 25-file limit test

**Status:** All 4 tests passed, documented in:
- `test-results.md`
- `test-results.md`
- `TIER2-20FILE-STRESS-RESULTS.md`
- `TIER2-25FILE-LIMIT-RESULTS.md`

### 2. Parallel Execution Tests (Completed ✅)
**File:** `test_multi-agent_parallel_execution.py`

Tests multiple agents running simultaneously:
- 5 Engineers working on different features in parallel
- Directory-based isolation
- No conflicts or race conditions

**Status:** Passed, documented in:
- `.ai/test-artifacts/multi-agent-parallel-*/test-results.md`

### 3. Sequential Workflow Tests (Pending)
**File:** `test_multi-agent_sequential_workflows.py` (to be created)

Tests multi-stage workflows with handoffs:
- PRD → Architect → Engineer → Reviewer → Tester
- Proper artifact passing between stages
- Gate enforcement

### 4. Complex Mixed Workflows (Pending)
**File:** `test_multi-agent_complex_workflows.py` (to be created)

Tests combining parallel and sequential:
- Orchestrator → Multiple Engineers (parallel) → Multiple Reviewers (parallel) → Tester

---

## Running Multi-Agent Tests

### Prerequisites
- Claude Code CLI installed and configured
- API access to Claude API (for spawned agents)
- Working directory: `/Users/brywoodruff/Projects/Vibe/ai-pack`

### Step 1: Prepare Test Infrastructure

Run the setup test to create contracts:

```bash
cd tests
python -m pytest test_multi-agent_parallel_execution.py::TestTier2ParallelExecution::test_01_prepare_parallel_5_engineers_task -v
```

This creates:
```
.ai/test-artifacts/multi-agent-parallel-{timestamp}/
├── tasks/
│   ├── feature-1-auth/task.md
│   ├── feature-2-api/task.md
│   ├── feature-3-cache/task.md
│   ├── feature-4-validator/task.md
│   └── feature-5-logger/task.md
├── output/ (empty, will be populated by agents)
├── README.md
├── test-metadata.json
└── verify_parallel_execution.py
```

### Step 2: Spawn Spawned Agents

**IMPORTANT:** This step must be done via Claude Code interactive session, NOT pytest.

In a Claude Code session:

```
User: "Spawn 5 Engineer agents in parallel for the contracts in
.ai/test-artifacts/multi-agent-parallel-{timestamp}/tasks/.
Each agent should read its contract and create 3 Python files."
```

Claude Code will use the Task tool to spawn 5 spawned agents:

```python
Task(
    subagent_type="general-purpose",
    description="Auth feature engineer",
    ,
    prompt="Read contract at .../feature-1-auth/task.md and create all files..."
)
# ... repeat for features 2-5
```

### Step 3: Monitor Execution

Claude Code will receive task notifications as agents complete:

```
<task-notification>
<task-id>aff1ecc</task-id>
<status>completed</status>
<summary>Agent "Auth feature engineer" completed</summary>
</task-notification>
```

Wait for all 5 agents to complete.

### Step 4: Verify Results

Run the verification script:

```bash
cd .ai/test-artifacts/multi-agent-parallel-{timestamp}
python verify_parallel_execution.py
```

Expected output:
```
✅ feature-1-auth: 3 files, 133 lines
✅ feature-2-api: 3 files, 158 lines
✅ feature-3-cache: 3 files, 166 lines
✅ feature-4-validator: 3 files, 169 lines
✅ feature-5-logger: 3 files, 183 lines

📊 Summary:
   Total files: 15
   Total lines: 800
   Features: 5

✅ PASS: All parallel agents completed successfully!
```

---

## Test Results Format

Each test execution should produce:
1. **Test artifacts directory:** `.ai/test-artifacts/multi-agent-{type}-{timestamp}/`
2. **Results document:** `TIER2-{TYPE}-TEST-RESULTS.md`
3. **Agent outputs:** `/tmp/claude/-Users-.../tasks/{agent_id}.output`
4. **Deliverable files:** In `output/` subdirectory

---

## Understanding Test Outputs

### During Execution

**Agent Progress Notifications:**
```
<system-reminder>
Agent a9bfe09 progress: 7 new tools used, 32642 new tokens.
The agent is still running.
</system-reminder>
```

**Agent Completion:**
```
<task-notification>
<task-id>a9bfe09</task-id>
<status>completed</status>
<summary>Agent "API feature engineer" completed</summary>
<result>... detailed completion report ...</result>
</task-notification>
```

### After Execution

**Agent Transcripts:**
- Location: `/tmp/claude/-Users-brywoodruff-Projects-Vibe-ai-pack/tasks/{agent_id}.output`
- Contains: Full conversation between Claude Code and the spawned agent
- Useful for: Debugging, understanding agent decisions

**Deliverable Files:**
- Location: `.ai/test-artifacts/multi-agent-{type}-{timestamp}/output/`
- Contents: Actual Python files created by agents
- Persisted: YES (in repository, not sandbox)

---

## File Locations Explained

### Repository Files (Persistent)
```
/Users/brywoodruff/Projects/Vibe/ai-pack/
├── .ai/test-artifacts/          # Test outputs (PERSISTED)
│   └── multi-agent-parallel-*/
│       ├── tasks/               # Test contracts
│       └── output/              # Agent deliverables (PYTHON FILES HERE)
│           ├── feature-1-auth/
│           │   ├── auth/authenticator.py
│           │   ├── auth/session_manager.py
│           │   └── tests/test_auth.py
│           ├── feature-2-api/
│           ├── feature-3-cache/
│           ├── feature-4-validator/
│           └── feature-5-logger/
└── tests/
    ├── test_multi-agent_parallel_execution.py  # Test setup
    └── test-results.md           # This file
```

### Temporary Files (Not Persisted)
```
/tmp/claude/-Users-brywoodruff-Projects-Vibe-ai-pack/tasks/
├── aff1ecc.output   # Agent transcript (Auth)
├── a9bfe09.output   # Agent transcript (API)
├── aa31b82.output   # Agent transcript (Cache)
├── aa99d85.output   # Agent transcript (Validator)
└── a18853c.output   # Agent transcript (Logger)
```

**Key Distinction:**
- **Deliverables** (Python files): `.ai/test-artifacts/.../output/` (REPOSITORY)
- **Agent logs** (transcripts): `/tmp/claude/.../tasks/` (TEMPORARY)

---

## Repeating Tests

### Re-running Parallel Test

**Option 1: Create New Test Run**
```bash
python -m pytest test_multi-agent_parallel_execution.py::TestTier2ParallelExecution::test_01_prepare_parallel_5_engineers_task -v
```
This creates a NEW timestamped directory with fresh contracts.

**Option 2: Reuse Existing Contracts**
```bash
# Find existing test directory
ls -lt .ai/test-artifacts/ | grep multi-agent-parallel

# In Claude Code session, reference existing contracts:
"Spawn 5 agents using contracts in .ai/test-artifacts/multi-agent-parallel-1768514135/tasks/"
```

### Creating Custom Tests

1. **Copy and modify** `test_multi-agent_parallel_execution.py`
2. **Adjust** feature contracts in the `features` list
3. **Run setup test** to generate contracts
4. **Spawn agents** via Claude Code
5. **Verify results** with verification script

---

## Success Criteria

### Per-Agent Success
- ✅ All files created at correct absolute paths
- ✅ Files in designated feature directory only
- ✅ Complete content (not truncated)
- ✅ Valid Python syntax
- ✅ No interference with other agents

### System-Level Success
- ✅ All agents complete without errors
- ✅ No silent failures (accurate status reporting)
- ✅ Files persist to repository (not sandbox)
- ✅ Proper isolation (no conflicts)
- ✅ Task notifications received for all agents

---

## Troubleshooting

### Issue: Files Not Found in output/

**Check:**
1. Did agents complete successfully? (Look for task-notification)
2. Are files in `.ai/test-artifacts/multi-agent-*/output/` or `/tmp/`?
3. Run: `find .ai/test-artifacts -name "*.py" -type f`

**Common Cause:** Agent created files in wrong directory
**Solution:** Ensure contracts specify ABSOLUTE paths

### Issue: Agent Timeout

**Symptoms:** Agent progress stops, no completion notification

**Check:**
1. Read agent output: `tail /tmp/claude/.../tasks/{agent_id}.output`
2. Look for errors or blocking prompts

**Common Cause:** Agent asking for user input
**Solution:** Ensure contracts are fully specified (no ambiguity)

### Issue: File Conflicts

**Symptoms:** Agents overwriting each other's files

**Check:**
1. Verify each feature has separate directory
2. Confirm contracts specify unique paths

**Solution:** Use directory-based isolation (feature-1-auth/, feature-2-api/, etc.)

---

## Next Test Development

### Creating Sequential Workflow Test

**File:** `test_multi-agent_sequential_workflows.py`

**Structure:**
```python
def test_01_prepare_prd_to_tester_workflow():
    """Setup: Create 5-stage workflow contracts"""
    # Create contracts:
    # 1. PRD agent (produces requirements.md)
    # 2. Architect agent (reads requirements.md, produces architecture.md)
    # 3. Engineer agent (reads architecture.md, produces code)
    # 4. Reviewer agent (reads code, produces review.md)
    # 5. Tester agent (reads code + review, produces test_results.md)
```

**Key Differences from Parallel:**
- Agents run SEQUENTIALLY (not simultaneously)
- Each agent READS previous agent's output
- Handoffs validated via gates
- Final deliverable integrates all stages

---

## Test Metrics

### What to Track

**Execution Metrics:**
- Agent completion time (start to finish)
- Token usage per agent
- Number of tool calls
- Completion order (for parallel tests)

**Quality Metrics:**
- Files created vs expected
- Lines of code produced
- Python syntax validation
- Contract compliance

**Reliability Metrics:**
- Success rate (agents completing successfully)
- Silent failure detection
- File persistence rate
- Isolation effectiveness (no conflicts)

### Example Metrics Report

```
Test: Multi-Agent Parallel Execution (5 Engineers)
Date: 2026-01-15
Duration: ~3-4 minutes (all agents)

Agents: 5
Total Files: 15
Total Lines: 800
Total Size: 24.3 KB

Avg Token Usage: 32,155 tokens/agent
Success Rate: 100% (5/5 agents)
Silent Failures: 0
Conflicts: 0
```

---

## Best Practices

### Contract Design

**DO:**
- ✅ Use absolute paths for all file deliverables
- ✅ Provide complete code specifications (not "implement X")
- ✅ Include acceptance criteria
- ✅ Specify working directory
- ✅ Warn about parallel context (if applicable)

**DON'T:**
- ❌ Use relative paths
- ❌ Leave implementation details ambiguous
- ❌ Omit file structure requirements
- ❌ Mix agents' working directories

### Agent Spawning

**DO:**
- ✅ Spawn all parallel agents in single Claude Code response
- ✅ Use descriptive agent names
- ✅ Set ``
- ✅ Provide full context in agent prompt

**DON'T:**
- ❌ Spawn agents one-by-one with delays
- ❌ Reuse agent IDs
- ❌ Assume agents share context (they don't)

### Verification

**DO:**
- ✅ Verify file count matches expected
- ✅ Check Python syntax with `py_compile`
- ✅ Validate directory isolation
- ✅ Document metrics and results

**DON'T:**
- ❌ Assume completion = success without verification
- ❌ Skip syntax validation
- ❌ Ignore agent transcripts (useful for debugging)

---

## Summary

**Multi-Agent tests validate:**
1. Real agent execution (not simulation)
2. Parallel and sequential workflows
3. File persistence to repository
4. Accurate status reporting
5. Framework reliability under stress

**Process:**
1. Run setup test → Creates contracts
2. Spawn agents via Claude Code → Real execution
3. Monitor completion → Wait for notifications
4. Verify results → Run verification script
5. Document outcomes → Create results markdown

**Current Status:**
- ✅ Sequential tests (3, 15, 20, 25 files)
- ✅ Parallel test (5 agents simultaneously)
- ⏳ Sequential workflows (pending)
- ⏳ Complex mixed workflows (pending)

---

**Last Updated:** 2026-01-15
**Framework Version:** AI-Pack (post Tier 1 validation)
