# AI-Pack Migration Guide

This guide helps you migrate existing projects through ai-pack version updates.

---

## Migrating to v2.0.0 (BREAKING CHANGES)

**Release Date:** 2026-01-14
**Codename:** Consolidation

### What Changed in v2.0.0

**BREAKING:** Agent tracking consolidation removes deprecated `agent-status-tracker.py` system.

### Breaking Changes

- ❌ **REMOVED:** `templates/.claude/scripts/agent-status-tracker.py`
- ❌ **REMOVED:** `.claude/.agent-status.json` file format
- ⚠️ **REQUIRED:** All agent tracking now uses Beads exclusively

### Why This Change?

Eliminates dual tracking system redundancy:
- **Before v2.0:** agent-status-tracker.py + Beads (redundant)
- **After v2.0:** Beads only (single source of truth)

### Migration Steps

**Prerequisites:**
- Project already on v1.1.0+ (Beads initialized)
- If not, migrate to v1.1.0 first (see below)

**Step 1: Update ai-pack submodule**

```bash
cd .ai-pack
git fetch origin
git checkout main
git pull origin main
cd ..
git add .ai-pack
git commit -m "Update ai-pack to v2.0.0"
```

**Step 2: Migrate agent status (if using agent-status-tracker.py)**

```bash
# Only if you have .claude/.agent-status.json file
python3 .ai-pack/scripts/migrate-agent-status-to-beads.py

# Dry run first to see what would happen
python3 .ai-pack/scripts/migrate-agent-status-to-beads.py --dry-run
```

**Step 3: Clean up old files**

```bash
# Remove old status file (if exists)
rm -f .claude/.agent-status.json
rm -f .claude/.agent-status.json.backup
```

**Step 4: Verify migration**

```bash
# Check Beads is working
bd list

# Check agent tasks
/ai-pack agents
# OR
bd list --json | jq '.[] | select(.title | startswith("Agent:"))'
```

### What to Expect

**Orchestrator behavior (unchanged):**
- Still spawns agents with Task tool
- Now creates Beads tasks (no longer uses agent-status-tracker.py)

**Agent monitoring (improved):**
- `/ai-pack agents` queries Beads directly
- Cross-session persistence (agents survive session restarts)
- Git-backed audit trail

**Worker behavior (enhanced):**
- Engineer/Tester/Reviewer can update their Beads tasks
- `bd block <task-id>` when blocked
- `bd close <task-id>` when complete

### Troubleshooting v2.0.0 Migration

**Problem: "agent-status-tracker.py not found"**

✅ **Expected!** This script was removed in v2.0.0. Use Beads:
```bash
# Old (removed)
python3 .claude/scripts/agent-status-tracker.py report

# New (v2.0.0)
/ai-pack agents
bd list --assignee "Engineer-*"
```

**Problem: Old .agent-status.json still exists**

```bash
# Migrate it
python3 .ai-pack/scripts/migrate-agent-status-to-beads.py

# Then remove
rm .claude/.agent-status.json
```

**Problem: No agents showing in /ai-pack agents**

Cause: Orchestrator not creating Beads tasks when spawning agents.

Solution: See Orchestrator role Section 2.13 (Agent Registration Protocol)

---

## Migrating to v1.1.0 (NON-BREAKING)

**Release Date:** 2026-01-12
**Codename:** Beads Integration

### What Changed in v1.1.0

**Beads Task Memory System** replaces session-based TodoWrite with persistent, git-backed task tracking:

- ✅ **Tasks persist across AI sessions** - No more "50 First Dates" problem
- ✅ **Git-backed storage** - Tasks versioned with code in `.beads/issues.jsonl`
- ✅ **Dependency tracking** - Full task graphs with automatic "ready" detection
- ✅ **Cross-machine sync** - Tasks sync via git pull/push
- ✅ **Multi-agent coordination** - Hash-based IDs prevent collisions
- ✅ **Cross-session continuity** - AI agents remember context between conversations

### Key Changes

| Aspect | Before (v1.0.0) | After (v1.1.0) |
|--------|-----------------|----------------|
| Task tracking | TodoWrite (session-only) | Beads (persistent) |
| Task persistence | ❌ Lost on session end | ✅ Survives sessions |
| Dependencies | ❌ Not supported | ✅ Full graphs |
| Storage | In-memory only | `.beads/issues.jsonl` (git) |
| Cross-machine sync | ❌ No | ✅ Via git |
| Multi-agent | ⚠️ Shared context | ✅ Hash IDs |

