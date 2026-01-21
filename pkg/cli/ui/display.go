package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/Giri-Aayush/starknet-faucet/internal/models"
)

var (
	// Colors - minimal palette for clean UI
	cyan   = color.New(color.FgCyan).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	dim    = color.New(color.Faint).SprintFunc()
	bold   = color.New(color.Bold).SprintFunc()
	white  = color.New(color.FgWhite).SprintFunc()

	// Symbols - clean and minimal
	symbolSuccess = green("✔")
	symbolError   = red("✖")
	symbolInfo    = cyan("›")
	symbolArrow   = dim("→")
)

// PrintBanner prints a clean minimal banner
func PrintBanner() {
	fmt.Println()
	fmt.Printf("  %s %s\n", bold("faucet"), dim("terminal"))
	fmt.Println()
}

// PrintNetworkInfo prints the selected network
func PrintNetworkInfo(network string) {
	fmt.Printf("  %s %s\n", dim("network"), white(network))
	fmt.Println()
}

// PrintSuccess prints a success message
func PrintSuccess(message string) {
	fmt.Printf("  %s %s\n", symbolSuccess, message)
}

// PrintError prints an error message
func PrintError(message string) {
	fmt.Printf("  %s %s\n", symbolError, red(message))
}

// PrintInfo prints an info message
func PrintInfo(message string) {
	fmt.Printf("  %s %s\n", symbolInfo, message)
}

// PrintStep prints a step in progress
func PrintStep(message string) {
	fmt.Printf("  %s %s\n", symbolArrow, dim(message))
}

// NewSpinner creates a new spinner with a message
func NewSpinner(message string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 80*time.Millisecond)
	s.Prefix = "  "
	s.Suffix = " " + message
	s.Color("cyan")
	return s
}

// PrintFaucetResponse prints a nicely formatted faucet response
func PrintFaucetResponse(resp *models.FaucetResponse) {
	fmt.Println()

	// Check if this is a BOTH token response (multiple transactions)
	if len(resp.Transactions) > 0 {
		for _, tx := range resp.Transactions {
			printTransactionBox(tx.Token, tx.Amount, tx.TxHash, tx.ExplorerURL)
		}
		fmt.Println()
		PrintSuccess(resp.Message)
		return
	}

	// Single token response
	printTransactionBox(resp.Token, resp.Amount, resp.TxHash, resp.ExplorerURL)
	fmt.Println()
	PrintSuccess("Tokens will arrive in ~30 seconds")
}

// printTransactionBox prints a clean transaction summary
func printTransactionBox(token, amount, txHash, explorerURL string) {
	fmt.Printf("  %s\n", dim(strings.Repeat("─", 52)))
	fmt.Println()
	fmt.Printf("    %s  %s %s\n", dim("amount"), bold(amount), token)
	fmt.Printf("    %s  %s\n", dim("tx"), shortenHash(txHash))
	fmt.Println()
	fmt.Printf("    %s\n", cyan(explorerURL))
	fmt.Println()
	fmt.Printf("  %s\n", dim(strings.Repeat("─", 52)))
}

// PrintStatusResponse prints a status response
func PrintStatusResponse(resp *models.StatusResponse, address string) {
	fmt.Println()
	fmt.Printf("  %s %s\n", dim("address"), shortenHash(address))
	fmt.Println()

	if resp.CanRequest {
		PrintSuccess("Ready to request tokens")
	} else {
		PrintError("Cooldown active")
		fmt.Println()
		if resp.NextRequestTime != nil {
			fmt.Printf("    %s %s\n", dim("available"), resp.NextRequestTime.Format("Jan 02 at 3:04 PM"))
		}
		if resp.RemainingHours != nil {
			fmt.Printf("    %s %s\n", dim("remaining"), formatDuration(*resp.RemainingHours))
		}
	}
	fmt.Println()
}

// PrintInfoResponse prints faucet information
func PrintInfoResponse(resp *models.InfoResponse) {
	fmt.Println()
	fmt.Printf("  %s\n", bold("Faucet Info"))
	fmt.Printf("  %s\n", dim(strings.Repeat("─", 40)))
	fmt.Println()

	fmt.Printf("  %s  %s\n", dim("network"), resp.Network)
	fmt.Println()

	fmt.Printf("  %s\n", dim("limits"))
	if resp.Limits.StrkPerRequest != "" && resp.Limits.StrkPerRequest != "0" {
		fmt.Printf("    STRK   %s per request\n", resp.Limits.StrkPerRequest)
	}
	if resp.Limits.EthPerRequest != "" && resp.Limits.EthPerRequest != "0" {
		fmt.Printf("    ETH    %s per request\n", resp.Limits.EthPerRequest)
	}
	fmt.Printf("    Daily  %d requests/IP\n", resp.Limits.DailyRequestsPerIP)
	fmt.Println()

	fmt.Printf("  %s\n", dim("proof of work"))
	fmt.Printf("    difficulty %d\n", resp.PoW.Difficulty)
	fmt.Println()

	fmt.Printf("  %s\n", dim("balance"))
	if resp.FaucetBalance.STRK != "" && resp.FaucetBalance.STRK != "0" {
		fmt.Printf("    STRK %s\n", resp.FaucetBalance.STRK)
	}
	if resp.FaucetBalance.ETH != "" && resp.FaucetBalance.ETH != "0" {
		fmt.Printf("    ETH  %s\n", resp.FaucetBalance.ETH)
	}
	fmt.Println()
}

// PrintCooldownError prints a cooldown error with details
func PrintCooldownError(nextRequestTime *time.Time, remainingHours *float64) {
	fmt.Println()
	PrintError("Cooldown active")
	fmt.Println()
	if nextRequestTime != nil {
		fmt.Printf("    %s %s\n", dim("available"), nextRequestTime.Format("Jan 02 at 3:04 PM"))
	}
	if remainingHours != nil {
		fmt.Printf("    %s %s\n", dim("remaining"), formatDuration(*remainingHours))
	}
	fmt.Println()
}

// PrintQuotaInfo prints quota information
func PrintQuotaInfo(used, total int, inCooldown bool) {
	fmt.Println()
	fmt.Printf("  %s\n", bold("Daily Quota"))
	fmt.Printf("  %s\n", dim(strings.Repeat("─", 30)))
	fmt.Println()
	fmt.Printf("  %s  %d/%d requests\n", dim("used"), used, total)
	if inCooldown {
		fmt.Printf("  %s  %s\n", dim("status"), red("in cooldown"))
	} else {
		fmt.Printf("  %s  %s\n", dim("status"), green("available"))
	}
	fmt.Println()
}

// Helper functions

func shortenHash(hash string) string {
	if len(hash) <= 20 {
		return hash
	}
	return hash[:10] + "..." + hash[len(hash)-8:]
}

func formatDuration(hours float64) string {
	if hours >= 24 {
		days := int(hours / 24)
		h := int(hours) % 24
		if h == 0 {
			return fmt.Sprintf("%dd", days)
		}
		return fmt.Sprintf("%dd %dh", days, h)
	}

	if hours >= 1 {
		h := int(hours)
		m := int((hours - float64(h)) * 60)
		if m == 0 {
			return fmt.Sprintf("%dh", h)
		}
		return fmt.Sprintf("%dh %dm", h, m)
	}

	m := int(hours * 60)
	return fmt.Sprintf("%dm", m)
}
