# Federated GraphQL Skills Summary

## Overview

Two complementary skills for working with Apollo GraphQL Federation:

1. **federated_graphql_designer** - Design federated GraphQL schemas
2. **federated_graphql_reviewer** - Review schemas for compliance, performance, and best practices

Both skills are **generic** and incorporate industry best practices from:
- Apollo Federation 2.0+ specification
- Netflix's production learnings at scale
- WunderGraph's federation patterns
- Generic GraphQL schema design principles

## Key Resources Loaded into UPK

The following knowledge has been stored in the upk (user personal knowledge) MCP server:

**Note:** The skills themselves are fully generic and work for any GraphQL Federation implementation.

### 1. Netflix Best Practices (Core to Skills)
- Entity design with stable keys
- Schema governance model (working group + team autonomy)
- Performance targets (<10ms query planning, DataLoader batching)
- Partial failure handling pattern
- Migration strategy (bridge layer, incremental extraction)
- When to use/not use federation (team size, infrastructure)

### 2. WunderGraph Architecture Patterns (Core to Skills)
- Subgraph/Router/Supergraph architecture
- Team autonomy and technology flexibility
- Query planning optimization
- Domain encapsulation patterns

### 3. Schema Review Signoff Checklist (Core to Skills)
- 13-point validation checklist
- Naming conventions
- Input argument patterns
- Entity entry points
- Type reuse and extensibility
- Documentation requirements
- Query plan validation
- Deployment verification across environments

## Skill 1: federated_graphql_designer

**Purpose:** Guide engineers through designing federated GraphQL schemas following Apollo Federation 2.0+ patterns.

