package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/Bharat1Rajput/workerService/internal/config"
	"github.com/Bharat1Rajput/workerService/internal/consumer"
	"github.com/Bharat1Rajput/workerService/internal/processor"
	"github.com/Bharat1Rajput/workerService/internal/repository"
	"database/sql"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	_ = godotenv.Load()

	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}
	defer logger.Sync()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}
	logger.Info("database url", zap.String("url", cfg.DatabaseURL))
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	if err := db.Ping(); err != nil {
		logger.Fatal("database ping failed", zap.Error(err))
	}
	defer db.Close()

	repo := repository.NewPostgresJobRepository(db)
	proc := processor.New(cfg, repo, logger)

	cons, err := consumer.New(cfg, proc, logger)
	if err != nil {
		logger.Fatal("failed to create consumer", zap.Error(err))
	}
	defer cons.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		if err := cons.Start(ctx); err != nil {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case <-stop:
		logger.Info("shutdown signal received")
		cancel()
	case err := <-errCh:
		return err
	}

	return nil
}

