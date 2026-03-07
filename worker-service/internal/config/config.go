package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL          string
	RabbitURL            string
	RabbitExchange       string
	RabbitQueue          string
	RabbitRoutingKey     string
	MaxRetries           int
	BackoffBaseMS        int
	HTTPClientTimeoutSec int
	WorkerConcurrency    int
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:          os.Getenv("DATABASE_URL"),
		RabbitURL:            os.Getenv("RABBITMQ_URL"),
		RabbitExchange:       getEnv("RABBITMQ_EXCHANGE", "webhooks"),
		RabbitQueue:          getEnv("RABBITMQ_QUEUE", "webhook.jobs"),
		RabbitRoutingKey:     getEnv("RABBITMQ_ROUTING_KEY", "webhook.jobs"),
		MaxRetries:           getEnvInt("MAX_RETRIES", 3),
		BackoffBaseMS:        getEnvInt("BACKOFF_BASE_MS", 1000),
		HTTPClientTimeoutSec: getEnvInt("HTTP_CLIENT_TIMEOUT_SEC", 10),
		WorkerConcurrency:    getEnvInt("WORKER_CONCURRENCY", 5),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("config: DATABASE_URL is required")
	}
	if cfg.RabbitURL == "" {
		return nil, fmt.Errorf("config: RABBITMQ_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

