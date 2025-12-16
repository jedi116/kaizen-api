package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jedi116/kaizen-api/internal/models"
	"gorm.io/gorm"
)

var (
	AccessTokenDuration  = 15 * time.Minute
	RefreshTokenDuration = 7 * 24 * time.Hour
	jwtSecret            = []byte(os.Getenv("JWT_SECRET"))
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// GenerateTokenPair creates both access and refresh tokens
func GenerateTokenPair(userID uint, email string, db *gorm.DB) (*TokenPair, error) {
	// Access Token
	accessJTI := fmt.Sprintf("access_%d_%d", userID, time.Now().UnixNano())
	accessClaims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        accessJTI,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(jwtSecret)
	if err != nil {
		return nil, err
	}

	// Refresh Token
	refreshJTI := fmt.Sprintf("refresh_%d_%d", userID, time.Now().UnixNano())
	refreshClaims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        refreshJTI,
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(jwtSecret)
	if err != nil {
		return nil, err
	}

	// Store tokens in PostgreSQL
	tokens := []models.Token{
		{
			ID:        accessJTI,
			UserID:    userID,
			Type:      "access",
			ExpiresAt: time.Now().Add(AccessTokenDuration),
		},
		{
			ID:        refreshJTI,
			UserID:    userID,
			Type:      "refresh",
			ExpiresAt: time.Now().Add(RefreshTokenDuration),
		},
	}

	if err := db.Create(&tokens).Error; err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}

// ValidateToken validates and parses a JWT token
func ValidateToken(tokenString string, db *gorm.DB) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Check if token exists and is not expired in database
	var dbToken models.Token
	if err := db.Where("id = ?", claims.ID).First(&dbToken).Error; err != nil {
		return nil, fmt.Errorf("token not found or revoked")
	}

	if dbToken.IsExpired() {
		return nil, fmt.Errorf("token has expired")
	}

	return claims, nil
}

// RevokeToken revokes a token by deleting it from database
func RevokeToken(tokenString string, db *gorm.DB) error {
	token, _ := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if claims, ok := token.Claims.(*Claims); ok {
		return db.Where("id = ?", claims.ID).Delete(&models.Token{}).Error
	}

	return nil
}

// RevokeAllUserTokens revokes all tokens for a user (logout from all devices)
func RevokeAllUserTokens(userID uint, db *gorm.DB) error {
	return db.Where("user_id = ?", userID).Delete(&models.Token{}).Error
}

// CleanupExpiredTokens removes expired tokens (run this periodically)
func CleanupExpiredTokens(db *gorm.DB) error {
	return db.Where("expires_at < ?", time.Now()).Delete(&models.Token{}).Error
}
