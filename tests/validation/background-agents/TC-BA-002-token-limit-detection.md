# TC-BA-002: Token Limit Detection and Error Reporting

**Category:** Background Agents
**Priority:** Critical
**Status:** Active
**Last Updated:** 2026-01-15

---

## Objective

Validate that orchestrators detect when background agents hit token limits and correctly report this as a failure rather than success.

## Background

**Production Failure (Harvana 2026-01-15):**
- Agent received verbose prompt with extensive instructions
- Agent generated long planning output
- Agent hit 32K output token limit during planning phase
- Agent made 0 Write() calls (never reached implementation)
- Orchestrator declared "completed successfully" despite total failure
- No error detection performed

**Root Cause:**
1. Verbose instructions consumed ~20K tokens before agent started work
2. Agent planning consumed remaining tokens
3. Token limit hit = silent failure (no error thrown)
4. Orchestrator didn't check for error patterns

**Impact:**
- Wasted hours of agent work
- False progress reporting
- No actual work completed
- Had to retry multiple times

## Prerequisites

- Project with complex task requiring substantial output
- Background agent with verbose prompt (intentionally)
- `.claude/settings.json` configured

## Test Scenario

### Setup Phase

1. Create test task packet for complex feature:
   ```bash
   mkdir -p .ai/tasks/2026-01-15_test-ba-002
   cp .ai-pack/templates/task-packet/* .ai/tasks/2026-01-15_test-ba-002/
   ```

2. Create intentionally verbose prompt (to trigger token limit):
   ```markdown
   ## Requirements
   Implement WunderGraph Cosmo Gateway integration (25 files)
   - Gateway configuration
   - Server setup
   - Schema definitions
   - Authentication
   - Deployment scripts
   ```

### Execution Phase - Verbose Prompt (Should Detect Failure)

3. **BAD EXAMPLE:** Orchestrator spawns agent with verbose prompt:
   ```python
   Task(
     subagent_type="general-purpose",
     description="Implement gateway (verbose - expect failure)",
     prompt="""You are the Engineer from .ai-pack/roles/engineer.md.

     TASK CONTEXT AND BACKGROUND:
     [3000 words of verbose context...]

     REQUIREMENTS IN DETAIL:
     [5000 words of detailed requirements...]

     TDD REMINDER (even though role already knows):
     [2000 words explaining TDD process...]

     ACCEPTANCE CRITERIA:
     [2000 words of acceptance criteria...]

     PLEASE FOLLOW THESE STEPS:
     [3000 words of step-by-step instructions...]

     Task packet: .ai/tasks/2026-01-15_test-ba-002/
     """,
     run_in_background=true
   )
   ```

4. Wait for agent completion (will hit token limit)

5. **CRITICAL: Orchestrator MUST detect token limit failure**

### Verification Phase (MANDATORY)

6. **Step 1: Check for error patterns** (THIS SHOULD CATCH TOKEN LIMIT)
   ```bash
   # Read agent output
   tail -1000 <agent-output-file>

   # Search for token limit pattern
   grep -iE "exceeded.*token.*maximum|output.*limit|token.*limit.*exceeded" <agent-output>

   # Expected: MATCH FOUND
   if grep -iE "exceeded.*token.*maximum" <agent-output>; then
     echo "⚠️ TOKEN LIMIT DETECTED"
     FAILURE_DETECTED=true
   fi
   ```

7. **Step 2: Verify Write() call count**
   ```bash
   WRITE_COUNT=$(grep -c "Write()" <agent-output>)
   echo "Write() calls: $WRITE_COUNT"

   if [ $WRITE_COUNT -eq 0 ]; then
     echo "⚠️ NO WRITE CALLS - Possible token limit or planning-only agent"
     SUSPICIOUS=true
   fi
   ```

