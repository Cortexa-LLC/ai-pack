# Cartographer Role — Model Performance

**Role:** `cartographer`  
**Description:** Codebase mapping and dependency analysis specialist  
**Configured Default:** (not specified in role spec)  
**Last Updated:** 2026-02-19

---

## Model Performance Grades

| Model | Grade | Success Rate | Attempts | Confidence | Project | Last Tested | Notes |
|-------|-------|-------------|----------|------------|---------|-------------|-------|
| claude-sonnet-4-5 | **F** | 25% (1/4) | 4 | 0.20 (low) | xasm++ | 2026-02-14 | **Worst absolute success rate in dataset** |
| gpt-4o-mini | ? | — | 0 | — | — | — | Not benchmarked |

> ⚠️ **Alert:** The cartographer role shows the **lowest success rate** across all benchmarked role × model combinations: 25% (1 out of 4 attempts).

---

## Analysis

### Why Cartographer Struggles
The cartographer role requires:
- **Systematic codebase traversal** — mapping every file, module, dependency
- **Relationship identification** — finding how components connect
- **Accurate representation** — outputting correct dependency graphs/maps
- **Completeness** — missing parts of the map is a failure

For the `xasm++` project specifically:
- Assembly codebase with non-standard structure
- Multiple subsystems with non-obvious dependencies
- Assembly-specific import/include patterns differ from high-level languages

### Is This a Model Problem or Project Problem?
With 4 attempts and 25% success, two explanations are plausible:
1. **Model limitation:** claude-sonnet-4-5 lacks the systematic rigor needed for complete codebase mapping
2. **Project complexity:** xasm++ has an unusually complex dependency structure for its size

Likely both contribute. More benchmarks on simpler projects would disambiguate.

---

## Role-Specific Model Considerations

### What the Cartographer Needs in a Model
1. **Systematic exhaustiveness** — must check every file, not just obvious ones
2. **Structured output** — maps must be in parseable formats (JSON, Mermaid, etc.)
3. **Large context** — big codebases require holding much in context
4. **Attention to detail** — one missed dependency breaks the map

These requirements suggest **larger models perform better** for this role.

### Recommended Model Hierarchy

```
claude-sonnet-4-5 → claude-sonnet-4-6 → claude-opus-4-6
```

Skip cheap models — incomplete maps are worse than no maps. The cost of a bad dependency map (wrong decisions downstream) exceeds the cost of a premium model.

---

## Task Types and Model Fit

| Task Type | Recommended Model | Notes |
|-----------|------------------|-------|
| Small project (<20 files) | claude-sonnet-4-5 | Manageable scope |
| Medium project (20-100 files) | claude-sonnet-4-6 | Needs more reliable coverage |
| Large project (100+ files) | claude-opus-4-6 | Complex enough to justify premium |
| Dependency update impact | claude-sonnet-4-5 | Scoped analysis |
| Full codebase audit | claude-opus-4-6 | Completeness critical |

---

## Benchmark History

### 2026-02-14: claude-sonnet-4-5 on xasm++ (Cartographer)
- **Result:** F (25% success rate) — lowest in dataset
- **Sample:** 4 attempts, 1 success, 3 failures
- **Confidence:** 0.20 (low)
- **Context:** xasm++ — niche assembly language project

---

## Gaps in Data

- [ ] No benchmarks on standard software projects (web apps, APIs, etc.)
- [ ] No benchmarks for claude-sonnet-4-6 or claude-opus-4-6
- [ ] Unclear what "failure" means for cartographer: incomplete map? wrong relationships? timeout?
- [ ] No quality metric: what % of dependencies correctly identified?

---

## Priority for Future Benchmarks

Given the lowest success rate, the cartographer role should be **prioritized for benchmarking** on non-assembly projects. This will determine if the issue is model × project specific or a fundamental model limitation.

---

*Last updated: 2026-02-19 | Grade data from `.claude/performance_grades_backup/`*
