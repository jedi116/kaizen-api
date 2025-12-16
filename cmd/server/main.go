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
	if err := config.Connect(); err != nil {
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
