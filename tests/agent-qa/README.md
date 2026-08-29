# Agent QA — Reviewer AUDIT MODE Harness (US-201)

A planted-defect harness that validates the reviewer agent's **AUDIT MODE**
(`plugin/agents/reviewer.md`). The fixture under `fixture/` is a small,
compiling Go record-store service with **six deliberately planted defects**,
each mapping to one of the audit techniques the reviewer prompt teaches.
The harness runs the reviewer prompt against the fixture and scores which
defects the resulting review found — so a prompt edit that weakens the
reviewer shows up as a score drop *before* merge.

## Layout

| Path | What it is |
|------|------------|
| `fixture/` | Self-contained Go module (`example.com/fixture`, no external deps) with the six planted defects. Compiles and its tests pass — one of them dishonestly (that is defect D3). |
| `defects.yaml` | The defect manifest: id, location, audit technique, description, and detection patterns per defect. |
| `run-harness.sh` | Extracts the reviewer system prompt, runs `claude -p` in audit mode against a throwaway temp-dir copy of the fixture (outside this repo, so the agent under test cannot read `defects.yaml` or this repo's git history), scores the output. |
| `score.py` | Stdlib-only scorer. Matches detection patterns (case-insensitive regex) against the review text. |
| `results/` | Run artifacts (`review-output.txt`). Gitignored. |

## How to run

**Dry run** — no claude invocation, no network, no quota. Validates prompt
extraction, prints the command a real run would execute, then generates a
synthetic review from the manifest descriptions and scores it (must be 6/6):

```bash
./run-harness.sh --dry-run
```

**Real run** — copies `fixture/` to a temp directory outside the repo (with a
fresh single-commit git history, cleaned up on exit), then invokes the
`claude` CLI there with the reviewer body as system prompt and the task
"Perform a full adversarial audit of this whole project." (which triggers
AUDIT MODE). The isolation matters: run in-tree, the reviewer could `cat
../defects.yaml` or read this repo's history and score 6/6 by cheating:

```bash
./run-harness.sh                  # results/review-output.txt + score
./run-harness.sh --output /tmp/qa --threshold 5
```

> **Cost note:** a real run consumes Claude subscription quota and typically
> takes several minutes (up to 60 agent turns reading and testing the
> fixture).

## The planted defects

| ID | Technique validated | Sketch |
|----|--------------------|--------|
| D1 | Diff parallel implementations | CLI insert paths call `store.Normalize`; the API `handleCreate` path skips it. |
| D2 | Falsify doc comments | `Load`'s comment claims a corrupt journal line is skipped; the code aborts the replay. |
| D3 | What would this test still pass with | `TestCompact` asserts on the compaction log sidecar while `Compact` silently drops records — data loss the test never sees. |
| D4 | Language traps | `TopRated` tie-breaks with `==` on computed float scores; the tie-break branch is dead. |
| D5 | Language traps / semantic | `Search` truncates to `limit` **before** filtering, dropping matches beyond the cap. |
| D6 | State/failure paths | `Append` returns an error after the journal write committed; `persistWithRetry` retries and duplicates the record. |

Exact locations and detection patterns live in `defects.yaml`.

## Scoring and thresholds

A defect counts as **found** when any of its `detection_patterns` matches the
review output (case-insensitive regex). The scorer prints a per-defect table
and `SCORE: n/6 (xx%)`, exiting 0 iff `n >= threshold`.

- **Default gate: 4/6.** The PRD's ≥80% detection target maps to **5/6**, but
  LLM runs are nondeterministic — a single flaky miss should not hard-fail
  the gate while confidence in run-to-run variance is still being built.
  Raise it with `--threshold 5` as evidence accumulates.
- The dry-run synthetic review must always score **6/6**, and the harness
  enforces that: the dry run scores with `score.py --require-all`, so anything
  less exits 1. A dry-run failure means a manifest description no longer
  contains any of its own patterns (a harness bug, not a reviewer regression).

## Adding a defect

1. Plant it in `fixture/` — embedded naturally, no "BUG HERE" comments; the
   fixture must still compile (`go build ./...`) and its tests still pass.
2. Add an entry to `defects.yaml` following the format comment at the top of
   that file (score.py parses a constrained YAML subset, not full YAML).
   Give it 2–4 detection patterns: think about what a review finding's text
   would plausibly say, and include function/file names as strong signals.
   The description must itself contain at least one of the patterns.
3. Map it to a named technique in reviewer.md's AUDIT MODE section.
4. Verify: `./run-harness.sh --dry-run` scores N/N.

## CLI flag dependency (`--max-turns`)

The harness caps the agent under test with `--max-turns 60`. That flag stopped
appearing in `claude --help` as of CLI 2.1.231 but still parses, which makes it an
**undocumented dependency**: the day the CLI drops it, the turn cap silently
disappears (or the run errors mid-flight) with nothing obvious to point at.

`run-harness.sh` therefore probes the flag before every run — dry or real. The probe
pairs `--max-turns` with a deliberately invalid flag and reads back which one the CLI
names as unknown; the CLI reports the first unrecognized option, and `--max-turns`
comes first, so being named means it is gone. Parsing rejects the invalid flag before
a session starts, so the probe costs no quota and no network.

The probe self-tests first: it runs the same check against a control flag that cannot
possibly exist and requires it to come back rejected. A probe that reports *everything*
as accepted has silently become a no-op, and its clean bill of health on `--max-turns`
would mean nothing — so a failed self-test downgrades to a warning instead of a
false all-clear.

| Probe outcome | Behavior |
|---|---|
| Flag accepted | Silent; run proceeds |
| Flag rejected as unknown | `error:` naming the CLI version and the fix, exit 1 |
| Neither error seen | `warning:`, run proceeds |
| Self-test failed | `warning:`, check skipped rather than trusted |
| `claude` not on PATH | `note:`, probe skipped (keeps `--dry-run` toolchain-free) |

In CI this surfaces on the real-run leg, which installs the pinned CLI. The dry-run
leg has no `claude` and skips the probe by design.

There is no supported replacement to adopt yet — as of CLI 2.1.236 the documented
flag list has no turn cap at all (`--max-budget-usd` caps spend, not turns). When one
lands, swap it in at all three sites in `run-harness.sh`: the real-run `claude`
invocation, the dry-run diagnostic `printf` that echoes the command, and the probe.

## CI integration

Implemented per [ADR-011](../../docs/adr/011-agent-qa-in-ci.md) (issue #17):
`.github/workflows/agent-qa.yml` publishes one required branch-protection
check, **`Agent QA / gate`**, with two legs behind change detection.

**Change detection.** PRs that touch none of `plugin/agents/**`,
`plugin/skills/**`, or `tests/agent-qa/**` pass the gate trivially. There is
deliberately no workflow-level `paths:` filter — a required check must report
on every PR or unrelated PRs hang on "Expected".

**Dry-run leg** (every prompt-change PR, always enforced). Runs
`./run-harness.sh --dry-run`: prompt extraction, the AUDIT MODE anchor, and
manifest self-consistency, scored with `--require-all` (must be 6/6).
Deterministic, free, seconds.

**Real-run leg** (same-repo, non-draft prompt-change PRs). Installs the
pinned `claude` CLI, runs `./run-harness.sh --threshold 4 --summary ...`,
uploads `review-output.txt` (plus the JSON summary) as the
`agent-qa-review-output` artifact, and upserts one sticky PR comment
(marker `<!-- agent-qa-score -->`) with the per-defect table, `SCORE: n/6`,
and the recorded baseline. Quota controls: per-PR concurrency with
cancel-in-progress, draft PRs skipped, 30-minute timeout.

**Threshold policy — Phase 2 is active.** [BASELINE.md](BASELINE.md) records
the first real-run baseline (6/6, 2026-08-25), so per ADR-011 the real-run
verdict is enforced: a score below **4/6** fails the gate (the 4/6 floor
absorbs a single flaky miss from the 5/6 PRD target). The comment warns when
the score is below both the baseline and 5/6. Ratchet: raise to
`--threshold 5` after five consecutive CI runs score ≥5/6, recording the
change in the ADR's status notes and BASELINE.md.

**Fork PRs.** The workflow triggers on plain `pull_request` only — never
`pull_request_target` — so fork PRs get no secrets and the real-run leg
auto-skips (the gate passes with a notice). This is deliberate: the harness
feeds the PR's `reviewer.md` to `claude` with Bash allowed, so a malicious
fork prompt could exfiltrate any reachable secret. Maintainer path for fork
contributions: review the diff, push the branch into this repo (or run
`workflow_dispatch` on it), and the funded leg executes from trusted context.

**Funding and rotation.** The real run is subscription-funded via the
`CLAUDE_CODE_OAUTH_TOKEN` repo secret; `ANTHROPIC_API_KEY` stays blanked so
no metered key can shadow it. Provision/rotate:

```bash
claude setup-token                      # mints a long-lived OAuth token locally
gh secret set CLAUDE_CODE_OAUTH_TOKEN   # paste it; stored as a repo Actions secret
```

Until the secret exists the real-run leg skips with a notice — it never
fails the gate. Registration of the required check itself lives in branch
protection (owner-only), not in the workflow file.
