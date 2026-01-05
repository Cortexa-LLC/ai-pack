# API and Interface Design Principles

> Principles for designing clean interfaces from [martinfowler.com](https://martinfowler.com)

## Command Query Separation (CQS)

**Definition:** Divide an object's methods into two distinct categories: queries and commands.

### The Two Categories

**Queries**
- Return a result
- Do NOT change observable state of the system
- Free of side effects
- Can be called multiple times safely
- Order of execution doesn't matter

**Commands (Modifiers)**
- Change the state of the system
- Do NOT return a value
- Have side effects
- Order and frequency matter
- Require careful consideration when calling

### Benefits of CQS

**Clarity and Confidence**
- Immediately understand method's impact from its signature
- Return type signals whether state changes
- Can reason about code more easily
- Reduces cognitive load

**Freedom with Queries**
- Use queries liberally without risk
- Reorder query calls without concern
- Introduce new queries anywhere
- No unintended consequences

**Caution with Commands**
- Explicit signal to exercise care
- Impact on system state is clear
- Easier to track state changes
- Better debugging and reasoning

### Examples

**Good - Follows CQS:**
```javascript
// Query - no side effects
const total = order.getTotal();
const items = cart.getItems();
const isValid = user.isAuthenticated();

// Command - changes state, returns nothing
cart.addItem(item);
user.logout();
order.submit();
```

**Problematic - Violates CQS:**
```javascript
// Returns value AND changes state
const item = stack.pop();  // Removes and returns
const user = cache.getOrCreate(id);  // Might create
const count = list.removeAll(predicate);  // Removes and counts
```

### Practical Considerations

**Exceptions Exist**

Some methods are useful enough to justify breaking the rule:
- `stack.pop()` - Returns and removes (convenience outweighs purity)
- `collection.remove(item)` - Returns success/failure (useful feedback)
- `cache.getOrCreate(key)` - Query with lazy initialization

**Meyer's Guidance:**
Follow the principle when feasible but recognize situations where flexibility matters more than purity.

**Language Support**
- C++ `const` modifier enforces query contracts
- Many languages lack built-in CQS support
- Rely on naming conventions and discipline

### Naming Conventions

**Queries - Use nouns or questions:**
- `getTotal()`, `isValid()`, `hasPermission()`
- `calculatePrice()`, `findUser()`, `countItems()`
- Names suggest information retrieval

**Commands - Use verbs:**
- `setStatus()`, `addItem()`, `removeUser()`
- `update()`, `delete()`, `save()`
- Names suggest action/mutation

### CQS in Different Contexts

**Object-Oriented Design**
- Instance methods separated by query/command
- Helps maintain immutability
- Clearer object contracts

**Functional Programming**
- Pure functions are ultimate queries
- Side effects isolated to specific boundaries
- CQS aligns with functional principles

**API Design**
- GET requests = queries (idempotent, safe)
- POST/PUT/DELETE = commands (state-changing)
- HTTP naturally enforces CQS

**Database Operations**
- SELECT = query
- INSERT/UPDATE/DELETE = commands
- SQL separates concerns explicitly

### CQRS - Taking CQS Further

**Command Query Responsibility Segregation:**
- Extends CQS to architectural level
- Separate models for reads and writes
- Different data stores for queries vs. commands
- Optimized independently

**When to use CQRS:**
- High read/write ratio imbalance
- Different scaling needs for reads vs. writes
- Complex domain with distinct read/write patterns
- Event sourcing architecture

### Common Violations to Watch For

**Hidden Side Effects in Queries**
```javascript
// BAD - getUser() shouldn't log or cache
function getUser(id) {
    logAccess(id);  // Side effect!
    const user = database.find(id);
    cache.store(id, user);  // Side effect!
    return user;
}

// GOOD - pure query
function getUser(id) {
    return database.find(id);
}
```

**Commands Returning Too Much**
```javascript
// BAD - update returns entire object
function updateEmail(userId, email) {
    user.email = email;
    return user;  // Mixing command with query
}

// BETTER - return success/void
function updateEmail(userId, email) {
    user.email = email;
    // Return nothing or just success indicator
}
```

### Testing Benefits

**Queries are Easy to Test**
- Pure functions without side effects
- No setup/teardown needed
- No mocking required
- Fast and reliable

**Commands Require More Care**
- Need to verify state changes
- Setup and teardown important
- May require mocks/stubs
- Potentially slower

### CQS and Immutability

**Natural Alignment:**
- Immutable objects only have queries
- Commands return new objects instead of mutating
- Functional programming style
- Thread-safe by default

**Example:**
```javascript
// Immutable style - no commands, only queries
const order = createOrder();
const withItem = order.addItem(item);  // Returns new order
const total = withItem.getTotal();  // Query
```

**Reference:** [Command Query Separation](https://martinfowler.com/bliki/CommandQuerySeparation.html)

## Naming Things - The Ongoing Challenge

> "There are only two hard things in Computer Science: cache invalidation and naming things." (Phil Karlton)

### Why Naming is Hard

**Universal Developer Challenge**
- Affects all skill levels equally
- Requires conveying purpose and behavior concisely
- Balance between precision and brevity
- No single "right" answer

**Long-Lasting Impact**
- Names persist in codebases for years
- Bad names compound confusion over time
- Renaming is expensive (but worthwhile)
- First impression matters

### Principles for Good Naming

**1. Reveal Intention**
- Name should clearly express purpose
- Reader shouldn't need to decipher meaning
- Connects to Beck's Design Rule #2
- Self-documenting code

**Examples:**
```javascript
// BAD
const d = 86400;  // What is d?
const temp = getUserData();  // What data?

// GOOD
const secondsPerDay = 86400;
const activeUserProfile = getUserProfile();
```

**2. Use Domain Language**
- Match business/domain terminology
- Makes code understandable to non-programmers
- Reduces translation between requirements and code
- Supports ubiquitous language (DDD)

**Examples:**
```javascript
// BAD - technical jargon
class RecordProcessor { }
class DataContainer { }

// GOOD - domain language
class InvoiceProcessor { }
class ShoppingCart { }
```

**3. Be Consistent**
- Same terminology for same concepts
- Follow established patterns in codebase
- Don't use synonyms for identical things
- Predictability helps comprehension

**Examples:**
```javascript
// BAD - inconsistent
getUserData()
fetchClientInfo()
retrieveCustomerDetails()  // All do the same thing!

// GOOD - consistent
getUserProfile()
getClientProfile()
getCustomerProfile()
```

**4. Avoid Meaningless Names**

**Generic names to avoid:**
- `data`, `info`, `item`, `object`
- `manager`, `handler`, `processor`, `helper`
- `doStuff()`, `process()`, `handle()`
- Single letters (except loop counters)

**When generic names are OK:**
- Loop indices: `i`, `j`, `k`
- Generic algorithms: `T` for type parameter
- Established conventions: `e` for error/event

**5. Choose Appropriate Length**

**Context matters:**
- Short scope → shorter names OK
- Long scope → more descriptive names needed
- Frequently used → can be shorter
- Rarely used → should be more descriptive

**Examples:**
```javascript
// Short scope - fine
for (let i = 0; i < items.length; i++) {
    const item = items[i];
    process(item);
}

// Long scope - needs clarity
class CustomerOrderProcessingService {
    processCustomerOrderWithValidation(order) {
        // Longer, more descriptive name appropriate
    }
}
```

**6. Avoid Abbreviations**

**Unless they're universally understood:**
```javascript
// BAD
const usrMgr = new UsrMgr();
const calcTot = () => { };

// GOOD
const userManager = new UserManager();
const calculateTotal = () => { };

// OK - widely understood
const html = "<div>...</div>";
const url = "https://...";
const id = user.id;
```

### Names for Different Constructs

**Classes and Types**
- Nouns or noun phrases
- Describe what it IS
- Examples: `User`, `OrderProcessor`, `PaymentGateway`

**Functions and Methods**
- Verbs or verb phrases
- Describe what it DOES
- Examples: `calculateTotal()`, `sendEmail()`, `validateInput()`

**Boolean Variables/Functions**
- Questions or assertions
- Start with `is`, `has`, `can`, `should`
- Examples: `isValid`, `hasPermission`, `canEdit`, `shouldRetry`

**Constants**
- UPPER_SNAKE_CASE traditionally
- Descriptive of value or purpose
- Examples: `MAX_RETRY_ATTEMPTS`, `DEFAULT_TIMEOUT_MS`

**Collections**
- Plural nouns
- Describe contents
- Examples: `users`, `orders`, `activeConnections`

### Warning Signs of Bad Names

**Need for Explanatory Comments**
```javascript
// BAD - name doesn't explain itself
const x = 5;  // Number of retry attempts

// GOOD - self-explanatory
const maxRetryAttempts = 5;
```

**Different Interpretations**
- Team members understand differently
- Leads to bugs and confusion
- Sign name is too vague or misleading

**Mismatched Name and Behavior**
```javascript
// BAD - does more than name suggests
function getUser(id) {
    const user = db.find(id);
    logAccess(user);  // Surprise!
    sendAnalytics(user);  // More surprises!
    return user;
}

// GOOD - name matches behavior
function getUserWithTracking(id) {
    const user = db.find(id);
    trackUserAccess(user);
    return user;
}
```

### Refactoring Names

**Don't Fear Renaming**
- Modern IDEs make renaming safe
- Better name improves entire codebase
- Worth the effort for frequently used identifiers
- Gradual improvement over time

**When to Rename**
- Understanding of domain improves
- Original name misleads
- Encountering confusion repeatedly
- During refactoring anyway

**How to Rename Safely**
- Use IDE refactoring tools
- Rename in small commits
- Update tests and documentation
- Use Parallel Change pattern for public APIs

### Semantic Diffusion and Naming

**Be Precise with Technical Terms**
- Don't let terms lose their meaning
- Use established terms correctly
- Define terms when introducing them
- Resist vague buzzwords

**Examples of term precision:**
- "Refactoring" means specific technique (see 07-development-practices.md)
- "Factory" means specific pattern
- "Service" has architectural meaning
- Use terms correctly or choose different words

### Cultural and Team Considerations

**Establish Naming Conventions**
- Document in style guide
- Discuss during code review
- Build shared vocabulary
- Refine over time

**Ubiquitous Language (DDD)**
- Use same terms in code as business uses
- Developers and business speak same language
- Reduces misunderstanding
- Code becomes domain model

**References:**
- [Two Hard Things](https://martinfowler.com/bliki/TwoHardThings.html)
- [Semantic Diffusion](https://martinfowler.com/bliki/SemanticDiffusion.html)

## API Design Principles Summary

### Interface Design Checklist

- [ ] Separate queries (return value, no side effects) from commands (mutate state, no return)
- [ ] Names clearly reveal intention
- [ ] Use domain language consistently
- [ ] Avoid generic names like "manager", "handler", "processor"
- [ ] Boolean functions start with is/has/can/should
- [ ] Method names match their actual behavior
- [ ] No hidden side effects in queries
- [ ] Consider immutability for complex state
- [ ] Document exceptions to CQS when necessary
- [ ] Use appropriate name length for scope

### Common API Design Mistakes

- Query methods with hidden side effects (logging, caching, etc.)
- Commands that return complex objects (mixing concerns)
- Misleading names that don't match behavior
- Inconsistent terminology across API
- Generic names that don't convey meaning
- Overly abbreviated names
- Breaking CQS without good reason
- Not considering caller's perspective
