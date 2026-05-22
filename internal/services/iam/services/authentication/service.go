package authentication

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/imama2/Genzite-Backend/internal/core/middleware"
	"github.com/imama2/Genzite-Backend/internal/services/iam/models"
	"github.com/imama2/Genzite-Backend/internal/services/iam/repository"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

var (
	ErrUserExists          = errors.New("user already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrGoogleNotConfigured = errors.New("google oauth is not configured")
	ErrGoogleAuthFailed    = errors.New("google authentication failed")
)

type Service struct {
	repo        repository.Repository
	cfg         *config.Config
	logger      *slog.Logger
	oauthConfig *oauth2.Config
	httpClient  *http.Client
}

type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

type LoginInput struct {
	Email    string
	Password string
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	VerifiedEmail bool   `json:"verified_email"`
}

func New(cfg *config.Config, repo repository.Repository, logger *slog.Logger) *Service {
	service := &Service{
		repo:       repo,
		cfg:        cfg,
		logger:     logger,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}

	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" && cfg.GoogleRedirectURL != "" {
		service.oauthConfig = &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Endpoint:     google.Endpoint,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
		}
	}

	return service
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*models.User, string, error) {
	email := normalizeEmail(input.Email)
	if email == "" || input.Password == "" {
		return nil, "", ErrInvalidCredentials
	}

	_, err := s.repo.GetUserByEmail(ctx, email)
	if err == nil {
		return nil, "", ErrUserExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("hash password: %w", err)
	}

	passwordHash := string(hash)
	user := &models.User{
		Email:        email,
		Name:         strings.TrimSpace(input.Name),
		PasswordHash: &passwordHash,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, "", err
	}

	token, err := s.issueToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*models.User, string, error) {
	email := normalizeEmail(input.Email)
	if email == "" || input.Password == "" {
		return nil, "", ErrInvalidCredentials
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	if user.PasswordHash == nil {
		return nil, "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	token, err := s.issueToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *Service) GoogleAuthURL(state string) (string, error) {
	if s.oauthConfig == nil {
		return "", ErrGoogleNotConfigured
	}
	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

func (s *Service) HandleGoogleCallback(ctx context.Context, code string) (*models.User, string, error) {
	if s.oauthConfig == nil {
		return nil, "", ErrGoogleNotConfigured
	}

	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, "", ErrGoogleAuthFailed
	}

	client := s.oauthConfig.Client(ctx, token)
	client.Timeout = s.httpClient.Timeout

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, "", ErrGoogleAuthFailed
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", ErrGoogleAuthFailed
	}

	var info GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, "", ErrGoogleAuthFailed
	}

	email := normalizeEmail(info.Email)
	if info.ID == "" || email == "" {
		return nil, "", ErrGoogleAuthFailed
	}

	user, err := s.repo.GetUserByGoogleID(ctx, info.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", err
	}

	if user == nil {
		userByEmail, err := s.repo.GetUserByEmail(ctx, email)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", err
		}

		if userByEmail != nil {
			userByEmail.GoogleID = &info.ID
			if userByEmail.Name == "" {
				userByEmail.Name = strings.TrimSpace(info.Name)
			}
			if err := s.repo.UpdateUser(ctx, userByEmail); err != nil {
				return nil, "", err
			}
			user = userByEmail
		} else {
			user = &models.User{
				Email:    email,
				Name:     strings.TrimSpace(info.Name),
				GoogleID: &info.ID,
			}
			if err := s.repo.CreateUser(ctx, user); err != nil {
				return nil, "", err
			}
		}
	}

	jwtToken, err := s.issueToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, jwtToken, nil
}

func (s *Service) LoadAuthContext(ctx context.Context, userID uint) (*middleware.AuthContext, error) {
	user, err := s.repo.GetUserWithRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	roles := make([]string, 0, len(user.Roles))
	permissions := make([]string, 0)
	seenPermissions := make(map[string]struct{})

	for _, role := range user.Roles {
		roles = append(roles, role.Name)
		for _, permission := range role.Permissions {
			if _, ok := seenPermissions[permission.Name]; !ok {
				seenPermissions[permission.Name] = struct{}{}
				permissions = append(permissions, permission.Name)
			}
		}
	}

	return &middleware.AuthContext{
		UserID:      user.ID,
		Email:       user.Email,
		Roles:       roles,
		Permissions: permissions,
	}, nil
}

func (s *Service) GetUserProfile(ctx context.Context, userID uint) (*models.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

func (s *Service) issueToken(user *models.User) (string, error) {
	now := time.Now()
	claims := middleware.Claims{
		Email: user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.cfg.JWTIssuer,
			Subject:   fmt.Sprintf("%d", user.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.JWTTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
