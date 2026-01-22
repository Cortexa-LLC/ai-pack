---
sidebar_position: 2
title: "Feature Workflow"
---

# Feature Workflow

**Version:** 1.0.0
**Last Updated:** 2026-01-07

## Overview

The Feature Workflow is specialized for adding new functionality to a system. It extends the Standard Workflow with additional emphasis on design, testing strategy, and rollout considerations.

**Extends:** [Standard Workflow](standard.md)

**Use for:** Adding new capabilities, implementing new user stories, creating new endpoints/interfaces.

---

## Key Differences from Standard Workflow

1. **Discovery Phase** - Extra emphasis on feature scoping and design
2. **Architecture Alignment** - Ensure feature fits system architecture
3. **Comprehensive Testing** - Unit, integration, and acceptance tests required
4. **Documentation** - User-facing documentation critical
5. **Rollout Strategy** - Consider feature flags and phased deployment

---

## Feature-Specific Phases

### Phase 0: Product Definition & Architecture Design (For Large/Complex Features)

**TRIGGER:** Large feature OR unclear requirements OR significant technical complexity

**DELEGATION STRATEGY:**

**Step 0.0: Strategic Analysis (Optional - for major initiatives)**
```text
IF new product OR major feature with market implications OR requires business case THEN
  Orchestrator delegates to Strategist
  Strategist conducts market research
  Strategist analyzes competitive landscape
  Strategist develops business case with ROI
  Strategist creates Market Requirements Document (MRD)
  Strategist recommends proceed/defer/do-not-pursue

  IF recommendation == "PROCEED" THEN
    Strategist persists MRD to docs/market/[product-name]/
    THEN proceed to Step 0.0a or 0.1 with validated market opportunity
  ELSE IF recommendation == "DEFER" OR "DO NOT PURSUE" THEN
    STOP workflow (market validation failed)
  END IF
END IF
```text

**Step 0.0a: Historical Context Investigation (Optional - for legacy systems)**
```text
IF feature integrates with OR extends legacy/unfamiliar code THEN
  Orchestrator delegates to Archaeologist
  Archaeologist investigates existing patterns
  Archaeologist reconstructs historical decisions
  Archaeologist identifies technical debt in target area
  Archaeologist assesses integration risks
  Archaeologist delivers historical context for feature work

  THEN proceed to Step 0.1 with understanding of existing system
END IF
```text

**Step 0.1: Product Definition (Optional - for large features)**
```text
IF feature is large OR requirements unclear THEN
  Orchestrator delegates to Cartographer
  Cartographer reviews MRD (if created by Strategist in Step 0.0)
  Cartographer creates PRD based on market requirements
  Cartographer defines epics and user stories
  Cartographer consults with Architect (if complex)
  Cartographer delivers requirements package

  THEN proceed to Phase 0.2 or Phase 1 with clear requirements
END IF
```text

**Step 0.1a: UX Design (Optional - for user-facing features)**
```text
IF feature has significant UI/UX work THEN
  Orchestrator delegates to Designer
  Designer reviews PRD (if exists)
  Designer conducts user research
  Designer creates user flows and journey maps
  Designer creates wireframes (HTML format for web/iOS/Android)
  Designer creates design specifications
  Designer specifies accessibility requirements
  Designer delivers design documentation

  THEN proceed to Phase 0.2 or Phase 1 with design specs
END IF
```text

**Step 0.2: Architecture Design (Optional - for complex features)**
```text
IF feature requires architecture design THEN
  Orchestrator delegates to Architect
  Architect reviews PRD (if exists)
  Architect reviews Design specs (if exist)
  Architect designs system architecture
  Architect defines API contracts and data models
  Architect creates ADRs for key decisions
  Architect delivers architecture documentation

  THEN proceed to Phase 1 with technical design
END IF
```text

**Selection Criteria:**

**Delegate to Strategist when:**
- New product initiative
- Entering new market or segment
- Major feature with competitive implications
- Requires business case justification
- Market opportunity needs validation
- Large investment decision
- Strategic direction unclear
- Competitive positioning needed

