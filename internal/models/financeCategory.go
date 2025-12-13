package models

import (
	"gorm.io/gorm"
)

// FinanceCategory represents categories for organizing transactions (e.g., Food, Rent, Salary, Freelance)
type FinanceCategory struct {
	gorm.Model
	UserID      uint   `gorm:"not null;index" json:"user_id"`         // Which user owns this category
	Name        string `gorm:"not null;size:100" json:"name"`         // e.g., "Groceries", "Salary", "Entertainment"
	Type        string `gorm:"not null;size:20;index" json:"type"`    // "income" or "expense"
	Description string `gorm:"size:500" json:"description"`           // Optional description
	Color       string `gorm:"size:7;default:'#000000'" json:"color"` // Hex color for UI (e.g., "#FF5733")
	Icon        string `gorm:"size:50" json:"icon"`                   // Icon name or emoji (e.g., "🍔", "💰")
	IsActive    bool   `gorm:"default:true" json:"is_active"`         // Soft disable without deleting

	// Relationships
	User     User             `gorm:"foreignKey:UserID" json:"-"`                      // Belongs to a user
	Journals []FinanceJournal `gorm:"foreignKey:CategoryID" json:"journals,omitempty"` // Has many journals
}

// TableName overrides the default table name
func (FinanceCategory) TableName() string {
	return "finance_categories"
}

// BeforeCreate hook to validate category type
func (fc *FinanceCategory) BeforeCreate(tx *gorm.DB) error {
	if fc.Type != "income" && fc.Type != "expense" {
		fc.Type = "expense" // Default to expense
	}
	return nil
}
