# TC-OR-001: Orchestrator Completion Verification Protocol

**Category:** Orchestrator
**Priority:** Critical
**Status:** Active
**Last Updated:** 2026-01-15

---

## Objective

Validate that orchestrators execute the complete verification protocol before declaring spawned agent work as "completed successfully" and detect false success scenarios.

## Background

**Production Failures (Harvana 2026-01-15):**

**Failure Pattern:**
1. Background agent spawned for implementation task
2. Agent hits token limit / permission failure / sandbox isolation
3. Agent produces no actual output (0 files created)
4. Orchestrator receives agent completion signal
5. Orchestrator declares "completed successfully" without verification
6. Reality: No work done, no files created, total failure

**Root Cause:**
- Orchestrator trusted agent completion signal without verification
- No check for error patterns in agent output
- No verification that claimed files actually exist
- No check that Write() tool calls were made

**Impact:**
- False progress reporting to user
- Hours of wasted work discovered later
- Build failures when files don't exist
- Loss of confidence in framework

## Prerequisites

- Project with spawned agent capability
- `.claude/settings.json` configured
- Test scenarios for both success and failure cases

## Test Scenario

### Scenario A: Agent Failure (Should Detect)

1. **Setup failed agent simulation:**
   ```bash
   mkdir -p .ai/tasks/2026-01-15_test-or-001-fail
   cp .ai-pack/templates/task-packet/* .ai/tasks/2026-01-15_test-or-001-fail/
   ```

2. **Create task requiring file creation:**
   ```markdown
   ## Requirements
   Create UserService.cs with CRUD operations

   ## Acceptance Criteria
   - [ ] File exists at src/services/UserService.cs
   - [ ] Contains Create, Read, Update, Delete methods
   - [ ] Has unit tests
   ```

3. **Orchestrator spawns agent (simulate failure scenario):**

   For this test, we simulate failure by:
   - Using verbose prompt (to hit token limit), OR
   - Removing Write permission (to cause permission failure), OR
   - Using incorrect working directory (to cause sandbox isolation)

   ```python
   Task(
     subagent_type="general-purpose",
     description="Create UserService (expect failure)",
     prompt="""[INTENTIONALLY VERBOSE PROMPT TO TRIGGER TOKEN LIMIT]

     Engineer role from .ai-pack/roles/engineer.md...
     [10,000 words of instructions]

     Create src/services/UserService.cs...
     """,
     
   )
   ```

4. **Wait for agent completion signal**

