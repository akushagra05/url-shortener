
package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/url-shortener/cache"
	"github.com/yourusername/url-shortener/config"
	"github.com/yourusername/url-shortener/database"
	"github.com/yourusername/url-shortener/internal/handlers"
	"github.com/yourusername/url-shortener/internal/repository"
	"github.com/yourusername/url-shortener/internal/service"
	"github.com/yourusername/url-shortener/middleware"
	"github.com/yourusername/url-shortener/pkg/snowflake"
	"github.com/yourusername/url-shortener/workers"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Initialize database
	if err := database.InitDatabase(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDatabase()

	// Run migrations
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Reserved keywords are hardcoded in url_service.go

	// Initialize Redis
	if err := cache.InitRedis(cfg); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer cache.CloseRedis()

	// Initialize Snowflake ID generator
	snowflakeGen, err := snowflake.NewGenerator(
		cfg.Snowflake.MachineID,
		cfg.Snowflake.EpochTimestamp,
	)
	if err != nil {
		log.Fatalf("Failed to initialize Snowflake generator: %v", err)
	}

	// Initialize layers (Repository -> Service -> Handler)
	db := database.GetDB()
	cacheTTL := time.Duration(cfg.Cache.TTLSeconds) * time.Second

	// Repository layer
	urlRepo := repository.NewURLRepository(db)

	// Service layer
	urlService := service.NewURLService(
		urlRepo,
		snowflakeGen,
		cfg.App.BaseURL,
		cacheTTL,
	)

	// Handler layer
	urlHandler := handlers.NewURLHandler(urlService)
	healthHandler := handlers.NewHealthHandler()

	// Start background workers
	go workers.StartClickCountSyncWorker(
		time.Duration(cfg.Cache.ClickBufferTTLSeconds) * time.Second,
	)

	// Initialize Gin router
	router := gin.Default()

	// Apply global middleware
	allowedOrigins := []string{cfg.App.FrontendURL, "*"}
	router.Use(middleware.CORS(allowedOrigins))

	// Health check endpoint (no rate limiting)
	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/api/health", healthHandler.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// URL shortening endpoint with rate limiting
		v1.POST("/shorten",
			middleware.CreateURLRateLimiter(cfg.RateLimit.CreateLimit),
			urlHandler.CreateShortURL,
		)

		// URL management endpoints
		v1.GET("/url/:shortCode", urlHandler.GetURLDetails)
		v1.DELETE("/url/:shortCode", urlHandler.DeleteURL)
		
		// Analytics endpoint
		v1.GET("/url/:shortCode/analytics", urlHandler.GetAnalytics)
	}

	// Redirect endpoint (short code to original URL) with rate limiting
	router.GET("/:shortCode",
		middleware.RedirectRateLimiter(cfg.RateLimit.RedirectLimit),
		urlHandler.RedirectToOriginal,
	)

	// Start server
	serverAddr := ":" + cfg.Server.Port
	log.Printf("Starting server on %s", serverAddr)
	log.Printf("Base URL: %s", cfg.App.BaseURL)
	log.Printf("Frontend URL: %s", cfg.App.FrontendURL)
	log.Printf("Machine ID: %d", cfg.Snowflake.MachineID)
	log.Println("Server is ready to accept requests")

	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
