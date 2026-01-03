// Package middleware provides HTTP middleware for the MCP server.
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// jsonRPCError represents a JSON-RPC error response for rate limiting.
type jsonRPCError struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      any              `json:"id"`
	Error   *jsonRPCErrorObj `json:"error"`
}

type jsonRPCErrorObj struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// limiterEntry wraps a rate limiter with last access time for TTL cleanup.
type limiterEntry struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// RateLimiter provides rate limiting functionality.
type RateLimiter struct {
	limiters map[string]*limiterEntry
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
	ttl      time.Duration
}

// RateLimiterConfig configures the rate limiter.
type RateLimiterConfig struct {
	RequestsPerMinute int
	BurstSize         int
	TTL               time.Duration // How long to keep unused limiters
}

// DefaultRateLimiterConfig returns default rate limiter configuration.
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerMinute: 60,
		BurstSize:         10,
		TTL:               10 * time.Minute,
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
	if config.TTL == 0 {
		config.TTL = 10 * time.Minute
	}

	return &RateLimiter{
		limiters: make(map[string]*limiterEntry),
		rate:     rate.Limit(float64(config.RequestsPerMinute) / 60.0),
		burst:    config.BurstSize,
		ttl:      config.TTL,
	}
}

// getLimiter gets or creates a rate limiter for a key.
func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.RLock()
	entry, exists := rl.limiters[key]
	rl.mu.RUnlock()

	if exists {
		// Update last access time
		rl.mu.Lock()
		entry.lastAccess = time.Now()
		rl.mu.Unlock()
		return entry.limiter
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock
	if entry, exists = rl.limiters[key]; exists {
		entry.lastAccess = time.Now()
		return entry.limiter
	}

	limiter := rate.NewLimiter(rl.rate, rl.burst)
	rl.limiters[key] = &limiterEntry{
		limiter:    limiter,
		lastAccess: time.Now(),
	}
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

				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(info.Limit))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(info.Reset.Unix(), 10))
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)

				// Use proper JSON encoding for consistency
				json.NewEncoder(w).Encode(jsonRPCError{
					JSONRPC: "2.0",
					ID:      nil,
					Error: &jsonRPCErrorObj{
						Code:    -32000,
						Message: "Rate limit exceeded. Retry after 60 seconds.",
					},
				})
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

	now := time.Now()
	for key, entry := range rl.limiters {
		if now.Sub(entry.lastAccess) > rl.ttl {
			delete(rl.limiters, key)
		}
	}
}

// StartCleanup starts a background goroutine that periodically cleans up expired limiters.
func (rl *RateLimiter) StartCleanup(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				rl.Cleanup()
			}
		}
	}()
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
