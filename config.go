package mnee

import (
	"context"
	"encoding/json"
	"net/http"
)

func (m *MNEE) GetConfig(ctx context.Context) (*SystemConfig, error) {

	select {

	case <-ctx.Done():
		return nil, ctx.Err()

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

	newRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		(m.mneeURL + "/v1/config" + m.mneeToken),
		nil,
	)
	if err != nil {
		return nil, err
	}

	configResponse, err := m.httpClient.Do(newRequest)
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
