# Persistence Gates

**Version:** 1.0.0
**Last Updated:** 2026-01-07

## Overview

Persistence gates govern all file operations and state management activities. These rules ensure data integrity, consistency, and recoverability throughout the workflow.

## File Operation Policies

### 1. Read Before Modify

**Rule:** Always read a file before editing or overwriting it.

**Rationale:**
- Understand existing code patterns
- Maintain consistency
- Avoid breaking functionality
- Preserve important context

**Implementation:**
```
BEFORE Edit(file_path):
  Read(file_path)
  analyze content
  identify patterns to follow
  plan modifications
END BEFORE

BEFORE Write(file_path) IF file exists:
  Read(file_path)
  IF overwrite necessary THEN
    confirm with rationale
  END IF
END BEFORE
```

**Exceptions:** None. This gate is mandatory.

---

### 2. Atomic Operations

**Rule:** File operations should be atomic and reversible.

**Guidelines:**
- Complete operations fully or not at all
- Don't leave files in partial states
- Use version control for safety
- Make commits at logical checkpoints

**Anti-patterns:**
- Half-written files
- Partial refactorings
- Uncommitted breaking changes
- Mixed concerns in single operation

---

### 3. Idempotency Requirements

**Rule:** Operations should be idempotent where possible.

**Definition:** Running the same operation multiple times produces the same result as running it once.

**Examples:**

**Idempotent (Good):**
```python
# Setting a value - repeatable safely
config['port'] = 8080

# Adding to set - no duplicates
allowed_hosts.add('localhost')

# Ensuring directory exists
os.makedirs('output', exist_ok=True)
```

**Non-Idempotent (Requires Care):**
```python
# Appending - duplicates on repeat
log_file.append('Started server')  # Check first

# Incrementing - changes on repeat
counter += 1  # Need init check

# Deleting - fails on repeat
os.remove(file)  # Check existence first
```

**Implementation:**
- Check current state before modifying
- Make operations repeatable
- Document non-idempotent operations
- Provide rollback mechanisms

---

### 4. State Consistency

**Rule:** Maintain consistency across related files and state.

**Scenarios:**

#### Multi-File Changes
When modifying multiple related files:
```
WHEN changing interface:
  update implementation files
  update test files
  update documentation
  verify consistency across all
END WHEN
```

#### Configuration Changes
```
WHEN updating config:
  validate new configuration
  check for dependent settings
  update related configs
  verify system still works
END WHEN
```

#### Refactoring
```
WHEN renaming/moving:
  update all references
  update imports/includes
  update tests
  update documentation
  verify build succeeds
END WHEN
```

---

### 5. Version Control Integration

**Rule:** All significant changes must go through version control.

**Requirements:**
- All work committed to git
- Clear, descriptive commit messages
- Atomic commits (one logical change)
- No destructive git operations

**Commit Message Format:**
```
<type>: <short summary>

<detailed description>

🤖 Generated with Claude Code
Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `refactor`: Code restructuring
- `test`: Test additions/modifications
- `docs`: Documentation
- `style`: Formatting changes
- `chore`: Build/tooling changes

**Prohibited Git Operations:**
```bash
# ❌ Destructive operations without user approval
git push --force
git reset --hard
git clean -fd
git rebase -i  # interactive not supported

# ❌ Amending pushed commits
git commit --amend  # only for unpushed local commits

# ❌ Skipping hooks
git commit --no-verify
```

---

### 6. Backup and Rollback

**Rule:** Critical operations require backup and rollback capability.

**Backup Triggers:**
- Large refactorings
- Schema changes
- Configuration updates
- Deployment operations

**Implementation:**
```
BEFORE critical operation:
  create git branch OR
  create backup copy OR
  document rollback steps
END BEFORE

IF operation fails THEN
  execute rollback
  restore previous state
  report issue to user
END IF
```

**Rollback Checklist:**
- [ ] Can we revert the git commit?
- [ ] Do we have backup of modified files?
- [ ] Are database migrations reversible?
- [ ] Can we redeploy previous version?
- [ ] Are configuration changes documented?

---

### 7. Batch vs. Incremental Changes

**Rule:** Choose the appropriate change strategy based on context.

### When to Batch
- Multiple independent operations
- No dependencies between changes
- Can parallelize safely
- All-or-nothing semantics needed

**Example:**
```
Batch: Creating multiple independent test files
✓ Can create all in parallel
✓ No dependencies between them
✓ Rollback is per-file
```

### When to Incremental
- Changes have dependencies
- Need to verify each step
- Progressive enhancement
- High risk operations

**Example:**
```
Incremental: Refactoring with behavior changes
1. Add new implementation
2. Run tests → verify
3. Migrate callers
4. Run tests → verify
5. Remove old implementation
6. Run tests → verify
```

---

### 8. File Creation Policy

**Rule:** Prefer editing existing files over creating new ones.

**Creation Checklist:**
```
BEFORE creating new file:
  ✓ Search for existing similar files
  ✓ Verify functionality doesn't exist
  ✓ Confirm new file is truly needed
  ✓ User explicitly requested it OR
  ✓ No existing file can serve the purpose
END BEFORE
```

**Prohibited Unless Requested:**
- Documentation files (README.md, CHANGELOG.md)
- Configuration files
- Helper/utility files
- Test fixtures (unless part of test)

---

### 9. Directory Structure

**Rule:** Maintain and respect the project's directory structure.

**Requirements:**
- Follow established patterns
- Place files in appropriate directories
- Create directories when needed
- Don't mix concerns across directories

**Structure Verification:**
```
BEFORE writing file:
  identify correct directory
  IF directory doesn't exist THEN
    verify parent directory exists
    create directory
  END IF
  write file to correct location
END BEFORE
```

---

### 10. Data Integrity

**Rule:** Ensure data integrity throughout all operations.

**Principles:**
- Validate inputs before processing
- Verify outputs after processing
- Check constraints and invariants
- Handle edge cases properly
- Report data inconsistencies

**Validation Points:**
```
Input:  Validate before processing
During: Check invariants maintained
Output: Verify results correct
After:  Confirm state consistent
```

---

## File Type-Specific Rules

### Source Code Files
- Must compile/parse successfully
- Must pass linting
- Must follow formatting standards
- Must maintain test coverage

### Configuration Files
- Must be valid format (JSON, YAML, etc.)
- Must not break existing functionality
- Must be documented if complex
- Must have defaults for missing values

### Test Files
- Must follow test file naming conventions
- Must be in correct test directory
- Must use project's test framework
- Must not duplicate existing tests

### Documentation Files
- Only create when explicitly requested
- Must be accurate and up-to-date
- Must follow project documentation style
- Must be linked from appropriate places

---

## Integration

Persistence gates integrate with:
- **[Global Gates](00-global-gates.md)** - Safety and quality requirements
- **[Tool Policy](20-tool-policy.md)** - Tool usage for file operations
- **[Verification Gates](30-verification.md)** - Post-operation verification

---

## Violation Recovery

If persistence gate violated:

1. **Assess Impact:**
   - What files were affected?
   - Is state inconsistent?
   - Can we rollback safely?

2. **Immediate Actions:**
   - Stop further operations
   - Check version control status
   - Identify recovery path

3. **Recovery Steps:**
   - Revert problematic changes
   - Restore from backup if needed
   - Re-apply changes correctly
   - Verify system integrity

4. **Prevention:**
   - Document what went wrong
   - Update gates if needed
   - Improve validation

---

**Last reviewed:** 2026-01-07
**Next review:** Quarterly or when persistence issues occur
