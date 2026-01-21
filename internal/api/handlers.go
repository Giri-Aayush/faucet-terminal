package api

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/Giri-Aayush/starknet-faucet/chains"
	"github.com/Giri-Aayush/starknet-faucet/internal/cache"
	"github.com/Giri-Aayush/starknet-faucet/internal/config"
	"github.com/Giri-Aayush/starknet-faucet/internal/models"
	"github.com/Giri-Aayush/starknet-faucet/internal/pow"
	"go.uber.org/zap"
)

// ChainProvider provides chain-specific configuration.
type ChainProvider interface {
	GetDripAmount(token string) string
	GetMaxTokensPerHour(token string) float64
	GetMaxTokensPerDay(token string) float64
	GetMinBalanceProtectPct() int
	GetFaucetAddress() string
}

// Handler contains dependencies for API handlers
type Handler struct {
	config            *config.Config
	logger            *zap.Logger
	redis             *cache.RedisClient
	chains            map[string]chains.Chain
	providers         map[string]ChainProvider
	powGenerator      *pow.Generator
	defaultNetwork    string
}

// NewMultiChainHandler creates a new multi-chain API handler
func NewMultiChainHandler(
	cfg *config.Config,
	logger *zap.Logger,
	redis *cache.RedisClient,
	chainRegistry map[string]chains.Chain,
	providerRegistry map[string]ChainProvider,
	powGenerator *pow.Generator,
) *Handler {
	// Determine default network (prefer starknet if available)
	defaultNetwork := ""
	if _, ok := chainRegistry["starknet"]; ok {
		defaultNetwork = "starknet"
	} else if _, ok := chainRegistry["ethereum"]; ok {
		defaultNetwork = "ethereum"
	} else {
		// Use first available
		for name := range chainRegistry {
			defaultNetwork = name
			break
		}
	}

	return &Handler{
		config:         cfg,
		logger:         logger,
		redis:          redis,
		chains:         chainRegistry,
		providers:      providerRegistry,
		powGenerator:   powGenerator,
		defaultNetwork: defaultNetwork,
	}
}

// NewHandler creates a new API handler (backward compatible - single chain)
func NewHandler(
	cfg *config.Config,
	logger *zap.Logger,
	redis *cache.RedisClient,
	chain chains.Chain,
	chainProvider ChainProvider,
	powGenerator *pow.Generator,
) *Handler {
	chainName := chain.GetChainName()
	return &Handler{
		config:         cfg,
		logger:         logger,
		redis:          redis,
		chains:         map[string]chains.Chain{chainName: chain},
		providers:      map[string]ChainProvider{chainName: chainProvider},
		powGenerator:   powGenerator,
		defaultNetwork: chainName,
	}
}

// getChain returns the chain and provider for the given network
func (h *Handler) getChain(network string) (chains.Chain, ChainProvider, error) {
	if network == "" {
		network = h.defaultNetwork
	}

	chain, ok := h.chains[network]
	if !ok {
		available := make([]string, 0, len(h.chains))
		for name := range h.chains {
			available = append(available, name)
		}
		return nil, nil, fmt.Errorf("unsupported network: %s (available: %s)", network, strings.Join(available, ", "))
	}

	provider, ok := h.providers[network]
	if !ok {
		return nil, nil, fmt.Errorf("provider not found for network: %s", network)
	}

	return chain, provider, nil
}

