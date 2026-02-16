# Role Metadata Template

This document defines the standard metadata format for all role definitions in the ai-pack framework.

## Required Metadata

Every role definition MUST include these metadata fields at the top of the markdown file:

```markdown
# Role Name

**Version:** X.Y.Z
**Last Updated:** YYYY-MM-DD
**Complexity:** [Low|Medium|High|Very High]
**Recommended Model:** [model-name]
**Cost Tier:** [Minimal|Low|Medium|High|Premium]
**Monthly Cost Estimate:** $X.XX-X.XX per 1M tokens
**Best For:** [comma-separated use cases]
**Avoid For:** [comma-separated anti-patterns]
```

## Metadata Field Definitions

### Version
- Format: Semantic versioning (MAJOR.MINOR.PATCH)
- Example: `1.2.0`
- Purpose: Track role definition changes over time

### Last Updated
- Format: ISO 8601 date (YYYY-MM-DD)
- Example: `2026-02-15`
- Purpose: Know when role was last modified

### Complexity
- Values: `Low`, `Medium`, `High`, `Very High`
- Purpose: Indicate cognitive/reasoning requirements
- Guidance:
  - **Low**: Simple pattern following, straightforward tasks
  - **Medium**: Some decision-making, standard engineering work
  - **High**: Complex reasoning, architectural decisions, coordination
  - **Very High**: Strategic planning, multi-agent orchestration, deep analysis

### Recommended Model
- Values: Specific model identifier (e.g., `gpt-4o-mini`, `claude-sonnet-4-5`)
- Purpose: Optimal model for this role's complexity level
- Guidance:
  - Low complexity: `gpt-4o-mini`, `claude-haiku-4-5`
  - Medium complexity: `gpt-4o`, `claude-sonnet-4-5`
  - High complexity: `claude-sonnet-4-5`, `claude-opus-4-6`
  - Very High complexity: `claude-opus-4-6`

### Cost Tier
- Values: `Minimal`, `Low`, `Medium`, `High`, `Premium`
- Purpose: Budget categorization for cost management
- Guidance:
  - **Minimal**: < $1/1M tokens (mini models)
  - **Low**: $1-5/1M tokens (haiku, small models)
  - **Medium**: $5-10/1M tokens (mid-tier models)
  - **High**: $10-20/1M tokens (sonnet, gpt-4)
  - **Premium**: > $20/1M tokens (opus, o-series)

### Monthly Cost Estimate
- Format: `$X.XX-X.XX per 1M tokens` (input-output range)
- Example: `$0.15-0.60 per 1M tokens`
- Purpose: Real pricing for budgeting and cost awareness
- Calculation: Based on recommended model's input/output token costs

### Best For
- Format: Comma-separated list of use cases
- Example: `Simple implementations, refactoring, test writing, documentation`
- Purpose: Clear guidance on when to use this role
- Should include 3-6 specific use cases

### Avoid For
- Format: Comma-separated list of anti-patterns
- Example: `Complex architecture, deep debugging, security-critical code`
- Purpose: Clear guidance on when NOT to use this role
- Should include 3-6 specific scenarios

## Example: Engineer Role

```markdown
# Engineer Role

**Version:** 1.3.0
**Last Updated:** 2026-01-31
**Complexity:** Medium
**Recommended Model:** claude-sonnet-4-5
**Cost Tier:** High
**Monthly Cost Estimate:** $3.00-15.00 per 1M tokens
**Best For:** Feature implementation, bug fixes, code refactoring, test creation
**Avoid For:** Project planning, multi-agent orchestration, strategic decisions

## Role Overview

The Engineer is an implementation specialist responsible for executing specific, well-defined tasks...
```

## Example: Orchestrator Role

```markdown
# Orchestrator Role

**Version:** 1.2.0
**Last Updated:** 2026-01-18
**Complexity:** Very High
**Recommended Model:** claude-opus-4-6
**Cost Tier:** Premium
**Monthly Cost Estimate:** $5.00-25.00 per 1M tokens
**Best For:** Task breakdown, multi-agent coordination, project planning, architectural decisions
**Avoid For:** Simple code edits, straightforward implementations, basic bug fixes

## Role Overview

The Orchestrator is a high-level coordinator responsible for breaking down complex work...
```

## Example: Engineer-Cheap Role

```markdown
# Cost-Optimized Engineer Agent

**Version:** 2.0
**Last Updated:** 2026-02-15
**Complexity:** Low
**Recommended Model:** gpt-4o-mini
**Cost Tier:** Minimal
**Monthly Cost Estimate:** $0.15-0.60 per 1M tokens
**Best For:** Simple implementations, refactoring, test writing, documentation
**Avoid For:** Complex architecture, deep debugging, security-critical code

You are a software engineer using **GPT-4o-mini** for maximum cost efficiency...
```

## Implementation Notes

### Parser Behavior
The `parseRoleMarkdown()` function extracts these metadata fields using pattern matching:
- Looks for `**Key:** value` patterns
- Stores in `AgentConfig.Metadata` map
- Keys are normalized to lowercase with underscores

### Cost Optimization Strategy
Use role metadata to implement smart routing:
1. **Route simple tasks to cheap agents** (engineer-cheap, inspector-lite)
2. **Route complex tasks to capable agents** (orchestrator, architect)
3. **Allow escalation paths** (engineer-cheap → engineer → orchestrator)
4. **Track actual costs** vs recommended costs in metrics

### Project Overrides
Projects can override role metadata in `.ai/roles/[role-name].md`:
```markdown
# Engineer Role (Project Override)

**Recommended Model:** gpt-4o-mini  # Override: use cheaper model for this project
**Cost Tier:** Minimal              # Override: reduce cost tier

## Project-Specific Additions
- Use our custom validation framework
- Follow project coding standards
```

## Updating Existing Roles

To add metadata to existing roles:
1. Add all required metadata fields at the top
2. Use role complexity to determine recommended model
3. Look up current pricing for cost estimates
4. Document clear use cases and anti-patterns
5. Update version and last updated date

## Validation

Roles without required metadata will trigger warnings but won't fail (graceful degradation).
However, best practice is to include ALL metadata fields for optimal cost management.
