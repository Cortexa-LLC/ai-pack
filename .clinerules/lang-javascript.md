# JavaScript/TypeScript Language-Specific Rules

> Based on Airbnb JavaScript Style Guide, Google JavaScript Style Guide, and TypeScript best practices

## Formatting Standards (JavaScript/TypeScript Specific)

**Indentation:** **2 spaces** (no tabs)

This follows Airbnb, Google, Prettier defaults, and the overwhelming majority of the JavaScript/TypeScript ecosystem.

**Note:** Cortexa LLC uses **language-specific indentation standards**:
- **C++**: 2 spaces (see lang-cpp.md)
- **Python**: 4 spaces (see lang-python.md)
- **JavaScript/TypeScript**: 2 spaces (this document)
- **Java**: 4 spaces (see lang-java.md)
- **Kotlin**: 4 spaces (see lang-kotlin.md)

**Example (TypeScript):**
```typescript
interface UserData {
  id: number;
  name: string;
  email: string;
}

class UserService {
  private users: Map<number, UserData> = new Map();

  async getUser(id: number): Promise<UserData | null> {
    // 2-space indentation throughout
    if (!this.users.has(id)) {
      return null;
    }

    const user = this.users.get(id);
    return user ?? null;
  }

  async createUser(data: Omit<UserData, 'id'>): Promise<UserData> {
    const id = this.generateId();
    const user: UserData = {
      id,
      ...data
    };

    this.users.set(id, user);
    return user;
  }

  private generateId(): number {
    return Math.floor(Math.random() * 1000000);
  }
}
```

---

## Overview

This file will contain JavaScript/TypeScript-specific best practices including:
- **Airbnb JavaScript Style Guide** - Industry standard
- **Google JavaScript Style Guide** - Google's conventions
- **TypeScript Best Practices** - Type-safe JavaScript
- **ESLint Rules** - Recommended configurations
- **Prettier** - Automated formatting
- **React/Vue/Angular** - Framework-specific patterns

---

## Quick Standards Summary

### Formatting
- **Indentation:** 2 spaces
- **Semicolons:** Use them (avoid ASI pitfalls)
- **Quotes:** Single quotes for strings, double quotes for JSX attributes
- **Line Length:** 100 characters (soft limit)
- **Trailing Commas:** Use them (better diffs)

### Naming
- `camelCase` - variables, functions
- `PascalCase` - classes, interfaces, types, React components
- `UPPER_CASE` - constants
- `_private` - prefix for private (TypeScript: use `private` keyword)

### TypeScript Specific
```typescript
// Use explicit types for function signatures
function calculateTotal(items: Item[], taxRate: number): number {
  return items.reduce((sum, item) => sum + item.price, 0) * (1 + taxRate);
}

// Use interfaces for object shapes
interface Config {
  apiUrl: string;
  timeout: number;
  retries?: number;  // Optional property
}

// Use type for unions and intersections
type Status = 'idle' | 'loading' | 'success' | 'error';
type Result<T> = { success: true; data: T } | { success: false; error: string };

// Use readonly for immutability
interface Point {
  readonly x: number;
  readonly y: number;
}

// Prefer const assertions for literal types
const config = {
  apiUrl: 'https://api.example.com',
  timeout: 5000
} as const;
```

### Modern JavaScript
```javascript
// Use const/let, never var
const immutableValue = 42;
let mutableValue = 0;

// Use arrow functions for callbacks
const doubled = numbers.map(n => n * 2);

// Use destructuring
const { name, email } = user;
const [first, second, ...rest] = array;

// Use spread operator
const newArray = [...oldArray, newItem];
const newObject = { ...oldObject, updatedField: 'value' };

// Use template literals
const message = `Hello, ${user.name}!`;

// Use optional chaining and nullish coalescing
const value = user?.profile?.bio ?? 'No bio available';

// Use async/await over Promise chains
async function fetchUserData(id: number): Promise<User> {
  const response = await fetch(`/api/users/${id}`);
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`);
  }
  return response.json();
}
```

---

## TODO: Full JavaScript/TypeScript Guidelines

This file will be expanded to include:
- [ ] Complete Airbnb style guide coverage
- [ ] TypeScript advanced patterns
- [ ] React/Vue/Angular best practices
- [ ] Testing with Jest/Vitest
- [ ] Error handling patterns
- [ ] Async patterns and Promises
- [ ] Module system (ESM vs CommonJS)
- [ ] Build tools configuration
- [ ] Common anti-patterns
- [ ] Performance optimization

---

**For now, always use 2-space indentation. Full guidelines coming soon.**
