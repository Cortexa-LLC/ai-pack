# Repository Overrides

**Project:** [Your Project Name]
**Last Updated:** [Date]

---

## Overview

This document tracks all project-specific customizations, extensions, and overrides to the base ai-pack framework.

**Purpose:** Central registry of all deviations from ai-pack defaults to ensure:
- ✅ Discoverability - Team members can find project-specific behavior
- ✅ Documentation - Rationale for customizations is recorded
- ✅ Maintainability - Updates to ai-pack can be evaluated against overrides
- ✅ Consistency - All extensions follow the same pattern

---

## Critical Reminder

**`.ai-pack/` is IMMUTABLE:**
```
❌ NEVER edit files in .ai-pack/
✅ DO create extensions in .ai/ (this directory)
✅ DO document all extensions here
```

---

## Role Extensions

Role extensions **augment** base roles from `.ai-pack/roles/` with project-specific additions.

### [Role Name] Extension

**Extension Location:** `.ai/roles/[role-name]-extension.md`
**Base Role:** `.ai-pack/roles/[role-name].md`
**Extension Summary:** [Brief description of what this extension adds]

**Key Additions:**
- [Addition 1]
- [Addition 2]
- [Addition 3]

**When to Use:** [Scenarios where this extension applies]

**Example:**
```markdown
### Tester Extension

**Extension Location:** `.ai/roles/tester-extension.md`
**Base Role:** `.ai-pack/roles/tester.md`
**Extension Summary:** Adds Docker container cleanup after test validation

**Key Additions:**
- Docker cleanup procedure
- Verification of cleanup success
- Integration with tools/cleanup-docker.sh

**When to Use:** Every test validation phase (Docker tests always used in this project)
```

---

## Custom Roles

Custom roles are entirely project-specific (not extensions of ai-pack roles).

### [Custom Role Name]

**Role Location:** `.ai/roles/[custom-role-name].md`
**Purpose:** [Why this custom role exists]
**Responsibilities:** [Brief list of responsibilities]
**When to Use:** [Scenarios requiring this custom role]

**Example:**
```markdown
### Compliance Checker

**Role Location:** `.ai/roles/compliance-checker.md`
**Purpose:** Verify HIPAA compliance before production deployment
**Responsibilities:**
- Check for PHI in logs
- Verify encryption settings
- Validate access controls
- Generate compliance report
**When to Use:** Before every production deployment
```

---

## Workflow Overrides

Workflow overrides modify or extend workflows from `.ai-pack/workflows/`.

### [Workflow Name] Override

**Base Workflow:** `.ai-pack/workflows/[workflow-name].md`
**Override Location:** `.ai/workflows/[workflow-name]-override.md` (if file created)
**Override Type:** [Addition | Modification | Replacement]

**Changes:**
- [Change 1]
- [Change 2]

**Rationale:** [Why this override is needed]

**Example:**
```markdown
### Deployment Workflow Addition

**Base Workflow:** `.ai-pack/workflows/feature.md`
**Override Type:** Addition (adds Phase 6: Production Deployment)

**Changes:**
- Add Phase 6: Production Deployment with approval gates
- Add smoke test verification
- Add rollback procedure

**Rationale:** Feature workflow ends at code review; we need production deployment steps
```

---

## Gate Overrides

Gate overrides modify enforcement rules from `.ai-pack/gates/`.

### [Gate Name] Override

**Base Gate:** `.ai-pack/gates/[gate-file].md`
**Override Type:** [Stricter | Relaxed | Additional]

**Changes:**
- [Change 1]
- [Change 2]

**Rationale:** [Why this override is needed]
**Approval:** [Who approved this override]

**Example:**
```markdown
### Code Coverage Gate Override

**Base Gate:** `.ai-pack/gates/35-code-quality-review.md`
**Override Type:** Stricter

**Changes:**
- Base requirement: 80-90% coverage
- Project requirement: 95% coverage minimum
- Critical paths: 100% coverage required

**Rationale:** Medical device software - FDA requires higher coverage
**Approval:** Engineering Director, 2026-01-10
```

---

## Standards Overrides

Standards overrides modify coding standards from `.ai-pack/quality/`.

### [Standard Name] Override

