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
MAX_ITER=10 (fix attempts — safety cap, convergence-gated; see A2)  MAX_WAIT_ITER=30 (polling rounds)

FETCH state: unresolved threads, review body, CI terminal?, verdict
  ACTIONABLE = threads > 0 OR body findings on a non-approved verdict
  CI_FAILING excludes the review gate check (it fails BY DESIGN on changes-requested)

IF ACTIONABLE == 0 AND THREAD_UNKNOWN == 0 AND CI_UNKNOWN == 0 AND CI all-terminal AND CI_FAILING == 0 AND verdict == APPROVED:
  → ✅ SUCCESS REPORT — stop, no wakeup        [checked before iter guards]

IF iter >= MAX_ITER → ESCALATE (too many fix attempts) — stop, no wakeup
IF wait_iter >= MAX_WAIT_ITER → ESCALATE (CI/reviewer stuck) — stop, no wakeup
IF same bot findings re-raised 2 consecutive passes → STUCK REPORT — stop, no wakeup  [see A2 convergence rule]

IF ACTIONABLE == 1:
  → Fix findings (threads and/or review body), commit, push, reply, resolve
  → ScheduleWakeup(270s, iter=$((ITER+1)), wait_iter=0)
  → stop

IF ACTIONABLE == 0 AND THREAD_UNKNOWN == 0 AND CI_FAILING > 0:
  → ⛔ CI FAILURE REPORT — stop, no wakeup     [before Route C — fail fast]

IF ACTIONABLE == 0 AND THREAD_UNKNOWN == 0 AND CI_PENDING > 0:
  → ScheduleWakeup(120s, iter=$ITER, wait_iter=$((WAIT_ITER+1)))
  → stop

IF THREAD_UNKNOWN == 1 OR (ACTIONABLE == 0 AND CI_UNKNOWN == 1):
  → ScheduleWakeup(120s, iter=$ITER, wait_iter=$((WAIT_ITER+1)))   # retry until fetches succeed
  → stop

IF ACTIONABLE == 0 AND THREAD_UNKNOWN == 0 AND CI_UNKNOWN == 0 AND CI_PENDING == 0 AND CI_FAILING == 0 AND verdict == PENDING:
  → ScheduleWakeup(90s, iter=$ITER, wait_iter=$((WAIT_ITER+1)))    # review not posted yet
  → stop

IF ACTIONABLE == 0 AND ... AND verdict is neither APPROVED nor PENDING:
  → 🔎 REVIEWED-NOT-APPROVED REPORT — stop, no wakeup   [terminal verdict, nothing to act on]
                                                        [complement, not a list: catches DISMISSED too]
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

# The review gate check FAILS BY DESIGN when the reviewer requests changes — that is
# the gate working, not broken CI. Counting it as a CI failure sends Route D's "fix
# CI" report at exactly the moment the real signal is a review finding. Exclude it
# from CI_FAILING by name; VERDICT already carries its meaning. While the gate is
# still running it counts toward CI_PENDING, which is exactly right — "the review
# has not finished". Once it reaches FAILURE it falls out of BOTH counts (the
# CI_PENDING query excludes FAILURE by state), and its signal is carried entirely
# by VERDICT. That is deliberate, not an oversight: a gate whose only meaning is
# "the reviewer said no" must not also read as a broken build.
# Override in a consumer repo whose gate check is named differently (issue #30).
REVIEW_GATE_CHECK="${REVIEW_GATE_CHECK:-Claude verdict}"
CI_PENDING=$(echo "$CI_JSON" | jq '[.[] | select(.state != "SUCCESS" and .state != "FAILURE" and .state != "SKIPPED" and .state != "CANCELLED" and .state != "TIMED_OUT" and .state != "ACTION_REQUIRED" and .state != "STARTUP_FAILURE" and .state != "ERROR" and .state != "NEUTRAL" and .state != "STALE")] | length')
CI_FAILING=$(echo "$CI_JSON" | jq --arg gate "$REVIEW_GATE_CHECK" '[.[] | select(.name != $gate) | select(.state == "FAILURE" or .state == "TIMED_OUT" or .state == "ACTION_REQUIRED" or .state == "STARTUP_FAILURE" or .state == "ERROR")] | length')

# Surface WAITING checks — environment approval gate, human action required
WAITING=$(echo "$CI_JSON" | jq -r '.[] | select(.state == "WAITING") | .name')
if [ -n "$WAITING" ]; then
  echo "Checks WAITING for environment approval — human action required: $WAITING"
  echo "Shepherd halted — approve pending environment(s) in GitHub ($WAITING), then re-run."
  exit 1
fi

