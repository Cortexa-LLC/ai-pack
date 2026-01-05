# SOLID Principles

> Practical application of SOLID principles based on Martin Fowler's guidance

## Single Responsibility Principle (SRP)

**Principle:** Each class/module should have one reason to change.

### Application

- Related to Beck's "reveals intention" - focused classes are easier to understand
- A class should have only one job or responsibility
- If a class has multiple reasons to change, split it into separate classes

### Warning Signs (Code Smells)

- **Large Class** - Too many fields/methods/lines of code
- **Divergent Change** - One class changes for multiple different reasons
- **Shotgun Surgery** - One change requires modifications across many classes

### Example

❌ **Violates SRP:**
```
class UserManager {
    validateUser()
    saveToDatabase()
    sendEmail()
    generateReport()
}
```

✅ **Follows SRP:**
```
class UserValidator { validateUser() }
class UserRepository { saveToDatabase() }
class EmailService { sendEmail() }
class ReportGenerator { generateReport() }
```

## Open/Closed Principle (OCP)

**Principle:** Software entities should be open for extension, closed for modification.

### Application

- Use polymorphism to organize calculations by type
- Prefer composition and interfaces over modification of existing code
- Add new functionality by adding new code, not changing existing code
- Related to refactoring flow: organize by type using polymorphism

### Implementation Strategies

- Abstract base classes or interfaces
- Strategy pattern for varying algorithms
- Template method pattern for varying steps
- Dependency injection for swappable implementations

## Liskov Substitution Principle (LSP)

**Principle:** Subtypes must be substitutable for their base types without breaking correctness.

### Application

- Ensure polymorphic behavior is predictable and consistent
- Derived classes should honor the contracts established by base classes
- Don't break behavioral expectations of calling code
- Don't weaken preconditions or strengthen postconditions

### Violation Examples

- Subclass throws exceptions not thrown by base class
- Subclass changes behavior in unexpected ways
- Subclass removes functionality present in base class

## Interface Segregation Principle (ISP)

**Principle:** Clients shouldn't depend on interfaces they don't use.

### Application

- Create focused, cohesive interfaces
- Related to Beck's "fewest elements" - remove unnecessary dependencies
- Split large interfaces into smaller, specific ones
- Prevents "fat interfaces" that force implementations to define unused methods

### Benefits

- Reduces coupling between components
- Makes code more flexible and easier to refactor
- Improves testability by minimizing dependencies

## Dependency Inversion Principle (DIP)

**Principle:** Depend on abstractions, not concretions.

### Core Rules

1. High-level modules shouldn't depend on low-level modules
2. Both should depend on abstractions (interfaces/protocols)
3. Abstractions shouldn't depend on details
4. Details should depend on abstractions

### Application

- Enables testability through test doubles
- Use interfaces to define contracts between components
- Inject dependencies rather than creating them directly
- Aligns with Dependency Injection principle

### Example

❌ **Violates DIP:**
```
class OrderProcessor {
    private database = new MySQLDatabase();  // Concrete dependency

    process() {
        this.database.save();
    }
}
```

✅ **Follows DIP:**
```
interface Database {
    save();
}

class OrderProcessor {
    constructor(private database: Database) {}  // Abstract dependency

    process() {
        this.database.save();
    }
}
```

## SOLID Principles Summary

When applied together, SOLID principles create:

- Code that's easier to understand and maintain
- Systems that are flexible and adaptable to change
- Components that are testable and mockable
- Architecture that supports growth without major rewrites
- Clear boundaries and responsibilities

Remember: These are guidelines, not absolute rules. Apply them pragmatically based on context and balance them against other design concerns.
