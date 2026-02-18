# Agent Policy

These rules apply to every agent regardless of role. They are non-negotiable framework constraints, not suggestions.

---

## Execute Explicit Steps First

If the task description or contract contains any of the following, **execute them immediately on turn 1** before reading any files or doing any exploration:

- A numbered list of steps (1. ... 2. ... 3. ...)
- A section labeled "CRITICAL STEPS", "Required Steps", or similar
- An explicit shell command (e.g. `cd ~/Projects/Foo && make clean && make`)

**These are orders, not suggestions.** The agent that wrote them knows what information is needed. Do not substitute your own research plan for explicit instructions.

**Example:** If the task says `1. cd ~/Projects/Apple/A2osX 2. make clean && make 3. Capture the error` — your first tool call is `bash: cd ~/Projects/Apple/A2osX && make clean && make 2>&1`. Not reading a source file. Not checking the directory structure. Run the command.

---

## Token Budget

This task has a finite token budget. The server will terminate you if you exceed it, and your work will be lost. Budget yourself proactively:

- **By turn 5**, you should have executed the explicit steps or started the core work. If you haven't, you are off-track.
- **By turn 15**, you should be approaching a result. If you are still in exploration mode, stop and report what you have.
- **If you haven't made measurable progress toward the stated goal in the last 5 turns**, stop exploring and report findings. Partial results with honest gaps are more valuable than burning the budget.

Do not read files speculatively "just to understand the codebase." Read only what the task requires.

---

## Turn Budget

Every API call has a cost. Treat every turn as precious.

- **Be decisive.** If you have enough information to proceed, proceed.
- **Stop exploring when you have what you need.** Reading 10 more files when 3 would suffice wastes the budget for everyone.
- **If you are stuck after 3 consecutive failed attempts**, stop trying variations and report what you found so far. A partial result with an honest "I couldn't find X" is more valuable than 50 more turns of spinning.

---

## Missing Files and Paths

- **1 attempt only.** If a file, directory, or path does not exist after your first attempt, move on immediately.
- **Never retry variations of a path that returned "not found".** If `.ai/tasks/foo/00-contract.md` doesn't exist, do not try `.ai/tasks/foo/contract.md`, `tasks/foo/00-contract.md`, `./foo/00-contract.md`, etc.
- **Task packets are optional context.** If the task description references a task packet directory that does not exist, skip it entirely and begin work from first principles using the task description itself.
- **Missing context is not a blocker.** Work with what exists.

---

## Error Handling

- **A tool error is information, not a reason to retry the same call.** Read the error, adjust your approach, move on.
- **If every tool call in a turn returns an error**, you are in a bad state. Stop, assess, and take a completely different approach — or report that you are blocked.
- **Don't confuse "I couldn't find it" with "it doesn't exist".** If your search strategy was wrong, try a different search strategy once. If that also fails, assume it doesn't exist and proceed.

---

## Beads Task Workflow

All agents that work on Beads tasks follow this lifecycle:

| Step | Command | When |
|------|---------|------|
| Check task | `bd show <id>` | Before starting |
| Claim task | `bd update --claim <id>` | When starting work |
| Block task | `bd block <id> "reason"` | When blocked on dependency |
| Unblock task | `bd unblock <id>` | When dependency resolved |
| Complete task | `bd close <id>` | When work is done |

**Beads task IDs** follow the format `<prefix>-<hash>` (e.g. `xasm++-qbxv`). Timestamped execution folder names (e.g. `xasm++-qbxv-20260218-084509`) are **not** Beads task IDs — never pass them to `bd` commands.

---

## Work Log

If your role uses a task packet, update `.ai/tasks/*/20-work-log.md` after each major phase. Include:
- What you did
- What you found
- What's next or what's blocking you

This lets the orchestrator track progress without reading your full output.

---

## Reporting When Stuck or Done

When you finish (or can't continue), produce a clear summary:
- What was accomplished
- What was NOT found or completed, and why
- Specific file:line references for any findings
- What the next agent or human needs to do

Do not end with "I need more information" and stop. Make your best determination with what you have and flag the uncertainty explicitly.
