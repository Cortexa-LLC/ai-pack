---
sidebar_position: 3
title: GitHub Integration
---

# GitHub Integration (Optional)

Optional integration for hosted GitHub.com repositories.

## Features

- Sync Beads tasks ↔ GitHub Issues bidirectionally
- Create Epics/Stories from Beads hierarchies
- Monitor CI/CD workflows and auto-create fix tasks
- Track work in your GitHub Repository

## Quick Start

```bash
# Initialize integration
.ai-pack/scripts/github-integration.py init

# Configure settings
export GITHUB_TOKEN="ghp_your_token_here"

# Start syncing
.ai-pack/scripts/github-integration.py sync
```text

## Configuration

All features configured via `.github-integration.yml`:

```yaml
github:
  enabled: true
  repository: "your-org/your-repo"

features:
  issue_sync:
    enabled: true
  epic_management:
    enabled: true
  ci_monitoring:
    enabled: true
  agent_triggers:
    enabled: true
```text

## Documentation

See project documentation for complete guides:
- GitHub Integration Setup
- GitHub Agent Triggers
- GitHub Integration Usage Guide
- Work Item Patterns

**Note:** GitHub integration is completely optional. AI-Pack works fully without it.
