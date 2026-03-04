package server

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/cortexa-llc/ai-pack/internal/monitoring"
)

// Initialize monitoring for all tests
func init() {
	monitoring.InitLogger(slog.LevelError) // Use error level to reduce noise in tests
}

// Helper to create test directory structure
func setupTestDir(t *testing.T) string {
	tmpDir := t.TempDir()

	// Create roles directory
	rolesDir := filepath.Join(tmpDir, "roles")
	if err := os.MkdirAll(rolesDir, 0755); err != nil {
		t.Fatalf("Failed to create roles dir: %v", err)
	}

	// Create .beads/tasks directory
	tasksDir := filepath.Join(tmpDir, ".beads", "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("Failed to create tasks dir: %v", err)
	}

	// Create test agent config
	testAgentConfig := `
name: test-agent
description: Test agent for unit tests

context:
  role_file: roles/test-agent.md
  gates:
    - validation

delegation:
  mode: delegate
  timeout: 5min
  max_context: 8000

tools:
  - read
  - write

success_criteria:
  - Task completed successfully
`
	configPath := filepath.Join(rolesDir, "test-agent.md")
	if err := os.WriteFile(configPath, []byte(testAgentConfig), 0644); err != nil {
		t.Fatalf("Failed to write test agent config: %v", err)
	}

	// Create test role file
	roleContent := "# Test Agent\n\nYou are a test agent for unit testing."
	rolePath := filepath.Join(tmpDir, "roles", "test-agent.md")
	if err := os.WriteFile(rolePath, []byte(roleContent), 0644); err != nil {
		t.Fatalf("Failed to write role file: %v", err)
	}

	return tmpDir
}

// Helper to clear environment variables
func clearAuthEnvVars(t *testing.T) func() {
	saved := map[string]string{
		"ANTHROPIC_API_TOKEN": os.Getenv("ANTHROPIC_API_TOKEN"),
		"ANTHROPIC_API_KEY":   os.Getenv("ANTHROPIC_API_KEY"),
	}

	os.Unsetenv("ANTHROPIC_API_TOKEN")
	os.Unsetenv("ANTHROPIC_API_KEY")

	return func() {
		for key, value := range saved {
			if value != "" {
				os.Setenv(key, value)
			}
		}
	}
}
