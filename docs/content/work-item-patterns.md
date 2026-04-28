---
sidebar_position: 3
title: "Work Item Patterns"
---

# Work Item Patterns: Epics, Stories, Tasks, Spikes, and Issues

**Version:** 1.0.0
**Date:** 2026-01-18
**Purpose:** Define standard patterns for work items across AI-Pack, Beads, and GitHub

## Overview

AI-Pack supports multiple types of work items that serve different purposes in the development workflow. This document defines how these work items map across three systems:

1. **Beads** - Local task memory and state tracking
2. **GitHub Issues** - Team visibility and collaboration
3. **Task Packets** - Detailed implementation artifacts

---

## Work Item Hierarchy

```
Epic (Large initiative, multiple sprints)
  └─ Story (User-facing feature, 1-5 days)
      ├─ Task (Implementation work, < 1 day)
      ├─ Spike (Research/Investigation, timeboxed)
      └─ Issue (Bug, defect, technical debt)
```

---

## 1. Epic

**Definition:** Large body of work that can be broken down into multiple Stories. Typically spans multiple sprints or weeks.

**When to use:**
- Major feature development (e.g., "User Authentication System")
- Large refactoring efforts (e.g., "Migrate to new API framework")
- Cross-cutting initiatives (e.g., "Improve test coverage to 90%")

### In Beads

```bash
# Create epic
epic_id=$(agent create "User Authentication System" --priority high --json | jq -r '.id')

# Create dependent stories
agent create "Design auth API" --depends-on ${epic_id}
agent create "Implement JWT tokens" --depends-on ${epic_id}
agent create "Add session management" --depends-on ${epic_id}
agent create "Write auth tests" --depends-on ${epic_id}
```

**Beads Representation:**
- Parent task with `--priority` set
- Multiple child tasks using `--depends-on`
- Use `agent show ${epic_id}` to view hierarchy

### In GitHub

```bash
# Sync epic to GitHub
python3 scripts/github-integration.py create-epic ${epic_id}
```

**GitHub Representation (2 options):**

**GitHub Representation:** ✅
- **Epic Issue** with label `epic`
- Title: "Epic: User Authentication System"
- Body: Task ID, description, checklist of Stories
- Story Issues linked to Epic
- Simple, lightweight, works everywhere

**Story Issues** (automatically created):
- Each dependent task becomes a Story Issue
- Label: `story`
- Links back to Epic
- Links to Beads: "**Beads Task:** bd-x1y2"

**Theme-Level Organization (Optional):**
For organizing multiple epics into a theme, manually create a GitHub Project:
```
GitHub Project: "Q1 Product Features" (Manual theme)
  ├── Epic Issue #42: User Authentication
  │   ├── Story Issue #43: Design auth API
  │   ├── Story Issue #44: Implement JWT
  │   └── Story Issue #45: Add tests
  ├── Epic Issue #50: Payment Processing
  │   └── Story Issue #51: Stripe integration
  └── Epic Issue #60: Analytics Dashboard
      └── Story Issue #61: Metrics visualization
```

**Configure epic naming in `.github-integration.yml`:**
```yaml
epics:
  naming_pattern: "Epic: {title}"
```

### In Task Packets

Epic-level task packet (optional):
```
.ai/tasks/local-20260118090000-user-authentication-system/
├── 00-contract.md         # Epic scope and acceptance criteria
├── 10-plan.md             # Epic breakdown and architecture
└── 20-work-log.md         # Epic-level progress tracking
```

**Best Practice:** Create task packets for Stories, not for the Epic itself. Epic serves as organizational container.

---

## 2. Story

**Definition:** User-facing feature or capability that delivers value. Should be completable within 1-5 days.

**When to use:**
- User-facing features
- API endpoints
- UI components
- End-to-end workflows

**Format:** User story format recommended
- "As a [user], I want [capability], so that [benefit]"
- Example: "As a developer, I want JWT authentication, so that API access is secure"

### In Beads

```bash
# Create story (part of epic)
story_id=$(agent create "Implement JWT tokens" --depends-on ${epic_id} --priority high --json | jq -r '.id')

# Start work
bd start ${story_id}

# Track progress
bd comment ${story_id} "Implemented token generation"
bd comment ${story_id} "Added token validation"

# Complete
agent close ${story_id}
```

### In GitHub

**Created automatically** when running `create-epic`, or manually:

```bash
# Export individual story
python3 scripts/github-integration.py export
```

**GitHub Issue:**
- Label: `story`
- Title: Story description
- Body: Includes epic reference, Task ID, acceptance criteria
- Linked to Epic issue

### In Task Packets

