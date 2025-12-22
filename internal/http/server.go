package http

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	_ "github.com/jedi116/kaizen-api/docs"
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

	// Swagger documentation
	s.GinEngine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Initialize handlers
	authHandler := &handlers.AuthHandler{DB: s.DB}
	userHandler := &handlers.UserHandler{DB: s.DB}
	categoryHandler := &handlers.FinanceCategoryHandler{DB: s.DB}
	journalHandler := &handlers.FinanceJournalHandler{DB: s.DB}

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

	// Finance Category routes (protected - requires JWT)
	categoriesGroup := api.Group("/categories")
	categoriesGroup.Use(auth.JWTAuthMiddleware(s.DB))
	{
		categoriesGroup.POST("", categoryHandler.CreateCategory)
		categoriesGroup.GET("", categoryHandler.ListCategories)
		categoriesGroup.GET("/:id", categoryHandler.GetCategory)
		categoriesGroup.PUT("/:id", categoryHandler.UpdateCategory)
		categoriesGroup.DELETE("/:id", categoryHandler.DeleteCategory)
	}

	// Finance Journal routes (protected - requires JWT)
	journalsGroup := api.Group("/journals")
	journalsGroup.Use(auth.JWTAuthMiddleware(s.DB))
	{
		journalsGroup.POST("", journalHandler.CreateJournal)
		journalsGroup.GET("", journalHandler.ListJournals)
		journalsGroup.GET("/summary", journalHandler.GetSummary)
		journalsGroup.GET("/:id", journalHandler.GetJournal)
		journalsGroup.PUT("/:id", journalHandler.UpdateJournal)
		journalsGroup.DELETE("/:id", journalHandler.DeleteJournal)
	}
}
