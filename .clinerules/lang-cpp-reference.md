# C++ Language-Specific Rules - Part 6: Quick Reference

> Quick reference checklist for C++ best practices
>
> **Part of a multi-file C++ rules series:**
> - lang-cpp-basics.md - Formatting, language basics, constructors, RAII
> - lang-cpp-design.md - Design, declarations, and implementations
> - lang-cpp-advanced.md - OOP, templates, and advanced topics
> - lang-cpp-modern.md - Modern C++ (C++11/14/17/20)
> - lang-cpp-guidelines.md - C++ Core Guidelines
> - **lang-cpp-reference.md** (this file) - Quick reference checklist

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
