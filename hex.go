package mnee

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// GetMNEETxHex fetches a MNEE transaction by its TXID and returns its full hex.
func (m *MNEE) GetMNEETxHex(ctx context.Context, txid string) (*string, error) {

	mneeHexRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		(m.mneeURL + "/v1/tx/" + txid + "?auth_token=" + m.mneeToken),
		nil,
	)
	if err != nil {
		return nil, err
	}

	mneeHexRequest.Header.Set("Content-Type", "application/json")

	mneeHexResponse, err := m.httpClient.Do(mneeHexRequest)
	if err != nil {
		return nil, err
	}

	defer mneeHexResponse.Body.Close()

	if mneeHexResponse.StatusCode == http.StatusForbidden {
		return nil, ErrForbidden
	}

	if mneeHexResponse.StatusCode != http.StatusOK {
		var errorResponse map[string]any
		err = json.NewDecoder(mneeHexResponse.Body).Decode(&errorResponse)
		if err != nil {
			return nil, err
		}

		errorMessage, ok := errorResponse["message"].(string)
		if !ok {
			return nil, fmt.Errorf("status received from mnee-cosigner -> %d", mneeHexResponse.StatusCode)
		}

		return nil, errors.New(errorMessage)
	}

	var mneeHexBody struct {
		RawTx *string `json:"rawtx"`
	}
	err = json.NewDecoder(mneeHexResponse.Body).Decode(&mneeHexBody)
	if err != nil {
		return nil, err
	}

	if mneeHexBody.RawTx == nil {
		return nil, errors.New("record not found")
	}

	mneeTxBytes, err := base64.StdEncoding.DecodeString(*mneeHexBody.RawTx)
	if err != nil {
		return nil, err
	}

	var mneeTxHex string = hex.EncodeToString(mneeTxBytes)

	return &mneeTxHex, nil
}
