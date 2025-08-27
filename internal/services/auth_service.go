package services

import (
	"errors"
	"time"

	"github.com/Run-Panel/VerTree/internal/database"
	"github.com/Run-Panel/VerTree/internal/models"
	"github.com/Run-Panel/VerTree/internal/utils"
	"gorm.io/gorm"
)

// AuthService handles authentication operations
type AuthService struct {
	jwtManager *utils.JWTManager
}

// NewAuthService creates a new auth service
func NewAuthService(jwtSecret string) *AuthService {
	return &AuthService{
		jwtManager: utils.NewJWTManager(jwtSecret),
	}
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(username, password string) (*models.LoginResponse, error) {
	db := database.GetDB()

	// Find user by username or email
	var admin models.Admin
	err := db.Where("username = ? OR email = ?", username, username).First(&admin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	// Check if account is active
	if !admin.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Verify password
	isValid, err := utils.VerifyPassword(password, admin.Password)
	if err != nil {
		return nil, err
	}
	if !isValid {
		return nil, errors.New("invalid credentials")
	}

	// Generate tokens
	accessToken, refreshToken, expiresAt, err := s.jwtManager.GenerateTokenPair(&admin)
	if err != nil {
		return nil, err
	}

	// Store refresh token in database
	refreshTokenRecord := models.RefreshToken{
		AdminID:   admin.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	if err := db.Create(&refreshTokenRecord).Error; err != nil {
		return nil, err
	}

	// Update last login
	admin.LastLogin = &time.Time{}
	*admin.LastLogin = time.Now()
	db.Save(&admin)

	// Clean up old refresh tokens
	go s.cleanupExpiredTokens(admin.ID)

	return &models.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User:         admin.ToAdminInfo(),
	}, nil
}

// RefreshToken generates a new access token using a refresh token
func (s *AuthService) RefreshToken(refreshToken string) (*models.LoginResponse, error) {
	db := database.GetDB()

	// Find refresh token in database
	var tokenRecord models.RefreshToken
	err := db.Preload("Admin").Where("token = ? AND expires_at > ?", refreshToken, time.Now()).First(&tokenRecord).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid refresh token")
		}
		return nil, err
	}

	// Check if admin is still active
	if !tokenRecord.Admin.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Generate new tokens
	accessToken, newRefreshToken, expiresAt, err := s.jwtManager.GenerateTokenPair(&tokenRecord.Admin)
	if err != nil {
		return nil, err
	}

	// Update refresh token in database
	tokenRecord.Token = newRefreshToken
	tokenRecord.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	if err := db.Save(&tokenRecord).Error; err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		Token:        accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
		User:         tokenRecord.Admin.ToAdminInfo(),
	}, nil
}

// Logout invalidates a refresh token
func (s *AuthService) Logout(refreshToken string) error {
	db := database.GetDB()

	// Delete refresh token from database
	result := db.Where("token = ?", refreshToken).Delete(&models.RefreshToken{})
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// CreateAdmin creates a new admin user
func (s *AuthService) CreateAdmin(req *models.CreateAdminRequest) (*models.AdminInfo, error) {
	db := database.GetDB()

	// Validate password strength
	if err := utils.ValidatePasswordStrength(req.Password); err != nil {
		return nil, err
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create admin
	admin := models.Admin{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     req.Role,
		IsActive: true,
	}

	if err := db.Create(&admin).Error; err != nil {
		return nil, err
	}

	adminInfo := admin.ToAdminInfo()
	return &adminInfo, nil
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(userID uint, currentPassword, newPassword string) error {
	db := database.GetDB()

	// Find user
	var admin models.Admin
	if err := db.First(&admin, userID).Error; err != nil {
		return err
	}

	// Verify current password
	isValid, err := utils.VerifyPassword(currentPassword, admin.Password)
	if err != nil {
		return err
	}
	if !isValid {
		return errors.New("current password is incorrect")
	}

	// Validate new password strength
	if err := utils.ValidatePasswordStrength(newPassword); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password
	admin.Password = hashedPassword
	if err := db.Save(&admin).Error; err != nil {
		return err
	}

	// Invalidate all refresh tokens for this user
	db.Where("admin_id = ?", userID).Delete(&models.RefreshToken{})

	return nil
}

// GetAdminByID retrieves an admin by ID
func (s *AuthService) GetAdminByID(adminID uint) (*models.AdminInfo, error) {
	db := database.GetDB()

	var admin models.Admin
	if err := db.First(&admin, adminID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("admin not found")
		}
		return nil, err
	}

	adminInfo := admin.ToAdminInfo()
	return &adminInfo, nil
}

// ListAdmins retrieves all admin users
func (s *AuthService) ListAdmins() ([]models.AdminInfo, error) {
	db := database.GetDB()

	var admins []models.Admin
	if err := db.Find(&admins).Error; err != nil {
		return nil, err
	}

	var adminInfos []models.AdminInfo
	for _, admin := range admins {
		adminInfos = append(adminInfos, admin.ToAdminInfo())
	}

	return adminInfos, nil
}

// UpdateAdmin updates admin information
func (s *AuthService) UpdateAdmin(adminID uint, req *models.UpdateAdminRequest) (*models.AdminInfo, error) {
	db := database.GetDB()

	var admin models.Admin
	if err := db.First(&admin, adminID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("admin not found")
		}
		return nil, err
	}

	// Update fields if provided
	if req.Email != "" {
		admin.Email = req.Email
	}
	if req.Role != "" {
		admin.Role = req.Role
	}
	if req.IsActive != nil {
		admin.IsActive = *req.IsActive
	}

	if err := db.Save(&admin).Error; err != nil {
		return nil, err
	}

	adminInfo := admin.ToAdminInfo()
	return &adminInfo, nil
}

// DeleteAdmin deletes an admin user
func (s *AuthService) DeleteAdmin(adminID uint) error {
	db := database.GetDB()

	// Delete refresh tokens first
	db.Where("admin_id = ?", adminID).Delete(&models.RefreshToken{})

	// Delete admin
	if err := db.Delete(&models.Admin{}, adminID).Error; err != nil {
		return err
	}

	return nil
}

// ValidateToken validates a JWT token
func (s *AuthService) ValidateToken(tokenString string) (*models.JWTClaims, error) {
	return s.jwtManager.ValidateAccessToken(tokenString)
}

// cleanupExpiredTokens removes expired refresh tokens for a user
func (s *AuthService) cleanupExpiredTokens(adminID uint) {
	db := database.GetDB()
	db.Where("admin_id = ? AND expires_at < ?", adminID, time.Now()).Delete(&models.RefreshToken{})
}
