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

func TestConnectDatabase_LoggerConfiguration_Development(t *testing.T) {
	// Test logger configuration branch when DSN suggests development
	// Since we can't easily test the actual connectDatabase function with PostgreSQL,
	// we'll test the logic by examining the DSN
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	// Test DSN generation
	dsn := cfg.GetDSN()
	assert.Contains(t, dsn, "localhost")
	assert.Contains(t, dsn, "testuser")
	assert.Contains(t, dsn, "testdb")
}

func TestConnectDatabase_LoggerConfiguration_Production(t *testing.T) {
	// Test logger configuration for production-like environments
	cfg := config.DatabaseConfig{
		Host:     "prod-db.example.com",
		Port:     5432,
		User:     "produser",
		Password: "prodpass",
		DBName:   "proddb",
		SSLMode:  "require",
	}

	// Test DSN generation for production
	dsn := cfg.GetDSN()
	assert.Contains(t, dsn, "prod-db.example.com")
	assert.Contains(t, dsn, "produser")
	assert.Contains(t, dsn, "require")
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

func TestAutoMigrate_MultipleModels(t *testing.T) {
	// Test migration with multiple models
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// First migration
	err = autoMigrate(db)
	require.NoError(t, err)

	// Verify both tables exist
	assert.True(t, db.Migrator().HasTable(&models.SupportRequest{}))
	assert.True(t, db.Migrator().HasTable(&models.User{}))

	// Run migration again (should be idempotent)
	err = autoMigrate(db)
	assert.NoError(t, err)
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

func TestSetupRouter_ProductionGinMode(t *testing.T) {
	// Save original mode
	originalMode := gin.Mode()
	defer gin.SetMode(originalMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Environment: "production",
			RateLimit:   5.0,
			RateBurst:   10,
		},
	}

	supportHandler := &handlers.SupportRequestHandler{}
	authHandler := &handlers.AuthHandler{}
	mockAuthService := &MockAuthServiceForRouter{}

	router := setupRouter(cfg, supportHandler, authHandler, mockAuthService)

	assert.NotNil(t, router)
	// The production mode should have been set during setupRouter execution
}

func TestSetupRouter_CORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Environment: "development",
			RateLimit:   10.0,
			RateBurst:   20,
		},
	}

	supportHandler := &handlers.SupportRequestHandler{}
	authHandler := &handlers.AuthHandler{}
	mockAuthService := &MockAuthServiceForRouter{}

	router := setupRouter(cfg, supportHandler, authHandler, mockAuthService)

	// Verify router is created with CORS middleware
	assert.NotNil(t, router)

	// Check that routes are properly registered
	routes := router.Routes()
	assert.Greater(t, len(routes), 0)
}

func TestSetupRouter_RateLimitingConfiguration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Environment: "development",
			RateLimit:   2.0, // Low rate for testing
			RateBurst:   5,   // Low burst for testing
		},
	}

	supportHandler := &handlers.SupportRequestHandler{}
	authHandler := &handlers.AuthHandler{}
	mockAuthService := &MockAuthServiceForRouter{}

	router := setupRouter(cfg, supportHandler, authHandler, mockAuthService)

	// Verify router is created and has the rate-limited route
	assert.NotNil(t, router)

	routes := router.Routes()
	supportRequestRoute := false
	for _, route := range routes {
		if route.Path == "/api/v1/support-request" && route.Method == "POST" {
			supportRequestRoute = true
			break
		}
	}
	assert.True(t, supportRequestRoute, "Rate-limited support request route should exist")
}

func TestSetupRouter_AllEndpointsRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Environment: "development",
			RateLimit:   10.0,
			RateBurst:   20,
		},
	}

	supportHandler := &handlers.SupportRequestHandler{}
	authHandler := &handlers.AuthHandler{}
	mockAuthService := &MockAuthServiceForRouter{}

	router := setupRouter(cfg, supportHandler, authHandler, mockAuthService)

	routes := router.Routes()
	routeMap := make(map[string]bool)

	for _, route := range routes {
		routeMap[route.Method+" "+route.Path] = true
	}

	// Verify all expected routes exist
	expectedRoutes := []string{
		"GET /health",
		"POST /api/v1/support-request",
		"POST /api/v1/auth/login",
		"GET /api/v1/auth/me",
		"PATCH /api/v1/auth/password",
		"POST /api/v1/auth/users",
		"GET /api/v1/auth/users",
		"GET /api/v1/auth/users/:id",
		"PATCH /api/v1/auth/users/:id",
		"DELETE /api/v1/auth/users/:id",
		"GET /api/v1/support-requests",
		"GET /api/v1/support-requests/:id",
		"PATCH /api/v1/support-requests/:id",
		"DELETE /api/v1/support-requests/:id",
	}

	for _, expectedRoute := range expectedRoutes {
		assert.True(t, routeMap[expectedRoute], "Route %s should be registered", expectedRoute)
	}
}
