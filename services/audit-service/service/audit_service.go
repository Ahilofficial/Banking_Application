package service

import (
	"time"

	appErrors "banking-microservices/pkg/common/errors"
	"banking-microservices/services/audit-service/dto"
	"banking-microservices/services/audit-service/models"
	"banking-microservices/services/audit-service/repository"

	"github.com/google/wire"
)

// ProviderSet provides AuditService dependency for Wire
var ProviderSet = wire.NewSet(NewAuditService)

type AuditService struct {
	repo *repository.AuditRepository
}

func NewAuditService(repo *repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Record(req dto.CreateAuditLogRequest) (*dto.AuditLogResponse, error) {
	log := &models.AuditLog{
		UserID:       req.UserID,
		Action:       req.Action,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		OldValue:     req.OldValue,
		NewValue:     req.NewValue,
		IPAddress:    req.IPAddress,
		UserAgent:    req.UserAgent,
		RequestID:    req.RequestID,
		CreatedAt:    time.Now(),
	}

	if err := s.repo.Create(log); err != nil {
		return nil, appErrors.ErrInternalServerError("Failed to record audit log")
	}

	return toAuditDTO(log), nil
}

func (s *AuditService) List(userID uint, action, resourceType string, limit, offset int) ([]dto.AuditLogResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	logs, err := s.repo.List(userID, action, resourceType, limit, offset)
	if err != nil {
		return nil, appErrors.ErrInternalServerError("Failed to query audit logs")
	}

	var res []dto.AuditLogResponse
	for _, l := range logs {
		res = append(res, *toAuditDTO(&l))
	}
	return res, nil
}

func toAuditDTO(l *models.AuditLog) *dto.AuditLogResponse {
	return &dto.AuditLogResponse{
		ID:           l.ID,
		UserID:       l.UserID,
		Action:       l.Action,
		ResourceType: l.ResourceType,
		ResourceID:   l.ResourceID,
		OldValue:     l.OldValue,
		NewValue:     l.NewValue,
		IPAddress:    l.IPAddress,
		RequestID:    l.RequestID,
		CreatedAt:    l.CreatedAt,
	}
}
