# Security Hardening Guide

## Overview

This document provides comprehensive security guidelines for the Wedding Invitation backend. Security is implemented in layers (defense in depth) to protect user data, prevent unauthorized access, and ensure system integrity.

---

## Security Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    PERIMETER SECURITY                       │
│  • HTTPS/TLS 1.2+                                          │
│  • WAF / DDoS Protection                                     │
│  • Rate Limiting                                             │
└─────────────────────────────────────────────────────────────┘
                             │
┌─────────────────────────────────────────────────────────────┐
│                    APPLICATION SECURITY                     │
│  • Authentication (JWT)                                      │
│  • Input Validation                                          │
│  • Output Encoding                                           │
│  • CORS Policy                                               │
└─────────────────────────────────────────────────────────────┘
                             │
┌─────────────────────────────────────────────────────────────┐
│                    DATA SECURITY                            │
│  • Encryption at Rest                                        │
│  • Encryption in Transit                                     │
│  • Password Hashing                                          │
│  • Access Controls                                           │
└─────────────────────────────────────────────────────────────┘
```

---

## 1. Authentication Security

### 1.1 JWT Implementation

**Algorithm Selection:**
- Use **RS256** (RSA with SHA-256) for production
- Never use symmetric algorithms (HS256) in production
- Asymmetric keys allow key rotation without downtime

**Token Structure:**

```go
package utils

import (
    "crypto/rsa"
    "crypto/x509"
    "encoding/pem"
    "fmt"
    "time"
    
    "github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims
type Claims struct {
    UserID    string `json:"user_id"`
    Email     string `json:"email"`
    TokenType string `json:"token_type"` // access or refresh
    jwt.RegisteredClaims
}

// TokenService handles JWT operations
type TokenService struct {
    privateKey *rsa.PrivateKey
    publicKey  *rsa.PublicKey
    config     TokenConfig
}

type TokenConfig struct {
    AccessTokenExpiry  time.Duration
    RefreshTokenExpiry time.Duration
    Issuer             string
}

// NewTokenService creates a new token service
func NewTokenService(privateKeyPEM, publicKeyPEM string, config TokenConfig) (*TokenService, error) {
    privateKey, err := parsePrivateKey(privateKeyPEM)
    if err != nil {
        return nil, fmt.Errorf("failed to parse private key: %w", err)
    }
    
    publicKey, err := parsePublicKey(publicKeyPEM)
    if err != nil {
        return nil, fmt.Errorf("failed to parse public key: %w", err)
    }
    
    return &TokenService{
        privateKey: privateKey,
        publicKey:  publicKey,
        config:     config,
    }, nil
}

// GenerateAccessToken creates a new access token
func (s *TokenService) GenerateAccessToken(userID, email string) (string, error) {
    claims := Claims{
        UserID:    userID,
        Email:     email,
        TokenType: "access",
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.AccessTokenExpiry)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    s.config.Issuer,
            Subject:   userID,
            ID:        generateTokenID(),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    return token.SignedString(s.privateKey)
}

// GenerateRefreshToken creates a new refresh token
func (s *TokenService) GenerateRefreshToken(userID string) (string, error) {
    claims := Claims{
        UserID:    userID,
        TokenType: "refresh",
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.RefreshTokenExpiry)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    s.config.Issuer,
            Subject:   userID,
            ID:        generateTokenID(),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    return token.SignedString(s.privateKey)
}

// ValidateToken validates and parses a JWT
func (s *TokenService) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // Verify signing method
        if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return s.publicKey, nil
    })
    
    if err != nil {
        return nil, fmt.Errorf("invalid token: %w", err)
    }
    
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, fmt.Errorf("invalid token claims")
}

// Key parsing helpers
func parsePrivateKey(pemString string) (*rsa.PrivateKey, error) {
    block, _ := pem.Decode([]byte(pemString))
    if block == nil {
        return nil, fmt.Errorf("failed to parse PEM block")
    }
    
    key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
    if err != nil {
        return nil, err
    }
    
    return key, nil
}

func parsePublicKey(pemString string) (*rsa.PublicKey, error) {
    block, _ := pem.Decode([]byte(pemString))
    if block == nil {
        return nil, fmt.Errorf("failed to parse PEM block")
    }
    
    key, err := x509.ParsePKIXPublicKey(block.Bytes)
    if err != nil {
        return nil, err
    }
    
    rsaKey, ok := key.(*rsa.PublicKey)
    if !ok {
        return nil, fmt.Errorf("not an RSA public key")
    }
    
    return rsaKey, nil
}

