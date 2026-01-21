package ethereum

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Giri-Aayush/starknet-faucet/internal/config"
	"github.com/joho/godotenv"
)

// Config holds Ethereum-specific configuration.
// Distribution settings come from local config.json, secrets from .env
type Config struct {
	// Network is the Ethereum network (sepolia, mainnet)
	Network string

	// RPCURL is the Ethereum RPC endpoint URL (from .env)
	RPCURL string

	// FaucetPrivateKey is the private key of the faucet wallet (from .env)
	FaucetPrivateKey string

	// FaucetAddress is the address of the faucet wallet (from .env)
	FaucetAddress string

	// ChainID is the chain ID for the network
	ChainID int64

	// Token configuration (from local config.json)
	Tokens map[string]config.TokenConfig

	// MinBalanceProtectPct stops distributing when balance drops to this percentage
	MinBalanceProtectPct int

	// ExplorerURL for transaction links
	ExplorerURL string
}

// getChainDir returns the directory where this chain's config.json is located
func getChainDir() string {
	// First, try relative to current working directory
	if _, err := os.Stat("chains/ethereum-sepolia/config.json"); err == nil {
		return "chains/ethereum-sepolia"
	}

	// Try relative to executable
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		chainDir := filepath.Join(execDir, "chains", "ethereum-sepolia")
		if _, err := os.Stat(filepath.Join(chainDir, "config.json")); err == nil {
			return chainDir
		}
	}

	// Fallback: try using runtime.Caller for development
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		return filepath.Dir(filename)
	}

	// Final fallback
	return "chains/ethereum-sepolia"
}

// LoadConfig loads Ethereum configuration from local config.json and .env (secrets)
func LoadConfig() (*Config, error) {
	// Load .env for secrets
	_ = godotenv.Load()

	// Load chain-specific config from this directory's config.json
	chainDir := getChainDir()
	chainConfig, err := config.LoadChainConfig(chainDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load ethereum config.json: %w", err)
	}

	// Load secrets from environment
	rpcURL := os.Getenv("ETHEREUM_RPC_URL")
	if rpcURL == "" {
		return nil, fmt.Errorf("ETHEREUM_RPC_URL is required in .env")
	}

	privateKey := os.Getenv("ETHEREUM_PRIVATE_KEY")
	if privateKey == "" {
		return nil, fmt.Errorf("ETHEREUM_PRIVATE_KEY is required in .env")
	}

	address := os.Getenv("ETHEREUM_ADDRESS")
	if address == "" {
		return nil, fmt.Errorf("ETHEREUM_ADDRESS is required in .env")
	}

	// Get chain ID from config
	var chainID int64 = 11155111 // Default to Sepolia
	if chainConfig.ChainID != nil {
		switch v := chainConfig.ChainID.(type) {
		case float64:
			chainID = int64(v)
		case int:
			chainID = int64(v)
		case int64:
			chainID = v
		}
	}

	// Get network name from chain ID
	network := "sepolia"
	switch chainID {
	case 1:
		network = "mainnet"
	case 5:
		network = "goerli"
	case 11155111:
		network = "sepolia"
	case 17000:
		network = "holesky"
	}

	cfg := &Config{
		Network:              network,
		RPCURL:               rpcURL,
		FaucetPrivateKey:     privateKey,
		FaucetAddress:        address,
		ChainID:              chainID,
		Tokens:               chainConfig.Tokens,
		MinBalanceProtectPct: chainConfig.MinBalanceProtectPct,
		ExplorerURL:          chainConfig.ExplorerURL,
	}

	return cfg, nil
}

// GetDripAmount returns the drip amount for a given token
func (c *Config) GetDripAmount(token string) string {
	if tc, ok := c.Tokens[token]; ok {
		return tc.DripAmount
	}
	return "0"
}

// GetMaxTokensPerHour returns the max hourly distribution limit for a token
func (c *Config) GetMaxTokensPerHour(token string) float64 {
	if tc, ok := c.Tokens[token]; ok {
		return tc.MaxPerHour
	}
	return 0
}

// GetMaxTokensPerDay returns the max daily distribution limit for a token
func (c *Config) GetMaxTokensPerDay(token string) float64 {
	if tc, ok := c.Tokens[token]; ok {
		return tc.MaxPerDay
	}
	return 0
}

// GetMinBalanceProtectPct returns the minimum balance protection percentage
func (c *Config) GetMinBalanceProtectPct() int {
	return c.MinBalanceProtectPct
}

// GetFaucetAddress returns the faucet wallet address
func (c *Config) GetFaucetAddress() string {
	return c.FaucetAddress
}

// GetExplorerURL returns the block explorer URL for transactions
func (c *Config) GetExplorerURL() string {
	return c.ExplorerURL
}
