package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/Bharat1Rajput/workerService/internal/config"
	"github.com/Bharat1Rajput/workerService/internal/database"
	"github.com/Bharat1Rajput/workerService/internal/store"
	"github.com/Bharat1Rajput/workerService/internal/worker"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	// Initialize database
	db, err := database.NewPostgres(database.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Connected to database successfully")
	defer db.Close()

	store := store.NewStore(db)
	w := worker.New(store)

	ctx, cancel := context.WithCancel(context.Background())
	go w.Start(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("\nShutdown signal received. Gracefully stopping system...")

	cancel()
	log.Println("Worker shutting down,no job were left unprocessed")
}