func generateTokenID() string {
    // Generate unique token ID for blacklisting
    return fmt.Sprintf("%d-%s", time.Now().UnixNano(), randomString(16))
}

func randomString(n int) string {
    const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    b := make([]byte, n)
    for i := range b {
        b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
    }
    return string(b)
}
```

### 1.2 Password Security

**Hashing:**
- Use **bcrypt** with cost factor 12 (takes ~250ms to hash)
- Never store passwords in plain text or reversible encryption
- Use unique salt per password (bcrypt handles this automatically)

```go
package utils

import (
    "fmt"
    
    "golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
    if len(password) < 8 {
        return "", fmt.Errorf("password must be at least 8 characters")
    }
    
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
    if err != nil {
        return "", fmt.Errorf("failed to hash password: %w", err)
    }
    
    return string(bytes), nil
}

// VerifyPassword checks if password matches hash
func VerifyPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

// Password validation rules
func ValidatePassword(password string) error {
    if len(password) < 8 {
        return fmt.Errorf("password must be at least 8 characters")
    }
    
    if len(password) > 72 {
        return fmt.Errorf("password must be less than 72 characters")
    }
    
    var (
        hasUpper   bool
        hasLower   bool
        hasNumber  bool
        hasSpecial bool
    )
    
    for _, char := range password {
        switch {
        case char >= 'A' && char <= 'Z':
            hasUpper = true
        case char >= 'a' && char <= 'z':
            hasLower = true
        case char >= '0' && char <= '9':
            hasNumber = true
        case char >= '!' && char <= '/':
            hasSpecial = true
        case char >= ':' && char <= '@':
            hasSpecial = true
        case char >= '[' && char <= '`':
            hasSpecial = true
        case char >= '{' && char <= '~':
            hasSpecial = true
        }
    }
    
    if !hasUpper || !hasLower || !hasNumber {
        return fmt.Errorf("password must contain uppercase, lowercase, and number")
    }
    
    return nil
}
```

### 1.3 Brute Force Protection

```go
package middleware

import (
    "net/http"
    "sync"
    "time"
    
    "github.com/gin-gonic/gin"
)

// LoginAttempt tracks login attempts
type LoginAttempt struct {
    Count     int
    FirstSeen time.Time
    LastSeen  time.Time
    Blocked   bool
}

// BruteForceProtector protects against brute force attacks
type BruteForceProtector struct {
    attempts map[string]*LoginAttempt
    mu       sync.RWMutex
    maxAttempts int
    blockDuration time.Duration
    windowDuration time.Duration
}

func NewBruteForceProtector() *BruteForceProtector {
    bfp := &BruteForceProtector{
        attempts: make(map[string]*LoginAttempt),
        maxAttempts: 5,
        blockDuration: 15 * time.Minute,
        windowDuration: 1 * time.Minute,
    }
    
    // Cleanup old entries periodically
    go bfp.cleanup()
    
    return bfp
}

func (bfp *BruteForceProtector) Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Only apply to login endpoint
        if c.Request.URL.Path != "/api/v1/auth/login" {
            c.Next()
            return
        }
        
        identifier := c.ClientIP()
        
        if bfp.isBlocked(identifier) {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "success": false,
                "error": gin.H{
                    "code": "BRUTE_FORCE_PROTECTION",
                    "message": "Too many login attempts. Please try again later.",
                    "retry_after": 900, // 15 minutes in seconds
                },
            })
            c.Abort()
            return
        }
        
        c.Next()
        
        // Record failed attempt if response status is 401
        if c.Writer.Status() == http.StatusUnauthorized {
            bfp.recordFailure(identifier)
        }
    }
}

func (bfp *BruteForceProtector) isBlocked(identifier string) bool {
    bfp.mu.RLock()
    defer bfp.mu.RUnlock()
    
    attempt, exists := bfp.attempts[identifier]
    if !exists {
        return false
    }
    
    if attempt.Blocked {
        if time.Since(attempt.LastSeen) > bfp.blockDuration {
            return false // Block expired
        }
        return true
    }
    
    return false
}

