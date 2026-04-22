# Federated GraphQL Skills

Two comprehensive skills for designing and reviewing Apollo GraphQL Federation schemas.

## Skills Overview

### 1. federated_graphql_designer.skill.md
**Purpose:** Guide engineers through designing federated GraphQL schemas  
**Version:** 2.0  
**Slot:** 35

Helps design Apollo Federation 2.0+ schemas with:
- Client-first design methodology
- Domain modeling (business concepts, not database structure)
- Entity ownership patterns (Owner, Extender, Stub)
- Stable key selection
- Federation directives (@key, @requires, @provides, @shareable)
- Performance optimization (DataLoader, N+1 prevention)
- Common production issue prevention

### 2. federated_graphql_reviewer.skill.md
**Purpose:** Review schemas for compliance, performance, and best practices  
**Version:** 1.0  
**Slot:** 36

Comprehensive 7-phase review process:
1. Initial Assessment
2. Federation Compliance
3. Performance Review
4. Schema Design Quality
5. Error Handling & Resilience
6. Governance & Standards
7. Generate Review Report

## Key Features

### Completely Generic Examples
All examples use **Apollo mission theme** (space exploration):
- `Astronaut` (entity) - crew members
- `Mission` (entity) - space missions
- `Launch` (entity) - launch events
- `Equipment` (entity) - spacecraft equipment
- `Spacecraft` (entity) - vehicles
- `Coordinate` (value type) - latitude/longitude
- `MissionStatus` (enum) - mission states

This theme is:
- ✅ Domain-agnostic (not tied to any specific industry)
- ✅ Memorable and engaging
- ✅ Matches Apollo GraphQL's own documentation style

### Industry Best Practices

