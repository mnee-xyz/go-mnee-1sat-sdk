package mnee

import (
	"github.com/lib/pq"
)

type Fee struct {
	MinAmt uint64 `json:"min"`
	MaxAmt uint64 `json:"max"`
	Fee    uint64 `json:"fee"`
}

type SystemConfig struct {
	Decimals    uint8   `json:"decimals"`
	Approver    *string `json:"approver,omitempty"`
	FeeAddress  *string `json:"feeAddress,omitempty"`
	BurnAddress *string `json:"burnAddress,omitempty"`
	MintAddress *string `json:"mintAddress,omitempty"`
	TokenId     *string `json:"tokenId,omitempty"`
	Fees        []Fee   `json:"fees,omitempty"`
}

type BalanceData struct {
	Amt      float64 `json:"amt"`
	Precised float64 `json:"precised"`
	Address  *string `json:"address"`
}

type TransferMneeDTO struct {
	Amount  uint64 `json:"amount"`
	Address string `json:"address,omitempty"`
}

type BsvData struct {
	Decimals uint8   `json:"dec"`
	Amt      uint64  `json:"amt"`
	Id       *string `json:"id,omitempty"`
	Op       *string `json:"op,omitempty"`
	Symbol   *string `json:"sym,omitempty"`
	Icon     *string `json:"icon,omitempty"`
}

type CosignData struct {
	Address  *string `json:"address,omitempty"`
	Cosigner *string `json:"cosigner,omitempty"`
}

type Data struct {
	Bsv21  *BsvData    `json:"bsv21,omitempty"`
	Cosign *CosignData `json:"cosign,omitempty"`
}

type MneeTxo struct {
	Satoshis uint16         `json:"satoshis,omitempty"`
	Height   uint64         `json:"height"`
	Idx      uint64         `json:"idx"`
	Score    uint64         `json:"score"`
	Vout     uint64         `json:"vout"`
	Outpoint *string        `json:"outpoint,omitempty"`
	Script   *string        `json:"script,omitempty"`
	Txid     *string        `json:"txid,omitempty"`
	Data     *Data          `json:"data,omitempty"`
	Owners   pq.StringArray `json:"owners,omitempty"`
	Senders  pq.StringArray `json:"senders,omitempty"`
}
