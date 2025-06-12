package main

import (
	"log"
	"support-app-backend/internal/config"
	"support-app-backend/internal/handlers"
	"support-app-backend/internal/middleware"
	"support-app-backend/internal/models"
	"support-app-backend/internal/repositories"
	"support-app-backend/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Connect to database
	db, err := connectDatabase(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate database
	if err := autoMigrate(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize repositories
	supportRepo := repositories.NewSupportRequestRepository(db)
	userRepo := repositories.NewUserRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg.JWT.SecretKey)
	supportService := services.NewSupportRequestService(supportRepo)

	// Create default admin account
	if err := authService.CreateDefaultAdmin(); err != nil {
		log.Printf("Warning: Failed to create default admin account: %v", err)
	} else {
		log.Println("✅ Default admin account ready (username: admin, password: securePassword@123)")
	}

	// Initialize handlers
	supportHandler := handlers.NewSupportRequestHandler(supportService)
	authHandler := handlers.NewAuthHandler(authService)

	// Setup router
	router := setupRouter(cfg, supportHandler, authHandler, authService)

	// Start server
	log.Printf("Starting server on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func connectDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	var gormLogger logger.Interface

	// Set appropriate log level based on environment
	if cfg.GetDSN() == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Error)
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
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

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
