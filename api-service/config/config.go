package config

import (
	"os"
	"strconv"
)

// Config holds the configuration values for the application.

type Config struct {
	ServerPort  string
	DBHost      string
	DBPort      int
	DBUser      string
	DBPassword  string
	DBName      string
	DBSSLMode   string
	WorkerCount int
	QueueSize   int
}

func Load() *Config {
	cfg := &Config{
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnvInt("DB_PORT", 5432),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", "postgres"),
		DBName:      getEnv("DB_NAME", "taskdb"),
		DBSSLMode:   getEnv("DB_SSL_MODE", "disable"),
		WorkerCount: getEnvInt("WORKER_COUNT", 5),
		QueueSize:   getEnvInt("QUEUE_SIZE", 100),
	}
	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
