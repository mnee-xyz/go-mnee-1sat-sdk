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

	"github.com/bsv-blockchain/go-sdk/transaction"
)

// SubmitRawTxSync submits a pre-built, partially signed transaction hex
// (as a string) to the MNEE synchronous transfer endpoint.
//
// This is an "expert" function. The rawTxHex must be a valid MNEE transaction
// (e.g., one created by PartialSign) and will be submitted directly to the
// cosigner. The function waits for the cosigner's response and returns
// the final, fully-signed transaction details.
func (m *MNEE) SubmitRawTxSync(ctx context.Context, rawTxHex string) (*TransferResponseDTO, error) {

	txBytes, err := hex.DecodeString(rawTxHex)
	if err != nil {
		return nil, fmt.Errorf("invalid raw transaction hex: %w", err)
	}

	base64EncodedTx := base64.StdEncoding.EncodeToString(txBytes)

	jsonBody, err := json.Marshal(map[string]string{"rawtx": base64EncodedTx})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sync request: %w", err)
	}

	transferRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		(m.mneeURL + "/v1/transfer?auth_token=" + m.mneeToken),
		bytes.NewBuffer(jsonBody),
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
	}

	return &TransferResponseDTO{
		Txid:  nil,
		Txhex: nil,
	}, nil
}

// SubmitRawTxAsync submits a pre-built, partially signed transaction hex
// (as a string) to the MNEE asynchronous transfer endpoint.
//
// This is an "expert" function. The rawTxHex must be a valid MNEE transaction
// (e.g., one created by PartialSign). It submits the transaction and
// immediately returns a ticketID for polling.
func (m *MNEE) SubmitRawTxAsync(ctx context.Context, rawTxHex string, callbackURL *string, callbackSecret *string) (*string, error) {

	txBytes, err := hex.DecodeString(rawTxHex)
	if err != nil {
		return nil, fmt.Errorf("invalid raw transaction hex: %w", err)
	}

	base64EncodedTx := base64.StdEncoding.EncodeToString(txBytes)

	var transferRequestDTO TransferRequestDTO = TransferRequestDTO{
		RawTx:          base64EncodedTx,
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
