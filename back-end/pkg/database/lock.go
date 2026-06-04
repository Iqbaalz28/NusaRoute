package database

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisMutex provides a simple distributed lock using Redis SETNX.
type RedisMutex struct {
	client *redis.Client
	key    string
}

// NewRedisMutex creates a new distributed lock.
func NewRedisMutex(client *redis.Client, key string) *RedisMutex {
	return &RedisMutex{
		client: client,
		key:    key,
	}
}

// Acquire tries to acquire the lock. Returns true if successful, false otherwise.
func (m *RedisMutex) Acquire(ctx context.Context, ttl time.Duration) bool {
	if m.client == nil {
		// Fallback for testing or if redis is unavailable (fail-open or fail-closed depending on risk)
		// For cron jobs, we prefer fail-open (execute anyway) or maybe fail-closed? 
		// Actually, if Redis is down, running multiple cron jobs might cause dupes, but not running them causes data stall.
		// We'll fail-open for now.
		return true
	}
	success, err := m.client.SetNX(ctx, m.key, "locked", ttl).Result()
	if err != nil {
		return false
	}
	return success
}

// Release releases the lock. Note: For simple cron jobs, letting the TTL expire is often sufficient
// to prevent concurrent overlapping runs, but Release can be called to free it early.
func (m *RedisMutex) Release(ctx context.Context) error {
	if m.client == nil {
		return nil
	}
	return m.client.Del(ctx, m.key).Err()
}
