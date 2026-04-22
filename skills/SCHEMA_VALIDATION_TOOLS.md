# GraphQL Schema Validation Tools

## Publicly Available Tools for Schema Review

These open-source and commercial tools can automatically detect many of the patterns documented in our validation skills.

---

## 1. Apollo Rover CLI

**Type:** Official Apollo CLI tool  
**License:** Free & Open Source  
**Best For:** Apollo Federation validation, schema checks, CI/CD integration

**Key Features:**
- `rover subgraph lint` - Lint individual subgraphs
- `rover graph lint` - Lint composed supergraph
- Schema composition validation
- Breaking change detection
- GraphOS integration for centralized governance
- CI/CD pipeline integration

**Installation:**
```bash
# macOS/Linux
curl -sSL https://rover.apollo.dev/nix/latest | sh

# npm
npm install -g @apollo/rover

# Check version (v0.16+ recommended)
rover --version
```

**Usage:**
```bash
# Lint local schema
rover subgraph lint my-subgraph \
  --schema ./schema.graphql

# Check for breaking changes
rover subgraph check my-graph@main \
  --schema ./schema.graphql \
  --name my-subgraph

# Validate composition
rover supergraph compose --config supergraph.yaml
```

**What It Catches:**
- Federation directive errors (@key, @requires, @provides)
- Composition conflicts
- Breaking changes
- Schema linting rules (configurable in GraphOS)
- Nullability inconsistencies
- Field description issues

