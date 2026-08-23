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

## CI integration

Deferred pending architect input (US-203). The harness is designed for it —
deterministic dry-run for smoke coverage, exit-code gate for real runs — but
when and where real (quota-consuming) runs happen in CI is an open design
question.
