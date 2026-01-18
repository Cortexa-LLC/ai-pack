# AI-Pack Scripts

This directory contains automation scripts for AI-Pack workflows.

## Available Scripts

### `github-integration.py`

**Optional** GitHub integration for hosted GitHub.com repositories.

**Features:**
- Sync Beads tasks ↔ GitHub Issues bidirectionally
- Create Epics/Stories from Beads task hierarchies
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
sync              Sync Beads tasks with GitHub issues
import            Import GitHub issues to Beads
export            Export Beads tasks to GitHub
monitor           Monitor CI/CD workflows
check-ci          Check current CI status
create-epic       Create epic from Beads task
status            Show integration status
help              Show help message
```

**Note:** GitHub integration is completely optional. AI-Pack works fully without it.

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
