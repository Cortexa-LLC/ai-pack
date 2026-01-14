# Role Extension Guide

**Version:** 1.0.0
**Last Updated:** 2026-01-14

## Overview

The ai-pack framework provides base roles that are **immutable and shared** across all projects. When you need project-specific behavior, you create **role extensions** in your project's `.ai/` directory.

---

## Critical Principles

### 🔒 Immutability Rule

```
❌ NEVER edit files in .ai-pack/
   - It's a git submodule
   - Managed externally
   - Changes will be lost on submodule update
   - Breaks other projects using ai-pack

✅ DO create extensions in .ai/
   - Project-specific additions
   - Safe from submodule updates
   - Local to your project only
```

### 📁 Directory Responsibilities

```
.ai-pack/ (IMMUTABLE - READ ONLY):
  - ❌ Never edit files here
  - ❌ Never add new files here
  - ✅ Only read and reference
  - ✅ Update via: git submodule update --remote .ai-pack

.ai/ (PROJECT-SPECIFIC - MUTABLE):
  - ✅ Task packets (.ai/tasks/)
  - ✅ Role extensions (.ai/roles/)
  - ✅ Custom workflows (if needed)
  - ✅ Project overrides (.ai/repo-overrides.md)

.claude/ (CLAUDE CODE INTEGRATION):
  - ✅ General project rules (.claude/rules/)
  - ✅ Claude-specific commands (.claude/commands/)
  - ❌ NOT for role extensions (use .ai/roles/ instead)
```

---

## When to Create a Role Extension

Create a role extension when:
- ✅ Need project-specific responsibilities
- ✅ Need domain-specific steps
- ✅ Need project-specific tools or scripts
- ✅ Need additional quality checks
- ✅ Need project-specific constraints

**Do NOT create an extension when:**
- ❌ The addition is universally applicable (contribute to ai-pack instead)
- ❌ It's a one-time task instruction (put in task packet)
- ❌ It's a general rule (put in .claude/rules/)

---

## Extension Pattern (Step-by-Step)

### Step 1: Create Extension File

```bash
# Create extension directory if needed
mkdir -p .ai/roles/

# Create extension file
touch .ai/roles/<role-name>-extension.md
```

**Example:** `.ai/roles/tester-extension.md`

### Step 2: Write Extension Content

**Template:**

```markdown
# <Role Name> Extension - [Project Name]

**Base Role:** `.ai-pack/roles/<role-name>.md` (immutable, managed by ai-pack)
**Extension Type:** Project-specific additions
**Last Updated:** [Date]

---

## Extension Purpose

[Brief explanation of why this extension exists]

---

## Additional Responsibilities

### [New Responsibility Name]

**When:** [When this applies]

**Steps:**
1. [Step 1]
2. [Step 2]
3. [Step 3]

**Example:**
```[language]
[Code example if applicable]
```

---

## Project-Specific Tools

### [Tool/Script Name]

**Location:** `[path/to/tool]`
**Purpose:** [What it does]
**Usage:**
```bash
[command]
```

---

## Additional Quality Gates

### [Gate Name]

**Requirement:** [What must be checked]
**Command:**
```bash
[verification command]
```

**Expected Result:** [Pass criteria]

---

## Integration with Base Role

This extension **augments** the base role defined in `.ai-pack/roles/<role-name>.md`.

**Base role responsibilities are unchanged.** This extension adds:
- [Addition 1]
- [Addition 2]

**Workflow:**
1. Follow base role procedure (`.ai-pack/roles/<role-name>.md`)
2. **Then** apply extension-specific steps (this document)
3. Report completion to Orchestrator

---

## References

- **Base Role:** [.ai-pack/roles/<role-name>.md](../.ai-pack/roles/<role-name>.md)
- **Project Overrides:** [.ai/repo-overrides.md](repo-overrides.md)
```

### Step 3: Document in repo-overrides.md

```bash
# Edit .ai/repo-overrides.md
```

**Add to "Role Extensions" section:**

```markdown
## Role Extensions

### <Role Name> Extension

**Extension Location:** `.ai/roles/<role-name>-extension.md`
**Base Role:** `.ai-pack/roles/<role-name>.md`
**Extension Summary:** [Brief description of what this extension adds]

**Key Additions:**
- [Addition 1]
- [Addition 2]
- [Addition 3]

**When to Use:** [Scenarios where this extension applies]
```

