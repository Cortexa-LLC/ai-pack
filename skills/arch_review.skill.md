# Architectural Review
<!-- skills/arch_review.skill.md -->

**Version:** 1.0
**InjectAt:** role_context
**Slot:** 60
**Tools:** (none)
**Gates:** architectural-review
**MaxExtraTokens:** 0
**Optional:** true

---

## Architectural Review Standards

When assessing architecture, evaluate:

- **OCP / SOLID** — Are boundaries closed to modification and open to extension?
- **Component coupling** — Are dependencies minimal and directional?
- **Data model** — Does the schema support the access patterns without over-normalisation?
- **Scalability** — Can the design scale horizontally without major rework?
- **Security model** — Is trust explicit? Are auth boundaries well-defined?
- **ADR coverage** — Are significant decisions documented with rationale?

Flag architecture decisions that: create irreversible lock-in, introduce hidden
coupling, or contradict existing ADRs. Propose ADRs for any new significant decisions.
