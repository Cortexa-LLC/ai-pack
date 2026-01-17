# Commit Message Guide

This repository uses structured commit messages to make changes easy to scan
and audit. Use ASCII only.

## Subject Line

- Imperative voice, sentence case.
- 72 characters or fewer.
- No trailing period.
- Prefer these prefixes when appropriate:
  - feat:
  - fix:
  - chore:
  - docs:
  - CRITICAL:
  - CRITICAL FIX:
- Reverts use: Revert "original subject"

## Body (Required for Significant Changes)

For significant changes, include a multi-line body with 3+ lines. Use labeled
sections and blank lines between them.

Recommended labels (use what fits the change):
- Issue: or Problem:
- Observed:
- Solution:
- Impact:
- Without this:
- With this:
- Reference:

Use short sentences or bullet points. Avoid escaped "\n" in a single line.

## Examples

Subject only:

```
feat: add Codex integration scaffolding
```

Structured body:

```
CRITICAL: add background agent permission verification

Issue: Background agents complete but fail to persist artifacts due to
missing Write(*) permissions.

Observed:
- PRDs written but not persisted
- ADRs referenced but missing

Solution:
- Add pre-flight permission checks in orchestrator workflow
- Block background agents if Write(*) is missing

Impact:
- Prevents silent failures
- Provides clear remediation steps

Reference:
- templates/.claude/PERMISSIONS.md
- templates/.claude/settings.json
```
