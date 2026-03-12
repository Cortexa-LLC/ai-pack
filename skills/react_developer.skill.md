# React Developer
<!-- skills/react_developer.skill.md -->

**Version:** 1.0
**InjectAt:** role_context
**Slot:** 50
**Tools:** Bash, Read, Write, Edit, Grep, Glob
**Gates:** react-best-practices, accessibility-wcag-aa
**MaxExtraTokens:** 5000
**Optional:** true

---

## React Web Development

Platform-specific capabilities for React development with TypeScript, modern hooks, and web tooling.

**Use when:** Building React web apps, working with TypeScript/JavaScript, or web-specific patterns.

---

## Platform Tooling

### Build and Development
```bash
# Start development server
npm run dev
# or
yarn dev

# Build for production
npm run build

# Preview production build
npm run preview

# Type check
npx tsc --noEmit

# Run tests
npm test

# Run tests in watch mode
npm test -- --watch
```

### Code Quality
```bash
# Run ESLint
npx eslint . --ext .ts,.tsx

# Fix ESLint issues
npx eslint . --ext .ts,.tsx --fix

# Run Prettier
npx prettier --check "src/**/*.{ts,tsx}"

# Format with Prettier
npx prettier --write "src/**/*.{ts,tsx}"

# Type check
npx tsc --noEmit
```

### Testing
```bash
# Run Jest tests
npm test

# Run tests with coverage
npm test -- --coverage

# Run specific test file
npm test -- UserForm.test.tsx

# Run tests in watch mode
npm test -- --watch
```

### Performance Auditing
```bash
# Run Lighthouse audit
npx lighthouse http://localhost:3000 --view

# Check bundle size
npx vite-bundle-visualizer
# or for webpack:
npx webpack-bundle-analyzer build/bundle-stats.json
```

---

## React Patterns and Best Practices

### Hooks Patterns

**State Management:**
```typescript
// ✅ Use useState for component state
const [count, setCount] = useState(0);

// ✅ Use lazy initialization for expensive initial state
const [data, setData] = useState(() => expensiveComputation());

// ✅ Functional updates when depending on previous state
setCount(prev => prev + 1);

// ❌ Avoid: Direct state mutation
setItems(items.push(newItem));  // WRONG
// ✅ Correct:
setItems([...items, newItem]);
```

**useEffect Dependencies:**
```typescript
// ✅ Include all dependencies
useEffect(() => {
    fetchData(userId);
}, [userId]);  // userId is used, so it's in deps

// ✅ Use useCallback/useMemo to stabilize references
const fetchData = useCallback(async (id: string) => {
    const data = await api.get(id);
    setData(data);
}, []);  // fetchData is now stable

useEffect(() => {
    fetchData(userId);
}, [fetchData, userId]);

// ❌ Avoid: Missing dependencies (ESLint will warn)
useEffect(() => {
    fetchData(userId);
}, []);  // WRONG: userId missing

// ❌ Avoid: Disabling exhaustive-deps rule
// eslint-disable-next-line react-hooks/exhaustive-deps
```

**Custom Hooks:**
```typescript
// ✅ Extract reusable logic into custom hooks
function useDebounce<T>(value: T, delay: number): T {
    const [debouncedValue, setDebouncedValue] = useState(value);

    useEffect(() => {
        const timer = setTimeout(() => setDebouncedValue(value), delay);
        return () => clearTimeout(timer);
    }, [value, delay]);

    return debouncedValue;
}

// Usage:
const debouncedSearch = useDebounce(searchTerm, 500);
```

### Component Patterns

**Composition Over Props:**
```typescript
// ✅ Use children for flexible composition
interface CardProps {
    children: React.ReactNode;
}

function Card({ children }: CardProps) {
    return <div className="card">{children}</div>;
}

// Usage:
<Card>
    <CardHeader />
    <CardBody />
    <CardFooter />
</Card>

// ❌ Avoid: Many optional props for variations
interface CardProps {
    showHeader?: boolean;
    showFooter?: boolean;
    headerContent?: React.ReactNode;
    // ... many more props
}
```

