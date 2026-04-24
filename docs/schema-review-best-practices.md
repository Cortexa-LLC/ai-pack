# Generic Federated GraphQL Schema Review Best Practices

> Universal patterns for reviewing Apollo Federation schemas before deployment

## Overview

This document provides a comprehensive checklist for reviewing federated GraphQL schemas based on industry best practices. Use these guidelines to ensure schemas are deployment-ready, performant, and maintainable.

## Review Process Framework

### Review Prerequisites

Before beginning a formal schema review:

1. **Schema files must be accessible** - Provide path to `.graphql` or `.gql` files
2. **Context must be clear** - Is this a new subgraph, entity extension, or schema update?
3. **Change scope identified** - Which types/fields are new vs modified?
4. **Federation version declared** - Must be Federation v2.0+

### Review SLA Expectations

- **Initial feedback**: Within 24 hours of submission
- **Approval timing**: Depends on compliance - no SLA for schemas violating standards
- **Iteration cycles**: Teams must address blocker issues before approval

## Core Schema Review Principles

1. **Federation Compliance** — Verify correct use of federation directives and patterns
2. **Performance Impact** — Assess query planning, N+1 risks, and resolver efficiency
3. **Domain Modeling** — Ensure schema reflects business concepts, not database structure
4. **Client Experience** — Validate schema serves client needs without over-fetching
5. **Documentation Standards** — Check naming conventions, descriptions, error handling
6. **Resilience** — Verify partial failure handling and error boundaries

---

## Schema Design Best Practices

### 1. Field Naming & Suffixes

**Input/Output Fields:**
```graphql
# ✅ GOOD: Proper suffixes
input CreateMissionInput {
  name: String!
  targetDate: DateTime!
}

type CreateMissionOutput {
  mission: Mission
  errors: [MissionError!]
}

# ❌ BAD: Generic names without suffixes
input MissionData {
  name: String!
}

type MissionResult {
  mission: Mission
}
```

**Mutations:**
```graphql
# ✅ GOOD: Use "create" for new entities
mutation {
  createMission(input: CreateMissionInput!): CreateMissionOutput
}

# ❌ BAD: Using "add" or generic verbs
mutation {
  addMission(data: MissionData): Mission
}
```

### 2. Nullable Fields - Avoid When Possible

**Problem:** Nullable fields are a common source of errors, maintenance burden, and bloat.

```graphql
# ❌ BAD: Excessive nullability
type Mission {
  id: ID!
  name: String  # When is this null?
  description: String  # Why nullable?
  startDate: DateTime  # Unclear null semantics
  endDate: DateTime
  status: MissionStatus  # Should be required!
}

# ✅ BETTER: Minimize nullability, use interfaces
type Mission {
  id: ID!
  name: String!
  description: String!  # Required, empty string if none
  startDate: DateTime!
  endDate: DateTime  # Null when: mission hasn't ended yet
  status: MissionStatus!
}
```

**When nullable fields are appropriate:**
- Optional relationships (one-to-one that may not exist)
- Future-dated fields (endDate on active mission)
- Scope-restricted fields (null when user lacks permission)

**Always document null semantics:**
```graphql
"""
Mission end date and time.
Null when: mission is still active and has not completed.
Updated when: mission completes successfully or is terminated.
"""
endDate: DateTime
```

### 3. Use Interfaces for Flexible Design

Instead of many nullable fields, use interfaces to model variations:

```graphql
# ✅ GOOD: Interface for type flexibility
interface Mission {
  id: ID!
  name: String!
  status: MissionStatus!
}

type RoboticMission implements Mission {
  id: ID!
  name: String!
  status: MissionStatus!
  roboticEquipment: [Equipment!]!
}

type CrewedMission implements Mission {
  id: ID!
  name: String!
  status: MissionStatus!
  crew: [Astronaut!]!
  lifeSupportSystems: [LifeSupport!]!
}
```

### 4. Entity Extension Patterns

