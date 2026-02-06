package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// SecurityConfig holds all security-related configuration
type SecurityConfig struct {
	// Rate limiting
	RateLimiting RateLimitingConfig
	
	// Security headers
	SecurityHeaders SecurityHeadersConfig
	
	// CORS
	CORS CORSSecurityConfig
	
	// Brute force protection
	BruteForce BruteForceConfig
	
	// Error handling
	ErrorHandling ErrorConfig
	
	// Environment (development, production, staging)
	Environment string
}

// RateLimitingConfig holds rate limiting configuration
type RateLimitingConfig struct {
	Enabled bool
	// Default rate limiter
	Default RateLimiterConfig
	// Endpoint-specific limiters
	Auth     RateLimiterConfig
	Public   RateLimiterConfig
	API      RateLimiterConfig
	Admin    RateLimiterConfig
	Analytics RateLimiterConfig
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		RateLimiting: RateLimitingConfig{
			Enabled: true,
			Default: DefaultRateLimiterConfig(),
			Auth: RateLimiterConfig{
				Rate:            rate.Every(12 * time.Second), // 5 requests per minute
				Burst:           5,
				CleanupInterval: 5 * time.Minute,
				EntryTTL:        15 * time.Minute,
			},
			Public: RateLimiterConfig{
				Rate:            rate.Every(600 * time.Millisecond), // 100 requests per minute
				Burst:           20,
				CleanupInterval: 5 * time.Minute,
				EntryTTL:        10 * time.Minute,
			},
			API: RateLimiterConfig{
				Rate:            rate.Every(60 * time.Millisecond), // 1000 requests per minute
				Burst:           50,
				CleanupInterval: 5 * time.Minute,
				EntryTTL:        5 * time.Minute,
			},
			Admin: RateLimiterConfig{
				Rate:            rate.Every(30 * time.Second), // 2 requests per minute
				Burst:           3,
				CleanupInterval: 5 * time.Minute,
				EntryTTL:        30 * time.Minute,
			},
			Analytics: RateLimiterConfig{
				Rate:            rate.Every(100 * time.Millisecond), // 600 requests per minute
				Burst:           50,
				CleanupInterval: 5 * time.Minute,
				EntryTTL:        5 * time.Minute,
			},
		},
		SecurityHeaders: DefaultSecurityHeadersConfig(),
		CORS:           DefaultCORSSecurityConfig(),
		BruteForce:     DefaultBruteForceConfig(),
		ErrorHandling:  DefaultErrorConfig(),
		Environment:    "development",
	}
}

// ProductionSecurityConfig returns production security configuration
func ProductionSecurityConfig(allowedOrigins []string) SecurityConfig {
	config := DefaultSecurityConfig()
	config.Environment = "production"
	config.SecurityHeaders = DefaultSecurityHeadersConfig()
	config.CORS = ProductionCORSSecurityConfig(allowedOrigins)
	config.ErrorHandling = DefaultErrorConfig()
	return config
}

// SecurityMiddleware integrates all security middleware
type SecurityMiddleware struct {
	config           SecurityConfig
	rateLimiter      *MultiRateLimiter
	bruteForceProtector *BruteForceProtector
	validator        *ValidationMiddleware
	errorHandler     *ErrorHandler
	logger           *zap.Logger
}

// NewSecurityMiddleware creates a new integrated security middleware
func NewSecurityMiddleware(logger *zap.Logger, config SecurityConfig) *SecurityMiddleware {
	sm := &SecurityMiddleware{
		config:   config,
		logger:   logger,
		validator: NewValidationMiddleware(),
	}
	
	// Initialize rate limiter
	if config.RateLimiting.Enabled {
		sm.rateLimiter = NewMultiRateLimiter()
	}
	
	// Initialize brute force protector
	sm.bruteForceProtector = NewBruteForceProtector(config.BruteForce)
	
	// Initialize error handler
	sm.errorHandler = NewErrorHandler(logger, config.ErrorHandling)
	
	return sm
}

// SetupMiddleware applies all security middleware to the Gin router
func (sm *SecurityMiddleware) SetupMiddleware(router *gin.Engine) {
	// 1. Security headers (first)
	router.Use(SecurityHeaders(sm.config.SecurityHeaders))
	
	// 2. CORS
	router.Use(CORSSecurityMiddleware(sm.config.CORS))
	
	// 3. Input sanitization
	router.Use(sm.validator.SanitizeInput())
	
	// 4. Rate limiting
	if sm.config.RateLimiting.Enabled && sm.rateLimiter != nil {
		router.Use(sm.rateLimiter.Middleware())
	}
	
	// 5. Brute force protection
	router.Use(sm.bruteForceProtector.Middleware())
	
	// 6. Error handling (last, to catch all errors)
	router.Use(sm.errorHandler.Middleware())
}

