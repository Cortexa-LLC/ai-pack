package knowledge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestNewEmbedder(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		model     string
		wantErr   bool
		wantModel string
	}{
		{
			name:      "default to OpenAI",
			apiKey:    "test-key",
			model:     "",
			wantErr:   false,
			wantModel: "text-embedding-3-small",
		},
		{
			name:      "explicit OpenAI model",
			apiKey:    "test-key",
			model:     "text-embedding-3-large",
			wantErr:   false,
			wantModel: "text-embedding-3-large",
		},
		{
			name:      "OpenAI with prefix",
			apiKey:    "test-key",
			model:     "openai:text-embedding-3-small",
			wantErr:   false,
			wantModel: "text-embedding-3-small",
		},
		{
			name:    "OpenAI missing API key",
			apiKey:  "",
			model:   "text-embedding-3-small",
			wantErr: true,
		},
		{
			name:      "Ollama with model name",
			apiKey:    "",
			model:     "ollama:nomic-embed-text",
			wantErr:   false,
			wantModel: "nomic-embed-text",
		},
		{
			name:      "Ollama direct model name",
			apiKey:    "",
			model:     "nomic-embed-text",
			wantErr:   false,
			wantModel: "nomic-embed-text",
		},
		{
			name:      "unsupported model defaults to Ollama",
			apiKey:    "",
			model:     "custom-model",
			wantErr:   false,
			wantModel: "custom-model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			embedder, err := NewEmbedder(tt.apiKey, tt.model, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEmbedder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && embedder.Model() != tt.wantModel {
				t.Errorf("NewEmbedder() model = %v, want %v", embedder.Model(), tt.wantModel)
			}
		})
	}
}

