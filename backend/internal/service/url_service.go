
package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"time"

	"github.com/yourusername/url-shortener/cache"
	"github.com/yourusername/url-shortener/internal/models"
	"github.com/yourusername/url-shortener/internal/repository"
	"github.com/yourusername/url-shortener/pkg/snowflake"
)

var (
	validURLRegex   = regexp.MustCompile(`^https?://`)
	validAliasRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	
	// Reserved keywords that cannot be used as custom aliases
	// These protect critical system routes and common paths
	reservedWords = map[string]bool{
		// API & Versioning
		"api": true, "v1": true, "v2": true, "v3": true,
		
		// System & Health
		"health": true, "metrics": true, "status": true, "ping": true,
		
		// Admin & Auth
		"admin": true, "dashboard": true,
		"login": true, "logout": true, "register": true, "signup": true,
		
		// Documentation
		"docs": true, "swagger": true, "openapi": true,
		
		// Static & Assets
		"static": true, "assets": true, "public": true,
		"www": true, "app": true, "web": true,
	}
)

type URLService interface {
	CreateShortURL(ctx context.Context, req *models.CreateURLRequest) (*models.CreateURLResponse, error)
	GetOriginalURL(ctx context.Context, shortCode string) (string, error)
	GetURLDetails(ctx context.Context, shortCode string) (*models.URLDetailsResponse, error)
	GetAnalytics(ctx context.Context, shortCode string) (*models.URLAnalyticsResponse, error)
	DeleteURL(ctx context.Context, shortCode string) error
}

type urlService struct {
	repo         repository.URLRepository
	snowflakeGen *snowflake.Generator
	baseURL      string
	cacheTTL     time.Duration
}

func NewURLService(
	repo repository.URLRepository,
	snowflakeGen *snowflake.Generator,
	baseURL string,
	cacheTTL time.Duration,
) URLService {
	return &urlService{
		repo:         repo,
		snowflakeGen: snowflakeGen,
		baseURL:      baseURL,
		cacheTTL:     cacheTTL,
	}
}

func (s *urlService) CreateShortURL(ctx context.Context, req *models.CreateURLRequest) (*models.CreateURLResponse, error) {
	// Validate URL
	if err := s.validateURL(req.URL); err != nil {
		return nil, err
	}

	var shortCode string
	customAlias := false

	// Handle custom alias
	if req.CustomAlias != "" {
		if err := s.validateAlias(req.CustomAlias); err != nil {
			return nil, err
		}

		// Check if exists
		exists, err := s.repo.ExistsByShortCode(ctx, req.CustomAlias)
		if err != nil {
			return nil, fmt.Errorf("failed to check alias: %w", err)
		}
		if exists {
			return nil, errors.New("ALIAS_EXISTS: Custom alias already in use")
		}

		shortCode = req.CustomAlias
		customAlias = true
	} else {
		// Generate short code using Snowflake
		var err error
		shortCode, err = s.snowflakeGen.GenerateString()
		if err != nil {
			return nil, fmt.Errorf("failed to generate ID: %w", err)
		}
	}

	// Calculate expiration time
	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		// Use explicit timestamp
		expiresAt = req.ExpiresAt
	} else if req.ExpiresIn != "" {
		// Parse duration string (e.g., "7d", "24h", "30m")
		expiry, err := s.parseDuration(req.ExpiresIn)
		if err == nil {
			expiresAt = &expiry
		}
	}
	
	// Create URL record
	urlRecord := &models.URL{
		OriginalURL:    req.URL,
		ShortCode:      shortCode,
		CustomAlias:    customAlias,
		CreatedAt:      time.Now(),
		ExpiresAt:      expiresAt,
		LastAccessedAt: nil,
		ClickCount:     0,
	}

	if err := s.repo.Create(ctx, urlRecord); err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	// Build response
	response := &models.CreateURLResponse{
		ShortCode:   shortCode,
		ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, shortCode),
		OriginalURL: req.URL,
		CustomAlias: customAlias,
		CreatedAt:   urlRecord.CreatedAt,
		ExpiresAt:   req.ExpiresAt,
	}

	return response, nil
}

