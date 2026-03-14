
package models

import (
	"time"

	"gorm.io/gorm"
)

// URL represents a shortened URL entry
type URL struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	ShortCode      string         `gorm:"uniqueIndex;size:20;not null" json:"short_code"`
	OriginalURL    string         `gorm:"type:text;not null" json:"original_url"`
	CustomAlias    bool           `gorm:"default:false" json:"custom_alias"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
	ClickCount     int64          `gorm:"default:0" json:"click_count"`
	LastAccessedAt *time.Time     `json:"last_accessed_at,omitempty"`
	UserID         string         `gorm:"size:255" json:"user_id,omitempty"`
	IsActive       bool           `gorm:"default:true;index" json:"is_active"`
	Metadata       *string        `gorm:"type:jsonb" json:"metadata,omitempty"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for URL model
func (URL) TableName() string {
	return "urls"
}

// BeforeCreate hook to validate short code length
func (u *URL) BeforeCreate(tx *gorm.DB) error {
	if len(u.ShortCode) < 4 || len(u.ShortCode) > 20 {
		return gorm.ErrInvalidField
	}
	return nil
}

// IsExpired checks if the URL has expired
func (u *URL) IsExpired() bool {
	if u.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*u.ExpiresAt)
}

// URLAnalytics represents analytics data for URL access
type URLAnalytics struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	ShortCode   string     `gorm:"size:20;not null;index:idx_analytics_short_code_time" json:"short_code"`
	AccessedAt  time.Time  `gorm:"index:idx_analytics_short_code_time;index:idx_analytics_accessed_at" json:"accessed_at"`
	IPAddress   string     `gorm:"type:inet" json:"ip_address"`
	UserAgent   string     `gorm:"type:text" json:"user_agent"`
	Referer     string     `gorm:"type:text" json:"referer"`
	CountryCode string     `gorm:"size:2" json:"country_code"`
	City        string     `gorm:"size:100" json:"city"`
	DeviceType  string     `gorm:"size:20" json:"device_type"`
}

// TableName specifies the table name for URLAnalytics model
func (URLAnalytics) TableName() string {
	return "url_analytics"
}

// ReservedKeyword represents reserved short codes that cannot be used
type ReservedKeyword struct {
	Keyword   string    `gorm:"primaryKey;size:20" json:"keyword"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name for ReservedKeyword model
func (ReservedKeyword) TableName() string {
	return "reserved_keywords"
}

// CreateURLRequest represents the request body for creating a short URL
type CreateURLRequest struct {
	URL          string     `json:"url" binding:"required,url"`
	CustomAlias  string     `json:"custom_alias,omitempty" binding:"omitempty,min=4,max=20"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`     // Exact timestamp: "2024-12-31T23:59:59Z"
	ExpiresIn    string     `json:"expires_in,omitempty"`     // Duration: "7d", "24h", "30m"
}

// CreateURLResponse represents the response for URL creation
type CreateURLResponse struct {
	ShortCode   string     `json:"short_code"`
	ShortURL    string     `json:"short_url"`
	OriginalURL string     `json:"original_url"`
	CustomAlias bool       `json:"custom_alias"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// URLDetailsResponse represents detailed URL information
type URLDetailsResponse struct {
	ShortCode      string     `json:"short_code"`
	ShortURL       string     `json:"short_url"`
	OriginalURL    string     `json:"original_url"`
	ClickCount     int64      `json:"click_count"`
	CreatedAt      time.Time  `json:"created_at"`
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`
	IsActive       bool       `json:"is_active"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

// URLAnalyticsResponse represents aggregate analytics for a URL
type URLAnalyticsResponse struct {
	ShortCode    string        `json:"short_code"`
	OriginalURL  string        `json:"original_url"`
	TotalClicks  int64         `json:"total_clicks"`  // Frontend expects this field name
	CreatedAt    time.Time     `json:"created_at"`
	LastAccessed *time.Time    `json:"last_accessed,omitempty"`
	Analytics    AnalyticsData `json:"analytics"`
}

// AnalyticsData contains detailed analytics
type AnalyticsData struct {
	ClicksByDate []ClicksByDate `json:"clicks_by_date"`
	TopCountries []TopItem      `json:"top_countries"`
	TopReferrers []TopItem      `json:"top_referrers"`
	Devices      DeviceStats    `json:"devices"`
}

// ClicksByDate represents clicks per date
type ClicksByDate struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// TopItem represents a top country or referrer
type TopItem struct {
	Name  string `json:"country,omitempty"` // For countries
	URL   string `json:"referrer,omitempty"` // For referrers
	Count int64  `json:"count"`
}

// DeviceStats represents device type statistics
type DeviceStats struct {
	Mobile  int64 `json:"mobile"`
	Desktop int64 `json:"desktop"`
	Tablet  int64 `json:"tablet"`
}

// APIResponse is the standard API response wrapper
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError represents an API error
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}



// CachedURL represents a cached URL entry in Redis
type CachedURL struct {
	OriginalURL string    `json:"original_url"`
	ClickCount  int64     `json:"click_count"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsActive    bool      `json:"is_active"`
}
