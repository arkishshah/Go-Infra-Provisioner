package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Add valid regions
var validRegions = map[string]bool{
	"us-east-1":      true,
	"us-east-2":      true,
	"us-west-1":      true,
	"us-west-2":      true,
	"eu-west-1":      true,
	"ap-southeast-1": true,
	// Add other regions as needed
}

type Config struct {
	AWSRegion    string
	AWSAccountID string
	Environment  string
	LogLevel     string
}

func Load() (*Config, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		// Try loading from configs/dev/app.env if .env doesn't exist
		if err := godotenv.Load("configs/dev/app.env"); err != nil {
			return nil, fmt.Errorf("error loading environment files: %w", err)
		}
	}

	region := getEnvOrDefault("AWS_REGION", "us-east-1")
	if !validRegions[region] {
		return nil, fmt.Errorf("invalid AWS region: %s", region)
	}

	config := &Config{
		AWSRegion:    region,
		AWSAccountID: os.Getenv("AWS_ACCOUNT_ID"),
		Environment:  getEnvOrDefault("ENVIRONMENT", "dev"),
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
