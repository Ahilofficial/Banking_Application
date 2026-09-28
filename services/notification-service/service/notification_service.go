package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"banking-microservices/pkg/client"
	pb "banking-microservices/pkg/proto/notification"
	"banking-microservices/services/notification-service/models"
	"banking-microservices/services/notification-service/repository"

	"github.com/google/wire"
)

// ProviderSet provides NotificationService dependency for Wire
var ProviderSet = wire.NewSet(NewNotificationService)

type NotificationService struct {
	repo        *repository.NotificationRepository
	auditClient *client.AuditClient
}

func NewNotificationService(repo *repository.NotificationRepository, auditClient *client.AuditClient) *NotificationService {
	return &NotificationService{
		repo:        repo,
		auditClient: auditClient,
	}
}

// SendEmail handles sending emails and persisting audit records
func (s *NotificationService) SendEmail(ctx context.Context, req *pb.SendEmailRequest) (*pb.SendEmailResponse, error) {
	if req.To == "" {
		return nil, fmt.Errorf("recipient email address 'to' cannot be empty")
	}

	metaJSON, _ := json.Marshal(req.Metadata)
	bodyContent := req.Body
	if bodyContent == "" && req.HtmlBody != "" {
		bodyContent = req.HtmlBody
	}

	record := &models.Notification{
		UserID:    uint(req.UserId),
		Type:      "EMAIL",
		Recipient: req.To,
		Subject:   req.Subject,
		Content:   bodyContent,
		Status:    "SENT",
		Metadata:  string(metaJSON),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(record); err != nil {
		return nil, fmt.Errorf("failed to persist email notification: %w", err)
	}

	log.Printf("[notification-service] [EMAIL] To: %s | Subject: %s | Sent successfully (ID: %d)", req.To, req.Subject, record.ID)

	// Record audit
	if s.auditClient != nil && req.UserId > 0 {
		s.auditClient.Log(client.AuditLogPayload{
			UserID:       uint(req.UserId),
			Action:       "SEND_EMAIL",
			ResourceType: "NOTIFICATION",
			ResourceID:   strconv.Itoa(int(record.ID)),
			NewValue:     fmt.Sprintf("To: %s, Subject: %s", req.To, req.Subject),
		})
	}

	return &pb.SendEmailResponse{
		NotificationId: strconv.Itoa(int(record.ID)),
		Status:         pb.NotificationStatus_SENT,
		Message:        "Email sent successfully",
		SentAt:         record.CreatedAt.Format(time.RFC3339),
	}, nil
}

// SendSMS handles sending SMS and persisting audit records
func (s *NotificationService) SendSMS(ctx context.Context, req *pb.SendSMSRequest) (*pb.SendSMSResponse, error) {
	if req.PhoneNumber == "" {
		return nil, fmt.Errorf("phone number cannot be empty")
	}
	if req.Message == "" {
		return nil, fmt.Errorf("SMS message content cannot be empty")
	}

	metaJSON, _ := json.Marshal(req.Metadata)
	record := &models.Notification{
		UserID:    uint(req.UserId),
		Type:      "SMS",
		Recipient: req.PhoneNumber,
		Content:   req.Message,
		Status:    "SENT",
		Metadata:  string(metaJSON),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(record); err != nil {
		return nil, fmt.Errorf("failed to persist SMS notification: %w", err)
	}

	log.Printf("[notification-service] [SMS] Phone: %s | Message: %s | Sent successfully (ID: %d)", req.PhoneNumber, req.Message, record.ID)

	if s.auditClient != nil && req.UserId > 0 {
		s.auditClient.Log(client.AuditLogPayload{
			UserID:       uint(req.UserId),
			Action:       "SEND_SMS",
			ResourceType: "NOTIFICATION",
			ResourceID:   strconv.Itoa(int(record.ID)),
			NewValue:     fmt.Sprintf("Phone: %s", req.PhoneNumber),
		})
	}

	return &pb.SendSMSResponse{
		NotificationId: strconv.Itoa(int(record.ID)),
		Status:         pb.NotificationStatus_SENT,
		Message:        "SMS sent successfully",
		SentAt:         record.CreatedAt.Format(time.RFC3339),
	}, nil
}

// SendNotification handles sending Push/In-App notifications
func (s *NotificationService) SendNotification(ctx context.Context, req *pb.SendNotificationRequest) (*pb.SendNotificationResponse, error) {
	typeStr := "PUSH"
	if req.Type == pb.NotificationType_IN_APP {
		typeStr = "IN_APP"
	} else if req.Type == pb.NotificationType_EMAIL {
		typeStr = "EMAIL"
	} else if req.Type == pb.NotificationType_SMS {
		typeStr = "SMS"
	}

	metaJSON, _ := json.Marshal(req.Metadata)
	record := &models.Notification{
		UserID:    uint(req.UserId),
		Type:      typeStr,
		Recipient: req.Recipient,
		Subject:   req.Title,
		Content:   req.Content,
		Status:    "SENT",
		Metadata:  string(metaJSON),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(record); err != nil {
		return nil, fmt.Errorf("failed to persist notification: %w", err)
	}

	log.Printf("[notification-service] [%s] Recipient: %s | Title: %s | Sent successfully (ID: %d)", typeStr, req.Recipient, req.Title, record.ID)

	if s.auditClient != nil && req.UserId > 0 {
		s.auditClient.Log(client.AuditLogPayload{
			UserID:       uint(req.UserId),
			Action:       "SEND_NOTIFICATION",
			ResourceType: "NOTIFICATION",
			ResourceID:   strconv.Itoa(int(record.ID)),
			NewValue:     fmt.Sprintf("Type: %s, Title: %s", typeStr, req.Title),
		})
	}

	return &pb.SendNotificationResponse{
		NotificationId: strconv.Itoa(int(record.ID)),
		Status:         pb.NotificationStatus_SENT,
		Message:        fmt.Sprintf("%s notification sent successfully", typeStr),
		SentAt:         record.CreatedAt.Format(time.RFC3339),
	}, nil
}

// GetNotifications queries and retrieves notifications for a user
func (s *NotificationService) GetNotifications(ctx context.Context, req *pb.GetNotificationsRequest) (*pb.GetNotificationsResponse, error) {
	typeStr := ""
	if req.Type != pb.NotificationType_NOTIFICATION_TYPE_UNSPECIFIED {
		typeStr = req.Type.String()
	}

	items, total, err := s.repo.FindByUserID(uint(req.UserId), typeStr, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}

	respItems := make([]*pb.NotificationItem, 0, len(items))
	for _, it := range items {
		nType := pb.NotificationType_NOTIFICATION_TYPE_UNSPECIFIED
		switch it.Type {
		case "EMAIL":
			nType = pb.NotificationType_EMAIL
		case "SMS":
			nType = pb.NotificationType_SMS
		case "PUSH":
			nType = pb.NotificationType_PUSH
		case "IN_APP":
			nType = pb.NotificationType_IN_APP
		}

		nStatus := pb.NotificationStatus_SENT
		switch it.Status {
		case "PENDING":
			nStatus = pb.NotificationStatus_PENDING
		case "DELIVERED":
			nStatus = pb.NotificationStatus_DELIVERED
		case "FAILED":
			nStatus = pb.NotificationStatus_FAILED
		}

		respItems = append(respItems, &pb.NotificationItem{
			Id:        strconv.Itoa(int(it.ID)),
			UserId:    uint32(it.UserID),
			Type:      nType,
			Recipient: it.Recipient,
			Title:     it.Subject,
			Content:   it.Content,
			Status:    nStatus,
			CreatedAt: it.CreatedAt.Format(time.RFC3339),
		})
	}

	return &pb.GetNotificationsResponse{
		Notifications: respItems,
		Total:         int32(total),
	}, nil
}

// GetNotificationStatus queries delivery status of a single notification
func (s *NotificationService) GetNotificationStatus(ctx context.Context, req *pb.GetNotificationStatusRequest) (*pb.NotificationStatusResponse, error) {
	id, err := strconv.Atoi(req.NotificationId)
	if err != nil {
		return nil, fmt.Errorf("invalid notification ID: %s", req.NotificationId)
	}

	record, err := s.repo.FindByID(uint(id))
	if err != nil {
		return nil, fmt.Errorf("notification not found: %w", err)
	}

	nStatus := pb.NotificationStatus_SENT
	switch record.Status {
	case "PENDING":
		nStatus = pb.NotificationStatus_PENDING
	case "DELIVERED":
		nStatus = pb.NotificationStatus_DELIVERED
	case "FAILED":
		nStatus = pb.NotificationStatus_FAILED
	}

	return &pb.NotificationStatusResponse{
		NotificationId: strconv.Itoa(int(record.ID)),
		Status:         nStatus,
		ErrorMessage:   record.ErrorMessage,
		UpdatedAt:      record.UpdatedAt.Format(time.RFC3339),
	}, nil
}
