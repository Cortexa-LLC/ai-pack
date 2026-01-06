# Clean Code Rules - Martin Fowler Principles

This `.clinerules` folder contains comprehensive guidelines based on Martin Fowler's principles from [martinfowler.com](https://martinfowler.com), covering SOLID principles, design patterns, refactoring techniques, and testability.

## Contents

### 01-design-principles.md
Core design principles that form the foundation of clean code:
- Beck's Four Rules of Simple Design
- Tell Don't Ask Principle
- Inversion of Control & Dependency Injection
- Seams for Testability

### 02-solid-principles.md
Practical application of SOLID principles with examples:
- Single Responsibility Principle (SRP)
- Open/Closed Principle (OCP)
- Liskov Substitution Principle (LSP)
- Interface Segregation Principle (ISP)
- Dependency Inversion Principle (DIP)

### 03-refactoring.md
Understanding and addressing code quality issues:
- Code Smells Catalog (Long Method, Data Class, Feature Envy, etc.)
- Refactoring Flow and Techniques
- When and How to Refactor Safely
- Anti-patterns to Avoid

### 04-testing.md
Comprehensive testing strategies and principles:
- The Test Pyramid (Unit → Integration → E2E)
- Unit Test Effectiveness
- Test Doubles (Dummy, Fake, Stub, Spy, Mock)
- Test Coverage as Diagnostic Tool
- Testability Through Design

### 05-architecture.md
Organizing and structuring larger systems:
- Bounded Contexts for Large Systems
- Code Organization Principles
- Layered Architecture
- Dependency Management
- Evolutionary Architecture

