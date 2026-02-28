# Role Performance Documentation

This directory contains per-role model recommendations and escalation paths.

## Quick Reference

| Role | Description | Documentation |
|------|-------------|---------------|
| [engineer](./engineer.md) | Code implementation, TDD, bug fixing | [engineer.md](./engineer.md) |
| [architect](./architect.md) | System design and architecture decisions | [architect.md](./architect.md) |
| [tester](./tester.md) | Test suite creation and quality assurance | [tester.md](./tester.md) |
| [reviewer](./reviewer.md) | Code review and improvement suggestions | [reviewer.md](./reviewer.md) |
| [orchestrator](./orchestrator.md) | Multi-agent task coordination and delegation | [orchestrator.md](./orchestrator.md) |
| [spelunker](./spelunker.md) | Codebase navigation and exploration | [spelunker.md](./spelunker.md) |
| [cartographer](./cartographer.md) | Architecture mapping and documentation | [cartographer.md](./cartographer.md) |
| [inspector](./inspector.md) | Bug root-cause analysis | [inspector.md](./inspector.md) |
| [archaeologist](./archaeologist.md) | Legacy code understanding and rationale recovery | [archaeologist.md](./archaeologist.md) |
| [product-manager](./product-manager.md) | Requirements specification and user story creation | [product-manager.md](./product-manager.md) |
| [designer](./designer.md) | UX flow design and user experience | [designer.md](./designer.md) |
| [strategist](./strategist.md) | Business strategy and planning | [strategist.md](./strategist.md) |

## Grade Legend

| Grade | Success Rate | Recommendation |
|-------|-------------|----------------|
| A | >90% | Strongly recommended — default choice |
| B | 75–90% | Recommended — reliable for production |
| C | 60–75% | Acceptable — monitor for edge cases |
| D | 40–60% | Marginal — consider alternatives |
| F | <40% | Failing — do not use |

## Understanding Grades

Grades are recorded from live task executions and stored in `.claude/performance_grades/`. Each grade file covers one model × role × project combination and includes a **confidence score** based on sample size.

The grade selector in `internal/monitoring/model_selector.go` automatically picks the cheapest model with Grade A or B for each role. New models start at Grade C (cold-start) and earn their grade through production evidence.

See `docs/adr/005-grade-seeding-redesign.md` for the full grade seeding policy.
