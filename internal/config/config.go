package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

const (
	defaultConfigFilename = "agent-server.json"
)

// ComplexityWeights controls how much each sub-score contributes to the
// composite risk score before the role multiplier is applied.
// All weights are normalised so they do not need to sum to 1.0;
// they are applied proportionally during scoring.
type ComplexityWeights struct {
	Scope       float64 `json:"scope"`       // default: 0.30
	MultiStep   float64 `json:"multi_step"`  // default: 0.25
	Uncertainty float64 `json:"uncertainty"` // default: 0.25
	Structural  float64 `json:"structural"`  // default: 0.20
}

// ComplexityGateConfig configures the v2 composite risk scorer.
type ComplexityGateConfig struct {
	Enabled           bool               `json:"enabled"`            // default: true
	WarnThreshold     float64            `json:"warn_threshold"`     // default: 0.50
	CriticalThreshold float64            `json:"critical_threshold"` // default: 0.75
	Weights           ComplexityWeights  `json:"weights"`
	RoleMultipliers   map[string]float64 `json:"role_multipliers"` // per-role weight
}

// Config holds all server configuration
type Config struct {
	Server          ServerConfig          `json:"server"`
	API             APIConfig             `json:"api"`
	Agent           AgentConfig           `json:"agent"`
	Logging         LoggingConfig         `json:"logging"`
	Metrics         MetricsConfig         `json:"metrics"`
	TaskCleanup     TaskCleanupConfig     `json:"task_cleanup"`
	ProviderCosts   ProviderCostsConfig   `json:"provider_costs"`
	GradingCriteria GradingCriteriaConfig `json:"grading_criteria"`
	MCP             MCPConfig             `json:"mcp"`
	ComplexityGate  ComplexityGateConfig  `json:"complexity_gate"`
	Projects           map[string]string            `json:"projects,omitempty"`            // map[projectPath]lastAccessed
	Providers          ProvidersConfig               `json:"providers,omitempty"`           // per-provider endpoint / key-env overrides
	ModelTranslations  map[string]map[string]string  `json:"model_translations,omitempty"`  // "from->to": {"model-id": "target-id", "*pattern*": "target", "*": "default"}
}

// ServerConfig holds server-specific settings
type ServerConfig struct {
	Host                string `json:"host"`
	Port                int    `json:"port"`
	MaxConcurrentAgents int    `json:"max_concurrent_agents"`
	WorkerPoolSize      int    `json:"worker_pool_size"`
	APIKey              string `json:"api_key"`
}

// APIConfig holds API-related settings
type APIConfig struct {
	AnthropicModel         string       `json:"anthropic_model"`
	MaxTokens              int          `json:"max_tokens"`
	TimeoutSeconds         int          `json:"timeout_seconds"`           // Overall task timeout (deprecated, use role Timeout field)
	RequestTimeoutSeconds  int          `json:"request_timeout_seconds"`   // Per-API-call timeout (default: 120s)
	Mode                   string       `json:"mode"`                      // "direct" or "proxy"
	AdaptiveModelSelection bool         `json:"adaptive_model_selection"`  // default: true; set false to always use anthropic_model
	Proxy                  *ProxyConfig `json:"proxy,omitempty"`           // Only used when mode = "proxy"
}

// AgentConfig holds agent behavior settings
type AgentConfig struct {
	MaxInactiveTurns         int `json:"max_inactive_turns"`           // Stop agent after N turns without progress
	MaxConsecutiveErrorTurns int `json:"max_consecutive_error_turns"`  // Stop agent after N turns where every tool call returned an error (0 = use MaxInactiveTurns)
}

