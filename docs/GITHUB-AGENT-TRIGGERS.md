# GitHub Integration: Agent/Role Triggers

## Overview

GitHub integration can be triggered automatically when AI agents perform actions, not just from manual user commands. This provides seamless integration where work in Beads automatically syncs to GitHub without explicit sync commands.

## How It Works

### Detection

The integration detects agent/role context from:
1. **task metadata** - Role assignee (e.g., `assignee: "Orchestrator"`)
2. **Task naming patterns** - Prefix patterns (e.g., `"SEC:"`, `"Epic:"`)
3. **Command context** - Which role is executing the command
4. **Work-log entries** - Role signatures in work-logs

### Automatic Triggers

When enabled, the following agent actions automatically trigger GitHub operations:

## Role-Based Triggers

### Orchestrator

**Epic Creation** (`orchestrator.epic_creation: true`)
```bash
# When Orchestrator creates epic in Beads:
agent create "Epic: User Authentication System" --assignee Orchestrator

# Automatically triggers:
${AI_PACK_ROOT}/scripts/github-integration.py create-epic <task-id>
```

**Work Breakdown** (`orchestrator.work_breakdown: true`)
```bash
# When Orchestrator breaks down epic into stories:
agent create "Story: Implement JWT" --depends-on bd-epic-123

# Automatically syncs to GitHub as story issues
```

**What Happens:**
- Epic created as GitHub Issue with label `epic`
- Stories created as GitHub Issues with label `story`
- Issues linked in epic checklist
- All synchronized automatically

---

### Program Manager

**Issue Creation** (`program_manager.issue_creation: true`)
```bash
# When PM creates tracking tasks:
agent create "Track Q1 OKRs" --assignee "Program-Manager"

# Automatically creates GitHub issue
```

**Milestone Updates** (`program_manager.milestone_updates: true`)
```bash
# When PM updates task with milestone info:
agent update bd-task-123 --metadata milestone="Q1-2026"

# Automatically syncs milestone to GitHub
```

---

### Security Role

**SEC Issue Creation** (`security.sec_issue_creation: true`)
```bash
# When Security role creates SEC tasks:
agent create "SEC: Investigation - SQL injection in auth" --assignee Security

# Automatically triggers:
# - Creates GitHub issue with labels: security, needs-review
# - Optionally makes issue private (org-level repos)
# - Assigns to security team
```

**Configuration:**
```yaml
agent_triggers:
  security:
    sec_issue_creation: true
    sec_labels:
      - "security"
      - "needs-review"
      - "vulnerability"
    sec_private: true  # Org repos only
    sec_assignees:
      - "security-team"
```

**What Happens:**
- Issue created with security labels
- Private visibility if supported
- Assigned to security team
- Links back to task

---

### Engineer

**Task Start** (`engineer.task_start: true`)
```bash
# When Engineer starts task:
bd start bd-story-123

# Automatically updates GitHub issue:
# - Adds "in-progress" label
# - Moves to "In Progress" column on Project boards
# - Comments: "🔨 Work started by Engineer"
```

**Task Complete** (`engineer.task_complete: true`)
```bash
# When Engineer completes task:
bd complete bd-story-123

# Automatically updates GitHub issue:
# - Adds "completed" label
# - Removes "in-progress" label
# - Comments: "✅ Completed by Engineer"
# - Checks off item in epic checklist
```

**Auto Draft PR** (`engineer.auto_draft_pr: false`)
```bash
# Optional: Auto-create draft PR when Engineer pushes to feature branch
# Disabled by default - enable if your workflow uses it
```

---

### Reviewer

**Review Comments** (`reviewer.review_comments: true`)
```bash
# When Reviewer completes code review:
# Automatically posts review results to GitHub PR comments

# Example comment posted:
# ✅ Code Review Complete
# - Clean code standards: PASS
# - Test coverage: 85% (threshold: 80%)
# - TDD compliance: PASS
#
# Approved by: Reviewer
# task: bd-review-456
```

