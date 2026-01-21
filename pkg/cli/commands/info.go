package commands

import (
	"encoding/json"
	"fmt"

	"github.com/Giri-Aayush/starknet-faucet/pkg/cli"
	"github.com/Giri-Aayush/starknet-faucet/pkg/cli/ui"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:     "info",
	Aliases: []string{"i"},
	Short:   "View faucet information",
	Long: `View faucet configuration, limits, and balances.

USAGE
  faucet-terminal info -n <network>

EXAMPLES
  faucet-terminal i -n eth
  faucet-terminal info -n sn`,
	RunE: runInfo,
}

func runInfo(cmd *cobra.Command, args []string) error {
	// Validate network first
	if err := ValidateNetwork(); err != nil {
		return err
	}

	// Create API client with correct URL for network
	client := cli.NewAPIClient(GetAPIURL())

	// Get info
	resp, err := client.GetInfo()
	if err != nil {
		return fmt.Errorf("failed to get info: %w", err)
	}

	// Print response
	if jsonOut {
		jsonBytes, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(jsonBytes))
	} else {
		ui.PrintBanner()
		ui.PrintInfoResponse(resp)
	}

	return nil
}