// ProxyConfig holds proxy-specific settings
type ProxyConfig struct {
	BaseURL string `json:"base_url"` // Full proxy base URL (e.g., "https://proxy.example.com/api")
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

// MetricsConfig holds metrics settings
type MetricsConfig struct {
	Enabled          bool `json:"enabled"`
	MaxTaskDurations int  `json:"max_task_durations"`
}

// TaskCleanupConfig holds task cleanup/archival settings
type TaskCleanupConfig struct {
	Enabled          bool `json:"enabled"`            // Enable automatic cleanup on startup
	ArchiveAfterDays int  `json:"archive_after_days"` // Archive tasks older than N days
}

// ProviderCostsConfig holds pricing for LLM providers
type ProviderCostsConfig struct {
	Models         map[string]ModelCost `json:"models"`
	LastUpdated    string               `json:"last_updated,omitempty"`    // ISO 8601 timestamp
	UpdateInterval int                  `json:"update_interval,omitempty"` // Days between updates (default: 30)
	FreeProviders  []string             `json:"free_providers,omitempty"`  // Providers with no per-token cost (e.g. local)
	ReferenceModel string               `json:"reference_model,omitempty"` // Baseline model for cost savings calculation
}

// ModelCost holds input/output token costs per 1M tokens in USD
type ModelCost struct {
	Provider    string  `json:"provider"`              // Provider name (e.g., "anthropic", "openai")
	InputCost   float64 `json:"input_cost_per_1m"`     // Cost per 1M input tokens in USD
	OutputCost  float64 `json:"output_cost_per_1m"`    // Cost per 1M output tokens in USD
	Description string  `json:"description,omitempty"` // Optional description
}

// GradingCriteriaConfig holds performance grading thresholds
type GradingCriteriaConfig struct {
	GradeA GradeThreshold `json:"grade_a"`
	GradeB GradeThreshold `json:"grade_b"`
	GradeC GradeThreshold `json:"grade_c"`
	GradeD GradeThreshold `json:"grade_d"`
}

// GradeThreshold defines success and retry rate thresholds for a grade
type GradeThreshold struct {
	MinSuccessRate float64 `json:"min_success_rate"` // Minimum success rate (0.0-1.0)
	MaxRetryRate   float64 `json:"max_retry_rate"`   // Maximum retry rate (0.0-1.0)
}

// MCPConfig holds MCP server configuration
type MCPConfig struct {
	Enabled        bool                       `json:"enabled"`         // Enable MCP integration
	Servers        map[string]MCPServerConfig `json:"servers"`         // Server configurations (can override user/project configs)
	EnabledServers []string                   `json:"enabled_servers"` // List of enabled server names (filters available servers)
}

// MCPServerConfig defines configuration for an MCP server
type MCPServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// ProviderConfig holds per-provider settings that can be set in
// agent-server.json under the "providers" key.
type ProviderConfig struct {
	Endpoint  string `json:"endpoint,omitempty"`    // Base URL override (local/proxy)
	APIKeyEnv string `json:"api_key_env,omitempty"` // Env var name for API key
}

// ProvidersConfig is a map of provider name -> per-provider settings.
// Supported provider keys: "anthropic", "openai", "gemini", "qwen".
type ProvidersConfig map[string]ProviderConfig

// DefaultConfig returns default configuration.
// NOTE: These defaults are only used as a baseline for merging config files.
// Users should have ~/.ai-pack/config.json created by 'make install'.
// The port default here should match scripts/init-config.py.
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:                "localhost",
			Port:                8080, // Match ~/.ai-pack/config.json default
			MaxConcurrentAgents: 10,
			WorkerPoolSize:      10,
		},
		API: APIConfig{
			AnthropicModel:         "claude-sonnet-4-6",
			MaxTokens:              24000,
			TimeoutSeconds:         600,
			Mode:                   "direct", // "direct" or "proxy"
			AdaptiveModelSelection: true,     // grade-based selection on by default
			Proxy:                  nil,
		},
		Agent: AgentConfig{
			MaxInactiveTurns: 10, // Stop after 10 turns without progress
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		Metrics: MetricsConfig{
			Enabled:          true,
			MaxTaskDurations: 1000,
		},
		TaskCleanup: TaskCleanupConfig{
			Enabled:          true,
			ArchiveAfterDays: 15,
		},
		ProviderCosts: ProviderCostsConfig{
			UpdateInterval: 30,                         // Update costs every 30 days
			Models:         make(map[string]ModelCost), // Empty by default - load from config file
		},
		GradingCriteria: GradingCriteriaConfig{
			GradeA: GradeThreshold{MinSuccessRate: 0.90, MaxRetryRate: 0.05},
			GradeB: GradeThreshold{MinSuccessRate: 0.80, MaxRetryRate: 0.10},
			GradeC: GradeThreshold{MinSuccessRate: 0.70, MaxRetryRate: 0.20},
			GradeD: GradeThreshold{MinSuccessRate: 0.60, MaxRetryRate: 0.30},
		},
		MCP: MCPConfig{
			Enabled:        false,                            // Disabled by default
			Servers:        make(map[string]MCPServerConfig), // Empty by default
			EnabledServers: []string{},                       // Empty means all available servers enabled
		},
		ComplexityGate: ComplexityGateConfig{
			Enabled:           true,
			WarnThreshold:     0.50,
			CriticalThreshold: 0.75,
			Weights: ComplexityWeights{
				Scope:       0.30,
				MultiStep:   0.25,
				Uncertainty: 0.25,
				Structural:  0.20,
			},
			RoleMultipliers: map[string]float64{},
		},
		Projects: make(map[string]string),
		ModelTranslations: map[string]map[string]string{
			"openai->anthropic": {
				"gpt-4o":            "claude-sonnet-4-6",
				"gpt-4o-2024-08-06": "claude-sonnet-4-6",
				"gpt-4.1-mini":      "claude-sonnet-4-6",
				"gpt-4o-mini":       "claude-haiku-4-5-20251022",
				"gpt-4.1-nano":      "claude-haiku-4-5-20251022",
				"*":                 "claude-sonnet-4-6",
			},
			"anthropic->openai": {
				"*opus*":   "gpt-4o",
				"*sonnet*": "gpt-4o",
				"*haiku*":  "gpt-4o-mini",
				"*":        "gpt-4o",
			},
			"qwen->anthropic": {
				"*": "claude-sonnet-4-6",
			},
			"gemini->anthropic": {
				"*pro*": "claude-sonnet-4-6",
				"*":     "claude-haiku-4-5",
			},
			"gemini->openai": {
				"*pro*": "gpt-4.1",
				"*":     "gpt-4.1-mini",
			},
		},
	}
}

