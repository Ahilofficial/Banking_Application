package repository

import (
	"errors"
	"fmt"
	"log"
	"time"

	"banking-microservices/services/auth-service/models"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet provides AuthRepository dependency for Wire
var ProviderSet = wire.NewSet(ProvideAuthRepository)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

// ProvideAuthRepository creates, auto-migrates, and seeds the AuthRepository
func ProvideAuthRepository(db *gorm.DB) (*AuthRepository, error) {
	repo := NewAuthRepository(db)
	if err := repo.AutoMigrate(); err != nil {
		return nil, fmt.Errorf("auth-service db migration failed: %w", err)
	}
	if err := repo.SeedDefaultRolesAndPermissions(); err != nil {
		log.Printf("[auth-service] Seeding roles warning: %v", err)
	}
	return repo, nil
}

func (r *AuthRepository) AutoMigrate() error {
	return r.db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.UserRole{},
		&models.RolePermission{},
		&models.Session{},
	)
}

func (r *AuthRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *AuthRepository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Roles.Permissions").Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) FindUserByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Roles.Permissions").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) ListUsers() ([]models.User, error) {
	var users []models.User
	err := r.db.Preload("Roles.Permissions").Find(&users).Error
	return users, err
}

func (r *AuthRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *AuthRepository) CreateSession(session *models.Session) error {
	return r.db.Create(session).Error
}

func (r *AuthRepository) FindSession(token string) (*models.Session, error) {
	var session models.Session
	err := r.db.Where("refresh_token = ?", token).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *AuthRepository) RevokeSession(token string) error {
	return r.db.Model(&models.Session{}).Where("refresh_token = ?", token).Update("is_revoked", true).Error
}

func (r *AuthRepository) FindRoleByName(name string) (*models.Role, error) {
	var role models.Role
	err := r.db.Preload("Permissions").Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *AuthRepository) AssignRole(userID, roleID uint) error {
	var existing models.UserRole
	err := r.db.Where("user_id = ? AND role_id = ?", userID, roleID).First(&existing).Error
	if err == nil {
		return nil // already assigned
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return r.db.Create(&models.UserRole{UserID: userID, RoleID: roleID}).Error
}

func (r *AuthRepository) GetPermissionsForUser(userID uint) ([]string, error) {
	var permissions []string
	query := `
		SELECT DISTINCT p.name
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN roles r ON rp.role_id = r.id
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = ?
	`
	err := r.db.Raw(query, userID).Scan(&permissions).Error
	return permissions, err
}

func (r *AuthRepository) GetRolesForUser(userID uint) ([]string, error) {
	var roles []string
	query := `
		SELECT r.name
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = ?
	`
	err := r.db.Raw(query, userID).Scan(&roles).Error
	return roles, err
}

// SeedDefaultRolesAndPermissions sets up initial RBAC matrix
func (r *AuthRepository) SeedDefaultRolesAndPermissions() error {
	permissions := []models.Permission{
		{Name: "customer:create", Description: "Create new customer profile"},
		{Name: "customer:view", Description: "View customer profiles"},
		{Name: "customer:update", Description: "Update customer details / KYC"},
		{Name: "customer:delete", Description: "Delete or archive customer"},

		{Name: "account:create", Description: "Open new bank account"},
		{Name: "account:view", Description: "View account balances and details"},
		{Name: "account:update", Description: "Update account settings"},
		{Name: "account:block", Description: "Block or suspend an account"},
		{Name: "account:unblock", Description: "Unblock an account"},
		{Name: "account:close", Description: "Close an account"},

		{Name: "transaction:create", Description: "Initiate money transfer, deposit or withdrawal"},
		{Name: "transaction:view", Description: "View transactions"},
		{Name: "transaction:approve", Description: "Approve large transactions"},
		{Name: "transaction:reverse", Description: "Reverse a failed or fraudulent transaction"},

		{Name: "ledger:view", Description: "Inspect double-entry ledger"},
		{Name: "audit:view", Description: "View system audit trail"},
		{Name: "user:manage", Description: "Manage users and assignments"},
		{Name: "role:manage", Description: "Manage roles and permissions"},
	}

	permMap := make(map[string]models.Permission)
	for _, p := range permissions {
		var existing models.Permission
		if err := r.db.Where("name = ?", p.Name).First(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				p.CreatedAt = time.Now()
				r.db.Create(&p)
				permMap[p.Name] = p
			}
		} else {
			permMap[p.Name] = existing
		}
	}

	// Roles and their permission matrix
	roleDefs := map[string][]string{
		"SUPER_ADMIN": {
			"customer:create", "customer:view", "customer:update", "customer:delete",
			"account:create", "account:view", "account:update", "account:block", "account:unblock", "account:close",
			"transaction:create", "transaction:view", "transaction:approve", "transaction:reverse",
			"ledger:view", "audit:view", "user:manage", "role:manage",
		},
		"BANK_ADMIN": {
			"customer:create", "customer:view", "customer:update",
			"account:create", "account:view", "account:update", "account:block", "account:unblock",
			"transaction:view", "transaction:approve", "ledger:view", "audit:view", "user:manage",
		},
		"BRANCH_MANAGER": {
			"customer:view", "customer:update",
			"account:create", "account:view", "account:block", "account:unblock",
			"transaction:view", "transaction:approve", "ledger:view",
		},
		"BANK_EMPLOYEE": {
			"customer:create", "customer:view",
			"account:create", "account:view",
			"transaction:view",
		},
		"TELLER": {
			"customer:view",
			"account:view",
			"transaction:create",
			"transaction:view",
		},
		"CUSTOMER": {
			"customer:view",
			"account:view",
			"transaction:create",
			"transaction:view",
		},
	}

	for roleName, perms := range roleDefs {
		var role models.Role
		if err := r.db.Where("name = ?", roleName).First(&role).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				role = models.Role{
					Name:        roleName,
					Description: "Standard role for " + roleName,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := r.db.Create(&role).Error; err != nil {
					return err
				}
			}
		}

		// Link permissions
		for _, permName := range perms {
			if perm, ok := permMap[permName]; ok {
				var rp models.RolePermission
				if err := r.db.Where("role_id = ? AND permission_id = ?", role.ID, perm.ID).First(&rp).Error; errors.Is(err, gorm.ErrRecordNotFound) {
					r.db.Create(&models.RolePermission{
						RoleID:       role.ID,
						PermissionID: perm.ID,
					})
				}
			}
		}
	}

	return nil
}
