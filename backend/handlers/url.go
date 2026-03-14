package handlers

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/url-shortener/cache"
	"github.com/yourusername/url-shortener/database"
	"github.com/yourusername/url-shortener/models"
	"github.com/yourusername/url-shortener/pkg/snowflake"
	"gorm.io/gorm"
)

var (
	snowflakeGen   *snowflake.Generator
	baseURL        string
	cacheTTL       time.Duration
	validURLRegex  = regexp.MustCompile(`^https?://`)
	validAliasRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// InitHandlers initializes the handlers with required dependencies
func InitHandlers(gen *snowflake.Generator, base string, ttl time.Duration) {
	snowflakeGen = gen
	baseURL = base
	cacheTTL = ttl
}

// CreateShortURL handles POST /api/v1/shorten
func CreateShortURL(c *gin.Context) {
	var req models.CreateURLRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body: " + err.Error(),
			},
		})
		return
	}

	// Validate URL
	if !isValidURL(req.URL) {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_URL",
				Message: "URL must start with http:// or https:// and be valid",
			},
		})
		return
	}

	// Check URL length
	if len(req.URL) > 2048 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "URL_TOO_LONG",
				Message: "URL must be less than 2048 characters",
			},
		})
		return
	}

	var shortCode string
	customAlias := false

	// Handle custom alias
	if req.CustomAlias != "" {
		if !isValidAlias(req.CustomAlias) {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "INVALID_ALIAS",
					Message: "Custom alias must be 4-20 characters and contain only letters, numbers, hyphens, and underscores",
				},
			})
			return
		}

		// Check if reserved
		if database.IsReservedKeyword(req.CustomAlias) {
			c.JSON(http.StatusConflict, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "ALIAS_RESERVED",
					Message: "This alias is reserved and cannot be used",
				},
			})
			return
		}

		// Check if already exists
		if _, err := database.GetURL(req.CustomAlias); err == nil {
			c.JSON(http.StatusConflict, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "ALIAS_ALREADY_EXISTS",
					Message: "Custom alias '" + req.CustomAlias + "' is already taken",
				},
			})
			return
		}

		shortCode = req.CustomAlias
		customAlias = true
	} else {
		// Generate short code using Snowflake
		code, err := snowflakeGen.GenerateString()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "GENERATION_ERROR",
					Message: "Failed to generate short code",
				},
			})
			return
		}
		shortCode = code
	}

	// Create URL entry
	urlEntry := &models.URL{
		ShortCode:   shortCode,
		OriginalURL: req.URL,
		CustomAlias: customAlias,
		ExpiresAt:   req.ExpiresAt,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := database.CreateURL(urlEntry); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to create short URL",
			},
		})
		return
	}

	// Cache the URL
	cachedURL := &models.CachedURL{
		OriginalURL: urlEntry.OriginalURL,
		ClickCount:  0,
		ExpiresAt:   urlEntry.ExpiresAt,
		IsActive:    true,
	}
	cache.SetCachedURL(shortCode, cachedURL, cacheTTL)

	// Build response
	response := models.CreateURLResponse{
		ShortCode:   shortCode,
		ShortURL:    baseURL + "/" + shortCode,
		OriginalURL: urlEntry.OriginalURL,
		CreatedAt:   urlEntry.CreatedAt,
		ExpiresAt:   urlEntry.ExpiresAt,
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    response,
	})
}

