# Breaking Changes

This document tracks breaking changes in AI-Pack that require action from consumers.

---

## 2026-01-24: Infrastructure Reorganization (v2.0.0)

**Version**: 2.0.0
**Severity**: High - Requires Submodule Reset
**Affects**: All projects using ai-pack as a Git submodule

### What Changed

Version 2.0.0 introduced a Go-based A2A agent server infrastructure that reorganized the repository structure. This change may cause Git submodule cache conflicts for existing projects.

### Symptoms

You may encounter this error when updating your submodule:

```
fatal: A git directory for '.ai-pack' is found locally with remote(s):
  origin https://github.com/Cortexa-LLC/ai-pack.git
If you want to reuse this local git directory instead of cloning again from
  https://github.com/Cortexa-LLC/ai-pack.git
use the '--force' option. If the local git directory is not the correct repo
or you are unsure what this means choose another name with the '--name' option.
```

### Required Action

**Option 1: Automated Reset (Recommended)**

Use the provided Python script:

```bash
# From your project root (the repo that contains .ai-pack)
python3 .ai-pack/scripts/reset-submodule.py
```

**Option 2: Manual Reset**

Follow the manual procedure in [SUBMODULE-RESET.md](SUBMODULE-RESET.md).

### What Gets Reset

The reset procedure:
1. Removes the existing `.ai-pack` submodule completely
2. Cleans all Git cache at `.git/modules/.ai-pack`
3. Removes all configuration entries
4. Re-adds the submodule with the updated structure
5. Commits both the removal and re-addition

### Impact

- **Low** - Two commits in your project history (removal + re-addition)
- **No data loss** - All your project code remains unchanged
- **Time required** - 1-2 minutes with automated script

### Documentation

- **Reset Guide**: [docs/SUBMODULE-RESET.md](SUBMODULE-RESET.md)
- **Automated Script**: [scripts/reset-submodule.py](../scripts/reset-submodule.py)
- **A2A Usage Guide**: [docs/content/framework/a2a-usage-guide.md](content/framework/a2a-usage-guide.md)

---

## Future Breaking Changes

Breaking changes will be documented here with:
- **Date** - When the change was introduced
- **Version** - Semantic version number
- **Severity** - Impact level (Low/Medium/High)
- **Required Action** - What you need to do
- **Migration Guide** - How to upgrade

---

**Last Updated**: 2026-01-24
