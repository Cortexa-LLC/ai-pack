# Verification Gates

**Version:** 1.0.0
**Last Updated:** 2026-01-07

## Overview

Verification gates define quality checks and validation requirements that must be satisfied at different stages of the workflow. These gates ensure work meets standards before progressing.

## Verification Stages

Verification occurs at three critical stages:
1. **Pre-implementation** - Before starting work
2. **Mid-implementation** - During development
3. **Post-implementation** - Before completion

---

## Pre-Implementation Checks

**Purpose:** Ensure readiness before beginning work.

### 1. Requirements Clarity

**Checkpoint:** Are requirements clear and unambiguous?

**Verification:**
```
✓ Task objectives are well-defined
✓ Success criteria are measurable
✓ Constraints are understood
✓ Acceptance criteria are clear
✓ Ambiguities have been resolved
```

**Action if not met:**
```
IF requirements unclear THEN
  use AskUserQuestion tool
  clarify ambiguities
  document clarifications
  wait for confirmation
  THEN proceed
END IF
```

---

### 2. Context Understanding

**Checkpoint:** Do we understand the codebase context?

**Verification:**
```
✓ Relevant code has been read
✓ Existing patterns identified
✓ Dependencies understood
✓ Architecture comprehended
✓ Test coverage assessed
```

**Required Activities:**
```
BEFORE implementation:
  read relevant files
  identify patterns to follow
  understand dependencies
  review existing tests
  check architecture alignment
END BEFORE
```

---

### 3. Approach Validation

**Checkpoint:** Is the implementation approach sound?

**Verification:**
```
✓ Approach aligns with requirements
✓ Follows existing patterns
✓ Considers edge cases
✓ Has testing strategy
✓ Has rollback plan
```

**For non-trivial tasks:**
```
IF task is non-trivial THEN
  enter plan mode
  create implementation plan
  present plan to user
  get approval
  THEN implement
END IF
```

---

## Mid-Implementation Validation

**Purpose:** Ensure quality during development.

### 4. Incremental Testing

**Checkpoint:** Tests pass at each step.

**Verification:**
```
AFTER each significant change:
  run relevant tests
  ✓ All tests pass
  ✓ No new failures introduced
  ✓ Test coverage maintained
  IF tests fail THEN
    fix immediately
    don't proceed until green
  END IF
END AFTER
```

**TDD Cycle:**
```
1. Write failing test (RED)
2. Write minimal code to pass (GREEN)
3. Refactor while keeping tests green (REFACTOR)
4. Repeat
```

---

### 5. Standards Compliance

**Checkpoint:** Code follows established standards.

**Verification:**
```
✓ Formatting correct (spaces, not tabs)
✓ Naming conventions followed
✓ Language-specific guidelines met
✓ No code smells introduced
✓ Design principles upheld
```

**Continuous Checks:**
```
DURING development:
  follow formatting standards
  use consistent naming
  apply SOLID principles
  avoid code smells
  maintain simplicity (YAGNI)
END DURING
```

---

### 6. Pattern Consistency

**Checkpoint:** New code follows existing patterns.

**Verification:**
```
✓ Error handling matches existing code
✓ Logging follows project conventions
✓ API design consistent
✓ Test structure similar
✓ Documentation style maintained
```

**Pattern Matching:**
```
WHEN writing new code:
  identify similar existing code
  follow established patterns
  maintain consistency
  IF deviation necessary THEN
    document rationale
    get user approval
  END IF
END WHEN
```

---

## Post-Implementation Verification

**Purpose:** Confirm work is complete and meets requirements.

### 7. Comprehensive Testing

**Checkpoint:** All tests pass, coverage meets target.

**Verification:**
```
✓ All unit tests pass
✓ All integration tests pass
✓ All acceptance tests pass
✓ Code coverage: 80-90%
✓ No flaky tests
✓ Test execution time reasonable
```

**Test Execution:**
```
RUN full test suite:
  unit tests
  integration tests
  acceptance tests

VERIFY results:
  ✓ 0 failures (zero tolerance)
  ✓ Coverage in target range
  ✓ No skipped tests (unless documented)

IF any test fails THEN
  STOP: do not proceed
  fix failing tests
  re-run full suite
  THEN continue
END IF
```

---

### 8. Build Verification (BLOCKING - ZERO WARNINGS REQUIRED)

