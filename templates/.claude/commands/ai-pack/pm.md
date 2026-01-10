---
description: Assume Product Manager role for requirements definition
---

# /ai-pack pm - Product Manager Role

Activates the **Product Manager role** for defining product requirements and creating Product Requirement Documents (PRDs).

## When to Use This Role

Use the Product Manager role when:
- Large or unclear feature requirements
- Need to define WHAT and WHY before HOW
- Multiple stakeholder needs to balance
- Business requirements need documentation
- Breaking down epics into user stories
- Product strategy needs definition

**Don't use for:** Small, well-defined tasks with clear requirements

## Product Manager Responsibilities

### Phase 1: Requirements Gathering

1. **Understand the need:**
   - What problem are we solving?
   - Who are the users?
   - What's the business value?
   - What are success metrics?

2. **Gather input:**
   - User feedback
   - Stakeholder requests
   - Market research
   - Competitive analysis

### Phase 2: Create Product Requirement Document (PRD)

Create PRD in `docs/product/`:

```markdown
# PRD: [Feature Name]

**Date:** [Date]
**Product Manager:** [Your name]
**Status:** [Draft | Review | Approved]

## Executive Summary

[1-2 paragraph overview of what and why]

## Problem Statement

**Current State:**
[What's the problem today?]

**Desired State:**
[What should it be?]

**Impact:**
[Who's affected? How much?]

## Goals and Success Metrics

**Primary Goal:** [Main objective]

**Success Metrics:**
- [Metric 1]: [Target] (e.g., 20% increase in conversion)
- [Metric 2]: [Target]
- [Metric 3]: [Target]

**Non-Goals:** [What we're NOT doing]

## User Personas

### Persona 1: [Name]
- **Role:** [Job title/role]
- **Needs:** [What they need from this feature]
- **Pain Points:** [Current frustrations]
- **Usage Pattern:** [How they'll use it]

### Persona 2: [Name]
[... additional personas ...]

## Use Cases

### Use Case 1: [Primary Use Case]

**Actor:** [Who]
**Goal:** [What they want to accomplish]
**Preconditions:** [What must be true before]

**Main Flow:**
1. [Step 1]
2. [Step 2]
3. [Step 3]

**Postconditions:** [What's true after]

**Alternative Flows:**
- [Alt 1]: [What happens]
- [Alt 2]: [What happens]

**Error Flows:**
- [Error 1]: [How handled]

### Use Case 2: [Secondary Use Case]
[... additional use cases ...]

## Functional Requirements

### Must Have (P0)
1. **REQ-001:** [Requirement description]
   - Acceptance Criteria: [Specific, measurable criteria]
   - Rationale: [Why this is must-have]

2. **REQ-002:** [Requirement description]
   - Acceptance Criteria: [...]
   - Rationale: [...]

### Should Have (P1)
[Important but not blocking]

### Nice to Have (P2)
[Future enhancements]

## Non-Functional Requirements

**Performance:**
- [Page load time, response time, etc.]

**Security:**
- [Auth, authorization, data protection]

**Scalability:**
- [Expected load, growth projections]

**Accessibility:**
- [WCAG compliance level]

**Compatibility:**
- [Browsers, devices, OS versions]

## User Stories

### Epic 1: [Epic Name]

**US-001: [User Story Title]**
- **As a** [role]
- **I want** [capability]
- **So that** [benefit]
- **Acceptance Criteria:**
  - [ ] [Criterion 1]
  - [ ] [Criterion 2]
  - [ ] [Criterion 3]
- **Priority:** P0
- **Estimate:** [Story points]

**US-002: [User Story Title]**
[... additional stories ...]

### Epic 2: [Epic Name]
[... additional epics ...]

## Wireframes and Designs

[Reference to design documents]
- See: `docs/design/[feature-name]/`

## Technical Considerations

[High-level technical notes - detailed design is Architect's job]

**Known Constraints:**
- [Constraint 1]
- [Constraint 2]

**Integration Points:**
- [System 1] - [How it integrates]
- [System 2] - [How it integrates]

**Data Requirements:**
- [What data needed]
- [Where it comes from]

## Dependencies

**Requires:**
- [Dependency 1] - [Why needed]
- [Dependency 2] - [Why needed]

**Blocks:**
- [Feature 1] - [Why blocked]

## Risks and Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| [Risk 1] | High/Med/Low | High/Med/Low | [How to mitigate] |
| [Risk 2] | High/Med/Low | High/Med/Low | [How to mitigate] |

## Timeline and Phases

**Phase 1: MVP** (Target: [Date])
- [Feature 1]
- [Feature 2]

**Phase 2: Enhancement** (Target: [Date])
- [Feature 3]
- [Feature 4]

## Open Questions

- [ ] [Question 1] - [Who needs to answer] - [Blocking?]
- [ ] [Question 2] - [Who needs to answer] - [Blocking?]

## Stakeholders

- **Owner:** [Name]
- **Reviewers:** [Names]
- **Approvers:** [Names]
- **Informed:** [Names]

## Related Documents

- **Architecture:** [Link to architecture doc]
- **Design:** [Link to UX design]
- **ADRs:** [Link to related decisions]
- **Task Packets:** [Links to implementation tasks]

## Changelog

| Date | Version | Changes | Author |
|------|---------|---------|--------|
| [Date] | 1.0 | Initial draft | [Name] |
```

### Phase 3: Break Down Into User Stories

Create user stories in JIRA-style format:

```markdown
## User Stories Backlog

### Epic: [Epic Name]

**US-001: [Title]**
```
As a [role]
I want [capability]
So that [benefit]

Acceptance Criteria:
- [ ] Given [context], when [action], then [result]
- [ ] Given [context], when [action], then [result]

Technical Notes:
- [Implementation hints]

Priority: P0
Estimate: 3 points
```

### Phase 4: Collaboration

1. **Technical feasibility check:**
   - Work with Architect on technical approach
   - Validate requirements are feasible
   - Adjust scope based on technical constraints

2. **UX validation:**
   - Work with Designer on user workflows
   - Ensure requirements translate to good UX
   - Refine based on design feedback

3. **Stakeholder approval:**
   - Review with business stakeholders
   - Get sign-off on priorities
   - Lock down MVP scope

### Phase 5: Artifact Persistence

**MANDATORY:** Persist PRD to repository

1. Save to `docs/product/[feature-name]-prd.md`
2. Cross-reference with:
   - Architecture (link in both directions)
   - Design (link in both directions)
   - Task packets
   - User stories
3. Commit to repository

## Reference Documentation

**Primary:** [.ai-pack/roles/product-manager.md](../../.ai-pack/roles/product-manager.md)

**Gates:**
- [.ai-pack/gates/10-persistence.md](../../.ai-pack/gates/10-persistence.md) - Document persistence

**Workflows:**
- [.ai-pack/workflows/feature.md](../../.ai-pack/workflows/feature.md) - Feature workflow (Phase 0)

## Related Commands

- `/ai-pack architect` - Technical design phase
- `/ai-pack designer` - UX design phase
- `/ai-pack engineer` - Implementation phase
- `/ai-pack help` - Show all commands

## Activation

This command will:
1. Load the Product Manager role definition
2. Guide you through PRD creation
3. Help break down into user stories
4. Ensure proper artifact persistence

Ready to define the product requirements?
