---
sidebar_position: 6
title: "Persistence Gates"
---

# Persistence Gates

**Version:** 1.0.0
**Last Updated:** 2026-01-07

## Overview

Persistence gates govern all file operations and state management activities. These rules ensure data integrity, consistency, and recoverability throughout the workflow.

## File Operation Policies

### 0. Absolute Path Requirement (CRITICAL - BLOCKING)

**Rule:** ALL file and directory creation MUST use absolute paths OR verify working directory first.

**⚠️ CRITICAL: This prevents disasters like nested `server/server/API/` directories (real Harvana incident)**

**Mandatory Verification Procedure:**
```text
BEFORE creating ANY file or directory:
  STEP 1: Get project root absolute path
    PROJECT_ROOT=$(git rev-parse --show-toplevel)
    # Or: PROJECT_ROOT=$(pwd)  # if already in root
    echo "Project root: $PROJECT_ROOT"

  STEP 2: Verify you're in expected directory
    pwd  # Check current location
    # Expected: /home/user/project
    # If wrong: cd "$PROJECT_ROOT"

  STEP 3: Use absolute paths for all Write/mkdir operations
    Write(file_path="$PROJECT_ROOT/src/api/routes.js")
    mkdir "$PROJECT_ROOT/server/API"

  ❌ NEVER do this:
    Write(file_path="server/api/file.js")  # Where does this go?
    mkdir server/API                        # Creates anywhere!
    mkdir -p server/routes/api              # Disaster waiting to happen!

  ✅ ALWAYS do this:
    Write(file_path="/home/user/project/server/api/file.js")
    mkdir /home/user/project/server/API
    mkdir -p /home/user/project/server/routes/api

  OR verify location first:
    cd /home/user/project
    pwd  # Confirm: /home/user/project
    mkdir server/API  # Now safe
END BEFORE
```text

**Why This Rule Exists:**
- **Real Example (Harvana):** Agent created nested `server/server/API/` and other duplicated folders because it wasn't in the expected directory when using relative paths
- Prevents accidental nested directories
- Eliminates file location ambiguity
- Ensures reproducible file structure
- Reduces debugging time by hours

**Common Mistake Pattern:**
```bash
# Agent's mental model: "I'm in project root"
# Reality: Agent is in /home/user/project/server/

mkdir server/API
# Creates: /home/user/project/server/server/API (WRONG! NESTED!)

# Also creates other duplicates:
mkdir -p server/routes/api
# Creates: /home/user/project/server/server/routes/api (DISASTER!)
```text

**Correct Approach:**
```bash
# Option 1: Always use absolute paths
PROJECT_ROOT=$(git rev-parse --show-toplevel)
mkdir -p "$PROJECT_ROOT/server/API"
mkdir -p "$PROJECT_ROOT/server/routes/api"

# Option 2: Verify location first
pwd  # Check: /home/user/project/server/ (not root!)
cd /home/user/project  # Go to actual root
pwd  # Verify: /home/user/project
mkdir server/API  # Creates: /home/user/project/server/API (CORRECT!)
```text

**Enforcement:** This is a BLOCKING rule. File operations without path verification will be REJECTED in code review.

**Detection:**
```bash
# If you see nested directories like these, absolute paths were not used:
server/server/API/          # Should be: server/API/
client/client/components/   # Should be: client/components/
docs/docs/architecture/     # Should be: docs/architecture/

# Fix by removing nested duplicates:
# WARNING: Verify contents before removing!
```text

---

### 1. Read Before Modify

**Rule:** Always read a file before editing or overwriting it.

**Rationale:**
- Understand existing code patterns
- Maintain consistency
- Avoid breaking functionality
- Preserve important context

**Implementation:**
```text
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
```text

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
```text

**Non-Idempotent (Requires Care):**
```python
# Appending - duplicates on repeat
log_file.append('Started server')  # Check first

# Incrementing - changes on repeat
counter += 1  # Need init check

# Deleting - fails on repeat
os.remove(file)  # Check existence first
```text

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
```text
WHEN changing interface:
  update implementation files
  update test files
  update documentation
  verify consistency across all
END WHEN
```text

#### Configuration Changes
```text
WHEN updating config:
  validate new configuration
  check for dependent settings
  update related configs
  verify system still works