**Extending entities owned by other subgraphs:**

```graphql
# ✅ GOOD: Proper extension
type Astronaut @key(fields: "id") {
  id: ID! @external
  missionHistory: [Mission!]!  # New field from this subgraph
}

# ❌ BAD: Missing @external
type Astronaut @key(fields: "id") {
  id: ID!  # Should be @external!
  missionHistory: [Mission!]!
}
```

**Reference-only stubs:**
```graphql
# ✅ GOOD: Non-resolvable reference
type Equipment @key(fields: "id", resolvable: false) {
  id: ID!
}

# Used in relationships without implementing resolver
type Mission {
  requiredEquipment: [Equipment!]!
}
```

### 5. Consistent Descriptions for Shared Types

**Problem:** Same type defined in multiple subgraphs with different descriptions causes composition warnings.

```graphql
# ❌ BAD: Custom descriptions differ across subgraphs
# Subgraph A:
"""Pagination parameters"""
type PaginationInput { ... }

# Subgraph B:
"""Input for pagination"""
type PaginationInput { ... }

# ✅ SOLUTION: Use {inherit} directive
"""
{inherit}
"""
type PaginationInput {
  maxPageSize: Int
  pageCursor: String
}
```

### 6. Proper @shareable for Value Types

```graphql
# ✅ GOOD: Value types marked @shareable
type Coordinate @shareable {
  latitude: Float!
  longitude: Float!
}

# ❌ BAD: Duplicate type without @shareable
type Coordinate {  # Composition error if in multiple subgraphs
  latitude: Float!
  longitude: Float!
}
```

### 7. No @external on @key Fields (Fed v2)

**Federation v1 relic - remove in v2:**

```graphql
# ❌ BAD: Unnecessary @external (Fed v1 syntax)
type Spacecraft @key(fields: "astronaut { id }") {
  astronaut: Astronaut! @external  # Not needed in Fed v2
  settings: SpacecraftSettings
}

# ✅ GOOD: Clean Fed v2 syntax
type Spacecraft @key(fields: "astronaut { id }") {
  astronaut: Astronaut!  # No @external needed
  settings: SpacecraftSettings
}
```

### 8. Consistent Nullability Across Subgraphs

```graphql
# ❌ BAD: Different nullability
# Subgraph A:
type Node {
  id: ID!  # Non-nullable
}

# Subgraph B:
type Node {
  id: ID  # Nullable - supergraph uses this (weakest)
}

# ✅ GOOD: Consistent across all subgraphs
type Node {
  id: ID!  # Always non-null everywhere
}
```

---

## Documentation Standards

### 1. Boolean Fields - Always Document True/False Meaning