// GetChallenge generates a new PoW challenge
func (h *Handler) GetChallenge(c *fiber.Ctx) error {
	ctx := context.Background()

	// Check challenge rate limit for this IP
	ip := c.IP()
	canRequest, err := h.redis.CheckChallengeRateLimit(ctx, ip)
	if err != nil {
		h.logger.Error("Failed to check challenge rate limit", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to check rate limit",
		})
	}
	if !canRequest {
		return c.Status(fiber.StatusTooManyRequests).JSON(models.ErrorResponse{
			Error: "[CHALLENGE LIMIT] Too many PoW challenge requests this hour. Try again later.",
		})
	}

	// Generate challenge
	response, challenge, err := h.powGenerator.GenerateChallenge()
	if err != nil {
		h.logger.Error("Failed to generate challenge", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to generate challenge",
		})
	}

	// Store challenge in Redis
	ttl := time.Duration(h.config.ChallengeTTL()) * time.Second
	if err := h.redis.StoreChallenge(ctx, challenge.ID, challenge.Challenge, ttl); err != nil {
		h.logger.Error("Failed to store challenge", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to store challenge",
		})
	}

	// Increment challenge rate limit counter
	if err := h.redis.IncrementChallengeRateLimit(ctx, ip); err != nil {
		h.logger.Error("Failed to increment challenge rate limit", zap.Error(err))
	}

	h.logger.Info("Challenge generated",
		zap.String("challenge_id", challenge.ID),
		zap.String("ip", ip),
	)

	return c.JSON(response)
}

