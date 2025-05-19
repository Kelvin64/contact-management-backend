package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	CORS     CORSConfig
}

// ServerConfig holds all server-related configuration
type ServerConfig struct {
	Port    string
	GinMode string
}

// DatabaseConfig holds all database-related configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	URL      string
}

// CORSConfig holds all CORS-related configuration
type CORSConfig struct {
	AllowedOrigins []string
}

// Load reads the environment variables and returns a Config struct
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found. Using environment variables.\n")
	}

	config := &Config{
		Server: ServerConfig{
			Port:    getEnvOrDefault("PORT", "8080"),
			GinMode: getEnvOrDefault("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvOrDefault("DB_PORT", "5432"),
			User:     getEnvOrDefault("DB_USER", "postgres"),
			Password: getEnvOrDefault("DB_PASSWORD", ""),
			DBName:   getEnvOrDefault("DB_NAME", "contactmanagement_db"),
			URL:      getEnvOrDefault("DATABASE_URL", ""),
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{getEnvOrDefault("ALLOWED_ORIGINS", "http://localhost:3000")},
		},
	}

	return config, nil
}

// GetDatabaseURL returns the database URL, either from DATABASE_URL env var or constructed from individual components
func (c *DatabaseConfig) GetDatabaseURL() string {
	if c.URL != "" {
		return c.URL
	}

	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		c.Host,
		c.User,
		c.Password,
		c.DBName,
		c.Port,
	)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
} 