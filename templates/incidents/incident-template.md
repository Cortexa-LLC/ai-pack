# Production Incident: [Incident Title]

**Incident ID:** INC-YYYY-NNN
**Severity:** [SEV-1 | SEV-2 | SEV-3 | SEV-4]
**Status:** [Investigating | Mitigating | Resolved | Closed]
**Start Time:** YYYY-MM-DD HH:MM:SS UTC
**End Time:** YYYY-MM-DD HH:MM:SS UTC (if resolved)
**Duration:** [X hours Y minutes]

---

## Executive Summary

[1-2 sentence summary of what happened, impact, and resolution]

---

## Impact

### User Impact

**Affected Users:** [Number/Percentage of users]
**Affected Features:** [List of features unavailable or degraded]
**Geographic Impact:** [Regions affected]

**Impact Level:**
- [ ] Total Outage (service completely down)
- [ ] Partial Outage (some features unavailable)
- [ ] Performance Degradation (slow but functional)
- [ ] Data Inconsistency (incorrect data displayed/stored)

### Business Impact

**Revenue Impact:** [Estimated loss or "None"]
**SLA Breach:** [Yes/No - Which SLAs]
**Customer Escalations:** [Number of complaints/tickets]
**Reputation Impact:** [Public visibility, media coverage]

---

## Timeline

All times in UTC. Use 24-hour format.

| Time (UTC) | Event | Action Taken | Owner |
|------------|-------|--------------|-------|
| HH:MM | Incident detected | [How detected] | [Name] |
| HH:MM | Initial investigation started | [Action] | [Name] |
| HH:MM | Root cause identified | [Cause] | [Name] |
| HH:MM | Mitigation started | [Action] | [Name] |
| HH:MM | Service partially restored | [Action] | [Name] |
| HH:MM | Service fully restored | [Action] | [Name] |
| HH:MM | Incident closed | [Final action] | [Name] |

### Key Milestones

- **Time to Detect (TTD):** [X minutes from start to detection]
- **Time to Mitigate (TTM):** [X minutes from detection to mitigation start]
- **Time to Resolve (TTR):** [X minutes from detection to full resolution]
- **Total Duration:** [X hours Y minutes]

---

## Incident Detection

### How Was It Detected?

- [ ] Automated Monitoring Alert
- [ ] Customer Report
- [ ] Internal Team Discovery
- [ ] Third-party Report
- [ ] Scheduled Check

**Detection Method:** [Specific alert, customer complaint, manual check, etc.]

**First Indicator:**
```
[Log entry, metric spike, error message that first indicated the issue]
```

