# GitHub Integration Usage Guide

**Version:** 1.0.0
**Date:** 2026-01-18
**Target:** AI-Pack Users with Hosted GitHub Repositories

## Overview

This guide explains how to integrate AI-Pack with hosted GitHub.com projects. The integration is **completely optional** and configured via YAML.

**What This Enables:**
- Sync Beads tasks ↔ GitHub Issues bidirectionally
- Create Epics/Stories from Beads task hierarchies
- Monitor CI/CD workflows and auto-create fix tasks
- Import GitHub issues into Beads work queue
- Track work across GitHub Projects and Beads

---

## Quick Start

### 1. Initialize Integration

```bash
cd your-project
./scripts/github-integration.py init
```

This creates `.github-integration.yml` with default settings.

### 2. Configure Your Settings

Edit `.github-integration.yml`:

```yaml
github:
  enabled: true
  repository: "your-org/your-repo"  # Auto-detected from git remote
  token: "${GITHUB_TOKEN}"          # Use environment variable

features:
  issue_sync:
    enabled: true                    # Sync Beads ↔ GitHub
  epic_management:
    enabled: true                    # Create epics/stories
  ci_monitoring:
    enabled: true                    # Monitor CI/CD
```

### 3. Set GitHub Token

```bash
export GITHUB_TOKEN="ghp_your_token_here"
# Or authenticate with GitHub CLI
gh auth login
```

### 4. Verify Status

```bash
./scripts/github-integration.py status
```

### 5. Start Syncing

```bash
# Bidirectional sync
./scripts/github-integration.py sync

# Or run individually
./scripts/github-integration.py export  # Beads → GitHub
./scripts/github-integration.py import  # GitHub → Beads
```

---

## Configuration Deep Dive

### Feature Toggles

Enable only the features you need:

```yaml
features:
  # Issue Management
  issue_sync:
    enabled: true
    auto_create_from_beads: true      # Auto-create GitHub issues
    auto_import_to_beads: true        # Auto-import GitHub issues
    bidirectional_sync: true          # Keep both in sync
    sync_interval: 300                # Sync every 5 minutes (0 = manual)

  # Epic/Story Management
  epic_management:
    enabled: true
    use_projects: true                # Use GitHub Projects
    default_project: 1                # Project number for epics

  # CI/CD Monitoring
  ci_monitoring:
    enabled: true
    watch_failures: true              # Watch for failures
    auto_create_failure_issues: true  # Create GitHub issues
    auto_create_failure_tasks: true   # Create Beads tasks
    check_interval: 60                # Check every minute

  # Pull Request Management
  pr_management:
    enabled: true
    auto_create_pr: false             # Don't auto-create PRs
    require_quality_gates: true       # Enforce quality gates
```

**Recommendation:** Start with `issue_sync` only, then enable others as needed.

### Sync Rules

Control what gets synced:

```yaml
sync_rules:
  # Beads → GitHub
  beads_to_github:
    statuses:
      - "open"
      - "in_progress"
      - "blocked"
    priorities:
      - "critical"
      - "high"
      - "normal"
    exclude_patterns:
      - "^Agent:"        # Don't sync agent tracking tasks
      - "^Internal:"     # Don't sync internal tasks

  # GitHub → Beads
  github_to_beads:
    required_labels:
      - "ai-pack"        # Only import issues with this label
    exclude_labels:
      - "wontfix"
      - "duplicate"
    only_open: true
```

**Best Practice:** Use `exclude_patterns` to keep internal tasks private.

### Labeling Strategy

Automatic label management:

```yaml
labels:
  auto_labels:
    - "ai-pack"
    - "beads-synced"

  priority_mapping:
    critical: "priority-critical"
    high: "priority-high"
    normal: "priority-normal"
    low: "priority-low"

  epic_label: "epic"
  story_label: "story"
  task_label: "task"
```

**GitHub Setup:** Create these labels in your repository settings for consistency.

### Task Packet Integration

Link GitHub issues in task packets:

```yaml
task_packets:
  link_in_contract: true               # Add GitHub link to contract
  auto_create_packets: true            # Create packets for imported issues
  directory_pattern: ".ai/tasks/{date}_{task_name}/"

  include_metadata: true
  metadata_fields:
    - "github_issue_number"
    - "github_issue_url"
    - "github_labels"
```

