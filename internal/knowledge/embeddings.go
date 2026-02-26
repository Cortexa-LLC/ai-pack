package knowledge

import (
	"context"
	"fmt"
	"strings"
)

// BatchEmbed processes all un-embedded entities and observations in a project.
// HNSW cache invalidation: the in-memory HNSW vector index is keyed on entity
// embeddings only (see buildIndex). After updating entity embeddings we call
// s.hnswIdx.invalidate(projectID) exactly once so the next VectorSearch
// rebuilds the index with the new vectors.
// Observation embeddings do not affect the entity-based HNSW index, so no
// action is necessary for observations.

// ...
