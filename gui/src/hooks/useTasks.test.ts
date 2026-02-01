import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useTasks } from './useTasks';
import React from 'react';

// Mock graphql client
vi.mock('../lib/graphql', () => ({
  graphqlClient: {
    query: vi.fn(),
  },
}));

import { graphqlClient } from '../lib/graphql';

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });
  return ({ children }: { children: React.ReactNode }) => (
    React.createElement(QueryClientProvider, { client: queryClient }, children)
  );
};

describe('useTasks', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch tasks data', async () => {
    const mockTasks = {
      tasks: [
        {
          taskID: 'task-1',
          role: 'engineer',
          task: 'Fix bug',
          status: 'running',
          createdAt: '2026-01-31T12:00:00Z',
          updatedAt: '2026-01-31T12:30:00Z',
          completedAt: null,
          result: null,
          error: null,
          metadata: {},
          beadsTaskID: 'bd-123',
        },
      ],
    };

    (graphqlClient.query as any).mockResolvedValueOnce({ data: mockTasks });

    const { result } = renderHook(() => useTasks(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockTasks);
    expect(result.current.data?.tasks).toHaveLength(1);
  });

  it('should handle empty task list', async () => {
    const mockTasks = { tasks: [] };
    (graphqlClient.query as any).mockResolvedValueOnce({ data: mockTasks });

    const { result } = renderHook(() => useTasks(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.tasks).toHaveLength(0);
  });
});