END WHEN
```text

#### Refactoring
```text
WHEN renaming/moving:
  update all references
  update imports/includes
  update tests
  update documentation
  verify build succeeds
END WHEN
```text

---

### 5. Version Control Integration

**Rule:** All significant changes must go through version control.

**Requirements:**
- All work committed to git
- Clear, descriptive commit messages
- Atomic commits (one logical change)
- No destructive git operations

**Commit Message Format:**
```text
<type>: <short summary>

<detailed description>

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```text

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `refactor`: Code restructuring
- `test`: Test additions/modifications
- `docs`: Documentation
- `style`: Formatting changes
- `chore`: Build/tooling changes

**Commit Message Rules (STRICT):**
```text
✅ DO:
- Plain ASCII text only
- Start with capitalized imperative verb
- Keep summary under 72 characters
- Co-Authored-By trailers acceptable

❌ NEVER:
- Use unicode characters or emoji (no 🤖 ✨ ✅ etc.)
- Add tool signatures ("Generated with [Tool]")
- Use markdown links in commit messages
- Write vague messages ("fix bug", "update code")
```text

**Rationale:**
- Commit messages document changes, not tools used
- Plain text ensures universal readability
- No encoding issues across environments
- Professional appearance in all contexts

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
```text

---

### 6. Backup and Rollback

**Rule:** Critical operations require backup and rollback capability.

**Backup Triggers:**
- Large refactorings
- Schema changes
- Configuration updates
- Deployment operations

**Implementation:**
```text
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
```text

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
```text
Batch: Creating multiple independent test files
✓ Can create all in parallel
✓ No dependencies between them
✓ Rollback is per-file
```text

### When to Incremental
- Changes have dependencies
- Need to verify each step
- Progressive enhancement
- High risk operations

**Example:**
```text
Incremental: Refactoring with behavior changes
1. Add new implementation
2. Run tests → verify
3. Migrate callers
4. Run tests → verify
5. Remove old implementation
6. Run tests → verify
```text

---

### 8. File Creation Policy

**Rule:** Prefer editing existing files over creating new ones.

**Creation Checklist:**
```text
BEFORE creating new file:
  ✓ Search for existing similar files
  ✓ Verify functionality doesn't exist
  ✓ Confirm new file is truly needed
  ✓ User explicitly requested it OR
  ✓ No existing file can serve the purpose
END BEFORE
```text

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
```text
BEFORE writing file:
  identify correct directory
  IF directory doesn't exist THEN
    verify parent directory exists
    create directory
  END IF
  write file to correct location
END BEFORE
```text

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
```text
Input:  Validate before processing
During: Check invariants maintained
Output: Verify results correct
After:  Confirm state consistent
```text

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

## Planning Artifact Persistence

### 11. Artifact Repository Persistence

**Rule:** Planning artifacts (PRDs, architecture docs, retrospectives) MUST be persisted to the repository when transitioning from planning to implementation/completion.

**Rationale:**
- Captures "why" decisions for future reference
- Provides long-term knowledge preservation
- Enables team onboarding and context understanding
- Creates single source of truth for requirements and designs
- Supports traceability from requirements through implementation

**Artifact Types and Destinations:**

```text
Product Manager Artifacts:
  .ai/tasks/[feature-id]/prd.md           → docs/product/YYYY-MM-DD-feature-name/prd.md
  .ai/tasks/[feature-id]/epics.md         → docs/product/YYYY-MM-DD-feature-name/epics.md
  .ai/tasks/[feature-id]/user-stories.md  → docs/product/YYYY-MM-DD-feature-name/user-stories.md

Strategist Artifacts:
  .ai/tasks/[product-id]/mrd.md           → docs/market/YYYY-MM-DD-product-name/mrd.md
  .ai/tasks/[product-id]/competitive.md   → docs/market/YYYY-MM-DD-product-name/competitive-analysis.md
  .ai/tasks/[product-id]/business-case.md → docs/market/YYYY-MM-DD-product-name/business-case.md

Architect Artifacts:
  .ai/tasks/[feature-id]/architecture.md  → docs/architecture/YYYY-MM-DD-feature-name/architecture.md
  .ai/tasks/[feature-id]/api-spec.md      → docs/architecture/YYYY-MM-DD-feature-name/api-spec.md
  .ai/tasks/[feature-id]/data-models.md   → docs/architecture/YYYY-MM-DD-feature-name/data-models.md
  .ai/tasks/[feature-id]/adrs/adr-NNN-*.md → docs/adr/adr-NNN-*.md

Inspector Artifacts:
  .ai/tasks/[bug-id]/retrospective.md     → docs/investigations/[bug-id]-[description].md
```text

