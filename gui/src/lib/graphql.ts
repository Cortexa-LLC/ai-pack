/**
 * GraphQL client for AI-Pack monitoring server
 * Provides type-safe queries and mutations to the backend GraphQL API
 */

const GRAPHQL_ENDPOINT = '/graphql';

interface GraphQLResponse<T = any> {
  data?: T;
  errors?: Array<{ message: string }>;
}

/**
 * GraphQL client for making queries and mutations
 */
export const graphqlClient = {
  /**
   * Execute a GraphQL query
   * @param query - GraphQL query string
   * @param variables - Optional query variables
   * @returns Promise with query result
   */
  async query<T = any>(query: string, variables?: Record<string, any>): Promise<GraphQLResponse<T>> {
    const response = await fetch(GRAPHQL_ENDPOINT, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        query,
        variables,
      }),
    });

    const result: GraphQLResponse<T> = await response.json();

    if (result.errors && result.errors.length > 0) {
      throw new Error(result.errors[0].message);
    }

    return result;
  },

  /**
   * Execute a GraphQL mutation
   * @param mutation - GraphQL mutation string
   * @param variables - Optional mutation variables
   * @returns Promise with mutation result
   */
  async mutate<T = any>(mutation: string, variables?: Record<string, any>): Promise<GraphQLResponse<T>> {
    return this.query<T>(mutation, variables);
  },
};
