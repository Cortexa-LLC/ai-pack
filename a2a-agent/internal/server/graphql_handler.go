package server

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/graphql"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
)

// SetupGraphQLHandlers sets up GraphQL endpoint and playground
func (s *AgentServer) SetupGraphQLHandlers(mux *http.ServeMux) {
	// Create GraphQL adapter
	adapter := NewGraphQLAdapter(s)

	// Create GraphQL resolver with dependencies
	resolver := graphql.NewResolver(adapter, monitoring.GlobalMetrics)

	// Create executable schema
	schema := graphql.NewExecutableSchema(graphql.Config{
		Resolvers: resolver,
	})

	// Create GraphQL handler
	graphqlHandler := handler.NewDefaultServer(schema)

	// GraphQL endpoint
	mux.Handle("/graphql", graphqlHandler)

	// GraphQL Playground (development UI)
	mux.Handle("/playground", playground.Handler("GraphQL Playground", "/graphql"))

	monitoring.Logger.Info("graphql_configured",
		"endpoint", "/graphql",
		"playground", "/playground")
}
