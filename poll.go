package mnee

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

func (m *MNEE) PollTicket(ctx context.Context, ticketID string, pollingInterval time.Duration) (*Ticket, error) {

	for {
		ticketRequest, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			(m.mneeURL + "/v2/ticket?auth_token=" + m.mneeToken + "&ticketID=" + ticketID),
			nil,
		)
		if err != nil {
			return nil, err
		}

		ticketResponse, err := m.httpClient.Do(ticketRequest)
		if err != nil {
			return nil, err
		}

		defer ticketResponse.Body.Close()

		if ticketResponse.StatusCode == http.StatusForbidden {
			return nil, errors.New("forbidden access to cosigner")
		}

		if ticketResponse.StatusCode != http.StatusOK {
			var errorResponse map[string]any
			err = json.NewDecoder(ticketResponse.Body).Decode(&errorResponse)
			if err != nil {
				return nil, err
			}

			errorMessage, ok := errorResponse["message"].(string)
			if !ok {
				return nil, fmt.Errorf("status received from mnee-cosigner -> %d", ticketResponse.StatusCode)
			}

			if errorMessage == "record not found" {
				time.Sleep(pollingInterval)
				continue
			} else {
				return nil, errors.New(errorMessage)
			}
		}

		var ticket Ticket
		err = json.NewDecoder(ticketResponse.Body).Decode(&ticket)
		if err != nil {
			return nil, err
		}

		return &ticket, nil
	}
}
