package service

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"banking-microservices/pkg/client"
	appErrors "banking-microservices/pkg/common/errors"
	"banking-microservices/pkg/common/money"
	accModels "banking-microservices/services/account-service/models"
	"banking-microservices/services/transaction-service/dto"
	"banking-microservices/services/transaction-service/models"
	"banking-microservices/services/transaction-service/repository"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet provides TransactionService dependency for Wire
var ProviderSet = wire.NewSet(NewTransactionService)

type TransactionService struct {
	repo        *repository.TransactionRepository
	auditClient *client.AuditClient
}

func NewTransactionService(repo *repository.TransactionRepository, auditClient *client.AuditClient) *TransactionService {
	return &TransactionService{
		repo:        repo,
		auditClient: auditClient,
	}
}

// Transfer performs an atomic, ACID double-entry money transfer with row-level locking
func (s *TransactionService) Transfer(
	req dto.TransferRequest,
	callerUserID uint,
	isPrivilegedStaff bool,
	idempotencyKey string,
	reqID string,
	ip string,
) (*dto.TransactionResponse, error) {
	if req.FromAccountID == req.ToAccountID {
		return nil, appErrors.ErrBadRequest(appErrors.ErrCodeSameAccountTransfer, "Cannot transfer money to the same account")
	}

	transferAmount := money.FromDecimal(req.Amount)
	if transferAmount <= 0 {
		return nil, appErrors.ErrBadRequest(appErrors.ErrCodeValidationFailed, "Transfer amount must be greater than zero")
	}

	// 1. Idempotency Check
	var idempRecord *models.IdempotencyKey
	if idempotencyKey != "" {
		existing, err := s.repo.FindIdempotencyKey(callerUserID, idempotencyKey)
		if err == nil && existing != nil {
			if existing.Status == models.IdempotencySuccess {
				var cachedResp dto.TransactionResponse
				if json.Unmarshal([]byte(existing.Response), &cachedResp) == nil {
					return &cachedResp, nil
				}
			}
			if existing.Status == models.IdempotencyProcessing {
				return nil, appErrors.NewAppError(409, appErrors.ErrCodeIdempotencyConflict, "A transaction with this idempotency key is already processing")
			}
		}

		reqJSON, _ := json.Marshal(req)
		idempRecord = &models.IdempotencyKey{
			UserID:      callerUserID,
			Key:         idempotencyKey,
			RequestHash: repository.ComputeHash(reqJSON),
			Status:      models.IdempotencyProcessing,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		_ = s.repo.CreateIdempotencyKey(idempRecord)
	}

	// 2. Resource/Business Authorization (Level 2)
	sourceAccount, err := s.repo.GetAccountByID(req.FromAccountID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.markIdempotencyFailed(idempRecord)
			return nil, appErrors.ErrResourceNotFound(appErrors.ErrCodeAccountNotFound, "Source account not found")
		}
		s.markIdempotencyFailed(idempRecord)
		return nil, appErrors.ErrInternalServerError("Failed to check source account")
	}

	if !isPrivilegedStaff && sourceAccount.UserID != callerUserID {
		s.markIdempotencyFailed(idempRecord)
		return nil, appErrors.ErrForbid("Forbidden: you are not authorized to debit this account")
	}

	refNumber, err := generateTxnReference()
	if err != nil {
		s.markIdempotencyFailed(idempRecord)
		return nil, appErrors.ErrInternalServerError("Failed to generate transaction reference")
	}

	var transaction *models.Transaction

	// 3. ACID Transaction Execution with Row-Level Locking
	err = s.repo.DB().Transaction(func(tx *gorm.DB) error {
		// Deadlock Prevention: Lock rows in deterministic ascending ID order
		firstID, secondID := req.FromAccountID, req.ToAccountID
		if firstID > secondID {
			firstID, secondID = req.ToAccountID, req.FromAccountID
		}

		firstAcc, err := s.repo.GetAccountForUpdate(tx, firstID)
		if err != nil {
			return appErrors.ErrResourceNotFound(appErrors.ErrCodeAccountNotFound, fmt.Sprintf("Account %d not found", firstID))
		}

		secondAcc, err := s.repo.GetAccountForUpdate(tx, secondID)
		if err != nil {
			return appErrors.ErrResourceNotFound(appErrors.ErrCodeAccountNotFound, fmt.Sprintf("Account %d not found", secondID))
		}

		// Map locked accounts back to source and destination
		var src, dest *accModels.Account
		if firstAcc.ID == req.FromAccountID {
			src = firstAcc
			dest = secondAcc
		} else {
			src = secondAcc
			dest = firstAcc
		}

		// Validate account statuses
		if src.Status != accModels.AccountStatusActive {
			return appErrors.ErrBadRequest(appErrors.ErrCodeAccountBlocked, fmt.Sprintf("Source account is %s", src.Status))
		}
		if dest.Status != accModels.AccountStatusActive {
			return appErrors.ErrBadRequest(appErrors.ErrCodeAccountBlocked, fmt.Sprintf("Destination account is %s", dest.Status))
		}

		// Balance validation
		if src.Balance < transferAmount {
			return appErrors.ErrBadRequest(appErrors.ErrCodeInsufficientFunds, fmt.Sprintf("Insufficient balance. Available: %s, Requested: %s", src.Balance.String(), transferAmount.String()))
		}

		// Perform Balance updates
		src.Balance -= transferAmount
		dest.Balance += transferAmount

		if err := s.repo.UpdateAccountBalance(tx, src.ID, src.Balance); err != nil {
			return appErrors.ErrInternalServerError("Failed to update source balance")
		}
		if err := s.repo.UpdateAccountBalance(tx, dest.ID, dest.Balance); err != nil {
			return appErrors.ErrInternalServerError("Failed to update destination balance")
		}

		// Create Transaction record
		now := time.Now()
		fromID := req.FromAccountID
		toID := req.ToAccountID
		currency := req.Currency
		if currency == "" {
			currency = src.Currency
		}

		transaction = &models.Transaction{
			ReferenceNumber: refNumber,
			FromAccountID:   &fromID,
			ToAccountID:     &toID,
			Amount:          transferAmount,
			Currency:        currency,
			TransactionType: models.TxTypeTransfer,
			Status:          models.TxStatusSuccess,
			Description:     req.Description,
			CreatedAt:       now,
			CompletedAt:     &now,
		}

		if err := s.repo.CreateTransaction(tx, transaction); err != nil {
			return appErrors.ErrInternalServerError("Failed to persist transaction")
		}

		// Create Double-Entry Ledger entries
		debitEntry := &models.LedgerEntry{
			TransactionID: transaction.ID,
			AccountID:     src.ID,
			EntryType:     models.LedgerDebit,
			Amount:        transferAmount,
			Currency:      currency,
			BalanceAfter:  src.Balance,
			CreatedAt:     now,
		}
		if err := s.repo.CreateLedgerEntry(tx, debitEntry); err != nil {
			return appErrors.ErrInternalServerError("Failed to record ledger debit entry")
		}

		creditEntry := &models.LedgerEntry{
			TransactionID: transaction.ID,
			AccountID:     dest.ID,
			EntryType:     models.LedgerCredit,
			Amount:        transferAmount,
			Currency:      currency,
			BalanceAfter:  dest.Balance,
			CreatedAt:     now,
		}
		if err := s.repo.CreateLedgerEntry(tx, creditEntry); err != nil {
			return appErrors.ErrInternalServerError("Failed to record ledger credit entry")
		}

		return nil
	})

	if err != nil {
		s.markIdempotencyFailed(idempRecord)
		return nil, err
	}

	resp := toTransactionDTO(transaction)

	// 4. Update Idempotency Record to SUCCESS with cached response
	if idempRecord != nil {
		respBytes, _ := json.Marshal(resp)
		idempRecord.Status = models.IdempotencySuccess
		idempRecord.Response = string(respBytes)
		idempRecord.UpdatedAt = time.Now()
		_ = s.repo.UpdateIdempotencyKey(idempRecord)
	}

	// 5. Emit Audit Log
	s.auditClient.Log(client.AuditLogPayload{
		UserID:       callerUserID,
		Action:       "TRANSFER_COMPLETED",
		ResourceType: "TRANSACTION",
		ResourceID:   fmt.Sprintf("%d", transaction.ID),
		NewValue:     fmt.Sprintf("From: %d, To: %d, Amount: %s", req.FromAccountID, req.ToAccountID, transferAmount.String()),
		IPAddress:    ip,
		RequestID:    reqID,
	})

	return resp, nil
}