**Checkpoint:** Project builds successfully with **ZERO WARNINGS**.

**⚠️ CRITICAL: Code MUST compile/build with ZERO WARNINGS. Warnings are treated as errors.**

**Verification:**
```
✓ Clean build succeeds
✓ No compilation errors
✓ No compilation warnings (BLOCKING)
✓ No linking errors
✓ No linting errors
✓ No type errors
✓ No style violations
```

**MANDATORY Build Process:**
```
RUN build WITH WARNINGS AS ERRORS:

  # C/C++ Projects
  IF C++ project THEN
    # Treat all warnings as errors
    cmake -DCMAKE_CXX_FLAGS="-Werror -Wall -Wextra" ..
    make
    # OR
    g++ -Werror -Wall -Wextra *.cpp
    ✓ MUST show "0 warnings"
  END IF

  # C# Projects
  IF C# project THEN
    # Treat warnings as errors (MANDATORY)
    dotnet build /warnaserror
    # OR verify in .csproj:
    # <TreatWarningsAsErrors>true</TreatWarningsAsErrors>
    ✓ MUST show "0 Warning(s)"
  END IF

  # Java Projects
  IF Java project THEN
    # Maven: treat warnings as errors
    mvn clean compile -Dmaven.compiler.showWarnings=true -Werror
    # Gradle: enable warnings as errors
    ./gradlew build -Pwarnings-as-errors
    ✓ MUST show "BUILD SUCCESS" with 0 warnings
  END IF

  # JavaScript/TypeScript Projects
  IF JS/TS project THEN
    # TypeScript: strict mode
    tsc --noEmitOnError --strict
    # ESLint: fail on warnings
    eslint . --max-warnings 0
    ✓ MUST show "0 problems"
  END IF

  # Python Projects
  IF Python project THEN
    # Flake8: fail on any issues
    flake8 . --count --show-source --statistics
    # mypy: strict type checking
    mypy . --strict --warn-unreachable
    ✓ MUST show "0 errors", "0 warnings"
  END IF

  # Go Projects
  IF Go project THEN
    # Go: vet checks
    go vet ./...
    # golangci-lint: strict
    golangci-lint run --max-issues-per-linter 0 --max-same-issues 0
    ✓ MUST show "0 issues"
  END IF

  # Rust Projects
  IF Rust project THEN
    # Cargo: deny warnings
    cargo build --release
    cargo clippy -- -D warnings
    ✓ MUST show "0 warnings"
  END IF

END RUN

**ENFORCEMENT:**
```
IF any warnings detected THEN
  STOP immediately
  BLOCK completion
  LIST all warnings
  FIX all warnings
  RE-RUN build
  ONLY proceed when: 0 warnings
END IF
```

**Pre-Commit Checklist (MANDATORY):**
```
BEFORE committing code:
  ✓ Run build with warnings-as-errors flag
  ✓ Verify output shows "0 warnings"
  ✓ Run linter/formatter
  ✓ Fix ALL issues
  ✓ Re-run build to confirm
  ✓ THEN commit
```

---

### 9. Requirement Satisfaction

**Checkpoint:** All requirements have been met.

**Verification:**
```
FOR each requirement:
  ✓ Implemented as specified
  ✓ Tests verify behavior
  ✓ Edge cases handled
  ✓ Error cases handled
  ✓ Documented if needed
END FOR
```

**Acceptance Criteria Check:**
```
REVIEW acceptance criteria from contract:
  ✓ All criteria met
  ✓ Success metrics achieved
  ✓ Constraints respected
  ✓ User expectations satisfied

IF criteria not met THEN
  identify gaps
  implement missing pieces
  re-verify
END IF
```

---

### 10. Integration Verification

**Checkpoint:** Changes integrate correctly with existing system.

**Verification:**
```
✓ No breaking changes to public APIs
✓ Backward compatibility maintained (if required)
✓ Dependencies updated correctly
✓ Configuration changes documented
✓ Migration path provided (if needed)
```

**Integration Tests:**
```
RUN integration checks:
  test with existing components
  verify no regressions
  check performance impact
  validate error handling

IF integration issues THEN
  fix integration problems
  update dependent code
  re-test integration
