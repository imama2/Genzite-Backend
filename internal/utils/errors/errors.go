package errors

import "errors"

var (
	ErrUserExists          = errors.New("user already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrGoogleNotConfigured = errors.New("google oauth is not configured")
	ErrGoogleAuthFailed    = errors.New("google authentication failed")
)
