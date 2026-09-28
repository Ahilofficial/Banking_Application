package service

import (
	"errors"
	"fmt"
	"time"

	"banking-microservices/pkg/client"
	"banking-microservices/pkg/common/config"
	appErrors "banking-microservices/pkg/common/errors"
	"banking-microservices/pkg/common/security"
	"banking-microservices/services/auth-service/dto"
	"banking-microservices/services/auth-service/models"
	"banking-microservices/services/auth-service/repository"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet provides AuthService dependency for Wire
var ProviderSet = wire.NewSet(ProvideAuthService)

type AuthService struct {
	repo        *repository.AuthRepository
	jwtSecret   string
	accessTTL   time.Duration
	refreshTTL  time.Duration
	auditClient *client.AuditClient
}

func NewAuthService(
	repo *repository.AuthRepository,
	jwtSecret string,
	accessTTL, refreshTTL time.Duration,
	auditClient *client.AuditClient,
) *AuthService {
	return &AuthService{
		repo:        repo,
		jwtSecret:   jwtSecret,
		accessTTL:   accessTTL,
		refreshTTL:  refreshTTL,
		auditClient: auditClient,
	}
}

// ProvideAuthService initializes AuthService using injected configuration
func ProvideAuthService(
	repo *repository.AuthRepository,
	cfg *config.Config,
	auditClient *client.AuditClient,
) *AuthService {
	return NewAuthService(repo, cfg.JWTSecret, cfg.AccessTTL, cfg.RefreshTTL, auditClient)
}

func (s *AuthService) Register(req dto.RegisterRequest, reqID, ip, ua string) (*dto.UserResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, appErrors.ErrBadRequest(appErrors.ErrCodeValidationFailed, "Email and password are required")
	}

	// Check existing
	if existing, _ := s.repo.FindUserByEmail(req.Email); existing != nil {
		return nil, appErrors.ErrBadRequest(appErrors.ErrCodeUserAlreadyExists, "User with this email already exists")
	}

	hash, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, appErrors.ErrInternalServerError("Failed to hash password")
	}

	roleName := req.Role
	if roleName == "" {
		roleName = "CUSTOMER"
	}

	role, err := s.repo.FindRoleByName(roleName)
	if err != nil {
		return nil, appErrors.ErrBadRequest(appErrors.ErrCodeValidationFailed, "Specified role does not exist: "+roleName)
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: hash,
		FullName:     req.FullName,
		Status:       "ACTIVE",
		Roles:        []models.Role{*role},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, appErrors.ErrInternalServerError("Failed to create user record")
	}

	// Assign role relationship
	_ = s.repo.AssignRole(user.ID, role.ID)

	roles, _ := s.repo.GetRolesForUser(user.ID)
	permissions, _ := s.repo.GetPermissionsForUser(user.ID)

	s.auditClient.Log(client.AuditLogPayload{
		UserID:       user.ID,
		Action:       "USER_REGISTERED",
		ResourceType: "USER",
		ResourceID:   fmt.Sprintf("%d", user.ID),
		IPAddress:    ip,
		RequestID:    reqID,
	})

	return &dto.UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		FullName:    user.FullName,
		Status:      user.Status,
		Roles:       roles,
		Permissions: permissions,
	}, nil
}

func (s *AuthService) Login(req dto.LoginRequest, reqID, ip, ua string) (*dto.TokenResponse, error) {
	user, err := s.repo.FindUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrBadRequest(appErrors.ErrCodeInvalidCredentials, "Invalid email or password")
		}
		return nil, appErrors.ErrInternalServerError("Database lookup failed")
	}

	if user.Status != "ACTIVE" {
		return nil, appErrors.ErrForbid("Account is currently " + user.Status)
	}

	if !security.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, appErrors.ErrBadRequest(appErrors.ErrCodeInvalidCredentials, "Invalid email or password")
	}

	roles, err := s.repo.GetRolesForUser(user.ID)
	if err != nil {
		return nil, appErrors.ErrInternalServerError("Failed to retrieve user roles")
	}

	permissions, err := s.repo.GetPermissionsForUser(user.ID)
	if err != nil {
		return nil, appErrors.ErrInternalServerError("Failed to retrieve user permissions")
	}

	tokens, err := security.GenerateTokenPair(
		s.jwtSecret,
		user.ID,
		user.Email,
		roles,
		permissions,
		s.accessTTL,
		s.refreshTTL,
	)
	if err != nil {
		return nil, appErrors.ErrInternalServerError("Failed to generate authentication tokens")
	}

	// Persist server-side session for refresh token revocation
	session := &models.Session{
		UserID:       user.ID,
		RefreshToken: tokens.RefreshToken,
		IsRevoked:    false,
		IPAddress:    ip,
		UserAgent:    ua,
		ExpiresAt:    time.Now().Add(s.refreshTTL),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	_ = s.repo.CreateSession(session)

	now := time.Now()
	user.LastLoginAt = &now
	_ = s.repo.UpdateUser(user)

	s.auditClient.Log(client.AuditLogPayload{
		UserID:       user.ID,
		Action:       "USER_LOGIN",
		ResourceType: "SESSION",
		ResourceID:   fmt.Sprintf("%d", session.ID),
		IPAddress:    ip,
		RequestID:    reqID,
	})

	return &dto.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		User: dto.UserResponse{
			ID:          user.ID,
			Email:       user.Email,
			FullName:    user.FullName,
			Status:      user.Status,
			Roles:       roles,
			Permissions: permissions,
		},
	}, nil
}

