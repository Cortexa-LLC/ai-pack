package knowledge

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// OpenAIEmbedder implements the Embedder interface using OpenAI's API
type OpenAIEmbedder struct {
	client openai.Client
	model  string
	dims   int
}

// NewOpenAIEmbedder creates an OpenAI embedder.
// dims controls embedding vector size; pass 0 to use the model default
// (1536 for text-embedding-3-small, 3072 for text-embedding-3-large).
func NewOpenAIEmbedder(apiKey, model string, dims int) (*OpenAIEmbedder, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key required (set OPENAI_API_KEY)")
	}

	if model == "" {
		model = "text-embedding-3-small"
	}

	if dims <= 0 {
		// Default dimensions for common models
		switch model {
		case "text-embedding-3-large":
			dims = 3072
		default:
			dims = 1536 // text-embedding-3-small default
		}
	}

	return &OpenAIEmbedder{
		client: openai.NewClient(option.WithAPIKey(apiKey)),
		model:  model,
		dims:   dims,
	}, nil
}

// Embed generates embeddings using OpenAI's API
func (e *OpenAIEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	params := openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{OfArrayOfStrings: texts},
		Model: openai.EmbeddingModel(e.model),
	}
	resp, err := e.client.Embeddings.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("openai embeddings API: %w", err)
	}
	result := make([][]float32, len(resp.Data))
	for i, d := range resp.Data {
		emb := make([]float32, len(d.Embedding))
		for j, v := range d.Embedding {
			emb[j] = float32(v)
		}
		result[i] = emb
	}
	return result, nil
}

func (e *OpenAIEmbedder) Dimensions() int {
	return e.dims
}

func (e *OpenAIEmbedder) Model() string {
	return e.model
}