**Resources:**
- [Apollo Rover Documentation](https://www.apollographql.com/docs/rover/)
- [Schema Linting Rules](https://www.apollographql.com/docs/graphos/platform/schema-management/linting/rules)
- [Schema Checks](https://www.apollographql.com/docs/graphos/platform/schema-management/checks)

---

## 2. graphql-schema-linter

**Type:** Standalone schema linter  
**License:** MIT (Open Source)  
**Best For:** General GraphQL schemas, custom rules, non-Federation projects

**Key Features:**
- Built-in validation rules
- Custom rule support
- Configuration file support
- CI-friendly exit codes
- Multiple output formats

**Installation:**
```bash
npm install -g graphql-schema-linter

# Or as dev dependency
npm install --save-dev graphql-schema-linter
```

**Usage:**
```bash
# Lint schema with default rules
graphql-schema-linter schema.graphql

# Custom configuration
graphql-schema-linter schema.graphql \
  --config-direction .graphql-schema-linterrc \
  --format json

# Custom rules
graphql-schema-linter schema.graphql \
  --custom-rule-paths ./custom-rules \
  --rules defined-types-are-used,deprecations-have-a-reason
```

**Built-in Rules:**
- `defined-types-are-used` - No unused types
- `deprecations-have-a-reason` - @deprecated needs reason
- `enum-values-sorted-alphabetically` - Consistent enum ordering
- `fields-are-camel-cased` - Naming conventions
- `types-are-capitalized` - Type name conventions
- `input-object-values-are-camel-cased` - Input naming
- `relay-connection-types-spec` - Relay pagination compliance
- And many more...

**What It Catches:**
- Naming convention violations
- Unused types/enums
- Missing deprecation reasons
- Description format issues
- Relay spec violations
- Custom rule violations

**Resources:**
- [GitHub: graphql-schema-linter](https://github.com/cjoudrey/graphql-schema-linter)
- [npm Package](https://www.npmjs.com/package/graphql-schema-linter)

---

## 3. GraphQL Inspector

**Type:** Schema comparison & validation tool  
**License:** MIT (Open Source)  
**Best For:** Schema diffing, CI/CD, detecting breaking changes

**Key Features:**
- Schema comparison (diff)
- Breaking change detection
- Coverage analysis
- Similar schema detection
- GitHub Action integration
- CI/CD friendly

**Installation:**
```bash
npm install -g @graphql-inspector/cli

# Or with specific plugins
npm install @graphql-inspector/cli \
  @graphql-inspector/diff-command \
  @graphql-inspector/validate-command
```

**Usage:**
```bash
# Compare schemas
graphql-inspector diff old-schema.graphql new-schema.graphql

# Validate schema
graphql-inspector validate schema.graphql

# Check coverage
graphql-inspector coverage schema.graphql operations/*.graphql

# Detect similar types
graphql-inspector similar schema.graphql
```

**What It Catches:**
- Breaking changes (field removals, type changes)
- Dangerous changes (enum additions, nullability)
- Schema evolution issues
- Similar/duplicate types
- Unused schema parts
- Operation coverage gaps

**Resources:**
- [GraphQL Inspector Website](https://graphql-inspector.com)
- [GitHub: GraphQL Inspector](https://github.com/kamilkisiela/graphql-inspector)
- [CI/CD Integration Guide](https://graphql-inspector.com/docs/essentials/ci)

---

## 4. GraphQL ESLint

**Type:** ESLint plugin for GraphQL  
**License:** MIT (Open Source)  
**Best For:** Linting both schemas AND operations, IDE integration

**Key Features:**
- Schema linting
- Operation (query/mutation) linting
- IDE integration (VSCode, etc.)
- Auto-fix support
- Federation support
- Custom rules

**Installation:**
```bash
npm install --save-dev @graphql-eslint/eslint-plugin

# Peer dependencies
npm install --save-dev graphql eslint
```

**Configuration:**
```javascript
// .eslintrc.js
module.exports = {
  overrides: [
    {
      files: ['*.graphql'],
      parser: '@graphql-eslint/eslint-plugin',
      plugins: ['@graphql-eslint'],
      rules: {
        '@graphql-eslint/known-type-names': 'error',
        '@graphql-eslint/no-typename-prefix': 'error',
        '@graphql-eslint/require-description': ['error', {
          types: true,
          FieldDefinition: true
        }],
        '@graphql-eslint/naming-convention': ['error', {
          types: 'PascalCase',
          FieldDefinition: 'camelCase'
        }]
      }
    }
  ]
};
```

**Usage:**
```bash
# Lint GraphQL files
eslint '**/*.graphql'

# With auto-fix
eslint '**/*.graphql' --fix
```

**What It Catches:**
- Naming convention violations
- Missing descriptions
- Type naming issues
- Unknown type references
- Deprecated usage
- Federation directive issues
- Operation validation

**Resources:**
- [GitHub: graphql-eslint](https://github.com/B2o5T/graphql-eslint)
- [Rule Documentation](https://the-guild.dev/graphql/eslint/docs)

---

## 5. WunderGraph Cosmo

**Type:** Open-source GraphQL Federation platform  
**License:** Apache 2.0 (Fully Open Source)  
**Best For:** Self-hosted federation, Apollo alternative, full lifecycle management

**Key Features:**
- Schema Registry with composition validation
- `wgc subgraph check` CLI for validation
- Breaking change detection with client usage analysis
- Schema linting for best practices and style guides
- Real-time composition checks
- Built-in observability (metrics, tracing, analytics)
- High-performance router
- Self-hosted or managed service options

**Installation:**
```bash
# Install Cosmo CLI
npm install -g wgc

# Or using specific version
npm install -g wgc@latest
```

**Usage:**
```bash
# Validate schema changes
wgc subgraph check \
  --name my-subgraph \
  --schema ./schema.graphql

# Push schema to registry
wgc subgraph publish \
  --name my-subgraph \
  --schema ./schema.graphql \
  --url https://api.example.com/graphql

# Check composition
wgc federated-graph check my-graph
```

**What It Catches:**
- Breaking changes with client usage analysis
- Schema composition errors
- Federation v1 & v2 compliance issues
- Style guide violations (via linting)
- Best practice violations
- Real-time client impact analysis

**Unique Features:**
- **Smart Safety Checks**: Analyzes real client usage to determine if "breaking" changes are actually safe
- **Open Source**: Apache 2.0 license, no vendor lock-in
- **Self-Hosted**: Full control via Kubernetes deployment
- **Unified Platform**: Schema registry + composition + observability + router in one

**Resources:**
- [WunderGraph Cosmo Documentation](https://cosmo-docs.wundergraph.com/overview)
- [GitHub: cosmo](https://github.com/wundergraph/cosmo)
- [Deploy with Confidence Tutorial](https://cosmo-docs.wundergraph.com/tutorial/deploy-federated-graphql-with-confidence)
- [Open Source Schema Registry Article](https://medium.com/@wundergraph/a-open-source-schema-registry-with-schema-checks-for-federated-graphql-242409787af4)

---

## 6. Apollo Studio (GraphOS)

**Type:** Cloud-based schema management platform  
**License:** Free tier available, paid plans for teams  
**Best For:** Centralized governance, team collaboration, production monitoring

**Key Features:**
- Schema registry
- Automated schema checks on PR
- Breaking change detection
- Schema linting with custom rules
- Field usage analytics
- Operation performance monitoring
- Schema changelog

**Setup:**
```bash
# Publish schema
rover subgraph publish my-graph@main \
  --schema ./schema.graphql \
  --name my-subgraph \
  --routing-url https://api.example.com/graphql

# Configure schema checks in CI
rover subgraph check my-graph@main \
  --schema ./schema.graphql \
  --name my-subgraph
```

**What It Catches:**
- All Rover CLI checks (above)
- Schema linting violations (configurable rules)
- Breaking changes with client impact analysis
- Field usage patterns
- Operation performance issues
- Composition errors across entire graph

**Resources:**
- [Apollo GraphOS](https://www.apollographql.com/docs/graphos/)
- [Schema Linting](https://www.apollographql.com/docs/graphos/platform/schema-management/linting)
- [Schema Checks](https://www.apollographql.com/docs/graphos/platform/schema-management/checks)

---

## Comparison Matrix

| Tool | Schema Linting | Breaking Changes | Federation | Custom Rules | CI/CD | IDE | Self-Hosted |
|------|----------------|------------------|------------|--------------|-------|-----|-------------|
| **Apollo Rover** | ✅ | ✅ | ✅ | ✅ (via GraphOS) | ✅ | ❌ | ❌ |
| **graphql-schema-linter** | ✅ | ❌ | ❌ | ✅ | ✅ | ❌ | ✅ |
| **GraphQL Inspector** | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ✅ |
| **GraphQL ESLint** | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **WunderGraph Cosmo** | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ |
| **Apollo Studio** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ (plugin) | ❌ |

---

## Recommended Tool Combinations

### For Apollo Federation Projects (Cloud):
```bash
# Primary: Apollo Rover + GraphOS
rover subgraph check --schema schema.graphql

# Secondary: GraphQL Inspector for local diffing
graphql-inspector diff old.graphql new.graphql
```

### For Apollo Federation Projects (Self-Hosted):
```bash
# Primary: WunderGraph Cosmo (fully open source)
wgc subgraph check --name my-subgraph --schema schema.graphql

# Secondary: GraphQL Inspector for local diffing
graphql-inspector diff old.graphql new.graphql
```

### For Non-Federation Projects:
```bash
# Primary: graphql-schema-linter
graphql-schema-linter schema.graphql

# Secondary: GraphQL Inspector for breaking changes
graphql-inspector diff old.graphql new.graphql
```

### For Maximum Coverage:
```bash
# 1. Apollo Rover for Federation
rover subgraph lint --schema schema.graphql

# 2. graphql-schema-linter for additional rules
graphql-schema-linter schema.graphql --rules all

# 3. GraphQL Inspector for diffing
graphql-inspector diff old.graphql new.graphql

# 4. GraphQL ESLint in IDE for real-time feedback
eslint '**/*.graphql'
```

---

## Integration with Skills

Our **federated_graphql_reviewer** skill complements these tools by:

1. **Validating tool output** - Understanding warnings/errors from these tools
2. **Manual review** - Catching patterns tools miss (domain modeling, client UX)
3. **Context-aware feedback** - Explaining WHY issues matter, not just WHAT
4. **Actionable fixes** - Providing specific code examples for each issue
5. **Education** - Teaching teams to write better schemas from the start

Our **federated_graphql_designer** skill helps engineers:

1. **Prevent issues** - Design schemas that pass tool validation first time
2. **Understand patterns** - Learn WHY certain patterns are problematic
3. **Apply best practices** - Netflix/Apollo patterns beyond what tools check
4. **Avoid common mistakes** - Issues from 100+ production PR reviews

---

## CI/CD Integration Examples

### GitHub Actions with Rover

```yaml
name: Schema Check
on: [pull_request]

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install Rover
        run: |
          curl -sSL https://rover.apollo.dev/nix/latest | sh
          echo "$HOME/.rover/bin" >> $GITHUB_PATH
      
      - name: Run Schema Check
        env:
          APOLLO_KEY: ${{ secrets.APOLLO_KEY }}
        run: |
          rover subgraph check my-graph@main \
            --schema ./schema.graphql \
            --name my-subgraph
```

### GitHub Actions with graphql-schema-linter

```yaml
name: Schema Lint
on: [pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Node
        uses: actions/setup-node@v3
        with:
          node-version: '18'
      
      - name: Install linter
        run: npm install -g graphql-schema-linter
      
      - name: Lint schema
        run: graphql-schema-linter schema/*.graphql
```

### GitHub Actions with GraphQL Inspector

```yaml
name: Schema Diff
on: [pull_request]

jobs:
  diff:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0
      
      - name: Get schema from main
        run: git show origin/main:schema.graphql > old-schema.graphql
      
      - name: Run GraphQL Inspector
        uses: kamilkisiela/graphql-inspector@master
        with:
          schema: 'old-schema.graphql'
          new-schema: 'schema.graphql'
```

### GitHub Actions with WunderGraph Cosmo

```yaml
name: Cosmo Schema Check
on: [pull_request]

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Node
        uses: actions/setup-node@v3
        with:
          node-version: '18'
      
      - name: Install Cosmo CLI
        run: npm install -g wgc
      
      - name: Check Schema
        env:
          COSMO_API_KEY: ${{ secrets.COSMO_API_KEY }}
        run: |
          wgc subgraph check \
            --name my-subgraph \
            --schema ./schema.graphql
```

---

## Apollo vs WunderGraph Cosmo

Both are excellent platforms for GraphQL Federation with different trade-offs:

| Feature | Apollo GraphOS | WunderGraph Cosmo |
|---------|----------------|-------------------|
| **License** | Proprietary (free tier) | Apache 2.0 (fully open) |
| **Hosting** | Cloud only | Self-hosted or managed |
| **Vendor Lock-in** | Yes | No |
| **Schema Registry** | ✅ | ✅ |
| **Breaking Changes** | ✅ | ✅ with client usage analysis |
| **Schema Linting** | ✅ | ✅ |
| **Observability** | ✅ | ✅ |
| **Router** | ✅ (separate) | ✅ (built-in) |
| **Federation Support** | v1 & v2 | v1 & v2 |
| **Client Impact Analysis** | ✅ (field usage) | ✅ (real-time traffic) |
| **Cost** | Free tier, then paid | Free (self-host) or managed |

**Choose Apollo GraphOS if:**
- You want fully managed cloud service
- You're already in the Apollo ecosystem
- You need enterprise support

**Choose WunderGraph Cosmo if:**
- You need self-hosted/on-premises deployment
- You want no vendor lock-in
- You prefer open source solutions
- You want unified platform (registry + router + observability)

---

## Sources

- [Apollo Rover Documentation](https://www.apollographql.com/docs/rover/)
- [Schema Linting - Apollo GraphQL Docs](https://www.apollographql.com/docs/graphos/platform/schema-management/linting)
- [Schema Checks - Apollo GraphQL Docs](https://www.apollographql.com/docs/graphos/platform/schema-management/checks)
- [graphql-schema-linter GitHub](https://github.com/cjoudrey/graphql-schema-linter)
- [GraphQL Inspector Website](https://graphql-inspector.com)
- [GraphQL ESLint Documentation](https://the-guild.dev/graphql/eslint/docs)
- [Standardize GraphQL schema linting policies with GraphOS](https://www.apollographql.com/blog/standardize-graphql-schema-linting-policies-with-graphos)
- [WunderGraph Cosmo Documentation](https://cosmo-docs.wundergraph.com/overview)
- [WunderGraph Cosmo GitHub](https://github.com/wundergraph/cosmo)
- [Deploy Federated GraphQL with Confidence](https://cosmo-docs.wundergraph.com/tutorial/deploy-federated-graphql-with-confidence)
- [Open Source Schema Registry with Schema Checks](https://medium.com/@wundergraph/a-open-source-schema-registry-with-schema-checks-for-federated-graphql-242409787af4)

---

**Note:** These tools detect mechanical/structural issues. Human reviewers (or AI agents with our skills) are still needed for:
- Domain modeling quality
- Client UX considerations
- Performance implications
- Business logic validation
- Architectural alignment
- Team coordination
