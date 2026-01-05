# C++ Language-Specific Rules

> Based on Scott Meyers' Effective C++ series and C++ Core Guidelines

## Formatting Standards (C++ Specific)

**Indentation:** **2 spaces** (no tabs)

This follows the Google C++ Style Guide, LLVM, and Chromium standards. C++'s verbose syntax (templates, namespaces, nested classes) benefits from narrower indentation.

**Note:** Cortexa LLC uses **language-specific indentation standards**:
- **C++**: 2 spaces (this document)
- **Python**: 4 spaces (PEP 8 - see lang-python.md)
- **JavaScript/TypeScript**: 2 spaces (see lang-javascript.md)
- **Java**: 4 spaces (see lang-java.md)
- **Kotlin**: 4 spaces (see lang-kotlin.md)

**Example:**
```cpp
namespace sourcerer {
namespace core {
  class Binary {
  public:
    Binary(const std::vector<uint8_t>& data, uint32_t load_address)
        : data_(data), load_address_(load_address) {
      // 2-space indentation throughout
    }

  private:
    std::vector<uint8_t> data_;
    uint32_t load_address_;
  };
}  // namespace core
}  // namespace sourcerer
```

---

## Overview

This file contains C++-specific best practices from:
- **Effective C++** (55 items) - 3rd Edition by Scott Meyers
- **Effective Modern C++** (42 items) - C++11/14 by Scott Meyers
- **C++ Core Guidelines** - ISO C++ community standards
- **More Effective C++** (35 items) - Advanced topics by Scott Meyers

## Chapter 1: Accustoming Yourself to C++

### Item 1: View C++ as a Federation of Languages

**Principle:** C++ is not one language but four sub-languages:
- **C** - Blocks, statements, preprocessor, built-in types, arrays, pointers
- **Object-Oriented C++** - Classes, encapsulation, inheritance, polymorphism, virtual functions
- **Template C++** - Generic programming, template metaprogramming
- **STL** - Containers, iterators, algorithms, function objects

**Application:**
- Rules change depending on which sub-language you're using
- Pass-by-value is efficient for C, but pass-by-reference-to-const is better for objects
- STL iterators and function objects are modeled on C pointers
- Understand the context when applying guidelines

### Item 2: Prefer const, enum, and inline to #define

**Principle:** "Prefer the compiler to the preprocessor"

**Why:**
- `#define` doesn't respect scope and can't be encapsulated
- Preprocessor macros don't appear in symbol tables (harder to debug)
- Can lead to multiple copies of the constant in object code

**Apply:**
```cpp
// BAD - preprocessor constant
#define ASPECT_RATIO 1.653

// GOOD - const variable
const double AspectRatio = 1.653;

// For class-specific constants
class GamePlayer {
private:
    static const int NumTurns = 5;  // Declaration
    int scores[NumTurns];
};

// Use enum hack when in-class initialization not allowed
class GamePlayer {
private:
    enum { NumTurns = 5 };  // Makes NumTurns a symbolic name for 5
    int scores[NumTurns];
};
```

**For function-like macros, use inline functions:**
```cpp
// BAD - macro
#define CALL_WITH_MAX(a, b) f((a) > (b) ? (a) : (b))

// GOOD - template inline function
template<typename T>
inline void callWithMax(const T& a, const T& b) {
    f(a > b ? a : b);
}
```

### Item 3: Use const Whenever Possible

**Principle:** The `const` keyword is remarkably versatile - use it to communicate to both compilers and other programmers.

**Applications:**

**Const with Pointers:**
```cpp
char greeting[] = "Hello";
char* p = greeting;                  // Non-const pointer, non-const data
const char* p = greeting;            // Non-const pointer, const data
char* const p = greeting;            // Const pointer, non-const data
const char* const p = greeting;      // Const pointer, const data
```

**Const with Iterators:**
```cpp
std::vector<int> vec;

// Iterator acts like T* pointer
const std::vector<int>::iterator iter = vec.begin();  // Like T* const
*iter = 10;  // OK - changes what iter points to
++iter;      // ERROR - iter itself is const

// Const_iterator acts like const T* pointer
std::vector<int>::const_iterator cIter = vec.begin();  // Like const T*
*cIter = 10;  // ERROR - can't change what cIter points to
++cIter;      // OK - can move cIter
```

**Const Member Functions:**
```cpp
class TextBlock {
public:
    // Const member function - can be called on const objects
    const char& operator[](std::size_t position) const {
        return text[position];
    }

    // Non-const member function - can only be called on non-const objects
    char& operator[](std::size_t position) {
        return text[position];
    }

private:
    std::string text;
};

// Usage
TextBlock tb("Hello");
std::cout << tb[0];  // Calls non-const operator[]

const TextBlock ctb("World");
std::cout << ctb[0];  // Calls const operator[]
```

**Const and Thread Safety:**
- Const member functions should be thread-safe
- Use `mutable` for implementation details that need to change in const functions

**Avoid const_cast:**
- Use it rarely and document why
- Casting away const to modify an object that was originally const is undefined behavior

### Item 4: Make Sure Objects are Initialized Before Use

**Principle:** Always initialize objects before use to avoid undefined behavior.

**Built-in Types:**
```cpp
int x = 0;                       // Manual initialization of int
const char* text = "A C-style string";  // Manual initialization of pointer
double d;                        // DANGEROUS - uninitialized
std::cin >> d;                   // Reading into uninitialized variable
```

**Member Initialization Lists:**
```cpp
// BAD - assignment, not initialization
class PhoneNumber { /* ... */ };
class ABEntry {
public:
    ABEntry(const std::string& name, const std::string& address,
            const std::list<PhoneNumber>& phones) {
        theName = name;         // Assignments, not initializations
        theAddress = address;
        thePhones = phones;
        numTimesConsulted = 0;
    }

private:
    std::string theName;
    std::string theAddress;
    std::list<PhoneNumber> thePhones;
    int numTimesConsulted;
};

// GOOD - use member initialization list
class ABEntry {
public:
    ABEntry(const std::string& name, const std::string& address,
            const std::list<PhoneNumber>& phones)
        : theName(name),           // Initialization
          theAddress(address),
          thePhones(phones),
          numTimesConsulted(0)
    { }

private:
    std::string theName;
    std::string theAddress;
    std::list<PhoneNumber> thePhones;
    int numTimesConsulted;
};
```

**Why Member Initialization Lists:**
- More efficient (one call instead of two)
- Required for const and reference members
- Required for types without default constructors

**Order matters:**
- Members initialized in order they're declared in class
- Not the order they appear in initialization list

**Static Objects:**
```cpp
// BAD - order of initialization across translation units is undefined
// file1.cpp
extern FileSystem tfs;  // Defined in another file

// file2.cpp
FileSystem tfs;         // When is this initialized relative to file1?

// GOOD - use local static objects (Meyers Singleton)
FileSystem& tfs() {
    static FileSystem fs;  // Initialized on first call
    return fs;
}
```

## Chapter 2: Constructors, Destructors, and Assignment Operators

### Item 5: Know What Functions C++ Silently Writes and Calls

**Compiler-Generated Functions:**

If you don't declare them, the compiler generates:
- Default constructor (if no constructors declared)
- Copy constructor
- Copy assignment operator
- Destructor

```cpp
class Empty { };

// Equivalent to:
class Empty {
public:
    Empty() { ... }                            // Default constructor
    Empty(const Empty& rhs) { ... }            // Copy constructor
    ~Empty() { ... }                           // Destructor
    Empty& operator=(const Empty& rhs) { ... } // Copy assignment operator
};
```

**When Compiler Won't Generate:**
- Won't generate copy assignment if class contains reference members
- Won't generate copy assignment if class contains const members
- Won't generate copy assignment if base class has private copy assignment

### Item 6: Explicitly Disallow Functions You Don't Want

**Pre-C++11 Approach:**
```cpp
class HomeForSale {
public:
    // ...

private:
    HomeForSale(const HomeForSale&);             // Declare but don't define
    HomeForSale& operator=(const HomeForSale&);  // Declare but don't define
};
```

**Modern C++11/14 Approach (Preferred):**
```cpp
class HomeForSale {
public:
    HomeForSale() = default;
    HomeForSale(const HomeForSale&) = delete;             // Explicitly delete
    HomeForSale& operator=(const HomeForSale&) = delete;  // Explicitly delete

    // For move semantics in modern C++
    HomeForSale(HomeForSale&&) = delete;
    HomeForSale& operator=(HomeForSale&&) = delete;
};
```

### Item 7: Declare Destructors Virtual in Polymorphic Base Classes

**Principle:** If a class is designed to be used polymorphically, it should have a virtual destructor.

```cpp
// BAD - non-virtual destructor in base class
class TimeKeeper {
public:
    TimeKeeper();
    ~TimeKeeper();  // NON-VIRTUAL - PROBLEM!
    // ...
};

class AtomicClock: public TimeKeeper { /* ... */ };

TimeKeeper* ptk = new AtomicClock;
// ...
delete ptk;  // UNDEFINED BEHAVIOR - only ~TimeKeeper called, not ~AtomicClock

// GOOD - virtual destructor
class TimeKeeper {
public:
    TimeKeeper();
    virtual ~TimeKeeper();  // Virtual destructor
    // ...
};
```

**Rules:**
- Polymorphic base classes should declare virtual destructors
- If a class has any virtual functions, it should have a virtual destructor
- Classes not designed to be base classes or not designed to be used polymorphically should NOT declare virtual destructors (adds vtable overhead)

### Item 8: Prevent Exceptions from Leaving Destructors

**Principle:** Destructors should never emit exceptions. If called functions may throw, catch exceptions and swallow them or terminate.

**Why:**
- If destructor throws during stack unwinding (from another exception), program terminates
- Gives no opportunity for cleanup
- Can lead to resource leaks

```cpp
// BAD - destructor can throw
class Widget {
public:
    ~Widget() {
        db.close();  // May throw!
    }
};

// GOOD - catch and handle
class Widget {
public:
    ~Widget() {
        try {
            db.close();
        } catch (...) {
            // Log error and swallow exception
            makeLogEntry("Error closing database");
        }
    }

    // Even better - provide explicit close() for clients to handle
    void close() {
        db.close();
        closed = true;
    }

    ~Widget() {
        if (!closed) {
            try {
                db.close();
            } catch (...) {
                // Log and swallow
            }
        }
    }

private:
    DBConnection db;
    bool closed = false;
};
```

### Item 9: Never Call Virtual Functions During Construction or Destruction

**Principle:** During construction/destruction, virtual functions don't behave virtually - they resolve to the constructor's/destructor's class version.

```cpp
class Transaction {
public:
    Transaction() {
        logTransaction();  // Calls Transaction::logTransaction, not derived version!
    }

    virtual void logTransaction() const = 0;
};

class BuyTransaction: public Transaction {
public:
    virtual void logTransaction() const {
        // Log buy-specific transaction
    }
};

BuyTransaction b;  // Constructor calls Transaction::logTransaction, not BuyTransaction::logTransaction!
```

**Solution:**
```cpp
class Transaction {
public:
    explicit Transaction(const std::string& logInfo) {
        logTransaction(logInfo);  // Non-virtual call
    }

    void logTransaction(const std::string& logInfo) const {
        // Do logging
    }
};

class BuyTransaction: public Transaction {
public:
    BuyTransaction(parameters)
        : Transaction(createLogString(parameters))  // Pass info to base class
    { }

private:
    static std::string createLogString(parameters) {
        // Return log info
    }
};
```

### Item 10: Have Assignment Operators Return a Reference to *this

**Principle:** Follow the convention for assignment operators to enable chaining.

```cpp
class Widget {
public:
    Widget& operator=(const Widget& rhs) {
        // ...
        return *this;
    }

    // Also applies to +=, -=, *=, etc.
    Widget& operator+=(const Widget& rhs) {
        // ...
        return *this;
    }

    // And even for assignments taking other types
    Widget& operator=(int rhs) {
        // ...
        return *this;
    }
};

// Enables chaining
Widget w1, w2, w3;
w1 = w2 = w3;  // Right-associative: w1 = (w2 = w3)
```

### Item 11: Handle Assignment to Self in operator=

**Principle:** Make sure assignment operators handle self-assignment correctly.

```cpp
// UNSAFE - fails if self-assignment and exception thrown
Widget& Widget::operator=(const Widget& rhs) {
    delete pb;               // Delete current resource
    pb = new Bitmap(*rhs.pb);  // If rhs is *this, we've deleted pb!
    return *this;
}

// BETTER - identity test
Widget& Widget::operator=(const Widget& rhs) {
    if (this == &rhs) return *this;  // Identity test

    delete pb;
    pb = new Bitmap(*rhs.pb);
    return *this;
}

// BEST - copy-and-swap idiom (exception-safe and handles self-assignment)
Widget& Widget::operator=(const Widget& rhs) {
    Widget temp(rhs);      // Make copy
    swap(temp);            // Swap with copy
    return *this;          // temp destroyed, releasing old resources
}

// Or even simpler with pass-by-value
Widget& Widget::operator=(Widget rhs) {  // Pass by value - copy made
    swap(rhs);                            // Swap with copy
    return *this;
}
```

### Item 12: Copy All Parts of an Object

**Principle:** When writing copy constructor or copy assignment, make sure to copy ALL data members and call base class copying functions.

```cpp
class PriorityCustomer: public Customer {
public:
    // BAD - forgot to copy base class parts
    PriorityCustomer(const PriorityCustomer& rhs)
        : priority(rhs.priority)  // Copies derived parts only!
    { }

    // GOOD - copy all parts
    PriorityCustomer(const PriorityCustomer& rhs)
        : Customer(rhs),          // Call base class copy constructor
          priority(rhs.priority)  // Copy derived parts
    { }

    // BAD - forgot to assign base class parts
    PriorityCustomer& operator=(const PriorityCustomer& rhs) {
        priority = rhs.priority;  // Assigns derived parts only!
        return *this;
    }

    // GOOD - assign all parts
    PriorityCustomer& operator=(const PriorityCustomer& rhs) {
        Customer::operator=(rhs);  // Assign base class parts
        priority = rhs.priority;   // Assign derived parts
        return *this;
    }

private:
    int priority;
};
```

**Warning:** Don't try to implement copy constructor in terms of copy assignment or vice versa. If needed, create a private init() helper.

## Chapter 3: Resource Management

### Item 13: Use Objects to Manage Resources (RAII)

**Principle:** Resource Acquisition Is Initialization - Use object lifetime to manage resources.

**The Problem:**
```cpp
void f() {
    Investment* pInv = createInvestment();
    // ...
    delete pInv;  // Might not be reached if exception thrown or early return!
}
```

**The Solution - Smart Pointers:**
```cpp
// Modern C++ (C++11 and later)
void f() {
    std::unique_ptr<Investment> pInv(createInvestment());
    // ... use pInv
    // Automatically deleted when pInv goes out of scope
}

// Or use auto with factory function
void f() {
    auto pInv = createInvestment();  // Returns std::unique_ptr<Investment>
    // ...
}

// For shared ownership
void f() {
    std::shared_ptr<Investment> pInv(createInvestment());
    // ... reference counting manages lifetime
}
```

