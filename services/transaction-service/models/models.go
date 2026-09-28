package models

import (
	"time"

	"banking-microservices/pkg/common/money"
)

type TransactionType string

const (
	TxTypeTransfer   TransactionType = "TRANSFER"
	TxTypeDeposit    TransactionType = "DEPOSIT"
	TxTypeWithdrawal TransactionType = "WITHDRAWAL"
	TxTypeReversal   TransactionType = "REVERSAL"
)

type TransactionStatus string

const (
	TxStatusPending    TransactionStatus = "PENDING"
	TxStatusProcessing TransactionStatus = "PROCESSING"
	TxStatusSuccess    TransactionStatus = "SUCCESS"
	TxStatusFailed     TransactionStatus = "FAILED"
	TxStatusReversed   TransactionStatus = "REVERSED"
)

type LedgerEntryType string

const (
	LedgerDebit  LedgerEntryType = "DEBIT"
	LedgerCredit LedgerEntryType = "CREDIT"
)

// Transaction represents a financial movement record
type Transaction struct {
	ID              uint              `gorm:"primaryKey;autoIncrement" json:"id"`
	ReferenceNumber string            `gorm:"uniqueIndex;not null;size:64" json:"reference_number"`
	FromAccountID   *uint             `gorm:"index" json:"from_account_id,omitempty"`
	ToAccountID     *uint             `gorm:"index" json:"to_account_id,omitempty"`
	Amount          money.Amount      `gorm:"not null" json:"amount"` // minor units (paise/cents)
	Currency        string            `gorm:"size:10;default:'INR'" json:"currency"`
	TransactionType TransactionType   `gorm:"size:30;not null" json:"transaction_type"`
	Status          TransactionStatus `gorm:"size:30;not null;default:'PENDING'" json:"status"`
	Description     string            `gorm:"size:255" json:"description"`
	FailureReason   string            `gorm:"size:255" json:"failure_reason,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	CompletedAt     *time.Time        `json:"completed_at,omitempty"`
}

// LedgerEntry represents a double-entry bookkeeping journal line
type LedgerEntry struct {
	ID            uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	TransactionID uint            `gorm:"index;not null" json:"transaction_id"`
	AccountID     uint            `gorm:"index;not null" json:"account_id"`
	EntryType     LedgerEntryType `gorm:"size:10;not null" json:"entry_type"` // DEBIT or CREDIT
	Amount        money.Amount    `gorm:"not null" json:"amount"`
	Currency      string          `gorm:"size:10;default:'INR'" json:"currency"`
	BalanceAfter  money.Amount    `gorm:"not null" json:"balance_after"`
	CreatedAt     time.Time       `json:"created_at"`
}

type IdempotencyStatus string

const (
	IdempotencyProcessing IdempotencyStatus = "PROCESSING"
	IdempotencySuccess    IdempotencyStatus = "SUCCESS"
	IdempotencyFailed     IdempotencyStatus = "FAILED"
)

// IdempotencyKey prevents duplicate financial operations
type IdempotencyKey struct {
	ID          uint              `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint              `gorm:"index;not null" json:"user_id"`
	Key         string            `gorm:"column:idempotency_key;uniqueIndex:idx_user_idemp_key,length:128;size:128;not null" json:"key"`
	RequestHash string            `gorm:"size:64;not null" json:"request_hash"`
	Response    string            `gorm:"type:text" json:"response"`
	Status      IdempotencyStatus `gorm:"size:20;not null;default:'PROCESSING'" json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}
