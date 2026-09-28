package service

import (
	"errors"
	"fmt"
	"time"

	"banking-microservices/pkg/client"
	appErrors "banking-microservices/pkg/common/errors"
	"banking-microservices/services/customer-service/dto"
	"banking-microservices/services/customer-service/models"
	"banking-microservices/services/customer-service/repository"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet provides CustomerService dependency for Wire
var ProviderSet = wire.NewSet(NewCustomerService)

type CustomerService struct {
	repo        *repository.CustomerRepository
	auditClient *client.AuditClient
}

func NewCustomerService(repo *repository.CustomerRepository, auditClient *client.AuditClient) *CustomerService {
	return &CustomerService{
		repo:        repo,
		auditClient: auditClient,
	}
}

func (s *CustomerService) Create(req dto.CreateCustomerRequest, callerUserID uint, reqID, ip string) (*dto.CustomerResponse, error) {
	if req.FirstName == "" || req.LastName == "" {
		return nil, appErrors.ErrBadRequest(appErrors.ErrCodeValidationFailed, "First name and last name are required")
	}

	targetUserID := req.UserID
	if targetUserID == 0 {
		targetUserID = callerUserID
	}

	// Check if customer profile already exists for this user_id
	if existing, _ := s.repo.FindByUserID(targetUserID); existing != nil {
		return nil, appErrors.ErrBadRequest(appErrors.ErrCodeConflict, "Customer profile already exists for this user")
	}

	customer := &models.Customer{
		UserID:      targetUserID,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		PhoneNumber: req.PhoneNumber,
		Address:     req.Address,
		NationalID:  req.NationalID,
		KYCStatus:   models.KYCStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(customer); err != nil {
		return nil, appErrors.ErrInternalServerError("Failed to create customer")
	}

	s.auditClient.Log(client.AuditLogPayload{
		UserID:       callerUserID,
		Action:       "CUSTOMER_CREATED",
		ResourceType: "CUSTOMER",
		ResourceID:   fmt.Sprintf("%d", customer.ID),
		IPAddress:    ip,
		RequestID:    reqID,
	})

	return toDTO(customer), nil
}

func (s *CustomerService) GetByID(id uint, callerUserID uint, hasGlobalView bool) (*dto.CustomerResponse, error) {
	customer, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrResourceNotFound(appErrors.ErrCodeCustomerNotFound, "Customer not found")
		}
		return nil, appErrors.ErrInternalServerError("Database lookup error")
	}

	// Level 2 Authorization: Customer can only view their own profile
	if !hasGlobalView && customer.UserID != callerUserID {
		return nil, appErrors.ErrForbid("Forbidden: you cannot access other customer profiles")
	}

	return toDTO(customer), nil
}

func (s *CustomerService) GetByUserID(userID uint, callerUserID uint, hasGlobalView bool) (*dto.CustomerResponse, error) {
	customer, err := s.repo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrResourceNotFound(appErrors.ErrCodeCustomerNotFound, "Customer profile not found for user")
		}
		return nil, appErrors.ErrInternalServerError("Database lookup error")
	}

	if !hasGlobalView && customer.UserID != callerUserID {
		return nil, appErrors.ErrForbid("Forbidden: you cannot access other customer profiles")
	}

	return toDTO(customer), nil
}

func (s *CustomerService) List(limit, offset int) ([]dto.CustomerResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	customers, err := s.repo.List(limit, offset)
	if err != nil {
		return nil, appErrors.ErrInternalServerError("Failed to query customers")
	}

	var res []dto.CustomerResponse
	for _, c := range customers {
		res = append(res, *toDTO(&c))
	}
	return res, nil
}

func (s *CustomerService) UpdateKYC(id uint, req dto.UpdateKYCRequest, adminID uint, reqID, ip string) (*dto.CustomerResponse, error) {
	customer, err := s.repo.FindByID(id)
	if err != nil {
		return nil, appErrors.ErrResourceNotFound(appErrors.ErrCodeCustomerNotFound, "Customer not found")
	}

	oldStatus := string(customer.KYCStatus)
	customer.KYCStatus = req.KYCStatus
	customer.UpdatedAt = time.Now()

	if err := s.repo.Update(customer); err != nil {
		return nil, appErrors.ErrInternalServerError("Failed to update KYC status")
	}

	s.auditClient.Log(client.AuditLogPayload{
		UserID:       adminID,
		Action:       "KYC_STATUS_UPDATED",
		ResourceType: "CUSTOMER",
		ResourceID:   fmt.Sprintf("%d", customer.ID),
		OldValue:     oldStatus,
		NewValue:     string(customer.KYCStatus),
		IPAddress:    ip,
		RequestID:    reqID,
	})

	return toDTO(customer), nil
}

func toDTO(c *models.Customer) *dto.CustomerResponse {
	return &dto.CustomerResponse{
		ID:          c.ID,
		UserID:      c.UserID,
		FirstName:   c.FirstName,
		LastName:    c.LastName,
		PhoneNumber: c.PhoneNumber,
		Address:     c.Address,
		NationalID:  c.NationalID,
		KYCStatus:   c.KYCStatus,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}
