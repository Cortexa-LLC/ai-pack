---
name: shepherd-pr
description: >
  Drive an open GitHub PR to merge-ready state by spawning the pr-shepherd agent.
  Handles CI failures, reviewer threads, and iterates until the PR is approved and
  all checks pass. Use when a PR needs to be nursed to a green, mergeable state
  without manual intervention.
  <example>shepherd PR #42 to merge-ready</example>
  <example>drive my PR to green</example>
  <example>fix the CI failures and reviewer comments on PR #15</example>
  <example>shepherd the current branch's PR until it's approved</example>
  <example>get my pull request merge-ready</example>
---

## Shepherd PR

This skill is a thin orchestrator. All iterative logic — watching CI, fixing
failures, resolving reviewer threads, and looping until merge-ready — lives in the
`pr-shepherd` agent role. This skill handles task-packet creation and spawning.

---

### Step 1 — Identify the PR

If the user supplied a PR number, use it directly. Otherwise, find the PR for
the current branch:

```bash
gh pr list --head "$(git branch --show-current)" --json number,title,url \
  | jq -r '.[] | "#\(.number)  \(.title)  \(.url)"'
```

If multiple PRs are returned, ask the user which one to shepherd.
If no PR is found, tell the user to open a PR first:

```bash
gh pr create --fill
```

Capture:
- `PR_NUMBER` — the PR number (integer)
- `REPO` — the `owner/repo` slug

```bash
REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner)
```

---

### Step 2 — Create the task packet

```bash
SLUG="shepherd-pr-${PR_NUMBER}"
TASK_DIR=".ai/tasks/${SLUG}-$(date +%Y%m%d%H%M%S)"
mkdir -p "$TASK_DIR"
```

Write `$TASK_DIR/task.md` with the following fields (the pr-shepherd role
requires all three):

```markdown
# PR Shepherd Contract

PR: <PR_NUMBER>
Repo: <REPO>
Working directory: <absolute path — output of `pwd`>

## Goal

Drive PR #<PR_NUMBER> in <REPO> to merge-ready:
- All CI checks green
- All reviewer threads resolved or acknowledged
- At least one approving review (if required by branch protection)

## Context

Branch: <git branch --show-current>
Base: <gh pr view <PR_NUMBER> --json baseRefName -q .baseRefName>
```

Include any additional context the user provided (e.g. specific CI jobs that are
failing, known reviewer concerns).

---

### Step 3 — Create and spawn the pr-shepherd task

```bash
TID=$(agent create "Shepherd PR #${PR_NUMBER} to merge-ready

Working directory: $(pwd)
Task packet: ${TASK_DIR}

PR: ${PR_NUMBER}
Repo: ${REPO}

Drive the PR to merge-ready: fix CI failures, resolve reviewer threads,
and loop until all checks pass and the PR is approved." \
  --json | jq -r '.id')

agent pr-shepherd "$TID" --stream
```

Stream the agent output so the user can see progress in real time.

---

### Step 4 — Report the result

When the pr-shepherd agent completes, read its output and report the outcome:

**If the PR is merge-ready:**

```
✅ PR #<PR_NUMBER> is merge-ready.
All CI checks passed and all reviewer threads are resolved.

To merge:
  gh pr merge <PR_NUMBER> --merge   # or --squash / --rebase
```

**If the agent halted before merge-ready:**

```
⚠️  PR shepherd halted before the PR was fully ready.

Reason: <paste the agent's final status>

Remaining issues:
<list unresolved CI failures or open reviewer threads>

Next steps:
<list any manual actions required>
```

---

### Notes

- The pr-shepherd agent operates on the remote PR, not the local working tree.
  Ensure `gh` is authenticated (`gh auth status`) before spawning.
- The agent will push commits to the PR branch. If the branch is protected (e.g.
  requires signed commits), the agent may not be able to push — note this in the
  contract if known.
- If the user wants to limit which reviewer threads are addressed (e.g. skip
  `[SUGGESTION]` threads), add that instruction to the contract in Step 2.
- If the PR requires manual approval from a human reviewer, the shepherd will
  complete all fixable work and then halt, noting that human approval is needed.
