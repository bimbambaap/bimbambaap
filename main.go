package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/bimbambaap/bimbambaap/database"
	"github.com/bimbambaap/bimbambaap/router"
)

func main() {
	// Laad .env lokaal (op Railway staan deze als environment variables)
	godotenv.Load()

	// Verbind met database
	database.Connect()
	database.Migrate()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := router.Setup()
	log.Printf("Server draait op :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Server kon niet starten:", err)
	}
}
