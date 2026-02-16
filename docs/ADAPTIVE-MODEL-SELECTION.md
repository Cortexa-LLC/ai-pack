# Adaptive Model Selection Strategy

## Philosophy

**Start cheap, escalate only when necessary.** Most tasks can be handled by cost-effective models. Only use premium models when complexity demands it or when cheaper models fail.

## Model Tiers

### Tier 1: Minimal Cost (Preferred Default)
- **Models**: gpt-4o-mini, claude-haiku-4-5
- **Cost**: $0.15-1.25 per 1M tokens
- **Use for**: Pattern following, simple edits, test writing, refactoring
- **Success rate target**: >80%

### Tier 2: Low Cost
- **Models**: gpt-4o, gpt-5.2-mini
- **Cost**: $0.60-10.00 per 1M tokens
- **Use for**: Standard engineering work, moderate complexity
- **Escalate from Tier 1 when**: Failure rate >20% or complexity indicators present

### Tier 3: Medium Cost
- **Models**: claude-sonnet-4-5
- **Cost**: $3.00-15.00 per 1M tokens
- **Use for**: Complex reasoning, architectural decisions
- **Escalate from Tier 2 when**: Failure rate >15% or high complexity indicators

### Tier 4: High Cost (Use Sparingly)
- **Models**: claude-opus-4-6
- **Cost**: $5.00-25.00 per 1M tokens
- **Use for**: Multi-agent orchestration, strategic planning, deep analysis
- **Escalate from Tier 3 when**: Failure rate >10% or critical/complex tasks

## Performance Grading System

### Metrics Tracked Per Model/Role/Project

```go
type PerformanceGrade struct {
    ModelID       string    // e.g., "gpt-4o-mini"
    RoleID        string    // e.g., "engineer"
    ProjectID     string    // Project path hash

    // Success metrics
    TotalAttempts int
    Successes     int
    Failures      int
    Retries       int

    // Quality indicators
    AverageTokens        int     // Efficiency
    AverageExecutionTime float64 // Speed
    ErrorRate            float64 // Calculated: Failures / TotalAttempts
    RetryRate            float64 // Calculated: Retries / TotalAttempts

    // Escalation tracking
    EscalationCount int     // How many times we had to escalate
    DowngradeCount  int     // How many times we successfully downgraded

    // Calculated grade
    Grade           string  // A, B, C, D, F
    ConfidenceScore float64 // 0.0-1.0 (higher = more data, more reliable)

    // Time-based tracking
    LastUsed      time.Time
    FirstUsed     time.Time
    SampleSize    int       // Number of tasks in grade calculation
}
```

### Grade Calculation

```
Grade A (90-100%): Success rate ≥ 90%, retry rate < 5%
  → Confidence: keep using this model, consider downgrading if over-powered

Grade B (80-89%): Success rate ≥ 80%, retry rate < 10%
  → Confidence: good performance, keep using

Grade C (70-79%): Success rate ≥ 70%, retry rate < 20%
  → Warning: monitor closely, consider escalating if trend continues

Grade D (60-69%): Success rate ≥ 60%, retry rate < 30%
  → Alert: escalate to next tier on next failure

Grade F (<60%): Success rate < 60% or retry rate > 30%
  → Action: immediately escalate to next tier
```

### Confidence Score

```
confidence = min(1.0, sampleSize / 20.0)

- < 5 samples: low confidence (0.0-0.25) - use defaults
- 5-10 samples: medium confidence (0.25-0.5) - start adapting
- 10-20 samples: high confidence (0.5-1.0) - trust the data
- 20+ samples: full confidence (1.0) - fully adaptive
```

## Adaptive Model Selection Algorithm

### On Task Creation

```go
func SelectModel(role, projectID, taskDescription string) string {
    // 1. Check complexity indicators in task description
    complexity := AnalyzeComplexity(taskDescription)

    // 2. Get performance history for this role + project
    grade := GetPerformanceGrade(role, projectID)

    // 3. Start with default tier for role
    tier := GetDefaultTier(role)

    // 4. Adjust based on performance history
    if grade.ConfidenceScore > 0.5 {
        if grade.Grade == "A" && tier > 1 {
            tier = max(1, tier - 1) // Try cheaper model
        } else if grade.Grade <= "C" && tier < 4 {
            tier = min(4, tier + 1) // Use more capable model
        }
    }

    // 5. Override tier if complexity demands it
    if complexity == "very_high" {
        tier = max(tier, 3)
    } else if complexity == "high" {
        tier = max(tier, 2)
    }

    // 6. Select best model from tier
    return SelectBestModelFromTier(tier, grade)
}
```

