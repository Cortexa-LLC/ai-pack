# Personal Knowledge Reader (UPK)
<!-- skills/upk_reader.skill.md -->

**Version:** 1.0
**InjectAt:** role_context
**Slot:** 15
**Tools:** mcp__upk__search_knowledge
**Gates:** knowledge-first
**MaxExtraTokens:** 5000
**Optional:** true

---

## Personal Knowledge — Check Before Searching

Your personal knowledge base contains learnings and insights from past conversations and work across ALL projects. **Search it before attempting cumbersome file searches or web searches** for patterns you may have already encountered.

### ⚠️ MANDATORY: This gate is ENFORCED

**FORBIDDEN operations without UPK check:**
- ❌ WebSearch before upk__search_knowledge (for "how to" questions)
- ❌ Grep across projects before checking if you've solved this before
- ❌ Re-deriving solutions you've already learned

**REQUIRED workflow:**
1. MUST search personal knowledge before external searches
2. ONLY IF knowledge returns nothing → use external search
3. MUST record new learnings back to personal knowledge

This is enforced by the **[Knowledge-First Gate](../gates/15-knowledge-first.md)**.

---

## When to Use Personal Knowledge

| Question Type | Use UPK | Instead of |
|---------------|---------|------------|
| "How have I solved X before?" | `upk__search_knowledge` | WebSearch / grep old projects |
| "What did I learn about Y?" | `upk__search_knowledge` | Re-researching from scratch |
| "Who did I talk to about Z?" | `upk__search_knowledge` | Searching chat logs manually |
| "What patterns work for A?" | `upk__search_knowledge` | Re-deriving from first principles |

### UPK vs Project KG vs Organizational Knowledge

```
Personal Knowledge (UPK):
  - Cross-project learnings
  - Conversation history
  - Design patterns you've discovered
  - Solutions to recurring problems
  ✅ USE: upk__search_knowledge

Project Code Knowledge (KG):
  - Current project's code/architecture
  - Project-specific decisions
  - Component relationships
  ✅ USE: kg__search_knowledge (see kg_reader skill)

Organizational Knowledge:
  - Team structures
  - Planning artifacts (M0/M1/M2)
  - Work items (Jira, GitHub)
  ✅ USE: org-specific MCP tools (compass, wiki, jira)
```

---

## Tool

- **`upk__search_knowledge`** `{query: string, limit?: number}` — Search across personal learnings and conversations. Returns relevant insights from ALL past work, not just current project.

---

## Required Workflow

### Before External Search

```
BEFORE WebSearch OR cross-project grep THEN
  result = upk__search_knowledge({query: "<problem or pattern>"})
  
  IF result contains answer THEN
    use learned solution
    adapt to current context
    SKIP external search
  ELSE
    proceed to external search
    record new findings with upk__add_learning (see upk_writer skill)
  END IF
END BEFORE
```

### Example Queries

```
✅ GOOD queries:
upk__search_knowledge({query: "rate limiting implementation"})
  → Find how you've implemented rate limiting before

upk__search_knowledge({query: "GraphQL federation patterns"})
  → Recall federation patterns you've used

upk__search_knowledge({query: "conversation with Alice about auth"})
  → Find specific past discussions

upk__search_knowledge({query: "React performance optimization"})
  → Retrieve performance lessons learned

❌ POOR queries:
upk__search_knowledge({query: "function"})
  → Too generic, use kg__search_knowledge for code

upk__search_knowledge({query: "team org chart"})
  → Use organizational MCP, not personal knowledge

upk__search_knowledge({query: "current project architecture"})
  → Use kg__search_knowledge for project-specific
```

---

## Knowledge-First Workflow

```
QUESTION from user: "How should I handle webhook retries?"

STEP 1: Check personal knowledge (MANDATORY)
  upk__search_knowledge({query: "webhook retry patterns"})
  → Result: "Learned: exponential backoff with max 5 retries,
             idempotency keys required, from project-x"

STEP 2: Apply to current context
  "Based on past learnings, I recommend exponential backoff..."
  Adapt pattern to current project

STEP 3: Record new variations (if different)
  upk__add_learning({
    content: "Webhook retries in Go: used exponential backoff...",
    source: "ai-pack implementation"
  })

EFFICIENCY GAIN:
  ❌ Without UPK: WebSearch (2000 tokens) + read docs (5000 tokens) = 7000 tokens
  ✅ With UPK: upk__search_knowledge (300 tokens) = 300 tokens
  SAVINGS: 95% token reduction
```

---

## Integration with Other Knowledge Systems

### Hybrid Knowledge Strategy

For comprehensive research, combine knowledge sources:

```
RESEARCH TASK: "Implement authentication system"

STEP 1: Personal patterns (UPK)
  upk__search_knowledge({query: "authentication implementation"})
  → "I've used JWT + refresh tokens, learned: rotate secrets, short expiry"

STEP 2: Project context (KG)
  kg__search_knowledge({query: "existing auth"})
  → "Current project has OAuth2 in pkg/oauth, but no JWT support"

STEP 3: Organizational standards (Org MCP)
  compass__query OR wiki__search({query: "auth standards"})
  → "Company requires SSO integration via Okta"

STEP 4: Synthesize approach
  "Implement JWT with Okta SSO, following pattern from personal experience,
   integrate with existing OAuth2 in pkg/oauth per company standards"

STEP 5: Record new synthesis (UPK)
  upk__add_learning({
    content: "Integrated Okta SSO with JWT refresh tokens...",
    source: "ai-pack auth implementation"
  })
```

---

## Compliance

**Token efficiency target:**
- Personal knowledge check MUST occur before external searches for "how to" questions
- Minimum 60% of "how to" queries should hit UPK before external search
- All new learnings MUST be recorded back (see upk_writer skill)

---

## See Also

- **[UPK Writer](upk_writer.skill.md)** - Recording learnings and conversations
- **[KG Reader](kg_reader.skill.md)** - Project code knowledge
- **[Knowledge-First Gate](../gates/15-knowledge-first.md)** - Enforcement policy
