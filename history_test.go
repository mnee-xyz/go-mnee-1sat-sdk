package mnee

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSpecificTransactionHistory_Integration(t *testing.T) {
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

	t.Log("Test Case: Fetching transaction history for a known address...")

	config, err := m.GetConfig(context.Background())
	if !assertions.NoError(err, "Failed to get config, cannot proceed with history test") {
		return
	}
	assertions.NotNil(config)
	assertions.NotNil(config.FeeAddress, "FeeAddress in config is nil")

	addressesToTest := []string{testAddress}

	history, err := m.GetSpecificTransactionHistory(context.Background(), addressesToTest, 0, 10)

	if !assertions.NoError(err, "GetSpecificTransactionHistory() should not return an error") {
		return
	}
	assertions.NotNil(history, "History response should not be nil")

	t.Logf("Found %d history items for the fee address", len(history))
	if len(history) > 0 {
		assertions.NotNil(history[0].Txid, "History item should have a Txid")
	}

	t.Logf("✅ Successfully fetched transaction history for %d addresses", len(addressesToTest))
}
