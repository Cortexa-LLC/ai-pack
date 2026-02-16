package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// GetModelForRole returns the model to use for a given role
// Loads the agent config and returns the model field, or default if not specified
func (s *AgentServer) GetModelForRole(role string) string {
	if role == "" {
		return s.model // Default model
	}

	// Try to load agent config
	config, err := s.loadAgentConfigForRole(role)
	if err != nil || config.Model == "" {
		// No model specified in config, use default
		return s.model
	}

	return config.Model
}

// loadAgentConfigForRole loads the agent config from the role .md file
func (s *AgentServer) loadAgentConfigForRole(role string) (*AgentConfig, error) {
	// Check multiple locations for role files
	locations := []string{
		filepath.Join(s.rootDir, ".ai-pack", "agents", fmt.Sprintf("%s.md", role)),
		filepath.Join(s.rootDir, "agents", fmt.Sprintf("%s.md", role)),
		filepath.Join(s.rootDir, "roles", fmt.Sprintf("%s.md", role)),
	}

	// Special case for orchestrator in chat mode
	if role == "orchestrator" {
		locations = append([]string{
			filepath.Join(s.rootDir, ".ai-pack", "agents", "orchestrator-chat.md"),
			filepath.Join(s.rootDir, "roles", "orchestrator-chat.md"),
		}, locations...)
	}

	for _, path := range locations {
		if _, err := os.Stat(path); err == nil {
			return s.parseAgentConfig(path)
		}
	}

	return nil, fmt.Errorf("agent config not found for role: %s", role)
}

// parseAgentConfig parses the YAML frontmatter from an agent .md file
func (s *AgentServer) parseAgentConfig(path string) (*AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)

	// Check for YAML frontmatter
	if !strings.HasPrefix(content, "---") {
		return &AgentConfig{}, nil // No frontmatter
	}

	// Extract frontmatter
	parts := strings.SplitN(content[3:], "---", 2)
	if len(parts) < 2 {
		return &AgentConfig{}, nil // Invalid frontmatter
	}

	// Parse YAML
	var config AgentConfig
	if err := yaml.Unmarshal([]byte(parts[0]), &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	return &config, nil
}
