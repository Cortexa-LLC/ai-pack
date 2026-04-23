# Knowledge-First Enforcement — Rollout Guide

**Version:** 1.0  
**Date:** 2026-04-22  
**Status:** Ready for Phase 2 (Enforced)

---

## Changes Overview

### New Components

1. **Gate:** `gates/15-knowledge-first.md`
   - Enforces knowledge check before file operations
   - Blocks grep/glob/read without prior knowledge search
   - Requires write-back of findings

2. **Skills:**
   - `skills/upk_reader.skill.md` (Slot 15) - Personal knowledge search
   - `skills/upk_writer.skill.md` (Slot 16) - Recording learnings/conversations
   - `skills/kg_reader.skill.md` (Updated v1.2) - Project knowledge with enforcement
   - `skills/kg_writer.skill.md` (Updated v1.2) - Project recording with enforcement

3. **Documentation:**
   - `docs/KNOWLEDGE-SYSTEMS.md` - Full architecture guide
   - `docs/KNOWLEDGE-QUICK-REF.md` - Quick reference card
   - `docs/KNOWLEDGE-ROLLOUT.md` - This rollout guide

---

## Required Role Updates

All roles need UPK skills added to their skill lists.

### Update Pattern

**Before:**
```yaml
**Skills:** general, kg_reader, kg_writer
```

**After:**
```yaml
**Skills:** general, upk_reader, upk_writer, kg_reader, kg_writer
```

**Rationale:** Load order is upk (slot 15-16) → kg (slot 20-25) → general (slot 50+)

### Roles to Update

| Role | Current Skills | Updated Skills |
|------|---------------|----------------|
| orchestrator | general, kg_reader, kg_writer | general, upk_reader, upk_writer, kg_reader, kg_writer |
| orchestrator-chat | general, kg_reader, kg_writer | general, upk_reader, upk_writer, kg_reader, kg_writer |
| engineer | general, kg_reader, kg_writer, tdd, code_review, ... | general, upk_reader, upk_writer, kg_reader, kg_writer, tdd, code_review, ... |
| architect | general, kg_reader, arch_review, kg_writer, ... | general, upk_reader, upk_writer, kg_reader, arch_review, kg_writer, ... |
| reviewer | general, kg_reader, code_review, kg_writer, ... | general, upk_reader, upk_writer, kg_reader, code_review, kg_writer, ... |
| tester | general, kg_reader, tdd, kg_writer | general, upk_reader, upk_writer, kg_reader, tdd, kg_writer |
| inspector | general, kg_reader, kg_writer, github_bug_analyzer, ... | general, upk_reader, upk_writer, kg_reader, kg_writer, github_bug_analyzer, ... |
| archaeologist | general, kg_reader, kg_writer | general, upk_reader, upk_writer, kg_reader, kg_writer |
| product-manager | general, kg_reader, kg_writer | general, upk_reader, upk_writer, kg_reader, kg_writer |
| strategist | general, kg_reader, kg_writer | general, upk_reader, upk_writer, kg_reader, kg_writer |
| designer | general, kg_reader, kg_writer, federated_graphql_designer | general, upk_reader, upk_writer, kg_reader, kg_writer, federated_graphql_designer |
| spelunker | general, kg_reader, kg_writer | general, upk_reader, upk_writer, kg_reader, kg_writer |

---

## Implementation Steps

### Step 1: Add UPK MCP Server

**Prerequisite:** UPK MCP server must be configured

Check `~/.claude/settings.json` has upk-server configured:
```json
{
  "mcpServers": {
    "upk": {
      "command": "npx",
      "args": ["-y", "@upk/mcp-server"]
    }
  }
}
```

**Test:**
```bash
# Should show upk__ tools
claude --list-tools | grep upk
```

### Step 2: Update All Roles

**Batch update script:**
```bash
#!/bin/bash
# update-roles-for-upk.sh

ROLES=(
  "orchestrator"
  "orchestrator-chat"
  "engineer"
  "architect"
  "reviewer"
  "tester"
  "inspector"
  "archaeologist"
  "product-manager"
  "strategist"
  "designer"
  "spelunker"
)

for role in "${ROLES[@]}"; do
  file="roles/${role}.md"
  
  # Replace skill line to add upk_reader, upk_writer after general
  sed -i.bak 's/\*\*Skills:\*\* general,/\*\*Skills:\*\* general, upk_reader, upk_writer,/' "$file"
  
  echo "Updated $file"
done

echo "All roles updated. Review git diff before committing."
```

**Manual verification:**
```bash
git diff roles/
# Verify each role has: general, upk_reader, upk_writer, kg_reader, kg_writer
```

### Step 3: Test Knowledge-First with One Role

**Test with Orchestrator:**
```bash
# Start agent
agent orchestrator test-task-id --stream

# Should see UPK skills loaded in system prompt
# Try knowledge search
upk__search_knowledge({query: "test"})

# Try grep without knowledge check - should get warning or block
```

**Expected behavior:**
- Skills load in order: upk_reader (15), upk_writer (16), kg_reader (20), kg_writer (25)
- Knowledge-first gate is active
- File operations prompt for knowledge check first

### Step 4: Gradual Rollout

**Week 1: Soft enforcement (warnings)**
- Gate logs violations but allows them
- Track compliance metrics
- Educate agents via warnings

**Week 2: Hard enforcement (blocks)**
- Gate blocks file ops without knowledge check
- Exceptions logged for review
- Monitor for false positives

