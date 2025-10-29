package mnee

import (
	"net/http"
	"sync"
	"time"
)

const (
	// EnvMain specifies the MNEE production environment URL.
	EnvMain string = "MAIN"
	// EnvSandbox specifies the MNEE sandbox environment URL.
	EnvSandbox string = "SANDBOX"
)

// MNEE provides the client for interacting with the MNEE API.
// It holds the API configuration, HTTP client, and caches the system config.
type MNEE struct {
	mneeURL      string
	mneeToken    string
	mutex        *sync.Mutex
	httpClient   *http.Client
	config       *SystemConfig
	refreshTimer <-chan time.Time
}

// NewMneeInstance creates a new MNEE client instance.
//
// It requires an environment (`EnvMain` or `EnvSandbox`) and an authToken.
// The client automatically fetches and caches the MNEE system configuration.
func NewMneeInstance(environment string, authToken string) (*MNEE, error) {

	var mnee MNEE

	switch environment {

	case EnvMain:
		{
			mnee.mneeURL = "https://proxy-api.mnee.net"
			mnee.mneeToken = authToken
		}

	case EnvSandbox:
		{
			mnee.mneeURL = "https://sandbox-proxy-api.mnee.net"
			mnee.mneeToken = authToken
		}

	default:
		return nil, ErrInvalidEnvironment
	}

	mnee.mutex = new(sync.Mutex)
	mnee.httpClient = &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			ForceAttemptHTTP2: false,
		},
		Timeout: 0,
	}
	mnee.refreshTimer = time.Tick(time.Hour)

	return &mnee, nil
}
