package models

import (
	"time"
)

// User represents an authenticable entity in the banking system
type User struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Email        string     `gorm:"uniqueIndex;not null;size:191" json:"email"`
	PasswordHash string     `gorm:"not null" json:"-"`
	FullName     string     `gorm:"size:255" json:"full_name"`
	Status       string     `gorm:"default:'ACTIVE';size:50" json:"status"` // ACTIVE, BLOCKED, SUSPENDED
	Roles        []Role     `gorm:"many2many:user_roles;" json:"roles,omitempty"`
	Sessions     []Session  `gorm:"foreignKey:UserID" json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
}

// Role represents a banking actor (e.g. SUPER_ADMIN, BANK_ADMIN, BRANCH_MANAGER, BANK_EMPLOYEE, TELLER, CUSTOMER)
type Role struct {
	ID          uint         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string       `gorm:"uniqueIndex;not null;size:100" json:"name"`
	Description string       `gorm:"size:255" json:"description"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// Permission represents a granular action permission (e.g. "account:create", "transaction:create")
type Permission struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null;size:100" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserRole links Users to Roles (many-to-many)
type UserRole struct {
	UserID uint `gorm:"primaryKey" json:"user_id"`
	RoleID uint `gorm:"primaryKey" json:"role_id"`
}

// RolePermission links Roles to Permissions (many-to-many)
type RolePermission struct {
	RoleID       uint `gorm:"primaryKey" json:"role_id"`
	PermissionID uint `gorm:"primaryKey" json:"permission_id"`
}

// Session tracks server-side sessions for revokable access/refresh tokens
type Session struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       uint      `gorm:"index;not null" json:"user_id"`
	RefreshToken string    `gorm:"uniqueIndex;not null;size:255" json:"refresh_token"`
	IsRevoked    bool      `gorm:"default:false" json:"is_revoked"`
	IPAddress    string    `gorm:"size:45" json:"ip_address"`
	UserAgent    string    `gorm:"size:255" json:"user_agent"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
