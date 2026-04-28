# AI-Pack Migration Guide

This guide helps you migrate existing projects through ai-pack version updates.

---

## Migrating to v2.2.0 (Task Database SQLite Migration)

**Release Date:** 2026-04-28
**Codename:** Task Database Migration

### What Changed in v2.2.0

**BREAKING:** Complete replacement of JSON-based task tracking with SQLite database.

### Breaking Changes

- ❌ **REMOVED:** `.beads/state.json` task tracking
- ❌ **REMOVED:** JSON-based task database
- ✅ **NEW:** SQLite database (`~/.ai-pack/tasks.db`) for centralized task tracking
- ⚠️ **REQUIRED:** One-time migration of existing task data

### Why This Change?

Improves performance, reliability, and data integrity:
- **Before v2.2:** JSON files (slow reads, no indexing, prone to corruption)
- **After v2.2:** SQLite database (fast queries, ACID compliance, structured data)

**Key Benefits:**
- ⚡ **10-100x faster** task queries
- 🔒 **ACID transactions** prevent data corruption
- 📊 **Indexed searches** for instant lookups
- 🔍 **Structured queries** with filters and joins
- 🧹 **Automatic archival** of old tasks
- 💾 **Centralized storage** at `~/.ai-pack/tasks.db`

### Migration Steps

**Prerequisites:**
- AI-Pack v2.0.0 or later
- Existing projects with tasks in `.beads/tasks/` directories

**Step 1: Update ai-pack repository**

```bash
cd /path/to/ai-pack
git pull origin main
```

**Step 2: Rebuild binaries with SQLite support**

```bash
make build
```

**Step 3: Run migration (import all tasks to SQLite)**

```bash
# Migrate all projects automatically
./bin/agent migrate --all

# OR migrate a specific project
./bin/agent migrate --project /path/to/project
```

**Step 4: Restart services**

```bash
make restart-all
```

**Step 5: Verify migration**

```bash
# Check task database location
ls -lh ~/.ai-pack/tasks.db

# List all tasks in database
agent status --all

# Verify specific project tasks
cd /path/to/project
agent status
```

### What Happens During Migration

The migration command:
1. ✅ Scans `.beads/tasks/*/metadata.json` files
2. ✅ Imports task data into SQLite database
3. ✅ Preserves task status, timestamps, and metadata
4. ✅ Skips already-migrated tasks (idempotent)
5. ✅ Leaves original files intact (non-destructive)

**Migration output example:**
```
Found project: /Users/brywoodruff/Projects/MyProject
  ✅ Migrated 34 tasks

📊 Migration complete:
   Projects scanned: 1
   Tasks migrated:   34
```

### Post-Migration Behavior

**Task tracking (improved):**
- All task operations use SQLite database
- Faster queries with indexed lookups
- Automatic archival after 7 days (configurable)
- Centralized storage across all projects

**Backward compatibility:**
- Old `.beads/tasks/` directories remain for execution logs
- Metadata files still created for individual task state
- No changes to task packet structure

**Agent workflow (unchanged):**
- `agent engineer <task-id> --stream` works as before
- `agent status <task-id>` queries SQLite database
- `agent logs <task-id>` reads from execution directory

### Configuration

Task database settings in `~/.ai-pack/agent-server.json`:

```json
{
  "task_cleanup": {
    "enabled": true,
    "archive_after_days": 7
  }
}
```

**archive_after_days:** Move completed task execution folders to `.beads/archive/YYYY-MM/`

### Troubleshooting

**Problem: "Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo"**

**Solution:** Rebuild with CGO enabled (already fixed in Makefile):
```bash
make clean
make build
```

**Problem: Migration reports 0 tasks**

**Cause:** No `.beads/tasks/` directories found or tasks already migrated.

**Solution:** Verify task directories exist:
```bash
find ~/Projects -name ".beads" -type d -exec ls -ld {} \;
```

**Problem: Migration fails with database error**

**Solution:** Reset database and re-migrate:
```bash
rm ~/.ai-pack/tasks.db
./bin/agent migrate --all
```

**Problem: Tasks missing after migration**

**Solution:** Check migration was run for correct project:
```bash
# List all tasks in database
sqlite3 ~/.ai-pack/tasks.db "SELECT id, beads_id, status FROM tasks;"
```

### Rolling Back (Not Recommended)

If you must revert to pre-v2.1 behavior:

```bash
# 1. Stop services
make stop-all

# 2. Checkout previous version
git checkout <commit-before-v2.1>

# 3. Rebuild
make build

# 4. Restart
make start-all
```

