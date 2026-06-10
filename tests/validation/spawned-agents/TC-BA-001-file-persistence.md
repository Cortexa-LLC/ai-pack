# TC-BA-001: Spawned Agent File Persistence Verification

**Category:** Spawned Agents
**Priority:** Critical
**Status:** Active
**Last Updated:** 2026-01-15

---

## Objective

Validate that spawned agents successfully persist files to the repository and that the orchestrator correctly verifies file persistence before declaring work complete.

## Background

**Production Failure (consumer-project 2026-01-15):**
- Background Engineer reported "✅ Created: server/Tests/Unit/GraphQL/UserMutationsTests.cs"
- Orchestrator declared "completed successfully"
- Reality: File did not exist in repository (`ls` showed "No such file or directory")
- Root cause: Agent worked in isolated sandbox, files not written back to repository

**Impact:**
- False progress reporting
- Wasted agent work hours
- Missing test files leaving coverage gaps
- Build failures downstream

## Prerequisites

- Project with `.claude/settings.json` configured
- Background agent permissions enabled (`Write(*)` in allow list)
- Test project with source directory structure

## Test Scenario

### Setup Phase

1. Create test task packet:
   ```bash
   mkdir -p .ai/tasks/local-20260115090000-test-ba-001
   cp .ai-pack/templates/task-packet/* .ai/tasks/local-20260115090000-test-ba-001/
   ```

2. Fill out contract with clear file creation requirements:
   ```markdown
   ## Requirements
   Create test file: src/tests/sample-test.js

   ## Acceptance Criteria
   - [ ] File exists at src/tests/sample-test.js
   - [ ] File contains valid test code
   - [ ] File is not empty
   ```

### Execution Phase

3. Orchestrator spawns background Engineer agent:
   ```python
   Task(
     subagent_type="general-purpose",
     description="Create sample test file",
     prompt="""You are the Engineer from .ai-pack/roles/engineer.md.

     CRITICAL WORKING DIRECTORY CONTEXT:
     - Repository root: /Users/brywoodruff/Projects/Vibe/test-project
     - All file paths MUST be absolute or relative to this root
     - Verify with pwd before creating files

     Task: Create file src/tests/sample-test.js with a simple test.

     Update work log at .ai/tasks/local-20260115090000-test-ba-001/result.md
     """,
     
   )
   ```

4. Wait for agent completion

5. **CRITICAL: Orchestrator MUST execute verification protocol**

### Verification Phase (MANDATORY)

6. **Step 1: Check for error patterns in agent output**
   ```bash
   # Read agent output
   tail -100 <agent-output-file>

   # Look for:
   # - "exceeded.*token.*maximum"
   # - "API Error"
   # - "rate limit"
   # - "permission denied"
   # - "tool.*failed"
   ```

7. **Step 2: Verify Write() calls made**
   ```bash
   # Check agent output for Write() tool usage
   grep -c "Write()" <agent-output-file>
   # Expected: > 0
   ```

8. **Step 3: Verify claimed files exist**
   ```bash
   # Extract claimed file from agent output
   CLAIMED_FILE=$(grep "Created.*:" <agent-output> | sed 's/.*Created: //')

   # Verify file exists
   if [ -f "$CLAIMED_FILE" ]; then
     echo "✓ File exists: $CLAIMED_FILE"
   else
     echo "❌ MISSING: $CLAIMED_FILE (agent claimed to create it)"
     exit 1
   fi
   ```

9. **Step 4: Verify file is not empty**
   ```bash
   if [ -s "$CLAIMED_FILE" ]; then
     echo "✓ File not empty"
   else
     echo "⚠️ WARNING: File exists but is EMPTY"
     exit 1
   fi
   ```

10. **Step 5: Report verification results**
    ```markdown
    ## Artifact Persistence Verification

    Expected files: 1
    Found files: 1
    Missing files: 0

    ✓ src/tests/sample-test.js (512 bytes)

    Status: ✅ VERIFIED - Files persisted successfully
    ```

## Expected Behavior

### If Working Correctly:

