# Federated GraphQL Designer
<!-- skills/federated_graphql_designer.skill.md -->

**Version:** 1.0
**InjectAt:** role_context
**Slot:** 35
**Tools:** kg__search_knowledge, kg__add_entity, kg__add_observation, Bash, Read, Write, Grep, Glob
**Gates:** federation-compliance-required
**MaxExtraTokens:** 8000
**Optional:** true

---

## Federated GraphQL Schema Design

Design GraphQL schemas following Apollo Federation 2.3+ patterns for federated architectures. Emphasizes client-first design, entity ownership patterns, and compliance with federation directives.

**Use when:** Designing new GraphQL subgraphs, adding queries/mutations to existing schemas, or designing federated entity relationships.

---

## Core Principles

1. **Client-First Design** — Map the client query flow BEFORE designing the schema
2. **Entity Ownership** — Understand Primary Owner vs Extender vs Stub patterns
3. **Federation Directives** — Correct use of @key, @requires, @provides, @external, @shareable
4. **Iterative Design** — Build one component at a time, get feedback, refine
5. **Pattern Reference** — Check KG and docs for established patterns

---

## Phase 1: Learn Federation Fundamentals

**FIRST ACTION:** Before designing any schema, understand Federation patterns.

### 1.1 Query KG for Existing Patterns

**If KG tools available:**
```javascript
// Search for established GraphQL patterns
kg__search_knowledge({
  query: "GraphQL federation patterns"
})

kg__search_knowledge({
  query: "entity ownership GraphQL"
})
```

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

Ask the user:

```
1. What is the client trying to accomplish?
   - User action or screen that triggers this
   - What data the client needs

2. What existing entities does the client query first?
   - User? Product? Order? CustomEntity?
   - Which subgraph owns these entities?

3. Where does the new schema fit in this flow?
   - Extension of existing entity? (add fields to User)
   - Separate query? (independent data fetch)
   - Mutation triggered by user action?

4. What is the order of Federation Gateway calls?
   - Single federated query? (Gateway orchestrates)
   - Sequential queries? (client makes multiple calls)

5. What does the complete client query look like?
```

**Example Analysis:**
```graphql
# Client perspective: Seller dashboard wants shipping metrics

query SellerDashboard($userId: ID!) {
  user(id: $userId) {
    # Existing fields from userprofile subgraph
    name
    email

    # NEW fields we're designing (from shipping subgraph)
    shippingMetrics {
      score
      recommendations { id message }
    }
  }
}

Decision: Extend User entity (not a new root query)
Reasoning: Client already queries User, adding field is seamless
```

### 2.2 Determine Entity Strategy

Based on client flow, choose ONE:

**A. Extend Existing Entity** (most common)
```graphql
# Your subgraph adds fields to an entity owned by another subgraph
type User @key(fields: "id") @extends {
  id: ID! @external
  shippingMetrics: ShippingMetrics
}
```

**B. Own New Entity**
```graphql
# Your subgraph owns a new entity
type ShippingInsight @key(fields: "id") {
  id: ID!
  userId: ID!
  score: Float!
}
```

**C. Reference Entity (Stub)**
```graphql
# Your subgraph references but doesn't extend an entity
type User @key(fields: "id", resolvable: false) {
  id: ID!
}
```

### 2.3 Identify Missing Information

Ask for specific details needed:

```
To design the schema, I need:

1. Entity definitions from Supergraph (if extending):
   - Which entity? (User, Product, Order, etc.)
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
**Choice:** Extend User entity
**Reasoning:** Client already queries User for dashboard data

### Operations
- **Queries:** None (data accessed via User extension)
- **Mutations:** updateShippingPreferences (if user can configure)

### Key Pattern
**Nested key:** `@key(fields: "id")` on User
**Requires:** May need `@requires(fields: "accountType")` if metrics differ by account

### Error Handling
**Pattern:** Union response type (success | error)
**Why:** Business rules (e.g., metrics not available for new users)
```

### 3.2 Show Schema Skeleton

```graphql
# Client query (how it will be used):
query {
  user(id: "123") {
    shippingMetrics { ... }  # NEW
  }
}

# Schema structure:
type User @key(fields: "id") @extends {
  id: ID! @external
  shippingMetrics: ShippingMetrics
}

type ShippingMetrics {
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

Step 1: Define User extension and key
Step 2: Define ShippingMetrics type
Step 3: Add Recommendation type
Step 4: Add mutation (if needed)
Step 5: Add error handling
Step 6: Document and validate
```

