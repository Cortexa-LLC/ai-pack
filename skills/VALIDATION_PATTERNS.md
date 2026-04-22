# GraphQL Federation Schema Validation Patterns

## Complete Reference of Production Issues

Extracted from 100+ real schema registry PR reviews. These are the actual validation warnings, errors, and feedback patterns that appear in production GraphQL Federation implementations.

---

## Category 1: Description & Documentation Issues

### 1.1 Inconsistent Descriptions (MOST COMMON)

**Pattern:**
```graphql
# Subgraph A:
"""Pagination parameters"""
type PaginationInput { ... }

# Subgraph B:  
"""Input for pagination"""
type PaginationInput { ... }
```

**Warning:**
```
WARN: Element PaginationInput has inconsistent descriptions across subgraphs.
```

**Solutions:**
- Use `{inherit}` directive to inherit from supergraph
- Match supergraph description exactly
- Coordinate descriptions across teams

**Impact:** Confusing API documentation, composition warnings

---

### 1.2 Boolean Field Documentation

**Pattern:**
```graphql
# ❌ INSUFFICIENT
type Mission {
  isActive: Boolean!
}
```

**Warning:**
```
WARN: Description for FieldName 'isActive' needs below changes:
Specify what true or false indicates
```

**Solution:**
```graphql
# ✅ CORRECT
type Mission {
  """
  Mission active status.
  true: Mission is currently active and accepting crew assignments.
  false: Mission is completed, cancelled, or not yet started.
  """
  isActive: Boolean!
}
```

**Impact:** Validation warning, unclear API semantics

---

### 1.3 String Field Constraints

**Pattern:**
```graphql
# ❌ INSUFFICIENT
type Astronaut {
  name: String!
  description: String
}
```

**Warnings:**
```
WARN: Description for FieldName 'name' needs below changes:
Provide format Ex: Alphanumeric

WARN: Description for FieldName 'description' needs below changes:
Provide maximum length
```

**Solution:**
```graphql
# ✅ CORRECT
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
  description: String
}
```

**Impact:** Validation warnings, unclear field constraints

---

### 1.4 Null Semantics Documentation

**Pattern:**
```graphql
# ❌ INSUFFICIENT
type Astronaut {
  specialty: String
}
```

**Required:**
```graphql
# ✅ CORRECT
type Astronaut {
  """
  Astronaut's mission specialty.
  Null when: astronaut is in general training (no specialty assigned yet).
  """
  specialty: String
}
```

**Impact:** Client confusion about when fields are null

---

### 1.5 Complex Nullability Rules

**Pattern:**
```graphql
# ❌ INSUFFICIENT
type LaunchManifest {
  equipment: [Equipment!]!
}
```

**Solution:**
```graphql
# ✅ CORRECT
type LaunchManifest {
  """
  Equipment required for launch.
  If manifest is provided, must contain at least one equipment entry.
  The array itself is non-nullable, each entry also non-nullable.
  """
  equipment: [Equipment!]!
}
```

**Impact:** Unclear array/nested nullability rules

---

## Category 2: Federation Directive Issues

### 2.1 Unnecessary @external on @key Fields

**Pattern:**
```graphql
# ❌ FEDERATION V1 RELIC
type Spacecraft @key(fields: "astronaut { id }") {
  astronaut: Astronaut! @external  # Unnecessary!
  capacity: Int!
}
```

**Warning:**
```
WARN: The entity extension "Spacecraft" defines a "@key" directive with field set "astronaut { id }".
The following field coordinates are declared "@external": "Spacecraft.astronaut"
Please note fields that form part of entity extension "@key" field sets are always provided.
Any such "@external" declarations are unnecessary relics of Federation Version 1 syntax.
```

**Solution:**
```graphql
# ✅ FEDERATION V2
type Spacecraft @key(fields: "astronaut { id }") {
  astronaut: Astronaut!  # No @external needed
  capacity: Int!
}
```

**Impact:** Cluttered schema, outdated syntax

---

### 2.2 Inconsistent Nullability Across Subgraphs

**Pattern:**
```graphql
# Subgraph A:
type Node {
  id: ID!  # Non-nullable
}

# Subgraph B:
type Node {
  id: ID   # Nullable
}
```

