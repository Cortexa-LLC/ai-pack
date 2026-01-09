# Federated GraphQL Schema Design Best Practices

> Enterprise-grade patterns for designing, implementing, and evolving GraphQL Federation architectures

## Overview

GraphQL Federation enables multiple GraphQL services (subgraphs) to compose into a unified API (supergraph) that clients can query as a single endpoint. This document provides comprehensive guidance for designing federated schemas that are maintainable, performant, and evolvable.

**Core Federation Concepts:**
- **Subgraph:** Independent GraphQL service owning specific domain data
- **Supergraph:** Unified schema composed from all subgraphs
- **Gateway:** Query planner that routes requests across subgraphs
- **Entity:** Type shared across subgraphs, identifiable by @key directive

## Federated Schema Design Principles

These principles distinguish federated GraphQL from monolithic GraphQL and guide effective schema design.

### 1. Think in Entities, Not Just Types

**Principle:** Define types as entities when objects can be uniquely identified by one or more fields.

Entities are the foundational building blocks of federation:
- Marked with `@key` directive
- Serve as primary nodes in the supergraph
- Enable cross-subgraph composition
- Must have well-defined resolution inputs
- Should mirror underlying data source structures

**Entity Identification Criteria:**
- ✓ Can this object be uniquely identified by specific fields?
- ✓ Does it have independent existence and lifecycle?
- ✓ Will other subgraphs need to reference or extend it?
- ✓ Can it be efficiently resolved from a data store?

**Example:**
```graphql
# ✓ Good entity - uniquely identifiable, independently managed
type Product @key(fields: "id") {
  id: ID!
  sku: String!
  name: String!
}

# ✓ Good value object - no independent identity needed
type Money {
  amount: Decimal!
  currency: CurrencyCode!
}

# ❌ Poor entity choice - ephemeral, not independently managed
type SearchResult @key(fields: "query timestamp") {
  query: String!
  timestamp: DateTime!
  results: [Product!]!
}
```

### 2. Design Domain-Oriented Subgraphs

**Principle:** Align subgraphs with business subdomains, not technical layers.

Apply Domain-Driven Design (DDD) principles:
- Each subgraph represents a bounded context
- Use ubiquitous language within each domain
- Entities at context boundaries become federated entities
- Internal implementation details remain private

**Good - Domain-Oriented:**
```
├── products-subgraph      # Product catalog domain
├── orders-subgraph        # Order management domain
├── users-subgraph         # User identity domain
├── reviews-subgraph       # Product reviews domain
└── inventory-subgraph     # Stock management domain
```

**Bad - Technology-Oriented:**
```
├── database-subgraph      # ❌ Technical layer
├── cache-subgraph         # ❌ Technical layer
├── api-subgraph           # ❌ Technical layer
└── legacy-subgraph        # ❌ Not a domain
```

**Establish Supergraph Conventions:**
- Field naming: camelCase vs snake_case
- Type naming patterns
- Error handling conventions
- Pagination approach (offset vs cursor)
- Date/time formats

### 3. Make Collaboration Strategy Explicit

**Principle:** Document and enforce how subgraphs share entity ownership.

Choose between two ownership models based on your organizational needs:

**Model A: Single Subgraph Ownership (Recommended for Most)**
- One subgraph authoritatively owns each entity
- Other subgraphs extend the entity with domain-specific fields
- Simpler governance and clear ownership
- Best for: Greenfield projects, small-to-medium teams, clear domain boundaries

```graphql
# Products subgraph (owner)
type Product @key(fields: "id") {
  id: ID!
  name: String!
  sku: String!
  price: Money!
}

# Reviews subgraph (extends)
type Product @key(fields: "id") {
  id: ID!
  reviews: [Review!]!
  averageRating: Float
}
```

**Model B: Shared Type Ownership**
- Multiple subgraphs define and provide fields for entities using @shareable
- Required when multiple teams have existing SLAs for same data
- More complex coordination required
- Best for: Enterprises with legacy systems, multiple teams owning same data

```graphql
# Products subgraph
type Product @key(fields: "id") {
  id: ID!
  name: String! @shareable
  category: Category! @shareable
}

# Catalog subgraph
type Product @key(fields: "id") {
  id: ID!
  name: String! @shareable
  category: Category! @shareable
  catalogMetadata: CatalogData!
}
```

**Document Your Choice:**
Include collaboration strategy in ADR (Architecture Decision Record) specifying:
- Which ownership model is used
- When to create new entities vs extend existing
- Approval process for entity changes
- Migration path if changing models

### 4. Design as a Platform, Not Point-to-Point

**Principle:** The supergraph is a reusable platform for multiple clients, not a 1-to-1 API.

Design for reusability across current and future clients:
- Avoid client-specific type variations
- Provide generic capabilities that multiple clients can use
- Think "product API" not "web API" or "mobile API"
- Enable future product development with existing infrastructure

**Anti-Pattern - Client-Specific Types:**
```graphql
# ❌ Don't create separate types for different clients
type ProductWeb {
  id: ID!
  name: String!
  detailedDescription: String!
  highResImages: [Image!]!
}

type ProductMobile {
  id: ID!
  name: String!
  shortDescription: String!
  thumbnails: [Image!]!
}
```

**Platform Pattern - Generic Types with Flexibility:**
```graphql
# ✓ Single type serving multiple clients
type Product @key(fields: "id") {
  id: ID!
  name: String!
  description(format: DescriptionFormat = FULL): String!
  images(size: ImageSize): [Image!]!
}

enum DescriptionFormat {
  FULL
  SHORT
  PLAIN_TEXT
}

enum ImageSize {
  THUMBNAIL
  MEDIUM
  HIGH_RES
}
```

### 5. Enable Subgraph-Focused Development

**Principle:** Teams should develop features independently without requiring full infrastructure.

Support independent subgraph development:
- Each subgraph can be developed and tested in isolation
- Mocking for entity references from other subgraphs
- Local composition testing without deploying all subgraphs
- CI/CD pipelines per subgraph

**Development Workflow Approaches:**

**Schema-First:**
```
1. Define GraphQL schema (.graphql files)
2. Generate types/interfaces from schema
3. Implement resolvers against generated types
4. Validate with Rover CLI
```

**Code-First:**
```
1. Define schema using code (TypeScript, Java, etc.)
2. Generate GraphQL SDL from code
3. Validate generated schema
4. Deploy subgraph
```

**Avoid Schema Autogeneration from Databases:**
- Creates bloated schemas with unused fields
- Poor alignment with client needs
- Violates principle of intentional API design
- Bypasses domain modeling