// LoadConfig loads configuration from file with environment variable overrides
// Config search order:
// Config is loaded by merging sources in priority order (later wins):
// 1. Built-in defaults
// 2. ~/.claude/agent-server.json (user baseline)
// 3. ./agent-server.json (project config — overrides user baseline)
// 4. AIPACK_AGENT_SERVER_CONFIG or explicit --config path (explicit override)
// 5. Environment variables (highest priority)
func LoadConfig(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	// Helper to merge a file into cfg (ignores missing files)
	mergeFile := func(path string) error {
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return fmt.Errorf("failed to read config file %s: %w", path, err)
		}
		if err := json.Unmarshal(data, cfg); err != nil {
			return fmt.Errorf("failed to parse config file %s: %w", path, err)
		}
		return nil
	}

	// 2. User baseline
	if homeDir, err := os.UserHomeDir(); err == nil {
		_ = mergeFile(homeDir + "/.claude/" + defaultConfigFilename)
	}

	// 3. Project config (overrides user baseline)
	if err := mergeFile(defaultConfigFilename); err != nil {
		return nil, err
	}

	// 4. Explicit path (--config flag or AIPACK_AGENT_SERVER_CONFIG env)
	if explicitPath := resolveExplicitConfigPath(configPath); explicitPath != "" {
		if err := mergeFile(explicitPath); err != nil {
			return nil, err
		}
	}

	// 5. Environment variables
	applyEnvOverrides(cfg)

	return cfg, nil
}

// resolveExplicitConfigPath returns an explicit override path from --config flag or env var.
// Returns "" if no explicit override is set.
func resolveExplicitConfigPath(explicitPath string) string {
	// 1. Explicit path provided (--config flag), but not the default filename
	if explicitPath != "" && explicitPath != defaultConfigFilename {
		return explicitPath
	}

	// 2. AIPACK_AGENT_SERVER_CONFIG environment variable
	if envPath := os.Getenv("AIPACK_AGENT_SERVER_CONFIG"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}
	}

	// 5. Return empty to use built-in defaults
	return ""
}

func applyServerOverrides(cfg *Config) {
	if val := os.Getenv("SERVER_HOST"); val != "" {
		cfg.Server.Host = val
	}
	if val := os.Getenv("AIPACK_AGENT_SERVER_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			cfg.Server.Port = port
		}
	}
	if val := os.Getenv("MAX_CONCURRENT_AGENTS"); val != "" {
		if max, err := strconv.Atoi(val); err == nil {
			cfg.Server.MaxConcurrentAgents = max
			cfg.Server.WorkerPoolSize = max
		}
	}
}

func applyAPIOverrides(cfg *Config) {
	if val := os.Getenv("ANTHROPIC_MODEL"); val != "" {
		cfg.API.AnthropicModel = val
	}
	if val := os.Getenv("MAX_TOKENS"); val != "" {
		if tokens, err := strconv.Atoi(val); err == nil {
			cfg.API.MaxTokens = tokens
		}
	}
	if val := os.Getenv("API_MODE"); val != "" {
		cfg.API.Mode = val
	}
	if val := os.Getenv("ANTHROPIC_BASE_URL"); val != "" {
		cfg.API.Mode = "proxy"
		if cfg.API.Proxy == nil {
			cfg.API.Proxy = &ProxyConfig{}
		}
		cfg.API.Proxy.BaseURL = val
	}
}

func applyLoggingOverrides(cfg *Config) {
	if val := os.Getenv("AIPACK_AGENT_SERVER_LOG_LEVEL"); val != "" {
		cfg.Logging.Level = val
	}
	if val := os.Getenv("LOG_FORMAT"); val != "" {
		cfg.Logging.Format = val
	}
}

