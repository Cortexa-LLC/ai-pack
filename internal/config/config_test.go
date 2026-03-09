package config

import (
	"os"
	"path/filepath"
	"testing"
)

// Helper to clear environment variables for test isolation
func clearEnvVars(t *testing.T) func() {
	// Save current values
	savedVars := map[string]string{
		"ANTHROPIC_BASE_URL": os.Getenv("ANTHROPIC_BASE_URL"),
		"ANTHROPIC_MODEL":    os.Getenv("ANTHROPIC_MODEL"),
		"SERVER_HOST":        os.Getenv("SERVER_HOST"),
		"SERVER_PORT":        os.Getenv("SERVER_PORT"),
		"API_MODE":           os.Getenv("API_MODE"),
	}

	// Clear all config-related env vars
	os.Unsetenv("ANTHROPIC_BASE_URL")
	os.Unsetenv("ANTHROPIC_MODEL")
	os.Unsetenv("SERVER_HOST")
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("API_MODE")

	// Return cleanup function
	return func() {
		for key, value := range savedVars {
			if value != "" {
				os.Setenv(key, value)
			}
		}
	}
}

func TestLoadConfigValidFile(t *testing.T) {
	cleanup := clearEnvVars(t)
	defer cleanup()

	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.json")

	configContent := `{
		"server": {
			"host": "localhost",
			"port": 9000,
			"max_concurrent_agents": 5
		},
		"api": {
			"anthropic_model": "claude-3-5-sonnet-20241022",
			"max_tokens": 4000,
			"mode": "direct"
		}
	}`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	cfg, err := LoadConfig(configPath)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if cfg.Server.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", cfg.Server.Host)
	}
	if cfg.Server.Port != 9000 {
		t.Errorf("Expected port 9000, got %d", cfg.Server.Port)
	}
	if cfg.API.MaxTokens != 4000 {
		t.Errorf("Expected max_tokens 4000, got %d", cfg.API.MaxTokens)
	}
	if cfg.API.Mode != "direct" {
		t.Errorf("Expected mode 'direct', got '%s'", cfg.API.Mode)
	}
}

func TestLoadConfigProxyMode(t *testing.T) {
	cleanup := clearEnvVars(t)
	defer cleanup()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "proxy-config.json")

	configContent := `{
		"api": {
			"mode": "proxy",
			"proxy": {
				"base_url": "https://proxy.example.com/api"
			}
		}
	}`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	cfg, err := LoadConfig(configPath)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if cfg.API.Mode != "proxy" {
		t.Errorf("Expected mode 'proxy', got '%s'", cfg.API.Mode)
	}
	if cfg.API.Proxy == nil {
		t.Fatal("Expected proxy config to be set")
	}
	if cfg.API.Proxy.BaseURL != "https://proxy.example.com/api" {
		t.Errorf("Expected proxy URL 'https://proxy.example.com/api', got '%s'", cfg.API.Proxy.BaseURL)
	}
}

