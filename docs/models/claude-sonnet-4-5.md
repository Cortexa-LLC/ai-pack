# claude-sonnet-4-5 Performance Profile

**Model ID:** `claude-sonnet-4-5-20250929`  
**Provider:** Anthropic  
**Tier:** 3 (Medium Cost)  
**Cost:** ~$3.00 per 1M input tokens, ~$15.00 per 1M output tokens  
**Context Window:** 200K tokens  
**Last Updated:** 2026-02-19

---

## Overview

Claude Sonnet 4.5 is Anthropic's mid-tier model, positioned as the default escalation target for roles needing moderate reasoning above cheap models (gpt-4o-mini, claude-haiku). It balances cost and capability but benchmark data from the `xasm++` project shows consistent failures across roles — likely due to project-specific complexity rather than fundamental model limitations.

---

## Performance Grades by Role

| Role | Grade | Success Rate | Attempts | Project | Last Tested | Notes |
|------|-------|-------------|----------|---------|-------------|-------|
| engineer | **F** | 33% (1/3) | 3 | ai-pack | 2026-02-19 | 2 failures, high error rate (67%) |
| engineer | **F** | 33% (1/3) | 3 | xasm++ | 2026-02-18 | Low confidence (0.15) |
| architect | **F** | 33% (4/12) | 12 | xasm++ | 2026-02-14 | Most attempts, partial credit likely |
| tester | **F** | 33% (1/3) | 3 | xasm++ | 2026-02-14 | Low confidence (0.15) |
| spelunker | **F** | 33% (1/3) | 3 | xasm++ | 2026-02-14 | Low confidence (0.15) |
| cartographer | **F** | 25% (1/4) | 4 | xasm++ | 2026-02-14 | Lowest success rate observed |
| orchestrator | **F** | 33% (2/6) | 6 | xasm++ | 2026-02-03 | Medium confidence (0.3) |
| reviewer | **F** | 33% (1/3) | 3 | xasm++ | 2026-02-14 | Low confidence (0.15) |

> ⚠️ **Important Caveat:** All `xasm++` benchmarks are from a single legacy 6502 assembler project with specialized domain knowledge requirements. These results may not generalize to typical software projects. The `ai-pack` engineer result (2026-02-19) is more representative of general use.

---

## Benchmark Summary Statistics

```
Total benchmark runs:   37 (across all roles)
Overall success rate:   ~32%
Overall grade:          F (current benchmarks)
Benchmark projects:     2 (ai-pack, xasm++)
Benchmark period:       2026-02-03 to 2026-02-19
```

### Usage Metrics (from daily metrics)
| Date | Calls | Input Tokens | Output Tokens | Cost |
|------|-------|-------------|---------------|------|
| 2026-02-17 | 463 | 12,208,755 | 118,866 | $38.41 |
| 2026-02-18 | 1,180 | 19,810,687 | 185,383 | $62.21 |
| 2026-02-19 | 21 | 828,747 | 2,343 | $2.52 |

---

## Observed Strengths

- **Large context window (200K):** Handles large codebases without truncation
- **Code generation:** Produces syntactically correct code in most cases
- **Instruction following:** Generally adheres to role-specific instructions
- **Multi-step reasoning:** Can break down complex tasks into steps

## Observed Weaknesses

- **Domain-specific knowledge:** Struggles with niche/legacy tech (6502 assembler, proprietary tools)
- **Error recovery:** Fails to recover gracefully from initial errors (high error rates seen)
- **Confidence calibration:** Low confidence scores even when succeeding
- **Consistency:** High variance across attempts on the same role/task type

---

## Recommended Use Cases

### ✅ Use claude-sonnet-4-5 For:
- Complex reasoning tasks that cheaper models fail
- Large-context analysis (>50K token inputs)
- Tasks escalated from gpt-4o-mini or claude-haiku
- Architecture review, security analysis, complex debugging
- Orchestration of multi-agent workflows

### ❌ Avoid claude-sonnet-4-5 For:
- Simple, pattern-following tasks (use gpt-4o-mini instead)
- High-volume repetitive work (cost inefficient)
- Projects with highly specialized domain knowledge (may need opus)
- When benchmark data shows consistent F grades for the specific role

---

## Escalation Path

```
gpt-4o-mini → gpt-4o → claude-sonnet-4-5 → claude-opus-4-6
```

**When to escalate TO claude-sonnet-4-5:**
- gpt-4o fails or produces poor quality output
- Task requires >32K context
- Complex multi-step reasoning needed

**When to escalate FROM claude-sonnet-4-5:**
- Consecutive failures (grade drops to F)
- Task requires deeper specialized reasoning
- Budget allows for opus-level quality

---

## Notes for Future Benchmarks

- [ ] Benchmark on diverse project types (not just xasm++)
- [ ] Test with simpler, well-defined tasks to establish baseline
- [ ] Compare against claude-sonnet-4-6 (newer version)
- [ ] Measure token efficiency per successful task
- [ ] Document specific failure modes (wrong output type vs. timeout vs. error)

---

*Last updated: 2026-02-19 | Data source: `.claude/performance_grades/` and `.claude/performance_grades_backup/`*
