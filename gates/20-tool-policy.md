# Tool Policy Gates

**Version:** 1.0.0
**Last Updated:** 2026-01-07

## Overview

Tool policy gates define when and how AI agents should use available tools. These policies ensure efficient, safe, and appropriate tool usage throughout workflows.

## Tool Selection Principles

### 1. Use Specialized Tools Over General Ones

**Rule:** Always prefer specialized tools when available.

**Rationale:**
- Specialized tools have built-in safety checks
- Better error messages
- Optimized for specific operations
- More reliable results

**Tool Preference Hierarchy:**

#### File Operations
```
✅ PREFERRED:
  Read tool      - for reading files
  Write tool     - for creating files
  Edit tool      - for modifying files
  Glob tool      - for finding files by pattern

❌ AVOID:
  Bash(cat)      - use Read instead
  Bash(echo)     - use Write instead
  Bash(sed/awk)  - use Edit instead
  Bash(find)     - use Glob instead
```

#### Search Operations
```
✅ PREFERRED:
  Grep tool      - for content search
  Glob tool      - for file name search

❌ AVOID:
  Bash(grep)     - use Grep tool instead
  Bash(rg)       - use Grep tool instead
  Bash(find)     - use Glob tool instead
```

#### Communication
```
✅ PREFERRED:
  Direct output  - speak directly to user

❌ AVOID:
  Bash(echo)     - not for user communication
  Code comments  - not for conversation
```

---

### 2. Parallel Tool Execution

**Rule:** Execute independent operations in parallel when possible.

**When to Parallelize:**
- Operations are independent (no shared state)
- No dependencies between operations
- All operations likely to succeed
- Faster results benefit user

**When NOT to Parallelize:**
- Operations have dependencies
- Later operations need earlier results
- Operations might fail conditionally
- Sequential order matters

**Examples:**

**✅ Good Parallelization:**
```
Single message with multiple tool calls:
- Read(file1.cpp)
- Read(file2.cpp)
- Read(file3.cpp)
→ All reads are independent, parallelize them
```

**❌ Bad Parallelization:**
```
DON'T DO:
- Read(file.cpp)        # need result first
- Edit(file.cpp)        # depends on read
→ Must be sequential
```

**Implementation:**
```
IF operations independent THEN
  make all tool calls in single message
  process results together
ELSE
  make tool calls sequentially
  wait for each result
END IF
```

---

### 3. Tool Approval Requirements

**Rule:** Some tools require user approval, others don't.

### Pre-Approved Tools

These can be used without asking:
```
✅ Safe to use freely:
- Read         - reading files
- Glob         - finding files
- Grep         - searching content
- WebFetch     - fetching web content (specific domains)
- WebSearch    - searching web

✅ Pre-approved Bash commands (read-only/non-destructive):
- ls, pwd, cd  - navigation and listing
- cat, head, tail, wc - file viewing (prefer Read tool)
- find, grep   - searching (prefer Glob/Grep tools)
- git status, git diff, git log - git inspection
- which, whereis - finding executables

✅ Pre-approved development commands (non-destructive):
Build/Test/Coverage:
- cmake, make, ninja - building code
- cmake --build <dir> - building projects
- ctest, pytest, jest, cargo test, go test, npm test - running tests
- ./build*/test_* --gtest_* - running specific test executables
- gcov, lcov, coverage - generating coverage reports
- clang-format --dry-run, black --check - checking formatting

Inspection/Analysis:
- clang-tidy, pylint, eslint - linting (read-only mode)
- valgrind, lldb --batch, gdb --batch - debugging tools
- nm, objdump, readelf - binary inspection
- tree, du, df - filesystem info
```

### Approval Required

These require user permission:
```
⚠️ Ask before using:
- Write        - creating new files (unless clearly needed for task)
- Edit         - modifying files (unless clearly part of task)
- Bash(rm)     - deleting files
- Bash(mv)     - moving files
- Bash(cp)     - copying files (may create new files)
- git add, git commit, git push - git write operations
- Package installation (npm install, pip install, cargo add, etc.)
- Configuration changes (settings, environment variables)
- Any destructive operation
- Any operation that modifies state
```

### Prohibited Without Explicit Request

```
❌ Never use without user explicitly requesting:
- git push --force
- git reset --hard
- rm -rf
- Database drops
- System configuration changes
```