**Result:** Task packets automatically include GitHub issue links.

---

## Usage Patterns

### Pattern 1: Beads-First Workflow

**Scenario:** You work primarily in Beads, want GitHub visibility

```bash
# 1. Create tasks in Beads as usual
bd create "Implement user authentication" --priority high
bd create "Add dark mode toggle" --priority normal

# 2. Export to GitHub for team visibility
./scripts/github-integration.py export

# 3. Work on tasks locally
bd start bd-a1b2
# ... implement ...
bd close bd-a1b2

# 4. Sync status back to GitHub
./scripts/github-integration.py sync
```

**When to use:** Solo development or small teams, Beads is primary tool.

### Pattern 2: GitHub-First Workflow

**Scenario:** Team uses GitHub Issues, you want Beads locally

```bash
# 1. Team creates issues in GitHub (with "ai-pack" label)

# 2. Import to Beads
./scripts/github-integration.py import

# 3. Find imported work
bd ready

# 4. Work on tasks
bd start bd-x1y2
# ... implement ...
bd close bd-x1y2

# 5. Sync status back to GitHub
./scripts/github-integration.py sync
```

**When to use:** Large teams, GitHub is primary tool, you want local Beads benefits.

### Pattern 3: Bidirectional Sync

**Scenario:** Mixed team, some use GitHub, some use Beads

```bash
# Run continuous sync (in background or separate terminal)
./scripts/github-integration.py sync

# Or set up periodic sync via cron
# Add to crontab:
# */5 * * * * cd /path/to/project && ./scripts/github-integration.py sync
```

**When to use:** Hybrid teams, both tools actively used.

### Pattern 4: CI/CD Monitoring

**Scenario:** Want automatic issue/task creation for CI failures

```bash
# Monitor CI continuously
./scripts/github-integration.py monitor

# Or check on-demand
./scripts/github-integration.py check-ci
```

**Configuration:**
```yaml
ci_config:
  on_failure:
    action: "both"          # Create both GitHub issue and Beads task
    assignee: "Engineer-CI"
    priority: "critical"
```

**Result:** CI failures automatically create actionable tasks.

---

## Epic and Story Management

### Creating Epics

```bash
# 1. Create epic task in Beads with subtasks
epic_id=$(bd create "User Authentication System" --priority high)

bd create "Design auth API" --depends-on ${epic_id}
bd create "Implement JWT tokens" --depends-on ${epic_id}
bd create "Add password hashing" --depends-on ${epic_id}
bd create "Write auth tests" --depends-on ${epic_id}

# 2. Create GitHub epic with stories
./scripts/github-integration.py create-epic ${epic_id}
```

**GitHub Result:**
- Epic issue created with checklist of subtasks
- Each subtask becomes a story issue
- All linked properly (epic ↔ stories ↔ Beads tasks)

### Epic Representation

Configure how epics appear in GitHub:

```yaml
epics:
  representation: "project"  # Options: "project" | "issue" | "milestone"
  naming_pattern: "Epic: {title}"

  story_template: |
    **Epic:** #{epic_number}
    **Beads Task:** {beads_id}

    {description}

    ## Acceptance Criteria
    {acceptance_criteria}

    ---
    Part of Epic #{epic_number}
    Managed by ai-pack
```

---

## CI/CD Integration

### Monitoring Workflows

**Check current status:**
```bash
./scripts/github-integration.py check-ci

# Output:
# Recent Workflows:
#   CI: completed - success
#   Tests: completed - success
#   Deploy: in_progress - running
```

**Continuous monitoring:**
```bash
./scripts/github-integration.py monitor

# Runs indefinitely, checks every minute (configurable)
# Creates issues/tasks automatically on failure
```

### Handling Failures

When CI fails, integration can:

1. **Create GitHub Issue:**
   - Title: "CI Failure: [Workflow Name]"
   - Labels: `ci-failure`, `priority-critical`
   - Body: Link to workflow run, error details
   - Auto-assigned

2. **Create Beads Task:**
   - Task: "Fix CI failure: [Workflow Name]"
   - Priority: `critical` (configurable)
   - Assignee: Configured engineer
   - Linked to GitHub issue

