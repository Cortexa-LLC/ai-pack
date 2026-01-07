# Clean Code Standards

**Comprehensive coding standards and best practices for software development**

A curated collection of clean code principles, design patterns, and language-specific guidelines based on industry-leading sources including Martin Fowler's design principles, SOLID principles, and modern language-specific best practices.

## Overview

This repository serves as a centralized, reusable set of coding standards that can be integrated into any project via Git submodules, symbolic links, or direct inclusion. These standards are designed to work seamlessly with AI coding assistants like Claude Code through `.clinerules` integration.

## Contents

### Universal Standards

- **[00-general-rules.md](.clinerules/00-general-rules.md)** - Universal rules: No tabs (spaces only), TDD workflow, test coverage guidelines (80-90%)

### Core Design Principles

- **[01-design-principles.md](.clinerules/01-design-principles.md)** - Beck's Four Rules of Simple Design, Tell Don't Ask, Dependency Injection, and Seams for Testability
- **[02-solid-principles.md](.clinerules/02-solid-principles.md)** - Complete SOLID principles with practical examples and code smells
- **[03-refactoring.md](.clinerules/03-refactoring.md)** - Code smells catalog and refactoring techniques
- **[04-testing.md](.clinerules/04-testing.md)** - Test Pyramid, test doubles, and testing best practices
- **[05-architecture.md](.clinerules/05-architecture.md)** - Bounded Contexts and architectural patterns
- **[06-code-review-checklist.md](.clinerules/06-code-review-checklist.md)** - Comprehensive code review guidelines

### Development Practices

- **[07-development-practices.md](.clinerules/07-development-practices.md)** - YAGNI, Frequency Reduces Difficulty, Continuous Integration, Technical Debt, Refactoring
- **[08-deployment-patterns.md](.clinerules/08-deployment-patterns.md)** - Feature Toggles, Blue-Green Deployment, Canary Release, Parallel Change
- **[09-system-evolution.md](.clinerules/09-system-evolution.md)** - Strangler Fig, Sacrificial Architecture, MonolithFirst, Semantic Diffusion

### API and Interface Design

- **[10-api-design.md](.clinerules/10-api-design.md)** - Command Query Separation, Naming conventions, API design principles

### Language-Specific Guidelines

Each language follows its community's established indentation standards:

| Language | Indentation | File |
|----------|-------------|------|
| C++ | 2 spaces | [lang-cpp.md](.clinerules/lang-cpp.md) |
| Python | 4 spaces | [lang-python.md](.clinerules/lang-python.md) |
| JavaScript/TypeScript | 2 spaces | [lang-javascript.md](.clinerules/lang-javascript.md) |
| Java | 4 spaces | [lang-java.md](.clinerules/lang-java.md) |
| Kotlin | 4 spaces | [lang-kotlin.md](.clinerules/lang-kotlin.md) |

**[lang-cpp.md](.clinerules/lang-cpp.md)** - Comprehensive C++ guidelines:
- All 55 items from Scott Meyers' *Effective C++*
- C++ Core Guidelines (P, F, I, C, R, ES, E, CP, Enum, Con, T, Per, SF, SL sections)
- Modern C++17/20 best practices
- 2-space indentation (Google C++ Style Guide)

**[lang-python.md](.clinerules/lang-python.md)** - Python best practices:
- PEP 8, PEP 20 (Zen of Python)
- Type hints and modern Python features
- 4-space indentation (PEP 8 mandatory)

**[lang-javascript.md](.clinerules/lang-javascript.md)** - JavaScript/TypeScript:
- Microsoft TypeScript Coding Guidelines
- Double quotes, prefer undefined over null, no I prefix
- 2-space indentation (JavaScript ecosystem standard)

**[lang-java.md](.clinerules/lang-java.md)** - Java guidelines:
- Oracle Java Code Conventions
- Effective Java (Joshua Bloch)
- Spring Framework patterns
- 4-space indentation (Oracle standard)

