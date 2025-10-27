package mnee

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig_Integration(t *testing.T) {
	assertions := assert.New(t)

	apiKey := os.Getenv("MNEE_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: MNEE_API_KEY environment variable not set")
	}

	m, err := NewMneeInstance(EnvSandbox, apiKey)

	if !assertions.NoError(err, "NewMneeInstance should not return an error") {
		return
	}
	assertions.NotNil(m, "MneeInstance should not be nil")

	t.Log("Calling GetConfig() against the sandbox API...")
	config, err := m.GetConfig(context.Background())

	if !assertions.NoError(err, "GetConfig() should not return an error") {
		return
	}
	assertions.NotNil(config, "Config response should not be nil")

	assertions.NotNil(config.TokenId, "Config should have a TokenId")
	assertions.NotNil(config.FeeAddress, "Config should have a FeeAddress")
	assertions.NotNil(config.MintAddress, "Config should have a MintAddress")
	assertions.Greater(len(config.Fees), 0, "Config should have at least one fee")
	assertions.GreaterOrEqual(config.Decimals, uint8(0), "Config should have valid decimals")

	t.Logf("✅ Successfully fetched config. Token ID: %s", *config.TokenId)
}
