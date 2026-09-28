package models

import "time"

// AuditLog provides an immutable banking audit trail
type AuditLog struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       uint      `gorm:"index" json:"user_id"`
	Action       string    `gorm:"index;size:100;not null" json:"action"`
	ResourceType string    `gorm:"index;size:100;not null" json:"resource_type"`
	ResourceID   string    `gorm:"index;size:100;not null" json:"resource_id"`
	OldValue     string    `gorm:"type:text" json:"old_value,omitempty"`
	NewValue     string    `gorm:"type:text" json:"new_value,omitempty"`
	IPAddress    string    `gorm:"size:45" json:"ip_address,omitempty"`
	UserAgent    string    `gorm:"size:255" json:"user_agent,omitempty"`
	RequestID    string    `gorm:"index;size:64" json:"request_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
