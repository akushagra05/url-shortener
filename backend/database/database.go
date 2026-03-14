package database

import (
	"fmt"
	"log"
	"time"

	"github.com/yourusername/url-shortener/config"
	"github.com/yourusername/url-shortener/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase initializes the database connection
func InitDatabase(cfg *config.Config) error {
	var err error
	
	dsn := cfg.GetDSN()
	
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connection established successfully")
	return nil
}

// AutoMigrate runs database migrations
func AutoMigrate() error {
	err := DB.AutoMigrate(
		&models.URL{},
		&models.URLAnalytics{},
		&models.ReservedKeyword{},
	)
	
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// SeedReservedKeywords adds reserved keywords to the database
func SeedReservedKeywords() error {
	reservedWords := []string{
		"api", "admin", "dashboard", "health", "metrics",
		"login", "logout", "register", "docs", "swagger",
		"static", "assets", "public", "www", "app",
		"v1", "v2", "version", "status", "ping",
	}

	for _, word := range reservedWords {
		keyword := models.ReservedKeyword{
			Keyword:   word,
			CreatedAt: time.Now(),
		}
		
		// Insert if not exists
		result := DB.Where("keyword = ?", word).FirstOrCreate(&keyword)
		if result.Error != nil {
			return fmt.Errorf("failed to seed reserved keyword %s: %w", word, result.Error)
		}
	}

	log.Printf("Seeded %d reserved keywords", len(reservedWords))
	return nil
}

// CloseDatabase closes the database connection
func CloseDatabase() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// IsReservedKeyword checks if a short code is reserved
func IsReservedKeyword(shortCode string) bool {
	var count int64
	DB.Model(&models.ReservedKeyword{}).Where("keyword = ?", shortCode).Count(&count)
	return count > 0
}

// GetURL retrieves a URL by short code
func GetURL(shortCode string) (*models.URL, error) {
	var url models.URL
	result := DB.Where("short_code = ? AND is_active = ?", shortCode, true).First(&url)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	return &url, nil
}

// CreateURL creates a new URL entry
func CreateURL(url *models.URL) error {
	result := DB.Create(url)
	return result.Error
}

// UpdateClickCount increments the click count for a URL
func UpdateClickCount(shortCode string, count int64) error {
	return DB.Model(&models.URL{}).
		Where("short_code = ?", shortCode).
		UpdateColumn("click_count", gorm.Expr("click_count + ?", count)).
		UpdateColumn("last_accessed_at", time.Now()).
		Error
}

// DeleteURL soft deletes a URL
func DeleteURL(shortCode string) error {
	return DB.Where("short_code = ?", shortCode).Delete(&models.URL{}).Error
}

// CreateAnalytics creates an analytics entry
func CreateAnalytics(analytics *models.URLAnalytics) error {
	return DB.Create(analytics).Error
}

// GetAnalytics retrieves analytics for a short code within a date range
func GetAnalytics(shortCode string, days int) (*models.URLAnalyticsResponse, error) {
	var url models.URL
	if err := DB.Where("short_code = ?", shortCode).First(&url).Error; err != nil {
		return nil, err
	}

	startDate := time.Now().AddDate(0, 0, -days)

	// Get clicks by date
	var clicksByDate []models.DateClickCount
	DB.Model(&models.URLAnalytics{}).
		Select("DATE(accessed_at) as date, COUNT(*) as count").
		Where("short_code = ? AND accessed_at >= ?", shortCode, startDate).
		Group("DATE(accessed_at)").
		Order("date DESC").
		Scan(&clicksByDate)

	// Get top countries
	var topCountries []models.CountryClickCount
	DB.Model(&models.URLAnalytics{}).
		Select("country_code as country, COUNT(*) as count").
		Where("short_code = ? AND accessed_at >= ? AND country_code != ''", shortCode, startDate).
		Group("country_code").
		Order("count DESC").
		Limit(10).
		Scan(&topCountries)

	// Get top referrers
	var topReferrers []models.ReferrerClickCount
	DB.Model(&models.URLAnalytics{}).
		Select("referer as referrer, COUNT(*) as count").
		Where("short_code = ? AND accessed_at >= ? AND referer != ''", shortCode, startDate).
		Group("referer").
		Order("count DESC").
		Limit(10).
		Scan(&topReferrers)

	// Get device distribution
	var devices models.DeviceDistribution
	DB.Model(&models.URLAnalytics{}).
		Select("SUM(CASE WHEN device_type = 'mobile' THEN 1 ELSE 0 END) as mobile, "+
			"SUM(CASE WHEN device_type = 'desktop' THEN 1 ELSE 0 END) as desktop, "+
			"SUM(CASE WHEN device_type = 'tablet' THEN 1 ELSE 0 END) as tablet").
		Where("short_code = ? AND accessed_at >= ?", shortCode, startDate).
		Scan(&devices)

	return &models.URLAnalyticsResponse{
		ShortCode:   shortCode,
		TotalClicks: url.ClickCount,
		Analytics: models.AnalyticsData{
			ClicksByDate: clicksByDate,
			TopCountries: topCountries,
			TopReferrers: topReferrers,
			Devices:      devices,
		},
	}, nil
}