## Schema Design Checklist

Use this checklist when designing or reviewing federated schemas:

### Entity Design
- [ ] Can each entity be uniquely identified by its @key fields?
- [ ] Do entity types closely mirror underlying data source structures?
- [ ] Are entity resolution inputs well-defined and efficient?
- [ ] Is the distinction clear between entities and value objects?
- [ ] Do entities align with domain boundaries?

### Subgraph Organization
- [ ] Does each subgraph represent a bounded context?
- [ ] Are technical concerns separated from domain concerns?
- [ ] Is the collaboration strategy documented?
- [ ] Are naming conventions consistent across the supergraph?
- [ ] Can subgraphs be developed and deployed independently?

### Field Ownership
- [ ] Does each field have exactly one authoritative subgraph?
- [ ] Are @external fields only used with @requires or in @key?
- [ ] Is @shareable used only for truly identical implementations?
- [ ] Are extended fields logically grouped by domain?

### Platform Design
- [ ] Are types generic enough for multiple clients?
- [ ] Are there no client-specific type variations?
- [ ] Do field arguments provide necessary flexibility?
- [ ] Can the schema support future, unknown use cases?

### Performance Considerations
- [ ] Are @provides hints used appropriately (stable data only)?
- [ ] Do reference resolvers support batching?
- [ ] Are entity keys efficiently resolvable from data stores?
- [ ] Are circular @requires dependencies avoided?

### Evolution and Maintenance
- [ ] Are breaking changes prevented by schema checks?
- [ ] Is deprecation strategy clear and documented?
- [ ] Are migration paths defined for field changes?
- [ ] Is schema versioning approach documented?

## Schema Design Patterns

### Entity Modeling and Boundaries

**Principle:** Each subgraph owns its domain entities and can extend entities from other domains.

**Entity Ownership Rules:**
1. **Single Source of Truth:** Each field has exactly one authoritative subgraph
2. **Domain Alignment:** Entities belong to the subgraph of their primary domain
3. **Encapsulation:** Subgraphs expose only what others need, hide internal details
4. **Value Types vs Entities:** Distinguish between shared value types and true entities

**Good Example - Clear Boundaries:**
```graphql
# Products subgraph (authoritative for product data)
type Product @key(fields: "id") {
  id: ID!
  sku: String!
  name: String!
  description: String!
  price: Money!
  category: Category!
}

# Reviews subgraph (extends Product with review data)
type Product @key(fields: "id") {
  id: ID!  # Reference field
  reviews: [Review!]!  # Owned by this subgraph
  averageRating: Float
  reviewCount: Int!
}

# Inventory subgraph (extends Product with stock data)
type Product @key(fields: "id") {
  id: ID!  # Reference field
  stockLevel: Int!  # Owned by this subgraph
  warehouseLocation: String!
  reorderPoint: Int!
}
```

**Bad Example - Blurred Boundaries:**
```graphql
# ❌ Products subgraph shouldn't own review aggregates
type Product @key(fields: "id") {
  id: ID!
  name: String!
  averageRating: Float  # ❌ Belongs in Reviews subgraph
  stockLevel: Int!      # ❌ Belongs in Inventory subgraph
}
```

### Choosing Effective Entity Keys

**Principle:** Keys must be stable, minimal, and efficiently resolvable.

**Key Selection Guidelines:**
1. **Immutable:** Keys should never change after creation
2. **Minimal:** Include only fields necessary for unique identification
3. **Resolvable:** Key fields must be efficiently queryable in your data store
4. **Accessible:** Other subgraphs must be able to obtain key values

**Good Example - Natural and Composite Keys:**
```graphql
# Simple natural key
type User @key(fields: "id") {
  id: ID!
  email: String!
  name: String!
}

# Composite key for multi-tenancy
type Document @key(fields: "organizationId tenantId documentId") {
  organizationId: ID!
  tenantId: ID!
  documentId: ID!
  title: String!
}

# Multiple keys for different access patterns
type Product
  @key(fields: "id")
  @key(fields: "sku") {
  id: ID!
  sku: String!
  name: String!
}
```

**Bad Example - Poor Key Choices:**
```graphql
# ❌ Mutable field as key
type User @key(fields: "email") {  # Email can change!
  email: String!
  name: String!
}

# ❌ Over-specified key
type Order @key(fields: "id userId createdAt status") {
  id: ID!  # ID alone is sufficient
  userId: ID!
  createdAt: DateTime!
  status: OrderStatus!
}

# ❌ Non-indexed field as key
type Product @key(fields: "name") {  # Name isn't indexed, slow lookup
  name: String!
  sku: String!
}
```

### Value Objects vs Entities

**Principle:** Not everything needs to be an entity. Use value objects for data without independent identity.

**Value Objects (Non-Entities):**
- Defined by their attributes, not identity
- Freely copyable across subgraphs
- No @key directive needed
- Examples: Money, Address, Coordinates, DateRange

**Entities:**
- Have unique identity
- Lifecycle managed by authoritative subgraph
- Extended by other subgraphs
- Marked with @key directive

**Example:**
```graphql
# Value object - freely reusable
type Money {
  amount: Decimal!
  currency: CurrencyCode!
}

# Value object - immutable data structure
type Address {
  street: String!
  city: String!
  state: String!
  postalCode: String!
  country: CountryCode!
}

# Entity - has identity and lifecycle
type User @key(fields: "id") {
  id: ID!
  email: String!
  billingAddress: Address  # Value object embedded
  shippingAddresses: [Address!]!
}

# Entity - owned by Orders subgraph
type Order @key(fields: "id") {
  id: ID!
  total: Money!  # Value object embedded
  shippingAddress: Address!
}
```

### Domain-Driven Design Alignment

**Principle:** Subgraph boundaries should align with bounded contexts from Domain-Driven Design.

**Bounded Context Mapping:**
1. Each subgraph represents one bounded context
2. Entities at context boundaries become federated entities
3. Context-internal types remain private to subgraph
4. Ubiquitous language of domain reflected in schema

**Example:**
```graphql
# E-Commerce Platform - Bounded Contexts as Subgraphs

# Catalog Context (catalog subgraph)
type Product @key(fields: "id") {
  id: ID!
  sku: String!
  name: String!
  description: String!
  catalogCategory: Category!
}

# Pricing Context (pricing subgraph)
type Product @key(fields: "id") {
  id: ID!
  price: PricingStrategy!
  discounts: [Discount!]!
  effectivePrice: Money!
}

# Order Context (orders subgraph)
type Order @key(fields: "id") {
  id: ID!
  items: [OrderItem!]!
  customer: Customer!
}

type OrderItem {
  product: Product!  # References Catalog context
  quantity: Int!
  unitPrice: Money!  # Snapshot from Pricing context
}
```

