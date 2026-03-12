# GitHub Bug Analyzer
<!-- skills/github_bug_analyzer.skill.md -->

**Version:** 1.0
**InjectAt:** role_context
**Slot:** 40
**Tools:** kg__add_entity, kg__add_observation, kg__link_entities, kg__search_knowledge, kg__query_graph, Bash, Read, Write
**Gates:** (none)
**MaxExtraTokens:** 15000
**Optional:** true

---

## GitHub Bug Analysis Pipeline

Automates end-to-end bug analysis: GitHub Issues → PR discovery → diff analysis → pattern mining → KG storage → structured reports.

**Use when:** User asks to analyze bugs from a GitHub milestone, label, or filter.

---

## Phase 1: Collect Inputs

### 1.1 Required Inputs

Ask the user for the following (skip if already provided):

```
1. Repository    — e.g., "github.com/A2osX/A2osX" or "A2osX/A2osX"
2. Filter Scope  — Choose ONE:
   - Milestone name (e.g., "v1.2.0")
   - Label filter (e.g., "bug" AND "fixed")
   - Issue number range (e.g., "#100-150")
```

### 1.2 Optional Inputs (show defaults)

```
3. Language/Stack focus — Which languages to analyze?
   Default: ALL (auto-detect from PR diffs)
   Options: Assembly, C, TypeScript, Java, Kotlin, Swift, etc.

4. Output file — Where to save the report?
   Default: docs/bug-analysis/<repo>-<milestone>-bugs.md
   Example: docs/bug-analysis/A2osX-v1.2.0-bugs.md
```

---

## Phase 2: Fetch Closed Issues

Use `gh` CLI to fetch bugs matching the filter:

```bash
# Example: Fetch by milestone and label
gh issue list \
  --repo A2osX/A2osX \
  --state closed \
  --label bug \
  --milestone "v1.2.0" \
  --limit 1000 \
  --json number,title,labels,closedAt \
  --jq '.[] | [.number, .title] | @tsv'

# Or by issue number range
gh issue list \
  --repo A2osX/A2osX \
  --state closed \
  --limit 1000 \
  --json number,title,labels \
  --jq '.[] | select(.number >= 100 and .number <= 150)'
```

Parse JSON output and print total count + first 10 issue titles to confirm.

---

## Phase 3: Extract Linked PRs

For each issue, extract PRs that closed it:

### Method A: Using gh CLI (primary)
```bash
# Get issue details including linked PRs
gh issue view $ISSUE_NUM \
  --repo $OWNER/$REPO \
  --json number,title,body,closedBy,comments \
  --jq '{
    number,
    title,
    closedBy: .closedBy.number,
    prs: [.body, .comments[].body] | join("\n") | scan("#([0-9]+)") | .[0]
  }'
```

### Method B: Scan Issue Body/Comments (fallback)
```bash
# Extract PR numbers from issue body and comments
gh issue view $ISSUE_NUM --repo $OWNER/$REPO --json body,comments \
  | jq -r '[.body, .comments[].body] | join("\n")' \
  | grep -oE '#[0-9]+|github\.com/.*/pull/[0-9]+' \
  | grep -oE '[0-9]+' \
  | sort -u
```

Collect all unique PRs per issue. Issues with no linked PRs are flagged but not excluded.

---

## Phase 4: Classify Language/Stack

For each unique repo found, spot-check one PR diff to detect language:

```bash
# Get PR diff and extract file extensions
gh pr diff $PR_NUM --repo $OWNER/$REPO 2>/dev/null \
  | grep -E '^(diff --git|\+\+\+)' \
  | grep -oE '\.(s|asm|a65|c|h|swift|kt|java|ts|tsx|js|jsx|go|py|rs)$' \
  | sort | uniq -c | sort -rn
```

**Language detection rules:**

| File Pattern | Language/Stack |
|--------------|----------------|
| `*.s`, `*.asm`, `*.a65` | Assembly |
| `*.c`, `*.h` | C |
| `*.swift` | Swift (iOS) |
| `*.kt`, `*.java` + `android/` paths | Android |
| `*.ts`, `*.tsx`, `*.js`, `*.jsx` | TypeScript/JavaScript |
| `*.go` | Go |
| `*.py` | Python |
| `*.rs` | Rust |