// GetRateLimiter returns the rate limiter instance
func (sm *SecurityMiddleware) GetRateLimiter() *MultiRateLimiter {
	return sm.rateLimiter
}

// GetBruteForceProtector returns the brute force protector instance
func (sm *SecurityMiddleware) GetBruteForceProtector() *BruteForceProtector {
	return sm.bruteForceProtector
}

// GetValidator returns the validator instance
func (sm *SecurityMiddleware) GetValidator() *ValidationMiddleware {
	return sm.validator
}

// GetErrorHandler returns the error handler instance
func (sm *SecurityMiddleware) GetErrorHandler() *ErrorHandler {
	return sm.errorHandler
}

// SecurityMiddlewareBuilder provides a fluent interface for building security middleware
type SecurityMiddlewareBuilder struct {
	config SecurityConfig
	logger *zap.Logger
}

// NewSecurityMiddlewareBuilder creates a new security middleware builder
func NewSecurityMiddlewareBuilder(logger *zap.Logger) *SecurityMiddlewareBuilder {
	return &SecurityMiddlewareBuilder{
		config: DefaultSecurityConfig(),
		logger: logger,
	}
}

// WithEnvironment sets the environment
func (smb *SecurityMiddlewareBuilder) WithEnvironment(env string) *SecurityMiddlewareBuilder {
	smb.config.Environment = env
	if env == "production" {
		smb.config.SecurityHeaders = DefaultSecurityHeadersConfig()
		smb.config.ErrorHandling = DefaultErrorConfig()
	} else {
		smb.config.SecurityHeaders = DevelopmentSecurityHeadersConfig()
		smb.config.ErrorHandling = DevelopmentErrorConfig()
	}
	return smb
}

// WithCORSOrigins sets allowed CORS origins
func (smb *SecurityMiddlewareBuilder) WithCORSOrigins(origins []string) *SecurityMiddlewareBuilder {
	smb.config.CORS.AllowedOrigins = origins
	return smb
}

// WithRateLimiting configures rate limiting
func (smb *SecurityMiddlewareBuilder) WithRateLimiting(enabled bool) *SecurityMiddlewareBuilder {
	smb.config.RateLimiting.Enabled = enabled
	return smb
}

// WithBruteForceProtection configures brute force protection
func (smb *SecurityMiddlewareBuilder) WithBruteForceProtection(maxAttempts int, window, blockDuration time.Duration) *SecurityMiddlewareBuilder {
	smb.config.BruteForce.MaxAttempts = maxAttempts
	smb.config.BruteForce.AttemptWindow = window
	smb.config.BruteForce.BlockDuration = blockDuration
	return smb
}

// WithCustomSecurityHeaders sets custom security headers
func (smb *SecurityMiddlewareBuilder) WithCustomSecurityHeaders(headers map[string]string) *SecurityMiddlewareBuilder {
	smb.config.SecurityHeaders.CustomHeaders = headers
	return smb
}

// Build creates the security middleware
func (smb *SecurityMiddlewareBuilder) Build() *SecurityMiddleware {
	return NewSecurityMiddleware(smb.logger, smb.config)
}

// ApplySecurityDefaults applies sensible security defaults to a router
func ApplySecurityDefaults(router *gin.Engine, logger *zap.Logger, environment string, allowedOrigins []string) {
	var config SecurityConfig
	
	if environment == "production" {
		config = ProductionSecurityConfig(allowedOrigins)
	} else {
		config = DefaultSecurityConfig()
		config.Environment = environment
		config.SecurityHeaders = DevelopmentSecurityHeadersConfig()
		config.ErrorHandling = DevelopmentErrorConfig()
		if len(allowedOrigins) > 0 {
			config.CORS.AllowedOrigins = allowedOrigins
		}
	}
	
	securityMiddleware := NewSecurityMiddleware(logger, config)
	securityMiddleware.SetupMiddleware(router)
	
	logger.Info("Security middleware applied",
		zap.String("environment", environment),
		zap.Bool("rate_limiting_enabled", config.RateLimiting.Enabled),
		zap.Strings("cors_origins", config.CORS.AllowedOrigins),
	)
}