# TC-BA-003: Working Directory Context Verification

**Category:** Spawned Agents
**Priority:** Critical
**Status:** Active
**Last Updated:** 2026-01-15

---

## Objective

Validate that orchestrators pass working directory context to ALL spawned agents and that agents write files to the correct repository location (not isolated sandboxes).

## Background

**Production Failure (consumer-project 2026-01-15):**
- Agent claimed: "Created tools/schema-validation/package.json"
- Orchestrator: "✅ Completed successfully"
- Reality: `ls tools/schema-validation/` → "No such file or directory"
- **Root cause:** Agent wrote to its temp workspace (sandbox), not repository

**The Problem:**
Background agents run in isolated contexts with different working directories. Without explicit repository root context, they write files successfully... but to the WRONG location.

**File Creation Flow (Without Fix):**
```
1. Orchestrator pwd: /Users/user/consumer-project
2. Spawns spawned agent
3. Agent pwd: /tmp/claude-agent-a9a6821/  ← DIFFERENT!
4. Agent: Write("tools/schema-validation/package.json")
5. File created: /tmp/claude-agent-a9a6821/tools/schema-validation/package.json ✅
6. Repository: /Users/user/consumer-project/tools/schema-validation/package.json ❌ MISSING
```

**Impact:**
- Files claim success but don't persist
- Build failures when files don't exist
- Test coverage gaps
- False progress reporting

## Prerequisites

- Project with spawned agent capability
- Test directory structure
- `.claude/settings.json` configured

## Test Scenario

### Setup Phase

1. **Create test task requiring file creation:**
   ```bash
   mkdir -p .ai/tasks/local-20260115090000-test-ba-003
   cp .ai-pack/templates/task-packet/* .ai/tasks/local-20260115090000-test-ba-003/
   ```

2. **Fill out contract:**
   ```markdown
   ## Requirements
   Create tools directory with package.json

   ## Acceptance Criteria
   - [ ] Directory exists: tools/schema-validation/
   - [ ] File exists: tools/schema-validation/package.json
   - [ ] File contains valid JSON
   ```

### Execution Phase - Without Working Directory Context (Should Fail)

3. **❌ WRONG: Orchestrator delegates without working directory context:**

   ```python
   # Orchestrator SKIPS pwd command
   # Does NOT capture PROJECT_ROOT

   Task(
     subagent_type="general-purpose",
     description="Create schema validation tools",
     prompt="""Engineer role from .ai-pack/roles/engineer.md

     Task: Create tools/schema-validation/package.json

     Requirements:
     - Directory: tools/schema-validation/
     - File: package.json with basic schema validation config

     Task packet: .ai/tasks/local-20260115090000-test-ba-003/
     Follow TDD. Update work log.
     """,
     
   )
   ```

4. **Agent executes (in sandbox):**
   ```
   Agent pwd: /tmp/claude-agent-xyz/  ← Wrong location!

   Agent: mkdir -p tools/schema-validation
   Agent: Write("tools/schema-validation/package.json", content)

   Agent: "✅ Created tools/schema-validation/package.json"
   Agent: "✅ Work complete"
   ```

5. **Verification reveals failure:**
   ```bash
   # Check repository
   ls /Users/user/project/tools/schema-validation/
   # Result: No such file or directory ❌

   # File exists in agent sandbox
   ls /tmp/claude-agent-xyz/tools/schema-validation/
   # Result: package.json ✅ (but wrong location!)
   ```

### Execution Phase - With Working Directory Context (Should Succeed)

6. **✅ CORRECT: Orchestrator MUST capture working directory FIRST:**

   ```bash
   # Orchestrator MUST run this BEFORE spawning agent
   PROJECT_ROOT=$(pwd)
   echo "Working directory: $PROJECT_ROOT"
   ```

7. **✅ CORRECT: Orchestrator passes working directory to agent:**

   ```python
   Task(
     subagent_type="general-purpose",
     description="Create schema validation tools",
     prompt="""Engineer role from .ai-pack/roles/engineer.md

     CRITICAL WORKING DIRECTORY CONTEXT:
     - Repository root: /Users/user/project
     - All file paths MUST be absolute or verified relative to this root
     - ALWAYS use absolute paths: Write("/Users/user/project/path/to/file")
     - OR verify location first: cd /Users/user/project && pwd && mkdir path

     Task: Create tools/schema-validation/package.json

     Requirements:
     - Directory: /Users/user/project/tools/schema-validation/
     - File: package.json with basic schema validation config

     Task packet: .ai/tasks/local-20260115090000-test-ba-003/
     Follow TDD. Update work log.
     """,
     
   )
   ```

