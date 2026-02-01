import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useGraphQLQuery } from './useGraphQLQuery';
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

describe('useGraphQLQuery', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch data successfully', async () => {
    const mockData = { health: { status: 'ok' } };
    (graphqlClient.query as any).mockResolvedValueOnce({ data: mockData });

    const { result } = renderHook(
      () => useGraphQLQuery('health', '{ health { status } }'),
      { wrapper: createWrapper() }
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockData);
  });

  it('should handle errors', async () => {
    (graphqlClient.query as any).mockRejectedValueOnce(new Error('Query failed'));

    const { result } = renderHook(
      () => useGraphQLQuery('health', '{ health { status } }'),
      { wrapper: createWrapper() }
    );

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toBeTruthy();
  });

  it('should show loading state', () => {
    (graphqlClient.query as any).mockImplementation(() => new Promise(() => {}));

    const { result } = renderHook(
      () => useGraphQLQuery('health', '{ health { status } }'),
      { wrapper: createWrapper() }
    );

    expect(result.current.isLoading).toBe(true);
  });
});
