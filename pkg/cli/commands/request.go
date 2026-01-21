package commands

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Giri-Aayush/starknet-faucet/internal/models"
	"github.com/Giri-Aayush/starknet-faucet/pkg/cli"
	"github.com/Giri-Aayush/starknet-faucet/pkg/cli/captcha"
	clipow "github.com/Giri-Aayush/starknet-faucet/pkg/cli/pow"
	"github.com/Giri-Aayush/starknet-faucet/pkg/cli/ui"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/sha3"
)

var (
	token            string
	both             bool
	skipVerification bool
)

var requestCmd = &cobra.Command{
	Use:     "request <address>",
	Aliases: []string{"req", "r"},
	Short:   "Request testnet tokens",
	Long: `Request testnet tokens for a blockchain address.

USAGE
  faucet-terminal req <address> -n <network> [flags]

NETWORKS
  starknet, sn    STRK (default) or ETH
  ethereum, eth   ETH

EXAMPLES
  faucet-terminal req 0x123...abc -n eth
  faucet-terminal req 0x123...abc -n sn
  faucet-terminal req 0x123...abc -n sn --token ETH
  faucet-terminal req 0x123...abc -n sn --both

FLAGS
  --token    Token to request (ETH, STRK)
  --both     Request both STRK and ETH (Starknet only)`,
	Args: cobra.ExactArgs(1),
	RunE: runRequest,
}

func init() {
	requestCmd.Flags().StringVar(&token, "token", "", "Token to request (ETH, STRK)")
	requestCmd.Flags().BoolVar(&both, "both", false, "Request both ETH and STRK (Starknet only)")
	// Hidden test flag - not shown in help, unique name to prevent exploitation
	requestCmd.Flags().BoolVar(&skipVerification, "skip-verification8922", false, "")
	requestCmd.Flags().MarkHidden("skip-verification8922")
}

func runRequest(cmd *cobra.Command, args []string) error {
	address := args[0]

	// First, validate that network is specified
	if err := ValidateNetwork(); err != nil {
		return err
	}

	selectedNetwork := GetNetwork()

	// Detect address type and check for mismatches
	detectedNetwork := detectAddressNetwork(address)

	// Check for network/address mismatch
	if detectedNetwork != "" && detectedNetwork != selectedNetwork {
		return fmt.Errorf(`address/network mismatch detected!

You provided:
  Address:  %s
  Network:  --%s

The address appears to be a %s address (length: %d chars).

Correct usage:
  faucet request %s --network %s

Address format rules:
  • Starknet: 0x + up to 64 hex chars (66 chars total when padded)
  • Ethereum: 0x + exactly 40 hex chars (42 chars total)`,
			address, selectedNetwork,
			detectedNetwork, len(address),
			address, detectedNetwork)
	}

	// Validate address format strictly
	if err := validateAddress(address, selectedNetwork); err != nil {
		return err
	}

	// Set default token based on network
	if token == "" {
		switch selectedNetwork {
		case "starknet":
			token = "STRK"
		case "ethereum":
			token = "ETH"
		default:
			token = "ETH"
		}
	}

	// Normalize token
	token = strings.ToUpper(token)

	// Handle "both" as token value
	if token == "BOTH" {
		both = true
	}

	// Validate token for the network
	if !both {
		if err := validateToken(token, selectedNetwork); err != nil {
			return err
		}
	}

	// Check if --both is valid for this network
	if both && selectedNetwork != "starknet" {
		return fmt.Errorf("--both flag is only supported on Starknet network")
	}

	// Create API client
	client := cli.NewAPIClient(GetAPIURL())

	// Print banner (unless JSON output)
	if !jsonOut {
		ui.PrintBanner()
		ui.PrintNetworkInfo(selectedNetwork)

		// Skip captcha if --skip-verification is set
		if !skipVerification {
			// Ask verification question (3 attempts)
			correct, err := captcha.AskQuestionWithRetries(3)
			if err != nil {
				return fmt.Errorf("verification failed: %w", err)
			}
			if !correct {
				return fmt.Errorf("verification failed - please try again later")
			}
			fmt.Println()
		} else {
			ui.PrintStep("skipping verification (test mode)")
		}
	}

	// Request tokens
	if both {
		// Request STRK first, then ETH
		if err := requestSingleToken(client, address, "STRK"); err != nil {
			return err
		}
		fmt.Println() // Add spacing
		if err := requestSingleToken(client, address, "ETH"); err != nil {
			return err
		}
	} else {
		if err := requestSingleToken(client, address, token); err != nil {
			return err
		}
	}

	return nil
}

