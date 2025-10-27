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
	assertions.NoError(err, "NewMneeInstance should not return an error")

	transferDTOs := []TransferMneeDTO{
		{
			Amount:  1000,
			Address: recipientAddress,
		},
	}
	wifs := []string{wif}

	t.Log("Attempting synchronous transfer...")
	transferResponse, err := m.SynchronousTransfer(context.Background(), wifs, transferDTOs, false, nil)

	assertions.NoError(err, "SynchronousTransfer() should not return an error")
	assertions.NotNil(transferResponse, "Transfer response should not be nil")
	assertions.NotNil(transferResponse.Txid, "Response should have a Txid")
	assertions.NotNil(transferResponse.Txhex, "Response should have a Txhex")

	t.Logf("✅ Successfully submitted transfer! Txid: %s", *transferResponse.Txid)
}

func TestSynchronousTransfer_WithTxos_Integration(t *testing.T) {
	assertions := assert.New(t)

	apiKey := os.Getenv("MNEE_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: MNEE_API_KEY not set")
	}

	wif := os.Getenv("MNEE_WIF")
	if wif == "" {
		t.Skip("Skipping integration test: MNEE_WIF not set")
	}

	testAddress := os.Getenv("MNEE_TEST_ADDRESS")
	if testAddress == "" {
		t.Skip("Skipping integration test: MNEE_TEST_ADDRESS not set")
	}

	recipientAddress := os.Getenv("MNEE_RECIPIENT_ADDRESS")
	if recipientAddress == "" {
		t.Skip("Skipping integration test: MNEE_RECIPIENT_ADDRESS not set")
	}

	m, err := NewMneeInstance(EnvSandbox, apiKey)
	assertions.NoError(err, "NewMneeInstance should not return an error")

	t.Log("Attempting to pre-fetch UTXOs...")
	mneeTxos, err := m.GetUnspentTxos(context.Background(), []string{testAddress})
	assertions.NoError(err, "GetUnspentTxos() failed, cannot test withTxos")
	assertions.NotEmpty(mneeTxos, "Test address has no UTXOs to spend")
	t.Logf("Successfully fetched %d UTXOs", len(mneeTxos))

	transferDTOs := []TransferMneeDTO{
		{
			Amount:  1000,
			Address: recipientAddress,
		},
	}
	wifs := []string{wif}

	t.Log("Attempting synchronous transfer with withTxos = true...")

	transferResponse, err := m.SynchronousTransfer(context.Background(), wifs, transferDTOs, true, mneeTxos)

	assertions.NoError(err, "SynchronousTransfer(withTxos=true) should not return an error")
	assertions.NotNil(transferResponse, "Transfer response should not be nil")
	assertions.NotNil(transferResponse.Txid, "Response should have a Txid")

	t.Logf("✅ Successfully submitted transfer with pre-fetched Txos! Txid: %s", *transferResponse.Txid)
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
	assertions.NoError(err, "NewMneeInstance should not return an error")

	transferDTOs := []TransferMneeDTO{
		{
			Amount:  1000,
			Address: recipientAddress,
		},
	}
	wifs := []string{wif}

	t.Log("Attempting asynchronous transfer...")
	ticketID, err := m.AsynchronousTransfer(context.Background(), wifs, transferDTOs, false, nil, nil, nil)

	assertions.NoError(err, "AsynchronousTransfer() should not return an error")
	assertions.NotNil(ticketID, "Ticket ID should not be nil")
	assertions.NotEmpty(*ticketID, "Ticket ID string should not be empty")

	t.Logf("✅ Successfully submitted async transfer! Ticket ID: %s", *ticketID)
}

func TestAsynchronousTransfer_WithTxos_Integration(t *testing.T) {
	assertions := assert.New(t)

	apiKey := os.Getenv("MNEE_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: MNEE_API_KEY not set")
	}

	wif := os.Getenv("MNEE_WIF")
	if wif == "" {
		t.Skip("Skipping integration test: MNEE_WIF not set")
	}

	testAddress := os.Getenv("MNEE_TEST_ADDRESS")
	if testAddress == "" {
		t.Skip("Skipping integration test: MNEE_TEST_ADDRESS not set")
	}

	recipientAddress := os.Getenv("MNEE_RECIPIENT_ADDRESS")
	if recipientAddress == "" {
		t.Skip("Skipping integration test: MNEE_RECIPIENT_ADDRESS not set")
	}

	m, err := NewMneeInstance(EnvSandbox, apiKey)
	assertions.NoError(err, "NewMneeInstance should not return an error")

	t.Log("Attempting to pre-fetch UTXOs...")
	mneeTxos, err := m.GetUnspentTxos(context.Background(), []string{testAddress})
	assertions.NoError(err, "GetUnspentTxos() failed, cannot test withTxos")
	assertions.NotEmpty(mneeTxos, "Test address has no UTXOs to spend")
	t.Logf("Successfully fetched %d UTXOs", len(mneeTxos))

	transferDTOs := []TransferMneeDTO{
		{
			Amount:  1000,
			Address: recipientAddress,
		},
	}
	wifs := []string{wif}

	t.Log("Attempting asynchronous transfer with withTxos = true...")

	ticketID, err := m.AsynchronousTransfer(context.Background(), wifs, transferDTOs, true, mneeTxos, nil, nil)

	assertions.NoError(err, "AsynchronousTransfer(withTxos=true) should not return an error")
	assertions.NotNil(ticketID, "Ticket ID should not be nil")
	assertions.NotEmpty(*ticketID, "Ticket ID string should not be empty")

	t.Logf("✅ Successfully submitted async transfer with pre-fetched Txos! Ticket ID: %s", *ticketID)
}
