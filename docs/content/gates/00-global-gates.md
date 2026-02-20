---
sidebar_position: 1
title: "Global Gates"
---

# Global Gates

**Version:** 1.1.1
**Last Updated:** 2026-01-17

## Overview

Global gates define universal rules and constraints that apply to all AI agent workflows, regardless of role or workflow type. These gates establish the foundational principles that cannot be violated.

## Core Principles

### 0. Task Packet Infrastructure (MANDATORY)

**Rule:** All non-trivial tasks MUST have task packet infrastructure before implementation.

**Requirement:**
```text
BEFORE starting implementation:
  IF task is non-trivial THEN
    REQUIRE: .ai/tasks/<beads-id>-<YYYYMMDDHHMMSS>-<short-desc>/ exists
    REQUIRE: 00-contract.md filled out
    REQUIRE: 10-plan.md created

    IF requirements not met THEN
      STOP all work
      CREATE task packet infrastructure FIRST
      ONLY THEN proceed
    END IF
  END IF
END BEFORE
```text

**Non-Trivial Task Definition:**
- Requires more than 2 simple steps
- Involves writing or modifying code
- Takes more than 30 minutes
- Requires tests or verification

**Why This Gate Exists:**
- Ensures clear requirements before work begins
- Provides audit trail of decisions
- Enables proper task coordination
- Prevents scope creep and confusion
- Documents acceptance criteria upfront

**Enforcement:** This gate CANNOT be bypassed. Task packets are mandatory infrastructure.

---

### 1. Safety First

**Rule:** Never perform destructive operations without explicit user approval.

**Destructive operations include:**
- Deleting files or directories
- Force-pushing to git repositories
- Dropping databases or tables
- Overwriting existing files without reading them first
- Running commands with `--force`, `--hard`, or similar flags

**Implementation:**
```text
IF operation is destructive THEN
  request user approval
  wait for confirmation
  proceed only if approved
END IF
```text

**Exceptions:** None. Safety gates cannot be bypassed.

---

### 2. Test-Driven Development (MANDATORY, BLOCKING)

**Rule:** All code MUST be developed using Test-Driven Development (TDD). This is MANDATORY and ENFORCED.

**Requirements:**
- Tests written BEFORE implementation (RED-GREEN-REFACTOR cycle)
- Git history shows test-first commits
- Test pyramid properly structured (65-80% unit, 15-25% integration, 5-10% e2e)
- Coverage targets met (80%+ overall, 95%+ critical paths)

**Implementation:**
```text
BEFORE writing any implementation code:
  STEP 1: Write failing test (RED)
  STEP 2: Write minimal code to pass (GREEN)
  STEP 3: Refactor while keeping tests green (REFACTOR)

  IF Engineer skips TDD THEN
    Tester BLOCKS approval
    Work is REJECTED
    Engineer must redo with TDD
  END IF
END BEFORE
```text

**Enforcement:** See [TDD Enforcement Gate](05-tdd-enforcement.md) for complete requirements.