8. **Agent executes (with context):**
   ```
   Agent receives: Repository root: /Users/user/project

   Agent: PROJECT_ROOT="/Users/user/project"
   Agent: mkdir -p "$PROJECT_ROOT/tools/schema-validation"
   Agent: Write("$PROJECT_ROOT/tools/schema-validation/package.json", content)

   Agent: "✅ Created /Users/user/project/tools/schema-validation/package.json"
   Agent: "✅ Work complete"
   ```

9. **Verification succeeds:**
   ```bash
   # Check repository
   ls /Users/user/project/tools/schema-validation/
   # Result: package.json ✅

   # Verify contents
   cat /Users/user/project/tools/schema-validation/package.json
   # Result: Valid JSON ✅
   ```

### Verification Checklist

10. **Orchestrator MUST verify:**

    ```bash
    # Step 1: Orchestrator captured working directory
    echo "PROJECT_ROOT=$PROJECT_ROOT"
    # Expected: /Users/user/project

    # Step 2: Working directory context in Task() prompt
    grep "CRITICAL WORKING DIRECTORY CONTEXT" <orchestrator-output>
    grep "Repository root:" <orchestrator-output>
    # Expected: Both found

    # Step 3: Files exist in repository (not sandbox)
    ls "$PROJECT_ROOT/tools/schema-validation/package.json"
    # Expected: File exists

    # Step 4: Files NOT in sandbox location
    # (Can't easily verify but lack of files in repo is the symptom)
    ```

## Expected Behavior

### Scenario A: Without Working Directory Context (WRONG)

**Orchestrator:**
```
❌ Does NOT run pwd
❌ Does NOT capture PROJECT_ROOT
❌ Does NOT pass working directory to agent
```

**Agent:**
```
Receives no repository location context
Works in default location (sandbox)
Creates files successfully in sandbox
Reports: "✅ Created tools/schema-validation/package.json"
```

**Verification:**
```
❌ File does NOT exist in repository
✅ File exists in agent sandbox (wrong location)
Status: FAILED - Files not persisted to repository
```

**What Should Happen:**
- TC-OR-001 verification detects missing files
- Orchestrator reports CRITICAL FAILURE
- Recovery initiated

### Scenario B: With Working Directory Context (CORRECT)

**Orchestrator:**
```
✅ Runs: PROJECT_ROOT=$(pwd)
✅ Captures: /Users/user/project
✅ Passes working directory in Task() prompt
✅ Uses "CRITICAL WORKING DIRECTORY CONTEXT:" header
```

**Agent:**
```
Receives repository root: /Users/user/project
Uses absolute paths: $PROJECT_ROOT/tools/...
Creates files in repository location
Reports: "✅ Created /Users/user/project/tools/schema-validation/package.json"
```

**Verification:**
```
✅ File exists in repository
✅ File contains expected content
Status: SUCCESS - Files persisted correctly
```

## Actual Behavior (Execution Record)

**Test Run:** [Date]

**Scenario A (Without Context):**
- Orchestrator captured pwd: [Yes/No]
- Working directory in prompt: [Yes/No]
- File exists in repository: [Yes/No]
- Agent claimed success: [Yes/No]
- Verification status: [Pass/Fail]

**Scenario B (With Context):**
- Orchestrator captured pwd: [Yes/No]
- Working directory in prompt: [Yes/No]
- Repository root correct: [Path]
- File exists in repository: [Yes/No]
- File location correct: [Yes/No]
- Verification status: [Pass/Fail]

**Deviations:**
[Any differences from expected]

## Pass/Fail Criteria

### PASS Criteria

**Orchestrator Behavior:**
✅ Orchestrator runs `pwd` BEFORE spawning agent
✅ Orchestrator captures PROJECT_ROOT
✅ Orchestrator includes "CRITICAL WORKING DIRECTORY CONTEXT:" in prompt
✅ Orchestrator passes repository root path explicitly
✅ All Task() calls include working directory context

**Agent Behavior:**
✅ Agent receives repository location
✅ Agent uses absolute paths or verified relative paths
✅ Agent creates files in repository location
✅ Agent reports absolute paths in success messages

**Verification:**
✅ Files exist in repository (not sandbox)
✅ Files at correct paths
✅ Verification protocol detects if files missing

### FAIL Criteria

❌ Orchestrator doesn't run pwd
❌ Orchestrator doesn't capture working directory
❌ No working directory context in Task() prompt
❌ Agent works in sandbox without repository context
❌ Files created in agent sandbox only
❌ Files missing from repository
❌ Relative paths used without location verification

## Known Issues

**Issue 1: Sandbox Isolation**
- Background agents may run in isolated temp directories
- Default working directory ≠ repository root
- Files written to relative paths go to sandbox
- **Mitigation:** ALWAYS pass absolute repository root

