# Federated GraphQL Designer
<!-- skills/federated_graphql_designer.skill.md -->

**Version:** 2.0
**InjectAt:** role_context
**Slot:** 35
**Tools:** kg__search_knowledge, kg__add_entity, kg__add_observation, mcp__upk__search_knowledge, mcp__upk__add_learning, Bash, Read, Write, Grep, Glob
**MaxExtraTokens:** 8000
**Optional:** true

---

## Federated GraphQL Schema Design

Design GraphQL schemas following Apollo Federation 2.3+ patterns for federated architectures. Emphasizes client-first design, entity ownership patterns, and compliance with federation directives.

**Use when:** Designing new GraphQL subgraphs, adding queries/mutations to existing schemas, or designing federated entity relationships.

---

## Core Principles

1. **Client-First Design** — Map the client query flow BEFORE designing the schema
2. **Domain Modeling** — Reflect business concepts, NOT database structure (Netflix principle)
3. **Entity Ownership** — Understand Primary Owner vs Extender vs Stub patterns
4. **Stable Keys** — Use stable primary key fields that won't change as systems evolve (Netflix)
5. **Federation Directives** — Correct use of @key, @requires, @provides, @external, @shareable
6. **Iterative Design** — Build one component at a time, get feedback, refine
7. **Pattern Reference** — Check KG and docs for established patterns

---

## Phase 1: Learn Federation Fundamentals

**FIRST ACTION:** Before designing any schema, understand Federation patterns.

### 1.1 Query Knowledge Base for Existing Patterns (Optional)

**If KG or upk tools are available and populated:**
```javascript
// Try KG first
kg__search_knowledge({
  query: "GraphQL federation patterns"
})

// Or upk
mcp__upk__search_knowledge({
  query: "GraphQL federation patterns"
})
```

**If knowledge base is empty or unavailable:**
- Skip to 1.2 (read project schemas)
- Rely on Apollo Federation best practices documented in this skill
- Can optionally fetch docs from apollographql.com if needed

### 1.2 Read Project GraphQL Conventions (if available)

```bash
# Look for project-specific patterns
find . -name "*graphql*conventions*.md" -o -name "*federation*guide*.md"
find . -path "*/schema/README.md"
```

### 1.3 Review Existing Subgraph Schemas

```bash
# Find existing schemas for reference
find . -name "*.graphql" -o -name "*.gql"
```

**Study existing schemas for:**
- Naming conventions (queries, mutations, types)
- Entity key patterns
- Error handling approaches
- Documentation style

---

## Phase 2: Analyze Requirements

### 2.1 Map the Client Query Flow

**CRITICAL:** Before designing the schema, understand how clients will consume it.

Ask the astronaut:

```
1. What is the client trying to accomplish?
   - Astronaut action or screen that triggers this
   - What data the client needs

2. What existing entities does the client query first?
   - Astronaut? Equipment? Launch? CustomEntity?
   - Which subgraph owns these entities?

3. Where does the new schema fit in this flow?
   - Extension of existing entity? (add fields to Astronaut)
   - Separate query? (independent data fetch)
   - Mutation triggered by astronaut action?

4. What is the order of Federation Gateway calls?
   - Single federated query? (Gateway orchestrates)
   - Sequential queries? (client makes multiple calls)

5. What does the complete client query look like?
```

**Example Analysis:**
```graphql
# Client perspective: Astronaut dashboard wants analytics metrics

query AstronautDashboard($astronautId: ID!) {
  astronaut(id: $astronautId) {
    # Existing fields from astronautprofile subgraph
    name
    email

    # NEW fields we're designing (from analytics subgraph)
    missionMetrics {
      score
      recommendations { id message }
    }
  }
}

Decision: Extend Astronaut entity (not a new root query)
Reasoning: Client already queries Astronaut, adding field is seamless
```

### 2.2 Determine Entity Strategy

Based on client flow, choose ONE:

**A. Extend Existing Entity** (most common - Netflix pattern)
```graphql
# Your subgraph adds fields to an entity owned by another subgraph
# Example: Reviews subgraph extends Movie owned by content subgraph
type Astronaut @key(fields: "id") @extends {
  id: ID! @external
  missionMetrics: MissionMetrics
}
```

**CRITICAL:** Use **stable identifiers** for @key fields - IDs that won't change as systems evolve.

**B. Own New Entity**
```graphql
# Your subgraph owns a new entity with primary ownership
type MissionInsight @key(fields: "id") {
  id: ID!
  astronautId: ID!
  score: Float!
}
```

**Composite Keys** (when single field insufficient):
```graphql
type LaunchManifest @key(fields: "launchId equipmentId") {
  launchId: ID!
  equipmentId: ID!
  quantity: Int!
}
```

