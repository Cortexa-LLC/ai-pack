# TC-BA-004: Absolute Path Requirements and Nested Directory Prevention

**Category:** Spawned Agents
**Priority:** Critical
**Status:** Active
**Last Updated:** 2026-01-15

---

## Objective

Validate that agents verify working directory before file operations and use absolute paths to prevent nested directory disasters like `server/server/API/`.

## Background

**Production Failure (Harvana 2026-01-14):**
- Agent created: `server/server/API/` (nested directory)
- Agent created: `server/server/Tests/` (nested directory)
- Agent created: `server/server/Models/` (nested directory)
- **Root cause:** Agent used relative paths while not in expected working directory
- **Impact:** Project structure corruption, significant debugging time, manual cleanup

**The Problem:**
```bash
# Agent's current directory (unknown to agent)
pwd
# /Users/user/harvana/server/

# Agent executes (thinking it's in project root)
mkdir server/API

# Result
# /Users/user/harvana/server/server/API/  ← DISASTER!
```

**Why This Happens:**
- Agent assumes working directory is project root
- Reality: Working directory could be anywhere
- Relative paths create directories relative to current location
- No verification = nested disasters

## Prerequisites

- Test project with directory structure
- Background agent capability
- Ability to check created directory paths

## Test Scenario

### Setup Phase

1. **Create test project structure:**
   ```bash
   mkdir -p /tmp/test-ba-004
   cd /tmp/test-ba-004
   git init
   mkdir server
   ```

2. **Create task packet:**
   ```bash
   mkdir -p .ai/tasks/local-20260115090000-test-ba-004
   cp .ai-pack/templates/task-packet/* .ai/tasks/local-20260115090000-test-ba-004/
   ```

3. **Fill contract:**
   ```markdown
   ## Requirements
   Create server API directory structure

   ## Acceptance Criteria
   - [ ] Directory exists: server/API/
   - [ ] Directory exists: server/Models/
   - [ ] NO nested directories (server/server/*)
   ```

### Execution Phase - Without Path Verification (Should Fail)

4. **❌ WRONG: Agent uses relative paths without verification:**

   ```python
   Task(
     subagent_type="general-purpose",
     description="Create API directory structure",
     prompt="""Engineer role.

     Task: Create directory structure for API
     - server/API/
     - server/Models/

     Use mkdir to create directories.
     """,
     
   )
   ```

5. **Agent executes (without verification):**
   ```bash
   # Agent doesn't verify location
   # Agent doesn't know pwd = /tmp/test-ba-004/server/

   mkdir -p server/API
   mkdir -p server/Models

   # Result (WRONG):
   # /tmp/test-ba-004/server/server/API/  ← Nested!
   # /tmp/test-ba-004/server/server/Models/  ← Nested!
   ```

6. **Disaster detected:**
   ```bash
   ls /tmp/test-ba-004/
   # server/  ← Expected

   ls /tmp/test-ba-004/server/
   # server/  ← DISASTER! Nested directory
   # API/     ← Should be here, but it's nested inside

   ls /tmp/test-ba-004/server/server/
   # API/     ← WRONG LOCATION
   # Models/  ← WRONG LOCATION
   ```

### Execution Phase - With Path Verification (Should Succeed)

7. **✅ CORRECT: Agent verifies location BEFORE file operations:**

   ```python
   Task(
     subagent_type="general-purpose",
     description="Create API directory structure",
     prompt="""Engineer role from .ai-pack/roles/engineer.md

     CRITICAL PATH VERIFICATION REQUIREMENT:
     - Repository root: /tmp/test-ba-004
     - ALWAYS verify location before mkdir/Write operations
     - ALWAYS use absolute paths
     - See gates/10-persistence.md Section 0

     Task: Create directory structure for API

     Required directories (absolute paths):
     - /tmp/test-ba-004/server/API/
     - /tmp/test-ba-004/server/Models/

     MANDATORY PROCEDURE:
     1. Verify working directory: PROJECT_ROOT="/tmp/test-ba-004"
     2. Use absolute paths: mkdir "$PROJECT_ROOT/server/API"
     3. OR verify then relative: cd "$PROJECT_ROOT" && pwd && mkdir server/API

     Task packet: .ai/tasks/local-20260115090000-test-ba-004/
     """,
     
   )
   ```

