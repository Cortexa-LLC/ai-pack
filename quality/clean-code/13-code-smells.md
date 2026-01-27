# Code Smells: Identification and Prevention Guide

**Version:** 1.0
**Last Updated:** 2026-01-27
**Reference:** [Refactoring.Guru - Code Smells](https://refactoring.guru/refactoring/smells)

## Table of Contents

1. [Introduction](#introduction)
2. [Bloaters](#bloaters)
3. [Object-Orientation Abusers](#object-orientation-abusers)
4. [Change Preventers](#change-preventers)
5. [Dispensables](#dispensables)
6. [Couplers](#couplers)
7. [Detection Strategy](#detection-strategy)
8. [Prevention Guidelines](#prevention-guidelines)
9. [Refactoring Priorities](#refactoring-priorities)

---

## Introduction

**Code smells** are indicators of deeper problems in your codebase. They're not bugs—the code works—but they signal design issues that will make future maintenance difficult, expensive, and error-prone.

### Why This Matters

- **Maintenance Cost:** Code smells increase the time needed to understand, modify, and test code
- **Bug Risk:** Smelly code is more likely to contain bugs and harder to debug
- **Team Velocity:** Poor code quality slows down development over time
- **Technical Debt:** Accumulated smells compound, making eventual refactoring more difficult

### How to Use This Guide

1. **During Code Review:** Check for these patterns before merging
2. **During Development:** Recognize smells as you write code and refactor immediately
3. **During Refactoring:** Use this guide to prioritize which smells to address first
4. **Agent Integration:** Reviewer agents should flag these patterns automatically

---

## Bloaters

**Definition:** Code, methods, and classes that have grown to "gargantuan proportions" and are hard to work with. These accumulate over time without active prevention.

### 1. Long Method

**Description:**
Methods exceeding 10-15 lines should raise questions. The longer a method, the harder it is to understand and maintain.

**How to Recognize:**
- Methods longer than one screen/page
- Difficulty naming the method concisely
- Multiple levels of indentation
- Scroll required to see entire method
- Comments needed to explain sections

**Why It's Bad:**
- Reduces readability and comprehension
- Makes testing more complex
- Hinders code reuse
- Increases bug risk (harder to spot edge cases)
- Complicates change impact analysis

**How to Avoid:**
- Follow Single Responsibility Principle (SRP)
- Extract helper methods early and often
- Aim for methods that fit on one screen (20-30 lines max)
- Each method should do ONE thing at ONE level of abstraction

**How to Fix:**
1. **Extract Method:** Pull out logical sections into named methods
2. **Replace Temp with Query:** Convert temporary variables to method calls
3. **Introduce Parameter Object:** Group related parameters
4. **Decompose Conditional:** Extract complex conditionals into named methods

**Example (Go):**

```go
// ❌ BAD: Long method doing too much
func ProcessOrder(order Order) error {
    // Validate order (10 lines)
    if order.Items == nil || len(order.Items) == 0 {
        return errors.New("empty order")
    }
    // ... more validation ...

    // Calculate totals (15 lines)
    var subtotal float64
    for _, item := range order.Items {
        subtotal += item.Price * float64(item.Quantity)
    }
    // ... tax, shipping calculations ...

    // Apply discounts (20 lines)
    // ... complex discount logic ...

    // Process payment (15 lines)
    // ... payment gateway integration ...

    // Send notifications (10 lines)
    // ... email/SMS logic ...

    return nil
}

// ✅ GOOD: Extracted into focused methods
func ProcessOrder(order Order) error {
    if err := validateOrder(order); err != nil {
        return err
    }

    totals := calculateOrderTotals(order)
    totals = applyDiscounts(order, totals)

    if err := processPayment(order, totals); err != nil {
        return err
    }

    sendOrderNotifications(order)
    return nil
}
```

---

### 2. Large Class

**Description:**
A class with too many fields, methods, or lines of code, indicating it has too many responsibilities.

**How to Recognize:**
- Class exceeds 200-300 lines
- Numerous instance variables (>7-10)
- Many methods (>15-20)
- Difficulty naming the class concisely
- Class name contains "And", "Manager", "Handler", "Utils"

**Why It's Bad:**
- Violates Single Responsibility Principle
- Low cohesion (unrelated functionality grouped together)
- Difficult to understand and maintain
- Increases likelihood of bugs
- Makes testing complex (too many scenarios)

**How to Avoid:**
- Design classes with ONE clear responsibility
- Limit class scope during initial design
- Review class size regularly
- Extract responsibilities early

**How to Fix:**
1. **Extract Class:** Move related functionality to new class
2. **Extract Subclass:** Create specialized subclasses
3. **Extract Interface:** Define clear contracts
4. **Replace Data Value with Object:** Convert primitive fields to objects

**Example (Python):**

```python
# ❌ BAD: Monolithic class doing everything
class OrderProcessor:
    def __init__(self):
        self.db_connection = None
        self.email_client = None
        self.payment_gateway = None
        self.inventory_system = None
        self.shipping_api = None
        # ... 10 more fields

    def process_order(self, order): ...
    def validate_order(self, order): ...
    def calculate_total(self, order): ...
    def apply_discount(self, order): ...
    def charge_payment(self, order): ...
    def update_inventory(self, order): ...
    def create_shipment(self, order): ...
    def send_confirmation(self, order): ...
    def send_tracking(self, order): ...
    # ... 15 more methods

# ✅ GOOD: Separated concerns
class OrderProcessor:
    def __init__(self, validator, calculator, payment_service, fulfillment_service):
        self.validator = validator
        self.calculator = calculator
        self.payment = payment_service
        self.fulfillment = fulfillment_service

    def process(self, order):
        self.validator.validate(order)
        totals = self.calculator.calculate_totals(order)
        self.payment.charge(order, totals)
        self.fulfillment.fulfill(order)

class OrderValidator:
    def validate(self, order): ...

class OrderCalculator:
    def calculate_totals(self, order): ...

class PaymentService:
    def charge(self, order, amount): ...

class FulfillmentService:
    def fulfill(self, order): ...
```

---

### 3. Primitive Obsession

**Description:**
Using primitive data types (int, string, bool) instead of small objects for simple tasks like currency, ranges, phone numbers, etc.

**How to Recognize:**
- Magic numbers scattered throughout code
- String constants used for types/statuses
- Complex validation logic for "simple" primitives
- Type information encoded in variable names (`userIdString`, `priceInCents`)
- Constants defined for coded information (`USER_ADMIN = 1`)

**Why It's Bad:**
- Loses type safety
- Scatters domain logic across codebase
- No centralized validation
- Obscures business concepts
- Difficult to add behavior or validation later

**How to Avoid:**
- Create value objects for domain concepts
- Use type systems to enforce constraints
- Encapsulate validation in objects
- Think "is this a concept?" not "is this data?"

**How to Fix:**
1. **Replace Data Value with Object:** Create class for the value
2. **Replace Type Code with Class:** Use proper objects instead of constants
3. **Replace Type Code with Subclasses:** Use inheritance for type variations
4. **Introduce Parameter Object:** Bundle related primitives

**Example (Go):**

```go
// ❌ BAD: Primitive obsession
func TransferMoney(fromAccount string, toAccount string, amount float64, currency string) error {
    if amount <= 0 {
        return errors.New("invalid amount")
    }
    if currency != "USD" && currency != "EUR" && currency != "GBP" {
        return errors.New("invalid currency")
    }
    // ... validation repeated everywhere
}

func CalculateInterest(principal float64, rate float64, years int) float64 {
    // No protection against negative values, rate > 1.0, etc.
    return principal * rate * float64(years)
}

// ✅ GOOD: Value objects
type Money struct {
    Amount   decimal.Decimal
    Currency Currency
}

func NewMoney(amount decimal.Decimal, currency Currency) (Money, error) {
    if amount.IsNegative() {
        return Money{}, errors.New("amount cannot be negative")
    }
    return Money{Amount: amount, Currency: currency}, nil
}

type Currency string

const (
    USD Currency = "USD"
    EUR Currency = "EUR"
    GBP Currency = "GBP"
)

type AccountID struct {
    value string
}

func NewAccountID(id string) (AccountID, error) {
    if !isValidAccountID(id) {
        return AccountID{}, errors.New("invalid account ID format")
    }
    return AccountID{value: id}, nil
}

func TransferMoney(from AccountID, to AccountID, amount Money) error {
    // Validation already done in constructors
    // Type safety enforced by compiler
}
```

---

### 4. Long Parameter List

**Description:**
Methods with more than 3-4 parameters are difficult to use correctly and indicate design problems.

**How to Recognize:**
- Method signatures with 4+ parameters
- Parameters with similar types (multiple strings, ints)
- Difficulty remembering parameter order
- Need to look up method signature frequently
- Parameters frequently passed together

**Why It's Bad:**
- Reduces readability
- Easy to mix up parameter order (especially with same types)
- Indicates possible missing abstractions
- Makes testing more complex
- Hinders method evolution (adding parameters breaks callers)

**How to Avoid:**
- Use parameter objects
- Limit methods to 3 parameters max
- Consider if parameters belong together (missing object)
- Use builder pattern for complex construction

**How to Fix:**
1. **Introduce Parameter Object:** Group related parameters
2. **Preserve Whole Object:** Pass entire object instead of fields
3. **Replace Parameter with Method Call:** Use existing relationships
4. **Use Builder Pattern:** For construction scenarios

**Example (Python):**

```python
# ❌ BAD: Too many parameters
def create_user(
    username: str,
    email: str,
    first_name: str,
    last_name: str,
    phone: str,
    address: str,
    city: str,
    state: str,
    zip_code: str,
    country: str
) -> User:
    # Easy to mix up parameter order
    # Hard to add new fields
    pass

# ✅ GOOD: Parameter object
@dataclass
class UserProfile:
    username: str
    email: str
    name: PersonName
    contact: ContactInfo
    address: Address

@dataclass
class PersonName:
    first: str
    last: str

@dataclass
class ContactInfo:
    email: str
    phone: str

@dataclass
class Address:
    street: str
    city: str
    state: str
    zip_code: str
    country: str

def create_user(profile: UserProfile) -> User:
    # Single, well-organized parameter
    # Easy to add fields to nested objects
    pass

# Even better: Use builder for optional parameters
class UserBuilder:
    def __init__(self):
        self._profile = UserProfile()

    def with_username(self, username: str) -> 'UserBuilder':
        self._profile.username = username
        return self

    def with_email(self, email: str) -> 'UserBuilder':
        self._profile.email = email
        return self

    def build(self) -> User:
        return create_user(self._profile)

# Usage: UserBuilder().with_username("alice").with_email("alice@example.com").build()
```

---

### 5. Data Clumps

**Description:**
Groups of variables that consistently appear together in multiple places suggest a missing abstraction.

**How to Recognize:**
- Same 2-3 variables passed together repeatedly
- Common parameter groups across methods
- Similar field groups in multiple classes
- Deleting one variable from the group breaks code in multiple places

**Why It's Bad:**
- Indicates missing domain concept
- Increases code duplication
- Changes require updates in multiple locations
- Obscures design intent
- Makes refactoring more difficult

**How to Avoid:**
- Identify concepts early
- Create value objects for related data
- Extract classes when patterns emerge
- Use domain-driven design principles

**How to Fix:**
1. **Extract Class:** Create object for the clump
2. **Introduce Parameter Object:** Bundle parameters
3. **Preserve Whole Object:** Pass complete object

**Example (Go):**

```go
// ❌ BAD: Data clumps everywhere
func ConnectToDatabase(host string, port int, username string, password string) (*DB, error) { ... }
func BackupDatabase(host string, port int, username string, password string) error { ... }
func MigrateDatabase(host string, port int, username string, password string) error { ... }

func CreateUser(firstName string, lastName string, email string) { ... }
func UpdateUser(userID int, firstName string, lastName string, email string) { ... }
func DisplayUser(firstName string, lastName string, email string) { ... }

// ✅ GOOD: Extract classes for clumps
type DatabaseConfig struct {
    Host     string
    Port     int
    Username string
    Password string
}

func ConnectToDatabase(config DatabaseConfig) (*DB, error) { ... }
func BackupDatabase(config DatabaseConfig) error { ... }
func MigrateDatabase(config DatabaseConfig) error { ... }

type UserInfo struct {
    FirstName string
    LastName  string
    Email     string
}

func CreateUser(info UserInfo) { ... }
func UpdateUser(userID int, info UserInfo) { ... }
func DisplayUser(info UserInfo) { ... }
```

---

## Object-Orientation Abusers

**Definition:** Incomplete or incorrect application of object-oriented programming principles. These smells indicate fundamental OOP design problems.

### 1. Switch Statements (Type Code)

**Description:**
Complex switch/case or if-else chains, especially when switching on type codes. This is procedural thinking in OO code.

**How to Recognize:**
- Multiple switch/case on same variable
- Switch on type field/code
- Same switch logic duplicated across methods
- New types require changes to multiple switch statements

**Why It's Bad:**
- Violates Open/Closed Principle
- Duplicated logic across switch statements
- Adding new types requires finding all switches
- Prevents leveraging polymorphism
- Makes extension difficult

**How to Avoid:**
- Use polymorphism instead of conditionals
- Apply Strategy pattern
- Use type systems (interfaces, abstract classes)
- Think "behavior" not "type checking"

**How to Fix:**
1. **Replace Conditional with Polymorphism:** Use subclasses
2. **Replace Type Code with State/Strategy:** Use pattern objects
3. **Introduce Null Object:** Eliminate null checks

**Example (Python):**

```python
# ❌ BAD: Switch statements on type
class ShapeRenderer:
    def render(self, shape):
        if shape.type == "circle":
            return self.render_circle(shape.radius)
        elif shape.type == "rectangle":
            return self.render_rectangle(shape.width, shape.height)
        elif shape.type == "triangle":
            return self.render_triangle(shape.base, shape.height)
        else:
            raise ValueError(f"Unknown shape type: {shape.type}")

    def calculate_area(self, shape):
        if shape.type == "circle":
            return 3.14 * shape.radius ** 2
        elif shape.type == "rectangle":
            return shape.width * shape.height
        elif shape.type == "triangle":
            return 0.5 * shape.base * shape.height
        # Duplicated switch logic!

# ✅ GOOD: Polymorphism
from abc import ABC, abstractmethod

class Shape(ABC):
    @abstractmethod
    def render(self, renderer) -> str:
        pass

    @abstractmethod
    def calculate_area(self) -> float:
        pass

class Circle(Shape):
    def __init__(self, radius):
        self.radius = radius

    def render(self, renderer):
        return renderer.render_circle(self.radius)

    def calculate_area(self):
        return 3.14 * self.radius ** 2

class Rectangle(Shape):
    def __init__(self, width, height):
        self.width = width
        self.height = height

    def render(self, renderer):
        return renderer.render_rectangle(self.width, self.height)

    def calculate_area(self):
        return self.width * self.height

class Triangle(Shape):
    def __init__(self, base, height):
        self.base = base
        self.height = height

    def render(self, renderer):
        return renderer.render_triangle(self.base, self.height)

    def calculate_area(self):
        return 0.5 * self.base * self.height

# Usage - adding new shape doesn't require changing existing code
shapes: List[Shape] = [Circle(5), Rectangle(4, 6), Triangle(3, 4)]
for shape in shapes:
    print(f"Area: {shape.calculate_area()}")
    print(shape.render(renderer))
```

---

### 2. Temporary Field

**Description:**
Instance variables that are only populated under certain circumstances and empty/null otherwise.

**How to Recognize:**
- Fields that are null most of the time
- Fields only used in specific methods
- Confusing object state (why is this field sometimes empty?)
- Initialization checks scattered throughout methods

**Why It's Bad:**
- Creates confusing object state
- Violates expectation that objects have consistent state
- Increases cognitive load
- Makes debugging difficult
- Suggests wrong object design

**How to Avoid:**
- Only include fields that are always meaningful
- Extract specialized classes for conditional state
- Use method parameters instead of temporary fields
- Consider if the field belongs in a different object

**How to Fix:**
1. **Extract Class:** Move field and methods that use it
2. **Replace Method with Method Object:** Create class for algorithm
3. **Introduce Null Object:** Eliminate null checks

**Example (Go):**

```go
// ❌ BAD: Temporary field pattern
type OrderProcessor struct {
    db *Database

    // Only used during bulk processing, nil otherwise
    bulkCache *BulkCache

    // Only used for discounted orders
    discountCalculator *DiscountCalculator

    // Only set when processing failed
    lastError error
}

func (p *OrderProcessor) ProcessOrder(order Order) error {
    // Confusing: why do I need to check if these are nil?
    if p.discountCalculator != nil {
        // Use it
    }
    // What state am I in?
}

// ✅ GOOD: Extracted classes, clear state
type OrderProcessor struct {
    db       *Database
    discount DiscountService // Always present, uses Null Object if no discount
}

func (p *OrderProcessor) ProcessOrder(order Order) error {
    // Always safe to call, discount service handles "no discount" case
    finalPrice := p.discount.ApplyDiscount(order)
    return p.db.SaveOrder(order)
}

// Bulk processing has its own class with appropriate state
type BulkOrderProcessor struct {
    db    *Database
    cache *BulkCache // Always present when bulk processing
}

func (p *BulkOrderProcessor) ProcessBatch(orders []Order) error {
    // cache is always available - it's fundamental to bulk processing
    for _, order := range orders {
        p.cache.Add(order)
    }
    return p.cache.Flush()
}
```

---

### 3. Refused Bequest

**Description:**
A subclass inherits methods/properties it doesn't need or use, indicating improper inheritance hierarchy.

**How to Recognize:**
- Subclass only uses small portion of inherited functionality
- Inherited methods throw "not supported" exceptions
- Subclass overrides parent methods to do nothing
- Inheritance is for code reuse, not "is-a" relationship

**Why It's Bad:**
- Violates Liskov Substitution Principle (LSP)
- Creates confusing inheritance hierarchies
- Misleads developers about class capabilities
- Increases coupling unnecessarily
- Makes maintenance difficult

**How to Avoid:**
- Use composition over inheritance
- Only inherit when true "is-a" relationship exists
- Prefer interface segregation
- Use delegation instead of inheritance

**How to Fix:**
1. **Replace Inheritance with Delegation:** Use composition
2. **Extract Superclass:** Create new parent with common functionality
3. **Extract Interface:** Define precise contracts

**Example (Python):**

```python
# ❌ BAD: Refused bequest
class Bird:
    def fly(self):
        return "Flying in the sky"

    def lay_egg(self):
        return "Laying an egg"

class Penguin(Bird):
    def fly(self):
        # Refuses the bequest - penguins can't fly!
        raise NotImplementedError("Penguins cannot fly")

    # Only uses lay_egg, not fly - wrong hierarchy

# ✅ GOOD: Composition and proper abstraction
from abc import ABC, abstractmethod

class Bird(ABC):
    @abstractmethod
    def move(self):
        pass

    def lay_egg(self):
        return "Laying an egg"

class FlyingBird(Bird):
    def move(self):
        return "Flying in the sky"

    def fly(self):
        return "Flying in the sky"

class FlightlessBird(Bird):
    def move(self):
        return "Walking or swimming"

class Sparrow(FlyingBird):
    pass

class Penguin(FlightlessBird):
    def swim(self):
        return "Swimming in water"

# Even better: Use composition for flight capability
class FlightCapability:
    def fly(self):
        return "Flying in the sky"

class SwimCapability:
    def swim(self):
        return "Swimming in water"

class Bird:
    def __init__(self, flight_capability=None, swim_capability=None):
        self._flight = flight_capability
        self._swim = swim_capability

    def move(self):
        if self._flight:
            return self._flight.fly()
        elif self._swim:
            return self._swim.swim()
        return "Walking"

# Usage
sparrow = Bird(flight_capability=FlightCapability())
penguin = Bird(swim_capability=SwimCapability())
```

---

### 4. Alternative Classes with Different Interfaces

**Description:**
Two or more classes perform the same function but have different method names/signatures.

**How to Recognize:**
- Similar functionality with different names
- Copy-pasted code with minor variations
- Multiple implementations of same concept
- Need to maintain parallel changes

**Why It's Bad:**
- Violates DRY principle
- Creates confusion about which to use
- Duplicate maintenance effort
- Makes refactoring difficult
- Prevents polymorphism

**How to Avoid:**
- Define common interfaces early
- Use consistent naming conventions
- Refactor duplicates immediately
- Apply interface segregation

**How to Fix:**
1. **Rename Method:** Use consistent names
2. **Extract Superclass:** Create common parent
3. **Extract Interface:** Define shared contract
4. **Pull Up Method:** Move common methods to parent

**Example (Go):**

```go
// ❌ BAD: Alternative classes with different interfaces
type FileLogger struct{}

func (l *FileLogger) WriteLog(message string) {
    // Write to file
}

type DatabaseLogger struct{}

func (l *DatabaseLogger) SaveLog(message string) {
    // Save to database
}

type NetworkLogger struct{}

func (l *NetworkLogger) SendLog(message string) {
    // Send over network
}

// Client code has to know about all variants
func LogMessage(message string, loggerType string) {
    switch loggerType {
    case "file":
        logger := &FileLogger{}
        logger.WriteLog(message)
    case "database":
        logger := &DatabaseLogger{}
        logger.SaveLog(message) // Different method name!
    case "network":
        logger := &NetworkLogger{}
        logger.SendLog(message) // Different method name!
    }
}

// ✅ GOOD: Common interface
type Logger interface {
    Log(message string) error
}

type FileLogger struct {
    path string
}

func (l *FileLogger) Log(message string) error {
    // Write to file
    return nil
}

type DatabaseLogger struct {
    db *Database
}

func (l *DatabaseLogger) Log(message string) error {
    // Save to database
    return nil
}

type NetworkLogger struct {
    endpoint string
}

func (l *NetworkLogger) Log(message string) error {
    // Send over network
    return nil
}

// Client code is simple and polymorphic
func LogMessage(logger Logger, message string) error {
    return logger.Log(message)
}

// Can compose loggers easily
type MultiLogger struct {
    loggers []Logger
}

func (m *MultiLogger) Log(message string) error {
    for _, logger := range m.loggers {
        if err := logger.Log(message); err != nil {
            return err
        }
    }
    return nil
}
```

---

## Change Preventers

**Definition:** Smells indicating that changes in one place require cascading changes elsewhere, making development expensive and risky.

### 1. Divergent Change

**Description:**
A single class requires changes in multiple unrelated methods whenever requirements change. The class has too many reasons to change.

**How to Recognize:**
- Modifying one feature requires changing multiple methods in same class
- Class handles multiple unrelated responsibilities
- Different types of changes affect the same class
- Difficult to describe class without using "and"

**Why It's Bad:**
- Violates Single Responsibility Principle
- Changes become fragmented and error-prone
- Increases risk of breaking unrelated functionality
- Makes testing difficult
- Hard to track change impact

**How to Avoid:**
- Design classes with single, focused responsibility
- Separate concerns early
- Follow domain-driven design principles
- Keep classes cohesive

**How to Fix:**
1. **Extract Class:** Separate responsibilities
2. **Extract Superclass:** Factor out commonality
3. **Extract Module:** Separate into different modules

**Example (Python):**

```python
# ❌ BAD: Divergent change - one class, many reasons to change
class ProductManager:
    def find_products(self, criteria):
        # Database query logic - changes when schema changes
        pass

    def display_products(self, products):
        # Display formatting - changes when UI changes
        pass

    def calculate_price(self, product, quantity):
        # Pricing logic - changes when pricing rules change
        pass

    def apply_discount(self, product, discount_code):
        # Discount logic - changes when promotion rules change
        pass

    def export_to_csv(self, products):
        # Export logic - changes when report format changes
        pass

    def send_low_stock_alert(self, product):
        # Notification logic - changes when notification system changes
        pass

# Changes to database, UI, pricing, discounts, exports, OR notifications
# all require modifying this class!

# ✅ GOOD: Separated concerns
class ProductRepository:
    """Changes only when database schema changes"""
    def find(self, criteria):
        pass

    def save(self, product):
        pass

class ProductDisplay:
    """Changes only when UI requirements change"""
    def format_product(self, product):
        pass

    def format_list(self, products):
        pass

class PricingEngine:
    """Changes only when pricing rules change"""
    def calculate_price(self, product, quantity):
        pass

class DiscountService:
    """Changes only when discount rules change"""
    def apply_discount(self, product, discount_code):
        pass

class ProductExporter:
    """Changes only when export format changes"""
    def export_to_csv(self, products):
        pass

class InventoryNotifier:
    """Changes only when notification system changes"""
    def send_low_stock_alert(self, product):
        pass

# Each class has ONE reason to change
```

---

### 2. Shotgun Surgery

**Description:**
Any change requires making many small modifications across multiple classes. Related functionality is scattered.

**How to Recognize:**
- Single feature change requires edits in many files
- Grep/search required to find all places to modify
- Easy to miss locations when making changes
- Repeated patterns of changes across classes

**Why It's Bad:**
- Time-consuming and error-prone
- Easy to forget locations, causing bugs
- Indicates poor organization
- Makes refactoring risky
- Discourages making improvements

**How to Avoid:**
- Keep related functionality together
- Use proper encapsulation
- Apply Law of Demeter
- Design clear module boundaries

**How to Fix:**
1. **Move Method/Field:** Consolidate related functionality
2. **Inline Class:** Merge scattered classes
3. **Extract Class:** Create central location for feature

**Example (Go):**

```go
// ❌ BAD: Shotgun surgery - adding new field requires changes everywhere
type User struct {
    ID    int
    Name  string
    Email string
    // Need to add "Phone" - requires changes in 20+ places
}

// In user_repository.go
func (r *UserRepository) Create(id int, name, email string) error {
    // Need to add phone parameter
}

// In user_service.go
func (s *UserService) Register(name, email string) error {
    // Need to add phone parameter
}

// In user_validator.go
func ValidateUser(name, email string) error {
    // Need to add phone validation
}

// In user_formatter.go
func FormatUser(id int, name, email string) string {
    // Need to add phone formatting
}

// In api_handler.go
func HandleCreateUser(w http.ResponseWriter, r *http.Request) {
    name := r.FormValue("name")
    email := r.FormValue("email")
    // Need to add phone extraction
}

// 20+ more locations...

// ✅ GOOD: Centralized user data
type User struct {
    ID    int
    Name  string
    Email string
    Phone string // Adding this requires minimal changes
}

type UserRepository struct {
    db *Database
}

func (r *UserRepository) Create(user User) error {
    // Takes complete user object - no change needed
    return r.db.Insert(user)
}

type UserService struct {
    repo *UserRepository
}

func (s *UserService) Register(user User) error {
    // Takes complete user object - no change needed
    if err := user.Validate(); err != nil {
        return err
    }
    return s.repo.Create(user)
}

// Validation is a method on User
func (u *User) Validate() error {
    if u.Name == "" || u.Email == "" || u.Phone == "" {
        return errors.New("missing required fields")
    }
    return nil
}

// Formatting is a method on User
func (u *User) String() string {
    return fmt.Sprintf("%s (%s) - %s", u.Name, u.Email, u.Phone)
}

// API handler
func HandleCreateUser(w http.ResponseWriter, r *http.Request) {
    user := User{
        Name:  r.FormValue("name"),
        Email: r.FormValue("email"),
        Phone: r.FormValue("phone"),
    }
    service.Register(user)
}

// Adding new field only requires:
// 1. Add to User struct
// 2. Update API handler to extract it
// 3. Update validation if needed
// Everything else "just works" because it passes complete User objects
```

---

### 3. Parallel Inheritance Hierarchies

**Description:**
Creating a subclass in one hierarchy forces you to create a corresponding subclass in another hierarchy.

**How to Recognize:**
- Two inheritance trees with similar structure
- Classes in one hierarchy reference classes in parallel hierarchy
- Adding subclass requires changes in multiple hierarchies
- Hierarchy names often mirror each other (UserController/User, OrderManager/Order)

**Why It's Bad:**
- Violates DRY principle
- Doubles maintenance effort
- Increases coupling between hierarchies
- Makes system rigid
- Suggests wrong abstraction

**How to Avoid:**
- Question need for parallel hierarchies
- Use composition over inheritance
- Collapse hierarchies when possible
- Apply strategy pattern

**How to Fix:**
1. **Move Methods/Fields:** Merge one hierarchy into the other
2. **Extract Class:** Create separate class for shared behavior
3. **Replace Inheritance with Delegation:** Use composition

**Example (Python):**

```python
# ❌ BAD: Parallel inheritance hierarchies
# Hierarchy 1: Shapes
class Shape:
    pass

class Circle(Shape):
    pass

class Rectangle(Shape):
    pass

class Triangle(Shape):
    pass

# Hierarchy 2: Shape Renderers (parallel structure)
class ShapeRenderer:
    pass

class CircleRenderer(ShapeRenderer):
    def render(self, circle):
        pass

class RectangleRenderer(ShapeRenderer):
    def render(self, rectangle):
        pass

class TriangleRenderer(ShapeRenderer):
    def render(self, triangle):
        pass

# Adding new shape requires BOTH:
# 1. New shape class (e.g., Pentagon)
# 2. New renderer class (e.g., PentagonRenderer)

# ✅ GOOD: Merged using visitor pattern or move behavior to shape
class Shape:
    @abstractmethod
    def render(self, renderer):
        pass

class Circle(Shape):
    def __init__(self, radius):
        self.radius = radius

    def render(self, renderer):
        return renderer.render_circle(self.radius)

class Rectangle(Shape):
    def __init__(self, width, height):
        self.width = width
        self.height = height

    def render(self, renderer):
        return renderer.render_rectangle(self.width, self.height)

class Triangle(Shape):
    def __init__(self, base, height):
        self.base = base
        self.height = height

    def render(self, renderer):
        return renderer.render_triangle(self.base, self.height)

# Single renderer with methods for each primitive
class Renderer:
    def render_circle(self, radius):
        return f"Circle with radius {radius}"

    def render_rectangle(self, width, height):
        return f"Rectangle {width}x{height}"

    def render_triangle(self, base, height):
        return f"Triangle base:{base} height:{height}"

# Adding new shape only requires:
# 1. New shape class with render() method
# 2. New method in Renderer (but same class)
```

---

## Dispensables

**Definition:** Unnecessary code whose removal would make the codebase cleaner, more efficient, and easier to understand.

### 1. Comments (Excessive)

**Description:**
Methods filled with explanatory comments indicating the code itself is unclear. Comments are often a sign of bad code.

**How to Recognize:**
- Comments explaining what code does (not why)
- Comments for every line or block
- Comments contradicting code (out of date)
- Commented-out code
- TODOs that never get done

**Why It's Bad:**
- Comments go out of date quickly
- Masks underlying readability problems
- Creates maintenance burden (code + comments)
- Developers ignore outdated comments
- Makes code harder to read (noise)

**How to Avoid:**
- Write self-explanatory code
- Use meaningful names
- Extract methods with clear names
- Only comment WHY, not WHAT

**When Comments Are Good:**
- Explaining WHY a non-obvious decision was made
- Documenting public APIs
- Explaining complex algorithms
- Warning about gotchas
- Providing examples of usage

**How to Fix:**
1. **Extract Method:** Replace comment with method name
2. **Rename Method/Variable:** Make intention clear
3. **Introduce Assertion:** Use code to express intent
4. **Delete Comment:** If code is clear

**Example (Go):**

```go
// ❌ BAD: Excessive comments explaining WHAT
func ProcessOrder(order Order) error {
    // Check if order is valid
    if order.Items == nil {
        return errors.New("invalid order")
    }

    // Calculate the total price
    var total float64
    // Loop through all items
    for _, item := range order.Items {
        // Multiply price by quantity
        total += item.Price * float64(item.Quantity)
    }

    // Apply 10% discount if total is over $100
    if total > 100 {
        // Calculate discount
        discount := total * 0.1
        // Subtract from total
        total -= discount
    }

    // TODO: Add tax calculation
    // TODO: Handle shipping
    // TODO: Process payment

    return nil
}

// ✅ GOOD: Self-explanatory code, meaningful comments
func ProcessOrder(order Order) error {
    if err := validateOrder(order); err != nil {
        return err
    }

    subtotal := calculateSubtotal(order)
    total := applyDiscounts(subtotal)

    // Note: Tax and shipping calculation deferred to checkout phase
    // per requirements in TICKET-123. Payment processing happens
    // asynchronously via payment gateway webhook.

    return saveOrder(order, total)
}

func validateOrder(order Order) error {
    if order.Items == nil {
        return errors.New("order must contain items")
    }
    return nil
}

func calculateSubtotal(order Order) float64 {
    var subtotal float64
    for _, item := range order.Items {
        subtotal += item.Price * float64(item.Quantity)
    }
    return subtotal
}

func applyDiscounts(subtotal float64) float64 {
    const bulkOrderThreshold = 100.0
    const bulkOrderDiscount = 0.1

    if subtotal > bulkOrderThreshold {
        return subtotal * (1 - bulkOrderDiscount)
    }
    return subtotal
}
```

---

### 2. Duplicate Code

**Description:**
Identical or nearly identical code appearing in multiple locations. The most pervasive code smell.

**How to Recognize:**
- Copy-pasted code blocks
- Similar logic with minor variations
- Same functionality in different classes
- Repeated patterns

**Why It's Bad:**
- Violates DRY principle
- Multiple points of maintenance
- Bug fixes must be replicated everywhere
- Increases codebase size unnecessarily
- Makes refactoring difficult

**How to Avoid:**
- Extract methods/functions immediately
- Create utilities for common operations
- Use inheritance/composition appropriately
- Apply design patterns

**How to Fix:**
1. **Extract Method:** Pull out duplicated code
2. **Pull Up Method:** Move to superclass
3. **Form Template Method:** Use template method pattern
4. **Extract Class:** Create reusable class

**Example (Python):**

```python
# ❌ BAD: Duplicate code
def create_user(name, email):
    # Validation
    if not name or len(name) < 2:
        raise ValueError("Invalid name")
    if not email or "@" not in email:
        raise ValueError("Invalid email")

    # Logging
    print(f"Creating user: {name}")

    # Database operation
    db.execute("INSERT INTO users (name, email) VALUES (?, ?)", (name, email))

    # Notification
    send_email(email, "Welcome!")

def create_admin(name, email):
    # Validation (duplicated!)
    if not name or len(name) < 2:
        raise ValueError("Invalid name")
    if not email or "@" not in email:
        raise ValueError("Invalid email")

    # Logging (duplicated!)
    print(f"Creating admin: {name}")

    # Database operation (similar)
    db.execute("INSERT INTO users (name, email, is_admin) VALUES (?, ?, ?)",
               (name, email, True))

    # Notification (duplicated!)
    send_email(email, "Welcome!")

# ✅ GOOD: Extract common functionality
def validate_user_input(name: str, email: str):
    if not name or len(name) < 2:
        raise ValueError("Invalid name")
    if not email or "@" not in email:
        raise ValueError("Invalid email")

def log_user_creation(name: str, user_type: str):
    print(f"Creating {user_type}: {name}")

def send_welcome_email(email: str):
    send_email(email, "Welcome!")

def create_user(name: str, email: str, is_admin: bool = False):
    validate_user_input(name, email)

    user_type = "admin" if is_admin else "user"
    log_user_creation(name, user_type)

    if is_admin:
        db.execute(
            "INSERT INTO users (name, email, is_admin) VALUES (?, ?, ?)",
            (name, email, True)
        )
    else:
        db.execute(
            "INSERT INTO users (name, email) VALUES (?, ?)",
            (name, email)
        )

    send_welcome_email(email)

# Even better: Use a class/object approach
@dataclass
class User:
    name: str
    email: str
    is_admin: bool = False

    def validate(self):
        if not self.name or len(self.name) < 2:
            raise ValueError("Invalid name")
        if not self.email or "@" not in self.email:
            raise ValueError("Invalid email")

    def save(self):
        self.validate()
        db.insert("users", {
            "name": self.name,
            "email": self.email,
            "is_admin": self.is_admin
        })
        self.send_welcome()

    def send_welcome(self):
        send_email(self.email, "Welcome!")
```

---

### 3. Lazy Class

**Description:**
A class that doesn't do enough to justify its existence and maintenance cost.

**How to Recognize:**
- Class with only one or two methods
- Class that just delegates to another class
- Class created for "future flexibility" but never needed
- Class that could be a function

**Why It's Bad:**
- Increases complexity without benefit
- Extra file/class to maintain
- Cognitive overhead
- Violates YAGNI principle

**How to Avoid:**
- Create classes when they provide value
- Don't create classes "just in case"
- Start with functions, extract classes when needed
- Delete classes that lose purpose

**How to Fix:**
1. **Inline Class:** Merge into related class
2. **Collapse Hierarchy:** Remove unnecessary abstraction
3. **Convert to Function:** Make it a simple function

**Example (Go):**

```go
// ❌ BAD: Lazy class - doesn't justify existence
type StringValidator struct{}

func (v *StringValidator) IsEmpty(s string) bool {
    return len(s) == 0
}

// Usage requires unnecessary ceremony
validator := &StringValidator{}
if validator.IsEmpty(name) {
    return errors.New("name is empty")
}

// Another lazy class
type MathHelper struct{}

func (h *MathHelper) Add(a, b int) int {
    return a + b
}

// ✅ GOOD: Just use functions
func IsEmpty(s string) bool {
    return len(s) == 0
}

// Usage is simpler
if IsEmpty(name) {
    return errors.New("name is empty")
}

// Or just use it inline
if len(name) == 0 {
    return errors.New("name is empty")
}

// Classes should have sufficient behavior to justify existence
type UserValidator struct {
    minNameLength    int
    emailPattern     *regexp.Regexp
    bannedDomains    []string
    profanityFilter  *ProfanityFilter
}

func (v *UserValidator) ValidateName(name string) error {
    if len(name) < v.minNameLength {
        return fmt.Errorf("name must be at least %d characters", v.minNameLength)
    }
    if v.profanityFilter.Contains(name) {
        return errors.New("name contains inappropriate content")
    }
    return nil
}

func (v *UserValidator) ValidateEmail(email string) error {
    if !v.emailPattern.MatchString(email) {
        return errors.New("invalid email format")
    }
    domain := getDomain(email)
    if contains(v.bannedDomains, domain) {
        return fmt.Errorf("domain %s is not allowed", domain)
    }
    return nil
}

// This class justifies its existence - it has state and multiple related behaviors
```

---

### 4. Data Class

**Description:**
A class that only contains fields and basic getters/setters with no business logic. Pure data containers.

**How to Recognize:**
- Only public fields or getters/setters
- No meaningful methods
- Other classes manipulate its data
- "Anemic domain model"

**Why It's Bad:**
- Violates encapsulation
- Business logic scattered across other classes
- Difficult to enforce invariants
- Data and behavior should be together

**When Data Classes Are OK:**
- DTOs (Data Transfer Objects) for API boundaries
- Value objects (immutable)
- Configuration objects
- Database entities (with proper boundaries)

**How to Avoid:**
- Add behavior to data classes
- Encapsulate related operations
- Protect invariants
- Make classes responsible for their data

**How to Fix:**
1. **Move Method:** Bring related behavior into class
2. **Extract Method:** Create methods for data manipulation
3. **Encapsulate Field:** Make fields private, add meaningful methods

**Example (Python):**

```python
# ❌ BAD: Anemic data class
class Order:
    def __init__(self):
        self.items = []
        self.total = 0.0
        self.discount = 0.0
        self.status = "pending"

# Business logic scattered in other classes
class OrderService:
    def calculate_total(self, order):
        order.total = sum(item.price * item.quantity for item in order.items)

    def apply_discount(self, order, discount_code):
        if discount_code == "SAVE10":
            order.discount = order.total * 0.1
            order.total -= order.discount

    def complete_order(self, order):
        if order.total <= 0:
            raise ValueError("Invalid total")
        order.status = "completed"

# ✅ GOOD: Rich domain model with behavior
class Order:
    def __init__(self):
        self._items = []
        self._status = OrderStatus.PENDING
        self._discount_applied = None

    def add_item(self, item: OrderItem):
        """Add item and recalculate total"""
        if self._status != OrderStatus.PENDING:
            raise ValueError("Cannot modify completed order")
        self._items.append(item)

    def remove_item(self, item: OrderItem):
        """Remove item and recalculate total"""
        if self._status != OrderStatus.PENDING:
            raise ValueError("Cannot modify completed order")
        self._items.remove(item)

    def calculate_subtotal(self) -> Decimal:
        """Calculate subtotal before discounts"""
        return sum(item.price * item.quantity for item in self._items)

    def apply_discount(self, discount: Discount):
        """Apply discount code, enforcing business rules"""
        if self._status != OrderStatus.PENDING:
            raise ValueError("Cannot apply discount to completed order")

        if not discount.is_valid():
            raise ValueError("Discount is not valid")

        subtotal = self.calculate_subtotal()
        if subtotal < discount.minimum_order_value:
            raise ValueError(f"Order must be at least {discount.minimum_order_value}")

        self._discount_applied = discount

    def calculate_total(self) -> Decimal:
        """Calculate final total with discounts"""
        subtotal = self.calculate_subtotal()

        if self._discount_applied:
            discount_amount = self._discount_applied.calculate_discount(subtotal)
            return subtotal - discount_amount

        return subtotal

    def complete(self):
        """Complete the order, enforcing business rules"""
        if not self._items:
            raise ValueError("Cannot complete empty order")

        if self.calculate_total() <= 0:
            raise ValueError("Order total must be positive")

        self._status = OrderStatus.COMPLETED

    def is_completed(self) -> bool:
        return self._status == OrderStatus.COMPLETED

    @property
    def items(self) -> List[OrderItem]:
        """Return read-only copy of items"""
        return list(self._items)

# Now business logic is where it belongs - in the Order class
# Invariants are protected (can't modify completed order)
# The class is responsible for its own data
```

---

### 5. Dead Code

**Description:**
Variables, parameters, fields, methods, or classes that are no longer used or executed.

**How to Recognize:**
- Unreferenced code
- Unused variables/parameters
- Obsolete methods
- Branches that can't be reached
- IDE warnings about unused code

**Why It's Bad:**
- Increases cognitive load
- Creates confusion
- Maintenance burden
- May mislead developers
- Bloats codebase

**How to Avoid:**
- Delete code immediately when no longer needed
- Use IDE warnings
- Run static analysis tools
- Regular code audits

**How to Fix:**
1. **Delete the code** - Version control preserves history
2. **Remove unused parameters**
3. **Eliminate unreachable branches**

**Example (Go):**

```go
// ❌ BAD: Dead code everywhere
type UserService struct {
    db       *Database
    cache    *Cache      // Never used
    oldCache *OldCache   // Legacy, should be removed
}

func (s *UserService) CreateUser(name, email string, unused int) error {
    // "unused" parameter never used

    // This old validation logic is commented out but still here
    // if name == "" {
    //     return errors.New("name required")
    // }

    user := User{Name: name, Email: email}

    // This branch is never reached because we removed the feature
    if false {
        s.sendDeprecatedNotification(user)
    }

    return s.db.Insert(user)
}

// This method is never called anywhere
func (s *UserService) sendDeprecatedNotification(user User) {
    // ... obsolete code ...
}

// This method was replaced by CreateUser but never deleted
func (s *UserService) CreateUserOld(name, email string) error {
    return errors.New("deprecated")
}

// ✅ GOOD: Clean, only what's used
type UserService struct {
    db *Database
}

func (s *UserService) CreateUser(name, email string) error {
    user := User{Name: name, Email: email}
    return s.db.Insert(user)
}

// If you need the old implementation, it's in version control:
// git log -p -- user_service.go
```

---

### 6. Speculative Generality

**Description:**
Unused classes, methods, fields, or parameters created "just in case" for future needs that never materialize.

**How to Recognize:**
- "Future-proof" abstractions
- Unused parameters or methods
- Complex designs for simple problems
- "We might need this later"
- Generic solutions without current use cases

**Why It's Bad:**
- Violates YAGNI (You Aren't Gonna Need It)
- Increases complexity without benefit
- Wastes development time
- Creates maintenance burden
- Future requirements may differ anyway

**How to Avoid:**
- Build for current needs only
- Refactor when new requirements emerge
- Trust your ability to refactor later
- Embrace simplicity

**How to Fix:**
1. **Delete unused abstractions**
2. **Collapse unnecessary hierarchies**
3. **Inline unnecessary methods**
4. **Remove unused parameters**

**Example (Python):**

```python
# ❌ BAD: Speculative generality
class AbstractDataProcessor(ABC):
    """We might want different processors in the future..."""

    @abstractmethod
    def process(self, data):
        pass

    @abstractmethod
    def validate(self, data):
        pass

    @abstractmethod
    def transform(self, data):
        pass

    @abstractmethod
    def format_output(self, data):
        pass

class CSVDataProcessor(AbstractDataProcessor):
    """Only implementation, but we made it abstract just in case"""

    def process(self, data):
        validated = self.validate(data)
        transformed = self.transform(validated)
        return self.format_output(transformed)

    def validate(self, data):
        return data  # No validation needed yet

    def transform(self, data):
        return data  # No transformation needed yet

    def format_output(self, data):
        return "\n".join(data)

# Unused parameters "just in case"
def save_data(data, format="csv", compression=None, encryption=None,
              backup=True, versioning=False):
    # Only uses data and format, rest are unused
    with open(f"data.{format}", "w") as f:
        f.write(data)

# ✅ GOOD: Build what you need now
def process_csv(data: List[str]) -> str:
    """Process CSV data. Simple and clear."""
    return "\n".join(data)

def save_csv(data: str):
    """Save CSV data. Add parameters when needed."""
    with open("data.csv", "w") as f:
        f.write(data)

# When you actually need multiple formats, THEN refactor:
class CSVProcessor:
    def process(self, data: List[str]) -> str:
        return "\n".join(data)

class JSONProcessor:
    def process(self, data: dict) -> str:
        return json.dumps(data, indent=2)

# Simple, focused, solves actual problem
```

---

## Couplers

**Definition:** Code smells that contribute to excessive coupling between classes or demonstrate problems with delegation.

### 1. Feature Envy

**Description:**
A method accesses data of another object more than its own data, suggesting the method is in the wrong class.

**How to Recognize:**
- Method uses more getters from other object than from own class
- Method seems more interested in other class
- Method would be simpler if it lived elsewhere
- Lots of calls across object boundary

**Why It's Bad:**
- Violates "Tell, Don't Ask" principle
- Increases coupling
- Logic is in wrong place
- Makes changes difficult

**How to Avoid:**
- Put behavior with the data it uses
- Follow "Tell, Don't Ask"
- Methods should primarily use their own data
- Consider object responsibilities

**How to Fix:**
1. **Move Method:** Move to the class with the data
2. **Extract Method:** Break apart and move pieces

**Example (Go):**

```go
// ❌ BAD: Feature envy
type Order struct {
    items []OrderItem
}

type OrderItem struct {
    Price    float64
    Quantity int
    TaxRate  float64
}

type OrderCalculator struct{}

// This method is envious of OrderItem's data
func (c *OrderCalculator) CalculateItemTotal(item OrderItem) float64 {
    subtotal := item.Price * float64(item.Quantity)
    tax := subtotal * item.TaxRate
    return subtotal + tax
}

// This method is envious of Order's data
func (c *OrderCalculator) CalculateOrderTotal(order Order) float64 {
    var total float64
    for _, item := range order.items {
        total += c.CalculateItemTotal(item)
    }
    return total
}

// ✅ GOOD: Methods with their data
type OrderItem struct {
    Price    float64
    Quantity int
    TaxRate  float64
}

// OrderItem knows how to calculate its own total
func (item OrderItem) CalculateTotal() float64 {
    subtotal := item.Price * float64(item.Quantity)
    tax := subtotal * item.TaxRate
    return subtotal + tax
}

type Order struct {
    items []OrderItem
}

// Order uses items' behavior instead of accessing their data
func (o *Order) CalculateTotal() float64 {
    var total float64
    for _, item := range o.items {
        total += item.CalculateTotal() // Tell, don't ask
    }
    return total
}

// No need for OrderCalculator at all!
```

---

### 2. Inappropriate Intimacy

**Description:**
One class excessively accesses internal details of another class, creating tight coupling.

**How to Recognize:**
- Classes accessing each other's private fields (even via getters)
- Bidirectional dependencies
- Classes that change together frequently
- Hard to reuse classes independently

**Why It's Bad:**
- High coupling makes changes difficult
- Reduces independence and reusability
- Violates encapsulation
- Creates maintenance nightmares

**How to Avoid:**
- Use proper encapsulation
- Define clear interfaces
- Minimize dependencies
- Apply information hiding

**How to Fix:**
1. **Move Method/Field:** Consolidate related functionality
2. **Extract Class:** Create mediator for shared behavior
3. **Hide Delegate:** Add delegation methods
4. **Replace Inheritance with Delegation:** Break intimate relationships

**Example (Python):**

```python
# ❌ BAD: Inappropriate intimacy
class User:
    def __init__(self):
        self.name = ""
        self.email = ""
        self.preferences = {}
        self.last_login = None
        self.login_count = 0

class UserProfile:
    def __init__(self, user):
        self.user = user

    def display(self):
        # Intimately knows User's internals
        print(f"Name: {self.user.name}")
        print(f"Email: {self.user.email}")
        print(f"Last login: {self.user.last_login}")
        print(f"Total logins: {self.user.login_count}")

        # Directly manipulates User's preferences
        if "theme" not in self.user.preferences:
            self.user.preferences["theme"] = "light"

class UserStats:
    def __init__(self, user):
        self.user = user

    def is_active(self):
        # Intimately accesses user internals
        if self.user.last_login is None:
            return False
        days_since = (datetime.now() - self.user.last_login).days
        return days_since < 30 and self.user.login_count > 5

# ✅ GOOD: Proper encapsulation
class User:
    def __init__(self):
        self._name = ""
        self._email = ""
        self._preferences = {}
        self._last_login = None
        self._login_count = 0

    # Public interface, hides internals
    def get_display_info(self):
        return {
            "name": self._name,
            "email": self._email,
            "last_login": self._last_login,
            "login_count": self._login_count
        }

    def get_preference(self, key, default=None):
        return self._preferences.get(key, default)

    def set_preference(self, key, value):
        self._preferences[key] = value

    def is_active(self):
        """User knows if it's active"""
        if self._last_login is None:
            return False
        days_since = (datetime.now() - self._last_login).days
        return days_since < 30 and self._login_count > 5

class UserProfile:
    def __init__(self, user):
        self.user = user

    def display(self):
        # Uses public interface only
        info = self.user.get_display_info()
        print(f"Name: {info['name']}")
        print(f"Email: {info['email']}")
        print(f"Last login: {info['last_login']}")
        print(f"Total logins: {info['login_count']}")

        # Uses public method
        if self.user.get_preference("theme") is None:
            self.user.set_preference("theme", "light")

# UserStats not needed - functionality moved to User
```

---

### 3. Message Chains

**Description:**
Long chains of method calls like `a.b().c().d()`, creating tight coupling and fragility.

**How to Recognize:**
- Multiple chained getter calls
- Code like `order.getCustomer().getAddress().getZipCode()`
- Changes in intermediate classes break many callers
- Violates Law of Demeter ("don't talk to strangers")

**Why It's Bad:**
- Creates tight coupling through chain
- Changes anywhere in chain break code
- Hard to test (must mock entire chain)
- Exposes internal structure

**How to Avoid:**
- Follow Law of Demeter
- Objects should only call methods on:
  - Themselves
  - Parameters
  - Objects they create
  - Direct components
- Use delegation methods

**How to Fix:**
1. **Hide Delegate:** Add delegation methods
2. **Extract Method:** Create helper for chain
3. **Move Method:** Put behavior closer to data

**Example (Go):**

```go
// ❌ BAD: Message chains
type Customer struct {
    address *Address
}

func (c *Customer) GetAddress() *Address {
    return c.address
}

type Address struct {
    city    string
    zipCode string
}

func (a *Address) GetCity() string {
    return a.city
}

func (a *Address) GetZipCode() string {
    return a.zipCode
}

type Order struct {
    customer *Customer
}

func (o *Order) GetCustomer() *Customer {
    return o.customer
}

// Client code has long chains
func CalculateShipping(order *Order) float64 {
    // Violates Law of Demeter - talks to strangers
    zipCode := order.GetCustomer().GetAddress().GetZipCode()
    city := order.GetCustomer().GetAddress().GetCity()

    // Fragile - any change in chain breaks this
    return calculateByZipCode(zipCode, city)
}

// ✅ GOOD: Hide delegates
type Customer struct {
    address *Address
}

// Delegate directly to what's needed
func (c *Customer) GetZipCode() string {
    if c.address == nil {
        return ""
    }
    return c.address.zipCode
}

func (c *Customer) GetCity() string {
    if c.address == nil {
        return ""
    }
    return c.address.city
}

type Order struct {
    customer *Customer
}

// Order provides direct access to needed info
func (o *Order) GetShippingZipCode() string {
    if o.customer == nil {
        return ""
    }
    return o.customer.GetZipCode()
}

func (o *Order) GetShippingCity() string {
    if o.customer == nil {
        return ""
    }
    return o.customer.GetCity()
}

// Client code is simple and doesn't know about chain
func CalculateShipping(order *Order) float64 {
    zipCode := order.GetShippingZipCode()
    city := order.GetShippingCity()
    return calculateByZipCode(zipCode, city)
}

// Even better: Have Order calculate its own shipping
func (o *Order) CalculateShipping() float64 {
    return calculateByZipCode(o.GetShippingZipCode(), o.GetShippingCity())
}
```

---

### 4. Middle Man

**Description:**
A class that exists primarily to delegate to another class, providing little value of its own.

**How to Recognize:**
- Most methods just delegate to another class
- Class acts as unnecessary wrapper
- Provides no additional functionality
- Created for "abstraction" without purpose

**Why It's Bad:**
- Adds unnecessary layer
- Complicates codebase
- Makes debugging harder
- No actual value provided

**How to Avoid:**
- Don't create wrappers without reason
- Delegation should add value (transformation, caching, security, etc.)
- Question abstractions
- Direct access is sometimes better

**How to Fix:**
1. **Remove Middle Man:** Give clients direct access
2. **Inline Method:** Eliminate delegation
3. **Replace Delegation with Inheritance:** If appropriate

**Example (Python):**

```python
# ❌ BAD: Middle man doing nothing useful
class PersonManager:
    def __init__(self, person):
        self._person = person

    def get_name(self):
        return self._person.get_name()  # Just delegates

    def set_name(self, name):
        self._person.set_name(name)  # Just delegates

    def get_age(self):
        return self._person.get_age()  # Just delegates

    def set_age(self, age):
        self._person.set_age(age)  # Just delegates

# Client code goes through unnecessary layer
person = Person()
manager = PersonManager(person)
name = manager.get_name()  # Why not use person directly?

# ✅ GOOD: Remove middle man
class Person:
    def __init__(self, name="", age=0):
        self._name = name
        self._age = age

    @property
    def name(self):
        return self._name

    @name.setter
    def name(self, value):
        self._name = value

    @property
    def age(self):
        return self._age

    @age.setter
    def age(self, value):
        if value < 0:
            raise ValueError("Age cannot be negative")
        self._age = value

# Client code uses Person directly
person = Person()
name = person.name  # Simple and direct

# Middle man is OK when it adds value:
class CachedPersonRepository:
    def __init__(self, repository, cache):
        self._repository = repository
        self._cache = cache

    def get_person(self, person_id):
        # Adds caching value
        if person_id in self._cache:
            return self._cache[person_id]

        person = self._repository.get_person(person_id)
        self._cache[person_id] = person
        return person

class SecurePersonRepository:
    def __init__(self, repository, auth):
        self._repository = repository
        self._auth = auth

    def get_person(self, person_id):
        # Adds security value
        if not self._auth.can_access_person(person_id):
            raise PermissionError("Access denied")

        return self._repository.get_person(person_id)

# These middle men justify their existence by adding functionality
```

---

## Detection Strategy

### Automated Tools

**Static Analysis:**
- **SonarQube:** Comprehensive code smell detection
- **ESLint/Pylint:** Language-specific linters
- **Go vet/staticcheck:** Go static analysis
- **Codacy, CodeClimate:** Cloud-based analysis

**Metrics Tools:**
- **Cyclomatic Complexity:** Detect long methods, complex conditionals
- **Code Coverage:** Find dead code
- **Duplication Detection:** Find duplicate code
- **Dependency Analysis:** Detect coupling issues

### Manual Detection

**Code Review Checklist:**
1. Is this method/class too long?
2. Does this class have too many responsibilities?
3. Is there duplicated code?
4. Are there long parameter lists?
5. Is there excessive coupling?
6. Are there magic numbers or primitives for domain concepts?
7. Is dead code present?
8. Are there excessive comments?

**Regular Practices:**
- Weekly/monthly code smell reviews
- Refactoring time allocated in sprint
- Pair programming for immediate detection
- Architectural review sessions

---

## Prevention Guidelines

### Design Phase

1. **Apply SOLID Principles:**
   - **S**ingle Responsibility
   - **O**pen/Closed
   - **L**iskov Substitution
   - **I**nterface Segregation
   - **D**ependency Inversion

2. **Follow Design Patterns:**
   - Use appropriate patterns for common problems
   - Don't over-engineer with patterns

3. **Domain-Driven Design:**
   - Model business domain accurately
   - Use ubiquitous language
   - Bounded contexts prevent bloat

### Development Phase

1. **Write Clean Code First:**
   - Meaningful names
   - Small, focused functions
   - Proper abstractions
   - Clear separation of concerns

2. **Refactor Continuously:**
   - Boy Scout Rule: "Leave code cleaner than you found it"
   - Refactor before adding features
   - Don't accumulate technical debt

3. **Test-Driven Development:**
   - Tests drive better design
   - Testable code is usually well-designed
   - Difficult to test = code smell

4. **Pair Programming:**
   - Immediate code smell detection
   - Knowledge sharing
   - Better design decisions

### Review Phase

1. **Code Review Focus:**
   - Check for code smells explicitly
   - Don't just check functionality
   - Require refactoring before merge

2. **Automated Checks:**
   - Fail builds on code smells
   - Enforce metrics thresholds
   - Use pre-commit hooks

3. **Regular Audits:**
   - Schedule refactoring time
   - Track technical debt
   - Prioritize problem areas

---

## Refactoring Priorities

### Critical (Fix Immediately)

1. **Duplicate Code** - Highest risk for inconsistent bugs
2. **Long Method** - Hard to understand and maintain
3. **Large Class** - Multiple responsibilities = high change risk
4. **Shotgun Surgery** - Changes cascade dangerously

### High Priority (Fix Soon)

5. **Feature Envy** - Logic in wrong place
6. **Switch Statements** - Prevents extensibility
7. **Primitive Obsession** - Loses type safety
8. **Long Parameter List** - Error-prone
9. **Inappropriate Intimacy** - High coupling

### Medium Priority (Fix When Touching Code)

10. **Data Clumps** - Missing abstractions
11. **Message Chains** - Fragile coupling
12. **Divergent Change** - Multiple responsibilities
13. **Refused Bequest** - Wrong hierarchy
14. **Middle Man** - Unnecessary complexity

### Low Priority (Fix During Refactoring Sessions)

15. **Lazy Class** - Small complexity increase
16. **Data Class** - Can be appropriate in some contexts
17. **Dead Code** - Easy to identify and remove
18. **Speculative Generality** - Low risk if not actively developed
19. **Comments** - Style issue unless excessive
20. **Temporary Field** - Isolated impact

---

## Conclusion

Code smells are not bugs—they're indicators of design problems that will cause issues later. The key is **prevention through good design and continuous refactoring**.

### Key Takeaways

1. **Recognize smells early** - Address them before they compound
2. **Refactor continuously** - Don't accumulate technical debt
3. **Use automation** - Tools catch what humans miss
4. **Make it a habit** - Include smell detection in your workflow
5. **Prioritize wisely** - Not all smells are equally urgent

### Resources

- **Refactoring.Guru:** https://refactoring.guru/refactoring/smells
- **Martin Fowler's Refactoring:** Classic book on refactoring techniques
- **Clean Code by Robert Martin:** Comprehensive guide to writing clean code
- **SonarQube Documentation:** https://www.sonarqube.org/

### Integration with AI-Pack

- **Reviewer agents** should flag code smells automatically
- **Pre-commit hooks** should run smell detection
- **Performance reports** should include code quality metrics
- **Documentation** should reference this guide

---

**Version History:**
- 1.0 (2026-01-27): Initial comprehensive guide

**Maintained by:** AI-Pack Quality Team
**Questions?** See docs/quality/ for related guidelines
