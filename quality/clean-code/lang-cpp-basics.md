# C++ Language-Specific Rules - Part 1: Fundamentals

> Based on Scott Meyers' Effective C++ series and C++ Core Guidelines
>
> **Part of a multi-file C++ rules series:**
> - **lang-cpp-basics.md** (this file) - Formatting, language basics, constructors, RAII
> - lang-cpp-design.md - Design, declarations, and implementations
> - lang-cpp-advanced.md - OOP, templates, and advanced topics
> - lang-cpp-modern.md - Modern C++ (C++11/14/17/20)
> - lang-cpp-guidelines.md - C++ Core Guidelines
> - lang-cpp-reference.md - Quick reference checklist

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
