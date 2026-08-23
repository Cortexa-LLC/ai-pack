---
name: pre-push
description: >
  Review and fix local commits before pushing to origin. Spawns a reviewer agent
  against the local diff, and if issues are found, spawns an engineer agent to fix
  them, amends the commit, then re-reviews. Loops until the reviewer approves or
  the iteration limit is reached. Use before `git push` to catch issues early.
  <example>review my commits before I push</example>
  <example>pre-push check</example>
  <example>check my changes before pushing to origin</example>
  <example>review and fix my local commits then tell me when it's safe to push</example>
---

## Pre-Push Shepherd

This skill runs a reviewer→engineer→amend loop against the local commits that have
not yet been pushed to `origin/main`. It halts when the reviewer approves or after
5 iterations, whichever comes first.

Agents are spawned natively with Claude Code's built-in `Agent` tool, using the
plugin's subagents (`ai-pack:reviewer`, `ai-pack:engineer`). Each spawn in this
loop depends on the previous step's result, so **always pass
`run_in_background: false`** — wait for each agent to finish before continuing.

Sub-agents have **no shared memory** with this conversation. Every prompt must be
fully self-contained: include the working directory, the diff scope, and complete
instructions in the prompt itself.

---

### Before You Start

Confirm the working repository and target branch:

```bash
git remote get-url origin          # confirm remote
git branch --show-current          # confirm current branch
git log --oneline origin/main..HEAD  # show commits to be reviewed
```

If there are no commits ahead of `origin/main`, stop and tell the user there is
nothing to push.

Set the iteration counter: `ITER = 1`.

---

### Step 1 — Compute the diff

```bash
BASE=$(git merge-base HEAD origin/main)
git diff "$BASE"..HEAD
```

Capture the full diff text. This is the code under review.

Also capture the commit list:

```bash
git log --oneline "$BASE"..HEAD
```

---

### Step 2 — Spawn the reviewer

Spawn the `ai-pack:reviewer` subagent with the `Agent` tool and
`run_in_background: false`. The prompt must be self-contained:

```text
Agent({
  subagent_type: "ai-pack:reviewer",
  description: "Pre-push review iteration <ITER>",
  run_in_background: false,
  prompt: "Review the local commits on branch <branch> in <absolute working
  directory> that have not been pushed to origin/main.

  Diff scope: run `git diff <BASE>..HEAD` (BASE = <merge-base SHA>) to obtain
  the exact diff under review. The commits in scope are:
  <paste output of git log --oneline BASE..HEAD>

  Review the diff for correctness, quality, security, and style. Return your
  results in your final message as:
  1. A verdict line: `Verdict: APPROVE`, `Verdict: REQUEST CHANGES`, or
     `Verdict: BLOCK`
  2. A numbered list of findings, each with file path, line reference, severity,
     and a concrete description of the problem and the expected fix.

  Do not modify any files. Do not push."
})
```

If the diff is small enough, paste it directly into the prompt instead of the
`git diff` instruction — the reviewer then needs zero exploration.

Wait for the agent to complete.

---

### Step 3 — Parse the verdict

Read the reviewer's final message. Look for one of:

- `Verdict: APPROVE`
- `Verdict: REQUEST CHANGES`
- `Verdict: BLOCK`

If the verdict is **APPROVE**:

```
✅ Reviewer approved. Ready to push.
Run: git push
```

Stop. Do not loop further.

---

### Step 4 — If verdict is REQUEST CHANGES or BLOCK

Extract the numbered findings from the reviewer's final message. These become the
brief for the engineer.

Spawn the `ai-pack:engineer` subagent with the `Agent` tool and
`run_in_background: false`, embedding the findings verbatim in the prompt:

```text
Agent({
  subagent_type: "ai-pack:engineer",
  description: "Fix pre-push findings iter <ITER>",
  run_in_background: false,
  prompt: "All context provided. Working directory: <absolute working directory>,
  branch <branch>.

  A pre-push code review of the local commits (diff scope:
  `git diff <BASE>..HEAD`) produced the findings below. Fix ALL of them.

  ## Reviewer findings
  <paste the numbered findings verbatim from the reviewer's final message>

  ## Constraints
  - The code is in the working directory; edit files in place.
  - Do NOT commit, amend, or push — only make and stage the changes
    (`git add -u` at most).
  - If a finding cannot be fixed (e.g. a design concern), skip it and say so
    explicitly in your final message.

  ## Acceptance criteria
  - Build and tests pass for the affected packages.
  - Every finding is either fixed or explicitly reported as not auto-fixable."
})
```

Wait for the engineer to complete. Note any findings it reported as not fixable.

---

### Step 5 — Amend the commit

After the engineer completes, amend the current commit to fold in the fixes.
**Do not create a new commit** — amending keeps the branch history clean.

```bash
git add -u
git commit --amend --no-edit
```

---

### Step 6 — Loop

Increment `ITER` and go back to **Step 1**.

```
ITER = ITER + 1
```

---

### Step 7 — Halting condition

If `ITER` exceeds **5**, stop the loop and print a summary:

```
⚠️  Pre-push shepherd reached the iteration limit (5 rounds).
The reviewer has not approved after 5 fix attempts.

Summary of remaining findings:
<paste the last reviewer output here>

Please review the findings manually and resolve them before pushing.
```

Do **not** push. Leave the branch in its current state for the user to inspect.

---

### Notes

- The amend in Step 5 rewrites history on the local branch. This is intentional
  and safe because the branch has not been pushed yet.
- If the branch has already been force-pushed before, remind the user they will
  need `git push --force-with-lease` after the shepherd approves.
- If the working tree is dirty (uncommitted changes), ask the user whether to
  stash them first: `git stash` before Step 1 and `git stash pop` after Step 7.
- Reviewer output may include findings that are not auto-fixable (e.g. design
  concerns). If the engineer reports it cannot fix a finding, record it in the
  summary and continue the loop — the reviewer may downgrade it to a warning on
  the next pass.
- Because each iteration's reviewer and engineer are fresh agents with no memory
  of previous iterations, always re-state the full context (branch, diff scope,
  findings) in every prompt.
