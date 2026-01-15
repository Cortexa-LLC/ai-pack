# Architecture Decision Records (ADRs)

This directory contains Architecture Decision Records for [Project Name].

## About ADRs

Architecture Decision Records (ADRs) document significant architectural and design decisions made during the project lifecycle. Each ADR captures:
- The context and problem being addressed
- The decision made
- The consequences (positive and negative)
- Alternatives that were considered

## ADR Index

### Active Decisions

| ADR | Title | Date | Status | Related Documents |
|-----|-------|------|--------|-------------------|
| [001](./001-example-decision.md) | Example Decision Title | YYYY-MM-DD | Accepted | [Architecture Doc](../architecture/feature-name/) |

### Superseded Decisions

| ADR | Title | Date | Status | Superseded By |
|-----|-------|------|--------|---------------|
| [XXX](./XXX-old-decision.md) | Old Decision Title | YYYY-MM-DD | Superseded | [ADR-YYY](./YYY-new-decision.md) |

### Rejected Decisions

| ADR | Title | Date | Status | Reason |
|-----|-------|------|--------|--------|
| [XXX](./XXX-rejected.md) | Rejected Decision Title | YYYY-MM-DD | Rejected | [Brief reason] |

## ADR Lifecycle

```
Proposed → Accepted → [Active]
       ↓
   Rejected

Accepted → Deprecated → Superseded
```

**Status Definitions:**
- **Proposed** - Under discussion, not yet decided
- **Accepted** - Approved and in effect
- **Rejected** - Proposed but not accepted
- **Deprecated** - No longer recommended, but not yet replaced
- **Superseded** - Replaced by a newer ADR

## Naming Convention

ADRs are numbered sequentially and use kebab-case for the title:

```
docs/adr/NNN-title-in-kebab-case.md
```

Examples:
- `001-use-graphql-federation.md`
- `002-postgresql-for-transactions.md`
- `003-microservices-architecture.md`

**Important:** ADR numbers are sequential across the **entire project**, not per feature.

## Creating a New ADR

1. Check the highest numbered ADR in this directory
2. Copy the template: `.ai-pack/templates/adr/adr-template.md`
3. Use next sequential number: `NNN+1`
4. Create in task packet: `.ai/tasks/[task-id]/adrs/adr-NNN-title.md`
5. Fill in all sections completely
6. When Architect phase completes, persist to `docs/adr/`
7. Update this README with new ADR entry

## Cross-References

ADRs should cross-reference:
- **PRD** - Product requirements that drove the decision
- **Architecture Docs** - System design context
- **Related ADRs** - Other ADRs that influenced or relate to this decision
- **Implementation** - Code that implements the decision (added by Engineers)

## Template

See: [.ai-pack/templates/adr/adr-template.md](.ai-pack/templates/adr/adr-template.md)

## By Category

### Infrastructure Decisions
- [ADR-XXX: Title](./XXX-title.md)

### Data Storage Decisions
- [ADR-XXX: Title](./XXX-title.md)

### API Design Decisions
- [ADR-XXX: Title](./XXX-title.md)

### Security Decisions
- [ADR-XXX: Title](./XXX-title.md)

### Deployment Decisions
- [ADR-XXX: Title](./XXX-title.md)

## By Feature

### [Feature Name]
- [ADR-XXX: Title](./XXX-title.md) - [Brief description]

---

**Last Updated:** YYYY-MM-DD
**Total ADRs:** N
**Active:** N | **Superseded:** N | **Rejected:** N
