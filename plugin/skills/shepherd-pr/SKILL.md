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

# Shepherd PR

Automates the cycle: **check state → fix threads → fix CI → push → reply + resolve →
schedule wakeup → repeat** until APPROVED with zero open conversations.

Each invocation does exactly ONE pass. When CI is still running after a push,
`ScheduleWakeup` re-fires this skill instead of blocking the main loop.

The automated reviewer is a GitHub Actions workflow that runs on every push and posts
its verdict via the GitHub Checks API. **No manual reviewer agent is needed.**

---

## Usage

```
/ai-pack:shepherd-pr [PR_NUMBER]
/ai-pack:shepherd-pr [PR_NUMBER] iter=N
```

If `PR_NUMBER` is omitted, auto-detect from the current branch.
On wakeup re-entries the prompt will include `iter=N` — parse it but do not show it to the user.

---

## State Machine (per invocation)

```
PARSE args → PR number + iter (default 0)
IF iter >= 10 → ESCALATE and stop (no wakeup)

FETCH state: unresolved threads, CI terminal?, verdict

IF threads == 0 AND CI all-terminal AND verdict == APPROVED:
  → SUCCESS REPORT — stop, no wakeup

IF threads > 0:
  → Fix all threads, commit, push, reply, resolve
  → ScheduleWakeup(270s, reason="waiting for CI after push to PR #N")
  → stop

IF threads == 0 AND CI has IN_PROGRESS checks:
  → ScheduleWakeup(120s, reason="CI still running on PR #N")
  → stop

IF threads == 0 AND CI all-terminal AND verdict != APPROVED:
  → ScheduleWakeup(90s, reason="waiting for reviewer to post threads on PR #N")
  → stop
```

**Delay rationale:**
- 270s after a push — automated reviewer takes ~3–4 min; stays within 5-min cache TTL
- 120s while CI in progress — avoids hammering GitHub, still responsive
- 90s waiting for reviewer threads — reviewer posts quickly after CI completes

---

## Step 1 — Parse Args and Resolve PR

```bash
# Parse PR and iter from args (e.g. "88" or "88 iter=2")
ARGS="${SKILL_ARGS:-}"
PR=$(echo "$ARGS" | grep -oE '^[0-9]+' || true)
ITER=$(echo "$ARGS" | grep -oE 'iter=([0-9]+)' | grep -oE '[0-9]+$' || echo "0")
MAX_ITER=10

if [ -z "$PR" ]; then
  PR=$(gh pr view --json number -q .number 2>/dev/null)
fi
if [ -z "$PR" ]; then
  echo "No open PR found. Create one first with: gh pr create"
  exit 1
fi

REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner)
OWNER="${REPO%/*}"
REPO_NAME="${REPO#*/}"
BRANCH=$(gh pr view "$PR" --json headRefName -q .headRefName)
HEAD_SHA=$(gh pr view "$PR" --json headRefOid -q .headRefOid)

echo "PR #$PR  iter=$ITER  branch: $BRANCH  HEAD: ${HEAD_SHA:0:7}"

if [ "$ITER" -ge "$MAX_ITER" ]; then
  echo "Max iterations ($MAX_ITER) reached — escalating. Review PR #$PR manually."
  exit 1
fi

git checkout "$BRANCH"
```

---

## Step 2 — Fetch State

```bash
# Unresolved review threads (always use first: 100 to avoid truncation)
THREADS=$(gh api graphql -f query="
{
  repository(owner: \"${OWNER}\", name: \"${REPO_NAME}\") {
    pullRequest(number: $PR) {
      reviewThreads(first: 100) {
        nodes {
          id isResolved path line
          comments(first: 10) {
            nodes { databaseId body author { login } }
          }
        }
      }
    }
  }
}" | jq '[.data.repository.pullRequest.reviewThreads.nodes[]
         | select(.isResolved == false)
         | {threadId:.id, commentId:.comments.nodes[0].databaseId,
            path:.path, line:.line,
            author:.comments.nodes[0].author.login,
            body:(.comments.nodes | map(.body) | join("\n\n"))}]')
THREAD_COUNT=$(echo "$THREADS" | jq 'length')

# CI checks — categorise into pending vs. failing
CHECKS_JSON=$(gh pr checks "$PR" --json name,state)
CI_PENDING=$(echo "$CHECKS_JSON" | jq '[.[] | select(
  .state != "SUCCESS" and .state != "FAILURE" and .state != "SKIPPED" and
  .state != "CANCELLED" and .state != "TIMED_OUT" and .state != "ACTION_REQUIRED" and
  .state != "STARTUP_FAILURE" and .state != "ERROR" and .state != "NEUTRAL" and
  .state != "STALE")] | length')
CI_FAILING=$(echo "$CHECKS_JSON" | jq '[.[] | select(
  .state == "FAILURE" or .state == "TIMED_OUT" or .state == "ACTION_REQUIRED" or
  .state == "STARTUP_FAILURE" or .state == "ERROR")] | length')

# Surface WAITING checks — environment approval gate, human action required
WAITING=$(echo "$CHECKS_JSON" | jq -r '.[] | select(.state == "WAITING") | .name')
if [ -n "$WAITING" ]; then
  echo "Checks WAITING for environment approval — human action required: $WAITING"
  echo "Shepherd halted — approve pending environment(s) in GitHub ($WAITING), then re-run."
  exit 1
fi

# Vacuous-true guard: only ALL_OK if there are actually checks and all are green
ALL_OK=$(echo "$CHECKS_JSON" | jq -r \
  '(length > 0) and ([.[].state] | all(. == "SUCCESS" or . == "SKIPPED" or . == "NEUTRAL" or . == "STALE" or . == "CANCELLED"))')

# Latest bot review verdict scoped to HEAD commit — uses gh pr view (allowlisted, avoids REST write surface)
VERDICT=$(gh pr view "$PR" --json reviews \
  | jq -r --arg sha "$HEAD_SHA" \
      '[.reviews[] | select(.commit.oid == $sha and (.author.login | endswith("[bot]")))] | last | .state // "PENDING"')

echo "Threads: $THREAD_COUNT  |  CI pending: $CI_PENDING  failing: $CI_FAILING  |  ALL_OK: $ALL_OK  |  Verdict: $VERDICT"
```