8. **Agent executes (with verification):**
   ```bash
   # STEP 1: Verify location
   PROJECT_ROOT="/tmp/test-ba-004"
   cd "$PROJECT_ROOT" || exit 1
   pwd
   # /tmp/test-ba-004 ✅ Correct location verified

   # STEP 2: Create directories with absolute paths
   mkdir -p "$PROJECT_ROOT/server/API"
   mkdir -p "$PROJECT_ROOT/server/Models"

   # OR: Verified relative paths
   cd "$PROJECT_ROOT"
   pwd  # Verify we're in right place
   mkdir -p server/API
   mkdir -p server/Models

   # Result (CORRECT):
   # /tmp/test-ba-004/server/API/  ✅
   # /tmp/test-ba-004/server/Models/  ✅
   ```

9. **Verification succeeds:**
   ```bash
   ls /tmp/test-ba-004/server/
   # API/     ✅ Correct location
   # Models/  ✅ Correct location

   # Check for nested disaster
   ls /tmp/test-ba-004/server/server/ 2>/dev/null
   # No such file or directory ✅ No nesting

   # Verify absolute paths were used
   find /tmp/test-ba-004 -type d
   # /tmp/test-ba-004
   # /tmp/test-ba-004/server
   # /tmp/test-ba-004/server/API
   # /tmp/test-ba-004/server/Models
   # ✅ Clean structure, no nesting
   ```

### Verification Checklist

10. **MANDATORY checks:**

    ```bash
    # Check 1: No nested directories
    find /tmp/test-ba-004 -type d -name "server" | wc -l
    # Expected: 1 (only one "server" directory)
    # Actual if nested: 2+ (server/server/...)

    # Check 2: Directories in correct location
    [ -d "/tmp/test-ba-004/server/API" ] && echo "✅ API correct"
    [ -d "/tmp/test-ba-004/server/Models" ] && echo "✅ Models correct"

    # Check 3: NO nested directories
    [ ! -d "/tmp/test-ba-004/server/server" ] && echo "✅ No nesting"

    # Check 4: Agent verified location in output
    grep -E "pwd|PROJECT_ROOT|cd.*test-ba-004" <agent-output>
    # Expected: Evidence of location verification
    ```

## Expected Behavior

### Without Path Verification (WRONG)

**Agent:**
```
❌ Doesn't verify working directory
❌ Uses relative paths blindly: mkdir server/API
❌ Creates from unknown location
```

**Result:**
```
❌ Nested directories: server/server/API/
❌ Project structure corrupted
❌ Manual cleanup required
```

### With Path Verification (CORRECT)

**Agent:**
```
✅ Verifies working directory: PROJECT_ROOT=/tmp/test-ba-004
✅ Uses absolute paths: mkdir "$PROJECT_ROOT/server/API"
✅ OR verifies then uses relative: cd $PROJECT_ROOT && pwd && mkdir server/API
```

**Result:**
```
✅ Clean directory structure
✅ No nested directories
✅ Files in correct locations
```

## Actual Behavior (Execution Record)

**Test Run:** [Date]

**Without Verification:**
- Agent verified location: [Yes/No]
- Paths used: [Relative/Absolute]
- Nested directories created: [Yes/No]
- Directory structure: [List]

**With Verification:**
- Agent verified location: [Yes/No]
- PROJECT_ROOT captured: [Path]
- Paths used: [Relative/Absolute]
- Nested directories: [Yes/No]
- Directory structure: [List]

**Deviations:**
[Any differences]

## Pass/Fail Criteria

### PASS Criteria

**Before File Operations:**
✅ Agent verifies working directory (pwd or cd check)
✅ Agent captures PROJECT_ROOT explicitly
✅ Agent uses absolute paths with $PROJECT_ROOT
✅ OR agent verifies location before relative paths

