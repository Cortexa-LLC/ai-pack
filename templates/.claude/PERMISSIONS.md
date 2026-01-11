# AI-Pack Permission Model

This document explains the permission strategy configured in [settings.json](./settings.json).

## Philosophy

Enable smooth workflow for both foreground (interactive) and background (agent) execution while requiring approval for critical/destructive operations.

## Auto-Approved Operations

These operations are pre-approved and won't prompt for confirmation:

### File Operations
- ✅ **Write(*)** - Create/overwrite any file
- ✅ **Edit(*)** - Modify any file

**Rationale:** Core development workflow. Constantly approving edits creates friction. Files are version-controlled (git), so changes are recoverable.

### Safe Bash Commands
- ✅ **mkdir, cp, mv** - File system operations
- ✅ **ls, cat, pwd, echo, which, cd, test** - Read-only utilities and navigation
- ✅ **npm, dotnet, jest, pytest, tsc** - Build/test/package managers
- ✅ **cargo, go, gradle, mvn** - Additional language tooling
- ✅ **node, python, python3** - Script execution

**Rationale:** Standard development commands needed for implementation and testing. Non-destructive or easily recoverable.

**Note on compound commands:** Commands can be chained with `&&`, `||`, `;`, or pipes (`|`). The permission check applies to each individual command in the chain. For example:
- `cd path && git log` - Both `cd` and `git log` must be in allow list
- `git status && git diff` - Both commands are auto-approved
- `cd path && git commit` - `cd` auto-approved, but `git commit` requires approval

### Safe Git Operations
- ✅ **git pull** - Update from remote (fast-forward)
- ✅ **git fetch** - Download remote changes
- ✅ **git status** - Check working tree status
- ✅ **git diff** - View changes
- ✅ **git log** - View commit history
- ✅ **git branch** - List/view branches
- ✅ **git checkout** - Switch branches/restore files

**Rationale:** Read-only or sync operations that don't modify git history. Background agents can check repository state and sync without approval. Pull operations that require merge will fail safely and can be handled manually.

## Require Approval

These operations require explicit user approval:

### Destructive Git Operations
- ⚠️ **git add** - Stage changes for commit
- ⚠️ **git commit** - Create commits
- ⚠️ **git push** - Push to remote
- ⚠️ **git merge** - Merge branches
- ⚠️ **git rebase** - Rewrite history
- ⚠️ **git reset** - Reset working tree

**Rationale:** These operations modify git history or remote state. User should review all changes and explicitly approve commits/pushes to maintain control over repository state.

### Destructive Operations
- ⚠️ **rm, rmdir** - File deletion
- ⚠️ **git reset --hard** - Destructive git operations
- ⚠️ **docker rm, docker rmi** - Container/image removal

**Rationale:** Permanent or difficult to recover operations. User should explicitly approve deletions.

### System Operations
- ⚠️ **sudo** - Elevated privileges
- ⚠️ **chmod, chown** - Permission changes
- ⚠️ **killall, pkill** - Process termination

**Rationale:** System-level changes that affect security or stability.

## Default Mode

```json
"defaultMode": "acceptEdits"
```

This setting auto-approves Write and Edit operations for smoother workflow.

## Benefits

### For Foreground (Interactive) Work
- No constant approval prompts for file edits
- Natural development flow
- Still protected from destructive operations

### For Background (Agent) Work
- Agents can write/edit files autonomously
- Tests and builds run without blocking
- Parallel execution works smoothly
- **Same restrictions**: Background agents also blocked from git/rm/destructive operations

## How It Works

The permissions in `settings.json` apply **globally** to:
- Main session (foreground/interactive work)
- Background agents spawned via Task tool
- All subprocesses and agent contexts

**This means:**
- ✅ Background agents CAN write/edit files (smooth parallel execution)
- ✅ Background agents CAN use safe git operations (pull, fetch, status, diff, log, branch, checkout)
- ⚠️ Background agents CANNOT commit/push/merge (must request approval)
- ⚠️ Background agents CANNOT use rm/destructive operations
- ⚠️ If a background agent needs restricted operations, it will **fail and report the issue**

