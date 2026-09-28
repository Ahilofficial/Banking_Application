package handlers

import (
	"strconv"

	"banking-microservices/pkg/common/errors"
	"banking-microservices/pkg/common/response"
	"banking-microservices/pkg/common/security"
	"banking-microservices/services/customer-service/dto"
	"banking-microservices/services/customer-service/service"

	"github.com/gofiber/fiber/v3"
)

type CustomerHandler struct {
	svc *service.CustomerService
}

func NewCustomerHandler(svc *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{svc: svc}
}

func (h *CustomerHandler) Create(c fiber.Ctx) error {
	var req dto.CreateCustomerRequest
	if err := c.Bind().Body(&req); err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, "Invalid customer payload")
	}

	callerUserID := security.GetUserID(c)
	reqID := security.GetRequestID(c)
	ip := c.IP()

	res, err := h.svc.Create(req, callerUserID, reqID, ip)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusCreated, res, "Customer profile created successfully")
}

func (h *CustomerHandler) GetByID(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, "Invalid customer ID")
	}

	callerUserID := security.GetUserID(c)
	hasGlobalView := security.HasPermission(c, "customer:view") && isStaffRole(c)

	res, err := h.svc.GetByID(uint(id), callerUserID, hasGlobalView)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, res)
}

func (h *CustomerHandler) GetByUserID(c fiber.Ctx) error {
	userIDStr := c.Params("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, "Invalid user ID")
	}

	callerUserID := security.GetUserID(c)
	hasGlobalView := security.HasPermission(c, "customer:view") && isStaffRole(c)

	res, err := h.svc.GetByUserID(uint(userID), callerUserID, hasGlobalView)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, res)
}

func isStaffRole(c fiber.Ctx) bool {
	for _, r := range security.GetUserRoles(c) {
		if r == "SUPER_ADMIN" || r == "BANK_ADMIN" || r == "TELLER" || r == "BRANCH_MANAGER" || r == "BANK_EMPLOYEE" {
			return true
		}
	}
	return false
}

func (h *CustomerHandler) List(c fiber.Ctx) error {
	limitStr := c.Query("limit", "20")
	offsetStr := c.Query("offset", "0")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	res, err := h.svc.List(limit, offset)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, res)
}

func (h *CustomerHandler) UpdateKYC(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, "Invalid customer ID")
	}

	var req dto.UpdateKYCRequest
	if err := c.Bind().Body(&req); err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, "Invalid payload")
	}

	adminID := security.GetUserID(c)
	reqID := security.GetRequestID(c)
	ip := c.IP()

	res, err := h.svc.UpdateKYC(uint(id), req, adminID, reqID, ip)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, res, "Customer KYC updated successfully")
}
