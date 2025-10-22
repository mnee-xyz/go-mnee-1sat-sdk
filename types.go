package mnee

import (
	"time"
)

type TokenOperation string
type TokenProtocol string
type TicketStatus string

const (
	TRANSFER    TokenOperation = "transfer"
	DEPLOY_MINT TokenOperation = "deploy+mint"
)

const (
	BROADCASTING = "BROADCASTING"
	SUCCESS      = "SUCCESS"
)

const (
	ACTION_DEPLOY   string = "deploy"
	ACTION_MINT     string = "mint"
	ACTION_TRANSFER string = "transfer"
	ACTION_REDEEM   string = "redeem"
)

const (
	BSV20 TokenProtocol = "bsv-20"
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

type Ticket struct {
	ID              *string      `json:"id,omitempty"`
	TxID            *string      `json:"tx_id,omitempty"`
	TxHex           *string      `json:"tx_hex,omitempty"`
	ActionRequested *string      `json:"action_requested,omitempty"`
	CallbackURL     *string      `json:"callback_url,omitempty"`
	CallbackSecret  *string      `json:"callback_secret,omitempty"`
	Status          TicketStatus `json:"status,omitempty,omitzero"`
	CreatedAt       *time.Time   `json:"createdAt,omitempty"`
	UpdatedAt       *time.Time   `json:"updatedAt,omitempty"`
	Errors          []string     `json:"errors"`
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
	Satoshis uint16   `json:"satoshis,omitempty"`
	Height   uint64   `json:"height"`
	Idx      uint64   `json:"idx"`
	Score    uint64   `json:"score"`
	Vout     uint64   `json:"vout"`
	Outpoint *string  `json:"outpoint,omitempty"`
	Script   *string  `json:"script,omitempty"`
	Txid     *string  `json:"txid,omitempty"`
	Data     *Data    `json:"data,omitempty"`
	Owners   []string `json:"owners,omitempty"`
	Senders  []string `json:"senders,omitempty"`
}

type TransferMneeDTO struct {
	Amount  uint64 `json:"amount"`
	Address string `json:"address,omitempty"`
}

type TransferRequestDTO struct {
	RawTx          string  `json:"rawtx,omitempty"`
	CallbackURL    *string `json:"callback_url,omitempty"`
	CallbackSecret *string `json:"callback_secret,omitempty"`
}

type TransferResponseDTO struct {
	Txid  *string `json:"txid,omitempty"`
	Txhex *string `json:"txhex,omitempty"`
}

type TransactionHistoryDTO struct {
	Height    uint64   `json:"height"`
	Idx       uint64   `json:"idx"`
	Score     uint64   `json:"score"`
	Rawtx     *string  `json:"rawtx,omitempty"`
	Txid      *string  `json:"txid,omitempty"`
	Outs      []uint64 `json:"outs,omitempty"`
	Senders   []string `json:"senders,omitempty"`
	Receivers []string `json:"receivers,omitempty"`
}

type BalanceDataDTO struct {
	Amt      float64 `json:"amt"`
	Precised float64 `json:"precised"`
	Address  *string `json:"address"`
}

type BaseTokenInscription struct {
	Protocol  TokenProtocol  `json:"p"`
	Amount    string         `json:"amt"`
	Operation TokenOperation `json:"op"`
	Decimal   string         `json:"dec"`
}

type TokenMetadata struct {
	CurrentSupply string `json:"currentSupply"`
	Action        string `json:"action"`
	Version       string `json:"version"`
}

type DeployChainInscription struct {
	BaseTokenInscription
	TokenID  string         `json:"id"`
	Metadata *TokenMetadata `json:"metadata,omitempty"`
}

type TransferTokenInscription struct {
	BaseTokenInscription
	TokenID string `json:"id"`
}
