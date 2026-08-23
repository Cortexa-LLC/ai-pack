---
sidebar_position: 4
title: "Research Workflow"
---

# Research Workflow

**Version:** 1.0.0
**Last Updated:** 2026-01-07

## Overview

The Research Workflow is specialized for investigating and understanding codebases, technologies, patterns, and problems. It emphasizes thorough exploration, knowledge synthesis, and clear documentation of findings.

**Extends:** [Standard Workflow](standard.md)

**Use for:** Understanding code, investigating issues, evaluating technologies, learning architecture, answering "how does X work?" questions.

---

## Key Differences from Standard Workflow

1. **Exploration Focus** - Deep investigation over quick answers
2. **Knowledge Synthesis** - Connect findings into coherent understanding
3. **Documentation Heavy** - Findings must be documented clearly
4. **No Implementation** - Research phase only, no code changes
5. **Iterative Deepening** - Start broad, then drill into specifics

---

## Research-Specific Phases

### Phase 1: Research Objectives

**Objective:** Define what needs to be understood and why.

#### 1.1 Question Formulation
```text
Define Research Questions:
□ What specifically needs to be understood?
□ Why is this information needed?
□ What decisions depend on this research?
□ How deep does understanding need to go?
□ What is success criteria for research?
```text

**Examples:**
```text
Broad: "How does the authentication system work?"
Specific: "What token validation flow is used?"

Broad: "How is data persisted?"
Specific: "What ORM pattern is used for database access?"

Broad: "What is the architecture?"
Specific: "What is the dependency structure of the API layer?"
```text

---

#### 1.2 Scope Definition
```text
□ What parts of codebase are in scope?
□ What technologies need investigation?
□ What documentation to review?
□ Are there time constraints?
□ How thorough should research be?
```text

**Thoroughness Levels:**
- **Quick:** Surface-level understanding, key files only
- **Medium:** Reasonable depth, major patterns identified
- **Deep:** Comprehensive understanding, all nuances covered

---

### Phase 2: Exploration Strategy

**Objective:** Plan systematic investigation approach.

#### 2.1 Entry Points Identification
```text
Where to Start Exploration:
□ README and documentation
□ Main entry points (main.js, app.py, main.cpp)
□ Configuration files
□ Directory structure
□ Test files (show usage patterns)
□ Package/dependency files
□ Recent commits (git log)
```text

---

#### 2.2 Exploration Techniques

**Top-Down Exploration:**
```text
1. Start with high-level architecture
2. Identify major components
3. Understand component relationships
4. Drill into specific components
5. Examine implementation details
```text

**Bottom-Up Exploration:**
```text
1. Start with specific code/feature
2. Trace through execution path
3. Identify called functions/classes
4. Map out related components
5. Build up to architecture view
```text

**Pattern Identification:**
```text
1. Find examples of common operations
2. Identify patterns used
3. Note conventions and standards
4. Document variations
5. Understand rationale
```text

---

### Phase 3: Investigation and Discovery

**Objective:** Gather information systematically.

#### 3.1 Codebase Exploration

**Tools and Techniques:**
```text
Use Task tool with Explore agent:
- "How does authentication work?"
- "Where are API endpoints defined?"
- "What is the database schema?"

Use Grep for pattern finding:
- Search for class definitions
- Find function implementations
- Locate configuration usage
- Identify error handling patterns

Use Glob for file discovery:
- Find all controllers: **/controllers/**/*.js
- Find all tests: **/tests/**/*.py
- Find all models: **/models/**/*.cpp

Use Read for detailed study:
- Read key files completely
- Understand implementation details
- Study test files for usage examples
```text

---

#### 3.2 Documentation Review
```text
Review All Available Documentation:
□ README files
□ API documentation
□ Architecture diagrams
□ Design documents
□ Code comments
□ Commit messages
□ Pull request descriptions
□ Issue discussions
```text

---

#### 3.3 Dynamic Investigation
```text
Interactive Exploration:
□ Run the application
□ Exercise features
□ Observe behavior
□ Check logs
□ Use debugger
□ Profile performance
```text

---

#### 3.4 Note Taking
```text
Document As You Go:
□ Key findings
□ Patterns observed
□ Questions that arise
□ Hypotheses formed
□ Areas needing deeper investigation
□ Surprising discoveries
```text

---

### Phase 4: Knowledge Synthesis

**Objective:** Connect findings into coherent understanding.

#### 4.1 Pattern Recognition
```text
Identify Patterns:
□ Architectural patterns (MVC, microservices, etc.)
□ Design patterns (Factory, Strategy, etc.)
□ Code organization patterns
□ Naming conventions
□ Error handling patterns
□ Testing patterns
```text

---

#### 4.2 Mental Model Construction
```text
Build Understanding:
□ How components interact
□ Data flow through system
□ Control flow and sequence
□ State management
□ Error handling strategy
□ Deployment architecture
```text

**Visualization Techniques:**
```text
Create diagrams:
- Component diagrams
- Sequence diagrams
- Data flow diagrams
- Directory structure trees
- Dependency graphs
```text

---

#### 4.3 Insight Generation
```text
Answer Research Questions:
□ Direct answers to questions
□ Supporting evidence
□ Confidence level in findings
□ Areas of uncertainty
□ Recommendations based on findings
```text

---

### Phase 5: Documentation and Recommendations

**Objective:** Document findings clearly for future reference.

