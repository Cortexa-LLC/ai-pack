# Tester Role — Model Performance

**Role:** `tester`  
**Description:** Testing specialist focused on comprehensive coverage (>80%)  
**Configured Default:** (not specified in role spec)  
**Last Updated:** 2026-02-19

---

## Model Performance Grades

| Model | Grade | Success Rate | Attempts | Confidence | Project | Last Tested | Notes |
|-------|-------|-------------|----------|------------|---------|-------------|-------|
| gpt-4o-mini | ? | — | 0 | — | — | — | Expected default (not confirmed) |
| claude-sonnet-4-5 | **F** | 33% (1/3) | 3 | 0.15 (low) | xasm++ | 2026-02-14 | Very low confidence |
| claude-opus-4-6 | ? | — | 0 | — | — | — | Not benchmarked |

---

## Role-Specific Model Considerations

### Tester Task Profile
Testing tasks are generally well-suited for cheaper models:
- Pattern-following (test file structure is repetitive)
- Clear success criteria (tests pass/fail)
- Domain knowledge less critical than for architect/spelunker
- Can work from existing test examples

This suggests **gpt-4o-mini should perform well** for the tester role — the F grade for claude-sonnet-4-5 on xasm++ likely reflects the project's domain complexity, not a general testing ability gap.

### xasm++ Testing Challenges
The 6502 assembler project requires:
- Understanding of assembly language
- Knowledge of 6502 architecture specifics
- Test patterns for low-level code
This is harder than typical application testing.

---

## Task Types and Model Fit

| Task Type | Recommended Model | Reasoning |
|-----------|------------------|-----------|
| Unit test generation | gpt-4o-mini | Template-driven, pattern following |
| Integration test design | gpt-4o-mini → gpt-4o | Slightly more reasoning needed |
| Edge case identification | gpt-4o | Requires broader thinking |
| Performance testing | gpt-4o → claude-sonnet-4-5 | Complex scenarios |
| Security test design | claude-sonnet-4-5 | Specialized knowledge needed |
| Assembly/low-level tests | claude-sonnet-4-5 | Domain knowledge required |

---

## Benchmark History

### 2026-02-14: claude-sonnet-4-5 on xasm++ (Tester)
- **Result:** F (33% success rate)
- **Sample:** 3 attempts, 1 success, 2 failures
- **Confidence:** 0.15 (very low)
- **Context:** Assembly language testing is a niche domain

---

## Gaps in Data

- [ ] No gpt-4o-mini benchmarks (most likely default for this role)
- [ ] No benchmarks on typical web/mobile application test generation
- [ ] Need to capture failure modes from the 2 failed attempts

---

*Last updated: 2026-02-19 | Grade data from `.claude/performance_grades_backup/`*
