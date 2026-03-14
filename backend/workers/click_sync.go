package workers

import (
	"log"
	"time"

	"github.com/yourusername/url-shortener/cache"
	"github.com/yourusername/url-shortener/database"
)

// StartClickCountSyncWorker starts a background worker to sync click counts
// from Redis to PostgreSQL periodically
func StartClickCountSyncWorker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Click count sync worker started (interval: %v)", interval)

	for range ticker.C {
		if err := syncClickCounts(); err != nil {
			log.Printf("Error syncing click counts: %v", err)
		}
	}
}

func syncClickCounts() error {
	// Get all click counters from Redis
	counters, err := cache.GetAllClickCounters()
	if err != nil {
		return err
	}

	if len(counters) == 0 {
		return nil
	}

	log.Printf("Syncing %d click counters to database", len(counters))

	// Update each URL's click count and last_accessed_at in the database
	shortCodes := make([]string, 0, len(counters))
	for shortCode, count := range counters {
		if count > 0 {
			// Get last access time from Redis
			lastAccessTime, err := cache.GetLastAccessTime(shortCode)
			if err != nil {
				log.Printf("Failed to get last access time for %s: %v", shortCode, err)
				continue
			}
			
			// If no timestamp in Redis, use current time
			if lastAccessTime == nil {
				now := time.Now()
				lastAccessTime = &now
			}
			
			// Update DB with count and actual last access time
			if err := database.UpdateClickCount(shortCode, count, *lastAccessTime); err != nil {
				log.Printf("Failed to update click count for %s: %v", shortCode, err)
				continue
			}
			shortCodes = append(shortCodes, shortCode)
		}
	}

	// Clear the Redis counters after successful sync
	if err := cache.FlushClickCounters(shortCodes); err != nil {
		log.Printf("Failed to flush click counters: %v", err)
		return err
	}

	log.Printf("Successfully synced %d click counters", len(counters))
	return nil
}
