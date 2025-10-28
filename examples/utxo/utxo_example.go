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

	// 1. Get Unspent UTXOs for an address
	fmt.Printf("Fetching UTXOs for address: %s...\n", testAddress)
	utxos, err := m.GetUnspentTxos(context.Background(), []string{testAddress})
	if err != nil {
		log.Fatalf("Error getting UTXOs: %v", err)
	}

	fmt.Printf("Found %d UTXOs:\n", len(utxos))
	if len(utxos) > 0 {
		firstUtxo := utxos[0]
		fmt.Printf(" - First UTXO Outpoint: %s\n", *firstUtxo.Outpoint)
		fmt.Printf("   Amount (Atomic): %d\n", firstUtxo.Data.Bsv21.Amt)
		fmt.Printf("   Owner: %s\n", firstUtxo.Owners[0])

		// 2. Get a single UTXO by its Outpoint
		outpointToGet := *firstUtxo.Outpoint
		fmt.Printf("\nFetching single UTXO by outpoint: %s...\n", outpointToGet)
		singleUtxo, err := m.GetTxo(context.Background(), outpointToGet)
		if err != nil {
			log.Fatalf("Error getting single UTXO: %v", err)
		}
		fmt.Printf("Successfully fetched single UTXO:\n")
		fmt.Printf(" - Txid: %s\n", *singleUtxo.Txid)
		fmt.Printf("   Vout: %d\n", singleUtxo.Vout)
		fmt.Printf("   Amount (Atomic): %d\n", singleUtxo.Data.Bsv21.Amt)
	} else {
		fmt.Println("No UTXOs found for this address.")
	}
}
