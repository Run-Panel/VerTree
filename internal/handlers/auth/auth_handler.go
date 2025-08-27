package auth

import (
	"net/http"
	"strconv"

	"github.com/Run-Panel/VerTree/internal/middleware"
	"github.com/Run-Panel/VerTree/internal/models"
	"github.com/Run-Panel/VerTree/internal/services"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(jwtSecret string) *AuthHandler {
	return &AuthHandler{
		authService: services.NewAuthService(jwtSecret),
	}
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid request format", err))
		return
	}

	// Validate request
	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Username and password are required", nil))
		return
	}

	response, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCodeAndError(
			http.StatusUnauthorized,
			err.Error(),
			"LOGIN_FAILED",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponseWithMessage("Login successful", response))
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid request format", err))
		return
	}

	response, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCodeAndError(
			http.StatusUnauthorized,
			err.Error(),
			"TOKEN_REFRESH_FAILED",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponseWithMessage("Token refreshed successfully", response))
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid request format", err))
		return
	}

	err := h.authService.Logout(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCodeAndError(
			http.StatusInternalServerError,
			"Failed to logout",
			"LOGOUT_FAILED",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponseWithMessage("Logout successful", nil))
}

// GetProfile returns the current user's profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCodeAndError(
			http.StatusUnauthorized,
			"User not authenticated",
			"NOT_AUTHENTICATED",
		))
		return
	}

	adminInfo, err := h.authService.GetAdminByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponseWithCodeAndError(
			http.StatusNotFound,
			"User not found",
			"USER_NOT_FOUND",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponseWithMessage("Profile retrieved successfully", adminInfo))
}

// ChangePassword handles password changes
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCodeAndError(
			http.StatusUnauthorized,
			"User not authenticated",
			"NOT_AUTHENTICATED",
		))
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid request format", err))
		return
	}

	err := h.authService.ChangePassword(userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCodeAndError(
			http.StatusBadRequest,
			err.Error(),
			"PASSWORD_CHANGE_FAILED",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponseWithMessage("Password changed successfully", nil))
}

// CreateAdmin creates a new admin user (superadmin only)
func (h *AuthHandler) CreateAdmin(c *gin.Context) {
	var req models.CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid request format", err))
		return
	}

	adminInfo, err := h.authService.CreateAdmin(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCodeAndError(
			http.StatusBadRequest,
			err.Error(),
			"ADMIN_CREATION_FAILED",
		))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponseWithMessage("Admin created successfully", adminInfo))
}

// ListAdmins returns all admin users (superadmin only)
func (h *AuthHandler) ListAdmins(c *gin.Context) {
	admins, err := h.authService.ListAdmins()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCodeAndError(
			http.StatusInternalServerError,
			"Failed to retrieve admins",
			"ADMIN_LIST_FAILED",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponseWithMessage("Admins retrieved successfully", admins))
}

// GetAdmin returns a specific admin user (superadmin only)
func (h *AuthHandler) GetAdmin(c *gin.Context) {
	adminIDStr := c.Param("id")
	adminID, err := strconv.ParseUint(adminIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid admin ID", nil))
		return
	}

	adminInfo, err := h.authService.GetAdminByID(uint(adminID))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponseWithCodeAndError(
			http.StatusNotFound,
			err.Error(),
			"ADMIN_NOT_FOUND",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponseWithMessage("Admin retrieved successfully", adminInfo))
}

// UpdateAdmin updates an admin user (superadmin only)
func (h *AuthHandler) UpdateAdmin(c *gin.Context) {
	adminIDStr := c.Param("id")
	adminID, err := strconv.ParseUint(adminIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid admin ID", nil))
		return
	}

	var req models.UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid request format", err))
		return
	}

	adminInfo, err := h.authService.UpdateAdmin(uint(adminID), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCodeAndError(
			http.StatusBadRequest,
			err.Error(),
			"ADMIN_UPDATE_FAILED",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponseWithMessage("Admin updated successfully", adminInfo))
}

// DeleteAdmin deletes an admin user (superadmin only)
func (h *AuthHandler) DeleteAdmin(c *gin.Context) {
	adminIDStr := c.Param("id")
	adminID, err := strconv.ParseUint(adminIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid admin ID", nil))
		return
	}

	// Prevent deleting self
	currentUserID, exists := middleware.GetCurrentUserID(c)
	if exists && currentUserID == uint(adminID) {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCodeAndError(
			http.StatusBadRequest,
			"Cannot delete your own account",
			"CANNOT_DELETE_SELF",
		))
		return
	}

	err = h.authService.DeleteAdmin(uint(adminID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCodeAndError(
			http.StatusInternalServerError,
			"Failed to delete admin",
			"ADMIN_DELETE_FAILED",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponseWithMessage("Admin deleted successfully", nil))
}
