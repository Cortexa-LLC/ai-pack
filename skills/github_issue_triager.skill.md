# GitHub Issue Triager
<!-- skills/github_issue_triager.skill.md -->

**Version:** 1.0
**InjectAt:** role_context
**Slot:** 45
**Tools:** kg__search_knowledge, kg__query_graph, Bash, Read, Grep, Glob, Write
**Gates:** (none)
**MaxExtraTokens:** 10000
**Optional:** true

---

## GitHub Issue Triage

Analyzes open GitHub issues using issue details, comments, linked PRs, and user-provided context to generate root cause hypotheses and identify missing data for confirmation.

**Use when:** User asks to triage open bugs, review issue backlog, or analyze specific GitHub issues.

---

## Phase 1: Collect Inputs

### 1.1 Issue Source (Required)

Ask the user for ONE of these:

**Option A — Label/Milestone Filter:**
```
Repository: github.com/<owner>/<repo>
Filter: label="needs-investigation" AND milestone="v1.3.0"
```

**Option B — Individual Issue URLs:**
```
Issue URLs (max 10 per session):
  https://github.com/A2osX/A2osX/issues/123
  https://github.com/A2osX/A2osX/issues/145
```

### 1.2 Triaging Context (Required)

**This is critical for quality analysis.** Ask the user:

```
What context should I use when triaging?

Provide any combination of:
1. Code repositories — Local paths or GitHub URLs to inspect
2. Architecture docs — Paths to design docs, ADRs, system diagrams
3. Product knowledge — Brief system description and component relationships
4. Known issues — Recent deployments, common failure modes, areas of concern
5. Related analysis — Paths to previous triage reports or bug pattern files

The more context you provide, the better the root cause hypotheses.
```

Store the context for use in Phase 3.

### 1.3 Output File (Optional)

```
Where to save the triage report?
Default: docs/issue-triage/triage-<YYYY-MM-DD>.md
```

---

## Phase 2: Fetch Issue Details

### 2.1 Extract Issue Numbers

**From Label/Milestone Filter:**
```bash
# List open issues with filters
gh issue list \
  --repo $OWNER/$REPO \
  --state open \
  --label "needs-investigation" \
  --milestone "$MILESTONE" \
  --limit 100 \
  --json number,title,labels \
  --jq '.[] | .number'
```

**From Individual URLs:**
```bash
# Parse issue numbers from URLs
echo "$URLS" | grep -oE 'issues/([0-9]+)' | cut -d/ -f2
```

### 2.2 Fetch Full Details for Each Issue

For each issue:

**A. Core Details:**
```bash
gh issue view $ISSUE_NUM \
  --repo $OWNER/$REPO \
  --json number,title,body,state,labels,assignee,createdAt,updatedAt
```

**B. Comments:**
```bash
gh issue view $ISSUE_NUM \
  --repo $OWNER/$REPO \
  --json comments \
  --jq '.comments[] | {author: .author.login, body: .body, created: .createdAt}'
```

**C. Timeline (for linked PRs/commits):**
```bash
# Extract linked PRs and commits from issue body/comments
gh issue view $ISSUE_NUM --repo $OWNER/$REPO --json body,comments \
  | jq -r '[.body, .comments[].body] | join("\n")' \
  | grep -oE '(#[0-9]+|[0-9a-f]{7,40}|github\.com/.*/pull/[0-9]+)'
```

**D. Labels Analysis:**
Extract signal from labels:
- `critical`, `high-priority` → Severity
- `regression`, `production` → Context
- `needs-repro`, `waiting-for-info` → Missing data

Print summary of each issue as fetched to confirm data loaded.

---

## Phase 3: Analyze Each Issue

**CRITICAL: Process issues ONE AT A TIME. Analyze → Write → Next issue.**

For each issue:

### 3.1 Read the Issue

Extract from fetched data:
- **Symptom:** What is the reported problem?
- **Expected vs Actual:** What should happen vs what does happen?
- **Specific Data:** User IDs, identifiers, error messages, timestamps
- **Scope:** Single occurrence or recurring? Specific to user/environment?
- **Attachments:** What do filenames suggest? (screenshots, logs, data dumps)
- **Reporter Context:** Is reporter a user, developer, or automated system?

