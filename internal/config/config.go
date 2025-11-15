package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port       string
	BaseURL    string
	ClientName string
}

func Load() (*Config, error) {
	godotenv.Load() // Load .env file if exists

	cfg := &Config{
		Port:       getEnv("PORT", "8181"),
		BaseURL:    getEnv("BASE_URL", "http://localhost:8181"),
		ClientName: getEnv("CLIENT_NAME", "AT Todo App"),
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
