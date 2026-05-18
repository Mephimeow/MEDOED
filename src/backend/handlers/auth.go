package handlers

import (
	"net/http"
	"strings"

	"github.com/Mephimeow/MEDOED/backend/models"
	"github.com/gin-gonic/gin"
)

var (
	authEnabled bool
	validKeys   []string
)

func InitAuth(enabled bool, apiKeys []string) {
	authEnabled = enabled
	validKeys = apiKeys
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !authEnabled || len(validKeys) == 0 {
			c.Next()
			return
		}

		key := extractAPIKey(c)
		if key == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Error: "missing API key"})
			return
		}

		if !isValidKey(key) {
			c.AbortWithStatusJSON(http.StatusForbidden, models.ErrorResponse{Error: "invalid API key"})
			return
		}

		c.Next()
	}
}

func extractAPIKey(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	apiKey := c.GetHeader("X-API-Key")
	if apiKey != "" {
		return apiKey
	}

	return c.Query("api_key")
}

func isValidKey(key string) bool {
	for _, k := range validKeys {
		if k == key {
			return true
		}
	}
	return false
}
