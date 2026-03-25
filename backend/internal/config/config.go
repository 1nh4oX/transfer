package config

import (
	"fmt"
	"os"
)

// Config keeps runtime settings for local/dev startup.
type Config struct {
	Port          string
	BaseURL       string
	DatabaseURL   string
	UploadDir     string
	JWTSecret     string
	DemoUsername  string
	DemoPassword  string
	TokenExpireIn int64
}

func Load() Config {
	cfg := Config{
		Port:          getEnv("PORT", "8080"),
		BaseURL:       getEnv("BASE_URL", "http://localhost:8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/transfer?sslmode=disable"),
		UploadDir:     getEnv("UPLOAD_DIR", "./uploads"),
		JWTSecret:     getEnv("JWT_SECRET", "dev-only-change-me"),
		DemoUsername:  getEnv("DEMO_USERNAME", "admin"),
		DemoPassword:  getEnv("DEMO_PASSWORD", "change_me"),
		TokenExpireIn: 7200,
	}

	return cfg
}

func (c Config) Addr() string {
	return fmt.Sprintf(":%s", c.Port)
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
