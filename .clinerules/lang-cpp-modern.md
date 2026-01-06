# C++ Language-Specific Rules - Part 4: Modern C++

> Based on Scott Meyers' Effective Modern C++ and C++11/14/17/20 features
>
> **Part of a multi-file C++ rules series:**
> - lang-cpp-basics.md - Formatting, language basics, constructors, RAII
> - lang-cpp-design.md - Design, declarations, and implementations
> - lang-cpp-advanced.md - OOP, templates, and advanced topics
> - **lang-cpp-modern.md** (this file) - Modern C++ (C++11/14/17/20)
> - lang-cpp-guidelines.md - C++ Core Guidelines
> - lang-cpp-reference.md - Quick reference checklist

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

