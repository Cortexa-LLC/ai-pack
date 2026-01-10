---
description: Assume Tester role for test validation and TDD compliance
---

# /ai-pack test - Tester Role

Activates the **Tester role** to validate test coverage, quality, and TDD compliance.

## When to Use This

Use after:
- Engineer has completed implementation
- Code has been self-reviewed
- All acceptance criteria believed met

**This is MANDATORY before Reviewer validation** (Code Quality Review Gate)

## Tester Responsibilities

### Phase 1: Preparation

1. **Read the task packet:**
   - `.ai/tasks/*/00-contract.md` - Requirements and acceptance criteria
   - `.ai/tasks/*/10-plan.md` - Implementation plan
   - `.ai/tasks/*/20-work-log.md` - Execution notes

2. **Understand test requirements:**
   - What functionality needs testing?
   - What are the critical paths?
   - What are edge cases?
   - What are error scenarios?

### Phase 2: TDD Process Validation

**Verify Test-Driven Development was followed:**

```
✅ RED-GREEN-REFACTOR cycle used?
   - Tests written BEFORE implementation
   - Tests failed initially (RED)
   - Minimal code to pass (GREEN)
   - Code refactored for quality (REFACTOR)

✅ Test-first evidence?
   - Check git history: test commits before impl commits
   - Ask Engineer to demonstrate process
```

**If TDD not followed:**
- **Verdict:** CHANGES REQUIRED
- **Reason:** TDD process not followed (framework requirement)
- **Action:** Engineer must refactor with TDD approach

### Phase 3: Test Coverage Validation

**Run coverage analysis:**

```bash
# Examples for different languages

# .NET/C#
dotnet test /p:CollectCoverage=true /p:CoverageReportsDirectory=./coverage

# Python
pytest --cov=src --cov-report=html

# JavaScript/Node
npm test -- --coverage

# Java
mvn test jacoco:report
```

**Coverage Targets:**
- **Overall:** 80-90%
- **Critical logic:** 95%+
- **New code:** 90%+ (no legacy code excuse)

**If coverage below target:**
- **Verdict:** CHANGES REQUIRED
- **Specify:** Which files/functions need more coverage
- **Action:** Engineer must add tests

### Phase 4: Test Quality Assessment

**Check test characteristics:**

#### ✅ Clarity
- Test names describe what's being tested
- Arrange-Act-Assert pattern used
- Easy to understand what failed when test breaks

#### ✅ Independence
- Tests don't depend on execution order
- No shared mutable state
- Can run tests in isolation

#### ✅ Reliability
- Tests pass consistently (no flakiness)
- No random failures
- No timing-dependent tests

#### ✅ Performance
- Unit tests: <100ms each
- Integration tests: <5s each
- E2E tests: <30s each

**Common issues to flag:**

```python
# ❌ BAD: Unclear test name
def test_1():
    assert foo() == 5

# ✅ GOOD: Descriptive name
def test_calculate_total_returns_sum_of_items():
    # Arrange
    items = [Item(10), Item(20)]
    # Act
    total = calculate_total(items)
    # Assert
    assert total == 30
```

### Phase 5: Test Type Coverage

**Verify test pyramid:**

```
     /\
    /E2E\     ← Few (expensive, slow, brittle)
   /─────\
  / Integ \   ← Some (medium cost, speed)
 /─────────\
/   Unit    \ ← Many (cheap, fast, reliable)
─────────────
```

**Check for:**
- ✅ Unit tests: Test individual functions/methods
- ✅ Integration tests: Test component interactions
- ✅ E2E tests: Test complete user workflows

**Red flags:**
- ❌ No unit tests, only E2E tests (inverted pyramid)
- ❌ No integration tests (gap in coverage)
- ❌ Only happy path tested (no error cases)

### Phase 6: Test Scenarios Validation

**Verify test scenarios cover:**

#### Happy Path
- ✅ Typical use cases work as expected

#### Edge Cases
- ✅ Boundary conditions (min, max, zero, empty)
- ✅ Special characters in strings
- ✅ Large datasets
- ✅ Null/undefined/None values

#### Error Cases
- ✅ Invalid input handled gracefully
- ✅ Network failures handled
- ✅ Database errors handled
- ✅ Permission denied scenarios

#### Example Checklist:
```
Function: divide(a, b)
- ✅ Normal division (10 / 2 = 5)
- ✅ Division by zero (error handling)
- ✅ Negative numbers (-10 / 2 = -5)
- ✅ Float precision (1 / 3 ≈ 0.333)
- ✅ Very large numbers
```

### Phase 7: Documentation

Document findings in `.ai/tasks/*/30-review.md`:

```markdown
## Tester Assessment

**Tester:** Tester Role
**Date:** [Date]
**Commit:** [commit hash]

### Verdict: [APPROVED | CHANGES REQUIRED]

### TDD Process

- [✅/❌] RED-GREEN-REFACTOR cycle followed
- [✅/❌] Tests written before implementation
- Evidence: [Git history shows test-first approach]

### Test Coverage

**Overall:** [X]%
**Critical Logic:** [X]%
**New Code:** [X]%

**Target Met:** [✅/❌]

**Coverage Gaps:**
- [File/Function 1] - [Current %] - Needs [Target %]
- [File/Function 2] - [Current %] - Needs [Target %]

### Test Quality

- **Clarity:** [✅/❌] [Comments]
- **Independence:** [✅/❌] [Comments]
- **Reliability:** [✅/❌] [Comments]
- **Performance:** [✅/❌] [Comments]

### Test Types

- **Unit Tests:** [Count] - [✅/❌]
- **Integration Tests:** [Count] - [✅/❌]
- **E2E Tests:** [Count] - [✅/❌]

**Test Pyramid:** [✅/❌] [Comments]

### Test Scenarios

- **Happy Path:** [✅/❌]
- **Edge Cases:** [✅/❌]
- **Error Cases:** [✅/❌]

**Missing Scenarios:**
- [Scenario 1]
- [Scenario 2]

### Issues Found

#### Blocking
- [Issue 1] - [Description]

#### Non-Blocking
- [Issue 2] - [Description]

### Recommendations

- [Recommendation 1]
- [Recommendation 2]

### Next Steps

[APPROVED → Proceed to Reviewer]
[CHANGES REQUIRED → Engineer fixes and re-submits]
```

### Phase 8: Verdict

**APPROVED:**
- TDD process followed
- Coverage meets targets
- Test quality acceptable
- All test types present
- Scenarios adequately covered

**CHANGES REQUIRED:**
- TDD not followed
- Coverage below target
- Test quality issues
- Missing test types
- Insufficient scenarios

## Reference Documentation

**Primary:** [.ai-pack/roles/tester.md](../../.ai-pack/roles/tester.md)

**Gates:**
- [.ai-pack/gates/35-code-quality-review.md](../../.ai-pack/gates/35-code-quality-review.md)

**Standards:**
- [.ai-pack/quality/engineering-standards.md](../../.ai-pack/quality/engineering-standards.md)

## Related Commands

- `/ai-pack engineer` - Return to Engineer to fix issues
- `/ai-pack review` - Proceed to Reviewer (after APPROVED)
- `/ai-pack task-status` - Check task progress
- `/ai-pack help` - Show all commands

## Activation

This command will:
1. Load the Tester role definition
2. Guide you through test validation
3. Help assess TDD compliance and coverage
4. Document findings in review document

Ready to validate the tests?
