package ethereum

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Client implements the chains.Chain interface for Ethereum.
type Client struct {
	client     *ethclient.Client
	privateKey *ecdsa.PrivateKey
	address    common.Address
	config     *Config
}

// NewClient creates a new Ethereum chain client.
func NewClient(cfg *Config) (*Client, error) {
	// Connect to Ethereum node
	client, err := ethclient.Dial(cfg.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	// Parse private key
	privateKey, err := crypto.HexToECDSA(stripHexPrefix(cfg.FaucetPrivateKey))
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	// Derive address from private key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to get public key")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Verify derived address matches configured address
	configuredAddr := common.HexToAddress(cfg.FaucetAddress)
	if address != configuredAddr {
		return nil, fmt.Errorf("private key does not match configured address: got %s, expected %s",
			address.Hex(), configuredAddr.Hex())
	}

	return &Client{
		client:     client,
		privateKey: privateKey,
		address:    address,
		config:     cfg,
	}, nil
}

// TransferTokens transfers ETH to a recipient.
func (c *Client) TransferTokens(
	ctx context.Context,
	recipient string,
	token string,
	amount *big.Int,
) (string, error) {
	// Ethereum faucet only supports native ETH
	if token != "ETH" {
		return "", fmt.Errorf("unsupported token: %s (only ETH supported)", token)
	}

	toAddress := common.HexToAddress(recipient)

	// Get the nonce for the faucet account
	nonce, err := c.client.PendingNonceAt(ctx, c.address)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get suggested gas price
	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get gas price: %w", err)
	}

	// Standard gas limit for ETH transfer
	gasLimit := uint64(21000)

	// Create the transaction
	tx := types.NewTransaction(nonce, toAddress, amount, gasLimit, gasPrice, nil)

	// Sign the transaction
	chainID := big.NewInt(c.config.ChainID)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), c.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send the transaction
	err = c.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTx.Hash().Hex(), nil
}

// GetBalance returns the ETH balance of an address.
func (c *Client) GetBalance(ctx context.Context, address string, token string) (*big.Int, error) {
	// Ethereum faucet only supports native ETH
	if token != "ETH" {
		return nil, fmt.Errorf("unsupported token: %s (only ETH supported)", token)
	}

	addr := common.HexToAddress(address)
	balance, err := c.client.BalanceAt(ctx, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

// WaitForTransaction waits for a transaction to be mined.
func (c *Client) WaitForTransaction(ctx context.Context, txHash string) error {
	hash := common.HexToHash(txHash)

	// Poll for transaction receipt
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			receipt, err := c.client.TransactionReceipt(ctx, hash)
			if err != nil {
				// Transaction not yet mined, continue waiting
				continue
			}

			if receipt.Status == types.ReceiptStatusSuccessful {
				return nil
			}
			return fmt.Errorf("transaction failed with status %d", receipt.Status)
		}
	}
}

// ValidateAddress validates an Ethereum address format.
func (c *Client) ValidateAddress(address string) error {
	return ValidateAddress(address)
}

// NormalizeAddress normalizes an Ethereum address.
func (c *Client) NormalizeAddress(address string) string {
	return NormalizeAddress(address)
}

// GetSupportedTokens returns the list of tokens supported by Ethereum faucet.
func (c *Client) GetSupportedTokens() []string {
	return []string{"ETH"}
}

// ValidateToken checks if a token is supported.
func (c *Client) ValidateToken(token string) error {
	return ValidateToken(token)
}

// GetExplorerURL returns the block explorer URL for a transaction.
func (c *Client) GetExplorerURL(txHash string) string {
	if c.config.ExplorerURL != "" {
		return c.config.ExplorerURL + txHash
	}
	// Fallback to default
	switch c.config.Network {
	case "mainnet":
		return fmt.Sprintf("https://etherscan.io/tx/%s", txHash)
	case "sepolia":
		return fmt.Sprintf("https://sepolia.etherscan.io/tx/%s", txHash)
	case "goerli":
		return fmt.Sprintf("https://goerli.etherscan.io/tx/%s", txHash)
	case "holesky":
		return fmt.Sprintf("https://holesky.etherscan.io/tx/%s", txHash)
	default:
		return fmt.Sprintf("https://sepolia.etherscan.io/tx/%s", txHash)
	}
}

// GetChainName returns the chain name.
func (c *Client) GetChainName() string {
	return "ethereum"
}

// GetNetworkName returns the network name.
func (c *Client) GetNetworkName() string {
	return c.config.Network
}

// GetConfig returns the chain configuration.
func (c *Client) GetConfig() *Config {
	return c.config
}

// Close closes the Ethereum client connection.
func (c *Client) Close() {
	c.client.Close()
}

// Helper function to strip 0x prefix from hex string
func stripHexPrefix(s string) string {
	if len(s) >= 2 && s[0:2] == "0x" {
		return s[2:]
	}
	return s
}