**Issue 2: Multiple Working Directories**
- Orchestrator pwd: /Users/user/project
- Agent pwd: /tmp/claude-agent-xyz/
- Engineer pwd in another session: /different/path/
- **Mitigation:** Explicit PROJECT_ROOT in every Task() call

**Issue 3: Relative Paths Are Dangerous**
- `Write("tools/file.json")` → writes to cwd (unknown location)
- `mkdir tools` → creates in cwd (might be sandbox)
- **Mitigation:** Always use absolute paths with $PROJECT_ROOT

## Working Directory Context Template

**MANDATORY pattern for all Task() calls:**

```python
# STEP 1: Capture working directory FIRST
PROJECT_ROOT=$(pwd)

# STEP 2: Pass to agent in EVERY Task() prompt
Task(
  subagent_type="general-purpose",
  description="[task description]",
  prompt="""[Role] from .ai-pack/roles/[role].md

  CRITICAL WORKING DIRECTORY CONTEXT:
  - Repository root: {PROJECT_ROOT}
  - All file paths MUST be absolute or verified relative to this root
  - ALWAYS use: Write("{PROJECT_ROOT}/path/to/file")
  - OR verify first: cd {PROJECT_ROOT} && pwd && [operation]

  Task: [task details]

  Task packet: .ai/tasks/[task-id]/
  Follow TDD. Update work log.
  """,
  
)
```

## Real-World Examples

**consumer-project Failures:**

**Example 1: Schema Validation**
```
❌ Agent: "Created tools/schema-validation/package.json"
❌ Reality: ls tools/schema-validation/ → No such file
✅ Fix: Pass PROJECT_ROOT, agent uses absolute paths
```

**Example 2: Test Files**
```
❌ Agent: "Created server/Tests/Unit/GraphQL/UserMutationsTests.cs"
❌ Reality: File doesn't exist in repository
✅ Fix: Working directory context in Task() prompt
```

**Example 3: Nested Directories**
```
❌ Agent: mkdir server/API (while pwd = /path/server/)
❌ Result: /path/server/server/API/ (nested disaster)
✅ Fix: Use PROJECT_ROOT=$(git rev-parse --show-toplevel)
        mkdir "$PROJECT_ROOT/server/API"
```

## Recovery Procedures

**If files not persisted:**

1. **Check agent output for file paths mentioned**
   ```bash
   grep -E "Created|Wrote|Generated" <agent-output>
   ```

2. **Determine if agent had working directory context**
   ```bash
   grep "CRITICAL WORKING DIRECTORY CONTEXT" <orchestrator-output>
   grep "Repository root:" <task-prompt>
   ```

3. **If context missing:**
   - Add PROJECT_ROOT=$(pwd) before Task()
   - Add working directory block to prompt
   - Re-spawn agent with context
   - Verify files in repository

4. **If files in sandbox:**
   - Extract file content from agent output
   - Manually create in repository
   - Verify correctness
   - Update task status

## Metrics

**Before Working Directory Context Fix:**
- File persistence rate: ~30% (many files in sandbox)
- Manual recovery required: ~70% of spawned agents
- Average time to discover: 10-30 minutes (after build fails)

**After Working Directory Context Fix:**
- File persistence rate: 100% (when context provided)
- Manual recovery required: 0%
- Immediate verification detects issues

## References

- **Commit:** `801fb14` - CRITICAL FIX: Pass working directory context to ALL spawned agents
- **Orchestrator Skill:** Lines 9-22 (FILE PERSISTENCE REQUIREMENT)
- **Orchestrator Skill:** Lines 135-148 (Basic delegation pattern with context)
- **Real Failure:** consumer-project tools/schema-validation/package.json not created
- **Related:** TC-BA-001 (File Persistence - reactive), TC-OR-001 (Completion Verification)

## Test Automation Hooks

```python
# Future automation
def test_working_directory_context():
    # 1. Spawn agent WITHOUT context
    agent1 = spawn_agent_no_context(task)
    assert file_exists_in_repo(expected_file) == False  # FAIL (expected)
    assert file_exists_in_sandbox(agent1.workspace, expected_file) == True

    # 2. Spawn agent WITH context
    project_root = os.getcwd()
    agent2 = spawn_agent_with_context(task, project_root)
    assert file_exists_in_repo(expected_file) == True  # SUCCESS
    assert file_path_is_absolute(agent2.created_files[0]) == True
```

---

**Related Tests:**
- TC-BA-001 (File Persistence Verification)
- TC-BA-004 (Absolute Path Requirements)
- TC-OR-001 (Completion Verification)