### Step 4: Reference in CLAUDE.md (Optional)

If the extension is commonly used, add a reference in `CLAUDE.md`:

```markdown
## Role Extensions

This project extends the following ai-pack roles:

### <Role Name>
**Base:** `.ai-pack/roles/<role-name>.md`
**Extension:** `.ai/roles/<role-name>-extension.md`
**Purpose:** [Brief description]

See: [.ai/roles/<role-name>-extension.md](.ai/roles/<role-name>-extension.md)
```

### Step 5: Commit Extension

```bash
# Stage extension and documentation
git add .ai/roles/<role-name>-extension.md
git add .ai/repo-overrides.md
git add CLAUDE.md  # If updated

# Commit
git commit -m "Add <role-name> extension for [project-specific need]"
```

---

## Real-World Example: Tester Extension

### Problem

Harvana project uses Docker containers for integration tests. Base Tester role doesn't know about Docker cleanup, leading to leftover containers.

### Solution

Created `.ai/roles/tester-extension.md`:

```markdown
# Tester Extension - Harvana Project

**Base Role:** `.ai-pack/roles/tester.md` (immutable, managed by ai-pack)
**Extension Type:** Project-specific Docker cleanup

---

## Additional Responsibilities

### Docker Cleanup After Tests

**When:** After completing test validation phase

**Steps:**
1. Run test suite (base Tester responsibility)
2. **Then** clean up Docker test containers:
   ```bash
   ./tools/cleanup-docker.sh
   ```
3. Verify cleanup successful:
   ```bash
   docker ps -a --filter "name=harvana-test-"
   # Should return: no containers
   ```

---

## Integration with Base Role

This extension **augments** the base Tester role.

**Workflow:**
1. Follow base Tester procedure (TDD validation, coverage check)
2. **Then** run Docker cleanup (this extension)
3. Report "APPROVED" or "CHANGES REQUIRED" to Orchestrator
```

### Documented in `.ai/repo-overrides.md`:

```markdown
## Role Extensions

### Tester Extension

**Extension Location:** `.ai/roles/tester-extension.md`
**Base Role:** `.ai-pack/roles/tester.md`
**Extension Summary:** Adds Docker container cleanup after test validation

**Key Additions:**
- Docker cleanup procedure
- Verification of cleanup success
- Integration with tools/cleanup-docker.sh

**When to Use:** Every test validation phase (Docker tests always used in Harvana)
```

---

## Common Extension Patterns

### 1. Adding Project-Specific Tools

```markdown
## Project-Specific Tools

### Database Migration Checker

**Location:** `tools/check-migrations.sh`
**Purpose:** Verify all migrations applied before tests
**Usage:**
```bash
./tools/check-migrations.sh
```

**Integration:** Run before base Tester test execution
```

### 2. Adding Domain-Specific Validation

```markdown
## Additional Quality Gates

### HIPAA Compliance Check

**Requirement:** Verify no PHI in logs or error messages
**Command:**
```bash
grep -r "ssn\|dob\|patient_id" logs/
```

**Expected Result:** No matches (exit code 1)

**Integration:** Run after base Reviewer code review
```

### 3. Adding Environment-Specific Steps

```markdown
## Environment-Specific Responsibilities

### Production Deployment Verification

**When:** Deploying to production environment

**Additional Steps:**
1. Verify feature flags configured
2. Check monitoring dashboards operational
3. Run smoke tests in production
4. Notify on-call team

**Integration:** After base Engineer deployment, before marking complete
```

---

## Extension Anti-Patterns

### ❌ Anti-Pattern 1: Editing .ai-pack/ Directly

```bash
# WRONG - Will be lost on submodule update
vim .ai-pack/roles/tester.md
```

**Why Wrong:**
- `.ai-pack/` is a git submodule
- Changes lost when submodule updates
- Breaks other projects using ai-pack
- Violates immutability principle

**Correct Approach:**
```bash
# Create extension instead
vim .ai/roles/tester-extension.md
```

### ❌ Anti-Pattern 2: Putting Role Extensions in .claude/rules/

```bash
# WRONG - .claude/rules/ is for general rules, not role extensions
vim .claude/rules/tester-extension.md
```

