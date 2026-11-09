package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	DatabaseURL        string
	Environment        string
	CORSAllowedOrigins []string
}

func Load() *Config {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://erp_user:erp_secret_password@localhost:5432/erp_db?sslmode=disable"
	}

	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	corsOriginsStr := os.Getenv("CORS_ALLOWED_ORIGINS")
	var corsOrigins []string
	if corsOriginsStr != "" {
		corsOrigins = strings.Split(corsOriginsStr, ",")
	} else {
		corsOrigins = []string{"http://localhost:5173", "http://localhost:3000"}
	}

	return &Config{
		Port:               port,
		DatabaseURL:        dbURL,
		Environment:        env,
		CORSAllowedOrigins: corsOrigins,
	}
}
