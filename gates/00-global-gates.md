# Global Gates

**Version:** 1.0.0
**Last Updated:** 2026-01-07

## Overview

Global gates define universal rules and constraints that apply to all AI agent workflows, regardless of role or workflow type. These gates establish the foundational principles that cannot be violated.

## Core Principles

### 0. Task Packet Infrastructure (MANDATORY)

**Rule:** All non-trivial tasks MUST have task packet infrastructure before implementation.

**Requirement:**
```
BEFORE starting implementation:
  IF task is non-trivial THEN
    REQUIRE: .ai/tasks/YYYY-MM-DD_task-name/ exists
    REQUIRE: 00-contract.md filled out
    REQUIRE: 10-plan.md created

    IF requirements not met THEN
      STOP all work
      CREATE task packet infrastructure FIRST
      ONLY THEN proceed
    END IF
  END IF
END BEFORE
```

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
```
IF operation is destructive THEN
  request user approval
  wait for confirmation
  proceed only if approved
END IF
```

**Exceptions:** None. Safety gates cannot be bypassed.

---

### 2. Quality Baseline

**Rule:** All code must meet minimum quality standards before completion.

**Requirements:**
- All tests must pass (zero tolerance for failures)
- Code coverage: 80-90% target
- No critical code review findings unresolved
- Follows language-specific style guides
- Properly formatted (spaces, not tabs)

**Implementation:**
```
BEFORE marking task complete:
  run all tests → must pass
  check coverage → must meet target
  verify formatting → must comply
  IF any check fails THEN
    fix issues
    re-run checks
  END IF
END BEFORE
```

---

### 3. Communication Protocol

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
```
IF uncertain OR multiple approaches OR user preference matters THEN
  use AskUserQuestion tool
  wait for response
  proceed with user's choice
ELSE
  execute autonomously
  report results
END IF
```

---

### 4. Error Handling and Recovery

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
```
Test suite failed with 3 failures

Failures:
1. test_authentication: Expected 200, got 401
2. test_database: Connection timeout
3. test_validation: Null pointer exception

Analysis: Database connection issue likely causing cascading failures.

Recovery attempt: Check database configuration...
```

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
```
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
```

---

### 6. Read Before Write

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
```
BEFORE creating new file:
  search for existing files with similar purpose
  IF existing file found THEN
    modify existing file
  ELSE IF truly necessary THEN
    create new file
  END IF
END BEFORE
```

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
