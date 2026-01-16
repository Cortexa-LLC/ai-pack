# TC-INT-001: Background Agent File Persistence Integration Test

**Category:** Integration
**Priority:** Critical
**Status:** Active
**Last Updated:** 2026-01-15

---

## Objective

Validate that background agents can actually create files in the repository when Write(*) permissions are configured correctly, using REAL background agent execution (not simulation).

**This is an INTEGRATION TEST:** Spawns actual background agents and verifies physical file creation.

## Background

**Problem:** Static permission checks aren't enough - need to verify permissions ACTUALLY work.

**User Request:**
> "Our test case should also validate that beyond just permissions in .claude/settings.json that it actually does work as expected with a validation test physically running in the background and generating files in a temporary folder in the repository"

**What This Tests:**
- ✅ Permissions configured correctly (.claude/settings.json)
- ✅ Background agent spawns successfully
- ✅ Agent receives correct working directory context
- ✅ Agent writes files to repository (not sandbox)
- ✅ Files persist after agent completes
- ✅ Files are in correct location
- ✅ End-to-end flow works

---

## Prerequisites

- Project with ai-pack framework
- .claude/settings.json configured with Write(*) permission
- Ability to spawn background agents
- Git repository initialized

---

## Test Scenario

### Setup Phase

1. **Verify permissions are configured:**
   ```bash
   # Run permission verification
   python3 tests/tools/verify-background-agent-permissions.py

   # Must see:
   # ✅ ALL CHECKS PASSED
   # Safe to spawn background agents!
   ```

2. **Create test directory:**
   ```bash
   # Create temporary test location IN REPOSITORY
   TIMESTAMP=$(date +%s)
   TEST_DIR=".ai/test-artifacts/tc-int-001-$TIMESTAMP"
   mkdir -p "$TEST_DIR"

   echo "Test directory created: $TEST_DIR"
   ```

3. **Capture repository root:**
   ```bash
   REPO_ROOT=$(git rev-parse --show-toplevel)
   echo "Repository root: $REPO_ROOT"
   ```

---

### Execution Phase: Spawn REAL Background Agent

4. **Define expected outputs:**
   ```bash
   # Files the agent should create
   EXPECTED_FILE_1="$TEST_DIR/test-file-1.md"
   EXPECTED_FILE_2="$TEST_DIR/test-file-2.txt"
   EXPECTED_FILE_3="$TEST_DIR/subdirectory/test-file-3.json"

   echo "Expected files:"
   echo "  - $EXPECTED_FILE_1"
   echo "  - $EXPECTED_FILE_2"
   echo "  - $EXPECTED_FILE_3"
   ```

5. **Spawn ACTUAL background agent:**
   ```python
   #!/usr/bin/env python3
   """
   This spawns a REAL background agent using Claude Code Task tool.
   Not a simulation - actual agent execution.
   """

   import time
   from pathlib import Path

   # Get paths
   repo_root = Path.cwd()
   test_dir = repo_root / f".ai/test-artifacts/tc-int-001-{int(time.time())}"
   test_dir.mkdir(parents=True, exist_ok=True)

   # Define expected files
   file_1 = test_dir / "test-file-1.md"
   file_2 = test_dir / "test-file-2.txt"
   file_3 = test_dir / "subdirectory" / "test-file-3.json"

   # CRITICAL: Spawn REAL background agent
   print("Spawning REAL background agent...")
   print(f"Test directory: {test_dir}")

   Task(
       subagent_type="general-purpose",
       description="TC-INT-001: Test background file creation",
       prompt=f"""
INTEGRATION TEST: TC-INT-001 Background Agent File Persistence

You are a background agent testing file creation capabilities.

CRITICAL CONTEXT:
- Repository root: {repo_root}
- Test directory: {test_dir}
- Current working directory may differ from repository root
- MUST use absolute paths to ensure files go to repository

TASK: Create 3 test files to verify background agent file persistence

FILE 1: {file_1}
Content:
```markdown
# TC-INT-001 Test File 1

