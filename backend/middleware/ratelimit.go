package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/url-shortener/cache"
	"github.com/yourusername/url-shortener/internal/models"
)

// RateLimiter middleware limits requests based on IP and endpoint
func RateLimiter(limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		endpoint := c.FullPath()
		key := ip + ":" + endpoint

		allowed, err := cache.CheckRateLimit(key, limit, int(window.Seconds()))
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "RATE_LIMIT_ERROR",
					Message: "Failed to check rate limit",
				},
			})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "RATE_LIMIT_EXCEEDED",
					Message: "Too many requests. Please try again later.",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CreateURLRateLimiter applies rate limiting for URL creation
func CreateURLRateLimiter(limit int) gin.HandlerFunc {
	return RateLimiter(limit, time.Minute)
}

// RedirectRateLimiter applies rate limiting for URL redirects
func RedirectRateLimiter(limit int) gin.HandlerFunc {
	return RateLimiter(limit, time.Minute)
}
