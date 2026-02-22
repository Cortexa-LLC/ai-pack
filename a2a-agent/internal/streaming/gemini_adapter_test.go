package streaming

import (
	"testing"
)

// ---------------------------------------------------------------------------
// GeminiFactory unit tests
// ---------------------------------------------------------------------------

func TestGeminiFactory_GetProviderName(t *testing.T) {
	f := NewGeminiFactory("test-key", 1024)
	if got := f.GetProviderName(); got != ProviderGemini {
		t.Errorf("GetProviderName() = %q, want %q", got, ProviderGemini)
	}
}

func TestGeminiFactory_SupportsModel(t *testing.T) {
	f := NewGeminiFactory("test-key", 1024)

	supported := []string{
		"gemini-2.5-flash-lite",
		"gemini-2.5-flash",
		"gemini-2.5-pro",
		"gemini-3-pro-preview",
		"GEMINI-2.5-FLASH", // case-insensitive
	}
	for _, m := range supported {
		if !f.SupportsModel(m) {
			t.Errorf("SupportsModel(%q) = false, want true", m)
		}
	}

	unsupported := []string{
		"gpt-4o",
		"claude-3-opus-20240229",
		"",
		"llama-3",
	}
	for _, m := range unsupported {
		if f.SupportsModel(m) {
			t.Errorf("SupportsModel(%q) = true, want false", m)
		}
	}
}

// ---------------------------------------------------------------------------
// buildGeminiSchema unit tests
// ---------------------------------------------------------------------------

func TestBuildGeminiSchema_Nil(t *testing.T) {
	schema, err := buildGeminiSchema(nil)
	if err != nil {
		t.Fatalf("buildGeminiSchema(nil) unexpected error: %v", err)
	}
	if schema == nil {
		t.Fatal("expected non-nil schema")
	}
}

func TestBuildGeminiSchema_SimpleObject(t *testing.T) {
	raw := map[string]interface{}{
		"type":        "object",
		"description": "test object",
		"required":    []interface{}{"name"},
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "the name",
			},
			"count": map[string]interface{}{
				"type": "integer",
			},
			"active": map[string]interface{}{
				"type": "boolean",
			},
			"ratio": map[string]interface{}{
				"type": "number",
			},
		},
	}

	schema, err := buildGeminiSchema(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(schema.Properties) != 4 {
		t.Errorf("expected 4 properties, got %d", len(schema.Properties))
	}
	if len(schema.Required) != 1 || schema.Required[0] != "name" {
		t.Errorf("expected required=[name], got %v", schema.Required)
	}
	if schema.Description != "test object" {
		t.Errorf("expected description 'test object', got %q", schema.Description)
	}
}

func TestBuildGeminiSchema_ArrayWithItems(t *testing.T) {
	raw := map[string]interface{}{
		"type": "array",
		"items": map[string]interface{}{
			"type": "string",
		},
	}
	schema, err := buildGeminiSchema(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Items == nil {
		t.Fatal("expected non-nil schema.Items")
	}
}

// ---------------------------------------------------------------------------
// buildGeminiFunctionDeclarations unit tests
// ---------------------------------------------------------------------------

func TestBuildGeminiFunctionDeclarations_Empty(t *testing.T) {
	decls, err := buildGeminiFunctionDeclarations(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decls) != 0 {
		t.Errorf("expected 0 decls, got %d", len(decls))
	}
}

func TestBuildGeminiFunctionDeclarations_Simple(t *testing.T) {
	tools := []Tool{
		{
			Name:        "read_file",
			Description: "Read a file",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string"},
				},
				"required": []interface{}{"path"},
			},
		},
	}
	decls, err := buildGeminiFunctionDeclarations(tools)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(decls))
	}
	if decls[0].Name != "read_file" {
		t.Errorf("expected Name=read_file, got %q", decls[0].Name)
	}
	if decls[0].Description != "Read a file" {
		t.Errorf("unexpected description: %q", decls[0].Description)
	}
}

// ---------------------------------------------------------------------------
// Service.translateModelForProvider — Gemini branch tests
// ---------------------------------------------------------------------------

