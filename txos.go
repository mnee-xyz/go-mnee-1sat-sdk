package mnee

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

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

	err = json.NewDecoder(utxosResponse.Body).Decode(&txos)
	if err != nil {
		return nil, err
	}

	return txos, nil
}

func (m *MNEE) GetTxo(ctx context.Context, outpoint string) (*MneeTxo, error) {

	utxoRequest, err := http.NewRequest(
		http.MethodGet,
		(m.mneeURL + "/v1/txo/" + outpoint + "?auth_token=" + m.mneeToken),
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

	var txo MneeTxo
	err = json.NewDecoder(utxoResponse.Body).Decode(&txo)
	if err != nil {
		return nil, err
	}

	return &txo, nil
}
