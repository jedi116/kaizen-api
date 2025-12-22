// internal/handlers/finance_journal_handler.go
package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/jedi116/kaizen-api/internal/auth"
	"github.com/jedi116/kaizen-api/internal/models"
)

type FinanceJournalHandler struct {
	DB *gorm.DB
}

type CreateJournalRequest struct {
	CategoryID    uint    `json:"category_id" binding:"required" example:"1"`
	Amount        float64 `json:"amount" binding:"required,gt=0" example:"25.50"`
	Title         string  `json:"title" binding:"required" example:"Grocery shopping"`
	Description   string  `json:"description" example:"Weekly groceries at Walmart"`
	Date          string  `json:"date" example:"2025-01-15"`
	PaymentMethod string  `json:"payment_method" example:"credit_card"`
	Location      string  `json:"location" example:"Walmart"`
	IsRecurring   bool    `json:"is_recurring" example:"false"`
	ReceiptURL    string  `json:"receipt_url" example:"https://example.com/receipt.jpg"`
}

type UpdateJournalRequest struct {
	CategoryID    *uint    `json:"category_id" example:"1"`
	Amount        *float64 `json:"amount" example:"25.50"`
	Title         string   `json:"title" example:"Grocery shopping"`
	Description   string   `json:"description" example:"Weekly groceries at Walmart"`
	Date          string   `json:"date" example:"2025-01-15"`
	PaymentMethod string   `json:"payment_method" example:"credit_card"`
	Location      string   `json:"location" example:"Walmart"`
	IsRecurring   *bool    `json:"is_recurring" example:"false"`
	ReceiptURL    string   `json:"receipt_url" example:"https://example.com/receipt.jpg"`
}

type JournalSummary struct {
	TotalIncome  float64 `json:"total_income" example:"5000.00"`
	TotalExpense float64 `json:"total_expense" example:"3500.00"`
	NetBalance   float64 `json:"net_balance" example:"1500.00"`
	StartDate    string  `json:"start_date" example:"2025-01-01"`
	EndDate      string  `json:"end_date" example:"2025-01-31"`
	EntryCount   int64   `json:"entry_count" example:"42"`
}

type JournalListResponse struct {
	Journals   []models.FinanceJournal `json:"journals"`
	TotalCount int64                   `json:"total_count"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"page_size"`
}

// CreateJournal godoc
// @Summary Create a journal entry
// @Description Create a new income or expense entry
// @Tags Finance Journals
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateJournalRequest true "Journal entry details"
// @Success 201 {object} models.FinanceJournal
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /journals [post]
func (h *FinanceJournalHandler) CreateJournal(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not found"})
		return
	}

	var req CreateJournalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Verify category belongs to user
	var category models.FinanceCategory
	if err := h.DB.Where("id = ? AND user_id = ?", req.CategoryID, userID).First(&category).Error; err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Category not found or doesn't belong to you"})
		return
	}

	// Parse date
	var entryDate time.Time
	if req.Date != "" {
		parsed, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid date format. Use YYYY-MM-DD"})
			return
		}
		entryDate = parsed
	} else {
		entryDate = time.Now()
	}

	journal := models.FinanceJournal{
		UserID:        userID,
		CategoryID:    req.CategoryID,
		Type:          category.Type, // Inherit type from category
		Amount:        req.Amount,
		Title:         req.Title,
		Description:   req.Description,
		Date:          entryDate,
		PaymentMethod: req.PaymentMethod,
		Location:      req.Location,
		IsRecurring:   req.IsRecurring,
		ReceiptURL:    req.ReceiptURL,
	}

	if err := h.DB.Create(&journal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create journal entry"})
		return
	}

	// Load category for response
	h.DB.Preload("Category").First(&journal, journal.ID)

	c.JSON(http.StatusCreated, journal)
}

// ListJournals godoc
// @Summary List journal entries
// @Description Get all journal entries with optional filters
// @Tags Finance Journals
// @Security BearerAuth
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param category_id query int false "Filter by category ID"
// @Param type query string false "Filter by type (income or expense)"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Items per page (default: 20, max: 100)"
// @Success 200 {object} JournalListResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /journals [get]
func (h *FinanceJournalHandler) ListJournals(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not found"})
		return
	}

	query := h.DB.Where("user_id = ?", userID)

	// Filter by date range
	if startDate := c.Query("start_date"); startDate != "" {
		if parsed, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("date >= ?", parsed)
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if parsed, err := time.Parse("2006-01-02", endDate); err == nil {
			query = query.Where("date <= ?", parsed)
		}
	}

	// Filter by category
	if categoryID := c.Query("category_id"); categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	// Filter by type
	if typeFilter := c.Query("type"); typeFilter != "" {
		if typeFilter == "income" || typeFilter == "expense" {
			query = query.Where("type = ?", typeFilter)
		}
	}

	// Get total count
	var totalCount int64
	query.Model(&models.FinanceJournal{}).Count(&totalCount)

	// Pagination
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if parsed, err := parseInt(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := parseInt(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}
	offset := (page - 1) * pageSize

	var journals []models.FinanceJournal
	if err := query.Preload("Category").Order("date DESC, created_at DESC").Offset(offset).Limit(pageSize).Find(&journals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch journal entries"})
		return
	}

	c.JSON(http.StatusOK, JournalListResponse{
		Journals:   journals,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	})
}

