package mnee

import (
	"context"
	"encoding/json"
	"net/http"
)

func (m *MNEE) GetConfig(ctx context.Context) (*SystemConfig, error) {

	m.mutex.Lock()
	defer m.mutex.Unlock()

	select {

	case <-ctx.Done():
		return nil, ctx.Err()

	case <-m.refreshTimer:
		break

	default:
		{
			if m.config != nil {
				return m.config, nil
			}
		}
	}

	configRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		(m.mneeURL + "/v1/config?auth_token" + m.mneeToken),
		nil,
	)
	if err != nil {
		return nil, err
	}

	configResponse, err := m.httpClient.Do(configRequest)
	if err != nil {
		return nil, err
	}

	defer configResponse.Body.Close()

	var systemConfig SystemConfig
	err = json.NewDecoder(configResponse.Body).Decode(&systemConfig)
	if err != nil {
		return nil, err
	}

	m.config = &systemConfig

	return &systemConfig, nil
}
