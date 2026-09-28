package repository

import (
	"fmt"

	"banking-microservices/services/notification-service/models"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet provides NotificationRepository dependency for Wire
var ProviderSet = wire.NewSet(ProvideNotificationRepository)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// ProvideNotificationRepository creates and auto-migrates the NotificationRepository
func ProvideNotificationRepository(db *gorm.DB) (*NotificationRepository, error) {
	repo := NewNotificationRepository(db)
	if err := repo.AutoMigrate(); err != nil {
		return nil, fmt.Errorf("notification-service db migration failed: %w", err)
	}
	return repo, nil
}

func (r *NotificationRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&models.Notification{})
}

func (r *NotificationRepository) Create(n *models.Notification) error {
	return r.db.Create(n).Error
}

func (r *NotificationRepository) FindByID(id uint) (*models.Notification, error) {
	var n models.Notification
	err := r.db.First(&n, id).Error
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *NotificationRepository) FindByUserID(userID uint, notifType string, limit, offset int) ([]models.Notification, int64, error) {
	var list []models.Notification
	var total int64

	query := r.db.Model(&models.Notification{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if notifType != "" && notifType != "NOTIFICATION_TYPE_UNSPECIFIED" {
		query = query.Where("type = ?", notifType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&list).Error
	return list, total, err
}

func (r *NotificationRepository) UpdateStatus(id uint, status string, errMsg string) error {
	updates := map[string]interface{}{
		"status":        status,
		"error_message": errMsg,
	}
	return r.db.Model(&models.Notification{}).Where("id = ?", id).Updates(updates).Error
}
