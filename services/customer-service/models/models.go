package models

import (
	"time"
)

type KYCStatus string

const (
	KYCStatusPending  KYCStatus = "PENDING"
	KYCStatusVerified KYCStatus = "VERIFIED"
	KYCStatusRejected KYCStatus = "REJECTED"
)

type Customer struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	FirstName   string    `gorm:"size:100;not null" json:"first_name"`
	LastName    string    `gorm:"size:100;not null" json:"last_name"`
	PhoneNumber string    `gorm:"size:25" json:"phone_number"`
	Address     string    `gorm:"size:255" json:"address"`
	NationalID  string    `gorm:"size:50" json:"national_id"` // PAN, Aadhaar, SSN
	KYCStatus   KYCStatus `gorm:"default:'PENDING';size:20" json:"kyc_status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
