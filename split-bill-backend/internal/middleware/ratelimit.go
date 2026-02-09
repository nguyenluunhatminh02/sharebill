package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/splitbill/backend/internal/utils"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int           // requests per interval
	interval time.Duration // time interval for rate
	cleanup  time.Duration // cleanup interval for stale entries
}

type visitor struct {
	tokens    int
	lastVisit time.Time
}

// NewRateLimiter creates a new rate limiter
// rate: maximum requests per interval
// interval: time window for rate limiting
func NewRateLimiter(rate int, interval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		interval: interval,
		cleanup:  time.Minute * 5,
	}

	// Start cleanup goroutine
	go rl.cleanupVisitors()

	return rl
}

// Allow checks if a request from the given IP is allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &visitor{
			tokens:    rl.rate - 1,
			lastVisit: time.Now(),
		}
		return true
	}

	// Calculate tokens to add based on time passed
	elapsed := time.Since(v.lastVisit)
	tokensToAdd := int(elapsed / rl.interval * time.Duration(rl.rate))
	v.tokens += tokensToAdd
	if v.tokens > rl.rate {
		v.tokens = rl.rate
	}
	v.lastVisit = time.Now()

	if v.tokens > 0 {
		v.tokens--
		return true
	}

	return false
}

// cleanupVisitors removes stale visitor entries
func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(rl.cleanup)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastVisit) > rl.interval*10 {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware returns a Gin middleware that rate limits requests
// by IP address using token bucket algorithm
func RateLimitMiddleware(requestsPerSecond int) gin.HandlerFunc {
	limiter := NewRateLimiter(requestsPerSecond, time.Second)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !limiter.Allow(ip) {
			utils.RespondError(c, http.StatusTooManyRequests, "Rate limit exceeded. Please try again later.")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByUser returns a rate limiter middleware that limits by user ID
// for authenticated routes
func RateLimitByUser(requestsPerMinute int) gin.HandlerFunc {
	limiter := NewRateLimiter(requestsPerMinute, time.Minute)

	return func(c *gin.Context) {
		// Try to use user ID first, fall back to IP
		key := c.ClientIP()
		if userID, exists := c.Get("firebase_uid"); exists {
			key = "user:" + userID.(string)
		}

		if !limiter.Allow(key) {
			utils.RespondError(c, http.StatusTooManyRequests, "Rate limit exceeded. Please try again later.")
			c.Abort()
			return
		}

		c.Next()
	}
}

// StrictRateLimitMiddleware returns a stricter rate limiter for sensitive endpoints
// like login, OCR scan, etc.
func StrictRateLimitMiddleware() gin.HandlerFunc {
	// 10 requests per minute for sensitive operations
	return RateLimitMiddleware(10)
}