func (bfp *BruteForceProtector) recordFailure(identifier string) {
    bfp.mu.Lock()
    defer bfp.mu.Unlock()
    
    attempt, exists := bfp.attempts[identifier]
    if !exists {
        bfp.attempts[identifier] = &LoginAttempt{
            Count:     1,
            FirstSeen: time.Now(),
            LastSeen:  time.Now(),
        }
        return
    }
    
    // Reset if outside window
    if time.Since(attempt.FirstSeen) > bfp.windowDuration {
        attempt.Count = 1
        attempt.FirstSeen = time.Now()
        attempt.Blocked = false
    } else {
        attempt.Count++
        if attempt.Count >= bfp.maxAttempts {
            attempt.Blocked = true
        }
    }
    
    attempt.LastSeen = time.Now()
}

func (bfp *BruteForceProtector) cleanup() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        bfp.mu.Lock()
        now := time.Now()
        for identifier, attempt := range bfp.attempts {
            if now.Sub(attempt.LastSeen) > bfp.blockDuration*2 {
                delete(bfp.attempts, identifier)
            }
        }
        bfp.mu.Unlock()
    }
}
```

---

## 2. Authorization

### 2.1 Resource Ownership Verification

```go
package middleware

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// WeddingOwnership middleware ensures user owns the wedding resource
func WeddingOwnership() gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, exists := c.Get("user_id")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "error": gin.H{
                    "code": "UNAUTHORIZED",
                    "message": "Authentication required",
                },
            })
            c.Abort()
            return
        }
        
        weddingID := c.Param("id")
        if weddingID == "" {
            c.Next()
            return
        }
        
        // Check ownership in database
        weddingService := c.MustGet("wedding_service").(WeddingService)
        
        objectID, err := primitive.ObjectIDFromHex(weddingID)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error": gin.H{
                    "code": "INVALID_ID",
                    "message": "Invalid wedding ID",
                },
            })
            c.Abort()
            return
        }
        
        wedding, err := weddingService.GetByID(c.Request.Context(), objectID)
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{
                "success": false,
                "error": gin.H{
                    "code": "WEDDING_NOT_FOUND",
                    "message": "Wedding not found",
                },
            })
            c.Abort()
            return
        }
        
        if wedding.UserID.Hex() != userID.(string) {
            c.JSON(http.StatusForbidden, gin.H{
                "success": false,
                "error": gin.H{
                    "code": "FORBIDDEN",
                    "message": "You don't have permission to access this resource",
                },
            })
            c.Abort()
            return
        }
        
        // Store wedding in context for handler use
        c.Set("wedding", wedding)
        c.Next()
    }
}
```

---

## 3. Input Validation

### 3.1 Request Validation

```go
package middleware

import (
    "net/http"
    "strings"
    
    "github.com/gin-gonic/gin"
    "github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
    validate = validator.New()
    
    // Custom validators
    validate.RegisterValidation("slug", validateSlug)
    validate.RegisterValidation("objectid", validateObjectID)
}

// ValidationMiddleware validates request body
func ValidationMiddleware(requestStruct interface{}) gin.HandlerFunc {
    return func(c *gin.Context) {
        if err := c.ShouldBindJSON(requestStruct); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error": gin.H{
                    "code":    "VALIDATION_ERROR",
                    "message": "Invalid request body",
                    "details":  formatValidationErrors(err),
                },
            })
            c.Abort()
            return
        }
        
        // Validate struct
        if err := validate.Struct(requestStruct); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error": gin.H{
                    "code":    "VALIDATION_ERROR",
                    "message": "Validation failed",
                    "details":  formatValidationErrors(err),
                },
            })
            c.Abort()
            return
        }
        
        c.Set("validated_request", requestStruct)
        c.Next()
    }
}

func validateSlug(fl validator.FieldLevel) bool {
    slug := fl.Field().String()
    
    // Slug rules: alphanumeric, hyphens only, 3-50 chars
    if len(slug) < 3 || len(slug) > 50 {
        return false
    }
    
    for _, char := range slug {
        if !((char >= 'a' && char <= 'z') || 
             (char >= '0' && char <= '9') || 
             char == '-') {
            return false
        }
    }
    
    // Cannot start or end with hyphen
    if strings.HasPrefix(slug, "-") || strings.HasSuffix(slug, "-") {
        return false
    }
    
    return true
}

func validateObjectID(fl validator.FieldLevel) bool {
    id := fl.Field().String()
    // MongoDB ObjectID is 24 hex characters
    if len(id) != 24 {
        return false
    }
    
    for _, char := range id {
        if !((char >= '0' && char <= '9') || 
             (char >= 'a' && char <= 'f')) {
            return false
        }
    }
    
    return true
}

