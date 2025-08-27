package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/Run-Panel/VerTree/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

var (
	// These should be loaded from environment variables in production
	jwtSecret       = []byte("your-super-secret-jwt-key-change-this-in-production")
	accessTokenTTL  = 24 * time.Hour
	refreshTokenTTL = 7 * 24 * time.Hour
)

// JWTManager handles JWT token operations
type JWTManager struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secret string) *JWTManager {
	if secret == "" {
		secret = "your-super-secret-jwt-key-change-this-in-production"
	}

	return &JWTManager{
		secret:          []byte(secret),
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

// GenerateTokenPair generates both access and refresh tokens
func (j *JWTManager) GenerateTokenPair(admin *models.Admin) (string, string, time.Time, error) {
	// Generate access token
	accessToken, expiresAt, err := j.GenerateAccessToken(admin)
	if err != nil {
		return "", "", time.Time{}, err
	}

	// Generate refresh token
	refreshToken, err := j.GenerateRefreshToken(admin)
	if err != nil {
		return "", "", time.Time{}, err
	}

	return accessToken, refreshToken, expiresAt, nil
}

// GenerateAccessToken generates a JWT access token
func (j *JWTManager) GenerateAccessToken(admin *models.Admin) (string, time.Time, error) {
	expiresAt := time.Now().Add(j.accessTokenTTL)

	claims := jwt.MapClaims{
		"user_id":  admin.ID,
		"username": admin.Username,
		"role":     admin.Role,
		"type":     "access",
		"exp":      expiresAt.Unix(),
		"iat":      time.Now().Unix(),
		"nbf":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// GenerateRefreshToken generates a secure refresh token
func (j *JWTManager) GenerateRefreshToken(admin *models.Admin) (string, error) {
	// Generate a random token
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

// ValidateAccessToken validates and parses a JWT access token
func (j *JWTManager) ValidateAccessToken(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Check token type
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "access" {
		return nil, errors.New("invalid token type")
	}

	// Extract claims
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, errors.New("invalid user_id in token")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return nil, errors.New("invalid username in token")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return nil, errors.New("invalid role in token")
	}

	return &models.JWTClaims{
		UserID:   uint(userID),
		Username: username,
		Role:     role,
		Type:     tokenType,
	}, nil
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func ExtractTokenFromHeader(authHeader string) string {
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}

// SetJWTSecret sets the JWT secret (should be called during app initialization)
func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}

// SetTokenTTL sets the token TTL values
func SetTokenTTL(accessTTL, refreshTTL time.Duration) {
	accessTokenTTL = accessTTL
	refreshTokenTTL = refreshTTL
}
