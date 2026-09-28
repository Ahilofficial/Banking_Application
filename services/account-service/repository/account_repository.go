package repository

import (
	"fmt"

	"banking-microservices/pkg/common/money"
	"banking-microservices/services/account-service/models"

	"github.com/google/wire"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProviderSet provides AccountRepository dependency for Wire
var ProviderSet = wire.NewSet(ProvideAccountRepository)

type AccountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// ProvideAccountRepository creates and auto-migrates the AccountRepository
func ProvideAccountRepository(db *gorm.DB) (*AccountRepository, error) {
	repo := NewAccountRepository(db)
	if err := repo.AutoMigrate(); err != nil {
		return nil, fmt.Errorf("account-service db migration failed: %w", err)
	}
	return repo, nil
}

func (r *AccountRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&models.Account{})
}

func (r *AccountRepository) DB() *gorm.DB {
	return r.db
}

func (r *AccountRepository) Create(account *models.Account) error {
	return r.db.Create(account).Error
}

func (r *AccountRepository) FindByID(id uint) (*models.Account, error) {
	var a models.Account
	err := r.db.First(&a, id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AccountRepository) FindByAccountNumber(accNum string) (*models.Account, error) {
	var a models.Account
	err := r.db.Where("account_number = ?", accNum).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AccountRepository) ListByCustomerID(customerID uint) ([]models.Account, error) {
	var accounts []models.Account
	err := r.db.Where("customer_id = ?", customerID).Find(&accounts).Error
	return accounts, err
}

func (r *AccountRepository) ListByUserID(userID uint) ([]models.Account, error) {
	var accounts []models.Account
	err := r.db.Where("user_id = ?", userID).Find(&accounts).Error
	return accounts, err
}

func (r *AccountRepository) ListAll(limit, offset int) ([]models.Account, error) {
	var accounts []models.Account
	err := r.db.Limit(limit).Offset(offset).Order("id DESC").Find(&accounts).Error
	return accounts, err
}

func (r *AccountRepository) Update(account *models.Account) error {
	return r.db.Save(account).Error
}

// GetForUpdate acquires an exclusive row lock (FOR UPDATE) within a transaction
func (r *AccountRepository) GetForUpdate(tx *gorm.DB, id uint) (*models.Account, error) {
	var a models.Account
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&a, id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AccountRepository) UpdateBalance(tx *gorm.DB, id uint, newBalance money.Amount) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Model(&models.Account{}).Where("id = ?", id).Update("balance", newBalance).Error
}
