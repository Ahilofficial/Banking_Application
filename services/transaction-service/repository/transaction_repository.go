package repository

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"banking-microservices/pkg/common/money"
	accModels "banking-microservices/services/account-service/models"
	"banking-microservices/services/transaction-service/models"

	"github.com/google/wire"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProviderSet provides TransactionRepository dependency for Wire
var ProviderSet = wire.NewSet(ProvideTransactionRepository)

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// ProvideTransactionRepository creates and auto-migrates the TransactionRepository
func ProvideTransactionRepository(db *gorm.DB) (*TransactionRepository, error) {
	repo := NewTransactionRepository(db)
	if err := repo.AutoMigrate(); err != nil {
		return nil, fmt.Errorf("transaction-service db migration failed: %w", err)
	}
	return repo, nil
}

func (r *TransactionRepository) AutoMigrate() error {
	return r.db.AutoMigrate(
		&models.Transaction{},
		&models.LedgerEntry{},
		&models.IdempotencyKey{},
		&accModels.Account{},
	)
}

func (r *TransactionRepository) DB() *gorm.DB {
	return r.db
}

// ComputeHash computes a SHA-256 hash of the request body
func ComputeHash(payload []byte) string {
	hash := sha256.Sum256(payload)
	return hex.EncodeToString(hash[:])
}

// Idempotency operations
func (r *TransactionRepository) FindIdempotencyKey(userID uint, key string) (*models.IdempotencyKey, error) {
	var item models.IdempotencyKey
	err := r.db.Where("user_id = ? AND idempotency_key = ?", userID, key).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *TransactionRepository) CreateIdempotencyKey(key *models.IdempotencyKey) error {
	return r.db.Create(key).Error
}

func (r *TransactionRepository) UpdateIdempotencyKey(key *models.IdempotencyKey) error {
	return r.db.Save(key).Error
}

// GetAccountForUpdate locks the account row exclusively within the active transaction
func (r *TransactionRepository) GetAccountForUpdate(tx *gorm.DB, accountID uint) (*accModels.Account, error) {
	var account accModels.Account
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&account, accountID).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// UpdateAccountBalance updates the locked balance in the transaction
func (r *TransactionRepository) UpdateAccountBalance(tx *gorm.DB, accountID uint, newBalance money.Amount) error {
	return tx.Model(&accModels.Account{}).Where("id = ?", accountID).Update("balance", newBalance).Error
}

// Transaction operations
func (r *TransactionRepository) CreateTransaction(tx *gorm.DB, t *models.Transaction) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Create(t).Error
}

func (r *TransactionRepository) UpdateTransaction(tx *gorm.DB, t *models.Transaction) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Save(t).Error
}

func (r *TransactionRepository) FindTransactionByID(id uint) (*models.Transaction, error) {
	var t models.Transaction
	err := r.db.First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TransactionRepository) FindTransactionByRef(ref string) (*models.Transaction, error) {
	var t models.Transaction
	err := r.db.Where("reference_number = ?", ref).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TransactionRepository) ListTransactions(accountID uint, limit, offset int) ([]models.Transaction, error) {
	var txs []models.Transaction
	query := r.db.Order("id DESC").Limit(limit).Offset(offset)
	if accountID > 0 {
		query = query.Where("from_account_id = ? OR to_account_id = ?", accountID, accountID)
	}
	err := query.Find(&txs).Error
	return txs, err
}

// Ledger operations
func (r *TransactionRepository) CreateLedgerEntry(tx *gorm.DB, l *models.LedgerEntry) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Create(l).Error
}

func (r *TransactionRepository) ListLedgerEntries(accountID uint, limit, offset int) ([]models.LedgerEntry, error) {
	var entries []models.LedgerEntry
	query := r.db.Order("id DESC").Limit(limit).Offset(offset)
	if accountID > 0 {
		query = query.Where("account_id = ?", accountID)
	}
	err := query.Find(&entries).Error
	return entries, err
}

// GetAccountByID reads an account without lock
func (r *TransactionRepository) GetAccountByID(accountID uint) (*accModels.Account, error) {
	var a accModels.Account
	err := r.db.First(&a, accountID).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}
