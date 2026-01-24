package server

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/config"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/protocol"
)

// Initialize monitoring for all tests
func init() {
	monitoring.InitLogger(slog.LevelError) // Use error level to reduce noise in tests
}

// Helper to create test directory structure
func setupTestDir(t *testing.T) string {
	tmpDir := t.TempDir()

	// Create .ai-pack/agents/lightweight directory
	agentDir := filepath.Join(tmpDir, ".ai-pack", "agents", "lightweight")
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		t.Fatalf("Failed to create agent dir: %v", err)
	}

	// Create .ai-pack/roles directory
	rolesDir := filepath.Join(tmpDir, ".ai-pack", "roles")
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
tier: lightweight

context:
  role_file: .ai-pack/roles/test-agent.md
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
	configPath := filepath.Join(agentDir, "test-agent.yml")
	if err := os.WriteFile(configPath, []byte(testAgentConfig), 0644); err != nil {
		t.Fatalf("Failed to write test agent config: %v", err)
	}

	// Create test role file
	roleContent := "# Test Agent\n\nYou are a test agent for unit testing."
	rolePath := filepath.Join(tmpDir, ".ai-pack", "roles", "test-agent.md")
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

func TestNewAgentServer_Success(t *testing.T) {
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)

	// Set required API key
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	apiCfg := &config.APIConfig{
		Mode: "direct",
	}

	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", apiCfg)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if server == nil {
		t.Fatal("Expected server to be created")
	}
	if server.rootDir != tmpDir {
		t.Errorf("Expected rootDir %s, got %s", tmpDir, server.rootDir)
	}
	if server.maxConcurrent != 3 {
		t.Errorf("Expected maxConcurrent 3, got %d", server.maxConcurrent)
	}
	if server.maxTokens != 4000 {
		t.Errorf("Expected maxTokens 4000, got %d", server.maxTokens)
	}
	if server.model != "claude-3-5-sonnet-20241022" {
		t.Errorf("Expected model claude-3-5-sonnet-20241022, got %s", server.model)
	}
	if server.client == nil {
		t.Error("Expected Anthropic client to be initialized")
	}
	if server.activeTasks == nil {
		t.Error("Expected activeTasks map to be initialized")
	}
	if server.taskQueue == nil {
		t.Error("Expected taskQueue to be initialized")
	}
	if server.workerPool == nil {
		t.Error("Expected workerPool to be initialized")
	}
}

func TestNewAgentServer_FallbackToClaudeCode(t *testing.T) {
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)

	apiCfg := &config.APIConfig{
		Mode: "direct",
	}

	// Should succeed via Claude Code helper fallback
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", apiCfg)

	// If Claude Code is configured, should succeed
	// If not configured, will fail - both are valid
	if err != nil {
		// No Claude Code helper available - this is expected in some environments
		t.Skip("Claude Code helper not available - skipping fallback test")
	}
	if server == nil {
		t.Fatal("Expected server to be created via Claude Code helper")
	}
}

func TestLoadAgentConfig_Success(t *testing.T) {
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	apiCfg := &config.APIConfig{Mode: "direct"}
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", apiCfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	cfg, err := server.loadAgentConfig("test-agent")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if cfg.Name != "test-agent" {
		t.Errorf("Expected name 'test-agent', got '%s'", cfg.Name)
	}
	if cfg.Tier != "lightweight" {
		t.Errorf("Expected tier 'lightweight', got '%s'", cfg.Tier)
	}
	if cfg.Context.RoleFile != ".ai-pack/roles/test-agent.md" {
		t.Errorf("Expected role_file '.ai-pack/roles/test-agent.md', got '%s'", cfg.Context.RoleFile)
	}
}

func TestLoadAgentConfig_NotFound(t *testing.T) {
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	apiCfg := &config.APIConfig{Mode: "direct"}
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", apiCfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	_, err = server.loadAgentConfig("nonexistent-agent")

	if err == nil {
		t.Error("Expected error when loading nonexistent agent config")
	}
}

func TestLoadRoleContext_Success(t *testing.T) {
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	apiCfg := &config.APIConfig{Mode: "direct"}
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", apiCfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	roleContext, err := server.loadRoleContext(".ai-pack/roles/test-agent.md")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if roleContext == "" {
		t.Error("Expected role context to be loaded")
	}
	if roleContext != "# Test Agent\n\nYou are a test agent for unit testing." {
		t.Errorf("Unexpected role context: %s", roleContext)
	}
}

