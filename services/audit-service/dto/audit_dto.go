package dto

import "time"

type CreateAuditLogRequest struct {
	UserID       uint   `json:"user_id"`
	Action       string `json:"action"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	OldValue     string `json:"old_value,omitempty"`
	NewValue     string `json:"new_value,omitempty"`
	IPAddress    string `json:"ip_address,omitempty"`
	UserAgent    string `json:"user_agent,omitempty"`
	RequestID    string `json:"request_id,omitempty"`
}

type AuditLogResponse struct {
	ID           uint      `json:"id"`
	UserID       uint      `json:"user_id"`
	Action       string    `json:"action"`
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id"`
	OldValue     string    `json:"old_value,omitempty"`
	NewValue     string    `json:"new_value,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	RequestID    string    `json:"request_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
