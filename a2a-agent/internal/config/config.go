package config

import (
	"encoding/json"
	"fmt"
	"os"
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
	Server          ServerConfig           `json:"server"`
	API             APIConfig              `json:"api"`
	Agent           AgentConfig            `json:"agent"`
	Logging         LoggingConfig          `json:"logging"`
	Metrics         MetricsConfig          `json:"metrics"`
	TaskCleanup     TaskCleanupConfig      `json:"task_cleanup"`
	ProviderCosts   ProviderCostsConfig    `json:"provider_costs"`
	GradingCriteria GradingCriteriaConfig  `json:"grading_criteria"`
	MCP             MCPConfig              `json:"mcp"`
	ComplexityGate  ComplexityGateConfig   `json:"complexity_gate"`
	Projects        map[string]string      `json:"projects,omitempty"` // map[projectPath]lastAccessed
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
	AnthropicModel string       `json:"anthropic_model"`
	MaxTokens      int          `json:"max_tokens"`
	TimeoutSeconds int          `json:"timeout_seconds"`
	Mode           string       `json:"mode"`            // "direct" or "proxy"
	Proxy          *ProxyConfig `json:"proxy,omitempty"` // Only used when mode = "proxy"
}

// AgentConfig holds agent behavior settings
type AgentConfig struct {
	MaxInactiveTurns int `json:"max_inactive_turns"` // Stop agent after N turns without progress
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
	Enabled         bool `json:"enabled"`           // Enable automatic cleanup on startup
	ArchiveAfterDays int  `json:"archive_after_days"` // Archive tasks older than N days
}

// ProviderCostsConfig holds pricing for LLM providers
type ProviderCostsConfig struct {
	Models         map[string]ModelCost `json:"models"`
	LastUpdated    string               `json:"last_updated,omitempty"`    // ISO 8601 timestamp
	UpdateInterval int                  `json:"update_interval,omitempty"` // Days between updates (default: 30)
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
	Enabled       bool                       `json:"enabled"`         // Enable MCP integration
	Servers       map[string]MCPServerConfig `json:"servers"`         // Server configurations (can override user/project configs)
	EnabledServers []string                   `json:"enabled_servers"` // List of enabled server names (filters available servers)
}

// MCPServerConfig defines configuration for an MCP server
type MCPServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:                "localhost",
			Port:                8080,
			MaxConcurrentAgents: 10,
			WorkerPoolSize:      10,
		},
		API: APIConfig{
			AnthropicModel: "claude-sonnet-4-5-20250929",
			MaxTokens:      24000,
			TimeoutSeconds: 600,
			Mode:           "direct", // "direct" or "proxy"
			Proxy:          nil,      // No proxy by default
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
			Enabled:        false,                          // Disabled by default
			Servers:        make(map[string]MCPServerConfig), // Empty by default
			EnabledServers: []string{},                      // Empty means all available servers enabled
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
			RoleMultipliers: map[string]float64{
				"engineer":     1.0,
				"orchestrator": 0.8,
				"inspector":    0.9,
				"architect":    0.7,
			},
		},
		Projects: make(map[string]string),
	}
}

// LoadConfig loads configuration from file with environment variable overrides
// Config search order:
// 1. Explicit configPath parameter (--config flag)
// 2. AGENT_SERVER_CONFIG environment variable
// 3. ~/.claude/agent-server.json (user config - DEFAULT)
// 4. ./agent-server.json (current directory - backward compat)
// 5. Built-in defaults
func LoadConfig(configPath string) (*Config, error) {
	// Start with defaults
	cfg := DefaultConfig()

	// Determine which config file to load
	resolvedPath := resolveConfigPath(configPath)

	// Load from file if it exists
	if resolvedPath != "" {
		data, err := os.ReadFile(resolvedPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to read config file %s: %w", resolvedPath, err)
			}
			// File doesn't exist, use defaults
		} else {
			if err := json.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config file %s: %w", resolvedPath, err)
			}
		}
	}

	// Override with environment variables
	applyEnvOverrides(cfg)

	return cfg, nil
}

// resolveConfigPath resolves the config file path using the search order
func resolveConfigPath(explicitPath string) string {
	// 1. Explicit path provided (--config flag)
	if explicitPath != "" && explicitPath != defaultConfigFilename {
		return explicitPath
	}

	// 2. AGENT_SERVER_CONFIG environment variable
	if envPath := os.Getenv("AGENT_SERVER_CONFIG"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}
	}

	// 3. ~/.claude/agent-server.json (user config - DEFAULT)
	if homeDir, err := os.UserHomeDir(); err == nil {
		claudePath := homeDir + "/.claude/" + defaultConfigFilename
		if _, err := os.Stat(claudePath); err == nil {
			return claudePath
		}
	}

	// 4. ./agent-server.json (current directory - backward compat)
	if _, err := os.Stat(defaultConfigFilename); err == nil {
		return defaultConfigFilename
	}

	// 5. Return empty to use built-in defaults
	return ""
}

func applyServerOverrides(cfg *Config) {
	if val := os.Getenv("SERVER_HOST"); val != "" {
		cfg.Server.Host = val
	}
	if val := os.Getenv("SERVER_PORT"); val != "" {
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
	if val := os.Getenv("LOG_LEVEL"); val != "" {
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
