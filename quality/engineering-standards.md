# Engineering Standards Quick Reference

**Version:** 1.1.0
**Last Updated:** 2026-01-08

This document serves as an entrypoint and navigation guide to the comprehensive clean-code standards maintained in this repository.

## Overview

The ai-pack quality standards provide industry-leading clean code principles, design patterns, and best practices synthesized from Martin Fowler, Robert C. Martin (Uncle Bob), Kent Beck, and other software engineering thought leaders.

## Core Principles

Start here for fundamental design and development principles:

### Design Foundations
- **[00-general-rules.md](clean-code/00-general-rules.md)** - Universal standards (formatting, TDD, test coverage)
- **[01-design-principles.md](clean-code/01-design-principles.md)** - Beck's Four Rules, Tell Don't Ask, Dependency Injection
- **[02-solid-principles.md](clean-code/02-solid-principles.md)** - Single Responsibility, Open-Closed, Liskov Substitution, Interface Segregation, Dependency Inversion

### Code Quality
- **[03-refactoring.md](clean-code/03-refactoring.md)** - Code smells catalog and refactoring techniques
- **[04-testing.md](clean-code/04-testing.md)** - Test Pyramid, test doubles, TDD practices
- **[06-code-review-checklist.md](clean-code/06-code-review-checklist.md)** - Comprehensive review guidelines

### Architecture & Deployment
- **[05-architecture.md](clean-code/05-architecture.md)** - Bounded Contexts and architectural patterns
- **[08-deployment-patterns.md](clean-code/08-deployment-patterns.md)** - Feature Toggles, Blue-Green, Canary Release
- **[10-api-design.md](clean-code/10-api-design.md)** - Command Query Separation, API naming conventions

### Development Practices
- **[07-development-practices.md](clean-code/07-development-practices.md)** - YAGNI, Frequency Reduces Difficulty, CI/CD
- **[09-system-evolution.md](clean-code/09-system-evolution.md)** - Managing technical debt and system growth
- **[11-documentation-standards.md](clean-code/11-documentation-standards.md)** - Inline docs, Markdown+Mermaid, ADRs

## Language-Specific Guidelines

Choose the guide for your programming language:

### C++ (Comprehensive Coverage)
- **[lang-cpp-basics.md](clean-code/lang-cpp-basics.md)** - Fundamental C++ concepts
- **[lang-cpp-modern.md](clean-code/lang-cpp-modern.md)** - Modern C++17/20 features
- **[lang-cpp-design.md](clean-code/lang-cpp-design.md)** - C++ design patterns
- **[lang-cpp-advanced.md](clean-code/lang-cpp-advanced.md)** - Advanced topics
- **[lang-cpp-guidelines.md](clean-code/lang-cpp-guidelines.md)** - C++ Core Guidelines (comprehensive)
- **[lang-cpp-reference.md](clean-code/lang-cpp-reference.md)** - Quick reference

### Other Languages
- **[lang-python.md](clean-code/lang-python.md)** - Python (PEP 8, type hints, best practices)
- **[lang-javascript.md](clean-code/lang-javascript.md)** - JavaScript/TypeScript
- **[lang-java.md](clean-code/lang-java.md)** - Java (Oracle conventions, Effective Java)
- **[lang-kotlin.md](clean-code/lang-kotlin.md)** - Kotlin (JetBrains standards)

## Quick Navigation by Use Case

### "I need to..."

**Write clean, maintainable code**
→ Start with [01-design-principles.md](clean-code/01-design-principles.md) and [02-solid-principles.md](clean-code/02-solid-principles.md)

**Refactor existing code**
→ See [03-refactoring.md](clean-code/03-refactoring.md) for code smells and refactoring patterns

**Improve my testing**
→ Read [04-testing.md](clean-code/04-testing.md) for comprehensive testing guidance

**Design an API**
→ Check [10-api-design.md](clean-code/10-api-design.md) for best practices

**Review code**
→ Use [06-code-review-checklist.md](clean-code/06-code-review-checklist.md)

**Deploy safely**
→ Learn deployment patterns in [08-deployment-patterns.md](clean-code/08-deployment-patterns.md)

**Document complex code**
→ See [11-documentation-standards.md](clean-code/11-documentation-standards.md) for inline docs, Markdown+Mermaid, and ADRs

**Work with a specific language**
→ See language-specific guides above

## Standards Overview

### Key Requirements
- **Formatting:** Spaces only (no tabs). 2-space for C++/JS, 4-space for Python/Java/Kotlin
- **Testing:** TDD workflow, 80-90% code coverage target
- **Quality:** All tests must pass (zero tolerance for failures)

### Core Guideline Patterns

These patterns apply across all supported languages (examples shown in C++ but principles translate):

#### Prefer Return Values Over Out-Parameters
```cpp
// ❌ Old style (out-parameters)
void GetSymbol(uint32_t addr, std::string* name, bool* found);

// ✅ Modern (return std::optional)
std::optional<std::string> GetSymbol(uint32_t addr);
```

#### Return Structs for Multiple Values
```cpp
// ❌ Multiple out-parameters
std::string Format(..., uint32_t* target, size_t* bytes, bool* success);

// ✅ Return struct
struct Result {
  std::string value;
  uint32_t target;
  size_t bytes;
  bool success;
};
Result Format(...);
```

#### Use Safe Container Views
```cpp
// ⚠️ Pointer + size (acceptable for compatibility)
void Process(const uint8_t* data, size_t size);

// ✅ Preferred (safer, modern)
void Process(std::span<const uint8_t> data);  // C++20
```

#### Manage Resources with RAII
```cpp
// ✅ Automatic resource management
std::unique_ptr<Resource> CreateResource();
std::vector<uint8_t> buffer;  // automatic cleanup
```

#### Document Intentional Deviations
When you must deviate from guidelines (compatibility, performance):
```cpp
// GUIDELINE DEVIATION (I.13): Binary data interface for plugins
// Will migrate to std::span in C++20
const uint8_t* GetPointer(uint32_t address) const;
```

### Attribution
These standards synthesize best practices from:
- Martin Fowler (design patterns, refactoring, deployment patterns)
- Kent Beck (Four Rules of Simple Design, TDD)
- Robert C. Martin (SOLID principles, Clean Code)
- Scott Meyers (Effective C++)
- ISO C++ Standards Committee (C++ Core Guidelines)
- Language-specific style guides (PEP 8, Oracle Java, JetBrains Kotlin)

## Additional Resources

- **[RULES_REFERENCE.md](clean-code/RULES_REFERENCE.md)** - Quick reference for all rules
- **[SETUP_TWO_TIER_RULES.md](clean-code/SETUP_TWO_TIER_RULES.md)** - Two-tier rule system setup
- **[CHANGELOG.md](clean-code/CHANGELOG.md)** - Version history

## Integration

These standards are designed to integrate with AI coding assistants through:
- Git submodules
- Symbolic links
- Direct copy into project `.ai-pack` directory

See the main README for integration instructions.

---

**Need help?** All files follow a consistent structure:
1. Overview and principles
2. Detailed guidance with examples
3. Anti-patterns to avoid
4. References and further reading