// Deposit credits money to an account
func (s *TransactionService) Deposit(
	req dto.DepositRequest,
	callerUserID uint,
	reqID string,
	ip string,
) (*dto.TransactionResponse, error) {
	amount := money.FromDecimal(req.Amount)
	if amount <= 0 {
		return nil, appErrors.ErrBadRequest(appErrors.ErrCodeValidationFailed, "Deposit amount must be greater than zero")
	}

	refNumber, err := generateTxnReference()
	if err != nil {
		return nil, appErrors.ErrInternalServerError("Reference generation failed")
	}

	var transaction *models.Transaction

	err = s.repo.DB().Transaction(func(tx *gorm.DB) error {
		account, err := s.repo.GetAccountForUpdate(tx, req.AccountID)
		if err != nil {
			return appErrors.ErrResourceNotFound(appErrors.ErrCodeAccountNotFound, "Account not found")
		}

		if account.Status != accModels.AccountStatusActive {
			return appErrors.ErrBadRequest(appErrors.ErrCodeAccountBlocked, "Account is not active")
		}

		account.Balance += amount
		if err := s.repo.UpdateAccountBalance(tx, account.ID, account.Balance); err != nil {
			return appErrors.ErrInternalServerError("Failed to update balance")
		}

		now := time.Now()
		toID := req.AccountID
		currency := req.Currency
		if currency == "" {
			currency = account.Currency
		}

		transaction = &models.Transaction{
			ReferenceNumber: refNumber,
			ToAccountID:     &toID,
			Amount:          amount,
			Currency:        currency,
			TransactionType: models.TxTypeDeposit,
			Status:          models.TxStatusSuccess,
			Description:     req.Description,
			CreatedAt:       now,
			CompletedAt:     &now,
		}
		if err := s.repo.CreateTransaction(tx, transaction); err != nil {
			return err
		}

		// Double-entry record for deposit (CREDIT to account)
		creditEntry := &models.LedgerEntry{
			TransactionID: transaction.ID,
			AccountID:     account.ID,
			EntryType:     models.LedgerCredit,
			Amount:        amount,
			Currency:      currency,
			BalanceAfter:  account.Balance,
			CreatedAt:     now,
		}
		return s.repo.CreateLedgerEntry(tx, creditEntry)
	})

	if err != nil {
		return nil, err
	}

	s.auditClient.Log(client.AuditLogPayload{
		UserID:       callerUserID,
		Action:       "DEPOSIT_COMPLETED",
		ResourceType: "TRANSACTION",
		ResourceID:   fmt.Sprintf("%d", transaction.ID),
		NewValue:     fmt.Sprintf("Account: %d, Amount: %s", req.AccountID, amount.String()),
		IPAddress:    ip,
		RequestID:    reqID,
	})

	return toTransactionDTO(transaction), nil
}

