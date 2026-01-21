package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/Giri-Aayush/starknet-faucet/chains"
	ethereum "github.com/Giri-Aayush/starknet-faucet/chains/ethereum-sepolia"
	starknet "github.com/Giri-Aayush/starknet-faucet/chains/starknet-sepolia"
	"github.com/Giri-Aayush/starknet-faucet/internal/api"
	"github.com/Giri-Aayush/starknet-faucet/internal/cache"
	"github.com/Giri-Aayush/starknet-faucet/internal/config"
	"github.com/Giri-Aayush/starknet-faucet/internal/pow"
	"github.com/Giri-Aayush/starknet-faucet/pkg/utils"
	"go.uber.org/zap"
)

func main() {
	// Load common configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger, err := utils.NewLogger(cfg.LogLevel())
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting Multi-Chain Faucet Server",
		zap.String("port", cfg.Port()),
	)

	// Initialize all available chains
	chainRegistry := make(map[string]chains.Chain)
	providerRegistry := make(map[string]api.ChainProvider)

	// Try to load Starknet
	starknetCfg, err := starknet.LoadConfig()
	if err != nil {
		logger.Warn("Starknet config not available", zap.Error(err))
	} else {
		logger.Info("Initializing Starknet client...")
		starknetClient, err := starknet.NewClient(starknetCfg)
		if err != nil {
			logger.Warn("Failed to create Starknet client", zap.Error(err))
		} else {
			chainRegistry["starknet"] = starknetClient
			providerRegistry["starknet"] = starknetCfg
			logger.Info("Starknet client initialized",
				zap.String("network", starknetCfg.Network),
				zap.String("faucet_address", starknetCfg.FaucetAddress),
			)
		}
	}

	// Try to load Ethereum
	ethereumCfg, err := ethereum.LoadConfig()
	if err != nil {
		logger.Warn("Ethereum config not available", zap.Error(err))
	} else {
		logger.Info("Initializing Ethereum client...")
		ethereumClient, err := ethereum.NewClient(ethereumCfg)
		if err != nil {
			logger.Warn("Failed to create Ethereum client", zap.Error(err))
		} else {
			chainRegistry["ethereum"] = ethereumClient
			providerRegistry["ethereum"] = ethereumCfg
			logger.Info("Ethereum client initialized",
				zap.String("network", ethereumCfg.Network),
				zap.Int64("chain_id", ethereumCfg.ChainID),
				zap.String("faucet_address", ethereumCfg.FaucetAddress),
			)
		}
	}

	// Ensure at least one chain is available
	if len(chainRegistry) == 0 {
		logger.Fatal("No chains configured. Please configure at least one chain in .env")
	}

	logger.Info("Chains loaded", zap.Int("count", len(chainRegistry)))

	// Initialize Redis
	logger.Info("Connecting to Redis...")
	redis, err := cache.NewRedisClient(
		cfg.RedisURL,
		cfg.MaxRequestsPerDayIP(),
		cfg.MaxChallengesPerHour(),
	)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redis.Close()
	logger.Info("Connected to Redis",
		zap.Int("max_requests_per_day_ip", cfg.MaxRequestsPerDayIP()),
		zap.Int("max_challenges_per_hour", cfg.MaxChallengesPerHour()),
	)

	// Initialize PoW generator
	powGenerator := pow.NewGenerator(cfg.PoWDifficulty(), cfg.ChallengeTTL())
	logger.Info("PoW generator initialized",
		zap.Int("difficulty", cfg.PoWDifficulty()),
	)

	// Create API handler with chain registries
	handler := api.NewMultiChainHandler(cfg, logger, redis, chainRegistry, providerRegistry, powGenerator)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:               "Multi-Chain Faucet API",
		DisableStartupMessage: false,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Setup routes
	api.SetupRoutes(app, handler)

	// Start server in goroutine
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port())
		logger.Info("Server starting", zap.String("addr", addr))
		if err := app.Listen(addr); err != nil {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		logger.Error("Server shutdown error", zap.Error(err))
	}

	logger.Info("Server stopped")
}
