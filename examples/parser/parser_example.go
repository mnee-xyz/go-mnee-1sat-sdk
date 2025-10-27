package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/bsv-blockchain/go-sdk/script" // Import the BSV SDK script package
	mnee "github.com/mnee-xyz/go-mnee-1sat-sdk"
)

func main() {
	apiKey := os.Getenv("MNEE_API_KEY")
	testAddress := os.Getenv("MNEE_TEST_ADDRESS")
	if apiKey == "" || testAddress == "" {
		log.Fatal("MNEE_API_KEY and MNEE_TEST_ADDRESS environment variables must be set")
	}

	m, err := mnee.NewMneeInstance(mnee.EnvSandbox, apiKey)
	if err != nil {
		log.Fatalf("Error creating MNEE instance: %v", err)
	}

	// 1. Get a real UTXO to get a valid MNEE script
	fmt.Println("Fetching a real UTXO to get its script...")
	utxos, err := m.GetUnspentTxos(context.Background(), []string{testAddress})
	if err != nil || len(utxos) == 0 {
		log.Fatalf("Could not get UTXOs for address %s: %v", testAddress, err)
	}

	// 2. Convert the Base64 script from the UTXO to ASM format
	base64Script := *utxos[0].Script
	scriptBytes, err := base64.StdEncoding.DecodeString(base64Script)
	if err != nil {
		log.Fatalf("Error decoding base64 script: %v", err)
	}
	s := script.NewFromBytes(scriptBytes)
	asmScript := s.ToASM()
	fmt.Printf("Got ASM script: %s\n", asmScript)

	// 3. Validate the ASM script
	fmt.Println("\nValidating the script...")
	isMnee, err := m.IsMneeScript(context.Background(), asmScript)
	if err != nil {
		log.Fatalf("Error validating script: %v", err)
	}

	if isMnee {
		fmt.Println("✅ Script is a valid MNEE script.")
	} else {
		fmt.Println("❌ Script is NOT a valid MNEE script.")
	}

	// 4. Example of an invalid script (simple P2PKH)
	fmt.Println("\nValidating a non-MNEE script...")
	p2pkhScript, _ := script.NewFromHex("1111111111111111111114oLvT2")
	invalidAsmScript := p2pkhScript.ToASM()
	fmt.Printf("Non-MNEE ASM script: %s\n", invalidAsmScript)

	isMnee, err = m.IsMneeScript(context.Background(), invalidAsmScript)
	if err != nil {
		log.Fatalf("Error validating script: %v", err)
	}

	if isMnee {
		fmt.Println("❌ Script IS a valid MNEE script (Error in logic!).")
	} else {
		fmt.Println("✅ Script is NOT a valid MNEE script (Correct).")
	}
}
