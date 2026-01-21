package commands

import (
	"encoding/json"
	"fmt"

	"github.com/Giri-Aayush/starknet-faucet/pkg/cli"
	"github.com/Giri-Aayush/starknet-faucet/pkg/cli/ui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:     "status <address>",
	Aliases: []string{"s"},
	Short:   "Check cooldown status",
	Long: `Check if an address is in cooldown and when it can request tokens.

USAGE
  faucet-terminal status <address> -n <network>

EXAMPLES
  faucet-terminal s 0x123...abc -n eth
  faucet-terminal status 0x123...abc -n sn`,
	Args: cobra.ExactArgs(1),
	RunE: runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	address := args[0]

	// Validate network first
	if err := ValidateNetwork(); err != nil {
		return err
	}

	// Create API client with correct URL for network
	client := cli.NewAPIClient(GetAPIURL())

	// Get status
	resp, err := client.GetStatus(address)
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	// Print response
	if jsonOut {
		jsonBytes, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(jsonBytes))
	} else {
		ui.PrintBanner()
		ui.PrintStatusResponse(resp, address)
	}

	return nil
}