**Info/Warning:**
```
INFO: Type of field "Node.id" is inconsistent but compatible across subgraphs:
will use type "ID" (nullable) in supergraph but "Node.id" has subtype "ID!" in subgraphs A, B, C...
```

**Impact:** Supergraph uses weakest constraint (nullable), may surprise clients expecting non-null

---

### 2.3 Missing @shareable on Value Types

**Pattern:**
```graphql
# Subgraph A:
type Coordinate {
  latitude: Float!
  longitude: Float!
}

# Subgraph B:
type Coordinate {  # ❌ Duplicate definition!
  latitude: Float!
  longitude: Float!
}
```

**Error:**
```
ERROR: Type "Coordinate" is defined in multiple subgraphs but is not marked @shareable
```

**Solution:**
```graphql
# ✅ BOTH SUBGRAPHS
type Coordinate @shareable {
  latitude: Float!
  longitude: Float!
}
```

**Impact:** Composition failure

---

## Category 3: Breaking Changes

### 3.1 Enum Value Additions

**Pattern:**
```graphql
# Original:
enum MissionStatus {
  ACTIVE
  INACTIVE
}

# Change:
enum MissionStatus {
  ACTIVE
  INACTIVE
  PENDING  # ⚠️ NEW VALUE
}
```

**Warning:**
```
WARN: Enum value 'PENDING' was added to enum 'MissionStatus'.
Adding an enum value may break existing clients that were not programming defensively
against an added case when querying an enum.
```

**Client Impact:**
```javascript
// Client code without default case breaks:
switch(mission.status) {
  case 'ACTIVE': ...
  case 'INACTIVE': ...
  // No default - PENDING causes failure!
}
```

**Mitigation:** Design enums with UNKNOWN value from start

---

### 3.2 Input Field Additions

**Pattern:**
```graphql
# Original:
input CreateMissionInput {
  name: String!
  date: DateTime!
}

# Change:
input CreateMissionInput {
  name: String!
  date: DateTime!
  priority: Priority  # ⚠️ NEW OPTIONAL FIELD
}
```

**Warning:**
```
WARN: Input field 'priority' of type 'Priority' was added to input object type 'CreateMissionInput'.
```

**Impact:** Clients with strict input validation may reject unknown fields

---

### 3.3 Union Type Member Additions

**Pattern:**
```graphql
# Original:
union SearchResult = Mission | Astronaut

# Change:
union SearchResult = Mission | Astronaut | Launch  # ⚠️ NEW MEMBER
```

**Warning:**
```
WARN: Member 'Launch' was added to Union type 'SearchResult'.
Adding a possible type to Unions may break existing clients that were not
programming defensively against a new possible type.
```

**Client Impact:**
```typescript
// TypeScript exhaustiveness checking breaks:
function handle(result: SearchResult) {
  switch(result.__typename) {
    case 'Mission': ...
    case 'Astronaut': ...
    // Error: 'Launch' not handled!
  }
}
```

---

## Category 4: Structural Issues

### 4.1 Unused Enum Types

**Pattern:**
```graphql
# ❌ Defined but never referenced
enum MissionPriority {
  HIGH
  MEDIUM
  LOW
}
# No field uses this enum!
```

**Debug Warning:**
```
DEBUG: Enum type "MissionPriority" is defined but unused.
It will be included in the supergraph with all values appearing in any subgraph
("as if" it was only used as an output type).
```

**Solution:** Remove unused enums or reference them in schema

---

### 4.2 Interface Without Implementations

**Pattern:**
```graphql
# ❌ Interface used but no concrete types
interface Vehicle {
  id: ID!
  capacity: Int!
}

type Query {
  vehicles: [Vehicle!]!  # No implementations exist!
}
```

**Warning:**
```
WARN: Subgraph "missions": The Interface "Vehicle" is used as an output type
without at least one Object type implementation defined in the schema.
```

**Solution:**
```graphql
# ✅ Add implementations
type Spacecraft implements Vehicle {
  id: ID!
  capacity: Int!
  fuelType: String!
}
```

---

### 4.3 Incomplete Interface Fields

