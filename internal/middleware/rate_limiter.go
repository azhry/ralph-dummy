package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiterConfig holds configuration for rate limiter
type RateLimiterConfig struct {
	// Requests per second
	Rate rate.Limit
	// Maximum burst size
	Burst int
	// Cleanup interval for old entries
	CleanupInterval time.Duration
	// TTL for entries
	EntryTTL time.Duration
}

// DefaultRateLimiterConfig returns default configuration
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		Rate:            rate.Limit(10), // 10 requests per second
		Burst:           20,             // Allow burst of 20 requests
		CleanupInterval: 5 * time.Minute,
		EntryTTL:        10 * time.Minute,
	}
}

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	visitors map[string]*rate.Limiter
	mu       sync.RWMutex
	config   RateLimiterConfig
	lastCleanup time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		config:   config,
		lastCleanup: time.Now(),
	}
	
	// Start cleanup goroutine
	go rl.cleanup()
	
	return rl
}

// getLimiter returns a rate limiter for the given key
func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	limiter, exists := rl.visitors[key]
	if !exists {
		limiter = rate.NewLimiter(rl.config.Rate, rl.config.Burst)
		rl.visitors[key] = limiter
	}
	
	return limiter
}

// Middleware returns the Gin middleware for rate limiting
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client identifier
		key := rl.getClientKey(c)
		
		limiter := rl.getLimiter(key)
		
		if !limiter.Allow() {
			rl.rateLimitExceeded(c)
			return
		}
		
		c.Next()
	}
}

// getClientKey generates a unique key for the client
func (rl *RateLimiter) getClientKey(c *gin.Context) string {
	// Use IP address as base key
	key := c.ClientIP()
	
	// If user is authenticated, include user ID for more specific limiting
	if userID, exists := c.Get("userID"); exists {
		key = key + "-" + userID.(string)
	}
	
	// Include endpoint for endpoint-specific limiting
	key = key + "-" + c.Request.URL.Path
	
	return key
}

// rateLimitExceeded handles rate limit exceeded responses
func (rl *RateLimiter) rateLimitExceeded(c *gin.Context) {
	c.JSON(http.StatusTooManyRequests, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "RATE_LIMIT_EXCEEDED",
			"message": "Too many requests. Please try again later.",
		},
	})
	c.Abort()
}

// cleanup removes old entries from the visitors map
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		
		for key := range rl.visitors {
			// Simple cleanup strategy - remove entries older than TTL
			// In a more sophisticated implementation, we'd track last access time
			if now.Sub(rl.lastCleanup) > rl.config.EntryTTL {
				delete(rl.visitors, key)
			}
		}
		
		rl.lastCleanup = now
		rl.mu.Unlock()
	}
}

// MultiRateLimiter manages multiple rate limiters for different endpoints
type MultiRateLimiter struct {
	limiters map[string]*RateLimiter
	defaultLimiter *RateLimiter
}

// NewMultiRateLimiter creates a new multi-rate limiter
func NewMultiRateLimiter() *MultiRateLimiter {
	mrl := &MultiRateLimiter{
		limiters: make(map[string]*RateLimiter),
	}
	
	// Setup default rate limiter
	mrl.defaultLimiter = NewRateLimiter(DefaultRateLimiterConfig())
	
	// Setup endpoint-specific rate limiters
	mrl.setupEndpointLimiters()
	
	return mrl
}

// setupEndpointLimiters configures rate limiters for different endpoint types
func (mrl *MultiRateLimiter) setupEndpointLimiters() {
	// Auth endpoints: stricter rate limiting
	authConfig := RateLimiterConfig{
		Rate:            rate.Every(12 * time.Second), // 5 requests per minute
		Burst:           5,
		CleanupInterval: 5 * time.Minute,
		EntryTTL:        15 * time.Minute,
	}
	mrl.limiters["/api/v1/auth"] = NewRateLimiter(authConfig)
	
	// Public endpoints: moderate rate limiting
	publicConfig := RateLimiterConfig{
		Rate:            rate.Every(600 * time.Millisecond), // 100 requests per minute
		Burst:           20,
		CleanupInterval: 5 * time.Minute,
		EntryTTL:        10 * time.Minute,
	}
	mrl.limiters["/api/v1/public"] = NewRateLimiter(publicConfig)
	
	// Analytics tracking: high rate limiting for tracking
	analyticsConfig := RateLimiterConfig{
		Rate:            rate.Every(100 * time.Millisecond), // 600 requests per minute
		Burst:           50,
		CleanupInterval: 5 * time.Minute,
		EntryTTL:        5 * time.Minute,
	}
	mrl.limiters["/api/v1/analytics"] = NewRateLimiter(analyticsConfig)
	
	// Admin endpoints: very strict rate limiting
	adminConfig := RateLimiterConfig{
		Rate:            rate.Every(30 * time.Second), // 2 requests per minute
		Burst:           3,
		CleanupInterval: 5 * time.Minute,
		EntryTTL:        30 * time.Minute,
	}
	mrl.limiters["/api/v1/admin"] = NewRateLimiter(adminConfig)
}

// Middleware returns the Gin middleware for multi-rate limiting
func (mrl *MultiRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		
		// Find matching rate limiter
		var limiter *RateLimiter
		for prefix, l := range mrl.limiters {
			if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
				limiter = l
				break
			}
		}
		
		// Use default limiter if no specific match
		if limiter == nil {
			limiter = mrl.defaultLimiter
		}
		
		// Apply rate limiting
		key := limiter.getClientKey(c)
		rateLimiter := limiter.getLimiter(key)
		
		if !rateLimiter.Allow() {
			limiter.rateLimitExceeded(c)
			return
		}
		
		c.Next()
	}
}

// GetLimiter returns a specific rate limiter by prefix
func (mrl *MultiRateLimiter) GetLimiter(prefix string) *RateLimiter {
	if limiter, exists := mrl.limiters[prefix]; exists {
		return limiter
	}
	return mrl.defaultLimiter
}