func TestOpenAIEmbedder(t *testing.T) {
	// Mock OpenAI API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/embeddings" {
			http.NotFound(w, r)
			return
		}

		var req struct {
			Input []string `json:"input"`
			Model string   `json:"model"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Generate mock embeddings
		data := make([]map[string]interface{}, len(req.Input))
		for i := range req.Input {
			// Create a simple mock embedding (1536 dimensions for text-embedding-3-small)
			embedding := make([]float64, 1536)
			for j := range embedding {
				embedding[j] = 0.1
			}

			data[i] = map[string]interface{}{
				"object":    "embedding",
				"index":     i,
				"embedding": embedding,
			}
		}

		response := map[string]interface{}{
			"object": "list",
			"data":   data,
			"model":  req.Model,
			"usage": map[string]int{
				"prompt_tokens": 10,
				"total_tokens":  10,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Note: We can't easily mock the OpenAI SDK without modifying the implementation
	// These tests verify the interface but won't make real API calls
	t.Run("basic properties", func(t *testing.T) {
		embedder, err := NewOpenAIEmbedder("test-key", "text-embedding-3-small", 0)
		if err != nil {
			t.Fatalf("NewOpenAIEmbedder() error = %v", err)
		}

		if embedder.Model() != "text-embedding-3-small" {
			t.Errorf("Model() = %v, want text-embedding-3-small", embedder.Model())
		}

		if embedder.Dimensions() != 1536 {
			t.Errorf("Dimensions() = %v, want 1536", embedder.Dimensions())
		}
	})

	t.Run("custom dims", func(t *testing.T) {
		embedder, err := NewOpenAIEmbedder("test-key", "text-embedding-3-small", 512)
		if err != nil {
			t.Fatalf("NewOpenAIEmbedder() error = %v", err)
		}

		if embedder.Dimensions() != 512 {
			t.Errorf("Dimensions() = %v, want 512", embedder.Dimensions())
		}
	})

	t.Run("large model default dims", func(t *testing.T) {
		embedder, err := NewOpenAIEmbedder("test-key", "text-embedding-3-large", 0)
		if err != nil {
			t.Fatalf("NewOpenAIEmbedder() error = %v", err)
		}

		if embedder.Dimensions() != 3072 {
			t.Errorf("Dimensions() = %v, want 3072", embedder.Dimensions())
		}
	})

	// These tests won't actually call the API without a real key
	t.Run("embed single text", func(t *testing.T) {
		embedder, _ := NewOpenAIEmbedder("test-key", "text-embedding-3-small", 0)
		_, err := embedder.Embed(context.Background(), []string{"test"})
		// We expect an error since we don't have a real API key
		if err == nil {
			t.Log("Note: Embedding succeeded (using real API key)")
		}
	})

	t.Run("embed multiple texts", func(t *testing.T) {
		embedder, _ := NewOpenAIEmbedder("test-key", "text-embedding-3-small", 0)
		_, err := embedder.Embed(context.Background(), []string{"test1", "test2"})
		// We expect an error since we don't have a real API key
		if err == nil {
			t.Log("Note: Embedding succeeded (using real API key)")
		}
	})

	t.Run("embed empty list", func(t *testing.T) {
		embedder, _ := NewOpenAIEmbedder("test-key", "text-embedding-3-small", 0)
		result, err := embedder.Embed(context.Background(), []string{})
		if err != nil {
			t.Fatalf("Embed() error = %v", err)
		}
		if len(result) != 0 {
			t.Errorf("Embed() returned %d embeddings, want 0", len(result))
		}
	})
}

func TestOllamaEmbedder(t *testing.T) {
	// Mock Ollama API server using the /api/embed batch endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/api/embed") {
			http.NotFound(w, r)
			return
		}

		var req struct {
			Model string   `json:"model"`
			Input []string `json:"input"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Generate mock embeddings (768 dimensions for nomic-embed-text)
		embeddings := make([][]float64, len(req.Input))
		for i := range req.Input {
			embedding := make([]float64, 768)
			for j := range embedding {
				embedding[j] = 0.1
			}
			embeddings[i] = embedding
		}

		response := map[string]interface{}{
			"model":      req.Model,
			"embeddings": embeddings,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	t.Run("basic properties", func(t *testing.T) {
		embedder := NewOllamaEmbedder("nomic-embed-text", 0)

		if embedder.Model() != "nomic-embed-text" {
			t.Errorf("Model() = %v, want nomic-embed-text", embedder.Model())
		}

		if embedder.Dimensions() != 768 {
			t.Errorf("Dimensions() = %v, want 768", embedder.Dimensions())
		}
	})

	t.Run("custom dims", func(t *testing.T) {
		embedder := NewOllamaEmbedder("nomic-embed-text", 512)
		if embedder.Dimensions() != 512 {
			t.Errorf("Dimensions() = %v, want 512", embedder.Dimensions())
		}
	})

	t.Run("mxbai-embed-large default dims", func(t *testing.T) {
		embedder := NewOllamaEmbedder("mxbai-embed-large", 0)
		if embedder.Dimensions() != 1024 {
			t.Errorf("Dimensions() = %v, want 1024", embedder.Dimensions())
		}
	})

	t.Run("all-minilm default dims", func(t *testing.T) {
		embedder := NewOllamaEmbedder("all-minilm", 0)
		if embedder.Dimensions() != 384 {
			t.Errorf("Dimensions() = %v, want 384", embedder.Dimensions())
		}
	})

	t.Run("embed single text - single batch call", func(t *testing.T) {
		embedder := NewOllamaEmbedder("nomic-embed-text", 0)
		embedder.baseURL = server.URL

		result, err := embedder.Embed(context.Background(), []string{"test"})
		if err != nil {
			t.Fatalf("Embed() error = %v", err)
		}

		if len(result) != 1 {
			t.Fatalf("Embed() returned %d embeddings, want 1", len(result))
		}

		if len(result[0]) != 768 {
			t.Errorf("Embedding has %d dimensions, want 768", len(result[0]))
		}
	})

	t.Run("embed multiple texts - single batch call", func(t *testing.T) {
		callCount := 0
		batchServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			var req struct {
				Model string   `json:"model"`
				Input []string `json:"input"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			embeddings := make([][]float64, len(req.Input))
			for i := range req.Input {
				emb := make([]float64, 768)
				for j := range emb {
					emb[j] = float64(i) * 0.1
				}
				embeddings[i] = emb
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"model":      req.Model,
				"embeddings": embeddings,
			})
		}))
		defer batchServer.Close()

		embedder := NewOllamaEmbedder("nomic-embed-text", 0)
		embedder.baseURL = batchServer.URL

		result, err := embedder.Embed(context.Background(), []string{"text1", "text2", "text3"})
		if err != nil {
			t.Fatalf("Embed() error = %v", err)
		}

		if len(result) != 3 {
			t.Fatalf("Embed() returned %d embeddings, want 3", len(result))
		}

		// Verify only one HTTP call was made (true batching)
		if callCount != 1 {
			t.Errorf("Embed() made %d HTTP calls, want 1 (true batching)", callCount)
		}
	})

	t.Run("embed empty list", func(t *testing.T) {
		embedder := NewOllamaEmbedder("nomic-embed-text", 0)
		result, err := embedder.Embed(context.Background(), []string{})
		if err != nil {
			t.Fatalf("Embed() error = %v", err)
		}
		if len(result) != 0 {
			t.Errorf("Embed() returned %d embeddings, want 0", len(result))
		}
	})

	t.Run("embed uses /api/embed endpoint", func(t *testing.T) {
		var capturedPath string
		endpointServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedPath = r.URL.Path
			var req struct {
				Model string   `json:"model"`
				Input []string `json:"input"`
			}
			json.NewDecoder(r.Body).Decode(&req)
			embeddings := [][]float64{{0.1, 0.2}}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"model":      req.Model,
				"embeddings": embeddings,
			})
		}))
		defer endpointServer.Close()

		embedder := NewOllamaEmbedder("nomic-embed-text", 0)
		embedder.baseURL = endpointServer.URL

		embedder.Embed(context.Background(), []string{"test"})

		if capturedPath != "/api/embed" {
			t.Errorf("Embed() called path %q, want /api/embed", capturedPath)
		}
	})
}

func TestNewEmbedderFromEnv(t *testing.T) {
	tests := []struct {
		name      string
		envModel  string
		envAPIKey string
		wantModel string
		wantErr   bool
	}{
		{
			name:      "default with API key",
			envModel:  "",
			envAPIKey: "test-key",
			wantModel: "text-embedding-3-small",
			wantErr:   false,
		},
		{
			name:      "custom OpenAI model",
			envModel:  "text-embedding-3-large",
			envAPIKey: "test-key",
			wantModel: "text-embedding-3-large",
			wantErr:   false,
		},
		{
			name:      "Ollama model",
			envModel:  "ollama:nomic-embed-text",
			envAPIKey: "",
			wantModel: "nomic-embed-text",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env vars first
			os.Unsetenv("KNOWLEDGE_EMBED_MODEL")
			os.Unsetenv("OPENAI_API_KEY")

			// Set environment variables
			if tt.envModel != "" {
				os.Setenv("KNOWLEDGE_EMBED_MODEL", tt.envModel)
				defer os.Unsetenv("KNOWLEDGE_EMBED_MODEL")
			}
			if tt.envAPIKey != "" {
				os.Setenv("OPENAI_API_KEY", tt.envAPIKey)
				defer os.Unsetenv("OPENAI_API_KEY")
			}

			embedder, err := NewEmbedderFromEnv()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEmbedderFromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && embedder.Model() != tt.wantModel {
				t.Errorf("NewEmbedderFromEnv() model = %v, want %v", embedder.Model(), tt.wantModel)
			}
		})
	}
}
