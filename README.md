<div align="center">

# faucet-terminal

**Get testnet tokens from your terminal**

[![npm](https://img.shields.io/npm/v/faucet-terminal?color=blue)](https://www.npmjs.com/package/faucet-terminal)
[![Downloads](https://img.shields.io/npm/dm/faucet-terminal)](https://www.npmjs.com/package/faucet-terminal)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

Starknet Sepolia · Ethereum Sepolia

</div>

---

```bash
npm install -g faucet-terminal
```

```
$ faucet-terminal req 0x742d35Cc6634C0532925a3b844Bc9e7595f8fE21 -n eth-sep

  faucet terminal

  network  ethereum

  ✔ Challenge received
  ✔ Challenge solved in 12.3s
  ✔ Transaction submitted!

  ────────────────────────────────────────────────────────

    amount  0.01 ETH
    tx      0x6139cd4b...82e8d55f

    https://sepolia.etherscan.io/tx/0x6139cd4b...

  ────────────────────────────────────────────────────────

  ✔ Tokens will arrive in ~30 seconds
```

## Quick Start

```bash
# Ethereum Sepolia
faucet-terminal req 0xYOUR_ADDRESS -n eth-sep

# Starknet Sepolia (STRK)
faucet-terminal req 0xYOUR_ADDRESS -n sn-sep

# Starknet Sepolia (ETH)
faucet-terminal req 0xYOUR_ADDRESS -n sn-sep --token ETH

```

## Networks

| Network | Aliases | Tokens |
|:--------|:--------|:-------|
| Starknet Sepolia | `starknet` `sn` `sn-sep` | STRK, ETH |
| Ethereum Sepolia | `ethereum` `eth` `eth-sep` | ETH |

## Commands

```
req, r      Request testnet tokens
status, s   Check cooldown status
info, i     View faucet info
quota, q    Check remaining quota
limits, l   Show rate limits
```

## Options

```
-n, --network   Network (required): eth-sep, sn-sep
--token         Token type: ETH, STRK
--json          JSON output
```

## Rate Limits

```
5 requests/day    per IP address
1 request/hour    per token type
24h cooldown      after daily limit
```

## How It Works

```
1. Submit address    →  Validate format
2. Verification      →  Answer simple question
3. Proof of Work     →  Solve SHA-256 challenge (~30s)
4. Receive tokens    →  Transaction submitted
```

## License

MIT

---

<div align="center">

[npm](https://www.npmjs.com/package/faucet-terminal) · [GitHub](https://github.com/Giri-Aayush/faucet-terminal) · [Issues](https://github.com/Giri-Aayush/faucet-terminal/issues)

Made with ❤️ by a developer, for developers

</div>