**Note:** Task data remains in SQLite database. Future updates will use it when you re-migrate.

### What's New in v2.2.0

- ✅ **SQLite Task Database**: Centralized task storage with 10-100x faster queries
- ✅ **Indexed Lookups**: Instant task searches by ID, status, project, role
- ✅ **ACID Transactions**: Prevent data corruption with atomic operations
- ✅ **Automatic Archival**: Old completed tasks moved to `.beads/archive/YYYY-MM/`
- ✅ **Migration Command**: `agent migrate --all` for one-time import
- ✅ **Orphan Detection**: Two-pass system (taskDB + metadata files)
- ✅ **Comprehensive Tests**: 28+ new tests for agent CLI commands
- ✅ **Agent CLI Commands**: `agent create`, `agent show`, `agent update`, `agent close`
- ❌ **REMOVED**: Beads CLI dependency (`bd` commands replaced with `agent` commands)

### Template and Documentation Updates for Projects Using ai-pack as Submodule

**IMPORTANT:** If your project uses ai-pack as a git submodule, you MUST update your project's templates and documentation after pulling the v2.2.0 update.

#### Changes Required in Your Project

The following terminology and commands have changed:

| **Old (v2.1.0 and earlier)** | **New (v2.2.0+)** |
|------------------------------|-------------------|
| `agent create "Task"` | `agent create "Task" --priority P1` |
| `agent show <task-id>` | `agent show <task-id>` |
| `agent close <task-id>` | `agent close <task-id>` |
| `agent update <task-id>` | `agent update <task-id> --status <status>` |
| `agent list` | `agent list --all` |
| `agent list --status queued` | `agent list --status queued` |
| "Task ID" terminology | "Task ID" terminology |
| "task" | "Task" |

#### Step-by-Step Update Process

**1. Update the ai-pack submodule:**

```bash
cd /path/to/your-project
git submodule update --remote .ai-pack
git add .ai-pack
git commit -m "Update ai-pack to v2.2.0 - SQLite task database"
```

**2. Copy updated templates to your project:**

```bash
# Backup your current .claude directory
cp -r .claude .claude.backup

# Copy updated templates from ai-pack
cp .ai-pack/templates/CLAUDE.md .claude/CLAUDE.md

# Update rules (if you use them)
cp .ai-pack/templates/.claude/rules/*.md .claude/rules/

# Update skills (if you use them)
cp .ai-pack/templates/.claude/skills/orchestrator/SKILL.md .claude/skills/orchestrator/
cp .ai-pack/templates/.claude/skills/coordinator/SKILL.md .claude/skills/coordinator/
```

**3. Update project-specific documentation:**

Search and replace across your project documentation:

```bash
# Find all references to bd commands
grep -r "agent create\|agent show\|agent close\|agent update\|agent list\|agent list --status queued" . --include="*.md"

# Find all "Task ID" references
grep -r "Task ID\|task\|task-id" . --include="*.md"
```

**4. Update your project's quick start / workflow documentation:**

**OLD workflow (v2.1.0):**
```bash
# Create task
BID=$(agent create "Implement feature" --priority high --json | jq -r '.id')

# Create task packet
TS=$(date +%Y%m%d%H%M%S)
mkdir -p .ai/tasks/${BID}-${TS}-feature/

# Spawn agent
agent engineer ${BID} --stream

# Complete task
agent close ${BID}
```

**NEW workflow (v2.2.0):**
```bash
# Create task
TASK_ID=$(agent create "Implement feature" --priority P1 --role engineer --json | jq -r '.task_id')

# Create task packet
TS=$(date +%Y%m%d%H%M%S)
mkdir -p .ai/tasks/${TASK_ID}-${TS}-feature/

# Spawn agent
agent engineer ${TASK_ID} --stream

# Complete task
agent close ${TASK_ID} -r "Completed"
```

**5. Update .claudeignore patterns (if applicable):**

```bash
# OLD pattern
.beads/tasks/

# NEW pattern (still valid - execution directories remain)
.beads/tasks/
```

**6. Review and update any custom scripts:**

If your project has custom scripts that use `bd` commands, update them to use `agent` commands:

```bash
# Example: Update monitoring script
# OLD
agent list --status in_progress

# NEW
agent list --running
```

**7. Update team documentation:**

Ensure your team knows about the command changes:

- Update README.md
- Update CONTRIBUTING.md
- Update any onboarding documentation
- Update CI/CD scripts if they reference `bd` commands