**Render Props / Compound Components:**
```typescript
// ✅ Use compound components for related UI
const Select = ({ children }: { children: React.ReactNode }) => {
    return <div className="select">{children}</div>;
};

Select.Option = ({ value, children }: { value: string; children: React.ReactNode }) => {
    return <option value={value}>{children}</option>;
};

// Usage:
<Select>
    <Select.Option value="1">One</Select.Option>
    <Select.Option value="2">Two</Select.Option>
</Select>
```

### Performance Optimization

**Memoization:**
```typescript
// ✅ Use React.memo for expensive components
const ExpensiveComponent = React.memo(({ data }: Props) => {
    return <div>{/* expensive rendering */}</div>;
});

// ✅ Use useMemo for expensive computations
const sortedItems = useMemo(() => {
    return items.sort((a, b) => a.name.localeCompare(b.name));
}, [items]);

// ✅ Use useCallback for stable function references
const handleClick = useCallback(() => {
    doSomething(value);
}, [value]);

// ❌ Avoid: Overusing memo (profile first)
// Not every component needs React.memo
```

**Code Splitting:**
```typescript
// ✅ Lazy load routes and heavy components
const Dashboard = lazy(() => import('./Dashboard'));
const Settings = lazy(() => import('./Settings'));

function App() {
    return (
        <Suspense fallback={<Loading />}>
            <Routes>
                <Route path="/dashboard" element={<Dashboard />} />
                <Route path="/settings" element={<Settings />} />
            </Routes>
        </Suspense>
    );
}
```

### Type Safety

**Component Props:**
```typescript
// ✅ Use interface for component props
interface ButtonProps {
    onClick: () => void;
    children: React.ReactNode;
    variant?: 'primary' | 'secondary';
    disabled?: boolean;
}

function Button({ onClick, children, variant = 'primary', disabled }: ButtonProps) {
    return (
        <button
            onClick={onClick}
            className={`btn btn-${variant}`}
            disabled={disabled}
        >
            {children}
        </button>
    );
}

// ✅ Use generic types for reusable components
interface ListProps<T> {
    items: T[];
    renderItem: (item: T) => React.ReactNode;
}

function List<T>({ items, renderItem }: ListProps<T>) {
    return <ul>{items.map(renderItem)}</ul>;
}
```

**Event Handlers:**
```typescript
// ✅ Use React's event types
function Form() {
    const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        // ...
    };

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setValue(e.target.value);
    };

    return (
        <form onSubmit={handleSubmit}>
            <input onChange={handleChange} />
        </form>
    );
}
```

---

## Accessibility (WCAG AA)

### Semantic HTML
```typescript
// ✅ Use semantic HTML elements
<header>
    <nav>
        <ul>
            <li><a href="/">Home</a></li>
        </ul>
    </nav>
</header>

// ❌ Avoid: Divs for everything
<div className="header">
    <div className="nav">
        <div className="link">Home</div>
    </div>
</div>
```

### ARIA Attributes
```typescript
// ✅ Use ARIA labels and roles
<button aria-label="Close dialog" onClick={onClose}>
    <XIcon />
</button>

// ✅ Use aria-describedby for form fields
<>
    <input
        id="email"
        type="email"
        aria-describedby="email-error"
    />
    <span id="email-error" role="alert">
        Invalid email format
    </span>
</>

// ✅ Use role for custom components
<div
    role="dialog"
    aria-modal="true"
    aria-labelledby="dialog-title"
>
    <h2 id="dialog-title">Confirm Action</h2>
</div>
```

### Keyboard Navigation
```typescript
// ✅ Support keyboard interaction
function Menu() {
    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Escape') {
            closeMenu();
        }
        if (e.key === 'Enter' || e.key === ' ') {
            selectItem();
        }
    };

    return (
        <div
            role="menu"
            onKeyDown={handleKeyDown}
            tabIndex={0}
        >
            {/* menu items */}
        </div>
    );
}
```

