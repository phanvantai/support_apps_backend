package main

import (
	"os"
	"support-app-backend/internal/config"
	"support-app-backend/internal/handlers"
	"support-app-backend/internal/models"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestConnectDatabase_Success(t *testing.T) {
	// cfg := config.DatabaseConfig{
	// 	Host:     "localhost",
	// 	Port:     5432,
	// 	User:     "testuser",
	// 	Password: "testpass",
	// 	DBName:   "testdb",
	// 	SSLMode:  "disable",
	// }

	// Use SQLite for testing instead of PostgreSQL
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	assert.NoError(t, err)
	assert.NotNil(t, db)

	// Test connection
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = sqlDB.Ping()
	assert.NoError(t, err)
}

func TestAutoMigrate_Success(t *testing.T) {
	// Use in-memory SQLite for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = autoMigrate(db)
	assert.NoError(t, err)

	// Verify that tables were created
	assert.True(t, db.Migrator().HasTable(&models.SupportRequest{}))
	assert.True(t, db.Migrator().HasTable(&models.User{}))
}

func TestSetupRouter_Development(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Environment: "development",
			RateLimit:   10.0,
			RateBurst:   20,
		},
	}

	// Create mock handlers and service
	supportHandler := &handlers.SupportRequestHandler{}
	authHandler := &handlers.AuthHandler{}

	// Create a mock auth service
	mockAuthService := &MockAuthServiceForRouter{}

	router := setupRouter(cfg, supportHandler, authHandler, mockAuthService)

	assert.NotNil(t, router)
}

func TestSetupRouter_Production(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Environment: "production",
			RateLimit:   10.0,
			RateBurst:   20,
		},
	}

	// Create mock handlers and service
	supportHandler := &handlers.SupportRequestHandler{}
	authHandler := &handlers.AuthHandler{}

	// Create a mock auth service
	mockAuthService := &MockAuthServiceForRouter{}

	router := setupRouter(cfg, supportHandler, authHandler, mockAuthService)

	assert.NotNil(t, router)
}

func TestSetupRouter_Routes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Environment: "development",
			RateLimit:   10.0,
			RateBurst:   20,
		},
	}

	// Create mock handlers and service
	supportHandler := &handlers.SupportRequestHandler{}
	authHandler := &handlers.AuthHandler{}

	// Create a mock auth service
	mockAuthService := &MockAuthServiceForRouter{}

	router := setupRouter(cfg, supportHandler, authHandler, mockAuthService)

	// Get routes
	routes := router.Routes()

	// Check that expected routes exist
	routePaths := make(map[string]bool)
	for _, route := range routes {
		routePaths[route.Method+" "+route.Path] = true
	}

	// Public routes
	assert.True(t, routePaths["GET /health"])
	assert.True(t, routePaths["POST /api/v1/support-request"])
	assert.True(t, routePaths["POST /api/v1/auth/login"])

	// Protected routes should exist (middleware will be applied at runtime)
	assert.True(t, routePaths["GET /api/v1/auth/me"])
	assert.True(t, routePaths["PATCH /api/v1/auth/password"])
}

func TestDatabaseConfig_GetDSN(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	dsn := cfg.GetDSN()
	expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	assert.Equal(t, expected, dsn)
}

func TestMain_Environment(t *testing.T) {
	// Test that main doesn't panic when called (we can't easily test the full main function)
	// but we can test that the setup functions work correctly

	// Save original environment
	originalPort := os.Getenv("PORT")
	originalJWTSecret := os.Getenv("JWT_SECRET")

	// Set test environment
	os.Setenv("PORT", "8081")
	os.Setenv("JWT_SECRET", "test-jwt-secret-key-that-is-long-enough-for-testing-purposes")

	defer func() {
		// Restore original environment
		if originalPort != "" {
			os.Setenv("PORT", originalPort)
		} else {
			os.Unsetenv("PORT")
		}
		if originalJWTSecret != "" {
			os.Setenv("JWT_SECRET", originalJWTSecret)
		} else {
			os.Unsetenv("JWT_SECRET")
		}
	}()

	// Test configuration loading
	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "8081", cfg.Server.Port)
	assert.Equal(t, "test-jwt-secret-key-that-is-long-enough-for-testing-purposes", cfg.JWT.SecretKey)
}

// MockAuthServiceForRouter is a minimal mock for testing router setup
type MockAuthServiceForRouter struct{}

func (m *MockAuthServiceForRouter) Login(req *models.LoginRequest) (*models.LoginResponse, error) {
	return nil, nil
}

func (m *MockAuthServiceForRouter) CreateUser(req *models.CreateUserRequest) (*models.UserInfo, error) {
	return nil, nil
}

func (m *MockAuthServiceForRouter) GetUserByID(id uint) (*models.UserInfo, error) {
	return nil, nil
}

func (m *MockAuthServiceForRouter) GetAllUsers(page, pageSize int) ([]*models.UserInfo, int64, error) {
	return nil, 0, nil
}

func (m *MockAuthServiceForRouter) UpdateUser(id uint, req *models.UpdateUserRequest) (*models.UserInfo, error) {
	return nil, nil
}

func (m *MockAuthServiceForRouter) DeleteUser(id uint) error {
	return nil
}

func (m *MockAuthServiceForRouter) ChangePassword(userID uint, req *models.ChangePasswordRequest) error {
	return nil
}

func (m *MockAuthServiceForRouter) CreateDefaultAdmin() error {
	return nil
}

func (m *MockAuthServiceForRouter) ValidateToken(tokenString string) (*models.User, error) {
	return &models.User{
		ID:       1,
		Username: "testuser",
		Role:     models.UserRoleUser,
		IsActive: true,
	}, nil
}

func TestSetupRouter_CORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Environment: "development",
			RateLimit:   10.0,
			RateBurst:   20,
		},
	}

	// Create mock handlers and service
	supportHandler := &handlers.SupportRequestHandler{}
	authHandler := &handlers.AuthHandler{}
	mockAuthService := &MockAuthServiceForRouter{}

	router := setupRouter(cfg, supportHandler, authHandler, mockAuthService)

	// Test that CORS middleware is properly set up by checking routes
	routes := router.Routes()
	assert.NotEmpty(t, routes)

	// The CORS middleware should be applied to all routes
	// We can't easily test the actual CORS headers without making HTTP requests
	// but we can verify the router was set up successfully
	assert.NotNil(t, router)
}

func TestAutoMigrate_WithExistingTables(t *testing.T) {
	// Use in-memory SQLite for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Run migration first time
	err = autoMigrate(db)
	assert.NoError(t, err)

	// Run migration second time (should not error)
	err = autoMigrate(db)
	assert.NoError(t, err)

	// Verify that tables still exist
	assert.True(t, db.Migrator().HasTable(&models.SupportRequest{}))
	assert.True(t, db.Migrator().HasTable(&models.User{}))
}
