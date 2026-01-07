# JavaScript/TypeScript Language-Specific Rules

> Based on Microsoft TypeScript Coding Guidelines with 2-space indentation

## Formatting Standards (JavaScript/TypeScript Specific)

**Indentation:** **2 spaces** (no tabs)

This follows the broader JavaScript ecosystem standard (Airbnb, Google, Prettier) while adopting Microsoft's TypeScript coding conventions. Microsoft's TypeScript compiler uses 4 spaces, but we adapt to 2 spaces for consistency with the JavaScript ecosystem.

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

  async getUser(id: number): Promise<UserData | undefined> {
    // 2-space indentation throughout
    // Note: Microsoft prefers undefined over null
    if (!this.users.has(id)) {
      return undefined;
    }

    const user = this.users.get(id);
    return user;
  }

  async createUser(data: Omit<UserData, "id">): Promise<UserData> {
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

This file follows Microsoft's TypeScript coding guidelines with adaptations for the broader JavaScript ecosystem:
- **Microsoft TypeScript Guidelines** - Official TypeScript team conventions
- **2-space indentation** - JavaScript ecosystem standard
- **Type Safety** - Comprehensive TypeScript patterns
- **Modern JavaScript** - ES6+ features and best practices

---

## Quick Standards Summary

### Formatting (Microsoft Standards)
- **Indentation:** 2 spaces (adapted from Microsoft's 4 spaces)
- **Quotes:** **Double quotes** for all strings (`"string"` not `'string'`)
- **Semicolons:** Use them consistently
- **Line Length:** No hard limit, but be reasonable
- **Trailing Commas:** Use them for better diffs

### Naming Conventions (Microsoft Standards)
- `camelCase` - functions, properties, local variables
- `PascalCase` - types, classes, interfaces, enum values
- `UPPER_CASE` - Not specified by Microsoft, but common for constants
- **No `I` prefix for interfaces** (Microsoft explicitly avoids this)
- **No underscore prefix for private properties** (use `private` keyword)
- Use complete words in identifiers

### Variable Declarations (Microsoft Standard)
```typescript
// ✅ One declaration per variable statement
let x = 1;
let y = 2;

// ❌ Avoid multiple declarations
let x = 1, y = 2;
```

### Null vs Undefined (Microsoft Standard)
```typescript
// ✅ Prefer undefined over null
function getUser(id: number): User | undefined {
  return users.get(id);
}

// ❌ Avoid null
function getUser(id: number): User | null {
  return users.get(id) || null;
}
```

### Arrow Functions (Microsoft Standard)
```typescript
// ✅ Prefer arrow functions over anonymous function expressions
setTimeout(() => {
  console.log("Delayed");
}, 1000);

// ❌ Avoid anonymous function expressions
setTimeout(function() {
  console.log("Delayed");
}, 1000);

// ✅ Only parenthesize arrow parameters when necessary
const double = (x: number) => x * 2;
const log = x => console.log(x);  // Single param, no parens needed
```

### Bracing (Microsoft Standard)
```typescript
// ✅ Always surround loop and conditional bodies with curly braces
if (condition) {
  doSomething();
}

// ✅ else goes on a separate line
if (condition) {
  doSomething();
}
else {
  doSomethingElse();
}

// ✅ Opening brace on same line
for (const item of items) {
  process(item);
}

// ❌ Never omit braces
if (condition) doSomething();  // Bad
```

---

## TypeScript Specific Standards

### Interface Naming (Microsoft Standard)
```typescript
// ✅ Do NOT use "I" prefix for interface names
interface User {
  id: number;
  name: string;
}

// ❌ Avoid "I" prefix
interface IUser {  // Bad
  id: number;
  name: string;
}
```

### Type Annotations
```typescript
// ✅ Use explicit types for function signatures
function calculateTotal(items: Item[], taxRate: number): number {
  return items.reduce((sum, item) => sum + item.price, 0) * (1 + taxRate);
}

// ✅ Use interfaces for object shapes
interface Config {
  apiUrl: string;
  timeout: number;
  retries?: number;  // Optional property
}

// ✅ Use type for unions and intersections
type Status = "idle" | "loading" | "success" | "error";
type Result<T> =
  | { success: true; data: T }
  | { success: false; error: string };

// ✅ Use readonly for immutability
interface Point {
  readonly x: number;
  readonly y: number;
}

// ✅ Prefer const assertions for literal types
const config = {
  apiUrl: "https://api.example.com",
  timeout: 5000
} as const;
```

### Exports (Microsoft Standard)
```typescript
// ✅ Do not export types/functions unless sharing across components
class InternalHelper {
  // Only used within this file
}

export class PublicService {
  // Shared across multiple files
}
```

### Type Definitions (Microsoft Standard)
```typescript
// ✅ Type definitions should precede other code within files
type UserId = number;
type UserRole = "admin" | "user" | "guest";

interface User {
  id: UserId;
  role: UserRole;
}

// Then implementation code follows
class UserService {
  // ...
}
```

---

## Modern JavaScript Standards

### Variables
```typescript
// ✅ Use const for immutable bindings
const API_URL = "https://api.example.com";
const user = { name: "Alice" };

// ✅ Use let for mutable bindings
let counter = 0;

// ❌ Never use var
var old = "avoid";  // Bad
```

### Destructuring
```typescript
// ✅ Use destructuring for objects and arrays
const { name, email } = user;
const [first, second, ...rest] = array;

// ✅ Use destructuring in function parameters
function greet({ name, age }: { name: string; age: number }) {
  console.log(`Hello ${name}, you are ${age}`);
}
```

### Spread Operator
```typescript
// ✅ Use spread for array/object composition
const newArray = [...oldArray, newItem];
const newObject = { ...oldObject, updatedField: "value" };

// ✅ Use rest parameters
function sum(...numbers: number[]): number {
  return numbers.reduce((a, b) => a + b, 0);
}
```

### Template Literals
```typescript
// ✅ Use template literals for string interpolation
const message = `Hello, ${user.name}!`;

// ✅ Use for multi-line strings
const query = `
  SELECT id, name, email
  FROM users
  WHERE active = true
`;
```

### Optional Chaining and Nullish Coalescing
```typescript
// ✅ Use optional chaining for nested properties
const bio = user?.profile?.bio;

// ✅ Use nullish coalescing for default values
const displayName = user.name ?? "Guest";

// ⚠️ Remember: undefined is preferred over null
const value = getValue() ?? "default";  // Works with undefined
```

### Async/Await
```typescript
// ✅ Use async/await over Promise chains
async function fetchUserData(id: number): Promise<User> {
  const response = await fetch(`/api/users/${id}`);

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`);
  }

  return response.json();
}

