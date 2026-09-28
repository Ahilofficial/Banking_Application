package repository

import (
	"fmt"

	"banking-microservices/services/audit-service/models"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet provides AuditRepository dependency for Wire
var ProviderSet = wire.NewSet(ProvideAuditRepository)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// ProvideAuditRepository creates and auto-migrates the AuditRepository
func ProvideAuditRepository(db *gorm.DB) (*AuditRepository, error) {
	repo := NewAuditRepository(db)
	if err := repo.AutoMigrate(); err != nil {
		return nil, fmt.Errorf("audit-service db migration failed: %w", err)
	}
	return repo, nil
}

func (r *AuditRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&models.AuditLog{})
}

func (r *AuditRepository) Create(log *models.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *AuditRepository) List(userID uint, action, resourceType string, limit, offset int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	query := r.db.Order("id DESC").Limit(limit).Offset(offset)

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}

	err := query.Find(&logs).Error
	return logs, err
}
