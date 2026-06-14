package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"forest-management/config"

	"github.com/gin-gonic/gin"
)

type rateEntry struct {
	Count       int
	WindowStart time.Time
	LastSeen    time.Time
}

type memoryRateLimiter struct {
	mu      sync.Mutex
	entries map[string]rateEntry
	limit   int
	window  time.Duration
}

func newMemoryRateLimiter(limit int, window time.Duration) *memoryRateLimiter {
	limiter := &memoryRateLimiter{entries: make(map[string]rateEntry), limit: limit, window: window}
	go limiter.cleanup()
	return limiter
}

func (l *memoryRateLimiter) allow(key string) (bool, int) {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	entry := l.entries[key]
	if entry.WindowStart.IsZero() || now.Sub(entry.WindowStart) >= l.window {
		entry = rateEntry{Count: 0, WindowStart: now}
	}
	entry.Count++
	entry.LastSeen = now
	l.entries[key] = entry
	remaining := l.limit - entry.Count
	if remaining < 0 {
		remaining = 0
	}
	return entry.Count <= l.limit, remaining
}

func (l *memoryRateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-2 * l.window)
		l.mu.Lock()
		for key, entry := range l.entries {
			if entry.LastSeen.Before(cutoff) {
				delete(l.entries, key)
			}
		}
		l.mu.Unlock()
	}
}

var (
	globalLimiter *memoryRateLimiter
	loginLimiter  = newMemoryRateLimiter(10, 5*time.Minute)
)

func GlobalRateLimit() gin.HandlerFunc {
	if globalLimiter == nil {
		globalLimiter = newMemoryRateLimiter(config.AppConfig.RateLimitPerMinute, time.Minute)
	}
	return rateLimitMiddleware(globalLimiter, "Too many requests")
}

func LoginRateLimit() gin.HandlerFunc {
	return rateLimitMiddleware(loginLimiter, "Too many login attempts. Please wait and try again.")
}

func rateLimitMiddleware(limiter *memoryRateLimiter, message string) gin.HandlerFunc {
	return func(c *gin.Context) {
		allowed, remaining := limiter.allow(c.ClientIP())
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		if !allowed {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"success": false, "message": message, "error": "rate_limited"})
			return
		}
		c.Next()
	}
}
