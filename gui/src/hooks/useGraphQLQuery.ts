import { useQuery, UseQueryResult } from '@tanstack/react-query';
import { graphqlClient } from '../lib/graphql';

/**
 * React Query hook for GraphQL queries
 * @param queryKey - Unique key for caching
 * @param query - GraphQL query string
 * @param variables - Optional query variables
 * @param options - Additional React Query options
 * @returns React Query result object
 */
export function useGraphQLQuery<T = any>(
  queryKey: string,
  query: string,
  variables?: Record<string, any>,
  options?: {
    refetchInterval?: number;
    enabled?: boolean;
  }
): UseQueryResult<T, Error> {
  return useQuery<T, Error>({
    queryKey: [queryKey, variables],
    queryFn: async () => {
      const response = await graphqlClient.query<T>(query, variables);
      return response.data!;
    },
    refetchInterval: options?.refetchInterval,
    enabled: options?.enabled,
    staleTime: 0, // Data is immediately stale
    refetchOnWindowFocus: true, // Refetch when window regains focus
    refetchOnReconnect: true, // Refetch when reconnecting
    structuralSharing: true, // Prevent re-renders if data hasn't actually changed
  });
}
