package main

import (
	"log"
	"time"

	"banking-microservices/pkg/common/config"
	"banking-microservices/pkg/common/database"
	"banking-microservices/pkg/common/money"
	"banking-microservices/pkg/common/security"

	accModels "banking-microservices/services/account-service/models"
	authModels "banking-microservices/services/auth-service/models"
	authRepo "banking-microservices/services/auth-service/repository"
	custModels "banking-microservices/services/customer-service/models"
	txModels "banking-microservices/services/transaction-service/models"
)

func main() {
	cfg := config.LoadConfig("seeder", 0)

	log.Printf("[Seeder] Connecting to MySQL using driver: %s, dsn: %s", cfg.DBDriver, cfg.DBDSN)
	db, err := database.Connect(cfg.DBDriver, cfg.DBDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate tables
	if err := db.AutoMigrate(
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
	); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// 1. Seed Roles and Permissions
	ar := authRepo.NewAuthRepository(db)
	if err := ar.SeedDefaultRolesAndPermissions(); err != nil {
		log.Fatalf("Roles seeding failed: %v", err)
	}
	log.Println("[Seeder] Roles and Permissions verified/seeded successfully.")

	// Helper to create user if not exists
	createUser := func(email, password, name, roleName string) *authModels.User {
		var user authModels.User
		if err := db.Where("email = ?", email).First(&user).Error; err == nil {
			return &user
		}

		hash, _ := security.HashPassword(password)
		role, err := ar.FindRoleByName(roleName)
		if err != nil {
			log.Fatalf("Role %s not found: %v", roleName, err)
		}

		user = authModels.User{
			Email:        email,
			PasswordHash: hash,
			FullName:     name,
			Status:       "ACTIVE",
			Roles:        []authModels.Role{*role},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		db.Create(&user)
		_ = ar.AssignRole(user.ID, role.ID)
		log.Printf("[Seeder] User created: %s (%s)", email, roleName)
		return &user
	}

	// 2. Create Initial Banking Actors
	adminUser := createUser("admin@bank.com", "Admin@123", "Super Administrator", "SUPER_ADMIN")
	tellerUser := createUser("teller@bank.com", "Teller@123", "Branch Teller John", "TELLER")
	aliceUser := createUser("alice@bank.com", "Alice@123", "Alice Sharma", "CUSTOMER")
	bobUser := createUser("bob@bank.com", "Bob@123", "Bob Verma", "CUSTOMER")

	// Helper to create customer profile
	createCustomer := func(user *authModels.User, first, last, phone, address, nationalID string) *custModels.Customer {
		var cust custModels.Customer
		if err := db.Where("user_id = ?", user.ID).First(&cust).Error; err == nil {
			return &cust
		}
		cust = custModels.Customer{
			UserID:      user.ID,
			FirstName:   first,
			LastName:    last,
			PhoneNumber: phone,
			Address:     address,
			NationalID:  nationalID,
			KYCStatus:   custModels.KYCStatusVerified,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		db.Create(&cust)
		log.Printf("[Seeder] Customer profile created for user %s (ID: %d)", user.Email, cust.ID)
		return &cust
	}

	aliceCust := createCustomer(aliceUser, "Alice", "Sharma", "+91-9876543210", "123 MG Road, Bangalore", "ABCDE1234F")
	bobCust := createCustomer(bobUser, "Bob", "Verma", "+91-9876543211", "456 Connaught Place, New Delhi", "XYZAB5678C")

	// Helper to create account
	createAccount := func(cust *custModels.Customer, accNum, branch string, accType accModels.AccountType, initialDeposit float64) *accModels.Account {
		var acc accModels.Account
		if err := db.Where("account_number = ?", accNum).First(&acc).Error; err == nil {
			return &acc
		}

		bal := money.FromDecimal(initialDeposit)
		acc = accModels.Account{
			AccountNumber: accNum,
			CustomerID:    cust.ID,
			UserID:        cust.UserID,
			BranchCode:    branch,
			AccountType:   accType,
			Currency:      "INR",
			Balance:       bal,
			Status:        accModels.AccountStatusActive,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		db.Create(&acc)
		log.Printf("[Seeder] Bank Account created: %s, Owner UserID: %d, Balance: %s", accNum, acc.UserID, bal.String())
		return &acc
	}

	// 3. Create Accounts
	// Alice: Account A = ₹50,000
	createAccount(aliceCust, "AC-MAIN-1000000001", "MAIN01", accModels.AccountTypeSavings, 50000.00)

	// Bob: Account B = ₹10,000
	createAccount(bobCust, "AC-MAIN-1000000002", "MAIN01", accModels.AccountTypeSavings, 10000.00)

	_ = adminUser
	_ = tellerUser

	log.Println("[Seeder] Seeding finished successfully on MySQL! All actors and accounts initialized.")
}