**Persistence Triggers:**

```text
WHEN Strategist phase completes THEN
  persist MRD, competitive analysis, business case to docs/market/YYYY-MM-DD-product-name/
  commit with message: "Add market requirements for [product-name]"
END WHEN

WHEN Product Manager phase completes THEN
  persist PRD, epics, user stories to docs/product/YYYY-MM-DD-feature-name/
  commit with message: "Add product requirements for [feature-name]"
END WHEN

WHEN Architect phase completes THEN
  persist architecture docs to docs/architecture/YYYY-MM-DD-feature-name/
  persist ADRs to docs/adr/
  commit with message: "Add architecture design for [feature-name]"
END WHEN

WHEN bug fix verified and accepted THEN
  persist retrospective to docs/investigations/
  commit with message: "Add retrospective for [bug-id]: [description]"
END WHEN
```text

**Cross-Reference Requirements (MANDATORY):**

All persisted artifacts MUST include a "Related Documents" section to enable traceability:

```text
MRD (docs/market/YYYY-MM-DD-product-name/mrd.md):
  ## Related Documents
  - PRD: [Link to docs/product/YYYY-MM-DD-feature-name/prd.md]
  - Business Case: [Link to business-case.md]
  - Competitive Analysis: [Link to competitive-analysis.md]

PRD (docs/product/YYYY-MM-DD-feature-name/prd.md):
  ## Related Documents
  - MRD: [Link to docs/market/YYYY-MM-DD-product-name/mrd.md] (if applicable)
  - Architecture: [Link to docs/architecture/YYYY-MM-DD-feature-name/]
  - User Stories: [Link to epics.md and user-stories.md]
  - ADRs: [Links to relevant ADRs]

Architecture Doc (docs/architecture/YYYY-MM-DD-feature-name/architecture.md):
  ## Related Documents
  - MRD: [Link to docs/market/YYYY-MM-DD-product-name/mrd.md] (if applicable)
  - PRD: [Link to docs/product/YYYY-MM-DD-feature-name/prd.md]
  - User Stories: [Link to docs/product/YYYY-MM-DD-feature-name/user-stories.md]
  - Related ADRs:
    - [ADR-NNN: Decision Title](../adr/NNN-decision-title.md)
  - Implementation: [Will be referenced by Engineers]

Bug Retrospective (docs/investigations/[bug-id]-description.md):
  ## Related Documents
  - Architecture: [Link to relevant architecture docs]
  - ADRs: [Link to relevant ADRs explaining design decisions]
  - Similar Bugs: [Links to related retrospectives]
  - Original Bug Report: [Reference]

Implementation (code comments and commits):
  // Implements: docs/architecture/2024-03-15-billing-system/api-spec.md
  // Requirement: FR-123 from docs/product/2024-03-10-billing-system/prd.md

  git commit message:
    feat: Add billing API endpoints

    Implements requirements from docs/architecture/2024-03-15-billing-system/api-spec.md
    Addresses FR-1, FR-2 from docs/product/2024-03-10-billing-system/prd.md

    Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```text

**Traceability Chain:**
```text
PRD (Requirements) → Architecture (Design) → Implementation (Code) → Tests (Verification) → Requirements (Validation)
```text

**Cross-Reference Verification:**
```text
WHEN verifying artifact persistence:
  CHECK for "Related Documents" section
  VERIFY links are valid and correct
  VERIFY bidirectional references (PRD ↔ Architecture)
  IF missing or incomplete THEN
    REQUEST specialist to add cross-references
    BLOCK implementation until added
  END IF
END WHEN
```text

**Verification Checklist:**

