# Task Contract

**Task ID:** [YYYY-MM-DD_task-name]
**Created:** [Date]
**Requestor:** [User/Team]
**Assigned Role:** [Orchestrator/Worker/Reviewer]
**Workflow:** [Standard/Feature/Bugfix/Refactor/Research]

---

## Task Description

[Clear, concise description of what needs to be done]

### Background and Context

[Why is this task needed? What problem does it solve? What is the larger context?]

### Current State

[What is the current situation? What exists today?]

### Desired State

[What should exist after this task is complete?]

---

## Success Criteria

Define objective, measurable criteria for completion:

```
✓ [Criterion 1: Specific, measurable outcome]
✓ [Criterion 2: Specific, measurable outcome]
✓ [Criterion 3: Specific, measurable outcome]
```

**Examples:**
- ✓ All tests passing (142/142)
- ✓ Code coverage ≥ 85%
- ✓ API endpoint returns expected response format
- ✓ User can complete workflow without errors
- ✓ Performance < 200ms response time

---

## Acceptance Criteria

Detailed checklist of requirements that must be met:

### Functional Requirements
```
□ [Requirement 1]
□ [Requirement 2]
□ [Requirement 3]
```

### Quality Requirements
```
□ All tests passing
□ Code coverage 80-90%
□ No linting errors
□ Code review approved
□ Documentation complete
```

### Non-Functional Requirements
```
□ Performance acceptable
□ Security validated
□ Accessibility considered
□ Error handling robust
```

---

## Constraints and Dependencies

### Constraints
```
□ [Technical constraint]
□ [Business constraint]
□ [Time constraint]
□ [Resource constraint]
```

### Dependencies
```
□ [Dependency on other task/feature]
□ [Dependency on external service]
□ [Dependency on team member]
□ [Dependency on tool/library]
```

### Out of Scope
```
✗ [Explicitly not included 1]
✗ [Explicitly not included 2]
✗ [Explicitly not included 3]
```

---

## Estimated Complexity

**Complexity:** [Trivial | Small | Medium | Large | Very Large]

**Rationale:**
- Number of files affected: [X]
- Lines of code estimate: [~X]
- New concepts/patterns: [Yes/No]
- Integration complexity: [Low/Medium/High]
- Risk level: [Low/Medium/High]

---

## Resources and References

### Relevant Files
```
- path/to/file1.ext - [Description]
- path/to/file2.ext - [Description]
```

### Documentation
```
- [Link to design doc]
- [Link to API spec]
- [Link to related issue]
```

### Examples
```
- path/to/example.ext - [Similar implementation]
- [External reference/tutorial]
```

---

## Assumptions

```
1. [Assumption 1]
2. [Assumption 2]
3. [Assumption 3]
```

*Note: If any assumption proves invalid, revisit this contract.*

---

## Risk Assessment

### Identified Risks
```
1. [Risk 1]
   - Probability: [Low/Medium/High]
   - Impact: [Low/Medium/High]
   - Mitigation: [Strategy]

2. [Risk 2]
   - Probability: [Low/Medium/High]
   - Impact: [Low/Medium/High]
   - Mitigation: [Strategy]
```

---

## Approvals and Sign-Off

**Contract Approved By:**
- [ ] Requestor: [Name] [Date]
- [ ] Agent: [Role] [Date]

**Changes to Contract:**
[Document any contract changes here with date and rationale]

---

## Notes

[Any additional notes, clarifications, or context]

---

**Contract Version:** 1.0
**Last Updated:** [Date]

---

## Usage Instructions

This template should be instantiated at: `.ai/tasks/YYYY-MM-DD_task-name/00-contract.md`

**When to create:**
- At the start of any new task
- Before planning or implementation begins

**Who creates it:**
- Orchestrator (for complex tasks)
- Worker (for assigned tasks)
- User (can provide initial version)

**Key principles:**
- Be specific and measurable
- Clarify ambiguities upfront
- Document assumptions
- Get agreement before proceeding
