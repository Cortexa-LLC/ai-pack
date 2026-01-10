---
paths: **/*
---

# Task Packet Requirements

Task packets are the fundamental unit of work tracking in the ai-pack framework.

## When Required

**MANDATORY for all non-trivial tasks:**
- Requires >2 steps
- Involves code changes
- Takes >30 minutes
- Needs verification

**NOT required for:**
- Trivial one-liners
- Reading/exploring code
- Answering questions

## Creating Task Packets

```bash
# Use the command
/ai-pack task-init <task-name>

# Creates:
.ai/tasks/YYYY-MM-DD_task-name/
├── 00-contract.md    # Requirements, acceptance criteria
├── 10-plan.md        # Implementation approach
├── 20-work-log.md    # Progress tracking
├── 30-review.md      # Quality assurance
└── 40-acceptance.md  # Sign-off
```

## Task Lifecycle

### Phase 1: Contract (00-contract.md)

**Define:**
- What needs to be done
- Why it's needed
- Acceptance criteria
- Stakeholders

**Example:**
```markdown
## Requirements

Implement user login with email/password authentication.

## Acceptance Criteria

- [ ] User can login with valid credentials
- [ ] Invalid credentials show error
- [ ] Session persists across page reloads
- [ ] Logout clears session

## Stakeholders

- Product: Jane Smith
- Tech Lead: John Doe
```

### Phase 2: Plan (10-plan.md)

**Document:**
- Implementation approach
- Workflow selection (feature, bugfix, refactor, etc.)
- Role assignment
- Execution strategy (parallel vs sequential)
- Dependencies and risks

**Example:**
```markdown
## Workflow

Using: .ai-pack/workflows/feature.md

## Approach

1. Create LoginForm component
2. Add authentication API endpoint
3. Implement session management
4. Add logout functionality

## Execution Strategy

Sequential (shared auth service):
1. API endpoint first (foundation)
2. UI components (depends on API)
3. Tests (validates both)

## Risks

- Session security concerns → Use HTTPOnly cookies
```

### Phase 3: Work Log (20-work-log.md)

**Track during implementation:**
- Progress updates
- Decisions made
- Blockers encountered
- Next steps

**Update frequently** - Document as you go.

**Example:**
```markdown
## 2026-01-10 10:00

**Progress:**
- Created LoginForm component with TDD
- Added form validation

**Decisions:**
- Using Formik for form handling
- Email regex validation client-side

**Blockers:**
- None

**Next:**
- Implement API endpoint
```

### Phase 4: Review (30-review.md)

**Quality assurance documentation:**
- Tester assessment
- Reviewer assessment
- Issues found
- Resolution status

**Filled by Tester and Reviewer roles.**

### Phase 5: Acceptance (40-acceptance.md)

**Final sign-off:**
- Verification all criteria met
- Deviations from plan
- Lessons learned
- Final status

## Location Rules

**✅ CORRECT:**
```
.ai/tasks/YYYY-MM-DD_task-name/
```

**❌ NEVER:**
```
.ai-pack/  (Framework is read-only)
```

## Task Packet Commands

```bash
# Create new task
/ai-pack task-init <name>

# Check status
/ai-pack task-status

# Choose role
/ai-pack engineer       # Simple tasks
/ai-pack orchestrate    # Complex tasks
```

## Reference

Full documentation: `.ai-pack/gates/00-global-gates.md`