**MANDATORY** - Every story must have a task packet:

```
.ai/tasks/local-20260118090000-implement-jwt-tokens/
├── 00-contract.md         # Requirements, acceptance criteria
├── 10-plan.md             # Implementation approach
├── 20-work-log.md         # Daily progress updates
├── 30-review.md           # Tester/Reviewer feedback
└── 40-acceptance.md       # Final sign-off
```

**Create with:**
```bash
/ai-pack task-init implement-jwt-tokens
```

---

## 3. Task

**Definition:** Granular implementation work. Should be completable in < 1 day, typically 2-4 hours.

**When to use:**
- Specific implementation work within a Story
- Subtasks that can be tracked independently
- Work that can be parallelized

**Examples:**
- "Write token generation function"
- "Add token validation middleware"
- "Update API documentation"

### In Beads

```bash
# Create task (part of story)
task_id=$(agent create "Write token generation function" --depends-on ${story_id} --json | jq -r '.id')

# Work on task
bd start ${task_id}
# ... implement ...
agent close ${task_id}
```

**Pattern for Stories with multiple Tasks:**

```bash
# Story
story_id=$(agent create "Implement JWT tokens" --priority high --json | jq -r '.id')

# Tasks
agent create "Write token generation function" --depends-on ${story_id}
agent create "Add token validation middleware" --depends-on ${story_id}
agent create "Add token refresh endpoint" --depends-on ${story_id}
agent create "Write token tests" --depends-on ${story_id}
```

### In GitHub

**Option 1:** Don't sync tasks to GitHub (keep lightweight)
- Tasks tracked only in Beads
- Story issue represents all tasks
- Use exclude patterns:
  ```yaml
  sync_rules:
    beads_to_github:
      exclude_patterns:
        - "^Write "
        - "^Add "
        - "^Update "
  ```

**Option 2:** Sync tasks as subtask issues
- Label: `task`
- Reference parent Story
- Useful for distributed teams

### In Task Packets

Tasks typically don't get separate task packets. They're tracked within the Story's task packet:

```markdown
## 20-work-log.md

### 2026-01-18 10:00 - Token Generation Function

**Task:** Write token generation function
**Status:** Complete
**Changes:**
- Implemented generateJWT() function
- Added expiration logic
- Added signing with secret key
```

---

## 4. Spike

**Definition:** Timeboxed research or investigation to answer a question or explore a solution. Produces knowledge, not production code.

**When to use:**
- Evaluating technology choices
- Investigating feasibility
- Understanding complex systems
- Prototyping approaches

**Key characteristic:** Timeboxed (usually 1-4 hours, max 1 day)

### In Beads

```bash
# Create spike
spike_id=$(agent create "Spike: Evaluate JWT libraries" --priority normal --json | jq -r '.id')

# Add timebox constraint
bd comment ${spike_id} "Timebox: 4 hours"
bd comment ${spike_id} "Goal: Recommend JWT library for Node.js"

# Start spike
bd start ${spike_id}

# Document findings
bd comment ${spike_id} "Evaluated: jsonwebtoken, jose, node-jose"
bd comment ${spike_id} "Recommendation: jsonwebtoken (most popular, well-maintained)"

# Complete
agent close ${spike_id}
```

**Naming convention:** Prefix with "Spike: " to indicate research nature

### In GitHub

```yaml
# Optionally sync spikes
sync_rules:
  beads_to_github:
    exclude_patterns:
      - "^Spike:"  # Don't sync spikes (keep internal)
```

**If synced:**
- Label: `spike`, `research`
- Indicate timebox in title: "Spike: Evaluate JWT libraries (4h)"

### In Task Packets

**Optional** - Use investigation template:

```
docs/investigations/2026-01-18_jwt-library-evaluation.md
```

**Or** research workflow:
```bash
# Use research workflow for spikes
# See: workflows/research.md
```

**Spike Output:**
- Recommendation document
- Prototype code (not for production)
- Architectural decision record (ADR)

---

## 5. Issue (Bug/Defect)

**Definition:** Problem with existing functionality. Something that doesn't work as intended or expected.

**When to use:**
- Bugs reported by users
- Test failures
- Defects found during review
- Production incidents

### In Beads

```bash
# Create bug
bug_id=$(agent create "Fix: Login fails with special characters in password" --priority critical --json | jq -r '.id')

# Start work
bd start ${bug_id}

# Track investigation
bd comment ${bug_id} "Root cause: Password not properly URL-encoded"
bd comment ${bug_id} "Fix: Add encodeURIComponent() before sending"

# Complete
agent close ${bug_id}
```

**Naming convention:** Prefix with "Fix: " for clarity

