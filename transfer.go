package mnee

import (
	"bytes"
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

func (m *MNEE) Transfer(wifs []string, mneeTransferDTO []TransferMneeDTO) error {
	var addressToWifMap map[string]*primitives.PrivateKey = make(map[string]*primitives.PrivateKey)
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

		addressToWifMap[address.AddressString] = privateKey
		addresses = append(addresses, address.AddressString)
	}

	var newClient *http.Client = &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			ForceAttemptHTTP2: false,
		},
		Timeout: 0,
	}

retry:
	configRequest, err := http.NewRequest(
		http.MethodGet,
		(m.mneeURL + "/v1/config?auth_token=" + m.mneeToken),
		nil,
	)
	if err != nil {
		return err
	}

	configResponse, err := newClient.Do(configRequest)
	if err != nil {
		return err
	}

	var config SystemConfig
	err = json.NewDecoder(configResponse.Body).Decode(&config)
	if err != nil {
		configResponse.Body.Close()
		return err
	}

	configResponse.Body.Close()
	if config.Approver == nil || config.FeeAddress == nil || config.Fees == nil || config.TokenId == nil {
		goto retry
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

		pubKey, err := primitives.PublicKeyFromString(*config.Approver)
		if err != nil {
			return err
		}

		lockingScript, err := lock(address, pubKey)
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

	utxosRequest, err := json.Marshal(&addresses)
	if err != nil {
		return err
	}

	var inputAddresses []string = make([]string, 0)
	var totalInputAmount uint64
	request, err := http.NewRequest(
		http.MethodPost,
		(m.mneeURL + "/v1/utxos?auth_token=" + m.mneeToken),
		bytes.NewBuffer(utxosRequest),
	)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	var txos []MneeTxo = make([]MneeTxo, 0)
	response, err := newClient.Do(request)
	if err != nil {
		return err
	}

	err = json.NewDecoder(response.Body).Decode(&txos)
	if err != nil {
		return err
	}

outer:
	for _, txo := range txos {
		if txo.Data == nil || txo.Data.Bsv21 == nil || txo.Txid == nil ||
			txo.Script == nil || txo.Data.Bsv21.Amt == 0 ||
			len(txo.Owners) == 0 {
			continue
		}

		totalInputAmount += txo.Data.Bsv21.Amt

		scriptBytes, err := base64.StdEncoding.DecodeString(*txo.Script)
		if err != nil {
			return err
		}

		sighashFlags := sighash.ForkID | sighash.All | sighash.AnyOneCanPay
		unlockingScriptTemplate, err := p2pkh.Unlock(addressToWifMap[txo.Owners[0]], &sighashFlags)
		if err != nil {
			return err
		}

		err = mneeTransaction.AddInputFrom(
			*txo.Txid,
			uint32(txo.Vout),
			hex.EncodeToString(scriptBytes),
			uint64(txo.Satoshis),
			unlockingScriptTemplate,
		)
		if err != nil {
			return err
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

						pubKey, err := primitives.PublicKeyFromString(*config.Approver)
						if err != nil {
							return err
						}

						feeLockingScript, err := lock(feeAddress, pubKey)
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

						changeAddress, err := script.NewAddressFromString(txo.Owners[0])
						if err != nil {
							return err
						}

						changeLockingScript, err := lock(changeAddress, pubKey)
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

						pubKey, err := primitives.PublicKeyFromString(*config.Approver)
						if err != nil {
							return err
						}

						feeLockingScript, err := lock(feeAddress, pubKey)
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

	transferRequest, err := http.NewRequest(
		http.MethodPost,
		(m.mneeURL + "/v2/transfer?auth_token=" + m.mneeToken),
		bytes.NewBuffer(fmt.Appendf(nil, "{\"rawtx\":\"%s\"}", base64.StdEncoding.EncodeToString(mneeTransaction.Bytes()))),
	)
	if err != nil {
		return err
	}

	transferResponse, err := newClient.Do(transferRequest)
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
			return fmt.Errorf("received from mnee-cosigner %d", transferResponse.StatusCode)
		}

		return errors.New(errorsMessage)
	}

	return nil
}
