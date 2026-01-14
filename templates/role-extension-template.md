# [Role Name] Extension - [Project Name]

**Base Role:** `.ai-pack/roles/[role-name].md` (immutable, managed by ai-pack)
**Extension Type:** Project-specific additions
**Last Updated:** [Date]

---

## Extension Purpose

[Brief explanation of why this extension exists. What project-specific need does it address?]

**Example:**
> This extension adds Docker container cleanup procedures to the base Tester role
> because consumer-project uses Docker for integration tests and needs automated cleanup
> to prevent resource exhaustion.

---

## Additional Responsibilities

### [New Responsibility Name]

**When:** [When does this additional responsibility apply?]

**Purpose:** [Why is this needed?]

**Steps:**
1. [Step 1 - be specific]
2. [Step 2 - be specific]
3. [Step 3 - be specific]

**Example:**
```[language]
# Example command or code
./tools/cleanup-docker.sh
```

**Expected Outcome:** [What should happen after these steps?]

**Integration Point:** [When in the base role workflow does this happen?]
- Before: [base role step X]
- After: [base role step Y]
- During: [base role phase Z]

---

## Project-Specific Tools

### [Tool/Script Name]

**Location:** `[path/to/tool]`
**Purpose:** [What does this tool do?]

**Usage:**
```bash
[command with arguments]
```

**Expected Output:**
```
[Example of successful output]
```

**Error Handling:**
```
IF [error condition] THEN
  [How to handle the error]
  [Whether to block or warn]
END IF
```

---

## Additional Quality Gates

### [Gate Name]

**Requirement:** [What must be checked or verified?]

**Verification Command:**
```bash
[command to run]
```

**Pass Criteria:** [What indicates success?]

**Fail Criteria:** [What indicates failure?]

**Action on Failure:**
- [ ] Block and report (prevents completion)
- [ ] Warn and continue (non-blocking)
- [ ] Fix automatically (if possible)

---

## Additional Configuration

### [Configuration Item]

**Location:** [Where is this configured?]

**Required Values:**
```[format]
[configuration example]
```

**Validation:**
```bash
[How to verify configuration is correct]
```

---

## Integration with Base Role

This extension **augments** the base role defined in `.ai-pack/roles/[role-name].md`.

**Base role responsibilities are unchanged.** This extension adds:
- [Addition 1]
- [Addition 2]
- [Addition 3]

**Workflow Integration:**
```
1. Follow base role procedure (.ai-pack/roles/[role-name].md)
   - [Base step 1]
   - [Base step 2]

2. THEN apply extension-specific steps (this document)
   - [Extension step 1]
   - [Extension step 2]

3. Report completion to Orchestrator
   - Include both base role outcomes AND extension outcomes
```

---

## Environment-Specific Behavior

### [Environment Name]

**When:** [In what environment does this apply?]

**Differences:**
- [Difference 1]
- [Difference 2]

**Example:**
```
IF environment == "production" THEN
  [production-specific steps]
ELSE IF environment == "staging" THEN
  [staging-specific steps]
END IF
```

---

## Common Issues and Solutions

### Issue: [Common problem]

**Symptoms:**
- [Symptom 1]
- [Symptom 2]

**Diagnosis:**
```bash
[How to confirm this is the issue]
```

**Resolution:**
```bash
[How to fix]
```

**Prevention:**
[How to avoid this issue in the future]

---

## Examples

### Example 1: [Scenario Name]

**Context:** [When does this scenario occur?]

**Procedure:**
```
1. [Step 1]
2. [Step 2]
3. [Step 3]
```

**Result:** [Expected outcome]

### Example 2: [Another Scenario Name]

**Context:** [When does this scenario occur?]

**Procedure:**
```
1. [Step 1]
2. [Step 2]
```

**Result:** [Expected outcome]

---

## References

- **Base Role:** [.ai-pack/roles/[role-name].md](../.ai-pack/roles/[role-name].md)
- **Project Overrides:** [.ai/repo-overrides.md](../repo-overrides.md)
- **Related Tools:** [tools/[tool-name].sh](../../tools/[tool-name].sh)
- **Project Documentation:** [docs/[relevant-doc].md](../../docs/[relevant-doc].md)

---

## Success Criteria

This extension is successful when:
- ✓ [Success criterion 1]
- ✓ [Success criterion 2]
- ✓ [Success criterion 3]
- ✓ Base role outcomes achieved
- ✓ Extension-specific outcomes achieved
- ✓ No conflicts with base role workflow

---

**Last reviewed:** [Date]
**Next review:** [When to review this extension - e.g., "Quarterly" or "When base role changes"]