**Auto Approve** (`reviewer.auto_approve: false`)
```bash
# Optional: Auto-approve PR if all checks pass
# Disabled by default - use with caution
```

---

### Tester

**Bug Creation** (`tester.bug_creation: true`)
```bash
# When Tester creates bug report:
agent create "Fix: Login fails with special characters" --assignee Tester

# Automatically creates GitHub issue:
# - Label: bug
# - Priority based on severity
# - Links to failed test case
```

**Bug Closure** (`tester.bug_closure: true`)
```bash
# When Tester marks bug as resolved:
bd complete bd-bug-789

# Automatically closes GitHub issue:
# - Comments: "✅ Verified fixed by Tester"
# - Closes issue
# - Updates related epic checklist
```

---

### All Roles

**Status Change Sync** (`all_roles.status_change_sync: true`)
```bash
# Any role changes task status:
bd start <task>    # → Issue labeled "in-progress"
bd block <task>    # → Issue labeled "blocked"
bd complete <task> # → Issue marked complete

# Automatic bidirectional sync
```

**Work-log Sync** (`all_roles.worklog_sync: false`)
```bash
# Optional: Sync work-log entries as issue comments
# Disabled by default - can be noisy
```

## Configuration

Enable agent triggers in `${AI_PACK_ROOT}/.github-integration.yml`:

```yaml
features:
  agent_triggers:
    enabled: true

    orchestrator:
      epic_creation: true
      work_breakdown: true

    program_manager:
      issue_creation: true
      milestone_updates: true

    security:
      sec_issue_creation: true
      sec_labels:
        - "security"
        - "needs-review"
      sec_private: true

    engineer:
      task_start: true
      task_complete: true
      auto_draft_pr: false

    reviewer:
      review_comments: true
      auto_approve: false

    tester:
      bug_creation: true
      bug_closure: true

    all_roles:
      status_change_sync: true
      worklog_sync: false
```

## Implementation

### Beads Hooks

Agent triggers use Beads hooks to detect actions:

**`.beads/hooks/post-create`**
```bash
#!/bin/bash
# Triggered after task creation

TASK_ID="$1"
TITLE=$(agent show "$TASK_ID" --format=json | jq -r '.title')
ASSIGNEE=$(agent show "$TASK_ID" --format=json | jq -r '.assignee')

# Check if agent triggers enabled
if ! grep -q "agent_triggers.enabled: true" ${AI_PACK_ROOT}/.github-integration.yml; then
  exit 0
fi

# Orchestrator epic creation
if [[ "$TITLE" == "Epic:"* ]] && [[ "$ASSIGNEE" == "Orchestrator" ]]; then
  if grep -q "orchestrator.epic_creation: true" ${AI_PACK_ROOT}/.github-integration.yml; then
    ${AI_PACK_ROOT}/scripts/github-integration.py create-epic "$TASK_ID" &
  fi
fi

# Security SEC issue creation
if [[ "$TITLE" == "SEC:"* ]] && [[ "$ASSIGNEE" == "Security" ]]; then
  if grep -q "security.sec_issue_creation: true" ${AI_PACK_ROOT}/.github-integration.yml; then
    ${AI_PACK_ROOT}/scripts/github-integration.py create-security-issue "$TASK_ID" &
  fi
fi
```

**`.beads/hooks/post-status-change`**
```bash
#!/bin/bash
# Triggered after status change

TASK_ID="$1"
OLD_STATUS="$2"
NEW_STATUS="$3"

# Check if status change sync enabled
if grep -q "all_roles.status_change_sync: true" ${AI_PACK_ROOT}/.github-integration.yml; then
  ${AI_PACK_ROOT}/scripts/github-integration.py sync-status "$TASK_ID" "$NEW_STATUS" &
fi
```

### Manual Override

Users can always manually trigger operations:

```bash
# Even with auto-triggers enabled, manual commands work
${AI_PACK_ROOT}/scripts/github-integration.py create-epic bd-epic-123
${AI_PACK_ROOT}/scripts/github-integration.py sync
${AI_PACK_ROOT}/scripts/github-integration.py export
```

