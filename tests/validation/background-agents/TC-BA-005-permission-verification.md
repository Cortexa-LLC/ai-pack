# TC-BA-005: Background Agent Permission Pre-Verification

**Category:** Background Agents
**Priority:** Critical
**Status:** Active
**Last Updated:** 2026-01-15

---

## Objective

Validate that orchestrators verify Write(*) permissions are configured BEFORE spawning background agents for artifact persistence tasks (Cartographer, Architect, Designer).

## Background

**Production Failure (consumer-project 2026-01-14):**
- Cartographer spawned to create PRD
- Agent generated 26KB of content (visible in agent logs)
- PRD never persisted to `docs/product/`
- Orchestrator continued as if successful
- **Root cause:** Write(*) permission not configured in `.claude/settings.json`
- **Impact:** Manual extraction from agent logs required

**Similar Failure (Architecture docs):**
- Architect generated 6 ADRs
- Referenced in architecture document
- ADRs never created (permission failure)
- Had to manually extract from agent output

**The Problem:**
Background agents (`run_in_background=true`) need Write(*) permission configured in `.claude/settings.json`. Without it:
- Write operations fail silently
- No error thrown
- Agent continues as if successful
- Files never persisted

## Prerequisites

- Project with `.claude/settings.json` support
- Ability to modify settings
- Background agent capability

## Test Scenario

### Setup Phase

1. **Create test task requiring artifact creation:**
   ```bash
   mkdir -p .ai/tasks/2026-01-15_test-ba-005
   cp .ai-pack/templates/task-packet/* .ai/tasks/2026-01-15_test-ba-005/
   ```

2. **Create two test scenarios:**
   - Scenario A: Missing Write permissions (should detect and block)
   - Scenario B: Correct Write permissions (should proceed)

### Scenario A: Missing Write Permissions (Should Block)

3. **Configure invalid permissions:**
   ```json
   {
     "permissions": {
       "allow": [
         "Read(*)",
         "Edit(*)",
         "Bash(git:*)"
         // ❌ Missing: "Write(*)"
       ],
       "defaultMode": "bypassPermissions"
     }
   }
   ```

4. **Orchestrator MUST verify permissions BEFORE spawning:**

   ```bash
   # Orchestrator verification script
   if [ ! -f .claude/settings.json ]; then
     echo "❌ ERROR: .claude/settings.json not found"
     exit 1
   fi

   if ! grep -q '"Write(\*)"' .claude/settings.json; then
     echo "❌ ERROR: Write(*) permission not configured"
     exit 1
   fi

   if ! grep -q '"bypassPermissions"' .claude/settings.json; then
     echo "❌ ERROR: defaultMode not set to bypassPermissions"
     exit 1
   fi

   echo "✅ Permissions verified"
   ```

5. **Expected: Orchestrator BLOCKS spawning:**

   ```
   ⚠️  CRITICAL: Background agent permissions not configured!

   Background agents need Write(*) permission to persist artifacts.

   Current configuration missing:
   - Write(*) in permissions.allow list

   SOLUTIONS:

   Option 1: Use .ai-pack setup (Recommended)
     python3 .ai-pack/templates/.claude-setup.py

   Option 2: Manual setup
     Copy template: cp .ai-pack/templates/.claude/settings.json .claude/
     Verify: grep 'Write(*)' .claude/settings.json

   Option 3: Run planning agents in FOREGROUND
     Remove run_in_background=true from Task calls
     Agents will prompt for Write permission interactively

   Cannot proceed with background planning agents until fixed.
   ```

6. **User fixes configuration:**
   ```json
   {
     "permissions": {
       "allow": [
         "Read(*)",
         "Write(*)",  // ✅ Added
         "Edit(*)",
         "Bash(git:*)"
       ],
       "defaultMode": "bypassPermissions"
     }
   }
   ```

### Scenario B: Correct Write Permissions (Should Proceed)

