package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireRoles(required ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(required) == 0 {
			c.Next()
			return
		}

		userRoles := getStringSlice(c, ContextRolesKey)
		if !hasAny(userRoles, required) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

func RequirePermissions(required ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(required) == 0 {
			c.Next()
			return
		}

		userPermissions := getStringSlice(c, ContextPermissionsKey)
		if !hasAll(userPermissions, required) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

func getStringSlice(c *gin.Context, key string) []string {
	value, ok := c.Get(key)
	if !ok {
		return nil
	}

	slice, ok := value.([]string)
	if !ok {
		return nil
	}

	return slice
}

func hasAny(values []string, required []string) bool {
	if len(values) == 0 {
		return false
	}

	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}

	for _, req := range required {
		if _, ok := set[req]; ok {
			return true
		}
	}

	return false
}

func hasAll(values []string, required []string) bool {
	if len(required) == 0 {
		return true
	}

	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}

	for _, req := range required {
		if _, ok := set[req]; !ok {
			return false
		}
	}

	return true
}
