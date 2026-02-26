# Test-Driven Development
<!-- skills/tdd.skill.md -->

**Version:** 1.0
**InjectAt:** role_context
**Slot:** 50
**Tools:** (none)
**Gates:** tdd-enforcement
**MaxExtraTokens:** 0
**Optional:** true

---

## TDD Discipline

You follow strict Test-Driven Development:

1. **Red** — write a failing test that captures the intended behaviour
2. **Green** — write the minimal code to make the test pass
3. **Refactor** — clean up with all tests green

Never skip the red phase. All new behaviour must be covered by tests before
the implementation is written. Run tests after every change to confirm they
pass. Do not proceed to the next step until the current step is green.