7. **Orchestrator verifies permissions:**
   ```bash
   # Check 1: settings.json exists
   [ -f .claude/settings.json ] && echo "✅ settings.json exists"

   # Check 2: Write(*) in allow list
   grep -q '"Write(\*)"' .claude/settings.json && echo "✅ Write(*) configured"

   # Check 3: defaultMode correct
   grep -q '"bypassPermissions"' .claude/settings.json && echo "✅ defaultMode correct"

   echo "✅ All permission checks passed"
   ```

8. **Orchestrator proceeds to spawn agent:**
   ```python
   # Permissions verified, safe to spawn background agent
   Task(
     subagent_type="general-purpose",
     description="Create PRD",
     prompt="""Cartographer role from .ai-pack/roles/cartographer.md

     Working directory: /Users/user/project

     Task: Create Product Requirements Document

     Persist to: docs/product/2026-01-15-feature-name/prd.md

     Task packet: .ai/tasks/2026-01-15_test-ba-005/
     """,
     run_in_background=true
   )
   ```

9. **Agent creates files successfully:**
   ```bash
   # Agent has Write permission
   # Successfully writes PRD
   Write("docs/product/2026-01-15-feature-name/prd.md", content)

   echo "✅ Created docs/product/2026-01-15-feature-name/prd.md"
   ```

10. **Verification confirms persistence:**
    ```bash
    ls docs/product/2026-01-15-feature-name/prd.md
    # ✅ File exists

    wc -c docs/product/2026-01-15-feature-name/prd.md
    # ✅ 26,854 bytes (PRD persisted successfully)
    ```

## Expected Behavior

### Scenario A: Missing Permissions

**Orchestrator Pre-Flight Check:**
```
Checking background agent permissions...

❌ ERROR: Write(*) not configured

BLOCKING: Cannot spawn background agents for artifact persistence
without Write(*) permission.

Please fix configuration using one of the provided solutions.
```

**Status:**
```
✅ Orchestrator detected missing permission
✅ Orchestrator BLOCKED spawning agent
✅ Orchestrator provided solution paths
✅ No silent failure
```

### Scenario B: Correct Permissions

**Orchestrator Pre-Flight Check:**
```
Checking background agent permissions...

✅ .claude/settings.json exists
✅ Write(*) in allow list
✅ defaultMode: bypassPermissions

Permissions verified. Proceeding to spawn background agent.
```

**Agent Execution:**
```
Cartographer generating PRD...

✅ Created docs/product/2026-01-15-feature-name/prd.md (26KB)

Work complete.
```

**Verification:**
```
✅ File persisted to repository
✅ Content complete (26KB)
✅ No manual extraction needed
```

## Actual Behavior (Execution Record)

**Test Run:** [Date]

**Scenario A (Missing Permissions):**
- Pre-flight check executed: [Yes/No]
- Missing permission detected: [Yes/No]
- Orchestrator blocked spawning: [Yes/No]
- Error message shown: [Yes/No]
- Solution paths provided: [Yes/No]

**Scenario B (Correct Permissions):**
- Pre-flight check executed: [Yes/No]
- All checks passed: [Yes/No]
- Agent spawned: [Yes/No]
- File persisted: [Yes/No]
- File size: [Bytes]

**Deviations:**
[Any differences]

## Pass/Fail Criteria

### PASS Criteria

**Scenario A (Missing Permissions):**
✅ Orchestrator runs pre-flight check
✅ Orchestrator detects missing Write(*)
✅ Orchestrator BLOCKS agent spawning
✅ Clear error message displayed
✅ Solution paths provided (3 options)
✅ User can fix and retry

**Scenario B (Correct Permissions):**
✅ Orchestrator runs pre-flight check
✅ All checks pass
✅ Orchestrator spawns agent
✅ Agent persists files successfully
✅ No manual extraction needed

### FAIL Criteria

❌ Orchestrator skips pre-flight check
❌ Orchestrator spawns agent without checking permissions
❌ Agent fails to persist but no error
❌ Silent failure with manual extraction required
❌ No guidance on fixing permissions

## Known Issues

**Issue 1: settings.local.json Override**
- May override settings.json
- **Detection:** Check both files
- **Fix:** Add Write(*) to settings.local.json or remove it