// Withdraw debits money from an account
func (s *TransactionService) Withdraw(
	req dto.WithdrawRequest,
	callerUserID uint,
	isPrivilegedStaff bool,
	reqID string,
	ip string,
) (*dto.TransactionResponse, error) {
	amount := money.FromDecimal(req.Amount)
	if amount <= 0 {
		return nil, appErrors.ErrBadRequest(appErrors.ErrCodeValidationFailed, "Withdrawal amount must be greater than zero")
	}

	// Verify ownership if caller is customer
	accountCheck, err := s.repo.GetAccountByID(req.AccountID)
	if err != nil {
		return nil, appErrors.ErrResourceNotFound(appErrors.ErrCodeAccountNotFound, "Account not found")
	}
	if !isPrivilegedStaff && accountCheck.UserID != callerUserID {
		return nil, appErrors.ErrForbid("Forbidden: you are not authorized to withdraw from this account")
	}

	refNumber, err := generateTxnReference()
	if err != nil {
		return nil, appErrors.ErrInternalServerError("Reference generation failed")
	}

	var transaction *models.Transaction

	err = s.repo.DB().Transaction(func(tx *gorm.DB) error {
		account, err := s.repo.GetAccountForUpdate(tx, req.AccountID)
		if err != nil {
			return appErrors.ErrResourceNotFound(appErrors.ErrCodeAccountNotFound, "Account not found")
		}

		if account.Status != accModels.AccountStatusActive {
			return appErrors.ErrBadRequest(appErrors.ErrCodeAccountBlocked, "Account is not active")
		}

		if account.Balance < amount {
			return appErrors.ErrBadRequest(appErrors.ErrCodeInsufficientFunds, "Insufficient balance")
		}

		account.Balance -= amount
		if err := s.repo.UpdateAccountBalance(tx, account.ID, account.Balance); err != nil {
			return appErrors.ErrInternalServerError("Failed to update balance")
		}

		now := time.Now()
		fromID := req.AccountID
		currency := req.Currency
		if currency == "" {
			currency = account.Currency
		}

		transaction = &models.Transaction{
			ReferenceNumber: refNumber,
			FromAccountID:   &fromID,
			Amount:          amount,
			Currency:        currency,
			TransactionType: models.TxTypeWithdrawal,
			Status:          models.TxStatusSuccess,
			Description:     req.Description,
			CreatedAt:       now,
			CompletedAt:     &now,
		}
		if err := s.repo.CreateTransaction(tx, transaction); err != nil {
			return err
		}

		// Double-entry record for withdrawal (DEBIT from account)
		debitEntry := &models.LedgerEntry{
			TransactionID: transaction.ID,
			AccountID:     account.ID,
			EntryType:     models.LedgerDebit,
			Amount:        amount,
			Currency:      currency,
			BalanceAfter:  account.Balance,
			CreatedAt:     now,
		}
		return s.repo.CreateLedgerEntry(tx, debitEntry)
	})

	if err != nil {
		return nil, err
	}

	s.auditClient.Log(client.AuditLogPayload{
		UserID:       callerUserID,
		Action:       "WITHDRAWAL_COMPLETED",
		ResourceType: "TRANSACTION",
		ResourceID:   fmt.Sprintf("%d", transaction.ID),
		NewValue:     fmt.Sprintf("Account: %d, Amount: %s", req.AccountID, amount.String()),
		IPAddress:    ip,
		RequestID:    reqID,
	})

	return toTransactionDTO(transaction), nil
}