func newServiceWithGemini(t *testing.T) *Service {
	t.Helper()
	selector := &alwaysDefaultSelector{model: "claude-haiku-4-5", provider: ProviderAnthropic}
	svc := NewService(selector, ProviderAnthropic)
	svc.RegisterProvider(NewGeminiFactory("test-key", 1024))
	return svc
}

func TestTranslateModelForProvider_GeminiToAnthropic(t *testing.T) {
	svc := newServiceWithGemini(t)
	tests := []struct {
		model    string
		expected string
	}{
		{"gemini-2.5-pro", "claude-sonnet-4-6"},
		{"gemini-3-pro-preview", "claude-sonnet-4-6"},
		{"gemini-2.5-flash", "claude-haiku-4-5"},
		{"gemini-2.5-flash-lite", "claude-haiku-4-5"},
	}
	for _, tt := range tests {
		got := svc.translateModelForProvider(tt.model, ProviderGemini, ProviderAnthropic)
		if got != tt.expected {
			t.Errorf("translateModelForProvider(%q, gemini->anthropic) = %q, want %q", tt.model, got, tt.expected)
		}
	}
}

func TestTranslateModelForProvider_GeminiToOpenAI(t *testing.T) {
	svc := newServiceWithGemini(t)
	tests := []struct {
		model    string
		expected string
	}{
		{"gemini-2.5-pro", "gpt-4.1"},
		{"gemini-3-pro-preview", "gpt-4.1"},
		{"gemini-2.5-flash", "gpt-4.1-mini"},
		{"gemini-2.5-flash-lite", "gpt-4.1-mini"},
	}
	for _, tt := range tests {
		got := svc.translateModelForProvider(tt.model, ProviderGemini, ProviderOpenAI)
		if got != tt.expected {
			t.Errorf("translateModelForProvider(%q, gemini->openai) = %q, want %q", tt.model, got, tt.expected)
		}
	}
}

func TestTranslateModelForProvider_GeminiToGemini(t *testing.T) {
	svc := newServiceWithGemini(t)
	model := "gemini-2.5-flash"
	got := svc.translateModelForProvider(model, ProviderGemini, ProviderGemini)
	if got != model {
		t.Errorf("translateModelForProvider(same provider) = %q, want %q", got, model)
	}
}

// ---------------------------------------------------------------------------
// Selector routing tests — gemini models route to gemini provider
// ---------------------------------------------------------------------------

func TestSimpleModelSelector_GeminiRouting(t *testing.T) {
	// Use a SimpleModelSelector with Gemini marked as available.
	selector := &SimpleModelSelector{
		defaultModel:    "claude-haiku-4-5",
		geminiAvailable: true,
	}

	geminiModels := []string{
		"gemini-2.5-flash-lite",
		"gemini-2.5-flash",
		"gemini-2.5-pro",
		"gemini-3-pro-preview",
	}
	for _, m := range geminiModels {
		model, provider, err := selector.SelectModel("engineer", m, 0)
		if err != nil {
			t.Errorf("SelectModel(%q) unexpected error: %v", m, err)
			continue
		}
		if model != m {
			t.Errorf("SelectModel(%q) model = %q, want %q", m, model, m)
		}
		if provider != ProviderGemini {
			t.Errorf("SelectModel(%q) provider = %q, want %q", m, provider, ProviderGemini)
		}
	}
}

func TestSimpleModelSelector_GeminiRouting_NotAvailable(t *testing.T) {
	selector := NewSimpleModelSelector("claude-haiku-4-5", false, nil)
	// geminiAvailable defaults to false

	_, provider, err := selector.SelectModel("engineer", "gemini-2.5-pro", 0)
	if err == nil {
		t.Error("expected error when gemini not available, got nil")
	}
	if provider == ProviderGemini {
		t.Errorf("expected fallback to anthropic, got %q", provider)
	}
}

// alwaysDefaultSelector is a tiny test stub.
type alwaysDefaultSelector struct {
	model    string
	provider string
}

func (s *alwaysDefaultSelector) SelectModel(_ string, requested string, _ int) (string, string, error) {
	if requested != "" {
		return requested, s.provider, nil
	}
	return s.model, s.provider, nil
}
