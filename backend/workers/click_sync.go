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

	// Update each URL's click count in the database
	for shortCode, count := range counters {
		if count > 0 {
			if err := database.UpdateClickCount(shortCode, count); err != nil {
				log.Printf("Failed to update click count for %s: %v", shortCode, err)
				continue
			}
		}
	}

	// Clear the Redis counters after successful sync
	if err := cache.FlushClickCounters(); err != nil {
		log.Printf("Failed to flush click counters: %v", err)
		return err
	}

	log.Printf("Successfully synced %d click counters", len(counters))
	return nil
}