func (s *urlService) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	// Try cache first
	cached, err := cache.GetCachedURL(shortCode)
	if err == nil && cached != nil {
		// Check if cached URL has expired (compare in UTC)
		if cached.ExpiresAt != nil {
			now := time.Now().UTC()
			expiry := cached.ExpiresAt.UTC()
			fmt.Printf("[DEBUG] Expiry check for %s:\n", shortCode)
			fmt.Printf("  Current UTC: %s\n", now.Format(time.RFC3339))
			fmt.Printf("  Expires UTC: %s\n", expiry.Format(time.RFC3339))
			fmt.Printf("  Is Expired?: %v\n", now.After(expiry))
			
			if now.After(expiry) {
				cache.DeleteCachedURL(shortCode)
				return "", errors.New("URL_EXPIRED: This short URL has expired")
			}
		}
		
		// Check if cached URL is active
		if !cached.IsActive {
			cache.DeleteCachedURL(shortCode)
			return "", errors.New("URL_NOT_FOUND: Short URL not found")
		}
		
		// Cache hit! Only update Redis counter (no DB call for performance)
		go cache.IncrementClickCount(shortCode)
		return cached.OriginalURL, nil
	}
	
	// Cache miss: fetch from database
	urlRecord, err := s.repo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	if urlRecord == nil {
		return "", errors.New("URL_NOT_FOUND: Short URL not found")
	}

	// Check if URL has expired (compare in UTC)
	if urlRecord.ExpiresAt != nil {
		now := time.Now().UTC()
		expiry := urlRecord.ExpiresAt.UTC()
		fmt.Printf("[DEBUG DB] Expiry check for %s:\n", shortCode)
		fmt.Printf("  Current UTC: %s\n", now.Format(time.RFC3339))
		fmt.Printf("  Expires UTC: %s\n", expiry.Format(time.RFC3339))
		fmt.Printf("  Is Expired?: %v\n", now.After(expiry))
		
		if now.After(expiry) {
			return "", errors.New("URL_EXPIRED: This short URL has expired")
		}
	}

	// Check if URL is active
	if !urlRecord.IsActive {
		return "", errors.New("URL_NOT_FOUND: Short URL not found")
	}

	// Store in cache for future requests
	go cache.SetCachedURL(shortCode, urlRecord.OriginalURL, urlRecord.ExpiresAt, urlRecord.IsActive)

	// Increment click count in DB
	if err := s.repo.IncrementClickCount(ctx, urlRecord.ID); err != nil {
		fmt.Printf("Warning: failed to increment click count: %v\n", err)
	}

	// Increment click count in Redis for aggregation
	go cache.IncrementClickCount(shortCode)

	return urlRecord.OriginalURL, nil
}

func (s *urlService) GetAnalytics(ctx context.Context, shortCode string) (*models.URLAnalyticsResponse, error) {
	// Fetch from database
	urlData, err := s.repo.GetAnalytics(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch analytics: %w", err)
	}
	if urlData == nil {
		return nil, errors.New("URL_NOT_FOUND: Short URL not found")
	}

	// Get buffered clicks from Redis
	totalClicks := urlData.ClickCount
	bufferedClicks, err := cache.GetClickCount(shortCode)
	if err == nil {
		totalClicks += bufferedClicks
	}

	// Build response matching frontend expectations
	response := &models.URLAnalyticsResponse{
		ShortCode:   urlData.ShortCode,
		OriginalURL: urlData.OriginalURL,
		TotalClicks: totalClicks,  // Frontend expects "total_clicks"
		CreatedAt:   urlData.CreatedAt,
		LastAccessed: urlData.LastAccessedAt,
		Analytics: models.AnalyticsData{
			// For MVP, return empty arrays (detailed analytics not yet implemented)
			ClicksByDate:   []models.ClicksByDate{},
			TopCountries:   []models.TopItem{},
			TopReferrers:   []models.TopItem{},
			Devices: models.DeviceStats{
				Mobile:  0,
				Desktop: 0,
				Tablet:  0,
			},
		},
	}

	return response, nil
}


