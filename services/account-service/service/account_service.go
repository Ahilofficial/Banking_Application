package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"banking-microservices/pkg/client"
	appErrors "banking-microservices/pkg/common/errors"
	"banking-microservices/pkg/common/money"
	"banking-microservices/services/account-service/dto"
	"banking-microservices/services/account-service/models"
	"banking-microservices/services/account-service/repository"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet provides AccountService dependency for Wire
var ProviderSet = wire.NewSet(NewAccountService)

type AccountService struct {
	repo        *repository.AccountRepository
	auditClient *client.AuditClient
}

func NewAccountService(repo *repository.AccountRepository, auditClient *client.AuditClient) *AccountService {
	return &AccountService{
		repo:        repo,
		auditClient: auditClient,
	}
}

func (s *AccountService) Create(req dto.CreateAccountRequest, callerUserID uint, reqID, ip string) (*dto.AccountResponse, error) {
	if req.CustomerID == 0 {
		return nil, appErrors.ErrBadRequest(appErrors.ErrCodeValidationFailed, "Customer ID is required")
	}

	branch := req.BranchCode
	if branch == "" {
		branch = "MAIN01"
	}

	accType := req.AccountType
	if accType == "" {
		accType = models.AccountTypeSavings
	}

	currency := req.Currency
	if currency == "" {
		currency = "INR"
	}

	initialBalance := money.FromDecimal(req.InitialDeposit)
	if initialBalance < 0 {
		return nil, appErrors.ErrBadRequest(appErrors.ErrCodeValidationFailed, "Initial deposit cannot be negative")
	}

	accNum, err := generateAccountNumber(branch)
	if err != nil {
		return nil, appErrors.ErrInternalServerError("Failed to generate account number")
	}

	targetUserID := req.UserID
	if targetUserID == 0 {
		targetUserID = callerUserID
	}

	account := &models.Account{
		AccountNumber: accNum,
		CustomerID:    req.CustomerID,
		UserID:        targetUserID,
		BranchCode:    branch,
		AccountType:   accType,
		Currency:      currency,
		Balance:       initialBalance,
		Status:        models.AccountStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.Create(account); err != nil {
		return nil, appErrors.ErrInternalServerError("Failed to create account in database")
	}

	s.auditClient.Log(client.AuditLogPayload{
		UserID:       callerUserID,
		Action:       "ACCOUNT_CREATED",
		ResourceType: "ACCOUNT",
		ResourceID:   fmt.Sprintf("%d", account.ID),
		NewValue:     fmt.Sprintf("Number: %s, Initial: %s", account.AccountNumber, account.Balance.String()),
		IPAddress:    ip,
		RequestID:    reqID,
	})

	return toAccountDTO(account), nil
}

func (s *AccountService) GetByID(id uint, callerUserID uint, hasGlobalView bool) (*dto.AccountResponse, error) {
	account, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrResourceNotFound(appErrors.ErrCodeAccountNotFound, "Account not found")
		}
		return nil, appErrors.ErrInternalServerError("Database lookup error")
	}

	// Level 2 Authorization: Verify customer ownership if user lacks global view
	if !hasGlobalView && account.UserID != callerUserID {
		return nil, appErrors.ErrForbid("Forbidden: you do not have permission to access this account")
	}

	return toAccountDTO(account), nil
}

func (s *AccountService) GetByAccountNumber(accNum string, callerUserID uint, hasGlobalView bool) (*dto.AccountResponse, error) {
	account, err := s.repo.FindByAccountNumber(accNum)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrResourceNotFound(appErrors.ErrCodeAccountNotFound, "Account not found")
		}
		return nil, appErrors.ErrInternalServerError("Database lookup error")
	}

	if !hasGlobalView && account.UserID != callerUserID {
		return nil, appErrors.ErrForbid("Forbidden: you do not have permission to access this account")
	}

	return toAccountDTO(account), nil
}

func (s *AccountService) List(customerID, userID uint, callerUserID uint, hasGlobalView bool, limit, offset int) ([]dto.AccountResponse, error) {
	var accounts []models.Account
	var err error

	if !hasGlobalView {
		// Strict Level 2 filter: Only caller's accounts
		accounts, err = s.repo.ListByUserID(callerUserID)
	} else if customerID > 0 {
		accounts, err = s.repo.ListByCustomerID(customerID)
	} else if userID > 0 {
		accounts, err = s.repo.ListByUserID(userID)
	} else {
		accounts, err = s.repo.ListAll(limit, offset)
	}

	if err != nil {
		return nil, appErrors.ErrInternalServerError("Failed to retrieve accounts")
	}

	var res []dto.AccountResponse
	for _, a := range accounts {
		res = append(res, *toAccountDTO(&a))
	}
	return res, nil
}

func (s *AccountService) UpdateStatus(id uint, newStatus models.AccountStatus, callerUserID uint, reqID, ip string) (*dto.AccountResponse, error) {
	account, err := s.repo.FindByID(id)
	if err != nil {
		return nil, appErrors.ErrResourceNotFound(appErrors.ErrCodeAccountNotFound, "Account not found")
	}

	oldStatus := string(account.Status)
	account.Status = newStatus
	account.UpdatedAt = time.Now()

	if err := s.repo.Update(account); err != nil {
		return nil, appErrors.ErrInternalServerError("Failed to update account status")
	}

	s.auditClient.Log(client.AuditLogPayload{
		UserID:       callerUserID,
		Action:       "ACCOUNT_STATUS_CHANGED",
		ResourceType: "ACCOUNT",
		ResourceID:   fmt.Sprintf("%d", account.ID),
		OldValue:     oldStatus,
		NewValue:     string(newStatus),
		IPAddress:    ip,
		RequestID:    reqID,
	})

	return toAccountDTO(account), nil
}

func (s *AccountService) VerifyAccount(accountID, userID uint) (*dto.VerifyAccountResponse, error) {
	account, err := s.repo.FindByID(accountID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrResourceNotFound(appErrors.ErrCodeAccountNotFound, "Account not found")
		}
		return nil, appErrors.ErrInternalServerError("Database lookup error")
	}

	// If userID is passed, ensure ownership
	if userID > 0 && account.UserID != userID {
		return nil, appErrors.ErrForbid("Account does not belong to specified user")
	}

	if account.Status != models.AccountStatusActive {
		return nil, appErrors.ErrBadRequest(appErrors.ErrCodeAccountBlocked, fmt.Sprintf("Account is %s", account.Status))
	}

	return &dto.VerifyAccountResponse{
		Valid:         true,
		AccountID:     account.ID,
		AccountNumber: account.AccountNumber,
		CustomerID:    account.CustomerID,
		UserID:        account.UserID,
		Balance:       int64(account.Balance),
		Status:        account.Status,
	}, nil
}

func generateAccountNumber(branchCode string) (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(9000000000))
	if err != nil {
		return "", err
	}
	num := n.Int64() + 1000000000
	cleanBranch := strings.ToUpper(strings.ReplaceAll(branchCode, "-", ""))
	return fmt.Sprintf("AC-%s-%d", cleanBranch, num), nil
}

func toAccountDTO(a *models.Account) *dto.AccountResponse {
	return &dto.AccountResponse{
		ID:             a.ID,
		AccountNumber:  a.AccountNumber,
		CustomerID:     a.CustomerID,
		UserID:         a.UserID,
		BranchCode:     a.BranchCode,
		AccountType:    a.AccountType,
		Currency:       a.Currency,
		Balance:        a.Balance.ToDecimal(),
		BalancePaise:   int64(a.Balance),
		BalanceDisplay: a.Balance.String(),
		Status:         a.Status,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}
}
