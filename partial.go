package mnee

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"slices"

	primitives "github.com/bsv-blockchain/go-sdk/primitives/ec"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/transaction"
	sighash "github.com/bsv-blockchain/go-sdk/transaction/sighash"
	"github.com/bsv-blockchain/go-sdk/transaction/template/p2pkh"
)

// PartialSign builds a MNEE transfer transaction and signs it *only* with the
// WIFs provided. It returns the partially signed transaction as a hex string.
func (m *MNEE) PartialSign(ctx context.Context, wifs []string, mneeTransferDTO []TransferMneeDTO, withTxos bool,
	mneeTxos []MneeTxo) (*string, error) {

	var addressToPrivateKey map[string]*primitives.PrivateKey = make(map[string]*primitives.PrivateKey)
	var addresses []string = make([]string, 0, len(wifs))
	for _, wif := range wifs {
		privateKey, err := primitives.PrivateKeyFromWif(wif)
		if err != nil {
			return nil, err
		}

		address, err := script.NewAddressFromPublicKey(privateKey.PubKey(), true)
		if err != nil {
			return nil, err
		}

		addressToPrivateKey[address.AddressString] = privateKey
		addresses = append(addresses, address.AddressString)
	}

	config, err := m.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	if config.Approver == nil || config.FeeAddress == nil || config.Fees == nil || config.TokenId == nil {
		return nil, ErrInvalidConfig
	}

	approverPubKey, err := primitives.PublicKeyFromString(*config.Approver)
	if err != nil {
		return nil, err
	}

	var mneeTransaction *transaction.Transaction = transaction.NewTransaction()
	var totalTransferAmt uint64
	for _, dto := range mneeTransferDTO {
		if dto.Amount == 0 {
			return nil, ErrTransferAmountGreaterThan0
		}

		address, err := script.NewAddressFromString(dto.Address)
		if err != nil {
			return nil, err
		}

		lockingScript, err := lock(address, approverPubKey)
		if err != nil {
			return nil, err
		}

		transferInscription, err := createTransferInscription(*config.TokenId, dto.Amount)
		if err != nil {
			return nil, err
		}

		err = mneeTransaction.Inscribe(&script.InscriptionArgs{
			ContentType:   "application/bsv-20",
			Data:          transferInscription,
			LockingScript: lockingScript,
		})
		if err != nil {
			return nil, err
		}

		totalTransferAmt += dto.Amount
	}

	var txos []MneeTxo = make([]MneeTxo, 0)
	if withTxos {
		txos = mneeTxos
	} else {
		txos, err = m.GetUnspentTxos(ctx, addresses)
		if err != nil {
			return nil, err
		}
	}

	var inputAddresses []string = make([]string, 0)
	var totalInputAmount uint64

outer:
	for i := range txos {
		if txos[i].Data == nil || txos[i].Data.Bsv21 == nil || txos[i].Txid == nil ||
			txos[i].Script == nil || txos[i].Data.Bsv21.Amt == 0 ||
			len(txos[i].Owners) == 0 {
			continue
		}

		if privateKey, ok := addressToPrivateKey[txos[i].Owners[0]]; !ok || privateKey == nil {
			continue
		}

		scriptBytes, err := base64.StdEncoding.DecodeString(*txos[i].Script)
		if err != nil {
			return nil, err
		}

		sighashFlags := sighash.ForkID | sighash.All | sighash.AnyOneCanPay
		unlockingScriptTemplate, err := p2pkh.Unlock(addressToPrivateKey[txos[i].Owners[0]], &sighashFlags)
		if err != nil {
			return nil, err
		}

		err = mneeTransaction.AddInputFrom(
			*txos[i].Txid,
			uint32(txos[i].Vout),
			hex.EncodeToString(scriptBytes),
			uint64(txos[i].Satoshis),
			unlockingScriptTemplate,
		)
		if err != nil {
			return nil, err
		}

		totalInputAmount += txos[i].Data.Bsv21.Amt
		if !slices.Contains(inputAddresses, txos[i].Owners[0]) {
			inputAddresses = append(inputAddresses, txos[i].Owners[0])
		}

		if totalInputAmount >= totalTransferAmt {
			var actualTransferAmt uint64
			for _, dto := range mneeTransferDTO {
				if slices.Contains(inputAddresses, dto.Address) {
					continue
				} else {
					actualTransferAmt += dto.Amount
				}
			}

			for _, fee := range config.Fees {
				if actualTransferAmt >= fee.MinAmt && actualTransferAmt <= fee.MaxAmt {
					if (totalInputAmount - totalTransferAmt) > fee.Fee {
						feeAddress, err := script.NewAddressFromString(*config.FeeAddress)
						if err != nil {
							return nil, err
						}

						feeLockingScript, err := lock(feeAddress, approverPubKey)
						if err != nil {
							return nil, err
						}

						feeInscription, err := createTransferInscription(*config.TokenId, fee.Fee)
						if err != nil {
							return nil, err
						}

						err = mneeTransaction.Inscribe(&script.InscriptionArgs{
							ContentType:   "application/bsv-20",
							Data:          feeInscription,
							LockingScript: feeLockingScript,
						})
						if err != nil {
							return nil, err
						}

						changeAddress, err := script.NewAddressFromString(txos[i].Owners[0])
						if err != nil {
							return nil, err
						}

						changeLockingScript, err := lock(changeAddress, approverPubKey)
						if err != nil {
							return nil, err
						}

						changeInscription, err := createTransferInscription(*config.TokenId, (totalInputAmount - totalTransferAmt - fee.Fee))
						if err != nil {
							return nil, err
						}

						err = mneeTransaction.Inscribe(&script.InscriptionArgs{
							ContentType:   "application/bsv-20",
							Data:          changeInscription,
							LockingScript: changeLockingScript,
						})
						if err != nil {
							return nil, err
						}

						break outer
					} else if (totalInputAmount - totalTransferAmt) == fee.Fee {
						feeAddress, err := script.NewAddressFromString(*config.FeeAddress)
						if err != nil {
							return nil, err
						}

						feeLockingScript, err := lock(feeAddress, approverPubKey)
						if err != nil {
							return nil, err
						}

						feeInscription, err := createTransferInscription(*config.TokenId, fee.Fee)
						if err != nil {
							return nil, err
						}

						err = mneeTransaction.Inscribe(&script.InscriptionArgs{
							ContentType:   "application/bsv-20",
							Data:          feeInscription,
							LockingScript: feeLockingScript,
						})
						if err != nil {
							return nil, err
						}

						break outer
					} else {
						continue outer
					}
				}
			}
		}
	}

	if totalInputAmount < totalTransferAmt {
		return nil, ErrInsufficientMneeBalance
	}

	err = mneeTransaction.Sign()
	if err != nil {
		return nil, err
	}

	var partialHex string = mneeTransaction.Hex()

	return &partialHex, nil
}