func formatValidationErrors(err error) []gin.H {
    var errors []gin.H
    
    if validationErrors, ok := err.(validator.ValidationErrors); ok {
        for _, e := range validationErrors {
            errors = append(errors, gin.H{
                "field":   e.Field(),
                "message": getErrorMessage(e),
            })
        }
    } else {
        errors = append(errors, gin.H{
            "message": err.Error(),
        })
    }
    
    return errors
}

func getErrorMessage(e validator.FieldError) string {
    switch e.Tag() {
    case "required":
        return e.Field() + " is required"
    case "email":
        return "Invalid email format"
    case "min":
        return e.Field() + " must be at least " + e.Param() + " characters"
    case "max":
        return e.Field() + " must be at most " + e.Param() + " characters"
    case "slug":
        return "Invalid slug format (use lowercase letters, numbers, and hyphens only)"
    default:
        return e.Field() + " validation failed on " + e.Tag()
    }
}
```

### 3.2 SQL/NoSQL Injection Prevention

```go
package repository

import (
    "context"
    "fmt"
    
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
)

// Never build queries with string concatenation
// BAD: collection.Find(ctx, bson.M{"email": userInput})
// GOOD: Use parameterized queries with proper type checking

type WeddingRepository struct {
    collection *mongo.Collection
}

// GetBySlug safely retrieves a wedding by slug
func (r *WeddingRepository) GetBySlug(ctx context.Context, slug string) (*Wedding, error) {
    // Validate slug format first
    if !isValidSlug(slug) {
        return nil, fmt.Errorf("invalid slug format")
    }
    
    // Use bson.M with properly typed values
    filter := bson.M{"slug": slug}
    
    var wedding Wedding
    err := r.collection.FindOne(ctx, filter).Decode(&wedding)
    if err != nil {
        return nil, err
    }
    
    return &wedding, nil
}

// GetByID safely retrieves by ObjectID
func (r *WeddingRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*Wedding, error) {
    filter := bson.M{"_id": id}
    
    var wedding Wedding
    err := r.collection.FindOne(ctx, filter).Decode(&wedding)
    if err != nil {
        return nil, err
    }
    
    return &wedding, nil
}

// Search with proper input sanitization
func (r *WeddingRepository) Search(ctx context.Context, query string, userID primitive.ObjectID) ([]Wedding, error) {
    // Sanitize query - remove special regex characters
    sanitized := sanitizeRegexInput(query)
    
    // Use regex for text search, but with sanitized input
    filter := bson.M{
        "user_id": userID,
        "$or": []bson.M{
            {"groom_name": bson.M{"$regex": sanitized, "$options": "i"}},
            {"bride_name": bson.M{"$regex": sanitized, "$options": "i"}},
        },
    }
    
    cursor, err := r.collection.Find(ctx, filter)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)
    
    var weddings []Wedding
    if err = cursor.All(ctx, &weddings); err != nil {
        return nil, err
    }
    
    return weddings, nil
}

func isValidSlug(slug string) bool {
    if len(slug) < 3 || len(slug) > 50 {
        return false
    }
    
    for _, char := range slug {
        if !((char >= 'a' && char <= 'z') || 
             (char >= '0' && char <= '9') || 
             char == '-') {
            return false
        }
    }
    
    return true
}

func sanitizeRegexInput(input string) string {
    // Remove MongoDB regex special characters that could cause issues
    // Only allow alphanumeric and spaces for search
    var result []rune
    for _, char := range input {
        if (char >= 'a' && char <= 'z') || 
           (char >= 'A' && char <= 'Z') || 
           (char >= '0' && char <= '9') || 
           char == ' ' {
            result = append(result, char)
        }
    }
    
    return string(result)
}
```

---

## 4. Transport Security

### 4.1 TLS Configuration

```go
package main

import (
    "crypto/tls"
    "net/http"
    "time"
)

func createSecureServer() *http.Server {
    tlsConfig := &tls.Config{
        MinVersion: tls.VersionTLS12,
        CurvePreferences: []tls.CurveID{
            tls.X25519,
            tls.CurveP256,
        },
        CipherSuites: []uint16{
            tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
            tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
            tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
            tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
        },
        PreferServerCipherSuites: true,
    }
    
    return &http.Server{
        Addr:         ":443",
        TLSConfig:    tlsConfig,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
    }
}
```

### 4.2 Security Headers Middleware

```go
package middleware

