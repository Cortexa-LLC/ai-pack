# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Language-Specific Indentation Standards:** Each language now follows its community's standard
  - C++: 2 spaces (Google C++ Style Guide)
  - Python: 4 spaces (PEP 8 mandatory)
  - JavaScript/TypeScript: 2 spaces (Microsoft TypeScript Guidelines adapted)
  - Java: 4 spaces (Oracle Java Conventions)
  - Kotlin: 4 spaces (Kotlin Coding Conventions)
- **lang-python.md:** Python best practices stub (PEP 8, type hints)
- **lang-javascript.md:** JavaScript/TypeScript guidelines following Microsoft TypeScript Coding Guidelines with 2-space indentation
  - Double quotes for strings
  - Prefer undefined over null
  - No I prefix for interfaces
  - One declaration per variable statement
  - Arrow functions over anonymous expressions
- **lang-java.md:** Java guidelines stub (Oracle, Effective Java, Spring)
- **lang-kotlin.md:** Kotlin conventions stub (JetBrains, Coroutines, Android)
- **Two-Tier Rule System:** Documentation for shared vs project-specific rules
- **PROJECT-*.md pattern:** Git-ignored pattern for project-specific rules in submodule

### Changed
- Updated company references to "Cortexa LLC" (full company name)
- Enhanced README with language-specific indentation table
- **lang-javascript.md:** Switched from Airbnb/Google style guides to Microsoft TypeScript Coding Guidelines (with 2-space indentation adaptation)

---

## [1.0.0] - 2025-01-04

### Added

#### Core Design Principles
- 01-design-principles.md: Beck's Four Rules, Tell Don't Ask, Dependency Injection, Seams
- 02-solid-principles.md: Complete SOLID principles with examples and code smells
- 03-refactoring.md: Code smells catalog and refactoring techniques
- 04-testing.md: Test Pyramid, test doubles, testing best practices
- 05-architecture.md: Bounded Contexts and architectural patterns
- 06-code-review-checklist.md: Comprehensive review guidelines

#### Development Practices
- 07-development-practices.md: YAGNI, Frequency Reduces Difficulty, CI, Technical Debt
- 08-deployment-patterns.md: Feature Toggles, Blue-Green, Canary, Parallel Change
- 09-system-evolution.md: Strangler Fig, Sacrificial Architecture, MonolithFirst
- 10-api-design.md: Command Query Separation, naming conventions

#### Language-Specific Guidelines
- lang-cpp.md: Complete C++ guidelines with:
  - All 55 items from Scott Meyers' Effective C++
  - C++ Core Guidelines (P, F, I, C, R, ES, E, CP, Enum, Con, T, Per, SF, SL sections)
  - Modern C++17/20 best practices
  - Extensive code examples and checklists

#### Repository Infrastructure
- README.md with comprehensive usage instructions
- LICENSE (MIT)
- CHANGELOG.md
- .gitignore
- Documentation on Git submodule usage
- Integration guides for AI assistants

### Attribution

Standards synthesized from:
- Martin Fowler (martinfowler.com)
- Kent Beck (Four Rules of Simple Design)
- Robert C. Martin (SOLID Principles)
- Scott Meyers (Effective C++)
- ISO C++ Standards Committee (C++ Core Guidelines)

---

## Versioning Guide

- **Major version (X.0.0)**: Breaking changes to structure, significant rewrites
- **Minor version (0.X.0)**: New guidelines or sections added
- **Patch version (0.0.X)**: Clarifications, typo fixes, improvements

## Future Planned Additions

### Language Guidelines (Planned)
- lang-python.md: Python best practices (PEP 8, PEP 20, typing)
- lang-javascript.md: JavaScript/TypeScript standards
- lang-rust.md: Rust guidelines and idioms
- lang-go.md: Go conventions and best practices

### Additional Patterns (Planned)
- 11-concurrency-patterns.md: Thread safety, async patterns
- 12-database-patterns.md: Repository pattern, ORM best practices
- 13-security-patterns.md: Common security vulnerabilities and mitigations

[1.0.0]: https://github.com/Cortexa-LLC/clean-code-standards/releases/tag/v1.0.0