**Key Features:**
- Client-first design approach (map query flow BEFORE schema)
- Domain modeling emphasis (business concepts, NOT database structure)
- Stable key selection (identifiers that won't change)
- Entity ownership patterns (Owner, Extender, Stub)
- Federation directive usage (@key, @requires, @provides, @shareable)
- Performance considerations (DataLoader batching, N+1 prevention)
- Iterative design process with user feedback loops
- Optional KG/upk integration for pattern lookup and decision storage

**Netflix Principles Integrated:**
- Stable primary keys for entity identification
- Composite keys when necessary
- Domain-first modeling (not DB schema exposure)
- Performance targets (<10ms query planning)
- Partial error handling patterns
- Team autonomy via subgraph ownership

**When to Use:**
- Designing new GraphQL subgraphs
- Adding queries/mutations to existing schemas
- Designing federated entity relationships
- Planning schema extensions

**Works With:**
- `kg__search_knowledge` / `mcp__upk__search_knowledge` (optional)
- `kg__add_entity` / `kg__add_observation` / `mcp__upk__add_learning` (optional)
- Falls back to built-in Apollo Federation best practices when KG/upk unavailable

## Skill 2: federated_graphql_reviewer

**Purpose:** Act as a "GraphQL Champion" or schema reviewer to validate schemas before deployment.

**Key Features:**
- Federation compliance validation (version, directives, entity keys)
- Performance risk assessment (N+1 queries, unbounded lists, query complexity)
- Schema quality review (domain modeling, naming, documentation)
- Error handling validation (partial failures, business vs technical errors)
- Governance checks (versioning, composition, breaking changes)
- Structured review report generation with severity levels (BLOCKER/WARNING/IMPROVEMENT)
- Actionable feedback with code examples

**Review Phases:**
1. Initial Assessment (context, schema location)
2. Federation Compliance (version, @key, @external, directives)
3. Performance Review (N+1 detection, pagination, resolver efficiency)
4. Schema Design Quality (domain modeling, naming, documentation)
5. Error Handling & Resilience (partial failures, authorization)
6. Governance & Standards (versioning, composition validation)
7. Report Generation (structured, actionable)

**Validation Checklist:**
- ✅ Federation version declared (v2.0+)
- ✅ Stable identifiers in @key fields
- ✅ Proper @external usage on extensions
- ✅ No circular @requires dependencies
- ✅ List fields paginated
- ✅ DataLoader pattern for batching
- ✅ Domain modeling (not DB structure)
- ✅ Comprehensive documentation with null semantics
- ✅ Error handling (union patterns or nullable fields)
- ✅ Schema composition success
- ✅ No version suffixes (V1, V2)
- ✅ Deprecation timelines documented

**When to Use:**
- Pre-deployment schema reviews
- Schema design validation
- Post-deployment audits
- Quality gate enforcement

**Works With:**
- Optional KG/upk lookup for anti-patterns
- `rover` or `@apollo/federation` CLI for composition validation
- Complements `federated_graphql_designer` skill

## Integration Between Skills

**Designer → Reviewer Workflow:**
1. Designer creates schema following best practices
2. Designer stores design decisions in KG/upk (optional)
3. Reviewer validates schema against checklist
4. Reviewer references stored patterns from KG/upk (optional)
5. Reviewer provides structured feedback
6. Designer iterates based on feedback

## Tool Requirements

**Both Skills:**
- `Bash`, `Read`, `Write`, `Grep`, `Glob` (always available)

**Optional (gracefully degrades):**
- `kg__search_knowledge`, `kg__add_entity`, `kg__add_observation` (KG tools)
- `mcp__upk__search_knowledge`, `mcp__upk__add_learning` (upk MCP server)

**Reviewer Additional:**
- `rover` CLI or `@apollo/federation` for composition validation (recommended)

## Netflix Production Learnings Applied

**Entity Design:**
- Stable primary keys (won't change)
- Composite keys support
- `__resolveReference` with DataLoader batching

**Performance:**
- Query planning <10ms overhead target
- Gateway processing thousands of queries/second
- Response times sub-100ms for most operations
- Field-level performance metrics

**Governance:**
- Schema working group for core types
- Team autonomy for owned fields
- Automated linting for naming
- Domain modeling over DB structure

**Migration Strategy:**
- Bridge layer to access monolith
- Incremental extraction
- Parallel deployments reduce risk
- Focus on willing early adopters

**When Federation Makes Sense:**
- ✅ 70+ services, hundreds of developers
- ✅ Multiple teams needing API ownership
- ✅ Schema complexity spanning domains
- ❌ <10 developers
- ❌ Strong transactional consistency needs
- ❌ Limited infrastructure investment

## WunderGraph Patterns Applied

**Architecture:**
- Subgraphs: Independent services with own schema/resolvers/data
- Router: Central orchestrator for query decomposition
- Supergraph: Unified schema composition

**Benefits:**
- Team autonomy (parallel development)
- Technology flexibility per subgraph
- Development velocity (independent evolution)
- Single client endpoint

**Implementation:**
- Domain-specific subgraph design
- Schema composition and validation
- Advanced query planning
- Real-time analytics

## Sources

Industry best practices and learnings from:

- [GraphQL Federation at Scale: The Netflix Engineering Blueprint](https://medium.com/@simardeep.oberoi/graphql-federation-at-scale-the-netflix-engineering-blueprint-85358b653e52)
- [How Netflix Scales Its API with GraphQL Federation - InfoQ](https://www.infoq.com/presentations/netflix-api-graphql-federation/)
- [An Unexpected Journey: How Netflix Transitioned to a Federated Supergraph - Apollo GraphQL Blog](https://www.apollographql.com/blog/an-unexpected-journey-how-netflix-transitioned-to-a-federated-supergraph)
- [Redefining API Strategy: Why Netflix Platform Engineering Chose Federated GraphQL - Apollo GraphQL Blog](https://www.apollographql.com/blog/redefining-api-strategy-why-netflix-platform-engineering-chose-federated-graphql)
- [A Brief Overview of Open Source GraphQL Federation - WunderGraph](https://wundergraph.com/blog/a-brief-overview-of-open-source-graphql-federation)
- Apollo Federation Specification: https://www.apollographql.com/docs/federation/
- Apollo Federated Schema Design Best Practices: https://www.apollographql.com/docs/federation/enterprise-guide/federated-schema-design

## Usage Examples

### Designer Example

```
User: "Design a GraphQL schema for astronaut mission metrics"
Context: "Extends Astronaut entity, data from mission analytics service"

Agent (using federated_graphql_designer):
1. Optionally queries KG/upk for patterns (skips if unavailable)
2. Reads existing project schemas
3. Maps client flow: Mission dashboard queries Astronaut
4. Proposes: Extend Astronaut entity (not new root query)
5. Designs iteratively:
   - Astronaut extension with stable @key
   - MissionMetrics type
   - Recommendation type with enum
   - Documentation with null semantics
6. Optionally stores design decisions
7. Generates complete schema with compliance checklist
8. Provides client query example
```

### Reviewer Example

```
User: "Review this GraphQL schema before deployment"
Context: Provides schema file path

Agent (using federated_graphql_reviewer):
1. Reads schema file
2. Validates federation compliance (version, @key, @external)
3. Checks for performance risks (N+1, unbounded lists, query complexity)
4. Reviews domain modeling and naming conventions
5. Validates error handling patterns
6. Tests schema composition (if tools available)
7. Generates structured review report:
   - BLOCKER: Must fix before deployment
   - WARNING: Should fix, may allow with sign-off
   - IMPROVEMENT: Nice to have
8. Provides specific code examples for all issues
9. Includes approval criteria and next steps
```

## File Locations

- **Designer Skill:** `skills/federated_graphql_designer.skill.md`
- **Reviewer Skill:** `skills/federated_graphql_reviewer.skill.md`
- **This Summary:** `skills/FEDERATED_GRAPHQL_SKILLS_SUMMARY.md`

## Version History

- **v2.0 (Designer):** Enhanced with Netflix/WunderGraph best practices, made generic
- **v1.0 (Reviewer):** Initial creation with comprehensive review framework
- **2026-04-22:** Initial release with industry best practices integrated