**Key Principles:**
- Resources are acquired during initialization
- Resources are released in destructors
- Never leak resources if exception thrown
- Prefer std::unique_ptr for exclusive ownership
- Use std::shared_ptr for shared ownership
- Avoid raw new/delete in application code

### Item 14: Think Carefully About Copying Behavior in Resource-Managing Classes

**Options for Copying RAII Objects:**

**1. Prohibit copying** (for exclusive resources):
```cpp
class Lock {
public:
    Lock(const Lock&) = delete;
    Lock& operator=(const Lock&) = delete;
    // ...
};
```

**2. Reference-count** the underlying resource:
```cpp
class Lock {
public:
    explicit Lock(Mutex* pm)
        : mutexPtr(pm, unlock)  // shared_ptr with custom deleter
    {
        lock(mutexPtr.get());
    }

private:
    std::shared_ptr<Mutex> mutexPtr;
};
```

**3. Copy the underlying resource** (deep copy):
```cpp
class String {
public:
    String(const String& rhs)
        : data(new char[strlen(rhs.data) + 1])
    {
        strcpy(data, rhs.data);
    }

private:
    char* data;
};
```

**4. Transfer ownership** (move semantics in C++11):
```cpp
class UniqueResource {
public:
    UniqueResource(UniqueResource&& other) noexcept
        : resource(std::exchange(other.resource, nullptr))
    { }

    UniqueResource& operator=(UniqueResource&& other) noexcept {
        if (this != &other) {
            delete resource;
            resource = std::exchange(other.resource, nullptr);
        }
        return *this;
    }

private:
    Resource* resource;
};
```

### Item 15: Provide Access to Raw Resources in Resource-Managing Classes

**Principle:** Provide ways to get the raw resource when needed (for C APIs or legacy code).

```cpp
class Investment {
public:
    // Explicit conversion
    Investment* get() const { return ptr.get(); }

    // Implicit conversion (use cautiously)
    operator Investment*() const { return ptr.get(); }

private:
    std::unique_ptr<Investment> ptr;
};

// Usage
void processInvestment(Investment* pi);

auto pInv = createInvestment();
processInvestment(pInv.get());  // Explicit - safer
```

### Item 16: Use the Same Form in Corresponding Uses of new and delete

**Principle:** Match array and non-array forms of new and delete.

```cpp
// BAD - mismatched new[]/delete
std::string* stringArray = new std::string[100];
delete stringArray;  // UNDEFINED BEHAVIOR - should be delete[]

// GOOD - matched pairs
std::string* stringPtr1 = new std::string;
std::string* stringArray1 = new std::string[100];

delete stringPtr1;      // delete matches new
delete[] stringArray1;  // delete[] matches new[]

// BEST - avoid manual new/delete
std::vector<std::string> stringArray2(100);  // No manual memory management!
auto stringPtr2 = std::make_unique<std::string>();
```

**Rule:** Never use arrays with manual new[]. Use std::vector or std::array instead.

### Item 17: Store newed Objects in Smart Pointers in Standalone Statements

**Principle:** Create smart pointers in their own statements to avoid resource leaks from exceptions.

```cpp
// BAD - potential resource leak
processWidget(std::shared_ptr<Widget>(new Widget), priority());

// Problem: Order of operations is not guaranteed. Could be:
// 1. new Widget
// 2. priority()  // If this throws exception...
// 3. shared_ptr constructor  // ...Widget leaks!

// GOOD - separate statement
std::shared_ptr<Widget> pw(new Widget);
processWidget(pw, priority());

// BEST - use make functions
processWidget(std::make_shared<Widget>(), priority());
```

## Chapter 4: Designs and Declarations

### Item 18: Make Interfaces Easy to Use Correctly and Hard to Use Incorrectly

**Principle:** Good interfaces are easy to use correctly and hard to use incorrectly.

**Techniques:**

**Use types to prevent errors:**
```cpp
// BAD - easy to swap parameters
Date(int month, int day, int year);
Date d(30, 3, 1995);  // Oops! Meant March 30th

// GOOD - types prevent errors
struct Day { explicit Day(int d) : val(d) {} int val; };
struct Month { explicit Month(int m) : val(m) {} int val; };
struct Year { explicit Year(int y) : val(y) {} int val; };

class Date {
public:
    Date(const Month& m, const Day& d, const Year& y);
};

Date d(Month(3), Day(30), Year(1995));  // Can't mix up order!
```

**Restrict values with types:**
```cpp
class Month {
public:
    static Month Jan() { return Month(1); }
    static Month Feb() { return Month(2); }
    // ...
private:
    explicit Month(int m);
};

Date d(Month::Mar(), Day(30), Year(1995));
```

**Be consistent with built-in types:**
- Follow conventions from standard types
- If a * b is legal, users expect b * a to work too

**Prevent resource leaks:**
```cpp
// BAD - can leak if exception between new and shared_ptr construction
std::shared_ptr<Investment> createInvestment() {
    return std::shared_ptr<Investment>(new Stock);
}

// GOOD - use make_shared
std::shared_ptr<Investment> createInvestment() {
    return std::make_shared<Stock>();
}
```

### Item 19: Treat Class Design as Type Design

**Principle:** Designing a class is defining a new type. Consider these questions:

1. **How should objects be created and destroyed?**
   - Constructors, destructors, memory allocation

2. **How should initialization differ from assignment?**
   - Behavior of constructors vs. assignment operators

3. **What does it mean to pass objects by value?**
   - Copy constructor defines pass-by-value

4. **What are the restrictions on legal values?**
   - Invariants that must be maintained

5. **Does it fit into an inheritance graph?**
   - Virtual or non-virtual destructor?
   - Virtual functions?

6. **What type conversions are allowed?**
   - Implicit vs. explicit conversions
   - Conversion operators

7. **What operators and functions make sense?**
   - Member vs. non-member functions

8. **What standard functions should be disallowed?**
   - Declare private or delete

9. **Who should have access to members?**
   - Public, protected, private
   - Friends

10. **What is the "undeclared interface"?**
    - Performance guarantees
    - Exception safety
    - Resource usage

11. **How general is it?**
    - Should it be a template?

12. **Do you really need a new type?**
    - Can you use existing types?

### Item 20: Prefer Pass-by-Reference-to-const to Pass-by-Value

**Principle:** For user-defined types, pass-by-reference-to-const is more efficient and avoids slicing.

```cpp
// BAD - copies entire object
bool validateStudent(Student s);

// GOOD - no copying, can't modify
bool validateStudent(const Student& s);
```

**Reasons:**

**1. Efficiency - avoids copying:**
```cpp
class Person {
public:
    Person();
    virtual ~Person();
private:
    std::string name;
    std::string address;
};

class Student: public Person {
public:
    Student();
    ~Student();
private:
    std::string schoolName;
    std::string schoolAddress;
};

// BAD - copies 6 strings total!
bool validateStudent(Student s);

// GOOD - no copying
bool validateStudent(const Student& s);
```

**2. Prevents slicing problem:**
```cpp
class Window {
public:
    virtual void display() const;
};

class WindowWithScrollBars: public Window {
public:
    virtual void display() const;
};

// BAD - slices off derived part!
void printNameAndDisplay(Window w) {
    w.display();  // Always calls Window::display!
}

// GOOD - polymorphism works
void printNameAndDisplay(const Window& w) {
    w.display();  // Calls correct version
}
```

**Exception:** Built-in types, STL iterators, and function objects are designed for pass-by-value:
```cpp
void process(int x);           // OK - built-in type
void process(Iterator iter);   // OK - STL iterator
```

### Item 21: Don't Return a Reference When You Must Return an Object

**Principle:** Never return a pointer or reference to a local object, heap-allocated object that must be deleted, or local static when multiple such objects may be needed.

```cpp
// BAD - returns reference to local
const Rational& operator*(const Rational& lhs, const Rational& rhs) {
    Rational result(lhs.n * rhs.n, lhs.d * rhs.d);
    return result;  // DISASTER - returns reference to destroyed object!
}

// BAD - heap allocation, who deletes?
const Rational& operator*(const Rational& lhs, const Rational& rhs) {
    Rational* result = new Rational(lhs.n * rhs.n, lhs.d * rhs.d);
    return *result;  // Who calls delete?
}

// BAD - static, fails with multiple objects
const Rational& operator*(const Rational& lhs, const Rational& rhs) {
    static Rational result;
    result = ...;
    return result;  // Fails: if ((a * b) == (c * d))
}

// GOOD - return by value (compiler can optimize with RVO/NRVO)
const Rational operator*(const Rational& lhs, const Rational& rhs) {
    return Rational(lhs.n * rhs.n, lhs.d * rhs.d);
}
```

### Item 22: Declare Data Members Private

**Principles:**

**1. Syntactic consistency** - clients access everything through functions:
```cpp
class AccessLevels {
public:
    int getReadOnly() const { return readOnly; }
    void setReadWrite(int value) { readWrite = value; }
    int getReadWrite() const { return readWrite; }
    void setWriteOnly(int value) { writeOnly = value; }

private:
    int noAccess;      // No access
    int readOnly;      // Read-only access
    int readWrite;     // Read-write access
    int writeOnly;     // Write-only access
};
```

**2. Fine-grained access control** - read-only, write-only, read-write, no access

**3. Encapsulation** - can change implementation without breaking clients:
```cpp
class SpeedDataCollection {
public:
    void addValue(int speed);
    double averageSoFar() const;

private:
    // Can change implementation later!
    // Option 1: store average
    double average;

    // Option 2: compute on demand
    std::vector<int> speeds;
};
```

**Protected is also not encapsulated:**
- Changing protected members breaks derived classes
- Protected is almost as unencapsulated as public

### Item 23: Prefer Non-member Non-friend Functions to Member Functions

**Principle:** Prefer non-member non-friend functions for better encapsulation.

```cpp
class WebBrowser {
public:
    void clearCache();
    void clearHistory();
    void removeCookies();
};

// LESS encapsulated - member function
class WebBrowser {
public:
    void clearEverything() {  // Member has access to private data
        clearCache();
        clearHistory();
        removeCookies();
    }
};

// MORE encapsulated - non-member function
void clearBrowser(WebBrowser& wb) {  // Can only access public interface
    wb.clearCache();
    wb.clearHistory();
    wb.removeCookies();
}
```

