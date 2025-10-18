package mnee

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func (m *MNEE) GetMneeBalance(addresses []string) ([]BalanceData, error) {

	addressesBuffer, err := json.Marshal(&addresses)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(
		http.MethodPost,
		(m.mneeURL + "/v2/balance?auth_token=" + m.mneeToken),
		bytes.NewBuffer(addressesBuffer),
	)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")

	var newClient *http.Client = &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			ForceAttemptHTTP2: false,
		},
		Timeout: 0,
	}

	var balances []BalanceData = make([]BalanceData, 0)
	response, err := newClient.Do(request)
	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(response.Body).Decode(&balances)
	if err != nil {
		return nil, err
	}

	return balances, nil
}
