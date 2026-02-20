# Reviewer Role — Model Performance

**Role:** `reviewer`  
**Description:** Code quality review and feedback specialist  
**Configured Variants:** reviewer.md (default), reviewer-opus.md, reviewer-sonnet.md, reviewer-codex.md, reviewer-codex-large.md, reviewer-codex-mini.md  
**Last Updated:** 2026-02-19

---

## Model Performance Grades

| Model | Grade | Success Rate | Attempts | Confidence | Project | Last Tested | Notes |
|-------|-------|-------------|----------|------------|---------|-------------|-------|
| claude-sonnet-4-5 | **F** | 33% (1/3) | 3 | 0.15 (low) | xasm++ | 2026-02-14 | Very low confidence |
| claude-opus-4-6 | ? | — | 0 | — | — | — | Has dedicated variant (reviewer-opus.md) |
| claude-sonnet-4-6 | ? | — | 0 | — | — | — | Has dedicated variant (reviewer-sonnet.md) |
| gpt-4o (codex) | ? | — | 0 | — | — | — | Has dedicated variants (reviewer-codex*.md) |

---

## Role Variants

This role has the **most variants** of any role in the system — reflecting that code review quality is highly sensitive to model capability:

| Variant File | Model | Use Case |
|-------------|-------|----------|
| `reviewer.md` | Default | Standard code review |
| `reviewer-opus.md` | claude-opus-4-6 | Premium, thorough review |
| `reviewer-sonnet.md` | claude-sonnet-4-6 | Quality + cost balance |
| `reviewer-codex.md` | gpt-4o (codex) | Fast code-focused review |
| `reviewer-codex-large.md` | gpt-4o-large | Large codebase review |
| `reviewer-codex-mini.md` | gpt-4o-mini (codex) | Budget review |

---

## Choosing the Right Reviewer Variant

| Scenario | Recommended Variant | Reasoning |
|----------|--------------------|-----------| 
| PR before merge (standard) | `reviewer-sonnet.md` | Good quality, reasonable cost |
| Security-sensitive code | `reviewer-opus.md` | Maximum thoroughness |
| Large PR (many files) | `reviewer-codex-large.md` | Handles breadth efficiently |
| Simple refactoring PR | `reviewer-codex-mini.md` | Cost-effective for low-risk changes |
| Architecture changes | `reviewer-opus.md` | Deep reasoning needed |
| Hot fix, time-sensitive | `reviewer-codex.md` | Fastest turnaround |

---

## Role-Specific Model Considerations

### Why Reviewer Has Multiple Variants
Code review quality scales significantly with model capability:
- **gpt-4o-mini:** Catches syntax issues, obvious bugs
- **gpt-4o (codex):** Catches logic errors, bad patterns
- **claude-sonnet:** Catches subtle design issues, security concerns
- **claude-opus:** Catches architectural problems, systemic issues

The team recognized this and created explicit variants rather than relying on dynamic escalation.

---

## Benchmark History

### 2026-02-14: claude-sonnet-4-5 on xasm++ (Reviewer)
- **Result:** F (33% success rate)
- **Sample:** 3 attempts, 1 success, 2 failures
- **Confidence:** 0.15 (very low)
- **Context:** Assembly code review requires specialized knowledge

---

## Gaps in Data

- [ ] No benchmarks for any of the variant-specific models
- [ ] No comparison between reviewer variants on the same PR
- [ ] No quality metric: what makes a review "successful"? (caught real bugs? missed critical issues?)
- [ ] No cost-per-review data across variants

---

*Last updated: 2026-02-19 | Grade data from `.claude/performance_grades_backup/`*
