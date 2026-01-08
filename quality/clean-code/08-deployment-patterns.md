# Deployment and Release Patterns

> Safe deployment strategies from [martinfowler.com](https://martinfowler.com)

## Feature Toggles (Feature Flags)

**Definition:** Configuration-based mechanisms that allow controlling feature visibility without changing code.

### Core Concept

A configuration file defines flags for pending features. The running application uses these toggles to decide whether to show or execute new functionality.

### Implementation Approaches

**User Interface Toggles**
```javascript
if (featureFlags.isEnabled('new-checkout')) {
    return <NewCheckoutComponent />;
} else {
    return <OldCheckoutComponent />;
}
```

**Application Logic Toggles**
```javascript
function calculatePrice(item) {
    if (featureFlags.isEnabled('new-pricing-algorithm')) {
        return newPricingAlgorithm(item);
    }
    return oldPricingAlgorithm(item);
}
```

### Types of Feature Toggles

**Release Toggles**
- Hide incomplete features
- Allow committing to mainline before feature is done
- Enable continuous integration with long-running features

**Experiment Toggles**
- Enable A/B testing
- Compare different implementations
- Data-driven decision making

**Ops Toggles**
- Give operations staff runtime controls
- Circuit breakers for external dependencies
- Performance tuning without redeployment

**Permissioning Toggles**
- Restrict features to user subsets
- Gradual rollout by user segment
- Premium features for paid users

### Best Practices

**Minimize Toggle Points**
- Protect only entry points, not every code path
- Reduces complexity and testing burden
- Makes toggle removal easier

**Testing Strategy**
- Validate next release's expected flag configuration
- Test all-flags-on scenario
- Similar to feature branch testing

**Retire Toggles Promptly**
- Release flags should be temporary
- Remove flags after deployment stabilizes
- Unmaintained toggles accumulate into technical debt

### When to Use

**Feature toggles should be a last resort:**
- Prioritize small, frequent releases first
- Use "Keystone Interface" approach when possible
- Only use toggles when necessary for:
  - Long-running feature development
  - Gradual rollouts
  - A/B testing requirements
  - Operational control needs

### Dangers

- Toggle proliferation creates complexity
- Combinatorial explosion of configurations to test
- Technical debt if not retired promptly
- Can obscure code logic if overused

**Reference:** [Feature Toggles](https://martinfowler.com/bliki/FeatureToggle.html)

## Blue-Green Deployment

**Definition:** Maintain two identical production environments where one serves live traffic while the other is idle, enabling rapid cutover with minimal downtime.

### How It Works

1. **Blue environment** serves all production traffic
2. Deploy new release to **green environment** (idle)
3. Perform final testing in green environment
4. Switch router to direct traffic to green
5. Blue environment becomes idle backup
6. Next deployment reverses roles

### Key Benefits

**Rapid Rollback**
- If issues arise, quickly redirect traffic back
- "If anything goes wrong you switch the router back to your blue environment"
- Minimal downtime from problems
- Reduced deployment risk

**Disaster Recovery Testing**
- Regularly test disaster-recovery procedures
- Both environments cycle between live, backup, and staging
- Validate recovery capabilities with every release
- Ensures backups actually work

**Schema Migration Support**
- Decouple database changes from application upgrades
- Apply database modifications supporting both versions first
- Verify stability before deploying new application
- Reduces deployment risk significantly

**Deployment Flexibility**
- Can work with separate hardware, VMs, or partitioned environments
- Vary granularity (just web tier vs. entire stack)
- Adapt to infrastructure constraints

### Implementation Considerations

**Database Handling**
- Option 1: Separate databases per environment (complete isolation)
- Option 2: Shared database (requires backward-compatible schema changes)
- Use Parallel Change pattern for schema evolution

**Testing in Production**
- Final verification happens in real production environment
- Catches environment-specific issues before cutover
- Smoke tests run in green before switching traffic

**Infrastructure Requirements**
- Need capacity for two full environments
- May not double cost (one environment idle)
- Cloud infrastructure makes this more affordable

### Variations

**Different granularities:**
- Just application tier (blue-green web/app servers)
- Include data tier (completely separate stacks)
- Mix of approaches based on what changes

**Canary within blue-green:**
- Switch small percentage to green first
- Gradually increase traffic to green
- Combines benefits of both patterns

**Reference:** [Blue-Green Deployment](https://martinfowler.com/bliki/BlueGreenDeployment.html)

## Canary Release

**Definition:** Deployment strategy that slowly rolls out changes to a small subset of users before rolling out to entire infrastructure.

### How It Works

1. Deploy new version to small portion of infrastructure
2. Route limited set of users to new version
3. Monitor for issues with careful observation
4. Gradually increase traffic as confidence grows
5. Complete rollout across all infrastructure
6. Decommission old version

### User Selection Strategies

**Random Sampling**
- Route X% of requests to new version
- Simple to implement
- Statistically representative

**Internal Users First**
- Roll out to employees before customers
- "Dogfooding" catches issues early
- Reduced customer impact

**Demographic-Based**
- Target specific user segments
- Test with less critical users first
- Geographic or feature-based selection

### When to Use Canary Releases

**Ideal Scenarios:**
- Production environments where risk mitigation is critical
- Systems requiring capacity testing in realistic conditions
- Applications with gradual deployment needs
- Services where quick rollback capability is essential

**Challenges:**
- Managing multiple software versions simultaneously
- Distributed software (mobile apps, desktop) with limited upgrade control
- Database schema changes requiring coordination
- Increased operational complexity

### Key Benefits

**Capacity Testing**
- "Test the new version in a production environment with safe rollback"
- Realistic load and data
- Catches performance issues before full rollout

**Risk Mitigation**
- Limits blast radius of problems
- Quick detection of issues
- Easy rollback by rerouting users

**Gradual Confidence Building**
- Start small, grow as confidence increases
- Monitor metrics at each stage
- Data-driven rollout decisions

### Monitoring and Metrics

**Critical to track:**
- Error rates in canary vs. baseline
- Performance metrics
- User behavior patterns
- Business metrics (conversion, engagement, etc.)

**Automated Detection:**
- "Cluster immune systems" can automatically detect problems
- Auto-rollback on anomalies
- Reduces manual monitoring burden

### Relationship to Other Patterns

**Works well with:**
- Blue-Green Deployment (canary within blue-green)
- Feature Toggles (control who sees canary)
- Parallel Change (manage version compatibility)

**Reference:** [Canary Release](https://martinfowler.com/bliki/CanaryRelease.html)

## Parallel Change (Expand-Contract)

**Definition:** Pattern for making breaking changes safely by dividing work into three phases: expand, migrate, contract.

### The Three Phases

**Phase 1: Expand**
- Add new interface alongside existing one
- Support both old and new signatures simultaneously
- No breaking changes yet

**Example:**
```javascript
class Grid {
    // Old method - keep for now
    addCell(x, y, cell) { ... }

    // New method - add alongside
    addCell(coordinate, cell) { ... }
}
```

**Phase 2: Migrate**
- Update all clients incrementally to use new interface
- Can be gradual, especially with external clients
- Each migration is safe and testable
- Old interface remains functional during migration

**Phase 3: Contract**
- Once migration complete, remove old methods
- Clean up redundant internal structures
- Codebase now uses only new interface

### Key Benefits

**Continuous Delivery Compatibility**
- "Particularly useful when practicing Continuous Delivery"
- Code can be released in any of these three phases
- No big bang cutover required
- Always in deployable state

**Risk Reduction**
- Prevents widespread breakage across codebase
- Each migration step is small and safe
- Easy to test incremental changes
- Rollback is straightforward

**Works with External Clients**
- Give external consumers time to migrate
- Deprecation warnings before contract phase
- Can version APIs during migration

### Real-World Applications

**Refactoring Method Signatures**
```javascript
// Expand: Add new method
function calculateTotal(order) { ... }
function calculateTotalWithTax(order) { ... }  // New

// Migrate: Update call sites one by one
// calculateTotal(order) → calculateTotalWithTax(order)

// Contract: Remove old method
```

**Database Schema Evolution**
- Expand: Add new column, keep old column
- Migrate: Dual-write to both columns, backfill data
- Contract: Remove old column

**Remote API Evolution**
- Expand: Support both v1 and v2 endpoints
- Migrate: Update clients to v2
- Contract: Deprecate and remove v1

**Deployment Strategies**
- Expand: Deploy new version alongside old (blue-green)
- Migrate: Gradually shift traffic (canary)
- Contract: Decommission old version

### Main Downside

**Maintaining Dual Versions**
- Temporary complexity during migration phase
- Need discipline to complete contract phase
- Can end up in confusing state if contract is forgotten
- Set deadlines for completing contract

### Best Practices

- Set explicit timelines for contract phase
- Document which phase you're in
- Automate detection of old interface usage
- Don't start new expansion until previous contract complete
- Communicate migration timeline to stakeholders

**Reference:** [Parallel Change](https://martinfowler.com/bliki/ParallelChange.html)

## Deployment Pattern Selection Guide

### Choose Blue-Green When:
- Need instant rollback capability
- Can maintain two complete environments
- Want to test in production before cutover
- Database changes can be backward-compatible

### Choose Canary When:
- Need gradual risk reduction
- Want real-world capacity testing
- Can monitor metrics effectively
- Managing multiple versions is feasible

### Choose Feature Toggles When:
- Feature development exceeds release cycle
- Need A/B testing capability
- Want to decouple deployment from release
- Require operational runtime controls

### Use Parallel Change For:
- API evolution with external clients
- Database schema migrations
- Refactoring breaking interfaces
- Any breaking change requiring gradual migration

### Combine Patterns:
- Blue-Green + Canary: Switch small traffic first, then all
- Feature Toggles + Canary: Control which users see new feature
- Parallel Change + Blue-Green: Support multiple versions during cutover
- All three together for maximum safety and control
