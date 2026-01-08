# Refactor Workflow

**Version:** 1.0.0
**Last Updated:** 2026-01-07

## Overview

The Refactor Workflow is specialized for improving code structure without changing its external behavior. It emphasizes safety, incremental changes, and continuous verification.

**Extends:** [Standard Workflow](standard.md)

**Use for:** Improving code quality, eliminating code smells, restructuring code, applying design patterns, reducing technical debt.

---

## Key Principles

**The Golden Rule of Refactoring:**
```
Refactoring changes the internal structure of code
without changing its external behavior.

If behavior changes → It's not refactoring
If structure changes AND behavior preserved → It's refactoring
```

---

## Key Differences from Standard Workflow

1. **Comprehensive Tests First** - Must have strong test coverage before refactoring
2. **Behavior Preservation** - External behavior must not change
3. **Small Steps** - Tiny, verified increments
4. **Continuous Testing** - Tests must pass after every change
5. **Performance Monitoring** - Ensure no performance degradation

---

## Refactoring-Specific Phases

### Phase 1: Assessment & Test Coverage

**Objective:** Ensure safe refactoring environment.

#### 1.1 Identify Refactoring Triggers
```
Code Smells:
□ Duplicated code
□ Long methods (>20 lines)
□ Large classes (too many responsibilities)
□ Long parameter lists
□ Divergent change (class changes for many reasons)
□ Shotgun surgery (change requires many small changes)
□ Feature envy (method uses another class's data)
□ Data clumps (same group of data everywhere)
□ Primitive obsession (primitives instead of objects)
□ Switch statements (type checking)
□ Lazy class (class doing too little)
□ Speculative generality (unused abstraction)
□ Temporary field (field only used sometimes)
□ Message chains (a.getB().getC().getD())
□ Middle man (class only delegating)
□ Inappropriate intimacy (classes too coupled)
□ Alternative classes with different interfaces
□ Incomplete library class
□ Data class (only getters/setters)
□ Refused bequest (subclass doesn't use parent)
□ Comments (explaining complex code)
```

---

#### 1.2 Establish Safety Net
```
CRITICAL: Before any refactoring

✓ Comprehensive test coverage (90%+)
✓ All tests passing
✓ Tests verify behavior, not implementation
✓ Integration tests cover key paths
✓ Performance baseline captured

IF tests insufficient THEN
  STOP: Add tests first
  Do NOT refactor without safety net
END IF
```

**Test Coverage Verification:**
```bash
# Check coverage
npm test -- --coverage
pytest --cov=src tests/
cargo testy --coverage

# Requirement: 90%+ for code being refactored
```

---

#### 1.3 Define Refactoring Scope
```
□ What code needs refactoring?
□ What is the target structure?
□ What design patterns apply?
□ What are the benefits?
□ What are the risks?
□ How long will it take?
```

---

### Phase 2: Refactoring Strategy

**Objective:** Plan incremental, safe refactoring steps.

#### 2.1 Select Refactoring Techniques

**Common Refactorings:**

**Extract Method:**
```javascript
// Before:
function processOrder(order) {
  // validate order
  if (!order.items || order.items.length === 0) {
    throw new Error('Empty order');
  }
  if (!order.customerId) {
    throw new Error('No customer');
  }

  // calculate total
  let total = 0;
  for (const item of order.items) {
    total += item.price * item.quantity;
  }

  // apply discount
  if (order.discountCode) {
    total *= 0.9;
  }

  return total;
}

// After:
function processOrder(order) {
  validateOrder(order);
  const subtotal = calculateSubtotal(order.items);
  return applyDiscount(subtotal, order.discountCode);
}

function validateOrder(order) { ... }
function calculateSubtotal(items) { ... }
function applyDiscount(total, discountCode) { ... }
```

**Extract Class:**
```javascript
// Before: Person class with address fields
class Person {
  name;
  streetAddress;
  city;
  state;
  zipCode;
}

// After: Separate Address class
class Person {
  name;
  address;  // Address object
}

class Address {
  streetAddress;
  city;
  state;
  zipCode;
}
```

