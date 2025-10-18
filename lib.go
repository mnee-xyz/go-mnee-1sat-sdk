package mnee

import (
	"errors"
	"net/http"
	"sync"
	"time"
)

const (
	EnvMain    string = "MAIN"
	EnvSandbox string = "SANDBOX"
)

type MNEE struct {
	mneeURL      string
	mneeToken    string
	mutex        *sync.Mutex
	httpClient   *http.Client
	config       *SystemConfig
	refreshTimer <-chan time.Time
}

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
		return nil, errors.New("invalid environment")
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
