package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"banking-microservices/pkg/client"
	"banking-microservices/pkg/common/database"
	"banking-microservices/pkg/common/errors"
	"banking-microservices/pkg/common/middleware"

	accHandlers "banking-microservices/services/account-service/handlers"
	accModels "banking-microservices/services/account-service/models"
	accRepo "banking-microservices/services/account-service/repository"
	accService "banking-microservices/services/account-service/service"

	auditHandlers "banking-microservices/services/audit-service/handlers"
	auditModels "banking-microservices/services/audit-service/models"
	auditRepo "banking-microservices/services/audit-service/repository"
	auditService "banking-microservices/services/audit-service/service"

	authDto "banking-microservices/services/auth-service/dto"
	authHandlers "banking-microservices/services/auth-service/handlers"
	authModels "banking-microservices/services/auth-service/models"
	authRepository "banking-microservices/services/auth-service/repository"
	authService "banking-microservices/services/auth-service/service"

	custHandlers "banking-microservices/services/customer-service/handlers"
	custModels "banking-microservices/services/customer-service/models"
	custRepo "banking-microservices/services/customer-service/repository"
	custService "banking-microservices/services/customer-service/service"

	txDto "banking-microservices/services/transaction-service/dto"
	txHandlers "banking-microservices/services/transaction-service/handlers"
	txModels "banking-microservices/services/transaction-service/models"
	txRepo "banking-microservices/services/transaction-service/repository"
	txService "banking-microservices/services/transaction-service/service"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"gorm.io/gorm"
)

var (
	testDB        *gorm.DB
	testJWTSecret = "test-secret-key-12345"

	authApp   *fiber.App
	custApp   *fiber.App
	accApp    *fiber.App
	txApp     *fiber.App
	auditApp  *fiber.App
)

func TestMain(m *testing.M) {
	_ = os.Remove("test_banking.db")
	var err error
	testDB, err = database.Connect("sqlite", "test_banking.db")
	if err != nil {
		panic(err)
	}

	// Auto migrate
	_ = testDB.AutoMigrate(
		&authModels.User{},
		&authModels.Role{},
		&authModels.Permission{},
		&authModels.UserRole{},
		&authModels.RolePermission{},
		&authModels.Session{},
		&custModels.Customer{},
		&accModels.Account{},
		&txModels.Transaction{},
		&txModels.LedgerEntry{},
		&txModels.IdempotencyKey{},
		&auditModels.AuditLog{},
	)

	// Seed roles & permissions
	aRepo := authRepository.NewAuthRepository(testDB)
	_ = aRepo.SeedDefaultRolesAndPermissions()

	auditR := auditRepo.NewAuditRepository(testDB)
	auditSvc := auditService.NewAuditService(auditR)
	auditH := auditHandlers.NewAuditHandler(auditSvc)

	auditApp = fiber.New(fiber.Config{ErrorHandler: errors.FiberErrorHandler})
	auditApp.Use(middleware.RequestID())
	auditApp.Post("/api/v1/audit/logs", auditH.Record)
	auditApp.Get("/api/v1/audit/logs", middleware.AuthRequired(testJWTSecret), auditH.List)

	auditClient := client.NewAuditClient("http://localhost:8005")

	authSvc := authService.NewAuthService(aRepo, testJWTSecret, time.Hour, 7*24*time.Hour, auditClient)
	authH := authHandlers.NewAuthHandler(authSvc)
	authApp = fiber.New(fiber.Config{ErrorHandler: errors.FiberErrorHandler})
	authApp.Use(middleware.RequestID())
	authApp.Post("/api/v1/auth/register", authH.Register)
	authApp.Post("/api/v1/auth/login", authH.Login)
	authApp.Get("/api/v1/auth/users", middleware.AuthRequired(testJWTSecret), middleware.RequirePermission("user:manage"), authH.ListUsers)

	cRepo := custRepo.NewCustomerRepository(testDB)
	cSvc := custService.NewCustomerService(cRepo, auditClient)
	cH := custHandlers.NewCustomerHandler(cSvc)
	custApp = fiber.New(fiber.Config{ErrorHandler: errors.FiberErrorHandler})
	custApp.Use(middleware.RequestID())
	custApp.Use(middleware.AuthRequired(testJWTSecret))
	custApp.Post("/api/v1/customers", cH.Create)
	custApp.Get("/api/v1/customers/:id", cH.GetByID)

	acRepo := accRepo.NewAccountRepository(testDB)
	acSvc := accService.NewAccountService(acRepo, auditClient)
	acH := accHandlers.NewAccountHandler(acSvc)
	accApp = fiber.New(fiber.Config{ErrorHandler: errors.FiberErrorHandler})
	accApp.Use(middleware.RequestID())
	accApp.Use(middleware.AuthRequired(testJWTSecret))
	accApp.Post("/api/v1/accounts", middleware.RequirePermission("account:create"), acH.Create)
	accApp.Get("/api/v1/accounts/:id", middleware.RequirePermission("account:view"), acH.GetByID)
	accApp.Get("/api/v1/accounts", middleware.RequirePermission("account:view"), acH.List)

	tRepo := txRepo.NewTransactionRepository(testDB)
	tSvc := txService.NewTransactionService(tRepo, auditClient)
	tH := txHandlers.NewTransactionHandler(tSvc)
	txApp = fiber.New(fiber.Config{ErrorHandler: errors.FiberErrorHandler})
	txApp.Use(middleware.RequestID())
	txApp.Use(recover.New())
	txApp.Use(middleware.AuthRequired(testJWTSecret))
	txApp.Post("/api/v1/transactions/transfer", middleware.RequirePermission("transaction:create"), tH.Transfer)
	txApp.Get("/api/v1/transactions/ledger", middleware.RequirePermission("ledger:view"), tH.ListLedger)

	code := m.Run()
	_ = os.Remove("test_banking.db")
	os.Exit(code)
}