// applyEnvOverrides applies environment variable overrides to config
func applyEnvOverrides(cfg *Config) {
	applyServerOverrides(cfg)
	applyAPIOverrides(cfg)
	applyLoggingOverrides(cfg)
}

// DataDir returns the root directory for all persistent runtime data
// (performance grades, metrics, execution logs).
//
// Resolution order:
//  1. AGENT_DATA_DIR environment variable — set this for service installs:
//     Linux/macOS service:  AGENT_DATA_DIR=/var/lib/ai-pack
//     Windows service:      AGENT_DATA_DIR=C:\ProgramData\ai-pack
//  2. Platform default (user-mode dev / interactive runs):
//     Windows:  %APPDATA%\ai-pack\   (e.g. C:\Users\name\AppData\Roaming\ai-pack)
//     macOS:    ~/.claude/            (aligns with Claude Code's own data dir)
//     Linux:    ~/.claude/            (aligns with Claude Code's own data dir)
//
// The path is always absolute and independent of the process working directory,
// so /usr/local/bin/agent-server and ./bin/agent-server store data identically.
func DataDir() (string, error) {
	if dir := os.Getenv("AGENT_DATA_DIR"); dir != "" {
		return dir, nil
	}
	if runtime.GOOS == "windows" {
		// %APPDATA% is the standard per-user application data location on Windows.
		// For system services, callers should set AGENT_DATA_DIR=C:\ProgramData\ai-pack.
		appData, err := os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine AppData directory: %w", err)
		}
		return filepath.Join(appData, "ai-pack"), nil
	}
	// Unix/macOS: co-locate with ~/.claude/agent-server.json so all Claude-
	// related runtime data lives in one well-known directory.
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".claude"), nil
}

// ProviderEndpoint returns the configured endpoint override for the given
// provider name, or "" if none is configured.
func (c *Config) ProviderEndpoint(provider string) string {
	if c == nil || c.Providers == nil {
		return ""
	}
	return c.Providers[provider].Endpoint
}

// ProviderAPIKeyEnv returns the configured env-var name for the given
// provider's API key, or the supplied defaultEnv if none is configured.
func (c *Config) ProviderAPIKeyEnv(provider, defaultEnv string) string {
	if c == nil || c.Providers == nil {
		return defaultEnv
	}
	if pc, ok := c.Providers[provider]; ok && pc.APIKeyEnv != "" {
		return pc.APIKeyEnv
	}
	return defaultEnv
}

// SaveConfig saves configuration to file
func SaveConfig(cfg *Config, configPath string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadMCPServers loads MCP server configurations from user and project configs
// Priority order (highest to lowest):
// 1. cfg.MCP.Servers (explicit server config in agent-server.json)
// 2. Project .mcp.json (if projectRoot is provided)
// 3. ~/.claude.json global MCP servers
//
// If cfg.MCP.EnabledServers is non-empty, only those servers will be loaded
func LoadMCPServers(cfg *MCPConfig, projectRoot string) (map[string]MCPServerConfig, error) {
	result := make(map[string]MCPServerConfig)

	// Load from ~/.claude.json (global user config)
	homeDir, err := os.UserHomeDir()
	if err == nil {
		claudeJSONPath := homeDir + "/.claude.json"
		if _, err := os.Stat(claudeJSONPath); err == nil {
			data, err := os.ReadFile(claudeJSONPath)
			if err == nil {
				var claudeConfig struct {
					MCPServers map[string]MCPServerConfig `json:"mcpServers"`
				}
				if err := json.Unmarshal(data, &claudeConfig); err == nil {
					for name, server := range claudeConfig.MCPServers {
						result[name] = server
					}
				}
			}
		}
	}

	// Load from project .mcp.json (if projectRoot provided)
	if projectRoot != "" {
		mcpJSONPath := projectRoot + "/.mcp.json"
		if _, err := os.Stat(mcpJSONPath); err == nil {
			data, err := os.ReadFile(mcpJSONPath)
			if err == nil {
				var mcpConfig struct {
					MCPServers map[string]MCPServerConfig `json:"mcpServers"`
				}
				if err := json.Unmarshal(data, &mcpConfig); err == nil {
					// Project config overrides global config
					for name, server := range mcpConfig.MCPServers {
						result[name] = server
					}
				}
			}
		}
	}

	// Override with explicit server configs from agent-server.json
	for name, server := range cfg.Servers {
		result[name] = server
	}

	// Filter by enabled servers list (if specified)
	if len(cfg.EnabledServers) > 0 {
		filtered := make(map[string]MCPServerConfig)
		for _, name := range cfg.EnabledServers {
			if server, exists := result[name]; exists {
				filtered[name] = server
			}
		}
		result = filtered
	}

	return result, nil
}
