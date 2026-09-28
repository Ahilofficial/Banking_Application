package handlers

import (
	"strconv"

	"banking-microservices/pkg/common/errors"
	"banking-microservices/pkg/common/response"
	"banking-microservices/pkg/common/security"
	"banking-microservices/services/transaction-service/dto"
	"banking-microservices/services/transaction-service/service"

	"github.com/gofiber/fiber/v3"
)

type TransactionHandler struct {
	svc *service.TransactionService
}

func NewTransactionHandler(svc *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{svc: svc}
}

func (h *TransactionHandler) Transfer(c fiber.Ctx) error {
	var req dto.TransferRequest
	if err := c.Bind().Body(&req); err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, "Invalid transfer payload")
	}

	callerUserID := security.GetUserID(c)
	isPrivileged := isStaffRole(c)
	idempotencyKey := c.Get(security.HeaderIdempotencyKey)
	reqID := security.GetRequestID(c)
	ip := c.IP()

	res, err := h.svc.Transfer(req, callerUserID, isPrivileged, idempotencyKey, reqID, ip)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, res, "Transfer completed successfully")
}

func (h *TransactionHandler) Deposit(c fiber.Ctx) error {
	var req dto.DepositRequest
	if err := c.Bind().Body(&req); err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, "Invalid deposit payload")
	}

	callerUserID := security.GetUserID(c)
	reqID := security.GetRequestID(c)
	ip := c.IP()

	res, err := h.svc.Deposit(req, callerUserID, reqID, ip)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, res, "Deposit completed successfully")
}

func (h *TransactionHandler) Withdraw(c fiber.Ctx) error {
	var req dto.WithdrawRequest
	if err := c.Bind().Body(&req); err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, "Invalid withdrawal payload")
	}

	callerUserID := security.GetUserID(c)
	isPrivileged := isStaffRole(c)
	reqID := security.GetRequestID(c)
	ip := c.IP()

	res, err := h.svc.Withdraw(req, callerUserID, isPrivileged, reqID, ip)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, res, "Withdrawal completed successfully")
}

func (h *TransactionHandler) GetByID(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, "Invalid transaction ID")
	}

	res, err := h.svc.GetByID(uint(id))
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, res)
}

func (h *TransactionHandler) List(c fiber.Ctx) error {
	accIDStr := c.Query("account_id")
	var accID uint
	if accIDStr != "" {
		if id, err := strconv.ParseUint(accIDStr, 10, 64); err == nil {
			accID = uint(id)
		}
	}

	limitStr := c.Query("limit", "20")
	offsetStr := c.Query("offset", "0")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	res, err := h.svc.List(accID, limit, offset)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, res)
}

func (h *TransactionHandler) ListLedger(c fiber.Ctx) error {
	accIDStr := c.Query("account_id")
	var accID uint
	if accIDStr != "" {
		if id, err := strconv.ParseUint(accIDStr, 10, 64); err == nil {
			accID = uint(id)
		}
	}

	limitStr := c.Query("limit", "50")
	offsetStr := c.Query("offset", "0")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	res, err := h.svc.ListLedger(accID, limit, offset)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, res)
}

func isStaffRole(c fiber.Ctx) bool {
	for _, r := range security.GetUserRoles(c) {
		if r == "SUPER_ADMIN" || r == "BANK_ADMIN" || r == "TELLER" || r == "BRANCH_MANAGER" {
			return true
		}
	}
	return false
}
