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

The package follows semantic versioning. Pin a tag in your own module when you
depend on a specific release.

## Getting Started

### Instantiate a client

```go
package main

import (
        "log"

        mnee "github.com/go-mnee-1sat-sdk"
)

func main() {
        client := mnee.NewClient(
                "https://cosigner.mnee.io", // Cosigner base URL
                "your-api-token",           // Auth token
                "1ExampleAddress...",       // Default sender address
        ).WithWIF("L1ExampleWIFKey...")

        // Use the client…
        _ = client
        log.Println("client initialised")
}
```

- `NewClient` normalises the cosigner URL, stores your API token, and remembers
  the sender address used for UTXO discovery.
- `WithWIF` is optional. You can either set a default signing key once or pass
  a WIF explicitly to the `Transfer` method call (passing an empty string uses
  the default).

### Fetch tracked UTXOs

Call `GetUTXOS` to retrieve 1Sat outputs that belong to the sender address.

```go
txos, err := client.GetUTXOS()
if err != nil {
        log.Fatalf("fetching UTXOs failed: %v", err)
}

for _, txo := range txos {
        if txo.Data != nil && txo.Data.Bsv21 != nil {
                log.Printf("%s -> %d tokens", *txo.Txid, txo.Data.Bsv21.Amt)
        }
}
```

Each entry is returned as an `MneeTxo` struct with decoded BSV-20 metadata:

```go
type MneeTxo struct {
        Satoshis uint16
        Height   uint64
        Idx      uint64
        Score    uint64
        Vout     uint64
        Script   *string
        Txid     *string
        Data     *Data // includes BSV-20 and cosigner details
        Owners   pq.StringArray
        Senders  pq.StringArray
}
```

Use the BSV-20 amount (`txo.Data.Bsv21.Amt`) to determine how many tokens can
be spent from each output.

### Create a transfer

```go
recipients := []mnee.TransferMneeDTO{
        {Address: "1RecipientAddress…", Amount: 1_000},
        {Address: "1AnotherAddress…", Amount: 500},
}

if err := client.Transfer("", recipients); err != nil { // empty string uses the default WIF
        log.Fatalf("transfer failed: %v", err)
}
```

Key behaviours to be aware of:

- The method builds BSV-20 inscriptions for each recipient and ensures the
  total available token balance is sufficient. If `Amount` is zero for any
  recipient an error is returned immediately.
- Fees are derived from the cosigner's `/v1/config` endpoint. When the change
  after applying the fee is positive, a change inscription addressed back to
  the sender is automatically appended to the transaction.
- Inputs are unlocked with `SIGHASH_ALL|SIGHASH_ANYONECANPAY|FORKID` as
  required by MNEE. If a raw transaction submission fails with an HTTP status
  code ≥ 300, the error response from the cosigner is returned.

> **Note**
> By default the SDK uses the sender address supplied during construction when
> querying UTXOs. If you need to spend from multiple addresses you can create
> additional clients or swap the `senderAddress` field by composing your own
> lightweight wrapper type.

### Transfer DTO helper

`TransferMneeDTO` is the payload consumed by `Transfer`:

```go
type TransferMneeDTO struct {
        Amount  uint64
        Address string
}
```

Amounts are denominated in the smallest divisible unit of the token (as
defined by the cosigner's configuration). Ensure you convert from human
readable amounts to integer units before calling `Transfer`.

## Error handling

All exported methods return Go errors. A few important ones to watch for:

- `transfer amount must be greater than 0` – validation guard for DTO entries.
- `insufficient mnee balance` – triggered when the summed UTXO balance cannot
  cover the requested transfer amount.
- `missing wif signing key` – returned when neither an explicit nor default WIF
  is provided for signing.
- `non-200 response code from mnee cosigner` or the cosigner's explicit error
  message when a remote call fails.

Wrap these errors in your own domain types or surface them to end users as
appropriate.

## Development

- Run `gofmt` before committing changes.
- Integration tests typically require access to a live cosigner service. Stub
  HTTP responses in unit tests when possible.
