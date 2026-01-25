package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

// Config holds all server configuration
type Config struct {
	Server  ServerConfig  `json:"server"`
	API     APIConfig     `json:"api"`
	Logging LoggingConfig `json:"logging"`
	Metrics MetricsConfig `json:"metrics"`
}

// ServerConfig holds server-specific settings
type ServerConfig struct {
	Host               string `json:"host"`
	Port               int    `json:"port"`
	MaxConcurrentAgents int   `json:"max_concurrent_agents"`
	WorkerPoolSize     int    `json:"worker_pool_size"`
}

// APIConfig holds API-related settings
type APIConfig struct {
	AnthropicModel string       `json:"anthropic_model"`
	MaxTokens      int          `json:"max_tokens"`
	TimeoutSeconds int          `json:"timeout_seconds"`
	Mode           string       `json:"mode"`            // "direct" or "proxy"
	Proxy          *ProxyConfig `json:"proxy,omitempty"` // Only used when mode = "proxy"
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

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:               "localhost",
			Port:               8080,
			MaxConcurrentAgents: 10,
			WorkerPoolSize:     10,
		},
		API: APIConfig{
			AnthropicModel: "claude-sonnet-4-5-20250929",
			MaxTokens:      8000,
			TimeoutSeconds: 600,
			Mode:           "direct", // "direct" or "proxy"
			Proxy:          nil,      // No proxy by default
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		Metrics: MetricsConfig{
			Enabled:          true,
			MaxTaskDurations: 1000,
		},
	}
}

// LoadConfig loads configuration from file with environment variable overrides
func LoadConfig(configPath string) (*Config, error) {
	// Start with defaults
	cfg := DefaultConfig()

	// Load from file if it exists
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}
			// File doesn't exist, use defaults
		} else {
			if err := json.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
		}
	}

	// Override with environment variables
	applyEnvOverrides(cfg)

	return cfg, nil
}

// applyEnvOverrides applies environment variable overrides to config
func applyEnvOverrides(cfg *Config) {
	// Server overrides
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

	// API overrides
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
		// If ANTHROPIC_BASE_URL is set, automatically switch to proxy mode
		cfg.API.Mode = "proxy"
		if cfg.API.Proxy == nil {
			cfg.API.Proxy = &ProxyConfig{}
		}
		cfg.API.Proxy.BaseURL = val
	}

	// Logging overrides
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		cfg.Logging.Level = val
	}
	if val := os.Getenv("LOG_FORMAT"); val != "" {
		cfg.Logging.Format = val
	}
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