---

### 4. Resource Consumption Limits

**Rule:** Be mindful of resource usage.

**Limits:**
- **File reads:** Read whole files when possible; use offset/limit only for very large files
- **Search operations:** Use appropriate scope (don't search entire filesystem unnecessarily)
- **Parallel operations:** Reasonable limit (4-5 parallel operations maximum)
- **Agent spawning:** Use Task tool judiciously; prefer direct operations for simple tasks

**Guidelines:**
```
BEFORE expensive operation:
  assess necessity
  check scope appropriateness
  consider alternatives
  IF truly needed THEN
    execute with appropriate limits
  END IF
END BEFORE
```

---

### 5. External Service Access

**Rule:** Follow policies for external service access.

### Web Access
```
✅ Allowed:
- WebSearch for current information
- WebFetch for documentation (approved domains)
- WebFetch for public APIs (with user context)

⚠️ Restricted:
- APIs requiring authentication
- Services with rate limits
- Commercial services
- User's production systems
```

### Pre-Approved Domains for WebFetch
```
- martinfowler.com
- docs.cline.bot
- gist.github.com
- google.github.io
- Other documentation sites (case-by-case)
```

---

### 6. Tool Fallback Strategies

**Rule:** Have fallback strategies when preferred tool fails.

**Strategy:**
```
TRY:
  use preferred specialized tool
CATCH error:
  assess error type
  IF tool unavailable THEN
    try alternative tool
  ELSE IF permission denied THEN
    request user approval
  ELSE IF invalid input THEN
    correct input and retry
  END IF
END TRY
```

**Example Fallbacks:**
```
Preferred: Grep tool
Fallback: Bash(grep) if Grep tool fails

Preferred: Read tool
Fallback: Bash(cat) if Read tool fails

Preferred: Task tool with Explore agent
Fallback: Direct Glob/Grep if agent unavailable
```

---

### 7. Agent Spawning Policy

**Rule:** Use Task tool with specialized agents appropriately.

### When to Use Task Tool

**✅ Use Task tool for:**
- Open-ended codebase exploration
- Complex multi-step tasks
- Tasks requiring specialized expertise
- Research that might need multiple search rounds
- Planning implementation approaches

**❌ Don't use Task tool for:**
- Reading a specific known file (use Read)
- Simple grep for known pattern (use Grep)
- Direct file operations
- Single-step operations
- Trivial tasks

### Agent Types and Usage

```
Explore agent:
  USE FOR: "Where is X handled?", "How does Y work?", "What's the structure?"
  DON'T USE FOR: "Read file.cpp", "Search for 'class Foo'"

Plan agent:
  USE FOR: Multi-file feature implementation planning
  DON'T USE FOR: Single-file edits, simple fixes

General-purpose agent:
  USE FOR: Complex tasks requiring multiple tool types
  DON'T USE FOR: Simple operations available directly
```

### Agent Limits and Enforced Configuration

**ENFORCED POLICY: Automatic parallelization analysis for work packages**

**This policy is enforced by:** [Execution Strategy Gate](25-execution-strategy.md)

```
Agent Spawning Policy (ENFORCED):
- MANDATORY: Analyze parallelization for 3+ subtasks
- MANDATORY: Spawn parallel workers when criteria met
- Maximum concurrent agents: 5
- REQUIRED: Launch agents in single message block (true parallelism)
- Resume agents when continuing their work
- Don't spawn agent for tasks you can do directly

Enforcement Rules:
✅ MANDATORY parallel execution when:
   - 3+ independent subtasks identified
   - Subtasks touch different files/modules
   - No cross-subtask dependencies
   - Each has isolated acceptance criteria
   - Shared context constraints respected

   IF above criteria met AND orchestrator proceeds sequentially THEN
     GATE VIOLATION (Execution Strategy Gate)
     REQUIRE documented justification OR
     REQUIRE switch to parallel execution
   END IF

⚠️ SEQUENCE when (must justify):
   - Subtasks have execution dependencies
   - Same files modified by multiple subtasks
   - Results needed for subsequent tasks
   - Shared context conflicts cannot be resolved

🔧 HYBRID approach when:
   - Mix of dependent and independent tasks
   - MANDATORY: Parallelize independent groups
   - REQUIRED: Sequence only dependent chains
```

**Shared Context Requirements (CRITICAL):**
```
Parallel workers operate in SHARED context:

✅ SHARED resources (all workers use same):
   - Source repository (no per-worker branches)
   - Build folders (no deletion/recreation)
   - Coverage data (merge, don't overwrite)
   - Test databases (coordinate access)

❌ FORBIDDEN during parallel execution:
   - Deleting build folders
   - Removing coverage files
   - Creating per-worker git branches
   - Destructive git operations
   - Context-invalidating operations

⚠️ REQUIRES COORDINATION:
   - Build operations
   - Coverage report generation
   - Database migrations
   - Git commits
```

**Parallel Launch Pattern (ENFORCED):**
```
// CORRECT: Single message with multiple Task calls
Task(worker, subtask1) + Task(worker, subtask2) + Task(worker, subtask3)
→ All 3 spawn truly in parallel (REQUIRED for 3+ independent)

// INCORRECT: Sequential messages (GATE VIOLATION if applicable)
Task(worker, subtask1)
[wait for result]
Task(worker, subtask2)
→ Serial execution - only acceptable if justified
```

---

### 8. Bash Tool Usage

**Rule:** Reserve Bash tool for actual system commands.

### Appropriate Bash Usage
```
✅ Good uses:
- git commands (status, diff, log, commit, push)
- Build commands (make, cmake, npm, pip)
- Test runners (pytest, jest, cargo test)
- Package managers (npm install, pip install)
- System utilities (docker, kubectl)
```

### Inappropriate Bash Usage
```
❌ Bad uses:
- cat/head/tail      → use Read tool
- grep/rg            → use Grep tool
- find               → use Glob tool
- sed/awk            → use Edit tool
- echo for output    → output directly
- echo > file        → use Write tool
```

### Bash Command Construction
```
Sequential (use &&):
  npm install && npm test && npm build
  → Operations depend on previous success

Parallel (use ; or separate calls):
  git status ; git diff
  → Independent operations

Quoted paths:
  cd "/path with spaces/dir"
  python "/path with spaces/script.py"
  → Always quote paths with spaces
```

---

### 9. Interactive Tools Prohibition

**Rule:** Never use interactive tools.

**Prohibited:**
```
❌ Interactive commands not supported:
- git rebase -i
- git add -i
- vim/nano/emacs
- Interactive debuggers
- Prompts requiring user input in tool
```

**Alternatives:**
```
Instead of: git rebase -i
Use: git rebase <branch> with --onto if needed

Instead of: vim file
Use: Edit tool

Instead of: interactive debugger
Use: logging and test-driven debugging
```

---

### 10. Tool Result Verification

**Rule:** Verify tool results before proceeding.

**Verification Steps:**
```
AFTER tool execution:
  check exit code/status
  verify expected output received
  validate result format
  IF tool failed THEN
    analyze error message
    determine recovery strategy
    retry OR report to user
  END IF
END AFTER
```

**Error Handling:**
```
Tool Error Detected:
  1. Parse error message
  2. Determine if recoverable
  3. IF recoverable THEN
       adjust and retry
     ELSE
       report to user
       suggest alternatives
     END IF
```

---

## Tool-Specific Policies

### Read Tool
- Prefer reading entire file
- Use offset/limit only for very large files (>10,000 lines)
- Always read before editing
- Don't make assumptions about file contents

### Write Tool
- Only create files when necessary
- Prefer Edit over Write for existing files
- Must read existing file first if overwriting
- Never create documentation unless requested

### Edit Tool
- Must read file first
- Preserve exact indentation from file
- Make targeted changes only
- Verify old_string uniqueness

### Glob Tool
- Use for file pattern matching
- Faster than recursive Bash find
- Supports standard glob patterns
- Use Task tool for open-ended searches

### Grep Tool
- Use for content search
- Supports full regex
- Set appropriate output_mode
- Use multiline:true for cross-line patterns

### Task Tool
- Use specialized agent types
- Launch in parallel when appropriate
- Resume agents when continuing work
- Don't overuse - prefer direct operations when simple

---

## Integration

Tool policies integrate with:
- **[Global Gates](00-global-gates.md)** - Safety requirements
- **[Persistence Gates](10-persistence.md)** - File operation rules
- **[Verification Gates](30-verification.md)** - Result validation

---

**Last reviewed:** 2026-01-07
**Next review:** When new tools added or policies need adjustment
