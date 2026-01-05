# System Evolution and Architecture Patterns

> Strategies for evolving systems over time from [martinfowler.com](https://martinfowler.com)

## Strangler Fig Pattern

**Definition:** Gradual approach to modernizing legacy systems by building new functionality separately and progressively replacing the old system.

### Core Concept

Instead of attempting complete system replacement, teams build new functionality separately from legacy codebase. As the new system matures and gains capabilities, it progressively "strangles" the old system by absorbing its functions until legacy code becomes unnecessary.

**Inspired by nature:** Strangler fig vines grow around host trees, eventually replacing them entirely.

### Four Essential Activities

**1. Establish Clear Outcomes**
- Align stakeholders on what modernization should achieve
- Define success criteria beyond "rewrite the system"
- Understand business drivers for modernization

**2. Break Down the Problem**
- Identify "seams" in legacy system
- Find isolatable components that can be replaced
- Map dependencies and interfaces
- Determine replacement order

**3. Deliver Incrementally**
- Replace small components with reduced risk
- Achieve faster value realization
- Deploy replacements independently
- Learn continuously about legacy behavior

**4. Evolve Organizational Practices**
- Introduce new development practices alongside new code
- Update team structures and processes
- Build capability while building software
- Cultural change happens gradually with technical change

### Why It Works

**Mitigates "Big Bang" Replacement Risks**
- Phased rollouts reduce investment risk
- Earlier delivery of business value
- Continuous learning about legacy system behavior
- Organization adapts alongside technology

**Acknowledges Reality**
- Legacy systems embed decades of business logic
- Organizational processes are tightly coupled to old system
- Complete replacement is impractical
- Gradual displacement ensures business continuity

### Implementation Strategy

**Routing Layer**
- Create facade/proxy in front of legacy system
- Route requests to either old or new system
- Gradually shift routes as new components are ready
- Eventually remove legacy system entirely

**Replacement Pattern**
```
Phase 1: [Legacy System] ← All traffic

Phase 2: [Routing Layer] → [Legacy System] ← Most traffic
                        ↘ [New Component A] ← Some traffic

Phase 3: [Routing Layer] → [Legacy System] ← Less traffic
                        ↘ [New System] ← Most traffic

Phase 4: [New System] ← All traffic (legacy retired)
```

### Best Practices

**Start with Edges**
- Replace peripheral functionality first
- Core logic replaced last
- Reduces risk, builds capability

**Maintain Compatibility**
- New components must integrate with legacy
- Use adapters and facades
- Database access patterns may need coordination

**Set Metrics**
- Track modernization progress
- Measure business value delivered
- Monitor system health during transition

**Celebrate Milestones**
- Recognize component retirements
- Build momentum for further modernization
- Maintain team morale during long effort

**Reference:** [Strangler Fig Application](https://martinfowler.com/bliki/StranglerFigApplication.html)

## Sacrificial Architecture

**Definition:** Intentionally designing systems with the expectation they'll be replaced within a few years.

### Core Philosophy

> "The best code you can write now is code you'll discard in a couple of years time."

Rather than treating replacement as failure, sacrificial architecture recognizes that early-stage systems serve a different purpose than mature systems.

### When It's Appropriate

**Early-Stage Systems with Uncertain Requirements**
- Exploring what users actually need
- Prioritize flexibility for feature changes over optimization
- Learning phase requires experimentation
- Architecture optimized for change, not scale

**Rapid Growth Scenarios**
- "Exponential growth isn't kind to architectural decisions"
- What works for thousands won't work for millions
- Companies often rebuild multiple times during growth
- Better to succeed and rebuild than over-engineer early

**Proof-of-Concept Phases**
- Building for limited user subsets to validate ideas
- Full investment not justified yet
- Learning whether idea has merit
- Architecture optimized for speed of learning

**Established Organizational Practice**
- Google designs systems for "10x current needs"
- Assume rewrite before reaching 100X growth
- Build what you need now, not what you might need
- Reduces premature optimization

### Important Caveats

**Don't Abandon Code Quality**
- Modularity remains essential for graceful replacement
- Clean code easier to understand when replacement time comes
- Good design makes strangling easier
- "Sacrificial" doesn't mean "sloppy"

**Monoliths Make Better Sacrificial Architecture**
- Simpler to understand as single unit
- Easier to replace in chunks via Strangler Fig
- Microservices add complexity without proven benefit
- See MonolithFirst pattern

**Original Team Should Decide Replacement**
- Team that built it understands trade-offs
- Not newcomers unfamiliar with context
- Avoid "not invented here" syndrome
- Respect decisions made under constraints

**Watch Accounting Practices**
- Don't let sunk cost thinking prevent necessary replacement
- Capitalized software assets may discourage rewriting
- Business value matters, not accounting treatment
- Optimize for outcomes, not asset preservation

### Relationship to Other Patterns

**Works Well With:**
- Strangler Fig Pattern for gradual replacement
- Feature Toggles to migrate users gradually
- Blue-Green Deployment for cutover
- MonolithFirst to keep initial architecture simple

**Reference:** [Sacrificial Architecture](https://martinfowler.com/bliki/SacrificialArchitecture.html)

## MonolithFirst

**Principle:** Build new applications as monoliths initially, even when anticipating eventual migration to microservices.

### Core Recommendation

> "You shouldn't start a new project with microservices, even if you're sure your application will be big enough."

### Key Reasons

**1. YAGNI (You Aren't Gonna Need It)**
- Early-stage applications prioritize speed and feedback cycles
- Microservices premium (management overhead) becomes unnecessary burden
- Premature distribution adds complexity without benefit
- Focus on learning and iteration, not infrastructure

**2. Boundary Definition is Hard**
- Microservices require stable, well-defined service boundaries
- Essentially correct Bounded Contexts
- "Any refactoring of functionality between services is much harder than it is in a monolith"
- Starting with monolith allows discovering optimal boundaries
- Once boundaries solidify, they become "a layer of treacle" in distributed system

**3. Evidence from Successful Systems**
- "Almost all successful microservice stories started with a monolith that got too big"
- Systems built as microservices from scratch frequently encounter serious problems
- Proven pattern: Monolith → Learn boundaries → Extract microservices
- Opposite pattern (microservices first) rarely succeeds

### Execution Strategies

**Gradual Extraction**
- Extract microservices at the edges while maintaining core monolith
- Start with coarse-grained services
- Progress to finer granularity as understanding improves
- Core monolith may remain indefinitely

**Design for Modularity**
- Build monolith with clear internal boundaries
- Use dependency injection and interfaces
- Keep modules loosely coupled
- Makes future extraction easier

**When to Extract Services**
- Wait until you understand domain boundaries
- Extract when team size justifies distribution
- Extract to enable independent deployment
- Extract to isolate scaling concerns

### The Monolith as Sacrificial Architecture

- Build monolith as disposable
- Prioritize market speed over permanence
- Plan for eventual evolution
- But don't over-engineer for hypothetical microservices

### When Microservices Might Be Appropriate Early

**Rare cases:**
- Team already has deep microservices expertise
- Clear, stable domain boundaries understood from start
- Organizational structure demands distribution
- But even then, consider monolith first

### Common Mistakes to Avoid

**Premature Decomposition**
- Splitting before understanding boundaries
- Results in chatty services
- Excessive coordination overhead
- Difficult to change boundaries later

**Under-Modularization**
- Building "big ball of mud" monolith
- Tangled dependencies prevent future extraction
- Makes eventual microservices migration impossible

**Over-Engineering**
- Adding abstraction for hypothetical future microservices
- Increases complexity without benefit
- Violates YAGNI principle

**Reference:** [MonolithFirst](https://martinfowler.com/bliki/MonolithFirst.html)

## Semantic Diffusion

**Definition:** Process where technical terms lose their original meaning as they spread through a community.

### How It Happens

Like "a succession of the telephone game," each retelling introduces distortions:

1. Originators coin term with careful definition
2. As popularity grows, secondary communicators discuss without consulting source
3. Each successive "hand-off" adds misunderstandings
4. Term's definition weakens progressively

### Why Popular Terms Suffer Most

**Ironic Effect:**
- Successful concepts face greatest semantic erosion
- Desirable-sounding terms attract more casual discussion
- Broader conceptual frameworks (vs. concrete tools) more vulnerable
- Less-defined boundaries make drift easier

### Real Examples

**"Agile"**
- Originally: Specific values and principles from Manifesto
- Diffused to: Any iterative development (even waterfall in sprints)
- Some believe it means "no planning"

**"Web 2.0"**
- Originally: Broader platform concept
- Diffused to: "Only means AJAX"

**"DevOps"**
- Originally: Cultural practice and collaboration
- Diffused to: Job title (inverts original meaning)

**"Microservices"**
- Originally: Specific architectural style with trade-offs
- Diffused to: Any small services or APIs

### Dangers of Semantic Diffusion

**Miscommunication**
- Same term means different things to different people
- Conversations happen at cross purposes
- Decisions based on misunderstanding

**Cargo Culting**
- Adopting practices without understanding principles
- "We do agile because we have standups"
- Missing the point of original concept

**Buzzword Bingo**
- Terms become marketing rather than technical
- Lose precision and usefulness
- Can't have meaningful technical discussions

### Combating Semantic Diffusion

**Personal Responsibility**
- Consult original sources before using terms
- Be precise in your own usage
- Correct misunderstandings when you encounter them
- Link to authoritative definitions

**Continuous Re-Articulation**
- Recognized authorities repeatedly clarify meanings
- Update definitions as understanding evolves
- Acknowledge how ideas naturally change
- Document canonical definitions

**The Silver Lining**
- Terms historically recover their integrity
- Example: "Object-oriented" regained meaning over time
- Community self-correction happens
- New terms emerge when old ones become too diffused

### Practical Implications

**When Using Technical Terms:**
- Define terms explicitly in documentation
- Link to authoritative sources
- Use examples to clarify meaning
- Don't assume shared understanding

**When Adopting Practices:**
- Research original sources
- Understand principles, not just practices
- Question interpretations
- Be skeptical of simplified explanations

**Reference:** [Semantic Diffusion](https://martinfowler.com/bliki/SemanticDiffusion.html)

## Evolution Pattern Selection

### Choose Strangler Fig When:
- Modernizing legacy system
- Can identify clear seams in old system
- Need to maintain business continuity
- Want to deliver value incrementally
- Have time for gradual replacement (months to years)

### Choose Sacrificial Architecture When:
- Building in uncertain domain
- Expect rapid growth
- In proof-of-concept phase
- Need to optimize for learning
- Startup or new product phase

### Choose MonolithFirst When:
- Starting new application
- Domain boundaries unclear
- Team is small/medium size
- Want to move fast and learn
- Can refactor easily later

### Avoid These Anti-Patterns:
- Big bang rewrites (use Strangler Fig instead)
- Premature microservices (use MonolithFirst)
- Over-engineering sacrificial systems
- Keeping sacrificial architecture too long
- Semantic diffusion of these patterns themselves!
