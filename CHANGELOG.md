# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2025-01-21

### Changed
- **BREAKING**: Renamed package from `faucet-cli` to `faucet-terminal`
- **BREAKING**: Binary renamed from `faucet` to `faucet-terminal`
- Restructured chains folder to use network-based naming (`ethereum-sepolia` instead of `ethereum-11155111`)
- Simplified `limits` command output with minimal UI
- Updated all help text for consistency

### Added
- Command aliases for faster usage:
  - `request` → `req`, `r`
  - `status` → `s`
  - `info` → `i`
  - `quota` → `q`
  - `limits` → `l`
- Network aliases:
  - `starknet` → `sn`, `sn-sep`
  - `ethereum` → `eth`, `eth-sep`
- Short flag `-n` for `--network`
- Test/production config separation with `FAUCET_TEST_MODE`

### Fixed
- Ethereum transactions now use EIP-1559 (Type 2) instead of legacy transactions
- All commands now properly validate network and use correct API URL

## [1.0.18] - 2024-01-XX

### Added
- Initial public release
- Starknet Sepolia support (STRK, ETH tokens)
- Ethereum Sepolia support (ETH token)
- Proof of Work anti-spam protection
- Rate limiting (5 requests/day per IP)
- Cross-platform binaries (Linux, macOS, Windows)

### Security
- SHA-256 based PoW challenge system
- Redis-backed rate limiting
- Address validation for all supported chains

---

## Version History

| Version | Date | Description |
|---------|------|-------------|
| 2.0.0 | 2025-01-21 | Rebrand to faucet-terminal, add command/network aliases |
| 1.0.18 | 2024-01 | Initial release with Starknet & Ethereum Sepolia |
