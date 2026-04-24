package middleware

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/imama2/Genzite-Backend/internal/core/config"
)

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func JWTAuth(cfg *config.Config, provider AuthProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		if provider == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "auth provider not configured"})
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		userID, err := strconv.ParseUint(claims.Subject, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token subject"})
			return
		}

		authContext, err := provider.LoadAuthContext(c.Request.Context(), uint(userID))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set(ContextUserIDKey, authContext.UserID)
		c.Set(ContextEmailKey, authContext.Email)
		c.Set(ContextRolesKey, authContext.Roles)
		c.Set(ContextPermissionsKey, authContext.Permissions)
		c.Next()
	}
}
