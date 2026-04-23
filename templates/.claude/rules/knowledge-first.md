---
paths: **/*
---

# Knowledge-First Enforcement

**⚠️ MANDATORY: Search knowledge BEFORE file operations**

This rule is **ENFORCED** by the Knowledge-First Gate.

## Three-Tier Knowledge Architecture

```
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│   Project    │  │   Personal   │  │    Org       │
│   KG (kg)    │  │   KG (upk)   │  │   Systems    │
└──────────────┘  └──────────────┘  └──────────────┘
      ↓                 ↓                  ↓
 This project      All projects      Organization
 code/arch         learnings          data/planning
```

## Required Workflow

### BEFORE File Operations

```
BEFORE Grep OR Glob OR Read THEN
  STEP 1: Determine scope
    - Project code/architecture? → kg__search_knowledge
    - Cross-project patterns? → upk__search_knowledge
    - Org/team info? → org MCP tools

  STEP 2: Search knowledge
    result = search_knowledge({query: "..."})
  
  STEP 3: Evaluate
    IF result answers question THEN
      use result, SKIP file search
    ELSE
      proceed to file search
      RECORD findings back to knowledge
    END IF
END BEFORE
```

### Specific Requirements

**Before `grep`:**
```javascript
// MANDATORY
kg__search_knowledge({query: "pattern or component"})

// Only if empty
grep -r "pattern"
```

**Before `Read(file)`:**
```javascript
// MANDATORY
kg__get_file_context({file: "path/to/file.go"})

// Then read only relevant sections
Read(file, offset, limit)
```

**Before WebSearch for "how to":**
```javascript
// MANDATORY
upk__search_knowledge({query: "how to X"})

// Only if empty
WebSearch({query: "how to X"})
```

## Write-Back Requirement

**AFTER finding information:**

```
IF project-specific (code/arch) THEN
  kg__add_entity OR kg__add_observation
ELSE IF cross-project learning THEN
  upk__add_learning
ELSE IF conversation context THEN
  upk__add_conversation
END IF
```

## Token Efficiency

```
WITHOUT knowledge-first:
  grep + read multiple files = 50,000-100,000 tokens

WITH knowledge-first:
  kg__search_knowledge + targeted read = 3,000-5,000 tokens

SAVINGS: 90-97% reduction
```

## Quick Decision Tree

```
QUESTION: "Where is X in THIS project?"
  → kg__search_knowledge({query: "X"})

QUESTION: "How have I solved Y before?"
  → upk__search_knowledge({query: "Y"})

QUESTION: "Who owns service Z?"
  → org MCP tools (compass, wiki, jira)

QUESTION: "What does function F do?"
  → kg__get_file_context({file: "path/to/file"})
  → Then Read specific function
```

## Violations

**❌ BLOCKED without knowledge check:**
- Direct grep without kg__search_knowledge
- Direct Read without kg__get_file_context
- WebSearch without upk__search_knowledge (for patterns)

**✅ EXCEPTIONS:**
- User provides explicit file path: "Read src/main.go" → direct Read allowed
- Following stack trace: error points to file:line → direct Read allowed
- Knowledge system unavailable: timeout → fallback to file search

## Reference

- **Full Guide:** `.ai-pack/docs/KNOWLEDGE-SYSTEMS.md`
- **Quick Ref:** `.ai-pack/docs/KNOWLEDGE-QUICK-REF.md`
- **Gate:** `.ai-pack/gates/15-knowledge-first.md`
- **Skills:** 
  - `.ai-pack/skills/upk_reader.skill.md`
  - `.ai-pack/skills/upk_writer.skill.md`
  - `.ai-pack/skills/kg_reader.skill.md`
  - `.ai-pack/skills/kg_writer.skill.md`