**Get user approval before proceeding.**

---

## Phase 4: Build Iteratively

### 4.1 Work One Component at a Time

For each component:

1. **Design** the component based on patterns
2. **Show** the schema for this component
3. **Explain** decisions (why this pattern? why these fields?)
4. **Get feedback** before next component

### 4.2 Apply Federation Best Practices

**Naming Conventions:**
```graphql
# ✅ Queries/Mutations: domainAction (no "get" prefix)
shippingMetrics(input: ShippingMetricsInput!): ShippingMetricsOutput!

# ❌ Avoid:
getShippingMetrics(userId: ID!): ShippingMetrics

# ✅ Input/Output: Unique per operation
input ShippingMetricsInput { userId: ID! }
type ShippingMetricsOutput { metrics: ShippingMetrics! }

# ✅ Types: Domain-specific names
type ShippingMetrics { ... }

# ❌ Avoid: Generic names
type Metrics { ... }
```

**Documentation:**
```graphql
"""
Shipping performance metrics for a seller.
Updated daily at 00:00 UTC.
"""
type ShippingMetrics {
  """
  Overall shipping performance score (0.0 - 100.0).
  Higher is better. Null when: user has no shipments in last 30 days.
  """
  score: Float

  """
  Personalized recommendations to improve shipping score.
  Always non-null, may be empty array.
  """
  recommendations: [Recommendation!]!
}
```

**Federation Directives:**

| Directive | When to Use | Example |
|-----------|-------------|---------|
| `@key` | Define entity identity | `@key(fields: "id")` |
| `@extends` | Extend entity from other subgraph | `type User @extends` |
| `@external` | Mark field resolved elsewhere | `id: ID! @external` |
| `@requires` | Need field from other subgraph | `@requires(fields: "accountType")` |
| `@provides` | Optimize by providing foreign field | `@provides(fields: "user { name }")` |
| `@shareable` | Value type used by multiple subgraphs | `type Price @shareable` |

### 4.3 Store Design Decisions in KG

**If KG tools available**, record architectural choices:

```javascript
kg__add_entity({
  name: "ShippingMetrics GraphQL Design",
  type: "design_decision",
  properties: {
    decision: "Extend User entity rather than create new root query",
    rationale: "Client already queries User, seamless addition",
    tradeoffs: "Tightly coupled to User schema changes",
    date: "2026-03-14"
  }
})

kg__add_observation({
  entity_id: design_id,
  content: `[DECISION] Used nested @key pattern because User is owned by userprofile subgraph.
Required federation 2.3+ for @shareable support.`
})
```

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
type User @key(fields: "id") @extends {
  id: ID! @external
  accountType: String! @external  # If needed for @requires
  shippingMetrics: ShippingMetrics
}

# New types
type ShippingMetrics {
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
  updateShippingPreferences(input: UpdateShippingPreferencesInput!): UpdateShippingPreferencesOutput
}

input UpdateShippingPreferencesInput {
  userId: ID!
  notifications: Boolean
}

type UpdateShippingPreferencesOutput {
  success: Boolean!
  user: User
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
query SellerDashboard($userId: ID!) {
  user(id: $userId) {
    # Existing fields
    name
    email
    accountType

    # New fields from our schema
    shippingMetrics {
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

- **Uses `kg_reader`** — Query existing GraphQL patterns
- **Uses `kg_writer`** — Store design decisions
- **Used by roles:** designer, architect
- **Complements:** API design workflows

---

## Example Usage

```
User: "Design a GraphQL schema for seller shipping insights"
Context: "Extends User entity, data from shipping analytics service"

Agent:
1. Queries KG for GraphQL federation patterns
2. Maps client flow: Seller dashboard queries User
3. Proposes: Extend User entity (not new root query)
4. Designs iteratively:
   - User extension with @key
   - ShippingMetrics type
   - Recommendation type
   - Documentation
5. Stores design decision in KG
6. Generates complete schema with compliance checklist
7. Provides client query example
```

---

## Notes

- **Client-first always:** Don't design schema in isolation from how it's consumed
- **Federation is complex:** When unsure, search for similar patterns in existing schemas
- **Document null semantics:** Every nullable field needs "Null when:" documentation
- **Avoid over-design:** Start simple, evolve schema based on real usage
- **Validate early:** Check federation composition before full implementation
