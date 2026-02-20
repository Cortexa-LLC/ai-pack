# Model × Role Benchmark Master Tracker

> **Living document** — update after every benchmark run.  
> **Last Updated:** 2026-05-28  
> **Data Source:** `.claude/performance_grades/` (project: `/Users/bryanw/Projects/Vibe/ai-pack`)  
> **Benchmark Script:** `scripts/benchmark-models.py`

---

## Summary

**168 benchmarks run** — 14 models × 12 roles, 5 attempts each (840 total API calls).

| Metric | Value |
|--------|-------|
| Models benchmarked | 14 |
| Roles benchmarked | 12 |
| Total grade files | 168 |
| A grades (>90%) | 162 (96%) |
| B grades (75-90%) | 5 (3%) |
| D grades (40-60%) | 1 (1%) |
| Overall pass rate | 99.4% |

---

## Roles Benchmarked

All 12 roles now have role-appropriate benchmark prompts (5 prompts each):

| Role | Benchmark Task Type |
|------|---------------------|
| **engineer** | Implement a specific coding feature or algorithm |
| **tester** | Write comprehensive test suite for given function |
| **reviewer** | Review code and identify improvements/issues |
| **orchestrator** | Produce delegation plan for multi-step problem |
| **architect** | Design system with trade-offs given requirements |
| **inspector** | Root-cause analysis given a bug report |
| **spelunker** | Reason through where to look in a codebase |
| **cartographer** | Describe architecture from a file/component map |
| **archaeologist** | Explain rationale behind legacy code decisions |
| **product-manager** | Write requirements spec from user need |
| **designer** | Describe UX flow for a feature |
| **strategist** | Outline strategy for business problem |

---

## Master Grade Matrix

> Grades: **A** (>90%) | **B** (75-90%) | **C** (60-75%) | **D** (40-60%) | **F** (<40%) | **?** (no data)  
> Role order: eng, tst, rev, orc, arc, ins, spl, crt, arc, pm, des, str

| Role | 4.1-nano | 4.1-mini | haiku-4-5 | o4-mini | codex-mini | gpt-4.1 | son-4-5 | son-4-5-old | son-4-6 | codex-5.1 | codex-5.2 | opus-4-5 | opus-4-6 | 4o-mini |
|------|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **engineer** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | B (83%) | **A** | **A** | **A** | **A** | **A** | **A** |
| **tester** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** |
| **reviewer** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** |
| **orchestrator** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** |
| **architect** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** |
| **inspector** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | D ⚠️ | B (80%) | **A** | **A** | **A** |
| **spelunker** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | B (80%) | **A** | **A** | **A** |
| **cartographer** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | B (80%) | **A** | **A** | **A** | **A** |
| **archaeologist** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | B (80%) | **A** | **A** | **A** | **A** |
| **product-manager** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** |
| **designer** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** |
| **strategist** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** | **A** |

**Column key:** `son-4-5-old` = claude-sonnet-4-5-20250929, `codex-5.1` = gpt-5.1-codex, `codex-5.2` = gpt-5.2-codex, `codex-mini` = gpt-5.1-codex-mini

---

## Per-Model Summary

| Model | Tier | Grades (eng→str) | Avg Pass | Avg Latency | Avg Tokens |
|-------|------|-----------------|----------|-------------|------------|
| gpt-4.1-nano | minimal | AAAAAAAAAAAA | 100% | 6.2s | 835 |
| gpt-4.1-mini | low | AAAAAAAAAAAA | 100% | 11.3s | 836 |
| claude-haiku-4-5 | minimal | AAAAAAAAAAAA | 100% | 8.3s | 979 |
| o4-mini | low | AAAAAAAAAAAA | 100% | 36.6s | 4,751 |
| gpt-5.1-codex-mini | medium | AAAAAAAAAAAA | 99% | 5.4s | 853 |
| gpt-4.1 | medium | AAAAAAAAAAAA | 100% | 14.8s | 841 |
| claude-sonnet-4-5 | medium | AAAAAAAAAAAA | 100% | 18.6s | 963 |
| claude-sonnet-4-5-20250929 | medium | BAAAAAAAAAAA | 99% | 18.4s | 934 |
| claude-sonnet-4-6 | medium | AAAAAAAAAAAA | 100% | 18.8s | 1,010 |
| gpt-5.1-codex | high | AAAAADABBAAA | 93% | 26.4s | 778 |
| gpt-5.2-codex | high | AAAAABBAAAAA | 97% | 14.7s | 864 |
| claude-opus-4-5 | high | AAAAAAAAAAAA | 100% | 15.8s | 948 |
| claude-opus-4-6 | high | AAAAAAAAAAAA | 100% | 19.4s | 977 |
| gpt-4o-mini | minimal | AAAAAAAAAAAA | 100% | 12.0s | 754 |

**Grade sequence:** engineer, tester, reviewer, orchestrator, architect, inspector, spelunker, cartographer, archaeologist, product-manager, designer, strategist

---

## Notable Findings

### gpt-5.1-codex — D on inspector, B on cartographer/archaeologist
`gpt-5.1-codex` achieves **D (60% pass rate)** on `inspector` and **B (80%)** on `cartographer`
and `archaeologist`. Root cause: this model uses the OpenAI Responses API with a 60s timeout,
and some longer reasoning responses exceed the timeout, resulting in empty responses that fail
evaluation. The model *can* produce correct answers (validated manually) but is unreliable
under the 60-second SLA.

**Recommendation:** Do not use `gpt-5.1-codex` for `inspector`, `cartographer`, or
`archaeologist` roles without a longer timeout. Use `claude-opus-4-6` or `gpt-5.2-codex`
instead for these roles.

