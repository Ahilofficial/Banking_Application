package models

import (
	"time"

	"banking-microservices/pkg/common/money"
)

type AccountType string

const (
	AccountTypeSavings  AccountType = "SAVINGS"
	AccountTypeChecking AccountType = "CHECKING"
	AccountTypeBusiness AccountType = "BUSINESS"
)

type AccountStatus string

const (
	AccountStatusActive  AccountStatus = "ACTIVE"
	AccountStatusBlocked AccountStatus = "BLOCKED"
	AccountStatusClosed  AccountStatus = "CLOSED"
)

type Account struct {
	ID            uint          `gorm:"primaryKey;autoIncrement" json:"id"`
	AccountNumber string        `gorm:"uniqueIndex;not null;size:50" json:"account_number"`
	CustomerID    uint          `gorm:"index;not null" json:"customer_id"`
	UserID        uint          `gorm:"index;not null" json:"user_id"` // Owner user ID for Level 2 auth
	BranchCode    string        `gorm:"size:20;not null" json:"branch_code"`
	AccountType   AccountType   `gorm:"size:20;not null" json:"account_type"`
	Currency      string        `gorm:"size:10;default:'INR'" json:"currency"`
	Balance       money.Amount  `gorm:"not null;default:0" json:"balance"` // In minor units (paise/cents)
	Status        AccountStatus `gorm:"size:20;default:'ACTIVE'" json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}
