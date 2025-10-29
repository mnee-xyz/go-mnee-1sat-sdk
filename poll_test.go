package mnee

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPollTicket_Integration(t *testing.T) {
	assertions := assert.New(t)

	apiKey := os.Getenv("MNEE_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: MNEE_API_KEY not set")
	}

	wif := os.Getenv("MNEE_WIF")
	if wif == "" {
		t.Skip("Skipping integration test: MNEE_WIF not set")
	}

	recipientAddress := os.Getenv("MNEE_RECIPIENT_ADDRESS")
	if recipientAddress == "" {
		t.Skip("Skipping integration test: MNEE_RECIPIENT_ADDRESS not set")
	}

	targetEnv := getTestEnvironment(t)

	m, err := NewMneeInstance(targetEnv, apiKey)
	if !assertions.NoError(err, "NewMneeInstance should not return an error") {
		return
	}

	t.Log("Submitting async transfer to get a ticket ID...")

	transferDTOs := []TransferMneeDTO{
		{
			Amount:  1000,
			Address: recipientAddress,
		},
	}
	wifs := []string{wif}

	ticketID, err := m.AsynchronousTransfer(context.Background(), wifs, transferDTOs, false, nil, nil, nil)
	if !assertions.NoError(err, "AsynchronousTransfer failed, cannot test PollTicket") {
		return
	}

	t.Log("Waiting 2 seconds for previous transactions to settle...")
	time.Sleep(5 * time.Second)

	assertions.NotNil(ticketID)

	t.Logf("Got ticket ID: %s. Polling for status...", *ticketID)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	ticket, err := m.PollTicket(ctx, *ticketID, 2*time.Second)

	if !assertions.NoError(err, "PollTicket() should not return an error") {
		return
	}
	assertions.NotNil(ticket, "Ticket should not be nil")
	assertions.Equal(*ticketID, *ticket.ID, "Ticket ID in response should match")

	assertions.Contains([]TicketStatus{SUCCESS, BROADCASTING}, ticket.Status, "Ticket status should be SUCCESS or BROADCASTING")

	if ticket.Status == SUCCESS {
		assertions.NotNil(ticket.TxID, "Successful ticket should have a TxID")
		t.Logf("✅ Successfully polled ticket and confirmed SUCCESS. TxID: %s", *ticket.TxID)
	} else {
		t.Logf("✅ Successfully polled ticket. Status: %s", ticket.Status)
	}
}
