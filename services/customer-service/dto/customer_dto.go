package dto

import (
	"banking-microservices/services/customer-service/models"
	"time"
)

type CreateCustomerRequest struct {
	UserID      uint   `json:"user_id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
	Address     string `json:"address"`
	NationalID  string `json:"national_id"`
}

type UpdateCustomerRequest struct {
	FirstName   *string `json:"first_name,omitempty"`
	LastName    *string `json:"last_name,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty"`
	Address     *string `json:"address,omitempty"`
}

type UpdateKYCRequest struct {
	KYCStatus models.KYCStatus `json:"kyc_status"`
	Reason    string           `json:"reason,omitempty"`
}

type CustomerResponse struct {
	ID          uint             `json:"id"`
	UserID      uint             `json:"user_id"`
	FirstName   string           `json:"first_name"`
	LastName    string           `json:"last_name"`
	PhoneNumber string           `json:"phone_number"`
	Address     string           `json:"address"`
	NationalID  string           `json:"national_id"`
	KYCStatus   models.KYCStatus `json:"kyc_status"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}