## Federation Directives

### @key - Entity Definition

**Purpose:** Declares that a type is an entity that can be resolved across subgraphs.

**Syntax:**
```graphql
type TypeName @key(fields: "field1 field2") {
  field1: Type!
  field2: Type!
}
```

**Multiple Keys:**
Use multiple @key directives when entities can be identified by different field combinations:

```graphql
type Product
  @key(fields: "id")
  @key(fields: "sku")
  @key(fields: "upc") {
  id: ID!
  sku: String!
  upc: String!
  name: String!
}
```

**Nested Keys:**
Keys can include nested fields for complex identification:

```graphql
type ShippingLabel @key(fields: "order { id } sequenceNumber") {
  order: Order!
  sequenceNumber: Int!
  trackingNumber: String!
}
```

**Reference Resolvers:**
Each subgraph with @key must implement a reference resolver:

```typescript
// TypeScript Apollo Server example
const resolvers = {
  Product: {
    __resolveReference(reference: { id: string }) {
      return fetchProductById(reference.id);
    }
  }
}
```

### @external - Field References

**Purpose:** Declares that a field is defined in another subgraph but needed for local resolution.

**When to Use:**
- Field is defined authoritatively in different subgraph
- Local subgraph needs field value for computation
- Always paired with @requires or used in @key

**Example:**
```graphql
# Shipping subgraph
type Product @key(fields: "id") {
  id: ID!
  weight: Float! @external
  dimensions: Dimensions! @external
  shippingCost(destination: String!): Money! @requires(fields: "weight dimensions")
}
```

**Common Mistake:**
```graphql
# ❌ Don't use @external without @requires or in @key
type Product @key(fields: "id") {
  id: ID!
  name: String! @external  # ❌ Not used anywhere, unnecessary
}
```

### @requires - Computed Fields

**Purpose:** Declares that resolving a field requires data from other subgraphs.

**Use Cases:**
- Computing derived values
- Authorization checks requiring external data
- Formatting that depends on external context

**Example - Computed Fields:**
```graphql
# Shipping subgraph
type Product @key(fields: "id") {
  id: ID!
  weight: Float! @external
  dimensions: Dimensions! @external

  # Requires weight and dimensions from Products subgraph
  shippingCost(destination: String!): Money!
    @requires(fields: "weight dimensions")
}
```

**Example - Authorization:**
```graphql
# Documents subgraph
type User @key(fields: "id") {
  id: ID!
  organizationId: ID! @external
  role: Role! @external
}

type Document @key(fields: "id") {
  id: ID!
  content: String!
    @requires(fields: "owner { organizationId role }")
  owner: User!
}
```

**Resolver Implementation:**
```typescript
const resolvers = {
  Document: {
    content(document, args, context) {
      // owner.organizationId and owner.role are guaranteed available
      if (!canAccessDocument(document.owner, document)) {
        throw new ForbiddenError('Insufficient permissions');
      }
      return document.content;
    }
  }
}
```

### @provides - Performance Optimization

**Purpose:** Allows a subgraph to provide fields normally resolved by another subgraph, avoiding extra roundtrips.

**When to Use:**
- Subgraph already has data that another owns
- Can avoid additional query to authoritative subgraph
- Typical in aggregation or caching scenarios

**Example:**
```graphql
# Orders subgraph (stores product snapshots at order time)
type Order @key(fields: "id") {
  id: ID!
  items: [OrderItem!]!
}

type OrderItem {
  product: Product! @provides(fields: "name price")
  quantity: Int!
  capturedPrice: Money!
}

type Product @key(fields: "id") {
  id: ID!
  name: String! @external
  price: Money! @external
}
```

**How It Works:**
1. Orders subgraph stores product name/price snapshot when order created
2. When querying order items, gateway can get name/price from Orders subgraph
3. Avoids roundtrip to Products subgraph if only name/price needed
4. If other product fields requested, gateway still queries Products subgraph

**Important Constraints:**
- @provides is a performance hint, not guarantee
- Data may be stale (snapshot at point in time)
- Authoritative subgraph remains source of truth
- Don't use for frequently changing data

### @shareable - Shared Fields

**Purpose:** Allows multiple subgraphs to define the same field when implementations are identical.

**When to Use:**
- Multiple subgraphs naturally have access to same data
- Implementations are guaranteed identical
- Common for value objects and reference data

**Example:**
```graphql
# Products subgraph
type Product @key(fields: "id") {
  id: ID!
  category: Category! @shareable
}

# Reviews subgraph
type Product @key(fields: "id") {
  id: ID!
  category: Category! @shareable  # Same implementation
}

type Category @shareable {
  id: ID!
  name: String!
  slug: String!
}
```

**When NOT to Use:**
```graphql
# ❌ Don't use @shareable for computed or derived fields
type Product @key(fields: "id") {
  id: ID!
  # Different subgraphs might compute this differently
  popularityScore: Float! @shareable  # ❌ Risky!
}
```

### @override - Migrating Field Ownership

**Purpose:** Safely migrate field ownership from one subgraph to another.

**Migration Process:**
```graphql
# Step 1: Original subgraph (Products)
type Product @key(fields: "id") {
  id: ID!
  inventory: Int!  # Currently owned here
}

# Step 2: New subgraph (Inventory) claims ownership
type Product @key(fields: "id") {
  id: ID!
  inventory: Int! @override(from: "Products")
}

# Step 3: After validation, remove from Products subgraph
```

**Use Cases:**
- Refactoring domain boundaries
- Extracting subgraphs
- Consolidating related functionality

## Performance Optimization

### Query Planning and N+1 Prevention

**Problem:** Federation can amplify N+1 query problems across service boundaries.

**The N+1 Problem in Federation:**
```graphql
query {
  orders {           # 1 query to Orders service
    customer {       # N queries to Users service (one per order)
      name
    }
    items {          # N queries to Products service
      product {      # N*M queries to Products service
        name
      }
    }
  }
}
```

**Solution: DataLoader Pattern**

Every subgraph should use DataLoader or equivalent batching:

