package mnee

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"

	primitives "github.com/bsv-blockchain/go-sdk/primitives/ec"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/transaction"
	sighash "github.com/bsv-blockchain/go-sdk/transaction/sighash"
	"github.com/bsv-blockchain/go-sdk/transaction/template/p2pkh"
)

// SynchronousTransfer builds, signs, and submits a MNEE transfer transaction,
// waiting for the final cosigned transaction response from the MNEE API.
//
// It automatically selects UTXOs unless `withTxos` is true and `mneeTxos` are provided.
// It calculates and includes the required MNEE fee based on the system config.
// Returns the final transaction details (Txid, Txhex) upon success.
// Use this function when you need immediate confirmation that the cosigner accepted the transaction.
func (m *MNEE) SynchronousTransfer(ctx context.Context, wifs []string, mneeTransferDTO []TransferMneeDTO, withTxos bool,
	mneeTxos []MneeTxo) (*TransferResponseDTO, error) {

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

	transferRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		(m.mneeURL + "/v1/transfer?auth_token=" + m.mneeToken),
		bytes.NewBuffer(fmt.Appendf(nil, "{\"rawtx\":\"%s\"}", base64.StdEncoding.EncodeToString(mneeTransaction.Bytes()))),
	)
	if err != nil {
		return nil, err
	}

	transferResponse, err := m.httpClient.Do(transferRequest)
	if err != nil {
		return nil, err
	}

	defer transferResponse.Body.Close()

	if transferResponse.StatusCode == http.StatusForbidden {
		return nil, ErrForbidden
	}

	if transferResponse.StatusCode != http.StatusOK {
		var errorResponse map[string]any
		err = json.NewDecoder(transferResponse.Body).Decode(&errorResponse)
		if err != nil {
			return nil, err
		}

		errorMessage, ok := errorResponse["message"].(string)
		if !ok {
			return nil, fmt.Errorf("status received from mnee-cosigner -> %d", transferResponse.StatusCode)
		}

		return nil, errors.New(errorMessage)
	}

	var transferResponseBody struct {
		Rawtx *string `json:"rawtx,omitempty"`
	}

	err = json.NewDecoder(transferResponse.Body).Decode(&transferResponseBody)
	if err != nil {
		return nil, err
	}

	if transferResponseBody.Rawtx != nil {
		transactionBytes, err := base64.StdEncoding.DecodeString(*transferResponseBody.Rawtx)
		if err != nil {
			return nil, err
		}

		finalTx, err := transaction.NewTransactionFromBytes(transactionBytes)
		if err != nil {
			return nil, err
		}

		var txHex string = finalTx.Hex()
		var txID string = finalTx.TxID().String()

		return &TransferResponseDTO{
			Txid:  &txID,
			Txhex: &txHex,
		}, nil
	} else {
		return &TransferResponseDTO{
			Txid:  nil,
			Txhex: nil,
		}, nil
	}
}

// AsynchronousTransfer builds, signs, and submits a MNEE transfer transaction,
// returning a ticket ID immediately without waiting for cosigner processing.
//
// It automatically selects UTXOs unless `withTxos` is true and `mneeTxos` are provided.
// It calculates and includes the required MNEE fee based on the system config.
// The status of the transfer can be tracked using the returned ticket ID with PollTicket
// or via webhooks (if callbackURL is provided).
// Use this function for non-blocking operations.
func (m *MNEE) AsynchronousTransfer(ctx context.Context, wifs []string, mneeTransferDTO []TransferMneeDTO, withTxos bool,
	mneeTxos []MneeTxo, callbackURL *string, callbackSecret *string) (*string, error) {

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

	var transferRequestDTO TransferRequestDTO = TransferRequestDTO{
		RawTx:          base64.StdEncoding.EncodeToString(mneeTransaction.Bytes()),
		CallbackURL:    callbackURL,
		CallbackSecret: callbackSecret,
	}

	transferRequestBuffer, err := json.Marshal(&transferRequestDTO)
	if err != nil {
		return nil, err
	}

	transferRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		(m.mneeURL + "/v2/transfer?auth_token=" + m.mneeToken),
		bytes.NewBuffer(transferRequestBuffer),
	)
	if err != nil {
		return nil, err
	}

	transferResponse, err := m.httpClient.Do(transferRequest)
	if err != nil {
		return nil, err
	}

	defer transferResponse.Body.Close()

	if transferResponse.StatusCode == http.StatusForbidden {
		return nil, ErrForbidden
	}

	if transferResponse.StatusCode != http.StatusOK {
		var errorResponse map[string]any
		err = json.NewDecoder(transferResponse.Body).Decode(&errorResponse)
		if err != nil {
			return nil, err
		}

		errorMessage, ok := errorResponse["message"].(string)
		if !ok {
			return nil, fmt.Errorf("status received from mnee-cosigner -> %d", transferResponse.StatusCode)
		}

		return nil, errors.New(errorMessage)
	}

	bodyBytes, err := io.ReadAll(transferResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if len(bodyBytes) == 0 {
		return nil, ErrReceivedEmptyTicketID
	}

	var ticketID string = string(bodyBytes)

	return &ticketID, nil
}
