# AI-Pack Scripts

This directory contains automation scripts for AI-Pack workflows.

## Available Scripts

### `github-integration.py`

**Optional** GitHub integration for hosted GitHub.com repositories.

**Features:**
- Sync tasks ↔ GitHub Issues bidirectionally
- Create Epics/Stories from task hierarchies
- Monitor CI/CD workflows and auto-create fix tasks
- Import GitHub issues into Beads work queue
- Track work across GitHub Projects and Beads

**Prerequisites:**
- `yq` - YAML parser: `brew install yq`
- `jq` - JSON parser: `brew install jq`
- `gh` - GitHub CLI: `brew install gh`
- `bd` - Beads (see quality/tooling/beads-integration.md)

**Quick Start:**
```bash
# Note: Replace ${AI_PACK_ROOT} with your actual path (.ai-pack, ai-pack, etc.)
# See Setup Guide for path detection helpers

# Initialize (from project root)
${AI_PACK_ROOT}/scripts/github-integration.py init

# Configure
# Edit ${AI_PACK_ROOT}/.github-integration.yml

# Set token
export GITHUB_TOKEN="ghp_your_token_here"

# Sync
${AI_PACK_ROOT}/scripts/github-integration.py sync
```

**Documentation:**
- [Setup Guide](../docs/GITHUB-INTEGRATION-SETUP.md) - Installation paths and environment setup
- [Agent Triggers](../docs/GITHUB-AGENT-TRIGGERS.md) - Auto-sync on role actions
- [Usage Guide](../docs/GITHUB-INTEGRATION-USAGE.md) - Complete feature guide
- [Configuration Example](../.github-integration.yml.example) - Config template
- [Work Item Patterns](../docs/WORK-ITEM-PATTERNS.md) - Epic/Story/Task patterns
- [Integration Summary](../docs/GITHUB-INTEGRATION-SUMMARY.md) - Overview and quick reference

**Commands:**
```bash
init              Initialize GitHub integration
sync              Sync tasks with GitHub issues
import            Import GitHub issues to Beads
export            Export tasks to GitHub
monitor           Monitor CI/CD workflows
check-ci          Check current CI status
create-epic       Create epic from task
status            Show integration status
help              Show help message
```

**Note:** GitHub integration is completely optional. AI-Pack works fully without it.

---

### `rspec_rule_extractor.py`