1. **Agent completes successfully**
   - Makes Write() tool calls
   - Creates file at correct path
   - File contains expected content
   - Agent reports completion

2. **Orchestrator verification protocol executes**
   - Checks agent output for errors → None found
   - Verifies Write() calls > 0 → Yes
   - Verifies file exists → Yes
   - Verifies file not empty → Yes

3. **Orchestrator reports success with verification**
   ```
   Background Engineer completed.

   Artifact Persistence Verification:
     Expected files: 1
     Found files: 1
     Missing files: 0

     ✓ src/tests/sample-test.js (512 bytes)

   Status: ✅ VERIFIED - Files persisted successfully
   ```

### If Failing (Regression):

1. **Agent completes but files missing**
   - Agent reports "created file"
   - But ls shows "No such file"
   - Verification detects: MISSING FILES

2. **Orchestrator MUST detect and report failure**
   ```
   Background Engineer completed.

   Artifact Persistence Verification:
     Expected files: 1
     Found files: 0
     Missing files: 1

     ❌ src/tests/sample-test.js (MISSING)

   Status: 🚨 CRITICAL FAILURE - No files persisted

   Root Cause: Agent worked in isolated sandbox

   Recovery Options:
     1. Extract content from agent output
     2. Re-run agent in foreground mode
     3. Fix permissions and re-run in background
   ```

## Actual Behavior (Execution Record)

**Test Run:** [Date]

**Result:** [Pass | Fail]

**Details:**
- Agent completion status:
- Verification step 1 (error check):
- Verification step 2 (Write calls):
- Verification step 3 (file exists):
- Verification step 4 (not empty):
- Orchestrator reporting:

**Deviations:**
[Any differences from expected behavior]

**Screenshots/Logs:**
[Attach relevant output]

## Pass/Fail Criteria

### PASS Criteria

✅ Agent creates file successfully
✅ File exists in repository after completion
✅ File is not empty
✅ Orchestrator executes ALL 5 verification steps
✅ Orchestrator reports verification results
✅ No false success reported

### FAIL Criteria

❌ Agent reports success but file doesn't exist
❌ Orchestrator skips verification steps
❌ Orchestrator declares success without verifying files
❌ File exists but is empty
❌ Agent hits token limit (verification should catch)

## Known Issues

**Issue 1: Sandbox Isolation**
- Some spawned agents may work in isolated sandboxes
- Files created in sandbox don't persist to repository
- **Mitigation:** Working directory context in prompt (Section 2.14 orchestrator.md)

**Issue 2: Permission Failures**
- Write(*) permission not configured
- Agent silently fails to write files
- **Mitigation:** Permission verification before spawning (Section 2.14 orchestrator.md)

**Issue 3: Token Limit**
- Agent hits 32K token limit before writing files
- Reports "completed" despite no work done
- **Mitigation:** Concise instructions + error detection (TC-BA-002)

## Recovery Procedures

### If Files Missing After Agent Completion:

1. **Check agent output for actual file content**
   ```bash
   # If agent output contains full file content
   # Extract and manually create file
   ```

2. **Re-run in foreground mode**
   ```python
   # Remove 
   Task(..., )
   ```

3. **Fix permissions and retry**
   ```bash
   # Verify .claude/settings.json
   grep "Write(\*)" .claude/settings.json
   # Re-run agent
   ```

## References

- **Commit:** `edf1d8a` - CRITICAL: Fix token limit failures + false success reporting
- **Orchestrator Role:** `roles/orchestrator.md` Section 2.15 (Artifact Persistence Verification)
- **Real Failure:** consumer-project 2026-01-15 (UserMutationsTests.cs not created)
- **Gate:** `gates/10-persistence.md` Section 11 (Artifact Repository Persistence)

## Test Automation Hooks

```bash
# Future automation script
./tools/test-ba-001.sh

# Expected:
# 1. Spawn spawned agent with file creation task
# 2. Wait for completion
# 3. Verify file exists
# 4. Verify orchestrator ran verification
# 5. Pass/Fail based on criteria
```

---

**Next Test:** TC-BA-002 (Token Limit Detection)
