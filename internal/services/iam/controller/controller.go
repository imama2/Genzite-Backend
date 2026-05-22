package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/imama2/Genzite-Backend/internal/core/middleware"
	"github.com/imama2/Genzite-Backend/internal/services/iam/models/dto"
	"github.com/imama2/Genzite-Backend/internal/services/iam/services/authentication"
	utilErrors "github.com/imama2/Genzite-Backend/internal/utils/errors"
	"github.com/imama2/Genzite-Backend/internal/utils/responses"
)

type AuthController struct {
	cfg     *config.Config
	service *authentication.Service
}

func NewAuthController(cfg *config.Config, service *authentication.Service) *AuthController {
	return &AuthController{
		cfg:     cfg,
		service: service,
	}
}

func (h *AuthController) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, token, err := h.service.Register(c.Request.Context(), dto.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	})
	if err != nil {
		switch {
		case errors.Is(err, utilErrors.ErrUserExists):
			c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
		case errors.Is(err, utilErrors.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register"})
		}
		return
	}

	responses.Ok(c, "user registered successfully", dto.AuthResponse{
		Token: token,
		User:  dto.UserPayloadResponse{ID: user.ID, Email: user.Email, Name: user.Name},
	})
}

func (h *AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, token, err := h.service.Login(c.Request.Context(), dto.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, utilErrors.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		}
		return
	}

	responses.Ok(c, "login successful", dto.AuthResponse{
		Token: token,
		User:  dto.UserPayloadResponse{ID: user.ID, Email: user.Email, Name: user.Name},
	})
}

func (h *AuthController) GoogleLogin(c *gin.Context) {
	state := generateState()
	authURL, err := h.service.GoogleOAuth(c.Request.Context(), state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	secure := h.cfg.Env == "production"
	c.SetCookie("oauth_state", state, 300, "/", "", secure, true)
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

func (h *AuthController) GoogleCallback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")
	if state == "" || code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing oauth state or code"})
		return
	}

	cookieState, err := c.Cookie("oauth_state")
	if err != nil || cookieState == "" || cookieState != state {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid oauth state"})
		return
	}

	user, token, err := h.service.HandleGoogleCallback(c.Request.Context(), code)
	if err != nil {
		switch {
		case errors.Is(err, utilErrors.ErrGoogleNotConfigured):
			c.JSON(http.StatusBadRequest, gin.H{"error": "google oauth is not configured"})
		default:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "google authentication failed"})
		}
		return
	}

	c.SetCookie("oauth_state", "", -1, "/", "", h.cfg.Env == "production", true)
	responses.Ok(c, "google login successful", dto.AuthResponse{
		Token: token,
		User:  dto.UserPayloadResponse{ID: user.ID, Email: user.Email, Name: user.Name},
	})
}

func (h *AuthController) Me(c *gin.Context) {
	userIDRaw, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDRaw.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := h.service.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load profile"})
		return
	}

	responses.Ok(c, "profile loaded successfully", dto.UserPayloadResponse{ID: user.ID, Email: user.Email, Name: user.Name})
}

func generateState() string {
	bytes := make([]byte, 16)
	if _, err := randReader.Read(bytes); err != nil {
		return "state"
	}
	return encodeState(bytes)
}
