package repository

import (
	"fmt"

	"banking-microservices/services/customer-service/models"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet provides CustomerRepository dependency for Wire
var ProviderSet = wire.NewSet(ProvideCustomerRepository)

type CustomerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

// ProvideCustomerRepository creates and auto-migrates the CustomerRepository
func ProvideCustomerRepository(db *gorm.DB) (*CustomerRepository, error) {
	repo := NewCustomerRepository(db)
	if err := repo.AutoMigrate(); err != nil {
		return nil, fmt.Errorf("customer-service db migration failed: %w", err)
	}
	return repo, nil
}

func (r *CustomerRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&models.Customer{})
}

func (r *CustomerRepository) Create(c *models.Customer) error {
	return r.db.Create(c).Error
}

func (r *CustomerRepository) FindByID(id uint) (*models.Customer, error) {
	var c models.Customer
	err := r.db.First(&c, id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CustomerRepository) FindByUserID(userID uint) (*models.Customer, error) {
	var c models.Customer
	err := r.db.Where("user_id = ?", userID).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CustomerRepository) List(limit, offset int) ([]models.Customer, error) {
	var customers []models.Customer
	err := r.db.Limit(limit).Offset(offset).Order("id DESC").Find(&customers).Error
	return customers, err
}

func (r *CustomerRepository) Update(c *models.Customer) error {
	return r.db.Save(c).Error
}
