# ADR 005: Grade Seeding Redesign — Cold-Start at Grade C, Real-Task Outcomes Override LiveBench

**Status:** Accepted
**Date:** 2026-02-26
**Deciders:** Architect, Bryan W
**Supersedes:** Previous stub (incorrect Grade B claim, omitted key evidence)

---

## Context and Problem Statement

The model selector needed a starting grade when a model had never been observed on a
particular role/project. The original approach was to derive a seed grade from
LiveBench coding-benchmark scores. This produced an inflated starting point:
- Several models seeded at Grade A based on LiveBench results corrupted files during
  multi-file Go refactoring tasks.
- LiveBench measures chat/completion capability, not agentic file-editing robustness.

The system needed a defensible cold-start policy that:
1. Does not over-trust benchmark data that predicts a different capability than the one
   being measured.
2. Allows grades to rise only on the basis of real task evidence.
3. Prevents a single bad run from permanently downgrading a model before enough data
   exists to be statistically meaningful.

---

## Decision

**Cold-start grade is C, not B and not A.**

Grade C was chosen as the cold-start default for all models at all role/project
combinations where no runtime data exists. The specific grade ladder is:

| Grade | Cold-start? | How earned |
|-------|-------------|------------|
| A     | No          | ≥5 real task successes AND error_rate < 0.10 |
| B     | No          | Partial real-data accumulation; not used for seeding |
| C     | **Yes**     | Default for any model with no observed runtime data |
| D     | No          | Emerges from runtime failures (success_rate < threshold) |
| F     | No          | Emerges from runtime failures (lowest threshold) |

`scripts/seed-grades.py` writes Grade C JSON files for all known models.
`GradeSourceLiveBench` (constant in `performance_grading.go`) marks these seeded files.

Until a model accumulates `minSamplesForRuntimeGrade = 5` real task runs, the letter
grade is *anchored* to the seeded value — runtime rates are updated for visibility but
the grade does not change. After 5 runs the runtime-calculated grade takes over.

---

## Rationale

### Why Grade C and not Grade B

Grade B implies partial positive evidence. A model with zero observed runs on a given
role has no positive evidence — it has no evidence at all. Starting at C is a
conservative neutral stance: the model selector will still use C-graded models but
will prefer higher-graded alternatives when they exist.

Starting at B would have masked cold-start uncertainty and contradicted the observed
pattern where benchmark-grade-A models caused regressions.

### Why LiveBench scores are no longer used as the primary signal

LiveBench coding scores reflect instruction-following and code-generation capability
in a single-turn eval harness. Agentic task execution requires:
- Multi-file coordination without corrupting unrelated files
- Tool-call discipline (no hallucinated paths, no spurious deletes)
- Recovery from partial failure states

Empirical evidence: models with LiveBench Grade A produced catastrophic file
corruption on multi-file Go refactoring tasks. Grade A performance on LiveBench did
not predict safe agentic behavior.

The `GradeSourceLiveBench` constant is retained in `performance_grading.go` for
backward compatibility (existing seeded files may carry this source tag), but no new
grades are derived from LiveBench data.

### Why the 5-run anchor

With fewer than 5 runs the success/failure rate is dominated by noise. A single
failure out of 2 runs produces a 50% error rate that would collapse the grade to D or
F. The anchor prevents premature downgrading and gives the model a fair evaluation
window. Five runs was chosen as the minimum for a statistically meaningful sample
while keeping the cold-start window short enough to matter operationally.

---

## Consequences

**Positive:**
- Models can no longer enter the system at Grade A on the basis of benchmark data
  alone. Grades must be earned through production evidence.
- Conservative cold-start reduces risk of catastrophic task failures caused by
  over-trusting models on new role/project combinations.
- The promotion ladder is explicit and auditable: Grade → Source → TotalAttempts.

**Negative:**
- Grade C models are subject to escalation to higher-capability models when the model
  selector logic deems the task complex. This imposes extra cost and latency during
  the 5-run warm-up window.
- Models that are genuinely Grade A capable will not receive Grade A routing until
  they have 5 runs, even if their benchmark scores would predict strong performance.
- LiveBench data that *was* predictive for some models (e.g., reasoning-heavy roles)
  is discarded; there is no hybrid weighting.

**Neutral:**
- Existing model grade files with `source: livebench` continue to function. The anchor
  logic in `recalculateGrade` treats them as seeded-C until 5 runs have accumulated.
- The `seed-grades.py` script must be re-run whenever a new model is added to the
  known-models list to ensure it has a C-grade file before the model selector first
  encounters it.

---

## Alternatives Considered

### Option A: Keep LiveBench as primary signal, cap at Grade B

Rejected. LiveBench scores still provided false confidence. A cap at B would only have
reduced severity, not addressed root cause. The file-corruption incidents involved
models seeded at A; moving the cap to B would have prevented the worst cases but
left intermediate risk. Furthermore, maintaining LiveBench ingestion code adds
maintenance burden.

### Option B: Start at Grade D (pessimistic cold-start)

Rejected. Grade D triggers aggressive escalation logic in the model selector. Forcing
all new models through Grade D would significantly increase cost and latency for
models that are actually capable. Grade C achieves conservative-neutral without
penalising untested models unnecessarily.

### Option C: No seeding — let grades emerge entirely from runtime data

Rejected. Without a seed, the first task run for a new model has no grade at all.
The model selector falls back to random or alphabetical selection, which is
unpredictable and untestable. A deterministic C seed ensures consistent baseline
behaviour and makes tests reproducible.

---

## Implementation Notes

- **`scripts/seed-grades.py`**: Writes `grade: "C"`, `source: "livebench-seed"` JSON
  files for every model in the known-models list. Re-run after adding models.
- **`internal/monitoring/performance_grading.go`**:
  - `minSamplesForRuntimeGrade = 5` — anchor threshold.
  - `recalculateGrade()` — skips letter-grade recalc while `source` is livebench-
    prefixed and `TotalAttempts < 5`.
  - `calculateLetterGrade()` — the A/B/C/D/F ladder driven by `GradingCriteriaConfig`.
- **`internal/monitoring/model_selector.go`**:
  - `ReloadGrades()` is called on every `SelectModel()` invocation, enabling live
    grade promotions without server restart.

## Open Items / Follow-on Work

- **Catastrophic failure detection** (engineer task filed separately): The current
  grade system records failures but does not distinguish routine failures (wrong
  output, exceeded retries) from catastrophic failures (file corruption, data
  deletion, security boundary violation). A dedicated detection layer is needed so
  that catastrophic events trigger an immediate grade penalty bypass of the 5-run
  anchor and alert the operator. See filed engineer task.

---

## Related Documents

- `scripts/seed-grades.py` — authoritative source of cold-start grade policy
- `internal/monitoring/performance_grading.go` — grade anchor and promotion logic
- `internal/monitoring/model_selector.go` — grade-aware routing
- ADR 001: Two-Tier Agent Architecture (grade-aware selection context)
