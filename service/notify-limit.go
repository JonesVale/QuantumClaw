package service

import (
	"sync"
	"time"
)

// NotificationRateLimiter limits the rate of notifications to prevent flooding.
type NotificationRateLimiter struct {
	mu       sync.Mutex
	cooldown map[string]time.Time
	interval time.Duration
}

var (
	defaultRateLimiter *NotificationRateLimiter
	rlOnce             sync.Once
)

// GetNotifyRateLimiter returns the global notification rate limiter instance.
func GetNotifyRateLimiter() *NotificationRateLimiter {
	rlOnce.Do(func() {
		defaultRateLimiter = &NotificationRateLimiter{
			cooldown: make(map[string]time.Time),
			interval: 5 * time.Minute,
		}
	})
	return defaultRateLimiter
}

// SetInterval sets the cooldown interval for the rate limiter.
func (rl *NotificationRateLimiter) SetInterval(d time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.interval = d
}

// Allow checks if a notification for the given key is allowed (not in cooldown).
func (rl *NotificationRateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	last, ok := rl.cooldown[key]
	now := time.Now()

	if ok && now.Before(last.Add(rl.interval)) {
		return false
	}

	rl.cooldown[key] = now
	return true
}

// Reset clears the cooldown for the given key.
func (rl *NotificationRateLimiter) Reset(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.cooldown, key)
}