**Replace Conditional with Polymorphism:**
```javascript
// Before:
function getSpeed(vehicle) {
  switch (vehicle.type) {
    case 'car': return vehicle.enginePower * 2;
    case 'bike': return vehicle.pedalPower * 3;
    case 'plane': return vehicle.enginePower * 10;
  }
}

// After:
class Car {
  getSpeed() { return this.enginePower * 2; }
}
class Bike {
  getSpeed() { return this.pedalPower * 3; }
}
class Plane {
  getSpeed() { return this.enginePower * 10; }
}
```

---

#### 2.2 Plan Incremental Steps
```
Break refactoring into smallest possible steps:

Example: Extract Method refactoring
1. Identify code to extract
2. Create new method with temp name
3. Copy code to new method
4. Identify parameters needed
5. Replace original code with method call
6. Run tests → verify passing
7. Rename method appropriately
8. Run tests → verify passing
9. Remove duplicated code if any
10. Run tests → verify passing
```

**Incremental Plan Template:**
```
Refactoring: [Name]
Trigger: [Code smell being addressed]

Steps:
1. [Smallest change possible]
   Test: Run full test suite
   Expected: All tests pass, no behavior change

2. [Next small change]
   Test: Run full test suite
   Expected: All tests pass, no behavior change

3. [Continue...]
```

---

### Phase 3: Incremental Refactoring

**Objective:** Refactor safely in small, verified steps.

#### 3.1 The Refactoring Cycle
```
FOR each refactoring step:
  1. Make one small change
  2. Run ALL tests
  3. Verify tests pass
  4. IF tests fail THEN
       revert change
       understand why
       retry with smaller step
     END IF
  5. Commit (optional, at logical checkpoints)
  6. Proceed to next step
END FOR

CRITICAL: Tests must pass after EVERY step
```

---

#### 3.2 Refactoring Rules

**DO:**
```
✅ Make tiny changes
✅ Run tests after every change
✅ Revert if tests fail
✅ Commit at checkpoints
✅ Refactor OR add features (not both)
✅ Maintain performance
```

**DON'T:**
```
❌ Change behavior
❌ Make multiple changes at once
❌ Skip running tests
❌ Continue if tests fail
❌ Refactor and add features simultaneously
❌ Refactor without tests
```

---

#### 3.3 Handling Test Failures
```
IF tests fail during refactoring THEN
  Options:
  1. Revert change (preferred)
  2. Fix the mistake immediately
  3. If test is implementation-dependent:
     - Update test to verify behavior not implementation
     - But ONLY if behavior unchanged

  NEVER:
  - Delete failing tests
  - Comment out failing tests
  - Proceed with failing tests
END IF
```

---

### Phase 4: Verification & Cleanup

**Objective:** Confirm refactoring success and clean up.

#### 4.1 Verification Checklist
```
□ All tests passing
□ Code coverage maintained or improved
□ No behavior changes
□ Performance not degraded
□ Code simpler and cleaner
□ Design patterns properly applied
□ No new code smells introduced
```

#### 4.2 Performance Verification
```
BEFORE refactoring:
  Capture performance baseline

AFTER refactoring:
  Run performance tests
  Compare with baseline
  IF significant degradation THEN
    investigate cause
    optimize OR
    reconsider refactoring
  END IF
```

#### 4.3 Code Review
```
Refactoring Review Focus:
□ Behavior truly unchanged?
□ Tests still meaningful?
□ Code quality improved?
□ Abstractions appropriate?
□ No over-engineering?
□ Documentation updated?
```

---

## Common Refactoring Patterns

### 1. Extract Method
**When:** Method too long or doing too much
**Safety:** HIGH (easily reversible)

### 2. Rename Variable/Method
**When:** Name unclear or misleading
**Safety:** HIGH (with good tests)

