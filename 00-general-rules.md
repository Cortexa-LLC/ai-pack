# General Rules

**Universal standards that apply to all code regardless of language or project**

These foundational rules establish baseline practices for all development work. They take precedence over language-specific conventions when conflicts arise.

## Formatting Standards

### No Tabs - Always Use Spaces

**Rule:** Never use tab characters for indentation. Always use spaces.

**Rationale:**
- Tabs render differently across editors, terminals, and environments
- Spaces provide consistent visual appearance everywhere
- Mixed tabs and spaces create alignment nightmares
- Most modern IDEs can be configured to insert spaces when Tab key is pressed
- Industry standard across all major style guides

**Implementation:**
- Configure your editor to insert spaces when Tab key is pressed
- Set indentation width per language standards:
  - C++, JavaScript/TypeScript: 2 spaces
  - Python, Java, Kotlin: 4 spaces
- Use `.editorconfig` to enforce across team
- Run linters to catch tab characters in code reviews

**Example `.editorconfig`:**
```ini
[*]
indent_style = space
indent_size = 2

[*.{py,java,kt}]
indent_size = 4
```

## Development Workflow

### Test-Driven Development (TDD)

**Rule:** Always use Test-Driven Development when generating new code.

**The TDD Cycle:**
1. **Red** - Write a failing test that defines desired behavior
2. **Green** - Write minimal code to make the test pass
3. **Refactor** - Improve code quality while keeping tests green

**Benefits:**
- Forces thinking about design and interfaces before implementation
- Ensures code is testable by design
- Provides immediate feedback on code correctness
- Creates comprehensive test suite as byproduct
- Reduces debugging time
- Enables confident refactoring

**When to Apply TDD:**
- New features and functionality
- Bug fixes (write test that reproduces bug first)
- Refactoring existing code (add tests first if missing)
- API and interface design

### Test Coverage Guidelines

**Rule:** Aim for test coverage in the upper 80-90% range, understanding the law of diminishing returns.

**Key Principles:**

> "Test coverage is a useful tool for finding untested parts of a codebase. Test coverage is of little use as a numeric statement of how good your tests are." - Martin Fowler

**Target: Upper 80-90% through thoughtful testing**
- This range naturally emerges from comprehensive TDD practices
- NOT a hard target to game with low-quality tests
- Focus on meaningful test quality, not hitting numbers

**The Law of Diminishing Returns:**
- Pursuing 100% coverage often wastes effort on trivial code
- Some code paths are not worth testing (simple getters, framework glue)
- Time spent on last 10% often better spent on other quality measures
- Suspiciously perfect 100% suggests "teaching to the test"

**Good Test Suites Ensure:**
- Bugs rarely escape to production
- Developers refactor confidently without fear
- Critical business logic has thorough coverage
- Edge cases and error conditions are validated

**Warning Signs:**

*Too Little Coverage (< 50%):*
- Frequent production bugs
- Fear of making changes
- Brittle code that breaks unexpectedly

*Coverage-Driven Testing (gaming metrics):*
- Tests with no assertions
- Tests that don't validate behavior
- Over-testing trivial code
- Under-testing complex logic

**What to Test Heavily:**
- Core business logic
- Complex algorithms
- Edge cases and boundary conditions
- Error handling paths
- Public APIs and interfaces
- Code that changes frequently

**What Not to Over-Test:**
- Simple property getters/setters
- Framework boilerplate
- Third-party library wrappers (test your usage, not the library)
- Generated code
- Trivial utility functions

**Use Coverage Analysis To:**
- Identify untested critical paths
- Find gaps worth investigating
- Guide where to focus testing efforts
- NOT as a quality gate or target metric

**Reference:** [Test Coverage - Martin Fowler](https://martinfowler.com/bliki/TestCoverage.html)

## Summary

These general rules establish the foundation for all development:
1. **No tabs, always spaces** - Consistent formatting everywhere
2. **TDD for all new code** - Design through tests, quality by default
3. **Target 80-90% coverage naturally** - Comprehensive testing without gaming metrics

Apply these principles universally, then layer on language-specific guidelines as appropriate.
