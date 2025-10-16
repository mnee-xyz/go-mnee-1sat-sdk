package mnee

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func (m *MNEE) GetUTXOS() ([]MneeTxo, error) {

	request, err := http.NewRequest(
		http.MethodPost,
		(m.mneeURL + "/v1/utxos?auth_token=" + m.mneeToken),
		bytes.NewBuffer(fmt.Appendf(nil, "[\"%s\"]", m.senderAddress)),
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

	var txos []MneeTxo = make([]MneeTxo, 0)
	response, err := newClient.Do(request)
	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(response.Body).Decode(&txos)
	if err != nil {
		return nil, err
	}

	return txos, nil
}
