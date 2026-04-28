---
sidebar_position: 1
title: Deployment Model
---

# Deployment Model

AI-Pack is designed as a git submodule that projects include at `.ai-pack/`.

## Directory Structure

```text
your-project/
├── .ai-pack/                        # Git submodule (read-only shared pack)
│   ├── quality/                     # Clean code standards
│   ├── gates/                       # Quality gates
│   ├── roles/                       # Agent roles
│   ├── workflows/                   # Development workflows
│   └── templates/                   # Task-packet templates
│
├── .beads/                          # task memory system (committed)
│   ├── issues.jsonl                 # Git-tracked task database
│   └── *.db                         # Local SQLite cache (git-ignored)
│
├── .ai/                             # Local workspace (your project)
│   ├── tasks/                       # Active task packets (temporary)
│   └── repo-overrides.md           # Optional project-specific deltas
│
├── docs/                            # Permanent documentation (committed)
└── CLAUDE.md                        # Bootstrap instructions for AI
```text

## Key Concepts

- **`.ai-pack/`** - Git submodule containing shared standards (READ-ONLY)
- **`.beads/`** - task memory system (PROJECT-SPECIFIC, COMMITTED)
- **`.ai/`** - Local workspace for task state (PROJECT-SPECIFIC, TEMPORARY)
- **`docs/`** - Permanent documentation (PROJECT-SPECIFIC, COMMITTED)

## Critical Invariants

- ✅ Task packets go in `.ai/tasks/` (never in `.ai-pack/`)
- ✅ `.ai-pack/` is read-only shared framework
- ✅ `.beads/issues.jsonl` is committed (source of truth)
- ✅ `.beads/*.db` is git-ignored (local cache only)
- ✅ Framework improvements happen in ai-pack repo