func TestCreateTaskPacket_Success(t *testing.T) {
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	apiCfg := &config.APIConfig{Mode: "direct"}
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", apiCfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	config := &AgentConfig{
		Name: "test-agent",
		Tier: "lightweight",
	}

	taskID := "task-test-20260124-120000-000000"
	err = server.createTaskPacket(taskID, "test-agent", "Test task", config)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify task directory created
	taskDir := filepath.Join(tmpDir, ".beads", "tasks", taskID)
	if _, err := os.Stat(taskDir); os.IsNotExist(err) {
		t.Error("Expected task directory to be created")
	}

	// Verify metadata file created
	metadataPath := filepath.Join(taskDir, "00-metadata.json")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		t.Error("Expected metadata file to be created")
	}

	// Verify plan file created
	planPath := filepath.Join(taskDir, "10-plan.md")
	if _, err := os.Stat(planPath); os.IsNotExist(err) {
		t.Error("Expected plan file to be created")
	}
}

func TestUpdateTaskStatus(t *testing.T) {
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	apiCfg := &config.APIConfig{Mode: "direct"}
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", apiCfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	config := &AgentConfig{
		Name: "test-agent",
		Tier: "lightweight",
	}

	taskID := "task-test-status-update"
	err = server.createTaskPacket(taskID, "test-agent", "Test task", config)
	if err != nil {
		t.Fatalf("Failed to create task packet: %v", err)
	}

	// Update status
	server.updateTaskStatus(taskID, "in_progress", "")

	// Read metadata and verify
	metadataPath := filepath.Join(tmpDir, ".beads", "tasks", taskID, "00-metadata.json")
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatalf("Failed to read metadata: %v", err)
	}

	if !contains(string(data), "in_progress") {
		t.Error("Expected status to be 'in_progress'")
	}
}

func TestBuildPrompt(t *testing.T) {
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	apiCfg := &config.APIConfig{Mode: "direct"}
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", apiCfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	config := &AgentConfig{
		Delegation: struct {
			Mode       string `yaml:"mode"`
			Timeout    string `yaml:"timeout"`
			MaxContext int    `yaml:"max_context"`
		}{
			Timeout: "5min",
		},
		Tools:           []string{"read", "write"},
		SuccessCriteria: []string{"Task complete"},
	}

	prompt := server.buildPrompt("engineer", "Create a function", "You are an engineer", config)

	if prompt == "" {
		t.Error("Expected prompt to be generated")
	}
	if !contains(prompt, "engineer") {
		t.Error("Expected prompt to contain role")
	}
	if !contains(prompt, "Create a function") {
		t.Error("Expected prompt to contain task")
	}
	if !contains(prompt, "You are an engineer") {
		t.Error("Expected prompt to contain role context")
	}
}

func TestTaskExecution_StreamChannel(t *testing.T) {
	execution := &TaskExecution{
		TaskID:     "test-task",
		Role:       "test",
		Task:       "test task",
		StartTime:  time.Now(),
		Status:     "queued",
		streamChan: make(chan *protocol.StreamEvent, 10),
		streamOpen: true,
	}

	// Test sending events
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	apiCfg := &config.APIConfig{Mode: "direct"}
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", apiCfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Send event
	server.sendStreamEvent(execution, "test_event", map[string]interface{}{
		"data": "test",
	})

	// Verify event received
	select {
	case event := <-execution.streamChan:
		if event.Type != "test_event" {
			t.Errorf("Expected event type 'test_event', got '%s'", event.Type)
		}
		if event.TaskID != "test-task" {
			t.Errorf("Expected task_id 'test-task', got '%s'", event.TaskID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive stream event")
	}
}

func TestCloseStream(t *testing.T) {
	cleanup := clearAuthEnvVars(t)
	defer cleanup()

	tmpDir := setupTestDir(t)
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	apiCfg := &config.APIConfig{Mode: "direct"}
	server, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-5-sonnet-20241022", apiCfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	execution := &TaskExecution{
		TaskID:     "test-task",
		streamChan: make(chan *protocol.StreamEvent, 10),
		streamOpen: true,
	}

	server.closeStream(execution)

	if execution.streamOpen {
		t.Error("Expected stream to be closed")
	}

	// Verify channel is closed
	_, ok := <-execution.streamChan
	if ok {
		t.Error("Expected channel to be closed")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr))))
}
