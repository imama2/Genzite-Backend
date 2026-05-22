package repository

import (
	"context"

	"github.com/imama2/Genzite-Backend/internal/services/iam/models"
)

type Repository interface {
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByGoogleID(ctx context.Context, googleID string) (*models.User, error)
	GetUserByID(ctx context.Context, userID uint) (*models.User, error)
	GetUserWithRoles(ctx context.Context, userID uint) (*models.User, error)
}
