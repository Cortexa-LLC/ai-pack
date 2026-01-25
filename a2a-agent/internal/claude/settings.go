package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Settings represents Claude Code settings
type Settings struct {
	Permissions struct {
		Allow       []string `json:"allow"`
		Deny        []string `json:"deny"`
		Ask         []string `json:"ask"`
		DefaultMode string   `json:"defaultMode"`
	} `json:"permissions"`
}

// LoadSettings loads Claude Code settings from tiered locations and merges them
// Precedence (highest to lowest):
// 1. Project: <project-root>/.claude/settings.json
// 2. Global: ~/.claude/settings.json
func LoadSettings(projectRoot string) (*Settings, error) {
	merged := &Settings{}

	// Load global settings first
	homeDir, err := os.UserHomeDir()
	if err == nil {
		globalPath := filepath.Join(homeDir, ".claude", "settings.json")
		if globalSettings, err := loadSettingsFile(globalPath); err == nil {
			merged = mergeSettings(merged, globalSettings)
		}
	}

	// Load project settings (override global)
	if projectRoot != "" {
		projectPath := filepath.Join(projectRoot, ".claude", "settings.json")
		if projectSettings, err := loadSettingsFile(projectPath); err == nil {
			merged = mergeSettings(merged, projectSettings)
		}
	}

	return merged, nil
}

// loadSettingsFile loads a single settings file
func loadSettingsFile(path string) (*Settings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// mergeSettings merges two settings, with override taking precedence
func mergeSettings(base, override *Settings) *Settings {
	result := &Settings{}

	// Merge permissions
	result.Permissions.Allow = append(base.Permissions.Allow, override.Permissions.Allow...)
	result.Permissions.Deny = append(base.Permissions.Deny, override.Permissions.Deny...)
	result.Permissions.Ask = append(base.Permissions.Ask, override.Permissions.Ask...)

	// Use override's default mode if set, otherwise use base's
	if override.Permissions.DefaultMode != "" {
		result.Permissions.DefaultMode = override.Permissions.DefaultMode
	} else {
		result.Permissions.DefaultMode = base.Permissions.DefaultMode
	}

	return result
}

// IsOperationAllowed checks if a tool operation on a path is allowed
func (s *Settings) IsOperationAllowed(tool, path string) (allowed bool, reason string) {
	// Normalize path - expand home directory
	if strings.HasPrefix(path, "~/") {
		homeDir, _ := os.UserHomeDir()
		path = filepath.Join(homeDir, path[2:])
	}
	path = filepath.Clean(path)

	// Check deny patterns first
	for _, pattern := range s.Permissions.Deny {
		if matchesPermissionPattern(pattern, tool, path) {
			return false, pattern
		}
	}

	// If not denied, it's allowed
	return true, ""
}

// matchesPermissionPattern checks if a tool operation matches a permission pattern
// Pattern format: "Tool(file_pattern)" e.g. "Read(.env)", "Edit(**/*.pem)", "Bash(sudo:*)"
func matchesPermissionPattern(pattern, tool, path string) bool {
	// Parse pattern: Tool(file_pattern)
	openParen := strings.Index(pattern, "(")
	closeParen := strings.LastIndex(pattern, ")")

	if openParen == -1 || closeParen == -1 {
		return false
	}

	patternTool := pattern[:openParen]
	filePattern := pattern[openParen+1 : closeParen]

	// Check if tool matches
	if !strings.EqualFold(patternTool, tool) {
		return false
	}

	// For Bash commands, match command prefix
	if tool == "Bash" || tool == "bash" {
		// Pattern like "Bash(sudo:*)" means any command starting with "sudo"
		if strings.HasSuffix(filePattern, ":*") {
			cmdPrefix := strings.TrimSuffix(filePattern, ":*")
			return strings.HasPrefix(path, cmdPrefix)
		}
		// Exact match
		return path == filePattern
	}

	// For file operations (Read, Write, Edit), match file path
	return matchesGlobPattern(filePattern, path)
}

// matchesGlobPattern matches a file path against a glob pattern
// Supports: *, **, .env, ./.env, **/.env, etc.
func matchesGlobPattern(pattern, path string) bool {
	// Expand home directory in pattern
	if strings.HasPrefix(pattern, "~/") {
		homeDir, _ := os.UserHomeDir()
		pattern = filepath.Join(homeDir, pattern[2:])
	}

	// Normalize both
	pattern = filepath.Clean(pattern)
	path = filepath.Clean(path)

	// Exact match
	if pattern == path {
		return true
	}

	// Pattern with **/ means match anywhere in path
	if strings.HasPrefix(pattern, "**/") {
		suffix := strings.TrimPrefix(pattern, "**/")
		// Check if path ends with suffix
		if strings.HasSuffix(path, suffix) {
			return true
		}
		// Check if any parent directory + suffix matches
		for i := 0; i < len(path); i++ {
			if path[i] == '/' {
				if matchesGlobPattern(suffix, path[i+1:]) {
					return true
				}
			}
		}
	}

	// Pattern with * wildcard
	if strings.Contains(pattern, "*") {
		return wildcardMatch(pattern, path)
	}

	// Basename match (e.g., ".env" matches "./.env" or "config/.env")
	if filepath.Base(path) == pattern {
		return true
	}

	return false
}

// wildcardMatch does simple wildcard matching
func wildcardMatch(pattern, str string) bool {
	// Simple implementation - split on *
	parts := strings.Split(pattern, "*")

	if len(parts) == 1 {
		return pattern == str
	}

	// Check prefix
	if !strings.HasPrefix(str, parts[0]) {
		return false
	}
	str = strings.TrimPrefix(str, parts[0])

	// Check suffix
	if !strings.HasSuffix(str, parts[len(parts)-1]) {
		return false
	}
	str = strings.TrimSuffix(str, parts[len(parts)-1])

	// Check middle parts
	for i := 1; i < len(parts)-1; i++ {
		idx := strings.Index(str, parts[i])
		if idx == -1 {
			return false
		}
		str = str[idx+len(parts[i]):]
	}

	return true
}
