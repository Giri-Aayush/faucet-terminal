<div align="center">

# faucet-terminal

**Get testnet tokens from your terminal**

[![npm](https://img.shields.io/npm/v/faucet-terminal?color=blue)](https://www.npmjs.com/package/faucet-terminal)
[![Downloads](https://img.shields.io/npm/dm/faucet-terminal)](https://www.npmjs.com/package/faucet-terminal)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

Starknet Sepolia · Ethereum Sepolia

</div>

---

## Installation

```bash
npm install -g faucet-terminal
```

## Quick Start

```bash
# Ethereum Sepolia (ETH)
faucet-terminal request 0xYOUR_ADDRESS --network ethereum
faucet-terminal req 0xYOUR_ADDRESS -n eth              # short version

# Starknet Sepolia (STRK - default)
faucet-terminal request 0xYOUR_ADDRESS --network starknet
faucet-terminal req 0xYOUR_ADDRESS -n sn               # short version

# Starknet Sepolia (ETH)
faucet-terminal request 0xYOUR_ADDRESS --network starknet --token ETH
faucet-terminal req 0xYOUR_ADDRESS -n sn --token ETH   # short version
```

## Example Output

```
$ faucet-terminal req 0x742d35Cc6634C0532925a3b844Bc9e7595f8fE21 -n eth

  faucet terminal

  network  ethereum

  [Verification] What is 5 + 7? 12
  Correct!

  ✔ Challenge received
  ✔ Challenge solved in 32.1s
  ✔ Transaction submitted!

  ────────────────────────────────────────────────────────

    amount  0.01 ETH
    tx      0x6139cd4b...82e8d55f

    https://sepolia.etherscan.io/tx/0x6139cd4b...

  ────────────────────────────────────────────────────────

  ✔ Tokens will arrive in ~30 seconds
```

## Supported Networks

| Network | Full Name | Aliases | Tokens |
|:--------|:----------|:--------|:-------|
| Starknet Sepolia | `starknet` | `sn`, `sn-sep` | STRK (default), ETH |
| Ethereum Sepolia | `ethereum` | `eth`, `eth-sep` | ETH |

## Commands

### Request Tokens

```bash
# Full command
faucet-terminal request <ADDRESS> --network <NETWORK> [--token <TOKEN>]

# Short version
faucet-terminal req <ADDRESS> -n <NETWORK> [--token <TOKEN>]

# Shortest version
faucet-terminal r <ADDRESS> -n <NETWORK>
```

**Examples:**
```bash
faucet-terminal request 0x123...abc --network ethereum           # ETH on Ethereum
faucet-terminal req 0x123...abc -n eth                           # same, shorter
faucet-terminal r 0x123...abc -n eth                             # same, shortest

faucet-terminal request 0x123...abc --network starknet           # STRK on Starknet
faucet-terminal req 0x123...abc -n sn                            # same, shorter

faucet-terminal request 0x123...abc --network starknet --token ETH  # ETH on Starknet
faucet-terminal req 0x123...abc -n sn --token ETH                   # same, shorter
```

### Check Status

```bash
# Full command
faucet-terminal status <ADDRESS> --network <NETWORK>

# Short version
faucet-terminal s <ADDRESS> -n <NETWORK>
```

**Examples:**
```bash
faucet-terminal status 0x123...abc --network ethereum
faucet-terminal s 0x123...abc -n eth
```

### View Faucet Info

```bash
# Full command
faucet-terminal info --network <NETWORK>

# Short version
faucet-terminal i -n <NETWORK>
```

**Examples:**
```bash
faucet-terminal info --network ethereum
faucet-terminal i -n eth
```

### Check Your Quota

```bash
# Full command
faucet-terminal quota

# Short version
faucet-terminal q
```

### View Rate Limits

```bash
# Full command
faucet-terminal limits

# Short version
faucet-terminal l
```

## Options

| Option | Short | Description |
|:-------|:------|:------------|
| `--network` | `-n` | Network to use (required for most commands) |
| `--token` | | Token to request: `ETH`, `STRK` |
| `--json` | | Output in JSON format |
| `--help` | `-h` | Show help |

## Rate Limits

```
5 requests/day     per IP address
1 request/hour     per token type (ETH and STRK tracked separately)
24h cooldown       after daily limit reached
```

## How It Works

```
1. Submit address    →  Validate format for the network
2. Verification      →  Answer a simple math question
3. Proof of Work     →  Solve SHA-256 challenge (~30-60s)
4. Receive tokens    →  Transaction submitted to blockchain
```

## Error Messages

The CLI provides clear error messages when you hit rate limits:

- `[HOURLY LIMIT]` - You requested this token within the last hour
- `[DAILY LIMIT]` - You've used all 5 daily requests
- `[FAUCET LIMIT]` - Faucet has temporarily reached its distribution limit
- `[LOW BALANCE]` - Faucet balance is too low

## License

MIT

---

<div align="center">

[npm](https://www.npmjs.com/package/faucet-terminal) · [GitHub](https://github.com/Giri-Aayush/faucet-terminal) · [Issues](https://github.com/Giri-Aayush/faucet-terminal/issues)

Made with ❤️ by a developer, for developers

</div>
