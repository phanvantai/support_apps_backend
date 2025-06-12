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

	err := validateConfig(config)
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

	err := validateConfig(config)
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

	err := validateConfig(config)
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

	err := validateConfig(config)
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

	err := validateConfig(config)
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

	err := validateConfig(config)
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

	err := validateConfig(config)
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
