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
		"ANTHROPIC_BASE_URL":  os.Getenv("ANTHROPIC_BASE_URL"),
		"ANTHROPIC_MODEL":     os.Getenv("ANTHROPIC_MODEL"),
		"SERVER_HOST":         os.Getenv("SERVER_HOST"),
		"SERVER_PORT":         os.Getenv("SERVER_PORT"),
		"API_MODE":            os.Getenv("API_MODE"),
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

func TestLoadConfig_ValidFile(t *testing.T) {
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

func TestLoadConfig_ProxyMode(t *testing.T) {
	cleanup := clearEnvVars(t)
	defer cleanup()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "proxy-config.json")

	configContent := `{
		"api": {
			"mode": "proxy",
			"proxy": {
				"type": "acme",
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
	if cfg.API.Proxy.Type != "acme" {
		t.Errorf("Expected proxy type '"acme"', got '%s'", cfg.API.Proxy.Type)
	}
	if cfg.API.Proxy.BaseURL != "https://proxy.example.com/api" {
		t.Errorf("Expected proxy URL 'https://proxy.example.com/api', got '%s'", cfg.API.Proxy.BaseURL)
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
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

func TestLoadConfig_InvalidJSON(t *testing.T) {
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

func TestLoadConfig_Defaults(t *testing.T) {
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

func TestLoadConfig_EnvVarOverrides(t *testing.T) {
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
