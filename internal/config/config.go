package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config holds the global application configuration loaded from root config.json and .env
// Chain-specific configuration is loaded separately by each chain from their own config.json
type Config struct {
	// Server settings
	Server ServerConfig `json:"server"`

	// PoW settings
	PoW PoWConfig `json:"pow"`

	// Rate limiting
	RateLimits RateLimitConfig `json:"rate_limits"`

	// From .env (secrets)
	RedisURL string `json:"-"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port     int    `json:"port"`
	LogLevel string `json:"log_level"`
}

// PoWConfig holds proof of work configuration
type PoWConfig struct {
	Difficulty      int `json:"difficulty"`
	ChallengeTTLSec int `json:"challenge_ttl_seconds"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	MaxRequestsPerDayIP  int `json:"max_requests_per_day_ip"`
	MaxChallengesPerHour int `json:"max_challenges_per_hour"`
}

// ChainConfig holds configuration for a specific chain (loaded from chain's config.json)
type ChainConfig struct {
	Name                 string                 `json:"name"`
	ChainID              interface{}            `json:"chain_id"` // Can be string or int
	Tokens               map[string]TokenConfig `json:"tokens"`
	MinBalanceProtectPct int                    `json:"min_balance_protect_pct"`
	ExplorerURL          string                 `json:"explorer_url"`
}

// TokenConfig holds configuration for a specific token
type TokenConfig struct {
	ContractAddress string  `json:"contract_address,omitempty"`
	DripAmount      string  `json:"drip_amount"`
	MaxPerHour      float64 `json:"max_per_hour"`
	MaxPerDay       float64 `json:"max_per_day"`
}

// Load loads global configuration from config directory and .env
// If FAUCET_TEST_MODE=true, loads config/config.test.json (relaxed settings)
// Otherwise, loads config/config.json (production settings)
func Load() (*Config, error) {
	// Load .env for secrets
	_ = godotenv.Load()

	// Determine which config to load based on FAUCET_TEST_MODE
	isTestMode := getEnv("FAUCET_TEST_MODE", "false") == "true"

	// Find the appropriate config file
	configPath, err := FindConfigFile(isTestMode)
	if err != nil {
		return nil, fmt.Errorf("failed to find config file: %w", err)
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", configPath, err)
	}

	config := &Config{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", configPath, err)
	}

	// Load secrets from environment
	config.RedisURL = getEnv("REDIS_URL", "redis://localhost:6379")

	// Validate
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// FindConfigFile looks for the appropriate config file based on test mode
// Test mode: config/config.test.json
// Production: config/config.json
func FindConfigFile(isTestMode bool) (string, error) {
	configFileName := "config.json"
	if isTestMode {
		configFileName = "config.test.json"
	}

	// Try config/ directory in current working directory
	configPath := filepath.Join("config", configFileName)
	if _, err := os.Stat(configPath); err == nil {
		return configPath, nil
	}

	// Try relative to executable
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		configPath = filepath.Join(execDir, "config", configFileName)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}

	// Try working directory
	wd, err := os.Getwd()
	if err == nil {
		configPath = filepath.Join(wd, "config", configFileName)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}

	// Fallback: try root level config.json for backward compatibility
	if !isTestMode {
		if _, err := os.Stat("config.json"); err == nil {
			return "config.json", nil
		}
	}

	return "", fmt.Errorf("%s not found in config/ directory", configFileName)
}

// LoadChainConfig loads a chain's config.json from the specified directory
func LoadChainConfig(chainDir string) (*ChainConfig, error) {
	configPath := filepath.Join(chainDir, "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", configPath, err)
	}

	config := &ChainConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", configPath, err)
	}

	return config, nil
}

// Validate checks if all required configuration is present
func (c *Config) Validate() error {
	if c.RedisURL == "" {
		return &ConfigError{Field: "REDIS_URL", Message: "is required (set in .env)"}
	}

	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}

	if c.PoW.Difficulty == 0 {
		c.PoW.Difficulty = 4
	}

	if c.PoW.ChallengeTTLSec == 0 {
		c.PoW.ChallengeTTLSec = 300
	}

	if c.RateLimits.MaxRequestsPerDayIP == 0 {
		c.RateLimits.MaxRequestsPerDayIP = 5
	}

	if c.RateLimits.MaxChallengesPerHour == 0 {
		c.RateLimits.MaxChallengesPerHour = 10
	}

	return nil
}

// ConfigError represents a configuration error
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return e.Field + " " + e.Message
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Accessor methods for global config

// Port returns the server port
func (c *Config) Port() string {
	return fmt.Sprintf("%d", c.Server.Port)
}

// LogLevel returns the log level
func (c *Config) LogLevel() string {
	return c.Server.LogLevel
}

// PoWDifficulty returns the PoW difficulty
func (c *Config) PoWDifficulty() int {
	return c.PoW.Difficulty
}

// ChallengeTTL returns the challenge TTL in seconds
func (c *Config) ChallengeTTL() int {
	return c.PoW.ChallengeTTLSec
}

// MaxRequestsPerDayIP returns the max requests per day per IP
func (c *Config) MaxRequestsPerDayIP() int {
	return c.RateLimits.MaxRequestsPerDayIP
}

// MaxChallengesPerHour returns the max challenges per hour
func (c *Config) MaxChallengesPerHour() int {
	return c.RateLimits.MaxChallengesPerHour
}
