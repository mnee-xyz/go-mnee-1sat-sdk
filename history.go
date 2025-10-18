package mnee

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func (m *MNEE) GetTransactionHistory(ctx context.Context, addresses []string) ([]TransactionHistoryDTO, error) {

	addressesBuffer, err := json.Marshal(&addresses)
	if err != nil {
		return nil, err
	}

	historyRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		(m.mneeURL + "/v1/sync?auth_token" + m.mneeToken),
		bytes.NewBuffer(addressesBuffer),
	)
	if err != nil {
		return nil, err
	}

	historyResponse, err := m.httpClient.Do(historyRequest)
	if err != nil {
		return nil, err
	}

	defer historyResponse.Body.Close()

	var history []TransactionHistoryDTO
	err = json.NewDecoder(historyResponse.Body).Decode(&history)
	if err != nil {
		return nil, err
	}

	return history, nil
}
