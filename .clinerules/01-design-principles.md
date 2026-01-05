# Core Design Principles

> Based on principles from [martinfowler.com](https://martinfowler.com)

## Beck's Four Rules of Simple Design

Apply these rules in priority order when designing code:

1. **Passes the tests** - Code must work correctly with tests proving it
2. **Reveals intention** - Code should clearly express its purpose to readers
3. **No duplication** - Express logic "Once and Only Once" (DRY/SPOT principle)
4. **Fewest elements** - Remove unnecessary code; avoid over-engineering for hypothetical futures

### Guidelines for Applying These Rules

- These are NOT subjective preferences - there are objectively better and worse designs
- When tension arises between rules, solve the underlying problem rather than papering over it
- Empathy for future code readers should guide decisions when technical metrics conflict
- As Kent Beck stated: "There are better and worse designs...you can evaluate them right now"

**Reference:** [Beck's Design Rules](https://martinfowler.com/bliki/BeckDesignRules.html)

## Tell Don't Ask Principle

**Core Concept:** Tell objects what to do rather than asking for data and making external decisions.

### Application

- Bundle data with the functions that operate on it for better encapsulation
- Move behavior inside objects alongside their data
- Since data and operations that modify it are interconnected, keep them together

### Example

❌ **Don't Ask:**
```
if (monitor.getValue() > monitor.getLimit()) {
    alarm.trigger();
}
```

✅ **Do Tell:**
```
monitor.setValue(newValue);  // Monitor handles alarm internally
```

### Important Caveat

Don't apply this principle dogmatically. Sometimes query methods prove valuable for effective object collaboration, and other design concerns like system layering may take priority. Good design requires balancing multiple factors.

**Reference:** [Tell Don't Ask](https://martinfowler.com/bliki/TellDontAsk.html)

## Inversion of Control & Dependency Injection

### Core Principles

- **Separate component configuration from component use**
- Components should not create their own dependencies
- An external assembler provides dependencies to components
- "The Hollywood Principle: Don't call us, we'll call you"

### Dependency Injection Styles

**Constructor Injection** (Preferred for required dependencies)
- Dependencies passed through constructors
- Makes requirements explicit
- Creates valid objects immediately

**Setter Injection** (For optional dependencies)
- Dependencies assigned via setter methods
- Offers flexibility when components have many optional dependencies

**Interface Injection**
- Components implement specific interfaces for dependency injection
- Container uses interfaces to inject dependencies

### Key Benefits

- **Testability:** Components can be tested in isolation without framework infrastructure
- **Reusability:** Components don't depend on external locators/containers
- **Flexibility:** Enables flexible deployment strategies and reduces coupling

### Critical Principle

> "The choice between Service Locator and Dependency Injection is less important than the principle of separating service configuration from the use of services."

**References:**
- [Inversion of Control](https://martinfowler.com/bliki/InversionOfControl.html)
- [Dependency Injection](https://martinfowler.com/articles/injection.html)

## Seams for Testability

**Definition:** "A seam is a place where you can alter behavior in your program without editing in that place" (Michael Feathers)

### Why Seams Matter

- Enable test doubles and mocking
- Allow legacy system improvements without major rewrites
- Facilitate observation and enhancement points
- Test-driven development naturally produces code with appropriate seams

### Creating Seams

- Design interfaces for dependencies
- Use dependency injection to provide different implementations
- Avoid hard-coded dependencies and direct instantiation
- Consider extension points for future behavior changes