5. **CRITICAL: Orchestrator MUST execute 5-step verification protocol:**

   **Step 1: Check for Error Patterns (MANDATORY)**
   ```bash
   # Read full agent output
   AGENT_OUTPUT=$(cat <agent-output-file>)

   # Check for token limit
   if echo "$AGENT_OUTPUT" | grep -iE "exceeded.*token.*maximum|output.*limit"; then
     FAILURE_DETECTED="TOKEN_LIMIT"
   fi

   # Check for API errors
   if echo "$AGENT_OUTPUT" | grep -iE "API Error|rate limit|timeout"; then
     FAILURE_DETECTED="API_ERROR"
   fi

   # Check for permission errors
   if echo "$AGENT_OUTPUT" | grep -iE "permission denied|access denied|forbidden"; then
     FAILURE_DETECTED="PERMISSION_ERROR"
   fi

   # Check for tool failures
   if echo "$AGENT_OUTPUT" | grep -iE "tool.*failed|command.*failed"; then
     FAILURE_DETECTED="TOOL_FAILURE"
   fi

   if [ -n "$FAILURE_DETECTED" ]; then
     echo "⚠️ ERROR DETECTED: $FAILURE_DETECTED"
     echo "Status: ❌ FAILED - Verification blocked"
     exit 1
   fi
   ```

   **Step 2: Verify Write() Call Count**
   ```bash
   WRITE_COUNT=$(echo "$AGENT_OUTPUT" | grep -c "Write()")

   if [ $WRITE_COUNT -eq 0 ]; then
     echo "⚠️ WARNING: No Write() calls detected"
     echo "Agent may have failed before reaching implementation"
     SUSPICIOUS=true
   else
     echo "✓ Write() calls detected: $WRITE_COUNT"
   fi
   ```

   **Step 3: Extract Claimed Files**
   ```bash
   # Parse agent output for file claims
   CLAIMED_FILES=($(echo "$AGENT_OUTPUT" | \
     grep -E "Created|Wrote to|Generated" | \
     sed -E 's/.*Created:? (.*)$/\1/' | \
     sed -E 's/.*Wrote to:? (.*)$/\1/'))

   echo "Claimed files: ${#CLAIMED_FILES[@]}"
   for file in "${CLAIMED_FILES[@]}"; do
     echo "  - $file"
   done
   ```

   **Step 4: Verify Each File Exists**
   ```bash
   MISSING_FILES=()
   EMPTY_FILES=()

   for file in "${CLAIMED_FILES[@]}"; do
     if [ ! -f "$file" ]; then
       MISSING_FILES+=("$file")
       echo "❌ MISSING: $file"
     elif [ ! -s "$file" ]; then
       EMPTY_FILES+=("$file")
       echo "⚠️ EMPTY: $file"
     else
       SIZE=$(wc -c < "$file")
       echo "✓ EXISTS: $file ($SIZE bytes)"
     fi
   done

   if [ ${#MISSING_FILES[@]} -ne 0 ]; then
     echo ""
     echo "🚨 CRITICAL: File persistence failure"
     echo "Missing ${#MISSING_FILES[@]} of ${#CLAIMED_FILES[@]} files"
     VERIFICATION_FAILED=true
   fi

   if [ ${#EMPTY_FILES[@]} -ne 0 ]; then
     echo ""
     echo "⚠️ WARNING: Empty files detected"
     echo "Empty ${#EMPTY_FILES[@]} of ${#CLAIMED_FILES[@]} files"
   fi
   ```

   **Step 5: Report Verification Results**
   ```markdown
   ## Agent Completion Verification

   ### Step 1: Error Pattern Check
   Result: ⚠️ TOKEN LIMIT DETECTED
   Pattern: "exceeded output token maximum"

   ### Step 2: Write() Call Analysis
   Result: ❌ FAILED
   Write() calls: 0

   ### Step 3: Claimed Files
   Result: No files claimed
   Count: 0

   ### Step 4: File Existence
   Result: N/A (no files claimed)

   ### Step 5: Overall Status
   Status: ❌ FAILED - Agent hit token limit before implementation

   ### Root Cause Analysis
   - Verbose prompt consumed token budget
   - Agent never reached implementation phase
   - No Write() calls made
   - No files created

   ### Required Action
   1. Reduce prompt verbosity (<500 tokens)
   2. Re-spawn agent with concise instructions
   3. Monitor for completion
   4. Re-verify
   ```

6. **Orchestrator MUST NOT declare success:**
   ```
   ❌ WRONG: "Background Engineer completed successfully"
   ✅ CORRECT: "🚨 CRITICAL: Background agent failed - token limit exceeded"
   ```

### Scenario B: Agent Success (Should Pass Verification)

7. **Setup successful agent scenario:**
   ```bash
   mkdir -p .ai/tasks/2026-01-15_test-or-001-success
   cp .ai-pack/templates/task-packet/* .ai/tasks/2026-01-15_test-or-001-success/
   ```

8. **Orchestrator spawns agent with concise prompt:**
   ```python
   Task(
     subagent_type="general-purpose",
     description="Create UserService (expect success)",
     prompt="""Engineer role (.ai-pack/roles/engineer.md)

     Working directory: /Users/user/project
     Task: Create src/services/UserService.cs
     Task packet: .ai/tasks/2026-01-15_test-or-001-success/
     Follow TDD. Update work log.
     """,
     
   )
   ```