## Migration Overview

**Migration is NON-BREAKING.** Existing projects continue to work. You migrate to gain persistent task memory.

**Estimated time:** 5-10 minutes

**Steps:**
1. Update `.ai-pack/` submodule to v1.1.0+
2. Install Beads CLI tool
3. Initialize Beads in your project
4. Update `.gitignore` (if needed)
5. Test the integration

## Prerequisites

- Git repository (required - Beads uses git for storage)
- Python 3.7+ (for migration script)
- Write access to project repository

## Migration Methods

### Method 1: Automated Migration (Recommended)

Use the provided Python script for guided migration:

```bash
# From your project root (e.g., ~/Projects/Harvana)
python .ai-pack/scripts/migrate-to-beads.py
```

The script will:
1. ✅ Check prerequisites (git repo, Beads installed)
2. ✅ Update `.ai-pack/` submodule to latest
3. ✅ Install Beads if not present
4. ✅ Initialize Beads (`bd init`)
5. ✅ Update `.gitignore` for `.beads/*.db`
6. ✅ Verify configuration
7. ✅ Show next steps

**Interactive:** Script will prompt before making changes.

### Method 2: Manual Migration

If you prefer manual control:

#### Step 1: Update AI-Pack Submodule

```bash
cd .ai-pack
git fetch origin
git checkout main
git pull origin main
cd ..
git add .ai-pack
git commit -m "Update ai-pack to v1.1.0 (Beads integration)"
```

Verify version:
```bash
cd .ai-pack && git log --oneline -1 | grep -E "(beads|v1.1)" && cd ..
```

#### Step 2: Install Beads

Choose your platform:

**macOS (Homebrew):**
```bash
brew install steveyegge/beads/beads
```

**Linux/macOS (curl):**
```bash
curl -fsSL https://raw.githubusercontent.com/steveyegge/beads/main/install.sh | bash
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/steveyegge/beads/main/install.ps1 | iex
```

**FreeBSD:**
```bash
pkg install beads
```

Verify installation:
```bash
bd --version
# Should output: beads version X.Y.Z
```

#### Step 3: Initialize Beads in Your Project

```bash
# From project root
bd init

# Verify .beads/ directory created
ls -la .beads/
# Should see: issues.jsonl and potentially *.db files
```

#### Step 4: Update .gitignore

Add Beads SQLite databases to `.gitignore` (JSONL is committed):

```bash
# Check if already present
grep -q "^\.beads/\*\.db$" .gitignore || echo ".beads/*.db" >> .gitignore
```

**IMPORTANT:** The `.beads/issues.jsonl` file MUST be committed to git. Only `*.db` files should be ignored.

#### Step 5: Commit Beads Configuration

```bash
git add .beads/issues.jsonl .gitignore
git commit -m "Initialize Beads task tracking

- Add .beads/issues.jsonl (git-backed task database)
- Ignore .beads/*.db (local SQLite cache)
- Enables persistent task memory across AI sessions"
```

#### Step 6: Verify Configuration

```bash
# Test Beads is working
bd list
# Should output: (empty list or existing tasks)

# Verify git tracking
git ls-files .beads/
# Should output: .beads/issues.jsonl

# Verify .db is ignored
git status --porcelain .beads/*.db
# Should output: (nothing - files are ignored)
```

## Verifying the Migration

After migration (automated or manual), verify:

### 1. Beads is Installed
```bash
bd --version
# Expected: beads version X.Y.Z
```

### 2. .beads/ Directory Exists
```bash
ls -la .beads/
# Expected:
# - issues.jsonl (committed to git)
# - *.db files (git-ignored, may not exist until first use)
```

### 3. Git Configuration Correct
```bash
# issues.jsonl is tracked
git ls-files .beads/
# Expected: .beads/issues.jsonl

# *.db files are ignored
grep "\.beads/\*\.db" .gitignore
# Expected: .beads/*.db
```

### 4. AI-Pack is Updated
```bash
cd .ai-pack && git log --oneline -1 && cd ..
# Expected: Commit message mentioning Beads or v1.1.0
```

### 5. Beads Commands Work
```bash
# Create test task
bd create "Test task for verification"

# List tasks
bd list

# Check task is in JSONL
cat .beads/issues.jsonl | tail -1
# Expected: JSON line with your test task

# Close test task
bd close <task-id>
```