**Delegate to Archaeologist when:**
- Feature must integrate with legacy code
- Existing patterns need understanding before design
- System has unclear historical design rationale
- Technical debt in target area needs assessment
- Team unfamiliar with existing codebase
- Feature may conflict with historical constraints
- Need to understand "why" before designing "what"

**Delegate to Cartographer when:**
- Large feature with multiple components
- Requirements unclear or incomplete
- Success metrics undefined
- Multiple potential approaches
- Stakeholder alignment needed

**Delegate to Designer when:**
- User-facing feature with significant UI
- New user workflows or customer journeys
- Complex forms or interactions
- Multiple user roles with different needs
- Customer experience mapping needed
- Accessibility requirements critical
- Mobile app development (iOS/Android)
- Responsive web application
- Product owner explicitly requests UX design

**Delegate to Architect when:**
- New architecture patterns needed
- Significant system changes
- Multiple system integration
- Performance/scale requirements
- Data model changes needed
- Technology decisions required

**Skip Phase 0 when:**
- Feature is small and straightforward
- Requirements are clear and complete
- Following established patterns
- No architectural changes needed
- No significant UI/UX work needed

**Deliverables from Phase 0:**
- **From Strategist (if invoked):** MRD, competitive analysis, business case, market research
- **From Cartographer (if invoked):** PRD, epics, user stories with acceptance criteria
- **From Designer (if invoked):** User research, user flows, wireframes (HTML), design specs, accessibility requirements
- **From Architect (if invoked):** Architecture document, API specifications, data models, ADRs

**CRITICAL: Artifact Persistence**

When Phase 0 completes and implementation begins, planning artifacts MUST be persisted to repository:

```text
WHEN Strategist phase complete:
  persist artifacts to docs/market/[product-name]/
  commit: MRD, competitive analysis, business case, market research
  see roles/strategist.md "Artifact Persistence" section

WHEN Cartographer phase complete:
  persist artifacts to docs/product/[feature-name]/
  commit: PRD, epics, user stories
  see roles/cartographer.md "Artifact Persistence" section

WHEN Designer phase complete:
  persist artifacts to docs/design/[feature-name]/
  persist wireframes to docs/design/[feature-name]/wireframes/
  commit: User research, user flows, wireframes (HTML), design specs
  see roles/designer.md "Artifact Persistence" section

WHEN Architect phase complete:
  persist artifacts to docs/architecture/[feature-name]/
  persist ADRs to docs/adr/
  commit: Architecture docs, API specs, data models, ADRs
  see roles/architect.md "Artifact Persistence" section

WHY:
  - Planning artifacts document decisions for years
  - Engineers reference docs during implementation
  - Future teams need context for "why" decisions were made
  - Version control tracks requirement/design evolution
  - Single source of truth for product and technical specifications
  - HTML wireframes can be opened in browser for easy review
```text

**Repository Structure After Phase 0:**
```text
project-root/
├── docs/
│   ├── product/[feature-name]/
│   │   ├── prd.md
│   │   ├── epics.md
│   │   └── user-stories.md
│   ├── design/[feature-name]/
│   │   ├── user-research.md
│   │   ├── user-flows.md
│   │   ├── design-specs.md
│   │   └── wireframes/
│   │       ├── wireframe-web.html
│   │       ├── wireframe-ios.html
│   │       └── wireframe-android.html
│   ├── architecture/[feature-name]/
│   │   ├── architecture.md
│   │   ├── api-spec.md
│   │   └── data-models.md
│   └── adr/
│       └── NNN-[decision-title].md
└── .ai/
    └── tasks/ (work-in-progress, temporary)
```text

---

### Phase 1: Feature Discovery & Scoping

**Objective:** Fully understand the feature and its implications.

#### 1.1 Feature Clarification
```text
□ What problem does this solve?
□ Who are the users?
□ What are the use cases?
□ What are the acceptance criteria?
□ What are non-requirements (out of scope)?
□ Are there UI/UX considerations?
```text

#### 1.2 Design Considerations
```text
□ How does this fit existing architecture?
□ What components are affected?
□ Are new abstractions needed?
□ What are the data models?
□ What are the interfaces/APIs?
□ Performance requirements?
□ Security requirements?
```text

