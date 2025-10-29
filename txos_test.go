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

	targetEnv := getTestEnvironment(t)

	m, err := NewMneeInstance(targetEnv, apiKey)
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

func TestGetTxo_Integration(t *testing.T) {
	assertions := assert.New(t)

	apiKey := os.Getenv("MNEE_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: MNEE_API_KEY environment variable not set")
	}

	testAddress := os.Getenv("MNEE_TEST_ADDRESS")
	if testAddress == "" {
		t.Skip("Skipping integration test: MNEE_TEST_ADDRESS environment variable not set")
	}

	targetEnv := getTestEnvironment(t)

	m, err := NewMneeInstance(targetEnv, apiKey)
	if !assertions.NoError(err, "NewMneeInstance should not return an error") {
		return
	}
	assertions.NotNil(m, "MneeInstance should not be nil")

	t.Log("Fetching UTXOs to get a valid outpoint...")
	utxos, err := m.GetUnspentTxos(context.Background(), []string{testAddress})
	if !assertions.NoError(err, "GetUnspentTxos() failed") {
		return
	}
	if !assertions.NotEmpty(utxos, "MNEE_TEST_ADDRESS has no UTXOs, cannot test GetTxo") {
		return
	}

	outpointToTest := *utxos[0].Outpoint
	expectedTxid := *utxos[0].Txid
	t.Logf("Got outpoint to test: %s", outpointToTest)

	txo, err := m.GetTxo(context.Background(), outpointToTest)

	if !assertions.NoError(err, "GetTxo() should not return an error") {
		return
	}
	if !assertions.NotNil(txo, "TXO response should not be nil") {
		return
	}

	assertions.NotNil(txo.Txid, "Returned TXO should have a Txid")
	assertions.Equal(expectedTxid, *txo.Txid, "Returned TXO's txid should match the one from GetUnspentTxos")
	assertions.Equal(outpointToTest, *txo.Outpoint, "Returned TXO's outpoint should match the requested outpoint")
	assertions.NotNil(txo.Data, "Returned TXO should have Data")
	assertions.NotNil(txo.Data.Bsv21, "Returned TXO Data should have Bsv21 info")
	assertions.Greater(txo.Data.Bsv21.Amt, uint64(0), "Returned TXO should have a positive amount")

	t.Log("✅ Successfully fetched single TXO by outpoint")
}