**Week 3: Optimization**
- Review compliance data
- Adjust gate thresholds
- Tune knowledge query patterns

### Step 5: Monitor Metrics

**Track in `.claude/metrics/knowledge-first/`:**
```json
{
  "date": "2026-04-22",
  "knowledge_searches": 150,
  "file_searches": 30,
  "knowledge_first_ratio": 0.833,
  "avg_tokens_with_knowledge": 3500,
  "avg_tokens_without_knowledge": 45000,
  "token_savings_pct": 92.2,
  "violations": {
    "grep_without_kg": 5,
    "read_without_context": 3,
    "search_without_upk": 2
  }
}
```

**Target metrics:**
- Knowledge-first ratio: ≥ 80%
- Token savings: ≥ 90%
- Violations: < 5% of operations

---

## Testing Checklist

### Unit Tests (Per Role)

- [ ] Role loads upk_reader skill (slot 15)
- [ ] Role loads upk_writer skill (slot 16)
- [ ] Role loads kg_reader skill (slot 20)
- [ ] Role loads kg_writer skill (slot 25)
- [ ] Skills load in correct order
- [ ] Knowledge-first gate is active
- [ ] Tools are available (upk__*, kg__*)

### Integration Tests

- [ ] Knowledge search before grep works
- [ ] File context before read works
- [ ] Write-back after finding works
- [ ] Dual recording (kg + upk) works
- [ ] Gate blocks violations
- [ ] Gate allows exceptions (user-provided paths)

### Performance Tests

- [ ] Token usage reduced ≥ 80%
- [ ] Knowledge searches complete < 2s
- [ ] File searches only when knowledge empty
- [ ] No performance degradation from gate checks

---

## Rollback Plan

If issues arise, rollback in reverse order:

### Rollback Step 1: Disable Gate Enforcement
```bash
# Edit gates/15-knowledge-first.md
# Change enforcement level to "warning" mode
```

### Rollback Step 2: Remove UPK from Roles
```bash
# Revert role files
git checkout HEAD -- roles/

# Or use sed to remove upk skills
for role in roles/*.md; do
  sed -i.bak 's/, upk_reader, upk_writer//' "$role"
done
```

### Rollback Step 3: Archive New Skills
```bash
# Move skills to archive
mkdir -p skills/archive
mv skills/upk_*.skill.md skills/archive/
```

---

## Success Criteria

### Phase 2 Completion

- ✅ All roles have upk_reader, upk_writer skills
- ✅ Knowledge-first gate enforced
- ✅ Knowledge-first ratio ≥ 80%
- ✅ Token savings ≥ 90%
- ✅ Violations < 5%
- ✅ No agent failures due to gate
- ✅ Documentation complete
- ✅ Team trained on knowledge systems

### Phase 3 Readiness (Future)

- Intelligent knowledge routing (auto-select tier)
- Hybrid queries (knowledge + file search)
- Automatic indexing on writes
- Predictive knowledge caching
- Cross-project knowledge federation

---

## Known Issues & Mitigations

### Issue 1: Knowledge System Unavailable

**Symptom:** KG or UPK server down, agents blocked

**Mitigation:**
```
IF knowledge_search TIMES OUT THEN
  log "Knowledge system unavailable - bypassing gate"
  allow file_search
  queue findings for write-back when system recovers
END IF
```

### Issue 2: Empty Knowledge Graph

**Symptom:** New projects have no KG data, all queries return empty

**Mitigation:**
```
IF kg__search_knowledge returns empty
   AND this is first query THEN
  log "KG not indexed - running initial index"
  kg__index_project()
  retry query
END IF
```

### Issue 3: Over-Querying UPK

**Symptom:** Agents search UPK for project-specific code

**Mitigation:**
- Improve skill documentation clarity
- Add examples of good vs bad queries
- Gate can detect project-file queries and redirect to kg

---

## Communication Plan

### For Users

**Announcement:**
```
Subject: New Knowledge-First System — Faster Agents, Lower Costs

We've deployed a knowledge-first architecture that makes agents:
- 10-100x more efficient (90%+ token savings)
- Faster to answer code questions
- Better at learning from past work

What's new:
- Agents search indexed knowledge before reading files
- Cross-project learnings are preserved and reused
- Important conversations are recorded for context

No action needed from you. Agents will automatically use the new system.
```

### For Agent Developers

**Training Materials:**
- [Knowledge Systems Architecture](KNOWLEDGE-SYSTEMS.md)
- [Quick Reference Card](KNOWLEDGE-QUICK-REF.md)
- [Knowledge-First Gate](../gates/15-knowledge-first.md)

**Workshop Topics:**
1. Three-tier knowledge architecture
2. When to use kg vs upk vs org MCP
3. Writing effective knowledge queries
4. Recording high-quality learnings
5. Dual recording pattern
6. Debugging gate violations

---

## Timeline

| Week | Phase | Activities |
|------|-------|------------|
| 1 | Setup | Add UPK MCP, create skills, update roles |
| 2 | Soft Launch | Warning mode, track metrics, educate |
| 3 | Enforcement | Enable gate blocking, monitor violations |
| 4 | Optimization | Tune queries, improve documentation |
| 5+ | Steady State | Monitor metrics, iterate on patterns |

---

## Contact & Support

**Questions:** #ai-pack-support  
**Issues:** GitHub Issues  
**Docs:** `/docs/KNOWLEDGE-*.md`

---

**Approved by:** [TBD]  
**Rollout date:** 2026-04-22  
**Review date:** 2026-05-22 (30 days post-rollout)