**Issue 2: Wrong defaultMode**
```json
// ❌ WRONG - Doesn't work for background agents
"defaultMode": "acceptEdits"

// ✅ CORRECT
"defaultMode": "bypassPermissions"
```

**Issue 3: VSCode Settings May Also Be Required**
- Some environments need VSCode-level permissions
- **Setting:** `claudeCode.allowDangerouslySkipPermissions: true`
- **See:** Commit 59f9864 for details

## Pre-Flight Check Procedure

**MANDATORY before spawning background planning agents:**

```bash
#!/bin/bash
# Permission Pre-Flight Check

echo "Checking background agent permissions..."

# Check 1: settings.json exists
if [ ! -f .claude/settings.json ]; then
  echo "❌ ERROR: .claude/settings.json not found"
  echo "Run: python3 .ai-pack/templates/.claude-setup.py"
  exit 1
fi

# Check 2: Write(*) in allow list
if ! grep -q '"Write(\*)"' .claude/settings.json; then
  echo "❌ ERROR: Write(*) not in permissions.allow"
  echo "Add: \"Write(*)\" to permissions.allow array"
  exit 1
fi

# Check 3: Edit(*) in allow list
if ! grep -q '"Edit(\*)"' .claude/settings.json; then
  echo "❌ ERROR: Edit(*) not in permissions.allow"
  echo "Add: \"Edit(*)\" to permissions.allow array"
  exit 1
fi

# Check 4: defaultMode correct
if ! grep -q '"bypassPermissions"' .claude/settings.json; then
  echo "❌ WARNING: defaultMode not set to bypassPermissions"
  echo "Set: \"defaultMode\": \"bypassPermissions\""
fi

echo "✅ All permission checks passed"
exit 0
```

## Required Configuration

**Minimum .claude/settings.json:**

```json
{
  "permissions": {
    "allow": [
      "Write(*)",
      "Edit(*)",
      "Read(*)",
      "Bash(git:*)",
      "Bash(npm:install)",
      "Bash(npm:test)",
      "Bash(dotnet:*)"
    ],
    "defaultMode": "bypassPermissions"
  }
}
```

## Alternative: Foreground Agents

**If permissions cannot be configured:**

```python
# Instead of background
Task(..., run_in_background=true)  # ❌ Needs Write(*) permission

# Use foreground (interactive)
Task(..., run_in_background=false)  # ✅ Can prompt for permissions
```

**Trade-offs:**
- Background: Parallel, autonomous, **requires Write(*)**
- Foreground: Sequential, interactive, **no config needed**

## Recovery from Permission Failures

**If agent completed but files missing:**

1. **Check agent output:**
   ```bash
   grep "Write(" <agent-output-file>
   # If no Write() calls, permission likely blocked
   ```

2. **Look for permission errors:**
   ```bash
   grep -i "permission denied\|access denied\|forbidden" <agent-output>
   ```

3. **Extract content from output:**
   - If agent output contains full file content
   - Manually create files
   - Verify correctness

4. **Fix permissions and retry:**
   - Add Write(*) to .claude/settings.json
   - Re-spawn agent
   - Verify files persist

## Metrics

**Before Permission Pre-Flight Check:**
- Silent permission failures: ~40% of background agents
- Manual extraction required: ~40%
- Time to discover failure: 10-30 minutes

**After Permission Pre-Flight Check:**
- Silent permission failures: 0% (blocked before spawning)
- Manual extraction required: 0%
- Time to discover misconfiguration: Immediate

## References

- **Commit:** `a944a7c` - Add mandatory background agent permission verification
- **Commit:** `59f9864` - Document VSCode setting requirement for background worker permissions
- **Orchestrator Role:** Section 2.14 (Background Agent Permission Verification)
- **PERMISSIONS.md:** templates/.claude/PERMISSIONS.md
- **settings.json template:** templates/.claude/settings.json
- **Real Failures:** consumer-project PRD (26KB not persisted), 6 ADRs not created

---

**Related Tests:**
- TC-BA-001 (File Persistence)
- TC-BA-003 (Working Directory Context)