// detectAddressNetwork tries to detect which network an address belongs to
func detectAddressNetwork(address string) string {
	if !strings.HasPrefix(address, "0x") {
		return ""
	}

	hexPart := address[2:]

	// Check if all chars are valid hex
	for _, c := range hexPart {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return ""
		}
	}

	// Ethereum addresses are exactly 40 hex chars (20 bytes)
	if len(hexPart) == 40 {
		return "ethereum"
	}

	// Starknet addresses can be 1-64 hex chars, but typically:
	// - Full addresses are 64 chars (252 bits / 4 bits per hex = 63, rounded to 64)
	// - Short addresses like 0x1 are valid but rare for user wallets
	if len(hexPart) > 40 && len(hexPart) <= 64 {
		return "starknet"
	}

	// Ambiguous - could be either a short Starknet address or invalid
	if len(hexPart) < 40 {
		// Most likely a short Starknet address (like contract addresses)
		return "starknet"
	}

	return ""
}

// validateAddress validates address format strictly based on network
func validateAddress(address, network string) error {
	if !strings.HasPrefix(address, "0x") {
		return fmt.Errorf(`invalid address: must start with 0x

You provided: %s

Correct format:
  • Starknet: 0x + up to 64 hex characters
  • Ethereum: 0x + exactly 40 hex characters`, address)
	}

	hexPart := address[2:]

	// Check for valid hex characters
	for i, c := range hexPart {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return fmt.Errorf(`invalid address: contains non-hex character '%c' at position %d

You provided: %s

Addresses must only contain: 0-9, a-f, A-F (after 0x prefix)`, c, i+2, address)
		}
	}

	switch network {
	case "starknet":
		return validateStarknetAddress(address, hexPart)
	case "ethereum":
		return validateEthereumAddress(address, hexPart)
	default:
		return fmt.Errorf("unsupported network: %s", network)
	}
}

// validateStarknetAddress validates a Starknet address
// Starknet addresses are 252-bit field elements, represented as up to 64 hex characters
func validateStarknetAddress(address, hexPart string) error {
	if len(hexPart) == 0 {
		return fmt.Errorf(`invalid Starknet address: empty hex part

You provided: %s

Starknet addresses must have at least one hex character after 0x`, address)
	}

	if len(hexPart) > 64 {
		return fmt.Errorf(`invalid Starknet address: too long (%d hex chars, max 64)

You provided: %s

Starknet addresses are 252-bit field elements (max 64 hex characters)`, len(hexPart), address)
	}

	// Check if this might be an Ethereum address used with wrong network
	if len(hexPart) == 40 {
		return fmt.Errorf(`address looks like an Ethereum address (40 hex chars)

You provided: %s
Network:      --network starknet

If this is an Ethereum address, use:
  faucet request %s --network ethereum

If this is a Starknet address, it should have more characters (typically 64)`, address, address)
	}

	return nil
}

// validateEthereumAddress validates an Ethereum address with EIP-55 checksum support
// Ethereum addresses are 20 bytes = 40 hex characters
func validateEthereumAddress(address, hexPart string) error {
	if len(hexPart) != 40 {
		// Check if it might be a Starknet address
		if len(hexPart) > 40 && len(hexPart) <= 64 {
			return fmt.Errorf(`address looks like a Starknet address (%d hex chars)

You provided: %s
Network:      --network ethereum

This appears to be a Starknet address. Use:
  faucet request %s --network starknet

Ethereum addresses are exactly 42 characters (0x + 40 hex chars)`, len(hexPart), address, address)
		}

		return fmt.Errorf(`invalid Ethereum address: wrong length (%d hex chars, need exactly 40)

You provided: %s

Ethereum addresses must be exactly 42 characters:
  • 0x prefix (2 chars)
  • 40 hex characters (20 bytes)

Example: 0x742d35Cc6634C0532925a3b844Bc9e7595f8d9f1`, len(hexPart), address)
	}

	// Validate EIP-55 checksum if address has mixed case
	if hasMixedCase(hexPart) {
		if !isValidEIP55Checksum(address) {
			// Get the correctly checksummed version
			correctChecksum := toEIP55Checksum(address)
			return fmt.Errorf(`invalid Ethereum address: EIP-55 checksum failed

You provided: %s
Expected:     %s

The address has invalid capitalization. EIP-55 uses mixed-case for error detection.
Either use the correctly checksummed address above, or use all lowercase:
  %s`, address, correctChecksum, strings.ToLower(address))
		}
	}

	return nil
}

// hasMixedCase checks if a hex string has both uppercase and lowercase letters
func hasMixedCase(s string) bool {
	hasUpper := false
	hasLower := false
	for _, c := range s {
		if c >= 'A' && c <= 'F' {
			hasUpper = true
		}
		if c >= 'a' && c <= 'f' {
			hasLower = true
		}
	}
	return hasUpper && hasLower
}

// isValidEIP55Checksum validates EIP-55 checksum
func isValidEIP55Checksum(address string) bool {
	return address == toEIP55Checksum(address)
}