#### 5.1 Research Report Structure
```text
## Research Summary

### Objectives
- What was investigated
- Why investigation was needed

### Key Findings
- Main discoveries (bullet points)
- Most important insights

### Detailed Findings
- Comprehensive documentation
- Code examples
- Diagrams
- Evidence

### Recommendations
- Actions based on findings
- Next steps
- Follow-up research needed

### References
- Files examined
- Documentation reviewed
- External resources consulted
```text

---

#### 5.2 Documentation Best Practices
```text
✅ Clear and concise
✅ Includes code examples
✅ Provides file references (file:line)
✅ Explains rationale
✅ Notes limitations/assumptions
✅ Actionable recommendations

❌ Vague or ambiguous
❌ Only high-level description
❌ No supporting evidence
❌ Unclear conclusions
```text

---

## Research Templates

### Architecture Research

```markdown
# Architecture Research: [System Name]

## High-Level Architecture

[Diagram or description]

## Major Components

### Component 1: [Name]
- Location: `path/to/component`
- Responsibility: [What it does]
- Dependencies: [What it depends on]
- Key files:
  - `file1.ext` - [Purpose]
  - `file2.ext` - [Purpose]

### Component 2: [Name]
...

## Data Flow

[How data moves through system]

## Patterns Used

- [Pattern 1]: [Where and why]
- [Pattern 2]: [Where and why]

## Key Insights

- [Insight 1]
- [Insight 2]

## Recommendations

- [Recommendation 1]
- [Recommendation 2]
```text

---

### Feature Research

```markdown
# Feature Research: [Feature Name]

## Feature Overview

What: [What the feature does]
Why: [Business purpose]
Users: [Who uses it]

## Implementation

### Entry Points
- API: `POST /api/endpoint` in `file.ext:42`
- UI: `ComponentName` in `file.ext:100`

### Core Logic
Location: `path/to/logic`

```code
// Key implementation snippet
```text

### Data Models
- Model 1: [Description]
- Model 2: [Description]

### Dependencies
- [Service 1]: [How used]
- [Service 2]: [How used]

## Testing

Test Coverage: X%
Key Tests: `test/feature.test.js`

## Observations

- [Observation 1]
- [Observation 2]

## Recommendations

- [Recommendation 1]
- [Recommendation 2]
```text

---

## Research Anti-Patterns

### Analysis Paralysis
```text
❌ Don't:
- Research forever without conclusions
- Explore every possible path
- Aim for perfect understanding

✅ Do:
- Set time limits
- Answer specific questions
- Acknowledge limitations
- Deliver findings incrementally
```text

### Shallow Investigation
```text
❌ Don't:
- Only read file names
- Rely on assumptions
- Skip key components
- Ignore edge cases

✅ Do:
- Read actual code
- Verify understanding
- Cover major components
- Note important details
```text

### Poor Documentation
```text
❌ Don't:
- Vague summaries
- No code references
- Conclusions without evidence
- Unclear recommendations

✅ Do:
- Specific findings
- File and line references
- Evidence-backed conclusions
- Actionable recommendations
```text

---

## Example Research Sessions

### Example 1: Authentication System Research

```markdown
# Research: Authentication System

## Objective
Understand how user authentication works to add OAuth support

## Key Findings

### Current Implementation
- **Strategy**: JWT tokens (stateless)
- **Location**: `src/auth/authService.js`
- **Flow**:
  1. User submits credentials → `POST /api/auth/login`
  2. Credentials validated → `authService.validateCredentials()`
  3. JWT token generated → `tokenService.generateToken()`
  4. Token returned to client
  5. Client includes token in `Authorization` header
  6. Server validates token → `authMiddleware.js:15`

### Token Structure
```javascript
{
  userId: "123",
  email: "user@example.com",
  role: "user",
  exp: 1234567890  // expires after 24h
}
```text

### Patterns Observed
- **Middleware pattern**: `authMiddleware` protects routes
- **Service layer**: Business logic in `authService`
- **Repository pattern**: Data access in `userRepository`

### Dependencies
- `jsonwebtoken` library for JWT operations
- `bcrypt` for password hashing
- `express` middleware integration

## Recommendations

### For OAuth Integration
1. Extract auth strategy into interface
2. Implement OAuth strategy alongside JWT
3. Use Strategy pattern for auth method selection
4. Maintain backward compatibility with JWT

### Code Locations to Modify
- `src/auth/authService.js` - Add OAuth flow
- `src/auth/authMiddleware.js` - Support both auth types
- `src/config/oauth.js` - New OAuth configuration

## Next Steps
1. Design OAuth integration architecture
2. Select OAuth library (Passport.js recommended)
3. Plan migration strategy for existing users
```text

---

## Research Completion Criteria

Research is complete when:
```text
✓ All research questions answered
✓ Key findings documented
✓ Evidence provided for conclusions
✓ Mental model constructed
✓ Recommendations provided
✓ Next steps identified
✓ Stakeholders informed
```text

---

## When Research is Complete vs. Needs More Depth

### Research Complete When:
```text
✓ Questions answered sufficiently
✓ Enough information to proceed
✓ Recommendations clear
✓ Confidence level acceptable
✓ Time box reached
```text

### Needs More Research When:
```text
❌ Critical gaps in understanding
❌ Contradictory findings
❌ High uncertainty
❌ Cannot make recommendations
❌ Missing key components
```text

---

## References

- [Standard Workflow](standard.md)
- [Agents](../agents.md) - Choosing an investigation agent (spelunker, inspector)

---

**Last reviewed:** 2026-01-07
**Next review:** Quarterly or when research practices evolve