```graphql
# ❌ BAD: Boolean without explanation
type Mission {
  isActive: Boolean!
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

### 2. String Fields - Provide Max Length and Format

```graphql
# ❌ BAD: Missing constraints
type Astronaut {
  name: String!
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

### 3. Enum Fields - Document Breaking Changes

```graphql
# ⚠️ WARNING: Adding enum values can break clients
enum MissionStatus {
  ACTIVE
  INACTIVE
  PENDING  # New value added - may break clients without default case
}

# Document enum evolution
"""
Mission status enumeration.
New values may be added - clients should handle unknown values gracefully.
"""
enum MissionStatus {
  ACTIVE
  INACTIVE
  PENDING
}
```

### 4. Remove Unused Enums

```graphql
# ❌ BAD: Enum defined but never used
enum MissionPriority {
  HIGH
  MEDIUM
  LOW
}
# No field references this enum - remove it!

# ✅ GOOD: Only define enums that are used
enum MissionStatus {
  ACTIVE
  INACTIVE
}

type Mission {
  status: MissionStatus!  # Enum is actually used
}
```

### 5. Interfaces Need Implementations

```graphql
# ❌ BAD: Interface without implementations
interface Vehicle {
  id: ID!
  capacity: Int!
}

type Query {
  vehicles: [Vehicle!]!  # No concrete types!
}

# ✅ GOOD: At least one implementation
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

### 6. Avoid Duplicate Types

```graphql
# ❌ BAD: Similar error types
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

---

## Schema Evolution & Migration

### 1. Standard Deprecation Process

```graphql
# Step 1: Add replacement field
type Astronaut {
  profile: AstronautProfile!
  legacyProfile: LegacyProfile @deprecated(
    reason: "Use 'profile' field. Removed after 2026-06-01"
  )
}

# Step 2: Wait until last client stops using deprecated field

# Step 3: Delete field from subgraph & publish to supergraph
```

### 2. Backward Compatibility During Schema/Implementation Gaps

Use internal `@inaccessible` stubs for interim support when schema and implementation deploy independently:

```graphql
# During migration period
type _MigrationStub @inaccessible {
  # Temporary support for old field during rollout
  oldField: String
}

# Remove stub once migration complete
```

### 3. Naming Consistency During Migrations

Keep names consistent during migrations to minimize client churn:

```graphql
# ✅ GOOD: Keep consistent naming
type Offer {
  id: ID!
  price: Money!
}

# If you must evolve, use @deprecated
type Offer {
  id: ID!
  price: Money!
  legacyPrice: Float @deprecated(reason: "Use 'price' field")
}

# ❌ BAD: Renaming without deprecation period
type Negotiation {  # Breaking change for clients!
  id: ID!
  price: Money!
}
```

---

## Performance Considerations

### 1. Composition Complexity

**Reduce composition overhead:**
- Single **canonical owner** per entity
- Minimize `@key` variants
- Flatten deep `@requires`/`@provides` chains
- Remove **circular references**
- Avoid redundant type extensions

**Default composition timeout:** 120s (can increase to 180-300s if needed)

### 2. Selective Gateway Exposure

```graphql
# Expose in both gateways
type ShippingPreference @tag(name: "public") @tag(name: "publicapi") {
  # Hide specific field in one gateway
  internalField: String @tag_directive(gateway: "publicapi", add: "inaccessible")
}
```

### 3. N+1 Query Prevention

Use DataLoader or batching in all subgraph resolvers:

```typescript
// ✅ GOOD: Batched resolution
const userLoader = new DataLoader(async (userIds: string[]) => {
  const users = await db.users.findByIds(userIds);
  return userIds.map(id => users.find(u => u.id === id));
});

const resolvers = {
  User: {
    __resolveReference(ref: { id: string }) {
      return userLoader.load(ref.id);
    }
  }
};
```

---

## Common Production Issues

### Issue 1: Inconsistent Descriptions (Most Common)
**Impact:** Composition warnings, confusing client documentation  
**Fix:** Use `{inherit}` or match supergraph description exactly

### Issue 2: Unnecessary @external on @key Fields
**Impact:** Legacy syntax, clutters schema  
**Fix:** Remove @external from @key fields in Fed v2

### Issue 3: Breaking Enum Value Additions
**Impact:** Clients without default cases may break  
**Fix:** Document enum changes, advise defensive programming

### Issue 4: Nullability Inconsistency
**Impact:** Supergraph uses nullable (weakest constraint)  
**Fix:** Ensure consistent nullability across all subgraphs

### Issue 5: Missing @shareable on Value Types
**Impact:** Composition fails  
**Fix:** Mark shared value types with @shareable

### Issue 6: Input Field Additions
**Impact:** Clients with strict validation may reject  
**Fix:** Document as potentially breaking, add as optional

### Issue 7: Boolean Fields Without Documentation
**Impact:** Validation warning - unclear true/false meaning  
**Fix:** Document what true and false indicate

### Issue 8: String Fields Missing Constraints
**Impact:** Validation warning - no max length or format  
**Fix:** Provide maximum length and format specification

### Issue 9: Unused Enum Types
**Impact:** DEBUG warning - enum included despite being unused  
**Fix:** Remove enum or reference it in a field

### Issue 10: Interface Without Implementations
**Impact:** Validation warning  
**Fix:** Provide at least one concrete implementation

### Issue 11: Potential Duplicate Types
**Impact:** Validation detects structural duplication  
**Fix:** Consolidate similar types or add discriminators

### Issue 12: Union Type Member Additions
**Impact:** May break clients without default type handling  
**Fix:** Document as potentially breaking, test client handling

---

## Review Checklist

### Federation Compliance
- [ ] Federation version declared (v2.0+)
- [ ] All entities have @key with stable identifiers
- [ ] Extensions properly use @external (except @key fields)
- [ ] No unnecessary @external on @key fields
- [ ] @shareable used correctly on value types
- [ ] Directives (@requires, @provides, @shareable) used correctly
- [ ] No circular dependencies

### Schema Quality
- [ ] Nullable fields minimized and documented
- [ ] Interfaces considered for type variations
- [ ] Consistent descriptions across subgraphs (use {inherit})
- [ ] Naming conventions followed (camelCase, PascalCase)
- [ ] No duplicate types detected
- [ ] All enums are used in schema
- [ ] All interfaces have implementations

### Documentation
- [ ] All types have descriptions
- [ ] Nullable fields include "Null when:" semantics
- [ ] Boolean fields document true/false meaning
- [ ] String fields specify max length and format
- [ ] Enum additions documented as potentially breaking
- [ ] Complex nullability explained
- [ ] Deprecations include removal timelines

### Performance
- [ ] Composition complexity reasonable
- [ ] No circular @requires dependencies
- [ ] DataLoader patterns considered for N+1 prevention
- [ ] Pagination used for list fields

### Evolution
- [ ] No version suffixes (V1, V2)
- [ ] Schema composes successfully
- [ ] Breaking changes identified
- [ ] Migration path documented

---

## Review Report Template

```markdown
# GraphQL Schema Review: [Subgraph Name]
**Reviewer:** [Name]  
**Date:** [YYYY-MM-DD]  
**Schema Version:** [Version/Commit]  

## Summary
[APPROVED / APPROVED WITH CHANGES / NEEDS REVISION]

## Federation Compliance
- [ ] ✅ Federation v2.3+ declared
- [ ] ✅ Entity keys use stable identifiers
- [ ] ⚠️ WARNING: Missing @external on Astronaut.id extension
- [ ] ❌ BLOCKER: Equipment entity missing @key directive

## Performance
- [ ] ✅ Pagination implemented on list fields
- [ ] ⚠️ CONCERN: Astronaut.missions could cause N+1

## Schema Quality
- [ ] ✅ Domain modeling follows business concepts
- [ ] ✅ Naming conventions consistent
- [ ] ⚠️ IMPROVEMENT: Add "Null when:" docs to nullable fields

## Documentation
- [ ] ⚠️ WARNING: 5 boolean fields missing true/false documentation
- [ ] ⚠️ WARNING: 8 string fields missing max length/format

## Required Changes (BLOCKER)
1. Add @key directive to Equipment entity
2. Fix Astronaut.id missing @external in reviews subgraph

## Recommended Improvements
1. Document null semantics on 12 nullable fields
2. Add true/false documentation to boolean fields
3. Provide max length/format for string fields

## Approved With Conditions
Schema approved pending BLOCKER fixes above.
Re-review not required for RECOMMENDED improvements.
```

---

## References

- [Apollo Federation Best Practices](https://www.apollographql.com/docs/federation/best-practices/)
- [Schema Design Guidelines](https://pages.github.corp.ebay.com/fgql/federated-graphql-docs/docs/standards/schema/all-schema-standards/)
- [Error Handling Standards](https://pages.github.corp.ebay.com/fgql/federated-graphql-docs/docs/standards/errors/error-standards/)

---

**Version:** 1.0  
**Last Updated:** 2026-04-24
