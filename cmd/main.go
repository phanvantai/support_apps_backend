// @title Support App Backend API
// @version 1.0
// @description RESTful API for support tickets management with authentication and authorization
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.support-app.io/support
// @contact.email support@support-app.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"fmt"
	"log"
	"support-app-backend/docs"
	"support-app-backend/internal/config"
	"support-app-backend/internal/handlers"
	"support-app-backend/internal/middleware"
	"support-app-backend/internal/models"
	"support-app-backend/internal/repositories"
	"support-app-backend/internal/services"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Application holds all application dependencies
type Application struct {
	Config         *config.Config
	DB             *gorm.DB
	AuthService    services.AuthService
	SupportService services.SupportRequestService
	AuthHandler    *handlers.AuthHandler
	SupportHandler *handlers.SupportRequestHandler
	Router         *gin.Engine
}

func main() {
	app, err := NewApplication()
	if err != nil {
		log.Fatal("Failed to initialize application:", err)
	}

	if err := app.Run(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// NewApplication creates and initializes a new application instance
func NewApplication() (*Application, error) {
	app := &Application{}

	// Initialize configuration
	if err := app.initializeConfig(); err != nil {
		return nil, fmt.Errorf("failed to initialize config: %w", err)
	}

	// Initialize database
	if err := app.initializeDatabase(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize services
	if err := app.initializeServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	// Initialize handlers
	if err := app.initializeHandlers(); err != nil {
		return nil, fmt.Errorf("failed to initialize handlers: %w", err)
	}

	// Setup router
	if err := app.setupRouter(); err != nil {
		return nil, fmt.Errorf("failed to setup router: %w", err)
	}

	return app, nil
}

// initializeConfig loads and validates the application configuration
func (app *Application) initializeConfig() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	app.Config = cfg
	return nil
}

// initializeDatabase connects to the database and runs migrations
func (app *Application) initializeDatabase() error {
	// Connect to database
	db, err := connectDatabase(app.Config.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	app.DB = db

	// Auto migrate database
	if err := autoMigrate(db); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

// initializeServices creates and initializes all application services
func (app *Application) initializeServices() error {
	// Initialize repositories
	supportRepo := repositories.NewSupportRequestRepository(app.DB)
	userRepo := repositories.NewUserRepository(app.DB)

	// Initialize services
	app.AuthService = services.NewAuthService(userRepo, app.Config.JWT.SecretKey)
	app.SupportService = services.NewSupportRequestService(supportRepo)

	// Create default admin account
	if err := app.createDefaultAdmin(); err != nil {
		log.Printf("Warning: Failed to create default admin account: %v", err)
	} else {
		log.Println("✅ Default admin account ready (username: admin, password: securePassword@123)")
	}

	return nil
}

// createDefaultAdmin creates the default admin account
func (app *Application) createDefaultAdmin() error {
	return app.AuthService.CreateDefaultAdmin()
}

// initializeHandlers creates and initializes all HTTP handlers
func (app *Application) initializeHandlers() error {
	app.SupportHandler = handlers.NewSupportRequestHandler(app.SupportService)
	app.AuthHandler = handlers.NewAuthHandler(app.AuthService)
	return nil
}

// setupRouter configures and sets up the HTTP router
func (app *Application) setupRouter() error {
	app.Router = setupRouter(app.Config, app.SupportHandler, app.AuthHandler, app.AuthService)
	return nil
}

// Run starts the HTTP server
func (app *Application) Run() error {
	log.Printf("Starting server on port %s", app.Config.Server.Port)

	// Log Swagger documentation URL
	var swaggerURL string
	if app.Config.Server.PublicDomain != "" && app.Config.Server.PublicDomain != "localhost:"+app.Config.Server.Port {
		// Railway deployment or custom domain
		swaggerURL = fmt.Sprintf("https://%s/swagger/index.html", app.Config.Server.PublicDomain)
	} else {
		// Local development
		swaggerURL = fmt.Sprintf("http://localhost:%s/swagger/index.html", app.Config.Server.Port)
	}
	log.Printf("📖 Swagger documentation available at: %s", swaggerURL)

	return app.Router.Run(":" + app.Config.Server.Port)
}

func connectDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	var gormLogger logger.Interface

	// Set appropriate log level based on environment
	if cfg.GetDSN() == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	// For testing with in-memory SQLite
	if cfg.DBName == ":memory:" {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			return nil, err
		}
		return db, nil
	}

	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	return db, nil
}

func autoMigrate(db *gorm.DB) error {
	// Run migrations for both tables
	// GORM will handle schema changes gracefully
	return db.AutoMigrate(&models.SupportRequest{}, &models.User{})
}

func setupRouter(cfg *config.Config, supportHandler *handlers.SupportRequestHandler, authHandler *handlers.AuthHandler, authService services.AuthService) *gin.Engine {
	// Set Gin mode based on environment
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		
		// Set CORS headers
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "false")
		c.Header("Access-Control-Max-Age", "86400")

		// Handle preflight OPTIONS request
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Initialize Swagger docs
	if cfg.Server.PublicDomain != "" {
		docs.SwaggerInfo.Host = cfg.Server.PublicDomain
	} else {
		// Fallback for Railway during initial deployment
		docs.SwaggerInfo.Host = "localhost:" + cfg.Server.Port
	}

	// Swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// Health check endpoint (no authentication required)
	router.GET("/health", supportHandler.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public endpoints (with rate limiting)
		rateLimiter := middleware.NewRateLimitMiddleware(cfg.Server.RateLimit, cfg.Server.RateBurst)
		v1.POST("/support-request", rateLimiter.Middleware(), supportHandler.CreateSupportRequest)

		// Authentication endpoints
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)

			// Protected auth endpoints (require authentication)
			authProtected := auth.Group("")
			authProtected.Use(middleware.AuthMiddleware(authService))
			{
				authProtected.GET("/me", authHandler.GetCurrentUser)
				authProtected.PATCH("/password", authHandler.ChangePassword)

				// Admin-only user management endpoints
				adminAuth := authProtected.Group("")
				adminAuth.Use(middleware.AdminOnlyMiddleware())
				{
					adminAuth.POST("/users", authHandler.CreateUser)
					adminAuth.GET("/users", authHandler.GetAllUsers)
					adminAuth.GET("/users/:id", authHandler.GetUser)
					adminAuth.PATCH("/users/:id", authHandler.UpdateUser)
					adminAuth.DELETE("/users/:id", authHandler.DeleteUser)
				}
			}
		}

		// Admin endpoints for support requests (require authentication)
		admin := v1.Group("/support-requests")
		admin.Use(middleware.AuthMiddleware(authService))
		admin.Use(middleware.AdminOnlyMiddleware())
		{
			admin.GET("", supportHandler.GetAllSupportRequests)
			admin.GET("/:id", supportHandler.GetSupportRequest)
			admin.PATCH("/:id", supportHandler.UpdateSupportRequest)
			admin.DELETE("/:id", supportHandler.DeleteSupportRequest)
		}
	}

	return router
}
