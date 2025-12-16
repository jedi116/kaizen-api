// internal/handlers/user_handler.go
package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/jedi116/kaizen-api/internal/auth"
	"github.com/jedi116/kaizen-api/internal/models"
)

type UserHandler struct {
	DB *gorm.DB
}

type UpdateProfileRequest struct {
	Name string `json:"name" binding:"required" example:"John Doe"`
}

type CreateAPIKeyRequest struct {
	Name      string     `json:"name" binding:"required" example:"Mobile App"`
	ExpiresAt *time.Time `json:"expires_at" example:"2025-12-31T23:59:59Z"`
}

type APIKeyResponse struct {
	ID        uint       `json:"id"`
	Name      string     `json:"name"`
	Key       string     `json:"key"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get current user profile
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.User
// @Failure 401 {object} ErrorResponse
// @Router /users/me [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not found"})
		return
	}

	var user models.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update current user profile
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "Profile data"
// @Success 200 {object} models.User
// @Failure 400 {object} ErrorResponse
// @Router /users/me [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not found"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	var user models.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	user.Name = req.Name

	if err := h.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// CreateAPIKey godoc
// @Summary Create API key
// @Description Create a new API key for the user
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateAPIKeyRequest true "API key details"
// @Success 201 {object} APIKeyResponse
// @Failure 400 {object} ErrorResponse
// @Router /users/api-keys [post]
func (h *UserHandler) CreateAPIKey(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not found"})
		return
	}

	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	apiKey := models.APIKey{
		UserID:    userID,
		Name:      req.Name,
		ExpiresAt: req.ExpiresAt,
		IsActive:  true,
	}

	if err := apiKey.GenerateKey(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate API key"})
		return
	}

	if err := h.DB.Create(&apiKey).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create API key"})
		return
	}

	c.JSON(http.StatusCreated, APIKeyResponse{
		ID:        apiKey.ID,
		Name:      apiKey.Name,
		Key:       apiKey.Key,
		CreatedAt: apiKey.CreatedAt,
		ExpiresAt: apiKey.ExpiresAt,
	})
}

// ListAPIKeys godoc
// @Summary List API keys
// @Description Get all API keys for the user
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.APIKey
// @Router /users/api-keys [get]
func (h *UserHandler) ListAPIKeys(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not found"})
		return
	}

	var apiKeys []models.APIKey
	if err := h.DB.Where("user_id = ?", userID).Find(&apiKeys).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch API keys"})
		return
	}

	c.JSON(http.StatusOK, apiKeys)
}

// DeleteAPIKey godoc
// @Summary Delete API key
// @Description Delete an API key
// @Tags Users
// @Security BearerAuth
// @Param id path int true "API Key ID"
// @Produce json
// @Success 200 {object} MessageResponse
// @Router /users/api-keys/{id} [delete]
func (h *UserHandler) DeleteAPIKey(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not found"})
		return
	}

	id := c.Param("id")

	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.APIKey{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete API key"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "API key deleted successfully"})
}