**C. Reference Entity (Stub)**
```graphql
# Your subgraph references but doesn't extend an entity
type Astronaut @key(fields: "id", resolvable: false) {
  id: ID!
}
```

### 2.3 Identify Missing Information

Ask for specific details needed:

```
To design the schema, I need:

1. Entity definitions from Supergraph (if extending):
   - Which entity? (Astronaut, Equipment, Launch, etc.)
   - Need: @key fields, any fields for @requires

2. Data source:
   - Where does data come from? (DB, API, computed)
   - Real-time or can be async?

3. Error handling:
   - Business errors expected? (out of stock, unauthorized)
   - Use union error pattern or throw exceptions?

4. Authorization:
   - Any scope-based access control?
   - Per-field authorization needed?
```

---

## Phase 3: Create High-Level Plan

### 3.1 Propose Approach

Present the design approach with options if applicable:

```markdown
## Proposed Approach

### Entity Strategy
**Choice:** Extend Astronaut entity
**Reasoning:** Client already queries Astronaut for dashboard data

### Operations
- **Queries:** None (data accessed via Astronaut extension)
- **Mutations:** updateMissionPreferences (if astronaut can configure)

### Key Pattern
**Nested key:** `@key(fields: "id")` on Astronaut
**Requires:** May need `@requires(fields: "accountType")` if metrics differ by account

### Error Handling
**Pattern:** Union response type (success | error)
**Why:** Business rules (e.g., metrics not available for new astronauts)
```

### 3.2 Show Schema Skeleton

```graphql
# Client query (how it will be used):
query {
  astronaut(id: "123") {
    missionMetrics { ... }  # NEW
  }
}

# Schema structure:
type Astronaut @key(fields: "id") @extends {
  id: ID! @external
  missionMetrics: MissionMetrics
}

type MissionMetrics {
  score: Float!
  recommendations: [Recommendation!]!
}

type Recommendation {
  id: ID!
  message: String!
}
```

### 3.3 Break Into Steps

```markdown
Let's build this incrementally:

Step 1: Define Astronaut extension and key
Step 2: Define MissionMetrics type
Step 3: Add Recommendation type
Step 4: Add mutation (if needed)
Step 5: Add error handling
Step 6: Document and validate
```

**Get astronaut approval before proceeding.**

---

## Phase 4: Build Iteratively

### 4.1 Work One Component at a Time

For each component:

1. **Design** the component based on patterns
2. **Show** the schema for this component
3. **Explain** decisions (why this pattern? why these fields?)
4. **Get feedback** before next component

### 4.2 Apply Federation Best Practices

**Performance Considerations (Netflix learnings):**
- Design resolvers to support DataLoader batching (avoid N+1 queries)
- Query planning overhead should be <10ms for the router/gateway
- Enable parallel execution where dependencies allow
- Consider query cost analysis for expensive operations
- Add field-level performance metrics to identify bottlenecks

**Error Handling (Netflix pattern):**
```graphql
# Return partial results rather than failing entire query
type MissionMetricsResult {
  data: MissionMetrics
  errors: [MissionError!]
}

type MissionError {
  code: String!
  message: String!
}
```

**Avoiding Common Equipmention Issues:**

**1. Consistent Descriptions for Shared Types**
```graphql
# ✅ GOOD: Use {inherit} for shared types
"""
{inherit}
"""
type PaginationInput {
  maxPageSize: Int
  pageCursor: String
}

# ❌ BAD: Custom description that differs from supergraph
"""
Input for pagination
"""
type PaginationInput { ... }  # Will cause composition warning
```

**2. Proper @shareable for Value Types**
```graphql
# ✅ GOOD: Shared value types need @shareable
type Coordinate @shareable {
  latitude: Float!
  longitude: String!
}

# ❌ BAD: Duplicate type without @shareable
type Coordinate {  # Composition error if defined in multiple subgraphs
  latitude: Float!
  longitude: String!
}
```

**3. No @external on @key Fields (Fed v2)**
```graphql
# ✅ GOOD: @key fields don't need @external
type Spacecraft @key(fields: "astronaut { id }") {
  astronaut: Astronaut!  # No @external needed
  settings: SpacecraftSettings
}

# ❌ BAD: Federation v1 syntax (relic)
type Spacecraft @key(fields: "astronaut { id }") {
  astronaut: Astronaut! @external  # Unnecessary in Fed v2
  settings: SpacecraftSettings
}
```

**4. Consistent Nullability**
```graphql
# ✅ GOOD: Consistent across all subgraphs
type Node {
  id: ID!  # Always non-null everywhere
}

# ❌ BAD: Different nullability in different subgraphs
# Subgraph A: id: ID!
# Subgraph B: id: ID   # Supergraph uses nullable (weakest constraint)
```