### Complexity Analysis

```go
func AnalyzeComplexity(description string) string {
    // High complexity indicators
    highIndicators := []string{
        "architecture", "design", "orchestrate", "coordinate",
        "multiple services", "refactor entire", "redesign",
    }

    // Medium complexity indicators
    mediumIndicators := []string{
        "implement feature", "new functionality", "integrate",
        "complex logic", "optimization", "debugging",
    }

    // Low complexity indicators
    lowIndicators := []string{
        "simple", "straightforward", "basic", "minor",
        "typo", "update documentation", "add test",
        "following pattern", "similar to",
    }

    // Score based on indicators
    // Return: "low", "medium", "high", "very_high"
}
```

### On Task Completion

```go
func RecordPerformance(taskID, model, role, projectID string, success bool, retries int) {
    grade := GetOrCreateGrade(model, role, projectID)

    grade.TotalAttempts++
    if success {
        grade.Successes++
    } else {
        grade.Failures++
    }
    grade.Retries += retries

    // Recalculate metrics
    grade.ErrorRate = float64(grade.Failures) / float64(grade.TotalAttempts)
    grade.RetryRate = float64(grade.Retries) / float64(grade.TotalAttempts)

    // Update grade
    grade.Grade = CalculateGrade(grade.ErrorRate, grade.RetryRate)
    grade.ConfidenceScore = min(1.0, float64(grade.TotalAttempts) / 20.0)

    // Track escalation decisions
    if wasEscalated {
        grade.EscalationCount++
    }
    if wasDowngraded {
        grade.DowngradeCount++
    }

    SaveGrade(grade)
}
```

## Updated Role Defaults

### Engineer Role (Most Common)
```markdown
**Recommended Model (Default)**: gpt-4o-mini (Tier 1)
**Escalation Path**: gpt-4o-mini → gpt-4o → claude-sonnet-4-5 → claude-opus-4-6
**Auto-escalate when**: Grade D or F
**Auto-downgrade when**: Grade A for 10+ consecutive tasks
```

### Orchestrator Role
```markdown
**Recommended Model (Default)**: gpt-4o (Tier 2)
**Escalation Path**: gpt-4o → claude-sonnet-4-5 → claude-opus-4-6
**Auto-escalate when**: Grade C or lower
**Auto-downgrade when**: Grade A for 15+ tasks with low complexity
```

### Inspector Role
```markdown
**Recommended Model (Default)**: gpt-4o-mini (Tier 1)
**Escalation Path**: gpt-4o-mini → claude-sonnet-4-5 → claude-opus-4-6
**Auto-escalate when**: Grade D or F
**Note**: Bug investigation often needs pattern recognition, not raw intelligence
```

## Implementation Priority

### Phase 1: Foundation (Immediate)
1. Update role metadata to use cheaper defaults
2. Add escalation path to role definitions
3. Create PerformanceGrade struct and storage

### Phase 2: Tracking (Week 1)
1. Record success/failure metrics per task
2. Calculate grades after each task
3. Display grades in GUI

### Phase 3: Adaptation (Week 2)
1. Implement SelectModel() with complexity analysis
2. Auto-escalate on poor performance
3. Auto-downgrade on consistent success

### Phase 4: Intelligence (Week 3)
1. Project-specific learning
2. Task-type pattern recognition
3. Cost optimization recommendations

## Cost Savings Example

### Scenario: 100 engineering tasks

**Old approach (always Sonnet):**
- 100 tasks × 50k tokens avg × $9/1M = $45.00

**New adaptive approach:**
- 80 tasks on gpt-4o-mini: 80 × 50k × $0.375/1M = $1.50
- 15 tasks escalated to gpt-4o: 15 × 50k × $6.25/1M = $4.69
- 5 tasks escalated to sonnet: 5 × 50k × $9/1M = $2.25
- **Total: $8.44 (81% cost reduction)**

## Monitoring Dashboard

Track and display:
- Current model tier per role per project
- Performance grades with trend arrows
- Cost savings vs. always-premium approach
- Escalation/downgrade frequency
- Model success rates

## Configuration Override

Allow users to force specific tiers:
```json
{
  "adaptive_models": {
    "enabled": true,
    "min_tier": 1,  // Never go below gpt-4o-mini
    "max_tier": 3,  // Never escalate to opus (cost control)
    "confidence_threshold": 0.5,  // Minimum confidence before adapting
    "downgrade_threshold": 10  // Consecutive successes before downgrade
  }
}
```