// toEIP55Checksum converts an Ethereum address to EIP-55 checksummed format
func toEIP55Checksum(address string) string {
	// Remove 0x prefix and lowercase
	addr := strings.ToLower(address[2:])

	// Keccak-256 hash of the lowercase address
	hash := sha3.NewLegacyKeccak256()
	hash.Write([]byte(addr))
	hashBytes := hash.Sum(nil)
	hashHex := hex.EncodeToString(hashBytes)

	// Apply checksum: uppercase if hash nibble >= 8
	result := "0x"
	for i, c := range addr {
		if c >= '0' && c <= '9' {
			result += string(c)
		} else {
			// Get the corresponding nibble from hash
			hashNibble := hashHex[i]
			if hashNibble >= '8' {
				result += strings.ToUpper(string(c))
			} else {
				result += string(c)
			}
		}
	}

	return result
}

// validateToken validates token for the network
func validateToken(token, network string) error {
	switch network {
	case "starknet":
		if token != "ETH" && token != "STRK" {
			return fmt.Errorf(`invalid token '%s' for Starknet

Supported tokens for Starknet Sepolia:
  • STRK  (default)
  • ETH

Usage:
  faucet request <ADDRESS> --network starknet             # STRK (default)
  faucet request <ADDRESS> --network starknet --token ETH # ETH
  faucet request <ADDRESS> --network starknet --both      # Both`, token)
		}
	case "ethereum":
		if token != "ETH" {
			return fmt.Errorf(`invalid token '%s' for Ethereum

Supported tokens for Ethereum Sepolia:
  • ETH  (only native ETH is supported)

Usage:
  faucet request <ADDRESS> --network ethereum`, token)
		}
	default:
		return fmt.Errorf("unsupported network: %s", network)
	}
	return nil
}

func requestSingleToken(client *cli.APIClient, address, token string) error {
	if !jsonOut {
		ui.PrintInfo(fmt.Sprintf("Requesting %s for %s", token, address))
		fmt.Println()
	}

	// Step 1: Get challenge
	var challengeResp *models.ChallengeResponse
	if !jsonOut {
		s := ui.NewSpinner("Fetching challenge...")
		s.Start()
		var err error
		challengeResp, err = client.GetChallenge()
		s.Stop()
		if err != nil {
			ui.PrintError(fmt.Sprintf("Failed to get challenge: %v", err))
			return err
		}
		ui.PrintSuccess("Challenge received")
		fmt.Println()
	} else {
		var err error
		challengeResp, err = client.GetChallenge()
		if err != nil {
			return err
		}
	}

	// Step 2: Solve PoW (or skip in test mode)
	var nonce int64
	var solveDuration time.Duration

	if skipVerification {
		// Test mode: use magic nonce -1 (server must have FAUCET_TEST_MODE=1)
		nonce = -1
		solveDuration = 0
		if !jsonOut {
			ui.PrintInfo("Skipping PoW (test mode, nonce: -1)")
			fmt.Println()
		}
	} else if !jsonOut {
		s := ui.NewSpinner(fmt.Sprintf("Solving proof of work (difficulty: %d)...", challengeResp.Difficulty))
		s.Start()

		solver := clipow.NewSolver()
		result, err := solver.Solve(challengeResp.Challenge, challengeResp.Difficulty, func(n int64, d time.Duration) {
			// Update spinner suffix with progress
			s.Suffix = fmt.Sprintf(" Solving proof of work (attempts: %d, time: %.1fs)...",
				n, d.Seconds())
		})

		s.Stop()

		if err != nil {
			ui.PrintError(fmt.Sprintf("Failed to solve challenge: %v", err))
			return err
		}

		nonce = result.Nonce
		solveDuration = result.Duration
		ui.PrintSuccess(fmt.Sprintf("Challenge solved in %.1fs (nonce: %d)", solveDuration.Seconds(), nonce))
		fmt.Println()
	} else {
		solver := clipow.NewSolver()
		result, err := solver.Solve(challengeResp.Challenge, challengeResp.Difficulty, nil)
		if err != nil {
			return err
		}
		nonce = result.Nonce
		solveDuration = result.Duration
	}

	// Step 3: Request tokens
	req := models.FaucetRequest{
		Address:     address,
		Token:       token,
		Network:     GetNetwork(),
		ChallengeID: challengeResp.ChallengeID,
		Nonce:       nonce,
	}

	var faucetResp *models.FaucetResponse
	if !jsonOut {
		s := ui.NewSpinner("Submitting request...")
		s.Start()
		var err error
		faucetResp, err = client.RequestTokens(req)
		s.Stop()
		if err != nil {
			ui.PrintError(fmt.Sprintf("Failed to request tokens: %v", err))
			return err
		}
		ui.PrintSuccess("Transaction submitted!")
	} else {
		var err error
		faucetResp, err = client.RequestTokens(req)
		if err != nil {
			return err
		}
	}

	// Print response
	if jsonOut {
		output := map[string]interface{}{
			"success":        faucetResp.Success,
			"tx_hash":        faucetResp.TxHash,
			"amount":         faucetResp.Amount,
			"token":          faucetResp.Token,
			"explorer_url":   faucetResp.ExplorerURL,
			"solve_duration": solveDuration.Seconds(),
		}
		jsonBytes, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(jsonBytes))
	} else {
		ui.PrintFaucetResponse(faucetResp)
	}

	return nil
}
