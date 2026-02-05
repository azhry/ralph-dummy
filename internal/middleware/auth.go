package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"wedding-invitation-backend/internal/services"
)

// AuthMiddleware creates a JWT authentication middleware
func AuthMiddleware(tokenService *services.TokenService, blacklistChecker BlacklistChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header or cookie
		tokenString := extractToken(c)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
			c.Abort()
			return
		}

		// Parse and validate token
		token, err := tokenService.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Check if token is blacklisted
		jti, exists := claims["jti"].(string)
		if exists {
			if isBlacklisted, err := blacklistChecker.IsBlacklisted(c, jti); err != nil || isBlacklisted {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
				c.Abort()
				return
			}
		}

		// Set user information in context
		userID, _ := claims["sub"].(string)
		role, _ := claims["role"].(string)
		deviceID, _ := claims["device_id"].(string)

		c.Set("userID", userID)
		c.Set("userRole", role)
		c.Set("deviceID", deviceID)
		c.Set("tokenJTI", jti)

		c.Next()
	}
}

// RequireRole creates a middleware that requires a specific role
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
			c.Abort()
			return
		}

		if userRole != requiredRole && userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin creates a middleware that requires admin role
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin")
}

// OptionalAuth creates an optional authentication middleware
func OptionalAuth(tokenService *services.TokenService, blacklistChecker BlacklistChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		if tokenString == "" {
			c.Next()
			return
		}

		token, err := tokenService.ParseToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.Next()
			return
		}

		// Check if token is blacklisted
		jti, exists := claims["jti"].(string)
		if exists {
			if isBlacklisted, err := blacklistChecker.IsBlacklisted(c, jti); err == nil && !isBlacklisted {
				userID, _ := claims["sub"].(string)
				role, _ := claims["role"].(string)
				deviceID, _ := claims["device_id"].(string)

				c.Set("userID", userID)
				c.Set("userRole", role)
				c.Set("deviceID", deviceID)
				c.Set("tokenJTI", jti)
			}
		}

		c.Next()
	}
}

// BlacklistChecker defines the interface for checking token blacklist status
type BlacklistChecker interface {
	IsBlacklisted(c *gin.Context, jti string) (bool, error)
}

// extractToken extracts JWT token from Authorization header or cookie
func extractToken(c *gin.Context) string {
	// Try Authorization header first
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// Try cookie
	token, err := c.Cookie("access_token")
	if err == nil {
		return token
	}

	return ""
}

// GetUserID retrieves the user ID from the context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return "", false
	}
	return userID.(string), true
}

// GetUserRole retrieves the user role from the context
func GetUserRole(c *gin.Context) (string, bool) {
	userRole, exists := c.Get("userRole")
	if !exists {
		return "", false
	}
	return userRole.(string), true
}

// GetDeviceID retrieves the device ID from the context
func GetDeviceID(c *gin.Context) (string, bool) {
	deviceID, exists := c.Get("deviceID")
	if !exists {
		return "", false
	}
	return deviceID.(string), true
}

// IsAuthenticated checks if the user is authenticated
func IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get("userID")
	return exists
}

// IsAdmin checks if the user has admin role
func IsAdmin(c *gin.Context) bool {
	role, exists := GetUserRole(c)
	return exists && role == "admin"
}
