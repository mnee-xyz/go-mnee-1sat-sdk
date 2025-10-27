# Golang MNEE-1Sat SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/mnee-xyz/go-mnee-1sat-sdk.svg)](https://pkg.go.dev/github.com/mnee-xyz/go-mnee-1sat-sdk)
[![Tests](https://github.com/mnee-xyz/go-mnee-1sat-sdk/actions/workflows/go-test.yml/badge.svg)](https://github.com/mnee-xyz/go-mnee-1sat-sdk/actions/workflows/go-test.yml)

The Golang MNEE SDK provides a robust and efficient way to interact with the MNEE USD token on the BSV blockchain. It leverages the [`bsv-blockchain/go-sdk`](https://github.com/bsv-blockchain/go-sdk) and offers features including balance checking, UTXO management, transaction signing, transfers, and validation.

## Documentation

📚 **Full documentation is available at [https://docs.mnee.io](https://docs.mnee.io)**

See the [examples](./examples) directory for runnable code snippets.

## Installation

```bash
go get github.com/mnee-xyz/go-mnee-1sat-sdk
```

## Requirements

- Go 1.24 or newer (matching the module's Go toolchain target in `go.mod`).
- An MNEE API token (obtainable from [MNEE](https://developer.mnee.net)).
- (For transfers) A Wallet Import Format (WIF) private key controlling MNEE UTXOs.

## Quick Start

### Basic Setup

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	mnee "github.com/mnee-xyz/go-mnee-1sat-sdk" 
)

func main() {
	// Read API Key from environment variable
	apiKey := os.Getenv("MNEE_API_KEY")
	if apiKey == "" {
		log.Fatal("MNEE_API_KEY environment variable not set")
	}

	// Initialize the SDK for the desired environment
	// Use mnee.EnvSandbox for testing, mnee.EnvMain for production
	m, err := mnee.NewMneeInstance(mnee.EnvSandbox, apiKey)
	if err != nil {
		log.Fatalf("Error creating MNEE instance: %v", err)
	}

	fmt.Println("MNEE SDK Initialized!")
	// Now you can use 'm' to call SDK functions
}

```
*(See `examples/config_example.go`)*

### Check Balance

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	mnee "github.com/mnee-xyz/go-mnee-1sat-sdk" 
)

func main() {
	// Read API Key from environment variable
	apiKey := os.Getenv("MNEE_API_KEY")
	if apiKey == "" { log.Fatal("MNEE_API_KEY needed") }

	m, err := mnee.NewMneeInstance(mnee.EnvSandbox, apiKey)
	if err != nil { log.Fatalf("Init failed: %v", err) }

	testAddress := os.Getenv("MNEE_TEST_ADDRESS")
	if testAddress == "" { log.Fatal("MNEE_TEST_ADDRESS needed") }

	// Get balances for multiple addresses
	addresses := []string{testAddress, "1111111111111111111114oLvT2"} // A known empty address
	balances, err := m.GetBalances(context.Background(), addresses)
	if err != nil {
		log.Fatalf("Error getting balances: %v", err)
	}

	fmt.Printf("Fetched balances:\n")
	for _, bal := range balances {
		fmt.Printf(" - %s: %.8f MNEE\n", *bal.Address, bal.Precised)
	}
}
```
*(See `examples/balance_example.go`)*

### Transfer MNEE (Asynchronous with Polling)

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	mnee "github.com/mnee-xyz/go-mnee-1sat-sdk"
)

func main() {
	apiKey := os.Getenv("MNEE_API_KEY")
	wif := os.Getenv("MNEE_WIF") // WIF for the sender address
	recipient := os.Getenv("MNEE_RECIPIENT_ADDRESS")
	if apiKey == "" || wif == "" || recipient == "" {
		log.Fatal("Need MNEE_API_KEY, MNEE_WIF, MNEE_RECIPIENT_ADDRESS")
	}

	m, err := mnee.NewMneeInstance(mnee.EnvSandbox, apiKey)
	if err != nil { log.Fatalf("Init failed: %v", err) }

	transferDTOs := []mnee.TransferMneeDTO{
		{Amount: 1000, Address: recipient}, // Amount in atomic units
	}
	wifs := []string{wif}

	fmt.Println("Submitting asynchronous transfer...")
	ticketID, err := m.AsynchronousTransfer(context.Background(), wifs, transferDTOs, false, nil, nil, nil)
	if err != nil {
		log.Fatalf("Transfer submission failed: %v", err)
	}
	fmt.Printf("Transfer submitted! Ticket ID: %s\n", *ticketID)

	fmt.Println("Polling ticket status...")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	ticket, err := m.PollTicket(ctx, *ticketID, 3*time.Second) // Poll every 3 seconds

	if err != nil {
		log.Fatalf("Polling failed: %v", err)
	}

	fmt.Printf("Polling complete! Final Status: %s\n", ticket.Status)
	if ticket.Status == mnee.SUCCESS {
		fmt.Printf("Transaction ID: %s\n", *ticket.TxID)
	}
}
```
*(See `examples/transfer_example.go`)*

## Key Features

- **Get Configuration:** Fetch current MNEE system parameters like `tokenId`, `approver` key, and `fee` structure.
- **Balance Checks:** Query MNEE balances for one or multiple addresses.
- **UTXO Management:** Retrieve Unspent Transaction Outputs (UTXOs) needed for transfers. Get all UTXOs for addresses or fetch a specific UTXO by its outpoint.
- **Transfers:**
    - `SynchronousTransfer`: Builds, signs, submits the transaction, and waits for the cosigner's response with the final transaction hex and ID. Use for immediate confirmation needs.
    - `AsynchronousTransfer`: Builds, signs, submits the transaction, and immediately returns a `ticketID`. Use for non-blocking operations or when combined with webhooks.
    - `PollTicket`: Checks the status of an asynchronous transfer using its `ticketID` until it succeeds or fails.
    - `withTxos` Option: Both transfer functions allow providing a pre-fetched list of UTXOs for optimization.
- **Transaction History:** Fetch historical MNEE transactions for specific addresses with pagination (`from`, `limit`).
- **Script Validation:** `IsMneeScript` function to check if a given ASM script is a valid MNEE token script according to the current configuration.
- **Partial Signing:** `PartialSign` function builds and signs the transaction inputs you provide WIFs for, returning the partially signed transaction hex. Useful for multi-signature or offline signing workflows.

## Support

- 📖 Documentation: [https://docs.mnee.io](https://docs.mnee.io)
- 🐛 Issues: Please open an issue on the repository.

## Contributing

Contributions are welcome! Please submit a pull request or open an issue to suggest improvements or report bugs.

## License

This project is licensed under the MIT License.