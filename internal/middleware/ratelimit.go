// Package middleware provides HTTP middleware for the MCP server.
package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter provides rate limiting functionality.
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// RateLimiterConfig configures the rate limiter.
type RateLimiterConfig struct {
	RequestsPerMinute int
	BurstSize         int
}

// DefaultRateLimiterConfig returns default rate limiter configuration.
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerMinute: 60,
		BurstSize:         10,
	}
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	if config.RequestsPerMinute == 0 {
		config.RequestsPerMinute = 60
	}
	if config.BurstSize == 0 {
		config.BurstSize = 10
	}

	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(float64(config.RequestsPerMinute) / 60.0),
		burst:    config.BurstSize,
	}
}

// getLimiter gets or creates a rate limiter for a key.
func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[key]
	rl.mu.RUnlock()

	if exists {
		return limiter
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists = rl.limiters[key]; exists {
		return limiter
	}

	limiter = rate.NewLimiter(rl.rate, rl.burst)
	rl.limiters[key] = limiter
	return limiter
}

// Allow checks if a request is allowed.
func (rl *RateLimiter) Allow(key string) bool {
	limiter := rl.getLimiter(key)
	return limiter.Allow()
}

// RateLimitInfo contains rate limit information for headers.
type RateLimitInfo struct {
	Limit     int
	Remaining int
	Reset     time.Time
}

// GetInfo returns rate limit info for a key.
func (rl *RateLimiter) GetInfo(key string) RateLimitInfo {
	limiter := rl.getLimiter(key)
	tokens := int(limiter.Tokens())
	if tokens < 0 {
		tokens = 0
	}

	return RateLimitInfo{
		Limit:     rl.burst,
		Remaining: tokens,
		Reset:     time.Now().Add(time.Duration(float64(time.Second) / float64(rl.rate))),
	}
}

// Middleware returns an HTTP middleware that applies rate limiting.
func (rl *RateLimiter) Middleware(keyFunc func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)

			if !rl.Allow(key) {
				info := rl.GetInfo(key)

				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(info.Limit))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(info.Reset.Unix(), 10))
				w.Header().Set("Retry-After", "60")

				http.Error(w, `{"jsonrpc":"2.0","id":null,"error":{"code":-32000,"message":"Rate limit exceeded. Retry after 60 seconds."}}`, http.StatusTooManyRequests)
				return
			}

			// Add rate limit headers to successful responses
			info := rl.GetInfo(key)
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(info.Limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(info.Reset.Unix(), 10))

			next.ServeHTTP(w, r)
		})
	}
}

// Cleanup removes old limiters to prevent memory leaks.
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Reset all limiters periodically
	// In production, you'd want to track last access time
	if len(rl.limiters) > 10000 {
		rl.limiters = make(map[string]*rate.Limiter)
	}
}

// GetRateLimitKey extracts the rate limit key from a request.
func GetRateLimitKey(r *http.Request) string {
	// Prefer session ID if present
	if sessionID := r.Header.Get("Mcp-Session-Id"); sessionID != "" {
		return "session:" + sessionID
	}

	// Fall back to IP address
	return "ip:" + GetClientIP(r)
}

// GetClientIP extracts the client IP from a request.
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
