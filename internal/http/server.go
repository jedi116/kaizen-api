package http

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/jedi116/kaizen-api/internal/auth"
	"github.com/jedi116/kaizen-api/internal/handlers"
)

type KaizenServer struct {
	GinEngine *gin.Engine
	DB        *gorm.DB
}

func (s KaizenServer) Start() error {
	return s.GinEngine.Run()
}

func (s KaizenServer) RegisterRoutes() {
	// Health check
	s.GinEngine.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Initialize handlers
	authHandler := &handlers.AuthHandler{DB: s.DB}
	userHandler := &handlers.UserHandler{DB: s.DB}

	// API routes
	api := s.GinEngine.Group("/api")

	// Auth routes (public)
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.RefreshToken)
	}

	// Auth routes (protected - requires JWT)
	authProtected := api.Group("/auth")
	authProtected.Use(auth.JWTAuthMiddleware(s.DB))
	{
		authProtected.POST("/logout", authHandler.Logout)
		authProtected.POST("/logout-all", authHandler.LogoutAllDevices)
	}

	// User routes (protected - requires JWT)
	usersGroup := api.Group("/users")
	usersGroup.Use(auth.JWTAuthMiddleware(s.DB))
	{
		usersGroup.GET("/me", userHandler.GetProfile)
		usersGroup.PUT("/me", userHandler.UpdateProfile)
		usersGroup.POST("/api-keys", userHandler.CreateAPIKey)
		usersGroup.GET("/api-keys", userHandler.ListAPIKeys)
		usersGroup.DELETE("/api-keys/:id", userHandler.DeleteAPIKey)
	}
}