// RedirectToOriginalURL handles GET /{short_code}
func RedirectToOriginalURL(c *gin.Context) {
	shortCode := c.Param("shortCode")

	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_SHORT_CODE",
				Message: "Short code is required",
			},
		})
		return
	}

	// Try cache first
	cachedURL, err := cache.GetCachedURL(shortCode)
	if err == nil && cachedURL != nil {
		// Check if expired
		if cachedURL.ExpiresAt != nil && time.Now().After(*cachedURL.ExpiresAt) {
			c.JSON(http.StatusGone, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "URL_EXPIRED",
					Message: "This short URL has expired",
				},
			})
			return
		}

		// Check if active
		if !cachedURL.IsActive {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "URL_NOT_FOUND",
					Message: "Short URL not found",
				},
			})
			return
		}

		// Increment click counter asynchronously
		go cache.IncrementClickCount(shortCode)
		go recordAnalytics(shortCode, c)

		c.Redirect(http.StatusFound, cachedURL.OriginalURL)
		return
	}

	// Cache miss - fetch from database
	urlEntry, err := database.GetURL(shortCode)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "URL_NOT_FOUND",
					Message: "Short URL not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to retrieve URL",
			},
		})
		return
	}

	// Check if expired
	if urlEntry.IsExpired() {
		c.JSON(http.StatusGone, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "URL_EXPIRED",
				Message: "This short URL has expired",
			},
		})
		return
	}

	// Cache the URL for future requests
	cache.CacheURLWithData(shortCode, urlEntry, cacheTTL)

	// Increment click counter
	go cache.IncrementClickCount(shortCode)
	go recordAnalytics(shortCode, c)

	c.Redirect(http.StatusFound, urlEntry.OriginalURL)
}

// GetURLDetails handles GET /api/v1/url/{short_code}
func GetURLDetails(c *gin.Context) {
	shortCode := c.Param("shortCode")

	urlEntry, err := database.GetURL(shortCode)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "URL_NOT_FOUND",
					Message: "Short URL not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to retrieve URL",
			},
		})
		return
	}

	// Get current click count from Redis
	redisClicks, _ := cache.GetClickCount(shortCode)
	totalClicks := urlEntry.ClickCount + redisClicks

	response := models.URLDetailsResponse{
		ShortCode:      shortCode,
		ShortURL:       baseURL + "/" + shortCode,
		OriginalURL:    urlEntry.OriginalURL,
		ClickCount:     totalClicks,
		CreatedAt:      urlEntry.CreatedAt,
		LastAccessedAt: urlEntry.LastAccessedAt,
		IsActive:       urlEntry.IsActive,
		ExpiresAt:      urlEntry.ExpiresAt,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    response,
	})
}

// GetURLAnalytics handles GET /api/v1/url/{short_code}/analytics
func GetURLAnalytics(c *gin.Context) {
	shortCode := c.Param("shortCode")
	days := 7 // Default to 7 days

	if daysParam := c.Query("days"); daysParam != "" {
		if d, err := time.ParseDuration(daysParam + "d"); err == nil {
			days = int(d.Hours() / 24)
		}
	}

	analytics, err := database.GetAnalytics(shortCode, days)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "URL_NOT_FOUND",
					Message: "Short URL not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to retrieve analytics",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    analytics,
	})
}

// DeleteURL handles DELETE /api/v1/url/{short_code}
func DeleteURL(c *gin.Context) {
	shortCode := c.Param("shortCode")

	if err := database.DeleteURL(shortCode); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to delete URL",
			},
		})
		return
	}

	// Invalidate cache
	cache.InvalidateURLCache(shortCode)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Short URL deleted successfully"},
	})
}

// Helper functions

func isValidURL(rawURL string) bool {
	if !validURLRegex.MatchString(rawURL) {
		return false
	}
	
	_, err := url.ParseRequestURI(rawURL)
	return err == nil
}

func isValidAlias(alias string) bool {
	if len(alias) < 4 || len(alias) > 20 {
		return false
	}
	return validAliasRegex.MatchString(alias)
}

func recordAnalytics(shortCode string, c *gin.Context) {
	analytics := &models.URLAnalytics{
		ShortCode:  shortCode,
		AccessedAt: time.Now(),
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
		Referer:    c.GetHeader("Referer"),
		DeviceType: detectDeviceType(c.GetHeader("User-Agent")),
	}

	database.CreateAnalytics(analytics)
}

func detectDeviceType(userAgent string) string {
	ua := strings.ToLower(userAgent)
	
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		return "mobile"
	}
	
	if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		return "tablet"
	}
	
	return "desktop"
}