3. **Both (recommended):**
   ```yaml
   ci_config:
     on_failure:
       action: "both"  # Create both issue and task
   ```

### Quality Gate Integration

Block work if CI is failing:

```yaml
features:
  quality_gates:
    enabled: true
    block_on_ci_failure: true       # Block new work if CI failing
    require_coverage: true
    coverage_threshold: 80
    require_tdd_validation: true
```

**Effect:** Engineers must fix CI before starting new features.

---

## Advanced Configuration

### Environment Variables

Override config with environment variables:

```bash
# Set token (REQUIRED)
export GITHUB_TOKEN="ghp_your_token_here"

# Override repository
export GITHUB_REPO="different-org/different-repo"

# For GitHub Enterprise
export GITHUB_API_URL="https://github.your-company.com/api/v3"
```

**Priority:** Environment variables > Config file

### Multiple Repositories

Use separate config files:

```bash
# Copy config template
cp .github-integration.yml.example .github-integration-api.yml
cp .github-integration.yml.example .github-integration-web.yml

# Edit each for different repos
# Then use:
CONFIG_FILE=.github-integration-api.yml ./scripts/github-integration.py sync
CONFIG_FILE=.github-integration-web.yml ./scripts/github-integration.py sync
```

### GitHub Enterprise

Configure API URL:

```yaml
github:
  api_url: "https://github.your-company.com/api/v3"
```

### Rate Limiting

GitHub API has rate limits (5000 requests/hour for authenticated users):

```yaml
advanced:
  rate_limit:
    max_requests: 4500      # Leave buffer
    backoff_strategy: "exponential"

  cache:
    enabled: true           # Cache responses
    ttl: 300                # 5 minutes
```

### Logging

Debug issues with detailed logging:

```yaml
advanced:
  logging:
    enabled: true
    level: "debug"          # debug | info | warning | error
    file: ".github-integration.log"
```

**View logs:**
```bash
tail -f .github-integration.log
```

### Dry Run Mode

Test without making changes:

```yaml
advanced:
  dry_run: true
```

**Effect:** Prints what would happen, doesn't create issues/tasks.

---

## Workflow Integration

### With Orchestrator

Orchestrator can trigger GitHub operations:

```bash
# In orchestrator role
# After creating Beads tasks:
./scripts/github-integration.py export

# After epic planning:
./scripts/github-integration.py create-epic ${epic_id}
```

### With Engineer

Engineer sees imported GitHub issues:

```bash
# Find work (includes imported GitHub issues)
bd ready

# Work on task
bd start bd-x1y2
# ... implement ...
bd close bd-x1y2

# Status automatically syncs to GitHub
```

### Automated Sync

Set up automatic syncing via cron or systemd timer:

**Cron (every 5 minutes):**
```bash
# Edit crontab
crontab -e

# Add:
*/5 * * * * cd /path/to/project && ./scripts/github-integration.py sync >> /path/to/project/.github-sync.log 2>&1
```

**Systemd Timer (more sophisticated):**
```ini
# /etc/systemd/user/github-sync.timer
[Unit]
Description=AI-Pack GitHub Sync Timer

[Timer]
OnBootSec=1min
OnUnitActiveSec=5min

[Install]
WantedBy=timers.target
```

```ini
# /etc/systemd/user/github-sync.service
[Unit]
Description=AI-Pack GitHub Sync

[Service]
Type=oneshot
WorkingDirectory=/path/to/project
ExecStart=/path/to/project/scripts/github-integration.py sync
```

```bash
# Enable timer
systemctl --user enable github-sync.timer
systemctl --user start github-sync.timer
```

---

## Security Best Practices

### Token Management

**DO:**
- Use environment variables: `export GITHUB_TOKEN="..."`
- Use GitHub CLI: `gh auth login` (token stored securely)
- Rotate tokens regularly (every 90 days)
- Use fine-grained tokens with minimal scopes

**DON'T:**
- Commit `.github-integration.yml` with token
- Share tokens via email/Slack
- Use tokens with more permissions than needed
- Leave tokens in shell history

### Minimal Permissions

Create token with only required scopes:

**Required:**
- `repo` (full control of private repositories)
- `workflow` (update GitHub Actions workflows)

