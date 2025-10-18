package mnee

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	primitives "github.com/bsv-blockchain/go-sdk/primitives/ec"
	"github.com/bsv-blockchain/go-sdk/script"
)

const (
	EnvMain    string = "MAIN"
	EnvSandbox string = "SANDBOX"
)

type MNEE struct {
	mneeURL     string
	mneeToken   string
	httpClient  *http.Client
	config      *SystemConfig
	configTimer <-chan time.Time
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
		return nil, errors.New("environment must be valid")
	}

	mnee.httpClient = &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			ForceAttemptHTTP2: false,
		},
		Timeout: 0,
	}
	mnee.configTimer = time.Tick(time.Hour)

	return &mnee, nil
}

func lock(a *script.Address, pubkey *primitives.PublicKey) (*script.Script, error) {

	if len(a.PublicKeyHash) != 20 {
		return nil, errors.New("invalid public key hash")
	}

	var s script.Script
	s.AppendOpcodes(script.OpDUP, script.OpHASH160)
	s.AppendPushData(a.PublicKeyHash)
	s.AppendOpcodes(script.OpEQUALVERIFY, script.OpCHECKSIGVERIFY)
	s.AppendPushData(pubkey.Compressed())
	s.AppendOpcodes(script.OpCHECKSIG)

	return &s, nil
}

func createTransferInscription(tokenID string, amt uint64) ([]byte, error) {

	var inscription map[string]string = make(map[string]string)
	inscription["p"] = "bsv-20"
	inscription["op"] = "transfer"
	inscription["id"] = tokenID
	inscription["amt"] = fmt.Sprintf("%d", amt)

	return json.Marshal(&inscription)
}