**Implication for orchestration:**
- Background agents should focus on implementation (write code, run tests)
- Background agents can check git status and sync with remote (pull/fetch)
- Git operations that modify history (commit, push, merge) should be done in main session after review
- Destructive operations should be done manually by user

## Customization

Projects can extend permissions by adding to the `allow` array:

```json
"allow": [
  "Write(*)",
  "Edit(*)",
  "Bash(your-custom-tool:*)"
]
```

## Security Considerations

1. **Repository Safety**: All file changes are version-controlled
2. **Destructive Operations**: Still require approval (git, rm, etc.)
3. **Rollback**: User can always revert commits or restore from backup
4. **Transparency**: All operations logged in conversation history

## Behavior in Background Agents

When a background agent tries to use a restricted operation:

1. **Agent cannot prompt for approval** (background = non-interactive)
2. **Operation fails immediately**
3. **Agent should detect failure and report to user**
4. **Work log documents the blocker**

**Example scenario:**
```
Background Engineer agent implements feature
  → Tries to commit: git commit -m "Add feature"
  → Permission denied (git not in allow list)
  → Agent reports: "Cannot commit - git requires approval"
  → User reviews code in main session
  → User manually commits after review
```

**This is by design:**
- Background agents focus on implementation
- User maintains control over git history
- All commits require human review

## Recommended Workflow Pattern

### Sequential Review (Option 1)

The recommended workflow separates implementation from git operations:

**Phase 1: Parallel Implementation (Background)**
```python
# Spawn multiple background agents for parallel work
Task(subagent_type="general-purpose",
     description="Implement feature A",
     prompt="Implement feature A...",
     run_in_background=true)

Task(subagent_type="general-purpose",
     description="Implement feature B",
     prompt="Implement feature B...",
     run_in_background=true)
```

**Phase 2: Wait and Monitor**
- Background agents work autonomously
- Agents can run git status, git diff, git pull to check repository state
- Check work logs: `.ai/tasks/*/20-work-log.md`
- Wait for completion or blockers

**Phase 3: Review and Commit (Foreground/Interactive)**
```bash
# Review changes
git diff

# Review agent work logs
cat .ai/tasks/*/20-work-log.md

# Commit with user approval
git add .
git commit -m "Implement features A and B"
git push
```

**Why this works:**
- ✅ Agents write code autonomously (no approval prompts)
- ✅ Agents can sync with remote (git pull, git fetch) without blocking
- ✅ Agents can check status (git status, git diff) to understand repository state
- ✅ User reviews all changes before committing
- ✅ Git history stays clean and intentional
- ✅ User maintains full control over commits and pushes

**Alternative: Foreground Git Agent**

If you want agent assistance with git operations:

```python
# After background implementation completes
Task(subagent_type="general-purpose",
     description="Review and commit changes",
     prompt="Review implementation and create commit...",
     run_in_background=false)  # Interactive, can prompt for git approval
```

## Examples

### Auto-Approved ✅
```bash
# Development workflow - no prompts
npm test
dotnet build
python script.py
mkdir src/components
cat README.md

# Safe git operations - no prompts
git status
git diff
git pull
git fetch
git log
git branch
git checkout feature-branch

# Compound commands with safe operations - no prompts
cd server && npm test
cd server/API && dotnet build
git fetch && git status
ls -la && cat README.md
```

**Works in both foreground and background.**

### Requires Approval ⚠️
```bash
# Destructive git operations - user confirms
git add .
git commit -m "message"
git push origin main
git merge feature-branch
git rebase main

# Destructive file operations - user confirms
rm -rf node_modules
sudo npm install -g package
```

**Foreground:** Prompts user for approval
**Background:** Fails immediately, agent reports blocker

## Related

- [settings.json](./settings.json) - Permission configuration
- [Claude Code Permissions Documentation](https://docs.anthropic.com/claude/docs/claude-code-permissions)
