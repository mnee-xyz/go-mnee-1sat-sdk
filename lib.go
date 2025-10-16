package mnee

import (
	"encoding/json"
	"errors"
	"fmt"

	primitives "github.com/bsv-blockchain/go-sdk/primitives/ec"
	"github.com/bsv-blockchain/go-sdk/script"
)

type MNEE struct {
	mneeURL       string
	mneeToken     string
	wif           string
	senderAddress string
	config        *SystemConfig
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
