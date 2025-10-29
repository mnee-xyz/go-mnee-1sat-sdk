package mnee

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/bsv-blockchain/go-sdk/script"
)

// IsMneeScript validates if a given ASM script is a valid MNEE token script.
// It checks the script structure, BSV-20 inscription, locking script,
// and token details against the current MNEE system configuration.
// 'asmScript' should be the ASM script format of every output of a transaction.
func (m *MNEE) IsMneeScript(ctx context.Context, asmScript string) (bool, error) {

	config, err := m.GetConfig(ctx)
	if err != nil {
		return false, err
	}

	var scriptTokens []string = strings.Split(asmScript, " ")
	if len(scriptTokens) != 13 && len(scriptTokens) != 15 {
		return false, nil
	}

	var valid bool = validateOrdInscription(scriptTokens)
	if !valid {
		return false, nil
	}

	if len(scriptTokens) == 13 {
		p2pkhScript, err := script.NewFromASM(strings.Join(scriptTokens[8:], " "))
		if err != nil {
			return false, nil
		}

		if !p2pkhScript.IsP2PKH() {
			return false, nil
		}

		valid = validateTransferLockingScript(scriptTokens[8:], config)
		if !valid {
			return false, nil
		}
	}

	if len(scriptTokens) == 15 {
		valid = validateTransferLockingScript(scriptTokens[8:], config)
		if !valid {
			return false, nil
		}
	}

	decoded, err := hex.DecodeString(scriptTokens[6])
	if err != nil {
		return false, nil
	}

	var transferInscription TransferTokenInscription

	err = json.Unmarshal(decoded, &transferInscription)
	if err != nil {
		var deployInscription DeployChainInscription

		err = json.Unmarshal(decoded, &deployInscription)
		if err != nil {
			return false, nil
		}

		return validateDeployChainInscription(&deployInscription, config), nil
	}

	return validateTransferInscription(&transferInscription, config), nil
}
