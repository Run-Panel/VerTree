package middleware

import (
	"net/http"

	"github.com/Run-Panel/VerTree/internal/models"
	"github.com/Run-Panel/VerTree/internal/utils"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware creates a JWT authentication middleware
func AuthMiddleware(jwtManager *utils.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCodeAndError(
				http.StatusUnauthorized,
				"Authorization header is required",
				"MISSING_AUTH_HEADER",
			))
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		token := utils.ExtractTokenFromHeader(authHeader)
		if token == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCodeAndError(
				http.StatusUnauthorized,
				"Invalid authorization header format",
				"INVALID_AUTH_HEADER",
			))
			c.Abort()
			return
		}

		// Validate token
		claims, err := jwtManager.ValidateAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCodeAndError(
				http.StatusUnauthorized,
				"Invalid or expired token",
				"INVALID_TOKEN",
			))
			c.Abort()
			return
		}

		// Store user information in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}

// RequireRole creates a middleware that requires specific roles
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCodeAndError(
				http.StatusUnauthorized,
				"User role not found in context",
				"MISSING_USER_ROLE",
			))
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCodeAndError(
				http.StatusInternalServerError,
				"Invalid user role type",
				"INVALID_ROLE_TYPE",
			))
			c.Abort()
			return
		}

		// Check if user has required role
		hasRole := false
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, models.ErrorResponseWithCodeAndError(
				http.StatusForbidden,
				"Insufficient permissions",
				"INSUFFICIENT_PERMISSIONS",
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireSuperAdmin requires superadmin role
func RequireSuperAdmin() gin.HandlerFunc {
	return RequireRole("superadmin")
}

// RequireAdmin requires admin or superadmin role
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin", "superadmin")
}

// OptionalAuth makes authentication optional (for endpoints that work with or without auth)
func OptionalAuth(jwtManager *utils.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			// No auth provided, continue without user context
			c.Next()
			return
		}

		token := utils.ExtractTokenFromHeader(authHeader)
		if token == "" {
			// Invalid header format, continue without user context
			c.Next()
			return
		}

		// Try to validate token
		claims, err := jwtManager.ValidateAccessToken(token)
		if err != nil {
			// Invalid token, continue without user context
			c.Next()
			return
		}

		// Store user information in context if token is valid
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}

// GetCurrentUser extracts current user information from context
func GetCurrentUser(c *gin.Context) (*models.JWTClaims, bool) {
	claims, exists := c.Get("jwt_claims")
	if !exists {
		return nil, false
	}

	jwtClaims, ok := claims.(*models.JWTClaims)
	return jwtClaims, ok
}

// GetCurrentUserID extracts current user ID from context
func GetCurrentUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	id, ok := userID.(uint)
	return id, ok
}
