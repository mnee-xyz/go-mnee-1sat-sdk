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

	m, err := NewMneeInstance(EnvSandbox, apiKey)
	assertions.Nil(err, "NewMneeInstance should not return an error")
	assertions.NotNil(m, "MneeInstance should not be nil")

	t.Log("Test Case 1: Fetching balances for known addresses...")

	config, err := m.GetConfig(context.Background())
	assertions.Nil(err, "Failed to get config, cannot proceed with balance test")
	assertions.NotNil(config)
	assertions.NotNil(config.FeeAddress, "FeeAddress in config is nil")

	feeAddress := *config.FeeAddress
	randomAddress := "1KtQkZKHPwmdy7WtHWXz8PyUyrVy11y4ED"

	addressesToTest := []string{feeAddress, randomAddress}

	balances, err := m.GetBalances(context.Background(), addressesToTest)

	assertions.Nil(err, "GetBalances() should not return an error")
	assertions.NotNil(balances, "Balances response should not be nil")
	assertions.Len(balances, 2, "Should get 2 balance results back")

	var foundFeeAddress, foundRandomAddress bool
	for _, balance := range balances {
		switch *balance.Address {
		case feeAddress:
			assertions.Greater(balance.Amt, float64(0), "Fee address should have a balance")
			foundFeeAddress = true
		case randomAddress:
			assertions.Equal(balance.Amt, float64(1000000), "Random address should have 0 balance")
			foundRandomAddress = true
		}
	}
	assertions.True(foundFeeAddress, "Did not find balance for feeAddress")
	assertions.True(foundRandomAddress, "Did not find balance for randomAddress")

	t.Logf("Successfully fetched balances for %d addresses", len(balances))

	t.Log("Test Case 2: Fetching balances for an empty address list...")

	emptyAddressesList := []string{}
	balances, err = m.GetBalances(context.Background(), emptyAddressesList)

	assertions.Nil(err, "GetBalances() with empty list should not error")
	assertions.NotNil(balances, "Balances response should not be nil (should be '[]')")
	assertions.Len(balances, 0, "Balances list should be empty")

	t.Log("Successfully handled empty address list (returned '[]')")
}
