package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
	"gopkg.in/yaml.v3"
)

// ConfigValidator validates agent configurations and API key availability
type ConfigValidator struct {
	server *AgentServer
}

// NewConfigValidator creates a new config validator
func NewConfigValidator(server *AgentServer) *ConfigValidator {
	return &ConfigValidator{server: server}
}

// ValidateAllConfigs validates all agent configs and reports issues
func (cv *ConfigValidator) ValidateAllConfigs() []ValidationWarning {
	var warnings []ValidationWarning

	// Check for agent config directories
	agentDirs := []string{
		filepath.Join(cv.server.rootDir, ".ai-pack", "agents"),
		filepath.Join(cv.server.rootDir, "agents"),
	}

	for _, dir := range agentDirs {
		if _, err := os.Stat(dir); err == nil {
			dirWarnings := cv.validateAgentDirectory(dir)
			warnings = append(warnings, dirWarnings...)
		}
	}

	// Log validation summary
	if len(warnings) > 0 {
		monitoring.Logger.Warn("config_validation_warnings", "count", len(warnings))
		for _, w := range warnings {
			monitoring.Logger.Warn("config_warning",
				"type", w.Type,
				"file", w.ConfigFile,
				"message", w.Message)
		}
	} else {
		monitoring.Logger.Info("config_validation_passed", "message", "All agent configs valid")
	}

	return warnings
}

// validateAgentDirectory scans and validates all configs in a directory
func (cv *ConfigValidator) validateAgentDirectory(dir string) []ValidationWarning {
	var warnings []ValidationWarning

	entries, err := os.ReadDir(dir)
	if err != nil {
		return warnings
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		configPath := filepath.Join(dir, entry.Name())
		configWarnings := cv.validateConfig(configPath)
		warnings = append(warnings, configWarnings...)
	}

	return warnings
}

// validateConfig validates a single agent config file
func (cv *ConfigValidator) validateConfig(configPath string) []ValidationWarning {
	var warnings []ValidationWarning

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		warnings = append(warnings, ValidationWarning{
			Type:       "read_error",
			ConfigFile: configPath,
			Message:    fmt.Sprintf("Failed to read config: %v", err),
			Severity:   "error",
		})
		return warnings
	}

	// Extract YAML frontmatter
	content := string(data)
	if !strings.HasPrefix(content, "---") {
		// No frontmatter, skip validation
		return warnings
	}

	parts := strings.SplitN(content[3:], "---", 2)
	if len(parts) < 2 {
		return warnings
	}

	// Parse YAML
	var config AgentConfig
	if err := yaml.Unmarshal([]byte(parts[0]), &config); err != nil {
		warnings = append(warnings, ValidationWarning{
			Type:       "parse_error",
			ConfigFile: configPath,
			Message:    fmt.Sprintf("Failed to parse YAML: %v", err),
			Severity:   "error",
		})
		return warnings
	}

	// Validate model availability
	if config.Model != "" {
		modelWarnings := cv.validateModelAvailability(config.Model, configPath)
		warnings = append(warnings, modelWarnings...)
	}

	return warnings
}

// validateModelAvailability checks if the specified model is available
func (cv *ConfigValidator) validateModelAvailability(model, configPath string) []ValidationWarning {
	var warnings []ValidationWarning

	// Check OpenAI models
	if strings.HasPrefix(model, "gpt-") {
		if cv.server.openaiProvider == nil {
			warnings = append(warnings, ValidationWarning{
				Type:       "missing_api_key",
				ConfigFile: configPath,
				Model:      model,
				Message: fmt.Sprintf(
					"Config uses '%s' but OPENAI_API_KEY is not set. "+
						"Agent will fall back to Claude Sonnet (higher cost). "+
						"Run './scripts/setup-api-keys.sh' to configure.",
					model),
				Severity:   "warning",
				Suggestion: "Set OPENAI_API_KEY environment variable or run setup script",
			})
		}
	}

	// Check Claude models
	if strings.Contains(model, "claude") {
		if cv.server.anthropicKey == "" {
			warnings = append(warnings, ValidationWarning{
				Type:       "missing_api_key",
				ConfigFile: configPath,
				Model:      model,
				Message: fmt.Sprintf(
					"Config uses '%s' but ANTHROPIC_API_KEY is not set. Agent will fail.",
					model),
				Severity:   "error",
				Suggestion: "Set ANTHROPIC_API_KEY environment variable",
			})
		}
	}

	return warnings
}

// ValidationWarning represents a config validation issue
type ValidationWarning struct {
	Type       string // "missing_api_key", "parse_error", "invalid_model"
	ConfigFile string
	Model      string
	Message    string
	Severity   string // "error", "warning", "info"
	Suggestion string
}

// PrintValidationReport prints a user-friendly validation report
func PrintValidationReport(warnings []ValidationWarning) {
	if len(warnings) == 0 {
		fmt.Println("✅ All agent configurations validated successfully")
		return
	}

	fmt.Printf("\n⚠️  Found %d configuration issue(s):\n\n", len(warnings))

	for i, w := range warnings {
		fmt.Printf("%d. ", i+1)

		switch w.Severity {
		case "error":
			fmt.Printf("❌ ERROR: ")
		case "warning":
			fmt.Printf("⚠️  WARNING: ")
		default:
			fmt.Printf("ℹ️  INFO: ")
		}

		fmt.Printf("%s\n", w.Message)
		fmt.Printf("   File: %s\n", w.ConfigFile)

		if w.Model != "" {
			fmt.Printf("   Model: %s\n", w.Model)
		}

		if w.Suggestion != "" {
			fmt.Printf("   💡 Suggestion: %s\n", w.Suggestion)
		}

		fmt.Println()
	}

	// Summary and action items
	errorCount := 0
	warningCount := 0
	for _, w := range warnings {
		if w.Severity == "error" {
			errorCount++
		} else if w.Severity == "warning" {
			warningCount++
		}
	}

	if errorCount > 0 {
		fmt.Printf("❌ %d error(s) must be fixed before agents can run\n", errorCount)
	}
	if warningCount > 0 {
		fmt.Printf("⚠️  %d warning(s) - agents will work with fallbacks (higher costs)\n", warningCount)
	}

	fmt.Println("\n💡 Quick fix: Run './scripts/setup-api-keys.sh' to configure API keys")
}
