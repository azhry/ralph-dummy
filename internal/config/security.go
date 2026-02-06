package config

import (
	"time"

	"golang.org/x/time/rate"
)

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	// Rate limiting settings
	RateLimiting RateLimitingConfig `yaml:"rate_limiting" json:"rate_limiting"`
	
	// Security headers settings
	SecurityHeaders SecurityHeadersConfig `yaml:"security_headers" json:"security_headers"`
	
	// CORS settings
	CORS CORSConfig `yaml:"cors" json:"cors"`
	
	// Brute force protection settings
	BruteForce BruteForceConfig `yaml:"brute_force" json:"brute_force"`
	
	// File upload security settings
	FileUpload FileUploadSecurityConfig `yaml:"file_upload" json:"file_upload"`
	
	// JWT security settings
	JWT JWTSecurityConfig `yaml:"jwt" json:"jwt"`
}

// RateLimitingConfig holds rate limiting configuration
type RateLimitingConfig struct {
	// Enable rate limiting
	Enabled bool `yaml:"enabled" json:"enabled"`
	
	// Default rate limiter settings
	Default DefaultRateLimiterConfig `yaml:"default" json:"default"`
	
	// Endpoint-specific rate limiters
	Auth     EndpointRateLimiterConfig `yaml:"auth" json:"auth"`
	Public   EndpointRateLimiterConfig `yaml:"public" json:"public"`
	API      EndpointRateLimiterConfig `yaml:"api" json:"api"`
	Admin    EndpointRateLimiterConfig `yaml:"admin" json:"admin"`
	Analytics EndpointRateLimiterConfig `yaml:"analytics" json:"analytics"`
}

// DefaultRateLimiterConfig holds default rate limiter settings
type DefaultRateLimiterConfig struct {
	// Requests per second (as a duration between requests)
	RequestsPerSecond string `yaml:"requests_per_second" json:"requests_per_second"`
	// Maximum burst size
	Burst int `yaml:"burst" json:"burst"`
	// Cleanup interval for old entries
	CleanupInterval string `yaml:"cleanup_interval" json:"cleanup_interval"`
	// TTL for entries
	EntryTTL string `yaml:"entry_ttl" json:"entry_ttl"`
}

// EndpointRateLimiterConfig holds endpoint-specific rate limiter settings
type EndpointRateLimiterConfig struct {
	// Requests per second (as a duration between requests)
	RequestsPerSecond string `yaml:"requests_per_second" json:"requests_per_second"`
	// Maximum burst size
	Burst int `yaml:"burst" json:"burst"`
	// Cleanup interval for old entries
	CleanupInterval string `yaml:"cleanup_interval" json:"cleanup_interval"`
	// TTL for entries
	EntryTTL string `yaml:"entry_ttl" json:"entry_ttl"`
}

