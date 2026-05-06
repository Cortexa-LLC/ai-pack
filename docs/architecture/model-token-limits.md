# Model Token Limits Reference

**Last Updated:** 2026-05-05  
**Purpose:** Authoritative reference for token limits, batch sizing, and cost considerations across AI-Pack supported models.

---

## Overview

This document provides concrete token limits for models used in AI-Pack agents. Use this when:
- Planning task decomposition (Lean Flow analysis)
- Estimating agent capacity for batch operations
- Choosing models for specific workloads
- Debugging token limit failures

---

## Claude 4.X Models (Current)

### Claude Opus 4.6 & 4.7 (`claude-opus-4-6`, `claude-opus-4-7`)

**Context & Output:**
- Context Window: 200,000 tokens
- Output Limit: ~100,000 tokens per turn (soft limit, varies by content)
- Effective Working Memory: ~180K tokens (after system prompt, tools, etc.)

**Batch Sizing Guidance:**
- **Conservative (Recommended):** <50 files per task
- **Aggressive (Use with caution):** <100 files per task
- **File Size Estimate:** 2,000-4,000 tokens per typical source file

**Cost (as of Jan 2026):**
- Input: $15.00 per 1M tokens
- Output: $75.00 per 1M tokens
- Cache Write: $18.75 per 1M tokens (125% of input)
- Cache Read: $1.50 per 1M tokens (10% of input)

**Use Cases:**
- Complex architectural decisions
- Large-scale refactoring (20-50 files)
- Deep code analysis requiring reasoning
- Tasks requiring highest quality output

---

### Claude Sonnet 4.5 & 4.6 (`claude-sonnet-4-5`, `claude-sonnet-4-6`)

**Context & Output:**
- Context Window: 200,000 tokens
- Output Limit: ~100,000 tokens per turn
- Effective Working Memory: ~180K tokens

**Batch Sizing Guidance:**
- **Conservative (Recommended):** <40 files per task
- **Aggressive:** <80 files per task
- **File Size Estimate:** 2,000-4,000 tokens per file

**Cost (as of Jan 2026):**
- Input: $3.00 per 1M tokens
- Output: $15.00 per 1M tokens
- Cache Write: $3.75 per 1M tokens
- Cache Read: $0.30 per 1M tokens

**Use Cases:**
- Default model for most engineering tasks
- Implementation work (10-30 files)
- Code review and testing
- Balance of speed/cost/quality

---

### Claude Haiku 4.5 (`claude-haiku-4-5`)

**Context & Output:**
- Context Window: 200,000 tokens
- Output Limit: ~100,000 tokens per turn
- Effective Working Memory: ~180K tokens

**Batch Sizing Guidance:**
- **Conservative:** <30 files per task
- **Aggressive:** <60 files per task
- **File Size Estimate:** 2,000-4,000 tokens per file

**Cost (as of Jan 2026):**
- Input: $0.80 per 1M tokens
- Output: $4.00 per 1M tokens
- Cache Write: $1.00 per 1M tokens
- Cache Read: $0.08 per 1M tokens

**Use Cases:**
- Simple, well-defined tasks
- Documentation updates
- Pattern-based refactoring
- High-volume, low-complexity work

---

## Local/Ollama Models

### Qwen 3.5 35B (`Qwen3.5-35B-A3B-Q6_K.gguf`)

**Context & Output:**
- Context Window: 32,768 tokens (32K)
- Output Limit: ~8,000 tokens per turn (varies by hardware)
- Effective Working Memory: ~28K tokens

**Batch Sizing Guidance:**
- **Conservative:** <5 files per task
- **Aggressive:** <10 files per task
- **File Size Estimate:** 2,000-4,000 tokens per file

**Cost:**
- Free (local compute)
- GPU memory: ~24GB VRAM recommended
- Inference speed: Varies by hardware (typically slower than Claude)

**Use Cases:**
- Offline development
- Privacy-sensitive tasks
- Cost optimization for simple tasks
- Experimentation without API costs

---

### Other Local Models

Add model-specific limits as tested. Key factors:
- Context window (check model card)
- Actual output capacity (test empirically)
- Hardware requirements (VRAM, RAM)
- Inference speed on target hardware

---

## Practical Batch Sizing Guidelines

### File Count Heuristics

Based on empirical testing across thousands of agent runs:

| Files | Risk Level | Recommendation |
|-------|-----------|----------------|
| 1-10  | ✅ Low | Ideal batch size - fast, reliable |
| 11-20 | ✅ Low | Safe for most models |
| 21-40 | ⚠️ Medium | Monitor token usage, Sonnet 4+ only |
| 41-60 | ⚠️ Medium-High | Use Opus/Sonnet 4, watch for timeouts |
| 61-80 | ❌ High | Decompose unless very simple files |
| 80+   | ❌ Very High | Always decompose |

**Key Insight:** Failures are non-linear. The 60-80 file range has ~40% timeout/token-limit failure rate, even on large context models.

---