9. **Wait for agent completion**

10. **Orchestrator executes same 5-step verification:**
    - Step 1: No errors found ✓
    - Step 2: Write() calls > 0 ✓
    - Step 3: Files claimed: 2 (UserService.cs, UserServiceTests.cs) ✓
    - Step 4: Both files exist and not empty ✓
    - Step 5: Overall status: SUCCESS ✓

11. **Orchestrator CORRECTLY declares success:**
    ```markdown
    ## Agent Completion Verification

    ✅ Background Engineer completed successfully

    Artifact Persistence Verification:
      Expected files: 2
      Found files: 2
      Missing files: 0

      ✓ src/services/UserService.cs (2,048 bytes)
      ✓ src/tests/UserServiceTests.cs (1,536 bytes)

    Status: ✅ VERIFIED - Files persisted successfully
    ```

## Expected Behavior

### For Failed Agent (Scenario A):

**Orchestrator MUST:**
1. ✅ Read full agent output
2. ✅ Detect error pattern (token limit / permission / etc.)
3. ✅ Note 0 Write() calls
4. ✅ Note 0 files claimed
5. ✅ **DECLARE FAILURE** (not success)
6. ✅ Provide root cause analysis
7. ✅ Provide recovery guidance
8. ✅ **NEVER say "completed successfully"**

**Example Correct Report:**
```
🚨 CRITICAL: Background agent verification FAILED

Error Detected: Token limit exceeded
Write() calls: 0
Files claimed: 0
Files missing: N/A

Root Cause: Verbose prompt (12,450 tokens) exceeded budget

Status: ❌ FAILED - No work completed

Recovery:
1. Reduce prompt to <500 tokens
2. Re-spawn agent
3. Monitor completion
```

### For Successful Agent (Scenario B):

**Orchestrator MUST:**
1. ✅ Read full agent output
2. ✅ Find no error patterns
3. ✅ Confirm Write() calls > 0
4. ✅ Extract claimed files (2)
5. ✅ Verify both files exist
6. ✅ Verify both files not empty
7. ✅ **DECLARE SUCCESS** with verification details

**Example Correct Report:**
```
✅ Background Engineer completed successfully

Verification Steps:
  ✓ No errors detected
  ✓ Write() calls: 2
  ✓ Files claimed: 2
  ✓ Files exist: 2/2
  ✓ Files not empty: 2/2

Artifact Persistence Verification:
  ✓ src/services/UserService.cs (2,048 bytes)
  ✓ src/tests/UserServiceTests.cs (1,536 bytes)

Status: ✅ VERIFIED
```

## Actual Behavior (Execution Record)

**Test Run:** [Date]

**Scenario A (Failure Detection):**
- Agent failure type:
- Step 1 (error check) executed: [Yes/No]
- Step 2 (Write count) executed: [Yes/No]
- Step 3 (file claims) executed: [Yes/No]
- Step 4 (file exists) executed: [Yes/No]
- Step 5 (reporting) executed: [Yes/No]
- Orchestrator declared success: [Yes/No - MUST BE NO]
- Orchestrator declared failure: [Yes/No - MUST BE YES]

**Scenario B (Success Verification):**
- Agent completion: [Success/Fail]
- All 5 verification steps executed: [Yes/No]
- Files verified: [Count]
- Orchestrator declared success: [Yes/No - MUST BE YES]

**Deviations:**
[Any differences from expected]

## Pass/Fail Criteria

### PASS Criteria

**Scenario A (Failure Detection):**
✅ All 5 verification steps executed
✅ Error pattern detected correctly
✅ Orchestrator DECLARES FAILURE (not success)
✅ Root cause analysis provided
✅ Recovery guidance provided
✅ Never says "completed successfully"

**Scenario B (Success Verification):**
✅ All 5 verification steps executed
✅ No errors detected
✅ Files verified to exist
✅ Orchestrator declares success
✅ Verification details included in report

### FAIL Criteria

