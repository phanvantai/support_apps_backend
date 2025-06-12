package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Success(t *testing.T) {
	// Set valid environment variables
	os.Setenv("JWT_SECRET", "this-is-a-very-secure-jwt-secret-key-that-is-at-least-32-characters-long")
	os.Setenv("DB_PASSWORD", "very-secure-database-password")
	os.Setenv("DB_USER", "customuser")
	os.Setenv("DB_SSLMODE", "require")
	os.Setenv("ENVIRONMENT", "production")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_SSLMODE")
		os.Unsetenv("ENVIRONMENT")
	}()

	config, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "this-is-a-very-secure-jwt-secret-key-that-is-at-least-32-characters-long", config.JWT.SecretKey)
	assert.Equal(t, "production", config.Server.Environment)
}

func TestLoad_DefaultValues(t *testing.T) {
	// Clear all environment variables
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("PORT")
	os.Unsetenv("ENVIRONMENT")

	// Set valid JWT secret for development
	os.Setenv("JWT_SECRET", "development-secret-key-that-is-long-enough-to-pass-validation")
	defer os.Unsetenv("JWT_SECRET")

	config, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "localhost", config.Database.Host)
	assert.Equal(t, 5432, config.Database.Port)
	assert.Equal(t, "postgres", config.Database.User)
	assert.Equal(t, "password", config.Database.Password)
	assert.Equal(t, "support_app", config.Database.DBName)
	assert.Equal(t, "8080", config.Server.Port)
	assert.Equal(t, "development", config.Server.Environment)
	assert.Equal(t, 10.0, config.Server.RateLimit)
	assert.Equal(t, 20, config.Server.RateBurst)
}

func TestValidateConfig_ProductionInsecureJWT(t *testing.T) {
	config := &Config{
		JWT: JWTConfig{
			SecretKey: "your-secret-key-change-in-production",
		},
		Server: ServerConfig{
			Environment: "production",
		},
	}

	err := validateConfig(config, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT secret is insecure")
}

func TestValidateConfig_ShortJWT(t *testing.T) {
	config := &Config{
		JWT: JWTConfig{
			SecretKey: "short",
		},
		Server: ServerConfig{
			Environment: "development",
		},
	}

	err := validateConfig(config, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT secret is insecure")
}

func TestValidateConfig_ProductionInsecureDatabase(t *testing.T) {
	config := &Config{
		JWT: JWTConfig{
			SecretKey: "this-is-a-very-secure-jwt-secret-key-that-is-at-least-32-characters-long",
		},
		Database: DatabaseConfig{
			Password: "password",
			User:     "postgres",
			SSLMode:  "disable",
		},
		Server: ServerConfig{
			Environment: "production",
		},
	}

	err := validateConfig(config, false) // Using individual env vars, not DATABASE_URL
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database password is insecure")
}

func TestValidateConfig_ProductionSSLDisabled(t *testing.T) {
	config := &Config{
		JWT: JWTConfig{
			SecretKey: "this-is-a-very-secure-jwt-secret-key-that-is-at-least-32-characters-long",
		},
		Database: DatabaseConfig{
			Password: "very-secure-database-password",
			User:     "customuser",
			SSLMode:  "disable",
		},
		Server: ServerConfig{
			Environment: "production",
		},
	}

	err := validateConfig(config, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SSL must be enabled")
}

func TestValidateConfig_ProductionDefaultUser(t *testing.T) {
	config := &Config{
		JWT: JWTConfig{
			SecretKey: "this-is-a-very-secure-jwt-secret-key-that-is-at-least-32-characters-long",
		},
		Database: DatabaseConfig{
			Password: "very-secure-database-password",
			User:     "postgres",
			SSLMode:  "require",
		},
		Server: ServerConfig{
			Environment: "production",
		},
	}

	err := validateConfig(config, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default database user 'postgres'")
}

func TestValidateConfig_InvalidEnvironment(t *testing.T) {
	config := &Config{
		JWT: JWTConfig{
			SecretKey: "this-is-a-very-secure-jwt-secret-key-that-is-at-least-32-characters-long",
		},
		Server: ServerConfig{
			Environment: "invalid",
		},
	}

	err := validateConfig(config, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid environment")
}

func TestValidateConfig_Success(t *testing.T) {
	config := &Config{
		JWT: JWTConfig{
			SecretKey: "this-is-a-very-secure-jwt-secret-key-that-is-at-least-32-characters-long",
		},
		Database: DatabaseConfig{
			Password: "very-secure-database-password",
			User:     "customuser",
			SSLMode:  "require",
		},
		Server: ServerConfig{
			Environment: "production",
		},
	}

	err := validateConfig(config, false)
	assert.NoError(t, err)
}

func TestDatabaseConfig_GetDSN(t *testing.T) {
	config := &DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	dsn := config.GetDSN()
	expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	assert.Equal(t, expected, dsn)
}

func TestDatabaseConfig_GetDatabaseURL(t *testing.T) {
	config := &DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		SSLMode:  "require",
	}

	url := config.GetDatabaseURL()
	expected := "postgresql://testuser:testpass@localhost:5432/testdb?sslmode=require"
	assert.Equal(t, expected, url)
}

func TestParseDatabaseURL_Success(t *testing.T) {
	databaseURL := "postgresql://testuser:testpass@localhost:5432/testdb?sslmode=require"

	config, err := parseDatabaseURL(databaseURL)
	assert.NoError(t, err)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 5432, config.Port)
	assert.Equal(t, "testuser", config.User)
	assert.Equal(t, "testpass", config.Password)
	assert.Equal(t, "testdb", config.DBName)
	assert.Equal(t, "require", config.SSLMode)
}

func TestParseDatabaseURL_DefaultPort(t *testing.T) {
	databaseURL := "postgresql://testuser:testpass@localhost/testdb"

	config, err := parseDatabaseURL(databaseURL)
	assert.NoError(t, err)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 5432, config.Port)         // Default port
	assert.Equal(t, "require", config.SSLMode) // Default SSL mode
}

func TestParseDatabaseURL_InvalidURL(t *testing.T) {
	tests := []struct {
		name        string
		databaseURL string
		expectError string
	}{
		{
			name:        "completely invalid URL",
			databaseURL: "invalid-url",
			expectError: "invalid DATABASE_URL scheme",
		},
		{
			name:        "wrong scheme",
			databaseURL: "mysql://user:pass@host:5432/db",
			expectError: "invalid DATABASE_URL scheme",
		},
		{
			name:        "missing host",
			databaseURL: "postgresql://user:pass@:5432/db",
			expectError: "missing host",
		},
		{
			name:        "missing database name",
			databaseURL: "postgresql://user:pass@localhost:5432/",
			expectError: "missing database name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseDatabaseURL(tt.databaseURL)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectError)
		})
	}
}

func TestParseDatabaseURL_InvalidPort(t *testing.T) {
	databaseURL := "postgresql://testuser:testpass@localhost:invalid/testdb"

	_, err := parseDatabaseURL(databaseURL)
	assert.Error(t, err)
	// The error comes from url.Parse first, so it's actually a parse error
	assert.Contains(t, err.Error(), "invalid DATABASE_URL format")
}

func TestLoad_WithDatabaseURL(t *testing.T) {
	// Clear individual DB environment variables
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_SSLMODE")

	// Set DATABASE_URL and JWT_SECRET
	os.Setenv("DATABASE_URL", "postgresql://railwayuser:railwaypass@railway.host:5432/railwaydb?sslmode=require")
	os.Setenv("JWT_SECRET", "railway-jwt-secret-key-that-is-long-enough-for-validation")
	defer os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("JWT_SECRET")

	config, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "railway.host", config.Database.Host)
	assert.Equal(t, 5432, config.Database.Port)
	assert.Equal(t, "railwayuser", config.Database.User)
	assert.Equal(t, "railwaypass", config.Database.Password)
	assert.Equal(t, "railwaydb", config.Database.DBName)
	assert.Equal(t, "require", config.Database.SSLMode)
}

