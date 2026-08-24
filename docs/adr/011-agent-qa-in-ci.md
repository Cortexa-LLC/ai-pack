# ADR-011: Agent-QA Harness Results in CI for Prompt-Change PRs

**Date:** 2026-08-23
**Status:** Proposed
**Deciders:** Bryan Woodruff
**Refs:** US-203 (docs/product/prd-framework-strengthening.md), issue #17; depends on issue #15 (recorded real-run baseline)

---

## Context

The agent-QA harness (`tests/agent-qa/`, US-201) scores the reviewer agent's AUDIT MODE
against a planted-defect fixture: a prompt edit that weakens `plugin/agents/reviewer.md`
shows up as a score drop. Today running it is remembered, not enforced — US-203 requires
that a PR touching `plugin/agents/` or `plugin/skills/` surfaces harness status on the PR
before merge.

Constraints that shape the design:

- **No metered API usage** (PRD, non-negotiable). Commit f85af6f3 deliberately blanks any
  inherited `ANTHROPIC_API_KEY` in `run-harness.sh` so real runs bill the Claude Max
  subscription. API-key-funded CI is off the table by product constraint, not preference.
- **GitHub-hosted runners have no subscription session by default.** The gap is closed by
  the pattern the PRD calls out: `claude setup-token` mints a long-lived OAuth token from
  the Max subscription; stored as the `CLAUDE_CODE_OAUTH_TOKEN` repo secret, the `claude`
  CLI (and claude-code-action) authenticates in CI and bills the subscription. The
  maintainer already runs this pattern for consumer-project PR review. It transfers here
  directly — the harness invokes `claude -p` itself, so we install the CLI in the job and
  export the token; no action wrapper needed.
- **A real run is expensive and nondeterministic:** up to 60 agent turns, several minutes,
  subscription quota shared with the maintainer's daily interactive use, and run-to-run
  score variance (README: default gate 4/6 precisely because a single flaky miss must not
  hard-fail while variance evidence accumulates).
- **The artifact under test is attacker-controlled on fork PRs.** The harness feeds the
  PR's `reviewer.md` body to `claude` as the system prompt with `Bash` allowed. A
  malicious fork PR could edit the prompt to exfiltrate any secret present. This forbids
  `pull_request_target` + token for fork PRs categorically. Solo maintainer today (fork
  PRs rare), team later (US-602: per-user tokens, no shared personal credentials).