## Using Beads with AI Agents

Once migrated, AI agents (Orchestrator, Engineer) will automatically use Beads:

### Orchestrator Workflow
```bash
bd create "Implement user authentication"
bd create "Add password reset feature"
bd dep add <auth-id> <reset-id>  # reset depends on auth
bd list --status open             # Monitor progress
```

### Engineer Workflow
```bash
bd ready                    # Find next available task
bd show bd-a1b2            # Review task details
bd start bd-a1b2           # Begin work
# ... implement code ...
bd close bd-a1b2           # Mark complete
bd ready                   # Find next task
```

See `quality/tooling/beads-integration.md` for comprehensive usage guide.

## Cross-Session Continuity

**Before Migration:**
```
Session 1: AI creates TodoWrite tasks
[Session ends]
Session 2: AI has no memory of tasks ❌
```

**After Migration:**
```
Session 1: AI creates Beads tasks (bd create)
[Session ends]
Session 2: AI reads tasks from git (bd list) ✅
```

## Team Collaboration

Beads enables multi-developer coordination:

```bash
# Developer A creates tasks
bd create "Fix auth bug" "Add caching" "Update docs"
git add .beads/issues.jsonl
git commit -m "Add tasks"
git push

# Developer B syncs tasks
git pull  # Beads auto-imports from issues.jsonl
bd list   # See Developer A's tasks

# Developer B starts work
bd ready
bd start bd-a1b2
```

## Troubleshooting

### Problem: `bd: command not found`

**Solution:** Beads not installed. Run installation for your platform (see Step 2).

### Problem: `.beads/` directory doesn't exist after `bd init`

**Solution:** Not in git repository. Beads requires git.
```bash
git rev-parse --is-inside-work-tree
# If false, initialize git first: git init
```

### Problem: `bd list` shows error about database

**Solution:** Corrupted SQLite cache. Regenerate:
```bash
rm .beads/*.db
bd list  # Will regenerate from issues.jsonl
```

### Problem: Tasks not syncing across machines

**Solution:** Ensure `.beads/issues.jsonl` is committed and pushed:
```bash
git add .beads/issues.jsonl
git commit -m "Sync tasks"
git push
```

### Problem: Merge conflicts in `.beads/issues.jsonl`

**Solution:** Beads uses append-only JSONL. Safe to take both versions:
```bash
# Git will merge line-by-line automatically
git add .beads/issues.jsonl
git commit -m "Merge tasks"
```

### Problem: Migration script fails with Python error

**Solution:** Ensure Python 3.7+:
```bash
python3 --version
# Use python3 if python points to Python 2.x
python3 .ai-pack/scripts/migrate-to-beads.py
```

## Rolling Back (If Needed)

If you need to revert the migration:

```bash
# 1. Remove Beads directory
rm -rf .beads/

# 2. Revert .gitignore changes
git checkout HEAD -- .gitignore

# 3. Revert ai-pack to v1.0.0
cd .ai-pack
git checkout v1.0.0  # Or previous commit
cd ..
git add .ai-pack
git commit -m "Rollback ai-pack to v1.0.0"
```

**Note:** You'll lose persistent task tracking but TodoWrite (session-only) will work as before.

## Next Steps After Migration

1. **Read Beads Integration Guide**: `.ai-pack/quality/tooling/beads-integration.md`
2. **Update Team Documentation**: Inform team about new workflow
3. **Test with AI Agent**: Start new Claude session and verify agents use Beads
4. **Create Initial Tasks**: Use `bd create` to populate task backlog

## References

- [Beads GitHub Repository](https://github.com/steveyegge/beads)
- [Beads Documentation](https://github.com/steveyegge/beads/tree/main/docs)
- [AI-Pack Beads Integration Guide](quality/tooling/beads-integration.md)
- [AI-Pack v1.1.0 Release Notes](.ai/context.md)

## Support

Issues with migration?

1. Check troubleshooting section above
2. Review `.ai-pack/quality/tooling/beads-integration.md`
3. Check Beads GitHub issues: https://github.com/steveyegge/beads/issues
4. Check AI-Pack GitHub issues: https://github.com/Cortexa-LLC/ai-pack/issues

---

**Version:** 1.0.0
**Last Updated:** 2026-01-12
**AI-Pack Version:** v1.1.0+