### 3.2 Consult User-Provided Context

Using context from Phase 1.3:

**1. Query KG for Known Patterns (if KG tools available)**
```javascript
// Search for patterns matching symptoms
kg__search_knowledge({
  query: "pattern: <symptom-description>"
})

// Example: If issue mentions "crash on startup"
kg__search_knowledge({
  query: "pattern: initialization crash"
})
```

**2. Inspect Code (if repos provided)**
```bash
# Search for error message in code
grep -r "<error-message-text>" src/ --include="*.{c,h,asm,ts,swift}"

# Find related functions/files
find . -name "*<feature-name>*" -type f
```

**3. Read Architecture Docs (if provided)**
Check relevant sections for:
- System design and component boundaries
- Data flows and integration points
- Known limitations or constraints

**4. Check Previous Triage Reports**
Look for similar symptoms in past reports.

### 3.3 Generate Root Cause Hypotheses

Produce **up to 2 hypotheses** ranked by likelihood.

**Confidence Levels:**
- **HIGH (80-95%):** Code evidence + issue details strongly support this
- **MEDIUM (50-80%):** Patterns or partial evidence support this; needs confirmation
- **LOW (20-50%):** Possible but requires significant investigation

**For each hypothesis include:**

```markdown
#### Hypothesis 1 (Most Likely): <Name>
**Confidence: HIGH / MEDIUM / LOW**

**What might be wrong:** <specific description, not vague>

**Evidence:**
- From issue: <specific details from issue body/comments>
- From KG: <matching patterns found>
- From code: <file:line references if repos were inspected>

**What this explains:** <which symptoms it accounts for>

**What this doesn't explain:** <gaps or contradictions>

**How to test:** <specific steps to confirm or rule out>
```

**If only one plausible cause exists**, provide just one hypothesis. Don't fabricate alternatives.

### 3.4 Identify Missing Data

For each issue, generate a **"Data Needed for Confirmation"** section.

Draw from these categories as relevant:

| Data Type | When to Request | Example |
|-----------|----------------|---------|
| **Logs** | Backend errors, crashes, timeouts | "Application logs with timestamps around <time>" |
| **Stack Traces** | Exceptions, crashes | "Full stack trace from error" |
| **Reproduction Steps** | Unclear or intermittent issues | "Exact steps starting from clean state" |
| **Environment** | Environment-specific behavior | "OS version, architecture, build configuration" |
| **Screenshots/Video** | UI issues, visual bugs | "Screenshot of error state" |
| **Data Samples** | Data-dependent issues | "Example input that triggers the issue" |
| **Configuration** | Config-dependent behavior | "Contents of config file or environment variables" |
| **Related Issues** | Pattern detection | "Links to similar issues" |
| **Performance Metrics** | Slowness, hangs | "Profiler output or timing measurements" |

**Prioritize requests.** Mark 1-2 items as **HIGHEST PRIORITY** — the data most likely to confirm or rule out the top hypothesis.

---

## Phase 4: Write the Report

### 4.1 Write Incrementally

After analyzing each issue, **append it to the output file immediately**. Do not batch.

This ensures progress is saved if task times out.

### 4.2 Report Structure