---

## Step 3 — Route

Evaluate in this order:

**A. Done:**
```
IF THREAD_COUNT == 0 AND ALL_OK == "true" AND CI_PENDING == 0 AND CI_FAILING == 0 AND VERDICT == "APPROVED"
  → print SUCCESS REPORT, stop (no ScheduleWakeup)
```

**B. Threads to fix** (fix immediately regardless of CI state):
```
IF THREAD_COUNT > 0
  → Steps 4–7 (fix CI if also failing, fix threads, commit, push, reply, resolve)
  → ScheduleWakeup(270s)
```

**C. Waiting for CI:**
```
IF THREAD_COUNT == 0 AND CI_PENDING > 0
  → ScheduleWakeup(120s, "CI still running on PR #$PR")
```

**D. CI done, waiting for reviewer verdict:**
```
IF THREAD_COUNT == 0 AND CI_PENDING == 0 AND VERDICT != "APPROVED"
  → ScheduleWakeup(90s, "waiting for reviewer on PR #$PR")
```

---

## Step 4 — Fix CI Failures (route B, if CI also failing)

When on route B and `CI_FAILING > 0`, fix CI failures first before addressing threads.

```bash
FAILURES=$(echo "$CHECKS_JSON" | jq -r '.[] | select(
  .state == "FAILURE" or .state == "TIMED_OUT" or .state == "ACTION_REQUIRED" or
  .state == "STARTUP_FAILURE" or .state == "ERROR") | .name')
```

For each failure, get the run log. `--workflow` expects the workflow filename (e.g.
`ci.yml`) — **not** the check name. A check like `CI / unit-tests` belongs to workflow
`ci.yml`; use `ci.yml` as the argument:

```bash
RUN_ID=$(gh run list --branch "$BRANCH" --workflow "<workflow-file.yml>" \
  --json databaseId,conclusion \
  --jq '[.[] | select(.conclusion == "failure" or .conclusion == "timed_out" or
         .conclusion == "action_required" or .conclusion == "startup_failure" or
         .conclusion == "error")] | first | .databaseId')
if [ -z "$RUN_ID" ] || [ "$RUN_ID" = "null" ]; then
  echo "No matching run found — skipping log fetch"
else
  gh run view "$RUN_ID" --log-failed 2>&1 | head -150
fi
```

**Fix matrix:**

| What failed | Method |
|---|---|
| Workflow YAML, markdown, shell scripts, config | Fix directly with Edit tool |
| Application code (`.kt`, `.swift`, `.ts`, `.go`, `.java`) | Delegate to engineer (see §Delegation) |
| Build / compile errors in app code | Delegate to engineer |
| Test failures in app code | Delegate to engineer |

After CI fixes, do **not** commit yet — include CI fixes in the single thread-fix commit below.

---

## Step 5 — Fix Threads (route B)

Read each unresolved thread's full body. Triage by severity prefix:

- **[BLOCKING]** — always fix
- **[SUGGESTION]** — fix if reasonable; reply with reasoning if declining
- No severity prefix (other bots or human reviewers) — fix if straightforward

Do not re-raise issues that are already resolved.

Fix strategy: workflow/config/markdown → fix directly; application code → delegate to
engineer (see §Delegation).

---

## Step 6 — Commit and Push

**Commit and push before replying** — replies must reference a real commit SHA.