func TestLoadConfigNonExistentFile(t *testing.T) {
	cleanup := clearEnvVars(t)
	defer cleanup()

	// Try to load non-existent file - should use defaults
	cfg, err := LoadConfig("/nonexistent/config.json")

	// Assert - should NOT error, should use defaults
	if err != nil {
		t.Errorf("LoadConfig should use defaults when file doesn't exist, got error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Expected default config, got nil")
	}
	// Verify defaults are applied
	if cfg.Server.Port == 0 {
		t.Error("Expected default port to be set")
	}
	if cfg.API.Mode == "" {
		t.Error("Expected default mode to be set")
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	cleanup := clearEnvVars(t)
	defer cleanup()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid-config.json")

	// Write invalid JSON
	err := os.WriteFile(configPath, []byte("{ invalid json }"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Try to load
	_, err = LoadConfig(configPath)

	// Assert
	if err == nil {
		t.Error("Expected error when loading invalid JSON")
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	cleanup := clearEnvVars(t)
	defer cleanup()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "minimal-config.json")

	// Minimal config - should use defaults
	configContent := `{
		"api": {
			"mode": "direct"
		}
	}`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	cfg, err := LoadConfig(configPath)

	// Assert defaults are applied
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if cfg.Server.Host == "" {
		t.Error("Expected default host to be set")
	}
	if cfg.Server.Port == 0 {
		t.Error("Expected default port to be set")
	}
	if cfg.API.MaxTokens == 0 {
		t.Error("Expected default max_tokens to be set")
	}
}

func TestLoadConfigEnvVarOverrides(t *testing.T) {
	cleanup := clearEnvVars(t)
	defer cleanup()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "base-config.json")

	configContent := `{
		"server": {
			"port": 8080
		},
		"api": {
			"mode": "direct"
		}
	}`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Set environment variable overrides
	os.Setenv("SERVER_PORT", "9999")
	os.Setenv("SERVER_HOST", "0.0.0.0")

	// Load config
	cfg, err := LoadConfig(configPath)

	// Assert env vars override config file
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if cfg.Server.Port != 9999 {
		t.Errorf("Expected port 9999 from env var, got %d", cfg.Server.Port)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Expected host '0.0.0.0' from env var, got '%s'", cfg.Server.Host)
	}
}

// ─── providers config tests ───────────────────────────────────────────────────

func TestProvidersConfigParsed(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "providers-config.json")

	configContent := `{
		"providers": {
			"qwen": {
				"endpoint": "http://localhost:1234/v1",
				"api_key_env": "MY_QWEN_KEY"
			},
			"anthropic": {
				"api_key_env": "CORP_ANTHROPIC_KEY"
			}
		}
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if cfg.Providers == nil {
		t.Fatal("expected Providers to be non-nil after parsing")
	}
	qwen := cfg.Providers["qwen"]
	if qwen.Endpoint != "http://localhost:1234/v1" {
		t.Errorf("qwen endpoint: got %q, want %q", qwen.Endpoint, "http://localhost:1234/v1")
	}
	if qwen.APIKeyEnv != "MY_QWEN_KEY" {
		t.Errorf("qwen api_key_env: got %q, want %q", qwen.APIKeyEnv, "MY_QWEN_KEY")
	}
	anth := cfg.Providers["anthropic"]
	if anth.APIKeyEnv != "CORP_ANTHROPIC_KEY" {
		t.Errorf("anthropic api_key_env: got %q, want %q", anth.APIKeyEnv, "CORP_ANTHROPIC_KEY")
	}
}

func TestProvidersConfigOmittedIsBackwardCompatible(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "no-providers-config.json")

	// Config without providers section
	configContent := `{
		"api": {
			"mode": "direct"
		}
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	// Helper methods must return defaults when providers section is absent
	if got := cfg.ProviderEndpoint("qwen"); got != "" {
		t.Errorf("ProviderEndpoint qwen: got %q, want empty", got)
	}
	if got := cfg.ProviderAPIKeyEnv("qwen", "LLAMA_API_KEY"); got != "LLAMA_API_KEY" {
		t.Errorf("ProviderAPIKeyEnv qwen default: got %q, want %q", got, "LLAMA_API_KEY")
	}
	if got := cfg.ProviderAPIKeyEnv("anthropic", "ANTHROPIC_API_KEY"); got != "ANTHROPIC_API_KEY" {
		t.Errorf("ProviderAPIKeyEnv anthropic default: got %q, want %q", got, "ANTHROPIC_API_KEY")
	}
}

func TestProviderEndpointOverridesQwenDefault(t *testing.T) {
	cfg := &Config{
		Providers: ProvidersConfig{
			"qwen": {Endpoint: "http://my-proxy:8080/v1"},
		},
	}

	got := cfg.ProviderEndpoint("qwen")
	if got != "http://my-proxy:8080/v1" {
		t.Errorf("ProviderEndpoint: got %q, want %q", got, "http://my-proxy:8080/v1")
	}
}

func TestProviderAPIKeyEnvOverrides(t *testing.T) {
	cfg := &Config{
		Providers: ProvidersConfig{
			"qwen":      {APIKeyEnv: "MY_QWEN_KEY"},
			"anthropic": {APIKeyEnv: "CORP_ANTHROPIC_KEY"},
			"openai":    {APIKeyEnv: "CORP_OPENAI_KEY"},
			"gemini":    {APIKeyEnv: "CORP_GEMINI_KEY"},
		},
	}

	cases := []struct {
		provider   string
		defaultEnv string
		wantEnv    string
	}{
		{"qwen", "LLAMA_API_KEY", "MY_QWEN_KEY"},
		{"anthropic", "ANTHROPIC_API_KEY", "CORP_ANTHROPIC_KEY"},
		{"openai", "OPENAI_API_KEY", "CORP_OPENAI_KEY"},
		{"gemini", "GEMINI_API_KEY", "CORP_GEMINI_KEY"},
	}
	for _, tc := range cases {
		got := cfg.ProviderAPIKeyEnv(tc.provider, tc.defaultEnv)
		if got != tc.wantEnv {
			t.Errorf("ProviderAPIKeyEnv(%q, %q): got %q, want %q", tc.provider, tc.defaultEnv, got, tc.wantEnv)
		}
	}
}

func TestProviderAPIKeyEnvFallsBackToDefault(t *testing.T) {
	// Provider exists but api_key_env is empty — fallback expected
	cfg := &Config{
		Providers: ProvidersConfig{
			"qwen": {Endpoint: "http://localhost:9000"},
		},
	}

	got := cfg.ProviderAPIKeyEnv("qwen", "LLAMA_API_KEY")
	if got != "LLAMA_API_KEY" {
		t.Errorf("expected fallback default LLAMA_API_KEY, got %q", got)
	}
}

func TestNilConfigProviderHelpers(t *testing.T) {
	var cfg *Config

	if got := cfg.ProviderEndpoint("qwen"); got != "" {
		t.Errorf("nil Config ProviderEndpoint: got %q, want empty", got)
	}
	if got := cfg.ProviderAPIKeyEnv("qwen", "LLAMA_API_KEY"); got != "LLAMA_API_KEY" {
		t.Errorf("nil Config ProviderAPIKeyEnv: got %q, want LLAMA_API_KEY", got)
	}
}
