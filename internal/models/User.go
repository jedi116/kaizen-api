// internal/models/user.go
package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name          string     `gorm:"not null;size:255" json:"name"`
	Email         string     `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Password      string     `gorm:"not null" json:"-"`
	EmailVerified bool       `gorm:"default:false" json:"email_verified"`
	LastLoginAt   *time.Time `json:"last_login_at"`

	// Relationships
	Categories []FinanceCategory `gorm:"foreignKey:UserID" json:"categories,omitempty"`
	Journals   []FinanceJournal  `gorm:"foreignKey:UserID" json:"journals,omitempty"`
	APIKeys    []APIKey          `gorm:"foreignKey:UserID" json:"api_keys,omitempty"`
	Tokens     []Token           `gorm:"foreignKey:UserID" json:"-"`
}

// TableName overrides the default table name
func (User) TableName() string {
	return "users"
}