```markdown
# GitHub Issue Triage Report

**Date:** <today>
**Repository:** <owner>/<repo>
**Source:** <filter description or "Individual issues">
**Total Issues:** <count>
**Context Used:** <brief summary of context provided>

---

## Executive Summary

| Issue | Priority | Title | Top Hypothesis | Confidence | Known Patterns |
|-------|----------|-------|----------------|------------|----------------|
| [#123](https://github.com/<owner>/<repo>/issues/123) | High | ... | ... | HIGH | Buffer Overflow (KG) |
| [#145](https://github.com/<owner>/<repo>/issues/145) | Med | ... | ... | MEDIUM | — |

---

## Detailed Issue Analysis

---

### [#123](https://github.com/<owner>/<repo>/issues/123): <Issue Title>

**Priority:** <from labels> | **Status:** Open | **Assignee:** <name> | **Created:** <date>

#### Bug Summary
<2-3 sentence summary of reported problem>

#### Evidence Reviewed
- ✅ Reviewed: Issue description, 5 comments, 2 linked PRs
- ❌ Unable to access: Attached log file (requires download)
- 📝 Evidence gaps: No reproduction steps provided

#### Context Analysis

**From KG:**
- Found matching pattern: "Buffer Overflow in String Copy" (HIGH severity)
- 3 historical instances in this repo

**From Code Inspection:** (if repos were provided)
- ✅ `src/kernel/memory.asm:156` — String copy routine lacks bounds check
- ⚠️  `src/kernel/init.asm:44` — Calls string copy with user input

---

#### Hypothesis 1 (Most Likely): Buffer Overflow in Boot Parameter Parsing

**Confidence: HIGH**

**What might be wrong:**
The boot parameter parser in `src/kernel/init.asm:44` calls the string copy routine without validating parameter length. Long boot parameters overflow the fixed-size buffer.

**Evidence:**
- Issue reports crash when boot parameter exceeds 32 chars
- Code shows 32-byte buffer allocated at `src/kernel/memory.asm:20`
- Copy routine at `:156` has no length check
- KG shows this pattern caused 3 previous bugs in this codebase

**What this explains:**
- Why crash only occurs with long parameters
- Why it's a kernel panic (buffer overflow corrupts stack)
- Why it's reproducible with specific input length

**What this doesn't explain:**
- Why only some users see it (may depend on bootloader)

**How to test:**
Boot with parameter exactly 33 chars and observe crash.

---

#### Hypothesis 2: Null Terminator Handling

**Confidence: MEDIUM**

<Only if a second plausible cause exists>

---

#### Data Needed for Confirmation

**🔴 HIGHEST PRIORITY:**
- [ ] Exact boot parameter string that triggers the crash
- [ ] Memory dump showing buffer contents at crash

**Additional helpful data:**
- [ ] Bootloader version and config
- [ ] Serial console log from boot to crash
- [ ] Kernel build configuration

---

<Repeat for each issue>

---

## Patterns Observed

<Summary of recurring patterns found across multiple issues>

| Pattern | Issues | Source |
|---------|--------|--------|
| Buffer Overflow | #123, #145 | KG + Code |
| Missing Null Check | #134 | Code inspection |

---

*Generated by ai-pack github_issue_triager skill*
```

### 4.3 Report Rules

- Issue IDs must be clickable hyperlinks
- Summaries must be self-contained (readable without product knowledge)
- Present findings as **hypotheses**, never as confirmed root causes
- Every hypothesis needs a confidence level
- Missing data requests must be specific and prioritized
- Include file:line references if code was inspected
- Note any evidence you couldn't access and flag reduced confidence

---

## Required Workflow

**Process issues one at a time:**
1. Read issue details + comments
2. Query KG for matching patterns
3. Inspect code (if repos provided)
4. Generate hypotheses (up to 2)
5. Identify missing data
6. **Write to report file** ← Do this immediately
7. Next issue

**Do NOT batch all analysis then write at once.**

---

## Example Usage

```
User: "Triage open bugs labeled 'needs-investigation' in A2osX/A2osX"
Context provided:
  - Local repo: ~/Projects/A2osX/A2osX
  - System knowledge: "Apple II operating system, assembly + C"

Agent:
1. Fetches 8 open issues with label
2. For each issue:
   - Reads description, comments, timeline
   - Queries KG: finds 2 matching patterns
   - Inspects code: finds relevant functions
   - Generates 1-2 hypotheses with confidence
   - Identifies missing data (logs, repro steps)
   - Appends to docs/issue-triage/triage-2026-03-12.md
3. Final report: 8 issues analyzed, 12 hypotheses, 5 known patterns matched
```

---

## Integration with Other Skills

- **Uses `kg_reader`** — Queries known patterns to inform hypotheses
- **Complements `github_bug_analyzer`** — Triages open issues vs analyzes closed bugs
- **Used by roles:** engineer, inspector, reviewer
- **Prerequisites:** Works best when bug patterns have been mined by `github_bug_analyzer`

---

## Notes

- **Context is critical:** Analysis quality is directly proportional to context provided
- **Read ALL comments:** Developers often investigate in comments
- **Don't assume reporter's hypothesis is correct:** Analyze independently
- **Be honest about confidence:** Better to say "MEDIUM confidence, need logs" than guess
- **Incremental writing:** Saves progress if task times out
