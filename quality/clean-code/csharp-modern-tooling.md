# Modern C# Code Quality Tooling (2026)

> **MANDATORY:** All C# projects MUST use the modern .NET tooling stack for code quality enforcement.

This document describes the modern, actively-maintained tooling stack for C# code quality in 2026.

---

## Table of Contents

1. [Overview](#overview)
2. [Why NOT StyleCop.Analyzers](#why-not-stylecop-analyzers)
3. [Modern Tooling Stack](#modern-tooling-stack)
4. [CSharpier - Automated Formatting](#csharpier---automated-formatting)
5. [.NET Analyzers - Built-in Quality](#net-analyzers---built-in-quality)
6. [Roslynator - Comprehensive Analysis](#roslynator---comprehensive-analysis)
7. [EditorConfig - IDE Preferences](#editorconfig---ide-preferences)
8. [Project Configuration](#project-configuration)
9. [Build Enforcement](#build-enforcement)
10. [Pre-Commit Workflow](#pre-commit-workflow)

---

## Overview

**Modern C# Tooling Philosophy:**
- Use **Microsoft-supported** tools as the foundation
- Add **actively-maintained** community tools for enhanced coverage
- Prefer **opinionated formatters** over configurable ones (zero debates)
- Enforce at **build time** to prevent violations

**The 2026 Stack:**
```
CSharpier (formatting)
  + .NET Analyzers (built-in code quality)
  + Roslynator (comprehensive analysis)
  + EditorConfig (IDE preferences)
= Professional C# Code Quality
```

---

## Why NOT StyleCop.Analyzers

### StyleCop.Analyzers is Obsolete (2026)

**Problems:**
- ❌ **Last stable release:** 2018 (8 years old)
- ❌ **Beta version:** Stuck at 1.2.0-beta.435 since 2016
- ❌ **Community-maintained:** Slow updates, not Microsoft-supported
- ❌ **Superseded:** Microsoft built equivalent functionality into .NET SDK
- ❌ **Outdated conventions:** Doesn't reflect modern C# patterns

**What replaced it:**
- ✅ **.NET Analyzers** (IDE* and CA* rules) - Built into .NET 5.0+ SDK
- ✅ **CSharpier** - Modern opinionated formatter
- ✅ **Roslynator** - Actively maintained (2023+), 500+ modern rules

**References:**
- [Microsoft: Code Analysis FAQ](https://learn.microsoft.com/en-us/visualstudio/code-quality/analyzers-faq?view=vs-2022)
- [C# Code Style by EditorConfig in .NET 5+ SDK](https://developers.mews.com/c-code-style-by-editorconfig-in-net-5-sdk-and-beyond/)
- [dotnet format or CSharpier?](http://coffeethinkcode.com/2023/05/05/dotnet-format-or-csharpier/)

---

## Modern Tooling Stack

### Stack Components

**1. CSharpier (Formatting - Opinionated)**
- **Purpose:** Automatic code formatting (like Prettier for JavaScript)
- **Status:** Actively maintained (2024+)
- **Philosophy:** Zero configuration, consistent formatting everywhere
- **Website:** https://csharpier.com/

**2. .NET Analyzers (Quality - Microsoft)**
- **Purpose:** Code quality and correctness (IDE* and CA* rules)
- **Status:** Built into .NET SDK 5.0+
- **Coverage:** Style (IDE*), Code Analysis (CA*), Design Guidelines
- **Docs:** https://learn.microsoft.com/en-us/dotnet/fundamentals/code-analysis/overview

**3. Roslynator (Comprehensive - Community)**
- **Purpose:** 500+ additional analyzers, refactorings, and fixes
- **Status:** Actively maintained (updated regularly)
- **Coverage:** Comprehensive modern C# patterns
- **GitHub:** https://github.com/dotnet/roslynator

**4. EditorConfig (Configuration)**
- **Purpose:** IDE preferences and .NET Analyzer severity configuration
- **Status:** Standard across all .NET IDEs
- **Scope:** Editor behavior + analyzer rule severity

---

## CSharpier - Automated Formatting

### What is CSharpier?

CSharpier is an **opinionated code formatter** for C#. It enforces a consistent style by parsing your code and re-printing it with its own rules.

**Philosophy:** Stop debating code style. Format automatically.

### Installation

**Global Tool (Recommended):**
```bash
dotnet tool install -g csharpier
```

**Local Tool (Per-Project):**
```bash
dotnet new tool-manifest
dotnet tool install csharpier
```

**NuGet Package (Build Integration):**
```xml
<ItemGroup>
  <PackageReference Include="CSharpier.MSBuild" Version="0.27.0">
    <PrivateAssets>all</PrivateAssets>
    <IncludeAssets>runtime; build; native; contentfiles; analyzers</IncludeAssets>
  </PackageReference>
</ItemGroup>
```

### Configuration

CSharpier requires **minimal configuration**. Create `.csharpierrc.json`:

```json
{
  "printWidth": 120,
  "useTabs": false,
  "tabWidth": 4,
  "endOfLine": "lf"
}
```

**That's it.** No debates about brace style, spacing, or line breaks.

### Usage

**Format entire solution:**
```bash
dotnet csharpier .
```

**Check formatting (CI/CD):**
```bash
dotnet csharpier . --check
```

**Format specific files:**
```bash
dotnet csharpier src/MyFile.cs
```

### What CSharpier Handles

CSharpier automatically enforces:
- ✅ Indentation (4 spaces)
- ✅ Line breaks and wrapping
- ✅ Brace placement (Allman style for C#)
- ✅ Spacing around operators
- ✅ Trailing commas in collections
- ✅ Consistent property/method formatting
- ✅ Array and object initializer formatting

**You never debate formatting again.**

### IDE Integration

**Visual Studio:**
- Install "CSharpier" extension
- Format on save: Tools → Options → CSharpier

**Visual Studio Code:**
- Install "CSharpier - Code formatter" extension
- Set as default formatter in settings.json

**JetBrains Rider:**
- Install "CSharpier" plugin
- Configure in Settings → Tools → CSharpier

---

## .NET Analyzers - Built-in Quality

### What are .NET Analyzers?

Built-in code analyzers that ship with the .NET SDK (5.0+). No external packages required.

**Rule Prefixes:**
- **IDE*** - Code style rules (formatting, naming, preferences)
- **CA*** - Code analysis rules (quality, design, performance, security)

### Enable in Project File

```xml
<PropertyGroup>
  <!-- Enable all analyzers -->
  <EnableNETAnalyzers>true</EnableNETAnalyzers>

  <!-- Analysis mode -->
  <AnalysisMode>AllEnabledByDefault</AnalysisMode>

  <!-- Treat warnings as errors in Release -->
  <TreatWarningsAsErrors Condition="'$(Configuration)' == 'Release'">true</TreatWarningsAsErrors>

  <!-- Enforce code style in build -->
  <EnforceCodeStyleInBuild>true</EnforceCodeStyleInBuild>
</PropertyGroup>
```

### Key Rule Categories

**IDE Rules (Code Style):**
```
IDE0001-IDE9999: Style preferences
- Naming conventions
- Use of 'var' vs explicit types
- Expression preferences (pattern matching, etc.)
- Code simplification suggestions
```

**CA Rules (Code Analysis):**
```
CA1000-CA1999: Design rules (inheritance, interfaces, naming)
CA2000-CA2999: Reliability rules (disposal, threading)
CA3000-CA3999: Security rules (injection, crypto)
CA5000-CA5999: Security rules (data flow analysis)
```

### Configuration via EditorConfig

Configure rule severity in `.editorconfig`:

```ini
[*.cs]

# Naming conventions
dotnet_naming_rule.async_methods_end_in_async.severity = warning
dotnet_naming_rule.interfaces_start_with_i.severity = warning

# Code style
dotnet_diagnostic.IDE0001.severity = warning
dotnet_diagnostic.IDE0005.severity = warning  # Remove unnecessary usings

# Code quality
dotnet_diagnostic.CA1001.severity = warning   # Types that own disposable fields
dotnet_diagnostic.CA1031.severity = suggestion # Do not catch general exceptions
dotnet_diagnostic.CA1716.severity = warning   # Identifiers should not match keywords
```

### Enforcement

**Build-time:**
```bash
dotnet build /warnaserror
```

**Continuous:**
```bash
dotnet build --no-incremental
```

---

## Roslynator - Comprehensive Analysis

### What is Roslynator?

A collection of **500+ analyzers**, refactorings, and code fixes for C#. Actively maintained, modern patterns.

**GitHub:** https://github.com/dotnet/roslynator
**Documentation:** https://josefpihrt.github.io/docs/roslynator/

### Installation

Add to your project file:

```xml
<ItemGroup>
  <PackageReference Include="Roslynator.Analyzers" Version="4.12.0">
    <PrivateAssets>all</PrivateAssets>
    <IncludeAssets>runtime; build; native; contentfiles; analyzers</IncludeAssets>
  </PackageReference>
</ItemGroup>
```

### Key Analyzer Categories

**Roslynator covers:**

1. **Simplification (RCS1*)**
   - Simplify boolean expressions
   - Remove redundant code
   - Use pattern matching
   - Inline variables

2. **Readability (RCS2*)**
   - Improve code clarity
   - Consistent formatting
   - Naming improvements

3. **Performance (RCS3*)**
   - Avoid allocations
   - Use efficient patterns
   - Optimize LINQ

4. **Design (RCS4*)**
   - SOLID principles
   - API design
   - Type design

5. **Maintainability (RCS5*)**
   - Reduce complexity
   - Extract methods
   - Simplify conditionals

### Example Rules

**RCS1033: Remove redundant boolean literal**
```csharp
// ❌ Before
if (condition == true)

// ✅ After
if (condition)
```

**RCS1036: Remove unnecessary blank line**
```csharp
// ❌ Before
public class Foo
{

    public void Bar()

// ✅ After (CSharpier handles this automatically)
```

**RCS1080: Use 'Count' property instead of 'Any' method**
```csharp
// ❌ Before
if (list.Any())

// ✅ After
if (list.Count > 0)
```

**RCS1179: Use return instead of assignment**
```csharp
// ❌ Before
bool result;
if (condition)
    result = true;
else
    result = false;
return result;

// ✅ After
return condition;
```

### Configuration

Configure rule severity in `.editorconfig`:

```ini
[*.cs]

# Roslynator rules
dotnet_diagnostic.RCS1033.severity = warning
dotnet_diagnostic.RCS1036.severity = silent  # CSharpier handles this
dotnet_diagnostic.RCS1080.severity = suggestion
dotnet_diagnostic.RCS1179.severity = warning
```

### Disable Rules

If a Roslynator rule conflicts with CSharpier:

```ini
# Disable formatting-related rules (CSharpier handles these)
dotnet_diagnostic.RCS1001.severity = none
dotnet_diagnostic.RCS1003.severity = none
dotnet_diagnostic.RCS1036.severity = none
```

---

## EditorConfig - IDE Preferences

### What is EditorConfig?

A configuration file (`.editorconfig`) that defines coding styles and preferences across editors.

**Purpose:**
- Configure .NET Analyzer rule severity
- Set IDE preferences
- Define naming conventions
- Ensure consistency across team

### Full .editorconfig Example

```ini
root = true

# All files
[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true

# C# files
[*.cs]
indent_style = space
indent_size = 4
tab_width = 4

# .NET formatting options
dotnet_sort_system_directives_first = true
dotnet_separate_import_directive_groups = false

# Code style
csharp_prefer_braces = true:warning
csharp_prefer_simple_using_statement = true:suggestion
csharp_style_namespace_declarations = file_scoped:warning
csharp_style_var_for_built_in_types = true:suggestion
csharp_style_var_when_type_is_apparent = true:suggestion

# Naming conventions
dotnet_naming_rule.interfaces_start_with_i.severity = warning
dotnet_naming_rule.interfaces_start_with_i.symbols = interface_symbols
dotnet_naming_rule.interfaces_start_with_i.style = interface_style

dotnet_naming_symbols.interface_symbols.applicable_kinds = interface
dotnet_naming_style.interface_style.capitalization = pascal_case
dotnet_naming_style.interface_style.required_prefix = I

# .NET Analyzer severity
dotnet_diagnostic.CA1001.severity = warning
dotnet_diagnostic.CA1031.severity = suggestion
dotnet_diagnostic.IDE0001.severity = warning
dotnet_diagnostic.IDE0005.severity = warning

# Roslynator severity
dotnet_diagnostic.RCS1033.severity = warning
dotnet_diagnostic.RCS1080.severity = suggestion

# Disable formatting rules (CSharpier handles these)
dotnet_diagnostic.IDE0055.severity = none
dotnet_diagnostic.RCS1036.severity = none
```

---

## Project Configuration

### Complete .csproj Setup

```xml
<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
    <Nullable>enable</Nullable>
    <ImplicitUsings>enable</ImplicitUsings>

    <!-- .NET Analyzers -->
    <EnableNETAnalyzers>true</EnableNETAnalyzers>
    <AnalysisMode>AllEnabledByDefault</AnalysisMode>
    <EnforceCodeStyleInBuild>true</EnforceCodeStyleInBuild>

    <!-- Treat warnings as errors in Release -->
    <TreatWarningsAsErrors Condition="'$(Configuration)' == 'Release'">true</TreatWarningsAsErrors>
  </PropertyGroup>

  <ItemGroup>
    <!-- CSharpier -->
    <PackageReference Include="CSharpier.MSBuild" Version="0.27.0">
      <PrivateAssets>all</PrivateAssets>
      <IncludeAssets>runtime; build; native; contentfiles; analyzers</IncludeAssets>
    </PackageReference>

    <!-- Roslynator -->
    <PackageReference Include="Roslynator.Analyzers" Version="4.12.0">
      <PrivateAssets>all</PrivateAssets>
      <IncludeAssets>runtime; build; native; contentfiles; analyzers</IncludeAssets>
    </PackageReference>
  </ItemGroup>

  <ItemGroup>
    <!-- Configuration files -->
    <AdditionalFiles Include=".editorconfig" />
  </ItemGroup>

</Project>
```

### Required Configuration Files

**Project Root Structure:**
```
MyProject/
├── .editorconfig           # .NET Analyzer + Roslynator configuration
├── .csharpierrc.json       # CSharpier configuration
├── MyProject.sln
└── src/
    └── MyProject.csproj
```

**.csharpierrc.json:**
```json
{
  "printWidth": 120,
  "useTabs": false,
  "tabWidth": 4,
  "endOfLine": "lf"
}
```

**.editorconfig:** See full example above

---

## Build Enforcement

### Build Commands

**Local Development:**
```bash
# Format code
dotnet csharpier .

# Build with analyzer checks
dotnet build

# Build treating warnings as errors
dotnet build /warnaserror
```

**CI/CD Pipeline:**
```bash
# Check formatting (don't modify)
dotnet csharpier . --check

# Build with strict enforcement
dotnet build --configuration Release /warnaserror
```

### GitHub Actions Example

```yaml
name: Code Quality

on: [push, pull_request]

jobs:
  quality:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup .NET
        uses: actions/setup-dotnet@v4
        with:
          dotnet-version: '8.0.x'

      - name: Restore dependencies
        run: dotnet restore

      - name: Check formatting
        run: dotnet csharpier . --check

      - name: Build with analyzers
        run: dotnet build --configuration Release /warnaserror

      - name: Run tests
        run: dotnet test --no-build --configuration Release
```

### Pre-Commit Hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash

echo "Running CSharpier..."
dotnet csharpier .

if [ $? -ne 0 ]; then
    echo "❌ Formatting failed. Commit aborted."
    exit 1
fi

echo "Running build with analyzers..."
dotnet build /warnaserror

if [ $? -ne 0 ]; then
    echo "❌ Build failed due to analyzer violations. Commit aborted."
    exit 1
fi

echo "✅ All checks passed."
exit 0
```

Make it executable:
```bash
chmod +x .git/hooks/pre-commit
```

---

## Pre-Commit Workflow

### Engineer Workflow

**Before Committing C# Code:**

```
STEP 1: Format code automatically
  $ dotnet csharpier .

STEP 2: Build with analyzer enforcement
  $ dotnet build /warnaserror

STEP 3: Run tests
  $ dotnet test

STEP 4: Commit if all pass
  $ git add .
  $ git commit -m "Your message"
```

**If build fails:**
```
IF analyzer violations found THEN
  1. Review the warnings/errors
  2. Fix the issues (use IDE quick fixes)
  3. Re-run: dotnet build /warnaserror
  4. Repeat until zero violations
END IF
```

### Reviewer Workflow

**Reviewing C# Code:**

```
STEP 1: Verify formatting
  $ dotnet csharpier . --check
  Expected: "All files are formatted correctly."

STEP 2: Verify build passes
  $ dotnet build /warnaserror
  Expected: "Build succeeded. 0 Warning(s) 0 Error(s)"

STEP 3: Review code quality
  - Check for logical issues
  - Verify tests comprehensive
  - Assess architecture decisions

STEP 4: Verdict
  IF formatting check fails THEN
    REJECT: "Code not formatted. Run: dotnet csharpier ."
  ELSE IF build has violations THEN
    REJECT: "Analyzer violations present. Fix and resubmit."
  ELSE IF code quality issues THEN
    REQUEST CHANGES: [List issues]
  ELSE
    APPROVE
  END IF
```

---

## Comparison: Old vs New

### StyleCop.Analyzers (Old - 2018)

```xml
<!-- ❌ OLD WAY (Don't use) -->
<ItemGroup>
  <PackageReference Include="StyleCop.Analyzers" Version="1.2.0-beta.435" />
  <AdditionalFiles Include="stylecop.json" />
</ItemGroup>
```

**Problems:**
- Last updated 2018
- Beta version stuck since 2016
- Manual configuration required
- Outdated conventions
- Not Microsoft-supported

### Modern Stack (New - 2026)

```xml
<!-- ✅ NEW WAY (Use this) -->
<PropertyGroup>
  <EnableNETAnalyzers>true</EnableNETAnalyzers>
  <AnalysisMode>AllEnabledByDefault</AnalysisMode>
  <EnforceCodeStyleInBuild>true</EnforceCodeStyleInBuild>
</PropertyGroup>

<ItemGroup>
  <PackageReference Include="CSharpier.MSBuild" Version="0.27.0" />
  <PackageReference Include="Roslynator.Analyzers" Version="4.12.0" />
</ItemGroup>
```

**Benefits:**
- ✅ Actively maintained (2024+)
- ✅ Microsoft-supported foundation
- ✅ Zero-config formatting
- ✅ 500+ modern analyzers
- ✅ Build-time enforcement
- ✅ IDE integration everywhere

---

## Summary

### MANDATORY Requirements

All C# projects MUST:

1. ✅ **Use CSharpier** for automatic formatting
   - Format before every commit: `dotnet csharpier .`

2. ✅ **Enable .NET Analyzers** in project file
   - `EnableNETAnalyzers=true`
   - `EnforceCodeStyleInBuild=true`

3. ✅ **Install Roslynator** for comprehensive analysis
   - Add Roslynator.Analyzers NuGet package

4. ✅ **Configure EditorConfig** for rule severity
   - Create `.editorconfig` with team standards

5. ✅ **Enforce at build time**
   - CI/CD must run: `dotnet build /warnaserror`
   - Local commits: pre-commit hook

### Zero Tolerance

**Build MUST pass with:**
- ✅ Zero formatting violations (CSharpier)
- ✅ Zero .NET Analyzer warnings
- ✅ Zero Roslynator warnings (configured severity)
- ✅ All tests passing

**Reviewers MUST block if:**
- ❌ Formatting check fails
- ❌ Build has analyzer violations
- ❌ Code quality issues present

---

## References

**Official Documentation:**
- [Microsoft: .NET Code Analysis](https://learn.microsoft.com/en-us/dotnet/fundamentals/code-analysis/overview)
- [Microsoft: Code Quality Analyzers FAQ](https://learn.microsoft.com/en-us/visualstudio/code-quality/analyzers-faq?view=vs-2022)
- [CSharpier Official Site](https://csharpier.com/)
- [Roslynator GitHub](https://github.com/dotnet/roslynator)
- [Roslynator Documentation](https://josefpihrt.github.io/docs/roslynator/)

**Articles:**
- [C# Code Style by EditorConfig in .NET 5+ SDK](https://developers.mews.com/c-code-style-by-editorconfig-in-net-5-sdk-and-beyond/)
- [dotnet format or CSharpier?](http://coffeethinkcode.com/2023/05/05/dotnet-format-or-csharpier/)
- [Roslynator Analyzers 2.3.1](https://www.infoq.com/news/2020/01/roslynator-analyzers-231/)

---

**Last Updated:** 2026-01-09
**Status:** ACTIVE - Modern tooling standard for all C# projects