### In GitHub

**Import bugs from GitHub:**
```bash
# Import issues labeled "bug"
python3 scripts/github-integration.py import
```

**Export bugs to GitHub:**
```bash
# Bugs sync automatically if priority matches rules
python3 scripts/github-integration.py export
```

**GitHub Issue:**
- Label: `bug`, priority label
- Template: Bug report template
- Links to task

### In Task Packets

Use **bugfix workflow**:

```
.ai/tasks/local-20260118090000-fix-login-special-chars/
├── 00-contract.md         # Bug description, reproduction steps
├── 10-plan.md             # Investigation and fix approach
├── 20-work-log.md         # Investigation notes, fix implementation
├── 30-review.md           # Regression testing verification
└── 40-acceptance.md       # Bug verified fixed
```

**Create with:**
```bash
/ai-pack task-init fix-login-special-chars
# Then follow workflows/bugfix.md
```

---

## Workflow Integration

### Orchestrator Role

**Creating work hierarchy:**

```bash
# 1. Create Epic in Beads
epic_id=$(agent create "User Authentication System" --priority high --json | jq -r '.id')

# 2. Break down into Stories
agent create "Design auth API" --depends-on ${epic_id}
agent create "Implement JWT tokens" --depends-on ${epic_id}
agent create "Add session management" --depends-on ${epic_id}

# 3. Optionally break Stories into Tasks (done by Engineer)

# 4. Sync to GitHub for team visibility
python3 scripts/github-integration.py create-epic ${epic_id}
```

### Engineer Role

**Working on Story:**

```bash
# 1. Find ready work
agent list --status queued

# 2. Start Story
bd start bd-x1y2

# 3. Create task packet
/ai-pack task-init implement-jwt-tokens

# 4. Optionally break into Tasks in Beads
agent create "Write token generation" --depends-on bd-x1y2
agent create "Add validation middleware" --depends-on bd-x1y2

# 5. Implement with TDD
# ... work ...

# 6. Complete
agent close bd-x1y2

# 7. GitHub automatically updates via sync
```

---

## GitHub Sync Patterns

### Pattern 1: Epic/Story Sync Only

**Use case:** Solo developer or small team

**Configuration:**
```yaml
sync_rules:
  beads_to_github:
    exclude_patterns:
      - "^Write "
      - "^Add "
      - "^Update "
      - "^Spike:"
```

**Result:**
- Epics and Stories sync to GitHub
- Tasks and Spikes stay in Beads
- Lightweight GitHub visibility

### Pattern 2: Full Sync

**Use case:** Distributed team, full traceability needed

**Configuration:**
```yaml
sync_rules:
  beads_to_github:
    statuses: ["open", "in_progress", "blocked"]
    priorities: ["critical", "high", "normal"]
    exclude_patterns:
      - "^Agent:"
      - "^Internal:"
```

**Result:**
- All work items sync to GitHub
- Full team visibility
- Use labels to distinguish: epic, story, task, spike, bug

### Pattern 3: Bugs Only

**Use case:** Public issue tracker

**Configuration:**
```yaml
sync_rules:
  beads_to_github:
    exclude_patterns:
      - ".*"  # Exclude everything
  github_to_beads:
    required_labels: ["bug"]
```

**Result:**
- Only import bugs from GitHub
- Internal work stays in Beads
- Public can report bugs via GitHub

---

## Labeling Strategy

### Beads → GitHub Label Mapping

Configure in `.github-integration.yml`:

```yaml
labels:
  # Work item types
  epic_label: "epic"
  story_label: "story"
  task_label: "task"

  # Custom type labels
  type_mapping:
    "^Spike:": "spike"
    "^Fix:": "bug"

  # Priority mapping
  priority_mapping:
    critical: "priority-critical"
    high: "priority-high"
    normal: "priority-normal"
    low: "priority-low"
```

---

## Best Practices

### 1. Beads is Source of Truth

**Always:**
- Create tasks in Beads first
- Use `bd` commands for state changes
- GitHub is a reflection, not the source

**Pattern:**
```bash
# ✅ Correct
agent create "Story"
python3 scripts/github-integration.py export

# ❌ Wrong
# Create in GitHub first, then try to import
```

### 2. Task Packets for Stories and Bugs

