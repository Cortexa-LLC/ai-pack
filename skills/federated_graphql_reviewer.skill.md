# Federated GraphQL Schema Reviewer
<!-- skills/federated_graphql_reviewer.skill.md -->

**Version:** 1.0
**InjectAt:** role_context
**Slot:** 36
**Tools:** kg__search_knowledge, mcp__upk__search_knowledge, Bash, Read, Grep, Glob
**MaxExtraTokens:** 8000
**Optional:** true

---

## Federated GraphQL Schema Review

Review and validate GraphQL schemas for Apollo Federation compliance, performance, best practices, and architectural soundness. Ensures quality before deployment using industry-standard patterns.

**Use when:** Reviewing new schemas, validating schema changes, conducting pre-deployment reviews, or auditing existing federated schemas.

**Reference Documentation:** See `docs/schema-review-best-practices.md` for complete generic best practices.

---

## Core Review Principles

1. **Federation Compliance** — Verify correct use of federation directives and patterns
2. **Performance Impact** — Assess query planning, N+1 risks, and resolver efficiency
3. **Domain Modeling** — Ensure schema reflects business concepts, not database structure
4. **Client Experience** — Validate schema serves client needs without over-fetching
5. **Documentation Standards** — Check naming conventions, descriptions, error handling
6. **Resilience** — Verify partial failure handling and error boundaries

## Review SLA Expectations

- **Initial feedback**: Within 24 hours of submission (when possible)
- **Approval timing**: Depends on compliance - schemas violating standards require fixes
- **Iteration cycles**: Address blocker issues before approval

---

## Phase 1: Initial Assessment

### 1.1 Understand Context

Ask the user:

```
1. What is the purpose of this schema?
   - New subgraph? Extension of existing entity?
   - Which domain/team owns it?

2. What stage is this review?
   - Pre-implementation design review?
   - Pre-deployment validation?
   - Post-deployment audit?

3. Are there specific concerns?
   - Performance issues?
   - Federation composition errors?
   - Client complaints?

4. Do you have the schema file path?
   - Or should I search for *.graphql files?

5. Are there any organization-specific review requirements?
   - Sign-off processes, compliance checklists, etc.
```

### 1.2 Locate Schema Files

```bash
# Find GraphQL schemas
find . -name "*.graphql" -o -name "*.gql" -o -name "schema.graphqls"

# Look for federation config
find . -name "federation.yaml" -o -name "supergraph.yaml"
```

### 1.3 Query Knowledge Base (Optional)

**If knowledge base available:**
```javascript
// Search for established patterns
kg__search_knowledge({ query: "GraphQL federation review checklist" })
mcp__upk__search_knowledge({ query: "GraphQL schema anti-patterns" })
```

---

## Phase 2: Federation Compliance Review

### 2.1 Federation Version Check

```graphql
# ✅ GOOD: Explicit federation version
schema
  @link(url: "https://specs.apollo.dev/federation/v2.3", import: ["@key", "@shareable"])
{
  query: Query
}

# ❌ BAD: No federation version declared
schema {
  query: Query
}
```

**Check:**
- [ ] Federation version explicitly declared (v2.0+)
- [ ] Required imports match directives used
- [ ] No deprecated v1 patterns (`@extends` without `@link`)

### 2.2 Entity Key Validation

**Critical Checks:**

```graphql
# ✅ GOOD: Stable identifier as key
type Astronaut @key(fields: "id") {
  id: ID!
  name: String!
}

# ❌ BAD: Business key that might change
type Astronaut @key(fields: "name") {
  id: ID!
  name: String!
}

# ✅ GOOD: Composite key when necessary
type LaunchManifest @key(fields: "launchId equipmentId") {
  launchId: ID!
  equipmentId: ID!
}

# ❌ BAD: Missing @key on entity
type Equipment {  # Should have @key!
  id: ID!
  name: String!
}
```

