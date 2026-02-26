package knowledge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// OllamaEmbedder implements the Embedder interface using Ollama's local API
type OllamaEmbedder struct {
	client  *http.Client
	baseURL string
	model   string
	dims    int
}

// NewOllamaEmbedder creates an Ollama embedder.
// dims controls embedding vector size; pass 0 to use the model default (768 for
// nomic-embed-text, 1024 for mxbai-embed-large, 384 for all-minilm).
func NewOllamaEmbedder(model string, dims int) *OllamaEmbedder {
	if model == "" {
		model = "nomic-embed-text"
	}

	if dims <= 0 {
		// Default dimensions for common models
		switch model {
		case "mxbai-embed-large":
			dims = 1024
		case "all-minilm":
			dims = 384
		default:
			dims = 768 // nomic-embed-text default
		}
	}

	return &OllamaEmbedder{
		client:  &http.Client{},
		baseURL: "http://localhost:11434",
		model:   model,
		dims:    dims,
	}
}

// Dimensions returns the embedding vector size
func (e *OllamaEmbedder) Dimensions() int {
	return e.dims
}

// Model returns the model name
func (e *OllamaEmbedder) Model() string {
	return e.model
}

// Embed generates embeddings for a batch of texts using Ollama's /api/embed endpoint.
// All texts are sent in a single HTTP request.
func (e *OllamaEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	// Use /api/embed which supports batch input via the "input" field
	reqBody := map[string]interface{}{
		"model": e.model,
		"input": texts,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama embed request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama embed returned status %d", resp.StatusCode)
	}

	var result struct {
		Embeddings [][]float32 `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(result.Embeddings) != len(texts) {
		return nil, fmt.Errorf("expected %d embeddings, got %d", len(texts), len(result.Embeddings))
	}

	return result.Embeddings, nil
}