Extract SonarQube code quality rules from the [rspec repository](https://github.com/SonarSource/rspec) for specific programming languages.

**Features:**
- Parse metadata.json files from rspec repository
- Multiple output formats: CSV (CI/CD), JSON (programmatic), Markdown (docs)
- Language name normalization (e.g., `cpp` → `cfamily`, `js` → `javascript`)
- Full rule metadata: impacts, tags, remediation costs, severity levels

**Usage:**
```bash
# Extract Go rules as CSV (for CI/CD)
python3 scripts/rspec_rule_extractor.py ~/Projects/Vibe/rspec go --format csv > go_rules.csv

# Extract Python rules as JSON (for tooling)
python3 scripts/rspec_rule_extractor.py ~/Projects/Vibe/rspec python --format json > python_rules.json

# Extract Java rules as Markdown (for docs)
python3 scripts/rspec_rule_extractor.py ~/Projects/Vibe/rspec java --format markdown > java_rules.md
```

**Supported Languages:**
C/C++, C#, Go, Python, JavaScript/TypeScript, Java, Kotlin, Swift, Ruby, PHP, Scala, Shell/Bash

**Documentation:** See [scripts/RSPEC-RULES.md](./RSPEC-RULES.md) for complete guide

---

### `generate_all_language_rules.sh`

Generate SonarQube rule sets for all languages in our clean code standards.

**Features:**
- Batch generation for all supported languages (8 languages)
- Creates CSV + JSON + Markdown for each language
- Organized directory structure with index README
- Progress indicators and error reporting
- Total: 4,131 rules across all languages

**Usage:**
```bash
# Generate all language rule sets
./scripts/generate_all_language_rules.sh ~/Projects/Vibe/rspec ./quality/sonarqube-rules

# Output:
# - quality/sonarqube-rules/cpp/rules.csv (916 rules)
# - quality/sonarqube-rules/go/rules.csv (108 rules)
# - quality/sonarqube-rules/python/rules.csv (493 rules)
# - ... and 5 more languages
```

**Integration:**
```bash
# Example: Find all critical Go rules for CI/CD
grep "CRITICAL" quality/sonarqube-rules/go/rules.csv

# Example: Get bug-type rules in Python
awk -F',' '$3=="BUG"' quality/sonarqube-rules/python/rules.csv
```

**Documentation:** See [scripts/RSPEC-RULES.md](./RSPEC-RULES.md) for integration examples

---

### `setup-sonarqube.py`

One-command setup for SonarQube Community Edition via Docker Compose.

**Features:**
- Starts SonarQube CE + PostgreSQL in Docker
- Waits for startup and validates health
- Changes default admin password
- Generates API token for agents
- Saves configuration automatically
- Cross-platform (Windows, macOS, Linux)

**Usage:**
```bash
# One-command setup
python3 scripts/setup-sonarqube.py

# Access web UI
open http://localhost:9000  # admin / admin123
```

**What it does:**
1. Starts Docker containers
2. Waits for SonarQube to be ready (~1 minute)
3. Creates API token
4. Saves to `.sonarqube-config`

**Documentation:** See [docs/SONARQUBE-INTEGRATION.md](../docs/SONARQUBE-INTEGRATION.md)

---

### `validate-with-sonarqube.py`

Validate code using SonarQube Community Edition. Agent-friendly with JSON output.

**Features:**
- Runs sonar-scanner on files/directories
- Fetches violations via SonarQube API
- Enriches with rule metadata (impacts, remediation cost, tags)
- Returns clean JSON for agent consumption
- Filters by severity (BLOCKER, CRITICAL, etc.)
- Auto-detects language from file extension

**Usage:**
```bash
# Validate single file
python3 scripts/validate-with-sonarqube.py src/server.go

# JSON output for agents
python3 scripts/validate-with-sonarqube.py src/server.go --format json

# Filter by severity
python3 scripts/validate-with-sonarqube.py src/ --severity BLOCKER,CRITICAL

# Validate directory
python3 scripts/validate-with-sonarqube.py ./src --project myproject
```

**JSON Output:**
```json
{
  "success": true,
  "violations": [
    {
      "rule_id": "S1192",
      "line": 42,
      "severity": "CRITICAL",
      "message": "Define a constant instead of duplicating...",
      "remediation_cost": "10min",
      "impacts": "MAINTAINABILITY:HIGH"
    }
  ],
  "summary": {"total": 1, "by_severity": {"CRITICAL": 1}}
}
```

**Documentation:** See [docs/SONARQUBE-INTEGRATION.md](../docs/SONARQUBE-INTEGRATION.md)

---

### `query-rules.py`

Search and filter SonarQube rules from extracted metadata.

**Features:**
- Search by language, type, severity, impact, tags
- Explain specific rules with full metadata
- Multiple output formats (table, JSON, markdown, summary)
- Filter by keyword in title/description
- Export rule sets for custom use

**Usage:**
```bash
# List all Go rules
python3 scripts/query-rules.py --language go

# Find critical bugs
python3 scripts/query-rules.py --language python --type BUG --severity CRITICAL

# Explain a specific rule
python3 scripts/query-rules.py --rule S1192

# Find security rules
python3 scripts/query-rules.py --language go --impact SECURITY:HIGH

# Export as JSON
python3 scripts/query-rules.py --language java --format json > java_rules.json
```

**Documentation:** See [scripts/RSPEC-RULES.md](./RSPEC-RULES.md)

---

## Adding New Scripts

When adding new scripts to this directory:

1. **Make executable:**
   ```bash
   # Python scripts
   chmod +x scripts/your-script.py

   # Shell scripts
   chmod +x scripts/your-script.sh
   ```

2. **Add shebang:**
   ```bash
   # Python (preferred for cross-platform)
   #!/usr/bin/env python3

   # Bash (macOS/Linux only)
   #!/bin/bash
   ```

3. **Include help:**
   ```python
   # Python example
   def show_help():
       print("""
   Your Script Name

   Usage: python3 scripts/your-script.py <command> [options]
   ...
   """)
   ```

4. **Document here** in this README

5. **Reference from docs/** if it's a major feature

**Note:** Prefer Python (.py) over Bash (.sh) for cross-platform compatibility (Windows, macOS, Linux).

---

## Future Scripts

Potential future additions:

- `slack-notifications.py` - Slack integration
- `metrics-dashboard.py` - Generate metrics
- `export-task-packets.py` - Export task packets as PDF/HTML
- `backup-beads.py` - Backup Beads database

---

**Last Updated:** 2026-01-18
