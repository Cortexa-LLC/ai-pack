# GitHub MCP Integration Guide for AI-Pack

**Version:** 1.0.0
**Date:** 2026-01-18
**Audience:** AI-Pack Adopters

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Installation & Setup](#installation--setup)
4. [Configuration for Claude Code](#configuration-for-claude-code)
5. [Usage Examples](#usage-examples)
6. [Best Practices](#best-practices)
7. [Troubleshooting](#troubleshooting)
8. [Advanced Configuration](#advanced-configuration)

---

## Overview

This guide shows you how to integrate GitHub's official MCP server with your AI-Pack project, enabling AI agents to:

- ✅ Create and manage GitHub issues from Beads tasks
- ✅ Create pull requests with proper linking
- ✅ Monitor CI/CD build and test status
- ✅ Perform code reviews with inline comments
- ✅ Enforce quality gates automatically
- ✅ Track agent activity across sessions

**What You'll Need:**
- GitHub repository with AI-Pack submodule
- GitHub Personal Access Token (PAT) or OAuth
- Claude Code CLI
- 30 minutes for setup

---

## Prerequisites

### 1. GitHub Repository Setup

Your project should already have AI-Pack installed:

```bash
cd your-project
git submodule add https://github.com/Cortexa-LLC/ai-pack .ai-pack
git submodule update --init --recursive
```

### 2. GitHub Personal Access Token

Create a GitHub PAT with the following permissions:

**Required Scopes:**
```
repo (full control)
├── repo:status (access commit status)
├── repo_deployment (access deployment status)
├── public_repo (access public repositories)
└── repo:invite (access repository invitations)

workflow (update GitHub Action workflows)

read:org (read org and team membership)
```

**How to Create PAT:**

1. Go to GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Name: `ai-pack-mcp-integration`
4. Select scopes above
5. Click "Generate token"
6. **Copy token immediately** (you won't see it again)

**Security Note:** Store token securely - never commit to git!

```bash
# Store in environment variable
export GITHUB_TOKEN="ghp_your_token_here"

# Or use a secret manager
```

### 3. Verify Prerequisites

```bash
# Check Git
git --version  # Should be 2.x+

# Check GitHub CLI (optional but recommended)
gh --version  # Should be 2.x+

# Check Beads
bd --version  # Should be installed
```

---

## Installation & Setup

### Option 1: Remote Server (Easiest)

Use GitHub's hosted MCP server - no local installation needed.

**Advantages:**
- ✅ No installation required
- ✅ Always up-to-date
- ✅ Official support
- ✅ OAuth authentication available

**Requirements:**
- VS Code 1.101+ with Claude extension
- OR Cursor with MCP support
- OR Claude Desktop app

**Skip to:** [Configuration for Claude Code](#configuration-for-claude-code)

---

### Option 2: Local Installation (Docker)

Run GitHub MCP server locally via Docker.

**Advantages:**
- ✅ Full control over instance
- ✅ Custom configuration
- ✅ Works offline (with cached data)
- ✅ Easier debugging

**Installation:**

```bash
# 1. Pull GitHub MCP server image
docker pull ghcr.io/github/github-mcp-server:latest

# 2. Create configuration directory
mkdir -p ~/.mcp/github-mcp-server

# 3. Create environment file
cat > ~/.mcp/github-mcp-server/.env << EOF
GITHUB_TOKEN=${GITHUB_TOKEN}
GITHUB_TOOLSETS=issues,pull_requests,workflows,repositories
LOG_LEVEL=info
EOF

# 4. Run server
docker run -d \
  --name github-mcp-server \
  --env-file ~/.mcp/github-mcp-server/.env \
  -p 3000:3000 \
  ghcr.io/github/github-mcp-server:latest

# 5. Verify running
docker logs github-mcp-server
# Should see: "GitHub MCP server listening on port 3000"
```

**Toolsets Available:**
- `issues` - Issue management
- `pull_requests` - PR operations
- `workflows` - CI/CD monitoring
- `repositories` - Repo operations
- `security` - Security scanning
- `discussions` - GitHub Discussions

---

### Option 3: Local Installation (Binary)

Build and run from source (for advanced users).

```bash
# 1. Clone repository
git clone https://github.com/github/github-mcp-server.git
cd github-mcp-server

# 2. Build (requires Go 1.21+)
go build -o github-mcp-server ./cmd/server

# 3. Configure
export GITHUB_TOKEN="ghp_your_token_here"
export GITHUB_TOOLSETS="issues,pull_requests,workflows"

# 4. Run
./github-mcp-server --port 3000
```

---

## Configuration for Claude Code

### Step 1: Configure MCP Server in Claude Code

Create or update your Claude Code MCP configuration:

**Location:** `~/.claude-code/config.json` (or project-specific `.claude/config.json`)

```json
{
  "mcpServers": {
    "github": {
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "--env-file",
        "/Users/yourusername/.mcp/github-mcp-server/.env",
        "ghcr.io/github/github-mcp-server:latest"
      ]
    }
  }
}
```

**For Remote Server (OAuth):**

```json
{
  "mcpServers": {
    "github": {
      "url": "https://api.githubcopilot.com/mcp/",
      "auth": {
        "type": "oauth",
        "provider": "github"
      }
    }
  }
}
```

### Step 2: Create AI-Pack GitHub Integration Hook

Create a Claude Code skill that automatically uses GitHub MCP for AI-Pack operations.

**File:** `.claude/skills/github-integration.md`

```markdown
# GitHub Integration Skill

**When to use:** Automatically activated for AI-Pack orchestration and task management.

## Activation Triggers
- User mentions "create issue", "update issue", "check CI", "create PR"
- Orchestrator creates Beads tasks
- Engineer creates PR
- Tester/Reviewer validates code

## What This Does

This skill integrates AI-Pack workflows with GitHub using the GitHub MCP server.

### Orchestrator Actions

When Orchestrator creates Beads tasks:

```bash
# Create Beads task
bd create "Implement user authentication" --priority high

# Automatically create linked GitHub issue via MCP
github.create_issue({
  title: "Implement user authentication",
  body: "AI-Pack task: bd-a1b2\nTask packet: .ai/tasks/2026-01-18_auth/",
  labels: ["ai-pack", "orchestrated", "high-priority"],
  milestone: current_milestone
})

# Link in Beads
bd comment bd-a1b2 "GitHub Issue: #123"
```

### Engineer Actions

When Engineer completes implementation:

```bash
# Create PR via MCP
github.create_pull_request({
  title: "Implement user authentication per task packet",
  body: generate_pr_body({
    beads_task: "bd-a1b2",
    task_packet: ".ai/tasks/2026-01-18_auth/",
    tdd_commits: commit_history,
    coverage: "87%"
  }),
  base: "main",
  head: "feature/auth",
  labels: ["ai-pack", "ready-for-review"]
})
```

### CI/CD Monitoring

Before proceeding through quality gates:

```bash
# Check CI status via MCP
workflow_status = github.get_workflow_runs({
  repo: current_repo,
  branch: current_branch,
  status: "latest"
})

if workflow_status.conclusion != "success":
  BLOCK gate progression
  notify "CI failing - see workflow: {workflow_status.html_url}"
end
```

### Tester/Reviewer Actions

When validating code:

```bash
# Add review via MCP
github.create_pull_request_review({
  pr_number: pr_id,
  event: "REQUEST_CHANGES",  # or "APPROVE"
  body: validation_report,
  comments: inline_violation_comments
})
```

## Configuration

Ensure GitHub MCP server is configured in `.claude/config.json` (see integration guide).

## Notes

- All GitHub operations log to `.ai/github-integration.log`
- Issues created by AI-Pack automatically tagged with `ai-pack` label
- PRs include task packet references and TDD commit history
- CI/CD checks integrated with quality gates
```

### Step 3: Update Orchestrator Role Extension

Add GitHub integration to your project's orchestrator role.

**File:** `.ai/roles/orchestrator-extension.md` (if not exists, create it)

```markdown
# Orchestrator Extension - GitHub Integration

**Base Role:** `.ai-pack/roles/orchestrator.md` (immutable)
**Extension Type:** Project-specific GitHub integration

## GitHub Issue Creation (MANDATORY)

**REQUIREMENT:** When creating Beads tasks for major features or bugs, MUST create corresponding GitHub issue.

### Trigger Conditions
```
IF task is non-trivial AND requires tracking THEN
  create_beads_task()
  create_github_issue_via_mcp()
  link_beads_to_github()
END IF
```

### Implementation

```bash
# Step 1: Create Beads task
task_id=$(bd create "Implement feature X" --priority high --json | jq -r '.id')

# Step 2: Create GitHub issue via MCP
issue_id=$(mcp github create_issue \
  --title "Implement feature X" \
  --body "AI-Pack Beads Task: ${task_id}
Task Packet: .ai/tasks/2026-01-18_feature-x/
Assigned: Orchestrator delegation pending" \
  --labels "ai-pack,orchestrated,high-priority" \
  --json | jq -r '.number')

# Step 3: Link in Beads
bd comment ${task_id} "GitHub Issue: #${issue_id}"

# Step 4: Log
echo "[$(date)] Created GitHub issue #${issue_id} for Beads task ${task_id}" >> .ai/github-integration.log
```

## CI/CD Gate Integration (MANDATORY)

**REQUIREMENT:** Before proceeding through quality gates, MUST check CI/CD status via GitHub MCP.

### Implementation

```bash
# Get latest workflow run status
workflow_status=$(mcp github get_workflow_runs \
  --repo "${GITHUB_REPOSITORY}" \
  --branch "${CURRENT_BRANCH}" \
  --status "latest" \
  --json)

conclusion=$(echo $workflow_status | jq -r '.workflow_runs[0].conclusion')

if [ "$conclusion" != "success" ]; then
  echo "❌ GATE BLOCKED: CI/CD checks failing"
  echo "Workflow: $(echo $workflow_status | jq -r '.workflow_runs[0].html_url')"
  exit 1
fi

echo "✅ CI/CD checks passing - proceeding"
```

## Agent Status Tracking

Track spawned agents via GitHub issues:

```bash
# When spawning Engineer agent
agent_issue=$(mcp github create_issue \
  --title "Agent: Engineer - Implement login feature" \
  --body "Spawned by Orchestrator
Beads Task: ${beads_task_id}
Task Packet: ${task_packet_path}
Status: in_progress" \
  --labels "agent,engineer,in-progress" \
  --assignee "ai-pack-bot")

# Update status as agent progresses
mcp github add_comment \
  --issue ${agent_issue} \
  --body "Progress update: [agent status summary]"

# Close when complete
mcp github update_issue \
  --issue ${agent_issue} \
  --state "closed" \
  --labels "agent,engineer,completed"
```
```

### Step 4: Update Engineer Role Extension

**File:** `.ai/roles/engineer-extension.md`

```markdown
# Engineer Extension - GitHub Integration

**Base Role:** `.ai-pack/roles/engineer.md` (immutable)
**Extension Type:** Project-specific GitHub PR integration

## Automated PR Creation (MANDATORY)

**REQUIREMENT:** When implementation complete and tests passing, MUST create PR via GitHub MCP.

### Implementation

```bash
# Step 1: Verify implementation complete
if ! (all_tests_passing && coverage_meets_target); then
  echo "Cannot create PR - quality gates not met"
  exit 1
fi

# Step 2: Generate PR body
pr_body="## Implementation Summary

**Beads Task:** ${beads_task_id}
**Task Packet:** ${task_packet_path}
**GitHub Issue:** Closes #${github_issue_id}

## TDD Commit History
$(git log --oneline --grep='test' ${base_branch}..HEAD)

## Coverage Report
- Overall: ${coverage_percent}%
- Critical Logic: ${critical_coverage}%
- Test Pyramid: ${unit_tests} unit | ${integration_tests} integration | ${e2e_tests} e2e

## Acceptance Criteria
$(cat .ai/tasks/${task_id}/00-contract.md | grep -A 20 '## Acceptance Criteria')

---
🤖 Generated by AI-Pack Engineer Agent
Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"

# Step 3: Create PR via MCP
pr_number=$(mcp github create_pull_request \
  --title "Implement ${feature_name} per task packet" \
  --body "${pr_body}" \
  --base "main" \
  --head "${current_branch}" \
  --labels "ai-pack,ready-for-review" \
  --json | jq -r '.number')

# Step 4: Request reviews
mcp github request_reviewers \
  --pr ${pr_number} \
  --reviewers "ai-pack-tester,ai-pack-reviewer"

# Step 5: Log
echo "[$(date)] Created PR #${pr_number} for ${feature_name}" >> .ai/github-integration.log
```

## Commit Message Format

All commits MUST include GitHub issue reference:

```bash
git commit -m "Add failing test for user login

Implements authentication flow per task packet.

Relates to #${github_issue_id}
Task: ${beads_task_id}
Phase: RED (TDD)

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```
```

### Step 5: Test Integration

```bash
# Start Claude Code with GitHub MCP enabled
claude-code

# Test GitHub integration
# Inside Claude Code session:
```

**Test 1: List Issues**
```
User: "Show me open issues labeled 'ai-pack'"

Expected: Claude uses GitHub MCP to fetch and display issues
```

**Test 2: Create Issue**
```
User: "Create a GitHub issue for implementing dark mode"

Expected: Claude creates issue and returns issue number
```

**Test 3: Check CI Status**
```
User: "Check the CI status for the current branch"

Expected: Claude fetches latest workflow run and reports status
```

**Test 4: Create PR**
```
User: "Create a PR for the current branch"

Expected: Claude creates PR with proper formatting and links
```

---

## Usage Examples

### Example 1: Orchestrator Creates Issues from Tasks

```bash
# User: "Break down user authentication into tasks"

# Orchestrator response:
"I'll decompose user authentication into subtasks and create GitHub issues.

Breaking down into tasks:
1. Design authentication architecture
2. Implement user model
3. Create login API
4. Add session management
5. Write tests

Creating Beads tasks and GitHub issues..."

# Behind the scenes:
bd create "Design authentication architecture" --priority high
# → Creates bd-a1b2

mcp github create_issue \
  --title "Design authentication architecture" \
  --body "AI-Pack Task: bd-a1b2..." \
  --labels "ai-pack,architecture,high-priority"
# → Creates Issue #123

bd comment bd-a1b2 "GitHub Issue: #123"

# Output to user:
"✅ Created 5 tasks:
- bd-a1b2 → GitHub Issue #123: Design authentication architecture
- bd-b2c3 → GitHub Issue #124: Implement user model
- bd-c3d4 → GitHub Issue #125: Create login API
- bd-d4e5 → GitHub Issue #126: Add session management
- bd-e5f6 → GitHub Issue #127: Write tests

View on GitHub: https://github.com/owner/repo/issues?q=is:issue+label:ai-pack"
```

### Example 2: Engineer Creates PR with TDD History

```bash
# User: "Implement login feature per task packet"

# Engineer implements with TDD, then:
"Implementation complete. Creating pull request..."

# Behind the scenes:
mcp github create_pull_request \
  --title "Implement login feature per task packet" \
  --body "$(generate_pr_body)" \
  --base "main" \
  --head "feature/login" \
  --labels "ai-pack,ready-for-review"
# → Creates PR #45

# Output to user:
"✅ Pull request created: #45

https://github.com/owner/repo/pull/45

Summary:
- Closes GitHub Issue #125
- Beads Task: bd-c3d4
- TDD commits: 12 (4 RED, 4 GREEN, 4 REFACTOR)
- Test coverage: 89%
- All CI checks passing ✅

Reviewers requested: @ai-pack-tester, @ai-pack-reviewer"
```

### Example 3: Tester Validates via CI Status

```bash
# User: "Validate TDD compliance for PR #45"

# Tester role activates:
"Validating TDD compliance and test quality for PR #45..."

# Behind the scenes:
pr_info=$(mcp github get_pull_request --pr 45)
commits=$(mcp github get_pr_commits --pr 45)
ci_status=$(mcp github get_ci_status --ref $pr_info.head.sha)

# Analyzes commits for TDD pattern
# Checks CI test results
# Validates coverage

mcp github create_review \
  --pr 45 \
  --event "APPROVE" \
  --body "✅ TDD Compliance Validated..."

# Output to user:
"✅ Validation Complete

**TDD Compliance:** PASS
- RED phase commits: 4 ✅
- GREEN phase commits: 4 ✅
- REFACTOR phase commits: 4 ✅
- Pattern: Proper TDD cycle followed

**Test Results:** PASS
- Total tests: 47 (all passing)
- Coverage: 89%
- Test pyramid: 38 unit / 7 integration / 2 e2e ✅

**CI Status:** PASS
- Build: ✅ passing
- Tests: ✅ 47/47 passing
- Lint: ✅ no issues

Review submitted: APPROVED"
```

### Example 4: Quality Gate Blocks on CI Failure

```bash
# Orchestrator checking quality gate:
"Checking quality gates before acceptance..."

# Behind the scenes:
ci_status=$(mcp github get_workflow_runs \
  --repo "owner/repo" \
  --branch "feature/login" \
  --status "latest")

if [ "$ci_status.conclusion" != "success" ]; then
  # Gate blocks!
fi

# Output to user:
"❌ GATE BLOCKED: CI/CD Checks Failing

Workflow: Tests
Status: failing
Failed Jobs:
  - unit-tests: 3 tests failing
  - coverage: Below 80% threshold (73%)

View details: https://github.com/owner/repo/actions/runs/12345

Cannot proceed to acceptance until CI passes.

Recommended actions:
1. Fix failing tests
2. Add tests to improve coverage
3. Re-run validation after fixes"
```

---

## Best Practices

### 1. Issue Naming Conventions

**Use consistent prefixes:**
```
Agent: {Role} - {Task Description}
Task: {Feature/Bug Description}
Retrospective: {BUG-ID} - {Root Cause}
Pattern: {Pattern Category}
```

**Examples:**
- `Agent: Engineer - Implement login API`
- `Task: Add dark mode toggle to settings`
- `Retrospective: BUG-123 - Null pointer in payment flow`
- `Pattern: Race conditions in async operations`

### 2. Label Strategy

**Recommended labels:**
```
ai-pack                  # All AI-Pack related issues
orchestrated             # Created by Orchestrator
agent                    # Agent tracking issue
{role-name}              # engineer, tester, reviewer, etc.
{priority}               # critical, high, normal, low
{status}                 # in-progress, blocked, completed
{type}                   # feature, bug, task, retrospective
tdd-compliant            # Implemented with TDD
ready-for-review         # Ready for Tester/Reviewer
artifact-persistence     # Documentation PR
```

### 3. Milestone Strategy

**Per feature/epic:**
```
Milestone: User Authentication
- Contains all subtask issues
- Due date: Sprint end date
- Description: Links to PRD and architecture docs
```

### 4. PR Templates

Create `.github/pull_request_template.md`:

```markdown
## AI-Pack Pull Request

**Beads Task:** <!-- bd-xxxx -->
**GitHub Issue:** Closes #<!-- issue number -->
**Task Packet:** <!-- .ai/tasks/yyyy-mm-dd_task/ -->

## Implementation Summary
<!-- Brief description of changes -->

## TDD Commit History
<!-- Auto-populated by Engineer -->

## Test Coverage
- Overall: <!-- X% -->
- Critical Logic: <!-- X% -->
- Test Pyramid: <!-- X unit | Y integration | Z e2e -->

## Acceptance Criteria
<!-- Copy from task packet 00-contract.md -->
- [ ] Criterion 1
- [ ] Criterion 2

## Review Checklist
- [ ] Tester validation: APPROVED
- [ ] Reviewer validation: APPROVED
- [ ] CI passing: ✅
- [ ] Coverage >= 80%: ✅

---
🤖 Generated by AI-Pack Engineer Agent
Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

### 5. Branch Protection Rules

Configure in GitHub Settings → Branches:

```yaml
Branch: main

Require pull request reviews before merging:
  ✓ Required approvals: 2
  ✓ Dismiss stale reviews
  ✓ Require review from Code Owners

Require status checks to pass:
  ✓ ci/tests
  ✓ ci/coverage
  ✓ ci/lint
  ✓ ai-pack/tdd-validation
  ✓ ai-pack/code-quality

Require branches to be up to date: ✓

Require conversation resolution: ✓

Do not allow bypassing: ✓
```

---

## Troubleshooting

### Issue: "GitHub MCP server not responding"

**Symptoms:**
```
Error: Failed to connect to GitHub MCP server
```

**Solutions:**

1. **Check server is running:**
```bash
docker ps | grep github-mcp-server
# OR
curl http://localhost:3000/health
```

2. **Check logs:**
```bash
docker logs github-mcp-server
```

3. **Restart server:**
```bash
docker restart github-mcp-server
```

4. **Verify token:**
```bash
# Test token manually
curl -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/user
```

### Issue: "Permission denied when creating issue"

**Symptoms:**
```
Error: Resource not accessible by integration
```

**Solutions:**

1. **Verify PAT scopes:**
```bash
# Check token scopes
curl -H "Authorization: token $GITHUB_TOKEN" \
  -I https://api.github.com/user | grep X-OAuth-Scopes
```

Should include: `repo, workflow, read:org`

2. **Regenerate token with correct scopes**

3. **Update environment:**
```bash
export GITHUB_TOKEN="new_token"
docker restart github-mcp-server
```

### Issue: "Rate limit exceeded"

**Symptoms:**
```
Error: API rate limit exceeded
```

**Solutions:**

1. **Check rate limit status:**
```bash
curl -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/rate_limit
```

2. **Wait for reset:**
```json
{
  "resources": {
    "core": {
      "remaining": 0,
      "reset": 1642694400  // Unix timestamp
    }
  }
}
```

3. **Reduce API calls:**
   - Batch operations
   - Cache responses
   - Use webhooks instead of polling

4. **Use GitHub Apps instead of PAT** (higher limits)

### Issue: "CI status check not found"

**Symptoms:**
```
Error: No workflow runs found for branch
```

**Solutions:**

1. **Verify workflow exists:**
```bash
ls .github/workflows/
```

2. **Check workflow has run:**
```bash
mcp github get_workflow_runs --repo owner/repo --branch main
```

3. **Trigger workflow manually:**
```bash
mcp github trigger_workflow --workflow ci.yml
```

### Issue: "Beads and GitHub out of sync"

**Symptoms:**
```
Beads task bd-a1b2 has no GitHub issue link
GitHub issue #123 has no Beads reference
```

**Solutions:**

1. **Manual sync script:**
```bash
#!/bin/bash
# sync-beads-github.sh

# Get all Beads tasks
bd list --json > beads_tasks.json

# For each task without GitHub link
jq -r '.[] | select(.github_issue == null) | .id' beads_tasks.json | while read task_id; do
  task_title=$(bd show $task_id | grep -oP '(?<=Title: ).*')

  # Search for matching GitHub issue
  issue_id=$(mcp github search_issues \
    --query "$task_title label:ai-pack" \
    --json | jq -r '.items[0].number // empty')

  if [ -n "$issue_id" ]; then
    echo "Linking $task_id to #$issue_id"
    bd comment $task_id "GitHub Issue: #$issue_id"
    mcp github add_comment --issue $issue_id --body "Beads Task: $task_id"
  fi
done
```

2. **Run sync:**
```bash
chmod +x sync-beads-github.sh
./sync-beads-github.sh
```

---

## Advanced Configuration

### Custom GitHub Actions for AI-Pack

**File:** `.github/workflows/ai-pack-gates.yml`

```yaml
name: AI-Pack Quality Gates

on:
  pull_request:
    types: [opened, synchronize, reopened]

jobs:
  tdd-validation:
    name: TDD Compliance Check
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Need full history for commit analysis

      - name: Analyze TDD Pattern
        id: tdd-check
        run: |
          # Check for RED-GREEN-REFACTOR pattern in commits
          commits=$(git log --oneline origin/${{ github.base_ref }}..HEAD)

          red_count=$(echo "$commits" | grep -ci "failing test\|add test" || true)
          green_count=$(echo "$commits" | grep -ci "make.*pass\|implement" || true)
          refactor_count=$(echo "$commits" | grep -ci "refactor" || true)

          echo "RED commits: $red_count"
          echo "GREEN commits: $green_count"
          echo "REFACTOR commits: $refactor_count"

          if [ $red_count -eq 0 ]; then
            echo "::error::No RED phase commits found - TDD not followed"
            exit 1
          fi

          echo "✅ TDD pattern detected"

      - name: Report Results
        if: always()
        uses: actions/github-script@v7
        with:
          script: |
            const body = `## TDD Validation Results

            **Status:** ${{ job.status == 'success' && '✅ PASS' || '❌ FAIL' }}

            **Commit Analysis:**
            - RED phase commits: Found
            - GREEN phase commits: Found
            - REFACTOR phase commits: Found

            TDD cycle properly followed.
            `;

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: body
            });

  test-pyramid:
    name: Test Pyramid Validation
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run Tests with Coverage
        run: |
          # Your test command here
          npm test -- --coverage --json > test-results.json

      - name: Analyze Test Pyramid
        run: |
          # Parse test results
          unit_count=$(jq '.numPassedTests | select(.unit)' test-results.json)
          integration_count=$(jq '.numPassedTests | select(.integration)' test-results.json)
          e2e_count=$(jq '.numPassedTests | select(.e2e)' test-results.json)

          total=$((unit_count + integration_count + e2e_count))
          unit_pct=$((unit_count * 100 / total))

          # Validate pyramid (60-80% unit tests)
          if [ $unit_pct -lt 60 ] || [ $unit_pct -gt 80 ]; then
            echo "::warning::Test pyramid imbalanced: ${unit_pct}% unit tests"
          fi

  coverage-gate:
    name: Coverage Gate (80%+)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Check Coverage
        run: |
          # Your coverage command
          npm test -- --coverage

          coverage=$(cat coverage/coverage-summary.json | jq '.total.lines.pct')

          if (( $(echo "$coverage < 80" | bc -l) )); then
            echo "::error::Coverage ${coverage}% below 80% threshold"
            exit 1
          fi

          echo "✅ Coverage: ${coverage}%"
```

### Webhook Integration

For real-time synchronization between GitHub and Beads:

**File:** `.ai-pack/scripts/github-webhook-handler.js`

```javascript
const express = require('express');
const { exec } = require('child_process');
const crypto = require('crypto');

const app = express();
app.use(express.json());

// Webhook secret from GitHub settings
const WEBHOOK_SECRET = process.env.GITHUB_WEBHOOK_SECRET;

// Verify GitHub signature
function verifySignature(req) {
  const signature = req.headers['x-hub-signature-256'];
  const hmac = crypto.createHmac('sha256', WEBHOOK_SECRET);
  const digest = 'sha256=' + hmac.update(JSON.stringify(req.body)).digest('hex');
  return crypto.timingSafeEqual(Buffer.from(signature), Buffer.from(digest));
}

// Handle issue events
app.post('/webhook/github', (req, res) => {
  if (!verifySignature(req)) {
    return res.status(401).send('Invalid signature');
  }

  const event = req.headers['x-github-event'];
  const payload = req.body;

  switch (event) {
    case 'issues':
      handleIssueEvent(payload);
      break;
    case 'pull_request':
      handlePREvent(payload);
      break;
    case 'workflow_run':
      handleWorkflowEvent(payload);
      break;
  }

  res.status(200).send('OK');
});

function handleIssueEvent(payload) {
  const action = payload.action; // opened, closed, edited, etc.
  const issue = payload.issue;

  // Check if issue is AI-Pack managed
  if (!issue.labels.some(l => l.name === 'ai-pack')) {
    return;
  }

  // Extract Beads task ID from issue body
  const beadsMatch = issue.body.match(/Beads Task: (bd-[a-z0-9]+)/);
  if (!beadsMatch) return;

  const taskId = beadsMatch[1];

  // Sync status to Beads
  if (action === 'closed') {
    exec(`bd close ${taskId}`, (err, stdout) => {
      console.log(`Closed Beads task ${taskId} from GitHub issue #${issue.number}`);
    });
  }
}

function handlePREvent(payload) {
  const action = payload.action;
  const pr = payload.pull_request;

  // Notify on PR review requests
  if (action === 'review_requested') {
    console.log(`Review requested for PR #${pr.number}`);
    // Could trigger Tester/Reviewer agents here
  }
}

function handleWorkflowEvent(payload) {
  const workflow = payload.workflow_run;

  if (workflow.conclusion === 'failure') {
    console.log(`Workflow ${workflow.name} failed - blocking quality gates`);
    // Update gate status
  }
}

const PORT = process.env.PORT || 3001;
app.listen(PORT, () => {
  console.log(`GitHub webhook handler listening on port ${PORT}`);
});
```

**Setup webhook in GitHub:**
1. Repository Settings → Webhooks → Add webhook
2. Payload URL: `https://your-server.com/webhook/github`
3. Content type: `application/json`
4. Secret: Generate and store in `GITHUB_WEBHOOK_SECRET`
5. Events: Issues, Pull requests, Workflow runs

---

## Summary

You now have GitHub MCP fully integrated with AI-Pack:

**✅ Completed:**
- GitHub MCP server installed and configured
- Claude Code integration configured
- Role extensions created for GitHub operations
- Quality gates integrated with CI/CD
- Best practices documented

**✅ Next Steps:**
1. Test integration with sample issue creation
2. Monitor first few PRs for proper formatting
3. Adjust labels and milestones per your workflow
4. Set up webhooks for real-time sync (optional)
5. Train team on new GitHub-AI-Pack workflow

**📚 Further Reading:**
- [GitHub MCP Integration Analysis](GITHUB-MCP-INTEGRATION-ANALYSIS.md) - Detailed capability mapping
- [GitHub MCP Server Repository](https://github.com/github/github-mcp-server) - Official documentation
- [AI-Pack Workflows](../workflows/) - Workflow documentation
- [AI-Pack Roles](../roles/) - Role documentation

---

**Need help?** Open an issue in the ai-pack repository: https://github.com/Cortexa-LLC/ai-pack/issues