**Pattern:**
```graphql
# Subgraph A defines interface with field:
interface Vehicle {
  id: ID!
  capacity: Int!
  fuelLevel: Float!  # Defined here
}

# Subgraph B defines same interface without field:
interface Vehicle {
  id: ID!
  capacity: Int!
  # fuelLevel missing!
}
```

**Debug Warning:**
```
DEBUG: Field "Vehicle.fuelLevel" of interface type "Vehicle" is defined in some
but not all subgraphs that define "Vehicle": "Vehicle.fuelLevel" is defined in
subgraphs "spacecraft" but not in subgraph "rover".
```

**Impact:** Inconsistent interface contracts across subgraphs

---

### 4.4 Potential Duplicate Types

**Pattern:**
```graphql
# Similar structure, different names
type MissionNotFound {
  errorCode: String!
  errorMessage: String!
}

type AstronautNotFound {
  errorCode: String!
  errorMessage: String!
}
```

**Warning:**
```
WARN: Potential duplicate type detected: 'MissionNotFound' in subgraph is similar
to 'AstronautNotFound' in supergraph. Reason: Both types represent error conditions
with identical fields 'errorCode' and 'errorMessage' having equivalent types.
```

**Recommendation:** Consolidate to shared error type with discriminator field

---

## Category 5: Performance & Design

### 5.1 Unbounded List Fields

**Pattern:**
```graphql
# ❌ No pagination
type Astronaut {
  missions: [Mission!]!  # Could be thousands!
}
```

**Recommendation:** Always paginate list fields

**Solution:**
```graphql
# ✅ With pagination
type Astronaut {
  missions(first: Int!, after: String): MissionConnection!
}
```

---

### 5.2 N+1 Query Risks

**Pattern:**
```graphql
type Query {
  astronauts: [Astronaut!]!
}

type Astronaut {
  spacecraft: Spacecraft!  # ⚠️ N+1 if not batched
}
```

**Solution:** Implement DataLoader for batching

---

## Validation Severity Levels

**ERROR** - Blocks composition, must fix:
- Missing @shareable on duplicate types
- Invalid @key references
- Circular @requires dependencies

**WARN** - Should fix, may block deployment:
- Inconsistent descriptions
- Breaking changes (enum/union additions)
- Missing @external cleanup
- Boolean/string documentation requirements

**DEBUG/INFO** - Advisory, good to know:
- Unused enums
- Nullability inconsistencies (compatible)
- Incomplete interface fields

**HINT** - Best practice suggestions:
- Potential duplicates
- Performance concerns

---

## Common Review Checklist

Based on production PRs, reviewers consistently check:

1. ✅ All shared types have consistent descriptions or use `{inherit}`
2. ✅ No @external on @key fields (Fed v1 relic)
3. ✅ Boolean fields document true/false semantics
4. ✅ String fields specify max length and format
5. ✅ Nullable fields document "Null when:" conditions
6. ✅ Complex arrays document nullability rules
7. ✅ Shared value types marked @shareable
8. ✅ Enum additions documented as potentially breaking
9. ✅ Input field additions noted for strict validators
10. ✅ Unused enums removed
11. ✅ Interfaces have at least one implementation
12. ✅ Similar types consolidated when appropriate
13. ✅ List fields paginated
14. ✅ No version suffixes (V1, V2)
15. ✅ Deprecations have removal timelines

---

## Quick Reference: Fix Patterns

| Issue | Fix |
|-------|-----|
| Inconsistent descriptions | Use `{inherit}` or match exactly |
| @external on @key | Remove @external |
| Boolean undocumented | Add "true: ..., false: ..." |
| String no constraints | Add "Max N chars, format: X" |
| Nullable unclear | Add "Null when: ..." |
| Missing @shareable | Add to all subgraphs |
| Enum addition | Document as breaking change |
| Unused enum | Remove or reference |
| Interface no impl | Add concrete type |
| Duplicate types | Consolidate with discriminator |
| Unbounded list | Add pagination |
| N+1 risk | Implement DataLoader |

---

This reference represents patterns from 100+ production schema reviews and serves as the foundation for both the Designer and Reviewer skills.
