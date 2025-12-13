package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name     string `gorm:"not null;size:255" json:"name"`
	Email    string `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Password string `gorm:"not null" json:"-"` // Never expose password in JSON

	// Relationships
	Categories []FinanceCategory `gorm:"foreignKey:UserID" json:"categories,omitempty"`
	Journals   []FinanceJournal  `gorm:"foreignKey:UserID" json:"journals,omitempty"`
}