```typescript
// User subgraph resolver with DataLoader
import DataLoader from 'dataloader';

const userLoader = new DataLoader(async (userIds: string[]) => {
  // Single batch query instead of N queries
  const users = await db.users.findByIds(userIds);

  // Return in same order as input IDs
  return userIds.map(id => users.find(u => u.id === id));
});

const resolvers = {
  User: {
    __resolveReference(ref: { id: string }) {
      return userLoader.load(ref.id);
    }
  }
};
```

**Entity Reference Batching:**
```typescript
// Product subgraph with batched reference resolution
const resolvers = {
  Product: {
    __resolveReference(references: Array<{ id: string }>) {
      // Gateway sends batch of product IDs
      const ids = references.map(ref => ref.id);
      return productLoader.loadMany(ids);
    }
  }
};
```

### Caching Strategies

**Entity Caching:**
```typescript
// Cache at reference resolver level
const productCache = new LRU({ max: 1000, ttl: 60000 });

const resolvers = {
  Product: {
    async __resolveReference(ref: { id: string }) {
      const cached = productCache.get(ref.id);
      if (cached) return cached;

      const product = await fetchProduct(ref.id);
      productCache.set(ref.id, product);
      return product;
    }
  }
};
```

**Gateway-Level Caching:**
```typescript
// Apollo Gateway with automatic persisted queries
const gateway = new ApolloGateway({
  supergraphSdl,
  buildService({ url }) {
    return new RemoteGraphQLDataSource({
      url,
      // Enable APQ for query result caching
      apq: true,
      // Add CDN cache headers
      willSendRequest({ request, context }) {
        request.http.headers.set('Cache-Control', 'public, max-age=60');
      }
    });
  }
});
```

**HTTP Caching Headers:**
```graphql
# Leverage HTTP caching for stable queries
query GetProduct($id: ID!) @cacheControl(maxAge: 300) {
  product(id: $id) {
    id
    name
    description
  }
}
```

### Connection Pooling

**Problem:** Each subgraph may maintain connections to shared databases.

**Solution: Shared Connection Pools**
```typescript
// Shared database connection pool
import { Pool } from 'pg';

const pool = new Pool({
  host: process.env.DB_HOST,
  database: process.env.DB_NAME,
  max: 20,  // Maximum connections
  idleTimeoutMillis: 30000,
  connectionTimeoutMillis: 2000,
});

// Use pool in resolvers
const resolvers = {
  Query: {
    async product(parent, { id }) {
      const client = await pool.connect();
      try {
        const result = await client.query(
          'SELECT * FROM products WHERE id = $1',
          [id]
        );
        return result.rows[0];
      } finally {
        client.release();
      }
    }
  }
};
```

### Query Complexity Limits

**Prevent Abusive Queries:**
```typescript
import { createComplexityLimitRule } from 'graphql-validation-complexity';

const server = new ApolloServer({
  schema,
  validationRules: [
    createComplexityLimitRule(1000, {
      onCost: (cost) => {
        console.log('Query cost:', cost);
      },
      formatErrorMessage: (cost) =>
        `Query too complex: ${cost}. Maximum allowed: 1000`
    })
  ]
});
```

**Depth Limiting:**
```typescript
import depthLimit from 'graphql-depth-limit';

const server = new ApolloServer({
  schema,
  validationRules: [depthLimit(10)]
});
```

### Monitoring and Tracing

**Distributed Tracing:**
```typescript
import { ApolloServerPluginInlineTrace } from '@apollo/server/plugin/inlineTrace';

const server = new ApolloServer({
  schema,
  plugins: [
    ApolloServerPluginInlineTrace({
      includeErrors: { transform: (err) => err }
    })
  ]
});
```

**Performance Metrics:**
```typescript
// Track resolver performance
const resolvers = {
  Query: {
    async product(parent, args, context, info) {
      const start = Date.now();
      try {
        const result = await fetchProduct(args.id);
        return result;
      } finally {
        const duration = Date.now() - start;
        metrics.recordResolverDuration('Query.product', duration);
      }
    }
  }
};
```

## Versioning and Evolution

### Schema Evolution Principles

**Backward Compatibility Rules:**
1. ✅ **Safe Changes (Non-Breaking):**
   - Adding new types
   - Adding new fields to types
   - Adding new optional arguments to fields
   - Adding new values to enums (with caution)
   - Adding new union members
   - Deprecating fields

2. ❌ **Breaking Changes:**
   - Removing types or fields
   - Renaming types or fields
   - Changing field types
   - Making optional arguments required
   - Removing enum values
   - Removing union members

### Deprecation Strategy

**Gradual Field Deprecation:**
```graphql
type Product @key(fields: "id") {
  id: ID!

  # Old field - deprecated
  price: Float! @deprecated(
    reason: "Use 'pricing.amount' instead. This field will be removed on 2024-06-01."
  )

  # New field - preferred
  pricing: Money!
}
```

**Deprecation Workflow:**
1. **Announce:** Add @deprecated with clear reason and timeline
2. **Monitor:** Track usage of deprecated field
3. **Migrate:** Work with client teams to update queries
4. **Remove:** After grace period and zero usage, remove field

**Monitoring Deprecated Fields:**
```typescript
const server = new ApolloServer({
  schema,
  plugins: [
    {
      async requestDidStart() {
        return {
          async didResolveField({ info }) {
            if (info.parentType && info.fieldName) {
              const field = info.parentType.getFields()[info.fieldName];
              if (field.deprecationReason) {
                metrics.increment('deprecated_field_usage', {
                  field: `${info.parentType.name}.${info.fieldName}`,
                  reason: field.deprecationReason
                });
              }
            }
          }
        };
      }
    }
  ]
});
```

### Field Migration Patterns

**Pattern 1: Parallel Fields**
```graphql
# Phase 1: Add new field alongside old
type User @key(fields: "id") {
  id: ID!
  name: String! @deprecated(reason: "Use firstName and lastName")
  firstName: String!
  lastName: String!
}

# Phase 2: After migration, remove old field
type User @key(fields: "id") {
  id: ID!
  firstName: String!
  lastName: String!
}
```

**Pattern 2: Field Type Evolution**
```graphql
# Phase 1: Original simple type
type Product @key(fields: "id") {
  id: ID!
  price: Float!
}

# Phase 2: Add structured type, deprecate simple
type Product @key(fields: "id") {
  id: ID!
  price: Float! @deprecated(reason: "Use pricing instead")
  pricing: Money!
}

type Money {
  amount: Decimal!
  currency: CurrencyCode!
}

# Phase 3: Remove deprecated field
type Product @key(fields: "id") {
  id: ID!
  pricing: Money!
}
```

