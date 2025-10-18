# Golang MNEE-1Sat SDK
This is official Golang mnee sdk

An  Go client for interacting with [MNEE](https://mnee.io) 1Sat
stablecoin. The SDK helps you compose [BSV-20](https://bsv20.com)
inscriptions, retrieve spendable UTXOs, and submit signed transfer
transactions to an MNEE-1sat stablecoin.

## Features

- Small, dependency-light wrapper around the
  [`github.com/bsv-blockchain/go-sdk`](https://github.com/bsv-blockchain/go-sdk)
  primitives used by MNEE.
- Convenience client for fetching MNEE tracked UTXOs.
- High level transfer helper that assembles BSV-20 inscriptions, enforces fee
  table rules, signs inputs, and sends the raw transaction to the cosigner.
- Ready to embed in CLI tools, backend services, and automation.

## Requirements

- Go 1.24 or newer (matching the module's Go toolchain target in `go.mod`).
- An MNEE API token and cosigner base URL.
- A Wallet Import Format (WIF) private key that controls the inputs you intend
  to spend.

## Installation

```bash
go get github.com/go-mnee-1sat-sdk
```