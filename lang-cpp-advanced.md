# C++ Language-Specific Rules - Part 3: Advanced Topics

> Based on Scott Meyers' Effective C++ series and C++ Core Guidelines
>
> **Part of a multi-file C++ rules series:**
> - lang-cpp-basics.md - Formatting, language basics, constructors, RAII
> - lang-cpp-design.md - Design, declarations, and implementations
> - **lang-cpp-advanced.md** (this file) - OOP, templates, and advanced topics
> - lang-cpp-modern.md - Modern C++ (C++11/14/17/20)
> - lang-cpp-guidelines.md - C++ Core Guidelines
> - lang-cpp-reference.md - Quick reference checklist

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

