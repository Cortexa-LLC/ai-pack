# Architecture and Code Organization

> Based on Martin Fowler's architectural guidance from [martinfowler.com](https://martinfowler.com)

## Bounded Contexts for Large Systems

**Core Concept:** Domain-Driven Design pattern for organizing large systems by dividing them into separate contexts with cohesive internal models.

### Why Bounded Contexts Matter

> "Total unification of the domain model for a large system will not be feasible or cost-effective."

Rather than forcing a single unified model across an entire organization:
- Divide systems into separate contexts
- Each context maintains its own cohesive model
- Use explicit boundaries and mapping between contexts

### Language and Vocabulary Boundaries

**The Problem:** Different teams use subtly different terminology

**Example:** A utility company where "meter" had three distinct meanings:
- Installation department: Physical device
- Billing department: Account identifier
- Technical department: Measurement reading

These linguistic variations (polysemes) create confusion in code but can be managed through bounded contexts.

### Shared vs. Separate Concepts

**Key Insight:** Contexts may share concept names like "Customer" or "Product" but model them differently.

- Each context has its own interpretation
- Mapping mechanisms translate between contexts
- Integration happens at explicit boundaries
- No forced unification of incompatible models

### Defining Context Boundaries

Primary factors that influence where boundaries form:

**Human Culture and Communication**
- Different contexts emerge where language changes
- Team structure and communication patterns
- Organizational boundaries

**Technical Representation**
- Distinct models for in-memory vs. database
- Different persistence strategies
- API vs. internal representations

**Historical and Organizational**
- Legacy systems
- Departmental divisions
- Existing code boundaries

### Benefits of Bounded Contexts

- **Independent Work** - Large teams work independently within their context
- **Internal Consistency** - Each context maintains its own unified model
- **Explicit Integration** - Clear architectural boundaries and integration points
- **Prevents Model Pollution** - No compromised unified model trying to serve all needs

**Reference:** [Bounded Context](https://martinfowler.com/bliki/BoundedContext.html)

## Code Organization Principles

### Decomposition Strategy

**Progressive refinement:**

1. **Break into small, focused functions/methods**
   - Each function does one thing well
   - Reveals intention through clear naming
   - Easy to understand and test

2. **Separate concerns**
   - Calculation from formatting
   - Business logic from presentation
   - Data access from business logic
   - Side effects from pure logic

3. **Use polymorphism for variations**
   - Type-based behavior uses polymorphism
   - Avoid switch statements on types
   - Strategy pattern for varying algorithms

### Cohesion: Keep Related Things Together

**Data and Behavior**
- Bundle data with functions that operate on it
- Behavior belongs with the data it manipulates
- Apply Tell Don't Ask principle

**Related Functionality**
- Group related functions/classes together
- Single Responsibility at module level
- Clear module boundaries

**Domain Concepts**
- Keep domain logic together
- Separate from technical infrastructure
- Bounded contexts for large domains

## Layered Architecture

### Common Layer Separation

**Presentation Layer**
- UI components
- User interaction handling
- Display formatting

**Application/Service Layer**
- Use cases and workflows
- Orchestration of domain logic
- Transaction boundaries

**Domain Layer**
- Core business logic
- Domain entities and value objects
- Business rules and invariants

**Infrastructure Layer**
- Database access
- External service integrations
- Technical concerns

### Layer Dependencies

- Higher layers depend on lower layers
- Lower layers don't know about higher layers
- Use Dependency Inversion to manage dependencies
- Interfaces defined by domain, implemented by infrastructure

## Module and Package Organization

### Organizing by Feature vs. Layer

**By Layer (Traditional)**
```
/controllers
/services
/repositories
/models
```

**By Feature (Preferred for larger systems)**
```
/user
  - UserController
  - UserService
  - UserRepository
  - User
/order
  - OrderController
  - OrderService
  - OrderRepository
  - Order
```

### Benefits of Feature-Based Organization

- Related code stays together
- Easier to understand and navigate
- Changes typically contained to one area
- Aligns with bounded contexts
- Supports independent teams

## Dependency Management

### Key Principles

**Acyclic Dependencies**
- No circular dependencies between modules
- Clear dependency direction
- Higher-level modules depend on lower-level

**Stable Dependencies Principle**
- Depend on things that change less frequently
- Unstable modules depend on stable modules
- Not the reverse

**Dependency Inversion**
- High-level modules don't depend on low-level modules
- Both depend on abstractions
- Details depend on abstractions

## Architectural Decision Making

### When Designing Architecture

**Consider:**
- What are the main use cases?
- Where are the natural boundaries?
- What things change together?
- What needs to be independently testable?
- What might need to be replaced or scaled?

**Avoid:**
- Over-engineering for hypothetical futures
- Premature optimization
- Copying patterns without understanding why
- Architecture for architecture's sake

### Evolutionary Architecture

- Start simple
- Refactor as understanding grows
- Let architecture emerge from needs
- Use refactoring to improve structure
- Tests enable safe architectural changes

## Cycle Time and Delivery

### Measuring Development Flow

**Cycle Time:** Measure time from idea to production

**Why It Matters:**
- Fast feedback loops enable agility
- Quick delivery provides business value sooner
- Short cycles reduce risk
- Enables rapid learning and adaptation

**Consistency Matters:**
- Use consistent start/stop points
- Makes comparison meaningful across teams
- Track trends over time
- Identify bottlenecks in process

## Anti-Patterns to Avoid

### Architectural

- **Big Ball of Mud** - No clear structure or boundaries
- **God Object** - One class/module knows/does everything
- **Spaghetti Code** - Tangled dependencies everywhere
- **Lasagna Code** - Too many layers obscuring simple operations
- **Premature Generalization** - Abstract before understanding patterns

### Organizational

- **Conway's Law Violations** - Architecture fights team structure
- **Ivory Tower Architecture** - Decisions disconnected from implementation reality
- **Not Invented Here** - Rebuilding everything from scratch
- **Golden Hammer** - Using same pattern/tool for everything
- **Analysis Paralysis** - Over-planning before starting

## Practical Guidelines

### Starting a New System

1. Start with simplest structure that works
2. Identify core domain concepts
3. Establish clear boundaries early
4. Use tests to validate architecture
5. Refactor as understanding grows

### Working with Legacy Systems

1. Understand existing structure first
2. Identify seams for change
3. Add tests before refactoring
4. Improve incrementally
5. Respect bounded contexts that exist
6. Don't try to fix everything at once

### Maintaining Systems

1. Keep architecture documentation current
2. Make dependencies explicit and visible
3. Regular refactoring to prevent decay
4. Address code smells promptly
5. Evolve architecture as needs change
