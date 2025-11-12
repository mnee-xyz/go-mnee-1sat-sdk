package main

import (
	"context"
	"fmt"
	"log"
	"os"

	mnee "github.com/mnee-xyz/go-mnee-1sat-sdk"
)

func main() {
	apiKey := os.Getenv("MNEE_API_KEY")
	wif := os.Getenv("MNEE_WIF") // WIF for the address holding funds
	recipientAddress := os.Getenv("MNEE_RECIPIENT_ADDRESS")
	if apiKey == "" || wif == "" || recipientAddress == "" {
		log.Fatal("MNEE_API_KEY, MNEE_WIF, and MNEE_RECIPIENT_ADDRESS env vars must be set")
	}

	m, err := mnee.NewMneeInstance(mnee.EnvSandbox, apiKey)
	if err != nil {
		log.Fatalf("Error creating MNEE instance: %v", err)
	}

	// --- 1. Create a transaction to get a valid TxID ---
	fmt.Println("Submitting a synchronous transfer to get a valid TxID...")
	transferDTOs := []mnee.TransferMneeDTO{
		{Amount: 1000, Address: recipientAddress},
	}
	wifs := []string{wif}

	syncResponse, err := m.SynchronousTransfer(context.Background(), wifs, transferDTOs, false, nil)
	if err != nil {
		log.Fatalf("Synchronous Transfer Error: %v", err)
	}

	txidToGet := *syncResponse.Txid
	fmt.Printf("✅ Transfer Submitted! Txid: %s\n", txidToGet)

	// --- 2. Call GetMNEETxHex ---
	fmt.Printf("\nFetching raw hex for TxID: %s...\n", txidToGet)
	txHex, err := m.GetMNEETxHex(context.Background(), txidToGet)
	if err != nil {
		log.Fatalf("GetMNEETxHex Error: %v", err)
	}

	fmt.Printf("✅ Successfully fetched transaction hex!\n")
	fmt.Printf("\n%s\n", *txHex)
}
