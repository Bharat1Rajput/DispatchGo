package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppPort           string
	HMACSecret        string
	RabbitURL         string
	RabbitExchange    string
	RabbitQueue       string
	RabbitRoutingKey  string
	ShutdownTimeoutSec int
}

func Load() (*Config, error) {
	cfg := &Config{
		AppPort:           getEnv("APP_PORT", "8080"),
		HMACSecret:        os.Getenv("HMAC_SECRET"),
		RabbitURL:         os.Getenv("RABBITMQ_URL"),
		RabbitExchange:    getEnv("RABBITMQ_EXCHANGE", "webhooks"),
		RabbitQueue:       getEnv("RABBITMQ_QUEUE", "webhook.jobs"),
		RabbitRoutingKey:  getEnv("RABBITMQ_ROUTING_KEY", "webhook.jobs"),
		ShutdownTimeoutSec: getEnvInt("SHUTDOWN_TIMEOUT_SEC", 15),
	}

	if cfg.HMACSecret == "" {
		return nil, fmt.Errorf("config: HMAC_SECRET is required")
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

