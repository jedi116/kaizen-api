// @title Kaizen API
// @version 1.0
// @description Personal finance and productivity API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@kaizen.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8000
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your bearer token in the format: Bearer <token>

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description Enter your API key

package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jedi116/kaizen-api/config"
	"github.com/jedi116/kaizen-api/internal/http"
	"github.com/joho/godotenv"
)

func main() {
	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, using environment variables")
		}
	}

	// Connect to database
	if err := config.ConnectToDataBase(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Get database instance
	db := config.GetDB()

	// Create server with database connection
	server := http.KaizenServer{
		GinEngine: gin.Default(),
		DB:        db,
	}

	server.RegisterRoutes()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Server starting on port %s\n", port)

	if err := server.Start(); err != nil {
		log.Fatal("Could not start the server: ", err)
	}
}
