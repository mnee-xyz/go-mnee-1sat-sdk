package mnee

import (
	"strconv"
)

func validateOrdInscription(tokens []string) bool {

	if tokens[0] != "OP_FALSE" || tokens[1] != "OP_IF" || tokens[2] != "6f7264" || tokens[3] != "OP_TRUE" ||
		tokens[4] != "6170706c69636174696f6e2f6273762d3230" || tokens[5] != "OP_FALSE" || tokens[7] != "OP_ENDIF" {
		return false
	}

	return true
}

func validateTransferLockingScript(tokens []string, config *SystemConfig) bool {

	if len(tokens) != 7 || config.Approver == nil {
		return false
	}

	if !(tokens[0] == "OP_DUP" &&
		tokens[1] == "OP_HASH160" &&
		tokens[3] == "OP_EQUALVERIFY" &&
		tokens[4] == "OP_CHECKSIGVERIFY" &&
		tokens[6] == "OP_CHECKSIG") {
		return false
	}

	if tokens[5] != *config.Approver {
		return false
	}

	return true
}

func validateTransferInscription(inscription *TransferTokenInscription, config *SystemConfig) bool {

	if inscription == nil || config.TokenId == nil {
		return false
	}

	if inscription.Protocol != BSV20 {
		return false
	}

	if inscription.Operation != TRANSFER {
		return false
	}

	if !isPositiveInteger(inscription.Amount) {
		return false
	}

	if inscription.TokenID != (*config.TokenId) {
		return false
	}

	return true
}

func validateDeployChainInscription(inscription *DeployChainInscription, config *SystemConfig) bool {

	if inscription == nil || config.TokenId == nil {
		return false
	}

	if inscription.Protocol != BSV20 {
		return false
	}

	if inscription.Operation != TRANSFER {
		return false
	}

	if !isPositiveInteger(inscription.Amount) {
		return false
	}

	if inscription.TokenID != (*config.TokenId) {
		return false
	}

	if inscription.Metadata == nil {
		return false
	}

	if inscription.Metadata.Action != ACTION_MINT && inscription.Metadata.Action != ACTION_REDEEM {
		return false
	}

	return true
}

func isPositiveInteger(value string) bool {

	v, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return false
	}

	return v > 0
}
