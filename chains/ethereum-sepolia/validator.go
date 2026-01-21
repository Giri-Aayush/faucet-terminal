package ethereum

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// Ethereum address regex: 0x followed by exactly 40 hex characters
	addressRegex = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
)

// ValidateAddress validates an Ethereum address format.
// Ethereum addresses are 20 bytes (40 hex characters) prefixed with 0x.
func ValidateAddress(address string) error {
	if address == "" {
		return fmt.Errorf("address cannot be empty")
	}

	if !strings.HasPrefix(address, "0x") {
		return fmt.Errorf("address must start with 0x")
	}

	if !addressRegex.MatchString(address) {
		return fmt.Errorf("invalid Ethereum address format: must be 0x followed by 40 hex characters")
	}

	return nil
}

// NormalizeAddress normalizes an Ethereum address to lowercase.
// Ethereum addresses are case-insensitive but conventionally lowercase
// (or checksummed with mixed case, but we normalize to lowercase for comparison).
func NormalizeAddress(address string) string {
	return strings.ToLower(address)
}

// ValidateToken validates a token type for Ethereum.
// For the basic Ethereum faucet, we only support native ETH.
func ValidateToken(token string) error {
	token = strings.ToUpper(token)
	if token != "ETH" {
		return fmt.Errorf("invalid token: Ethereum faucet only supports ETH")
	}
	return nil
}
