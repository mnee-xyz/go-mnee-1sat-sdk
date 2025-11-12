package mnee

import (
	"context"
	"encoding/hex"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPartialSign_Integration(t *testing.T) {
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

	transferDTOs := []TransferMneeDTO{
		{
			Amount:  1000,
			Address: recipientAddress,
		},
	}
	wifs := []string{wif}

	t.Log("Attempting to partially sign a transfer...")
	partialHex, err := m.PartialSign(context.Background(), wifs, transferDTOs, false, nil)

	if !assertions.NoError(err, "PartialSign() should not return an error") {
		return
	}
	if !assertions.NotNil(partialHex, "Returned partial hex string should not be nil") {
		return
	}

	assertions.NotEmpty(*partialHex, "Partial hex string should not be empty")
	_, err = hex.DecodeString(*partialHex)
	assertions.NoError(err, "Returned string is not valid hex")

	t.Logf("✅ Successfully created and signed partial transaction hex.")
}
