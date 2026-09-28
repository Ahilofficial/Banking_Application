package models

import (
	"time"
)

type Notification struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       uint      `gorm:"index;not null" json:"user_id"`
	Type         string    `gorm:"size:32;not null;index" json:"type"` // EMAIL, SMS, PUSH, IN_APP
	Recipient    string    `gorm:"size:255;not null" json:"recipient"` // email, phone, token
	Subject      string    `gorm:"size:255" json:"subject,omitempty"`
	Content      string    `gorm:"type:text;not null" json:"content"`
	Status       string    `gorm:"size:32;not null;default:'SENT';index" json:"status"` // PENDING, SENT, DELIVERED, FAILED
	ErrorMessage string    `gorm:"size:512" json:"error_message,omitempty"`
	Metadata     string    `gorm:"type:text" json:"metadata,omitempty"` // JSON string
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