**Why non-member is better:**
- More encapsulation (doesn't increase functions with access to private data)
- Greater packaging flexibility (can be in different headers)
- Increased extensibility (clients can add their own convenience functions)

**Common in C++ standard library:**
```cpp
namespace std {
    template<typename T>
    class vector { ... };

    // Lots of non-member functions operating on vector
    template<typename T>
    void sort(vector<T>& v);
}
```

### Item 24: Declare Non-member Functions When Type Conversions Should Apply to All Parameters

**Principle:** If you need implicit type conversions on all parameters (including `this`), make the function non-member.

```cpp
class Rational {
public:
    Rational(int numerator = 0, int denominator = 1);  // Not explicit - allows implicit conversion

    int numerator() const;
    int denominator() const;

private:
    int n, d;
};

// BAD - member function, asymmetric behavior
class Rational {
public:
    const Rational operator*(const Rational& rhs) const;
};

Rational oneHalf(1, 2);
Rational result = oneHalf * 2;    // OK: oneHalf.operator*(2)
                                  // 2 implicitly converted to Rational(2)
result = 2 * oneHalf;             // ERROR! No way to convert this

// GOOD - non-member function, symmetric behavior
const Rational operator*(const Rational& lhs, const Rational& rhs) {
    return Rational(lhs.numerator() * rhs.numerator(),
                   lhs.denominator() * rhs.denominator());
}

Rational result = oneHalf * 2;    // OK: operator*(oneHalf, Rational(2))
result = 2 * oneHalf;             // OK: operator*(Rational(2), oneHalf)
```

### Item 25: Consider Support for a Non-throwing swap

**Principle:** Provide an efficient, non-throwing swap function for your types.

**Default std::swap:**
```cpp
namespace std {
    template<typename T>
    void swap(T& a, T& b) {
        T temp(a);  // Potentially expensive for some types
        a = b;
        b = temp;
    }
}
```

**For pimpl idiom:**
```cpp
class Widget {
public:
    Widget(const Widget& rhs);
    Widget& operator=(const Widget& rhs) {
        *pImpl = *(rhs.pImpl);  // Copy the Impl object
    }

    void swap(Widget& other) {
        using std::swap;
        swap(pImpl, other.pImpl);  // Just swap pointers - efficient!
    }

private:
    WidgetImpl* pImpl;
};

// Non-member swap that calls member swap
namespace std {
    template<>
    void swap<Widget>(Widget& a, Widget& b) {
        a.swap(b);
    }
}
```

**For templates:**
```cpp
template<typename T>
class WidgetImpl { ... };

template<typename T>
class Widget {
public:
    void swap(Widget& other) {
        using std::swap;
        swap(pImpl, other.pImpl);
    }

private:
    WidgetImpl<T>* pImpl;
};

// Can't partially specialize std::swap, so use own namespace
namespace WidgetStuff {
    template<typename T>
    class Widget { ... };

    template<typename T>
    void swap(Widget<T>& a, Widget<T>& b) {  // Non-member in same namespace
        a.swap(b);
    }
}
```

## Chapter 5: Implementations

### Item 26: Postpone Variable Definitions as Long as Possible

**Principle:** Define variables when you have initialization values, not before.

```cpp
// BAD - wastes construction/destruction if exception thrown
std::string encryptPassword(const std::string& password) {
    std::string encrypted;  // Default constructed here
    if (password.length() < MinimumPasswordLength) {
        throw logic_error("Password is too short");  // encrypted wasted
    }
    // ...
    encrypted = ...;  // Assignment, not initialization
    return encrypted;
}

// BETTER - delay until really needed
std::string encryptPassword(const std::string& password) {
    if (password.length() < MinimumPasswordLength) {
        throw logic_error("Password is too short");
    }
    std::string encrypted;  // Default constructed here
    encrypted = ...;
    return encrypted;
}

// BEST - initialize directly
std::string encryptPassword(const std::string& password) {
    if (password.length() < MinimumPasswordLength) {
        throw logic_error("Password is too short");
    }
    std::string encrypted(computeEncryption(password));  // Initialized!
    return encrypted;
}
```

**In loops:**
```cpp
// Approach A: define outside loop
Widget w;
for (int i = 0; i < n; ++i) {
    w = some value dependent on i;
    ...
}
// Cost: 1 constructor + 1 destructor + n assignments

// Approach B: define inside loop
for (int i = 0; i < n; ++i) {
    Widget w(some value dependent on i);
    ...
}
// Cost: n constructors + n destructors

// Prefer Approach B unless:
// 1. Assignment is less expensive than constructor/destructor pair
// 2. You're dealing with performance-sensitive code
```

### Item 27: Minimize Casting

**Principle:** Avoid casts whenever possible. When necessary, use C++ style casts and hide them in functions.

**C++ cast syntax:**
```cpp
const_cast<T>(expression)        // Cast away const
dynamic_cast<T>(expression)      // Safe downcasting
reinterpret_cast<T>(expression)  // Low-level casts (dangerous!)
static_cast<T>(expression)       // Force implicit conversions
```

**Why C++ style casts are better:**
- Easy to identify in code (searchable)
- More specific about intent
- Compiler can check more strictly

**Common mistake with casts:**
```cpp
class Window {
public:
    virtual void onResize() { ... }
};

class SpecialWindow: public Window {
public:
    virtual void onResize() {
        static_cast<Window>(*this).onResize();  // WRONG! Calls on copy, not *this
        ...  // Do SpecialWindow-specific stuff
    }
};

// CORRECT way:
class SpecialWindow: public Window {
public:
    virtual void onResize() {
        Window::onResize();  // Calls on *this
        ...
    }
};
```

**dynamic_cast is usually slow:**
```cpp
// BAD - cascading dynamic_casts
class Window { ... };
class SpecialWindow1: public Window { ... };
class SpecialWindow2: public Window { ... };

typedef std::vector<std::shared_ptr<Window>> VPW;
VPW winPtrs;

for (VPW::iterator iter = winPtrs.begin(); iter != winPtrs.end(); ++iter) {
    if (SpecialWindow1* psw1 = dynamic_cast<SpecialWindow1*>(iter->get())) {
        ...
    } else if (SpecialWindow2* psw2 = dynamic_cast<SpecialWindow2*>(iter->get())) {
        ...
    }
}

// BETTER - use virtual functions
class Window {
public:
    virtual void blink() { }  // Default implementation does nothing
};

class SpecialWindow1: public Window {
public:
    virtual void blink() { ... }  // Custom blink
};

for (VPW::iterator iter = winPtrs.begin(); iter != winPtrs.end(); ++iter) {
    (*iter)->blink();  // Polymorphism!
}
```

### Item 28: Avoid Returning Handles to Object Internals

**Principle:** Don't return handles (references, pointers, iterators) to internal data. It breaks encapsulation and can lead to dangling handles.

```cpp
class Point {
public:
    Point(int x, int y);
    void setX(int newVal);
    void setY(int newVal);
};

struct RectData {
    Point ulhc;  // upper left-hand corner
    Point lrhc;  // lower right-hand corner
};

class Rectangle {
public:
    // BAD - returns reference to internal data
    Point& upperLeft() { return pData->ulhc; }
    Point& lowerRight() { return pData->lrhc; }

private:
    std::shared_ptr<RectData> pData;
};

// Problem 1: Breaks encapsulation
Point coord1(0, 0);
Point coord2(100, 100);
const Rectangle rec(coord1, coord2);
rec.upperLeft().setX(50);  // rec is supposed to be const!

// BETTER - return const reference
class Rectangle {
public:
    const Point& upperLeft() const { return pData->ulhc; }
    const Point& lowerRight() const { return pData->lrhc; }
};

// Problem 2: Dangling handles
class GUIObject { ... };
const Rectangle boundingBox(const GUIObject& obj);  // Returns temp by value

GUIObject* pgo;
const Point* pUpperLeft = &(boundingBox(*pgo).upperLeft());  // DANGER! Temp destroyed
// pUpperLeft now dangles!

// BEST - return by value
class Rectangle {
public:
    Point upperLeft() const { return pData->ulhc; }
    Point lowerRight() const { return pData->lrhc; }
};
```

### Item 29: Strive for Exception-Safe Code

**Principle:** Exception-safe functions offer one of three guarantees:

1. **Basic guarantee**: If exception thrown, program is in valid state (no leaked resources, no corrupted data)

2. **Strong guarantee**: If exception thrown, program state unchanged (commit-or-rollback semantics)

3. **Nothrow guarantee**: Promise never to throw exceptions (declared `noexcept` in C++11)

```cpp
class PrettyMenu {
public:
    void changeBackground(std::istream& imgSrc);

private:
    Mutex mutex;
    Image* bgImage;
    int imageChanges;
};

// BAD - not exception-safe
void PrettyMenu::changeBackground(std::istream& imgSrc) {
    lock(&mutex);
    delete bgImage;
    ++imageChanges;
    bgImage = new Image(imgSrc);  // If throws, mutex not released, bgImage dangles
    unlock(&mutex);
}

// BETTER - basic guarantee
void PrettyMenu::changeBackground(std::istream& imgSrc) {
    Lock ml(&mutex);  // RAII lock
    delete bgImage;
    ++imageChanges;  // Incremented even if new throws
    bgImage = new Image(imgSrc);
}

// BEST - strong guarantee (copy-and-swap)
class PrettyMenu {
public:
    void changeBackground(std::istream& imgSrc);

private:
    Mutex mutex;
    std::shared_ptr<Image> bgImage;  // Smart pointer
    int imageChanges;
};

struct PMImpl {
    std::shared_ptr<Image> bgImage;
    int imageChanges;
};

void PrettyMenu::changeBackground(std::istream& imgSrc) {
    Lock ml(&mutex);

    std::shared_ptr<PMImpl> pNew(new PMImpl(*pImpl));  // Copy
    pNew->bgImage.reset(new Image(imgSrc));  // Modify copy
    ++pNew->imageChanges;

    std::swap(pImpl, pNew);  // Swap (nothrow)
    // Old data automatically deleted
}
```

### Item 30: Understand the Ins and Outs of Inlining

**Principle:** Limit inlining to small, frequently-called functions.

**Inline is a request, not command:**
```cpp
// Implicitly inline
class Person {
public:
    int age() const { return theAge; }  // Implicitly inline
private:
    int theAge;
};

// Explicitly inline
inline void f() { ... }  // Inline request

template<typename T>
void g(T x) { ... }  // Templates usually in headers, often inlined
```

**Problems with inline:**

1. **Code bloat** - inline replaces each call with function body
2. **Paging issues** - larger code can reduce cache hit rate
3. **Can't debug** - can't set breakpoint in inline function
4. **Binary compatibility** - changing inline function requires recompiling all clients

**When NOT to inline:**
- Functions with loops or recursion
- Virtual functions (usually can't be inlined)
- Functions called through pointers
- Constructors/destructors (often do more than you think)

```cpp
// Looks simple but probably shouldn't be inline
class Base {
public:
    Base() { ... }  // Calls base constructors, initializes members
    ~Base() { ... }  // Calls destructors, exception handling
};
```

### Item 31: Minimize Compilation Dependencies Between Files

**Principle:** Depend on declarations, not definitions. Use pimpl idiom and interface classes.

**Problem - include files increase dependencies:**
```cpp
// Person.h
#include <string>
#include "date.h"
#include "address.h"

class Person {
public:
    Person(const std::string& name, const Date& birthday, const Address& addr);
    std::string name() const;
    // ...

private:
    std::string theName;  // Requires string definition
    Date theBirthDate;    // Requires Date definition
    Address theAddress;   // Requires Address definition
};
```

**Solution 1: Pimpl (Pointer to Implementation) Idiom:**
```cpp
// Person.h
#include <memory>  // For shared_ptr

class PersonImpl;  // Forward declaration
class Date;        // Forward declaration
class Address;     // Forward declaration

class Person {
public:
    Person(const std::string& name, const Date& birthday, const Address& addr);
    std::string name() const;
    // ...

private:
    std::shared_ptr<PersonImpl> pImpl;  // Pointer to implementation
};

// Person.cpp
#include "Person.h"
#include "PersonImpl.h"  // Implementation details

Person::Person(const std::string& name, const Date& birthday, const Address& addr)
    : pImpl(new PersonImpl(name, birthday, addr))
{ }

std::string Person::name() const {
    return pImpl->name();
}
```

**Solution 2: Interface Classes (Abstract Base Classes):**
```cpp
// Person.h
class Person {
public:
    virtual ~Person();
    virtual std::string name() const = 0;
    virtual std::string birthDate() const = 0;

    static std::shared_ptr<Person> create(const std::string& name,
                                         const Date& birthday,
                                         const Address& addr);
};

// RealPerson.cpp
class RealPerson: public Person {
public:
    RealPerson(const std::string& name, const Date& birthday, const Address& addr)
        : theName(name), theBirthDate(birthday), theAddress(addr)
    { }

    virtual ~RealPerson() { }
    std::string name() const { return theName; }
    // ...

private:
    std::string theName;
    Date theBirthDate;
    Address theAddress;
};

std::shared_ptr<Person> Person::create(const std::string& name,
                                      const Date& birthday,
                                      const Address& addr) {
    return std::shared_ptr<Person>(new RealPerson(name, birthday, addr));
}

// Client code
std::shared_ptr<Person> pp(Person::create(name, dateOfBirth, address));
std::cout << pp->name();
```

## Chapter 6: Inheritance and Object-Oriented Design

### Item 32: Make Sure Public Inheritance Models "Is-a"

**Principle:** Public inheritance means "is-a". Everything applicable to base class must apply to derived class.

```cpp
class Bird {
public:
    virtual void fly();  // Birds can fly
};

class Penguin: public Bird {
    // Penguins are birds but can't fly!
    // This inheritance is wrong!
};

// BETTER - more accurate hierarchy
class Bird { ... };
class FlyingBird: public Bird {
public:
    virtual void fly();
};
class Penguin: public Bird {
    // No fly function
};
```

**Square-Rectangle problem:**
```cpp
// Mathematically, square is-a rectangle
class Rectangle {
public:
    virtual void setHeight(int newHeight);
    virtual void setWidth(int newWidth);
    virtual int height() const;
    virtual int width() const;
};

class Square: public Rectangle { ... };

void makeBigger(Rectangle& r) {
    int oldHeight = r.height();
    r.setWidth(r.width() + 10);
    assert(r.height() == oldHeight);  // Should be true for rectangles
                                      // But fails for squares!
}
```

### Item 33: Avoid Hiding Inherited Names

**Problem:** Derived class scope hides base class scope.

```cpp
class Base {
public:
    virtual void mf1() = 0;
    virtual void mf1(int);
    virtual void mf2();
    void mf3();
    void mf3(double);
};

class Derived: public Base {
public:
    virtual void mf1();  // Hides Base::mf1(int)!
    void mf3();          // Hides both Base::mf3()!
    void mf4();
};

Derived d;
d.mf1();     // OK - calls Derived::mf1
d.mf1(10);   // ERROR! Derived::mf1 hides Base::mf1(int)
d.mf2();     // OK - calls Base::mf2
d.mf3();     // OK - calls Derived::mf3
d.mf3(5.5);  // ERROR! Derived::mf3 hides Base::mf3(double)
```

**Solutions:**

**1. Using declarations:**
```cpp
class Derived: public Base {
public:
    using Base::mf1;  // Make all Base::mf1 versions visible
    using Base::mf3;  // Make all Base::mf3 versions visible

    virtual void mf1();
    void mf3();
};

Derived d;
d.mf1();     // OK - calls Derived::mf1
d.mf1(10);   // OK - calls Base::mf1(int)
d.mf3();     // OK - calls Derived::mf3
d.mf3(5.5);  // OK - calls Base::mf3(double)
```

**2. Forwarding functions (for selective unhiding):**
```cpp
class Derived: private Base {  // Private inheritance
public:
    virtual void mf1() {  // Forwarding function
        Base::mf1();      // Inline call to Base::mf1
    }
};

Derived d;
d.mf1();     // OK - Derived::mf1 which calls Base::mf1
d.mf1(10);   // ERROR - Base::mf1(int) not inherited
```

### Item 34: Differentiate Between Inheritance of Interface and Inheritance of Implementation

**Three types of inheritance:**

**1. Pure virtual - inherit interface only:**
```cpp
class Shape {
public:
    virtual void draw() const = 0;  // Must be redefined by derived classes
    virtual void error(const std::string& msg);  // Can be redefined
    int objectID() const;  // Must not be redefined
};
```

**2. Simple virtual - inherit interface + default implementation:**
```cpp
class Airport { ... };

class Airplane {
public:
    virtual void fly(const Airport& destination);  // Default fly behavior
};

class ModelA: public Airplane { ... };  // Uses default fly
class ModelB: public Airplane { ... };  // Uses default fly

// Later, add ModelC that SHOULD have different fly()
class ModelC: public Airplane { ... };  // Oops! Forgot to override fly()
                                       // Uses default - probably wrong!
```

**Better approach - separate interface from default:**
```cpp
class Airplane {
public:
    virtual void fly(const Airport& destination) = 0;  // Pure virtual interface

protected:
    void defaultFly(const Airport& destination);  // Default implementation
};

void Airplane::defaultFly(const Airport& destination) {
    // Default code for flying
}

class ModelA: public Airplane {
public:
    virtual void fly(const Airport& destination) {
        defaultFly(destination);  // Explicitly opt into default
    }
};

class ModelC: public Airplane {
public:
    virtual void fly(const Airport& destination);  // Must provide implementation
};
```

**3. Non-virtual - inherit interface + mandatory implementation:**
```cpp
class Shape {
public:
    int objectID() const { return oid; }  // Never override this!

private:
    int oid;
};
```

### Item 35: Consider Alternatives to Virtual Functions

**1. Non-Virtual Interface (NVI) Idiom:**
```cpp
class GameCharacter {
public:
    int healthValue() const {  // Non-virtual public interface
        // Pre-logic
        int retVal = doHealthValue();  // Call virtual
        // Post-logic
        return retVal;
    }

private:
    virtual int doHealthValue() const {  // Virtual implementation
        ...  // Default calculation
    }
};
```

**2. Function Pointers:**
```cpp
class GameCharacter;
int defaultHealthCalc(const GameCharacter& gc);

class GameCharacter {
public:
    typedef int (*HealthCalcFunc)(const GameCharacter&);

    explicit GameCharacter(HealthCalcFunc hcf = defaultHealthCalc)
        : healthFunc(hcf)
    { }

    int healthValue() const {
        return healthFunc(*this);
    }

private:
    HealthCalcFunc healthFunc;
};
```

**3. std::function (Modern C++):**
```cpp
class GameCharacter {
public:
    typedef std::function<int (const GameCharacter&)> HealthCalcFunc;

    explicit GameCharacter(HealthCalcFunc hcf = defaultHealthCalc)
        : healthFunc(hcf)
    { }

    int healthValue() const {
        return healthFunc(*this);
    }

private:
    HealthCalcFunc healthFunc;
};
```

**4. Strategy Pattern:**
```cpp
class GameCharacter;

class HealthCalcFunc {
public:
    virtual int calc(const GameCharacter& gc) const { ... }
};

HealthCalcFunc defaultHealthCalc;

class GameCharacter {
public:
    explicit GameCharacter(HealthCalcFunc* phcf = &defaultHealthCalc)
        : pHealthCalc(phcf)
    { }

    int healthValue() const {
        return pHealthCalc->calc(*this);
    }

private:
    HealthCalcFunc* pHealthCalc;
};
```

### Item 36-40: Additional Inheritance Guidelines

**Item 36:** Never redefine an inherited non-virtual function
- Non-virtual functions are statically bound
- Redefinition leads to inconsistent behavior

**Item 37:** Never redefine a function's inherited default parameter value
- Default parameters are statically bound
- Virtual functions are dynamically bound
- Leads to confusion

**Item 38:** Model "has-a" or "is-implemented-in-terms-of" through composition
- Has-a: Car has-a Engine
- Use composition, not inheritance

**Item 39:** Use private inheritance judiciously
- Means "is-implemented-in-terms-of"
- Composition is usually better
- Use private inheritance when accessing protected members or redefining virtuals

**Item 40:** Use multiple inheritance judiciously
- Can lead to diamond problem
- Use virtual inheritance to solve it
- Virtual inheritance has costs (size, speed, initialization complexity)
- Prefer single inheritance or composition

## Chapter 7: Templates and Generic Programming

### Item 41: Understand Implicit Interfaces and Compile-Time Polymorphism

**Principle:** Templates support implicit interfaces and compile-time polymorphism, unlike classes which support explicit interfaces and runtime polymorphism.

```cpp
// Explicit interface with runtime polymorphism
class Widget {
public:
    Widget();
    virtual ~Widget();
    virtual std::size_t size() const;
    virtual void normalize();
    void swap(Widget& other);
};

void doProcessing(Widget& w) {
    if (w.size() > 10 && w != someNastyWidget) {
        Widget temp(w);
        temp.normalize();
        temp.swap(w);
    }
}

// Implicit interface with compile-time polymorphism
template<typename T>
void doProcessing(T& w) {
    // Implicit interface:
    // - T must support size() returning something comparable to 10
    // - T must support operator!=
    // - T must support copy constructor
    // - T must support normalize()
    // - T must support swap()

    if (w.size() > 10 && w != someNastyWidget) {
        T temp(w);
        temp.normalize();
        temp.swap(w);
    }
}
```

**Key differences:**
- Classes have **explicit interfaces** based on function signatures
- Templates have **implicit interfaces** based on valid expressions
- Classes have **runtime polymorphism** via virtual functions
- Templates have **compile-time polymorphism** via template instantiation

### Item 42: Understand the Two Meanings of typename

**Principle:** Use `class` and `typename` interchangeably for template parameters, but use `typename` to identify nested dependent type names.

**For template parameters - same meaning:**
```cpp
template<class T> class Widget;     // Uses class
template<typename T> class Widget;  // Uses typename - equivalent
```

**For nested dependent names - must use typename:**
```cpp
template<typename C>
void print2nd(const C& container) {
    if (container.size() >= 2) {
        C::const_iterator iter(container.begin());  // ERROR! Is const_iterator a type?
        // Compiler doesn't know if C::const_iterator is a type or a static member
    }
}

// CORRECT - tell compiler it's a type
template<typename C>
void print2nd(const C& container) {
    if (container.size() >= 2) {
        typename C::const_iterator iter(container.begin());  // OK!
        ++iter;
        int value = *iter;
        std::cout << value;
    }
}
```

**Exception - don't use typename:**
```cpp
template<typename T>
class Derived: public Base<T>::Nested {  // NOT: typename Base<T>::Nested
public:
    explicit Derived(int x)
        : Base<T>::Nested(x)  // NOT: typename Base<T>::Nested
    {
        typename Base<T>::Nested temp;  // But DO use typename here!
        ...
    }
};
```

**Typedef convenience:**
```cpp
template<typename IterT>
void workWithIterator(IterT iter) {
    typedef typename std::iterator_traits<IterT>::value_type value_type;  // Convenient typedef

    value_type temp(*iter);  // Now easier to use
    ...
}
```

### Item 43: Know How to Access Names in Templatized Base Classes

**Problem:** Compilers won't look in templatized base classes for names.

```cpp
class CompanyA {
public:
    void sendCleartext(const std::string& msg);
    void sendEncrypted(const std::string& msg);
};

class CompanyB {
public:
    void sendCleartext(const std::string& msg);
    void sendEncrypted(const std::string& msg);
};

class MsgInfo { ... };

template<typename Company>
class MsgSender {
public:
    void sendClear(const MsgInfo& info) {
        std::string msg;
        // Create msg from info
        Company c;
        c.sendCleartext(msg);
    }
    void sendSecret(const MsgInfo& info) { ... }
};

// Problem: Derived class can't find base class names
template<typename Company>
class LoggingMsgSender: public MsgSender<Company> {
public:
    void sendClearMsg(const MsgInfo& info) {
        // Log before sending
        sendClear(info);  // ERROR! sendClear not found!
        // Log after sending
    }
};
```

**Solutions:**

**1. this-> prefix:**
```cpp
template<typename Company>
class LoggingMsgSender: public MsgSender<Company> {
public:
    void sendClearMsg(const MsgInfo& info) {
        this->sendClear(info);  // OK! Assumes sendClear will be inherited
    }
};
```

**2. using declaration:**
```cpp
template<typename Company>
class LoggingMsgSender: public MsgSender<Company> {
public:
    using MsgSender<Company>::sendClear;  // Tell compiler to assume it's in base

    void sendClearMsg(const MsgInfo& info) {
        sendClear(info);  // OK!
    }
};
```

**3. Explicit qualification:**
```cpp
template<typename Company>
class LoggingMsgSender: public MsgSender<Company> {
public:
    void sendClearMsg(const MsgInfo& info) {
        MsgSender<Company>::sendClear(info);  // OK but disables virtual binding
    }
};
```

### Item 44: Factor Parameter-Independent Code Out of Templates

**Principle:** Avoid template bloat by factoring out non-dependent code.

```cpp
// BAD - n separate functions generated
template<typename T, std::size_t n>
class SquareMatrix {
public:
    void invert();  // Inverts matrix in place
};

SquareMatrix<double, 5> sm1;
sm1.invert();  // Instantiates SquareMatrix<double, 5>::invert

SquareMatrix<double, 10> sm2;
sm2.invert();  // Instantiates SquareMatrix<double, 10>::invert
// Two identical copies of invert() except for constant 5 vs. 10!

// BETTER - factor out size-independent code
template<typename T>
class SquareMatrixBase {
protected:
    SquareMatrixBase(std::size_t n, T* pMem)
        : size(n), pData(pMem) {}

    void setDataPtr(T* ptr) { pData = ptr; }
    void invert(std::size_t matrixSize);  // Size-independent invert

private:
    std::size_t size;
    T* pData;
};

template<typename T, std::size_t n>
class SquareMatrix: private SquareMatrixBase<T> {
public:
    SquareMatrix()
        : SquareMatrixBase<T>(n, data) {}

    void invert() { this->invert(n); }  // Calls base class version

private:
    T data[n*n];
};
```

### Item 45: Use Member Function Templates to Accept "All Compatible Types"

**Principle:** Use member templates for generalized copying.

```cpp
template<typename T>
class SmartPtr {
public:
    explicit SmartPtr(T* realPtr);  // Constructor from raw pointer

    // Member template for "generalized copy constructor"
    template<typename U>
    SmartPtr(const SmartPtr<U>& other)
        : heldPtr(other.get()) { }  // Initialize from compatible SmartPtr

    T* get() const { return heldPtr; }

private:
    T* heldPtr;
};

// Usage
class Top { ... };
class Middle: public Top { ... };
class Bottom: public Middle { ... };

SmartPtr<Top> pt1 = SmartPtr<Middle>(new Middle);  // OK - Middle* converts to Top*
SmartPtr<Top> pt2 = SmartPtr<Bottom>(new Bottom);  // OK - Bottom* converts to Top*
SmartPtr<const Top> pct2 = pt1;                    // OK - non-const to const
```

**Note:** Member templates don't replace compiler-generated functions:
```cpp
template<typename T>
class SmartPtr {
public:
    SmartPtr(const SmartPtr& other);  // Copy constructor (generated by compiler)

    template<typename U>
    SmartPtr(const SmartPtr<U>& other);  // Generalized copy constructor (member template)
};
```

### Item 46: Define Non-member Functions Inside Templates When Type Conversions Are Desired

**Principle:** For template classes, declare non-member functions inside the class template to enable implicit type conversions.

```cpp
template<typename T>
class Rational {
public:
    Rational(const T& numerator = 0, const T& denominator = 1);

    const T numerator() const;
    const T denominator() const;

    // Friend function defined inside class template
    friend const Rational operator*(const Rational& lhs, const Rational& rhs) {
        return Rational(lhs.numerator() * rhs.numerator(),
                       lhs.denominator() * rhs.denominator());
    }
};

// Usage
Rational<int> oneHalf(1, 2);
Rational<int> result = oneHalf * 2;     // OK! 2 converts to Rational<int>
result = 2 * oneHalf;                   // OK! Friend function enables this
```

### Item 47: Use Traits Classes for Information About Types

**Principle:** Traits provide compile-time information about types.

```cpp
// Standard library iterator traits
template<typename IterT>
struct iterator_traits {
    typedef typename IterT::iterator_category iterator_category;
};

// Partial specialization for pointers
template<typename T>
struct iterator_traits<T*> {
    typedef random_access_iterator_tag iterator_category;
};

// Usage - different algorithm for different iterator types
template<typename IterT, typename DistT>
void advance(IterT& iter, DistT d) {
    doAdvance(iter, d, typename std::iterator_traits<IterT>::iterator_category());
}

// Random access version - O(1)
template<typename IterT, typename DistT>
void doAdvance(IterT& iter, DistT d, std::random_access_iterator_tag) {
    iter += d;
}

// Bidirectional version - O(n)
template<typename IterT, typename DistT>
void doAdvance(IterT& iter, DistT d, std::bidirectional_iterator_tag) {
    if (d >= 0) { while (d--) ++iter; }
    else { while (d++) --iter; }
}

// Input iterator version - O(n)
template<typename IterT, typename DistT>
void doAdvance(IterT& iter, DistT d, std::input_iterator_tag) {
    if (d < 0) {
        throw std::out_of_range("Negative distance");
    }
    while (d--) ++iter;
}
```

### Item 48: Be Aware of Template Metaprogramming

**Principle:** Template metaprogramming moves work from runtime to compile time.

```cpp
// Factorial at compile time
template<unsigned n>
struct Factorial {
    enum { value = n * Factorial<n-1>::value };
};

template<>
struct Factorial<0> {
    enum { value = 1 };
};

int main() {
    std::cout << Factorial<5>::value;  // Prints 120, computed at compile time!
    std::cout << Factorial<10>::value; // Prints 3628800, computed at compile time!
}
```

**Benefits:**
- Makes some computations easier to express
- Shifts work from runtime to compile time
- Enables early error detection
- Can lead to smaller executables and faster code

**Drawbacks:**
- Complex syntax
- Long compile times
- Difficult to debug

## Chapter 8: Customizing new and delete

### Item 49: Understand the Behavior of the New-Handler

**Principle:** `set_new_handler` allows you to specify a function to call when memory allocation fails.

```cpp
namespace std {
    typedef void (*new_handler)();
    new_handler set_new_handler(new_handler p) throw();
}

// Custom new-handler
void outOfMem() {
    std::cerr << "Unable to satisfy request for memory\n";
    std::abort();
}

int main() {
    std::set_new_handler(outOfMem);

    int* pBigDataArray = new int[100000000L];  // If fails, calls outOfMem
}
```

**New-handler function should do one of:**
1. **Make more memory available** - e.g., release reserved memory
2. **Install a different new-handler** - if current one can't make more memory available
3. **Deinstall the new-handler** - pass null to `set_new_handler`
4. **Throw an exception** - of type `bad_alloc` or derived
5. **Not return** - typically by calling `abort` or `exit`

**Per-class new-handlers:**
```cpp
class Widget {
public:
    static std::new_handler set_new_handler(std::new_handler p) throw();
    static void* operator new(std::size_t size) throw(std::bad_alloc);

private:
    static std::new_handler currentHandler;
};

std::new_handler Widget::currentHandler = 0;

std::new_handler Widget::set_new_handler(std::new_handler p) throw() {
    std::new_handler oldHandler = currentHandler;
    currentHandler = p;
    return oldHandler;
}

void* Widget::operator new(std::size_t size) throw(std::bad_alloc) {
    NewHandlerHolder h(std::set_new_handler(currentHandler));  // Install Widget's handler
    return ::operator new(size);  // Allocate or throw
    // h's destructor restores global handler
}
```

### Item 50: Understand When It Makes Sense to Replace new and delete

**Reasons to replace:**

**1. Detect usage errors:**
```cpp
static const int signature = 0xDEADBEEF;

typedef unsigned char Byte;

void* operator new(std::size_t size) throw(std::bad_alloc) {
    size_t realSize = size + 2 * sizeof(int);
    void* pMem = malloc(realSize);
    if (!pMem) throw std::bad_alloc();

    // Write signature at front and back
    *(static_cast<int*>(pMem)) = signature;
    *(reinterpret_cast<int*>(static_cast<Byte*>(pMem) + realSize - sizeof(int))) = signature;

    return static_cast<Byte*>(pMem) + sizeof(int);
}

void operator delete(void* pMemory) throw() {
    // Check signatures, report corruption if found
}
```

**2. Improve efficiency:**
- Custom allocators can be faster for specific usage patterns
- Reduce overhead of general-purpose allocators
- Pool allocators for fixed-size objects

**3. Collect usage statistics:**
- Track allocation sizes, lifetimes, patterns
- Identify memory leaks
- Optimize allocation strategy

**4. Increase speed:**
- Object pools
- Arena allocators
- Stack-based allocators

**5. Reduce memory overhead:**
- Custom allocators can have less overhead than default

**6. Accommodate special platforms:**
- Shared memory, memory-mapped files, etc.

### Item 51: Adhere to Convention When Writing new and delete

**Conventions for operator new:**

```cpp
void* operator new(std::size_t size) throw(std::bad_alloc) {
    using namespace std;

    if (size == 0) {
        size = 1;  // Handle 0-byte requests
    }

    while (true) {
        // Attempt allocation
        void* p = malloc(size);

        if (p) {
            return p;  // Success
        }

        // Allocation failed; find current new-handler
        new_handler globalHandler = set_new_handler(0);
        set_new_handler(globalHandler);

        if (globalHandler) {
            (*globalHandler)();  // Call it
        } else {
            throw bad_alloc();  // No handler, throw
        }
    }
}
```

**Class-specific operator new must handle requests for wrong size:**
```cpp
class Base {
public:
    static void* operator new(std::size_t size) throw(std::bad_alloc);
};

class Derived: public Base { ... };

void* Base::operator new(std::size_t size) throw(std::bad_alloc) {
    if (size != sizeof(Base)) {  // Wrong size!
        return ::operator new(size);  // Let standard operator new handle it
    }
    // Handle size == sizeof(Base) case
}
```

**Conventions for operator delete:**
```cpp
void operator delete(void* rawMemory) throw() {
    if (rawMemory == 0) return;  // Do nothing if null pointer

    // Deallocate memory
    free(rawMemory);
}

// Class-specific version
class Base {
public:
    static void* operator new(std::size_t size) throw(std::bad_alloc);
    static void operator delete(void* rawMemory, std::size_t size) throw();
};

void Base::operator delete(void* rawMemory, std::size_t size) throw() {
    if (rawMemory == 0) return;

    if (size != sizeof(Base)) {  // Wrong size - derived object!
        ::operator delete(rawMemory);  // Let standard handle it
        return;
    }

    // Deallocate memory for Base-sized object
    return;
}
```

### Item 52: Write Placement delete If You Write Placement new

**Principle:** If you write placement new, write corresponding placement delete.

```cpp
class Widget {
public:
    // Normal new
    static void* operator new(std::size_t size) throw(std::bad_alloc);

    // Placement new (takes ostream)
    static void* operator new(std::size_t size, std::ostream& logStream) throw(std::bad_alloc);

    // Normal delete
    static void operator delete(void* pMemory) throw();

    // Placement delete (matches placement new)
    static void operator delete(void* pMemory, std::ostream& logStream) throw();
};

// If Widget constructor throws after placement new...
Widget* pw = new (std::cerr) Widget;  // Calls placement new
// If constructor throws, placement delete called automatically
```

**Problem - name hiding:**
```cpp
class Base {
public:
    static void* operator new(std::size_t size, std::ostream& logStream) throw(std::bad_alloc);
};

class Derived: public Base {
public:
    static void* operator new(std::size_t size) throw(std::bad_alloc);
};

Derived* p = new (std::clog) Derived;  // ERROR! Base's placement new is hidden!
```

**Solution:**
```cpp
class StandardNewDeleteForms {
public:
    // Normal new/delete
    static void* operator new(std::size_t size) throw(std::bad_alloc) {
        return ::operator new(size);
    }
    static void operator delete(void* pMemory) throw() {
        ::operator delete(pMemory);
    }

    // Placement new/delete
    static void* operator new(std::size_t size, void* ptr) throw() {
        return ::operator new(size, ptr);
    }
    static void operator delete(void* pMemory, void* ptr) throw() {
        return ::operator delete(pMemory, ptr);
    }

    // Nothrow new/delete
    static void* operator new(std::size_t size, const std::nothrow_t& nt) throw() {
        return ::operator new(size, nt);
    }
    static void operator delete(void* pMemory, const std::nothrow_t&) throw() {
        ::operator delete(pMemory);
    }
};

class Widget: public StandardNewDeleteForms {
public:
    using StandardNewDeleteForms::operator new;    // Make standard forms visible
    using StandardNewDeleteForms::operator delete;

    static void* operator new(std::size_t size, std::ostream& logStream) throw(std::bad_alloc);
    static void operator delete(void* pMemory, std::ostream& logStream) throw();
};
```

## Chapter 9: Miscellany

### Item 53: Pay Attention to Compiler Warnings

**Principle:** Take compiler warnings seriously and strive for warning-free code.

```cpp
class B {
public:
    virtual void f() const;
};

class D: public B {
public:
    virtual void f();  // Forgot const
};

// Compiler warning:
// "D::f() hides virtual B::f()"

// This is BAD! D::f doesn't override B::f, it hides it!
```

**Best practices:**
- Treat warnings as errors (compiler flags like `-Werror`)
- Understand what each warning means
- Fix warnings, don't suppress them without good reason
- Different compilers have different warnings - test with multiple compilers
- Warning-free code at max warning level shows code quality

### Item 54: Familiarize Yourself with the Standard Library

**Standard C++ Library includes:**

**From C:**
- `<cstddef>`, `<cstdlib>`, `<cstring>`, etc.

**iostreams:**
- `<iostream>`, `<fstream>`, `<sstream>`, `<iomanip>`
- Better type safety than C's printf/scanf

**Containers:**
- `<vector>`, `<list>`, `<deque>`, `<set>`, `<map>`, `<stack>`, `<queue>`
- `<array>`, `<forward_list>`, `<unordered_set>`, `<unordered_map>` (C++11)

**Algorithms:**
- `<algorithm>`, `<numeric>`
- `find`, `sort`, `transform`, `accumulate`, etc.

**Iterators:**
- Pointer-like objects for containers
- Input, output, forward, bidirectional, random access

**Function Objects:**
- `<functional>`
- Objects that act like functions

**Smart Pointers (C++11):**
- `<memory>`
- `unique_ptr`, `shared_ptr`, `weak_ptr`

**Strings:**
- `<string>`
- Better than char arrays

**Regular Expressions (C++11):**
- `<regex>`

**Random Numbers (C++11):**
- `<random>`
- Better than C's rand()

**Threading (C++11):**
- `<thread>`, `<mutex>`, `<condition_variable>`, `<atomic>`, `<future>`

**Time (C++11):**
- `<chrono>`

**Tuples (C++11):**
- `<tuple>`

### Item 55: Familiarize Yourself with Boost

**Boost libraries provide:**

**String and text processing:**
- Boost.Regex (regular expressions)
- Boost.Spirit (parser framework)
- Boost.StringAlgo

**Containers:**
- Boost.Array (fixed-size arrays)
- Boost.MultiIndex (multiple indexed views)
- Boost.Bimap (bidirectional maps)

**Function objects and higher-order programming:**
- Boost.Bind (function binding)
- Boost.Lambda (lambda expressions - predates C++11)
- Boost.Function (function wrappers - now in std::function)

**Generic programming:**
- Boost.ConceptCheck (concept checking)
- Boost.TypeTraits (type traits - many now in std)

**Template metaprogramming:**
- Boost.MPL (metaprogramming library)
- Boost.Fusion (compile-time and runtime algorithms)

**Math and numerics:**
- Boost.Math (math functions)
- Boost.Random (random number generators - many now in std)

**Correctness and testing:**
- Boost.Test (unit testing framework)
- Boost.StaticAssert (compile-time assertions)

**Data structures:**
- Boost.Variant (discriminated union)
- Boost.Optional (optional values)
- Boost.Any (typesafe container for single values)

**Inter-language support:**
- Boost.Python (Python/C++ interoperability)

**Memory:**
- Boost.SmartPtr (smart pointers - many now in std)
- Boost.PointerContainer (containers for pointers)

**Miscellaneous:**
- Boost.DateTime (date and time handling)
- Boost.Filesystem (filesystem paths and operations - now in std)
- Boost.Thread (threading - many features now in std)
- Boost.Asio (asynchronous I/O)

**Why use Boost:**
- Peer-reviewed code
- Production-quality
- Cross-platform
- Many Boost libraries become C++ Standard Library components
- Actively maintained and improved

## Modern C++ (C++11/14/17/20)

### Smart Pointers Best Practices

**std::unique_ptr** (Preferred):
- Zero overhead compared to raw pointers
- Exclusive ownership
- Movable but not copyable
- Can have custom deleters
```cpp
auto pw = std::make_unique<Widget>();
std::unique_ptr<Widget, decltype(&deleter)> pw(new Widget, deleter);
```

**std::shared_ptr**:
- Reference counted
- Thread-safe reference counting
- Slightly heavier than unique_ptr
- Can have custom deleters
```cpp
auto pw = std::make_shared<Widget>();  // More efficient than new
```

**std::weak_ptr**:
- Breaks cycles in shared_ptr graphs
- Doesn't affect reference count
- Must check if resource still exists before use
```cpp
std::weak_ptr<Widget> wpw = spw;
if (auto spt = wpw.lock()) {  // Check if still alive
    // Use spt
}
```

### Rule of Zero/Three/Five

**Rule of Zero** (Preferred):
- Don't define any special member functions
- Let compiler generate them
- Use std::unique_ptr, std::vector, etc.
```cpp
class Widget {
    std::string name;
    std::vector<int> data;
    std::unique_ptr<Impl> pImpl;
    // Compiler-generated special members work perfectly!
};
```

**Rule of Three** (Pre-C++11):
- If you define one of: destructor, copy constructor, copy assignment
- You probably need to define all three

**Rule of Five** (C++11+):
- If you define any special member function, define or delete all five:
  - Destructor
  - Copy constructor
  - Copy assignment operator
  - Move constructor
  - Move assignment operator
```cpp
class Resource {
public:
    ~Resource();                              // Destructor
    Resource(const Resource&);                // Copy constructor
    Resource& operator=(const Resource&);     // Copy assignment
    Resource(Resource&&) noexcept;            // Move constructor
    Resource& operator=(Resource&&) noexcept; // Move assignment
};
```

### Move Semantics

**When to provide move operations:**
- When copying is expensive
- When managing resources
- For containers and large objects

```cpp
class Buffer {
public:
    // Move constructor
    Buffer(Buffer&& other) noexcept
        : data(std::exchange(other.data, nullptr)),
          size(std::exchange(other.size, 0))
    { }

    // Move assignment
    Buffer& operator=(Buffer&& other) noexcept {
        if (this != &other) {
            delete[] data;
            data = std::exchange(other.data, nullptr);
            size = std::exchange(other.size, 0);
        }
        return *this;
    }

private:
    char* data;
    size_t size;
};
```

**noexcept** is important for move operations:
- Enables optimizations (like std::vector growth)
- Should be noexcept whenever possible

## References and Sources

**Books:**
- [Effective C++: 55 Specific Ways to Improve Your Programs and Designs](https://www.amazon.com/Effective-Specific-Improve-Programs-Designs/dp/0321334876) by Scott Meyers
- [Effective Modern C++: 42 Specific Ways to Improve Your Use of C++11 and C++14](https://www.amazon.com/Effective-Modern-Specific-Ways-Improve/dp/1491903996) by Scott Meyers
- [More Effective C++: 35 New Ways to Improve Your Programs and Designs](https://www.amazon.com/More-Effective-Improve-Programs-Designs/dp/020163371X) by Scott Meyers

**Online Resources:**
- [C++ Core Guidelines](https://isocpp.github.io/CppCoreGuidelines/CppCoreGuidelines) - ISO C++ community standards
- [Effective C++ Items Summary](https://gist.github.com/asambol/fa234c747ba4a677dee7b2ddaa64778d) - Community reference
- [C++ Core Guidelines: Resource Management](https://www.modernescpp.com/index.php/c-core-guidelines-rules-to-resource-management/)
- [C++ Core Guidelines: Smart Pointers](https://www.modernescpp.com/index.php/c-core-guidelines-rules-to-smart-pointers/)

## C++ Core Guidelines

The [C++ Core Guidelines](https://isocpp.github.io/CppCoreGuidelines/CppCoreGuidelines) are a set of tried-and-true guidelines, rules, and best practices about coding in C++ maintained by the ISO C++ Standards Committee.

### Philosophy (P): General Guiding Principles

**P.1: Express ideas directly in code**
- Use language facilities to express intent
- Don't hide information needed to understand behavior
- Make types match abstractions

**P.2: Write in ISO Standard C++**
- Use standard C++ features
- Avoid compiler extensions when possible
- Enables portability

**P.3: Express intent**
- Code should clearly state what it intends to do
- Good names, clear structure, meaningful comments
- Easier to understand and maintain

**P.4: Ideally, a program should be statically type safe**
- Type system is your first line of defense
- Catch errors at compile time
- Use strong types over weak types

**P.5: Prefer compile-time checking to run-time checking**
- Compile-time checking is faster and more reliable
- No runtime overhead
- Impossible to violate at runtime

**P.6: What cannot be checked at compile time should be checkable at run time**
- Use assertions, contracts, exceptions
- Fail fast when invariants violated
- Enable debugging and testing

**P.7: Catch run-time errors early**
- Check preconditions, postconditions, invariants
- Don't let errors propagate
- Fail fast and loudly

**P.8: Don't leak any resources**
- Use RAII for all resources
- Smart pointers, containers manage memory
- No bare new/delete in application code

**P.9: Don't waste time or space**
- Efficiency matters but measure first
- No premature optimization
- Consider algorithmic complexity

**P.10: Prefer immutable data to mutable data**
- Immutable data is easier to reason about
- Thread-safe by default
- Use const whenever possible

**P.11: Encapsulate messy constructs**
- Don't let implementation details leak
- Provide clean abstractions
- Isolate low-level code

**P.12: Use supporting tools as appropriate**
- Static analyzers, sanitizers, profilers
- Automated testing frameworks
- Build systems and package managers

**P.13: Use support libraries as appropriate**
- Don't reinvent the wheel
- Standard library, Boost, etc.
- Peer-reviewed, tested code

### Functions (F): Design and Declaration

**F.1: "Package" meaningful operations as carefully named functions**
- Functions are basic building blocks
- Each function should do one logical thing
- Good names document intent

**F.2: A function should perform a single logical operation**
- Single Responsibility Principle
- Easier to test, understand, reuse
- Compose functions to build complexity

**F.3: Keep functions short and simple**
- Prefer functions that fit on a screen
- One level of abstraction per function
- Extract complex logic to helper functions

**F.7: For general use, take T* or T& arguments rather than smart pointers**
- Smart pointers express ownership
- Non-owning parameters should use raw pointers/references
- Clearer intent, less coupling

**F.8: Prefer pure functions**
- No side effects
- Same inputs always produce same outputs
- Easier to reason about, test, parallelize

**F.15: Prefer simple and conventional ways of passing information**
```cpp
void f(int x);              // Pass by value for cheap-to-copy types
void f(const string& s);    // Pass by const reference for expensive types
void f(string& s);          // Pass by reference to modify
void f(string&& s);         // Pass by rvalue reference to move from
```

### Interfaces (I): Making Interfaces Explicit

**I.1: Make interfaces explicit**
- State assumptions in interface
- Use preconditions, postconditions
- Document requirements

**I.2: Avoid non-const global variables**
- Breaks encapsulation
- Hard to reason about
- Thread-safety issues

**I.4: Make interfaces precisely and strongly typed**
- Type system prevents errors
- Better than comments or documentation
```cpp
// BAD
void draw_rect(int, int, int, int);  // What do these ints mean?

// GOOD
void draw_rect(Point top_left, Point bottom_right);
void draw_rect(Point top_left, Size size);
```

**I.10: Use exceptions to signal a failure to perform a required task**
- Don't use error codes
- Exceptions can't be ignored
- Separate error handling from normal logic

**I.11: Never transfer ownership by a raw pointer (T\*) or reference (T&)**
- Use unique_ptr for exclusive ownership transfer
- Use shared_ptr for shared ownership
- Raw pointers don't express ownership

**I.13: Do not pass an array as a single pointer**
- Use span<T> or vector<T>
- Prevents buffer overruns
```cpp
// BAD
void read(int* p, int n);  // Size separate from pointer

// GOOD
void read(span<int> s);    // Size bundled with pointer
```

### Classes (C): Class Design and Hierarchies

**C.1: Organize related data into structures (structs or classes)**
- If data is related, reflect that in code
- Improves comprehension
- Enables better abstractions

**C.2: Use class if the class has an invariant; use struct if the data members can vary independently**
- Classes maintain invariants
- Structs are just data bundles
- Signals intent to readers

**C.3: Represent the distinction between an interface and an implementation using a class**
- Public interface, private implementation
- Encapsulation
- Can change implementation without affecting clients

**C.4: Make a function a member only if it needs direct access to the representation of a class**
- Prefer free functions for operations not needing access
- Better encapsulation
- Easier testing

**C.9: Minimize exposure of members**
- Prefer private data members
- Public interface, private implementation
- Enables change without breaking clients

**C.21: If you define or =delete any default operation, define or =delete them all**
- Rule of Five/Zero
- Prevents subtle bugs
- Be explicit about intent

**C.30: Define a destructor if a class needs an explicit action at object destruction**
- RAII pattern
- Resource cleanup
- But prefer Rule of Zero

**C.31: All resources acquired by a class must be released by the class's destructor**
- No leaks
- Exception safety
- Use smart pointers/RAII wrappers

**C.35: A base class destructor should be either public and virtual, or protected and non-virtual**
- Public virtual: for polymorphic use
- Protected non-virtual: prevent deletion through base pointer
- Prevents undefined behavior

**C.45: Don't define a default constructor that only initializes data members; use in-class member initializers instead**
```cpp
// BAD
class X {
    int i;
    string s;
public:
    X() : i(0), s("") {}
};

// GOOD
class X {
    int i = 0;
    string s = "";
public:
    X() = default;
};
```

**C.46: By default, declare single-argument constructors explicit**
- Prevents implicit conversions
- Avoids surprises
```cpp
class String {
public:
    explicit String(int size);  // Not a conversion from int
};
```

### Resource Management (R): RAII and Lifetimes

**R.1: Manage resources automatically using resource handles and RAII**
- Constructor acquires resource
- Destructor releases resource
- No manual new/delete needed
- Exception-safe by default

**R.2: In interfaces, use raw pointers to denote individual objects (only)**
- Single object, not arrays
- Non-owning
- Use span<T> for arrays

**R.3: A raw pointer (a T\*) is non-owning**
- Ownership via smart pointers
- Raw pointers for observation only
- Clear intent

**R.4: A raw reference (a T&) is non-owning**
- References never own
- Can't be null, can't be reseated
- Safe alternative to pointers

**R.5: Prefer scoped objects, don't heap-allocate unnecessarily**
- Stack allocation is faster
- Automatic lifetime management
- Use heap only when necessary
```cpp
// BAD
auto p = new Widget();  // Manual management

// GOOD
Widget w;  // Automatic management
```

**R.10: Avoid malloc() and free()**
- No type safety
- No constructor/destructor calls
- Use new/delete or better, smart pointers

**R.11: Avoid calling new and delete explicitly**
- Use make_unique/make_shared
- Use containers (vector, string)
- RAII wrappers

**R.12: Immediately give the result of an explicit resource allocation to a manager object**
```cpp
// BAD
Widget* p = new Widget;
// ... could throw before smart pointer created

// GOOD
auto p = make_unique<Widget>();  // Atomic, exception-safe
```

**R.20: Use unique_ptr or shared_ptr to represent ownership**
- unique_ptr for exclusive ownership
- shared_ptr for shared ownership
- Automatic cleanup

**R.21: Prefer unique_ptr over shared_ptr unless you need to share ownership**
- unique_ptr is lighter (no ref counting)
- Makes ownership clear
- Use shared_ptr only when actually sharing

### Expressions and Statements (ES): Code Clarity

**ES.1: Prefer the standard library to other libraries and to "handcrafted code"**
- Well-tested and optimized
- Portable and maintained
- Better abstractions than raw loops

**ES.5: Keep scopes small**
- Declare variables close to use
- Reduces cognitive load
- Limits variable lifetime

**ES.6: Declare names in for-statement initializers and conditions to limit scope**
```cpp
for (int i = 0; i < max; ++i) {  // i scoped to loop
    // ...
}
// i not accessible here
```

**ES.10: Declare one name (only) per declaration**
```cpp
// BAD
int* p, q;  // q is int, not int*!

// GOOD
int* p;
int* q;
```

**ES.11: Use auto to avoid redundant repetition of type names**
```cpp
auto v = make_unique<Widget>();  // Type is obvious from context
```

**ES.20: Always initialize an object**
- Uninitialized variables cause bugs
- Modern C++ makes initialization easy
```cpp
int x = 0;        // OK
int y{0};         // Also OK
auto z = 0;       // Also OK
int w;            // BAD - uninitialized
```

**ES.21: Don't introduce a variable (or constant) before you need to use it**
- Declare close to first use
- Easier to understand
- Smaller scope

**ES.22: Don't declare a variable until you have a value to initialize it with**
```cpp
// BAD
int x;
// ... lots of code
x = calculate();

// GOOD
int x = calculate();
```

**ES.23: Prefer the {}-initializer syntax**
- Most vexing parse prevention
- Prevents narrowing conversions
- Consistent syntax
```cpp
int x{5};           // OK
int y = {5};        // OK
int z(5);           // OK but confusing
auto v = vector<int>{1, 2, 3};  // Clear initialization
```

**ES.28: Use lambdas for complex initialization, especially of const variables**
```cpp
const int x = [&] {
    // Complex calculation
    return result;
}();
```

**ES.46: Avoid lossy (narrowing, truncating) arithmetic conversions**
```cpp
int i = 7.2;   // BAD - truncates
int j{7.2};    // ERROR - {} prevents narrowing
```

**ES.49: If you must use a cast, use a named cast**
- static_cast, const_cast, reinterpret_cast, dynamic_cast
- Never use C-style casts
- Named casts are searchable and explicit

**ES.50: Don't cast away const**
- const exists for a reason
- Casting it away is usually a design error
- If necessary, document why

**ES.56: Write std::move() only when you need to explicitly move an object to another scope**
- Don't move unnecessarily
- Compiler moves automatically when appropriate
- Moving can leave object in valid-but-unspecified state

**ES.60: Avoid new and delete outside resource management functions**
- Use make_unique/make_shared
- Use containers
- RAII wrappers

**ES.61: Delete arrays using delete[] and non-arrays using delete**
- Mismatched delete is undefined behavior
- Better: use vector or array instead

**ES.65: Don't dereference an invalid pointer**
- Always check pointers before use
- Use optional<T> or smart pointers
- Null pointer errors are common bugs

### Error Handling (E): Exception Safety and RAII

The E section focuses on systematic error handling using exceptions and RAII.

**E.1: Develop an error-handling strategy early in a design**
- Error handling is hard to retrofit
- Strategy should be systematic, robust, and non-repetitive
- Decide early: exceptions vs error codes vs terminate
- Document error handling approach in design

**E.2: Throw an exception to signal that a function can't perform its assigned task**
- Exceptions are for exceptional circumstances
- Can't be ignored (unlike return codes)
- Non-intrusive signaling of failure
- Constructor failure requires exception
```cpp
// GOOD - use exceptions
class File {
public:
    File(const string& name) {
        f = fopen(name.c_str(), "r");
        if (!f) {
            throw runtime_error("Could not open " + name);
        }
    }
private:
    FILE* f;
};

// BAD - can't signal constructor failure without exceptions
class File {
public:
    File(const string& name) {
        f = fopen(name.c_str(), "r");
        // How to signal failure? Can't return error code!
    }
    bool is_valid() const { return f != nullptr; }  // Client must remember to check
};
```

**E.3: Use exceptions for error handling only**
- Don't use exceptions for normal control flow
- Exceptions are for errors that can't be handled locally
- Performance cost of throwing is acceptable for rare events
```cpp
// BAD - using exceptions for control flow
try {
    while (true) {
        // ... do something
        if (done()) throw Done();
    }
} catch (Done&) {
    // normal exit
}

// GOOD - use normal control flow
while (!done()) {
    // ... do something
}
```

**E.4: Design your error-handling strategy around invariants**
- Establish what invariants your code maintains
- Use exceptions when invariants are violated
- RAII maintains invariants automatically

**E.5: Let a constructor establish an invariant, and throw if it cannot**
- Constructor should fully initialize object or throw
- No partially-constructed objects
- After constructor succeeds, object is valid
```cpp
class Vector {
public:
    Vector(int size) : elem(new double[size]), sz(size) {
        if (size < 0) throw length_error("negative size");
        // Object now valid, invariant established
    }
private:
    double* elem;
    int sz;
};
```

**E.6: Use RAII to prevent leaks**
- RAII is the simplest, most systematic way to prevent leaks
- Resource lifetime tied to object lifetime
- C++ runtime manages object lifetimes
- Automatic cleanup on scope exit or exception
```cpp
// BAD - manual resource management, can leak
void f(const string& name) {
    FILE* f = fopen(name.c_str(), "r");
    // ... use f ...
    fclose(f);  // Might not be reached if exception thrown
}

// GOOD - RAII, no leaks
void f(const string& name) {
    ifstream f(name);  // Constructor acquires, destructor releases
    // ... use f ...
}  // Automatically closed, even if exception thrown
```

**E.7: State your preconditions**
- Document what must be true before function is called
- Use assertions to check preconditions
- Consider contracts (C++20)
```cpp
int area(int height, int width) {
    // Precondition: height and width are positive
    assert(height > 0 && width > 0);
    return height * width;
}
```

**E.8: State your postconditions**
- Document what will be true after function returns
- Use assertions or return value checks

**E.12: Use noexcept when exiting a function because of a throw is impossible or unacceptable**
- `noexcept` enables optimizations
- Reduces alternative execution paths
- Speeds up failure handling
- Terminated immediately if violated
```cpp
// Functions that should never throw
void swap(T& a, T& b) noexcept {
    T tmp = move(a);
    a = move(b);
    b = move(tmp);
}

// Move operations should be noexcept when possible
class Vector {
public:
    Vector(Vector&& v) noexcept
        : elem(v.elem), sz(v.sz)
    {
        v.elem = nullptr;
        v.sz = 0;
    }
};
```

**E.13: Never throw while being the direct owner of an object**
- Direct ownership means using raw pointers with new/delete
- Exception before delete causes leak
- Use RAII instead
```cpp
// BAD - may leak
void leak(int x) {
    auto p = new int{7};
    if (x < 0) throw Get_me_out_of_here{};  // Leak!
    delete p;
}

// GOOD - RAII prevents leak
void no_leak(int x) {
    auto p = make_unique<int>(7);
    if (x < 0) throw Get_me_out_of_here{};  // No leak
}
```

**E.14: Use purpose-designed user-defined types as exceptions (not built-in types)**
```cpp
// BAD - hard to catch selectively
void f() {
    throw 7;  // int exception
}

// GOOD - specific exception type
class MyError : public runtime_error {
public:
    MyError(const string& msg) : runtime_error(msg) {}
};

void f() {
    throw MyError("something went wrong");
}
```

**E.15: Throw by value, catch by reference**
```cpp
// GOOD
try {
    throw MyException("error");
} catch (const MyException& e) {  // Catch by const reference
    // Handle exception
}

// BAD - slicing
catch (MyException e) {  // Catch by value - slices derived classes
}

// BAD - can throw null pointer
catch (MyException* e) {  // Catch by pointer
}
```

**E.16: Destructors, deallocation, and swap must never fail**
- These operations are fundamental to RAII and exception safety
- Mark them `noexcept`
- If they might fail, design differently

**E.25: If you can't throw exceptions, simulate RAII for resource management**
- In no-exception environments (embedded, real-time, legacy)
- Systematically check object validity after construction
- Still release all resources in destructor
```cpp
// No-exception RAII
class File {
public:
    File(const string& name) : f(fopen(name.c_str(), "r")) {}

    bool valid() const { return f != nullptr; }  // Check validity

    ~File() { if (f) fclose(f); }  // Always release

private:
    FILE* f;
};
```

**E.27: If you can't throw exceptions, use error codes systematically**
- Return error codes or use output parameters
- Check every error code
- Document which functions can fail
- Consider `std::expected` or `std::optional` (C++23/17)

**E.28: Avoid error handling based on global state (errno)**
```cpp
// BAD - errno is global, can be overwritten
double d = sqrt(-1);
if (errno == EDOM) { /* ... */ }  // Unreliable

// GOOD - exception or return value
auto result = safe_sqrt(-1);  // Returns optional<double> or throws
```

### Concurrency and Parallelism (CP): Thread Safety

The CP section addresses safe concurrent programming with threads.

**CP.1: Assume that your code will run as part of a multi-threaded program**
- Even if single-threaded now, may be multi-threaded later
- Design for thread safety from the start
- Use const, immutable data, and proper synchronization

**CP.2: Avoid data races**
- Data race = concurrent access to same memory, at least one is write, no synchronization
- Data races are undefined behavior
- Use synchronization (mutex, atomic) or immutable data
```cpp
// BAD - data race
int counter = 0;
void increment() {
    ++counter;  // Data race if called from multiple threads!
}

// GOOD - mutex synchronization
mutex mtx;
int counter = 0;
void increment() {
    lock_guard<mutex> lock(mtx);
    ++counter;
}

// BETTER - atomic
atomic<int> counter{0};
void increment() {
    ++counter;  // Atomic operation, no race
}
```

**CP.3: Minimize explicit sharing of writable data**
- Shared mutable state is source of concurrency bugs
- Pass data by value to threads when possible
- Use message passing instead of shared memory
- Prefer immutable data structures
```cpp
// GOOD - pass by value, no sharing
void worker(vector<int> data) {  // Copy passed to thread
    // Process data independently
}

thread t(worker, myData);  // myData copied to thread
```

**CP.4: Think in terms of tasks, rather than threads**
- Tasks are higher-level abstraction
- Use `std::async`, thread pools, or task libraries
- Let runtime manage thread creation
```cpp
// BETTER - task-based
auto result = async(launch::async, []{ return compute(); });
// ... do other work ...
auto value = result.get();

// WORSE - manual thread management
int result;
thread t([&result]{ result = compute(); });
// ... do other work ...
t.join();
```

**CP.8: Don't try to use volatile for synchronization**
- `volatile` is for hardware memory, not thread synchronization
- Use `atomic` or mutexes for thread synchronization
```cpp
// BAD - volatile doesn't provide thread safety
volatile bool ready = false;
// ... other thread sets ready = true ...
while (!ready) { }  // Not guaranteed to work!

// GOOD - use atomic
atomic<bool> ready{false};
// ... other thread sets ready = true ...
while (!ready) { }  // Guaranteed to see update
```

**CP.20: Use RAII, never plain lock()/unlock()**
- Always wrap mutex in RAII lock guard
- Prevents forgetting to unlock
- Exception-safe
```cpp
// BAD - manual lock/unlock
mutex mtx;
void f() {
    mtx.lock();
    // ... code that might throw ...
    mtx.unlock();  // Might not be reached!
}

// GOOD - RAII lock guard
mutex mtx;
void f() {
    lock_guard<mutex> lock(mtx);  // Automatically unlocks
    // ... code that might throw ...
}  // Automatically unlocked here
```

**CP.21: Use std::lock() or std::scoped_lock to acquire multiple mutexes**
- Lock multiple mutexes atomically to avoid deadlock
- Always lock in same order, or use `std::lock`/`std::scoped_lock`
```cpp
// BAD - potential deadlock
mutex m1, m2;
void f() {
    lock_guard<mutex> lock1(m1);
    lock_guard<mutex> lock2(m2);  // Different order in different threads = deadlock!
}

// GOOD - C++11: use std::lock
void f() {
    unique_lock<mutex> lock1(m1, defer_lock);
    unique_lock<mutex> lock2(m2, defer_lock);
    lock(lock1, lock2);  // Atomic, deadlock-free
}

// BETTER - C++17: use scoped_lock
void f() {
    scoped_lock lock(m1, m2);  // Locks both atomically
}
```

**CP.22: Never call unknown code while holding a lock (e.g., a callback)**
- Unknown code might try to acquire same lock (deadlock)
- Unknown code might be slow (poor performance)
- Release lock before calling callbacks
```cpp
// BAD - calling callback while holding lock
mutex mtx;
function<void()> callback;

void notify() {
    lock_guard<mutex> lock(mtx);
    callback();  // Callback might deadlock or be slow!
}

// GOOD - release lock before callback
void notify() {
    function<void()> cb;
    {
        lock_guard<mutex> lock(mtx);
        cb = callback;
    }
    cb();  // Call outside lock
}
```

**CP.23: Think of a joining thread as a scoped container**
- Join threads before they go out of scope
- Or explicitly detach
- Use RAII wrapper for threads
```cpp
// GOOD - RAII thread wrapper
class joining_thread {
public:
    template<typename... Args>
    joining_thread(Args&&... args)
        : t(forward<Args>(args)...) {}

    ~joining_thread() {
        if (t.joinable()) t.join();
    }

private:
    thread t;
};
```

**CP.24: Think of a thread as a global container**
- Detached threads live until program ends
- Be very careful with detached threads
- Prefer joinable threads

**CP.25: Prefer gsl::joining_thread over std::thread**
- Automatically joins on destruction
- No risk of termination from unjoi ned thread

**CP.26: Don't detach() a thread**
- Detached threads are hard to reason about
- No way to know when they finish
- No way to get return value or exceptions
- Prefer joinable threads

**CP.31: Pass small amounts of data between threads by value, rather than by reference or pointer**
- Avoids sharing and synchronization
- Simpler and safer

**CP.32: To share ownership between unrelated threads use shared_ptr**
- When thread lifetime is independent
- `shared_ptr` manages lifetime safely
```cpp
void worker(shared_ptr<Data> data) {
    // Use data - lifetime extended automatically
}

auto data = make_shared<Data>();
thread t1(worker, data);
thread t2(worker, data);  // Both threads share ownership
```

**CP.41: Minimize thread creation and destruction**
- Thread creation is expensive
- Use thread pools
- Reuse threads for multiple tasks

**CP.42: Don't wait without a condition**
- Busy-waiting wastes CPU
- Use condition variables or futures
```cpp
// BAD - busy-wait
while (!ready) { }  // Wastes CPU

// GOOD - condition variable
unique_lock<mutex> lock(mtx);
cv.wait(lock, []{ return ready; });  // Efficient wait
```

**CP.43: Minimize time spent in a critical section**
- Shorter critical sections = better parallelism
- Do work outside lock when possible
```cpp
// BAD - long critical section
lock_guard<mutex> lock(mtx);
auto result = expensive_computation(data);  // Don't compute under lock!
shared_data = result;

// GOOD - short critical section
auto result = expensive_computation(data);  // Compute outside lock
lock_guard<mutex> lock(mtx);
shared_data = result;  // Only assignment under lock
```

**CP.44: Remember to name your lock_guards and unique_locks**
```cpp
// BAD - temporary lock, unlocked immediately!
lock_guard<mutex>(mtx);  // Creates and destroys lock immediately
// Critical section not protected!

// GOOD - named lock
lock_guard<mutex> lock(mtx);  // Lives until end of scope
// Critical section properly protected
```

**CP.50: Define a mutex together with the data it protects**
```cpp
// GOOD - mutex with data it protects
class BankAccount {
public:
    void deposit(int amount) {
        lock_guard<mutex> lock(mtx);
        balance += amount;
    }

private:
    mutex mtx;       // Protects balance
    int balance{0};  // Protected data
};
```

**CP.60: Use a future to return a value from a concurrent task**
```cpp
// GOOD - use future for return value
future<int> result = async(launch::async, []{
    return expensive_computation();
});
// ... do other work ...
int value = result.get();  // Get result when ready
```

**CP.61: Use async() to spawn concurrent tasks**
- Simpler than manual thread management
- Returns future automatically
- Runtime chooses execution strategy

**CP.110: Do not write your own double-checked locking for initialization**
- Extremely error-prone
- Use `std::call_once` instead
```cpp
// BAD - broken double-checked locking
Widget* instance = nullptr;
mutex mtx;

Widget* get_instance() {
    if (!instance) {  // First check (unsafe)
        lock_guard<mutex> lock(mtx);
        if (!instance) {  // Second check
            instance = new Widget();  // Can be reordered!
        }
    }
    return instance;
}

// GOOD - use call_once
Widget* instance = nullptr;
once_flag flag;

Widget* get_instance() {
    call_once(flag, []{ instance = new Widget(); });
    return instance;
}

// BETTER - use local static (C++11 thread-safe)
Widget& get_instance() {
    static Widget instance;  // Thread-safe initialization
    return instance;
}
```

**CP.111: Use a conventional pattern if you really need double-checked locking**
- Only if proven necessary by profiling
- Use atomic with proper memory ordering
- Better: avoid need for it

---

## Enum: Enumerations

Enumerations provide type-safe named constants. Modern C++ prefers `enum class` over traditional `enum`.

**Enum.1: Prefer enumerations over macros**
- Enumerations are type-safe and scoped
- Macros pollute namespace and lack type safety
```cpp
// BAD - macros
#define RED 0
#define GREEN 1
#define BLUE 2

// GOOD - enumeration
enum class Color { red, green, blue };
```

**Enum.2: Use enumerations to represent sets of related named constants**
```cpp
enum class TrafficLight { red, yellow, green };
enum class Month { jan = 1, feb, mar, apr, may, jun, jul, aug, sep, oct, nov, dec };
```

**Enum.3: Prefer enum classes over "plain" enums**
- `enum class` prevents implicit conversions and name pollution
- Safer and more explicit
```cpp
// BAD - plain enum
enum Color { red, green, blue };
int x = red;  // Implicit conversion to int (dangerous!)

// GOOD - enum class
enum class Color { red, green, blue };
Color c = Color::red;
// int x = Color::red;  // Error: no implicit conversion (safe!)
int x = static_cast<int>(Color::red);  // Explicit conversion required
```

**Enum.4: Define operations on enumerations for safe and simple use**
```cpp
enum class Day { mon, tue, wed, thu, fri, sat, sun };

// Define increment operator
Day& operator++(Day& d) {
    return d = (d == Day::sun) ? Day::mon : static_cast<Day>(static_cast<int>(d) + 1);
}

// Usage
Day today = Day::mon;
++today;  // Now tue
```

**Enum.5: Don't use ALL_CAPS for enumerators**
- Conflicts with macro conventions
- Use lowercase or CamelCase
```cpp
// BAD
enum class Color { RED, GREEN, BLUE };

// GOOD
enum class Color { red, green, blue };
// or
enum class Color { Red, Green, Blue };
```

**Enum.6: Avoid unnamed enumerations**
```cpp
// BAD - unnamed enum
enum { red, green, blue };  // What does this represent?

// GOOD - named enum
enum class Color { red, green, blue };
```

**Enum.7: Specify the underlying type of an enumeration only when necessary**
```cpp
// Usually not needed
enum class Color { red, green, blue };

// Specify when needed for binary compatibility or size control
enum class SmallFlag : uint8_t { off, on };
enum class Status : int32_t { ok = 0, error = -1 };  // For C API compatibility
```

**Enum.8: Specify enumerator values only when necessary**
```cpp
// Default sequential values are fine
enum class Color { red, green, blue };  // 0, 1, 2

// Specify when needed for compatibility or semantics
enum class HttpStatus {
    ok = 200,
    not_found = 404,
    server_error = 500
};
```

---

## Con: Constants and Immutability

Use `const` and `constexpr` to express immutability and enable compiler optimizations.

**Con.1: By default, make objects immutable**
- Immutable objects are easier to reason about
- Thread-safe by default
- Enable optimizations
```cpp
// GOOD - immutable by default
const int max_attempts = 3;
const string greeting = "Hello";

// Mutable only when necessary
int counter = 0;  // Needs to change
```

**Con.2: By default, make member functions const**
- Mark member functions `const` if they don't modify object state
- Enables use with const objects
- Documents intent
```cpp
class Point {
public:
    int get_x() const { return x; }  // const - doesn't modify
    int get_y() const { return y; }  // const - doesn't modify

    void set_x(int new_x) { x = new_x; }  // non-const - modifies

private:
    int x, y;
};
```

**Con.3: By default, pass pointers and references to const**
```cpp
// GOOD - const when not modifying
void print(const string& s);
void process(const vector<int>& data);
int find(const int* array, size_t size, int value);

// Non-const only when modifying
void update(vector<int>& data);
void reset(int* value);
```

**Con.4: Use const to define objects with values that do not change after construction**
```cpp
const int buffer_size = 1024;
const string application_name = "MyApp";
const Point origin{0, 0};
```

**Con.5: Use constexpr for values that can be computed at compile time**
- `constexpr` enables compile-time evaluation
- Better performance than `const` alone
- Can be used in constant expressions
```cpp
// GOOD - computed at compile time
constexpr int square(int x) { return x * x; }
constexpr int buffer_size = 1024;
constexpr int max_size = square(32);  // Computed at compile time

// Can use in array sizes, template arguments, etc.
array<int, square(10)> data;  // OK: constexpr
```

**Con.6: Use constexpr for constant expressions**
```cpp
constexpr int factorial(int n) {
    return n <= 1 ? 1 : n * factorial(n - 1);
}

constexpr int f10 = factorial(10);  // Computed at compile time
```

---

## T: Templates and Generic Programming (Detailed)

Templates enable generic programming and compile-time polymorphism.

**T.1: Use templates to raise the level of abstraction of code**
```cpp
// Instead of duplicating code for each type
int max(int a, int b) { return a > b ? a : b; }
double max(double a, double b) { return a > b ? a : b; }

// Write once, use with any type
template<typename T>
T max(T a, T b) { return a > b ? a : b; }
```

**T.2: Use templates to express algorithms that apply to many argument types**
```cpp
template<typename Iterator, typename Value>
Iterator find(Iterator first, Iterator last, const Value& val) {
    while (first != last) {
        if (*first == val) return first;
        ++first;
    }
    return last;
}
```

**T.3: Use templates to express containers and ranges**
```cpp
template<typename T>
class Vector {
public:
    void push_back(const T& value);
    T& operator[](size_t index);
    size_t size() const;
private:
    T* data;
    size_t sz;
};
```

**T.10: Specify concepts for all template arguments**
- Concepts (C++20) express template requirements clearly
```cpp
// C++20 - use concepts
template<typename T>
concept Sortable = requires(T a, T b) {
    { a < b } -> std::convertible_to<bool>;
};

template<Sortable T>
void sort(vector<T>& v) {
    // ...
}

// Pre-C++20 - use SFINAE or static_assert
template<typename T>
void sort(vector<T>& v) {
    static_assert(std::is_arithmetic_v<T>, "T must be arithmetic");
    // ...
}
```

**T.11: Whenever possible use standard concepts**
```cpp
#include <concepts>

// Use standard concepts
template<std::integral T>
T gcd(T a, T b) { /*...*/ }

template<std::floating_point T>
T sqrt(T x) { /*...*/ }
```

**T.12: Prefer concept names over auto for local variables**
```cpp
// GOOD - concept names are clearer
std::integral auto x = 42;
std::floating_point auto y = 3.14;

// Less clear
auto x = 42;
auto y = 3.14;
```

**T.13: Prefer the shorthand notation for simple, single-type argument concepts**
```cpp
// GOOD - shorthand
void sort(Sortable auto& container);

// More verbose but equivalent
template<Sortable T>
void sort(T& container);
```

**T.40: Use function objects to pass operations to algorithms**
```cpp
// GOOD - function objects
sort(v.begin(), v.end(), greater<int>());
auto it = find_if(v.begin(), v.end(), [](int x) { return x > 0; });
```

**T.41: Require only essential properties in a template's concepts**
- Don't over-constrain templates
```cpp
// BAD - requires too much
template<typename T>
requires Sortable<T> && Hashable<T> && Printable<T>
void process(T value);

// GOOD - require only what's needed
template<typename T>
requires Sortable<T>
void sort_items(vector<T>& items);
```

**T.42: Use template aliases to simplify notation and hide implementation details**
```cpp
template<typename T>
using StringMap = map<string, T>;

StringMap<int> word_count;  // Clearer than map<string, int>
```

**T.43: Prefer using over typedef for defining aliases**
```cpp
// GOOD - using (works with templates)
template<typename T>
using Vec = vector<T>;

// BAD - typedef (doesn't work with templates)
typedef vector<int> IntVec;  // Can't template this
```

**T.44: Use function templates to deduce class template argument types**
```cpp
// Instead of:
pair<int, string> p{42, "hello"};

// Use make_pair to deduce types:
auto p = make_pair(42, "hello");
```

**T.60: Minimize a template's context dependencies**
- Templates should be self-contained when possible
```cpp
// GOOD - minimal dependencies
template<typename T>
class Stack {
    vector<T> elements;  // Uses standard library, not custom types
public:
    void push(const T& elem) { elements.push_back(elem); }
    T pop();
};
```

**T.61: Do not over-parameterize members (SCARY)**
- Keep template parameters at class level when shared
```cpp
// BAD - over-parameterized
template<typename T>
class Vector {
public:
    template<typename Alloc>
    class iterator { /*...*/ };  // Doesn't need to know about Alloc
};

// GOOD
template<typename T>
class Vector {
public:
    class iterator { /*...*/ };  // Simpler
};
```

**T.64: Use specialization to provide alternative implementations of class templates**
```cpp
// General template
template<typename T>
class Vector { /*...*/ };

// Specialization for bool
template<>
class Vector<bool> {
    // Bit-packed implementation
};
```

**T.65: Use tag dispatch to provide alternative implementations of functions**
```cpp
template<typename Iterator>
void advance_impl(Iterator& it, int n, random_access_iterator_tag) {
    it += n;  // O(1) for random access
}

template<typename Iterator>
void advance_impl(Iterator& it, int n, input_iterator_tag) {
    while (n--) ++it;  // O(n) for input iterators
}

template<typename Iterator>
void advance(Iterator& it, int n) {
    advance_impl(it, n, typename iterator_traits<Iterator>::iterator_category{});
}
```

**T.69: Inside a template, don't make an unqualified non-member function call unless you intend it to be a customization point**
```cpp
template<typename T>
void f(T x) {
    // BAD - might call wrong function
    swap(x, y);

    // GOOD - qualified call to std::swap
    std::swap(x, y);

    // GOOD - ADL-enabled customization point
    using std::swap;
    swap(x, y);  // Finds T::swap or std::swap
}
```

---

## Per: Performance

**Per.1: Don't optimize without reason**
- Premature optimization is the root of much evil
- Correct first, fast second
```cpp
// Write clear code first
vector<int> filter_positive(const vector<int>& data) {
    vector<int> result;
    for (int x : data) {
        if (x > 0) result.push_back(x);
    }
    return result;
}

// Optimize only if profiling shows it's a bottleneck
```

**Per.2: Don't optimize prematurely**
- "Premature optimization is the root of all evil" - Donald Knuth
- Most code isn't performance critical
- Focus on correctness and clarity first

**Per.3: Don't optimize something that's not performance critical**
- Profile to find actual bottlenecks
- 90% of execution time is in 10% of code
```cpp
// Don't optimize this:
string get_greeting() {
    return "Hello";  // Called once at startup
}

// Do optimize this if profiling shows it's slow:
void process_million_records(const vector<Record>& records) {
    // Actually performance-critical
}
```

**Per.4: Don't assume that complicated code is necessarily faster than simple code**
- Compilers are very good at optimizing simple code
- Complex code is harder to maintain and often slower
```cpp
// Simple and fast
int sum = 0;
for (int x : data) {
    sum += x;
}

// Complex and not necessarily faster
int sum = std::accumulate(data.begin(), data.end(), 0,
    [](int a, int b) { return a + b; });
```

**Per.5: Don't assume that low-level code is necessarily faster than high-level code**
- High-level abstractions often compile to efficient code
- Standard library is heavily optimized
```cpp
// Don't write this thinking it's faster:
char* p = buffer;
while (*p != '\0') {
    // manual loop
    ++p;
}

// This is likely just as fast and clearer:
string s(buffer);
for (char c : s) {
    // ...
}
```

**Per.6: Don't make claims about performance without measurements**
- Always profile
- Measure before and after optimization
- Performance is often counterintuitive
```cpp
// Measure, don't guess!
auto start = chrono::high_resolution_clock::now();
// ... code to measure ...
auto end = chrono::high_resolution_clock::now();
auto duration = chrono::duration_cast<chrono::microseconds>(end - start);
cout << "Time: " << duration.count() << " microseconds\n";
```

**Per.7: Design to enable optimization**
- Loose coupling enables optimization
- Immutability enables optimization
- Value semantics enable optimization
```cpp
// GOOD - enables move optimization
vector<int> create_data() {
    vector<int> result(1000000);
    // ... populate result ...
    return result;  // Moved, not copied
}
```

**Per.10: Rely on the static type system**
- Compile-time polymorphism (templates) faster than runtime polymorphism (virtual)
- Use when performance matters
```cpp
// Runtime polymorphism - virtual function call overhead
class Animal {
public:
    virtual void make_sound() = 0;
};

// Compile-time polymorphism - zero overhead
template<typename Animal>
void make_sound(Animal& a) {
    a.make_sound();  // Resolved at compile time
}
```

**Per.11: Move computation from run time to compile time**
- Use `constexpr` when possible
- Template metaprogramming for complex compile-time computation
```cpp
// Computed at runtime
int factorial(int n) {
    return n <= 1 ? 1 : n * factorial(n - 1);
}

// Computed at compile time
constexpr int factorial(int n) {
    return n <= 1 ? 1 : n * factorial(n - 1);
}

constexpr int f10 = factorial(10);  // Computed once at compile time
```

**Per.13: Use support libraries, don't reinvent the wheel**
- Standard library implementations are highly optimized
- Tested and maintained
```cpp
// Don't write your own sort
// Use std::sort - highly optimized

sort(v.begin(), v.end());
```

**Per.15: Do not allocate on a critical branch**
- Memory allocation is expensive
- Reuse allocations in hot paths
```cpp
// BAD - allocates in loop
for (int i = 0; i < 1000000; ++i) {
    vector<int> temp(100);  // Allocates every iteration!
    // ... use temp ...
}

// GOOD - allocate once
vector<int> temp(100);
for (int i = 0; i < 1000000; ++i) {
    temp.clear();  // Doesn't deallocate
    // ... use temp ...
}
```

**Per.19: Access memory predictably**
- Cache-friendly access patterns
- Linear traversal better than random access
```cpp
// GOOD - cache-friendly (row-major order)
for (int i = 0; i < rows; ++i) {
    for (int j = 0; j < cols; ++j) {
        process(matrix[i][j]);
    }
}

// BAD - cache-unfriendly (column-major in row-major array)
for (int j = 0; j < cols; ++j) {
    for (int i = 0; i < rows; ++i) {
        process(matrix[i][j]);  // Jumps around memory
    }
}
```

---

## SF: Source Files

**SF.1: Use a .cpp suffix for code files and .h for interface files**
- Consistent naming helps tools and developers
```
// GOOD
myclass.h      // Interface
myclass.cpp    // Implementation
```

**SF.2: A .h file must not contain object definitions or non-inline function definitions**
```cpp
// myclass.h
// BAD - definition in header
int global_counter = 0;  // Multiple definition error!

// GOOD - declaration only
extern int global_counter;  // Declaration

// myclass.cpp
int global_counter = 0;  // Definition
```

**SF.3: Use .h files for all declarations used in multiple source files**
```cpp
// utils.h - shared declarations
void log(const string& message);
class Logger { /*...*/ };

// main.cpp
#include "utils.h"
// ... use log() ...

// other.cpp
#include "utils.h"
// ... use log() ...
```

**SF.4: Include .h files before other declarations in a .cpp file**
```cpp
// myclass.cpp
#include "myclass.h"  // First!
#include <string>
#include <vector>

// ... other code ...
```

**SF.5: A .cpp file must include the .h file(s) that defines its interface**
- Ensures declarations and definitions match
```cpp
// widget.h
class Widget {
public:
    void process();
};

// widget.cpp
#include "widget.h"  // Must include!

void Widget::process() {
    // Implementation
}
```

**SF.6: Use using namespace directives for transition, for foundation libraries, or within a local scope only**
```cpp
// BAD - in header file
using namespace std;  // Pollutes all includers!

// OK - in implementation file
// main.cpp
using namespace std;  // Only affects this file

// BETTER - in local scope
void f() {
    using namespace std;  // Only affects this function
    cout << "Hello\n";
}

// BEST - don't use at all for small projects
void f() {
    std::cout << "Hello\n";
}
```

**SF.7: Don't write using namespace at global scope in a header file**
- Forces all includers to import that namespace
```cpp
// BAD - header.h
using namespace std;  // Affects everyone who includes this!

// GOOD - header.h
// No using namespace

// implementation.cpp
using namespace std;  // OK in .cpp file
```

**SF.8: Use #include guards for all .h files**
```cpp
// myclass.h

// Traditional include guards
#ifndef MYCLASS_H
#define MYCLASS_H

class MyClass {
    // ...
};

#endif  // MYCLASS_H

// Alternative: #pragma once (non-standard but widely supported)
#pragma once

class MyClass {
    // ...
};
```

**SF.9: Avoid cyclic dependencies among source files**
- Cyclic dependencies complicate builds and understanding
```cpp
// BAD
// a.h includes b.h
// b.h includes a.h
// = cycle!

// GOOD - break cycle
// Forward declare instead of including
class B;  // Forward declaration

class A {
    B* ptr;  // Pointer/reference to B (doesn't need full definition)
};
```

**SF.10: Avoid dependencies on implicitly #included names**
- Explicitly include what you use
```cpp
// BAD - relies on transitive include
// some_header.h includes <string>
#include "some_header.h"
// ... use std::string ...  // Fragile!

// GOOD - explicitly include what you use
#include "some_header.h"
#include <string>  // Explicitly include
// ... use std::string ...  // Robust
```

**SF.11: Header files should be self-contained**
- Header should compile on its own
```cpp
// myclass.h - should compile standalone
#include <string>   // Include dependencies
#include <vector>

class MyClass {
    std::string name;
    std::vector<int> data;
};
```

**SF.12: Prefer the quoted form of #include for files relative to the including file**
```cpp
// For project files
#include "myclass.h"      // Quoted
#include "utils/helper.h"

// For system/library files
#include <string>         // Angle brackets
#include <vector>
```

---

## SL: The Standard Library

**SL.1: Use libraries wherever possible**
- Don't reinvent the wheel
- Standard library is well-tested and optimized
```cpp
// Don't write your own:
// - sort → use std::sort
// - find → use std::find
// - dynamic array → use std::vector
```

**SL.2: Prefer the standard library to other libraries**
- Standard library is portable
- Well-documented
- Maintained
```cpp
// GOOD - standard library
#include <vector>
#include <algorithm>
#include <string>

vector<int> data;
sort(data.begin(), data.end());
```

**SL.3: Do not add non-standard entities to namespace std**
- Undefined behavior
- Breaks ABI
```cpp
// BAD - undefined behavior!
namespace std {
    void my_function();  // DON'T DO THIS!
}

// GOOD - use your own namespace
namespace mylib {
    void my_function();
}
```

**SL.4: Use the standard library in a type-safe manner**
```cpp
// BAD - type unsafe
void* ptr = malloc(sizeof(int) * 10);
// ...
free(ptr);

// GOOD - type safe
vector<int> data(10);
```

**SL.con.1: Prefer using STL array or vector instead of a C array**
```cpp
// BAD - C array
int data[100];
int size = sizeof(data) / sizeof(data[0]);  // Error-prone

// GOOD - std::array for fixed size
array<int, 100> data;
int size = data.size();  // Safe

// GOOD - std::vector for dynamic size
vector<int> data(100);
data.resize(200);  // Can grow
```

**SL.con.2: Prefer using STL vector by default unless you have a reason to use a different container**
- Vector is fast, flexible, and cache-friendly
```cpp
// Default choice
vector<int> data;

// Other containers only when you need their specific properties:
// - deque: frequent insertion/deletion at both ends
// - list: frequent insertion/deletion in middle
// - set: sorted unique elements
// - map: key-value pairs
```

**SL.con.3: Avoid bounds errors**
```cpp
// BAD - no bounds checking
vector<int> v(10);
v[100] = 42;  // Undefined behavior!

// GOOD - bounds checking
v.at(100) = 42;  // Throws out_of_range exception

// BETTER - avoid indexing
for (int& x : v) {
    // ...
}
```

**SL.str.1: Use std::string to own character sequences**
```cpp
// BAD - manual memory management
char* s = new char[100];
strcpy(s, "hello");
// ... easy to forget delete[] ...
delete[] s;

// GOOD - automatic memory management
string s = "hello";
// ... automatically cleaned up ...
```

**SL.str.2: Use std::string_view or gsl::span to refer to character sequences**
```cpp
// GOOD - non-owning view (C++17)
void print(string_view sv) {
    cout << sv;
}

print("hello");           // From string literal
print(some_string);       // From std::string
print(some_string.substr(0, 5));  // From substring
```

**SL.io.1: Use character-level input only when you have to**
```cpp
// BAD - character by character
char c;
while (cin.get(c)) {
    // ... process one char at a time (slow!)
}

// GOOD - line by line
string line;
while (getline(cin, line)) {
    // ... process whole lines (faster)
}
```

**SL.io.2: When reading, always consider ill-formed input**
```cpp
int x;
if (cin >> x) {
    // Successfully read integer
} else {
    // Handle error (non-integer input)
    cin.clear();  // Clear error state
    cin.ignore(numeric_limits<streamsize>::max(), '\n');  // Skip bad input
}
```

**SL.io.3: Prefer iostream for I/O**
- Type-safe
- Extensible
- Don't use printf/scanf
```cpp
// BAD - printf (type-unsafe)
printf("%d", some_value);  // What if some_value isn't an int?

// GOOD - iostream (type-safe)
cout << some_value;  // Works with any printable type
```

**SL.C.1: Don't use setjmp/longjmp**
- Use exceptions for error handling
- setjmp/longjmp bypass destructors (breaks RAII)
```cpp
// BAD - setjmp/longjmp
jmp_buf buf;
if (setjmp(buf)) {
    // Error handling
}

// GOOD - exceptions
try {
    // ... code that might throw ...
} catch (const exception& e) {
    // Error handling
}
```

## Quick Reference Checklist

When writing C++ code, check:

### Effective C++ (Scott Meyers)
- [ ] Used const wherever possible
- [ ] Initialized all variables before use
- [ ] Used member initialization lists in constructors
- [ ] Made destructors virtual in polymorphic base classes
- [ ] Prevented exceptions from leaving destructors
- [ ] Used smart pointers instead of raw new/delete
- [ ] Matched new/delete and new[]/delete[] correctly
- [ ] Made assignment operators handle self-assignment
- [ ] Copied all parts in copy constructor/assignment (including base class)
- [ ] Returned reference to *this from assignment operators
- [ ] Used Rule of Zero when possible (no manual resource management)
- [ ] Defined Rule of Five consistently when managing resources
- [ ] Made move operations noexcept when possible
- [ ] Preferred std::unique_ptr over std::shared_ptr
- [ ] Used std::make_unique and std::make_shared
- [ ] Never called virtual functions from constructors/destructors

### C++ Core Guidelines
- [ ] Expressed ideas directly in code (P.1)
- [ ] Made interfaces explicit (I.1)
- [ ] Used RAII for all resource management (R.1, P.8)
- [ ] Preferred stack allocation to heap allocation (R.5)
- [ ] Used smart pointers for ownership (R.20)
- [ ] Preferred unique_ptr over shared_ptr (R.21)
- [ ] Avoided raw new/delete (R.11, ES.60)
- [ ] Kept functions short and simple (F.3)
- [ ] Made functions perform single logical operations (F.2)
- [ ] Passed expensive types by const reference (F.15)
- [ ] Used strong types in interfaces (I.4)
- [ ] Made single-argument constructors explicit (C.46)
- [ ] Used in-class member initializers (C.45)
- [ ] Made base class destructors public/virtual or protected/non-virtual (C.35)
- [ ] Preferred standard library to handcrafted code (ES.1)
- [ ] Initialized all objects (ES.20)
- [ ] Used auto to avoid type repetition (ES.11)
- [ ] Preferred {} initialization syntax (ES.23)
- [ ] Used named casts, never C-style casts (ES.49)
- [ ] Kept scopes small (ES.5)
- [ ] Declared one name per declaration (ES.10)
- [ ] Used enum class over plain enum (Enum.3)
- [ ] Defined operations on enumerations (Enum.4)
- [ ] Made constants const or constexpr (Con.1)
- [ ] Made member functions const by default (Con.2)
- [ ] Passed pointers to const for data not modified (Con.3)
- [ ] Used constexpr for compile-time evaluation (Con.5)
- [ ] Used concepts to express template requirements (T.10)
- [ ] Minimized template argument requirements (T.11)
- [ ] Avoided unnecessary template parameters (T.12)
- [ ] Avoided premature optimization (Per.1)
- [ ] Measured before optimizing (Per.2)
- [ ] Used #include guards or #pragma once (SF.8)
- [ ] Avoided cycles in #include dependencies (SF.9)
- [ ] Preferred standard library over custom code (SL.1)
- [ ] Preferred algorithm calls over hand-written loops (SL.2)
