package handlers

import (
	"strconv"

	"banking-microservices/pkg/common/errors"
	"banking-microservices/pkg/common/response"
	"banking-microservices/pkg/common/security"
	pb "banking-microservices/pkg/proto/notification"
	"banking-microservices/services/notification-service/service"

	"github.com/gofiber/fiber/v3"
	"github.com/google/wire"
)

// ProviderSet provides NotificationHandler dependency for Wire
var ProviderSet = wire.NewSet(NewNotificationHandler)

type NotificationHandler struct {
	svc *service.NotificationService
}

func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// SendEmail handles POST /api/v1/notifications/email
func (h *NotificationHandler) SendEmail(c fiber.Ctx) error {
	var req pb.SendEmailRequest
	if err := c.Bind().Body(&req); err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, "Invalid request payload")
	}

	if req.UserId == 0 {
		req.UserId = uint32(security.GetUserID(c))
	}

	res, err := h.svc.SendEmail(c.Context(), &req)
	if err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, err.Error())
	}

	return response.Success(c, fiber.StatusOK, res, "Email notification queued/sent")
}

// SendSMS handles POST /api/v1/notifications/sms
func (h *NotificationHandler) SendSMS(c fiber.Ctx) error {
	var req pb.SendSMSRequest
	if err := c.Bind().Body(&req); err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, "Invalid request payload")
	}

	if req.UserId == 0 {
		req.UserId = uint32(security.GetUserID(c))
	}

	res, err := h.svc.SendSMS(c.Context(), &req)
	if err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, err.Error())
	}

	return response.Success(c, fiber.StatusOK, res, "SMS notification queued/sent")
}

// SendNotification handles POST /api/v1/notifications/send
func (h *NotificationHandler) SendNotification(c fiber.Ctx) error {
	var req pb.SendNotificationRequest
	if err := c.Bind().Body(&req); err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, "Invalid request payload")
	}

	if req.UserId == 0 {
		req.UserId = uint32(security.GetUserID(c))
	}

	res, err := h.svc.SendNotification(c.Context(), &req)
	if err != nil {
		return errors.ErrBadRequest(errors.ErrCodeInvalidRequest, err.Error())
	}

	return response.Success(c, fiber.StatusOK, res, "Notification queued/sent")
}

// List handles GET /api/v1/notifications
func (h *NotificationHandler) List(c fiber.Ctx) error {
	userID := security.GetUserID(c)
	if paramUID := c.Query("user_id"); paramUID != "" {
		if uid, err := strconv.Atoi(paramUID); err == nil && uid > 0 {
			userID = uint(uid)
		}
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	res, err := h.svc.GetNotifications(c.Context(), &pb.GetNotificationsRequest{
		UserId: uint32(userID),
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return errors.ErrInternalServerError(err.Error())
	}

	return response.Success(c, fiber.StatusOK, res, "Notifications retrieved")
}

// GetByID handles GET /api/v1/notifications/:id
func (h *NotificationHandler) GetByID(c fiber.Ctx) error {
	id := c.Params("id")
	res, err := h.svc.GetNotificationStatus(c.Context(), &pb.GetNotificationStatusRequest{
		NotificationId: id,
	})
	if err != nil {
		return errors.ErrResourceNotFound(errors.ErrCodeNotFound, err.Error())
	}

	return response.Success(c, fiber.StatusOK, res, "Notification status retrieved")
}
