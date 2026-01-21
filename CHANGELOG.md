# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Restructured chains folder to use network-based naming (e.g., `ethereum-sepolia` instead of `ethereum-11155111`)

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
| 1.0.18 | 2024-01 | Initial release with Starknet & Ethereum Sepolia |
