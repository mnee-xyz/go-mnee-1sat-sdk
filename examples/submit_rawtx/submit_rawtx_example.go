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
	wif := os.Getenv("MNEE_WIF") // WIF for the address holding funds
	recipientAddress := os.Getenv("MNEE_RECIPIENT_ADDRESS")
	if apiKey == "" || wif == "" || recipientAddress == "" {
		log.Fatal("MNEE_API_KEY, MNEE_WIF, and MNEE_RECIPIENT_ADDRESS env vars must be set")
	}

	m, err := mnee.NewMneeInstance(mnee.EnvSandbox, apiKey)
	if err != nil {
		log.Fatalf("Error creating MNEE instance: %v", err)
	}

	// --- 1. Create a partially signed transaction ---
	// We need a valid, partially-signed hex to submit.
	// We'll use PartialSign to create one.
	fmt.Println("Creating a partially-signed transaction hex...")
	transferDTOs := []mnee.TransferMneeDTO{
		{Amount: 1000, Address: recipientAddress},
	}
	wifs := []string{wif}

	partialHex, err := m.PartialSign(context.Background(), wifs, transferDTOs, false, nil)
	if err != nil {
		log.Fatalf("PartialSign Error: %v", err)
	}
	fmt.Printf("✅ Got partially-signed hex: %s...\n", (*partialHex)[:60])

	// --- 2. Submit the hex using SubmitRawTxSync ---
	fmt.Println("\nAttempting to submit raw tx synchronously...")
	syncResponse, err := m.SubmitRawTxSync(context.Background(), *partialHex)
	if err != nil {
		log.Fatalf("SubmitRawTxSync Error: %v", err)
	}
	fmt.Printf("✅ Synchronous Submission Success! Txid: %s\n", *syncResponse.Txid)

	// --- 3. Create another partial tx for the async example ---
	// We need a new tx, as the previous one's UTXOs are now spent
	fmt.Println("\nCreating a second partially-signed transaction hex...")
	transferDTOs2 := []mnee.TransferMneeDTO{
		{Amount: 1000, Address: recipientAddress},
	}
	partialHex2, err := m.PartialSign(context.Background(), wifs, transferDTOs2, false, nil)
	if err != nil {
		log.Fatalf("PartialSign Error (for async): %v", err)
	}
	fmt.Printf("✅ Got second partially-signed hex: %s...\n", (*partialHex2)[:60])

	// --- 4. Submit the second hex using SubmitRawTxAsync ---
	fmt.Println("\nAttempting to submit raw tx asynchronously...")
	ticketID, err := m.SubmitRawTxAsync(context.Background(), *partialHex2, nil, nil)
	if err != nil {
		log.Fatalf("SubmitRawTxAsync Error: %v", err)
	}
	fmt.Printf("✅ Asynchronous Submission Submitted! Ticket ID: %s\n", *ticketID)

	// --- 5. Poll the async ticket ---
	fmt.Printf("Polling ticket %s for status...\n", *ticketID)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	ticket, pollErr := m.PollTicket(ctx, *ticketID, 3*time.Second) // Poll every 3 seconds
	if pollErr != nil {
		log.Printf("Error polling ticket: %v", pollErr)
	} else {
		fmt.Printf("✅ Ticket Polled! Final Status: %s\n", ticket.Status)
		if ticket.Status == mnee.SUCCESS {
			fmt.Printf("   Transaction ID: %s\n", *ticket.TxID)
		}
	}
}
