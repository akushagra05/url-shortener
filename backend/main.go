package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/url-shortener/cache"
	"github.com/yourusername/url-shortener/config"
	"github.com/yourusername/url-shortener/database"
	"github.com/yourusername/url-shortener/handlers"
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

	// Seed reserved keywords
	if err := database.SeedReservedKeywords(); err != nil {
		log.Fatalf("Failed to seed reserved keywords: %v", err)
	}

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

	// Initialize handlers
	cacheTTL := time.Duration(cfg.Cache.TTLSeconds) * time.Second
	handlers.InitHandlers(snowflakeGen, cfg.App.BaseURL, cacheTTL)

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
	router.GET("/health", handlers.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// URL shortening endpoint with rate limiting
		v1.POST("/shorten",
			middleware.CreateURLRateLimiter(cfg.RateLimit.CreateLimit),
			handlers.CreateShortURL,
		)

		// URL details and analytics endpoints
		v1.GET("/url/:shortCode", handlers.GetURLDetails)
		v1.GET("/url/:shortCode/analytics", handlers.GetURLAnalytics)
		v1.DELETE("/url/:shortCode", handlers.DeleteURL)
	}

	// Redirect endpoint (short code to original URL) with rate limiting
	router.GET("/:shortCode",
		middleware.RedirectRateLimiter(cfg.RateLimit.RedirectLimit),
		handlers.RedirectToOriginalURL,
	)

	// Start server
	serverAddr := ":" + cfg.Server.Port
	log.Printf("Starting server on %s", serverAddr)
	log.Printf("Base URL: %s", cfg.App.BaseURL)
	log.Printf("Frontend URL: %s", cfg.App.FrontendURL)
	log.Printf("Machine ID: %d", cfg.Snowflake.MachineID)

	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