### 06-code-review-checklist.md
Practical checklist for reviewing code quality:
- Design Quality (Beck's Four Rules)
- SOLID Principles Compliance
- Code Smells Detection
- Testability Assessment
- Architecture and Organization

### 07-development-practices.md
Daily development workflow principles:
- YAGNI (You Aren't Gonna Need It)
- Frequency Reduces Difficulty
- Continuous Integration
- Technical Debt Management
- True Definition of Refactoring
- Naming Principles

### 08-deployment-patterns.md
Safe deployment and release strategies:
- Feature Toggles (Feature Flags)
- Blue-Green Deployment
- Canary Release
- Parallel Change (Expand-Contract)
- Pattern Selection Guide

### 09-system-evolution.md
Long-term system evolution strategies:
- Strangler Fig Pattern for Legacy Modernization
- Sacrificial Architecture
- MonolithFirst Approach
- Semantic Diffusion Awareness
- Evolution Pattern Selection

### 10-api-design.md
Interface and API design principles:
- Command Query Separation (CQS)
- Naming Things - The Hard Problem
- Interface Design Checklist
- Common API Design Mistakes

### C++ Language-Specific Rules (Multi-File Series)

C++ best practices organized across six focused files:

#### lang-cpp-basics.md
Fundamentals and resource management:
- Formatting standards (2-space indentation for C++)
- Language basics and const correctness
- Constructors, destructors, and assignment operators
- RAII (Resource Acquisition Is Initialization)
- Smart pointers and resource management

#### lang-cpp-design.md
Design and implementation principles:
- Interface design and declarations
- Pass-by-reference vs pass-by-value
- Encapsulation and data member access
- Exception safety guarantees
- Compilation dependencies and the pimpl idiom

#### lang-cpp-advanced.md
Advanced OOP and templates:
- Inheritance hierarchies and polymorphism
- Virtual functions and alternatives
- Template programming fundamentals
- Custom new/delete operators
- Standard library familiarity

#### lang-cpp-modern.md
Modern C++ features (C++11/14/17/20):
- Move semantics and rvalue references
- Lambda expressions and auto keyword
- Smart pointer best practices
- Range-based for loops
- Modern threading and concurrency

#### lang-cpp-guidelines.md
C++ Core Guidelines from ISO C++:
- Comprehensive guidelines from the C++ Standards Committee
- Enum, constants, and immutability patterns
- Template and generic programming patterns
- Performance guidelines
- Standard library usage

#### lang-cpp-reference.md
Quick reference checklist:
- Condensed best practices for daily use
- Common patterns and anti-patterns
- Quick decision guides

## How to Use These Rules

These rules are automatically processed by Claude Code and inform code review, refactoring decisions, and new code development.

### For Development
- Reference principles when making design decisions
- Use code smell catalog to identify improvement opportunities
- Apply refactoring techniques systematically
- Design for testability from the start

### For Code Review
- Use the checklist in `06-code-review-checklist.md`
- Reference specific principles when providing feedback
- Balance pragmatism with principle adherence

### For Learning
- Read one file at a time to build understanding
- Study examples and anti-patterns
- Discuss principles with your team
- Apply incrementally as you learn

## Key Principles Summary

**Beck's Four Rules (Priority Order):**
1. Passes the tests
2. Reveals intention
3. No duplication
4. Fewest elements

**Core Maxims:**
- Tell Don't Ask - Tell objects what to do, don't query and decide externally
- YAGNI - You Aren't Gonna Need It; avoid building for hypothetical futures
- Frequency Reduces Difficulty - If it hurts, do it more often
- Separate configuration from use - Dependency injection over service location
- Test at the right level - Push tests down the pyramid
- Coverage is diagnostic, not a target - Focus on thoughtful testing
- MonolithFirst - Start simple, extract services when boundaries are clear
- Command Query Separation - Separate state-changing from state-querying methods

## References

All principles derived from Martin Fowler's work at [martinfowler.com](https://martinfowler.com):

**Design & Architecture:**
- [Beck's Design Rules](https://martinfowler.com/bliki/BeckDesignRules.html)
- [Tell Don't Ask](https://martinfowler.com/bliki/TellDontAsk.html)
- [Bounded Context](https://martinfowler.com/bliki/BoundedContext.html)
- [Command Query Separation](https://martinfowler.com/bliki/CommandQuerySeparation.html)

**Refactoring & Quality:**
- [Code Smells](https://martinfowler.com/bliki/CodeSmell.html)
- [Refactoring Malapropism](https://martinfowler.com/bliki/RefactoringMalapropism.html)
- [Technical Debt](https://martinfowler.com/bliki/TechnicalDebt.html)

**Testing:**
- [Practical Test Pyramid](https://martinfowler.com/articles/practical-test-pyramid.html)
- [Test Doubles](https://martinfowler.com/bliki/TestDouble.html)
- [Test Coverage](https://martinfowler.com/bliki/TestCoverage.html)
- [Unit Tests](https://martinfowler.com/bliki/UnitTest.html)

**Dependency Management:**
- [Dependency Injection](https://martinfowler.com/articles/injection.html)
- [Inversion of Control](https://martinfowler.com/bliki/InversionOfControl.html)

**Development Practices:**
- [YAGNI](https://martinfowler.com/bliki/Yagni.html)
- [Frequency Reduces Difficulty](https://martinfowler.com/bliki/FrequencyReducesDifficulty.html)
- [Continuous Integration](https://martinfowler.com/articles/continuousIntegration.html)
- [Two Hard Things](https://martinfowler.com/bliki/TwoHardThings.html)

**Deployment Patterns:**
- [Feature Toggles](https://martinfowler.com/bliki/FeatureToggle.html)
- [Blue-Green Deployment](https://martinfowler.com/bliki/BlueGreenDeployment.html)
- [Canary Release](https://martinfowler.com/bliki/CanaryRelease.html)
- [Parallel Change](https://martinfowler.com/bliki/ParallelChange.html)

**System Evolution:**
- [Strangler Fig Application](https://martinfowler.com/bliki/StranglerFigApplication.html)
- [Sacrificial Architecture](https://martinfowler.com/bliki/SacrificialArchitecture.html)
- [MonolithFirst](https://martinfowler.com/bliki/MonolithFirst.html)
- [Semantic Diffusion](https://martinfowler.com/bliki/SemanticDiffusion.html)

## Context and Balance

These are guidelines, not rigid rules. Apply them:
- **Pragmatically** based on your context
- **Incrementally** as your understanding grows
- **Balanced** against other design concerns
- **Thoughtfully** with empathy for future maintainers

The goal is better code, not perfect adherence to principles.
