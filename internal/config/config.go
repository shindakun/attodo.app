package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	BaseURL           string
	ClientName        string
	Environment       string
	DBPath            string
	MigrationsDir     string
	VAPIDPublicKey    string
	VAPIDPrivateKey   string
	VAPIDSubscriber   string
}

func Load() (*Config, error) {
	godotenv.Load() // Load .env file if exists

	cfg := &Config{
		Port:            getEnv("PORT", "8181"),
		BaseURL:         getEnv("BASE_URL", "http://localhost:8181"),
		ClientName:      getEnv("CLIENT_NAME", "AT Todo App"),
		Environment:     getEnv("ENVIRONMENT", "production"),
		DBPath:          getEnv("DB_PATH", "./data/app.db"),
		MigrationsDir:   getEnv("MIGRATIONS_DIR", "./migrations"),
		VAPIDPublicKey:  getEnv("VAPID_PUBLIC_KEY", ""),
		VAPIDPrivateKey: getEnv("VAPID_PRIVATE_KEY", ""),
		VAPIDSubscriber: getEnv("VAPID_SUBSCRIBER", "mailto:admin@attodo.app"),
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
