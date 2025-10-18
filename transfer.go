package mnee

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"

	primitives "github.com/bsv-blockchain/go-sdk/primitives/ec"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/transaction"
	sighash "github.com/bsv-blockchain/go-sdk/transaction/sighash"
	"github.com/bsv-blockchain/go-sdk/transaction/template/p2pkh"
)

func (m *MNEE) Transfer(ctx context.Context, wifs []string, mneeTransferDTO []TransferMneeDTO, withTxos bool, mneeTxos []MneeTxo) error {

	var addressToPrivateKey map[string]*primitives.PrivateKey = make(map[string]*primitives.PrivateKey)
	var addresses []string = make([]string, 0, len(wifs))
	for _, wif := range wifs {
		privateKey, err := primitives.PrivateKeyFromWif(wif)
		if err != nil {
			return err
		}

		address, err := script.NewAddressFromPublicKey(privateKey.PubKey(), true)
		if err != nil {
			return err
		}

		addressToPrivateKey[address.AddressString] = privateKey
		addresses = append(addresses, address.AddressString)
	}

	config, err := m.GetConfig(ctx)
	if err != nil {
		return err
	}

	if config.Approver == nil || config.FeeAddress == nil || config.Fees == nil || config.TokenId == nil {
		return errors.New("invalid config")
	}

	approverPubKey, err := primitives.PublicKeyFromString(*config.Approver)
	if err != nil {
		return err
	}

	var mneeTransaction *transaction.Transaction = transaction.NewTransaction()
	var totalTransferAmt uint64
	for _, dto := range mneeTransferDTO {
		if dto.Amount == 0 {
			return errors.New("transfer amount must be greater than 0")
		}

		address, err := script.NewAddressFromString(dto.Address)
		if err != nil {
			return err
		}

		lockingScript, err := lock(address, approverPubKey)
		if err != nil {
			return err
		}

		transferInscription, err := createTransferInscription(*config.TokenId, dto.Amount)
		if err != nil {
			return err
		}

		err = mneeTransaction.Inscribe(&script.InscriptionArgs{
			ContentType:   "application/bsv-20",
			Data:          transferInscription,
			LockingScript: lockingScript,
		})
		if err != nil {
			return err
		}

		totalTransferAmt += dto.Amount
	}

	var txos []MneeTxo = make([]MneeTxo, 0)
	if withTxos {
		txos = mneeTxos
	} else {
		txos, err = m.GetUnspentTxos(ctx, addresses)
		if err != nil {
			return err
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

		scriptBytes, err := base64.StdEncoding.DecodeString(*txos[i].Script)
		if err != nil {
			return err
		}

		sighashFlags := sighash.ForkID | sighash.All | sighash.AnyOneCanPay
		unlockingScriptTemplate, err := p2pkh.Unlock(addressToPrivateKey[txos[i].Owners[0]], &sighashFlags)
		if err != nil {
			return err
		}

		err = mneeTransaction.AddInputFrom(
			*txos[i].Txid,
			uint32(txos[i].Vout),
			hex.EncodeToString(scriptBytes),
			uint64(txos[i].Satoshis),
			unlockingScriptTemplate,
		)
		if err != nil {
			return err
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
							return err
						}

						feeLockingScript, err := lock(feeAddress, approverPubKey)
						if err != nil {
							return err
						}

						feeInscription, err := createTransferInscription(*config.TokenId, fee.Fee)
						if err != nil {
							return err
						}

						err = mneeTransaction.Inscribe(&script.InscriptionArgs{
							ContentType:   "application/bsv-20",
							Data:          feeInscription,
							LockingScript: feeLockingScript,
						})
						if err != nil {
							return err
						}

						changeAddress, err := script.NewAddressFromString(txos[i].Owners[0])
						if err != nil {
							return err
						}

						changeLockingScript, err := lock(changeAddress, approverPubKey)
						if err != nil {
							return err
						}

						changeInscription, err := createTransferInscription(*config.TokenId, (totalInputAmount - totalTransferAmt - fee.Fee))
						if err != nil {
							return err
						}

						err = mneeTransaction.Inscribe(&script.InscriptionArgs{
							ContentType:   "application/bsv-20",
							Data:          changeInscription,
							LockingScript: changeLockingScript,
						})
						if err != nil {
							return err
						}

						break outer
					} else if (totalInputAmount - totalTransferAmt) == fee.Fee {
						feeAddress, err := script.NewAddressFromString(*config.FeeAddress)
						if err != nil {
							return err
						}

						feeLockingScript, err := lock(feeAddress, approverPubKey)
						if err != nil {
							return err
						}

						feeInscription, err := createTransferInscription(*config.TokenId, fee.Fee)
						if err != nil {
							return err
						}

						err = mneeTransaction.Inscribe(&script.InscriptionArgs{
							ContentType:   "application/bsv-20",
							Data:          feeInscription,
							LockingScript: feeLockingScript,
						})
						if err != nil {
							return err
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
		return errors.New("insufficient mnee balance")
	}

	err = mneeTransaction.Sign()
	if err != nil {
		return err
	}

	transferRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		(m.mneeURL + "/v2/transfer?auth_token=" + m.mneeToken),
		bytes.NewBuffer(fmt.Appendf(nil, "{\"rawtx\":\"%s\"}", base64.StdEncoding.EncodeToString(mneeTransaction.Bytes()))),
	)
	if err != nil {
		return err
	}

	transferResponse, err := m.httpClient.Do(transferRequest)
	if err != nil {
		return err
	}

	defer transferResponse.Body.Close()

	if transferResponse.StatusCode > 299 {
		var errorResponse map[string]any
		err = json.NewDecoder(transferResponse.Body).Decode(&errorResponse)
		if err != nil {
			return err
		}

		errorsMessage, ok := errorResponse["message"].(string)
		if !ok {
			return fmt.Errorf("status received from mnee-cosigner -> %d", transferResponse.StatusCode)
		}

		return errors.New(errorsMessage)
	}

	return nil
}
