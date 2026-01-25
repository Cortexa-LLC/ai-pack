# Resetting the `.ai-pack` Git Submodule (Clean Removal + Re-Add)

**Effective Date**: 2026-01-24
**Breaking Change**: Infrastructure reorganization requires submodule reset

---

## Why This Is Needed

As of **2026-01-24**, the AI-Pack repository underwent an infrastructure reorganization (v2.0.0) that may cause Git submodule cache conflicts. If you see an error like:

> `fatal: A git directory for '.ai-pack' is found locally with remote(s): ... use '--force' ...`

This means Git has cached submodule repository data at `.git/modules/.ai-pack` that conflicts with the updated structure.

## Quick Fix (Automated)

Use the provided Python script to automatically reset the submodule:

```bash
# From your project root (the repo that contains .ai-pack as a submodule)
python3 .ai-pack/scripts/reset-submodule.py
```

The script will:
1. Safely remove the existing submodule
2. Clean all Git cache and configuration
3. Re-add the submodule with the latest structure
4. Initialize and update the submodule

---

## Manual Reset Procedure

If you prefer to reset manually, follow these steps:

### Preconditions

* Run all commands from the **root** of your parent repository (the repo that contains the submodule)
* Replace `.ai-pack` and the URL as needed for other submodules

### 1) Identify Current State (Optional)

```bash
git submodule status || true
git config -f .gitmodules --get-regexp '^submodule\.' || true
ls -la .git/modules 2>/dev/null || true
```

### 2) Fully Remove the Submodule

#### A. Deinitialize the submodule

```bash
git submodule deinit -f .ai-pack || true
```

#### B. Remove from index and working tree

```bash
git rm -f .ai-pack || true
```

#### C. Remove cached submodule repository (critical step)

This prevents the "git directory is found locally" error:

```bash
rm -rf .git/modules/.ai-pack
```

#### D. Remove configuration entries

```bash
git config -f .gitmodules --remove-section submodule..ai-pack 2>/dev/null || true
git config --remove-section submodule..ai-pack 2>/dev/null || true
```

#### E. Remove working directory

```bash
rm -rf .ai-pack
```

### 3) Commit the Removal

```bash
git add .gitmodules 2>/dev/null || true
git commit -m "Remove .ai-pack submodule for v2.0.0 reset"
```

**Note**: If there are no changes to commit, Git will tell you (this is expected).

### 4) Re-Add the Submodule

```bash
git submodule add https://github.com/Cortexa-LLC/ai-pack.git .ai-pack
git submodule update --init --recursive
git commit -m "Re-add .ai-pack submodule (v2.0.0 structure)"
```

### 5) Verify

```bash
git submodule status
cd .ai-pack && git log --oneline -5
```

You should see the latest commits including Phase 2 infrastructure.

---

## One-Shot Script (Drop-in)

```bash
SUBMODULE_PATH=".ai-pack"
SUBMODULE_URL="https://github.com/Cortexa-LLC/ai-pack.git"

# Full cleanup
git submodule deinit -f "$SUBMODULE_PATH" || true
git rm -f "$SUBMODULE_PATH" || true
rm -rf ".git/modules/$SUBMODULE_PATH"
git config -f .gitmodules --remove-section "submodule.$SUBMODULE_PATH" 2>/dev/null || true
git config --remove-section "submodule.$SUBMODULE_PATH" 2>/dev/null || true
rm -rf "$SUBMODULE_PATH"

# Commit removal
git add .gitmodules 2>/dev/null || true
git commit -m "Reset submodule $SUBMODULE_PATH (v2.0.0)" || true

# Re-add
git submodule add "$SUBMODULE_URL" "$SUBMODULE_PATH"
git submodule update --init --recursive
git commit -m "Re-add submodule $SUBMODULE_PATH (v2.0.0)"
```

---

## For New Clones

If someone clones your repo fresh after you've completed the reset:

```bash
git clone <YOUR_REPO_URL>
cd <YOUR_REPO_DIR>
git submodule update --init --recursive
```

If someone already cloned but is missing submodules:

```bash
git submodule update --init --recursive
```

---

## Common Pitfalls & Diagnostics

### Error: "A git directory for '.ai-pack' is found locally…"

**Cause**: Leftover cache at `.git/modules/.ai-pack`

**Fix**:
```bash
rm -rf .git/modules/.ai-pack
```

### Submodule Path Changed

If you changed the path (e.g., `.ai` vs `.ai-pack`), remove the old cached directory:

```bash
rm -rf .git/modules/<old-path>
```

### `.gitmodules` Still Contains Old Entries

Check:
```bash
cat .gitmodules
```

Remove the matching section manually and commit.

### Permission Denied Errors

If you get permission errors during removal:

```bash
chmod -R u+w .ai-pack
rm -rf .ai-pack
```

---

## SSH vs HTTPS

The automated script uses **HTTPS** by default:
```
https://github.com/Cortexa-LLC/ai-pack.git
```

If you prefer **SSH**:
```bash
git submodule add git@github.com:Cortexa-LLC/ai-pack.git .ai-pack
```

Or edit `.gitmodules` after adding:
```ini
[submodule ".ai-pack"]
    path = .ai-pack
    url = git@github.com:Cortexa-LLC/ai-pack.git
```

Then:
```bash
git submodule sync
git submodule update --init --recursive
```

---

## Troubleshooting

### Script Fails with "Not a Git repository"

Make sure you run the script from your project root (the parent repo), not from inside `.ai-pack`.

### Changes Not Reflected After Reset

1. Verify you're on the latest commit:
   ```bash
   cd .ai-pack
   git log --oneline -1
   ```

2. Force update if needed:
   ```bash
   cd .ai-pack
   git fetch origin main
   git reset --hard origin/main
   cd ..
   git add .ai-pack
   git commit -m "Update .ai-pack to latest"
   ```

### Submodule Shows as Modified

If `git status` shows `.ai-pack` as modified:

```bash
cd .ai-pack
git checkout main
git pull origin main
cd ..
git add .ai-pack
git commit -m "Update .ai-pack submodule reference"
```

---

## Support

For issues with the reset procedure:
- [GitHub Issues](https://github.com/Cortexa-LLC/ai-pack/issues)
- [GitHub Discussions](https://github.com/Cortexa-LLC/ai-pack/discussions)

---

**Last Updated**: 2026-01-24
**Version**: 2.0.0