**Optional (if using):**
- `read:org` (read org/team membership)
- `project` (manage GitHub Projects)

### Add to .gitignore

```bash
echo ".github-integration.yml" >> .gitignore
git add .gitignore
git commit -m "Ignore GitHub integration config"
```

**Reason:** Config may contain sensitive data or environment-specific settings.

---

## Troubleshooting

### "GitHub CLI not authenticated"

```bash
gh auth login
# Follow prompts to authenticate
```

### "yq not found"

```bash
brew install yq
```

### "Beads not found"

```bash
# Install Beads
curl -fsSL https://raw.githubusercontent.com/steveyegge/beads/main/scripts/install.sh | bash
```

### "No issues synced"

**Check:**
1. Is issue_sync enabled in config?
2. Do Beads tasks match sync rules?
3. Are exclusion patterns too broad?
4. Check logs: `tail -f .github-integration.log`

### "Rate limit exceeded"

**Solutions:**
1. Reduce `sync_interval` in config
2. Enable caching: `cache.enabled: true`
3. Use webhooks instead of polling
4. Wait for rate limit reset (check with `gh api rate_limit`)

### "Import not working"

**Check:**
1. Do GitHub issues have required label (`ai-pack`)?
2. Are issues in correct state (open)?
3. Are excluded labels applied?
4. Run with debug logging:
   ```bash
   # Set in config
   logging.level: "debug"
   ```

---

## Examples

### Example 1: Sprint Planning with GitHub Projects

```bash
# 1. Create sprint epic in Beads
sprint_id=$(bd create "Sprint 23 - User Features" --priority high)

# 2. Add user stories
bd create "User profile page" --depends-on ${sprint_id}
bd create "Settings page" --depends-on ${sprint_id}
bd create "Notifications" --depends-on ${sprint_id}

# 3. Create GitHub epic with project
./scripts/github-integration.py create-epic ${sprint_id}

# 4. Team sees epic in GitHub Projects
# 5. Engineers pull work from Beads
bd ready

# 6. Work syncs bidirectionally
```

### Example 2: Bug Triage Workflow

```bash
# 1. Team reports bugs as GitHub issues (label: "bug", "ai-pack")

# 2. Import to Beads for triage
./scripts/github-integration.py import

# 3. Triage in Beads
bd list --json | jq '.[] | select(.title | contains("bug"))'

# 4. Prioritize
bd comment bd-x1y2 "Priority: critical"

# 5. Assign
bd start bd-x1y2

# 6. Status syncs back to GitHub automatically
```

### Example 3: CI-Driven Development

```bash
# 1. Enable CI monitoring
# In .github-integration.yml:
# ci_monitoring.enabled: true
# ci_monitoring.auto_create_failure_tasks: true

# 2. Start monitoring
./scripts/github-integration.py monitor

# 3. When CI fails:
#    - Beads task auto-created with priority: critical
#    - GitHub issue auto-created with link to workflow
#    - Engineer gets automatic work item

# 4. Engineer finds task
bd ready
# Shows: "Fix CI failure: Tests" [priority: critical]

# 5. Fix and push
bd start bd-fail123
# ... fix ...
git push
bd close bd-fail123

# 6. CI passes, issue closed automatically
```

---

## Summary

**GitHub integration is completely optional and configurable.**

**Key Benefits:**
- ✅ Team visibility (GitHub) + Personal productivity (Beads)
- ✅ Automatic sync - no manual copying
- ✅ CI/CD monitoring with automatic task creation
- ✅ Epic/Story management across both systems
- ✅ Quality gate enforcement

**Getting Started:**
```bash
./scripts/github-integration.py init
# Edit .github-integration.yml
export GITHUB_TOKEN="your_token"
./scripts/github-integration.py sync
```

**Best Practice:** Start with just `issue_sync` enabled, add features as needed.

---

**See Also:**
- [GitHub MCP Integration Analysis](GITHUB-MCP-INTEGRATION-ANALYSIS.md)
- [GitHub MCP Integration Guide](GITHUB-MCP-INTEGRATION-GUIDE.md)
- [Beads Integration](../quality/tooling/beads-integration.md)
- [Orchestrator Role](../roles/orchestrator.md)
- [Engineer Role](../roles/engineer.md)
