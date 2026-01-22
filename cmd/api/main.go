package main

import (
	"log"

	"github.com/Bharat1Rajput/task-management/config"
	"github.com/Bharat1Rajput/task-management/internal/database"
	"github.com/joho/godotenv"
)

func main() {
	// Load configuration
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	cfg := config.Load()

	// Initialize database
	db, err := database.NewPostgresDB(database.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	})
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}

	log.Println("Database connection established")

	defer db.Close()

}