// GetJournal godoc
// @Summary Get a journal entry
// @Description Get a single journal entry by ID
// @Tags Finance Journals
// @Security BearerAuth
// @Produce json
// @Param id path int true "Journal ID"
// @Success 200 {object} models.FinanceJournal
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /journals/{id} [get]
func (h *FinanceJournalHandler) GetJournal(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not found"})
		return
	}

	id := c.Param("id")

	var journal models.FinanceJournal
	if err := h.DB.Preload("Category").Where("id = ? AND user_id = ?", id, userID).First(&journal).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Journal entry not found"})
		return
	}

	c.JSON(http.StatusOK, journal)
}

// UpdateJournal godoc
// @Summary Update a journal entry
// @Description Update an existing journal entry
// @Tags Finance Journals
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Journal ID"
// @Param request body UpdateJournalRequest true "Journal data"
// @Success 200 {object} models.FinanceJournal
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /journals/{id} [put]
func (h *FinanceJournalHandler) UpdateJournal(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not found"})
		return
	}

	id := c.Param("id")

	var journal models.FinanceJournal
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&journal).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Journal entry not found"})
		return
	}

	var req UpdateJournalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Update category if provided
	if req.CategoryID != nil {
		var category models.FinanceCategory
		if err := h.DB.Where("id = ? AND user_id = ?", *req.CategoryID, userID).First(&category).Error; err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Category not found or doesn't belong to you"})
			return
		}
		journal.CategoryID = *req.CategoryID
		journal.Type = category.Type
	}

	// Update other fields
	if req.Amount != nil && *req.Amount > 0 {
		journal.Amount = *req.Amount
	}
	if req.Title != "" {
		journal.Title = req.Title
	}
	if req.Description != "" {
		journal.Description = req.Description
	}
	if req.Date != "" {
		if parsed, err := time.Parse("2006-01-02", req.Date); err == nil {
			journal.Date = parsed
		}
	}
	if req.PaymentMethod != "" {
		journal.PaymentMethod = req.PaymentMethod
	}
	if req.Location != "" {
		journal.Location = req.Location
	}
	if req.IsRecurring != nil {
		journal.IsRecurring = *req.IsRecurring
	}
	if req.ReceiptURL != "" {
		journal.ReceiptURL = req.ReceiptURL
	}

	if err := h.DB.Save(&journal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update journal entry"})
		return
	}

	// Load category for response
	h.DB.Preload("Category").First(&journal, journal.ID)

	c.JSON(http.StatusOK, journal)
}

// DeleteJournal godoc
// @Summary Delete a journal entry
// @Description Delete a journal entry (soft delete)
// @Tags Finance Journals
// @Security BearerAuth
// @Produce json
// @Param id path int true "Journal ID"
// @Success 200 {object} MessageResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /journals/{id} [delete]
func (h *FinanceJournalHandler) DeleteJournal(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not found"})
		return
	}

	id := c.Param("id")

	var journal models.FinanceJournal
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&journal).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Journal entry not found"})
		return
	}

	if err := h.DB.Delete(&journal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete journal entry"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Journal entry deleted successfully"})
}

// GetSummary godoc
// @Summary Get financial summary
// @Description Get income, expense, and balance summary for a date range
// @Tags Finance Journals
// @Security BearerAuth
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD, default: first day of current month)"
// @Param end_date query string false "End date (YYYY-MM-DD, default: today)"
// @Success 200 {object} JournalSummary
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /journals/summary [get]
func (h *FinanceJournalHandler) GetSummary(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not found"})
		return
	}

	// Default date range: current month
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := now

	// Parse custom date range
	if sd := c.Query("start_date"); sd != "" {
		if parsed, err := time.Parse("2006-01-02", sd); err == nil {
			startDate = parsed
		}
	}
	if ed := c.Query("end_date"); ed != "" {
		if parsed, err := time.Parse("2006-01-02", ed); err == nil {
			endDate = parsed
		}
	}

	// Calculate totals
	var totalIncome float64
	var totalExpense float64
	var entryCount int64

	h.DB.Model(&models.FinanceJournal{}).
		Where("user_id = ? AND date >= ? AND date <= ? AND type = ?", userID, startDate, endDate, "income").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalIncome)

	h.DB.Model(&models.FinanceJournal{}).
		Where("user_id = ? AND date >= ? AND date <= ? AND type = ?", userID, startDate, endDate, "expense").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalExpense)

	h.DB.Model(&models.FinanceJournal{}).
		Where("user_id = ? AND date >= ? AND date <= ?", userID, startDate, endDate).
		Count(&entryCount)

	summary := JournalSummary{
		TotalIncome:  totalIncome,
		TotalExpense: totalExpense,
		NetBalance:   totalIncome - totalExpense,
		StartDate:    startDate.Format("2006-01-02"),
		EndDate:      endDate.Format("2006-01-02"),
		EntryCount:   entryCount,
	}

	c.JSON(http.StatusOK, summary)
}

// Helper function to parse int from string
func parseInt(s string) (int, error) {
	var result int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		} else {
			return 0, nil
		}
	}
	return result, nil
}