# Latest review verdict on HEAD SHA — REST, NOT `gh pr view --json reviews`: that field
# is GraphQL-backed and reports app logins WITHOUT the `[bot]` suffix (`cortexa-llc-reviewer`)
# and carries no type field, so a bot filter there never matches an app-posted review.
# REST exposes `.user.type == "Bot"`; `--paginate` merges pages into one array.
VERDICT_RAW=$(gh api "repos/$REPO/pulls/$PR/reviews" --paginate 2>/dev/null) \
  || { echo "⚠️  VERDICT fetch failed (network/auth) — defaulting to PENDING"; VERDICT_RAW='[]'; }
VERDICT=$(echo "$VERDICT_RAW" | jq -r --arg sha "$HEAD_SHA" \
    '[.[] | select(.commit_id == $sha and .user.type == "Bot")] | last | .state // "PENDING"')

# Findings do not only arrive as inline threads. A reviewer that posts with
# `gh pr review --body-file` submits a SINGLE top-level review body and cannot
# attach line comments, so reviewThreads is empty and every finding lives in the
# body. Keying the state machine on THREAD_COUNT alone makes those reviewers
# invisible: THREAD_COUNT is permanently 0, the fix route never fires, and the
# shepherd waits out its polling budget next to a review full of findings.
REVIEW_BODY=$(echo "$VERDICT_RAW" | jq -r --arg sha "$HEAD_SHA" \
    '[.[] | select(.commit_id == $sha and .user.type == "Bot")] | last | .body // ""')

# A body is actionable when the verdict is not an approval and the body has
# substance. An APPROVED review's body may still hold nits, but approval is a
# terminal success — do not reopen it (the treadmill guard in Step 5 explains why).
BODY_FINDINGS=0
if [ "$VERDICT" != "APPROVED" ] && [ "$VERDICT" != "PENDING" ] && [ "${#REVIEW_BODY}" -gt 40 ]; then
  BODY_FINDINGS=1
fi

# One actionable-work signal for the router, from either source.
ACTIONABLE=0
if [ "$THREAD_COUNT" -gt 0 ] || [ "$BODY_FINDINGS" -eq 1 ]; then ACTIONABLE=1; fi

echo "Threads: $THREAD_COUNT  |  body findings: $BODY_FINDINGS  |  CI pending: $CI_PENDING  failing: $CI_FAILING  |  Verdict: $VERDICT"
```

---

## Step 3 — Route

Evaluate in this order:

**A. Done:**
```
IF ACTIONABLE == 0 AND THREAD_UNKNOWN == 0 AND CI_UNKNOWN == 0 AND CI_PENDING == 0 AND CI_FAILING == 0 AND VERDICT == "APPROVED"
  → settle check: sleep 30, re-run the thread query from Step 2 once
    (an APPROVING review can post SUGGESTION threads seconds AFTER its verdict)
  → IF threads still 0: print SUCCESS REPORT, stop (no ScheduleWakeup)
  → ELSE: treat as Route B this same pass (late threads are usually
    SUGGESTION-level — see the treadmill guard in Step 5)
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

**Convergence rule:** the budget is generous by design — clean convergence against a
thorough automated reviewer routinely takes more than 5 fix rounds; 5 is the minimum
meaningful budget, so keep `MAX_ITER` at least that high. If the reviewer re-raises the
same findings unchanged on 2 consecutive passes, stop then regardless of `iter` —
report **"stuck"** with an analysis of the repeated findings (a stall, not a failure).
Detecting this needs no cross-wakeup state: the full review history is durable on
GitHub — fetch it read-only with `gh api "repos/$REPO/pulls/$PR/reviews"`, take the
bot's two most recent reviews (`jq '[.[] | select(.user.type == "Bot")] | .[-2:]'`),
and compare their Critical/Major findings; substantially the same finding sets on both
— same file:line anchors with equivalent descriptions, regardless of exact wording —
means non-converging. (Fewer than two bot reviews yet = trivially still converging.
`$VERDICT_RAW` from Step 2 is this same REST payload, so reuse it instead of re-fetching.)

If the `MAX_ITER` guard fires while rounds were still converging (each round's
findings fewer or smaller than the last), say so in the report and recommend
re-running `/ai-pack:shepherd-pr $PR` to continue.

**B. Findings to fix** (fix immediately regardless of CI state):
```
IF ACTIONABLE == 1          # inline threads, a review body, or both
  → Steps 4–8 (fix CI if also failing, fix findings, commit, push, reply, resolve)
  → ScheduleWakeup(270s, iter=$((ITER+1)), wait_iter=0)   # reset wait counter after each push
```
`ACTIONABLE` deliberately covers both shapes of reviewer. A thread-posting reviewer
sets `THREAD_COUNT > 0`; a body-posting reviewer sets `BODY_FINDINGS == 1` with
`THREAD_COUNT == 0`. Routing on `THREAD_COUNT` alone is what made the second kind
invisible — see Step 2.