If user specified language filter, only analyze PRs matching that language.

---

## Phase 5: Fetch PR Diffs and Analyze

For each bug with at least one PR in target language(s):

### 5.1 Fetch PR Details
```bash
# Get PR metadata
gh pr view $PR_NUM \
  --repo $OWNER/$REPO \
  --json number,title,body,state,mergedAt,files

# Get PR diff (first 8000 chars for analysis)
gh pr diff $PR_NUM --repo $OWNER/$REPO | head -c 8000
```

### 5.2 Analyze Root Cause

For each bug, synthesize a 2-3 sentence summary:
1. **What was wrong** — The specific anti-pattern or missing code
2. **Why it caused the symptom** — The mechanism
3. **What the fix did** — The code change in plain terms

**Focus on what is visible in the diff** — exclude domain-specific reasoning requiring external knowledge.

### 5.3 Exclude Non-Signal PRs

Skip PRs that are:
- Pure logging additions (no logic change)
- Dependency version bumps only
- Config file changes only

Note excluded PRs in a footnote for transparency.

---

## Phase 6: Mine Patterns and Store in KG

For each bug, evaluate if the root cause is **diff-detectable** (no domain knowledge required).

### 6.1 Check if Pattern Exists

**If KG tools available:**
```javascript
kg__search_knowledge({
  query: "pattern: <root-cause-description>"
})
```

### 6.2 If Pattern is New, Create It

**If KG tools available:**
```javascript
kg__add_entity({
  name: "Buffer Overflow in Assembly String Copy",
  type: "pattern",
  properties: {
    category: "Memory Safety",
    language: "Assembly",
    severity: "HIGH",
    detection_rule: "String copy without length check"
  }
})
```

### 6.3 Record Pattern Instance

**If KG tools available:**
```javascript
// Link bug to pattern
kg__link_entities({
  from: bug_entity_id,
  relation: "EXHIBITS_PATTERN",
  to: pattern_id
})

// Link affected file to pattern
kg__link_entities({
  from: file_entity_id,
  relation: "CONTAINS_PATTERN",
  to: pattern_id
})

// Add observation with context
kg__add_observation({
  entity_id: pattern_id,
  content: `[INSTANCE] GitHub Issue #${issue_num}: ${symptom}
File: ${file_path}:${line_num}
Fix: ${fix_description}
PR: ${pr_url}`
})
```

**If KG tools NOT available:**
Store patterns in markdown for manual review later.

---

## Phase 7: Group Bugs by Domain

After analyzing all bugs, group them by feature area or error type:

**Suggested groupings (adapt to actual bugs):**

| Language | Suggested Groups |
|----------|------------------|
| **Assembly/C** | Memory Safety, Null/Bounds Checking, Register Management, Logic Errors |
| **Swift/iOS** | UI Lifecycle, Memory Management (retain cycles), Async/Concurrency, API Integration |
| **Android** | Lifecycle Issues, Memory Leaks, Threading, UI State |
| **TypeScript/React** | State Management, Hooks Dependencies, Type Safety, Async Handling |
| **Java/Kotlin** | Null Safety, Exception Handling, Concurrency, API Contracts |

Use "Other" as a catch-all. Aim for 3-7 groups per language.

---

## Phase 8: Query KG for Report Data

**If KG tools available**, query for pattern statistics:

```javascript
// Pattern frequency
kg__query_graph({
  query: `
    MATCH (p:Pattern)-[:APPEARS_IN]->(f:File)
    WHERE f.project = '<project-name>'
    RETURN p.name, p.category, p.severity, count(f) as instances
    ORDER BY instances DESC
  `
})

// Top affected files
kg__query_graph({
  query: `
    MATCH (f:File)-[:CONTAINS_PATTERN]->(p:Pattern)
    WHERE f.project = '<project-name>'
    RETURN f.path, collect(p.name) as patterns, count(p) as pattern_count
    ORDER BY pattern_count DESC
    LIMIT 10
  `
})
```