func doReq(app *fiber.App, method, url string, body any, token string, extraHeaders map[string]string) (*http.Response, map[string]any, error) {
	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}

	req, _ := http.NewRequest(method, url, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	resp, err := app.Test(req)
	if err != nil {
		return nil, nil, err
	}

	respBytes, _ := io.ReadAll(resp.Body)
	var respMap map[string]any
	_ = json.Unmarshal(respBytes, &respMap)

	return resp, respMap, nil
}

func TestCompleteBankingMicroservicesFlow(t *testing.T) {
	// 1. Register Alice (Customer) and Bob (Customer)
	resp, data, err := doReq(authApp, http.MethodPost, "/api/v1/auth/register", authDto.RegisterRequest{
		Email:    "alice_test@bank.com",
		Password: "Password@123",
		FullName: "Alice Test",
		Role:     "CUSTOMER",
	}, "", nil)
	if err != nil || resp.StatusCode != http.StatusCreated {
		t.Fatalf("Alice registration failed: %v, resp: %v", err, data)
	}

	_, _, err = doReq(authApp, http.MethodPost, "/api/v1/auth/register", authDto.RegisterRequest{
		Email:    "bob_test@bank.com",
		Password: "Password@123",
		FullName: "Bob Test",
		Role:     "CUSTOMER",
	}, "", nil)
	if err != nil {
		t.Fatalf("Bob registration failed: %v", err)
	}

	// Register Admin
	_, _, err = doReq(authApp, http.MethodPost, "/api/v1/auth/register", authDto.RegisterRequest{
		Email:    "admin_test@bank.com",
		Password: "Password@123",
		FullName: "Admin Test",
		Role:     "SUPER_ADMIN",
	}, "", nil)
	if err != nil {
		t.Fatalf("Admin registration failed: %v", err)
	}

	// 2. Login Alice and Bob
	resp, data, err = doReq(authApp, http.MethodPost, "/api/v1/auth/login", authDto.LoginRequest{
		Email:    "alice_test@bank.com",
		Password: "Password@123",
	}, "", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Alice login failed: %v", data)
	}
	tokenData := data["data"].(map[string]any)
	aliceToken := tokenData["access_token"].(string)
	aliceUser := tokenData["user"].(map[string]any)
	aliceUserID := uint(aliceUser["id"].(float64))

	resp, data, _ = doReq(authApp, http.MethodPost, "/api/v1/auth/login", authDto.LoginRequest{
		Email:    "bob_test@bank.com",
		Password: "Password@123",
	}, "", nil)
	bobToken := data["data"].(map[string]any)["access_token"].(string)
	bobUser := data["data"].(map[string]any)["user"].(map[string]any)
	bobUserID := uint(bobUser["id"].(float64))

	resp, data, _ = doReq(authApp, http.MethodPost, "/api/v1/auth/login", authDto.LoginRequest{
		Email:    "admin_test@bank.com",
		Password: "Password@123",
	}, "", nil)
	adminToken := data["data"].(map[string]any)["access_token"].(string)

	// 3. Test Level 1 RBAC: Alice (Customer) trying to access admin endpoint -> 403 Forbidden
	resp, data, _ = doReq(authApp, http.MethodGet, "/api/v1/auth/users", nil, aliceToken, nil)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("Expected 403 Forbidden for customer accessing admin route, got %d", resp.StatusCode)
	}

	// Admin accessing admin endpoint -> 200 OK
	resp, data, _ = doReq(authApp, http.MethodGet, "/api/v1/auth/users", nil, adminToken, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK for admin accessing /users, got %d", resp.StatusCode)
	}

	// 4. Create Customers
	resp, data, _ = doReq(custApp, http.MethodPost, "/api/v1/customers", map[string]any{
		"user_id":      aliceUserID,
		"first_name":   "Alice",
		"last_name":    "Test",
		"phone_number": "1234567890",
		"address":      "Street A",
		"national_id":  "ID-ALICE-1",
	}, aliceToken, nil)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create customer for Alice: %v", data)
	}
	aliceCustID := uint(data["data"].(map[string]any)["id"].(float64))

	resp, data, _ = doReq(custApp, http.MethodPost, "/api/v1/customers", map[string]any{
		"user_id":      bobUserID,
		"first_name":   "Bob",
		"last_name":    "Test",
		"phone_number": "0987654321",
		"address":      "Street B",
		"national_id":  "ID-BOB-2",
	}, bobToken, nil)
	bobCustID := uint(data["data"].(map[string]any)["id"].(float64))

	// 5. Test Level 2 Authorization: Alice cannot view Bob's customer profile
	resp, _, _ = doReq(custApp, http.MethodGet, fmt.Sprintf("/api/v1/customers/%d", bobCustID), nil, aliceToken, nil)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("Expected 403 Forbidden when Alice views Bob's customer profile, got %d", resp.StatusCode)
	}

	// 6. Create Accounts
	// Alice Account: ₹50,000 (5,000,000 paise)
	resp, data, _ = doReq(accApp, http.MethodPost, "/api/v1/accounts", map[string]any{
		"customer_id":     aliceCustID,
		"user_id":         aliceUserID,
		"branch_code":     "BLR01",
		"account_type":    "SAVINGS",
		"currency":        "INR",
		"initial_deposit": 50000.00,
	}, adminToken, nil)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create Alice account: %v", data)
	}
	aliceAccID := uint(data["data"].(map[string]any)["id"].(float64))

	// Bob Account: ₹10,000 (1,000,000 paise)
	resp, data, _ = doReq(accApp, http.MethodPost, "/api/v1/accounts", map[string]any{
		"customer_id":     bobCustID,
		"user_id":         bobUserID,
		"branch_code":     "BLR01",
		"account_type":    "SAVINGS",
		"currency":        "INR",
		"initial_deposit": 10000.00,
	}, adminToken, nil)
	bobAccID := uint(data["data"].(map[string]any)["id"].(float64))

	// 7. Test Level 2 Authorization on Accounts: Alice trying to view Bob's account details -> 403 Forbidden
	resp, _, _ = doReq(accApp, http.MethodGet, fmt.Sprintf("/api/v1/accounts/%d", bobAccID), nil, aliceToken, nil)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("Expected 403 Forbidden when Alice accesses Bob's account, got %d", resp.StatusCode)
	}

	// 8. Money Transfer: Alice transfers ₹10,000 to Bob with Idempotency Key
	idempKey := "trans-idemp-key-001"
	headers := map[string]string{
		"Idempotency-Key": idempKey,
		"X-Request-ID":    "req-test-trace-1234",
	}
	resp, data, _ = doReq(txApp, http.MethodPost, "/api/v1/transactions/transfer", txDto.TransferRequest{
		FromAccountID: aliceAccID,
		ToAccountID:   bobAccID,
		Amount:        10000.00,
		Currency:      "INR",
		Description:   "Dinner repayment",
	}, aliceToken, headers)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Transfer failed: %v", data)
	}

	// 9. Verify Balances after Transfer
	// Alice balance should be ₹40,000
	resp, data, _ = doReq(accApp, http.MethodGet, fmt.Sprintf("/api/v1/accounts/%d", aliceAccID), nil, aliceToken, nil)
	aliceBal := data["data"].(map[string]any)["balance"].(float64)
	if aliceBal != 40000.00 {
		t.Fatalf("Expected Alice balance to be 40000, got %.2f", aliceBal)
	}

	// Bob balance should be ₹20,000
	resp, data, _ = doReq(accApp, http.MethodGet, fmt.Sprintf("/api/v1/accounts/%d", bobAccID), nil, bobToken, nil)
	bobBal := data["data"].(map[string]any)["balance"].(float64)
	if bobBal != 20000.00 {
		t.Fatalf("Expected Bob balance to be 20000, got %.2f", bobBal)
	}

	// 10. Test Idempotency: Replaying exact same transfer request with same Idempotency-Key
	resp2, data2, _ := doReq(txApp, http.MethodPost, "/api/v1/transactions/transfer", txDto.TransferRequest{
		FromAccountID: aliceAccID,
		ToAccountID:   bobAccID,
		Amount:        10000.00,
		Currency:      "INR",
		Description:   "Dinner repayment",
	}, aliceToken, headers)

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("Idempotent replay failed: %v", data2)
	}

	// Balances MUST NOT change!
	resp, data, _ = doReq(accApp, http.MethodGet, fmt.Sprintf("/api/v1/accounts/%d", aliceAccID), nil, aliceToken, nil)
	aliceBal = data["data"].(map[string]any)["balance"].(float64)
	if aliceBal != 40000.00 {
		t.Fatalf("Idempotency violation! Alice was debited twice: %.2f", aliceBal)
	}

	// 11. Test Insufficient Funds: Alice attempts to transfer ₹100,000 (she only has ₹40,000)
	resp, data, _ = doReq(txApp, http.MethodPost, "/api/v1/transactions/transfer", txDto.TransferRequest{
		FromAccountID: aliceAccID,
		ToAccountID:   bobAccID,
		Amount:        100000.00,
		Currency:      "INR",
		Description:   "Overdrawn transfer",
	}, aliceToken, nil)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected 400 Bad Request for insufficient funds, got %d", resp.StatusCode)
	}
	errObj := data["error"].(map[string]any)
	if errObj["code"] != errors.ErrCodeInsufficientFunds {
		t.Fatalf("Expected error code INSUFFICIENT_FUNDS, got %v", errObj["code"])
	}

	// 12. Test Double-Entry Ledger Verification
	resp, data, _ = doReq(txApp, http.MethodGet, "/api/v1/transactions/ledger", nil, adminToken, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to fetch ledger: %v", data)
	}
	entries := data["data"].([]any)
	if len(entries) < 2 {
		t.Fatalf("Expected at least 2 ledger entries for transfer, got %d", len(entries))
	}

	// Verify Debit and Credit entries match ₹10,000
	var debitFound, creditFound bool
	for _, raw := range entries {
		entry := raw.(map[string]any)
		entryType := entry["entry_type"].(string)
		amt := entry["amount"].(float64)
		if entryType == "DEBIT" && amt == 10000.00 {
			debitFound = true
		}
		if entryType == "CREDIT" && amt == 10000.00 {
			creditFound = true
		}
	}

	if !debitFound || !creditFound {
		t.Fatalf("Double entry ledger mismatch! DebitFound: %v, CreditFound: %v", debitFound, creditFound)
	}

	t.Log(">>> All Banking Microservices Integration Tests Passed Successfully! <<<")
}
