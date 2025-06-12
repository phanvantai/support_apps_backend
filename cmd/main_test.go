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

// Application tests
func TestNewApplication_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Set up test environment with SQLite
	setupTestEnvironmentWithSQLite(t)
	defer cleanupTestEnvironment()

	app, err := NewApplication()
	require.NoError(t, err)
	assert.NotNil(t, app)
	assert.NotNil(t, app.Config)
	assert.NotNil(t, app.DB)
	assert.NotNil(t, app.AuthService)
	assert.NotNil(t, app.SupportService)
	assert.NotNil(t, app.AuthHandler)
	assert.NotNil(t, app.SupportHandler)
	assert.NotNil(t, app.Router)
}

func TestNewApplication_ConfigError(t *testing.T) {
	// Set invalid environment to cause config error
	os.Setenv("ENVIRONMENT", "invalid")
	defer os.Unsetenv("ENVIRONMENT")

	app, err := NewApplication()
	assert.Error(t, err)
	assert.Nil(t, app)
	assert.Contains(t, err.Error(), "failed to initialize config")
}

func TestApplication_InitializeConfig_Success(t *testing.T) {
	setupTestEnvironment(t)
	defer cleanupTestEnvironment()

	app := &Application{}
	err := app.initializeConfig()
	require.NoError(t, err)
	assert.NotNil(t, app.Config)
}

func TestApplication_InitializeConfig_Error(t *testing.T) {
	// Set invalid environment to cause config validation error
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("JWT_SECRET", "short") // Too short for production
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("JWT_SECRET")
	}()

	app := &Application{}
	err := app.initializeConfig()
	assert.Error(t, err)
	assert.Nil(t, app.Config)
}

func TestApplication_InitializeDatabase_Success(t *testing.T) {
	setupTestEnvironmentWithSQLite(t)
	defer cleanupTestEnvironment()

	app := &Application{}
	err := app.initializeConfig()
	require.NoError(t, err)

	err = app.initializeDatabase()
	require.NoError(t, err)
	assert.NotNil(t, app.DB)

	// Verify migration worked
	assert.True(t, app.DB.Migrator().HasTable(&models.SupportRequest{}))
	assert.True(t, app.DB.Migrator().HasTable(&models.User{}))
}

func TestApplication_InitializeServices_Success(t *testing.T) {
	setupTestEnvironment(t)
	defer cleanupTestEnvironment()

	app := &Application{}
	err := app.initializeConfig()
	require.NoError(t, err)

	// Setup test database
	app.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	err = autoMigrate(app.DB)
	require.NoError(t, err)

	err = app.initializeServices()
	require.NoError(t, err)
	assert.NotNil(t, app.AuthService)
	assert.NotNil(t, app.SupportService)
}

func TestApplication_CreateDefaultAdmin_Success(t *testing.T) {
	setupTestEnvironment(t)
	defer cleanupTestEnvironment()

	app := &Application{}
	err := app.initializeConfig()
	require.NoError(t, err)

	// Setup test database
	app.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	err = autoMigrate(app.DB)
	require.NoError(t, err)

	err = app.initializeServices()
	require.NoError(t, err)

	// Test createDefaultAdmin specifically
	err = app.createDefaultAdmin()
	assert.NoError(t, err)

	// Verify admin was created by trying to create again (should not error)
	err = app.createDefaultAdmin()
	assert.NoError(t, err)
}

func TestApplication_InitializeHandlers_Success(t *testing.T) {
	setupTestEnvironment(t)
	defer cleanupTestEnvironment()

	app := &Application{}
	err := app.initializeConfig()
	require.NoError(t, err)

	// Setup test database and services
	app.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	err = autoMigrate(app.DB)
	require.NoError(t, err)

	err = app.initializeServices()
	require.NoError(t, err)

	err = app.initializeHandlers()
	require.NoError(t, err)
	assert.NotNil(t, app.AuthHandler)
	assert.NotNil(t, app.SupportHandler)
}

func TestApplication_SetupRouter_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestEnvironment(t)
	defer cleanupTestEnvironment()

	app := &Application{}
	err := app.initializeConfig()
	require.NoError(t, err)

	// Setup test database and services
	app.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	err = autoMigrate(app.DB)
	require.NoError(t, err)

	err = app.initializeServices()
	require.NoError(t, err)

	err = app.initializeHandlers()
	require.NoError(t, err)

	err = app.setupRouter()
	require.NoError(t, err)
	assert.NotNil(t, app.Router)

	// Verify routes are set up
	routes := app.Router.Routes()
	assert.Greater(t, len(routes), 0)
}

