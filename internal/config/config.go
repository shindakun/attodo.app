package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	BaseURL     string
	ClientName  string
	Environment string
}

func Load() (*Config, error) {
	godotenv.Load() // Load .env file if exists

	cfg := &Config{
		Port:        getEnv("PORT", "8181"),
		BaseURL:     getEnv("BASE_URL", "http://localhost:8181"),
		ClientName:  getEnv("CLIENT_NAME", "AT Todo App"),
		Environment: getEnv("ENVIRONMENT", "production"),
	}

	return cfg, nil
}

// IsDev returns true if running in development mode
func (c *Config) IsDev() bool {
	return c.Environment == "dev" || c.Environment == "development"
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