// ✅ Handle errors with try/catch
async function safeLoad(id: number): Promise<User | undefined> {
  try {
    return await fetchUserData(id);
  }
  catch (error) {
    console.error("Failed to load user:", error);
    return undefined;
  }
}
```

### Iteration (Microsoft Standard)
```typescript
// ✅ Use Array methods over loops when practical
const doubled = numbers.map(n => n * 2);
const evens = numbers.filter(n => n % 2 === 0);
const sum = numbers.reduce((a, b) => a + b, 0);

// ✅ Use for...of for iteration
for (const item of items) {
  process(item);
}

// ❌ Avoid for...in for arrays (Microsoft guideline)
for (const key in array) {  // Bad for arrays
  process(array[key]);
}

// ✅ for...in is acceptable for objects
for (const key in object) {
  if (object.hasOwnProperty(key)) {
    process(object[key]);
  }
}
```

---

## Code Organization (Microsoft Standard)

### File Structure
```typescript
// 1. Type definitions at the top
type UserId = number;

interface User {
  id: UserId;
  name: string;
}

// 2. Constants
const MAX_RETRIES = 3;

// 3. Class/function implementations
export class UserService {
  private users = new Map<UserId, User>();

  getUser(id: UserId): User | undefined {
    return this.users.get(id);
  }
}

// 4. Helper functions (unexported)
function formatError(error: Error): string {
  return `Error: ${error.message}`;
}
```

### Immutability (Microsoft Standard)
```typescript
// ✅ Treat objects and arrays as immutable outside creating component
function updateUser(user: User, updates: Partial<User>): User {
  // Return new object instead of mutating
  return { ...user, ...updates };
}

// ❌ Avoid mutation
function updateUser(user: User, updates: Partial<User>): User {
  Object.assign(user, updates);  // Bad: mutates parameter
  return user;
}
```

---

## String Formatting (Microsoft Standard)

All examples use **double quotes**:

```typescript
const greeting = "Hello, world!";
const name = "Alice";
const message = `Welcome, ${name}`;

// Object properties
const config = {
  apiUrl: "https://api.example.com",
  timeout: 5000
};

// Import paths
import { User } from "./models/user";
import { ApiClient } from "../services/api";
```

---

## Summary of Key Microsoft Standards

1. **Double quotes** for all strings
2. **One declaration per variable statement**
3. **Prefer `undefined` over `null`**
4. **No `I` prefix for interfaces**
5. **No underscore prefix** (use `private` keyword)
6. **Arrow functions** preferred over anonymous expressions
7. **Always use curly braces** for control flow
8. **`else` on separate line**
9. **Type definitions precede implementation**
10. **Treat objects/arrays as immutable** outside creating component
11. **Avoid `for...in` for arrays**
12. **Do not export** unless needed across components

---

## TODO: Full JavaScript/TypeScript Guidelines

This file will be expanded to include:
- [ ] Complete Microsoft TypeScript patterns
- [ ] React/Vue/Angular best practices
- [ ] Testing with Jest/Vitest
- [ ] Error handling patterns
- [ ] Advanced async patterns
- [ ] Module system best practices
- [ ] Build tools configuration
- [ ] Common anti-patterns
- [ ] Performance optimization

---

**Standards based on Microsoft TypeScript Coding Guidelines with 2-space indentation for JavaScript ecosystem compatibility.**
