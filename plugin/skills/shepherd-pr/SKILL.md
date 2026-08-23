---
name: shepherd-pr
description: >
  Drive an open GitHub PR to merge-ready state via a non-blocking state machine:
  one pass per invocation (check state → fix threads → fix CI → push → reply),
  then a scheduled wakeup until the PR is approved and all checks pass. This is
  the preferred entry point for shepherding a PR from the main session; spawn the
  ai-pack:pr-shepherd agent instead only when delegating PR-shepherding as one
  background workstream among several.
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
/ai-pack:shepherd-pr [PR_NUMBER] iter=N wait_iter=N
```

If `PR_NUMBER` is omitted, auto-detect from the current branch.
On wakeup re-entries the prompt will include `iter=N wait_iter=N` — parse them but do not show them to the user.

---

## State Machine (per invocation)

```
PARSE args → PR number + iter (default 0) + wait_iter (default 0)
MAX_ITER=10 (fix attempts)  MAX_WAIT_ITER=30 (polling rounds)

FETCH state: unresolved threads, CI terminal?, verdict

IF threads == 0 AND THREAD_UNKNOWN == 0 AND CI_UNKNOWN == 0 AND CI all-terminal AND CI_FAILING == 0 AND verdict == APPROVED:
  → ✅ SUCCESS REPORT — stop, no wakeup        [checked before iter guards]

IF iter >= MAX_ITER → ESCALATE (too many fix attempts) — stop, no wakeup
IF wait_iter >= MAX_WAIT_ITER → ESCALATE (CI/reviewer stuck) — stop, no wakeup

IF threads > 0:
  → Fix all threads, commit, push, reply, resolve
  → ScheduleWakeup(270s, iter=$((ITER+1)), wait_iter=0)
  → stop

IF threads == 0 AND THREAD_UNKNOWN == 0 AND CI_FAILING > 0:
  → ⛔ CI FAILURE REPORT — stop, no wakeup     [before Route C — fail fast]

IF threads == 0 AND THREAD_UNKNOWN == 0 AND CI_PENDING > 0:
  → ScheduleWakeup(120s, iter=$ITER, wait_iter=$((WAIT_ITER+1)))
  → stop

IF THREAD_UNKNOWN == 1 OR (threads == 0 AND CI_UNKNOWN == 1):
  → ScheduleWakeup(120s, iter=$ITER, wait_iter=$((WAIT_ITER+1)))   # retry until fetches succeed
  → stop

IF threads == 0 AND THREAD_UNKNOWN == 0 AND CI_UNKNOWN == 0 AND CI_PENDING == 0 AND CI_FAILING == 0 AND verdict != APPROVED:
  → ScheduleWakeup(90s, iter=$ITER, wait_iter=$((WAIT_ITER+1)))
  → stop
```

**Delay rationale:**
- 270s after a push — automated reviewer takes ~3–4 min; stays within 5-min cache TTL
- 120s while CI in progress — avoids hammering GitHub, still responsive
- 90s waiting for reviewer threads — reviewer posts quickly after CI completes

---

## Step 1 — Parse Args and Resolve PR

```bash
# Parse PR, iter, and wait_iter from args (e.g. "88 iter=2 wait_iter=5")
ARGS="${SKILL_ARGS:-}"
PR=$(echo "$ARGS" | grep -oE '^[0-9]+' || true)
ITER=$(echo " $ARGS" | grep -oE ' iter=([0-9]+)' | grep -oE '[0-9]+$' || echo "0")  # leading space prevents matching inside wait_iter=
WAIT_ITER=$(echo "$ARGS" | grep -oE 'wait_iter=([0-9]+)' | grep -oE '[0-9]+$' || echo "0")
MAX_ITER=10
MAX_WAIT_ITER=30

if [ -z "$PR" ]; then
  PR=$(gh pr view --json number -q .number 2>/dev/null)
