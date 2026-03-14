
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/yourusername/url-shortener/internal/models"
	"gorm.io/gorm"
)

type URLRepository interface {
	Create(ctx context.Context, url *models.URL) error
	FindByShortCode(ctx context.Context, shortCode string) (*models.URL, error)
	FindByOriginalURLAndUser(ctx context.Context, originalURL string, userID string) (*models.URL, error)
	IncrementClickCount(ctx context.Context, id uint) error
	GetAnalytics(ctx context.Context, shortCode string) (*models.URL, error)
	ExistsByShortCode(ctx context.Context, shortCode string) (bool, error)
	Delete(ctx context.Context, shortCode string) error
}

type urlRepository struct {
	db *gorm.DB
}

func NewURLRepository(db *gorm.DB) URLRepository {
	return &urlRepository{db: db}
}

func (r *urlRepository) Create(ctx context.Context, url *models.URL) error {
	return r.db.WithContext(ctx).Create(url).Error
}

func (r *urlRepository) FindByShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	var url models.URL
	err := r.db.WithContext(ctx).Where("short_code = ?", shortCode).First(&url).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &url, nil
}

func (r *urlRepository) FindByOriginalURLAndUser(ctx context.Context, originalURL string, userID string) (*models.URL, error) {
	var url models.URL
	query := r.db.WithContext(ctx).Where("original_url = ? AND is_active = ?", originalURL, true)
	
	// If userID is provided, filter by user
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	} else {
		// If no userID, look for URLs without user (public URLs)
		query = query.Where("user_id IS NULL OR user_id = ?", "")
	}
	
	err := query.First(&url).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &url, nil
}

func (r *urlRepository) IncrementClickCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&models.URL{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"click_count":      gorm.Expr("click_count + ?", 1),
			"last_accessed_at": time.Now(),
		}).Error
}

func (r *urlRepository) GetAnalytics(ctx context.Context, shortCode string) (*models.URL, error) {
	var url models.URL
	err := r.db.WithContext(ctx).
		Where("short_code = ?", shortCode).
		First(&url).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &url, nil
}

func (r *urlRepository) ExistsByShortCode(ctx context.Context, shortCode string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.URL{}).
		Where("short_code = ?", shortCode).
		Count(&count).Error
	return count > 0, err
}

func (r *urlRepository) Delete(ctx context.Context, shortCode string) error {
	return r.db.WithContext(ctx).
		Model(&models.URL{}).
		Where("short_code = ?", shortCode).
		Update("is_active", false).Error
}
