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

	// --- 1. Create an async transfer to get a Ticket ID ---
	fmt.Println("Attempting Asynchronous Transfer...")
	transferDTOs := []mnee.TransferMneeDTO{
		{Amount: 1000, Address: recipientAddress},
	}
	wifs := []string{wif}

	ticketID, err := m.AsynchronousTransfer(context.Background(), wifs, transferDTOs, false, nil, nil, nil)
	if err != nil {
		log.Fatalf("Asynchronous Transfer Error: %v", err)
	}

	fmt.Printf("✅ Asynchronous Transfer Submitted! Ticket ID: %s\n", *ticketID)

	// --- 2. Poll the ticket for its status ---
	fmt.Printf("Polling ticket %s for status...\n", *ticketID)

	// Set a 1-minute timeout for the polling context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Poll every 3 seconds
	ticket, pollErr := m.PollTicket(ctx, *ticketID, 3*time.Second)
	if pollErr != nil {
		log.Fatalf("Error polling ticket: %v", pollErr)
	}

	fmt.Printf("✅ Ticket Polled! Final Status: %s\n", ticket.Status)
	if ticket.Status == mnee.SUCCESS {
		fmt.Printf("   Transaction ID: %s\n", *ticket.TxID)
	}
}
