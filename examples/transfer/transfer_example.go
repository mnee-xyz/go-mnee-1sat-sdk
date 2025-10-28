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

	// --- Example 1: Synchronous Transfer ---
	fmt.Println("Attempting Synchronous Transfer...")
	transferDTOs := []mnee.TransferMneeDTO{
		{Amount: 1000, Address: recipientAddress}, // Send 1000 atomic units (0.00001 MNEE if 5 decimals)
	}
	wifs := []string{wif}

	syncResponse, err := m.SynchronousTransfer(context.Background(), wifs, transferDTOs, false, nil)
	if err != nil {
		log.Printf("Synchronous Transfer Error: %v", err)
	} else {
		fmt.Printf("✅ Synchronous Transfer Submitted! Txid: %s\n", *syncResponse.Txid)
	}

	// Wait a bit before next transfer to avoid double-spend issues in example
	fmt.Println("\nWaiting 5 seconds...")
	time.Sleep(5 * time.Second)

	// --- Example 2: Asynchronous Transfer & Polling ---
	fmt.Println("Attempting Asynchronous Transfer...")
	// You can add callbackURL and callbackSecret here if needed
	ticketID, err := m.AsynchronousTransfer(context.Background(), wifs, transferDTOs, false, nil, nil, nil)
	if err != nil {
		log.Printf("Asynchronous Transfer Error: %v", err)
	} else {
		fmt.Printf("✅ Asynchronous Transfer Submitted! Ticket ID: %s\n", *ticketID)

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
}
