package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/Bharat1Rajput/apiService/internal/broker"
	"github.com/Bharat1Rajput/apiService/internal/config"
	"github.com/Bharat1Rajput/apiService/internal/handler"
	"github.com/Bharat1Rajput/apiService/internal/middleware"
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

	pub, err := broker.NewRabbitPublisher(
		cfg.RabbitURL,
		cfg.RabbitExchange,
		cfg.RabbitQueue,
		cfg.RabbitRoutingKey,
		logger, 
	)
	if err != nil {
		logger.Fatal("failed to create rabbit publisher", zap.Error(err))
	}
	defer pub.Close()

	r := chi.NewRouter()
	r.Use(middleware.RequestLogger(logger))
	r.Group(func(r chi.Router) {
    r.Get("/health", handler.HealthHandler)
})
	r.Group(func(r chi.Router) {
		r.Use(middleware.HMACAuth(cfg.HMACSecret, logger))
		r.Mount("/", handler.NewWebhookHandler(cfg, pub, logger).Routes())
	})

	addr := ":" + cfg.AppPort
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	} 

	errCh := make(chan error, 1)

	go func() {
		logger.Info("api-service starting", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logger.Info("shutdown signal received", zap.String("signal", sig.String()))
	case err := <-errCh:
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ShutdownTimeoutSec)*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", zap.Error(err))
		return err
	}

	logger.Info("api-service stopped gracefully")
	return nil
}

