package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	JWT      JWTConfig
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string
	Environment  string
	RateLimit    float64
	RateBurst    int
	PublicDomain string // For Railway deployment or custom domain
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file (optional)
	_ = godotenv.Load()

	var databaseConfig DatabaseConfig
	var usingDatabaseURL bool

	// Check if DATABASE_URL is provided (Railway style)
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		var err error
		databaseConfig, err = parseDatabaseURL(databaseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse DATABASE_URL: %w", err)
		}
		usingDatabaseURL = true
	} else {
		// Use individual environment variables (development style)
		databaseConfig = DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "support_app"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		}
		usingDatabaseURL = false
	}

	config := &Config{
		Database: databaseConfig,
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			Environment:  getEnv("ENVIRONMENT", "development"),
			RateLimit:    getEnvAsFloat("RATE_LIMIT", 10.0), // 10 requests per second
			RateBurst:    getEnvAsInt("RATE_BURST", 20),     // burst of 20 requests
			PublicDomain: getPublicDomain(),
		},
		JWT: JWTConfig{
			SecretKey: getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		},
	}

	// Validate configuration for security
	if err := validateConfig(config, usingDatabaseURL); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// validateConfig validates the configuration for security issues
func validateConfig(config *Config, usingDatabaseURL bool) error {
	// Validate JWT secret
	if config.JWT.SecretKey == "your-secret-key-change-in-production" ||
		config.JWT.SecretKey == "your-jwt-secret-key-change-this" ||
		len(config.JWT.SecretKey) < 32 {
		return fmt.Errorf("JWT secret is insecure: must be at least 32 characters and not use default values")
	}

	// Validate database configuration in production
	if config.Server.Environment == "production" {
		// If using DATABASE_URL (Railway/cloud deployment), trust the provider's security
		if !usingDatabaseURL {
			// Only validate individual env vars for manual configuration
			if config.Database.Password == "password" || len(config.Database.Password) < 12 {
				return fmt.Errorf("database password is insecure for production: must be at least 12 characters and not use default values")
			}

			// Check for default database user
			if config.Database.User == "postgres" {
				return fmt.Errorf("default database user 'postgres' should not be used in production")
			}
		}

		// Always ensure SSL is enabled in production (regardless of DATABASE_URL)
		if config.Database.SSLMode == "disable" {
			return fmt.Errorf("SSL must be enabled for production database connections")
		}
	}

	// Validate environment
	validEnvironments := []string{"development", "staging", "production"}
	isValidEnv := false
	for _, env := range validEnvironments {
		if config.Server.Environment == env {
			isValidEnv = true
			break
		}
	}
	if !isValidEnv {
		return fmt.Errorf("invalid environment '%s': must be one of %s",
			config.Server.Environment, strings.Join(validEnvironments, ", "))
	}

	return nil
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// GetDatabaseURL returns the PostgreSQL connection string in URL format (Railway style)
func (c *DatabaseConfig) GetDatabaseURL() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvAsInt gets an environment variable as integer with a fallback value
func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

// getEnvAsFloat gets an environment variable as float with a fallback value
func getEnvAsFloat(key string, fallback float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return fallback
}

// getPublicDomain determines the public domain for the application
// Returns Railway domain if deployed there, otherwise localhost for development
func getPublicDomain() string {
	// Check for Railway-specific environment variables
	if railwayPublicDomain := os.Getenv("RAILWAY_PUBLIC_DOMAIN"); railwayPublicDomain != "" {
		return railwayPublicDomain
	}

	// Check for generic public domain override
	if publicDomain := os.Getenv("PUBLIC_DOMAIN"); publicDomain != "" {
		return publicDomain
	}

	// Check if we're running on Railway (they set specific env vars)
	if os.Getenv("RAILWAY_ENVIRONMENT") != "" || os.Getenv("RAILWAY_PROJECT_ID") != "" {
		// On Railway but no public domain set yet - this happens during initial deployment
		// Return empty string so we can handle it gracefully
		return ""
	}

	// Default to localhost for local development
	port := getEnv("PORT", "8080")
	return "localhost:" + port
}

// parseDatabaseURL parses a DATABASE_URL environment variable (Railway style)
// Expected format: postgresql://username:password@host:port/database?sslmode=require
func parseDatabaseURL(databaseURL string) (DatabaseConfig, error) {
	parsedURL, err := url.Parse(databaseURL)
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("invalid DATABASE_URL format: %w", err)
	}

	// Validate scheme
	if parsedURL.Scheme != "postgresql" && parsedURL.Scheme != "postgres" {
		return DatabaseConfig{}, fmt.Errorf("invalid DATABASE_URL scheme: expected postgresql or postgres, got %s", parsedURL.Scheme)
	}

	// Extract components
	host := parsedURL.Hostname()
	if host == "" {
		return DatabaseConfig{}, fmt.Errorf("invalid DATABASE_URL: missing host")
	}

	port := parsedURL.Port()
	if port == "" {
		port = "5432" // Default PostgreSQL port
	}

	var user, password string
	if parsedURL.User != nil {
		user = parsedURL.User.Username()
		password, _ = parsedURL.User.Password()
	}

	dbName := strings.TrimPrefix(parsedURL.Path, "/")
	if dbName == "" {
		return DatabaseConfig{}, fmt.Errorf("invalid DATABASE_URL: missing database name")
	}

	// Extract SSL mode from query parameters
	sslMode := "require" // Default for production
	if parsedURL.Query().Get("sslmode") != "" {
		sslMode = parsedURL.Query().Get("sslmode")
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("invalid port in DATABASE_URL: %w", err)
	}

	return DatabaseConfig{
		Host:     host,
		Port:     portInt,
		User:     user,
		Password: password,
		DBName:   dbName,
		SSLMode:  sslMode,
	}, nil
}