## Use Cases

### 1. Orchestrator Creates Epic

```bash
# Orchestrator: Break down Q1 feature
agent create "Epic: User Authentication System" --assignee Orchestrator
agent create "Story: JWT implementation" --depends-on bd-epic-123 --assignee Engineer
agent create "Story: Password reset flow" --depends-on bd-epic-123 --assignee Engineer

# With agent_triggers.orchestrator.epic_creation: true
# Automatically creates:
# - GitHub Issue #100: Epic: User Authentication System [epic]
# - GitHub Issue #101: Story: JWT implementation [story]
# - GitHub Issue #102: Story: Password reset flow [story]
# - Issues #101 and #102 linked in #100 checklist
```

### 2. Security Investigation

```bash
# Security: Found vulnerability
agent create "SEC: SQL injection in user search" --assignee Security --priority critical

# With agent_triggers.security.sec_issue_creation: true
# Automatically creates:
# - GitHub Issue #103 [security, needs-review, critical]
# - Private visibility (if org repo)
# - Assigned to @security-team
# - Comments: "Created from task bd-sec-456"
```

### 3. Engineer Workflow

```bash
# Engineer: Start working on story
bd start bd-story-101

# With agent_triggers.engineer.task_start: true
# Automatically updates GitHub Issue #101:
# - Label: in-progress
# - Comment: "🔨 Work started"
# - Project board: Moved to "In Progress"

# Engineer: Complete story
bd complete bd-story-101

# With agent_triggers.engineer.task_complete: true
# Automatically updates GitHub Issue #101:
# - Label: completed
# - Removes: in-progress
# - Comment: "✅ Completed"
# - Epic checklist: Checked off
```

### 4. Tester Finds Bug

```bash
# Tester: Test failure
agent create "Fix: Login button not responsive on mobile" --assignee Tester --priority high

# With agent_triggers.tester.bug_creation: true
# Automatically creates:
# - GitHub Issue #104 [bug, priority-high]
# - Links to failed test case
# - Assigned to Engineer
```

## Benefits

1. **Seamless Integration** - No manual sync commands needed
2. **Real-time Updates** - GitHub reflects current state immediately
3. **Role-Specific Behavior** - Different actions for different roles
4. **Audit Trail** - All actions logged and traceable
5. **Team Visibility** - Non-AI team members see updates in GitHub

## Best Practices

1. **Start Conservative** - Enable only triggers you need
2. **Monitor Initially** - Watch for unexpected behavior
3. **Test in Staging** - Verify triggers work as expected
4. **Document Custom Triggers** - If you add project-specific triggers
5. **Use Async Triggers** - Background jobs don't block Beads operations

## Troubleshooting

### Triggers Not Firing

```bash
# Check if enabled
grep "agent_triggers.enabled" ${AI_PACK_ROOT}/.github-integration.yml

# Check Beads hooks
ls -la .beads/hooks/

# Check hook permissions
chmod +x .beads/hooks/post-*

# Check logs
tail -f ${AI_PACK_ROOT}/.github-integration.log
```

### Too Many GitHub Operations

```bash
# Disable noisy triggers
# In .github-integration.yml:
all_roles:
  status_change_sync: false  # Reduce noise
  worklog_sync: false         # Disable work-log comments
```

### Security Issues Visible

```bash
# Ensure private issues enabled (org repos only)
security:
  sec_private: true

# Or exclude from sync
sync_rules:
  beads_to_github:
    exclude_patterns:
      - "^SEC:"  # Don't sync security issues
```

## See Also

- [GitHub Integration Setup](GITHUB-INTEGRATION-SETUP.md) - Installation paths
- [GitHub Integration Usage](GITHUB-INTEGRATION-USAGE.md) - Full feature guide
- [Work Item Patterns](WORK-ITEM-PATTERNS.md) - Epic/Story/Task hierarchy
- [Beads Hooks](../quality/tooling/beads-integration.md) - Hook system details
