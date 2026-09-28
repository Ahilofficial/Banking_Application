package handlers

import (
	"strconv"

	"banking-microservices/pkg/common/errors"
	"banking-microservices/pkg/common/response"
	"banking-microservices/services/audit-service/dto"
	"banking-microservices/services/audit-service/service"

	"github.com/gofiber/fiber/v3"
)

type AuditHandler struct {
	svc *service.AuditService
}

func NewAuditHandler(svc *service.AuditService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

func (h *AuditHandler) Record(c fiber.Ctx) error {
	var req dto.CreateAuditLogRequest
	if err := c.Bind().Body(&req); err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, "Invalid audit log payload")
	}

	if req.IPAddress == "" {
		req.IPAddress = c.IP()
	}
	if req.UserAgent == "" {
		req.UserAgent = c.Get("User-Agent")
	}
	if req.RequestID == "" {
		req.RequestID = c.Get("X-Request-ID")
	}

	res, err := h.svc.Record(req)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusCreated, res)
}

func (h *AuditHandler) List(c fiber.Ctx) error {
	userIDStr := c.Query("user_id")
	action := c.Query("action")
	resourceType := c.Query("resource_type")
	limitStr := c.Query("limit", "50")
	offsetStr := c.Query("offset", "0")

	var userID uint
	if userIDStr != "" {
		if id, err := strconv.ParseUint(userIDStr, 10, 64); err == nil {
			userID = uint(id)
		}
	}
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	res, err := h.svc.List(userID, action, resourceType, limit, offset)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, res)
}