import (
    "net/http"
    "strings"
    
    "github.com/gin-gonic/gin"
)

// SecurityHeaders adds security headers to all responses
func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Prevent MIME type sniffing
        c.Header("X-Content-Type-Options", "nosniff")
        
        // Prevent clickjacking
        c.Header("X-Frame-Options", "DENY")
        
        // XSS Protection (legacy browsers)
        c.Header("X-XSS-Protection", "1; mode=block")
        
        // Strict Transport Security (force HTTPS)
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
        
        // Content Security Policy
        csp := []string{
            "default-src 'self'",
            "script-src 'self'",
            "style-src 'self' 'unsafe-inline'",
            "img-src 'self' https://cdn.example.com data:",
            "font-src 'self'",
            "connect-src 'self' https://api.example.com",
            "frame-ancestors 'none'",
            "base-uri 'self'",
            "form-action 'self'",
        }
        c.Header("Content-Security-Policy", strings.Join(csp, "; "))
        
        // Referrer Policy
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        
        // Permissions Policy
        c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
        
        c.Next()
    }
}
```

---

## 5. API Security

### 5.1 Rate Limiting

```go
package middleware

import (
    "net/http"
    "sync"
    "time"
    
    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
)

// RateLimiter implements token bucket algorithm
type RateLimiter struct {
    visitors map[string]*rate.Limiter
    mu       sync.RWMutex
    rate     rate.Limit
    burst    int
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
    return &RateLimiter{
        visitors: make(map[string]*rate.Limiter),
        rate:     r,
        burst:    b,
    }
}

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    limiter, exists := rl.visitors[key]
    if !exists {
        limiter = rate.NewLimiter(rl.rate, rl.burst)
        rl.visitors[key] = limiter
    }
    
    return limiter
}

// RateLimit middleware
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get client identifier (IP + User ID if authenticated)
        key := c.ClientIP()
        if userID, exists := c.Get("user_id"); exists {
            key = key + "-" + userID.(string)
        }
        
        limiter := rl.getLimiter(key)
        
        if !limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "success": false,
                "error": gin.H{
                    "code":    "RATE_LIMIT_EXCEEDED",
                    "message": "Too many requests. Please try again later.",
                },
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// Different rate limits for different endpoints
func SetupRateLimits(router *gin.Engine) {
    // Auth endpoints: 5 requests per minute
    authLimiter := NewRateLimiter(rate.Every(12*time.Second), 5)
    
    // Public API: 100 requests per minute
    publicLimiter := NewRateLimiter(rate.Every(600*time.Millisecond), 100)
    
    // API endpoints: 1000 requests per minute
    apiLimiter := NewRateLimiter(rate.Every(60*time.Millisecond), 1000)
    
    router.Use(func(c *gin.Context) {
        path := c.Request.URL.Path
        
        if strings.HasPrefix(path, "/api/v1/auth") {
            authLimiter.Middleware()(c)
        } else if strings.HasPrefix(path, "/api/v1/public") {
            publicLimiter.Middleware()(c)
        } else {
            apiLimiter.Middleware()(c)
        }
    })
}
```

### 5.2 CORS Configuration

```go
package middleware

import (
    "net/http"
    "strings"
    
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)

func CORSConfig(allowedOrigins []string) gin.HandlerFunc {
    config := cors.Config{
        AllowOrigins:     allowedOrigins,
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
        ExposeHeaders:    []string{"Content-Length", "Content-Type"},
        AllowCredentials: true,
        MaxAge:           86400, // 24 hours
        AllowOriginFunc: func(origin string) bool {
            // Strict origin checking in production
            for _, allowed := range allowedOrigins {
                if origin == allowed || allowed == "*" {
                    return true
                }
                // Support wildcards like https://*.example.com
                if strings.HasPrefix(allowed, "*.") {
                    domain := allowed[2:]
                    if strings.HasSuffix(origin, domain) {
                        return true
                    }
                }
            }
            return false
        },
    }
    
    return cors.New(config)
}
```

---

## 6. File Upload Security

### 6.1 File Validation

```go
package utils

import (
    "bytes"
    "fmt"
    "io"
    "mime/multipart"
)

// FileValidator validates uploaded files
type FileValidator struct {
    MaxSize      int64                  // Max file size in bytes
    AllowedTypes map[string][]byte      // MIME type -> magic number
}

