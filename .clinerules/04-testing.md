# Testing Principles

> Based on Martin Fowler's testing guidance from [martinfowler.com](https://martinfowler.com)

## The Test Pyramid

**Core Principle:** "Write lots of small and fast unit tests. Write some more coarse-grained tests and very few high-level tests that test your application from end to end." (Mike Cohn)

### Pyramid Structure

```
        /\
       /E2E\      ← Very few: Full system validation
      /------\
     / Integ  \   ← Some: System interactions
    /----------\
   /   Unit     \  ← Lots: Fast, isolated component tests
  /--------------\
```

1. **Unit Tests (Base)** - Fast, isolated tests of individual components
2. **Integration Tests (Middle)** - Tests of system interactions with external dependencies
3. **End-to-End Tests (Top)** - Full system validation through user interfaces

**Reference:** [Practical Test Pyramid](https://martinfowler.com/articles/practical-test-pyramid.html)

## Key Testing Principles

### Test at the Right Level

- **Push tests down the pyramid whenever possible**
- Higher-level tests should only verify what lower-level tests cannot
- If a higher-level test spots an error with no lower-level test failing, write a lower-level test

### Avoid Test Duplication

> "If a higher-level test spots an error and there's no lower-level test failing, you need to write a lower-level test."

- Redundant tests slow development without adding confidence
- Each test should validate something unique
- Don't test the same thing at multiple levels

### Structure Tests Consistently

**Follow the "Arrange, Act, Assert" (AAA) pattern:**

```
// Arrange - Set up test data and conditions
const user = new User('test@example.com');

// Act - Execute the behavior being tested
const result = user.validate();

// Assert - Verify the outcome
expect(result).toBe(true);
```

### Prioritize Test Code Quality

- **Test code deserves the same care as production code**
- One focused assertion per test improves clarity
- Clear test names that describe what's being tested
- Keep tests simple and readable
- Refactor tests when they become unclear

### Speed and Feedback Loops

- **Fast feedback loops are essential**
- Run quick unit tests first
- Reserve slower integration and end-to-end tests for critical paths
- Tests must be fast enough to run constantly during development

## Unit Tests

**Definition:** "Unit tests are low-level, focusing on a small part of the software system."

### What Makes Unit Tests Effective

**Speed**
- The MOST critical factor
- Must run rapidly enough to execute constantly during coding
- If tests are slow, developers won't run them
- Should provide immediate feedback

**Independence**
- Isolate the code being tested
- Choose between:
  - **Solitary tests** - Use test doubles for all dependencies
  - **Sociable tests** - Use real collaborators when reasonable

**Focus**
- Test small portions of the system
- Typically individual classes or functions
- Clear scope: what constitutes a "unit" for your context

**Frequency**
- Run after every code change
- Catch bugs immediately when introduced
- Part of normal coding workflow

### Common Unit Test Pitfalls

- **Slowness** - Tests taking too long discourages frequent execution
- **Poor scope definitions** - Unclear what qualifies as a "unit"
- **Over-isolation** - Mocking everything creates brittle tests
- **Infrequent execution** - Not running tests means bugs go undetected

**Reference:** [Unit Tests](https://martinfowler.com/bliki/UnitTest.html)

## Test Doubles

**Definition:** "Test Double is a generic term for any case where you replace a production object for testing purposes."

### Types of Test Doubles

**Dummy Objects**
- Placeholders that satisfy parameters but are never used
- For irrelevant dependencies in the test scenario

**Fake Objects**
- Working implementations with shortcuts unsuitable for production
- Example: In-memory database instead of real database
- Have real behavior but simplified

**Stubs**
- Provide predetermined responses to specific calls
- Simple response control for test scenarios
- Ignore interactions outside programmed responses

**Spies**
- Record information about how they were invoked
- Track behavior without side effects
- Example: Email service that counts messages sent without actually sending

**Mocks**
- Pre-programmed with expectations about incoming calls
- Actively verify correct usage patterns
- Throw exceptions for unexpected interactions
- Confirm all anticipated calls occurred during verification

### Choosing the Right Test Double

- **Dummy** - When you need to satisfy a parameter but won't use it
- **Fake** - When you need working behavior with a simpler implementation
- **Stub** - When you need simple predetermined responses
- **Spy** - When you need to track how something was called
- **Mock** - When you need to verify interaction contracts

**Reference:** [Test Doubles](https://martinfowler.com/bliki/TestDouble.html)

## Test Coverage

### Coverage as a Tool, Not a Target

> "Test coverage is a diagnostic tool, not a quality metric."

**Key Principles:**

- Coverage identifies untested code areas
- Don't treat coverage as a quality metric
- Setting percentage thresholds incentivizes low-quality tests
- Focus on thoughtful testing, not hitting numbers

### What Good Coverage Looks Like

**Natural coverage in upper 80s-90s** through thoughtful testing practices

**Good test suites ensure:**
- Bugs rarely escape to production
- Developers change code confidently without fear

### Warning Signs

**Too Little Testing (< 50% coverage)**
- Frequent production bugs
- Fear of refactoring code
- Uncertainty about impact of changes

**Too Much Testing (Coverage-driven)**
- Simple code changes requiring excessive test modifications
- Test duplication slowing development
- Tests taking excessive time to run
- Tests written just to hit coverage numbers

### The Real Value

> "If a part of your test suite is weak in a way that coverage can detect, it's likely also weak in a way coverage can't detect." (Brian Marick)

Use coverage analysis to:
- Discover gaps worth investigating
- Understand which critical paths lack validation
- Guide where to focus testing efforts
- NOT to hit arbitrary percentage targets

**Reference:** [Test Coverage](https://martinfowler.com/bliki/TestCoverage.html)

## Testability Through Design

### Seams for Testing

"A seam is a place where you can alter behavior in your program without editing in that place"

**Enable testability by:**
- Using dependency injection
- Designing to interfaces/abstractions
- Avoiding hard-coded dependencies
- Creating extension points
- Test-driven development naturally produces appropriate seams

### Design Patterns That Improve Testability

- **Dependency Injection** - Inject dependencies rather than creating them
- **Strategy Pattern** - Swap algorithms for testing
- **Observer Pattern** - Verify notifications without side effects
- **Repository Pattern** - Replace data access with test doubles
- **Factory Pattern** - Control object creation in tests

## Testing Anti-Patterns to Avoid

- **Setting coverage percentage targets** - Creates bad incentives
- **Only writing E2E tests** - Inverted pyramid, slow feedback
- **Over-mocking** - Brittle tests that don't reflect reality
- **Test duplication** - Same behavior tested at multiple levels
- **Slow tests** - Discourages running them frequently
- **Testing implementation details** - Breaks when refactoring
- **Unclear test names** - Hard to understand what's being tested
- **Multiple assertions per test** - Unclear what failed
- **Shared test state** - Tests that affect each other
- **Ignoring failing tests** - Erodes confidence in suite