**5. Enum Evolution Strategy**
```graphql
# ⚠️ CAUTION: Adding enum values can break clients
enum MissionStatus {
  ACTIVE
  INACTIVE
  PENDING  # Adding this may break clients without default case
}

# ✅ BETTER: Design enums defensively with UNKNOWN
enum MissionStatus {
  UNKNOWN  # Allows graceful handling of new values
  ACTIVE
  INACTIVE
}

# ✅ ALSO: Remove unused enums
# Don't define enums that no field references - causes DEBUG warnings
```

**6. Interface Design**
```graphql
# ❌ BAD: Interface without implementations
interface Vehicle {
  id: ID!
  capacity: Int!
}

type Query {
  vehicles: [Vehicle!]!  # No concrete types!
}

# ✅ GOOD: Provide at least one implementation
type Spacecraft implements Vehicle {
  id: ID!
  capacity: Int!
  fuelType: String!
}
```

**7. Avoid Duplicate Types**
```graphql
# ❌ BAD: Similar error types
type MissionNotFound {
  code: String!
  message: String!
}

type AstronautNotFound {
  code: String!
  message: String!
}

# ✅ BETTER: Shared type with discriminator
type NotFoundError {
  code: String!
  message: String!
  resourceType: String!  # "Mission", "Astronaut"
}
```

**8. Union Type Additions**
```graphql
# ⚠️ CAUTION: Adding members to unions can break clients
union SearchResult = Mission | Astronaut

# Adding Launch may break clients:
union SearchResult = Mission | Astronaut | Launch

# Document union additions as potentially breaking
```

**Naming Conventions:**
```graphql
# ✅ Queries/Mutations: domainAction (no "get" prefix)
missionMetrics(input: MissionMetricsInput!): MissionMetricsOutput!

# ❌ Avoid:
getMissionMetrics(astronautId: ID!): MissionMetrics

# ✅ Input/Output: Unique per operation
input MissionMetricsInput { astronautId: ID! }
type MissionMetricsOutput { metrics: MissionMetrics! }

# ✅ Types: Domain-specific names
type MissionMetrics { ... }

# ❌ Avoid: Generic names
type Metrics { ... }
```

**Documentation:**
```graphql
"""
Astronaut mission metrics for performance tracking.
Updated daily at 00:00 UTC.
"""
type MissionMetrics {
  """
  Overall mission score (0.0 - 100.0).
  Higher is better. Null when: astronaut has no mission in last 30 days.
  """
  score: Float

  """
  Personalized recommendations to improve performance.
  Always non-null, may be empty array.
  """
  recommendations: [Recommendation!]!
  
  """
  Indicates if astronaut is currently on active duty.
  true: Astronaut is assigned to active mission.
  false: Astronaut is in training or between missions.
  """
  isActive: Boolean!
}
```

**Documentation Requirements:**

**Boolean Fields** - Always document what true/false means:
```graphql
# ❌ BAD
isReady: Boolean!

# ✅ GOOD
"""
true: Spacecraft pre-flight checks complete, ready for launch.
false: Spacecraft requires maintenance or awaiting clearance.
"""
isReady: Boolean!
```

**String Fields** - Provide max length and format:
```graphql
# ❌ BAD
description: String!

# ✅ GOOD
"""
Mission description and objectives.
Max 500 characters, alphanumeric with punctuation.
"""
description: String!
```

**Complex Nullability** - Document rules for nested structures:
```graphql
"""
Launch manifest entries.
If provided, must contain at least one equipment entry.
The array itself is non-nullable, each entry also non-nullable.
"""
equipment: [LaunchEquipment!]!
```

**Federation Directives:**

| Directive | When to Use | Example |
|-----------|-------------|---------|
| `@key` | Define entity identity | `@key(fields: "id")` |
| `@extends` | Extend entity from other subgraph | `type Astronaut @extends` |
| `@external` | Mark field resolved elsewhere | `id: ID! @external` |
| `@requires` | Need field from other subgraph | `@requires(fields: "accountType")` |
| `@provides` | Optimize by providing foreign field | `@provides(fields: "astronaut { name }")` |
| `@shareable` | Value type used by multiple subgraphs | `type Price @shareable` |

### 4.3 Store Design Decisions (Optional)

**If KG tools available**, record architectural choices:

```javascript
kg__add_entity({
  name: "MissionMetrics GraphQL Design",
  type: "design_decision",
  properties: {
    decision: "Extend Astronaut entity rather than create new root query",
    rationale: "Client already queries Astronaut, seamless addition",
    tradeoffs: "Tightly coupled to Astronaut schema changes",
    date: "2026-04-22"
  }
})

kg__add_observation({
  entity_id: design_id,
  content: `[DECISION] Used nested @key pattern because Astronaut is owned by astronautprofile subgraph.