// RequestTokens handles faucet requests
func (h *Handler) RequestTokens(c *fiber.Ctx) error {
	ctx := context.Background()

	// Parse request
	var req models.FaucetRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Get the chain for the specified network
	chain, chainProvider, err := h.getChain(req.Network)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: err.Error(),
		})
	}

	// Validate address using chain-specific validation
	if err := chain.ValidateAddress(req.Address); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Invalid address: %s", err.Error()),
		})
	}

	// Validate token using chain-specific validation
	req.Token = strings.ToUpper(req.Token)
	if req.Token != "BOTH" {
		if err := chain.ValidateToken(req.Token); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Error: err.Error(),
			})
		}
	}

	// NEW SIMPLIFIED RATE LIMITING
	ip := c.IP()

	// 1. Check IP daily limit (5 requests/day) and 24h cooldown
	canRequest, currentCount, cooldownEnd, err := h.redis.CheckIPDailyLimit(ctx, ip)
	if err != nil {
		h.logger.Error("Failed to check IP daily limit", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to check rate limit",
		})
	}

	// If in 24h cooldown after hitting limit
	if !canRequest && cooldownEnd != nil {
		remaining := time.Until(*cooldownEnd)
		hours := int(remaining.Hours())
		minutes := int(remaining.Minutes()) % 60
		var timeStr string
		if hours > 0 {
			timeStr = fmt.Sprintf("%dh %dm", hours, minutes)
		} else {
			timeStr = fmt.Sprintf("%dm", minutes)
		}
		errorMsg := fmt.Sprintf("[DAILY LIMIT] You've used all 5 daily requests. 24-hour cooldown: %s remaining.", timeStr)
		return c.Status(fiber.StatusTooManyRequests).JSON(models.ErrorResponse{
			Error: errorMsg,
		})
	}

	// Calculate how many requests this will consume (1 for single token, 2 for BOTH)
	requestCost := 1
	if req.Token == "BOTH" {
		requestCost = 2
	}

	// Check if there's enough quota
	if !canRequest || (currentCount+requestCost) > h.config.MaxRequestsPerDayIP() {
		used, _, _, _ := h.redis.GetIPDailyQuota(ctx, ip)
		errorMsg := fmt.Sprintf("[DAILY LIMIT] Request would exceed daily limit (%d/%d used). Wait for quota reset.",
			used, h.config.MaxRequestsPerDayIP())
		return c.Status(fiber.StatusTooManyRequests).JSON(models.ErrorResponse{
			Error: errorMsg,
		})
	}

	// 2. Check per-token hourly throttle (per-network: Starknet ETH and Ethereum ETH have separate throttles)
	network := req.Network
	if network == "" {
		network = h.defaultNetwork
	}

	if req.Token == "BOTH" {
		// For BOTH, check all supported tokens for this network
		supportedTokens := chain.GetSupportedTokens()
		for _, token := range supportedTokens {
			canRequestToken, nextTime, err := h.redis.CheckTokenHourlyThrottle(ctx, ip, network, token)
			if err != nil {
				h.logger.Error("Failed to check token throttle", zap.Error(err), zap.String("token", token))
				return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
					Error: "Failed to check rate limit",
				})
			}
			if !canRequestToken {
				minutesRemaining := int(time.Until(*nextTime).Minutes()) + 1 // +1 to round up
				errorMsg := fmt.Sprintf("[HOURLY LIMIT] %s on %s: 1 request per hour. Try again in %d minutes.",
					token, network, minutesRemaining)
				return c.Status(fiber.StatusTooManyRequests).JSON(models.ErrorResponse{
					Error: errorMsg,
				})
			}
		}
	} else {
		// For single token, check that token's throttle on this network
		canRequestToken, nextAvailable, err := h.redis.CheckTokenHourlyThrottle(ctx, ip, network, req.Token)
		if err != nil {
			h.logger.Error("Failed to check token throttle", zap.Error(err), zap.String("token", req.Token))
			return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
				Error: "Failed to check rate limit",
			})
		}
		if !canRequestToken {
			minutesRemaining := int(time.Until(*nextAvailable).Minutes()) + 1 // +1 to round up
			errorMsg := fmt.Sprintf("[HOURLY LIMIT] %s on %s: 1 request per hour. Try again in %d minutes.",
				req.Token, network, minutesRemaining)
			return c.Status(fiber.StatusTooManyRequests).JSON(models.ErrorResponse{
				Error: errorMsg,
			})
		}
	}

	// Verify challenge exists
	storedChallenge, err := h.redis.GetChallenge(ctx, req.ChallengeID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid or expired challenge",
		})
	}

	// Verify PoW solution
	if !h.powGenerator.VerifyPoW(storedChallenge, req.Nonce, h.config.PoWDifficulty()) {
		h.logger.Warn("Invalid PoW solution",
			zap.String("challenge_id", req.ChallengeID),
			zap.Int64("nonce", req.Nonce),
			zap.String("ip", ip),
		)
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid proof of work solution",
		})
	}

	// Delete challenge to prevent reuse
	if err := h.redis.DeleteChallenge(ctx, req.ChallengeID); err != nil {
		h.logger.Error("Failed to delete challenge", zap.Error(err))
	}

	// Handle BOTH token request
	if req.Token == "BOTH" {
		return h.handleBothTokensRequest(c, ctx, req, ip, chain, chainProvider)
	}

	// Determine amount (single token) using chain provider
	amountStr := chainProvider.GetDripAmount(req.Token)
	amountFloat, _ := strconv.ParseFloat(amountStr, 64)
	maxHourly := chainProvider.GetMaxTokensPerHour(req.Token)
	maxDaily := chainProvider.GetMaxTokensPerDay(req.Token)

	// Check global distribution limits (anti-drain protection)
	canDistribute, err := h.redis.TrackGlobalDistribution(ctx, req.Token, amountFloat, maxHourly, maxDaily)
	if err != nil {
		h.logger.Error("Failed to check global distribution limits", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to process request",
		})
	}
	if !canDistribute {
		h.logger.Warn("Global distribution limit reached",
			zap.String("token", req.Token),
			zap.String("ip", ip),
		)
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ErrorResponse{
			Error: "[FAUCET LIMIT] Faucet has temporarily reached its distribution limit. Please try again in an hour.",
		})
	}

	// Check minimum balance protection (stop at configured percentage)
	currentBalance, err := chain.GetBalance(ctx, chainProvider.GetFaucetAddress(), req.Token)
	if err != nil {
		h.logger.Error("Failed to check faucet balance", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to check faucet balance",
		})
	}

	// Convert amount to wei for comparison
	amountWei := chains.AmountToWei(amountFloat)

	// Check if balance would drop below minimum threshold
	minBalancePct := float64(chainProvider.GetMinBalanceProtectPct()) / 100.0
	currentBalanceFloat := chains.WeiToAmount(currentBalance)
	minBalanceRequired := currentBalanceFloat * minBalancePct
	balanceAfterTransfer := currentBalanceFloat - amountFloat

	if balanceAfterTransfer < minBalanceRequired {
		h.logger.Warn("Balance protection triggered",
			zap.String("token", req.Token),
			zap.Float64("current_balance", currentBalanceFloat),
			zap.Float64("min_balance_required", minBalanceRequired),
			zap.String("ip", ip),
		)
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("[LOW BALANCE] Faucet %s balance too low (%.4f). Please try again later.", req.Token, currentBalanceFloat),
		})
	}

	// Transfer tokens
	h.logger.Info("Transferring tokens",
		zap.String("network", req.Network),
		zap.String("recipient", req.Address),
		zap.String("token", req.Token),
		zap.String("amount", amountStr),
		zap.String("ip", ip),
	)

	txHash, err := chain.TransferTokens(ctx, req.Address, req.Token, amountWei)
	if err != nil {
		h.logger.Error("Failed to transfer tokens",
			zap.Error(err),
			zap.String("recipient", req.Address),
			zap.String("token", req.Token),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to send tokens. Please try again later.",
		})
	}

	// Increment IP daily counter (1 for single token)
	if err := h.redis.IncrementIPDailyLimit(ctx, ip, 1); err != nil {
		h.logger.Error("Failed to increment IP daily limit", zap.Error(err))
	}

	// Set token hourly throttle (1 hour cooldown for this token on this network)
	if err := h.redis.SetTokenHourlyThrottle(ctx, ip, network, req.Token); err != nil {
		h.logger.Error("Failed to set token throttle", zap.Error(err))
	}

	// Build response
	response := models.FaucetResponse{
		Success:     true,
		TxHash:      txHash,
		Amount:      amountStr,
		Token:       req.Token,
		ExplorerURL: chain.GetExplorerURL(txHash),
		Message:     "Tokens sent successfully",
	}

	h.logger.Info("Tokens sent successfully",
		zap.String("tx_hash", txHash),
		zap.String("recipient", req.Address),
		zap.String("token", req.Token),
	)

	return c.JSON(response)
}