**Purpose:** Verify background agent can create markdown files

**Timestamp:** {time.time()}
**Test:** Integration test for background agent permissions
**Agent Type:** Background worker
**Working Directory:** $(pwd when this file was created)

This file proves:
- Background agent has Write(*) permission
- Files persist to repository (not sandbox)
- Absolute paths work correctly
```

FILE 2: {file_2}
Content:
```
TC-INT-001 Test File 2
Plain text file to verify basic file creation
Timestamp: {time.time()}
```

FILE 3: {file_3}
Content:
```json
{{
  "test": "TC-INT-001",
  "purpose": "Verify nested directory creation",
  "timestamp": {time.time()},
  "agent": "background",
  "success": true
}}
```

MANDATORY PROCEDURE:
1. Verify current working directory: pwd
2. Echo repository root: echo "Repository root: {repo_root}"
3. Create subdirectory: mkdir -p {test_dir / "subdirectory"}
4. Create File 1 using Write() tool with ABSOLUTE path: {file_1}
5. Create File 2 using Write() tool with ABSOLUTE path: {file_2}
6. Create File 3 using Write() tool with ABSOLUTE path: {file_3}
7. Verify all files exist:
   - ls -la {file_1}
   - ls -la {file_2}
   - ls -la {file_3}
8. Report file sizes:
   - wc -c {file_1}
   - wc -c {file_2}
   - wc -c {file_3}

CRITICAL: Use ABSOLUTE paths for all Write() calls.
DO NOT use relative paths - they may write to sandbox.

EXPECTED OUTCOME:
All 3 files created in repository at specified absolute paths.
       """,
       run_in_background=True  # ACTUAL background execution
   )

   print("Background agent spawned!")
   print("Agent is now running in background...")
   print("Waiting for completion...")
   ```

6. **Poll for agent completion:**
   ```python
   # Wait for files to appear (indicates agent completed)
   import time

   max_wait = 180  # 3 minutes max
   poll_interval = 5  # Check every 5 seconds
   files_created = []

   print("\nPolling for file creation...")

   for elapsed in range(0, max_wait, poll_interval):
       # Check if all files exist
       all_exist = all([
           file_1.exists(),
           file_2.exists(),
           file_3.exists()
       ])

       if all_exist:
           print(f"\n✅ All files detected after {elapsed}s")
           files_created = [file_1, file_2, file_3]
           break

       # Report progress
       existing = sum([file_1.exists(), file_2.exists(), file_3.exists()])
       print(f"  [{elapsed}s] Files created: {existing}/3", end='\r')

       time.sleep(poll_interval)
   else:
       # Timeout
       print(f"\n❌ TIMEOUT: Files not created within {max_wait}s")
       print("\nDiagnostics:")
       print(f"  File 1 exists: {file_1.exists()}")
       print(f"  File 2 exists: {file_2.exists()}")
       print(f"  File 3 exists: {file_3.exists()}")
       raise TimeoutError("Background agent did not complete in time")
   ```

---

### Verification Phase: Confirm ACTUAL File Persistence

7. **Verify File 1 (Markdown):**
   ```bash
   FILE_1="$TEST_DIR/test-file-1.md"

   # Check exists
   if [ ! -f "$FILE_1" ]; then
       echo "❌ FAIL: File 1 not found: $FILE_1"
       exit 1
   fi
   echo "✅ File 1 exists"

   # Check size
   SIZE=$(wc -c < "$FILE_1")
   if [ "$SIZE" -lt 100 ]; then
       echo "❌ FAIL: File 1 too small ($SIZE bytes)"
       exit 1
   fi
   echo "✅ File 1 has content ($SIZE bytes)"

   # Check content
   if ! grep -q "TC-INT-001" "$FILE_1"; then
       echo "❌ FAIL: File 1 missing expected content"
       exit 1
   fi
   echo "✅ File 1 content correct"
   ```

