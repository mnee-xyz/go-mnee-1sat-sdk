package mnee

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSynchronousTransfer_Integration(t *testing.T) {
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

	m, err := NewMneeInstance(EnvSandbox, apiKey)
	assertions.Nil(err, "NewMneeInstance should not return an error")

	transferDTOs := []TransferMneeDTO{
		{
			Amount:  1000,
			Address: recipientAddress,
		},
	}
	wifs := []string{wif}

	t.Log("Attempting synchronous transfer...")
	transferResponse, err := m.SynchronousTransfer(context.Background(), wifs, transferDTOs, false, nil)

	assertions.Nil(err, "SynchronousTransfer() should not return an error")
	assertions.NotNil(transferResponse, "Transfer response should not be nil")
	assertions.NotNil(transferResponse.Txid, "Response should have a Txid")
	assertions.NotNil(transferResponse.Txhex, "Response should have a Txhex")

	t.Logf("✅ Successfully submitted transfer! Txid: %s", *transferResponse.Txid)
}

func TestAsynchronousTransfer_Integration(t *testing.T) {
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

	m, err := NewMneeInstance(EnvSandbox, apiKey)
	assertions.Nil(err, "NewMneeInstance should not return an error")

	transferDTOs := []TransferMneeDTO{
		{
			Amount:  1000,
			Address: recipientAddress,
		},
	}
	wifs := []string{wif}

	t.Log("Attempting asynchronous transfer...")
	ticketID, err := m.AsynchronousTransfer(context.Background(), wifs, transferDTOs, false, nil, nil, nil)

	assertions.Nil(err, "AsynchronousTransfer() should not return an error")
	assertions.NotNil(ticketID, "Ticket ID should not be nil")
	assertions.NotEmpty(*ticketID, "Ticket ID string should not be empty")

	t.Logf("✅ Successfully submitted async transfer! Ticket ID: %s", *ticketID)
}