// GetStatus returns the status of an address
func (h *Handler) GetStatus(c *fiber.Ctx) error {
	ctx := context.Background()

	address := c.Params("address")
	network := c.Query("network", h.defaultNetwork)

	// Get the chain for the specified network
	chain, _, err := h.getChain(network)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: err.Error(),
		})
	}

	// Validate address using chain-specific validation
	if err := chain.ValidateAddress(address); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Invalid address: %s", err.Error()),
		})
	}

	// Get IP from request
	ip := c.IP()

	// Get IP daily quota
	used, remaining, cooldownEnd, err := h.redis.GetIPDailyQuota(ctx, ip)
	if err != nil {
		h.logger.Error("Failed to get IP daily quota", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to check status",
		})
	}

	canRequest := remaining > 0 && cooldownEnd == nil

	response := models.StatusResponse{
		Address:    address,
		CanRequest: canRequest,
	}

	h.logger.Info("Status check",
		zap.String("address", address),
		zap.String("ip", ip),
		zap.Int("daily_quota_used", used),
		zap.Int("daily_quota_remaining", remaining),
	)

	return c.JSON(response)
}

// GetInfo returns information about the faucet
func (h *Handler) GetInfo(c *fiber.Ctx) error {
	ctx := context.Background()
	network := c.Query("network", h.defaultNetwork)

	// Get the chain for the specified network
	chain, chainProvider, err := h.getChain(network)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: err.Error(),
		})
	}

	// Get supported tokens for this chain
	supportedTokens := chain.GetSupportedTokens()

	// Get faucet balances for each supported token
	balances := make(map[string]string)
	for _, token := range supportedTokens {
		balance, err := chain.GetBalance(ctx, chainProvider.GetFaucetAddress(), token)
		if err != nil {
			h.logger.Error("Failed to get balance", zap.Error(err), zap.String("token", token))
			balances[token] = "0"
		} else {
			// Use different precision for different tokens
			if token == "ETH" {
				balances[token] = fmt.Sprintf("%.4f", chains.WeiToAmount(balance))
			} else {
				balances[token] = fmt.Sprintf("%.2f", chains.WeiToAmount(balance))
			}
		}
	}

	// Build balance info (for backward compatibility with existing API)
	balanceInfo := models.BalanceInfo{
		STRK: balances["STRK"],
		ETH:  balances["ETH"],
	}
	if balanceInfo.STRK == "" {
		balanceInfo.STRK = "0"
	}
	if balanceInfo.ETH == "" {
		balanceInfo.ETH = "0"
	}

	// Get available networks
	availableNetworks := make([]string, 0, len(h.chains))
	for name := range h.chains {
		availableNetworks = append(availableNetworks, name)
	}

	response := models.InfoResponse{
		Network: chain.GetNetworkName(),
		Limits: models.LimitInfo{
			StrkPerRequest:     chainProvider.GetDripAmount("STRK"),
			EthPerRequest:      chainProvider.GetDripAmount("ETH"),
			DailyRequestsPerIP: h.config.MaxRequestsPerDayIP(),
			TokenThrottleHours: 1, // 1 hour throttle per token
		},
		PoW: models.PoWInfo{
			Enabled:    true,
			Difficulty: h.config.PoWDifficulty(),
		},
		FaucetBalance:     balanceInfo,
		AvailableNetworks: availableNetworks,
	}

	return c.JSON(response)
}