**Pattern 3: Entity Refactoring**
```graphql
# Phase 1: Monolithic type
type Order @key(fields: "id") {
  id: ID!
  shippingStreet: String!
  shippingCity: String!
  shippingState: String!
  billingStreet: String!
  billingCity: String!
  billingState: String!
}

# Phase 2: Add structured types, deprecate flat fields
type Order @key(fields: "id") {
  id: ID!
  shippingAddress: Address!
  billingAddress: Address!

  shippingStreet: String! @deprecated(reason: "Use shippingAddress.street")
  shippingCity: String! @deprecated(reason: "Use shippingAddress.city")
  # ... etc
}

# Phase 3: Remove deprecated fields
type Order @key(fields: "id") {
  id: ID!
  shippingAddress: Address!
  billingAddress: Address!
}
```

### Subgraph Versioning

**Versioned Subgraph Deployment:**
```typescript
// gateway-config.ts
const gateway = new ApolloGateway({
  supergraphSdl: new IntrospectAndCompose({
    subgraphs: [
      { name: 'products-v2', url: 'https://products.api/v2/graphql' },
      { name: 'users', url: 'https://users.api/graphql' },
      { name: 'orders', url: 'https://orders.api/graphql' },
    ],
  }),
});
```

**Feature Flags for Schema Changes:**
```typescript
// Feature-flag controlled field resolution
const resolvers = {
  Product: {
    enhancedDescription(product, args, context) {
      if (!context.features.isEnabled('enhanced-descriptions')) {
        return product.description;
      }
      return enrichDescription(product);
    }
  }
};
```

### Schema Checks and CI/CD

**Automated Schema Validation:**
```yaml
# .github/workflows/schema-check.yml
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

      - name: Schema Check
        run: |
          rover subgraph check my-graph@main \
            --name products \
            --schema ./schema.graphql
```

**Breaking Change Detection:**
```bash
# Rover CLI schema validation
rover subgraph check my-graph@production \
  --name products \
  --schema ./products-schema.graphql

# Output:
# ✗ Breaking changes detected:
#   - Field 'Product.price' was removed
#   - Type 'OldCategory' was removed
# ✓ Safe changes:
#   - Field 'Product.rating' was added
```

## Multi-Tenant Federation Patterns

Multi-tenancy in federated GraphQL means a single API instance serves multiple tenant groups with isolated data while sharing infrastructure. Two primary approaches exist, each with distinct trade-offs.

### Approach 1: Transport-Based Multi-Tenancy

**Mechanism:** Tenant information flows through request transport (JWT tokens, HTTP headers, GraphQL context) rather than being explicit in the schema.

**Implementation:**
```typescript
// Gateway context extraction
const server = new ApolloServer({
  gateway,
  context: ({ req }) => {
    const token = req.headers.authorization;
    const decoded = verifyJWT(token);
    return {
      tenantId: decoded.tenantId,
      userId: decoded.userId
    };
  }
});

// Subgraph resolver using context
const resolvers = {
  Query: {
    products: (parent, args, context) => {
      // Tenant extracted from context, not schema
      return fetchProducts(context.tenantId);
    }
  }
};
```

**Advantages:**
- Subgraphs remain tenant-agnostic in schema definition
- Lightweight implementation
- Easy to add tenant info without schema changes
- Supports zero-trust security via JWT validation

**Disadvantages:**
- Analytics/monitoring tools cannot differentiate tenants from schema alone
- Tenant extraction logic duplicated across services (consistency risk)
- Tenant dependencies are implicit, not enforced by schema
- Harder to reason about resolver behavior from schema
- Cannot use GraphQL tooling to validate tenant handling

**When to Use:**
- Simple tenant isolation requirements
- Single team maintaining all subgraphs
- Tenant information is straightforward (single ID)
- Monitoring tenant usage is not critical

### Approach 2: Schema-Based Multi-Tenancy (Recommended)

**Mechanism:** Tenant information becomes a first-class element in the GraphQL schema, typically as an entity that other subgraphs extend.

**Core Pattern - Tenant as Entry Point:**

```graphql
# Identity subgraph (owns Tenant entity)
type Tenant @key(fields: "id") {
  id: ID!
  name: String!
  plan: SubscriptionPlan!
  settings: TenantSettings!
}

type Query {
  # Tenant is the entry point
  currentTenant: Tenant!
}

# Products subgraph (extends Tenant)
type Tenant @key(fields: "id") {
  id: ID!
  products: [Product!]!
  productCategories: [Category!]!
}

type Product @key(fields: "id tenantId") {
  id: ID!
  tenantId: ID!
  name: String!
  sku: String!
}

# Orders subgraph (extends Tenant)
type Tenant @key(fields: "id") {
  id: ID!
  orders(status: OrderStatus): [Order!]!
  recentOrders(limit: Int = 10): [Order!]!
}

type Order @key(fields: "id tenantId") {
  id: ID!
  tenantId: ID!
  items: [OrderItem!]!
  total: Money!
}
```

**Client Usage:**
```graphql
query GetTenantData {
  currentTenant {
    name                    # From Identity subgraph
    plan                    # From Identity subgraph
    products {              # From Products subgraph
      name
      sku
    }
    orders(status: PENDING) {  # From Orders subgraph
      total
      items {
        product {
          name
        }
      }
    }
  }
}
```

**Composite Keys for Multi-Tenant Entities:**
```graphql
# Entity exists within tenant context
type Product @key(fields: "id tenantId") {
  id: ID!
  tenantId: ID!
  name: String!
}

# Reference resolver receives both keys
const resolvers = {
  Product: {
    __resolveReference(ref: { id: string, tenantId: string }) {
      // Both tenant and product ID available
      return fetchProduct(ref.tenantId, ref.id);
    }
  }
};
```

**Complex Tenant Context with Composite Keys:**
```graphql
# Tenant information includes multiple dimensions
type Tenant @key(fields: "id info { country currency }") {
  id: ID!
  info: TenantInfo!
}

type TenantInfo {
  country: CountryCode!
  currency: CurrencyCode!
  timezone: String!
}

# Product pricing varies by tenant context
type Product @key(fields: "id tenant { id info { country currency } }") {
  id: ID!
  tenant: Tenant!
  # Price calculation uses full tenant context
  price: Money!
  taxRate: Float!
}
```