// SecurityHeadersConfig holds security headers configuration
type SecurityHeadersConfig struct {
	// Content Security Policy
	CSPEnabled bool   `yaml:"csp_enabled" json:"csp_enabled"`
	CSPPolicy  string `yaml:"csp_policy" json:"csp_policy"`
	
	// HSTS (HTTP Strict Transport Security)
	HSTSEnabled           bool   `yaml:"hsts_enabled" json:"hsts_enabled"`
	HSTSMaxAge            string `yaml:"hsts_max_age" json:"hsts_max_age"`
	HSTSIncludeSubDomains bool   `yaml:"hsts_include_sub_domains" json:"hsts_include_sub_domains"`
	HSTSPreload           bool   `yaml:"hsts_preload" json:"hsts_preload"`
	
	// Other security headers
	XFrameOptions      string `yaml:"x_frame_options" json:"x_frame_options"`
	XContentTypeOptions string `yaml:"x_content_type_options" json:"x_content_type_options"`
	XSSProtection      string `yaml:"xss_protection" json:"xss_protection"`
	ReferrerPolicy     string `yaml:"referrer_policy" json:"referrer_policy"`
	PermissionsPolicy  string `yaml:"permissions_policy" json:"permissions_policy"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	// Allowed origins
	AllowedOrigins []string `yaml:"allowed_origins" json:"allowed_origins"`
	// Allowed methods
	AllowedMethods []string `yaml:"allowed_methods" json:"allowed_methods"`
	// Allowed headers
	AllowedHeaders []string `yaml:"allowed_headers" json:"allowed_headers"`
	// Exposed headers
	ExposedHeaders []string `yaml:"exposed_headers" json:"exposed_headers"`
	// Allow credentials
	AllowCredentials bool `yaml:"allow_credentials" json:"allow_credentials"`
	// Max age
	MaxAge int `yaml:"max_age" json:"max_age"`
	// Strict origin checking
	StrictOriginChecking bool `yaml:"strict_origin_checking" json:"strict_origin_checking"`
}

// BruteForceConfig holds brute force protection configuration
type BruteForceConfig struct {
	// Maximum number of attempts allowed
	MaxAttempts int `yaml:"max_attempts" json:"max_attempts"`
	// Time window for attempts
	AttemptWindow string `yaml:"attempt_window" json:"attempt_window"`
	// Duration to block the identifier
	BlockDuration string `yaml:"block_duration" json:"block_duration"`
	// Cleanup interval for old entries
	CleanupInterval string `yaml:"cleanup_interval" json:"cleanup_interval"`
	// Whether to track by IP, email, or both
	TrackByIP    bool `yaml:"track_by_ip" json:"track_by_ip"`
	TrackByEmail bool `yaml:"track_by_email" json:"track_by_email"`
}

// FileUploadSecurityConfig holds file upload security configuration
type FileUploadSecurityConfig struct {
	// Maximum file size in bytes
	MaxFileSize int64 `yaml:"max_file_size" json:"max_file_size"`
	// Allowed MIME types
	AllowedMimeTypes []string `yaml:"allowed_mime_types" json:"allowed_mime_types"`
	// Allowed file extensions
	AllowedExtensions []string `yaml:"allowed_extensions" json:"allowed_extensions"`
	// Whether to validate file magic numbers
	ValidateMagicNumbers bool `yaml:"validate_magic_numbers" json:"validate_magic_numbers"`
	// Whether to scan for malware
	ScanForMalware bool `yaml:"scan_for_malware" json:"scan_for_malware"`
}

// JWTSecurityConfig holds JWT security configuration
type JWTSecurityConfig struct {
	// Token algorithm
	Algorithm string `yaml:"algorithm" json:"algorithm"`
	// Access token lifetime
	AccessTokenLifetime string `yaml:"access_token_lifetime" json:"access_token_lifetime"`
	// Refresh token lifetime
	RefreshTokenLifetime string `yaml:"refresh_token_lifetime" json:"refresh_token_lifetime"`
	// Whether to use blacklisting
	UseBlacklisting bool `yaml:"use_blacklisting" json:"use_blacklisting"`
	// Issuer
	Issuer string `yaml:"issuer" json:"issuer"`
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		RateLimiting: RateLimitingConfig{
			Enabled: true,
			Default: DefaultRateLimiterConfig{
				RequestsPerSecond: "10ms",  // 100 requests per second
				Burst:             20,
				CleanupInterval:   "5m",
				EntryTTL:          "10m",
			},
			Auth: EndpointRateLimiterConfig{
				RequestsPerSecond: "12s",  // 5 requests per minute
				Burst:             5,
				CleanupInterval:   "5m",
				EntryTTL:          "15m",
			},
			Public: EndpointRateLimiterConfig{
				RequestsPerSecond: "600ms", // 100 requests per minute
				Burst:             20,
				CleanupInterval:   "5m",
				EntryTTL:          "10m",
			},
			API: EndpointRateLimiterConfig{
				RequestsPerSecond: "60ms", // 1000 requests per minute
				Burst:             50,
				CleanupInterval:   "5m",
				EntryTTL:          "5m",
			},
			Admin: EndpointRateLimiterConfig{
				RequestsPerSecond: "30s", // 2 requests per minute
				Burst:             3,
				CleanupInterval:   "5m",
				EntryTTL:          "30m",
			},
			Analytics: EndpointRateLimiterConfig{
				RequestsPerSecond: "100ms", // 600 requests per minute
				Burst:             50,
				CleanupInterval:   "5m",
				EntryTTL:          "5m",
			},
		},
		SecurityHeaders: SecurityHeadersConfig{
			CSPEnabled:           true,
			CSPPolicy:            "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
			HSTSEnabled:          true,
			HSTSMaxAge:           "31536000",
			HSTSIncludeSubDomains: true,
			HSTSPreload:          true,
			XFrameOptions:        "DENY",
			XContentTypeOptions:  "nosniff",
			XSSProtection:        "1; mode=block",
			ReferrerPolicy:       "strict-origin-when-cross-origin",
			PermissionsPolicy:    "geolocation=(), microphone=(), camera=(), payment=()",
		},
		CORS: CORSConfig{
			AllowedOrigins:     []string{"http://localhost:3000", "http://localhost:8080"},
			AllowedMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
			ExposedHeaders:     []string{"Content-Length", "Content-Type"},
			AllowCredentials:   true,
			MaxAge:             86400,
			StrictOriginChecking: true,
		},
		BruteForce: BruteForceConfig{
			MaxAttempts:     5,
			AttemptWindow:   "15m",
			BlockDuration:   "1h",
			CleanupInterval: "5m",
			TrackByIP:       true,
			TrackByEmail:    true,
		},
		FileUpload: FileUploadSecurityConfig{
			MaxFileSize:         5242880, // 5MB
			AllowedMimeTypes:    []string{"image/jpeg", "image/png", "image/webp"},
			AllowedExtensions:   []string{".jpg", ".jpeg", ".png", ".webp"},
			ValidateMagicNumbers: true,
			ScanForMalware:      false,
		},
		JWT: JWTSecurityConfig{
			Algorithm:            "RS256",
			AccessTokenLifetime:  "15m",
			RefreshTokenLifetime: "7d",
			UseBlacklisting:      true,
			Issuer:               "wedding-invitation-api",
		},
	}
}

// ProductionSecurityConfig returns production security configuration
func ProductionSecurityConfig(allowedOrigins []string) SecurityConfig {
	config := DefaultSecurityConfig()
	
	// Stricter rate limiting in production
	config.RateLimiting.Default.RequestsPerSecond = "100ms" // 10 requests per second
	config.RateLimiting.Default.Burst = 10
	
	// Stricter CSP in production
	config.SecurityHeaders.CSPPolicy = "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"
	
	// Production CORS origins
	config.CORS.AllowedOrigins = allowedOrigins
	
	// Stricter brute force protection in production
	config.BruteForce.MaxAttempts = 3
	config.BruteForce.AttemptWindow = "10m"
	config.BruteForce.BlockDuration = "2h"
	
	// Enable malware scanning in production
	config.FileUpload.ScanForMalware = true
	
	return config
}

// ParseDuration parses a duration string and returns time.Duration
func ParseDuration(duration string) (time.Duration, error) {
	return time.ParseDuration(duration)
}

// ParseRate parses a rate string and returns rate.Limit
func ParseRate(rateStr string) (rate.Limit, error) {
	duration, err := time.ParseDuration(rateStr)
	if err != nil {
		return 0, err
	}
	return rate.Every(duration), nil
}