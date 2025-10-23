package mnee

import (
	"context"
	"encoding/base64"
	"os"
	"testing"

	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/stretchr/testify/assert"
)

func TestIsMneeScript_Integration(t *testing.T) {
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
	assertions.Nil(err, "NewMneeInstance should not return an error")
	assertions.NotNil(m, "MneeInstance should not be nil")

	t.Log("Test Case 1: Checking a real, valid MNEE script...")

	config, err := m.GetConfig(context.Background())
	assertions.Nil(err, "Failed to get config, cannot proceed")
	assertions.NotNil(config.FeeAddress)

	utxos, err := m.GetUnspentTxos(context.Background(), []string{testAddress})
	assertions.Nil(err, "Failed to get UTXOs")
	assertions.NotEmpty(utxos, "Fee address should have at least one UTXO to test")

	base64Script := *utxos[0].Script
	base64ScriptBytes, err := base64.StdEncoding.DecodeString(base64Script)
	assertions.Nil(err, "Failed to parse script from Base64")

	s := script.NewFromBytes(base64ScriptBytes)
	assertions.NotNil(s, "Failed to create script from bytes")

	asmScript := s.ToASM()

	isMnee, err := m.IsMneeScript(context.Background(), asmScript)
	assertions.Nil(err, "IsMneeScript should not return an error")
	assertions.True(isMnee, "A real UTXO from the feeAddress should be a valid MNEE script")

	t.Log("✅ Correctly identified a valid MNEE script")
}
