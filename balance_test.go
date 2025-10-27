package mnee

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBalances_Integration(t *testing.T) {
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
	assertions.NoError(err, "NewMneeInstance should not return an error")
	assertions.NotNil(m, "MneeInstance should not be nil")

	t.Log("Test Case 1: Fetching balances for known addresses...")

	config, err := m.GetConfig(context.Background())
	assertions.NoError(err, "Failed to get config, cannot proceed with balance test")
	assertions.NotNil(config)
	assertions.NotNil(config.FeeAddress, "FeeAddress in config is nil")

	feeAddress := *config.FeeAddress

	addressesToTest := []string{feeAddress, testAddress}

	balances, err := m.GetBalances(context.Background(), addressesToTest)

	assertions.NoError(err, "GetBalances() should not return an error")
	assertions.NotNil(balances, "Balances response should not be nil")
	assertions.Len(balances, 2, "Should get 2 balance results back")

	var foundFeeAddress, foundTestAddress bool
	for _, balance := range balances {
		switch *balance.Address {
		case feeAddress:
			assertions.Greater(balance.Amt, float64(0), "Fee address should have a balance")
			foundFeeAddress = true
		case testAddress:
			assertions.GreaterOrEqual(balance.Amt, float64(1000000), "Test address should have balance greater than or equal to 10")
			foundTestAddress = true
		}
	}
	assertions.True(foundFeeAddress, "Did not find balance for feeAddress")
	assertions.True(foundTestAddress, "Did not find balance for testAddress")

	t.Logf("Successfully fetched balances for %d addresses", len(balances))

	t.Log("Test Case 2: Fetching balances for an empty address list...")

	emptyAddressesList := []string{}
	balances, err = m.GetBalances(context.Background(), emptyAddressesList)

	assertions.NoError(err, "GetBalances() with empty list should not error")
	assertions.NotNil(balances, "Balances response should not be nil (should be '[]')")
	assertions.Len(balances, 0, "Balances list should be empty")

	t.Log("✅ Successfully handled empty address list (returned '[]')")
}
