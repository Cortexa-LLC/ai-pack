# C++ Language-Specific Rules - Part 2: Design & Implementation

> Based on Scott Meyers' Effective C++ series and C++ Core Guidelines
>
> **Part of a multi-file C++ rules series:**
> - lang-cpp-basics.md - Formatting, language basics, constructors, RAII
> - **lang-cpp-design.md** (this file) - Design, declarations, and implementations
> - lang-cpp-advanced.md - OOP, templates, and advanced topics
> - lang-cpp-modern.md - Modern C++ (C++11/14/17/20)
> - lang-cpp-guidelines.md - C++ Core Guidelines
> - lang-cpp-reference.md - Quick reference checklist

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