func (s *AuthService) Refresh(req dto.RefreshRequest, reqID, ip, ua string) (*dto.TokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, appErrors.ErrBadRequest(appErrors.ErrCodeValidationFailed, "Refresh token is required")
	}

	session, err := s.repo.FindSession(req.RefreshToken)
	if err != nil || session.IsRevoked || time.Now().After(session.ExpiresAt) {
		return nil, appErrors.ErrUnauth("Invalid, revoked, or expired refresh token")
	}

	user, err := s.repo.FindUserByID(session.UserID)
	if err != nil {
		return nil, appErrors.ErrUnauth("User no longer exists")
	}

	// Revoke old session (Refresh Token Rotation)
	_ = s.repo.RevokeSession(session.RefreshToken)

	roles, _ := s.repo.GetRolesForUser(user.ID)
	permissions, _ := s.repo.GetPermissionsForUser(user.ID)

	tokens, err := security.GenerateTokenPair(
		s.jwtSecret,
		user.ID,
		user.Email,
		roles,
		permissions,
		s.accessTTL,
		s.refreshTTL,
	)
	if err != nil {
		return nil, appErrors.ErrInternalServerError("Token generation error")
	}

	newSession := &models.Session{
		UserID:       user.ID,
		RefreshToken: tokens.RefreshToken,
		IsRevoked:    false,
		IPAddress:    ip,
		UserAgent:    ua,
		ExpiresAt:    time.Now().Add(s.refreshTTL),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	_ = s.repo.CreateSession(newSession)

	return &dto.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		User: dto.UserResponse{
			ID:          user.ID,
			Email:       user.Email,
			FullName:    user.FullName,
			Status:      user.Status,
			Roles:       roles,
			Permissions: permissions,
		},
	}, nil
}

func (s *AuthService) Logout(req dto.LogoutRequest) error {
	if req.RefreshToken != "" {
		return s.repo.RevokeSession(req.RefreshToken)
	}
	return nil
}

func (s *AuthService) Introspect(token string) (*security.BankingClaims, error) {
	return security.ValidateToken(token, s.jwtSecret)
}

func (s *AuthService) GetUser(userID uint) (*dto.UserResponse, error) {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return nil, appErrors.ErrResourceNotFound(appErrors.ErrCodeNotFound, "User not found")
	}
	roles, _ := s.repo.GetRolesForUser(user.ID)
	permissions, _ := s.repo.GetPermissionsForUser(user.ID)

	return &dto.UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		FullName:    user.FullName,
		Status:      user.Status,
		Roles:       roles,
		Permissions: permissions,
	}, nil
}

func (s *AuthService) ListUsers() ([]dto.UserResponse, error) {
	users, err := s.repo.ListUsers()
	if err != nil {
		return nil, appErrors.ErrInternalServerError("Failed to list users")
	}

	var res []dto.UserResponse
	for _, u := range users {
		roles, _ := s.repo.GetRolesForUser(u.ID)
		perms, _ := s.repo.GetPermissionsForUser(u.ID)
		res = append(res, dto.UserResponse{
			ID:          u.ID,
			Email:       u.Email,
			FullName:    u.FullName,
			Status:      u.Status,
			Roles:       roles,
			Permissions: perms,
		})
	}
	return res, nil
}

func (s *AuthService) AssignRole(req dto.AssignRoleRequest, adminID uint, reqID string) error {
	role, err := s.repo.FindRoleByName(req.Role)
	if err != nil {
		return appErrors.ErrBadRequest(appErrors.ErrCodeNotFound, "Role not found")
	}

	if err := s.repo.AssignRole(req.UserID, role.ID); err != nil {
		return appErrors.ErrInternalServerError("Failed to assign role")
	}

	s.auditClient.Log(client.AuditLogPayload{
		UserID:       adminID,
		Action:       "ROLE_ASSIGNED",
		ResourceType: "USER",
		ResourceID:   fmt.Sprintf("%d", req.UserID),
		NewValue:     req.Role,
		RequestID:    reqID,
	})

	return nil
}