### Focus Management
```typescript
// ✅ Manage focus for modals/dialogs
function Dialog({ isOpen }: { isOpen: boolean }) {
    const dialogRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (isOpen) {
            dialogRef.current?.focus();
        }
    }, [isOpen]);

    return (
        <div
            ref={dialogRef}
            role="dialog"
            tabIndex={-1}
            aria-modal="true"
        >
            {/* dialog content */}
        </div>
    );
}
```

---

## Common React Anti-Patterns to Avoid

| Anti-Pattern | Risk | Fix |
|--------------|------|-----|
| **Missing useEffect dependencies** | Stale closures, bugs | Include all deps or use ESLint rule |
| **Mutating state directly** | State not updating | Use setState with new object/array |
| **Inline object/function in JSX** | Unnecessary re-renders | Extract to variable or use memo/callback |
| **No key prop in lists** | Incorrect rendering, perf issues | Use stable unique keys (not index) |
| **useEffect without cleanup** | Memory leaks, race conditions | Return cleanup function |
| **Too many useState** | Hard to manage | Consider useReducer or external state |
| **No error boundaries** | White screen on error | Add ErrorBoundary components |
| **Missing alt text on images** | Screen readers can't describe | Always add `alt` attribute |

---

## Testing Guidelines

### Component Tests
```typescript
// ✅ Test user interactions
import { render, screen, fireEvent } from '@testing-library/react';

test('increments counter on button click', () => {
    render(<Counter />);

    const button = screen.getByRole('button', { name: /increment/i });
    fireEvent.click(button);

    expect(screen.getByText('Count: 1')).toBeInTheDocument();
});

// ✅ Test accessibility
test('form has proper labels', () => {
    render(<LoginForm />);

    expect(screen.getByLabelText('Email')).toBeInTheDocument();
    expect(screen.getByLabelText('Password')).toBeInTheDocument();
});
```

### Hook Tests
```typescript
// ✅ Test custom hooks
import { renderHook, act } from '@testing-library/react';

test('useCounter increments', () => {
    const { result } = renderHook(() => useCounter());

    act(() => {
        result.current.increment();
    });

    expect(result.current.count).toBe(1);
});
```

---

## Build Configuration

### Vite Config
```typescript
// vite.config.ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
    plugins: [react()],
    build: {
        rollupOptions: {
            output: {
                manualChunks: {
                    'react-vendor': ['react', 'react-dom'],
                    'router': ['react-router-dom'],
                },
            },
        },
    },
});
```

### TypeScript Config
```json
{
    "compilerOptions": {
        "target": "ES2020",
        "lib": ["ES2020", "DOM", "DOM.Iterable"],
        "jsx": "react-jsx",
        "strict": true,
        "noUncheckedIndexedAccess": true,
        "noImplicitReturns": true,
        "noFallthroughCasesInSwitch": true
    }
}
```

---

## Integration with Other Skills

- **Uses:** general (core capabilities)
- **Complements:** code_review (React-specific patterns)
- **Used by roles:** engineer, designer
- **Gates enforced:** react-best-practices, accessibility-wcag-aa

---

## Example Usage

```
User: "Build and test the React app"

Agent (with react_developer skill):
1. Runs ESLint for code quality
2. Runs TypeScript type checking
3. Builds production bundle
4. Runs Jest tests
5. Runs Lighthouse accessibility audit
6. Reports results with React best practice violations
```

---

## Notes

- **Prefer functional components** with hooks over class components
- **Use TypeScript** for type safety
- **Follow React 18+ patterns** (concurrent features, automatic batching)
- **Test with React Testing Library** (user-centric tests)
- **Profile with React DevTools** for performance issues
- **Support modern browsers** (ES2020+), use polyfills sparingly
