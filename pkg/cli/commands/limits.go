package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var limitsCmd = &cobra.Command{
	Use:     "limits",
	Aliases: []string{"l"},
	Short:   "Show rate limit rules",
	Long: `View detailed rate limiting rules for the faucet.

USAGE
  faucet-terminal limits -n <network>

EXAMPLES
  faucet-terminal l -n eth
  faucet-terminal limits -n sn`,
	RunE: runLimits,
}

func runLimits(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println("  Rate Limits")
	fmt.Println("  ────────────────────────────────────────")
	fmt.Println()

	fmt.Println("  daily limit (per IP)")
	fmt.Println("    5 requests per day")
	fmt.Println("    single token = 1 request")
	fmt.Println("    --both = 2 requests (Starknet only)")
	fmt.Println("    24h cooldown after 5th request")
	fmt.Println()

	fmt.Println("  hourly throttle (per token)")
	fmt.Println("    1 request per token per hour")
	fmt.Println("    STRK and ETH tracked separately")
	fmt.Println()

	fmt.Println("  amounts per request")
	fmt.Println("    STRK  10 STRK")
	fmt.Println("    ETH   0.01 ETH")
	fmt.Println()

	fmt.Println("  proof of work")
	fmt.Println("    required for each request")
	fmt.Println("    8 challenges per hour max")
	fmt.Println()

	return nil
}
