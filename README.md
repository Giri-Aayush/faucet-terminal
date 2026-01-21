# faucet-terminal

Multi-chain testnet faucet CLI. Get tokens on Starknet and Ethereum testnets from your terminal.

```bash
npm install -g faucet-terminal
```

## Usage

```bash
# Ethereum Sepolia
faucet-terminal req 0x742d35Cc6634C0532925a3b844Bc9e7595f8fE21 -n eth-sep

# Starknet Sepolia
faucet-terminal req 0x049d36570d4e46f48e99674bd3fcc84644ddd6b96f7c741b1562b82f9e004dc7 -n sn-sep

# Starknet with ETH token
faucet-terminal req 0x049d...dc7 -n sn-sep --token ETH

# Starknet with both STRK and ETH
faucet-terminal req 0x049d...dc7 -n sn-sep --both
```

## Networks

| Network | Aliases | Tokens |
|---------|---------|--------|
| Starknet Sepolia | `starknet`, `sn`, `sn-sep` | STRK, ETH |
| Ethereum Sepolia | `ethereum`, `eth`, `eth-sep` | ETH |

## Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `request <address>` | `req`, `r` | Request testnet tokens |
| `status <address>` | `s` | Check cooldown status |
| `info` | `i` | View faucet info |
| `quota` | `q` | Check remaining quota |
| `limits` | `l` | Show rate limit rules |

### Examples

```bash
# Request tokens
faucet-terminal req 0x123...abc -n eth-sep
faucet-terminal r 0x123...abc -n sn-sep --token ETH

# Check status
faucet-terminal s 0x123...abc -n eth-sep

# View faucet info
faucet-terminal i -n sn-sep

# Check your quota
faucet-terminal q -n eth-sep

# Show rate limits
faucet-terminal l -n sn-sep
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--network` | `-n` | Network to use (required) |
| `--token` | | Token type: `ETH`, `STRK` |
| `--both` | | Request both tokens (Starknet only) |
| `--json` | | JSON output |
| `--verbose` | `-v` | Verbose output |

## Rate Limits

- 5 requests per day per IP
- 1 request per token per hour
- 24h cooldown after daily limit

## How It Works

1. Submit wallet address
2. Answer verification question
3. Solve proof of work (~30s)
4. Receive tokens

## Development

```bash
git clone https://github.com/Giri-Aayush/faucet-cli.git
cd faucet-cli
go build -o faucet-terminal ./cmd/cli
./faucet-terminal --help
```

## License

MIT

---

[npm](https://www.npmjs.com/package/faucet-terminal) ·
[GitHub](https://github.com/Giri-Aayush/faucet-cli) ·
[Issues](https://github.com/Giri-Aayush/faucet-cli/issues)