8. **Verify File 2 (Plain text):**
   ```bash
   FILE_2="$TEST_DIR/test-file-2.txt"

   if [ ! -f "$FILE_2" ]; then
       echo "❌ FAIL: File 2 not found"
       exit 1
   fi
   echo "✅ File 2 exists"

   if ! grep -q "Plain text file" "$FILE_2"; then
       echo "❌ FAIL: File 2 content incorrect"
       exit 1
   fi
   echo "✅ File 2 content correct"
   ```

9. **Verify File 3 (JSON in subdirectory):**
   ```bash
   FILE_3="$TEST_DIR/subdirectory/test-file-3.json"

   if [ ! -f "$FILE_3" ]; then
       echo "❌ FAIL: File 3 not found (subdirectory creation failed)"
       exit 1
   fi
   echo "✅ File 3 exists in subdirectory"

   # Validate JSON
   if ! python3 -m json.tool "$FILE_3" > /dev/null 2>&1; then
       echo "❌ FAIL: File 3 is not valid JSON"
       exit 1
   fi
   echo "✅ File 3 is valid JSON"
   ```

10. **Verify files are IN REPOSITORY (not sandbox):**
    ```bash
    # Get repository root
    REPO_ROOT=$(git rev-parse --show-toplevel)

    # Check each file is within repository
    for FILE in "$FILE_1" "$FILE_2" "$FILE_3"; do
        # Get absolute path
        ABS_FILE=$(cd "$(dirname "$FILE")" && pwd)/$(basename "$FILE")

        # Check if within repository
        if [[ "$ABS_FILE" != "$REPO_ROOT"* ]]; then
            echo "❌ FAIL: File is OUTSIDE repository"
            echo "   Repository: $REPO_ROOT"
            echo "   File: $ABS_FILE"
            echo "   This indicates sandbox isolation issue!"
            exit 1
        fi
    done

    echo "✅ All files are within repository"
    ```

11. **Verify working directory context was correct:**
    ```bash
    # Check if File 1 contains working directory info
    if grep -q "Working Directory:" "$FILE_1"; then
        AGENT_WD=$(grep "Working Directory:" "$FILE_1" | cut -d: -f2 | tr -d ' ')
        echo "Agent working directory was: $AGENT_WD"

        # Verify it's the repository root
        if [ "$AGENT_WD" != "$REPO_ROOT" ]; then
            echo "⚠️  WARNING: Agent working directory differs from repo root"
            echo "   This is OK if absolute paths were used correctly"
        fi
    fi
    ```

12. **Verify no files in common sandbox locations:**
    ```bash
    # Check common sandbox locations
    SANDBOX_PATTERNS=(
        "/tmp/claude-*"
        "/var/tmp/claude-*"
        "$HOME/.claude/temp/*"
    )

    echo "Checking for files in sandbox locations..."
    for PATTERN in "${SANDBOX_PATTERNS[@]}"; do
        if find $PATTERN -name "test-file-*.md" 2>/dev/null | grep -q .; then
            echo "⚠️  WARNING: Test files found in sandbox: $PATTERN"
            echo "   Files should only be in repository!"
        fi
    done
    echo "✅ No test files in common sandbox locations"
    ```

---

### Cleanup Phase

13. **Remove test artifacts:**
    ```bash
    # Clean up test directory
    rm -rf "$TEST_DIR"
    echo "✅ Test artifacts removed"

    # Verify cleanup
    if [ -d "$TEST_DIR" ]; then
        echo "⚠️  WARNING: Test directory still exists"
    else
        echo "✅ Test directory cleaned up successfully"
    fi
    ```

---

## Expected Behavior

**Agent Spawn:**
```
✅ Background agent spawns successfully
✅ Agent receives task prompt
✅ Agent begins execution in background
✅ No errors during spawn
```

