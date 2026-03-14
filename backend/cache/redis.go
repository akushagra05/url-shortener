package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/yourusername/url-shortener/config"
	"github.com/yourusername/url-shortener/models"
)

var (
	Client *redis.Client
	ctx    = context.Background()
)

// InitRedis initializes the Redis client
func InitRedis(cfg *config.Config) error {
	Client = redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test connection
	_, err := Client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Redis connection established successfully")
	return nil
}

// CloseRedis closes the Redis connection
func CloseRedis() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

// Cache key patterns
const (
	URLCacheKeyPrefix      = "url:"
	ClickCounterKeyPrefix  = "clicks:"
	RateLimitKeyPrefix     = "ratelimit:"
)

// GetCachedURL retrieves a URL from cache
func GetCachedURL(shortCode string) (*models.CachedURL, error) {
	key := URLCacheKeyPrefix + shortCode
	
	val, err := Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, err
	}

	var cachedURL models.CachedURL
	if err := json.Unmarshal([]byte(val), &cachedURL); err != nil {
		return nil, err
	}

	return &cachedURL, nil
}

// SetCachedURL stores a URL in cache
func SetCachedURL(shortCode string, cachedURL *models.CachedURL, ttl time.Duration) error {
	key := URLCacheKeyPrefix + shortCode
	
	data, err := json.Marshal(cachedURL)
	if err != nil {
		return err
	}

	return Client.Set(ctx, key, data, ttl).Err()
}

// DeleteCachedURL removes a URL from cache
func DeleteCachedURL(shortCode string) error {
	key := URLCacheKeyPrefix + shortCode
	return Client.Del(ctx, key).Err()
}

// IncrementClickCount increments the click counter in Redis
func IncrementClickCount(shortCode string) error {
	key := ClickCounterKeyPrefix + shortCode
	return Client.Incr(ctx, key).Err()
}

// GetClickCount retrieves the current click count from Redis
func GetClickCount(shortCode string) (int64, error) {
	key := ClickCounterKeyPrefix + shortCode
	
	val, err := Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	var count int64
	fmt.Sscanf(val, "%d", &count)
	return count, nil
}

// ResetClickCount resets the click counter to 0
func ResetClickCount(shortCode string) error {
	key := ClickCounterKeyPrefix + shortCode
	return Client.Set(ctx, key, 0, 0).Err()
}

// DeleteClickCount removes the click counter
func DeleteClickCount(shortCode string) error {
	key := ClickCounterKeyPrefix + shortCode
	return Client.Del(ctx, key).Err()
}

// CheckRateLimit checks if an IP has exceeded the rate limit
func CheckRateLimit(ip, endpoint string, limit int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("%s%s:%s", RateLimitKeyPrefix, ip, endpoint)
	
	// Get current count
	count, err := Client.Get(ctx, key).Int()
	if err == redis.Nil {
		// First request, set counter to 1 with expiration
		err = Client.Set(ctx, key, 1, window).Err()
		return true, err
	}
	if err != nil {
		return false, err
	}

	// Check if limit exceeded
	if count >= limit {
		return false, nil
	}

	// Increment counter
	Client.Incr(ctx, key)
	return true, nil
}

// GetRateLimitInfo returns current rate limit status
func GetRateLimitInfo(ip, endpoint string) (int, error) {
	key := fmt.Sprintf("%s%s:%s", RateLimitKeyPrefix, ip, endpoint)
	
	count, err := Client.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}

// CacheURLWithData is a helper to cache URL data from database model
func CacheURLWithData(shortCode string, url *models.URL, ttl time.Duration) error {
	cachedURL := &models.CachedURL{
		OriginalURL: url.OriginalURL,
		ClickCount:  url.ClickCount,
		ExpiresAt:   url.ExpiresAt,
		IsActive:    url.IsActive,
	}
	return SetCachedURL(shortCode, cachedURL, ttl)
}

// InvalidateURLCache removes all cache entries for a URL
func InvalidateURLCache(shortCode string) error {
	if err := DeleteCachedURL(shortCode); err != nil {
		return err
	}
	return DeleteClickCount(shortCode)
}

// GetAllClickCounters retrieves all click counters (for batch flushing)
func GetAllClickCounters() (map[string]int64, error) {
	pattern := ClickCounterKeyPrefix + "*"
	
	keys, err := Client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	counters := make(map[string]int64)
	
	for _, key := range keys {
		val, err := Client.Get(ctx, key).Result()
		if err != nil {
			continue
		}
		
		var count int64
		fmt.Sscanf(val, "%d", &count)
		
		// Extract short code from key
		shortCode := key[len(ClickCounterKeyPrefix):]
		counters[shortCode] = count
	}

	return counters, nil
}

// FlushClickCounters deletes all click counter keys
func FlushClickCounters() error {
	pattern := ClickCounterKeyPrefix + "*"
	
	keys, err := Client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return Client.Del(ctx, keys...).Err()
	}

	return nil
}

// Health check for Redis
func HealthCheck() error {
	_, err := Client.Ping(ctx).Result()
	return err
}

// GetCacheStats returns basic cache statistics
func GetCacheStats() (map[string]interface{}, error) {
	info, err := Client.Info(ctx, "stats").Result()
	if err != nil {
		return nil, err
	}

	dbSize, err := Client.DBSize(ctx).Result()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"db_size": dbSize,
		"info":    info,
	}

	return stats, nil
}