**If KG tools NOT available**, compute statistics from in-memory analysis.

---

## Phase 9: Generate Markdown Report

Write the report to the output path (default: `docs/bug-analysis/<repo>-<milestone>-bugs.md`):

```markdown
# <Repo> <Milestone> Bug Analysis

**Date:** <today>
**Source:** <GitHub filter description>
**Total Bugs Analyzed:** <count>
**Languages:** <detected languages>

---

## Pattern Summary

| Pattern | Category | Severity | Instances |
|---------|----------|----------|-----------|
| Buffer Overflow in String Copy | Memory Safety | HIGH | 5 |
| Missing Null Check | Null Safety | MEDIUM | 3 |
...

---

## Diff-Detectable Patterns

**Only include patterns detectable from code alone (no domain knowledge required)**

**[#45](https://github.com/<owner>/<repo>/issues/45)** — Buffer overflow in string copy routine (Missing Null Check pattern)

**[#67](https://github.com/<owner>/<repo>/issues/67)** — Use-after-free in memory allocator (Memory Safety pattern)

---

## Full Bug List

### Memory Safety (5 bugs)

| Bug ID | Title | Summary |
|--------|-------|---------|
| [#45](https://github.com/A2osX/A2osX/issues/45) | Kernel panic on long filename | String copy routine did not bounds-check destination buffer, causing overflow into adjacent memory. Fix added length validation before copy. |
| [#67](https://github.com/A2osX/A2osX/issues/67) | Crash in memory allocator | Free routine did not null the pointer after deallocation, leading to use-after-free. Fix sets pointer to null after free. |

### Null/Bounds Checking (3 bugs)

...

---

## Top Affected Files

| File | Patterns | Count |
|------|----------|-------|
| src/kernel/memory.asm | Buffer Overflow, Use-After-Free | 3 |
| src/drivers/serial.c | Null Pointer Dereference | 2 |

---

## Excluded PRs

<count> PRs excluded (logging-only, dependency bumps, config-only):
- PR #123 — Added CAL logging (no logic change)
- PR #145 — Dependency version bump

---

*Generated by ai-pack github_bug_analyzer skill*
```

**Report rules:**
- Bug IDs must be clickable hyperlinks
- Summaries must be self-contained (readable without knowing the codebase)
- Group by domain/category for readability
- Include pattern statistics from KG
- Note excluded PRs for transparency

---

## Phase 10: Commit Report (Optional)

If user wants to commit the report:

```bash
git add docs/bug-analysis/<filename>.md
git commit -m "Add bug analysis for <milestone>

Analyzed <count> bugs, identified <pattern-count> patterns.

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Required Workflow

**Process bugs incrementally** — analyze → store in KG → next bug.

Do NOT batch all analysis then write at once. If the task times out, KG observations are preserved.

---

## Example Usage

```
User: "Analyze all bugs fixed in A2osX v1.2.0"

Agent:
1. Fetches closed issues with milestone="v1.2.0", label="bug"
2. Extracts linked PRs from issue timelines
3. Detects languages: Assembly (80%), C (20%)
4. Analyzes PR diffs for root causes
5. Stores 8 new patterns in KG, matches 5 existing patterns
6. Groups bugs: Memory Safety (5), Null Checking (3), Logic (2)
7. Queries KG for pattern stats
8. Generates: docs/bug-analysis/A2osX-v1.2.0-bugs.md
9. Commits report to git
```

---

## Integration with Other Skills

- **Uses `kg_writer`** — Stores patterns, observations, relationships
- **Uses `kg_reader`** — Checks for existing patterns before creating new ones
- **Used by roles:** engineer, reviewer, inspector
- **Complements:** github_issue_triager (triage open bugs vs analyze closed bugs)

---

## Notes

- **GitHub API rate limits:** Be mindful when analyzing 100+ bugs. Use `per_page: 100` and pagination.
- **Large diffs:** Use `max_tokens` parameter to limit diff size. Focus on changed logic, not boilerplate.
- **Multi-repo support:** If bugs span multiple repos, process each repo separately then combine reports.