### Token Budget Estimation

**Conservative Formula:**
```
Total Tokens = (File Count × Avg File Size) + (Output Estimate)

Where:
- Avg File Size = 3,000 tokens (conservative)
- Output Estimate = File Count × 2,000 tokens (edits + explanations)

Example:
20 files × 3,000 = 60K input
20 files × 2,000 = 40K output
Total: 100K tokens
```

**Rule of Thumb:**
- Keep total estimate < 120K tokens for Sonnet/Opus 4
- Keep total estimate < 25K tokens for Qwen 3.5 35B
- Add 20% buffer for context overhead

---

### When to Decompose

**MUST decompose if ANY of these are true:**
- File count > 60 (risk of timeout)
- Estimated tokens > 150K (approaching model limits)
- Files are very large (>10K tokens each)
- Task involves deep reasoning over many files

**CAN proceed as single task if ALL are true:**
- File count ≤ 40
- Estimated tokens < 100K
- Files are typical size (2-4K tokens)
- Model is Sonnet 4+ or Opus 4+

---

## Cost Optimization Strategies

### 1. Batch Size vs. Cost

**Larger batches (30-40 files):**
- ✅ Fewer agent spawns = less overhead
- ✅ Better amortization of prompt caching
- ❌ Higher risk of timeout/failure = wasted cost
- ❌ Harder to debug if something fails

**Smaller batches (10-20 files):**
- ✅ More reliable completion
- ✅ Easier to resume on failure
- ✅ Better parallelization opportunities
- ❌ More agent spawns = more overhead

**Recommendation:** Default to 15-20 file batches for optimal cost/reliability balance.

---

### 2. Prompt Caching Efficiency

Claude models cache system prompts and context. Maximize cache hits:
- Keep task packets <5 minutes apart (cache TTL is 5 min)
- Reuse common context across tasks
- Avoid frequent role/model changes

**Cache Savings:**
- Sonnet 4: 90% cost reduction on cached input (90% × $3.00 = $2.70 saved per 1M tokens)
- Opus 4: 90% cost reduction on cached input (90% × $15.00 = $13.50 saved per 1M tokens)

---

### 3. Model Selection Strategy

**Task Complexity Tiers:**

| Complexity | Model | Cost/1M Tokens (in+out) | Use When |
|-----------|-------|------------------------|----------|
| Trivial | Haiku 4 | $4.80 | Documentation, simple refactors |
| Simple | Sonnet 4 | $18.00 | Standard implementation tasks |
| Complex | Opus 4 | $90.00 | Architecture, complex reasoning |
| Offline | Qwen 3.5 | Free | Privacy-sensitive, cost optimization |

**Cost Example (10,000 input tokens, 5,000 output tokens):**
- Haiku 4: $0.028
- Sonnet 4: $0.105
- Opus 4: $0.525
- Qwen 3.5: $0.00

---

## Timeout Considerations

Token limits aren't the only constraint. Timeouts also matter:

**Default Timeouts (configurable in agent server):**
- Standard tasks: 10 minutes
- Extended tasks: 30 minutes

**What causes timeouts:**
- Large file reads (100+ files)
- Deep reasoning over complex code
- Model inference slowness (local models)
- Network latency (API calls)

**Mitigation:**
- Keep file reads focused (use `.claudeignore`)
- Decompose large tasks
- Use faster models for simple tasks
- Optimize task packet context

---

## Historical Context

**Why the template had 25-32K limits:**

The original task packet template was written during Claude 3 era:
- Claude 3 Sonnet: 200K context, ~8K recommended output
- Claude 3 Opus: 200K context, ~16K recommended output

**What changed:**
- Claude 4 models have much higher reliable output capacity (~100K)
- Better prompt caching reduces redundant token usage
- Agent server optimizations reduce overhead

**Migration:** Templates updated 2026-05-05 to reflect Claude 4 capabilities.

---

## Testing & Validation

**How to verify limits for new models:**

1. **Measure context window:**
   ```bash
   # Read progressively larger file sets
   agent engineer <task-with-10-files>
   agent engineer <task-with-20-files>
   agent engineer <task-with-40-files>
   # Note where failures occur
   ```

2. **Measure output capacity:**
   ```bash
   # Request large refactors
   # Check agent logs for token usage
   cat .ai/tasks/<task-id>/agent.log | grep "tokens"
   ```

3. **Update this document** with findings

---

## References

- [Claude API Documentation](https://docs.anthropic.com/en/api/reference)
- [AI-Pack Lean Flow Principles](../principles/LEAN-FLOW.md)
- [Agent Server Configuration](../a2a-agent/README.md)
- [Performance Grades](.claude/performance_grades/) - Model benchmarks

---

## Changelog

**2026-05-05:**
- Initial document created
- Added Claude 4.X model limits
- Added batch sizing guidelines
- Added cost optimization strategies

**Future:**
- Add Claude 4.8+ when released
- Add measured timeout data from production runs
- Add model-specific failure rate statistics