// handleBothTokensRequest handles requests for both STRK and ETH tokens
func (h *Handler) handleBothTokensRequest(c *fiber.Ctx, ctx context.Context, req models.FaucetRequest, ip string, chain chains.Chain, chainProvider ChainProvider) error {
	// Process all supported tokens for this chain
	tokens := chain.GetSupportedTokens()
	var transactions []models.TransactionInfo
	var failedToken string

	for _, token := range tokens {
		// Determine amount using chain provider
		amountStr := chainProvider.GetDripAmount(token)
		amountFloat, _ := strconv.ParseFloat(amountStr, 64)
		maxHourly := chainProvider.GetMaxTokensPerHour(token)
		maxDaily := chainProvider.GetMaxTokensPerDay(token)

		// Check global distribution limits
		canDistribute, err := h.redis.TrackGlobalDistribution(ctx, token, amountFloat, maxHourly, maxDaily)
		if err != nil {
			h.logger.Error("Failed to check global distribution limits", zap.Error(err), zap.String("token", token))
			failedToken = token
			break
		}
		if !canDistribute {
			h.logger.Warn("Global distribution limit reached", zap.String("token", token), zap.String("ip", ip))
			failedToken = token
			break
		}

		// Check minimum balance protection
		currentBalance, err := chain.GetBalance(ctx, chainProvider.GetFaucetAddress(), token)
		if err != nil {
			h.logger.Error("Failed to check faucet balance", zap.Error(err), zap.String("token", token))
			failedToken = token
			break
		}

		amountWei := chains.AmountToWei(amountFloat)
		minBalancePct := float64(chainProvider.GetMinBalanceProtectPct()) / 100.0
		currentBalanceFloat := chains.WeiToAmount(currentBalance)
		minBalanceRequired := currentBalanceFloat * minBalancePct
		balanceAfterTransfer := currentBalanceFloat - amountFloat

		if balanceAfterTransfer < minBalanceRequired {
			h.logger.Warn("Balance protection triggered", zap.String("token", token), zap.Float64("current_balance", currentBalanceFloat))
			failedToken = token
			break
		}

		// Transfer tokens
		h.logger.Info("Transferring tokens", zap.String("recipient", req.Address), zap.String("token", token), zap.String("amount", amountStr))

		txHash, err := chain.TransferTokens(ctx, req.Address, token, amountWei)
		if err != nil {
			h.logger.Error("Failed to transfer tokens", zap.Error(err), zap.String("token", token))
			failedToken = token
			break
		}

		// Add to transactions list
		transactions = append(transactions, models.TransactionInfo{
			Token:       token,
			Amount:      amountStr,
			TxHash:      txHash,
			ExplorerURL: chain.GetExplorerURL(txHash),
		})

		h.logger.Info("Tokens sent successfully", zap.String("tx_hash", txHash), zap.String("token", token))
	}

	// If any token failed and we have partial success, still return success with what worked
	if len(transactions) > 0 {
		// Increment IP daily counter by 2 (BOTH = 1 STRK + 1 ETH)
		if err := h.redis.IncrementIPDailyLimit(ctx, ip, 2); err != nil {
			h.logger.Error("Failed to increment IP daily limit", zap.Error(err))
		}

		// Set hourly throttle for both tokens on this network
		network := req.Network
		if network == "" {
			network = h.defaultNetwork
		}
		for _, tx := range transactions {
			if err := h.redis.SetTokenHourlyThrottle(ctx, ip, network, tx.Token); err != nil {
				h.logger.Error("Failed to set token throttle", zap.Error(err), zap.String("token", tx.Token))
			}
		}

		message := "Both tokens sent successfully"
		if failedToken != "" {
			message = fmt.Sprintf("Sent %d token(s) successfully, but %s failed", len(transactions), failedToken)
		}

		return c.JSON(models.FaucetResponse{
			Success:      true,
			Transactions: transactions,
			Message:      message,
		})
	}

	// If no transactions succeeded, return error
	return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
		Error: fmt.Sprintf("Failed to send %s tokens. Please try again later.", failedToken),
	})
}