func (s *TransactionService) GetByID(id uint) (*dto.TransactionResponse, error) {
	t, err := s.repo.FindTransactionByID(id)
	if err != nil {
		return nil, appErrors.ErrResourceNotFound(appErrors.ErrCodeNotFound, "Transaction not found")
	}
	return toTransactionDTO(t), nil
}

func (s *TransactionService) List(accountID uint, limit, offset int) ([]dto.TransactionResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	txs, err := s.repo.ListTransactions(accountID, limit, offset)
	if err != nil {
		return nil, appErrors.ErrInternalServerError("Failed to query transactions")
	}

	var res []dto.TransactionResponse
	for _, t := range txs {
		res = append(res, *toTransactionDTO(&t))
	}
	return res, nil
}

func (s *TransactionService) ListLedger(accountID uint, limit, offset int) ([]dto.LedgerEntryResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	entries, err := s.repo.ListLedgerEntries(accountID, limit, offset)
	if err != nil {
		return nil, appErrors.ErrInternalServerError("Failed to query ledger entries")
	}

	var res []dto.LedgerEntryResponse
	for _, e := range entries {
		res = append(res, dto.LedgerEntryResponse{
			ID:            e.ID,
			TransactionID: e.TransactionID,
			AccountID:     e.AccountID,
			EntryType:     e.EntryType,
			Amount:        e.Amount.ToDecimal(),
			AmountMinor:   int64(e.Amount),
			BalanceAfter:  e.BalanceAfter.ToDecimal(),
			Currency:      e.Currency,
			CreatedAt:     e.CreatedAt,
		})
	}
	return res, nil
}

func (s *TransactionService) markIdempotencyFailed(idemp *models.IdempotencyKey) {
	if idemp != nil {
		idemp.Status = models.IdempotencyFailed
		idemp.UpdatedAt = time.Now()
		_ = s.repo.UpdateIdempotencyKey(idemp)
	}
}

func generateTxnReference() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(90000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("TXN-%d-%08d", time.Now().UnixNano()/1e6, n.Int64()+10000000), nil
}

func toTransactionDTO(t *models.Transaction) *dto.TransactionResponse {
	return &dto.TransactionResponse{
		ID:              t.ID,
		ReferenceNumber: t.ReferenceNumber,
		FromAccountID:   t.FromAccountID,
		ToAccountID:     t.ToAccountID,
		Amount:          t.Amount.ToDecimal(),
		AmountMinor:     int64(t.Amount),
		AmountDisplay:   t.Amount.String(),
		Currency:        t.Currency,
		TransactionType: t.TransactionType,
		Status:          t.Status,
		Description:     t.Description,
		CreatedAt:       t.CreatedAt,
		CompletedAt:     t.CompletedAt,
	}
}
