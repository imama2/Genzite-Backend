package errors

import "errors"

var (
	ErrUserExists          = errors.New("user already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrGoogleNotConfigured = errors.New("google oauth is not configured")
	ErrGoogleAuthFailed    = errors.New("google authentication failed")
	ErrInvalidInput        = errors.New("invalid input")
	ErrSlugTaken           = errors.New("slug already taken")
	ErrSiteNotFound        = errors.New("site not found")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrRenderFailed        = errors.New("render failed")
	ErrNotPublished        = errors.New("site not published")
	ErrPaymentNotFound     = errors.New("payment not found")
	ErrInvalidSignature    = errors.New("invalid signature")
	ErrWebBuilderMissing   = errors.New("web-builder service not configured")
)