**Netflix Production Learnings:**
- Stable primary keys (won't change)
- Query planning <10ms overhead
- DataLoader batching for N+1 prevention
- Partial error handling (return partial results)
- Domain modeling over database structure
- When to use federation (70+ services, hundreds of developers)
- When NOT to use (<10 developers, strong transactional consistency)

**WunderGraph Patterns:**
- Subgraph architecture (independent services)
- Router orchestration (query decomposition)
- Supergraph composition
- Team autonomy and technology flexibility

**Real Production Issues (from Schema Registry PRs):**

**Common Warnings:**
1. Inconsistent descriptions across subgraphs → Use `{inherit}`
2. Unnecessary @external on @key fields → Fed v1 relic, remove it
3. Breaking enum value additions → Clients may not handle new cases
4. Nullability inconsistency → Supergraph uses weakest constraint
5. Shared types missing @shareable → Composition fails
6. Input field additions → May break strict validation

**Documentation Warnings:**
7. Boolean fields missing true/false explanation → Must document semantics
8. String fields missing max length → Provide character limits
9. String fields missing format → Specify allowed characters (alphanumeric, etc.)
10. Complex nullability undocumented → Explain array/nested requirements

**Structural Issues:**
11. Unused enum types → Remove or reference in schema
12. Interface without implementations → Need at least one concrete type
13. Incomplete interface fields → Field defined in some subgraphs but not all
14. Potential duplicate types → Similar types with same structure
15. Union type additions → May break client type narrowing

## Tools Integration

**Both Skills:**
- Standard tools: `Bash`, `Read`, `Write`, `Grep`, `Glob`

**Optional (gracefully degrades if unavailable):**
- `kg__search_knowledge`, `kg__add_entity`, `kg__add_observation` (KG tools)
- `mcp__upk__search_knowledge`, `mcp__upk__add_learning` (upk MCP server)

**Reviewer Additional:**
- `rover` CLI or `@apollo/federation` for composition validation

## Usage Patterns

### Designer Workflow
```
User: "Design a GraphQL schema for astronaut mission assignments"

Agent:
1. Queries KG/upk for patterns (optional)
2. Reads existing schemas
3. Maps client query flow
4. Proposes: Extend Astronaut entity
5. Designs iteratively:
   - Astronaut extension with stable @key
   - Mission type
   - Assignment relationship
   - Documentation with null semantics
6. Stores decisions (optional)
7. Generates schema + compliance checklist
8. Provides client query example
```

### Reviewer Workflow
```
User: "Review this schema before deployment"

Agent:
1. Reads schema file
2. Validates federation compliance (v2.0+, @key, @external)
3. Checks performance risks (N+1, pagination, complexity)
4. Reviews domain modeling and naming
5. Validates error handling
6. Tests composition (if tools available)
7. Generates structured report:
   - BLOCKER: Must fix
   - WARNING: Should fix
   - IMPROVEMENT: Nice to have
8. Provides code examples for fixes
```

## Common Issues Prevented/Caught

### Designer Prevents:
- ❌ Using business keys (name, email) as @key
- ❌ Missing @shareable on shared value types
- ❌ Inconsistent descriptions across subgraphs
- ❌ Unnecessary @external on @key fields
- ❌ Nullability inconsistency
- ❌ Missing pagination on list fields
- ❌ Generic type names (Metrics vs MissionMetrics)
- ❌ Boolean fields without true/false documentation
- ❌ String fields without max length/format
- ❌ Unused enum types
- ❌ Interfaces without implementations
- ❌ Duplicate error/value types

### Reviewer Catches:
- ❌ Federation v1 syntax relics
- ❌ Breaking enum additions without documentation
- ❌ Missing documentation on nullable fields
- ❌ N+1 query risks
- ❌ Unbounded list fields
- ❌ Composition warnings and errors
- ❌ Version suffixes (V1, V2)
- ❌ Missing deprecation timelines
- ❌ Boolean fields without semantic documentation
- ❌ String fields missing constraints (length/format)
- ❌ Incomplete interface field coverage
- ❌ Potential type duplication
- ❌ Breaking union type additions

## Knowledge Base (upk)

The following learnings are stored in upk:

1. **Netflix GraphQL Federation Best Practices**
   - Entity design with stable keys
   - Schema governance model
   - Performance targets
   - Migration strategy

2. **WunderGraph Architecture Patterns**
   - Subgraph/Router/Supergraph model
   - Team autonomy benefits
   - Query planning optimization

3. **Generic Schema Review Checklist**
   - 14-point validation checklist
   - Naming conventions
   - Entity patterns
   - Documentation requirements

4. **Common Production Issues**
   - Inconsistent descriptions
   - @external anti-patterns
   - Enum breaking changes
   - Nullability conflicts

## Sources

- [Netflix GraphQL Federation at Scale](https://medium.com/@simardeep.oberoi/graphql-federation-at-scale-the-netflix-engineering-blueprint-85358b653e52)
- [Netflix's Federated Supergraph Journey](https://www.apollographql.com/blog/an-unexpected-journey-how-netflix-transitioned-to-a-federated-supergraph)
- [WunderGraph Federation Overview](https://wundergraph.com/blog/a-brief-overview-of-open-source-graphql-federation)
- Apollo Federation Specification: https://www.apollographql.com/docs/federation/
- Real production PR reviews from enterprise schema registries

## When to Use Each Skill

**Use Designer When:**
- Starting new GraphQL subgraph design
- Adding queries/mutations to existing schemas
- Planning federated entity relationships
- Need guidance on federation patterns
- Want to avoid common mistakes upfront

**Use Reviewer When:**
- Pre-deployment schema validation
- Schema design review/approval
- Post-deployment audits
- Quality gate enforcement
- Teaching/mentoring on best practices

## Integration Example

```graphql
# Designer creates this schema:
type Astronaut @key(fields: "id") @extends {
  id: ID! @external
  assignedMissions: [Mission!]!
}

type Mission @key(fields: "id") {
  id: ID!
  name: String!
  status: MissionStatus!
}

# Reviewer validates:
✅ Fed v2.0+ declared
✅ Stable @key (id, not name)
✅ @external on extended entity
✅ Pagination... ⚠️ assignedMissions needs pagination
✅ Documentation... ❌ Missing descriptions
```

## Version History

- **v2.0 (Designer)** - 2026-04-22
  - Added Netflix/WunderGraph best practices
  - Apollo mission-themed examples
  - Production issue prevention
  - Made fully generic
  
- **v1.0 (Reviewer)** - 2026-04-22
  - Initial release
  - Real production issues from PR reviews
  - Apollo mission-themed examples
  - 7-phase review framework
  - Comprehensive validation checklist