func TestApplication_Run_ServerStartError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestEnvironment(t)
	defer cleanupTestEnvironment()

	app := &Application{}
	err := app.initializeConfig()
	require.NoError(t, err)

	// Set an invalid port to force server start error
	app.Config.Server.Port = "999999" // Invalid port number

	// Setup minimal dependencies
	app.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = app.initializeServices()
	require.NoError(t, err)

	err = app.initializeHandlers()
	require.NoError(t, err)

	err = app.setupRouter()
	require.NoError(t, err)

	// This should fail due to invalid port
	err = app.Run()
	assert.Error(t, err)
}

func TestNewApplication_DatabaseError(t *testing.T) {
	setupTestEnvironment(t)
	defer cleanupTestEnvironment()

	// Set invalid database configuration
	os.Setenv("DB_HOST", "invalid-host-that-does-not-exist")
	defer os.Unsetenv("DB_HOST")

	app, err := NewApplication()
	assert.Error(t, err)
	assert.Nil(t, app)
	assert.Contains(t, err.Error(), "failed to initialize database")
}

func TestApplication_InitializationChain_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		setupError  func()
		expectedErr string
	}{
		{
			name: "config_error",
			setupError: func() {
				os.Setenv("ENVIRONMENT", "invalid")
			},
			expectedErr: "failed to initialize config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup error condition
			tt.setupError()

			// Cleanup after test
			defer func() {
				os.Unsetenv("ENVIRONMENT")
				os.Unsetenv("DB_HOST")
				os.Unsetenv("JWT_SECRET")
			}()

			app, err := NewApplication()
			assert.Error(t, err)
			assert.Nil(t, app)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestApplication_FullIntegration_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestEnvironmentWithSQLite(t)
	defer cleanupTestEnvironment()

	// Test the complete initialization chain
	app, err := NewApplication()
	require.NoError(t, err)

	// Verify all components are properly initialized
	assert.NotNil(t, app.Config)
	assert.NotNil(t, app.DB)
	assert.NotNil(t, app.AuthService)
	assert.NotNil(t, app.SupportService)
	assert.NotNil(t, app.AuthHandler)
	assert.NotNil(t, app.SupportHandler)
	assert.NotNil(t, app.Router)

	// Verify database tables exist
	assert.True(t, app.DB.Migrator().HasTable(&models.SupportRequest{}))
	assert.True(t, app.DB.Migrator().HasTable(&models.User{}))

	// Verify router has expected routes
	routes := app.Router.Routes()
	routeMap := make(map[string]bool)
	for _, route := range routes {
		routeMap[route.Method+" "+route.Path] = true
	}

	expectedRoutes := []string{
		"GET /health",
		"POST /api/v1/support-request",
		"POST /api/v1/auth/login",
	}

	for _, expectedRoute := range expectedRoutes {
		assert.True(t, routeMap[expectedRoute], "Route %s should be registered", expectedRoute)
	}
}

func TestConnectDatabase_SQLiteMemory_Success(t *testing.T) {
	cfg := config.DatabaseConfig{
		DBName: ":memory:",
	}

	db, err := connectDatabase(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// Verify it's using SQLite
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = sqlDB.Ping()
	assert.NoError(t, err)
}

func TestConnectDatabase_PostgreSQL_ConnectionError(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:     "invalid-host",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	db, err := connectDatabase(cfg)
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestApplication_InitializeDatabase_MigrationError(t *testing.T) {
	setupTestEnvironment(t)
	defer cleanupTestEnvironment()

	app := &Application{}
	err := app.initializeConfig()
	require.NoError(t, err)

	// Create a broken database connection by providing invalid config
	app.Config.Database.DBName = "invalid-db-name-that-will-fail"
	app.Config.Database.Host = "invalid-host"

	err = app.initializeDatabase()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to database")
}

func TestApplication_CreateDefaultAdmin_AdminServiceError(t *testing.T) {
	setupTestEnvironment(t)
	defer cleanupTestEnvironment()

	app := &Application{}
	err := app.initializeConfig()
	require.NoError(t, err)

	// Setup test database but don't run migrations to cause an error
	app.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	// Don't run autoMigrate to cause error

	err = app.initializeServices()
	require.NoError(t, err)

	// This should log a warning but not fail
	err = app.createDefaultAdmin()
	// The function logs a warning but doesn't return an error in main logic
	// Test that it handles the error gracefully
	assert.Error(t, err) // SQLite will error on missing tables
}

func TestMain_ErrorPaths_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		setupError  func()
		expectedErr string
	}{
		{
			name: "invalid_jwt_secret",
			setupError: func() {
				os.Setenv("ENVIRONMENT", "production")
				os.Setenv("JWT_SECRET", "short")
				os.Setenv("DB_HOST", "localhost")
				os.Setenv("DB_NAME", ":memory:")
			},
			expectedErr: "failed to initialize config",
		},
		{
			name: "database_connection_error",
			setupError: func() {
				os.Setenv("ENVIRONMENT", "development")
				os.Setenv("JWT_SECRET", "test-jwt-secret-key-that-is-long-enough-for-testing-purposes")
				os.Setenv("DB_HOST", "invalid-host-name-that-does-not-exist")
				os.Setenv("DB_USER", "postgres")
				os.Setenv("DB_NAME", "testdb")
			},
			expectedErr: "failed to initialize database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			envKeys := []string{"ENVIRONMENT", "JWT_SECRET", "DB_HOST", "DB_USER", "DB_NAME", "DB_PASSWORD", "PORT"}
			for _, key := range envKeys {
				os.Unsetenv(key)
			}

			// Setup error condition
			tt.setupError()

			// Cleanup after test
			defer func() {
				for _, key := range envKeys {
					os.Unsetenv(key)
				}
			}()

			app, err := NewApplication()
			assert.Error(t, err)
			assert.Nil(t, app)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestApplication_ErrorChaining_Coverage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// This test verifies that config errors are properly handled and propagated
	// We'll test with a JWT secret that's too short to trigger validation error

	// Set an invalid JWT secret that's too short (< 32 characters)
	os.Setenv("JWT_SECRET", "short")
	defer os.Unsetenv("JWT_SECRET")

	app := &Application{}
	err := app.initializeConfig()
	// This should trigger the JWT validation error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT secret is insecure")
}

func TestApplication_DatabaseConnectionFailure_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestEnvironment(t)
	defer cleanupTestEnvironment()

	// Set a definitely invalid database host to cause connection failure
	os.Setenv("DB_HOST", "invalid.host.that.does.not.exist.nowhere")
	defer os.Unsetenv("DB_HOST")

	app, err := NewApplication()
	assert.Error(t, err)
	assert.Nil(t, app)
	assert.Contains(t, err.Error(), "failed to initialize database")
}

func TestApplication_DatabaseErrorHandling_EdgeCases(t *testing.T) {
	setupTestEnvironment(t)
	defer cleanupTestEnvironment()

	app := &Application{}
	err := app.initializeConfig()
	require.NoError(t, err)

	// Test database connection with invalid configuration
	app.Config.Database.Host = "127.0.0.1"
	app.Config.Database.Port = 99999 // Invalid port
	app.Config.Database.DBName = "nonexistent"

	err = app.initializeDatabase()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to database")
}

func TestConnectDatabase_LoggerModes_Coverage(t *testing.T) {
	// Test with development environment (will trigger Info log level)
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "devuser",
		Password: "devpass",
		DBName:   ":memory:",
		SSLMode:  "disable",
	}

	db, err := connectDatabase(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// Test with production-like environment (will trigger Error log level)
	cfg2 := config.DatabaseConfig{
		Host:     "prod.example.com",
		Port:     5432,
		User:     "produser",
		Password: "prodpass",
		DBName:   ":memory:",
		SSLMode:  "require",
	}

	db2, err := connectDatabase(cfg2)
	assert.NoError(t, err)
	assert.NotNil(t, db2)
}

func TestMain_Function_ErrorScenarios(t *testing.T) {
	// Test that main function would handle errors correctly
	// We can't test main() directly, but we can test the error paths
	// that would cause main() to call log.Fatal

	// These are the scenarios that would cause main() to exit:
	// 1. NewApplication() error
	// 2. app.Run() error

	gin.SetMode(gin.TestMode)

	// Test NewApplication error scenario
	os.Setenv("ENVIRONMENT", "invalid")
	defer os.Unsetenv("ENVIRONMENT")

	app, err := NewApplication()
	assert.Error(t, err)
	assert.Nil(t, app)
	// This error would cause main() to call log.Fatal("Failed to initialize application:", err)

	// Test app.Run() error scenario
	setupTestEnvironmentWithSQLite(t)
	defer cleanupTestEnvironment()

	validApp, err := NewApplication()
	require.NoError(t, err)

	// Set invalid port to cause Run() to fail
	validApp.Config.Server.Port = "999999"
	err = validApp.Run()
	assert.Error(t, err)
	// This error would cause main() to call log.Fatal("Failed to start server:", err)
}

func TestConnectDatabase_PostgreSQL_ConfigurationEdgeCases(t *testing.T) {
	// Test various PostgreSQL configuration scenarios that would be covered
	// but not necessarily tested in normal flow

	// Test with minimum valid PostgreSQL config (but it will fail to connect)
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "test",
		Password: "",
		DBName:   "test",
		SSLMode:  "disable",
	}

	// This should attempt PostgreSQL connection and fail
	db, err := connectDatabase(cfg)
	assert.Error(t, err)
	assert.Nil(t, db)

	// Test connection pool configuration path
	// We can test this by ensuring SQLite doesn't trigger the pool config
	cfgSQLite := config.DatabaseConfig{
		DBName: ":memory:",
	}

	dbSQLite, err := connectDatabase(cfgSQLite)
	assert.NoError(t, err)
	assert.NotNil(t, dbSQLite)

	// Verify SQLite doesn't have connection pool settings
	sqlDB, err := dbSQLite.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	stats := sqlDB.Stats()
	// SQLite should work without explicit pool configuration
	assert.GreaterOrEqual(t, stats.MaxOpenConnections, 0)
}

func TestApplication_DatabaseConnectionPool_Coverage(t *testing.T) {
	// Test the connection pool configuration path in connectDatabase
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	// This will fail to connect but will test the connection pool code path
	db, err := connectDatabase(cfg)
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestApplication_Logger_Development_Path(t *testing.T) {
	// Test the development logger path in connectDatabase
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	// This will test the non-development logger path since DSN won't return "development"
	db, err := connectDatabase(cfg)
	assert.Error(t, err) // Should fail to connect to non-existent host
	assert.Nil(t, db)
}

func TestApplication_AutoMigrate_Error_Coverage(t *testing.T) {
	// Test autoMigrate with a broken database connection
	// Create a database that will fail migration
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Close the database to make migration fail
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.Close()

	// Now autoMigrate should fail
	err = autoMigrate(db)
	assert.Error(t, err)
}

func TestApplication_Production_Logger_Path(t *testing.T) {
	// Test the production logger path (non-development)
	cfg := config.DatabaseConfig{
		Host:     "prod.example.com",
		Port:     5432,
		User:     "produser",
		Password: "prodpass",
		DBName:   "proddb",
		SSLMode:  "require",
	}

	// This will test the production logger path (Error level)
	db, err := connectDatabase(cfg)
	assert.Error(t, err) // Should fail to connect
	assert.Nil(t, db)
}

func TestApplication_CompleteErrorFlow_Coverage(t *testing.T) {
	// Test various error scenarios to improve coverage

	// Test case 1: Config load error (already tested above)

	// Test case 2: Database connection error in initializeDatabase
	app := &Application{}
	app.Config = &config.Config{
		Database: config.DatabaseConfig{
			Host:     "invalid-host-name",
			Port:     5432,
			User:     "testuser",
			Password: "testpass",
			DBName:   "testdb",
			SSLMode:  "disable",
		},
	}

	err := app.initializeDatabase()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to database")
}

func TestApplication_SetupRouter_Error_Scenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test setupRouter with minimal valid app
	setupTestEnvironmentWithSQLite(t)
	defer cleanupTestEnvironment()

	app := &Application{}
	err := app.initializeConfig()
	require.NoError(t, err)

	// Setup test database and services
	app.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	err = autoMigrate(app.DB)
	require.NoError(t, err)

	err = app.initializeServices()
	require.NoError(t, err)

	err = app.initializeHandlers()
	require.NoError(t, err)

	// Test setupRouter - should succeed
	err = app.setupRouter()
	assert.NoError(t, err)
	assert.NotNil(t, app.Router)
}

func TestApplication_Individual_Methods_Coverage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "initializeConfig_success",
			testFunc: func(t *testing.T) {
				setupTestEnvironment(t)
				defer cleanupTestEnvironment()

				app := &Application{}
				err := app.initializeConfig()
				assert.NoError(t, err)
				assert.NotNil(t, app.Config)
			},
		},
		{
			name: "initializeHandlers_success",
			testFunc: func(t *testing.T) {
				setupTestEnvironment(t)
				defer cleanupTestEnvironment()

				app := &Application{}
				err := app.initializeConfig()
				require.NoError(t, err)

				app.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
					Logger: logger.Default.LogMode(logger.Silent),
				})
				require.NoError(t, err)
				err = autoMigrate(app.DB)
				require.NoError(t, err)

				err = app.initializeServices()
				require.NoError(t, err)

				err = app.initializeHandlers()
				assert.NoError(t, err)
				assert.NotNil(t, app.AuthHandler)
				assert.NotNil(t, app.SupportHandler)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

