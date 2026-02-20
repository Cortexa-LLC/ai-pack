## Orchestrator-Engineer Handoff Protocol

**Purpose:** Prevent engineers from burning context on research that the orchestrator already did.
An engineer with a vague brief will read files to understand the codebase. An engineer with a
pre-cooked brief will write code immediately.

**Evidence:** Task `ai-pack-8x0` received a 4-bullet description of what to fix. The engineer spent
all 600 turns reading files to understand the codebase — zero code written, ~$40 burned. The retry
task `ai-pack-ms7` had complete code inline. The engineer wrote all changes in 12 turns.

---

## Required Elements of Every Engineer Brief

Every engineer brief MUST contain all four elements:

### 1. Exact Absolute File Paths
Not "the streaming adapter" — the actual path:
```
TARGET FILE: a2a-agent/internal/streaming/openai_adapter.go
```

### 2. Specific Changes with Code
Not "fix the tool call handling" — the exact diff:
```go
// Replace Next() with:
func (a *OpenAIChatStreamAdapter) Next() bool {
    if a.done { return false }
    ...
}
```

### 3. "All context provided" signal
This exact phrase triggers Execution Mode Bypass in engineer.md — the engineer skips all
planning phases and proceeds directly to writing code.

### 4. Acceptance Criteria as Shell Commands
Not "verify it works" — commands that pass or fail:
```bash
go build ./...               # must exit 0
grep -c 'ToolCalls' openai_adapter.go  # must be >= 2
```

---

## Worked Example: ai-pack-8x0 (Bad) vs ai-pack-ms7 (Good)

### Bad Brief (ai-pack-8x0) — 600 turns, $40, zero code

```
Fix OpenAI multi-turn tool use and unblock go build

1. Fix GraphQL compile errors (three functions trapped in comment block
   in schema.resolvers.go)
2. Rewrite OpenAIChatStreamAdapter to emit tool_use events (currently
   only handles text deltas)
3. Fix multi-role message handling in CreateStream() Chat path
4. Rewrite Responses API stream adapter for tool calls
```

**Why it failed:** No file paths, no code, no signal to skip research. The engineer read
`server.go`, `openai_adapter.go`, `interfaces.go`, `schema.resolvers.go`, every imported package,
and the entire openai-go SDK source before writing a single line of code.

### Good Brief (ai-pack-ms7) — 12 turns, $3, all changes correct

```
All context provided. Previous run (ai-pack-8x0) burned 600 turns reading.

TARGET FILE: internal/streaming/openai_adapter.go (read this file first)

--- CHANGE 1: Fix GraphQL compile errors ---
File: internal/graphql/schema.resolvers.go
Find the block starting ~line 432:
  /*
     func convertTaskInfoToAgentTask ...
  */
Remove the /* and */ delimiters only. Keep function bodies intact.
Verify: go build ./internal/graphql/... passes.

--- CHANGE 2: Rewrite OpenAIChatStreamAdapter ---
Add fields: acc openai.ChatCompletionAccumulator, pendingEvents []StreamEvent, done bool

Replace Next() with:
  func (a *OpenAIChatStreamAdapter) Next() bool {
    if a.done { return false }
    if len(a.pendingEvents) > 0 { ... }
    ...
  }

COMPILE STEP: Run go build ./... after each change.
```

**Why it worked:** The engineer read the target file once, wrote the code, built, done.

---

## Orchestrator Checklist

Before spawning any engineer, verify:

- [ ] Brief contains at least one **exact absolute file path**
- [ ] Brief contains **specific code** to write (or exact line numbers + what to change)
- [ ] Brief contains the phrase **"All context provided"**
- [ ] **Acceptance criteria are shell commands** that exit 0 on success
- [ ] Brief explains the **context in 1–2 sentences** (what failed, why this fixes it)
- [ ] If the task requires reading first, the brief says **which file to read** (not "understand the codebase")

**If you cannot write the brief without telling the engineer to do research, do the research
yourself first** (using spelunker role) and include the findings inline.

---

## When to Use Spelunker First

Spawn a spelunker before the engineer when:

- You don't know the exact file paths
- You're not sure what the current code looks like
- You need to understand an unfamiliar subsystem before specifying changes

The spelunker's job is to produce a design doc / set of exact findings. The orchestrator then
incorporates those findings into the engineer brief.

**Pattern:**
```
spelunker task → design doc → orchestrator reads doc → pre-cooks engineer brief → engineer task
```

This is exactly how `ai-pack-jf9` (spelunker for pause/resume design) fed into the
`ai-pack-7bd` engineer task.
