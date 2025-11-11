package mnee

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createTestPartialHex(t *testing.T, m *MNEE, wif, recipientAddress string) *string {
	transferDTOs := []TransferMneeDTO{
		{
			Amount:  1000,
			Address: recipientAddress,
		},
	}
	wifs := []string{wif}

	partialHex, err := m.PartialSign(context.Background(), wifs, transferDTOs, false, nil)
	if err != nil {
		t.Fatalf("Test setup failed: PartialSign failed: %v", err)
	}
	if partialHex == nil || *partialHex == "" {
		t.Fatal("Test setup failed: PartialSign returned nil or empty hex")
	}
	return partialHex
}

func TestSubmitRawTxSync_Integration(t *testing.T) {
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

	t.Log("Creating a partial tx to submit...")
	partialHex := createTestPartialHex(t, m, wif, recipientAddress)

	t.Log("Submitting raw tx synchronously...")
	syncResponse, err := m.SubmitRawTxSync(context.Background(), *partialHex)

	if !assertions.NoError(err, "SubmitRawTxSync should not return an error") {
		return
	}
	assertions.NotNil(syncResponse, "Sync response should not be nil")
	assertions.NotNil(syncResponse.Txid, "Response Txid should not be nil")
	assertions.NotNil(syncResponse.Txhex, "Response Txhex should not be nil")
	assertions.NotEmpty(*syncResponse.Txid, "Response Txid should not be empty")

	t.Logf("✅ Successfully submitted raw tx synchronously! TxID: %s", *syncResponse.Txid)
}

func TestSubmitRawTxAsync_Integration(t *testing.T) {
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

	t.Log("Creating a partial tx to submit...")
	partialHex := createTestPartialHex(t, m, wif, recipientAddress)

	t.Log("Submitting raw tx asynchronously...")
	ticketID, err := m.SubmitRawTxAsync(context.Background(), *partialHex, nil, nil)

	if !assertions.NoError(err, "SubmitRawTxAsync should not return an error") {
		return
	}
	assertions.NotNil(ticketID, "Ticket ID should not be nil")
	assertions.NotEmpty(*ticketID, "Ticket ID should not be empty")

	t.Logf("✅ Successfully submitted raw tx asynchronously! Ticket ID: %s", *ticketID)
	t.Log("Waiting 5 seconds before polling...")
	time.Sleep(5 * time.Second)

	t.Logf("Polling ticket %s for status...", *ticketID)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	ticket, pollErr := m.PollTicket(ctx, *ticketID, 2*time.Second)
	if !assertions.NoError(pollErr, "PollTicket failed") {
		return
	}
	assertions.NotNil(ticket, "Polled ticket should not be nil")
	assertions.Equal(*ticketID, *ticket.ID, "Ticket ID in response should match")
	assertions.Contains([]TicketStatus{SUCCESS, BROADCASTING}, ticket.Status, "Ticket status should be SUCCESS or BROADCASTING")

	t.Logf("✅ Successfully polled ticket. Final Status: %s", ticket.Status)
}
