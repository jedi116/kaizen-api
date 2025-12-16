package models

import (
	"time"
)

type Token struct {
	ID        string    `gorm:"primaryKey;size:100"`
	UserID    uint      `gorm:"not null;index"`
	Type      string    `gorm:"not null;size:20"` // "access" or "refresh"
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time

	User User `gorm:"foreignKey:UserID"`
}

// TableName overrides the default table name
func (Token) TableName() string {
	return "tokens"
}

// IsExpired checks if token is expired
func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}