END IF
```

---

## Documentation Requirements

### 11. Code Documentation

**Checkpoint:** Code is adequately documented.

**Verification:**
```
✓ Public APIs documented
✓ Complex logic explained
✓ Non-obvious decisions documented
✓ Examples provided (if helpful)
✓ Comments accurate and current
```

**Documentation Standards:**
```
Document:
  ✅ Public interfaces and APIs
  ✅ Complex algorithms
  ✅ Non-obvious design decisions
  ✅ Deviation from standards (with rationale)

Don't document:
  ❌ Obvious code
  ❌ Self-explanatory functions
  ❌ Code that doesn't need explanation
```

---

### 12. Change Documentation

**Checkpoint:** Changes are properly documented.

**Verification:**
```
✓ Commit messages clear and descriptive
✓ Work log updated (in .ai/tasks/)
✓ Breaking changes documented
✓ Migration guide provided (if needed)
✓ Changelog updated (if project has one)
```

**Commit Message Quality:**
```
Commit should include:
- Type (feat/fix/refactor/test/docs)
- Short summary (50 chars max)
- Detailed description (if needed)
- References to issues (if applicable)
- Co-authored-by attribution
```

---

## Task State Management

### 13. Task Packet Updates

**Checkpoint:** Task packet in `.ai/tasks/` is current.

**Verification:**
```
✓ Brief documented (task.md)
✓ Result written (result.md)
```

**Update Requirements:**
```
DURING task execution:
  initialize task packet
  update work log incrementally
  document decisions
  track issues and resolutions

UPON completion:
  finalize work log
  conduct review
  complete acceptance checklist
END DURING
```

---

## Specialized Verifications

### For Refactoring

**Additional Checks:**
```
✓ Behavior unchanged (tests prove it)
✓ Test suite unchanged (unless improving tests)
✓ Performance not degraded
✓ All tests still pass
✓ Code quality improved
```

### For Bug Fixes

**Additional Checks:**
```
✓ Bug reproduced first
✓ Test added for bug
✓ Test fails before fix
✓ Test passes after fix
✓ Root cause understood and documented
✓ Similar bugs checked for
```

### For New Features

**Additional Checks:**
```
✓ Feature complete (not partial)
✓ User documentation updated
✓ Tests cover happy path + edge cases
✓ Error handling comprehensive
✓ Performance acceptable
✓ Security considered
```

### For Performance Optimization

**Additional Checks:**
```
✓ Baseline metrics captured
✓ Improvement measured
✓ No functionality broken
✓ Trade-offs understood
✓ Performance gains documented
```

---

## Rollback Triggers

**When to rollback:**

```
ROLLBACK if:
  ❌ Tests fail and can't be fixed quickly
  ❌ Build breaks
  ❌ Performance degrades significantly
  ❌ Security vulnerability introduced
  ❌ Breaking change not approved
  ❌ Requirements misunderstood
```

**Rollback Process:**
```
1. Stop current work
2. Revert changes (git revert/reset)
3. Verify system restored
4. Analyze what went wrong
5. Revise approach
6. Get user approval
7. Re-attempt correctly
```

---

## Quality Metrics

### Target Metrics

```
Test Coverage:     80-90%
Test Pass Rate:    100% (zero tolerance for failures)
Build Success:     100%
Code Review:       No critical findings
Standards:         100% compliance
Documentation:     All public APIs documented
```

### Measurement

```
MEASURE quality:
  run coverage tool → verify 80-90%
  run test suite → verify 100% pass
  run linter → verify 0 errors
  run security scan → verify no criticals
  review code → verify standards met
END MEASURE
```

---

## Integration

Verification gates integrate with:
- **[Global Gates](00-global-gates.md)** - Overall quality requirements
- **[Persistence Gates](10-persistence.md)** - File operation validation
- **[Tool Policy](20-tool-policy.md)** - Tool result verification
- **[Workflows](../workflows/)** - Stage-specific verification
- **[Task Packets](../templates/task-packet/)** - Progress tracking

---

## Continuous Verification

**Philosophy:** Verify continuously, not just at end.

```
Pre-implementation:  Verify readiness
During implementation: Verify incrementally
Post-implementation: Verify completion
Before user handoff:  Verify acceptance criteria
```

**Benefits:**
- Catch issues early
- Maintain quality throughout
- Reduce rework
- Build confidence incrementally
- Enable safe, incremental progress

---

**Last reviewed:** 2026-01-07
**Next review:** Quarterly or when verification processes need refinement
