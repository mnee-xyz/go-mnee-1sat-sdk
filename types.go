package mnee

import (
	"errors"
	"time"
)

// ErrForbidden is returned when the MNEE API responds with a 403 Forbidden status.
// This almost always indicates an invalid or missing API key.
var ErrForbidden = errors.New("forbidden access to cosigner")

// ErrInvalidConfig is returned by functions when the fetched
// MNEE system config is missing required fields (like Approver or TokenId).
var ErrInvalidConfig = errors.New("invalid config")

// ErrInvalidEnvironment is returned by NewMneeInstance if the
// environment string is not 'MAIN' or 'SANDBOX'.
var ErrInvalidEnvironment = errors.New("invalid environment")

// ErrInsufficientMneeBalance is returned by transfer, partial sign functions when the
// wallet's UTXOs do not have enough MNEE tokens to cover the transfer amount + fee.
var ErrInsufficientMneeBalance = errors.New("insufficient mnee balance")

// ErrTransferAmountGreaterThan0 is returned by transfer, partial sign functions
// if any recipient amount is 0.
var ErrTransferAmountGreaterThan0 = errors.New("transfer amount must be greater than 0")

// ErrInvalidPublicKeyHash is returned by the internal 'lock' function
// if the provided address's public key hash is not 20 bytes.
var ErrInvalidPublicKeyHash = errors.New("invalid public key hash")

// ErrReceivedEmptyTicketID is returned by AsynchronousTransfer if the
// API returns a 200 OK but the response body (ticket ID) is empty.
var ErrReceivedEmptyTicketID = errors.New("received an empty ticket ID from server")

// TokenOperation defines the type of BSV-20 operation.
type TokenOperation string

// TokenProtocol defines the token protocol name.
type TokenProtocol string

// TicketStatus defines the status of an asynchronous transfer.
type TicketStatus string

const (
	// TRANSFER is the "transfer" operation for BSV-20.
	TRANSFER TokenOperation = "transfer"
	// DEPLOY_MINT is the "deploy+mint" operation for BSV-20.
	DEPLOY_MINT TokenOperation = "deploy+mint"
)

const (
	// BROADCASTING indicates the transaction is in the mempool.
	BROADCASTING TicketStatus = "BROADCASTING"
	// SUCCESS indicates the transaction has been confirmed.
	SUCCESS TicketStatus = "SUCCESS"
)

const (
	// ACTION_DEPLOY is the "deploy" metadata action.
	ACTION_DEPLOY string = "deploy"
	// ACTION_MINT is the "mint" metadata action.
	ACTION_MINT string = "mint"
	// ACTION_TRANSFER is the "transfer" metadata action.
	ACTION_TRANSFER string = "transfer"
	// ACTION_REDEEM is the "redeem" metadata action.
	ACTION_REDEEM string = "redeem"
)

const (
	// BSV20 is the protocol name "bsv-20".
	BSV20 TokenProtocol = "bsv-20"
)

// Fee represents a single fee tier from the MNEE config.
type Fee struct {
	MinAmt uint64 `json:"min"`
	MaxAmt uint64 `json:"max"`
	Fee    uint64 `json:"fee"`
}

// SystemConfig holds the MNEE system configuration, fetched from the API.
type SystemConfig struct {
	Decimals    uint8   `json:"decimals"`
	Approver    *string `json:"approver,omitempty"`
	FeeAddress  *string `json:"feeAddress,omitempty"`
	BurnAddress *string `json:"burnAddress,omitempty"`
	MintAddress *string `json:"mintAddress,omitempty"`
	TokenId     *string `json:"tokenId,omitempty"`
	Fees        []Fee   `json:"fees,omitempty"`
}

// Ticket represents the response for an asynchronous transfer,
// used for polling its status.
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

// BsvData holds BSV-21 specific data from a UTXO.
type BsvData struct {
	Decimals uint8   `json:"dec"`
	Amt      uint64  `json:"amt"`
	Id       *string `json:"id,omitempty"`
	Op       *string `json:"op,omitempty"`
	Symbol   *string `json:"sym,omitempty"`
	Icon     *string `json:"icon,omitempty"`
}

// CosignData holds cosigner-specific data from a UTXO.
type CosignData struct {
	Address  *string `json:"address,omitempty"`
	Cosigner *string `json:"cosigner,omitempty"`
}

// Data is a container for UTXO data, including BSV-21 and cosigner info.
type Data struct {
	Bsv21  *BsvData    `json:"bsv21,omitempty"`
	Cosign *CosignData `json:"cosign,omitempty"`
}

// MneeTxo represents a single MNEE UTXO.
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

// TransferMneeDTO defines a single recipient for a transfer.
type TransferMneeDTO struct {
	Amount  uint64 `json:"amount"`
	Address string `json:"address,omitempty"`
}

// TransferRequestDTO is the JSON body for the asynchronous transfer request.
type TransferRequestDTO struct {
	RawTx          string  `json:"rawtx,omitempty"`
	CallbackURL    *string `json:"callback_url,omitempty"`
	CallbackSecret *string `json:"callback_secret,omitempty"`
}

// TransferResponseDTO is the successful response from a SynchronousTransfer.
type TransferResponseDTO struct {
	Txid  *string `json:"txid,omitempty"`
	Txhex *string `json:"txhex,omitempty"`
}

// TransactionHistoryDTO represents a single item in the transaction history.
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

// BalanceDataDTO represents the MNEE balance for a single address.
type BalanceDataDTO struct {
	Amt      float64 `json:"amt"`
	Precised float64 `json:"precised"`
	Address  *string `json:"address"`
}

// BaseTokenInscription defines the common fields for BSV-20 inscriptions.
type BaseTokenInscription struct {
	Protocol  TokenProtocol  `json:"p"`
	Amount    string         `json:"amt"`
	Operation TokenOperation `json:"op"`
	Decimal   string         `json:"dec"`
}

// TokenMetadata defines the "metadata" field within a deploy inscription.
type TokenMetadata struct {
	CurrentSupply string `json:"currentSupply"`
	Action        string `json:"action"`
	Version       string `json:"version"`
}

// DeployChainInscription represents a "deploy+mint" BSV-20 inscription.
type DeployChainInscription struct {
	BaseTokenInscription
	TokenID  string         `json:"id"`
	Metadata *TokenMetadata `json:"metadata,omitempty"`
}

// TransferTokenInscription represents a "transfer" BSV-20 inscription.
type TransferTokenInscription struct {
	BaseTokenInscription
	TokenID string `json:"id"`
}
