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

	// Use EnvSandbox for examples
	m, err := mnee.NewMneeInstance(mnee.EnvSandbox, apiKey)
	if err != nil {
		log.Fatalf("Error creating MNEE instance: %v", err)
	}

	fmt.Println("Attempting to partially sign a transfer...")
	transferDTOs := []mnee.TransferMneeDTO{
		{Amount: 1000, Address: recipientAddress}, // Send 1000 atomic units
	}
	wifs := []string{wif}

	// Call PartialSign
	// This builds the transaction, selects UTXOs, and signs with *our* WIF only.
	partialHex, err := m.PartialSign(context.Background(), wifs, transferDTOs, false, nil)
	if err != nil {
		log.Fatalf("PartialSign Error: %v", err)
	}

	fmt.Printf("✅ Successfully created partially-signed transaction:\n")
	fmt.Printf("\n%s\n", *partialHex)
}
