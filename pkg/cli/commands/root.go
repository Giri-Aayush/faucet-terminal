package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	apiURL  string
	network string
	verbose bool
	jsonOut bool
)

// Network to API URL mapping
var networkURLs = map[string]string{
	"starknet": "https://disgusted-melodee-aayushgiri-575fc666.koyeb.app",
	"ethereum": "https://disgusted-melodee-aayushgiri-575fc666.koyeb.app",
}

// Network aliases for developer convenience
var networkAliases = map[string]string{
	"sn":      "starknet",
	"stark":   "starknet",
	"sn-sep":  "starknet",
	"eth":     "ethereum",
	"eth-sep": "ethereum",
}

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "faucet-terminal",
	Short: "Multi-chain testnet token faucet",
	Long: `faucet-terminal — Multi-chain testnet tokens

USAGE
  faucet-terminal <command> [flags]

NETWORKS
  starknet   Starknet Sepolia (aliases: sn, sn-sep)
  ethereum   Ethereum Sepolia (aliases: eth, eth-sep)

EXAMPLES
  faucet-terminal req 0x123...abc -n eth
  faucet-terminal req 0x123...abc -n sn --token ETH
  faucet-terminal req 0x123...abc -n starknet --both
  faucet-terminal info -n eth
  faucet-terminal quota -n sn

COMMANDS
  req, request    Request testnet tokens
  status          Check address cooldown status
  info            View faucet information
  quota           Check your remaining quota
  limits          Show rate limit rules

FLAGS
  -n, --network   Network to use (required)
  -h, --help      Show help

https://github.com/Giri-Aayush/faucet-cli`,
	Version: "1.1.0",
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags with short versions
	rootCmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Network: starknet|sn-sep, ethereum|eth-sep")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "", "Override API URL")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "JSON output")

	// Add subcommands
	rootCmd.AddCommand(requestCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(limitsCmd)
	rootCmd.AddCommand(quotaCmd)
}

// resolveNetwork converts network aliases to full network names
func resolveNetwork(n string) string {
	n = strings.ToLower(n)
	if alias, ok := networkAliases[n]; ok {
		return alias
	}
	return n
}

// ValidateNetwork checks if network is specified and valid
func ValidateNetwork() error {
	if network == "" {
		return fmt.Errorf(`network required

Use -n or --network:
  -n starknet    Starknet Sepolia (or: sn, sn-sep)
  -n ethereum    Ethereum Sepolia (or: eth, eth-sep)

Example:
  faucet-terminal req 0xADDRESS -n eth`)
	}

	// Resolve aliases
	network = resolveNetwork(network)

	validNetworks := []string{"starknet", "ethereum"}
	for _, valid := range validNetworks {
		if network == valid {
			return nil
		}
	}

	return fmt.Errorf(`invalid network: %s

Supported:
  starknet (sn)    Starknet Sepolia
  ethereum (eth)   Ethereum Sepolia`, network)
}

// GetAPIURL returns the API URL for the selected network
func GetAPIURL() string {
	if apiURL != "" {
		return apiURL
	}

	resolved := resolveNetwork(network)
	if url, ok := networkURLs[resolved]; ok {
		return url
	}

	return networkURLs["starknet"]
}

// GetNetwork returns the selected network (resolved from alias)
func GetNetwork() string {
	return resolveNetwork(network)
}
