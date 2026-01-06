# C++ Language-Specific Rules - Part 5: Core Guidelines

> Based on the C++ Core Guidelines from the ISO C++ Standards Committee
>
> **Part of a multi-file C++ rules series:**
> - lang-cpp-basics.md - Formatting, language basics, constructors, RAII
> - lang-cpp-design.md - Design, declarations, and implementations
> - lang-cpp-advanced.md - OOP, templates, and advanced topics
> - lang-cpp-modern.md - Modern C++ (C++11/14/17/20)
> - **lang-cpp-guidelines.md** (this file) - C++ Core Guidelines
> - lang-cpp-reference.md - Quick reference checklist

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

