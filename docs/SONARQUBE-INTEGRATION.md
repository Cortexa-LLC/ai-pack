# SonarQube Integration Guide

Complete guide for using SonarQube Community Edition to validate code with your AI agents.

## Table of Contents
- [Overview](#overview)
- [Quick Start](#quick-start)
- [Architecture](#architecture)
- [Setup](#setup)
- [Agent Integration](#agent-integration)
- [Rule Queries](#rule-queries)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)

## Overview

This integration provides comprehensive code validation using SonarQube Community Edition:

- ✅ **Complete Rule Coverage**: All 4,131+ SonarQube rules across 8 languages
- ✅ **Agent-Friendly**: Simple JSON API for programmatic access
- ✅ **Free**: Community Edition for unlimited private repositories
- ✅ **Cross-Platform**: Python scripts work on Windows, macOS, Linux
- ✅ **Metadata Enrichment**: Violations enhanced with impact, remediation cost, tags

**Tools:**
1. **setup-sonarqube.py** - One-command Docker deployment
2. **validate-with-sonarqube.py** - Agent-friendly code validation
3. **query-rules.py** - Search and explain rules

## Quick Start

### 1. Start SonarQube (One-Time Setup)

```bash
# Start SonarQube Community Edition in Docker
python3 scripts/setup-sonarqube.py

# Wait ~1 minute for startup
# Web UI will be available at: http://localhost:9000
```

### 2. Install sonar-scanner

**macOS:**
```bash
brew install sonar-scanner
```

**Linux:**
```bash
# Download from https://docs.sonarqube.org/latest/analysis/scan/sonarscanner/
unzip sonar-scanner-cli-*.zip
sudo mv sonar-scanner-*/ /opt/sonar-scanner
export PATH="/opt/sonar-scanner/bin:$PATH"
```

**Windows:**
```powershell
# Download from https://docs.sonarqube.org/latest/analysis/scan/sonarscanner/
# Extract and add to PATH
```

### 3. Validate Code

```bash
# Validate a file
python3 scripts/validate-with-sonarqube.py src/server.go

# Get JSON output for agents
python3 scripts/validate-with-sonarqube.py src/server.go --format json
```

## Architecture

```
┌─────────────────────┐
│   Your Agent        │
│  (a2a-agent, etc.)  │
└──────────┬──────────┘
           │
           │ calls
           ▼
┌─────────────────────────────────┐
│  validate-with-sonarqube.py     │
│  ┌─────────────────────────┐   │
│  │ 1. Run sonar-scanner    │   │
│  │ 2. Fetch violations API │   │
│  │ 3. Enrich with metadata │   │
│  │ 4. Return JSON          │   │
│  └─────────────────────────┘   │
└──────────┬────────────┬─────────┘
           │            │
           ▼            ▼
┌──────────────────┐  ┌─────────────────────┐
│  SonarQube CE    │  │ Extracted Rules     │
│  (validates)     │  │ (metadata)          │
│                  │  │                     │
│  - All 4,131+    │  │  - Rule titles      │
│    rules         │  │  - Impacts          │
│  - Returns       │  │  - Remediation cost │
│    violations    │  │  - Tags             │
└──────────────────┘  └─────────────────────┘
```

## Setup

### Prerequisites

- **Docker Desktop** - https://www.docker.com/products/docker-desktop
- **Python 3.7+** - Comes with macOS/Linux, download for Windows
- **sonar-scanner** - See Quick Start above

### Step-by-Step Setup

**1. Clone/Update ai-pack:**
```bash
cd /path/to/ai-pack
git pull  # Get latest scripts
```

**2. Start SonarQube:**
```bash
python3 scripts/setup-sonarqube.py
```

This script:
- Starts SonarQube + PostgreSQL in Docker
- Waits for startup (~1 minute)
- Changes default admin password
- Generates API token
- Saves config to `.sonarqube-config`

**3. Verify Setup:**
```bash
# Check web UI
open http://localhost:9000

# Login: admin / admin123

# Test validation
python3 scripts/validate-with-sonarqube.py a2a-agent/hello.py
```

### Configuration File

The setup script creates `.sonarqube-config`:

```bash
# SonarQube Configuration for AI-Pack
# Generated: 2026-01-27 10:30:00

SONARQUBE_URL=http://localhost:9000
SONARQUBE_TOKEN=squ_1234567890abcdef
```

**Keep this file secure!** Add to `.gitignore`.

## Agent Integration

### Basic Validation

```python
import subprocess
import json

def validate_code(file_path):
    """Validate code file using SonarQube."""
    result = subprocess.run(
        ["python3", "scripts/validate-with-sonarqube.py", file_path, "--format", "json"],
        capture_output=True,
        text=True
    )

    if result.returncode != 0:
        return json.loads(result.stdout)

    return {"success": True, "violations": []}
```

### JSON Output Format

```json
{
  "success": true,
  "source": "src/server.go",
  "project_key": "aipack_src_server_go",
  "language": "go",
  "violations": [
    {
      "rule_id": "S1192",
      "rule_key": "go:S1192",
      "file": "server.go",
      "line": 42,
      "message": "Define a constant instead of duplicating this literal \"text\" 3 times.",
      "severity": "CRITICAL",
      "type": "CODE_SMELL",
      "status": "OPEN",
      "rule_title": "String literals should not be duplicated",
      "impacts": "MAINTAINABILITY:HIGH",
      "remediation_cost": "Constant/Issue: 10min",
      "tags": ["design"],
      "scope": "All"
    }
  ],
  "summary": {
    "total": 1,
    "by_severity": {
      "CRITICAL": 1
    },
    "by_type": {
      "CODE_SMELL": 1
    }
  }
}
```

### Filtering by Severity

```bash
# Only check BLOCKER and CRITICAL
python3 scripts/validate-with-sonarqube.py src/ \
  --severity BLOCKER,CRITICAL \
  --format json
```

### Example: Agent Pre-Commit Hook

```python
#!/usr/bin/env python3
"""Agent task: Validate code before commit."""

import subprocess
import json
import sys

def validate_changed_files():
    """Validate all changed files."""
    # Get changed files
    result = subprocess.run(
        ["git", "diff", "--name-only", "--cached"],
        capture_output=True,
        text=True
    )

    files = result.stdout.strip().split("\n")
    code_files = [f for f in files if f.endswith((".go", ".py", ".java", ".js"))]

    violations_found = False

    for file in code_files:
        print(f"Validating {file}...")

        # Run validation
        result = subprocess.run(
            ["python3", "scripts/validate-with-sonarqube.py", file,
             "--severity", "BLOCKER,CRITICAL",
             "--format", "json"],
            capture_output=True,
            text=True
        )

        data = json.loads(result.stdout)

        if data.get("violations"):
            violations_found = True
            print(f"\n❌ {file}: {len(data['violations'])} critical violations")

            for v in data["violations"]:
                print(f"  Line {v['line']}: [{v['severity']}] {v['message']}")

    if violations_found:
        print("\n🚫 Commit blocked due to critical violations")
        sys.exit(1)
    else:
        print("\n✅ All files pass validation")

if __name__ == "__main__":
    validate_changed_files()
```

## Rule Queries

### Search Rules

```bash
# List all Go rules
python3 scripts/query-rules.py --language go

# Find critical bugs
python3 scripts/query-rules.py --language python --type BUG --severity CRITICAL

# Search by keyword
python3 scripts/query-rules.py --language java --search "string literal"
```

### Explain a Rule

```bash
# Get full details for S1192
python3 scripts/query-rules.py --rule S1192
```

Output:
```
======================================================================
Rule: S1192
======================================================================

Title: String literals should not be duplicated
Type: CODE_SMELL
Severity: Critical
Status: ready
Scope: All

Impacts: MAINTAINABILITY:HIGH
Tags: design
Remediation Cost: Constant/Issue: 10min
Quality Profiles: Sonar way
Languages: go;java;python;javascript;...

Metadata: /path/to/rspec/rules/S1192/metadata.json
```

### Filter by Impact

```bash
# Find all high-security-impact rules
python3 scripts/query-rules.py --language go --impact SECURITY:HIGH

# Get JSON for programmatic use
python3 scripts/query-rules.py --language go --impact SECURITY:HIGH --format json
```

### Export Rules

```bash
# Export as JSON
python3 scripts/query-rules.py --language go --format json > go_rules.json

# Export as Markdown
python3 scripts/query-rules.py --language go --format markdown > go_rules.md

# Summary statistics
python3 scripts/query-rules.py --language go --format summary
```

## Examples

### Example 1: Agent Code Review

```python
#!/usr/bin/env python3
"""Agent task: Review code changes."""

import subprocess
import json

def review_code_quality(file_path):
    """Review code quality and return feedback."""
    # Run validation
    result = subprocess.run(
        ["python3", "scripts/validate-with-sonarqube.py", file_path,
         "--format", "json"],
        capture_output=True,
        text=True
    )

    data = json.loads(result.stdout)

    if not data.get("violations"):
        return "✅ Code looks good! No violations found."

    # Generate review feedback
    feedback = []
    feedback.append(f"📋 Code Review for {file_path}\n")
    feedback.append(f"Found {data['summary']['total']} issues:\n")

    # Group by severity
    for severity in ["BLOCKER", "CRITICAL", "MAJOR", "MINOR"]:
        count = data['summary']['by_severity'].get(severity, 0)
        if count > 0:
            feedback.append(f"  • {severity}: {count}")

    feedback.append("\n🔍 Top Issues:\n")

    # Show top 5 violations
    for i, v in enumerate(data['violations'][:5], 1):
        feedback.append(f"{i}. [{v['severity']}] Line {v['line']}")
        feedback.append(f"   {v['message']}")
        if v.get('remediation_cost'):
            feedback.append(f"   Fix time: {v['remediation_cost']}")
        feedback.append("")

    return "\n".join(feedback)

# Usage
print(review_code_quality("src/server.go"))
```

### Example 2: Quality Gate Check

```python
#!/usr/bin/env python3
"""Check if code meets quality gate."""

import subprocess
import json
import sys

def check_quality_gate(source_path):
    """Check if code passes quality gate."""
    # Run validation
    result = subprocess.run(
        ["python3", "scripts/validate-with-sonarqube.py", source_path,
         "--format", "json"],
        capture_output=True,
        text=True
    )

    data = json.loads(result.stdout)

    # Quality gate rules:
    # - No BLOCKER violations
    # - No more than 5 CRITICAL violations
    # - No more than 10 MAJOR violations

    blockers = data['summary']['by_severity'].get('BLOCKER', 0)
    criticals = data['summary']['by_severity'].get('CRITICAL', 0)
    majors = data['summary']['by_severity'].get('MAJOR', 0)

    failed = []

    if blockers > 0:
        failed.append(f"❌ {blockers} BLOCKER violations (max: 0)")

    if criticals > 5:
        failed.append(f"❌ {criticals} CRITICAL violations (max: 5)")

    if majors > 10:
        failed.append(f"❌ {majors} MAJOR violations (max: 10)")

    if failed:
        print("🚫 Quality Gate: FAILED\n")
        for f in failed:
            print(f)
        sys.exit(1)
    else:
        print("✅ Quality Gate: PASSED")

# Usage
check_quality_gate("src/")
```

### Example 3: Generate Fix Suggestions

```python
#!/usr/bin/env python3
"""Generate fix suggestions based on rules."""

import subprocess
import json

def generate_fix_suggestions(file_path):
    """Generate fix suggestions for violations."""
    # Validate code
    result = subprocess.run(
        ["python3", "scripts/validate-with-sonarqube.py", file_path,
         "--format", "json"],
        capture_output=True,
        text=True
    )

    data = json.loads(result.stdout)

    if not data.get("violations"):
        return "No violations found."

    suggestions = []
    suggestions.append(f"🔧 Fix Suggestions for {file_path}\n")

    for v in data['violations']:
        suggestions.append(f"Line {v['line']}: {v['rule_id']}")
        suggestions.append(f"  Issue: {v['message']}")

        # Add fix suggestion based on rule
        if v['rule_id'] == 'S1192':
            suggestions.append(f"  💡 Fix: Extract the duplicated string to a constant")
        elif v['rule_id'] == 'S1066':
            suggestions.append(f"  💡 Fix: Combine the if statements into a single condition")
        elif 'error' in v['message'].lower():
            suggestions.append(f"  💡 Fix: Add proper error handling")

        if v.get('remediation_cost'):
            suggestions.append(f"  ⏱️  Estimated time: {v['remediation_cost']}")

        suggestions.append("")

    return "\n".join(suggestions)

# Usage
print(generate_fix_suggestions("src/server.go"))
```

## Troubleshooting

### SonarQube won't start

**Symptom:** Docker container exits immediately

**Solution:**
```bash
# Check logs
docker logs sonarqube

# Common issue: Not enough memory
# Docker Desktop → Settings → Resources → Memory: 4GB minimum

# Restart
docker compose -f docker-compose.sonarqube.yml down
docker compose -f docker-compose.sonarqube.yml up -d
```

### sonar-scanner not found

**Symptom:** `sonar-scanner: command not found`

**Solution:**
```bash
# macOS
brew install sonar-scanner

# Verify
sonar-scanner -v
```

### Violations not found

**Symptom:** Validation succeeds but returns no violations

**Possible causes:**
1. **Project not indexed yet** - Wait 10 seconds, try again
2. **Wrong project key** - Check SonarQube UI
3. **File not analyzed** - Check sonar-scanner output

**Debug:**
```bash
# Run with verbose output (remove --format json)
python3 scripts/validate-with-sonarqube.py file.go

# Check SonarQube UI
open http://localhost:9000
```

### Token invalid

**Symptom:** `401 Unauthorized`

**Solution:**
```bash
# Regenerate token
python3 scripts/setup-sonarqube.py

# Or manually create token:
# 1. Login to http://localhost:9000
# 2. My Account → Security → Generate Tokens
# 3. Update .sonarqube-config
```

### Port 9000 already in use

**Symptom:** `Bind for 0.0.0.0:9000 failed: port is already allocated`

**Solution:**
```bash
# Find process using port 9000
lsof -i :9000

# Kill it or change SonarQube port in docker-compose.sonarqube.yml:
ports:
  - "9001:9000"  # Use port 9001 instead
```

## Advanced Usage

### Custom Quality Profiles

```bash
# 1. Login to http://localhost:9000
# 2. Quality Profiles → Create
# 3. Activate only specific rules
# 4. Set as default for language
```

### Multiple Projects

```bash
# Specify different project keys
python3 scripts/validate-with-sonarqube.py src/ --project my-frontend
python3 scripts/validate-with-sonarqube.py api/ --project my-backend
```

### CI/CD Integration

```yaml
# .github/workflows/code-quality.yml
name: Code Quality

on: [push, pull_request]

jobs:
  sonarqube:
    runs-on: ubuntu-latest
    services:
      sonarqube:
        image: sonarqube:community
        ports:
          - 9000:9000

    steps:
      - uses: actions/checkout@v2

      - name: Setup Python
        uses: actions/setup-python@v2
        with:
          python-version: '3.9'

      - name: Install sonar-scanner
        run: |
          wget https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-4.7.0.2747-linux.zip
          unzip sonar-scanner-cli-*.zip
          echo "$PWD/sonar-scanner-*/bin" >> $GITHUB_PATH

      - name: Validate Code
        run: |
          python3 scripts/validate-with-sonarqube.py . \
            --url http://localhost:9000 \
            --token ${{ secrets.SONAR_TOKEN }} \
            --severity BLOCKER,CRITICAL
```

## References

- [SonarQube Documentation](https://docs.sonarqube.org/latest/)
- [SonarQube Rules](https://rules.sonarsource.com/)
- [sonar-scanner Documentation](https://docs.sonarqube.org/latest/analysis/scan/sonarscanner/)
- [Our Extracted Rules](../quality/sonarqube-rules/)
- [Rule Extraction Guide](../scripts/RSPEC-RULES.md)

---

**Last Updated:** 2026-01-27
