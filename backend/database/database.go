
package database

import (
	"fmt"
	"log"
	"time"

	"github.com/yourusername/url-shortener/config"
	"github.com/yourusername/url-shortener/internal/models"
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

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
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

// Note: Reserved keywords are managed in url_service.go as a hardcoded map
// for optimal performance. The reserved_keywords table exists for future
// extensibility if dynamic keyword management is needed.

// CloseDatabase closes the database connection
func CloseDatabase() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// UpdateClickCount updates the click count and last_accessed_at for a URL
// lastAccessTime should be passed from Redis (actual access time, not sync time)
func UpdateClickCount(shortCode string, count int64, lastAccessTime time.Time) error {
	return DB.Model(&models.URL{}).
		Where("short_code = ?", shortCode).
		Updates(map[string]interface{}{
			"click_count":      gorm.Expr("click_count + ?", count),
			"last_accessed_at": lastAccessTime,
		}).Error
}
