package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/url-shortener/cache"
	"github.com/yourusername/url-shortener/database"
)

// HealthCheck handles GET /health
func HealthCheck(c *gin.Context) {
	dbStatus := "up"
	redisStatus := "up"

	// Check database
	sqlDB, err := database.DB.DB()
	if err != nil || sqlDB.Ping() != nil {
		dbStatus = "down"
	}

	// Check Redis
	if err := cache.HealthCheck(); err != nil {
		redisStatus = "down"
	}

	overallStatus := "healthy"
	statusCode := http.StatusOK
	
	if dbStatus == "down" || redisStatus == "down" {
		overallStatus = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"status":    overallStatus,
		"timestamp": c.Request.Context().Value("timestamp"),
		"services": gin.H{
			"database": dbStatus,
			"redis":    redisStatus,
		},
	})
}