// GetQuota returns the current rate limit quota for the requesting IP
func (h *Handler) GetQuota(c *fiber.Ctx) error {
	ctx := context.Background()
	ip := c.IP()

	// Get IP daily quota (global across all networks)
	used, remaining, cooldownEnd, err := h.redis.GetIPDailyQuota(ctx, ip)
	if err != nil {
		h.logger.Error("Failed to get IP daily quota", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to get quota",
		})
	}

	// Build per-network throttle status
	networkThrottles := make(map[string]map[string]interface{})
	for networkName, chain := range h.chains {
		tokens := chain.GetSupportedTokens()
		tokenThrottles := make(map[string]interface{})

		for _, token := range tokens {
			available, nextTime, err := h.redis.CheckTokenHourlyThrottle(ctx, ip, networkName, token)
			if err != nil {
				h.logger.Error("Failed to check token throttle", zap.Error(err), zap.String("network", networkName), zap.String("token", token))
				continue
			}
			tokenThrottles[strings.ToLower(token)] = map[string]interface{}{
				"available":       available,
				"next_request_at": nextTime,
			}
		}
		networkThrottles[networkName] = tokenThrottles
	}

	response := map[string]interface{}{
		"daily_limit": map[string]interface{}{
			"total":        h.config.MaxRequestsPerDayIP(),
			"used":         used,
			"remaining":    remaining,
			"cooldown_end": cooldownEnd,
			"in_cooldown":  cooldownEnd != nil,
		},
		"hourly_throttle_by_network": networkThrottles,
	}

	return c.JSON(response)
}

// Health returns the health status of the API
func (h *Handler) Health(c *fiber.Ctx) error {
	ctx := context.Background()

	// Check Redis
	if err := h.redis.Ping(ctx); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ErrorResponse{
			Error: "Redis unavailable",
		})
	}

	return c.JSON(models.HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().Unix(),
	})
}
