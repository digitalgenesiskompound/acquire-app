package config

import (
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	Host        string
	Port        string
	HTTPPort    string
	Environment string
	Debug       bool
}

// Load creates a new configuration from environment variables
func Load() *Config {
	cfg := &Config{
		Host:        getEnv("HOST", "0.0.0.0"),
		Port:        getEnv("PORT", "8080"),
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		Environment: getEnv("ENV", "development"),
		Debug:       getEnvBool("DEBUG", true),
	}
	
	return cfg
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvBool gets a boolean environment variable with a fallback value
func getEnvBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return fallback
}