func TestValidateConfig_WithDatabaseURL_Production(t *testing.T) {
	// Test that DATABASE_URL bypasses individual credential validation
	config := &Config{
		JWT: JWTConfig{
			SecretKey: "this-is-a-very-secure-jwt-secret-key-that-is-at-least-32-characters-long",
		},
		Database: DatabaseConfig{
			Password: "short",    // This would normally fail validation
			User:     "postgres", // This would normally fail validation
			SSLMode:  "require",
		},
		Server: ServerConfig{
			Environment: "production",
		},
	}

	// With usingDatabaseURL=true, validation should pass despite short password and default user
	err := validateConfig(config, true)
	assert.NoError(t, err)

	// With usingDatabaseURL=false, validation should fail
	err = validateConfig(config, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database password is insecure")
}

func TestGetEnv_WithValue(t *testing.T) {
	os.Setenv("TEST_ENV", "test_value")
	defer os.Unsetenv("TEST_ENV")

	result := getEnv("TEST_ENV", "fallback")
	assert.Equal(t, "test_value", result)
}

func TestGetEnv_WithFallback(t *testing.T) {
	os.Unsetenv("MISSING_ENV")

	result := getEnv("MISSING_ENV", "fallback")
	assert.Equal(t, "fallback", result)
}

func TestGetEnvAsInt_WithValue(t *testing.T) {
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	result := getEnvAsInt("TEST_INT", 10)
	assert.Equal(t, 42, result)
}

func TestGetEnvAsInt_WithInvalidValue(t *testing.T) {
	os.Setenv("TEST_INT", "invalid")
	defer os.Unsetenv("TEST_INT")

	result := getEnvAsInt("TEST_INT", 10)
	assert.Equal(t, 10, result)
}

func TestGetEnvAsInt_WithFallback(t *testing.T) {
	os.Unsetenv("MISSING_INT")

	result := getEnvAsInt("MISSING_INT", 10)
	assert.Equal(t, 10, result)
}

func TestGetEnvAsFloat_WithValue(t *testing.T) {
	os.Setenv("TEST_FLOAT", "3.14")
	defer os.Unsetenv("TEST_FLOAT")

	result := getEnvAsFloat("TEST_FLOAT", 1.0)
	assert.Equal(t, 3.14, result)
}

func TestGetEnvAsFloat_WithInvalidValue(t *testing.T) {
	os.Setenv("TEST_FLOAT", "invalid")
	defer os.Unsetenv("TEST_FLOAT")

	result := getEnvAsFloat("TEST_FLOAT", 1.0)
	assert.Equal(t, 1.0, result)
}

func TestGetEnvAsFloat_WithFallback(t *testing.T) {
	os.Unsetenv("MISSING_FLOAT")

	result := getEnvAsFloat("MISSING_FLOAT", 1.0)
	assert.Equal(t, 1.0, result)
}
