package mnee

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// GetBalances fetches the MNEE balance for a list of addresses.
// It returns a slice of BalanceDataDTO, one for each address.
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

	if balancesResponse.StatusCode == http.StatusForbidden {
		return nil, ErrForbidden
	}

	if balancesResponse.StatusCode != http.StatusOK {
		var errorResponse map[string]any
		err = json.NewDecoder(balancesResponse.Body).Decode(&errorResponse)
		if err != nil {
			return nil, err
		}

		errorMessage, ok := errorResponse["message"].(string)
		if !ok {
			return nil, fmt.Errorf("status received from mnee-cosigner -> %d", balancesResponse.StatusCode)
		}

		return nil, errors.New(errorMessage)
	}

	err = json.NewDecoder(balancesResponse.Body).Decode(&balances)
	if err != nil {
		return nil, err
	}

	return balances, nil
}
