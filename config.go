package mnee

import (
	"encoding/json"
	"net/http"
)

func (m *MNEE) GetConfig() (*SystemConfig, error) {

	select {

	case <-m.configTimer:
		{
			m.config = nil
		}

	default:
		{
			if m.config != nil {
				return m.config, nil
			}
		}
	}

	var newClient *http.Client = &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			ForceAttemptHTTP2: false,
		},
		Timeout: 0,
	}

	newRequest, err := http.NewRequest(http.MethodGet, (m.mneeURL + "/v1/config" + m.mneeToken), nil)
	if err != nil {
		return nil, err
	}

	configResponse, err := newClient.Do(newRequest)
	if err != nil {
		return nil, err
	}

	defer configResponse.Body.Close()

	var systemConfig SystemConfig
	err = json.NewDecoder(configResponse.Body).Decode(&systemConfig)
	if err != nil {
		return nil, err
	}

	return &systemConfig, nil
}
