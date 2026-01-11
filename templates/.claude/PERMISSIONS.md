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
- ✅ **ls, cat, pwd, echo, which** - Read-only utilities
- ✅ **npm, dotnet, jest, pytest, tsc** - Build/test/package managers
- ✅ **cargo, go, gradle, mvn** - Additional language tooling
- ✅ **node, python, python3** - Script execution

**Rationale:** Standard development commands needed for implementation and testing. Non-destructive or easily recoverable.

## Require Approval

These operations require explicit user approval:

### Git Operations
- ⚠️ **git add, commit, push, pull, merge, rebase, etc.**

**Rationale:** Git operations affect repository state and history. User should review commits and approve pushes.

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

## Examples

### Auto-Approved ✅
```bash
# Development workflow - no prompts
npm test
dotnet build
python script.py
mkdir src/components
cat README.md
```

### Requires Approval ⚠️
```bash
# Critical operations - user confirms
git commit -m "message"
git push origin main
rm -rf node_modules
sudo npm install -g package
```

## Related

- [settings.json](./settings.json) - Permission configuration
- [Claude Code Permissions Documentation](https://docs.anthropic.com/claude/docs/claude-code-permissions)