- **No enforcement threshold exists yet:** the harness has never had a recorded real-run
  baseline (issue #15). Enforcing a score gate before knowing the baseline would gate on
  a guess.
- The repo already has required checks on `main` (`Release consistency` et al. in
  `.github/workflows/ci.yml`); "no resident services" is a standing product decision.

## Decision

**A single required check, `Agent QA / gate`, in a new workflow
`.github/workflows/agent-qa.yml`, with two legs: a deterministic dry-run leg that is
always enforced, and a subscription-funded real-run leg that starts advisory and becomes
enforced once the baseline (issue #15) is recorded.**

Mechanics:

1. **Change detection.** A `changes` job diffs the PR against its merge base. If neither
   `plugin/agents/**`, `plugin/skills/**`, nor `tests/agent-qa/**` changed, `gate` passes
   immediately ("not a prompt change"). No workflow-level `paths:` filter — a required
   check must report on every PR or unrelated PRs hang on "Expected".
2. **Dry-run leg (every prompt-change PR, enforced from day one).** Runs
   `./run-harness.sh --dry-run` on `ubuntu-latest`: validates prompt extraction, the
   AUDIT MODE anchor, and manifest self-consistency (`--require-all`, must be 6/6).
   Deterministic, free, seconds. This is the US-202 smoke tier.
3. **Real-run leg (prompt-change PRs from branches in this repo only).** Installs the
   `claude` CLI, exports `CLAUDE_CODE_OAUTH_TOKEN` from repo secrets, runs
   `./run-harness.sh --threshold 4`, uploads `review-output.txt` as an artifact, and
   upserts one sticky PR comment with the per-defect table, `SCORE: n/6`, the recorded
   baseline, and the artifact link. Quota controls: skip draft PRs; `concurrency` group
   per-PR with `cancel-in-progress: true`; 30-minute job timeout; `workflow_dispatch`
   input to re-run manually on a given ref.
4. **Fork PRs:** the plain `pull_request` event exposes no secrets, so the real-run leg
   auto-skips with an explanatory neutral notice; the maintainer reviews the diff, pushes
   the branch into the repo (or triggers `workflow_dispatch` on it), and the funded run
   executes from trusted context. This is the deliberate security posture, not a gap.
5. **What the PR shows before merge** (acceptance criterion): the required
   `Agent QA / gate` status (green/red), plus the sticky score comment on any PR where
   the real run executed.

**Score threshold policy:**

- **Phase 1 (now → baseline recorded):** dry-run leg failure fails `gate`; real-run leg
  executes and reports (comment + artifact) but cannot fail `gate`. Merging with a bad
  score is visible, not blocked — better than today, honest about not yet knowing what
  "bad" is.
- **Phase 2 (after issue #15 lands `tests/agent-qa/BASELINE.md`):** real-run leg fails
  `gate` below **4/6** (harness default; a single flaky miss from the 5/6 PRD target
  does not block). The comment warns when the score is below both baseline and 5/6.
- **Ratchet:** raise to `--threshold 5` after five consecutive CI runs score ≥5/6;
  record the change by updating this ADR's status notes and BASELINE.md.

## Alternatives Rejected

- **Local pre-merge runs + maintainer posts a commit status via `gh api`:** cheapest and
  zero secrets in CI, but it is exactly the "remembered, not enforced" failure mode
  US-203 exists to kill — the gate's integrity would rest on the same human discipline it
  replaces. Survives only as the fork-PR manual path (run via `workflow_dispatch`, which
  posts a real check, rather than hand-crafted statuses).
- **Self-hosted runner on the maintainer's machine:** solves auth (local session) but is
  a resident service (violates a standing product decision), a single point of failure
  (laptop asleep = PRs blocked), and a security liability once outside PRs exist.
- **API-key-funded CI runs:** directly violates the PRD's non-negotiable zero-metered-cost
  constraint; f85af6f3 exists specifically to prevent this accidentally.
- **Scheduled nightly real runs as the gate:** decouples the score from the PR that
  caused it — fails the acceptance criterion ("visible on the PR before merge"). May
  later complement the gate for drift detection; out of scope.
- **`pull_request_target` so fork PRs get the token:** secret exfiltration by prompt
  edit, see Context. Categorically rejected.
- **Exact-match / golden-output comparison instead of thresholded score:** LLM output is
  nondeterministic; exact matching guarantees flakiness. The detection-pattern scorer with
  a threshold below target is the correct falsifiability/flakiness trade.

## Consequences

- Positive: the gate is enforced by branch protection, not memory; prompt regressions
  surface as a score drop on the PR itself; zero metered cost; unrelated PRs pay nothing.
- Negative (accepted): prompt-change PRs consume subscription quota per push (bounded by
  concurrency-cancel + draft-skip; prompt PRs are rare); a long-lived OAuth token lives in
  repo secrets (rotate via `claude setup-token`; team era moves to per-user tokens,
  US-602); occasional flaky sub-4/6 runs will need a manual re-run; fork contributors
  wait on a maintainer for the funded leg.
- The `Release consistency` check is untouched; `Agent QA / gate` is added alongside it
  in branch protection.

## Implementation Work Breakdown

Engineer-executable; no further design input needed. Steps 1–5 can land now (Phase 1);
step 6 flips Phase 2 and is blocked on issue #15.

1. **Harness: machine-readable score.** Add `--summary <file>` to
   `tests/agent-qa/score.py` writing one JSON object
   (`{"score": n, "total": 6, "threshold": t, "per_defect": {...}}`); `run-harness.sh`
   passes it through. Exit-code semantics unchanged. Dry-run gate must still be 6/6.
2. **Workflow: `.github/workflows/agent-qa.yml`.**
   - Triggers: `pull_request` (all types except draft-only churn), `workflow_dispatch`
     (input: `ref`).
   - `changes` job: `git diff --name-only origin/${{ base }}...HEAD` matched against
     `plugin/agents/**`, `plugin/skills/**`, `tests/agent-qa/**`; expose a boolean output.
   - `dry-run` job (needs `changes`, if prompt-change): checkout, run
     `./run-harness.sh --dry-run`.
   - `audit` job (needs `changes`; if prompt-change **and**
     `github.event.pull_request.head.repo.full_name == github.repository` **and** not
     draft): install CLI (`npm install -g @anthropic-ai/claude-code`), `go` toolchain
     (fixture builds), env `CLAUDE_CODE_OAUTH_TOKEN: ${{ secrets.CLAUDE_CODE_OAUTH_TOKEN }}`,
     run `./run-harness.sh --threshold 4 --output "$RUNNER_TEMP/agent-qa"`; upload
     `review-output.txt` artifact; upsert sticky PR comment (marker
     `<!-- agent-qa-score -->`) from the JSON summary via `gh api`. Phase 1: job sets
     `continue-on-error: true` and publishes its verdict as an output.
     `concurrency: agent-qa-${{ github.ref }}`, `cancel-in-progress: true`,
     `timeout-minutes: 30`.
   - `gate` job (needs all, `if: always()`): fail iff dry-run failed, or (Phase 2) audit
     verdict failed; pass with notice when not a prompt change or when audit skipped
     (fork/draft). This job's context, **`Agent QA / gate`**, is the only required check.
3. **Secret provisioning (maintainer, manual, documented):** run `claude setup-token`
   locally; store as repo secret `CLAUDE_CODE_OAUTH_TOKEN`; note rotation procedure in
   `tests/agent-qa/README.md`.
4. **Branch protection:** add `Agent QA / gate` to required status checks on `main`.
5. **Docs:** replace the README's "CI integration — deferred pending architect input"
   section with the two-leg description, fork-PR posture, and threshold policy; link this
   ADR.
6. **Phase 2 flip (blocked on issue #15):** once `tests/agent-qa/BASELINE.md` records the
   first real-run baseline, remove `continue-on-error` from `audit` and make `gate`
   honor its verdict. Single-line workflow change plus README note.
