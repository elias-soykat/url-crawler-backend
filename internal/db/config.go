package db

import (
	"os"
	"time"
)

// Config holds database configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	MaxOpen  int
	MaxIdle  int
	Timeout  time.Duration
}

// NewConfig creates a new database configuration from environment variables
func NewConfig() *Config {
	return &Config{
		Host:     getEnvOrDefault("MYSQL_HOST", "localhost"),
		Port:     getEnvOrDefault("MYSQL_PORT", "3306"),
		User:     getEnvOrDefault("MYSQL_USER", "root"),
		Password: getEnvOrDefault("MYSQL_PASSWORD", ""),
		Database: getEnvOrDefault("MYSQL_DATABASE", "url_crawler"),
		MaxOpen:  25,
		MaxIdle:  5,
		Timeout:  30 * time.Second,
	}
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
} 