❌ Orchestrator skips any verification step
❌ Orchestrator declares success when agent failed
❌ Orchestrator doesn't detect error patterns
❌ Orchestrator ignores 0 Write() calls
❌ Orchestrator doesn't verify file existence
❌ No root cause analysis when failure detected
❌ No recovery guidance provided

## Known Issues

**Issue 1: Completion Signal ≠ Success**
- Agent completion just means "agent stopped"
- Doesn't mean "work completed successfully"
- Could be: token limit, error, timeout, permission failure
- **Mitigation:** MUST verify, not trust completion signal

**Issue 2: False Success Reporting**
- Old behavior: "Agent completed" → "Success!"
- New behavior: "Agent completed" → "Let me verify..."
- **Mitigation:** Mandatory 5-step verification protocol

**Issue 3: Silent Failures**
- Token limits don't throw errors, just truncate
- Permission failures may not be obvious
- Sandbox isolation has no error message
- **Mitigation:** Multiple verification dimensions (errors + Write calls + files)

## Verification Protocol Template

```bash
#!/bin/bash
# Orchestrator Agent Verification Protocol

AGENT_OUTPUT_FILE="$1"

# Step 1: Check for error patterns (BLOCKING)
echo "Step 1: Checking for error patterns..."
if grep -iE "exceeded.*token.*maximum|API Error|rate limit|permission denied|tool.*failed" "$AGENT_OUTPUT_FILE"; then
  echo "❌ ERROR DETECTED - Agent failed"
  exit 1
fi
echo "✓ No error patterns found"

# Step 2: Verify Write() calls
echo "Step 2: Checking Write() call count..."
WRITE_COUNT=$(grep -c "Write()" "$AGENT_OUTPUT_FILE")
if [ $WRITE_COUNT -eq 0 ]; then
  echo "⚠️ WARNING: No Write() calls (agent may have failed early)"
fi
echo "✓ Write() calls: $WRITE_COUNT"

# Step 3: Extract claimed files
echo "Step 3: Extracting claimed files..."
# [Implementation to parse agent output for file claims]

# Step 4: Verify files exist
echo "Step 4: Verifying file existence..."
# [Implementation to check each file]

# Step 5: Report results
echo "Step 5: Verification complete"
if [ all_checks_passed ]; then
  echo "✅ VERIFIED - Agent completed successfully"
else
  echo "❌ FAILED - Verification failed"
fi
```

## Recovery Procedures

### If Verification Fails:

1. **Analyze failure type:**
   - Token limit → Reduce prompt verbosity
   - Permission → Fix .claude/settings.json
   - Sandbox → Add working directory context
   - Tool failure → Check permissions, try foreground

2. **Do NOT mark work complete**
3. **Do NOT proceed to next phase**
4. **Fix root cause**
5. **Re-spawn agent**
6. **Re-verify**

### Never Skip Verification

```
❌ FORBIDDEN:
"Agent completed, marking task complete"

✅ REQUIRED:
"Agent completed, running verification...
 [5 verification steps]
 ...verification passed, marking complete"
```

## References

- **Commit:** `edf1d8a` - Lines 699-786 (Agent Completion Verification)
- **Orchestrator Role:** `roles/orchestrator.md` Section 2.15
- **Orchestrator Skill:** `templates/.claude/skills/orchestrator/SKILL.md`
- **Real Failures:** Harvana 2026-01-15 (Multiple false successes)

## Related Test Cases

- TC-BA-001 (File Persistence)
- TC-BA-002 (Token Limit Detection)
- TC-OR-002 (Artifact Persistence Enforcement)

## Metrics

**Before Fix:**
- False success rate: ~100% (all failures reported as success)
- User discovers failure: After builds break
- Wasted work: Hours per failed agent

**After Fix:**
- False success rate target: 0%
- Failure detection: Immediate (verification step)
- User notified: Before marking complete

---

**Version:** 1.0.0
**Last Reviewed:** 2026-01-15
