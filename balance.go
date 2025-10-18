package mnee

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func (m *MNEE) GetBalances(ctx context.Context, addresses []string) ([]BalanceDataDTO, error) {

	addressesBuffer, err := json.Marshal(&addresses)
	if err != nil {
		return nil, err
	}

	balancesRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		(m.mneeURL + "/v2/balance?auth_token=" + m.mneeToken),
		bytes.NewBuffer(addressesBuffer),
	)
	if err != nil {
		return nil, err
	}

	balancesRequest.Header.Set("Content-Type", "application/json")

	var balances []BalanceDataDTO = make([]BalanceDataDTO, 0)
	balancesResponse, err := m.httpClient.Do(balancesRequest)
	if err != nil {
		return nil, err
	}

	defer balancesResponse.Body.Close()

	err = json.NewDecoder(balancesResponse.Body).Decode(&balances)
	if err != nil {
		return nil, err
	}

	return balances, nil
}