```bash
git add <changed files>
if ! git diff --cached --quiet; then
  git commit -m "fix(pr${PR}): address reviewer feedback — <brief list>"
  if ! git push origin "$BRANCH"; then
    echo "Push failed — leaving threads open for next iteration"
    exit 1
  fi
else
  echo "Nothing to commit — all fixes were reply-only or delegated"
fi
```

**Push must succeed before resolving threads.** A push failure exits immediately,
leaving all threads open so the next iteration retries them.

---

## Step 7 — Reply and Resolve

```bash
NEW_SHA=$(gh pr view "$PR" --json headRefOid -q .headRefOid | cut -c1-7)

# For each thread: post reply then resolve immediately (do not batch)
gh api "repos/${REPO}/pulls/${PR}/comments/${COMMENT_ID}/replies" \
  -f body="Fixed in ${NEW_SHA} — <one-sentence description>."

gh api graphql -f query="mutation {
  resolveReviewThread(input: {threadId: \"${THREAD_ID}\"}) {
    thread { isResolved }
  }
}"
```

Every thread gets a reply — including declined suggestions (explain why, then resolve).
Resolve each thread immediately after replying — do not batch.

---

## Step 8 — Schedule Wakeup

After fixing threads and pushing (route B):

```
ScheduleWakeup(
  delaySeconds: 270,
  reason: "waiting for CI after push to PR #$PR (iter $((ITER+1))/$MAX_ITER)",
  prompt: "/ai-pack:shepherd-pr $PR iter=$((ITER+1))"
)
```

For route C (CI in progress):

```
ScheduleWakeup(
  delaySeconds: 120,
  reason: "CI still running on PR #$PR (iter $((ITER+1))/$MAX_ITER)",
  prompt: "/ai-pack:shepherd-pr $PR iter=$((ITER+1))"
)
```

For route D (CI passed, waiting for reviewer verdict):

```
ScheduleWakeup(
  delaySeconds: 90,
  reason: "waiting for reviewer to post threads on PR #$PR (iter $((ITER+1))/$MAX_ITER)",
  prompt: "/ai-pack:shepherd-pr $PR iter=$((ITER+1))"
)
```

---

## Delegation to Engineer

When a fix requires application code changes, delegate to the engineer sub-agent.
Provide a dense brief with exact file paths and specific changes (vague briefs cause
the engineer to spend turns reading the codebase instead of writing code):

**Brief template for engineer delegation:**

```
All context provided

## Context
PR #<N> reviewer flagged: "<exact quote from thread body>"
Root cause: <your assessment from reading the diff>

## Task
Fix <specific issue> in <file>.

## Files to change
- `<ABSOLUTE_PATH/to/file>` — <exact change description>

## Exact changes
<paste relevant code or precise instruction>

## Acceptance criteria
- [ ] `git diff --name-only` shows <file>
- [ ] <relevant build or test command> exits 0
```

After the engineer completes, pull its commits:

```bash
DIRTY=$(git status --porcelain)
[ -n "$DIRTY" ] && git stash
git pull --rebase origin "$BRANCH"
[ -n "$DIRTY" ] && git stash pop
```

If the engineer delegation fails, leave the thread open — it will be retried next iteration.

---

## Termination Report

### Success

```
PR #$PR is ready to merge
   Branch:     $BRANCH
   Iterations: $ITER
   Threads:    0 unresolved
   CI:         all green
   Verdict:    APPROVED on ${HEAD_SHA:0:7}
```

### Max Iterations Reached

```
PR #$PR — max iterations ($MAX_ITER) reached without APPROVED

Open threads:
  - path:line — first line of body
  ...

Last verdict: $VERDICT on ${HEAD_SHA:0:7}
Next step: review the threads manually or re-run /ai-pack:shepherd-pr $PR
```

---

## Escalation Table

Stop and report (no ScheduleWakeup) if:

| Condition | Action |
|-----------|--------|
| `iter >= MAX_ITER` | Max iterations report above |
| Same thread IDs unresolved 2 passes in a row | Report: "Thread stuck — may need human decision" |
| `CI_FAILING > 0` and no reviewer threads after 2 waits | Report CI failure details; ask user to investigate |
| Any check stuck IN_PROGRESS > 10 min | Ask user to check workflow logs |
| Verdict is `DISMISSED` after 2 passes | Ask user — bot may be broken |
| `WAITING` check detected | Halt with environment-approval message (handled in Step 2) |

---

## Constraints

- Never force-push. Use `git push` only.
- At most two commits per iteration: one for CI failures (Step 4) and one for thread
  fixes (Step 6). When there are no CI failures, only the Step-6 commit is created.
- Do not modify test expectations unless the reviewer explicitly requested it.
- Do not add features — only fix what CI or the reviewer flagged.
- Stay on the PR branch throughout. Never switch to main.