**Create task packets for:**
- ✅ Stories (always)
- ✅ Bugs (always)
- ⚠️ Epics (optional, usually not needed)
- ❌ Tasks (tracked in Story's packet)
- ❌ Spikes (use investigation docs instead)

### 3. Use Dependencies for Hierarchy

```bash
# Epic → Stories → Tasks
epic_id=$(agent create "Epic")
story_id=$(agent create "Story" --depends-on ${epic_id})
task_id=$(agent create "Task" --depends-on ${story_id})
```

### 4. Meaningful Naming

**Good:**
- "Epic: User Authentication System"
- "Implement JWT token generation"
- "Fix: Login fails with special characters"
- "Spike: Evaluate JWT libraries (4h)"

**Bad:**
- "Auth stuff"
- "Fix bug"
- "Research"
- "TODO"

### 5. Timebox Spikes

Always add timebox to spikes:
```bash
bd comment ${spike_id} "Timebox: 4 hours"
bd comment ${spike_id} "Goal: Choose between Library A or B"
```

---

## Example: Full Epic Workflow

### 1. Orchestrator: Plan Epic

```bash
# Create epic
epic_id=$(agent create "User Authentication System" --priority high --json | jq -r '.id')

# Break into stories
agent create "Design auth API" --depends-on ${epic_id} --json
agent create "Implement JWT tokens" --depends-on ${epic_id} --json
agent create "Add password hashing" --depends-on ${epic_id} --json
agent create "Write auth tests" --depends-on ${epic_id} --json

# Run spike first if needed
agent create "Spike: Evaluate JWT libraries (4h)" --depends-on ${epic_id} --json

# Sync to GitHub
python3 scripts/github-integration.py create-epic ${epic_id}
```

**Result in GitHub:**
- Epic Issue #42 with checklist of stories
- Story Issues #43, #44, #45, #46 linked to epic
- Spike not synced (excluded by pattern)
- Full bidirectional traceability maintained

**Optional:** Manually add epic issues to a GitHub Project for theme-level organization.

### 2. Engineer: Work on Story

```bash
# Find ready work
agent list --status queued
# Shows: bd-x1y2 "Implement JWT tokens"

# Start story
bd start bd-x1y2

# Create task packet
/ai-pack task-init implement-jwt-tokens

# Break into tasks (optional)
agent create "Write token generation" --depends-on bd-x1y2
agent create "Add validation middleware" --depends-on bd-x1y2
agent create "Write token tests" --depends-on bd-x1y2

# Implement with TDD
bd start bd-a1b2  # "Write token generation"
# ... RED-GREEN-REFACTOR ...
agent close bd-a1b2

# Continue with other tasks...

# Complete story
agent close bd-x1y2
```

**GitHub automatically updates:**
- Issue #43 status updated
- Story issue shows progress

### 3. Tester: Validate

```bash
/ai-pack test
# Reviews test coverage, TDD compliance
# Updates .ai/tasks/.../30-review.md
```

### 4. Reviewer: Code Review

```bash
/ai-pack review
# Reviews code quality, standards compliance
# Updates .ai/tasks/.../30-review.md
```

### 5. Complete Story

```bash
# Sign off in acceptance document
# .ai/tasks/.../40-acceptance.md
```

**GitHub:**
- Story issue closed
- Epic checklist item checked ✓

### 6. Repeat for Remaining Stories

Continue until all stories complete.

### 7. Close Epic

```bash
agent close ${epic_id}
```

**GitHub:**
- Epic issue closed
- All stories verified complete

---

## Summary

| Work Item | Beads | GitHub | Task Packet | Typical Duration |
|-----------|-------|--------|-------------|------------------|
| **Epic** | Parent task | Issue with checklist | Optional | Weeks to months |
| **Story** | Task with dependencies | Issue (label: story) | Required | 1-5 days |
| **Task** | Dependent task | Optional issue | No (in Story's packet) | 2-4 hours |
| **Spike** | Task (prefix "Spike:") | Usually excluded | Investigation doc | 1-4 hours (timeboxed) |
| **Bug** | Task (prefix "Fix:") | Issue (label: bug) | Required (bugfix workflow) | Hours to days |

**Note:** Epic issues can be manually added to GitHub Projects for theme-level organization.

**Hierarchy:**
```
Theme (GitHub Project) - Optional, contains multiple epics
  └─ Epic (GitHub Issue) - Contains stories
      └─ Story (GitHub Issue) - Contains tasks
          └─ Task (Beads only) - Implementation work
```

**Key Principles:**
1. ✅ Beads is the source of truth for task state
2. ✅ GitHub provides team visibility and collaboration
3. ✅ Task Packets capture implementation details
4. ✅ Use dependencies to model hierarchy
5. ✅ Sync patterns control GitHub visibility

---

**See Also:**
- [Beads Enforcement Gate](./gates/beads-enforcement)
- [Orchestrator Role](./roles/orchestrator)
- [Engineer Role](./roles/engineer)
- [Workflows](./workflows/bugfix)
