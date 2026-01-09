package main

import (
	"log"
	"os"

	"github.com/brian-l-johnson/Redteam-Dashboard-go/v2/models"
	"github.com/brian-l-johnson/Redteam-Dashboard-go/v2/server"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Failed to load .env file, using environment variables")
	}

	// Validate required environment variables
	if os.Getenv("API_BASE_URL") == "" {
		log.Fatal("API_BASE_URL environment variable is required")
	}

	// Validate session secret exists
	if os.Getenv("SESSION_SECRET") == "" {
		log.Println("WARNING: SESSION_SECRET not set. Generating random secret (sessions will not persist across restarts)")
	}

	// Initialize database
	models.Init()

	// Start server
	server.Init()
}
