package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AWSRegion    string
	AWSAccountID string
	Environment  string
	LogLevel     string
}

func Load() (*Config, error) {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	// Load environment-specific .env file
	envFile := fmt.Sprintf("configs/%s/app.env", env)
	if err := godotenv.Load(envFile); err != nil {
		return nil, fmt.Errorf("error loading %s: %w", envFile, err)
	}

	config := &Config{
		AWSRegion:    getEnvOrDefault("AWS_REGION", "us-east-1"),
		AWSAccountID: os.Getenv("AWS_ACCOUNT_ID"),
		Environment:  env,
		LogLevel:     getEnvOrDefault("LOG_LEVEL", "info"),
	}

	if config.AWSAccountID == "" {
		return nil, fmt.Errorf("AWS_ACCOUNT_ID is required")
	}

	return config, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
