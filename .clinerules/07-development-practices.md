# Development Practices and Workflow

> Practical principles for day-to-day development from [martinfowler.com](https://martinfowler.com)

## YAGNI - You Aren't Gonna Need It

**Definition:** Don't build capabilities you presume your software will need in the future because you likely won't actually require them.

### Core Arguments for YAGNI

**Requirements are Uncertain**
- Requirements change frequently
- Research by Kohavi et al.: Only 1/3 of features built with advance analysis actually improved their intended metrics
- Even with careful planning, you may be wrong about future needs

**Multiple Costs of Building Early**

1. **Build Cost** - Wasted effort analyzing, programming, and testing unused features
2. **Cost of Delay** - Resources spent on presumptive features delay more urgent priorities and revenue
3. **Cost of Carry** - Extra code complexity makes codebase harder to modify and debug
4. **Cost of Repair** - Features built wrong months ago accumulate technical debt

### When to Apply YAGNI

Apply when introducing extra complexity now for later use, but only when:

- **Code is malleable** - Refactoring, self-testing code, and continuous delivery are prerequisites
- **Complexity threshold** - Only avoid features that genuinely increase complexity
- **Real options exist** - You can actually defer the decision

### When NOT to Apply YAGNI

YAGNI doesn't justify:
- Neglecting code health or quality
- Skipping tests or refactoring
- Avoiding effort that makes software easier to modify
- Ignoring necessary infrastructure

> "Refactoring and testing infrastructure aren't YAGNI violations—they enable evolutionary design."

**Reference:** [YAGNI](https://martinfowler.com/bliki/Yagni.html)

## Frequency Reduces Difficulty

**Core Principle:** "If it hurts, do it more often."

### The Counterintuitive Truth

Pain and time intervals have an exponential relationship. When integration, testing, or migrations happen infrequently, accumulated complexity makes them increasingly agonizing. Breaking these activities into smaller, more frequent iterations dramatically reduces friction.

### Three Key Mechanisms

**1. Decomposition**
- Large tasks become manageable when split into incremental chunks
- One small schema change is far simpler than restructuring multiple tables
- Break painful activities into smaller pieces

**2. Feedback Loops**
- Regular check-ins enable quick course corrections
- Frequent feedback accelerates learning
- Teams adapt faster to changing circumstances

**3. Practice and Automation**
- Repeating activities builds expertise
- Reveals automation opportunities
- Like a surgeon performing procedures frequently, developers who deploy regularly develop superior skills

### Practical Applications

**Continuous Integration**
- Integrating code daily makes what was once a nightmare task nearly painless
- No code sits unintegrated for more than a couple of hours

**Database Migrations**
- Frequent small changes instead of infrequent large ones
- Each migration is simple and testable

**Deployments**
- Deploy multiple times per day instead of quarterly
- Each deployment is low-risk and routine

**Refactoring**
- Continuous small improvements instead of big bang rewrites
- Code stays maintainable through constant attention

### When Facing Painful Activities

Evaluate whether these three forces apply (decomposition, feedback, practice), then increase frequency accordingly. This transforms sources of stress into sources of competence and efficiency.

**Reference:** [Frequency Reduces Difficulty](https://martinfowler.com/bliki/FrequencyReducesDifficulty.html)

## Continuous Integration

**Definition:** Development practice where team members merge changes into a shared codebase at least daily, with each integration verified by automated builds and tests.

### Essential Practices

**Version Control & Mainline**
- Single, clear mainline branch where all code converges
- Everything needed to build the product lives in version control:
  - Source code
  - Tests
  - Database schemas
  - Configuration files
  - Build scripts

**Automated Building & Testing**
- Builds must be fully automated
- "Computers are designed to perform simple, repetitive tasks"
- Build should be quick (ideally under 10 minutes) for rapid feedback
- Self-testing builds catch errors immediately

**High-Frequency Integration**
- Developers commit to mainline every day, ideally multiple times
- "No code sits unintegrated for more than a couple of hours"
- Keeps conflicts small and manageable
- Reduces integration pain through frequency

**Immediate Feedback**
- Every mainline commit triggers automated CI service build
- Build runs in reference environment (not just developer's machine)
- Broken builds become high-priority issues requiring immediate attention
- Team drops everything to fix broken mainline

**Production-Like Testing**
- Testing occurs in environments that mirror production
- Reduces risk of environment-specific failures
- "Test in a clone of the production environment"

### Key Benefits

- **Reduced delivery risk** - Predictable, small integrations instead of big bang releases
- **Fewer bugs** - Comprehensive automated testing catches issues early
- **Faster feature delivery** - Enables regular refactoring without fear
- **Business control** - Release timing decided by business, not technical constraints

### The CI Mindset

> "Continuous Integration doesn't get rid of bugs, but it does make them dramatically easier to find and remove."

- Fix broken builds immediately
- Keep builds fast
- Test in production-like environments
- Everyone can see build status
- Automate deployment

**Reference:** [Continuous Integration](https://martinfowler.com/articles/continuousIntegration.html)

## Technical Debt

**Definition:** Deficiencies in internal code quality that accumulate over time.

### The Metaphor

Like financial debt, technical debt has:
- **Principal** - The cruft in the codebase
- **Interest** - Extra effort required to add new features due to poor code quality

> "The extra effort that it takes to add new features is the interest paid on the debt." (Ward Cunningham)

**Example:** A confusing module structure adds 2 unnecessary days to a 6-day feature. Those 2 days are the interest cost.

### Types of Technical Debt

**Deliberate vs. Unintentional**
- Did cruft accumulate consciously or through careless development?

**Prudent vs. Reckless**
- Was the tradeoff well-reasoned or poorly considered?

### Managing Technical Debt Effectively

**Gradual Principal Repayment**
- Remove cruft incrementally during regular feature work
- Don't tackle large separate cleanup efforts
- Naturally focuses improvements where they matter most
- Pay down debt in high-activity areas

**Strategic Prioritization**
- Location determines urgency
- Stable but messy code can remain unchanged
- Frequently modified areas demand "zero-tolerance" attention
- Prevent compounding costs in active code

**Avoid the Debt Trap**
- Taking on debt to accelerate delivery often backfires
- Teams that neglect quality end up slower overall
- "Cruft has a quick impact, slowing down the very new features that are needed quickly"
- Quality enables speed in the long run

### The Debt Quadrant

```
                  Reckless              Prudent
Deliberate    "We don't have      "We must ship now,
              time for design"    deal with consequences"

Inadvertent   "What's layering?"  "Now we know how we
                                  should have done it"
```

### Key Principles

- Not all technical debt is bad - some is a reasonable tradeoff
- The problem is letting debt accumulate without paying it down
- High-quality code is faster to modify than poor-quality code
- Pay debt down in small increments where it hurts most

**Reference:** [Technical Debt](https://martinfowler.com/bliki/TechnicalDebt.html)

## What Refactoring Actually Means

**True Definition:** "A very specific technique built on small behavior-preserving transformations."

### Key Characteristics of Real Refactoring

**Behavior Preservation**
- System's functionality remains unchanged
- External behavior is identical
- Only internal structure changes

**Small, Incremental Steps**
- Make tiny modifications one at a time
- Each step is safe and verifiable
- Build on previous steps

**Minimal Downtime**
- "Your system should not be broken for more than a few minutes at a time"
- Always in working state
- Can stop and deploy at any point

**Well-Defined Behavior**
- Can only refactor code with clearly understood functionality
- Need tests to verify behavior preservation
- Without clear behavior, it's not refactoring

### Common Misuses

**❌ Long-Duration Restructuring**
- System broken for days during work
- This is general restructuring, not refactoring
- May be valuable but it's different

**❌ Non-Code Restructuring**
- "Refactoring" documents or other non-behavioral entities
- Not true refactoring - no behavior to preserve

### Why Precision Matters

> "I'd like us to be clear about what we mean when we use this word."

- Refactoring has specific meaning and value
- Other restructuring techniques may be valuable but are different
- Clear terminology prevents confusion
- Enables better communication about techniques

### True Refactoring Flow

1. Ensure comprehensive tests exist
2. Make one small behavior-preserving change
3. Run tests to verify behavior unchanged
4. Commit if tests pass
5. Repeat

**Reference:** [Refactoring Malapropism](https://martinfowler.com/bliki/RefactoringMalapropism.html)

## The Two Hard Things in Computer Science

> "There are only two hard things in Computer Science: cache invalidation and naming things." (Phil Karlton)

### Why Naming Matters

**Naming is Universally Difficult**
- Developers at all skill levels struggle with naming
- Requires conveying purpose and behavior clearly
- Bad names create confusion and misunderstanding
- Good names make code self-documenting

### Naming Principles

**Reveal Intention**
- Names should clearly express purpose
- Reader shouldn't need to decipher meaning
- Related to Beck's "reveals intention" rule

**Be Consistent**
- Use same terminology for same concepts
- Follow established patterns in codebase
- Avoid synonyms for identical things

**Use Domain Language**
- Names should match business/domain terminology
- Makes code understandable to non-programmers
- Reduces translation between code and requirements

**Avoid Semantic Diffusion**
- Don't let terms lose their meaning
- Be precise with technical terminology
- Resist vague, trendy buzzwords

### Warning Signs of Bad Names

- Need for explanatory comments
- Different team members interpret differently
- Names don't match what code actually does
- Abbreviations that aren't obvious
- Generic names like "data", "info", "manager", "handler"

**Reference:** [Two Hard Things](https://martinfowler.com/bliki/TwoHardThings.html)

## Git Commit Message Guidelines

**Core Principle:** Commit messages are documentation for future maintainers. Write them clearly, concisely, and without unnecessary formatting.

### Required Format

**Structure:**
```
Short summary line (50-72 characters)

Detailed explanation of what changed and why (if needed).
Wrap at 72 characters per line.

Additional paragraphs as needed.
```

### Strict Rules

**DO:**
- Use plain ASCII text only
- Start with a capitalized imperative verb (Add, Fix, Update, Remove, Refactor)
- Keep summary line under 72 characters
- Separate summary from body with blank line
- Wrap body text at 72 characters
- Explain what and why, not how
- Reference issue/ticket numbers if applicable

**DO NOT:**
- Use unicode characters or emoji
- Add tool signature footers like "Generated with [Tool]"
- Use markdown links or formatting in commit messages
- Write vague messages like "fix bug" or "update code"
- Include timestamps (git tracks this)

**ACCEPTABLE:**
- Co-Authored-By trailers for pair programming or AI assistance

### Examples

**Good:**
```
Refactor authentication module to use dependency injection

Replaced global singleton pattern with constructor injection to improve
testability and follow SOLID principles. Updated tests to use mocks
instead of real authentication service.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Bad (vague with tool signature):**
```
Update auth stuff

Made some changes to authentication. Should work better now.

Generated with [Claude Code](https://claude.com/claude-code)
```

**Bad (unicode characters):**
```
Fix authentication bug

- Add proper token validation
- Update error handling
- Improve logging

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

### Rationale

**Plain text ensures:**
- Universal readability across all terminals and git clients
- No encoding issues in different environments
- Clean git logs that can be parsed by tools
- Professional appearance in all contexts

**No tool signature footers because:**
- Commit messages document code changes, not the tools used
- Footer clutter reduces readability
- Tools change over time; the change itself matters
- Co-Authored-By is acceptable for attribution, not tool advertising

**Reference:** Industry standard practices and [Git documentation](https://git-scm.com/book/en/v2/Distributed-Git-Contributing-to-a-Project)
