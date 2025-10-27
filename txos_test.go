package mnee

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUnspentTxos_Integration(t *testing.T) {
	assertions := assert.New(t)

	apiKey := os.Getenv("MNEE_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: MNEE_API_KEY environment variable not set")
	}

	testAddress := os.Getenv("MNEE_TEST_ADDRESS")
	if testAddress == "" {
		t.Skip("Skipping integration test: MNEE_TEST_ADDRESS environment variable not set")
	}

	m, err := NewMneeInstance(EnvSandbox, apiKey)
	if !assertions.NoError(err, "NewMneeInstance should not return an error") {
		return
	}
	assertions.NotNil(m, "MneeInstance should not be nil")

	t.Log("Test Case 1: Fetching UTXOs for a known address...")

	config, err := m.GetConfig(context.Background())
	if !assertions.NoError(err, "Failed to get config, cannot proceed with UTXO test") {
		return
	}
	assertions.NotNil(config)
	assertions.NotNil(config.FeeAddress, "FeeAddress in config is nil")

	addressesToTest := []string{testAddress}

	txos, err := m.GetUnspentTxos(context.Background(), addressesToTest)

	if !assertions.NoError(err, "GetUnspentTxos() should not return an error") {
		return
	}
	assertions.NotNil(txos, "TXOs response should not be nil")

	t.Logf("Found %d UTXOs for the fee address", len(txos))
	if len(txos) > 0 {
		assertions.NotNil(txos[0].Txid, "UTXO should have a Txid")
		assertions.NotNil(txos[0].Script, "UTXO should have a Script")
	}

	t.Log("Test Case 2: Fetching UTXOs for an empty address list...")

	emptyAddressesList := []string{}
	txos, err = m.GetUnspentTxos(context.Background(), emptyAddressesList)

	if !assertions.NoError(err, "GetUnspentTxos() with empty list should not error") {
		return
	}
	assertions.NotNil(txos, "TXOs response should not be nil (should be '[]')")
	assertions.Len(txos, 0, "TXOs list should be empty")

	t.Log("✅ Successfully handled empty address list (returned '[]')")
}