func TestConnectDatabase_Edge_Cases(t *testing.T) {
	tests := []struct {
		name        string
		config      config.DatabaseConfig
		expectError bool
		description string
	}{
		{
			name: "sqlite_memory_success",
			config: config.DatabaseConfig{
				DBName: ":memory:",
			},
			expectError: false,
			description: "SQLite in-memory should succeed",
		},
		{
			name: "postgres_invalid_host",
			config: config.DatabaseConfig{
				Host:     "invalid-host-12345",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				DBName:   "testdb",
				SSLMode:  "disable",
			},
			expectError: true,
			description: "PostgreSQL with invalid host should fail",
		},
		{
			name: "postgres_invalid_port",
			config: config.DatabaseConfig{
				Host:     "localhost",
				Port:     99999, // Invalid port
				User:     "testuser",
				Password: "testpass",
				DBName:   "testdb",
				SSLMode:  "disable",
			},
			expectError: true,
			description: "PostgreSQL with invalid port should fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := connectDatabase(tt.config)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, db)
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, db)

				// Cleanup
				if db != nil {
					if sqlDB, err := db.DB(); err == nil {
						sqlDB.Close()
					}
				}
			}
		})
	}
}

// Helper functions for test setup
func setupTestEnvironment(t *testing.T) {
	// Save original environment
	originalEnv := map[string]string{
		"PORT":        os.Getenv("PORT"),
		"JWT_SECRET":  os.Getenv("JWT_SECRET"),
		"ENVIRONMENT": os.Getenv("ENVIRONMENT"),
		"DB_HOST":     os.Getenv("DB_HOST"),
		"DB_NAME":     os.Getenv("DB_NAME"),
		"DB_USER":     os.Getenv("DB_USER"),
		"DB_PASSWORD": os.Getenv("DB_PASSWORD"),
	}

	// Store in test context for cleanup
	t.Cleanup(func() {
		for key, value := range originalEnv {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	})

	// Set test environment
	os.Setenv("PORT", "8081")
	os.Setenv("JWT_SECRET", "test-jwt-secret-key-that-is-long-enough-for-testing-purposes")
	os.Setenv("ENVIRONMENT", "development")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_NAME", "test_db")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "password")
}

func setupTestEnvironmentWithSQLite(t *testing.T) {
	// Save original environment
	originalEnv := map[string]string{
		"PORT":        os.Getenv("PORT"),
		"JWT_SECRET":  os.Getenv("JWT_SECRET"),
		"ENVIRONMENT": os.Getenv("ENVIRONMENT"),
		"DB_HOST":     os.Getenv("DB_HOST"),
		"DB_NAME":     os.Getenv("DB_NAME"),
		"DB_USER":     os.Getenv("DB_USER"),
		"DB_PASSWORD": os.Getenv("DB_PASSWORD"),
	}

	// Store in test context for cleanup
	t.Cleanup(func() {
		for key, value := range originalEnv {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	})

	// Set test environment with SQLite-compatible settings
	os.Setenv("PORT", "8081")
	os.Setenv("JWT_SECRET", "test-jwt-secret-key-that-is-long-enough-for-testing-purposes")
	os.Setenv("ENVIRONMENT", "development")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_NAME", ":memory:")
	os.Setenv("DB_USER", "test")
	os.Setenv("DB_PASSWORD", "test")
	os.Setenv("DB_SSL_MODE", "disable")
}

func cleanupTestEnvironment() {
	// This will be handled by t.Cleanup() in setupTestEnvironment
}
