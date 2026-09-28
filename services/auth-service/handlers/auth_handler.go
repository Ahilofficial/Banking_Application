package handlers

import (
	"strconv"

	"banking-microservices/pkg/common/errors"
	"banking-microservices/pkg/common/response"
	"banking-microservices/pkg/common/security"
	"banking-microservices/services/auth-service/dto"
	"banking-microservices/services/auth-service/service"

	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(c fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, "Invalid request payload")
	}

	reqID := security.GetRequestID(c)
	ip := c.IP()
	ua := c.Get("User-Agent")

	user, err := h.svc.Register(req, reqID, ip, ua)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusCreated, user, "User registered successfully")
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, "Invalid request payload")
	}

	reqID := security.GetRequestID(c)
	ip := c.IP()
	ua := c.Get("User-Agent")

	res, err := h.svc.Login(req, reqID, ip, ua)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, res, "Login successful")
}

func (h *AuthHandler) Refresh(c fiber.Ctx) error {
	var req dto.RefreshRequest
	if err := c.Bind().Body(&req); err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, "Invalid request payload")
	}

	reqID := security.GetRequestID(c)
	ip := c.IP()
	ua := c.Get("User-Agent")

	res, err := h.svc.Refresh(req, reqID, ip, ua)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, res, "Token refreshed successfully")
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	var req dto.LogoutRequest
	_ = c.Bind().Body(&req)

	if err := h.svc.Logout(req); err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, nil, "Logged out successfully")
}

func (h *AuthHandler) Me(c fiber.Ctx) error {
	userID := security.GetUserID(c)
	if userID == 0 {
		return errors.ErrUnauth("Authentication required")
	}

	user, err := h.svc.GetUser(userID)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, user)
}

func (h *AuthHandler) ListUsers(c fiber.Ctx) error {
	users, err := h.svc.ListUsers()
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, users)
}

func (h *AuthHandler) AssignRole(c fiber.Ctx) error {
	var req dto.AssignRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, "Invalid request payload")
	}

	adminID := security.GetUserID(c)
	reqID := security.GetRequestID(c)

	if err := h.svc.AssignRole(req, adminID, reqID); err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, nil, "Role assigned successfully")
}

func (h *AuthHandler) Introspect(c fiber.Ctx) error {
	var req dto.IntrospectRequest
	if err := c.Bind().Body(&req); err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, "Invalid request payload")
	}

	claims, err := h.svc.Introspect(req.Token)
	if err != nil {
		return errors.ErrUnauth("Token is invalid or expired")
	}

	return response.Success(c, fiber.StatusOK, fiber.Map{
		"active":      true,
		"user_id":     claims.UserID,
		"email":       claims.Email,
		"roles":       claims.Roles,
		"permissions": claims.Permissions,
	})
}

func (h *AuthHandler) GetUserByID(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, "Invalid user ID")
	}

	user, err := h.svc.GetUser(uint(id))
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, user)
}