```text
Before marking planning phase complete:
  ✓ Artifacts moved from .ai/tasks/ to docs/
  ✓ docs/market/ updated (if Strategist phase occurred)
  ✓ docs/product/ updated (if Cartographer phase occurred)
  ✓ docs/architecture/ updated (if Architect phase occurred)
  ✓ docs/adr/ updated (if ADRs created)
  ✓ docs/security/ updated (if security issues documented)
  ✓ docs/incidents/ updated (if incident documented)
  ✓ docs/investigations/ updated (if retrospective completed)
  ✓ Cross-references added where appropriate
  ✓ Changes committed to repository
  ✓ Commit message follows format standards
```text

**Directory Structure Standards:**

```text
docs/
├── market/                               # Market research and business cases
│   ├── 2024-01-10-mobile-app/           # Date-prefixed directories
│   │   ├── mrd.md                        # Market Requirements Document
│   │   ├── competitive-analysis.md
│   │   └── business-case.md
│   └── 2024-06-15-enterprise-platform/
│       └── mrd.md
├── product/                              # Product requirements
│   ├── 2024-01-15-user-authentication/  # Date-prefixed directories
│   │   ├── prd.md                        # Product Requirements Document
│   │   ├── epics.md
│   │   └── user-stories.md
│   └── 2024-03-10-billing-system/
│       ├── prd.md
│       └── epics.md
├── architecture/                         # Technical design
│   ├── 2024-01-20-user-authentication/  # Date-prefixed directories
│   │   ├── architecture.md
│   │   ├── api-spec.md
│   │   └── data-models.md
│   └── 2024-03-15-billing-system/
│       ├── architecture.md
│       └── api-spec.md
├── adr/                                  # Architecture Decision Records
│   ├── 001-decision-title.md            # Sequential across entire project
│   ├── 002-another-decision.md
│   └── README.md                         # Index of all ADRs
├── security/                             # Security vulnerabilities
│   ├── SEC-2024-001-sql-injection.md
│   ├── SEC-2024-002-xss-profile.md
│   └── README.md
├── incidents/                            # Production incidents
│   ├── INC-2024-001-database-outage.md
│   ├── INC-2024-002-api-timeout.md
│   └── README.md
└── investigations/                       # Bug retrospectives
    ├── BUG-2024-001-null-pointer.md
    ├── BUG-2024-002-race-condition.md
    └── README.md
```text

**Temporary vs Permanent:**

```text
.ai/tasks/               → Temporary work-in-progress (not committed)
docs/                    → Permanent documentation (committed)

.ai/tasks/[task-id]/     → Deleted after task completion
docs/product/            → Long-lived product requirements
docs/architecture/       → Long-lived technical designs
docs/investigations/     → Long-lived learning and patterns
docs/adr/                → Long-lived decision records
```text

**Enforcement:**

This gate is MANDATORY for:
- Product Manager role after deliverables approved
- Architect role after deliverables approved
- Inspector role after bug fix verified
- Orchestrator role to verify persistence occurred

**Violation Consequences:**
- Planning artifacts lost when .ai/tasks/ cleaned up
- Loss of institutional knowledge
- No traceability from requirements to implementation
- Future teams lack context for decisions
- Repeat past mistakes due to missing retrospectives

**References:**
- [Product Manager Role](../roles/product-manager.md) - Section: "Artifact Persistence to Repository"
- [Architect Role](../roles/architect.md) - Section: "Artifact Persistence to Repository"
- [Inspector Role](../roles/inspector.md) - Section: "Artifact Persistence to Repository"
- [Feature Workflow](../workflows/feature.md) - Phase 0: Artifact persistence
- [Bugfix Workflow](../workflows/bugfix.md) - Post-Fix: Retrospective persistence
- [Standard Workflow](../workflows/standard.md) - Section: "Artifact Persistence"

---

## Work Log → Result File

### 12. Result File Management

**Rule:** Agents write findings to `result.md` in the task packet directory.

**Rationale:**
- Task packets use a two-file format: `task.md` (brief) + `result.md` (findings)
- The KG handles crash-recovery and resume context
- result.md is the human-readable completion record

**Result File Format:**
See template at `templates/task-packet/result.md`.

**References:**
- [Engineer Role](../roles/engineer.md) - Result file requirements
- [Orchestrator Role](../roles/orchestrator.md) - Progress monitoring
- [Result Template](../templates/task-packet/result.md)

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