**Why Wrong:**
- `.claude/rules/` is for general project rules
- Role extensions belong in `.ai/roles/`
- Mixing concerns causes confusion
- Not discoverable via .ai/repo-overrides.md

**Correct Approach:**
```bash
# Put role extensions in .ai/roles/
vim .ai/roles/tester-extension.md

# Document in repo-overrides
vim .ai/repo-overrides.md
```

### ❌ Anti-Pattern 3: Not Documenting Extensions

```bash
# WRONG - Extension exists but not documented
ls .ai/roles/tester-extension.md  # Exists
grep "tester-extension" .ai/repo-overrides.md  # No mention
```

**Why Wrong:**
- Extensions not discoverable
- New team members won't find them
- Orchestrator won't know to delegate with extensions
- Maintenance nightmare

**Correct Approach:**
```bash
# Always document in repo-overrides.md
vim .ai/repo-overrides.md
# Add to "Role Extensions" section
```

### ❌ Anti-Pattern 4: Duplicating Base Role Content

```markdown
# WRONG - Copying base role content
# Tester Extension

## Base Role Responsibilities (COPIED)
[Entire base role content duplicated here]

## Additional Responsibilities
[Extension-specific content]
```

**Why Wrong:**
- Duplication causes sync issues
- Base role updates not reflected
- Increases maintenance burden
- Creates confusion about source of truth

**Correct Approach:**
```markdown
# CORRECT - Reference base role, don't duplicate
# Tester Extension - Harvana Project

**Base Role:** `.ai-pack/roles/tester.md` (see there for base responsibilities)

## Additional Responsibilities
[Extension-specific content only]
```

---

## Verification Checklist

After creating a role extension:

```
✅ Extension file exists in .ai/roles/<role-name>-extension.md
✅ Extension references base role path
✅ Extension documented in .ai/repo-overrides.md
✅ Extension committed to git
✅ Base role in .ai-pack/ unchanged (verify: git status .ai-pack/)
✅ Extension does NOT duplicate base role content
✅ Team members aware of extension (via CLAUDE.md or team docs)
```

---

## Contributing Extensions to ai-pack

If your extension is **universally applicable** (not project-specific), consider contributing it to ai-pack itself:

1. **Discuss with team** - Is this truly universal?
2. **Open issue** - Propose addition to base role
3. **Submit PR** - Update base role in ai-pack repo
4. **Update submodule** - After merged, update all projects

**Example of universal addition:**
- Adding C# quality checks to Tester role
- Adding accessibility validation to Reviewer role

**Example of project-specific (keep as extension):**
- Harvana's Docker cleanup (not all projects use Docker)
- HIPAA compliance checks (not all projects are healthcare)

---

## FAQ

### Q: Can I extend any ai-pack role?

**A:** Yes! Any role in `.ai-pack/roles/` can be extended. Common extensions:
- Tester (project-specific quality checks)
- Engineer (domain-specific steps)
- Reviewer (compliance checks)
- Orchestrator (coordination patterns)

### Q: Can I create completely new roles?

**A:** Yes! Create in `.ai/roles/new-role.md` (not extension pattern). Document in `.ai/repo-overrides.md` under "Custom Roles" section.

### Q: What if I need to override base role behavior?

**A:** Extensions **augment**, they don't override. If you need to change base behavior:
1. Discuss with ai-pack maintainers
2. Consider if base role should be updated
3. If truly project-specific, create extension with clear guidance

### Q: Can extensions reference other extensions?

**A:** Yes, but keep it simple. Document dependencies in `.ai/repo-overrides.md`.

### Q: What if base role changes?

**A:** Extensions reference base role, so updates propagate automatically. Review extensions after submodule updates to ensure compatibility.

---

## Summary

**The Golden Rule:**
```
.ai-pack/ = IMMUTABLE (read only, shared across projects)
.ai/      = MUTABLE (project-specific, your territory)
```

**Extension Pattern:**
1. Create `.ai/roles/<role-name>-extension.md`
2. Reference base role (don't duplicate)
3. Document in `.ai/repo-overrides.md`
4. Commit and push

**Never:**
- Edit `.ai-pack/` files
- Put role extensions in `.claude/`
- Leave extensions undocumented

---

**Last reviewed:** 2026-01-14
**Next review:** When extension patterns evolve or issues arise