#### 1.3 Dependencies and Integration
```text
□ What existing systems integrate?
□ Are external services needed?
□ Database schema changes required?
□ API version implications?
□ Backward compatibility needed?
```text

---

### Phase 2: Feature Design & Planning

**Objective:** Design a complete, well-integrated feature.

#### 2.1 Architecture Alignment
```text
□ Review with system architecture
□ Identify layer placements:
  - Presentation layer
  - Business logic layer
  - Data access layer
□ Define interfaces/contracts
□ Plan for extensibility
□ Consider SOLID principles
```text

**Example Architecture:**
```text
┌─────────────────────────────────┐
│ Presentation Layer              │
│ - API endpoints                 │
│ - Input validation              │
│ - Response formatting           │
└─────────────────────────────────┘
           ↓
┌─────────────────────────────────┐
│ Business Logic Layer            │
│ - Feature service               │
│ - Business rules                │
│ - Orchestration                 │
└─────────────────────────────────┘
           ↓
┌─────────────────────────────────┐
│ Data Access Layer               │
│ - Repository/DAO                │
│ - Database queries              │
│ - External service calls        │
└─────────────────────────────────┘
```text

#### 2.2 Implementation Approach
```text
□ Break feature into stories/tasks
□ Define implementation sequence
□ Plan for incremental delivery
□ Identify vertical slices (end-to-end)
□ Plan for testing at each layer
```text

**Incremental Delivery Strategy:**
```text
Iteration 1: Core functionality (minimal viable)
Iteration 2: Error handling and edge cases
Iteration 3: Performance optimization
Iteration 4: Additional features/polish
```text

#### 2.3 Testing Strategy
```text
Testing Pyramid for Features:

          /\
         /  \        E2E/Acceptance Tests
        /────\       - User flows
       /      \      - Happy path + key scenarios
      /────────\     Integration Tests
     /          \    - Component integration
    /────────────\   - API contracts
   /──────────────\  Unit Tests
  /                \ - Business logic
 /──────────────────\- Edge cases, validation
```text

**Test Coverage Requirements:**
- Unit tests: 90%+ of business logic
- Integration tests: All component interactions
- Acceptance tests: All user stories/scenarios
- Performance tests: If feature has performance requirements

---

### Phase 3: Feature Implementation

**Objective:** Build feature with quality and testability.

#### 3.1 TDD with Feature Development
```text
FOR each feature component:
  1. Write acceptance test (defines feature behavior)
  2. Write integration test (defines component contracts)
  3. Write unit test (defines logic)
  4. Implement to pass tests
  5. Refactor for quality
  6. Verify all tests pass
END FOR
```text

#### 3.2 Implementation Checklist
```text
□ API/Interface layer implemented
□ Business logic implemented
□ Data access implemented
□ Input validation comprehensive
□ Error handling robust
□ Logging appropriate
□ Tests comprehensive
□ Performance acceptable
```text

---

### Phase 4: Feature Validation & Documentation

**Objective:** Ensure feature is complete, documented, and ready for users.

#### 4.1 Feature Validation
```text
□ All acceptance criteria met
□ All user stories completed
□ All test scenarios pass
□ Edge cases handled
□ Error scenarios handled
□ Performance meets requirements
□ Security validated
```text

#### 4.2 Documentation Requirements
```text
Required Documentation:
□ User documentation (how to use feature)
□ API documentation (for developers)
□ Configuration documentation (if applicable)
□ Migration guide (if breaking changes)
□ Known limitations documented
```text

#### 4.3 Deployment Considerations
```text
□ Feature flags (if phased rollout)
□ Database migrations (if schema changes)
□ Configuration changes
□ Dependencies updated
□ Monitoring/metrics in place
□ Rollback plan defined
```text

---

## Feature Flags and Rollout

### When to Use Feature Flags

```text
USE feature flags when:
✓ Feature is large/complex
✓ Phased rollout desired
✓ A/B testing needed
✓ Risk is high
✓ Easy rollback required
```text

### Feature Flag Implementation