### 3. Extract Class
**When:** Class has too many responsibilities
**Safety:** MEDIUM (affects dependencies)

### 4. Move Method
**When:** Method in wrong class
**Safety:** MEDIUM (affects callers)

### 5. Replace Conditional with Polymorphism
**When:** Type checking or complex switches
**Safety:** MEDIUM (structural change)

### 6. Introduce Parameter Object
**When:** Long parameter lists
**Safety:** LOW (affects all callers)

### 7. Replace Magic Numbers with Constants
**When:** Unexplained numbers in code
**Safety:** HIGH (simple change)

---

## Refactoring Anti-Patterns

### Big Bang Refactoring
```
❌ Don't:
- Refactor entire system at once
- Make hundreds of changes
- Go weeks without running tests

✅ Do:
- Refactor incrementally
- Small, verified steps
- Run tests continuously
```

### Refactoring Without Tests
```
❌ Don't:
- Refactor code with low test coverage
- Trust manual testing
- "I'll add tests later"

✅ Do:
- Add tests FIRST
- Achieve 90%+ coverage
- Then refactor safely
```

### Changing Behavior During Refactoring
```
❌ Don't:
- Fix bugs during refactoring
- Add features during refactoring
- "Improve" behavior during refactoring

✅ Do:
- Refactor OR fix bugs (separate commits)
- Preserve exact behavior
- Add features after refactoring
```

---

## Example Refactoring Session

### Code Smell: Long Method
```javascript
// Before: 50-line method handling order processing
function processOrder(order) {
  // 10 lines of validation
  // 15 lines of calculation
  // 10 lines of discount logic
  // 10 lines of email sending
  // 5 lines of database updates
}
```

### Refactoring Plan
```
Step 1: Extract validation
Step 2: Extract calculation
Step 3: Extract discount logic
Step 4: Extract email sending
Step 5: Extract database updates
Step 6: Rename for clarity
```

### Execution
```javascript
// Step 1: Extract validation
function validateOrder(order) {
  // validation logic
}

function processOrder(order) {
  validateOrder(order);
  // rest of logic
}
// → Run tests ✓

// Step 2: Extract calculation
function calculateOrderTotal(order) {
  // calculation logic
}

function processOrder(order) {
  validateOrder(order);
  const total = calculateOrderTotal(order);
  // rest of logic
}
// → Run tests ✓

// Continue for each step...

// Final result:
function processOrder(order) {
  validateOrder(order);
  const total = calculateOrderTotal(order);
  const finalTotal = applyDiscount(total, order.discountCode);
  await sendOrderEmail(order, finalTotal);
  await updateDatabase(order, finalTotal);
  return finalTotal;
}

// Each function is now small, focused, testable
// → Run all tests ✓
```

---

## Refactoring Safety Checklist

### Before Starting
```
□ Test coverage >90% for code to refactor
□ All tests passing
□ Performance baseline captured
□ Behavior documented
□ Refactoring plan defined
```

### During Refactoring
```
□ Making small changes
□ Running tests after each change
□ All tests still passing
□ No behavior changes
□ Committing at checkpoints
```

### After Completing
```
□ All tests passing
□ Coverage maintained
□ Performance verified
□ Code review completed
□ Documentation updated
```

---

## Success Criteria

A refactoring is successful when:
```
✓ All tests passing
✓ Code coverage maintained or improved
✓ No behavior changes
✓ Code simpler and cleaner
✓ Performance maintained
✓ Code smells eliminated
✓ Team can understand code better
```

---

## References

- [Standard Workflow](standard.md)
- [Refactoring Catalog](../quality/clean-code/03-refactoring.md)
- [Design Principles](../quality/clean-code/01-design-principles.md)
- [SOLID Principles](../quality/clean-code/02-solid-principles.md)
- [Testing Guidelines](../quality/clean-code/04-testing.md)

---

**Last reviewed:** 2026-01-07
**Next review:** Quarterly or when refactoring practices evolve
