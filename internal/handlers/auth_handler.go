// internal/handlers/auth_handler.go
package handlers

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/jedi116/kaizen-api/internal/auth"
	"github.com/jedi116/kaizen-api/internal/models"
)

type AuthHandler struct {
	DB *gorm.DB
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required" example:"John Doe"`
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"password123"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

type AuthResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         UserProfile `json:"user"`
}

type UserProfile struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := h.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Email already registered"})
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to hash password"})
		return
	}

	// Create new user
	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create user"})
		return
	}

	// Generate tokens
	tokens, err := auth.GenerateTokenPair(user.ID, user.Email, h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate tokens"})
		return
	}

	// Set cookies
	h.setAuthCookies(c, tokens)

	c.JSON(http.StatusCreated, AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User: UserProfile{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	})
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return tokens
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Find user
	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid email or password"})
		return
	}

	// Check password
	if !auth.CheckPassword(user.Password, req.Password) {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid email or password"})
		return
	}

	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	h.DB.Save(&user)

	// Generate tokens
	tokens, err := auth.GenerateTokenPair(user.ID, user.Email, h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate tokens"})
		return
	}

	// Set cookies
	h.setAuthCookies(c, tokens)

	c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User: UserProfile{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	})
}

// Logout godoc
// @Summary Logout user
// @Description Revoke user tokens
// @Tags Authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} MessageResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Revoke access token from header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		auth.RevokeToken(tokenString, h.DB)
	}

	// Revoke refresh token from cookie
	refreshToken, _ := c.Cookie("refresh_token")
	if refreshToken != "" {
		auth.RevokeToken(refreshToken, h.DB)
	}

	// Clear cookies
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, MessageResponse{Message: "Logged out successfully"})
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param refresh_token body object{refresh_token=string} false "Refresh token (optional if using cookies)"
// @Success 200 {object} auth.TokenPair
// @Failure 401 {object} ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Try to get refresh token from cookie
	refreshToken, _ := c.Cookie("refresh_token")

	// If not in cookie, try request body
	if refreshToken == "" {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := c.ShouldBindJSON(&req); err == nil {
			refreshToken = req.RefreshToken
		}
	}

	if refreshToken == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "No refresh token provided"})
		return
	}

	// Validate refresh token
	claims, err := auth.ValidateToken(refreshToken, h.DB)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid refresh token"})
		return
	}

	// Revoke old tokens
	auth.RevokeToken(refreshToken, h.DB)

	// Generate new token pair
	tokens, err := auth.GenerateTokenPair(claims.UserID, claims.Email, h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate tokens"})
		return
	}

	// Set new cookies
	h.setAuthCookies(c, tokens)

	c.JSON(http.StatusOK, tokens)
}

// LogoutAllDevices godoc
// @Summary Logout from all devices
// @Description Revoke all user tokens
// @Tags Authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} MessageResponse
// @Router /auth/logout-all [post]
func (h *AuthHandler) LogoutAllDevices(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not found"})
		return
	}

	if err := auth.RevokeAllUserTokens(userID, h.DB); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to logout from all devices"})
		return
	}

	// Clear cookies
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, MessageResponse{Message: "Logged out from all devices"})
}

// Helper to set auth cookies
func (h *AuthHandler) setAuthCookies(c *gin.Context, tokens *auth.TokenPair) {
	secure := os.Getenv("ENV") == "production"

	c.SetCookie(
		"access_token",
		tokens.AccessToken,
		int(auth.AccessTokenDuration.Seconds()),
		"/",
		"",
		secure, // Secure only in production (HTTPS)
		true,   // HttpOnly
	)

	c.SetCookie(
		"refresh_token",
		tokens.RefreshToken,
		int(auth.RefreshTokenDuration.Seconds()),
		"/",
		"",
		secure, // Secure only in production (HTTPS)
		true,   // HttpOnly
	)
}
