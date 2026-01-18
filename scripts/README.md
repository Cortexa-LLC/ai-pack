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
# Initialize
./scripts/github-integration.py init

# Configure
# Edit .github-integration.yml

# Set token
export GITHUB_TOKEN="ghp_your_token_here"

# Sync
./scripts/github-integration.py sync
```

**Documentation:**
- [Usage Guide](../docs/GITHUB-INTEGRATION-USAGE.md) - Complete guide
- [Configuration Example](../.github-integration.yml.example) - Config template
- [Integration Analysis](../docs/GITHUB-MCP-INTEGRATION-ANALYSIS.md) - Feature analysis

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
   chmod +x scripts/your-script.sh
   ```

2. **Add shebang:**
   ```bash
   #!/bin/bash
   ```

3. **Include help:**
   ```bash
   show_help() {
       cat << EOF
   Your Script Name

   Usage: $0 <command> [options]
   ...
   EOF
   }
   ```

4. **Document here** in this README

5. **Reference from docs/** if it's a major feature

---

## Future Scripts

Potential future additions:

- `slack-notifications.sh` - Slack integration
- `metrics-dashboard.sh` - Generate metrics
- `export-task-packets.sh` - Export task packets as PDF/HTML
- `backup-beads.sh` - Backup Beads database

---

**Last Updated:** 2026-01-18
