package authentication

import (
	"context"

	"github.com/imama2/Genzite-Backend/internal/services/iam/models/dto"
	dbentities "github.com/imama2/Genzite-Backend/internal/services/iam/models/entities"
)

type AuthServiceInterface interface {
	Register(ctx context.Context, input dto.RegisterRequest) (*dbentities.Users, string, error)
	Login(ctx context.Context, input dto.LoginRequest) (*dbentities.Users, string, error)
	GoogleOAuth(ctx context.Context, state string) (string, error)
}
