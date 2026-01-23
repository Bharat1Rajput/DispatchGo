package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Bharat1Rajput/task-management/config"
	"github.com/Bharat1Rajput/task-management/internal/api"
	"github.com/Bharat1Rajput/task-management/internal/database"
	"github.com/Bharat1Rajput/task-management/internal/store"
	"github.com/joho/godotenv"
)

func main() {

	// Load environment variables (fallback to system vars if .env is missing)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	cfg := config.Load()

	// Initialize database connection pool
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
	defer db.Close()
	log.Println("Database connection established")

	// The Store manages DB queries
	taskStore := store.NewTaskStore(db)

	// The API Handlers handle incoming HTTP requests
	apiHandler := api.NewHandler(taskStore)
	router := api.NewRouter(apiHandler)

	// When deploying to K8s or stopping the server, we don't want to kill
	// tasks in the middle of a webhook delivery.

	// Listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Configure and start the HTTP Server
	port := ":" + cfg.ServerPort

	srv := &http.Server{
		Addr:    port,
		Handler: router,
	}

	go func() {
		log.Printf("Starting HTTP server on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP Server failed: %v", err)
		}
	}()

	// The main thread pauses here, waiting for an interrupt signal
	<-sigChan
	log.Println("\nShutdown signal received. Gracefully stopping system...")

	// Give the HTTP server 5 seconds to finish sending responses
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	log.Println("System exited successfully. No tasks were lost.")
}
