package middleware

import (
	"net/http"
	"strings"

	"github.com/Run-Panel/VerTree/internal/models"
	"github.com/Run-Panel/VerTree/internal/services"
	"github.com/gin-gonic/gin"
)

// APIKeyAuth creates a middleware that validates API key authentication for external clients
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.UnauthorizedResponse("Missing Authorization header"))
			c.Abort()
			return
		}

		// Check if it starts with "Bearer "
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.JSON(http.StatusUnauthorized, models.UnauthorizedResponse("Invalid authorization format. Use 'Bearer <app_id>:<api_key>'"))
			c.Abort()
			return
		}

		// Extract the token part
		token := strings.TrimPrefix(authHeader, bearerPrefix)

		// Parse app_id and api_key
		parts := strings.SplitN(token, ":", 2)
		if len(parts) != 2 {
			c.JSON(http.StatusUnauthorized, models.UnauthorizedResponse("Invalid token format. Use '<app_id>:<api_key>'"))
			c.Abort()
			return
		}

		appID := parts[0]
		apiKey := parts[1]

		if appID == "" || apiKey == "" {
			c.JSON(http.StatusUnauthorized, models.UnauthorizedResponse("App ID and API key cannot be empty"))
			c.Abort()
			return
		}

		// Validate the API key
		appService := services.NewApplicationService()
		app, key, err := appService.ValidateAPIKey(appID, apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.UnauthorizedResponse("Invalid credentials"))
			c.Abort()
			return
		}

		// Store application and key info in context for use in handlers
		c.Set("app_id", app.AppID)
		c.Set("app", app)
		c.Set("api_key", key)
		c.Set("api_key_permissions", key.Permissions)

		c.Next()
	}
}

// RequirePermission creates a middleware that checks if the API key has the required permission
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("api_key_permissions")
		if !exists {
			c.JSON(http.StatusForbidden, models.ErrorResponseWithCode(403, "Permissions not found in context", nil))
			c.Abort()
			return
		}

		permissionList, ok := permissions.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, models.ErrorResponseWithCode(403, "Invalid permissions format", nil))
			c.Abort()
			return
		}

		// Check if the required permission exists
		hasPermission := false
		for _, p := range permissionList {
			if p == permission || p == "*" { // "*" means all permissions
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, models.ErrorResponseWithCode(403, "Insufficient permissions", nil))
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAPIKeyAuth creates a middleware that optionally validates API key but doesn't fail if missing
// This is useful for endpoints that can work both with and without authentication
func OptionalAPIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth header, continue without authentication
			c.Next()
			return
		}

		// Check if it starts with "Bearer "
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			// Invalid format but optional, continue without authentication
			c.Next()
			return
		}

		// Extract the token part
		token := strings.TrimPrefix(authHeader, bearerPrefix)

		// Parse app_id and api_key
		parts := strings.SplitN(token, ":", 2)
		if len(parts) != 2 {
			// Invalid format but optional, continue without authentication
			c.Next()
			return
		}

		appID := parts[0]
		apiKey := parts[1]

		if appID == "" || apiKey == "" {
			// Empty credentials but optional, continue without authentication
			c.Next()
			return
		}

		// Validate the API key
		appService := services.NewApplicationService()
		app, key, err := appService.ValidateAPIKey(appID, apiKey)
		if err != nil {
			// Invalid credentials but optional, continue without authentication
			c.Next()
			return
		}

		// Store application and key info in context for use in handlers
		c.Set("app_id", app.AppID)
		c.Set("app", app)
		c.Set("api_key", key)
		c.Set("api_key_permissions", key.Permissions)

		c.Next()
	}
}
