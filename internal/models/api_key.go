// internal/models/api_key.go
package models

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"gorm.io/gorm"
)

type APIKey struct {
	gorm.Model
	UserID     uint       `gorm:"not null;index" json:"user_id"`
	Name       string     `gorm:"not null;size:100" json:"name"`
	Key        string     `gorm:"uniqueIndex;not null;size:64" json:"key"`
	LastUsedAt *time.Time `json:"last_used_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	IsActive   bool       `gorm:"default:true" json:"is_active"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// GenerateKey creates a new random API key
func (a *APIKey) GenerateKey() error {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return err
	}
	a.Key = hex.EncodeToString(bytes)
	return nil
}

// TableName overrides the default table name
func (APIKey) TableName() string {
	return "api_keys"
}
