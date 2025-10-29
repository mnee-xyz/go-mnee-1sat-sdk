package mnee

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// PollTicket polls the MNEE API for the status of an asynchronous transfer ticket.
// It will continue to poll at the specified `pollingInterval` until the context
// is canceled or the ticket status is no longer "record not found".
// It returns the final Ticket details.
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
			return nil, ErrForbidden
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
				select {
				case <-time.After(pollingInterval):
					continue
				case <-ctx.Done():
					return nil, ctx.Err()
				}
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