fi
if [ -z "$PR" ]; then
  echo "No open PR found. Create one first with: gh pr create"
  exit 1
fi

REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner)
BRANCH=$(gh pr view "$PR" --json headRefName -q .headRefName)
HEAD_SHA=$(gh pr view "$PR" --json headRefOid -q .headRefOid)

echo "PR #$PR  iter=$ITER  wait_iter=$WAIT_ITER  branch: $BRANCH  HEAD: ${HEAD_SHA:0:7}"

git checkout "$BRANCH" || { echo "❌ git checkout $BRANCH failed — aborting"; exit 1; }
```

---

## Step 2 — Fetch State

```bash
# Unresolved review threads (always use first: 100 to avoid truncation)
THREAD_UNKNOWN=0
THREADS_RAW=$(gh api graphql -f query="
{
  repository(owner: \"${REPO%/*}\", name: \"${REPO#*/}\") {
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
}" 2>/dev/null) \
  || { echo "⚠️  Thread fetch failed (network/auth) — thread state unknown"; THREAD_UNKNOWN=1; THREADS_RAW='{"data":{"repository":{"pullRequest":{"reviewThreads":{"nodes":[]}}}}}'; }
THREADS=$(echo "$THREADS_RAW" | jq '[.data.repository.pullRequest.reviewThreads.nodes[]
         | select(.isResolved == false)
         | {threadId:.id, commentId:.comments.nodes[0].databaseId,
            path:.path, line:.line,
            author:.comments.nodes[0].author.login,
            body:(.comments.nodes | map(.body) | join("\n\n"))}]')
THREAD_COUNT=$(echo "$THREADS" | jq 'length')

# CI: fetch once, derive both counts from the same snapshot
CI_UNKNOWN=0
CI_JSON=$(gh pr checks "$PR" --json name,state 2>/dev/null) \
  || { echo "⚠️  CI fetch failed (network/auth) — CI state unknown"; CI_JSON="[]"; CI_UNKNOWN=1; }
CI_PENDING=$(echo "$CI_JSON" | jq '[.[] | select(.state != "SUCCESS" and .state != "FAILURE" and .state != "SKIPPED" and .state != "CANCELLED" and .state != "TIMED_OUT" and .state != "ACTION_REQUIRED" and .state != "STARTUP_FAILURE" and .state != "ERROR" and .state != "NEUTRAL" and .state != "STALE")] | length')
CI_FAILING=$(echo "$CI_JSON" | jq '[.[] | select(.state == "FAILURE" or .state == "TIMED_OUT" or .state == "ACTION_REQUIRED" or .state == "STARTUP_FAILURE" or .state == "ERROR")] | length')

# Surface WAITING checks — environment approval gate, human action required
WAITING=$(echo "$CI_JSON" | jq -r '.[] | select(.state == "WAITING") | .name')
if [ -n "$WAITING" ]; then
  echo "Checks WAITING for environment approval — human action required: $WAITING"
  echo "Shepherd halted — approve pending environment(s) in GitHub ($WAITING), then re-run."
  exit 1
fi

# Latest review verdict on HEAD SHA — uses gh pr view (already allowlisted, avoids REST write surface)
VERDICT_RAW=$(gh pr view "$PR" --json reviews 2>/dev/null) \
  || { echo "⚠️  VERDICT fetch failed (network/auth) — defaulting to PENDING"; VERDICT_RAW='{"reviews":[]}'; }
VERDICT=$(echo "$VERDICT_RAW" | jq -r --arg sha "$HEAD_SHA" \
    '[.reviews[] | select(.commit.oid == $sha and (.author.login | endswith("[bot]")))] | last | .state // "PENDING"')

echo "Threads: $THREAD_COUNT  |  CI pending: $CI_PENDING  failing: $CI_FAILING  |  Verdict: $VERDICT"
```

---

## Step 3 — Route

Evaluate in this order:

**A. Done:**
```
IF THREAD_COUNT == 0 AND THREAD_UNKNOWN == 0 AND CI_UNKNOWN == 0 AND CI_PENDING == 0 AND CI_FAILING == 0 AND VERDICT == "APPROVED"
  → print SUCCESS REPORT, stop (no ScheduleWakeup)
```
Note: `CI_UNKNOWN==1` or `THREAD_UNKNOWN==1` skips Route A and falls through to Route F to retry — unknown state is not the same as clean state.

**A2. Escalation guards (after Route A so final-iteration approvals succeed):**
```bash
if [ "$ITER" -ge "$MAX_ITER" ]; then
  echo "⚠️  Max fix iterations ($MAX_ITER) reached — escalating. Review PR #$PR manually."
  exit 1
fi
if [ "$WAIT_ITER" -ge "$MAX_WAIT_ITER" ]; then
  echo "⚠️  Max wait iterations ($MAX_WAIT_ITER) reached — CI or reviewer appears stuck. Review PR #$PR manually."
  exit 1
fi
```

**B. Threads to fix** (fix immediately regardless of CI state):
```
IF THREAD_COUNT > 0
  → Steps 4–7 (fix CI if also failing, fix threads, commit, push, reply, resolve)
  → ScheduleWakeup(270s, iter=$((ITER+1)), wait_iter=0)   # reset wait counter after each push
```

**D. CI failed — escalate** (before C — fail fast even while other checks still run):
```
IF THREAD_COUNT == 0 AND THREAD_UNKNOWN == 0 AND CI_FAILING > 0
  → print CI failure report, stop (no ScheduleWakeup)
```
Note: `THREAD_UNKNOWN==1` bypasses Route D and falls to Route F. This is intentional — if the thread fetch failed we cannot confirm there are zero threads, so we retry. A real CI failure will still be caught once the network recovers.

**C. Waiting for CI:**
```
IF THREAD_COUNT == 0 AND THREAD_UNKNOWN == 0 AND CI_PENDING > 0
  → ScheduleWakeup(120s, iter=$ITER, wait_iter=$((WAIT_ITER+1)))
```

**F. Fetch failed — retry:**
```
IF THREAD_UNKNOWN == 1 OR (THREAD_COUNT == 0 AND CI_UNKNOWN == 1)
  → ScheduleWakeup(120s, iter=$ITER, wait_iter=$((WAIT_ITER+1)))
```
Note: covers dead-ends where thread or CI fetch failed — both set PENDING=0/FAILING=0 for their domain, which would otherwise allow silent exits or false Route A success.

**E. CI passed, waiting for reviewer verdict:**
```
IF THREAD_COUNT == 0 AND THREAD_UNKNOWN == 0 AND CI_UNKNOWN == 0 AND CI_PENDING == 0 AND CI_FAILING == 0 AND VERDICT != "APPROVED"
  → ScheduleWakeup(90s, iter=$ITER, wait_iter=$((WAIT_ITER+1)))
```
Note: `CI_UNKNOWN==0` guard is explicit — Route F fires first when CI fetch failed, so Route E only runs on confirmed-clean CI.

---

## Step 4 — Fix CI Failures (route B, if CI also failing)

When on route B and `CI_FAILING > 0`, fix CI failures first before addressing threads.

```bash
FAILURES=$(echo "$CI_JSON" | jq -r '.[] | select(
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
git commit -m "fix(pr${PR}): address reviewer feedback — <brief summary>" \
  || { echo "❌ git commit failed — aborting"; exit 1; }
git push origin "$BRANCH" || { echo "❌ git push failed — aborting"; exit 1; }
```

---

## Step 7 — Reply to Each Thread

```bash
NEW_SHA=$(gh pr view "$PR" --json headRefOid -q .headRefOid | cut -c1-7)

if [ -z "${COMMENT_ID}" ] || [ "${COMMENT_ID}" = "null" ]; then
  echo "⚠️  Thread $THREAD_ID has no comment ID — skipping reply"
else
  gh api "repos/$REPO/pulls/$PR/comments/${COMMENT_ID}/replies" \
    -f body="Fixed in ${NEW_SHA} — <one-sentence description>."
fi
```

Every thread gets a reply — including declined suggestions (explain why, then resolve).
Resolve each thread immediately after replying — do not batch.

---

## Step 8 — Resolve Threads

```bash
gh api graphql -f query="mutation {
  resolveReviewThread(input: {threadId: \"$THREAD_ID\"}) {
    thread { isResolved }
  }
}"
```

---

## Step 9 — Schedule Wakeup

After fixing threads and pushing (route B):

```
ScheduleWakeup(
  delaySeconds: 270,
  reason: "waiting for CI after push to PR #$PR (iter $((ITER+1))/$MAX_ITER)",
  prompt: "/ai-pack:shepherd-pr $PR iter=$((ITER+1)) wait_iter=0"
)
```

For route C (CI in progress) — increments wait_iter only, not iter:

```
ScheduleWakeup(
  delaySeconds: 120,
  reason: "CI still running on PR #$PR (wait $((WAIT_ITER+1))/$MAX_WAIT_ITER)",
  prompt: "/ai-pack:shepherd-pr $PR iter=$ITER wait_iter=$((WAIT_ITER+1))"
)
```

For route E (CI passed, waiting for reviewer verdict) — increments wait_iter only:

```
ScheduleWakeup(
  delaySeconds: 90,
  reason: "waiting for reviewer to post threads on PR #$PR (wait $((WAIT_ITER+1))/$MAX_WAIT_ITER)",
  prompt: "/ai-pack:shepherd-pr $PR iter=$ITER wait_iter=$((WAIT_ITER+1))"
)
```

For route F (thread or CI fetch failed — retry) — increments wait_iter only:

```
ScheduleWakeup(
  delaySeconds: 120,
  reason: "thread or CI fetch failed on PR #$PR — retrying (wait $((WAIT_ITER+1))/$MAX_WAIT_ITER)",
  prompt: "/ai-pack:shepherd-pr $PR iter=$ITER wait_iter=$((WAIT_ITER+1))"
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
✅ PR #$PR is ready to merge
   Branch:     $BRANCH
   Iterations: $ITER
   Threads:    0 unresolved
   CI:         all green
   Verdict:    APPROVED on ${HEAD_SHA:0:7}
```

### Max Iterations Reached

```
⚠️  PR #$PR — max iterations ($MAX_ITER) reached without APPROVED

Open threads:
  • path:line — first line of body
  ...

Last verdict: $VERDICT on ${HEAD_SHA:0:7}
Next step: review the threads manually or re-run /ai-pack:shepherd-pr $PR
```

### CI Failure (Route D)

```
⛔ PR #$PR — CI is failing on ${HEAD_SHA:0:7}

Failing checks:
$(echo "$CI_JSON" | jq -r '.[] | select(.state == "FAILURE" or .state == "TIMED_OUT" or .state == "ACTION_REQUIRED" or .state == "STARTUP_FAILURE" or .state == "ERROR") | "  • \(.name) [\(.state)]"')

Next step: investigate CI failures, fix, commit, and re-run /ai-pack:shepherd-pr $PR
```

---

## Escalation Table

Stop and report (no ScheduleWakeup) if:

| Condition | Action |
|-----------|--------|
| `iter >= MAX_ITER` | Max iterations report above |
| `wait_iter >= MAX_WAIT_ITER` | CI/reviewer stuck — escalate; Routes C, E, F all increment this counter, so slow reviewer or perpetually-pending CI exhausts it without any fix iteration occurring |
| Same thread IDs unresolved 2 passes in a row | Report: "Thread stuck — may need human decision" |
| `CI_FAILING > 0` | Route D fires immediately — CI failure report, no wakeup |
| `Claude PR Review` check stuck IN_PROGRESS > 10 min | Ask user to check workflow logs |
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
