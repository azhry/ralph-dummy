package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityHeadersConfig holds configuration for security headers
type SecurityHeadersConfig struct {
	// Content Security Policy
	CSPEnabled bool
	CSPPolicy  string
	
	// HSTS (HTTP Strict Transport Security)
	HSTSEnabled         bool
	HSTSMaxAge          string
	HSTSIncludeSubDomains bool
	HSTSPreload         bool
	
	// Other security headers
	XFrameOptions         string
	XContentTypeOptions    string
	XSSProtection          string
	ReferrerPolicy         string
	PermissionsPolicy      string
	
	// Custom headers
	CustomHeaders map[string]string
}

// DefaultSecurityHeadersConfig returns default security configuration
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		CSPEnabled: true,
		CSPPolicy: strings.Join([]string{
			"default-src 'self'",
			"script-src 'self' 'unsafe-inline'", // Allow inline scripts for simplicity
			"style-src 'self' 'unsafe-inline'",
			"img-src 'self' data: https:",
			"font-src 'self'",
			"connect-src 'self'",
			"frame-ancestors 'none'",
			"base-uri 'self'",
			"form-action 'self'",
		}, "; "),
		
		HSTSEnabled:           true,
		HSTSMaxAge:            "31536000", // 1 year
		HSTSIncludeSubDomains: true,
		HSTSPreload:           true,
		
		XFrameOptions:      "DENY",
		XContentTypeOptions: "nosniff",
		XSSProtection:      "1; mode=block",
		ReferrerPolicy:     "strict-origin-when-cross-origin",
		PermissionsPolicy:  "geolocation=(), microphone=(), camera=(), payment=()",
		
		CustomHeaders: make(map[string]string),
	}
}

// DevelopmentSecurityHeadersConfig returns less strict configuration for development
func DevelopmentSecurityHeadersConfig() SecurityHeadersConfig {
	config := DefaultSecurityHeadersConfig()
	
	// More permissive CSP for development
	config.CSPPolicy = strings.Join([]string{
		"default-src 'self'",
		"script-src 'self' 'unsafe-inline' 'unsafe-eval'", // Allow eval for development
		"style-src 'self' 'unsafe-inline'",
		"img-src 'self' data: https: http:",
		"font-src 'self'",
		"connect-src 'self' ws: wss:",
		"frame-ancestors 'none'",
		"base-uri 'self'",
		"form-action 'self'",
	}, "; ")
	
	// Disable HSTS in development
	config.HSTSEnabled = false
	
	return config
}

// SecurityHeaders adds security headers to all responses
func SecurityHeaders(config SecurityHeadersConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Content Security Policy
		if config.CSPEnabled && config.CSPPolicy != "" {
			c.Header("Content-Security-Policy", config.CSPPolicy)
		}
		
		// HTTP Strict Transport Security
		if config.HSTSEnabled {
			hstsValue := "max-age=" + config.HSTSMaxAge
			if config.HSTSIncludeSubDomains {
				hstsValue += "; includeSubDomains"
			}
			if config.HSTSPreload {
				hstsValue += "; preload"
			}
			c.Header("Strict-Transport-Security", hstsValue)
		}
		
		// X-Frame-Options
		if config.XFrameOptions != "" {
			c.Header("X-Frame-Options", config.XFrameOptions)
		}
		
		// X-Content-Type-Options
		if config.XContentTypeOptions != "" {
			c.Header("X-Content-Type-Options", config.XContentTypeOptions)
		}
		
		// X-XSS-Protection
		if config.XSSProtection != "" {
			c.Header("X-XSS-Protection", config.XSSProtection)
		}
		
		// Referrer Policy
		if config.ReferrerPolicy != "" {
			c.Header("Referrer-Policy", config.ReferrerPolicy)
		}
		
		// Permissions Policy
		if config.PermissionsPolicy != "" {
			c.Header("Permissions-Policy", config.PermissionsPolicy)
		}
		
		// Additional security headers
		c.Header("X-Download-Options", "noopen")
		c.Header("X-Permitted-Cross-Domain-Policies", "none")
		c.Header("Cross-Origin-Embedder-Policy", "require-corp")
		c.Header("Cross-Origin-Opener-Policy", "same-origin")
		c.Header("Cross-Origin-Resource-Policy", "same-origin")
		
		// Remove server information
		c.Header("Server", "")
		
		// Custom headers
		for key, value := range config.CustomHeaders {
			c.Header(key, value)
		}
		
		c.Next()
	}
}

// CORSSecurityConfig holds CORS configuration
type CORSSecurityConfig struct {
	AllowedOrigins     []string
	AllowedMethods     []string
	AllowedHeaders     []string
	ExposedHeaders     []string
	AllowCredentials  bool
	MaxAge            int
	// Strict origin checking
	StrictOriginChecking bool
}

// DefaultCORSSecurityConfig returns default CORS configuration
func DefaultCORSSecurityConfig() CORSSecurityConfig {
	return CORSSecurityConfig{
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:8080"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID", "X-Client-Version"},
		ExposedHeaders: []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:          86400, // 24 hours
		StrictOriginChecking: true,
	}
}

// ProductionCORSSecurityConfig returns production CORS configuration
func ProductionCORSSecurityConfig(allowedOrigins []string) CORSSecurityConfig {
	return CORSSecurityConfig{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposedHeaders: []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:          86400,
		StrictOriginChecking: true,
	}
}

// CORSSecurityMiddleware provides secure CORS handling
func CORSSecurityMiddleware(config CORSSecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		
		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			// Check if origin is allowed
			if !isOriginAllowed(origin, config.AllowedOrigins, config.StrictOriginChecking) {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
			c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
			c.Header("Access-Control-Max-Age", string(rune(config.MaxAge)))
			
			if config.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
			
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		// Handle actual requests
		if isOriginAllowed(origin, config.AllowedOrigins, config.StrictOriginChecking) {
			c.Header("Access-Control-Allow-Origin", origin)
			
			if len(config.ExposedHeaders) > 0 {
				c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
			}
			
			if config.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
		}
		
		c.Next()
	}
}

// isOriginAllowed checks if the origin is allowed based on configuration
func isOriginAllowed(origin string, allowedOrigins []string, strictChecking bool) bool {
	if origin == "" {
		return true // Allow same-origin requests
	}
	
	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return true // Wildcard allows all origins
		}
		
		if allowed == origin {
			return true // Exact match
		}
		
		// Support wildcard subdomains like https://*.example.com
		if strictChecking && strings.HasPrefix(allowed, "*.") {
			domain := allowed[2:]
			if strings.HasSuffix(origin, domain) {
				// Check that the protocol matches
				protocol := "https://"
				if strings.HasPrefix(origin, protocol) && strings.HasPrefix(allowed, protocol) {
					return true
				}
			}
		}
	}
	
	return false
}