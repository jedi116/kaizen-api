// internal/handlers/finance_category_handler.go
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/jedi116/kaizen-api/internal/auth"
	"github.com/jedi116/kaizen-api/internal/models"
)

type FinanceCategoryHandler struct {
	DB *gorm.DB
}

type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required" example:"Groceries"`
	Type        string `json:"type" binding:"required,oneof=income expense" example:"expense"`
	Description string `json:"description" example:"Food and household items"`
	Color       string `json:"color" example:"#FF5733"`
	Icon        string `json:"icon" example:"🛒"`
}

type UpdateCategoryRequest struct {
	Name        string `json:"name" example:"Groceries"`
	Type        string `json:"type" binding:"omitempty,oneof=income expense" example:"expense"`
	Description string `json:"description" example:"Food and household items"`
	Color       string `json:"color" example:"#FF5733"`
	Icon        string `json:"icon" example:"🛒"`
	IsActive    *bool  `json:"is_active" example:"true"`
}

// CreateCategory godoc
// @Summary Create a finance category
// @Description Create a new income or expense category
// @Tags Finance Categories
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateCategoryRequest true "Category details"
// @Success 201 {object} models.FinanceCategory
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories [post]
func (h *FinanceCategoryHandler) CreateCategory(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not found"})
		return
	}

	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	category := models.FinanceCategory{
		UserID:      userID,
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
		IsActive:    true,
	}

	if err := h.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create category"})
		return
	}

	c.JSON(http.StatusCreated, category)
}

// ListCategories godoc
// @Summary List finance categories
// @Description Get all categories for the current user with optional type filter
// @Tags Finance Categories
// @Security BearerAuth
// @Produce json
// @Param type query string false "Filter by type (income or expense)"
// @Param active query bool false "Filter by active status"
// @Success 200 {array} models.FinanceCategory
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories [get]
func (h *FinanceCategoryHandler) ListCategories(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not found"})
		return
	}

	query := h.DB.Where("user_id = ?", userID)

	// Filter by type
	if typeFilter := c.Query("type"); typeFilter != "" {
		if typeFilter == "income" || typeFilter == "expense" {
			query = query.Where("type = ?", typeFilter)
		}
	}

	// Filter by active status
	if activeFilter := c.Query("active"); activeFilter != "" {
		if activeFilter == "true" {
			query = query.Where("is_active = ?", true)
		} else if activeFilter == "false" {
			query = query.Where("is_active = ?", false)
		}
	}

	var categories []models.FinanceCategory
	if err := query.Order("name ASC").Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch categories"})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// GetCategory godoc
// @Summary Get a finance category
// @Description Get a single category by ID
// @Tags Finance Categories
// @Security BearerAuth
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} models.FinanceCategory
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /categories/{id} [get]
func (h *FinanceCategoryHandler) GetCategory(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not found"})
		return
	}

	id := c.Param("id")

	var category models.FinanceCategory
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Category not found"})
		return
	}

	c.JSON(http.StatusOK, category)
}

// UpdateCategory godoc
// @Summary Update a finance category
// @Description Update an existing category
// @Tags Finance Categories
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param request body UpdateCategoryRequest true "Category data"
// @Success 200 {object} models.FinanceCategory
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories/{id} [put]
func (h *FinanceCategoryHandler) UpdateCategory(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not found"})
		return
	}

	id := c.Param("id")

	var category models.FinanceCategory
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Category not found"})
		return
	}

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Update fields if provided
	if req.Name != "" {
		category.Name = req.Name
	}
	if req.Type != "" {
		category.Type = req.Type
	}
	if req.Description != "" {
		category.Description = req.Description
	}
	if req.Color != "" {
		category.Color = req.Color
	}
	if req.Icon != "" {
		category.Icon = req.Icon
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if err := h.DB.Save(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update category"})
		return
	}

	c.JSON(http.StatusOK, category)
}

// DeleteCategory godoc
// @Summary Delete a finance category
// @Description Delete a category (soft delete)
// @Tags Finance Categories
// @Security BearerAuth
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} MessageResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories/{id} [delete]
func (h *FinanceCategoryHandler) DeleteCategory(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not found"})
		return
	}

	id := c.Param("id")

	// Check if category exists and belongs to user
	var category models.FinanceCategory
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Category not found"})
		return
	}

	// Check if category has journals
	var journalCount int64
	h.DB.Model(&models.FinanceJournal{}).Where("category_id = ?", id).Count(&journalCount)
	if journalCount > 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Cannot delete category with existing journal entries. Delete the entries first or deactivate the category."})
		return
	}

	if err := h.DB.Delete(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete category"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Category deleted successfully"})
}

