package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jedi116/kaizen-api/config"
	"github.com/jedi116/kaizen-api/internal/http"
	"github.com/joho/godotenv"

	"log"
	"os"
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

	// listen and serve on
	server := http.KaizenServer{
		GinEngine: gin.Default(),
	}

	server.RegisterRoutes()
	err := server.Start()

	port := os.Getenv("PORT")
	if err != nil {
		log.Fatal("Could not start the server: ", err)
	} else {
		log.Printf("🚀 Server starting on port %s\n", port)
	}
}