Required federation 2.3+ for @shareable support.`
})
```

**Or if upk is available:**
```javascript
mcp__upk__add_learning({
  content: "Used nested @key pattern for Astronaut extension. Rationale: Astronaut owned by astronautprofile subgraph. Required federation 2.3+ for @shareable support.",
  source: "MissionMetrics schema design - 2026-04-22"
})
```

**If neither available:** Skip this step - design decisions can be documented in schema comments or separate markdown files.

---

## Phase 5: Validate and Document

### 5.1 Generate Complete Schema

Combine all components:

```graphql
schema
  @link(url: "https://specs.apollo.dev/federation/v2.3", import: ["@key", "@shareable", "@requires", "@external"])
{
  query: Query
  mutation: Mutation  # If applicable
}

# Entity extensions
type Astronaut @key(fields: "id") @extends {
  id: ID! @external
  accountType: String! @external  # If needed for @requires
  missionMetrics: MissionMetrics
}

# New types
type MissionMetrics {
  score: Float
  recommendations: [Recommendation!]!
}

type Recommendation {
  id: ID!
  message: String!
  priority: Priority!
}

enum Priority {
  HIGH
  MEDIUM
  LOW
}

# Mutations (if applicable)
type Mutation {
  updateMissionPreferences(input: UpdateMissionPreferencesInput!): UpdateMissionPreferencesOutput
}

input UpdateMissionPreferencesInput {
  astronautId: ID!
  notifications: Boolean
}

type UpdateMissionPreferencesOutput {
  success: Boolean!
  astronaut: Astronaut
}
```

### 5.2 Compliance Checklist

Verify:
- [ ] Federation 2.3 or higher
- [ ] Unique Input/Output types per operation
- [ ] All fields documented with null semantics
- [ ] No version suffixes (V1, V2) — use schema evolution instead
- [ ] `@shareable` on value types used across subgraphs
- [ ] `resolvable: false` on reference-only entities
- [ ] Business error handling (if mutations present)

### 5.3 Provide Client Example

```graphql
# Complete client query demonstrating usage:
query AstronautDashboard($astronautId: ID!) {
  astronaut(id: $astronautId) {
    # Existing fields
    name
    email
    accountType

    # New fields from our schema
    missionMetrics {
      score
      recommendations {
        id
        message
        priority
      }
    }
  }
}
```

---

## Integration with Other Skills

- **Optionally uses `kg_reader` or `mcp__upk__search_knowledge`** — Query existing GraphQL patterns if available
- **Optionally uses `kg_writer` or `mcp__upk__add_learning`** — Store design decisions if available
- **Used by roles:** designer, architect
- **Complements:** API design workflows
- **Fallback:** Works standalone with Apollo Federation best practices when KG/upk unavailable

---

## Example Usage

```
Astronaut: "Design a GraphQL schema for astronaut mission insights"
Context: "Extends Astronaut entity, data from analytics service"

Agent:
1. Tries to query KG/upk for GraphQL federation patterns (skips if unavailable)
2. Reads existing project schemas for patterns
3. Maps client flow: Astronaut dashboard queries Astronaut
4. Proposes: Extend Astronaut entity (not new root query)
5. Designs iteratively:
   - Astronaut extension with @key
   - MissionMetrics type
   - Recommendation type
   - Documentation
6. Optionally stores design decision in KG/upk (if available)
7. Generates complete schema with compliance checklist
8. Provides client query example
```

---

## Notes

- **Client-first always:** Don't design schema in isolation from how it's consumed
- **Domain over Database:** Model business concepts, not storage details (Netflix principle)
- **Federation is complex:** When unsure, search for similar patterns in existing schemas
- **Document null semantics:** Every nullable field needs "Null when:" documentation
- **Avoid over-design:** Start simple, evolve schema based on real usage
- **Validate early:** Check federation composition before full implementation
- **Stable keys:** Use IDs that won't change - avoid business keys that might be updated
- **Team autonomy:** Each subgraph should encapsulate domain-specific functionality (WunderGraph pattern)

## When Federation Makes Sense (Netflix guidance)

**Good Fit:**
- Multiple teams needing API ownership
- Schema complexity spanning distinct domains
- 70+ services, hundreds of developers
- Need for technology flexibility per domain

**Poor Fit:**
- Teams with fewer than 10 developers
- Strong transactional consistency requirements across domains
- Limited platform infrastructure investment
- Single team managing entire API surface

**Infrastructure Reality:** Netflix spent a year building platform capabilities before self-service adoption
