# Refactoring Principles

> Based on Martin Fowler's refactoring guidance from [martinfowler.com](https://martinfowler.com)

## Understanding Code Smells

**Definition:** "A code smell is a surface indication that usually corresponds to a deeper problem in the system." (Kent Beck)

### Key Characteristics

- **Easy to spot** - Quick to identify through visual inspection
- **Not always problematic** - Smells are warning signs, not definitive problems
- **Prompt investigation** - Use as indicators to look deeper, not triggers for automatic refactoring

### Using Smells Effectively

> "The most useful smells are easily detectable and frequently point toward genuine improvements."

**Reference:** [Code Smells](https://martinfowler.com/bliki/CodeSmell.html)

## Common Code Smells Catalog

### Long Method/Function

**Description:** Methods/functions that are too long

**Why it matters:**
- Quick to identify visually
- Often indicates multiple responsibilities
- Makes code harder to understand and test

**Solutions:**
- Extract Method/Function
- Decompose into smaller, focused functions
- Each function should do one thing well

### Data Class

**Description:** Classes with only fields and no behavior (getters/setters only)

**Why it matters:**
- Violates object-oriented principle of bundling data with behavior
- Behavior that operates on this data likely belongs in the class

**Solutions:**
- Move methods that use the data into the data class
- Apply Tell Don't Ask principle
- Add behavior alongside the data

### Large Class

**Description:** Classes with too many responsibilities, fields, or methods

**Why it matters:**
- Violates Single Responsibility Principle
- Multiple reasons to change
- Difficult to understand and maintain

**Solutions:**
- Extract Class for cohesive subsets of behavior
- Identify distinct responsibilities and split them
- Look for groups of fields/methods that belong together

### Long Parameter List

**Description:** Functions with too many parameters

**Why it matters:**
- Hard to understand what each parameter means
- Difficult to call correctly
- Often indicates missing abstraction

**Solutions:**
- Introduce Parameter Object
- Pass entire object instead of individual fields
- Use Builder pattern for complex construction

### Divergent Change

**Description:** One class changes for multiple different reasons

**Why it matters:**
- Violates Single Responsibility Principle
- Different concerns tangled together
- Changes to one concern affect unrelated functionality

**Solutions:**
- Split class by responsibility
- Extract classes for each reason to change
- Identify and separate distinct concerns

### Shotgun Surgery

**Description:** One change requires modifications across many classes

**Why it matters:**
- Related behavior is scattered
- Hard to find all places that need changing
- Easy to miss necessary updates

**Solutions:**
- Move Method/Field to consolidate behavior
- Inline Class if behavior is too fragmented
- Create new class to centralize related behavior

### Feature Envy

**Description:** Method more interested in another class's data than its own

**Why it matters:**
- Behavior separated from data it operates on
- Violates encapsulation
- Indicates wrong responsibility placement

**Solutions:**
- Move Method closer to the data it uses
- Apply Tell Don't Ask principle
- Consider if behavior belongs in the other class

### Primitive Obsession

**Description:** Overuse of primitives instead of small objects

**Why it matters:**
- Loses type safety and semantic meaning
- Behavior scattered across codebase
- Validation logic duplicated

**Solutions:**
- Introduce Value Objects
- Replace primitive with object (e.g., Money, PhoneNumber, EmailAddress)
- Encapsulate validation and behavior in the object

## Refactoring Flow

Martin Fowler describes a systematic approach to refactoring:

1. **Break into functions** - Decompose large methods into smaller ones
2. **Separate concerns** - Separate calculation from formatting, business logic from presentation
3. **Organize by type** - Use polymorphism for variant behavior
4. **Apply patterns** - Progress from decomposition → separation → object-oriented patterns

## When to Refactor

### The Rule of Three

First time: Just do it
Second time: Wince and duplicate
Third time: Refactor

### Specific Triggers

**Before adding features**
- Clean up the area where you'll be working
- Makes the new feature easier to add
- Prevents piling new code on top of messy code

**When you encounter code smells**
- Investigate whether they indicate real problems
- Refactor if they point to design issues

**When tests are hard to write**
- Design for testability
- Add seams for dependency injection
- Separate concerns to isolate what you're testing

**When making changes feels dangerous**
- Add tests first to establish safety net
- Then refactor to make changes safer
- Small, safe steps with test coverage

## Refactoring Safety

### Prerequisites

1. **Have tests** - Comprehensive test coverage before refactoring
2. **Small steps** - Make tiny changes and test after each
3. **Commit often** - Easy to revert if something goes wrong

### Process

1. Ensure tests pass before starting
2. Make one small refactoring
3. Run tests
4. Commit if tests pass
5. Repeat

## Anti-Patterns to Avoid

- **Refactoring without tests** - Dangerous; add tests first
- **Big bang refactoring** - Small incremental changes are safer
- **Ignoring code smells** - They usually indicate real problems
- **Premature abstraction** - Wait for duplication before abstracting
- **Over-engineering** - Simplest solution that works is usually best

## Educational Approach

> "An effective teaching approach involves selecting individual smells periodically and having teams identify and discuss them collaboratively. This gradual process helps develop programming expertise across teams."

- Pick one smell at a time to focus on
- Have team discussions about examples
- Build shared understanding gradually
- Even junior developers can spot smells before fully understanding problems
