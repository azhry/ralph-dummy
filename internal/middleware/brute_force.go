package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// BruteForceConfig holds configuration for brute force protection
type BruteForceConfig struct {
	// Maximum number of attempts allowed
	MaxAttempts int
	// Time window for attempts (e.g., 15 minutes)
	AttemptWindow time.Duration
	// Duration to block the identifier (e.g., 1 hour)
	BlockDuration time.Duration
	// Cleanup interval for old entries
	CleanupInterval time.Duration
	// Whether to track by IP, email, or both
	TrackByIP    bool
	TrackByEmail bool
}

// DefaultBruteForceConfig returns default configuration
func DefaultBruteForceConfig() BruteForceConfig {
	return BruteForceConfig{
		MaxAttempts:     5,
		AttemptWindow:   15 * time.Minute,
		BlockDuration:   1 * time.Hour,
		CleanupInterval: 5 * time.Minute,
		TrackByIP:       true,
		TrackByEmail:    true,
	}
}

// LoginAttempt tracks login attempts for an identifier
type LoginAttempt struct {
	Count      int
	FirstSeen  time.Time
	LastSeen   time.Time
	Blocked    bool
	BlockedUntil time.Time
}

// BruteForceProtector protects against brute force attacks
type BruteForceProtector struct {
	attempts map[string]*LoginAttempt
	mu       sync.RWMutex
	config   BruteForceConfig
	lastCleanup time.Time
}

// NewBruteForceProtector creates a new brute force protector
func NewBruteForceProtector(config BruteForceConfig) *BruteForceProtector {
	bfp := &BruteForceProtector{
		attempts: make(map[string]*LoginAttempt),
		config:   config,
		lastCleanup: time.Now(),
	}
	
	// Start cleanup goroutine
	go bfp.cleanup()
	
	return bfp
}

// Middleware returns the Gin middleware for brute force protection
func (bfp *BruteForceProtector) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to authentication endpoints
		if !bfp.shouldApply(c) {
			c.Next()
			return
		}
		
		// Get identifiers to track
		identifiers := bfp.getIdentifiers(c)
		
		// Check if any identifier is blocked
		for _, identifier := range identifiers {
			if bfp.isBlocked(identifier) {
				bfp.blockedResponse(c)
				return
			}
		}
		
		c.Next()
		
		// Record failed attempt if response status indicates authentication failure
		if bfp.isAuthFailure(c) {
			for _, identifier := range identifiers {
				bfp.recordFailure(identifier)
			}
		} else if bfp.isAuthSuccess(c) {
			// Clear successful attempts
			for _, identifier := range identifiers {
				bfp.clearAttempts(identifier)
			}
		}
	}
}

// shouldApply determines if brute force protection should be applied to this request
func (bfp *BruteForceProtector) shouldApply(c *gin.Context) bool {
	path := c.Request.URL.Path
	method := c.Request.Method
	
	// Apply to authentication endpoints
	authEndpoints := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/forgot-password",
		"/api/v1/auth/reset-password",
	}
	
	for _, endpoint := range authEndpoints {
		if path == endpoint && method == "POST" {
			return true
		}
	}
	
	return false
}

// getIdentifiers extracts identifiers to track from the request
func (bfp *BruteForceProtector) getIdentifiers(c *gin.Context) []string {
	var identifiers []string
	
	// Track by IP address
	if bfp.config.TrackByIP {
		identifiers = append(identifiers, "ip:"+c.ClientIP())
	}
	
	// Track by email (extract from request body)
	if bfp.config.TrackByEmail {
		if email := bfp.extractEmail(c); email != "" {
			identifiers = append(identifiers, "email:"+email)
		}
	}
	
	return identifiers
}

// extractEmail extracts email from request body
func (bfp *BruteForceProtector) extractEmail(c *gin.Context) string {
	// Try to get email from JSON body
	var body struct {
		Email string `json:"email"`
	}
	
	if err := c.ShouldBindJSON(&body); err == nil && body.Email != "" {
		return body.Email
	}
	
	// Try to get from form data
	if email := c.PostForm("email"); email != "" {
		return email
	}
	
	return ""
}

