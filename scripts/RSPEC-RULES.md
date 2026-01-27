# SonarQube RSpec Rule Extraction

Complete guide for extracting and using SonarQube code quality rules from the rspec repository.

## Table of Contents
- [Overview](#overview)
- [Quick Start](#quick-start)
- [Scripts](#scripts)
- [File Formats](#file-formats)
- [CI/CD Integration](#cicd-integration)
- [Adding New Languages](#adding-new-languages)
- [Rule Interpretation](#rule-interpretation)
- [Examples](#examples)

## Overview

These scripts extract code quality rules from SonarQube's [rspec repository](https://github.com/SonarSource/rspec) and generate language-specific rule sets in multiple formats for integration with CI/CD pipelines, static analysis tools, and documentation.

**Total Rules by Language:**
- C++: 916 rules
- C#: 598 rules
- Go: 108 rules
- Python: 493 rules
- JavaScript/TypeScript: 677 rules
- Java: 986 rules
- Kotlin: 187 rules
- Swift: 166 rules

**Total: 4,131 rules**

## Quick Start

### Generate All Language Rule Sets

```bash
# Clone rspec repository (one-time setup)
cd ~/Projects/Vibe
git clone https://github.com/SonarSource/rspec.git

# Generate rules for all languages
cd /path/to/ai-pack
./scripts/generate_all_language_rules.sh ~/Projects/Vibe/rspec ./quality/sonarqube-rules
```

### Extract Rules for a Single Language

```bash
# CSV format (for CI/CD)
python3 scripts/rspec_rule_extractor.py ~/Projects/Vibe/rspec go --format csv > go_rules.csv

# JSON format (for programmatic use)
python3 scripts/rspec_rule_extractor.py ~/Projects/Vibe/rspec python --format json > python_rules.json
```

## Scripts

### rspec_rule_extractor.py

Extract rules for a specific language from the rspec repository.

**Parameters:**
- `rspec_root` - Path to rspec repository
- `language` - Language code (go, python, java, cpp, etc.)
- `--format` - Output format: csv, json, or markdown
- `--output` - Output file (default: stdout)

**Examples:**
```bash
# Basic usage
python3 rspec_rule_extractor.py ~/Projects/Vibe/rspec go --format csv

# Save to file
python3 rspec_rule_extractor.py ~/Projects/Vibe/rspec python \
  --format json \
  --output python_rules.json

# Use language aliases
python3 rspec_rule_extractor.py ~/Projects/Vibe/rspec cpp --format csv  # C++
python3 rspec_rule_extractor.py ~/Projects/Vibe/rspec js --format csv   # JavaScript
```

**Language Code Mapping:**
- `cpp`, `c++`, `c` → cfamily
- `csharp`, `c#` → csharp
- `go` → go
- `python`, `py` → python
- `javascript`, `js`, `typescript`, `ts` → javascript
- `java` → java
- `kotlin`, `kt` → kotlin
- `swift` → swift
- `ruby`, `rb` → ruby
- `php` → php
- `scala` → scala
- `shell`, `bash`, `sh` → shell

### generate_all_language_rules.sh

Generate rule sets for all supported languages.

**Parameters:**
- `$1` - Path to rspec repository (default: `~/Projects/Vibe/rspec`)
- `$2` - Output directory (default: `./quality/sonarqube-rules`)

**Output Structure:**
```
quality/sonarqube-rules/
├── README.md                    # Index with statistics
├── cpp/
│   ├── rules.csv               # Machine-readable (CI/CD)
│   ├── rules.json              # Programmatic use
│   └── rules.md                # Documentation
├── go/
│   ├── rules.csv
│   ├── rules.json
│   └── rules.md
└── ... (8 languages total)
```

## File Formats

### CSV Format (Primary - CI/CD)

**Use Cases:**
- Shell scripts and command-line tools
- CI/CD pipeline integration
- Grep/awk/cut filtering
- Spreadsheet import

**Columns:**
1. `rule_id` - Rule identifier (e.g., S1192)
2. `title` - Human-readable description
3. `type` - CODE_SMELL, BUG, VULNERABILITY, SECURITY_HOTSPOT
4. `severity` - BLOCKER, CRITICAL, MAJOR, MINOR, INFO
5. `status` - ready, deprecated, closed
6. `scope` - All, Main, Test
7. `languages` - Semicolon-separated list
8. `tags` - Semicolon-separated list
9. `impacts` - Impact:Level pairs (e.g., "MAINTAINABILITY:HIGH")
10. `remediation_cost` - Estimated fix time
11. `default_quality_profiles` - Semicolon-separated list
12. `language_specific_docs` - Available doc files
13. `metadata_path` - Source metadata.json path

**Example CSV Row:**
```csv
S1192,"String literals should not be duplicated",CODE_SMELL,Critical,ready,All,"go;java;python",design,MAINTAINABILITY:HIGH,"Constant/Issue: 10min","Sonar way",rule.adoc,/path/to/rspec/rules/S1192/metadata.json
```

### JSON Format (Primary - Programmatic)

**Use Cases:**
- Language-specific tooling
- Web dashboards
- Complex rule analysis
- API integration

**Structure:**
```json
[
  {
    "rule_id": "S1192",
    "title": "String literals should not be duplicated",
    "type": "CODE_SMELL",
    "severity": "Critical",
    "status": "ready",
    "scope": "All",
    "languages": ["go", "java", "python"],
    "tags": ["design"],
    "impacts": {
      "MAINTAINABILITY": "HIGH"
    },
    "remediation_cost": "Constant/Issue: 10min",
    "default_quality_profiles": ["Sonar way"],
    "language_specific_docs": ["rule.adoc"],
    "metadata_path": "/path/to/rspec/rules/S1192/metadata.json"
  }
]
```

### Markdown Format (Optional - Documentation)

**Use Cases:**
- Documentation and reference
- GitHub/GitLab browsing
- Code review guidelines

**Structure:**
- Grouped by rule type
- Tables with key information
- Links to rspec documentation

## CI/CD Integration

### Example: Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Get blocker/critical rules for current language
RULES_FILE="quality/sonarqube-rules/go/rules.csv"

if [[ ! -f "$RULES_FILE" ]]; then
    echo "Rules file not found: $RULES_FILE"
    exit 1
fi

# Extract blocker and critical rule IDs
BLOCKER_RULES=$(awk -F',' '$4=="BLOCKER" {print $1}' "$RULES_FILE" | tr '\n' '|')
CRITICAL_RULES=$(awk -F',' '$4=="CRITICAL" {print $1}' "$RULES_FILE" | tr '\n' '|')

# Run your linter and check for violations
# (Implementation depends on your linter)
echo "Checking against ${BLOCKER_RULES}..."
```

### Example: Find Critical Rules by Type

```bash
#!/bin/bash
# find-critical-bugs.sh

RULES_FILE="quality/sonarqube-rules/go/rules.csv"

# Find critical bugs
awk -F',' '
  $3=="BUG" && $4=="CRITICAL" {
    printf "Rule: %s\n", $1
    printf "  Title: %s\n", $2
    printf "  Cost: %s\n\n", $10
  }
' "$RULES_FILE"
```

### Example: Generate Quality Report

```bash
#!/bin/bash
# quality-report.sh

for lang_dir in quality/sonarqube-rules/*/; do
    lang=$(basename "$lang_dir")
    csv="${lang_dir}rules.csv"

    if [[ -f "$csv" ]]; then
        total=$(tail -n +2 "$csv" | wc -l | tr -d ' ')
        blocker=$(awk -F',' '$4=="BLOCKER"' "$csv" | wc -l | tr -d ' ')
        critical=$(awk -F',' '$4=="CRITICAL"' "$csv" | wc -l | tr -d ' ')

        printf "%-15s Total: %4d | Blocker: %3d | Critical: %3d\n" \
            "$lang" "$total" "$blocker" "$critical"
    fi
done
```

### Example: Python Integration

```python
#!/usr/bin/env python3
import csv
import json
from collections import defaultdict

def load_rules(lang):
    """Load rules for a language."""
    with open(f'quality/sonarqube-rules/{lang}/rules.json') as f:
        return json.load(f)

def analyze_rules(rules):
    """Analyze rules by category."""
    by_type = defaultdict(list)
    by_severity = defaultdict(list)

    for rule in rules:
        by_type[rule['type']].append(rule)
        by_severity[rule['severity']].append(rule)

    return by_type, by_severity

def get_critical_bugs(lang):
    """Get all critical bugs for a language."""
    rules = load_rules(lang)
    return [r for r in rules if r['type'] == 'BUG' and r['severity'] == 'Critical']

# Usage
go_rules = load_rules('go')
by_type, by_severity = analyze_rules(go_rules)

print(f"Go Rules Summary:")
print(f"  Total: {len(go_rules)}")
print(f"  Bugs: {len(by_type['BUG'])}")
print(f"  Code Smells: {len(by_type['CODE_SMELL'])}")
print(f"  Vulnerabilities: {len(by_type['VULNERABILITY'])}")
```

### Example: Shell Script Analysis

```bash
#!/bin/bash
# analyze-impact.sh

LANG="$1"
RULES_FILE="quality/sonarqube-rules/${LANG}/rules.csv"

echo "Impact Analysis for ${LANG}:"
echo

# Count by impact level
echo "Maintainability Impact:"
grep "MAINTAINABILITY:HIGH" "$RULES_FILE" | wc -l | xargs printf "  HIGH:   %d\n"
grep "MAINTAINABILITY:MEDIUM" "$RULES_FILE" | wc -l | xargs printf "  MEDIUM: %d\n"
grep "MAINTAINABILITY:LOW" "$RULES_FILE" | wc -l | xargs printf "  LOW:    %d\n"

echo
echo "Reliability Impact:"
grep "RELIABILITY:HIGH" "$RULES_FILE" | wc -l | xargs printf "  HIGH:   %d\n"
grep "RELIABILITY:MEDIUM" "$RULES_FILE" | wc -l | xargs printf "  MEDIUM: %d\n"
grep "RELIABILITY:LOW" "$RULES_FILE" | wc -l | xargs printf "  LOW:    %d\n"

echo
echo "Security Impact:"
grep "SECURITY:HIGH" "$RULES_FILE" | wc -l | xargs printf "  HIGH:   %d\n"
grep "SECURITY:MEDIUM" "$RULES_FILE" | wc -l | xargs printf "  MEDIUM: %d\n"
grep "SECURITY:LOW" "$RULES_FILE" | wc -l | xargs printf "  LOW:    %d\n"
```

## Adding New Languages

To add support for a new language:

1. **Check rspec support:**
   ```bash
   # List supported languages in rspec
   ls -1 ~/Projects/Vibe/rspec/rules/S100/
   ```

2. **Add to `rspec_rule_extractor.py`:**
   ```python
   # In the LANGUAGE_MAP dictionary
   LANGUAGE_MAP = {
       # ... existing mappings ...
       'rust': 'rust',        # If directory name matches
       'dart': 'dart',
   }
   ```

3. **Add to `generate_all_language_rules.sh`:**
   ```bash
   # In the LANGUAGES array
   LANGUAGES=(
       # ... existing languages ...
       "Rust:rust"
       "Dart:dart"
   )
   ```

4. **Test extraction:**
   ```bash
   python3 scripts/rspec_rule_extractor.py ~/Projects/Vibe/rspec rust --format csv
   ```

5. **Regenerate all rules:**
   ```bash
   ./scripts/generate_all_language_rules.sh
   ```

6. **Update documentation:**
   - Add language to this file
   - Update rule counts
   - Add to README.md

## Rule Interpretation

### Rule Types

| Type | Description | Priority |
|------|-------------|----------|
| **CODE_SMELL** | Maintainability issues | Medium |
| **BUG** | Reliability issues | High |
| **VULNERABILITY** | Security vulnerabilities | Critical |
| **SECURITY_HOTSPOT** | Security-sensitive code | Review |

### Severity Levels

| Severity | Description | Action |
|----------|-------------|--------|
| **BLOCKER** | Must fix immediately | Block deployment |
| **CRITICAL** | Fix ASAP | High priority |
| **MAJOR** | Should fix | Medium priority |
| **MINOR** | Consider fixing | Low priority |
| **INFO** | Informational | No action required |

### Impact Categories

| Category | Description |
|----------|-------------|
| **MAINTAINABILITY** | Ease of understanding, modifying, extending |
| **RELIABILITY** | Correctness and predictability |
| **SECURITY** | Resistance to attacks |

### Impact Levels

| Level | Description |
|-------|-------------|
| **HIGH** | Significant impact |
| **MEDIUM** | Moderate impact |
| **LOW** | Minor impact |

## Examples

### Find All String Duplication Rules

```bash
grep -i "string.*duplicate" quality/sonarqube-rules/*/rules.csv
```

### Get Rules with High Security Impact

```bash
awk -F',' '$9 ~ /SECURITY:HIGH/' quality/sonarqube-rules/go/rules.csv
```

### Count Rules by Type for All Languages

```bash
for lang in cpp csharp go python javascript java kotlin swift; do
    echo "$lang:"
    awk -F',' 'NR>1 {print $3}' quality/sonarqube-rules/$lang/rules.csv | sort | uniq -c
    echo
done
```

### Find Quick Wins (Low Remediation Cost)

```bash
# Rules that take ≤5 minutes to fix
grep "5min" quality/sonarqube-rules/go/rules.csv | awk -F',' '{print $1, $2}'
```

### Generate Custom Rule List

```bash
# Create custom rule list for code review
awk -F',' '
  BEGIN {
    print "# Code Review Checklist"
    print ""
  }
  $4=="CRITICAL" && $3=="BUG" {
    printf "- [ ] **%s**: %s\n", $1, $2
  }
' quality/sonarqube-rules/go/rules.csv > code-review-checklist.md
```

## Updating Rules

The rspec repository is updated regularly. To refresh your rule sets:

```bash
# Update rspec repository
cd ~/Projects/Vibe/rspec
git pull

# Regenerate all rule sets
cd /path/to/ai-pack
./scripts/generate_all_language_rules.sh ~/Projects/Vibe/rspec ./quality/sonarqube-rules

# Review changes
git diff quality/sonarqube-rules/

# Commit updates
git add quality/sonarqube-rules/
git commit -m "chore: update SonarQube rule sets from rspec $(date +%Y-%m-%d)"
```

## Troubleshooting

### "No rules found for language"

**Cause:** Language not supported in rspec or incorrect language code

**Solution:**
```bash
# Check available languages for a rule
ls ~/Projects/Vibe/rspec/rules/S100/

# Verify language mapping in rspec_rule_extractor.py
python3 -c "from scripts.rspec_rule_extractor import RSpecExtractor; print(RSpecExtractor.LANGUAGE_MAP)"
```

### "metadata.json parse error"

**Cause:** Corrupted rspec repository or JSON syntax error

**Solution:**
```bash
# Update rspec repository
cd ~/Projects/Vibe/rspec
git fetch origin
git reset --hard origin/master

# Try extraction again
cd /path/to/ai-pack
python3 scripts/rspec_rule_extractor.py ~/Projects/Vibe/rspec go --format csv
```

### "Permission denied"

**Cause:** Scripts not executable

**Solution:**
```bash
chmod +x scripts/rspec_rule_extractor.py
chmod +x scripts/generate_all_language_rules.sh
```

## Performance Notes

- **CSV generation:** ~100-1000 rules/second
- **JSON generation:** ~50-500 rules/second
- **Markdown generation:** ~20-200 rules/second
- **Full generation (8 languages, all formats):** ~30-60 seconds

## References

- [SonarQube Rules Documentation](https://rules.sonarsource.com/)
- [rspec Repository](https://github.com/SonarSource/rspec)
- [SonarSource Analyzers](https://github.com/SonarSource)
- [Our Clean Code Standards](../quality/clean-code/)
- [Sonar Rule Specification](https://github.com/SonarSource/rspec/blob/master/README.adoc)

## License

These scripts are part of the ai-pack project. The SonarQube rspec repository is maintained by SonarSource SA under the LGPL v3 license.

---

**Last Updated:** 2026-01-27
