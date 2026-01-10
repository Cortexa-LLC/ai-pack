# AI-Pack Skills

This directory contains **Skills** that Claude Code automatically discovers and applies based on user requests.

## What Are Skills?

Skills are AI-triggered capabilities. Unlike slash commands (which you invoke manually like `/ai-pack engineer`), Skills are **automatically activated** by Claude when it detects relevant keywords or patterns in your request.

## Provided Skills

The ai-pack framework provides Skills for each role:

- `orchestrator/` - Auto-activates for complex coordination tasks
- `engineer/` - Auto-activates for implementation tasks
- (Add similar skills for other roles as needed)

## Skill Pattern

Each skill directory contains a `SKILL.md` file with:

```yaml
---
name: skill-name
description: When to use this skill. Keywords trigger auto-activation.
---

# Skill Content

Instructions for Claude when this skill activates.
```

## Creating Additional Skills

To add skills for other ai-pack roles, create directories matching this pattern:

```
skills/
├── orchestrator/SKILL.md  ✅ Provided
├── engineer/SKILL.md      ✅ Provided
├── reviewer/SKILL.md      ⚠️ TODO: Add if desired
├── tester/SKILL.md        ⚠️ TODO: Add if desired
├── inspector/SKILL.md     ⚠️ TODO: Add if desired
├── architect/SKILL.md     ⚠️ TODO: Add if desired
├── designer/SKILL.md      ⚠️ TODO: Add if desired
└── product-manager/SKILL.md ⚠️ TODO: Add if desired
```

**Note:** The provided `orchestrator` and `engineer` skills demonstrate the pattern. You can add skills for other roles as needed by copying the pattern and referencing the appropriate role definition from `.ai-pack/roles/*.md`.

## When Skills vs Commands

**Use Skills when:**
- You want auto-activation based on keywords
- Behavior should apply across many requests
- Pattern matching is sufficient

**Use Commands (`/ai-pack <command>`) when:**
- Explicit invocation needed
- User needs to see available actions
- Discoverable via `/help`

## Skill Activation Examples

**Orchestrator Skill:**
- "Orchestrate this feature implementation"
- "Coordinate building these three components"
- "Break down and delegate this work"

**Engineer Skill:**
- "Implement the login feature"
- "Write code for user authentication"
- "Build the API endpoint"

## References

- Skills Documentation: https://code.claude.com/docs/en/skills.md
- Role Definitions: `.ai-pack/roles/*.md`
