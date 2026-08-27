# Reviewer AUDIT MODE — Real-Run Baseline

First real (non-dry-run) execution of the planted-defect harness (US-201 / issue #15).
This is the baseline that unlocked Phase 2 of ADR-011's CI gate: the gate
(`.github/workflows/agent-qa.yml`, issue #17) is implemented with Phase 2 active — the
real-run leg enforces `--threshold 4`, ratcheting to 5 after five consecutive ≥5/6 runs.

## Run record

| Field | Value |
|---|---|
| Date | 2026-08-25 |
| Plugin version | 3.1.0 |
| Main commit | d874904d |
| claude CLI | 2.1.231 |
| Invocation | `tests/agent-qa/run-harness.sh` (real mode, default `--threshold 4`) |
| Funding | Claude subscription (`ANTHROPIC_API_KEY` blanked by the harness) |
| Review output | `results/review-output.txt` (50 lines) |

Pre-flight: dry run scored 6/6; CLI flags verified against claude 2.1.231 — all current
(`--max-turns` no longer appears in `--help` but still parses; no fixes needed).

## Score: 6/6 (100%)

| ID | Technique | Result | Matched pattern |
|---|---|---|---|
| D1 | diff parallel implementations | FOUND | `normaliz` |
| D2 | falsify doc comments | FOUND | `corrupt.{0,40}(skip\|fatal\|abort)` |
| D3 | what would this test still pass with | FOUND | `TestCompact` |
| D4 | language traps (float equality) | FOUND | `TopRated` |
| D5 | language traps / semantic (truncation before filter) | FOUND | `truncat` |
| D6 | state/failure paths (error after commit + retry) | FOUND | `persistWithRetry` |

## Notes

- Single-run baseline. LLM review runs are stochastic; per ADR-011 the enforced CI
  threshold stays at 4/6 to absorb single-run variance — a 6/6 baseline does not raise
  the gate.
- Ratchet condition (ADR-011): raise to 5/6 after five consecutive ≥5/6 real runs;
  append subsequent run records to this file when they inform threshold changes.
