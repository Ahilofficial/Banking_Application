package handlers

import (
	"strconv"

	"banking-microservices/pkg/common/errors"
	"banking-microservices/pkg/common/response"
	"banking-microservices/pkg/common/security"
	"banking-microservices/services/account-service/dto"
	"banking-microservices/services/account-service/models"
	"banking-microservices/services/account-service/service"

	"github.com/gofiber/fiber/v3"
)

type AccountHandler struct {
	svc *service.AccountService
}

func NewAccountHandler(svc *service.AccountService) *AccountHandler {
	return &AccountHandler{svc: svc}
}

func (h *AccountHandler) Create(c fiber.Ctx) error {
	var req dto.CreateAccountRequest
	if err := c.Bind().Body(&req); err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, "Invalid account payload")
	}

	callerUserID := security.GetUserID(c)
	reqID := security.GetRequestID(c)
	ip := c.IP()

	account, err := h.svc.Create(req, callerUserID, reqID, ip)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusCreated, account, "Account created successfully")
}

func (h *AccountHandler) GetByID(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, "Invalid account ID")
	}

	callerUserID := security.GetUserID(c)
	hasGlobalView := security.HasPermission(c, "account:view") && (containsRole(c, "SUPER_ADMIN") || containsRole(c, "BANK_ADMIN") || containsRole(c, "TELLER") || containsRole(c, "BRANCH_MANAGER"))

	account, err := h.svc.GetByID(uint(id), callerUserID, hasGlobalView)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, account)
}

func (h *AccountHandler) GetByAccountNumber(c fiber.Ctx) error {
	accNum := c.Params("account_number")
	callerUserID := security.GetUserID(c)
	hasGlobalView := security.HasPermission(c, "account:view") && (containsRole(c, "SUPER_ADMIN") || containsRole(c, "BANK_ADMIN") || containsRole(c, "TELLER") || containsRole(c, "BRANCH_MANAGER"))

	account, err := h.svc.GetByAccountNumber(accNum, callerUserID, hasGlobalView)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, account)
}

func (h *AccountHandler) List(c fiber.Ctx) error {
	callerUserID := security.GetUserID(c)
	hasGlobalView := containsRole(c, "SUPER_ADMIN") || containsRole(c, "BANK_ADMIN") || containsRole(c, "TELLER") || containsRole(c, "BRANCH_MANAGER")

	customerIDStr := c.Query("customer_id")
	userIDStr := c.Query("user_id")

	var customerID, userID uint
	if customerIDStr != "" {
		if cid, err := strconv.ParseUint(customerIDStr, 10, 64); err == nil {
			customerID = uint(cid)
		}
	}
	if userIDStr != "" {
		if uid, err := strconv.ParseUint(userIDStr, 10, 64); err == nil {
			userID = uint(uid)
		}
	}

	limitStr := c.Query("limit", "20")
	offsetStr := c.Query("offset", "0")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	accounts, err := h.svc.List(customerID, userID, callerUserID, hasGlobalView, limit, offset)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, accounts)
}

func (h *AccountHandler) Block(c fiber.Ctx) error {
	return h.changeStatus(c, models.AccountStatusBlocked)
}

func (h *AccountHandler) Unblock(c fiber.Ctx) error {
	return h.changeStatus(c, models.AccountStatusActive)
}

func (h *AccountHandler) Close(c fiber.Ctx) error {
	return h.changeStatus(c, models.AccountStatusClosed)
}

func (h *AccountHandler) changeStatus(c fiber.Ctx, status models.AccountStatus) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, "Invalid account ID")
	}

	callerUserID := security.GetUserID(c)
	reqID := security.GetRequestID(c)
	ip := c.IP()

	account, err := h.svc.UpdateStatus(uint(id), status, callerUserID, reqID, ip)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, account, "Account status updated to "+string(status))
}

func (h *AccountHandler) InternalVerify(c fiber.Ctx) error {
	var req dto.VerifyAccountRequest
	if err := c.Bind().Body(&req); err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, "Invalid payload")
	}

	res, err := h.svc.VerifyAccount(req.AccountID, req.UserID)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, res)
}

func containsRole(c fiber.Ctx, role string) bool {
	for _, r := range security.GetUserRoles(c) {
		if r == role {
			return true
		}
	}
	return false
}