**File Creation:**
```
✅ File 1 created within 30s
✅ File 2 created within 30s
✅ File 3 created within 30s (with subdirectory)
✅ All files have correct content
✅ All files have reasonable size (>100 bytes each)
```

**Location Verification:**
```
✅ All files in repository (not sandbox)
✅ Absolute paths resolve to repository locations
✅ Subdirectory created correctly
✅ No files in /tmp or other sandbox locations
```

**Permissions:**
```
✅ Write(*) permission used successfully
✅ No permission errors in agent output
✅ Files writable by repository owner
```

---

## Actual Behavior (Execution Record)

**Test Run:** [Date]

**Setup:**
- Repository root: [Path]
- Test directory: [Path]
- Permissions verified: [Yes/No]

**Agent Spawn:**
- Spawn successful: [Yes/No]
- Spawn time: [Seconds]
- Errors during spawn: [None/List]

**File Creation:**
- File 1 created: [Yes/No] - Time: [Seconds]
- File 2 created: [Yes/No] - Time: [Seconds]
- File 3 created: [Yes/No] - Time: [Seconds]
- Total completion time: [Seconds]

**Verification:**
- File 1 size: [Bytes]
- File 2 size: [Bytes]
- File 3 size: [Bytes]
- All in repository: [Yes/No]
- Content correct: [Yes/No]
- JSON valid: [Yes/No]

**Issues Found:**
[List any issues, errors, or unexpected behavior]

**Deviations from Expected:**
[Any differences from expected behavior]

---

## Pass/Fail Criteria

### PASS Criteria

**Permissions:**
✅ .claude/settings.json configured correctly
✅ Write(*) in permissions.allow
✅ defaultMode: bypassPermissions
✅ Permission verification passes

**Agent Execution:**
✅ Background agent spawns successfully
✅ Agent completes within 3 minutes
✅ No errors in agent execution
✅ Agent output shows successful file creation

**File Persistence:**
✅ All 3 files created
✅ Files in correct repository location
✅ Files have expected content
✅ Files have reasonable size
✅ JSON file is valid
✅ Subdirectory created correctly

**Location Verification:**
✅ Files within repository root
✅ Absolute paths used correctly
✅ NO files in sandbox locations (/tmp, etc.)
✅ Working directory context correct

**Cleanup:**
✅ Test artifacts removed successfully
✅ No leftover files

---

### FAIL Criteria

**Critical Failures:**
❌ Permission verification fails
❌ Background agent fails to spawn
❌ Agent times out (>3 minutes)
❌ Files NOT created
❌ Files in sandbox instead of repository
❌ Files have wrong content
❌ Files not readable

**Warning Failures:**
⚠️ Slow completion (>60s for all files)
⚠️ Files found in both repository AND sandbox
⚠️ Incorrect working directory
⚠️ Cleanup incomplete

---

## Troubleshooting

### Issue 1: Agent Spawns But Files Not Created

**Symptoms:**
- Agent spawns successfully
- No error messages
- Files never appear

**Diagnosis:**
```bash
# Check agent logs
# Look for Write() tool calls
grep "Write(" <agent-output-file>

# Check for permission errors
grep -i "permission\|denied\|forbidden" <agent-output-file>
```

**Likely Cause:** Permission failure (even though pre-check passed)

**Solution:**
```bash
# Verify settings.local.json doesn't override
cat .claude/settings.local.json

# Re-run permission verification
python3 tests/tools/verify-background-agent-permissions.py

# Try foreground agent
# Change run_in_background=False to see interactive errors
```

---

### Issue 2: Files Created in Sandbox Instead of Repository

**Symptoms:**
- Files created successfully
- But not in expected repository location
- Found in /tmp or similar

**Diagnosis:**
```bash
# Search for files
find /tmp -name "test-file-*.md"
find ~ -name "test-file-*.md"
```

**Likely Cause:** Working directory context not passed OR relative paths used

