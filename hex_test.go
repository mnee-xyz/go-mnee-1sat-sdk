package mnee

import (
	"context"
	"encoding/hex"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMNEETxHex_Integration(t *testing.T) {
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

	t.Log("Submitting a transfer to get a valid TxID...")
	transferDTOs := []TransferMneeDTO{
		{
			Amount:  1000,
			Address: recipientAddress,
		},
	}
	wifs := []string{wif}

	syncResponse, err := m.SynchronousTransfer(context.Background(), wifs, transferDTOs, false, nil)
	if !assertions.NoError(err, "Test setup failed: SynchronousTransfer failed") {
		return
	}
	if !assertions.NotNil(syncResponse) || !assertions.NotNil(syncResponse.Txid) {
		t.Fatal("Test setup failed: SynchronousTransfer did not return a Txid")
	}

	txidToTest := *syncResponse.Txid
	expectedHex := *syncResponse.Txhex
	t.Logf("Got TxID to test: %s", txidToTest)

	txHex, err := m.GetMNEETxHex(context.Background(), txidToTest)

	if !assertions.NoError(err, "GetMNEETxHex should not return an error") {
		return
	}
	if !assertions.NotNil(txHex, "Returned TxHex should not be nil") {
		return
	}

	assertions.NotEmpty(*txHex, "Returned TxHex should not be empty")
	assertions.Equal(expectedHex, *txHex, "Returned hex should match the hex from SynchronousTransfer")

	_, err = hex.DecodeString(*txHex)
	assertions.NoError(err, "Returned string is not valid hex")

	t.Logf("✅ Successfully fetched and validated TxHex for %s", txidToTest)
}
