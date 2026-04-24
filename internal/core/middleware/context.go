package middleware

import "context"

const (
	ContextUserIDKey      = "userID"
	ContextEmailKey       = "email"
	ContextRolesKey       = "roles"
	ContextPermissionsKey = "permissions"
)

type AuthContext struct {
	UserID      uint
	Email       string
	Roles       []string
	Permissions []string
}

type AuthProvider interface {
	LoadAuthContext(ctx context.Context, userID uint) (*AuthContext, error)
}
