import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import App from './App';

// Mock hooks
vi.mock('./hooks/useMetrics', () => ({
  useMetrics: () => ({
    data: {
      metrics: {
        tasksSpawned: 10,
        tasksCompleted: 8,
        tasksFailed: 1,
        tasksActive: 1,
        averageDurationMs: 1500.5,
        tokenUsage: {
          totalTokens: 5000,
          inputTokens: 3000,
          outputTokens: 2000,
        },
        apiCalls: {
          total: 50,
          success: 48,
          failed: 2,
        },
        performance: {
          cpuUsage: 25.5,
          memoryUsageMB: 512.0,
          goroutines: 10,
          uptime: '2h30m',
        },
      },
    },
    isLoading: false,
    isError: false,
  }),
}));

vi.mock('./hooks/useTasks', () => ({
  useTasks: () => ({
    data: {
      tasks: [
        {
          taskID: 'task-1',
          role: 'engineer',
          task: 'Build feature',
          status: 'running',
          progress: 0.5,
          createdAt: '2026-01-31T12:00:00Z',
          updatedAt: '2026-01-31T12:30:00Z',
          completedAt: null,
          result: null,
          error: null,
          metadata: null,
          beadsTaskID: 'bd-123',
        },
      ],
    },
    isLoading: false,
    isError: false,
  }),
}));

const queryClient = new QueryClient();

describe('App', () => {
  it('should render app title', () => {
    render(
      <QueryClientProvider client={queryClient}>
        <App />
      </QueryClientProvider>
    );
    
    expect(screen.getByText(/AI-Pack Monitor/i)).toBeInTheDocument();
  });

  it('should render metrics section', () => {
    render(
      <QueryClientProvider client={queryClient}>
        <App />
      </QueryClientProvider>
    );
    
    expect(screen.getByText('Tasks Spawned')).toBeInTheDocument();
    expect(screen.getByText('Tasks Completed')).toBeInTheDocument();
    expect(screen.getByText('8')).toBeInTheDocument(); // Tasks Completed value
  });

  it('should render tasks section', () => {
    render(
      <QueryClientProvider client={queryClient}>
        <App />
      </QueryClientProvider>
    );
    
    expect(screen.getByText('Build feature')).toBeInTheDocument();
    expect(screen.getByText('engineer')).toBeInTheDocument();
  });
});
