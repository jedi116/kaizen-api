package models

import (
	"time"

	"gorm.io/gorm"
)

// FinanceJournal represents individual income/expense transactions
type FinanceJournal struct {
	gorm.Model
	UserID        uint      `gorm:"not null;index" json:"user_id"`             // Which user made this transaction
	CategoryID    uint      `gorm:"not null;index" json:"category_id"`         // Which category (Food, Rent, etc.)
	Type          string    `gorm:"not null;size:20;index" json:"type"`        // "income" or "expense"
	Amount        float64   `gorm:"type:decimal(15,2);not null" json:"amount"` // Transaction amount (always positive)
	Title         string    `gorm:"not null;size:255" json:"title"`            // e.g., "Grocery shopping at Walmart"
	Description   string    `gorm:"type:text" json:"description"`              // Optional detailed notes
	Date          time.Time `gorm:"not null;index;type:date" json:"date"`      // Transaction date (for daily tracking)
	PaymentMethod string    `gorm:"size:50" json:"payment_method"`             // "cash", "credit_card", "bank_transfer", etc.
	Location      string    `gorm:"size:255" json:"location"`                  // Where transaction occurred (optional)
	IsRecurring   bool      `gorm:"default:false" json:"is_recurring"`         // Is this a recurring transaction?

	// Attachments (optional - for receipts)
	ReceiptURL string `gorm:"size:500" json:"receipt_url"` // URL to uploaded receipt image

	// Relationships
	User     User            `gorm:"foreignKey:UserID" json:"-"`                      // Belongs to a user
	Category FinanceCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"` // Belongs to a category
}

// TableName overrides the default table name
func (FinanceJournal) TableName() string {
	return "finance_journals"
}

// BeforeCreate hook to auto-set type from category and validate
func (fj *FinanceJournal) BeforeCreate(tx *gorm.DB) error {
	// Auto-set type from category if not set
	if fj.Type == "" {
		var category FinanceCategory
		if err := tx.First(&category, fj.CategoryID).Error; err == nil {
			fj.Type = category.Type
		}
	}

	// Validate type
	if fj.Type != "income" && fj.Type != "expense" {
		fj.Type = "expense" // Default to expense
	}

	// Ensure amount is positive
	if fj.Amount < 0 {
		fj.Amount = -fj.Amount
	}

	// Set date to today if not provided
	if fj.Date.IsZero() {
		fj.Date = time.Now()
	}

	return nil
}
