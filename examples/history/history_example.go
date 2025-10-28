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
	testAddress := os.Getenv("MNEE_TEST_ADDRESS")
	if apiKey == "" || testAddress == "" {
		log.Fatal("MNEE_API_KEY and MNEE_TEST_ADDRESS environment variables must be set")
	}

	m, err := mnee.NewMneeInstance(mnee.EnvSandbox, apiKey)
	if err != nil {
		log.Fatalf("Error creating MNEE instance: %v", err)
	}

	fmt.Printf("Fetching transaction history for address: %s...\n", testAddress)
	// Example: Get the 10 most recent transactions (limit=10) starting from the beginning (from=0)
	history, err := m.GetSpecificTransactionHistory(context.Background(), []string{testAddress}, 0, 10)
	if err != nil {
		log.Fatalf("Error getting transaction history: %v", err)
	}

	fmt.Printf("Found %d history items:\n", len(history))
	for i, item := range history {
		fmt.Printf(" - Item %d:\n", i+1)
		fmt.Printf("   Txid: %s\n", *item.Txid)
		fmt.Printf("   Height: %d\n", item.Height)
		fmt.Printf("   Senders: %v\n", item.Senders)
		fmt.Printf("   Receivers: %v\n", item.Receivers)
	}
}