func (s *urlService) GetURLDetails(ctx context.Context, shortCode string) (*models.URLDetailsResponse, error) {
	// Fetch from database
	urlRecord, err := s.repo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	if urlRecord == nil {
		return nil, errors.New("URL_NOT_FOUND: Short URL not found")
	}

	// Get buffered clicks from Redis
	bufferedClicks, err := cache.GetClickCount(shortCode)
	if err == nil {
		urlRecord.ClickCount += bufferedClicks
	}
	
	// Get last access time from Redis (more recent than DB)
	lastAccessedAt := urlRecord.LastAccessedAt
	if redisLastAccess, err := cache.GetLastAccessTime(shortCode); err == nil && redisLastAccess != nil {
		// Use Redis timestamp if it's newer than DB
		if lastAccessedAt == nil || redisLastAccess.After(*lastAccessedAt) {
			lastAccessedAt = redisLastAccess
		}
	}

	return &models.URLDetailsResponse{
		ShortCode:      urlRecord.ShortCode,
		ShortURL:       s.baseURL + "/" + urlRecord.ShortCode,
		OriginalURL:    urlRecord.OriginalURL,
		ClickCount:     urlRecord.ClickCount,
		CreatedAt:      urlRecord.CreatedAt,
		LastAccessedAt: lastAccessedAt,
		IsActive:       urlRecord.IsActive,
		ExpiresAt:      urlRecord.ExpiresAt,
	}, nil
}

func (s *urlService) DeleteURL(ctx context.Context, shortCode string) error {
	// Check if URL exists
	urlRecord, err := s.repo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return fmt.Errorf("failed to fetch URL: %w", err)
	}
	if urlRecord == nil {
		return errors.New("URL_NOT_FOUND: Short URL not found")
	}

	// Soft delete by marking as inactive
	if err := s.repo.Delete(ctx, shortCode); err != nil {
		return fmt.Errorf("failed to delete URL: %w", err)
	}

	// Invalidate cache
	go cache.DeleteCachedURL(shortCode)

	return nil
}


// Helper methods

func (s *urlService) validateURL(rawURL string) error {
	if rawURL == "" {
		return errors.New("INVALID_URL: URL is required")
	}

	if !validURLRegex.MatchString(rawURL) {
		return errors.New("INVALID_URL: URL must start with http:// or https://")
	}

	if len(rawURL) > 2048 {
		return errors.New("URL_TOO_LONG: URL must be less than 2048 characters")
	}

	// Parse and validate URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil || parsedURL.Host == "" {
		return errors.New("INVALID_URL: Malformed URL")
	}

	return nil
}

// parseDuration converts human-readable duration strings to time
// Supports: "7d", "24h", "30m", "1w"
func (s *urlService) parseDuration(duration string) (time.Time, error) {
	if duration == "" {
		return time.Time{}, errors.New("empty duration")
	}
	
	// Get numeric part and unit
	var value int
	var unit string
	
	if _, err := fmt.Sscanf(duration, "%d%s", &value, &unit); err != nil {
		return time.Time{}, errors.New("invalid duration format")
	}
	
	var d time.Duration
	switch unit {
	case "m", "min", "mins", "minute", "minutes":
		d = time.Duration(value) * time.Minute
	case "h", "hr", "hrs", "hour", "hours":
		d = time.Duration(value) * time.Hour
	case "d", "day", "days":
		d = time.Duration(value) * 24 * time.Hour
	case "w", "week", "weeks":
		d = time.Duration(value) * 7 * 24 * time.Hour
	case "M", "month", "months":
		d = time.Duration(value) * 30 * 24 * time.Hour
	case "y", "year", "years":
		d = time.Duration(value) * 365 * 24 * time.Hour
	default:
		return time.Time{}, errors.New("unsupported duration unit")
	}
	
	return time.Now().Add(d), nil
}

func (s *urlService) validateAlias(alias string) error {
	if len(alias) < 4 || len(alias) > 20 {
		return errors.New("INVALID_ALIAS: Alias must be 4-20 characters")
	}

	if !validAliasRegex.MatchString(alias) {
		return errors.New("INVALID_ALIAS: Alias can only contain letters, numbers, hyphens, and underscores")
	}

	if reservedWords[alias] {
		return errors.New("ALIAS_RESERVED: This alias is reserved")
	}

	return nil
}