**[lang-kotlin.md](.clinerules/lang-kotlin.md)** - Kotlin conventions:
- Kotlin Coding Conventions (JetBrains)
- Coroutines and Flow patterns
- Android best practices
- 4-space indentation (Kotlin standard)

## Usage

### Option 1: Git Submodule (Recommended for Teams)

Add these standards to your project as a submodule:

```bash
cd your-project
git submodule add https://github.com/Cortexa-LLC/clean-code-standards .clinerules
git submodule update --init --recursive
```

Update standards in your project:

```bash
git submodule update --remote
git add .clinerules
git commit -m "Update clean code standards"
```

### Option 2: Symbolic Link (For Local Development)

Create a symbolic link to a single shared copy:

```bash
cd your-project
ln -s /path/to/clean-code-standards/.clinerules .clinerules
```

### Option 3: Direct Copy (For Standalone Projects)

Copy the standards directly into your project:

```bash
cp -r /path/to/clean-code-standards/.clinerules your-project/.clinerules
```

## Integration with AI Assistants

These standards are designed to work with Claude Code and other AI assistants that support `.clinerules`:

1. Add this repository as a submodule or link in your project
2. The `.clinerules` folder will be automatically discovered by Claude Code
3. AI assistants will apply these standards during code generation and review

## Project-Specific Customization

### Two-Tier Rule System

This repository supports a **two-tier approach** for managing shared and project-specific rules:

**Tier 1: Shared Standards** (from this submodule)
- Core design principles
- SOLID principles
- Language-specific guidelines
- Universal best practices

**Tier 2: Project-Specific Rules** (in your project)
- Project conventions
- Team preferences
- Technology stack specifics
- Workflow and tooling

### How to Add Project-Specific Rules

When you add this repository as a submodule to `.clinerules/`, you can also add project-specific rule files directly to the same directory. These files are git-ignored by the submodule but tracked in your project repository.

**Naming Convention for Project Files:**
- `PROJECT-*.md` - Project-specific rules (e.g., `PROJECT-sourcerer.md`)
- `PROJECT-README.md` - Overview of your project's rule structure

**Example Setup:**

```bash
# Add submodule
git submodule add git@github.com:Cortexa-LLC/clean-code.git .clinerules

# Add project-specific rules to the same directory
cat > .clinerules/PROJECT-README.md << 'EOF'
# Sourcerer Coding Standards

This project uses a **two-tier rule system**:

## Tier 1: Shared Standards (Submodule)
All files without `PROJECT-` prefix come from the Cortexa clean-code standards.

## Tier 2: Project-Specific Rules
- `PROJECT-sourcerer.md` - Sourcerer-specific conventions
- `PROJECT-architecture.md` - Plugin architecture rules

**Both tiers are automatically discovered by Claude Code and other AI assistants.**
EOF

# Add your project rules
cat > .clinerules/PROJECT-sourcerer.md << 'EOF'
# Sourcerer-Specific Rules

## Formatting
- NO TABS - Use 2-space indentation
- Line length: 100 chars (soft), 120 chars (hard)

## Architecture
- CodeAnalyzer must be CPU-agnostic
- Use plugin pattern for extensibility
EOF

# Commit project files (submodule files are not committed to parent)
git add .clinerules/PROJECT-*.md
git commit -m "Add project-specific coding rules"
```

### How AI Assistants Discover Both Tiers

**Claude Code and similar tools automatically:**
1. ✅ Read all `.md` files in `.clinerules/` directory
2. ✅ Include both submodule files (shared standards)
3. ✅ Include project-specific files (PROJECT-*.md pattern)
4. ✅ Apply both sets of rules during code generation and review

**No additional configuration needed!** Just place your project files in `.clinerules/` with the `PROJECT-` prefix.

### Example Directory Structure

After setup, your project's `.clinerules/` contains both shared and project files:

```
.clinerules/                              # Git submodule + project files
├── README.md                            # Shared (from submodule)
├── 01-design-principles.md              # Shared (from submodule)
├── 02-solid-principles.md               # Shared (from submodule)
├── lang-cpp.md                          # Shared (from submodule)
├── PROJECT-README.md                    # Project-specific (your file)
├── PROJECT-sourcerer.md                 # Project-specific (your file)
└── PROJECT-architecture.md              # Project-specific (your file)
```

**Git Behavior:**
- Submodule tracks: All files except `PROJECT-*`
- Parent project tracks: Only `PROJECT-*` files
- Updates to submodule don't conflict with your project files

## Sources and Attribution

This repository synthesizes best practices from:

- **Martin Fowler** - [martinfowler.com](https://martinfowler.com)
  - Design Patterns, Refactoring, Deployment Patterns
- **Kent Beck** - Four Rules of Simple Design
- **Robert C. Martin (Uncle Bob)** - SOLID Principles
- **Scott Meyers** - *Effective C++*
- **ISO C++ Standards Committee** - C++ Core Guidelines
- **Google** - Industry best practices

## Using These Standards in Your Projects

### Code Reviews

Reference specific guidelines during code reviews:

```
This violates the Single Responsibility Principle (02-solid-principles.md).
Consider extracting the database logic into a separate repository class.
```

### CI/CD Integration

Add automated checks based on these standards:

```yaml
# .github/workflows/code-quality.yml
- name: Check code compliance
  run: |
    # Run linters configured per these standards
    clang-tidy --config-file=.clinerules/clang-tidy-config
```

### Team Onboarding

Use as onboarding material for new team members:

1. **Week 1:** Read design principles (01-06)
2. **Week 2:** Study development practices (07-09)
3. **Week 3:** Review language-specific guidelines (lang-*)

## Contributing

To suggest improvements or additions:

1. Fork this repository
2. Create a feature branch (`git checkout -b feature/new-guideline`)
3. Make your changes with clear documentation
4. Submit a pull request

### Contribution Guidelines

- Provide concrete code examples for each principle
- Include both "good" and "bad" examples
- Cite authoritative sources
- Keep language neutral in core principles (language-specific details go in `lang-*` files)

## Repository Structure

```
clean-code-standards/
├── README.md                          # This file
├── LICENSE                            # MIT License
├── .gitignore                         # Git ignore rules
└── .clinerules/                       # Standards directory
    ├── README.md                      # Quick reference guide
    ├── 01-design-principles.md        # Core design principles
    ├── 02-solid-principles.md         # SOLID principles
    ├── 03-refactoring.md              # Refactoring guidelines
    ├── 04-testing.md                  # Testing standards
    ├── 05-architecture.md             # Architecture patterns
    ├── 06-code-review-checklist.md    # Review checklist
    ├── 07-development-practices.md    # Development workflow
    ├── 08-deployment-patterns.md      # Deployment strategies
    ├── 09-system-evolution.md         # System evolution
    ├── 10-api-design.md               # API design principles
    └── lang-cpp.md                    # C++ specific guidelines
```

## Versioning

This repository uses [Semantic Versioning](https://semver.org/):

- **Major version** (X.0.0): Breaking changes to structure or significant rewrites
- **Minor version** (0.X.0): New guidelines or sections added
- **Patch version** (0.0.X): Clarifications, typo fixes, minor improvements

Current version: **1.0.0**

## License

MIT License - See [LICENSE](LICENSE) file for details.

Copyright (c) 2025 Cortexa LLC

## Related Projects

- **[Sourcerer](https://github.com/Cortexa-LLC/sourcerer)** - A project following these standards

## Support

For questions or discussions:

- Open an [issue](https://github.com/Cortexa-LLC/clean-code-standards/issues)
- Start a [discussion](https://github.com/Cortexa-LLC/clean-code-standards/discussions)

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history.

---

**Made with ❤️ by Cortexa LLC**

*Building better software through better standards*