**Advantages:**
- **Declarative and Enforceable:** Schema checks validate tenant handling
- **Explicit Contracts:** Clear expectations between teams
- **Better Tooling:** GraphQL tooling understands tenant relationships
- **Predictable Behavior:** Resolver inputs explicit in schema
- **Centralized Logic:** Tenant validation in one place
- **Observable:** Monitoring tools can track tenant usage from queries
- **Type-Safe:** Compiler/tooling can validate tenant propagation

**Disadvantages:**
- More complex initial setup
- Requires careful key design for complex tenant structures
- Schema changes needed to modify tenant information
- More verbose queries (explicit tenant navigation)

**When to Use (Recommended for Most):**
- Multiple teams maintaining subgraphs
- Governance and safety important at scale
- Complex tenant information (geography, plan, settings)
- Need to monitor per-tenant usage
- Want explicit contracts and validation

### Implementation Best Practices

**1. Tenant Validation Middleware:**
```typescript
// Gateway-level tenant validation
const server = new ApolloServer({
  gateway,
  context: async ({ req }) => {
    const token = req.headers.authorization;
    const decoded = verifyJWT(token);

    // Validate tenant access
    const tenant = await validateTenantAccess(decoded.tenantId, decoded.userId);
    if (!tenant) {
      throw new AuthenticationError('Invalid tenant access');
    }

    return { tenant, userId: decoded.userId };
  }
});
```

**2. Tenant Isolation in Resolvers:**
```typescript
// Products subgraph with tenant isolation
const resolvers = {
  Query: {
    // Schema-based: tenant from arguments
    products: (parent, { tenantId }, context) => {
      // Verify user has access to requested tenant
      if (context.tenant.id !== tenantId) {
        throw new ForbiddenError('Tenant access denied');
      }
      return fetchProducts(tenantId);
    }
  },

  Tenant: {
    __resolveReference(ref: { id: string }) {
      return { id: ref.id };
    },

    products(tenant, args, context) {
      // Tenant ID from entity, validated at gateway
      return fetchProducts(tenant.id);
    }
  }
};
```

**3. Database-Level Tenant Isolation:**
```typescript
// Row-level security enforcement
async function fetchProducts(tenantId: string) {
  // Include tenant in WHERE clause for every query
  return db.products.findMany({
    where: {
      tenantId: tenantId,
      deletedAt: null
    }
  });
}

// Or use database-level row security policies
// PostgreSQL RLS example:
// CREATE POLICY tenant_isolation ON products
//   USING (tenant_id = current_setting('app.current_tenant')::uuid);
```

**4. Tenant Context Propagation:**
```typescript
// Pass tenant context to downstream services
const productsDataSource = {
  async getProduct(id: string, tenantId: string) {
    return this.get(`/products/${id}`, {
      headers: {
        'X-Tenant-ID': tenantId
      }
    });
  }
};
```

### Migration Strategy: Transport to Schema-Based

If migrating from transport-based to schema-based multi-tenancy:

**Phase 1: Add Tenant Entity (Non-Breaking)**
```graphql
# Add Tenant entity without removing existing APIs
type Tenant @key(fields: "id") {
  id: ID!
  name: String!
}

type Query {
  currentTenant: Tenant!
  # Keep existing query working
  products: [Product!]!  # Uses context.tenantId
}
```

**Phase 2: Extend Tenant with Domain Data**
```graphql
type Tenant @key(fields: "id") {
  id: ID!
  name: String!
  # New tenant-scoped API
  products: [Product!]!
}

type Query {
  currentTenant: Tenant!
  # Deprecate old query
  products: [Product!]! @deprecated(reason: "Use currentTenant.products")
}
```

**Phase 3: Remove Deprecated APIs**
```graphql
type Tenant @key(fields: "id") {
  id: ID!
  name: String!
  products: [Product!]!
}

type Query {
  currentTenant: Tenant!
  # products query removed
}
```

### Hybrid Approach

In practice, both approaches can coexist:

```typescript
// Schema-based for core business logic
const resolvers = {
  Tenant: {
    products(tenant, args, context) {
      // Tenant from schema
      return fetchProducts(tenant.id);
    }
  },

  // Transport-based for cross-cutting concerns
  Query: {
    products: (parent, args, context) => {
      // Audit logging uses context
      auditLog.record({
        action: 'QUERY_PRODUCTS',
        tenantId: context.tenant.id,
        userId: context.userId,
        timestamp: new Date()
      });

      // Business logic uses schema
      return fetchProducts(context.tenant.id);
    }
  }
};
```

**Recommendation:** Use schema-based multi-tenancy as the primary pattern for business logic, with transport-based context for auxiliary concerns like logging, tracing, and authentication.

## Anti-Patterns and Common Mistakes

### Anti-Pattern 1: Shared Database Access

**Problem:**
```graphql
# ❌ Multiple subgraphs directly querying same database tables
# Products subgraph queries products table
# Orders subgraph queries products table
# Inventory subgraph queries products table
```

**Why It's Bad:**
- Violates service encapsulation
- Creates tight coupling
- Makes schema evolution difficult
- Breaks federation principles

**Solution:**
```graphql
# ✓ Products subgraph owns product data
# Other subgraphs extend Product entity

# Products subgraph (authoritative)
type Product @key(fields: "id") {
  id: ID!
  name: String!
  sku: String!
}

# Orders subgraph (extends, doesn't query products DB)
type Product @key(fields: "id") {
  id: ID!
  orderedCount: Int!  # From orders database
}
```

### Anti-Pattern 2: Circular Dependencies

**Problem:**
```graphql
# ❌ Users subgraph requires field from Orders
type User @key(fields: "id") {
  id: ID!
  orderCount: Int! @requires(fields: "orders { id }")
  orders: [Order!]! @external  # From Orders subgraph
}

# ❌ Orders subgraph requires field from Users
type Order @key(fields: "id") {
  id: ID!
  customerName: String! @requires(fields: "customer { name }")
  customer: User! @external  # From Users subgraph
}
```

**Why It's Bad:**
- Creates circular query planning
- Poor performance
- Difficult to reason about

**Solution:**
```graphql
# ✓ Each subgraph owns its aggregations

# Users subgraph
type User @key(fields: "id") {
  id: ID!
  name: String!
  # Query orders through gateway, compute in resolver
  orderCount: Int!
}

# Orders subgraph
type Order @key(fields: "id") {
  id: ID!
  customerId: ID!  # Store denormalized
  customerName: String!  # Snapshot at order time
}
```

### Anti-Pattern 3: Chatty Schemas