8. **Step 3: Check for claimed files**
   ```bash
   CLAIMED_FILES=$(grep -E "Created.*:|Wrote.*:" <agent-output>)
   if [ -z "$CLAIMED_FILES" ]; then
     echo "⚠️ NO FILES CLAIMED - Agent likely hit token limit before implementation"
     FAILURE_DETECTED=true
   fi
   ```

9. **Step 4: Orchestrator MUST report token limit as FAILURE**
   ```markdown
   ## Agent Completion Analysis

   🚨 CRITICAL: Token limit failure detected

   Error Pattern Found:
     "exceeded output token maximum"

   Diagnostic Evidence:
     - Write() calls: 0
     - Files claimed: 0
     - Agent output truncated at 32K tokens

   Root Cause: Verbose prompt consumed token budget before implementation

   Status: ❌ FAILED - Agent hit token limit, no work completed

   Required Action:
     1. Reduce prompt verbosity (see concise instruction guidelines)
     2. Decompose task into smaller chunks (see task size guidelines)
     3. Re-spawn agent with concise instructions
   ```

### Execution Phase - Concise Prompt (Should Succeed)

10. **GOOD EXAMPLE:** Orchestrator re-spawns with concise prompt:
    ```python
    Task(
      subagent_type="general-purpose",
      description="Implement gateway (concise - expect success)",
      prompt="""Engineer role (.ai-pack/roles/engineer.md)

      Working directory: /path/to/repo
      Task: Implement gateway foundation (4 files)
      Task packet: .ai/tasks/2026-01-15_test-ba-002/
      Follow TDD. Update work log.
      """,
      run_in_background=true
    )
    ```

11. Wait for completion

12. Run verification protocol again

13. **Expected:** No token limit, Write() calls > 0, files created

## Expected Behavior

### Scenario 1: Verbose Prompt (Token Limit)

**Agent Behavior:**
- Receives verbose prompt
- Generates extensive planning output
- Hits 32K token limit during planning
- Never reaches implementation phase
- No Write() calls made
- Output truncated

**Orchestrator Detection (CRITICAL):**
1. ✅ Reads agent output
2. ✅ Finds error pattern: "exceeded.*token.*maximum"
3. ✅ Notes Write() call count: 0
4. ✅ Notes files claimed: 0
5. ✅ **DECLARES FAILURE** (not success)

**Orchestrator Report:**
```
🚨 CRITICAL: Token limit failure detected

Agent hit 32K output token limit before completing work.

Evidence:
  - Error: "exceeded output token maximum"
  - Write() calls: 0
  - Files created: 0

Status: ❌ FAILED

Recommendation: Reduce prompt verbosity and re-spawn
```

### Scenario 2: Concise Prompt (Success)

**Agent Behavior:**
- Receives concise prompt (~500 tokens)
- Minimal planning output
- Proceeds to implementation
- Makes Write() calls
- Creates files
- Completes successfully

**Orchestrator Detection:**
1. ✅ Reads agent output
2. ✅ No error patterns found
3. ✅ Write() calls: >0
4. ✅ Files claimed: Yes
5. ✅ Files exist: Yes
6. ✅ **DECLARES SUCCESS**

**Orchestrator Report:**
```
✅ Background Engineer completed successfully

Artifact Persistence Verification:
  Expected files: 4
  Found files: 4
  Missing files: 0

Status: ✅ VERIFIED
```

## Actual Behavior (Execution Record)

**Test Run:** [Date]

**Scenario 1 (Verbose Prompt):**
- Agent output length:
- Token limit error found: [Yes/No]
- Write() call count:
- Orchestrator detected failure: [Yes/No]
- Orchestrator reported correctly: [Yes/No]

**Scenario 2 (Concise Prompt):**
- Agent output length:
- Token limit error found: [Yes/No]
- Write() call count:
- Files created:
- Orchestrator reported success: [Yes/No]

**Deviations:**
[Any differences from expected behavior]

## Pass/Fail Criteria

### PASS Criteria