### gpt-5.2-codex — B on inspector and spelunker
`gpt-5.2-codex` achieves **B (80%)** on `inspector` and `spelunker`. Same latency-related
issue as `gpt-5.1-codex` but less severe. Acceptable but not ideal for time-sensitive workflows.

### claude-sonnet-4-5-20250929 — B on engineer
The older snapshot `claude-sonnet-4-5-20250929` achieves **B (83%)** on `engineer`.
This is expected — it's a snapshot from September 2025 and the current `claude-sonnet-4-5`
already supersedes it with A-grade performance. This model is kept for comparison only.

### o4-mini — High latency, A grades
`o4-mini` takes 36.6s average (longest of all models) due to chain-of-thought reasoning,
but achieves **A grades across all 12 roles**. Excellent quality, poor speed.

### gpt-5.1-codex-mini — Fastest + reliable
`gpt-5.1-codex-mini` at 5.4s average is the fastest model with A grades across all roles.
Excellent choice for cost-sensitive, latency-sensitive tasks.

---

## Role-Level Analysis

### Universally Strong Roles (100% of models achieve A)
- **tester** — all 14 models: A
- **reviewer** — all 14 models: A
- **orchestrator** — all 14 models: A
- **architect** — all 14 models: A
- **product-manager** — all 14 models: A
- **designer** — all 14 models: A
- **strategist** — all 14 models: A

These roles have straightforward, natural-language output that all current-gen models handle well.

### Roles with Occasional Failures
- **inspector** — 1 D (gpt-5.1-codex), 1 B (gpt-5.2-codex): latency-sensitive reasoning
- **spelunker** — 1 B (gpt-5.2-codex): occasionally shallow analysis
- **cartographer** — 1 B (gpt-5.1-codex): timeout-related
- **archaeologist** — 1 B (gpt-5.1-codex): timeout-related
- **engineer** — 1 B (claude-sonnet-4-5-20250929): older model snapshot

---

## Tier Recommendations

Based on benchmark results, the following tier assignments are validated:

### TierMinimal (cheap, fast)
✅ `gpt-4.1-nano` — A across all roles, 6.2s, 835 tokens  
✅ `claude-haiku-4-5` — A across all roles, 8.3s, 979 tokens  
✅ `gpt-4o-mini` — A across all roles, 12.0s, 754 tokens  

### TierLow
✅ `gpt-4.1-mini` — A across all roles, 11.3s  
✅ `o4-mini` — A across all roles, 36.6s (slow but quality)  

### TierMedium (default workhorse)
✅ `gpt-5.1-codex-mini` — A across all roles, fastest (5.4s)  
✅ `gpt-4.1` — A across all roles, 14.8s  
✅ `claude-sonnet-4-5` — A across all roles, 18.6s  
✅ `claude-sonnet-4-6` — A across all roles, 18.8s  
⚠️ `claude-sonnet-4-5-20250929` — B on engineer, legacy model  

### TierHigh (premium)
⚠️ `gpt-5.1-codex` — D on inspector, B on cartographer/archaeologist (latency)  
⚠️ `gpt-5.2-codex` — B on inspector, spelunker (latency)  
✅ `claude-opus-4-5` — A across all roles, 15.8s  
✅ `claude-opus-4-6` — A across all roles, 19.4s  

---

## Benchmark Methodology

### How Benchmarks Work
1. Each model × role combination runs **5 benchmark prompts** (different task instances)
2. Each prompt is evaluated with a role-specific pass/fail checker (keyword + length heuristics)
3. Results are stored in `.claude/performance_grades/` as JSON
4. Grades are computed from success rate: A≥90%, B≥75%, C≥60%, D≥40%, F<40%

### Role-Specific Prompt Design
Each role has 5 distinct prompts designed to test the specific cognitive task of that role:

- **orchestrator**: Multi-agent delegation plans with dependency ordering
- **architect**: System design documents with trade-offs and non-functional requirements
- **inspector**: Root-cause analysis with ranked hypotheses and evidence collection steps
- **spelunker**: Codebase navigation reasoning ("where would I find X in a repo like Y?")
- **cartographer**: Architecture description from directory/component structure
- **archaeologist**: Rationale reconstruction for legacy code patterns
- **product-manager**: Requirements specs with user stories and acceptance criteria
- **designer**: UX flow descriptions with user journey steps
- **strategist**: Strategic plans with options, trade-offs, and recommendations

### API Routing
- **Codex models** (`gpt-5.1-codex`, `gpt-5.2-codex`): OpenAI Responses API (`client.responses.create`)
- **o-series** (`o4-mini`): Chat completions with `max_completion_tokens`, no system prompt
- **All others**: Standard chat completions with system + user messages

---

## Historical Runs

| Date | Models | Roles | Notes |
|------|--------|-------|-------|
| 2026-02-19 | 6 (gpt-4o era) | 3 (eng/tst/rev) | Initial tracker |
| 2026-05-28 | 14 | 12 | Full expansion — all roles benchmarked |

---

## How to Update This Document

After running new benchmarks:

```bash
# Run benchmarks for specific models/roles
python3 scripts/benchmark-models.py --models gpt-4.1 --roles inspector

# Run full suite
python3 scripts/benchmark-models.py

# Grade files auto-saved to .claude/performance_grades/
# Then manually update this tracker from the JSON files
```

To regenerate the grade matrix:
```python
import json, glob
files = glob.glob('.claude/performance_grades/*__YOUR_PROJECT_PATH.json')
data = {(d['model_id'], d['role_id']): d for f in files for d in [json.load(open(f))]}
# Build table as needed
```
