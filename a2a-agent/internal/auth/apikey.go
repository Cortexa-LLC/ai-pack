package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ClaudeSettings represents the Claude Code settings.json structure
type ClaudeSettings struct {
	APIKeyHelper string            `json:"apiKeyHelper"`
	Env          map[string]string `json:"env"`
}

// GetAPIKey attempts to get an API key from multiple sources in priority order:
// 1. ANTHROPIC_API_TOKEN environment variable (bearer token for corporate proxies)
// 2. ANTHROPIC_API_KEY environment variable (standard API key)
// 3. Claude Code API key helper (from ~/.claude/settings.json)
// Returns: (key/token, isBearerToken, error)
func GetAPIKey() (string, bool, error) {
	// Try bearer token first (for corporate proxies)
	if token := os.Getenv("ANTHROPIC_API_TOKEN"); token != "" {
		return token, true, nil
	}

	// Try standard API key
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		return apiKey, false, nil
	}

	// Try Claude Code API key helper
	apiKey, err := getAPIKeyFromClaudeCode()
	if err == nil && apiKey != "" {
		return apiKey, false, nil
	}

	return "", false, fmt.Errorf("ANTHROPIC_API_TOKEN or ANTHROPIC_API_KEY not set and Claude Code API key helper not available")
}

// getAPIKeyFromClaudeCode retrieves API key using Claude Code's API key helper
func getAPIKeyFromClaudeCode() (string, error) {
	// Read Claude Code settings
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return "", fmt.Errorf("failed to read Claude Code settings: %w", err)
	}

	var settings ClaudeSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return "", fmt.Errorf("failed to parse Claude Code settings: %w", err)
	}

	if settings.APIKeyHelper == "" {
		return "", fmt.Errorf("no API key helper configured in Claude Code settings")
	}

	// Execute the API key helper command
	apiKey, err := executeAPIKeyHelper(settings.APIKeyHelper)
	if err != nil {
		return "", fmt.Errorf("failed to execute API key helper: %w", err)
	}

	return apiKey, nil
}

// executeAPIKeyHelper runs the API key helper command and returns the token
func executeAPIKeyHelper(helperCmd string) (string, error) {
	// Parse the command (e.g., "npx @company/api-token get_token")
	parts := strings.Fields(helperCmd)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty API key helper command")
	}

	// Create command
	ctx := exec.Command(parts[0], parts[1:]...)
	ctx.Env = os.Environ()

	// Set timeout
	done := make(chan error, 1)
	var output []byte
	var cmdErr error

	go func() {
		output, cmdErr = ctx.CombinedOutput()
		done <- cmdErr
	}()

	// Wait with timeout
	select {
	case err := <-done:
		if err != nil {
			return "", fmt.Errorf("API key helper command failed: %w (output: %s)", err, string(output))
		}
	case <-time.After(30 * time.Second):
		ctx.Process.Kill()
		return "", fmt.Errorf("API key helper command timed out")
	}

	// The output should be the API key
	apiKey := strings.TrimSpace(string(output))
	if apiKey == "" {
		return "", fmt.Errorf("API key helper returned empty result")
	}

	return apiKey, nil
}

// GetBaseURL returns the Anthropic base URL from Claude Code settings or default
// Returns the URL as-is from settings - this is the FULL base URL that the SDK will use
func GetBaseURL() string {
	// Try environment variable first
	baseURL := os.Getenv("ANTHROPIC_BASE_URL")

	// If not set, try Claude Code settings
	if baseURL == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return ""
		}

		settingsPath := filepath.Join(homeDir, ".claude", "settings.json")
		data, err := os.ReadFile(settingsPath)
		if err != nil {
			return ""
		}

		var settings ClaudeSettings
		if err := json.Unmarshal(data, &settings); err != nil {
			return ""
		}

		if url, ok := settings.Env["ANTHROPIC_BASE_URL"]; ok {
			baseURL = url
		}
	}

	// Return as-is - the URL in Claude Code settings is the complete base URL
	return baseURL
}
