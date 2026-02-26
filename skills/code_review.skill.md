# Code Review
<!-- skills/code_review.skill.md -->

**Version:** 1.0
**InjectAt:** role_context
**Slot:** 55
**Tools:** (none)
**Gates:** code-quality-review
**MaxExtraTokens:** 0
**Optional:** true

---

## Code Review Standards

When reviewing code, assess each of the following dimensions:

- **Correctness** — Does it do what it claims? Are edge cases handled?
- **Test coverage** — Are new behaviours covered by tests?
- **Readability** — Is the code easy to understand and maintain?
- **Security** — Are inputs validated? Are credentials handled safely?
- **Performance** — Are there obvious bottlenecks or unnecessary allocations?
- **API consistency** — Does it follow the project's naming and style conventions?

Provide specific, actionable feedback. Approve only when all blocking issues
are resolved. Non-blocking suggestions are welcome but must be clearly labelled.
