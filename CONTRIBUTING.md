# Contributing to faucet-cli

Thank you for your interest in contributing to faucet-cli! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Adding a New Chain](#adding-a-new-chain)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)

## Code of Conduct

Please be respectful and constructive in all interactions. We welcome contributors of all backgrounds and experience levels.

## Getting Started

1. Fork the repository
2. Clone your fork locally
3. Create a new branch for your feature/fix
4. Make your changes
5. Submit a pull request

## Development Setup

### Prerequisites

- Go 1.21 or later
- Node.js 14 or later (for npm package testing)
- Redis (for local backend testing)
- Docker (optional, for containerized development)

### Local Development

```bash
# Clone the repository
git clone https://github.com/Giri-Aayush/faucet-cli.git
cd faucet-cli

# Install Go dependencies
go mod download

# Build the CLI
go build -o faucet ./cmd/cli

# Build the server
go build -o server ./cmd/server

# Run tests
go test ./...
```

### Environment Variables

Copy `.env.example` to `.env` and fill in the required values:

```bash
cp .env.example .env
```

Required variables:
- `REDIS_URL` - Redis connection string
- `STARKNET_RPC_URL` - Alchemy Starknet RPC endpoint
- `STARKNET_PRIVATE_KEY` - Faucet wallet private key
- `STARKNET_ADDRESS` - Faucet wallet address
- `ETHEREUM_RPC_URL` - Alchemy Ethereum RPC endpoint
- `ETHEREUM_PRIVATE_KEY` - Faucet wallet private key
- `ETHEREUM_ADDRESS` - Faucet wallet address

## Project Structure

```
faucet-cli/
├── chains/                 # Chain implementations (one folder per network)
│   ├── chain.go           # Chain interface definition
│   ├── starknet-sepolia/  # Starknet Sepolia implementation
│   └── ethereum-sepolia/  # Ethereum Sepolia implementation
├── cmd/
│   ├── cli/               # CLI entry point
│   └── server/            # Backend API entry point
├── internal/              # Server-side internal packages
│   ├── api/               # HTTP handlers and routes
│   ├── cache/             # Redis rate limiting
│   ├── config/            # Configuration loading
│   ├── models/            # Data models
│   └── pow/               # Proof of Work verification
├── pkg/                   # Shared packages
│   ├── cli/               # CLI client code
│   └── utils/             # Shared utilities
└── deployments/           # Deployment configurations
```

## Adding a New Chain

To add support for a new blockchain network:

1. Create a new folder under `chains/` with the network name (e.g., `arbitrum-sepolia/`)

2. Implement the required files:
   - `client.go` - Chain client with transaction logic
   - `config.go` - Configuration loading
   - `config.json` - Token amounts, limits, explorer URLs
   - `validator.go` - Address validation

3. Implement the `Chain` interface from `chains/chain.go`:
   ```go
   type Chain interface {
       TransferTokens(ctx context.Context, recipient string, token string) (*TransferResult, error)
       GetBalance(ctx context.Context, token string) (*big.Float, error)
       WaitForTransaction(ctx context.Context, txHash string) error
       ValidateAddress(address string) bool
       GetExplorerURL(txHash string) string
       GetConfig() *ChainConfig
   }
   ```

4. Register the chain in the server's chain registry

5. Update the CLI to recognize the new network

6. Add tests for the new chain implementation

## Pull Request Process

1. **Create a descriptive branch name**: `feature/add-arbitrum-support`, `fix/rate-limit-bug`

2. **Write clear commit messages**: Follow conventional commits format
   ```
   feat: add Arbitrum Sepolia support
   fix: correct rate limit calculation
   docs: update README with new chain
   ```

3. **Update documentation**: If your change affects usage, update the README

4. **Add tests**: New features should include tests

5. **Request review**: Tag maintainers for review

## Coding Standards

### Go Code

- Follow standard Go conventions (`go fmt`, `go vet`)
- Use descriptive variable and function names
- Add comments for exported functions
- Handle errors explicitly (no silent failures)
- Use context for cancellation and timeouts

### Testing

- Write unit tests for new functionality
- Use table-driven tests where appropriate
- Mock external dependencies (RPC, Redis)

### Security

- Never commit private keys or secrets
- Validate all user input
- Use parameterized queries/calls
- Follow OWASP guidelines

## Questions?

Open an issue or reach out to the maintainers. We're happy to help!
