package seed

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/imama2/Genzite-Backend/internal/core/config"
	dbentities "github.com/imama2/Genzite-Backend/internal/services/iam/models/entities"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	adminRoleName        = "admin"
	migrationsPermission = "migrations:*"
	adminRoleDescription = "System administrator"
	migrationsPermDesc   = "Run migrations and seeders"
	defaultAdminName     = "Administrator"
)

func IAM(ctx context.Context, db *gorm.DB, cfg *config.Config, logger *slog.Logger) error {
	email := strings.ToLower(strings.TrimSpace(cfg.AdminEmail))
	if email == "" || cfg.AdminPassword == "" {
		return errors.New("ADMIN_EMAIL and ADMIN_PASSWORD are required to seed admin user")
	}

	adminName := strings.TrimSpace(cfg.AdminName)
	if adminName == "" {
		adminName = defaultAdminName
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		permission, err := ensurePermission(tx, migrationsPermission, migrationsPermDesc)
		if err != nil {
			return err
		}

		role, err := ensureRole(tx, adminRoleName, adminRoleDescription)
		if err != nil {
			return err
		}

		if err := tx.Model(role).Association("Permissions").Append(permission); err != nil {
			return fmt.Errorf("assign permission: %w", err)
		}

		user, created, err := ensureAdminUser(tx, email, adminName, cfg.AdminPassword)
		if err != nil {
			return err
		}

		if err := tx.Model(user).Association("Roles").Append(role); err != nil {
			return fmt.Errorf("assign admin role: %w", err)
		}

		if created {
			logger.Info("seeded admin user", "email", user.Email)
		}

		return nil
	})
}

func ensurePermission(tx *gorm.DB, name, description string) (*dbentities.Permissions, error) {
	var perm dbentities.Permissions
	err := tx.Where("name = ?", name).First(&perm).Error
	if err == nil {
		return &perm, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	perm = dbentities.Permissions{Name: name, Description: description}
	if err := tx.Create(&perm).Error; err != nil {
		return nil, err
	}
	return &perm, nil
}

func ensureRole(tx *gorm.DB, name, description string) (*dbentities.Roles, error) {
	var role dbentities.Roles
	err := tx.Where("name = ?", name).First(&role).Error
	if err == nil {
		return &role, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	role = dbentities.Roles{Name: name, Description: description}
	if err := tx.Create(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func ensureAdminUser(tx *gorm.DB, email, name, password string) (*dbentities.Users, bool, error) {
	var user dbentities.Users
	err := tx.Where("email = ?", email).First(&user).Error
	if err == nil {
		return &user, false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, false, fmt.Errorf("hash admin password: %w", err)
	}

	passwordHash := string(hash)
	user = dbentities.Users{
		Email:        email,
		Name:         name,
		PasswordHash: &passwordHash,
	}

	if err := tx.Create(&user).Error; err != nil {
		return nil, false, err
	}

	return &user, true, nil
}
