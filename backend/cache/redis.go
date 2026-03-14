
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/yourusername/url-shortener/config"
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
	ClickCounterKeyPrefix = "clicks:"
	RateLimitKeyPrefix    = "ratelimit:"
)

// IncrementClickCount increments the click counter and updates last access time in Redis
func IncrementClickCount(shortCode string) error {
	key := ClickCounterKeyPrefix + shortCode
	lastAccessKey := "last_access:" + shortCode
	
	// Pipeline for atomic operations
	pipe := Client.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Set(ctx, lastAccessKey, time.Now().Unix(), 0)  // Store timestamp
	_, err := pipe.Exec(ctx)
	
	return err
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

// GetLastAccessTime retrieves the last access timestamp from Redis
func GetLastAccessTime(shortCode string) (*time.Time, error) {
	key := "last_access:" + shortCode
	
	val, err := Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // No access recorded
	}
	if err != nil {
		return nil, err
	}
	
	var timestamp int64
	fmt.Sscanf(val, "%d", &timestamp)
	t := time.Unix(timestamp, 0)
	return &t, nil
}

// ResetClickCount resets the click counter to 0
func ResetClickCount(shortCode string) error {
	key := ClickCounterKeyPrefix + shortCode
	return Client.Set(ctx, key, 0, 0).Err()
}

// URL caching with TTL refresh on access
const (
	URLCacheKeyPrefix = "url:"
	URLCacheTTL       = 24 * time.Hour
)

// URLCacheData represents cached URL data
type URLCacheData struct {
	OriginalURL string     `json:"original_url"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsActive    bool       `json:"is_active"`
}

// SetCachedURL stores URL in Redis with 24hr TTL
func SetCachedURL(shortCode string, originalURL string, expiresAt *time.Time, isActive bool) error {
	key := URLCacheKeyPrefix + shortCode
	data := URLCacheData{
		OriginalURL: originalURL,
		ExpiresAt:   expiresAt,
		IsActive:    isActive,
	}
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	
	return Client.Set(ctx, key, jsonData, URLCacheTTL).Err()
}

// GetCachedURL retrieves URL from cache and refreshes TTL
func GetCachedURL(shortCode string) (*URLCacheData, error) {
	key := URLCacheKeyPrefix + shortCode
	
	val, err := Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, err
	}
	
	var data URLCacheData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, err
	}
	
	// Refresh TTL on access (keeps hot URLs cached)
	go Client.Expire(ctx, key, URLCacheTTL)
	
	return &data, nil
}

// DeleteCachedURL removes URL from cache
func DeleteCachedURL(shortCode string) error {
	key := URLCacheKeyPrefix + shortCode
	return Client.Del(ctx, key).Err()
}

// CheckRateLimit checks if rate limit is exceeded
func CheckRateLimit(key string, limit int, window int) (bool, error) {
	fullKey := RateLimitKeyPrefix + key
	
	count, err := Client.Incr(ctx, fullKey).Result()
	if err != nil {
		return false, err
	}
	
	if count == 1 {
		Client.Expire(ctx, fullKey, time.Duration(window)*time.Second)
	}
	
	return count <= int64(limit), nil
}

// GetAllClickCounters gets all click counters using SCAN (non-blocking)
func GetAllClickCounters() (map[string]int64, error) {
	pattern := ClickCounterKeyPrefix + "*"
	counters := make(map[string]int64)
	
	// Use SCAN instead of KEYS for non-blocking iteration
	var cursor uint64
	for {
		var keys []string
		var err error
		
		// SCAN returns cursor and keys in batches (non-blocking)
		keys, cursor, err = Client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}
		
		// Process this batch of keys
		for _, key := range keys {
			val, err := Client.Get(ctx, key).Result()
			if err != nil {
				continue
			}
			var count int64
			fmt.Sscanf(val, "%d", &count)
			shortCode := key[len(ClickCounterKeyPrefix):]
			counters[shortCode] = count
		}
		
		// cursor == 0 means iteration complete
		if cursor == 0 {
			break
		}
	}
	
	return counters, nil
}

// FlushClickCounters deletes all click counters
func FlushClickCounters(shortCodes []string) error {
	if len(shortCodes) == 0 {
		return nil
	}
	
	keys := make([]string, len(shortCodes))
	for i, code := range shortCodes {
		keys[i] = ClickCounterKeyPrefix + code
	}
	
	return Client.Del(ctx, keys...).Err()
}
