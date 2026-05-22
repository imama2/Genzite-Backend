package repository

import (
	"context"

	dbentities "github.com/imama2/Genzite-Backend/internal/services/iam/models/entities"
)

type Repository interface {
	CreateUser(ctx context.Context, user *dbentities.Users) error
	UpdateUser(ctx context.Context, user *dbentities.Users) error
	GetUserByEmail(ctx context.Context, email string) (*dbentities.Users, error)
	GetUserByGoogleID(ctx context.Context, googleID string) (*dbentities.Users, error)
	GetUserByID(ctx context.Context, userID uint) (*dbentities.Users, error)
	GetUserWithRoles(ctx context.Context, userID uint) (*dbentities.Users, error)
}
