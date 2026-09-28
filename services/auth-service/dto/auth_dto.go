package dto

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Role     string `json:"role"` // Optional, defaults to "CUSTOMER"
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AssignRoleRequest struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
}

type IntrospectRequest struct {
	Token string `json:"token"`
}

type UserResponse struct {
	ID          uint     `json:"id"`
	Email       string   `json:"email"`
	FullName    string   `json:"full_name"`
	Status      string   `json:"status"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

type TokenResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
	User         UserResponse `json:"user"`
}