func NewFileValidator() *FileValidator {
    return &FileValidator{
        MaxSize: 5 * 1024 * 1024, // 5MB
        AllowedTypes: map[string][]byte{
            "image/jpeg": {0xFF, 0xD8, 0xFF},
            "image/png":  {0x89, 0x50, 0x4E, 0x47},
            "image/webp": {0x52, 0x49, 0x46, 0x46},
        },
    }
}

func (v *FileValidator) Validate(fileHeader *multipart.FileHeader, file multipart.File) error {
    // Check file size
    if fileHeader.Size > v.MaxSize {
        return fmt.Errorf("file too large: %d bytes (max: %d)", fileHeader.Size, v.MaxSize)
    }
    
    // Read magic number
    buffer := make([]byte, 512)
    n, err := file.Read(buffer)
    if err != nil && err != io.EOF {
        return fmt.Errorf("failed to read file: %w", err)
    }
    buffer = buffer[:n]
    
    // Reset file reader
    file.Seek(0, 0)
    
    // Verify magic number
    valid := false
    for mimeType, magic := range v.AllowedTypes {
        if bytes.HasPrefix(buffer, magic) {
            valid = true
            // Verify declared MIME type matches detected
            if fileHeader.Header.Get("Content-Type") != mimeType {
                return fmt.Errorf("file type mismatch: declared %s, detected %s", 
                    fileHeader.Header.Get("Content-Type"), mimeType)
            }
            break
        }
    }
    
    if !valid {
        return fmt.Errorf("invalid file type")
    }
    
    return nil
}
```

---

## 7. Error Handling

### 7.1 Generic Error Messages

```go
package middleware

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// ErrorHandler catches panics and errors
func ErrorHandler(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                // Log full error details internally
                logger.Error("panic recovered",
                    zap.Any("error", err),
                    zap.String("path", c.Request.URL.Path),
                    zap.String("method", c.Request.Method),
                    zap.String("client_ip", c.ClientIP()),
                )
                
                // Return generic error to client
                c.JSON(http.StatusInternalServerError, gin.H{
                    "success": false,
                    "error": gin.H{
                        "code":    "INTERNAL_ERROR",
                        "message": "An unexpected error occurred",
                    },
                })
            }
        }()
        
        c.Next()
    }
}
```

---

## 8. Security Checklist

### Pre-Deployment Checklist

- [ ] JWT using RS256 with strong keys (min 2048-bit RSA)
- [ ] bcrypt with cost 12 for passwords
- [ ] Rate limiting on all endpoints
- [ ] Input validation on all user inputs
- [ ] CORS properly configured (no wildcards in production)
- [ ] Security headers enabled
- [ ] HTTPS enforced (HSTS)
- [ ] File uploads validated (magic numbers, not just extensions)
- [ ] SQL/NoSQL injection prevention (parameterized queries)
- [ ] XSS prevention (output encoding)
- [ ] Resource ownership checks
- [ ] Audit logging enabled
- [ ] Secrets not hardcoded (use environment variables)
- [ ] Error messages don't leak sensitive info
- [ ] TLS 1.2+ configured
- [ ] Dependencies scanned for vulnerabilities
- [ ] Security headers tested

### Ongoing Security Tasks

- [ ] Regular dependency updates (weekly)
- [ ] Security scanning (monthly)
- [ ] Access log review (weekly)
- [ ] Failed authentication monitoring (daily)
- [ ] SSL certificate renewal monitoring
- [ ] Database backup testing (monthly)
- [ ] Penetration testing (quarterly)

---

## 9. Common Vulnerabilities & Mitigations

| Vulnerability | Risk | Mitigation |
|--------------|------|-----------|
| **Broken Authentication** | High | JWT with RS256, bcrypt, brute force protection |
| **Injection Attacks** | High | Parameterized queries, input validation |
| **XSS** | Medium | Output encoding, CSP headers |
| **CSRF** | Low | SameSite cookies (if using sessions) |
| **Sensitive Data Exposure** | High | HTTPS, encryption at rest, minimal data collection |
| **Broken Access Control** | High | Resource ownership checks, authorization middleware |
| **Security Misconfiguration** | Medium | Environment-specific configs, security headers |
| **Insecure File Uploads** | Medium | Magic number validation, size limits, secure storage |
| **Insufficient Logging** | Medium | Structured logging, audit trails, monitoring |

---

**Version:** 1.0  
**Last Updated:** 2024-01-15  
**Next Review:** 2024-04-15
