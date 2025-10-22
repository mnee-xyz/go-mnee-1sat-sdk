package mnee

import (
	"context"
	"encoding/json"
	"errors"
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

		if ticketResponse.StatusCode > 299 {
			var errorResponse struct {
				Message string `json:"message,omitempty"`
			}
			err = json.NewDecoder(ticketResponse.Body).Decode(&errorResponse)
			if err != nil {
				return nil, err
			}

			if errorResponse.Message == "record not found" {
				time.Sleep(pollingInterval)
				continue
			} else {
				return nil, errors.New(errorResponse.Message)
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