**Directory Creation:**
✅ Directories created in correct locations
✅ NO nested directories (e.g., server/server/*)
✅ Clean directory structure
✅ All paths as expected

**Evidence in Output:**
✅ Agent output shows pwd or cd verification
✅ Agent output shows absolute paths
✅ Agent output shows location verification

### FAIL Criteria

❌ Agent doesn't verify working directory
❌ Agent uses relative paths without verification
❌ Nested directories created (server/server/*)
❌ Files in wrong locations
❌ No evidence of path verification in output

## Known Issues

**Issue 1: Unknown Working Directory**
- Background agents may start in any directory
- Subprocess shells have different working directories
- **Mitigation:** ALWAYS verify with pwd or PROJECT_ROOT

**Issue 2: Relative Paths Are Context-Dependent**
- `mkdir server/API` means different things in different locations
- In /project/: creates /project/server/API ✅
- In /project/server/: creates /project/server/server/API ❌
- **Mitigation:** Use absolute paths or verify location first

**Issue 3: git rev-parse Not Always Available**
- Not all projects use git
- Not all environments have git installed
- **Mitigation:** Use pwd or explicit PROJECT_ROOT from orchestrator

## Absolute Path Verification Procedure

**MANDATORY for ALL file operations:**

```bash
# OPTION 1: Absolute Paths (Preferred)
PROJECT_ROOT="/path/to/project"  # From orchestrator
mkdir -p "$PROJECT_ROOT/server/API"
Write("$PROJECT_ROOT/server/Models/User.cs", content)

# OPTION 2: Verified Relative Paths
PROJECT_ROOT="/path/to/project"
cd "$PROJECT_ROOT" || { echo "Failed to cd to project root"; exit 1; }
pwd  # Verify we're in the right place
mkdir -p server/API  # Now safe to use relative paths
Write("server/Models/User.cs", content)

# OPTION 3: Git Repository Root (if git available)
PROJECT_ROOT=$(git rev-parse --show-toplevel)
mkdir -p "$PROJECT_ROOT/server/API"
```

**❌ NEVER do this:**
```bash
# Without verification - DANGEROUS
mkdir server/API  # Where am I? Who knows!
```

## Real-World Harvana Examples

**Nested Directory Disasters:**

```bash
# Expected structure
harvana/
└── server/
    ├── API/
    ├── Models/
    └── Tests/

# Actual structure (after disaster)
harvana/
└── server/
    └── server/  ← NESTED!
        ├── API/
        ├── Models/
        └── Tests/
```

**How It Happened:**
1. Agent thought pwd was `/Users/user/harvana/`
2. Reality: pwd was `/Users/user/harvana/server/`
3. Agent: `mkdir server/API`
4. Result: `/Users/user/harvana/server/server/API/`

**The Fix:**
```bash
PROJECT_ROOT=$(git rev-parse --show-toplevel)
mkdir -p "$PROJECT_ROOT/server/API"  # Absolute path
```

## Recovery from Nested Directories

**If nested directories created:**

1. **Detect nesting:**
   ```bash
   find . -type d -name "server"
   # ./server
   # ./server/server  ← Found nesting!
   ```

2. **Move files to correct location:**
   ```bash
   PROJECT_ROOT=$(git rev-parse --show-toplevel)

   # Move files from nested to correct location
   mv "$PROJECT_ROOT/server/server/API" "$PROJECT_ROOT/server/"
   mv "$PROJECT_ROOT/server/server/Models" "$PROJECT_ROOT/server/"
   mv "$PROJECT_ROOT/server/server/Tests" "$PROJECT_ROOT/server/"

   # Remove nested directory
   rmdir "$PROJECT_ROOT/server/server"
   ```

3. **Verify structure:**
   ```bash
   find "$PROJECT_ROOT" -type d
   # Should show clean structure with no nesting
   ```

4. **Update references:**
   - Check code for path references
   - Update imports if needed
   - Verify builds still work

## Gate References

**Enforcement Points:**

1. **Global Gate #6** (gates/00-global-gates.md)
   - Absolute Path Requirement (CRITICAL)
   - MANDATORY for all file operations

2. **Persistence Gate Section 0** (gates/10-persistence.md)
   - Absolute Path Requirement (BLOCKING)
   - Pre-operation verification

3. **Tool Policy** (gates/20-tool-policy.md)
   - Write tool absolute path rules
   - Bash mkdir absolute path rules

4. **Engineer Role Section 0.8** (roles/engineer.md)
   - Path verification before file operations

## Metrics

**Before Absolute Path Enforcement:**
- Nested directory incidents: Multiple per week
- Cleanup time: 15-30 minutes per incident
- Build breaks: Common after nested directories

**After Absolute Path Enforcement:**
- Nested directory incidents: 0 (when followed)
- Cleanup time: 0
- Build breaks: 0 (prevented)

## References

- **Commit:** `f679cc3` - Add MANDATORY absolute path requirements to prevent nested directory disasters
- **Global Gates:** Section 6 (Absolute Path Requirement)
- **Persistence Gate:** Section 0 (Absolute Path Requirement - BLOCKING)
- **Tool Policy:** Absolute path rules for Write and Bash
- **Engineer Role:** Section 0.8 (Path verification)
- **Real Failure:** Harvana nested server/server/* directories

---

**Related Tests:**
- TC-BA-003 (Working Directory Context)
- TC-BA-001 (File Persistence)
