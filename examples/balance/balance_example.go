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
	testAddress := os.Getenv("MNEE_TEST_ADDRESS") // Your sandbox address with some balance
	if apiKey == "" || testAddress == "" {
		log.Fatal("MNEE_API_KEY and MNEE_TEST_ADDRESS environment variables must be set")
	}

	m, err := mnee.NewMneeInstance(mnee.EnvSandbox, apiKey)
	if err != nil {
		log.Fatalf("Error creating MNEE instance: %v", err)
	}

	fmt.Println("Fetching balances...")
	// Example: Get balance for your test address and a known empty one
	addresses := []string{testAddress, "1111111111111111111114oLvT2"}
	balances, err := m.GetBalances(context.Background(), addresses)
	if err != nil {
		log.Fatalf("Error getting balances: %v", err)
	}

	fmt.Printf("Successfully fetched balances for %d addresses:\n", len(balances))
	for _, bal := range balances {
		fmt.Printf(" - Address: %s\n", *bal.Address)
		fmt.Printf("   Amount (Atomic): %.0f\n", bal.Amt) // Amt is float64 in DTO
		fmt.Printf("   Amount (MNEE): %.8f\n", bal.Precised)
	}
}
