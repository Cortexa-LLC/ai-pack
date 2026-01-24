package auth

import (
	"os"
	"testing"
)

func TestGetAPIKey_FromAPIToken(t *testing.T) {
	// Setup
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token-123")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")
	os.Unsetenv("ANTHROPIC_API_KEY")

	// Test
	apiKey, err := GetAPIKey()

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if apiKey != "test-token-123" {
		t.Errorf("Expected 'test-token-123', got '%s'", apiKey)
	}
}

func TestGetAPIKey_FromAPIKey(t *testing.T) {
	// Setup
	os.Unsetenv("ANTHROPIC_API_TOKEN")
	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-456")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	// Test
	apiKey, err := GetAPIKey()

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if apiKey != "sk-ant-test-456" {
		t.Errorf("Expected 'sk-ant-test-456', got '%s'", apiKey)
	}
}

func TestGetAPIKey_PriorityOrder(t *testing.T) {
	// Setup - both set, token should take priority
	os.Setenv("ANTHROPIC_API_TOKEN", "bearer-token")
	os.Setenv("ANTHROPIC_API_KEY", "api-key")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	// Test
	apiKey, err := GetAPIKey()

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if apiKey != "bearer-token" {
		t.Errorf("Expected bearer token to take priority, got '%s'", apiKey)
	}
}

func TestGetAPIKey_NoEnvVars(t *testing.T) {
	// Setup
	os.Unsetenv("ANTHROPIC_API_TOKEN")
	os.Unsetenv("ANTHROPIC_API_KEY")

	// Test
	apiKey, err := GetAPIKey()

	// Assert - should fail or fallback to Claude Code helper
	// If Claude Code is not configured, should return error
	if err == nil && apiKey == "" {
		t.Error("Expected either an error or a valid API key from Claude Code helper")
	}
}