**Base Standard:** `.ai-pack/quality/[standard-file].md`
**Override Type:** [Addition | Modification]

**Changes:**
- [Change 1]
- [Change 2]

**Rationale:** [Why this override is needed]

**Example:**
```markdown
### Python Standards Addition

**Base Standard:** `.ai-pack/quality/clean-code/python-standards.md`
**Override Type:** Addition

**Changes:**
- Add: Type hints required for all public APIs
- Add: Pydantic models required for all data classes
- Add: Async/await required for all I/O operations

**Rationale:** High-performance async API project requires strict async patterns
```

---

## Tool and Script Overrides

Project-specific tools that interact with ai-pack workflows.

### [Tool Name]

**Location:** `tools/[tool-name].[ext]`
**Purpose:** [What this tool does]
**Integration:** [How it integrates with ai-pack workflows]

**Usage:**
```bash
[command]
```

**Example:**
```markdown
### Docker Cleanup Script

**Location:** `tools/cleanup-docker.sh`
**Purpose:** Clean up test Docker containers after test validation
**Integration:** Called by Tester role extension after test phase

**Usage:**
```bash
./tools/cleanup-docker.sh
```
```

---

## Project-Specific Constraints

Additional constraints beyond ai-pack standards.

### [Constraint Name]

**Applies to:** [Which roles/workflows/files]
**Constraint:** [What the constraint is]
**Enforcement:** [How it's enforced]
**Rationale:** [Why this constraint exists]

**Example:**
```markdown
### Database Migration Review

**Applies to:** All database migrations
**Constraint:** Migrations must be reviewed by DBA before merge
**Enforcement:** Required reviewer in GitHub CODEOWNERS
**Rationale:** Database changes affect multiple services; DBA review prevents conflicts
```

---

## Documentation Overrides

Project-specific documentation requirements beyond ai-pack standards.

### [Documentation Type]

**Requirement:** [What documentation is required]
**Location:** [Where it should be stored]
**Format:** [What format is expected]
**Trigger:** [When this documentation must be created]

**Example:**
```markdown
### API Change Documentation

**Requirement:** All public API changes require changelog entry
**Location:** `docs/api/CHANGELOG.md`
**Format:** Markdown with version, date, change type (breaking/feature/fix)
**Trigger:** Any commit modifying files in `src/api/`
```

---

## Technology-Specific Additions

Technology or language-specific requirements not covered by ai-pack.

### [Technology Name]

**Requirements:**
- [Requirement 1]
- [Requirement 2]

**Tools:**
- [Tool 1]
- [Tool 2]

**References:**
- [Link to docs]

**Example:**
```markdown
### GraphQL Schema Requirements

**Requirements:**
- All schema changes require backward compatibility check
- Schema deprecation warnings required before removal
- Schema documentation auto-generated from comments

**Tools:**
- `graphql-inspector` for compatibility checking
- `graphql-markdown` for documentation generation

**References:**
- [GraphQL Best Practices](https://graphql.org/learn/best-practices/)
```

---

## Review History

Track when overrides were added, modified, or removed.

### [Date] - [Change Type]

**Changed:** [What was changed]
**Rationale:** [Why the change]
**Approved by:** [Who approved]

**Example:**
```markdown
### 2026-01-14 - Added Tester Extension

**Changed:** Added Tester role extension for Docker cleanup
**Rationale:** Integration tests leave Docker containers, need automated cleanup
**Approved by:** Engineering Lead

### 2026-01-10 - Modified Coverage Gate

**Changed:** Increased coverage requirement from 80% to 95%
**Rationale:** FDA requirements for medical device software
**Approved by:** Engineering Director + QA Director
```

---

## Summary

**Total Overrides:** [Count]
- Role Extensions: [Count]
- Custom Roles: [Count]
- Workflow Overrides: [Count]
- Gate Overrides: [Count]
- Standards Overrides: [Count]
- Tools: [Count]
- Constraints: [Count]
- Documentation Requirements: [Count]
- Technology-Specific: [Count]

**Last Review:** [Date]
**Next Review:** [Date or trigger - e.g., "Quarterly" or "After ai-pack update"]

---

**Maintained by:** [Team or person responsible]
**Questions:** [Contact info or Slack channel]
