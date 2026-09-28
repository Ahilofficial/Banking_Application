package dto

import (
	"time"

	"banking-microservices/services/account-service/models"
)

type CreateAccountRequest struct {
	CustomerID     uint               `json:"customer_id"`
	UserID         uint               `json:"user_id"` // Owner
	BranchCode     string             `json:"branch_code"`
	AccountType    models.AccountType `json:"account_type"`
	Currency       string             `json:"currency"`
	InitialDeposit float64            `json:"initial_deposit"` // e.g. 50000.00
}

type AccountResponse struct {
	ID             uint                 `json:"id"`
	AccountNumber  string               `json:"account_number"`
	CustomerID     uint                 `json:"customer_id"`
	UserID         uint                 `json:"user_id"`
	BranchCode     string               `json:"branch_code"`
	AccountType    models.AccountType   `json:"account_type"`
	Currency       string               `json:"currency"`
	Balance        float64              `json:"balance"` // Decimal formatted for display
	BalancePaise   int64                `json:"balance_minor_units"`
	BalanceDisplay string               `json:"balance_display"`
	Status         models.AccountStatus `json:"status"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

type UpdateStatusRequest struct {
	Status models.AccountStatus `json:"status"`
	Reason string               `json:"reason,omitempty"`
}

type VerifyAccountRequest struct {
	AccountID uint `json:"account_id"`
	UserID    uint `json:"user_id"` // Optional ownership verification
}

type VerifyAccountResponse struct {
	Valid         bool                 `json:"valid"`
	AccountID     uint                 `json:"account_id"`
	AccountNumber string               `json:"account_number"`
	CustomerID    uint                 `json:"customer_id"`
	UserID        uint                 `json:"user_id"`
	Balance       int64                `json:"balance_minor_units"`
	Status        models.AccountStatus `json:"status"`
}