**Solution:**
```bash
# Ensure prompt includes:
# - Absolute paths for all files
# - Repository root context
# - Explicit mkdir with absolute path

# Example correct prompt:
REPO_ROOT=$(git rev-parse --show-toplevel)
Write("$REPO_ROOT/path/to/file.md", content)  # ← Absolute
```

---

### Issue 3: Agent Times Out

**Symptoms:**
- Agent spawns
- Runs for >3 minutes
- Never completes

**Diagnosis:**
```bash
# Check if agent is still running
ps aux | grep claude

# Check agent output so far
tail -f <agent-output-file>
```

**Likely Cause:** Complex task, network issues, or hung on permission prompt

**Solution:**
- Simplify task (fewer files)
- Check network connectivity
- Verify defaultMode: bypassPermissions
- Increase timeout if task is genuinely complex

---

### Issue 4: JSON File Invalid

**Symptoms:**
- File 3 created
- But fails JSON validation

**Diagnosis:**
```bash
# Validate JSON
python3 -m json.tool .ai/test-artifacts/*/subdirectory/test-file-3.json

# Check content
cat .ai/test-artifacts/*/subdirectory/test-file-3.json
```

**Likely Cause:** Agent generated invalid JSON or corrupted content

**Solution:**
- Check agent prompt has correct JSON template
- Verify no string interpolation issues
- Try re-running test

---

## Integration with Other Tests

**This test validates the END-TO-END flow of:**

1. **TC-BA-005:** Permission pre-verification (this test requires it)
2. **TC-BA-003:** Working directory context (verified by file location)
3. **TC-BA-001:** File persistence (core of this test)
4. **Gate 08:** Background agent permissions (enforced before spawn)

**Test Flow:**
```
1. Gate 08: Verify permissions ← Pre-requisite
   ↓ PASS
2. TC-INT-001: Spawn REAL agent ← This test
   ↓
3. Agent creates files with Write()
   ↓
4. TC-BA-001: Verify files exist ← Validation
   ↓
5. TC-BA-003: Verify location correct ← Validation
   ↓ PASS
6. Test complete
```

---

## Metrics

**Performance Benchmarks:**
```
Ideal Performance:
- Agent spawn: <5 seconds
- File creation (all 3): <30 seconds
- Total test time: <60 seconds

Acceptable Performance:
- Agent spawn: <10 seconds
- File creation: <90 seconds
- Total test time: <120 seconds

Unacceptable:
- Agent spawn: >15 seconds (investigate)
- File creation: >180 seconds (timeout)
```

**Reliability Targets:**
```
Pass Rate: >95% (allow for occasional network issues)
False Positives: 0% (never pass when should fail)
False Negatives: <5% (rare timeout on slow systems)
```

---

## Success Criteria

**This test is successful when:**

✅ Background agent spawns and completes reliably
✅ Files created in repository (never sandbox)
✅ Permissions work as configured
✅ Test completes in <2 minutes
✅ No manual intervention needed
✅ Test can run repeatedly without issues
✅ Cleanup leaves no artifacts

**Leading Indicators:**
- Permission verification always runs first
- Agent spawn never hangs
- Files appear quickly (<30s)
- Location always correct
- No permission errors in logs

---

## References

**Related Tests:**
- TC-BA-005: Permission Pre-Verification
- TC-BA-003: Working Directory Context
- TC-BA-001: File Persistence Verification

**Tools:**
- `tests/tools/verify-background-agent-permissions.py`

**Gates:**
- Gate 08: Background Agent Permissions

**Documentation:**
- `roles/orchestrator.md` - Section 2.14
- `templates/.claude/settings.json`

**Production Context:**
- Based on consumer-project failures where files weren't persisted
- User request for physical validation test

---

**Version:** 1.0.0
**Type:** Integration Test
**Execution:** Manual (requires human to spawn agent)
**Duration:** ~2 minutes
**Frequency:** Before major releases, after permission changes