**Scenario 1 (Verbose - Should Fail):**
✅ Agent hits token limit
✅ Orchestrator detects error pattern
✅ Orchestrator notes 0 Write() calls
✅ Orchestrator DECLARES FAILURE (not success)
✅ Orchestrator provides recovery guidance

**Scenario 2 (Concise - Should Succeed):**
✅ Agent completes without token limit
✅ Agent makes Write() calls
✅ Files created successfully
✅ Orchestrator verifies files
✅ Orchestrator declares success

### FAIL Criteria

❌ Orchestrator declares success when agent hit token limit
❌ Orchestrator skips error pattern check
❌ Orchestrator ignores 0 Write() calls
❌ Orchestrator doesn't distinguish between success and failure
❌ No recovery guidance provided

## Known Issues

**Issue 1: Token Limit = Silent Failure**
- Token limit doesn't throw error in agent output
- Just truncates at 32K tokens
- Easy to miss without explicit check
- **Mitigation:** Mandatory error pattern check (Section 2.15 orchestrator.md)

**Issue 2: Verbose Prompts**
- Old pattern: Include full context in prompt
- Consumed 10-20K tokens before work started
- **Mitigation:** Concise instruction guidelines (Orchestrator skill lines 232-304)

**Issue 3: Large Tasks**
- 25-file tasks exceed token budget even with concise prompts
- **Mitigation:** Task size guidelines + decomposition (Orchestrator skill lines 306-488)

## Error Patterns to Detect

Required regex patterns for orchestrator verification:

```bash
# Token limits
"exceeded.*token.*maximum"
"output.*limit.*reached"
"token.*limit.*exceeded"
"maximum.*tokens.*exceeded"

# API errors
"API Error"
"rate limit"
"timeout"

# Permission errors
"permission denied"
"access denied"
"forbidden"

# Tool failures
"tool.*failed"
"command.*failed"
```

## Recovery Procedures

### If Token Limit Detected:

1. **Analyze prompt verbosity**
   - Count tokens in instruction section
   - Target: <500 tokens for instructions

2. **Apply concise instruction guidelines**
   - Reference task packets, don't repeat
   - Trust role knowledge
   - Bullet points not paragraphs
   - Remove meta-commentary

3. **Check task size**
   - If >10 files, decompose into smaller tasks
   - Each subtask: 3-8 files max

4. **Re-spawn with corrected prompt**

### Concise Instruction Template:

```python
"""Engineer role (.ai-pack/roles/engineer.md)

Working directory: {repo_root}
Task: {brief_description}
Task packet: .ai/tasks/{task_id}/
Follow TDD. Update work log.
"""
```

## References

- **Commit:** `edf1d8a` - CRITICAL: Fix token limit failures + false success reporting
- **Commit:** `e1764ec` - Add task decomposition guidance to prevent token limit failures
- **Orchestrator Skill:** `templates/.claude/skills/orchestrator/SKILL.md` Lines 232-304 (Concise Instructions)
- **Orchestrator Skill:** Lines 306-488 (Task Decomposition)
- **Orchestrator Skill:** Lines 699-786 (Agent Completion Verification)
- **Real Failure:** Harvana 2026-01-15 (WunderGraph gateway - 5 failed attempts)

## Metrics

**Before Fix (Harvana):**
- Attempts to complete 25-file task: 5
- Successes: 0
- Token limit failures: 5
- Failure detection rate: 0% (all marked "success")

**After Fix (Expected):**
- Token limit detection rate: 100%
- False success rate: 0%
- Recovery guidance provided: Always

## Test Automation Hooks

```bash
# Future automation
./tools/test-ba-002.sh

# Test flow:
# 1. Spawn agent with intentionally verbose prompt
# 2. Wait for token limit
# 3. Verify orchestrator detects failure
# 4. Verify orchestrator doesn't declare success
# 5. Spawn agent with concise prompt
# 6. Verify success this time
# 7. Compare results
```

---

**Related Tests:**
- TC-BA-001 (File Persistence)
- TC-OR-005 (Task Decomposition)
