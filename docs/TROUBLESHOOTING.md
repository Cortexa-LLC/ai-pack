# AI-Pack Troubleshooting Guide

## Hook Execution Errors

### Error: "can't open file '.ai-pack/.claude/hooks/...'"

**Symptom:**
```
UserPromptSubmit operation blocked by hook:
  [python3 .claude/hooks/check-task-packet.py]:
  can't open file '/path/to/project/.ai-pack/.claude/hooks/check-task-packet.py': [Errno 2] No such file or directory
```

**Root Cause:**
Older versions of ai-pack templates used relative paths without ensuring the correct working directory. When Claude Code executes hooks, the current working directory may not be the project root, causing Python to resolve paths incorrectly.

**Solution:**
Update your `.claude/settings.json` to use the `cd` command before executing Python hooks:

**OLD (broken):**
```json
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "command": "python3 .claude/hooks/check-task-packet.py"
          }
        ]
      }
    ]
  }
}
```

**NEW (correct):**
```json
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "command": "cd $(git rev-parse --show-toplevel 2>/dev/null || pwd) && python3 .claude/hooks/check-task-packet.py"
          }
        ]
      }
    ]
  }
}
```

**Quick Fix:**
Replace all hook commands in your `.claude/settings.json`:

```bash
# Backup first
cp .claude/settings.json .claude/settings.json.backup

# Update to use cd before python3
sed -i.bak 's|"python3 \.claude/hooks/|"cd $(git rev-parse --show-toplevel 2>/dev/null \|\| pwd) \&\& python3 .claude/hooks/|g' .claude/settings.json
```

**Alternative: Run Update Script:**
```bash
python3 .ai-pack/templates/.claude-update.py
```

This will copy the latest template with correct hook paths.

---

## Updating from Old ai-pack Versions

If you installed ai-pack before 2026-04-22, you may have the old hook path format. Run the update script:

```bash
# From your project root
python3 .ai-pack/templates/.claude-update.py
```

This will:
- ✅ Update all Claude Code integration files
- ✅ Fix hook paths in settings.json
- ✅ Preserve your custom commands/skills/rules
- ✅ Create a backup of your existing .claude/ directory

---

## Why This Happened

The issue occurred because:

1. **Old Template Format (before 2026-04-22):**
   - Used relative paths: `python3 .claude/hooks/script.py`
   - Assumed CWD would be project root
   - Broke when Claude Code executed hooks from different directories

2. **New Template Format (after 2026-04-22):**
   - Explicitly sets CWD: `cd $(git rev-parse --show-toplevel 2>/dev/null || pwd) && python3 .claude/hooks/script.py`
   - Works regardless of Claude Code's execution context
   - Falls back to `pwd` if not in a git repo

---

## Verification

After applying the fix, verify it works:

```bash
# Test the hook directly
echo '{"user_input": "test implement"}' | cd $(git rev-parse --show-toplevel) && python3 .claude/hooks/check-task-packet.py

# Exit code should be:
# - 0 if task packets exist
# - 2 if no task packets (gate violation)
# - 1 if technical error
```

---

## Related Issues

- [GitHub Issue #XX] - Hook path resolution on non-root CWD
- [PR #XX] - Fix hook paths in template

