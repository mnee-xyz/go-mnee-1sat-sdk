package mnee

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// GetUnspentTxos fetches all MNEE UTXOs for a given list of addresses.
func (m *MNEE) GetUnspentTxos(ctx context.Context, addresses []string) ([]MneeTxo, error) {

	addressesBuffer, err := json.Marshal(&addresses)
	if err != nil {
		return nil, err
	}

	utxosRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		(m.mneeURL + "/v1/utxos?auth_token=" + m.mneeToken),
		bytes.NewBuffer(addressesBuffer),
	)
	if err != nil {
		return nil, err
	}

	utxosRequest.Header.Set("Content-Type", "application/json")

	var txos []MneeTxo = make([]MneeTxo, 0)
	utxosResponse, err := m.httpClient.Do(utxosRequest)
	if err != nil {
		return nil, err
	}

	defer utxosResponse.Body.Close()

	if utxosResponse.StatusCode == http.StatusForbidden {
		return nil, ErrForbidden
	}

	if utxosResponse.StatusCode != http.StatusOK {
		var errorResponse map[string]any
		err = json.NewDecoder(utxosResponse.Body).Decode(&errorResponse)
		if err != nil {
			return nil, err
		}

		errorMessage, ok := errorResponse["message"].(string)
		if !ok {
			return nil, fmt.Errorf("status received from mnee-cosigner -> %d", utxosResponse.StatusCode)
		}

		return nil, errors.New(errorMessage)
	}

	err = json.NewDecoder(utxosResponse.Body).Decode(&txos)
	if err != nil {
		return nil, err
	}

	return txos, nil
}

// GetPaginatedUnspentTxos fetches MNEE UTXOs for a given list of addresses with pagination.
func (m *MNEE) GetPaginatedUnspentTxos(ctx context.Context, addresses []string, page int, size int) ([]MneeTxo, error) {

	addressesBuffer, err := json.Marshal(&addresses)
	if err != nil {
		return nil, err
	}

	utxosRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		(m.mneeURL + "/v2/utxos?page=" + fmt.Sprintf("%d", page) + "&size=" + fmt.Sprintf("%d", size) + "&auth_token=" + m.mneeToken),
		bytes.NewBuffer(addressesBuffer),
	)
	if err != nil {
		return nil, err
	}

	utxosRequest.Header.Set("Content-Type", "application/json")

	var txos []MneeTxo = make([]MneeTxo, 0)
	utxosResponse, err := m.httpClient.Do(utxosRequest)
	if err != nil {
		return nil, err
	}

	defer utxosResponse.Body.Close()

	if utxosResponse.StatusCode == http.StatusForbidden {
		return nil, ErrForbidden
	}

	if utxosResponse.StatusCode != http.StatusOK {
		var errorResponse map[string]any
		err = json.NewDecoder(utxosResponse.Body).Decode(&errorResponse)
		if err != nil {
			return nil, err
		}

		errorMessage, ok := errorResponse["message"].(string)
		if !ok {
			return nil, fmt.Errorf("status received from mnee-cosigner -> %d", utxosResponse.StatusCode)
		}

		return nil, errors.New(errorMessage)
	}

	err = json.NewDecoder(utxosResponse.Body).Decode(&txos)
	if err != nil {
		return nil, err
	}

	return txos, nil
}

// GetTxo fetches a single MNEE UTXO by its outpoint string (e.g., "txid_vout").
func (m *MNEE) GetTxo(ctx context.Context, outpoint string) (*MneeTxo, error) {

	utxoRequest, err := http.NewRequest(
		http.MethodGet,
		(m.mneeURL + "/v2/txos/" + outpoint + "?auth_token=" + m.mneeToken),
		nil,
	)
	if err != nil {
		return nil, err
	}

	utxoResponse, err := m.httpClient.Do(utxoRequest)
	if err != nil {
		return nil, err
	}

	defer utxoResponse.Body.Close()

	if utxoResponse.StatusCode == http.StatusForbidden {
		return nil, ErrForbidden
	}

	if utxoResponse.StatusCode != http.StatusOK {
		var errorResponse map[string]any
		err = json.NewDecoder(utxoResponse.Body).Decode(&errorResponse)
		if err != nil {
			return nil, err
		}

		errorMessage, ok := errorResponse["message"].(string)
		if !ok {
			return nil, fmt.Errorf("status received from mnee-cosigner -> %d", utxoResponse.StatusCode)
		}

		return nil, errors.New(errorMessage)
	}

	var txo MneeTxo
	err = json.NewDecoder(utxoResponse.Body).Decode(&txo)
	if err != nil {
		return nil, err
	}

	return &txo, nil
}
