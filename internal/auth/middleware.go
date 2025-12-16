// internal/auth/middleware.go
package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/jedi116/kaizen-api/internal/models"
)

// JWTAuthMiddleware validates JWT tokens (PostgreSQL version)
func JWTAuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// Try to get token from Authorization header first
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// If not in header, try cookie (for web clients)
		if tokenString == "" {
			tokenString, _ = c.Cookie("access_token")
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No authorization token provided"})
			c.Abort()
			return
		}

		// Validate token
		claims, err := ValidateToken(tokenString, db)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)

		c.Next()
	}
}

// APIKeyMiddleware validates API keys (unchanged)
func APIKeyMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No API key provided"})
			c.Abort()
			return
		}

		var key models.APIKey
		if err := db.Where("key = ? AND is_active = ?", apiKey, true).First(&key).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key has expired"})
			c.Abort()
			return
		}

		// Update last used time
		go func() {
			now := time.Now()
			db.Model(&key).Update("last_used_at", now)
		}()

		c.Set("user_id", key.UserID)
		c.Set("auth_method", "api_key")

		c.Next()
	}
}

// OptionalAuth tries JWT first, then API key
func OptionalAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try JWT first
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			JWTAuthMiddleware(db)(c)
			return
		}

		// Try API key
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			APIKeyMiddleware(db)(c)
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{"error": "No authentication provided"})
		c.Abort()
	}
}

func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	return userID.(uint), true
}
