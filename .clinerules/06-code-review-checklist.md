# Code Review Checklist

> Practical checklist based on Martin Fowler principles for reviewing code quality, design, and testability

## Design Quality (Beck's Four Rules)

### 1. Does it pass tests?
- [ ] All tests pass
- [ ] New tests added for new functionality
- [ ] Tests cover edge cases and error conditions
- [ ] Tests are meaningful, not just achieving coverage

### 2. Does it reveal intention?
- [ ] Clear, descriptive names for classes, functions, and variables
- [ ] Code is self-documenting through good naming
- [ ] Comments only where logic isn't self-evident
- [ ] Functions/methods do what their name suggests

### 3. Is there duplication?
- [ ] No copy-pasted code blocks
- [ ] Logic expressed "Once and Only Once"
- [ ] Similar code consolidated through abstraction
- [ ] BUT: Not abstracted prematurely (wait for third occurrence)

### 4. Are there unnecessary elements?
- [ ] No dead code or commented-out code
- [ ] No speculative features for hypothetical futures
- [ ] No unused variables, parameters, or imports
- [ ] No over-engineering or premature optimization

## Responsibilities and SOLID

### Single Responsibility
- [ ] Each class has one clear responsibility
- [ ] Each function does one thing well
- [ ] No "god objects" that know/do everything
- [ ] Changes to one concern don't affect unrelated code

### Open/Closed
- [ ] Uses polymorphism for type-based variations
- [ ] New features added through extension, not modification
- [ ] Appropriate abstractions for varying behavior

### Liskov Substitution
- [ ] Subtypes can replace base types without breaking behavior
- [ ] Polymorphic behavior is predictable
- [ ] No surprising exceptions or behavior changes in subclasses

### Interface Segregation
- [ ] Interfaces are focused and cohesive
- [ ] No "fat interfaces" forcing unused method implementations
- [ ] Clients don't depend on interfaces they don't use

### Dependency Inversion
- [ ] Depends on abstractions, not concrete implementations
- [ ] High-level modules independent of low-level details
- [ ] Dependencies injected, not created internally

## Object-Oriented Design

### Tell Don't Ask
- [ ] Objects are told what to do, not queried for decisions
- [ ] Behavior located with the data it operates on
- [ ] No excessive getter/setter chains
- [ ] Data and operations on that data are bundled together

### Encapsulation
- [ ] Internal state properly hidden
- [ ] Appropriate access modifiers used
- [ ] No direct field access from outside
- [ ] Behavior exposes intent, not implementation details

## Testability

### Test Coverage
- [ ] Critical paths have test coverage
- [ ] New code includes tests
- [ ] Coverage is thoughtful, not just hitting percentages
- [ ] Tests validate behavior, not just exercise code

### Test Quality
- [ ] Tests follow Arrange-Act-Assert pattern
- [ ] One focused assertion per test
- [ ] Clear test names describing what's being tested
- [ ] Tests are fast enough to run frequently
- [ ] No test duplication at multiple levels

### Testable Design
- [ ] Dependencies can be injected for testing
- [ ] Appropriate seams for test doubles
- [ ] No hard-coded dependencies or globals
- [ ] Side effects separated from pure logic

### Test Doubles Usage
- [ ] Appropriate test double type used (dummy, fake, stub, spy, mock)
- [ ] Not over-mocked (avoiding brittle tests)
- [ ] Mocks verify interactions when needed
- [ ] Stubs provide test data appropriately

## Code Smells

### Methods and Functions
- [ ] No long methods (check if it needs decomposition)
- [ ] No long parameter lists (consider parameter objects)
- [ ] No feature envy (method more interested in other class's data)
- [ ] Functions at single level of abstraction

### Classes and Modules
- [ ] No large classes with too many responsibilities
- [ ] No data classes (just fields, no behavior)
- [ ] No god classes that know/do everything
- [ ] Classes are cohesive

### Changes and Dependencies
- [ ] No divergent change (one class changing for multiple reasons)
- [ ] No shotgun surgery (one change affecting many classes)
- [ ] No inappropriate intimacy (classes too coupled)
- [ ] No circular dependencies

### Data and Types
- [ ] No primitive obsession (use value objects where appropriate)
- [ ] No magic numbers (use named constants)
- [ ] Appropriate data structures for the task

## Refactoring Quality

### Code Structure
- [ ] Broken into small, focused functions
- [ ] Concerns separated (calculation from formatting, etc.)
- [ ] Polymorphism used for type-based variations
- [ ] Appropriate design patterns applied

### Refactoring Safety
- [ ] Tests existed before refactoring
- [ ] Small, incremental changes
- [ ] Tests still pass after changes
- [ ] Commits show safe progression

## Architecture and Organization

### Module Organization
- [ ] Related code grouped together
- [ ] Clear module boundaries
- [ ] Dependencies flow in one direction (no cycles)
- [ ] Stable dependencies principle followed

### Bounded Contexts
- [ ] Clear context boundaries in large systems
- [ ] Each context has cohesive internal model
- [ ] Explicit integration points between contexts
- [ ] No forced unification of incompatible models

### Layering
- [ ] Appropriate separation of concerns across layers
- [ ] Business logic separate from presentation
- [ ] Infrastructure separate from domain
- [ ] Dependencies point inward (dependency inversion)

## Dependency Management

### Dependency Injection
- [ ] Dependencies injected via constructor or setter
- [ ] Required dependencies via constructor
- [ ] Optional dependencies via setter
- [ ] No service locator pattern in components

### Inversion of Control
- [ ] Components don't create their dependencies
- [ ] Configuration separated from usage
- [ ] Framework calls component, not vice versa

## General Code Quality

### Readability
- [ ] Code is easy to understand
- [ ] Consistent formatting and style
- [ ] Appropriate level of abstraction
- [ ] No clever tricks that obscure meaning

### Maintainability
- [ ] Easy to modify and extend
- [ ] Changes are localized
- [ ] No tight coupling
- [ ] Well-organized and structured

### Performance
- [ ] No premature optimization
- [ ] Efficient algorithms where it matters
- [ ] No obvious performance issues
- [ ] Performance trade-offs justified

## Anti-Patterns to Reject

- [ ] No test coverage targets driving low-quality tests
- [ ] No over-engineering for hypothetical futures
- [ ] No asking objects for data then making external decisions
- [ ] No creating dependencies inside components
- [ ] No inverted test pyramid (only E2E tests)
- [ ] No slow tests discouraging execution
- [ ] No test duplication without additional value
- [ ] No ignoring code smells without investigation
- [ ] No refactoring without tests
- [ ] No components dependent on framework locators

## Review Process

### Before Approving
- [ ] Code meets acceptance criteria
- [ ] All checklist items addressed or have justification
- [ ] Tests are comprehensive and passing
- [ ] No unaddressed code smells
- [ ] Architecture aligns with system design
- [ ] Documentation updated if needed

### Providing Feedback
- Reference specific code locations
- Explain why something is problematic
- Suggest concrete improvements
- Link to relevant principles/patterns
- Balance critical feedback with positive observations
- Focus on the code, not the person

## Context and Balance

Remember:
- These are guidelines, not absolute rules
- Apply principles pragmatically based on context
- Balance multiple concerns in design decisions
- Simple solutions are often the best solutions
- Perfect is the enemy of good
