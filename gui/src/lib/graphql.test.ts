import { describe, it, expect, vi, beforeEach } from 'vitest';
import { graphqlClient } from './graphql';

// Mock fetch
(globalThis as any).fetch = vi.fn();

describe('graphqlClient', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should send GraphQL queries', async () => {
    const mockResponse = { data: { health: { status: 'ok' } } };
    const mockFetch = globalThis.fetch as any;
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponse,
    });

    const result = await graphqlClient.query(`
      query {
        health {
          status
        }
      }
    `);

    expect(result).toEqual(mockResponse);
    expect(mockFetch).toHaveBeenCalledWith(
      '/graphql',
      expect.objectContaining({
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      })
    );
  });

  it('should handle GraphQL errors', async () => {
    const mockError = { errors: [{ message: 'Query failed' }] };
    const mockFetch = globalThis.fetch as any;
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockError,
    });

    await expect(
      graphqlClient.query('{ invalid }')
    ).rejects.toThrow('Query failed');
  });

  it('should handle network errors', async () => {
    const mockFetch = globalThis.fetch as any;
    mockFetch.mockRejectedValueOnce(new Error('Network error'));

    await expect(
      graphqlClient.query('{ health }')
    ).rejects.toThrow('Network error');
  });
});
