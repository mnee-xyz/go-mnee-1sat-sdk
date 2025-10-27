package main

import (
	"context"
	"fmt"
	"log"
	"os"

	mnee "github.com/mnee-xyz/go-mnee-1sat-sdk"
)

func main() {
	apiKey := os.Getenv("MNEE_API_KEY") // Use sandbox key for example
	if apiKey == "" {
		log.Fatal("MNEE_API_KEY environment variable not set")
	}

	// Use EnvSandbox for testing examples
	m, err := mnee.NewMneeInstance(mnee.EnvSandbox, apiKey)
	if err != nil {
		log.Fatalf("Error creating MNEE instance: %v", err)
	}

	fmt.Println("Fetching MNEE configuration...")
	config, err := m.GetConfig(context.Background())
	if err != nil {
		log.Fatalf("Error getting config: %v", err)
	}

	fmt.Printf("Successfully fetched config!\n")
	fmt.Printf("Token ID: %s\n", *config.TokenId)
	fmt.Printf("Decimals: %d\n", config.Decimals)
	fmt.Printf("Fee Address: %s\n", *config.FeeAddress)
	fmt.Printf("Approver PubKey: %s\n", *config.Approver)
	fmt.Printf("Number of Fee Tiers: %d\n", len(config.Fees))
}
