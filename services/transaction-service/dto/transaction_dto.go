package dto

import (
	"time"

	"banking-microservices/services/transaction-service/models"
)

type TransferRequest struct {
	FromAccountID uint    `json:"from_account_id"`
	ToAccountID   uint    `json:"to_account_id"`
	Amount        float64 `json:"amount"` // e.g. 10000.00
	Currency      string  `json:"currency"`
	Description   string  `json:"description"`
}

type DepositRequest struct {
	AccountID   uint    `json:"account_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
}

type WithdrawRequest struct {
	AccountID   uint    `json:"account_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
}

type TransactionResponse struct {
	ID              uint                     `json:"id"`
	ReferenceNumber string                   `json:"reference_number"`
	FromAccountID   *uint                    `json:"from_account_id,omitempty"`
	ToAccountID     *uint                    `json:"to_account_id,omitempty"`
	Amount          float64                  `json:"amount"`
	AmountMinor     int64                    `json:"amount_minor_units"`
	AmountDisplay   string                   `json:"amount_display"`
	Currency        string                   `json:"currency"`
	TransactionType models.TransactionType   `json:"transaction_type"`
	Status          models.TransactionStatus `json:"status"`
	Description     string                   `json:"description"`
	CreatedAt       time.Time                `json:"created_at"`
	CompletedAt     *time.Time               `json:"completed_at,omitempty"`
}

type LedgerEntryResponse struct {
	ID            uint                   `json:"id"`
	TransactionID uint                   `json:"transaction_id"`
	AccountID     uint                   `json:"account_id"`
	EntryType     models.LedgerEntryType `json:"entry_type"`
	Amount        float64                `json:"amount"`
	AmountMinor   int64                  `json:"amount_minor_units"`
	BalanceAfter  float64                `json:"balance_after"`
	Currency      string                 `json:"currency"`
	CreatedAt     time.Time              `json:"created_at"`
}