```javascript
// Example feature flag check
if (featureFlags.isEnabled('new-user-profile')) {
  // New feature implementation
  return renderNewUserProfile(user);
} else {
  // Fallback to existing implementation
  return renderLegacyUserProfile(user);
}
```text

### Rollout Strategy

```text
Phase 1: Internal testing (dev team)
Phase 2: Beta users (subset of users)
Phase 3: Gradual rollout (10% → 50% → 100%)
Phase 4: Full deployment
Phase 5: Remove feature flag (after stable period)
```text

---

## Feature-Specific Gates

### Design Gate

```text
✓ Architecture reviewed and approved
✓ Design aligns with system patterns
✓ Interfaces well-defined
✓ No architectural violations
✓ Scalability considered
```text

### Implementation Gate

```text
✓ All layers implemented
✓ Unit tests comprehensive
✓ Integration tests complete
✓ Acceptance tests pass
✓ Performance acceptable
✓ Security validated
```text

### Documentation Gate

```text
✓ User documentation complete
✓ API documentation complete
✓ Examples provided
✓ Edge cases documented
✓ Limitations noted
```text

### Deployment Readiness Gate

```text
✓ Feature flags configured (if used)
✓ Migrations tested
✓ Rollback plan defined
✓ Monitoring in place
✓ Team trained (if needed)
```text

---

## Example: User Authentication Feature

### Discovery
```text
Feature: User authentication system
Users: All application users
Problem: Need secure user login and session management

Scope:
✓ Email/password registration
✓ Login with JWT tokens
✓ Logout (token invalidation)
✓ Password reset via email

Out of Scope:
✗ OAuth/social login (future)
✗ Two-factor authentication (future)
✗ Single sign-on (future)
```text

### Design
```text
Architecture:
- POST /api/auth/register
  - Validates input
  - Hashes password (bcrypt)
  - Creates user record
  - Returns JWT token

- POST /api/auth/login
  - Validates credentials
  - Generates JWT token
  - Returns token + user info

- POST /api/auth/logout
  - Invalidates token
  - Clears session

Components:
- AuthController (API endpoints)
- AuthService (business logic)
- UserRepository (data access)
- TokenService (JWT operations)
- EmailService (password reset)
```text

### Implementation
```text
Week 1: Core authentication
- Registration endpoint
- Login endpoint
- Token generation
- Basic tests

Week 2: Security & Edge Cases
- Password hashing
- Token validation
- Error handling
- Comprehensive tests

Week 3: Password Reset
- Reset request flow
- Email integration
- Reset token handling
- Tests

Week 4: Documentation & Deployment
- API documentation
- User guide
- Security review
- Deployment
```text

---

## Common Pitfalls to Avoid

### Over-Engineering
```text
❌ Building for future requirements
❌ Premature abstractions
❌ Too many configuration options
✅ Build for current needs
✅ Refactor when needed
✅ YAGNI principle
```text

### Under-Testing
```text
❌ Only happy path tests
❌ Missing integration tests
❌ No acceptance tests
✅ Test pyramid approach
✅ Edge cases covered
✅ Error paths tested
```text

### Poor Integration
```text
❌ Ignoring existing patterns
❌ Inconsistent error handling
❌ Different naming conventions
✅ Follow established patterns
✅ Maintain consistency
✅ Review integration points
```text

---

## Success Criteria

A feature is complete when:
```text
✓ All acceptance criteria met
✓ All user stories implemented
✓ Tests comprehensive and passing
✓ Documentation complete
✓ Security validated
✓ Performance acceptable
✓ Deployed successfully
✓ User feedback positive
```text

---

## References

- [Standard Workflow](standard.md) - Base workflow
- [Engineering Standards](../quality/engineering-standards.md)
- [Design Principles](../quality/clean-code/01-design-principles.md)
- [SOLID Principles](../quality/clean-code/02-solid-principles.md)
- [Testing Guidelines](../quality/clean-code/04-testing.md)
- [Deployment Patterns](../quality/clean-code/08-deployment-patterns.md)

---

**Last reviewed:** 2026-01-07
**Next review:** Quarterly or when feature workflow evolves