**References:**
- Martin Fowler: [Practical Test Pyramid](https://martinfowler.com/articles/practical-test-pyramid.html)
- Kent Beck: *Test-Driven Development: By Example*
- Robert C. Martin: *The Clean Coder* - Chapter 5

---

### 3. Quality Baseline (BLOCKING)

**Rule:** All code must meet minimum quality standards before completion.

**Requirements:**
- All tests must pass (zero tolerance for failures)
- Code coverage: 80-90% minimum (enforced by TDD gate)
- **Build with ZERO WARNINGS (zero tolerance - BLOCKING)**
- No critical code review findings unresolved
- Follows language-specific style guides
- Properly formatted (spaces, not tabs)

**⚠️ CRITICAL: Zero Warnings Policy**

Code MUST compile/build with ZERO WARNINGS. This is NON-NEGOTIABLE.

**Why Zero Warnings:**
- Warnings indicate quality issues
- Warnings become production bugs
- Professional code has zero warnings
- Teams ignoring warnings accumulate technical debt

**Implementation:**
```text
BEFORE marking task complete:
  run all tests → must pass
  check coverage → must meet target
  run build with warnings-as-errors → MUST show 0 warnings
  verify formatting → must comply

  IF any check fails THEN
    fix issues
    re-run checks
  END IF

  IF any warnings exist THEN
    STOP - BLOCK completion
    FIX all warnings
    RE-RUN build
    ONLY proceed when: 0 warnings
  END IF
END BEFORE
```text

**Language-Specific Enforcement:**
```text
C/C++:        -Werror -Wall -Wextra
C#:           /warnaserror
Java:         -Werror
TypeScript:   --strict, --max-warnings 0
Python:       flake8, mypy --strict
Go:           go vet, golangci-lint --max-issues-per-linter 0
Rust:         clippy -- -D warnings
```text

---

### 5. Communication Protocol

**Rule:** Know when to ask versus when to execute autonomously.

**Ask the user when:**
- Requirements are ambiguous or unclear
- Multiple valid approaches exist with trade-offs
- User preferences might influence the decision
- You encounter unexpected errors or blockers
- You need to deviate from established patterns

**Execute autonomously when:**
- Requirements are clear and unambiguous
- Standard patterns apply
- The approach is well-established
- You're confident in the solution
- Following explicit user instructions
- Running non-destructive commands (tests, builds, coverage, etc.)

**Implementation:**
```text
IF uncertain OR multiple approaches OR user preference matters THEN
  use AskUserQuestion tool
  wait for response
  proceed with user's choice
ELSE
  execute autonomously
  report results
END IF
```text

---

### 5. Error Handling and Recovery

**Rule:** Handle errors gracefully and recover when possible.

**Requirements:**
- Never silently ignore errors
- Always report errors to the user
- Attempt recovery when possible
- Provide actionable error messages
- Log failures for debugging

**Error Response Protocol:**
1. **Detect** the error or failure
2. **Analyze** the cause and impact
3. **Report** to user with context
4. **Suggest** potential solutions
5. **Attempt** recovery if safe
6. **Escalate** if unable to resolve

**Example:**
```text
Test suite failed with 3 failures

Failures:
1. test_authentication: Expected 200, got 401
2. test_database: Connection timeout
3. test_validation: Null pointer exception

Analysis: Database connection issue likely causing cascading failures.

Recovery attempt: Check database configuration...
```text

---

### 5. Incremental Progress

**Rule:** Make progress in small, verifiable steps.

**Principles:**
- Break large tasks into smaller subtasks
- Verify each step before proceeding
- Commit working changes frequently
- Never make sweeping changes without validation
- Test early, test often

**Implementation:**
```text
FOR each subtask:
  implement change
  run tests
  verify success
  IF tests pass THEN
    commit (if appropriate)
    proceed to next subtask
  ELSE
    fix issue
    retry
  END IF
END FOR
```text

---

### 6. Absolute Paths for File Operations (CRITICAL)

**Rule:** ALWAYS use absolute paths for file creation and directory operations. NEVER use relative paths without verifying current working directory.

**⚠️ CRITICAL: This prevents nested directory disasters like `server/server/API/`**

**Requirements:**
```text
BEFORE creating files or directories:
  STEP 1: Verify current working directory
    pwd  # Check where you are

  STEP 2: Use absolute paths for ALL file operations
    Write tool: /absolute/path/to/file.ext
    mkdir: /absolute/path/to/directory

  STEP 3: If using relative paths REQUIRED
    VERIFY you're in correct directory FIRST
    IF not in expected directory THEN
      cd to correct directory
      VERIFY with pwd
      THEN use relative path
    END IF

  ❌ NEVER do this:
    mkdir server/API        # Could create anywhere!
    Write("api/file.js")    # Where will this go?

  ✅ ALWAYS do this:
    mkdir /home/user/project/server/API
    Write("/home/user/project/api/file.js")

  OR verify location first:
    pwd  # Check: /home/user/project
    mkdir server/API  # Now safe, creates /home/user/project/server/API
END BEFORE
```text

**Why This Rule Exists:**
- Prevents nested directory disasters (`server/server/API/`)
- Eliminates ambiguity about file locations
- Avoids creating files in wrong directories
- Ensures reproducible file operations
- Reduces debugging time

**Common Mistake - Harvana Example:**
```text
# ❌ WRONG: Agent not in project root
pwd  # Returns: /home/user/project/server
mkdir server/API  # Creates: /home/user/project/server/server/API (DISASTER!)

# ✅ CORRECT: Use absolute path
mkdir /home/user/project/server/API  # Always creates in right place
```text

**Enforcement:** This gate CANNOT be bypassed. File operations with ambiguous paths will be rejected.

---

### 7. Read Before Write

**Rule:** Always read existing code before modifying it.

**Requirements:**
- Use Read tool before Edit tool
- Understand context before making changes
- Identify patterns to follow
- Avoid breaking existing functionality
- Maintain consistency with codebase

**Rationale:** Writing code without reading it first leads to:
- Breaking existing patterns
- Introducing inconsistencies
- Missing important context
- Creating merge conflicts
- Violating architectural principles

---

### 7. Prefer Editing Over Creating

**Rule:** Always prefer editing existing files over creating new ones.

**Guidelines:**
- Search for existing files that could be modified
- Create new files only when truly necessary
- Avoid duplicating functionality
- Maintain single source of truth
- Don't create documentation files unless explicitly requested

**Implementation:**
```text
BEFORE creating new file:
  search for existing files with similar purpose
  IF existing file found THEN
    modify existing file
  ELSE IF truly necessary THEN
    create new file
  END IF
END BEFORE
```text

---

### 8. Test-Driven Development

**Rule:** Write tests first, then implementation.

**TDD Cycle:**
1. **Red:** Write a failing test
2. **Green:** Write minimal code to pass
3. **Refactor:** Improve code while keeping tests green

**Requirements:**
- Tests written before implementation
- Tests define expected behavior
- Implementation satisfies tests
- Refactoring preserves test success
- Coverage meets 80-90% target

---

### 9. No Over-Engineering

**Rule:** Only solve the problem at hand, not hypothetical future problems.

**Principles:**
- YAGNI (You Aren't Gonna Need It)
- Don't add features not requested
- Don't create abstractions for one use case
- Don't optimize prematurely
- Keep solutions simple and focused

**Anti-patterns to avoid:**
- Adding configurability not needed
- Creating frameworks for single use
- Designing for scale not required
- Adding error handling for impossible cases
- Implementing "nice to have" features

---

### 9. Documentation Standards (MANDATORY)

**Rule:** All documentation artifacts MUST follow established documentation standards, including mandatory Mermaid for diagrams.

**Requirements:**
- All diagrams MUST use Mermaid format (no ASCII art, no text diagrams, no Unicode box drawings)
- Inline documentation for complex algorithms
- ADRs for significant architectural decisions
- Self-documenting code as primary documentation

**Diagram Standard (MANDATORY, BLOCKING):**
```text
BEFORE creating any diagram in documentation:
  ✅ MUST: Use Mermaid syntax in markdown code blocks
  ✅ MUST: Use appropriate diagram type (flowchart, sequence, class, state)
  ❌ NEVER: Use ASCII art, text diagrams, or Unicode box drawings for conceptual diagrams
  ❌ NEVER: Use external tools like PlantUML, Draw.io, etc.

  ✅ EXCEPTION: ASCII art IS acceptable for directory/file structure layouts
    Example: project/
             ├── src/
             └── docs/

  IF diagram is not Mermaid AND not a directory layout THEN
    REJECT documentation
    REQUIRE conversion to Mermaid
  END IF
END BEFORE
```text

**Rationale:**
- Mermaid renders properly in GitHub, GitLab, and modern documentation tools
- Mermaid syntax is version-controllable and diffable
- Mermaid diagrams are maintainable without artistic skill
- Text diagrams break with formatting changes

**Reference:** See [Documentation Standards](../quality/clean-code/11-documentation-standards.md) for complete requirements.

---

### 10. Code Must Be Reversible

**Rule:** All changes must be easily reversible.

**Requirements:**
- Use version control (git)
- Make atomic commits
- Write clear commit messages
- Don't use destructive git operations
- Keep feature branches manageable

**Git Protocol:**
- Commit working changes frequently
- Use descriptive commit messages
- Never use `--force` on shared branches
- Never use `--amend` on pushed commits
- Keep commits focused and atomic

---

## Gate Enforcement

These gates are **mandatory** and **cannot be bypassed**. They apply to:
- All agent roles (orchestrator, worker, reviewer)
- All workflows (standard, feature, bugfix, refactor, research)
- All tasks, regardless of size or complexity

### Violation Consequences

If a global gate is violated:
1. **Stop immediately**
2. **Report violation to user**
3. **Assess damage and impact**
4. **Propose remediation plan**
5. **Wait for user approval before proceeding**

## Integration

These global gates integrate with:
- **[Persistence Gates](10-persistence.md)** - File operation specifics
- **[Tool Policy Gates](20-tool-policy.md)** - Tool usage rules
- **[Execution Strategy Gate](25-execution-strategy.md)** - Parallel execution enforcement
- **[Verification Gates](30-verification.md)** - Quality checks
- **[Code Quality Review Gate](35-code-quality-review.md)** - Mandatory Tester and Reviewer validation for code changes

All specialized gates must comply with these global principles.

---

**Last reviewed:** 2026-01-08
**Next review:** Quarterly or when adding new gate types