**Problem:**
```graphql
# ❌ Requiring many small trips across subgraphs
type Order @key(fields: "id") {
  id: ID!
  items: [OrderItem!]!
}

type OrderItem {
  productId: ID!
  product: Product!  # Trip to Products subgraph
  quantity: Int!
  price: Money!
}

type Product @key(fields: "id") {
  id: ID!
  category: Category!  # Trip to Catalog subgraph
}

type Category @key(fields: "id") {
  id: ID!
  department: Department!  # Another trip
}
```

**Why It's Bad:**
- Many sequential roundtrips
- Poor query performance
- Amplifies latency

**Solution:**
```graphql
# ✓ Denormalize when appropriate
type OrderItem {
  productId: ID!
  product: Product! @provides(fields: "name category { name }")
  quantity: Int!
  price: Money!

  # Snapshot data at order time
  productName: String!
  categoryName: String!
}

# Balance: Can still navigate to full Product, but have essentials local
```

### Anti-Pattern 4: Over-Granular Subgraphs

**Problem:**
```graphql
# ❌ Too many tiny subgraphs
# - Product-Name subgraph (just product names)
# - Product-Price subgraph (just prices)
# - Product-Description subgraph (just descriptions)
# - Product-Images subgraph (just images)
```

**Why It's Bad:**
- Excessive coordination overhead
- Poor query performance
- Operational complexity

**Solution:**
```graphql
# ✓ Right-sized subgraphs aligned with bounded contexts
# Single Products subgraph owns all product catalog data
type Product @key(fields: "id") {
  id: ID!
  name: String!
  price: Money!
  description: String!
  images: [Image!]!
}
```

### Anti-Pattern 5: God Subgraph

**Problem:**
```graphql
# ❌ Core subgraph that everything depends on
# Core subgraph owns: Users, Products, Orders, Inventory, Shipping, Reviews
```

**Why It's Bad:**
- Defeats federation purpose
- Single point of failure
- Deployment coupling
- Team scaling bottleneck

**Solution:**
```graphql
# ✓ Properly decomposed domains
# Users subgraph -> User entity
# Products subgraph -> Product entity
# Orders subgraph -> Order entity
# Each independently deployable
```

## Testing Strategies

### Unit Testing Resolvers

```typescript
// products-resolver.test.ts
import { resolvers } from './resolvers';

describe('Product Resolvers', () => {
  describe('__resolveReference', () => {
    it('resolves product by ID', async () => {
      const product = await resolvers.Product.__resolveReference(
        { id: '123' },
        { dataSources: { productsAPI: mockProductsAPI } }
      );

      expect(product).toEqual({
        id: '123',
        name: 'Widget',
        sku: 'WDG-001'
      });
    });

    it('returns null for non-existent product', async () => {
      const product = await resolvers.Product.__resolveReference(
        { id: 'invalid' },
        { dataSources: { productsAPI: mockProductsAPI } }
      );

      expect(product).toBeNull();
    });
  });
});
```

### Integration Testing Subgraphs

```typescript
// subgraph-integration.test.ts
import { ApolloServer } from '@apollo/server';
import { buildSubgraphSchema } from '@apollo/subgraph';
import { typeDefs, resolvers } from './schema';

describe('Products Subgraph Integration', () => {
  let server: ApolloServer;

  beforeAll(() => {
    server = new ApolloServer({
      schema: buildSubgraphSchema({ typeDefs, resolvers })
    });
  });

  it('executes product query', async () => {
    const result = await server.executeOperation({
      query: `
        query GetProduct($id: ID!) {
          product(id: $id) {
            id
            name
            price
          }
        }
      `,
      variables: { id: '123' }
    });

    expect(result.body.kind).toBe('single');
    expect(result.body.singleResult.data).toEqual({
      product: {
        id: '123',
        name: 'Widget',
        price: 19.99
      }
    });
  });

  it('resolves entity reference', async () => {
    const result = await server.executeOperation({
      query: `
        query {
          _entities(representations: [
            { __typename: "Product", id: "123" }
          ]) {
            ... on Product {
              id
              name
            }
          }
        }
      `
    });

    expect(result.body.kind).toBe('single');
    expect(result.body.singleResult.data._entities[0]).toEqual({
      id: '123',
      name: 'Widget'
    });
  });
});
```

### Contract Testing Across Subgraphs

```typescript
// products-contract.test.ts
import { buildSubgraphSchema } from '@apollo/subgraph';
import { printSchema } from 'graphql';

describe('Products Subgraph Contract', () => {
  it('provides required Product fields for Orders subgraph', () => {
    const schema = buildSubgraphSchema({ typeDefs, resolvers });
    const productType = schema.getType('Product');

    // Orders subgraph expects these fields
    expect(productType.getFields()).toHaveProperty('id');
    expect(productType.getFields()).toHaveProperty('name');
    expect(productType.getFields()).toHaveProperty('sku');
  });

  it('maintains Product entity key stability', () => {
    const schema = buildSubgraphSchema({ typeDefs, resolvers });
    const productType = schema.getType('Product');
    const keyDirective = productType
      .astNode
      .directives
      .find(d => d.name.value === 'key');

    expect(keyDirective.arguments[0].value.value).toBe('id');
  });
});
```

### Gateway Composition Testing

```typescript
// gateway-composition.test.ts
import { ApolloGateway } from '@apollo/gateway';
import { readFileSync } from 'fs';

describe('Gateway Composition', () => {
  it('composes all subgraphs without errors', async () => {
    const gateway = new ApolloGateway({
      supergraphSdl: readFileSync('./supergraph.graphql', 'utf-8')
    });

    const { schema, executor } = await gateway.load();
    expect(schema).toBeDefined();
  });

  it('resolves cross-subgraph query', async () => {
    const gateway = new ApolloGateway({
      supergraphSdl: readFileSync('./supergraph.graphql', 'utf-8')
    });

    const { executor } = await gateway.load();

    const result = await executor({
      document: gql`
        query {
          product(id: "123") {
            name              # From Products subgraph
            reviews {         # From Reviews subgraph
              rating
              comment
            }
            stockLevel        # From Inventory subgraph
          }
        }
      `
    });

    expect(result.errors).toBeUndefined();
    expect(result.data.product).toBeDefined();
  });
});
```

## Security Considerations

### Authorization Patterns

**Gateway-Level Authorization:**
```typescript
// Check permissions before query reaches subgraphs
const server = new ApolloServer({
  gateway,
  context: ({ req }) => ({
    user: authenticateUser(req),
    permissions: getUserPermissions(req)
  }),
  validationRules: [
    createAuthorizationRule({
      getUserPermissions: (context) => context.permissions
    })
  ]
});
```

