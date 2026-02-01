# AI-Pack Monitoring GUI

A modern React-based monitoring dashboard for AI-Pack orchestration system.

## Features

- **Real-time Metrics Dashboard**
  - Task statistics (spawned, completed, failed, active)
  - Token usage tracking (total, input, output)
  - API call monitoring (total, successful, failed)
  - Performance metrics (CPU, memory, goroutines, uptime)
  - Average task duration

- **Active Tasks View**
  - Real-time task status
  - Progress tracking
  - Agent role visualization
  - Task creation timestamps

## Tech Stack

- **React 18** with TypeScript
- **Vite** for fast development and builds
- **TailwindCSS** for styling
- **React Query** for data fetching and caching
- **Vitest** for testing
- **Testing Library** for component testing

## Getting Started

### Prerequisites

- Node.js 18+ and npm
- AI-Pack server running on `http://localhost:8080`

### Installation

```bash
npm install
```

### Development

```bash
npm run dev
```

Opens the app at `http://localhost:5173`

### Testing

```bash
# Run all tests
npm test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:coverage
```

### Build

```bash
npm run build
```

Outputs production-ready files to `dist/`

## Project Structure

```
gui/
├── src/
│   ├── components/       # React components
│   │   ├── MetricsCard.tsx
│   │   └── TaskList.tsx
│   ├── hooks/            # Custom React hooks
│   │   ├── useGraphQLQuery.ts
│   │   ├── useMetrics.ts
│   │   └── useTasks.ts
│   ├── lib/              # Utilities
│   │   └── graphql.ts
│   ├── types/            # TypeScript types
│   │   └── metrics.ts
│   ├── App.tsx           # Main application
│   └── main.tsx          # Entry point
├── tests/                # Test files
└── dist/                 # Production build
```

## API Integration

The GUI communicates with the AI-Pack server via GraphQL at `/graphql` endpoint:

### Metrics Query
```graphql
query {
  metrics {
    tasksSpawned
    tasksCompleted
    tasksFailed
    tasksActive
    tokenUsage {
      totalTokens
      inputTokens
      outputTokens
    }
    apiCalls {
      total
      success
      failed
    }
    performance {
      cpuUsage
      memoryUsageMB
      goroutines
      uptime
    }
    averageDurationMs
  }
}
```

### Tasks Query
```graphql
query {
  tasks {
    id
    beadsId
    agentRole
    status
    progress
    createdAt
  }
}
```

## Configuration

- **API Endpoint**: Configured in `src/lib/graphql.ts` (default: `/graphql`)
- **Refresh Interval**: Configured in hooks (default: 5 seconds)
- **Query Options**: Configured in `src/main.tsx`

## Testing

The project uses Test-Driven Development (TDD) approach:

- **Unit Tests**: All components and hooks have comprehensive tests
- **Integration Tests**: App component tests verify full integration
- **Coverage**: 80-90% code coverage target

Test files are co-located with source files using `.test.tsx` or `.test.ts` suffix.

## Development Guidelines

1. **Follow TDD**: Write tests before implementation
2. **Type Safety**: Use TypeScript for all code
3. **Clean Code**: Follow React best practices
4. **Error Handling**: Handle loading, error, and empty states
5. **Documentation**: Document complex logic and APIs

## License

Part of AI-Pack orchestration system.