**D. CI failed — escalate** (before C — fail fast even while other checks still run):
```
IF ACTIONABLE == 0 AND THREAD_UNKNOWN == 0 AND CI_FAILING > 0
  → print CI failure report, stop (no ScheduleWakeup)
```
`CI_FAILING` excludes `$REVIEW_GATE_CHECK` by construction (Step 2). That check fails
by design when the reviewer requests changes, and counting it here produced the exact
wrong instruction: "CI is failing — investigate CI failures" on a PR whose builds were
all green and whose real signal was a review finding. A `CHANGES_REQUESTED` verdict
belongs to Route B, not Route D.
Note: `THREAD_UNKNOWN==1` bypasses Route D and falls to Route F. This is intentional — if the thread fetch failed we cannot confirm there are zero threads, so we retry. A real CI failure will still be caught once the network recovers.

**C. Waiting for CI:**
```
IF ACTIONABLE == 0 AND THREAD_UNKNOWN == 0 AND CI_PENDING > 0
  → ScheduleWakeup(120s, iter=$ITER, wait_iter=$((WAIT_ITER+1)))
```

**F. Fetch failed — retry:**
```
IF THREAD_UNKNOWN == 1 OR (ACTIONABLE == 0 AND CI_UNKNOWN == 1)
  → ScheduleWakeup(120s, iter=$ITER, wait_iter=$((WAIT_ITER+1)))
```
Note: covers dead-ends where thread or CI fetch failed — both set PENDING=0/FAILING=0 for their domain, which would otherwise allow silent exits or false Route A success.

**E. CI passed, waiting for the reviewer to run:**
```
IF ACTIONABLE == 0 AND THREAD_UNKNOWN == 0 AND CI_UNKNOWN == 0 AND CI_PENDING == 0 AND CI_FAILING == 0 AND VERDICT == "PENDING"
  → ScheduleWakeup(90s, iter=$ITER, wait_iter=$((WAIT_ITER+1)))
```
Note: `CI_UNKNOWN==0` guard is explicit — Route F fires first when CI fetch failed, so Route E only runs on confirmed-clean CI.

**Only `PENDING` waits here.** `COMMENTED` and `CHANGES_REQUESTED` are *terminal
verdicts* — the review ran and reached a conclusion, and neither becomes `APPROVED`
without a new push. Waiting on them burns the entire polling budget
(30 × 90s ≈ 45 minutes) beside a finished review, then escalates as "reviewer appears
stuck" when nothing was stuck at all. Those verdicts carry findings, so they route to
B via `BODY_FINDINGS`; the only case that lands here is a review that has not been
posted for this HEAD yet.

**G. Reviewed, nothing actionable, but not approved:**
```
IF ACTIONABLE == 0 AND THREAD_UNKNOWN == 0 AND CI_UNKNOWN == 0 AND CI_PENDING == 0 AND CI_FAILING == 0
   AND VERDICT is not APPROVED and not PENDING
  → print REVIEWED-NOT-APPROVED REPORT, stop (no ScheduleWakeup)
```
Reached when the reviewer concluded but left nothing this skill can act on — a
`COMMENTED` verdict with an empty or trivial body, or findings already dispositioned.
Report the verdict, quote the body, and hand the decision to the human. Do **not**
schedule a wakeup: no further automated event is coming.

The guard is the **complement** of the two handled states, not a list of
`COMMENTED`/`CHANGES_REQUESTED`. `APPROVED` exits at Route A and `PENDING` waits at
Route E; every other value — `DISMISSED`, or any state this skill does not
recognise — is terminal-with-nothing-to-do and belongs here. Enumerating the two
known verdicts instead would let `DISMISSED` match no route at all and exit silently
with neither a wakeup nor a report, which is the one outcome the state machine must
never produce.

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

## Step 5 — Fix Findings (route B)

Findings arrive in one of two shapes, and a PR can carry both. Handle whichever is
present.

**Shape 1 — inline threads** (`THREAD_COUNT > 0`). Read each unresolved thread's full
body, triage, fix, then reply and resolve per Steps 7–8.

**Shape 2 — a review body** (`BODY_FINDINGS == 1`). The reviewer posted with
`gh pr review --body-file`, which submits one top-level body and cannot attach line
comments, so there is nothing to reply to or resolve. `$REVIEW_BODY` from Step 2 holds
the whole review: parse its findings out of the body text and triage them the same way.
Because there is no thread to resolve, the disposition goes in a single PR comment
after the push (Step 7), naming each finding and what you did about it. A finding you
decline still gets a line — silence reads as an oversight.

Triage by severity prefix:

- **[BLOCKING]** — always fix
- **[SUGGESTION]** — fix if reasonable; reply with reasoning if declining
- No severity prefix (other bots or human reviewers) — fix if straightforward

**Treadmill guard:** every push triggers a fresh automated review, which can mint new
cosmetic findings against your fix (and re-post earlier findings as duplicate threads).
Once the verdict on HEAD is APPROVED and every remaining unresolved thread is
SUGGESTION-level, prefer **reply-with-reasoning + resolve, without pushing** — that is
a legitimate maintainer response and it ends the cycle. Resolve a duplicate thread with
a short reply referencing the original. Never resolve a [BLOCKING] finding without
fixing it.

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

## Step 7 — Reply

```bash
NEW_SHA=$(gh pr view "$PR" --json headRefOid -q .headRefOid | cut -c1-7)
```

**If the findings came from threads** — reply in each thread:

```bash
if [ -z "${COMMENT_ID}" ] || [ "${COMMENT_ID}" = "null" ]; then
  echo "⚠️  Thread $THREAD_ID has no comment ID — skipping reply"
else
  gh api "repos/$REPO/pulls/$PR/comments/${COMMENT_ID}/replies" \
    -f body="Fixed in ${NEW_SHA} — <one-sentence description>."
fi
```

Every thread gets a reply — including declined suggestions (explain why, then resolve).
Resolve each thread immediately after replying — do not batch.

**If the findings came from a review body** — there is no thread to reply into, so post
one PR comment covering the round:

```bash
gh pr comment "$PR" --body "$(cat <<EOF
Addressed in ${NEW_SHA}:

- <finding> — fixed, <one sentence>
- <finding> — declined, <why>
EOF
)"
```

One comment per round, not one per finding. Declined findings are listed with the
reasoning; a body-posting reviewer has no resolve affordance, so the comment is the
only record that a finding was considered.

---

## Step 8 — Resolve Threads

Only applies to inline threads; skip when the round's findings came from a review body
(there is nothing to resolve — the Step 7 comment is the disposition record).

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
   Body:       no outstanding findings
   CI:         all green (excluding $REVIEW_GATE_CHECK, which tracks the verdict)
   Verdict:    APPROVED on ${HEAD_SHA:0:7}
```

### Max Iterations Reached

```
⚠️  PR #$PR — max iterations ($MAX_ITER) reached without APPROVED

Open threads:
  • path:line — first line of body
  ...

Last verdict: $VERDICT on ${HEAD_SHA:0:7}
Convergence: <still converging — re-run to continue | stalled — same findings re-raised, see analysis>
Next step: review the threads manually or re-run /ai-pack:shepherd-pr $PR
```

### Stuck (Non-Converging)

```
🔁 PR #$PR — stuck: reviewer re-raised the same findings 2 passes in a row

Repeated findings (from the bot's last two reviews):
  • path:line — finding summary
  ...

Analysis: <why the fixes did not satisfy the reviewer — disagreement, missed root
cause, or a finding that needs a human decision>

Last verdict: $VERDICT on ${HEAD_SHA:0:7}
Next step: resolve the repeated findings manually (or dispute them on the PR), then
re-run /ai-pack:shepherd-pr $PR
```

### Reviewed, Not Approved (Route G)

```
🔎 PR #$PR — the reviewer concluded but did not approve

Verdict:  $VERDICT on ${HEAD_SHA:0:7}
Threads:  0 unresolved
CI:       all green

Review body:
  <quote the review, or "empty">

Nothing here is actionable by this skill — the verdict is terminal and will not
become APPROVED without a new push. Decide whether to address the comments, dispute
them, or merge as-is.
```

Do not schedule a wakeup for this state: no further automated event is coming.

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
| `wait_iter >= MAX_WAIT_ITER` | CI/reviewer stuck — escalate; Routes C, E, F all increment this counter, so slow reviewer or perpetually-pending CI exhausts it without any fix iteration occurring. Route E now waits only on `PENDING`, so a concluded-but-unapproved review no longer burns the budget |
| Same thread IDs unresolved 2 passes in a row | Report: "Thread stuck — may need human decision" |
| Same findings re-raised unchanged 2 passes in a row (compare the bot's last two reviews via `gh api "repos/$REPO/pulls/$PR/reviews"` — see A2 convergence rule) | Non-converging — Stuck (Non-Converging) report above, with analysis of the repeated findings; a stall, not a failure |
| `CI_FAILING > 0` (excludes `$REVIEW_GATE_CHECK`) | Route D fires immediately — CI failure report, no wakeup |
| Verdict `COMMENTED` / `CHANGES_REQUESTED` with nothing actionable left | Route G — Reviewed, Not Approved report above; terminal, no wakeup (never poll a concluded review) |
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