// isBlocked checks if an identifier is currently blocked
func (bfp *BruteForceProtector) isBlocked(identifier string) bool {
	bfp.mu.RLock()
	defer bfp.mu.RUnlock()
	
	attempt, exists := bfp.attempts[identifier]
	if !exists {
		return false
	}
	
	if attempt.Blocked {
		// Check if block has expired
		if time.Now().After(attempt.BlockedUntil) {
			return false
		}
		return true
	}
	
	return false
}

// recordFailure records a failed authentication attempt
func (bfp *BruteForceProtector) recordFailure(identifier string) {
	bfp.mu.Lock()
	defer bfp.mu.Unlock()
	
	now := time.Now()
	attempt, exists := bfp.attempts[identifier]
	
	if !exists {
		bfp.attempts[identifier] = &LoginAttempt{
			Count:     1,
			FirstSeen: now,
			LastSeen:  now,
		}
		return
	}
	
	// Reset if outside attempt window
	if now.Sub(attempt.FirstSeen) > bfp.config.AttemptWindow {
		attempt.Count = 1
		attempt.FirstSeen = now
		attempt.Blocked = false
	} else {
		attempt.Count++
		
		// Block if max attempts reached
		if attempt.Count >= bfp.config.MaxAttempts {
			attempt.Blocked = true
			attempt.BlockedUntil = now.Add(bfp.config.BlockDuration)
		}
	}
	
	attempt.LastSeen = now
}

// clearAttempts clears all attempts for an identifier (after successful auth)
func (bfp *BruteForceProtector) clearAttempts(identifier string) {
	bfp.mu.Lock()
	defer bfp.mu.Unlock()
	
	delete(bfp.attempts, identifier)
}

// isAuthFailure checks if the response indicates an authentication failure
func (bfp *BruteForceProtector) isAuthFailure(c *gin.Context) bool {
	status := c.Writer.Status()
	return status == http.StatusUnauthorized || status == http.StatusBadRequest || status == http.StatusForbidden
}

// isAuthSuccess checks if the response indicates successful authentication
func (bfp *BruteForceProtector) isAuthSuccess(c *gin.Context) bool {
	status := c.Writer.Status()
	return status == http.StatusOK || status == http.StatusCreated
}

// blockedResponse returns the response for blocked requests
func (bfp *BruteForceProtector) blockedResponse(c *gin.Context) {
	c.JSON(http.StatusTooManyRequests, gin.H{
		"success": false,
		"error": gin.H{
			"code":       "BRUTE_FORCE_PROTECTION",
			"message":    "Too many authentication attempts. Please try again later.",
			"retry_after": int(bfp.config.BlockDuration.Seconds()),
		},
	})
	c.Abort()
}

// cleanup removes old entries from the attempts map
func (bfp *BruteForceProtector) cleanup() {
	ticker := time.NewTicker(bfp.config.CleanupInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		bfp.mu.Lock()
		now := time.Now()
		
		for identifier, attempt := range bfp.attempts {
			// Remove entries that are older than the block duration * 2
			// and not currently blocked
			if !attempt.Blocked && now.Sub(attempt.LastSeen) > bfp.config.BlockDuration*2 {
				delete(bfp.attempts, identifier)
			}
			
			// Unblock expired blocks
			if attempt.Blocked && now.After(attempt.BlockedUntil) {
				delete(bfp.attempts, identifier)
			}
		}
		
		bfp.lastCleanup = now
		bfp.mu.Unlock()
	}
}

// GetAttempts returns information about attempts for an identifier
func (bfp *BruteForceProtector) GetAttempts(identifier string) *LoginAttempt {
	bfp.mu.RLock()
	defer bfp.mu.RUnlock()
	
	if attempt, exists := bfp.attempts[identifier]; exists {
		// Return a copy to avoid concurrent modification
		return &LoginAttempt{
			Count:        attempt.Count,
			FirstSeen:    attempt.FirstSeen,
			LastSeen:     attempt.LastSeen,
			Blocked:      attempt.Blocked,
			BlockedUntil: attempt.BlockedUntil,
		}
	}
	
	return nil
}

// IsBlocked checks if a specific identifier is blocked
func (bfp *BruteForceProtector) IsBlocked(identifier string) bool {
	return bfp.isBlocked(identifier)
}

// Unblock manually unblocks an identifier
func (bfp *BruteForceProtector) Unblock(identifier string) {
	bfp.mu.Lock()
	defer bfp.mu.Unlock()
	
	delete(bfp.attempts, identifier)
}