**Subgraph-Level Authorization:**
```typescript
// Products subgraph resolver
const resolvers = {
  Query: {
    product: (parent, { id }, context) => {
      if (!context.user) {
        throw new AuthenticationError('Not authenticated');
      }
      return fetchProduct(id);
    }
  },
  Product: {
    internalNotes: (product, args, context) => {
      if (!context.user.hasRole('ADMIN')) {
        throw new ForbiddenError('Admin access required');
      }
      return product.internalNotes;
    }
  }
};
```

**Field-Level Authorization with @requires:**
```graphql
type Document @key(fields: "id") {
  id: ID!
  title: String!
  owner: User!

  # Content requires owner context for authorization
  content: String! @requires(fields: "owner { id role }")
}
```

### Input Validation

**Schema-Level Validation:**
```graphql
input CreateProductInput {
  name: String! @constraint(minLength: 1, maxLength: 200)
  sku: String! @constraint(pattern: "^[A-Z]{3}-\\d{3}$")
  price: Decimal! @constraint(min: 0, max: 1000000)
}
```

**Resolver-Level Validation:**
```typescript
import { UserInputError } from '@apollo/server/errors';

const resolvers = {
  Mutation: {
    createProduct: (parent, { input }, context) => {
      // Validate business rules
      if (input.price < 0) {
        throw new UserInputError('Price must be positive');
      }

      if (input.name.trim().length === 0) {
        throw new UserInputError('Product name required');
      }

      return productService.create(input);
    }
  }
};
```

### Rate Limiting

**Gateway Rate Limiting:**
```typescript
import rateLimit from 'express-rate-limit';

const limiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 100, // Limit each IP to 100 requests per windowMs
  message: 'Too many requests from this IP'
});

app.use('/graphql', limiter);
```

**Operation-Specific Rate Limiting:**
```typescript
const operationLimits = {
  'CreateOrder': { max: 10, window: 60000 },  // 10 per minute
  'SearchProducts': { max: 100, window: 60000 }  // 100 per minute
};

const server = new ApolloServer({
  gateway,
  plugins: [
    {
      async requestDidStart({ request, context }) {
        const operationName = request.operationName;
        const limit = operationLimits[operationName];

        if (limit) {
          await checkRateLimit(context.user.id, operationName, limit);
        }
      }
    }
  ]
});
```

## Summary Checklist

### Federated Design Principles
- [ ] Types defined as entities only when uniquely identifiable
- [ ] Entity types mirror underlying data source structures
- [ ] Subgraphs organized by domain, not technical layers
- [ ] Collaboration strategy documented (single vs shared ownership)
- [ ] Schema designed as platform, not client-specific
- [ ] Supergraph naming conventions established and followed
- [ ] Subgraph-focused development workflow enabled

### Schema Design
- [ ] Subgraphs align with bounded contexts (DDD)
- [ ] Entity ownership clearly defined
- [ ] Keys are stable, minimal, and efficiently resolvable
- [ ] Value objects distinguished from entities
- [ ] Each field has single authoritative subgraph
- [ ] No client-specific type variations (ProductWeb, ProductMobile, etc.)
- [ ] Types provide generic capabilities for multiple clients
- [ ] Field arguments used for flexibility instead of type proliferation

### Federation Directives
- [ ] All entities have @key directives
- [ ] Reference resolvers implemented for all keys
- [ ] @external used only with @requires or in @key
- [ ] @provides used carefully (data may be stale)
- [ ] @shareable used only for truly identical implementations
- [ ] Circular @requires dependencies avoided

### Multi-Tenancy (if applicable)
- [ ] Multi-tenancy approach chosen and documented
- [ ] Schema-based multi-tenancy preferred for governance
- [ ] Tenant entity serves as entry point if schema-based
- [ ] Composite keys used for multi-tenant entities
- [ ] Tenant validation implemented at gateway
- [ ] Database-level tenant isolation enforced
- [ ] Tenant context properly propagated to downstream services

### Performance
- [ ] DataLoader implemented in all subgraphs
- [ ] Entity resolution batched
- [ ] Appropriate caching strategies
- [ ] Query complexity limits enforced
- [ ] Distributed tracing enabled
- [ ] Connection pooling configured properly

### Evolution
- [ ] Deprecation strategy defined
- [ ] Schema checks in CI/CD pipeline
- [ ] Breaking changes prevented
- [ ] Migration patterns documented
- [ ] Deprecated field usage monitored
- [ ] Field migration paths clearly defined

### Testing
- [ ] Unit tests for all resolvers
- [ ] Integration tests for subgraphs
- [ ] Contract tests across subgraphs
- [ ] Gateway composition tests
- [ ] Performance tests under load
- [ ] Multi-tenant isolation tested (if applicable)

### Security
- [ ] Authentication at gateway
- [ ] Authorization at subgraph level
- [ ] Input validation comprehensive
- [ ] Rate limiting implemented
- [ ] Sensitive data access controlled
- [ ] Tenant isolation security validated (if applicable)

## References and Further Reading

**Official Documentation:**
- [Apollo Federation Documentation](https://www.apollographql.com/docs/federation/)
- [Federation 2 Specification](https://www.apollographql.com/docs/federation/federation-spec/)

**Schema Design Best Practices:**
- [Federated Schema Design Principles (Apollo)](https://www.apollographql.com/blog/federated-schema-design) - Think in entities, domain-oriented design, collaboration strategies
- [Multi-Tenant Federated GraphQL (WunderGraph)](https://wundergraph.com/blog/graphql-schema-design-multi-tenant-federated-graph) - Schema-based vs transport-based multi-tenancy
- [Principled GraphQL](https://principledgraphql.com/) - Integrity, agility, and operations principles
- [GraphQL Schema Design Guide](https://www.apollographql.com/docs/apollo-server/schema/schema/)
- [Domain-Driven Design (Eric Evans)](https://www.domainlanguage.com/ddd/)

**Performance:**
- [DataLoader Documentation](https://github.com/graphql/dataloader)
- [Query Planning in Federation](https://www.apollographql.com/docs/federation/query-plans/)

**Tools:**
- [Rover CLI](https://www.apollographql.com/docs/rover/) - Schema management and validation
- [Apollo Studio](https://studio.apollographql.com/) - Schema registry and governance
- [GraphQL Inspector](https://graphql-inspector.com/) - Schema comparison and validation

---

**Last Updated:** 2026-01-09
**Version:** 1.1.0
