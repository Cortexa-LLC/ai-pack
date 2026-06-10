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

### Step 2 — Create a reviewer task and spawn the reviewer

Create a task packet directory for this review iteration. Use a short slug that
includes the iteration number (e.g. `pre-push-review-iter-1`):

```bash
TASK_DIR=".ai/tasks/pre-push-review-iter-${ITER}"
mkdir -p "$TASK_DIR"
```

Write `$TASK_DIR/task.md` with:
- The diff (or a note that it can be obtained with `git diff "$BASE"..HEAD`)
- The commit list
- Instruction: "Review this diff for correctness, quality, security, and style.
  Produce a verdict of APPROVE, REQUEST CHANGES, or BLOCK with inline findings."

Create the task using the `agent` CLI and note the task ID:

```bash
TID=$(agent create "Review local commits before push

Working directory: $(pwd)
Task packet: $TASK_DIR

Review the diff in task.md for correctness, quality, security, and
style. Produce a verdict of APPROVE, REQUEST CHANGES, or BLOCK with
inline findings." --json | jq -r '.id')
```

Spawn the reviewer agent and stream its output:

```bash
agent reviewer "$TID" --stream
```

Wait for the agent to complete.

---

### Step 3 — Parse the verdict

Read the review output. The reviewer writes its verdict to `$TASK_DIR/result.md`
and also prints the verdict line to stdout. Look for one of:

- `**APPROVE**` or `Verdict: APPROVE`
- `**REQUEST CHANGES**` or `Verdict: REQUEST CHANGES`
- `**BLOCK**` or `Verdict: BLOCK`

If the verdict is **APPROVE**:

```
✅ Reviewer approved. Ready to push.
Run: git push
```

Stop. Do not loop further.

---

### Step 4 — If verdict is REQUEST CHANGES or BLOCK

Extract the list of findings from the reviewer output (the numbered issues listed
under the verdict). These become the brief for the engineer.

Create an engineer task:

```bash
ENG_DIR=".ai/tasks/pre-push-fix-iter-${ITER}"
mkdir -p "$ENG_DIR"
```

Write `$ENG_DIR/task.md` with:
- The reviewer findings (copy verbatim from reviewer output)
- Instruction: "Fix all findings listed above. The code is in the working
  directory. Do not push — only make and stage the changes."

Create and spawn the engineer:

```bash
ETID=$(agent create "Fix pre-push review findings (iteration ${ITER})

Working directory: $(pwd)
Task packet: $ENG_DIR

Fix all findings listed in task.md. Do not push. Stage all changes
when done." --json | jq -r '.id')

agent engineer "$ETID" --stream
```

Wait for the engineer to complete.

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
