package models

import (
	"time"

	"gorm.io/gorm"
)

// Admin represents an admin user
type Admin struct {
	ID         uint           `json:"id" gorm:"primarykey"`
	Username   string         `json:"username" gorm:"type:varchar(50);uniqueIndex;not null" validate:"required,min=3,max=50"`
	Email      string         `json:"email" gorm:"type:varchar(100);uniqueIndex;not null" validate:"required,email"`
	Password   string         `json:"-" gorm:"type:varchar(255);not null"` // Never expose password in JSON
	Role       string         `json:"role" gorm:"type:varchar(20);not null;default:'admin'" validate:"required,oneof=admin superadmin"`
	IsActive   bool           `json:"is_active" gorm:"default:true"`
	FirstLogin bool           `json:"first_login" gorm:"default:true"`
	LastLogin  *time.Time     `json:"last_login"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" validate:"required" binding:"required"`
	Password string `json:"password" validate:"required,min=6" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         AdminInfo `json:"user"`
}

// AdminInfo represents public admin information
type AdminInfo struct {
	ID         uint   `json:"id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	FirstLogin bool   `json:"first_login"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" binding:"required"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"omitempty" binding:"omitempty"`
	NewPassword     string `json:"new_password" validate:"required,min=6" binding:"required"`
}

// CreateAdminRequest represents a request to create a new admin
type CreateAdminRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50" binding:"required"`
	Email    string `json:"email" validate:"required,email" binding:"required"`
	Password string `json:"password" validate:"required,min=6" binding:"required"`
	Role     string `json:"role" validate:"required,oneof=admin superadmin" binding:"required"`
}

// UpdateAdminRequest represents a request to update admin info
type UpdateAdminRequest struct {
	Email    string `json:"email" validate:"omitempty,email"`
	Role     string `json:"role" validate:"omitempty,oneof=admin superadmin"`
	IsActive *bool  `json:"is_active"`
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Type     string `json:"type"` // "access" or "refresh"
}

// RefreshToken represents a stored refresh token
type RefreshToken struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	AdminID   uint           `json:"admin_id" gorm:"not null;index"`
	Token     string         `json:"token" gorm:"type:varchar(255);uniqueIndex;not null"`
	ExpiresAt time.Time      `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	Admin     Admin          `json:"-" gorm:"foreignKey:AdminID"`
}

// TableName sets the table name for Admin model
func (Admin) TableName() string {
	return "admins"
}

// TableName sets the table name for RefreshToken model
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// ToAdminInfo converts Admin to AdminInfo
func (a *Admin) ToAdminInfo() AdminInfo {
	return AdminInfo{
		ID:         a.ID,
		Username:   a.Username,
		Email:      a.Email,
		Role:       a.Role,
		FirstLogin: a.FirstLogin,
	}
}