**Monitoring Gaps:**
[What should have alerted us earlier but didn't?]

---

## Root Cause Analysis

### What Happened?

[Detailed technical explanation of the root cause]

**Primary Cause:**
- **Category:** [Code Bug | Configuration Error | Infrastructure Failure | Dependency Failure | Human Error | External Factor]
- **Component:** [Service/system that failed]
- **Specific Issue:** [Exact technical cause]

**Contributing Factors:**
1. [Factor 1 that made this worse or enabled it]
2. [Factor 2]
3. [Factor 3]

### Why Did It Happen?

**5 Whys Analysis:**
1. Why did [symptom] occur? [Answer]
2. Why did [answer 1] happen? [Answer]
3. Why did [answer 2] happen? [Answer]
4. Why did [answer 3] happen? [Answer]
5. Why did [answer 4] happen? [Root cause]

### Technical Details

**System State Before Incident:**
```
[Description of normal operating state]
```

**Triggering Event:**
```
[What specific action/event triggered the failure]
```

**Failure Mode:**
```
[How the system failed - cascade, timeout, resource exhaustion, etc.]
```

**Code/Configuration:**
```language
// Problematic code or configuration
[Code snippet or config that caused the issue]
```

**Logs:**
```
[Relevant log entries showing the failure]
```

**Metrics:**
[Charts or data showing system behavior during incident]

---

## Resolution

### Immediate Mitigation

**Actions Taken:**
1. [First action to stop the bleeding]
2. [Second action]
3. [Third action]

**Mitigation Effectiveness:**
[How well did the immediate actions work?]

### Permanent Fix

**Solution Implemented:**
[Description of the permanent fix]

**Implementation Details:**
```language
// Fix code
[Code changes made]
```

**Configuration Changes:**
```yaml
# Configuration fix
[Config changes]
```

**Deployment:**
- Deployed to: [Environment]
- Deployment method: [Rolling/Blue-green/etc.]
- Verification: [How we confirmed the fix]

### Rollback Plan (if needed)

[What was the rollback strategy in case the fix made things worse?]

---

## Response Evaluation

### What Went Well

- [Success 1]
- [Success 2]
- [Success 3]

### What Went Poorly

- [Issue 1]
- [Issue 2]
- [Issue 3]

### Communication

**Internal Communication:**
- [ ] Engineering team notified
- [ ] Leadership notified
- [ ] Support team notified
- [ ] Incident channel created: [Link]

**External Communication:**
- [ ] Status page updated: [Link]
- [ ] Customers notified: [Method]
- [ ] Public statement issued: [Link]

**Communication Timeline:**
| Time | Audience | Channel | Message |
|------|----------|---------|---------|
| HH:MM | Internal | Slack | [Summary] |
| HH:MM | Customers | Email | [Summary] |
| HH:MM | Public | Status Page | [Summary] |

---

## Prevention & Follow-up

### Immediate Action Items

- [ ] [Action 1] - Owner: [Name] - Due: YYYY-MM-DD - Status: [Open/Done]
- [ ] [Action 2] - Owner: [Name] - Due: YYYY-MM-DD - Status: [Open/Done]
- [ ] [Action 3] - Owner: [Name] - Due: YYYY-MM-DD - Status: [Open/Done]

### Short-term Improvements (This Sprint)

- [ ] Improve monitoring for [specific metric]
- [ ] Add alerting for [specific condition]
- [ ] Update runbook for [procedure]
- [ ] Conduct training on [topic]

### Long-term Improvements (Next Quarter)

- [ ] Architectural change: [Description]
- [ ] Process improvement: [Description]
- [ ] Tool/infrastructure upgrade: [Description]

### Monitoring & Alerting Improvements

**New Alerts Needed:**
- [ ] Alert: [Description] - Threshold: [Value] - Owner: [Name]
- [ ] Alert: [Description] - Threshold: [Value] - Owner: [Name]

**Dashboard Updates:**
- [ ] Add metric: [Description]
- [ ] Add graph: [Description]

### Documentation Updates

- [ ] Update runbook: [Which runbook]
- [ ] Update architecture docs: [What changes]
- [ ] Update on-call guide: [What to add]
- [ ] Create new ADR: [Decision to document]

---

## Lessons Learned

### Technical Lessons

**What we learned about the system:**
- [Learning 1]
- [Learning 2]
- [Learning 3]

**How to prevent this class of incident:**
- [Prevention 1]
- [Prevention 2]

### Process Lessons

**What we learned about our response:**
- [Learning 1]
- [Learning 2]

**How to improve incident response:**
- [Improvement 1]
- [Improvement 2]

### Similar Incidents

**Have we seen this before?**
- [Link to similar incident INC-YYYY-NNN]
- [Link to similar incident INC-YYYY-NNN]

**Pattern Analysis:**
[Is this part of a broader pattern? What does it tell us about system weaknesses?]

---

## Related Documents

- **Task Packet:** `.ai/tasks/[incident-id]/`
- **Investigation Report:** `.ai/tasks/[incident-id]/runtime-report.md`
- **Architecture:** [Link to relevant architecture docs]
- **ADRs:** [Links to related architecture decisions]
- **Runbook:** [Link to relevant runbook]
- **Fix PR:** [Link to pull request with fix]
- **Post-mortem Meeting:** [Link to meeting notes]

---

## Severity Classification

### SEV-1 (Critical)
- **Definition:** Complete service outage or data loss
- **Response:** Immediate all-hands, page on-call
- **SLA:** Resolve within 1 hour

### SEV-2 (High)
- **Definition:** Major feature unavailable or severe degradation
- **Response:** Immediate on-call response
- **SLA:** Resolve within 4 hours

### SEV-3 (Medium)
- **Definition:** Minor feature degradation or partial unavailability
- **Response:** Normal business hours response
- **SLA:** Resolve within 1 business day

### SEV-4 (Low)
- **Definition:** Cosmetic issue or minimal impact
- **Response:** Backlog prioritization
- **SLA:** No specific SLA

---

## Sign-off

**Incident Commander:** [Name] - [Date]
**Technical Lead:** [Name] - [Date]
**Engineering Manager:** [Name] - [Date]

**Post-mortem Complete:** [Yes/No]
**Action Items Tracked:** [Yes/No - Link to tracker]
**Documentation Updated:** [Yes/No]

---

**Document Version:** 1.0
**Template Version:** 1.0.0
**Last Updated:** 2026-01-14
**Template Location:** `.ai-pack/templates/incidents/incident-template.md`

## Usage Instructions

1. Copy this template to `docs/incidents/INC-YYYY-NNN-short-description.md`
2. Fill in incident ID: `INC-YYYY-NNN`
3. Update timeline in real-time during incident
4. Complete RCA after resolution
5. Conduct post-mortem meeting
6. Track action items to completion
7. Update `docs/incidents/README.md` index
8. Archive when all follow-ups complete