**Validate:**
- [ ] All entities have `@key` directive
- [ ] Key fields use stable identifiers (won't change)
- [ ] Composite keys only when single field insufficient
- [ ] Key fields are non-nullable
- [ ] `@external` used correctly on extended entity keys
- [ ] **No unnecessary @external on @key fields** (Federation v1 relic - @key fields are always provided)

### 2.3 Entity Extension Patterns

**Owner vs Extender:**

```graphql
# ✅ GOOD: Proper extension with @external
type Astronaut @key(fields: "id") {
  id: ID! @external
  preferences: AstronautPreferences  # New field
}

# ❌ BAD: Missing @external on key field
type Astronaut @key(fields: "id") {
  id: ID!  # Should be @external!
  preferences: AstronautPreferences
}

# ✅ GOOD: Reference-only stub
type Equipment @key(fields: "id", resolvable: false) {
  id: ID!
}

# ❌ BAD: Extending without marking key @external
type Equipment @key(fields: "id") {
  id: ID!  # Missing @external
  reviews: [Review!]!
}
```

**Check:**
- [ ] Extended entities mark key fields as `@external`
- [ ] Reference-only stubs use `resolvable: false`
- [ ] No duplicate field definitions across subgraphs (unless `@shareable`)

### 2.4 Federation Directives Usage

**@requires:**
```graphql
# ✅ GOOD: Requires field from owning subgraph
type Astronaut @key(fields: "id") {
  id: ID! @external
  rank: String! @external
  bonus: Float! @requires(fields: "rank")
}

# ❌ BAD: Requires own field
type Astronaut @key(fields: "id") {
  id: ID!
  name: String!
  greeting: String! @requires(fields: "name")  # Just use name directly!
}
```

**@provides:**
```graphql
# ✅ GOOD: Optimizes foreign field fetch
type Launch @key(fields: "id") {
  id: ID!
  astronaut: Astronaut! @provides(fields: "name rank")
}

# ⚠️ WARNING: Over-providing can break boundaries
# Only provide fields your resolver actually has
```

**@shareable:**
```graphql
# ✅ GOOD: Value type shared across subgraphs
type Coordinate @shareable {
  latitude: Float!
  longitude: String!
}

# ❌ BAD: Entity shared without coordination
type Astronaut @shareable {  # Dangerous! Multiple resolvers for entity
  id: ID!
  name: String!
}
```

**Validate:**
- [ ] `@requires` only references `@external` fields
- [ ] `@provides` used judiciously (performance optimization)
- [ ] `@shareable` only on value types, not entities with state
- [ ] No circular `@requires` dependencies

---

## Phase 3: Performance Review

### 3.1 N+1 Query Detection

**Anti-pattern:**
```graphql
type Query {
  orders: [Launch!]!
}

type Launch {
  id: ID!
  astronaut: Astronaut!  # ⚠️ N+1 risk if not using DataLoader
}

type Astronaut @key(fields: "id") {
  id: ID!
  name: String!
}
```

**Questions to ask:**
- [ ] Are list fields resolved with batching (DataLoader)?
- [ ] Do resolvers support batch loading via `__resolveReference`?
- [ ] Are there deeply nested list-of-list patterns?

### 3.2 Query Complexity

```graphql
# ⚠️ WARNING: Unbounded lists
type Astronaut {
  orders: [Launch!]!  # How many orders? Could be thousands!
}

# ✅ BETTER: Pagination required
type Astronaut {
  orders(first: Int!, after: String): LaunchConnection!
}

# ✅ GOOD: Relay-style pagination
type LaunchConnection {
  edges: [LaunchEdge!]!
  pageInfo: PageInfo!
}
```

**Check:**
- [ ] List fields have pagination or limits
- [ ] Query depth limits considered
- [ ] Field-level cost analysis for expensive operations
- [ ] Router query planning overhead expected <10ms (Netflix standard)

### 3.3 Resolver Efficiency

**Ask about implementation:**
- [ ] Are database queries optimized (indexes, select specific fields)?
- [ ] Is caching implemented for frequently accessed data?
- [ ] Are downstream service calls batched?
- [ ] Is parallel execution possible for independent fields?

---

## Phase 4: Schema Design Quality

### 4.1 Domain Modeling

**Netflix Principle:** Schema should reflect business concepts, not database structure.

```graphql
# ❌ BAD: Exposing database structure
type Astronaut {
  astronaut_id: Int!  # Database column name
  astronaut_data_json: String!  # Raw DB field
  created_ts: Int!  # Timestamp as int
}

# ✅ GOOD: Business domain model
type Astronaut {
  id: ID!
  profile: AstronautSpacecraft!
  createdAt: DateTime!
}
```

**Check:**
- [ ] Field names reflect domain language, not database columns
- [ ] No raw JSON strings or opaque data blobs
- [ ] Proper type usage (DateTime, not Int for timestamps)
- [ ] Relationships modeled as GraphQL types, not foreign keys

### 4.2 Naming Conventions

```graphql
# ✅ GOOD: Consistent naming
type Query {
  astronaut(id: ID!): Astronaut
  products(filter: EquipmentFilter): [Equipment!]!
}

type Mutation {
  updateAstronaut(input: UpdateAstronautInput!): UpdateAstronautPayload!
}

# ❌ BAD: Inconsistent naming
type Query {
  getAstronaut(astronautId: ID!): Astronaut  # Don't prefix with "get"
  fetchEquipments: [Equipment!]!  # Inconsistent verb
}

type Mutation {
  astronaut_update(data: AstronautData): Astronaut  # Snake case, generic names
}
```

**Validate:**
- [ ] camelCase for fields and arguments
- [ ] PascalCase for types
- [ ] No "get" prefixes on queries
- [ ] Unique Input/Output types per operation (no reuse)
- [ ] Meaningful, domain-specific names

### 4.3 Documentation Quality

```graphql
# ✅ GOOD: Complete documentation
"""
Astronaut in the mission program.
Represents crew members assigned to space missions.
"""
type Astronaut @key(fields: "id") {
  """
  Unique astronaut identifier.
  Stable across system lifetime, never changes.
  """
  id: ID!
  
  """
  Astronaut's mission preferences and specializations.
  Null when: astronaut has not yet configured mission preferences.
  Updated when: astronaut updates their specialization profile.
  """
  preferences: MissionPreferences
}

# ❌ BAD: No documentation or vague
type Astronaut @key(fields: "id") {
  id: ID!
  preferences: MissionPreferences  # What does this contain?
}
```

**Check:**
- [ ] All types have descriptions
- [ ] All fields documented with purpose
- [ ] Nullable fields include "Null when:" semantics
- [ ] Complex types explain relationships
- [ ] Deprecation reasons documented with `@deprecated`
- [ ] **Shared types have consistent descriptions** across all subgraphs (or use `{inherit}`)

---

## Phase 5: Error Handling & Resilience

### 5.1 Error Patterns

**Netflix Pattern:** Return partial results with clear errors.

```graphql
# ✅ GOOD: Union error pattern
type Query {
  astronaut(id: ID!): AstronautResult!
}

union AstronautResult = Astronaut | AstronautNotFound | UnauthorizedError

type AstronautNotFound {
  message: String!
  requestedId: ID!
}

# ✅ ALSO GOOD: Nullable with errors array
type Query {
  astronaut(id: ID!): Astronaut
}
# (Errors returned in GraphQL errors array)

# ❌ BAD: Throwing exceptions for business logic
# (Forces client to parse error messages)
```

**Validate:**
- [ ] Business errors handled explicitly (union types or nullable fields)
- [ ] Technical errors allowed to propagate to GraphQL errors array
- [ ] Error messages are actionable for clients
- [ ] No sensitive information in error messages

### 5.2 Authorization Boundaries

```graphql
# ✅ GOOD: Clear authorization model
type Astronaut @key(fields: "id") {
  id: ID!
  name: String!  # Public
  medicalRecords: [MedicalRecord!]  # Null when: viewer lacks medical scope
}

# ⚠️ REVIEW: Field-level authorization complexity
type Astronaut {
  orders: [Launch!]!  # Filtered by viewer permissions? Or should throw?
}
```

**Check:**
- [ ] Authorization model documented (field vs object level)
- [ ] Null semantics clear for unauthorized access
- [ ] Sensitive fields appropriately restricted

---

## Phase 6: Governance & Standards

### 6.1 Versioning

```graphql
# ❌ BAD: Version suffixes
type Astronaut {
  profileV1: AstronautSpacecraftV1
  profileV2: AstronautSpacecraftV2
}

# ✅ GOOD: Schema evolution with @deprecated
type Astronaut {
  profile: AstronautSpacecraft!
  legacySpacecraft: LegacyAstronautSpacecraft @deprecated(reason: "Use 'profile' field. Removed after 2026-06-01")
}
```

**Check:**
- [ ] No version suffixes (V1, V2, etc.)
- [ ] Deprecated fields include removal timeline
- [ ] Breaking changes documented in migration guide

### 6.2 Schema Validation

**Run composition check:**
```bash
# Apollo Rover
rover subgraph check <GRAPH_REF> --schema ./schema.graphql

# Or federation CLI
npx @apollo/federation compose --config supergraph.yaml
```

**Validate:**
- [ ] Schema composes successfully with other subgraphs
- [ ] No composition errors or warnings
- [ ] Breaking change detection run
- [ ] Lint rules pass (if automated linting configured)
- [ ] **Nullability consistency**: shared fields have consistent nullability (ID vs ID!) across subgraphs
- [ ] **Enum additions**: new enum values documented, clients should handle gracefully
- [ ] **Input field additions**: optional input fields noted as potentially breaking for strict validation

---

## Phase 6.5: Common Production Issues

### Real-World Patterns from Schema Reviews

**Reference:** See `docs/schema-review-best-practices.md` for complete details on each issue.

These issues appear frequently in GraphQL federation schemas:

**1. Inconsistent Descriptions Across Subgraphs** (Most Common)

```graphql
# ❌ PROBLEM: Same type, different descriptions in subgraphs
# Subgraph A:
"""Pagination parameters"""
type PaginationInput { ... }

# Subgraph B:
"""Input for pagination"""
type PaginationInput { ... }

# ✅ SOLUTION 1: Use {inherit} directive
"""
{inherit}
"""
type PaginationInput { ... }

# ✅ SOLUTION 2: Match supergraph description exactly
"""
Pagination parameters for query results
"""
type PaginationInput { ... }
```

**Impact:** Composition warnings, confusing documentation for clients.

**2. Unnecessary @external on @key Fields** (Federation v1 Relic)

```graphql
# ❌ BAD: @external on @key field in extension
type CommanderSpacecraft @key(fields: "astronaut { id }") {
  astronaut: Astronaut! @external  # Unnecessary - @key fields always provided
  metrics: CommanderMetrics
}

# ✅ GOOD: Remove @external from @key fields
type CommanderSpacecraft @key(fields: "astronaut { id }") {
  astronaut: Astronaut!  # No @external needed
  metrics: CommanderMetrics
}
```

**Impact:** Legacy syntax, ignored by Fed v2 but clutters schema.

**3. Breaking Enum Value Additions**

```graphql
# ⚠️ WARNING: Adding enum values can break clients
enum MissionStatus {
  ACTIVE
  INACTIVE
  PENDING  # New value added
}

# Clients with switch statements may not handle PENDING:
# switch(status) {
#   case ACTIVE: ...
#   case INACTIVE: ...
#   // No default case - PENDING causes failure!
# }
```

**Impact:** Clients not programming defensively may break.

**Mitigation:** Document enum changes, advise clients to use default cases.

**4. Nullability Inconsistency Across Subgraphs**

```graphql
# ❌ PROBLEM: Different nullability across subgraphs
# Subgraph A:
type Node {
  id: ID!  # Non-nullable
}

# Subgraph B:
type Node {
  id: ID  # Nullable
}

# Supergraph uses looser type: ID (nullable)
```

**Impact:** Supergraph uses nullable type (weakest constraint), may surprise clients expecting non-null.

**5. Shared Value Types Needing @shareable**

```graphql
# ❌ PROBLEM: Common types defined in multiple subgraphs
# Without @shareable, composition fails

# Subgraph A:
type Coordinate {
  latitude: Float!
  longitude: String!
}

# Subgraph B:
type Coordinate {  # Duplicate definition!
  latitude: Float!
  longitude: String!
}

# ✅ SOLUTION: Mark as @shareable
type Coordinate @shareable {
  latitude: Float!
  longitude: String!
}
```

**6. Input Field Additions**

```graphql
# ⚠️ WARNING: Adding optional fields to input types
input CreateAstronautInput {
  name: String!
  rank: String!
  specialty: String  # Newly added optional field
}
```

**Impact:** Clients with strict input validation may reject unknown fields.

**7. Boolean Field Documentation**

```graphql
# ❌ BAD: Boolean without explanation
type Mission {
  isActive: Boolean!  # What does true/false mean?
}

# ✅ GOOD: Clear documentation
type Mission {
  """
  Mission active status.
  true: Mission is currently active and accepting crew assignments.
  false: Mission is completed, cancelled, or not yet started.
  """
  isActive: Boolean!
}
```

**Impact:** Validation warning - must specify what true/false indicates.

**8. String Field Format/Length**

```graphql
# ❌ BAD: Missing format/length constraints
type Astronaut {
  name: String!  # How long? What characters allowed?
  bio: String
}

# ✅ GOOD: Document constraints
type Astronaut {
  """
  Astronaut full name.
  Max 100 characters, alphanumeric with spaces and hyphens.
  """
  name: String!
  
  """
  Astronaut biography.
  Max 2000 characters, supports markdown formatting.
  """
  bio: String
}
```

**Impact:** Validation warning - provide maximum length and format specification.

**9. Unused Enum Types**

```graphql
# ⚠️ WARNING: Enum defined but never used
enum MissionPriority {
  HIGH
  MEDIUM
  LOW
}
# No field references this enum!
```

**Impact:** DEBUG warning - enum included in supergraph despite being unused.

**10. Interface Without Implementations**

```graphql
# ❌ PROBLEM: Interface used as output without implementations
interface Vehicle {
  id: ID!
  capacity: Int!
}

type Query {
  vehicles: [Vehicle!]!  # No concrete types implement Vehicle!
}

# ✅ SOLUTION: Provide implementations
type Spacecraft implements Vehicle {
  id: ID!
  capacity: Int!
  fuelType: String!
}

type Rover implements Vehicle {
  id: ID!
  capacity: Int!
  terrain: String!
}
```

**Impact:** Validation warning - interface needs at least one concrete implementation.

**11. Potential Duplicate Types**

```graphql
# ⚠️ WARNING: Similar types detected
type MissionNotFound {
  errorCode: String!
  errorMessage: String!
}

type AstronautNotFound {
  errorCode: String!
  errorMessage: String!
}

# ✅ BETTER: Shared error type
type NotFoundError {
  errorCode: String!
  errorMessage: String!
  resourceType: String!  # "Mission", "Astronaut", etc.
}
```

**Impact:** Validation detects structural duplication, suggests consolidation.

**12. Union Type Member Additions**

```graphql
# ⚠️ WARNING: Adding new type to union
union SearchResult = Mission | Astronaut | Launch  # Adding Launch

# Clients may not handle new type:
# switch(result.__typename) {
#   case 'Mission': ...
#   case 'Astronaut': ...
#   // No case for 'Launch'!
# }
```

**Impact:** May break clients without default/exhaustive type handling.

---

## Phase 7: Generate Review Report

### 7.1 Structure Report

```markdown
# GraphQL Schema Review: [Subgraph Name]
**Reviewer:** [Your Name]  
**Date:** [YYYY-MM-DD]  
**Schema Version:** [Version/Commit]  

## Summary
[Overall assessment - APPROVED / APPROVED WITH CHANGES / NEEDS REVISION]

## Federation Compliance
- [ ] ✅ Federation v2.3+ declared
- [ ] ✅ Entity keys use stable identifiers
- [ ] ⚠️ WARNING: Missing @external on Astronaut.id extension
- [ ] ❌ BLOCKER: Equipment entity missing @key directive

## Performance
- [ ] ✅ Pagination implemented on list fields
- [ ] ⚠️ CONCERN: Astronaut.missions could cause N+1 without DataLoader

## Schema Quality
- [ ] ✅ Domain modeling follows business concepts
- [ ] ✅ Naming conventions consistent
- [ ] ⚠️ IMPROVEMENT: Add "Null when:" docs to nullable fields

## Error Handling
- [ ] ✅ Union error pattern used for mutations
- [ ] ✅ Business vs technical errors separated

## Required Changes (BLOCKER)
1. Add @key directive to Equipment entity
2. Fix Astronaut.id missing @external in reviews subgraph

## Recommended Improvements
1. Add DataLoader batching for Astronaut.missions resolver
2. Document null semantics on 12 nullable fields
3. Add deprecation timeline to legacyAddress field

## Approved With Conditions
Schema approved pending BLOCKER fixes above.
Re-review not required for RECOMMENDED improvements.
```

### 7.2 Provide Actionable Feedback

For each issue, include:
- **Severity:** BLOCKER / WARNING / IMPROVEMENT
- **Location:** Type.field or line number
- **Issue:** What's wrong
- **Why:** Impact on federation, performance, or clients
- **Fix:** Specific code example showing correction

---

## Review Checklist Summary

**Federation Compliance:**
- [ ] Federation version declared (v2.0+)
- [ ] All entities have @key with stable identifiers
- [ ] Extensions properly use @external
- [ ] Directives (@requires, @provides, @shareable) used correctly
- [ ] No circular dependencies

**Performance:**
- [ ] List fields paginated or limited
- [ ] N+1 queries prevented (DataLoader pattern)
- [ ] Query complexity bounded
- [ ] Resolver efficiency considered

**Schema Quality:**
- [ ] Domain modeling (business concepts, not DB structure)
- [ ] Naming conventions consistent
- [ ] Comprehensive documentation
- [ ] Null semantics documented

**Error Handling:**
- [ ] Business errors handled explicitly
- [ ] Partial failure support
- [ ] Authorization boundaries clear

**Governance:**
- [ ] No version suffixes
- [ ] Schema composes successfully
- [ ] Breaking changes documented
- [ ] Deprecations have timelines

---

## Integration with Other Skills

- **Complements `federated_graphql_designer`** — Reviews designs from designer
- **Optionally uses knowledge base** — Checks against established patterns
- **Used by roles:** reviewer, architect, schema champion

---

## Example Usage

```
Astronaut: "Review this GraphQL schema before deployment"
Context: Provides schema file path

Agent:
1. Reads schema file
2. Validates federation compliance (version, @key, @external)
3. Checks for performance risks (N+1, unbounded lists)
4. Reviews domain modeling and naming
5. Validates error handling patterns
6. Tests schema composition
7. Generates structured review report with:
   - BLOCKER issues (must fix)
   - WARNING issues (should fix)
   - IMPROVEMENT suggestions (nice to have)
8. Provides specific code examples for fixes
```

---

## Notes

- **Be thorough but pragmatic:** Balance perfection with development velocity
- **Explain the "why":** Help teams learn, don't just flag issues
- **Distinguish severity:** Not every issue blocks deployment
- **Offer solutions:** Show the correct pattern, don't just point out problems
- **Consider context:** Early-stage prototypes need less rigor than deployment-ready schemas
- **Build knowledge:** Record common issues in KG/upk for future reviews
- **Reference best practices:** Point teams to `docs/schema-review-best-practices.md` for generic patterns
- **Adapt to organization:** If the organization has specific review processes (sign-offs, compliance checklists), incorporate them into the review workflow
