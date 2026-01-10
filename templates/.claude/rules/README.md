# AI-Pack Modular Rules

This directory contains modular rules that Claude Code automatically loads based on file paths.

## What Are Rules?

Rules are instructions that Claude follows for specific file patterns. They're automatically loaded when working in matching paths, providing context-specific guidance.

## Provided Rules

- **gates.md** - Mandatory quality gates (all paths)
- **task-packets.md** - Task packet requirements (all paths)
- **workflows.md** - Workflow selection guide (all paths)

## Rule Structure

Each rule file has frontmatter specifying which paths it applies to:

```yaml
---
paths: **/*
---

# Rule Content

Instructions that apply to the specified paths.
```

## Path Patterns

- `**/*` - All files (global rules)
- `src/**/*.py` - Python files in src/
- `tests/**/*` - All files in tests/
- `*.md` - All markdown files

## Custom Rules

Projects can add custom rules by creating additional `.md` files in this directory:

```
rules/
├── gates.md              ✅ Provided (global)
├── task-packets.md       ✅ Provided (global)
├── workflows.md          ✅ Provided (global)
├── project-specific.md   ⚠️ Add if needed
└── language-specific.md  ⚠️ Add if needed
```

## Example: Language-Specific Rule

Create `rules/python-standards.md`:

```yaml
---
paths: **/*.py
---

# Python Coding Standards

When working with Python files:
- Follow PEP 8
- Use type hints
- Docstrings required
- ... etc
```

## Rule Priority

Rules are applied in this order:
1. Framework rules (from ai-pack)
2. Project rules (from .claude/rules/)
3. More specific path patterns override general ones

## References

- Rules Documentation: https://code.claude.com/docs/en/memory.md#modular-rules-with-claude-rulesmd
- AI-Pack Gates: `.ai-pack/gates/`
- AI-Pack Workflows: `.ai-pack/workflows/`