#### Files to Review in Your Project

Common files that may need updates:

```
Your Project Root/
├── .claude/
│   ├── CLAUDE.md          ← MUST update (task commands changed)
│   ├── rules/*.md         ← Update if using ai-pack rules
│   └── skills/*/SKILL.md  ← Update if using ai-pack skills
├── README.md              ← Update quick start examples
├── CONTRIBUTING.md        ← Update workflow documentation
├── docs/
│   ├── workflows/*.md     ← Update any workflow guides
│   └── onboarding/*.md    ← Update setup instructions
└── scripts/
    └── *.sh               ← Update any scripts using bd commands
```

#### Verification

After updating, verify the changes:

```bash
# 1. Check no bd commands remain in documentation
grep -r "agent create\|agent show\|agent close" . --include="*.md" | grep -v "backup"

# 2. Check no "Task ID" references remain
grep -r "Task ID" . --include="*.md" | grep -v "backup"

# 3. Test the new workflow
agent create "Test task" --priority P1 --role engineer --json
agent list --all
```

#### Example: Before and After

**Before (CLAUDE.md with Beads):**
```markdown
## Quick Start

1. Create task:
   ```bash
   BID=$(agent create "Fix bug" --priority high --json | jq -r '.id')
   ```

2. Show task details:
   ```bash
   agent show ${BID}
   ```

3. Complete task:
   ```bash
   agent close ${BID}
   ```
```

**After (CLAUDE.md with Agent CLI):**
```markdown
## Quick Start

1. Create task:
   ```bash
   TASK_ID=$(agent create "Fix bug" --priority P0 --role engineer --json | jq -r '.task_id')
   ```

2. Show task details:
   ```bash
   agent show ${TASK_ID}
   ```

3. Complete task:
   ```bash
   agent close ${TASK_ID} -r "Bug fixed"
   ```
```

#### Common Migration Issues

**Issue: Scripts still reference `bd` command**

```bash
# Find all shell scripts using bd
grep -r "bd " . --include="*.sh"

# Update each occurrence
# OLD: agent create "Task"
# NEW: agent create "Task" --priority P1
```

**Issue: Documentation references Beads terminology**

```bash
# Find and replace in all markdown files
find . -name "*.md" -type f -exec sed -i '' 's/Task ID/Task ID/g' {} +
find . -name "*.md" -type f -exec sed -i '' 's/task/task/g' {} +
```

**Issue: Task packet slugs still reference "task-id"**

The task packet directory format remains the same:
```
.ai/tasks/<task-id>-<YYYYMMDDHHMMSS>-<short-desc>/
```

No changes needed for existing task packet directories.

---

## Migrating to v2.0.0 (BREAKING CHANGES)

**Release Date:** 2026-01-14
**Codename:** Consolidation

### What Changed in v2.0.0

**BREAKING:** Agent tracking consolidation removes deprecated `agent-status-tracker.py` system.

### Breaking Changes

- ❌ **REMOVED:** `templates/.claude/scripts/agent-status-tracker.py`
- ❌ **REMOVED:** `.claude/.agent-status.json` file format
- ⚠️ **REQUIRED:** All agent tracking now uses task database exclusively

### Migration Steps

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

**Step 2: Clean up old files**

```bash
# Remove old status file (if exists)
rm -f .claude/.agent-status.json
rm -f .claude/.agent-status.json.backup
```

**Step 3: Verify migration**

```bash
# Check agent tasks
agent status --all
```

---

## Next Steps After Migration

1. **Verify Task Database**: Check `~/.ai-pack/tasks.db` exists
2. **Test Agent Commands**: Run `agent status` to query database
3. **Configure Archival**: Adjust `archive_after_days` in config if needed
4. **Monitor Performance**: Task queries should be noticeably faster

## References

- [SQLite Task Database Schema](internal/taskdb/schema.go)
- [Agent Server Configuration](docs/USAGE-GUIDE.md)
- [Task Migration Implementation](cmd/agent/commands/migrate.go)

## Support

Issues with migration?

1. Check troubleshooting section above
2. Verify CGO is enabled in build (`make build` output)
3. Check database file exists: `ls -lh ~/.ai-pack/tasks.db`
4. Review migration logs: `./bin/agent migrate --all`
5. Open issue: https://github.com/Cortexa-LLC/ai-pack/issues

---

**Version:** 2.2.0
**Last Updated:** 2026-04-28
**AI-Pack Version:** v2.2.0+
