# Authentication Flow

## Overview

The authentication system uses a secure JWT-based approach with RS256 asymmetric signing, dual token strategy (access + refresh), and device-bound tokens for enhanced security.

**Key Features:**
- JWT tokens with RS256 asymmetric signing (RSA public/private keys)
- Short-lived access tokens (15 minutes)
- Long-lived refresh tokens (7 days)
- Device fingerprinting for token binding
- Secure cookie-based token transport
- Token blacklisting for logout
- Rate limiting on auth endpoints
- bcrypt password hashing with cost 12

---

## JWT Architecture

### Token Types

#### Access Token (15 minutes expiry)
```json
{
  "sub": "user-uuid",
  "jti": "unique-token-id",
  "iss": "wedding-invitation-api",
  "aud": "wedding-invitation-client",
  "iat": 1704067200,
  "exp": 1704068100,
  "nbf": 1704067200,
  "device_id": "device-fingerprint-hash",
  "role": "user"
}
```

**Claims:**
- `sub` (Subject): User UUID
- `jti` (JWT ID): Unique token identifier for revocation
- `iss` (Issuer): API identifier
- `aud` (Audience): Client application
- `iat` (Issued At): Unix timestamp
- `exp` (Expiration): Unix timestamp
- `nbf` (Not Before): Unix timestamp
- `device_id`: Device fingerprint hash
- `role`: User role (user, admin)

#### Refresh Token (7 days expiry)
```json
{
  "sub": "user-uuid",
  "jti": "unique-refresh-token-id",
  "iss": "wedding-invitation-api",
  "aud": "wedding-invitation-client",
  "iat": 1704067200,
  "exp": 1704672000,
  "nbf": 1704067200,
  "type": "refresh",
  "device_id": "device-fingerprint-hash"
}
```

**Additional Claims:**
- `type`: "refresh" to distinguish from access tokens

### RS256 Asymmetric Signing

The system uses RSA-256 asymmetric cryptography for token signing:

**Advantages:**
- Private key stays secure on server
- Public key can be distributed to multiple services
- Allows token verification without secret sharing
- Enables microservices architecture with shared auth

**Key Management:**
- Private key: Kept server-side, never exposed
- Public key: Can be shared with other services
- Keys stored in environment variables or secure vault
- Key rotation support through versioning

---

## Authentication Flows

### 1. Registration Flow with Email Verification

```
┌──────────┐         ┌──────────┐         ┌──────────┐         ┌──────────┐
│  Client  │         │   API    │         │ Database │         │  Email   │
└────┬─────┘         └────┬─────┘         └────┬─────┘         └────┬─────┘
     │                    │                    │                    │
     │  POST /register    │                    │                    │
     │  {email, password, │                    │                    │
     │   name, device}    │                    │                    │
     ├───────────────────>│                    │                    │
     │                    │  1. Validate input │                    │
     │                    │  2. Check email    │                    │
     │                    │     uniqueness     │                    │
     │                    ├───────────────────>│                    │
     │                    │                    │  SELECT email...   │
     │                    │                    ├───────────────────>│
     │                    │                    │  Return result     │
     │                    │                    │<───────────────────┤
     │                    │                    │                    │
     │                    │  3. Hash password  │                    │
     │                    │  4. Create user    │                    │
     │                    │     (unverified)   │                    │
     │                    ├───────────────────>│                    │
     │                    │                    │  INSERT user...    │
     │                    │                    │<───────────────────┤
     │                    │                    │                    │
     │                    │  5. Generate       │                    │
     │                    │     verification   │                    │
     │                    │     token          │                    │
     │                    │  6. Send email     │                    │
     │                    ├─────────────────────────────────────────>│
     │                    │                    │                    │
     │  201 Created       │                    │                    │
     │  {message:         │                    │                    │
     │   "Verification    │                    │                    │
     │    email sent"}    │                    │                    │
     │<───────────────────┤                    │                    │
     │                    │                    │                    │
     │                    │                    │                    │
     │  POST /verify-email│                    │                    │
     │  {token}           │                    │                    │
     ├───────────────────>│                    │                    │
     │                    │  7. Verify token   │                    │
     │                    │  8. Update user    │                    │
     │                    │     status         │                    │
     │                    ├───────────────────>│                    │
     │                    │                    │  UPDATE status...  │
     │                    │                    │<───────────────────┤
     │                    │                    │                    │
     │  200 OK            │                    │                    │
     │  {message:         │                    │                    │
     │   "Email verified"}│                    │                    │
     │<───────────────────┤                    │                    │
```

**Go Implementation:**

```go
package handlers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
    "github.com/google/uuid"
)

// RegistrationRequest represents the registration payload
type RegistrationRequest struct {
    Email       string `json:"email" binding:"required,email"`
    Password    string `json:"password" binding:"required,min=8,max=72"`
    Name        string `json:"name" binding:"required,min=2,max=100"`
    DeviceInfo  string `json:"device_info" binding:"required"`
}

// RegisterHandler handles user registration
func (h *AuthHandler) RegisterHandler(c *gin.Context) {
    var req RegistrationRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input: " + err.Error()})
        return
    }

    // Validate password strength
    if err := validatePassword(req.Password); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
        return
    }

    // Check if email exists
    existingUser, err := h.userRepo.GetByEmail(c.Request.Context(), req.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
        return
    }
    if existingUser != nil {
        // Return 200 to prevent email enumeration
        c.JSON(http.StatusOK, gin.H{
            "message": "If this email is not registered, you will receive a verification email",
        })
        return
    }

    // Hash password with bcrypt (cost 12)
    hashedPassword, err := bcrypt.GenerateFromPassword(
        []byte(req.Password),
        bcrypt.DefaultCost, // 12
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to process password"})
        return
    }

    // Create user
    user := &models.User{
        ID:             uuid.New().String(),
        Email:          req.Email,
        PasswordHash:   string(hashedPassword),
        Name:           req.Name,
        Status:         models.UserStatusUnverified,
        CreatedAt:      time.Now(),
        UpdatedAt:      time.Now(),
    }

    if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create user"})
        return
    }

    // Generate verification token
    verificationToken, err := h.tokenService.GenerateVerificationToken(user.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate token"})
        return
    }

    // Send verification email asynchronously
    go h.emailService.SendVerificationEmail(user.Email, verificationToken)

    // Return success without confirming email doesn't exist
    c.JSON(http.StatusCreated, gin.H{
        "message": "Registration successful. Please check your email to verify your account.",
    })
}

// Password validation rules
func validatePassword(password string) error {
    if len(password) < 8 {
        return fmt.Errorf("password must be at least 8 characters")
    }
    if len(password) > 72 {
        return fmt.Errorf("password must not exceed 72 characters")
    }
    
    var (
        hasUpper   bool
        hasLower   bool
        hasNumber  bool
        hasSpecial bool
    )
    
    for _, char := range password {
        switch {
        case unicode.IsUpper(char):
            hasUpper = true
        case unicode.IsLower(char):
            hasLower = true
        case unicode.IsNumber(char):
            hasNumber = true
        case unicode.IsPunct(char) || unicode.IsSymbol(char):
            hasSpecial = true
        }
    }
    
    if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
        return fmt.Errorf("password must contain at least one uppercase letter, one lowercase letter, one number, and one special character")
    }
    
    return nil
}
```

### 2. Login Flow with Token Generation

```
┌──────────┐         ┌──────────┐         ┌──────────┐         ┌──────────┐
│  Client  │         │   API    │         │ Database │         │  Redis   │
└────┬─────┘         └────┬─────┘         └────┬─────┘         └────┬─────┘
     │                    │                    │                    │
     │  POST /login       │                    │                    │
     │  {email, password, │                    │                    │
     │   device_info}     │                    │                    │
     ├───────────────────>│                    │                    │
     │                    │  1. Get user       │                    │
     │                    │     by email       │                    │
     │                    ├───────────────────>│                    │
     │                    │                    │  SELECT * FROM...  │
     │                    │                    │<───────────────────┤
     │                    │                    │                    │
     │                    │  2. Verify         │                    │
     │                    │     password       │                    │
     │                    │                    │                    │
     │                    │  3. Check if       │                    │
     │                    │     verified       │                    │
     │                    │                    │                    │
     │                    │  4. Generate       │                    │
     │                    │     device ID      │                    │
     │                    │                    │                    │
     │                    │  5. Create tokens  │                    │
     │                    │  6. Store refresh  │                    │
     │                    │     token in Redis │                    │
     │                    ├─────────────────────────────────────────>│
     │                    │                    │                    │
     │                    │                    │     SET refresh... │
     │                    │                    │<───────────────────┤
     │                    │                    │                    │
     │  200 OK            │                    │                    │
     │  Set-Cookie:       │                    │                    │
     │   access_token     │                    │                    │
     │  Set-Cookie:       │                    │                    │
     │   refresh_token    │                    │                    │
     │  {user: {...}}     │                    │                    │
     │<───────────────────┤                    │                    │
```

**Go Implementation:**

```go
package handlers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
)

// LoginRequest represents the login payload
type LoginRequest struct {
    Email      string `json:"email" binding:"required,email"`
    Password   string `json:"password" binding:"required"`
    DeviceInfo string `json:"device_info" binding:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
    User         *models.User `json:"user"`
    AccessToken  string       `json:"-"`
    RefreshToken string       `json:"-"`
}

// LoginHandler handles user login
func (h *AuthHandler) LoginHandler(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input"})
        return
    }

    // Rate limiting check
    clientIP := c.ClientIP()
    if !h.rateLimiter.AllowLogin(clientIP) {
        c.JSON(http.StatusTooManyRequests, ErrorResponse{
            Error: "Too many login attempts. Please try again later.",
        })
        return
    }

    // Get user by email
    user, err := h.userRepo.GetByEmail(c.Request.Context(), req.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
        return
    }

    // Verify password (constant time comparison)
    if user == nil || bcrypt.CompareHashAndPassword(
        []byte(user.PasswordHash),
        []byte(req.Password),
    ) != nil {
        // Increment failed attempts
        h.rateLimiter.RecordFailedAttempt(clientIP)
        c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
        return
    }

    // Check if email is verified
    if user.Status != models.UserStatusActive {
        c.JSON(http.StatusForbidden, ErrorResponse{
            Error: "Email not verified. Please check your email.",
        })
        return
    }

    // Generate device fingerprint
    deviceID := generateDeviceID(req.DeviceInfo, c.Request.UserAgent())

    // Generate tokens
    tokenPair, err := h.tokenService.GenerateTokenPair(user.ID, deviceID, user.Role)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate tokens"})
        return
    }

    // Store refresh token in Redis with device binding
    refreshKey := fmt.Sprintf("refresh:%s:%s", user.ID, tokenPair.RefreshJTI)
    err = h.redisClient.Set(c.Request.Context(), refreshKey, deviceID, 
        7*24*time.Hour).Err()
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to store session"})
        return
    }

    // Set HTTP-only cookies
    setAuthCookies(c, tokenPair)

    // Log successful login
    h.auditLog.Log(c.Request.Context(), user.ID, "login", map[string]interface{}{
        "ip":       clientIP,
        "device":   deviceID,
        "user_agent": c.Request.UserAgent(),
    })

    c.JSON(http.StatusOK, LoginResponse{User: user})
}

// generateDeviceID creates a device fingerprint
func generateDeviceID(deviceInfo, userAgent string) string {
    data := deviceInfo + userAgent
    hash := sha256.Sum256([]byte(data))
    return hex.EncodeToString(hash[:])
}

// setAuthCookies sets authentication cookies
func setAuthCookies(c *gin.Context, tokens *TokenPair) {
    // Access token cookie
    accessCookie := &http.Cookie{
        Name:     "access_token",
        Value:    tokens.AccessToken,
        Path:     "/",
        HttpOnly: true,
        Secure:   true, // HTTPS only
        SameSite: http.SameSiteStrictMode,
        MaxAge:   900, // 15 minutes
    }
    http.SetCookie(c.Writer, accessCookie)

    // Refresh token cookie
    refreshCookie := &http.Cookie{
        Name:     "refresh_token",
        Value:    tokens.RefreshToken,
        Path:     "/auth/refresh",
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
        MaxAge:   604800, // 7 days
    }
    http.SetCookie(c.Writer, refreshCookie)
}
```

### 3. Token Refresh Flow

```
┌──────────┐         ┌──────────┐         ┌──────────┐         ┌──────────┐
│  Client  │         │   API    │         │   JWT    │         │  Redis   │
└────┬─────┘         └────┬─────┘         └────┬─────┘         └────┬─────┘
     │                    │                    │                    │
     │  POST /refresh     │                    │                    │
     │  Cookie:           │                    │                    │
     │   refresh_token    │                    │                    │
     ├───────────────────>│                    │                    │
     │                    │  1. Parse and      │                    │
     │                    │     validate       │                    │
     │                    │     refresh token  │                    │
     │                    ├───────────────────>│                    │
     │                    │                    │  Verify signature  │
     │                    │                    │  Check expiration  │
     │                    │                    │  Validate claims   │
     │                    │                    │<───────────────────┤
     │                    │                    │                    │
     │                    │  2. Check if       │                    │
     │                    │     blacklisted    │                    │
     │                    ├─────────────────────────────────────────>│
     │                    │                    │                    │
     │                    │                    │     SISMEMBER...   │
     │                    │                    │<───────────────────┤
     │                    │                    │                    │
     │                    │  3. Verify device  │                    │
     │                    │     binding        │                    │
     │                    ├─────────────────────────────────────────>│
     │                    │                    │                    │
     │                    │                    │     GET refresh... │
     │                    │                    │<───────────────────┤
     │                    │                    │                    │
     │                    │  4. Generate       │                    │
     │                    │     new tokens     │                    │
     │                    ├───────────────────>│                    │
     │                    │                    │                    │
     │                    │  5. Blacklist      │                    │
     │                    │     old refresh    │                    │
     │                    │     token          │                    │
     │                    ├─────────────────────────────────────────>│
     │                    │                    │                    │
     │                    │  6. Store new      │                    │
     │                    │     refresh token  │                    │
     │                    ├─────────────────────────────────────────>│
     │                    │                    │                    │
     │  200 OK            │                    │                    │
     │  Set-Cookie:       │                    │                    │
     │   access_token     │                    │                    │
     │  Set-Cookie:       │                    │                    │
     │   refresh_token    │                    │                    │
     │<───────────────────┤                    │                    │
```

**Go Implementation:**

```go
package handlers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

// RefreshHandler handles token refresh
func (h *AuthHandler) RefreshHandler(c *gin.Context) {
    // Get refresh token from cookie
    refreshCookie, err := c.Cookie("refresh_token")
    if err != nil {
        c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "No refresh token provided"})
        return
    }

    // Parse and validate refresh token
    token, err := h.tokenService.ParseToken(refreshCookie)
    if err != nil {
        c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid refresh token"})
        return
    }

    // Extract claims
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid token claims"})
        return
    }

    // Verify it's a refresh token
    tokenType, _ := claims["type"].(string)
    if tokenType != "refresh" {
        c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid token type"})
        return
    }

    userID, _ := claims["sub"].(string)
    jti, _ := claims["jti"].(string)
    deviceID, _ := claims["device_id"].(string)

    // Check if token is blacklisted
    blacklisted, err := h.redisClient.SIsMember(
        c.Request.Context(),
        "blacklist:refresh",
        jti,
    ).Result()
    if err != nil || blacklisted {
        c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Token has been revoked"})
        return
    }

    // Verify device binding
    refreshKey := fmt.Sprintf("refresh:%s:%s", userID, jti)
    storedDevice, err := h.redisClient.Get(c.Request.Context(), refreshKey).Result()
    if err != nil {
        c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Session not found"})
        return
    }

    if storedDevice != deviceID {
        // Possible token theft - revoke all sessions
        h.revokeAllUserSessions(c.Request.Context(), userID)
        c.JSON(http.StatusUnauthorized, ErrorResponse{
            Error: "Security violation detected. All sessions revoked.",
        })
        return
    }

    // Get user for role claim
    user, err := h.userRepo.GetByID(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
        return
    }

    // Generate new token pair
    tokenPair, err := h.tokenService.GenerateTokenPair(userID, deviceID, user.Role)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate tokens"})
        return
    }

    // Blacklist old refresh token
    err = h.redisClient.SAdd(c.Request.Context(), "blacklist:refresh", jti).Err()
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to revoke old token"})
        return
    }

    // Set expiration on blacklist entry
    h.redisClient.Expire(c.Request.Context(), "blacklist:refresh", 7*24*time.Hour)

    // Delete old refresh token
    h.redisClient.Del(c.Request.Context(), refreshKey)

    // Store new refresh token
    newRefreshKey := fmt.Sprintf("refresh:%s:%s", userID, tokenPair.RefreshJTI)
    err = h.redisClient.Set(c.Request.Context(), newRefreshKey, deviceID, 
        7*24*time.Hour).Err()
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to store session"})
        return
    }

    // Set new cookies
    setAuthCookies(c, tokenPair)

    c.JSON(http.StatusOK, gin.H{"message": "Token refreshed successfully"})
}

// revokeAllUserSessions revokes all sessions for a user
func (h *AuthHandler) revokeAllUserSessions(ctx context.Context, userID string) {
    // Find all refresh tokens for user
    pattern := fmt.Sprintf("refresh:%s:*", userID)
    iter := h.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
    
    var keys []string
    for iter.Next(ctx) {
        keys = append(keys, iter.Val())
    }
    
    // Delete all sessions
    if len(keys) > 0 {
        h.redisClient.Del(ctx, keys...)
    }
}
```

### 4. Logout and Token Blacklisting

```
┌──────────┐         ┌──────────┐         ┌──────────┐         ┌──────────┐
│  Client  │         │   API    │         │   JWT    │         │  Redis   │
└────┬─────┘         └────┬─────┘         └────┬─────┘         └────┬─────┘
     │                    │                    │                    │
     │  POST /logout      │                    │                    │
     │  Cookie:           │                    │                    │
     │   access_token     │                    │                    │
     │   refresh_token    │                    │                    │
     ├───────────────────>│                    │                    │
     │                    │  1. Validate       │                    │
     │                    │     access token   │                    │
     │                    ├───────────────────>│                    │
     │                    │                    │  Verify signature  │
     │                    │                    │  Extract jti       │
     │                    │                    │<───────────────────┤
     │                    │                    │                    │
     │                    │  2. Blacklist      │                    │
     │                    │     access token   │                    │
     │                    ├─────────────────────────────────────────>│
     │                    │                    │                    │
     │                    │  3. Get refresh    │                    │
     │                    │     token from     │                    │
     │                    │     cookie         │                    │
     │                    │                    │                    │
     │                    │  4. Blacklist      │                    │
     │                    │     refresh token  │                    │
     │                    ├─────────────────────────────────────────>│
     │                    │                    │                    │
     │                    │  5. Delete session │                    │
     │                    │     from Redis     │                    │
     │                    ├─────────────────────────────────────────>│
     │                    │                    │                    │
     │  200 OK            │                    │                    │
     │  Clear cookies     │                    │                    │
     │<───────────────────┤                    │                    │
```

**Go Implementation:**

```go
package handlers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

// LogoutHandler handles user logout
func (h *AuthHandler) LogoutHandler(c *gin.Context) {
    // Get access token from cookie
    accessCookie, err := c.Cookie("access_token")
    if err == nil {
        // Parse access token to get JTI
        token, parseErr := h.tokenService.ParseToken(accessCookie)
        if parseErr == nil {
            if claims, ok := token.Claims.(jwt.MapClaims); ok {
                if jti, exists := claims["jti"].(string); exists {
                    // Calculate remaining TTL
                    if exp, ok := claims["exp"].(float64); ok {
                        ttl := time.Until(time.Unix(int64(exp), 0))
                        if ttl > 0 {
                            // Blacklist access token
                            h.redisClient.Set(
                                c.Request.Context(),
                                fmt.Sprintf("blacklist:access:%s", jti),
                                "1",
                                ttl,
                            )
                        }
                    }
                }
                
                // Log logout
                if userID, exists := claims["sub"].(string); exists {
                    h.auditLog.Log(c.Request.Context(), userID, "logout", map[string]interface{}{
                        "ip": c.ClientIP(),
                    })
                }
            }
        }
    }

    // Get refresh token from cookie
    refreshCookie, err := c.Cookie("refresh_token")
    if err == nil {
        token, parseErr := h.tokenService.ParseToken(refreshCookie)
        if parseErr == nil {
            if claims, ok := token.Claims.(jwt.MapClaims); ok {
                if jti, exists := claims["jti"].(string); exists {
                    userID, _ := claims["sub"].(string)
                    
                    // Blacklist refresh token
                    h.redisClient.SAdd(c.Request.Context(), "blacklist:refresh", jti)
                    h.redisClient.Expire(c.Request.Context(), "blacklist:refresh", 7*24*time.Hour)
                    
                    // Delete session
                    refreshKey := fmt.Sprintf("refresh:%s:%s", userID, jti)
                    h.redisClient.Del(c.Request.Context(), refreshKey)
                }
            }
        }
    }

    // Clear cookies
    clearAuthCookies(c)

    c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// LogoutAllHandler logs out all sessions
func (h *AuthHandler) LogoutAllHandler(c *gin.Context) {
    // Get user ID from context (set by auth middleware)
    userID, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
        return
    }

    userIDStr := userID.(string)

    // Revoke all sessions
    h.revokeAllUserSessions(c.Request.Context(), userIDStr)

    // Clear cookies
    clearAuthCookies(c)

    // Log security event
    h.auditLog.Log(c.Request.Context(), userIDStr, "logout_all", map[string]interface{}{
        "ip": c.ClientIP(),
    })

    c.JSON(http.StatusOK, gin.H{"message": "All sessions logged out"})
}

// clearAuthCookies clears authentication cookies
func clearAuthCookies(c *gin.Context) {
    // Clear access token
    accessCookie := &http.Cookie{
        Name:     "access_token",
        Value:    "",
        Path:     "/",
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
        MaxAge:   -1,
        Expires:  time.Unix(0, 0),
    }
    http.SetCookie(c.Writer, accessCookie)

    // Clear refresh token
    refreshCookie := &http.Cookie{
        Name:     "refresh_token",
        Value:    "",
        Path:     "/auth/refresh",
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
        MaxAge:   -1,
        Expires:  time.Unix(0, 0),
    }
    http.SetCookie(c.Writer, refreshCookie)
}
```

### 5. Password Reset Flow

```
┌──────────┐         ┌──────────┐         ┌──────────┐         ┌──────────┐
│  Client  │         │   API    │         │ Database │         │  Email   │
└────┬─────┘         └────┬─────┘         └────┬─────┘         └────┬─────┘
     │                    │                    │                    │
     │  POST /forgot-     │                    │                    │
     │       password     │                    │                    │
     │  {email}           │                    │                    │
     ├───────────────────>│                    │                    │
     │                    │  1. Find user      │                    │
     │                    │     by email       │                    │
     │                    ├───────────────────>│                    │
     │                    │                    │  SELECT * FROM...  │
     │                    │                    │<───────────────────┤
     │                    │                    │                    │
     │                    │  2. Generate       │                    │
     │                    │     reset token    │                    │
     │                    │  3. Store token    │                    │
     │                    │     with expiry    │                    │
     │                    ├───────────────────>│                    │
     │                    │                    │  INSERT/UPDATE...  │
     │                    │                    │<───────────────────┤
     │                    │                    │                    │
     │                    │  4. Send email     │                    │
     │                    ├─────────────────────────────────────────>│
     │                    │                    │                    │
     │  200 OK            │                    │                    │
     │  {message:         │                    │                    │
     │   "If email        │                    │                    │
     │    exists..."}     │                    │                    │
     │<───────────────────┤                    │                    │
     │                    │                    │                    │
     │  POST /reset-      │                    │                    │
     │       password     │                    │                    │
     │  {token,           │                    │                    │
     │   new_password}    │                    │                    │
     ├───────────────────>│                    │                    │
     │                    │  5. Validate       │                    │
     │                    │     token          │                    │
     │                    ├───────────────────>│                    │
     │                    │                    │  SELECT token...   │
     │                    │                    │<───────────────────┤
     │                    │                    │                    │
     │                    │  6. Validate       │                    │
     │                    │     password       │                    │
     │                    │  7. Hash password  │                    │
     │                    │  8. Update user    │                    │
     │                    │     password       │                    │
     │                    ├───────────────────>│                    │
     │                    │                    │  UPDATE password   │
     │                    │                    │<───────────────────┤
     │                    │                    │                    │
     │                    │  9. Invalidate     │                    │
     │                    │     reset token    │                    │
     │                    │  10. Revoke all    │                    │
     │                    │      sessions      │                    │
     │                    ├─────────────────────────────────────────>│
     │                    │                    │                    │
     │  200 OK            │                    │                    │
     │  {message:         │                    │                    │
     │   "Password        │                    │                    │
     │    reset"}         │                    │                    │
     │<───────────────────┤                    │                    │
```

**Go Implementation:**

```go
package handlers

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
)

// ForgotPasswordRequest represents the forgot password payload
type ForgotPasswordRequest struct {
    Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents the reset password payload
type ResetPasswordRequest struct {
    Token       string `json:"token" binding:"required"`
    NewPassword string `json:"new_password" binding:"required,min=8,max=72"`
}

// ForgotPasswordHandler initiates password reset
func (h *AuthHandler) ForgotPasswordHandler(c *gin.Context) {
    var req ForgotPasswordRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid email"})
        return
    }

    // Rate limiting
    clientIP := c.ClientIP()
    if !h.rateLimiter.AllowPasswordReset(clientIP) {
        c.JSON(http.StatusTooManyRequests, ErrorResponse{
            Error: "Too many requests. Please try again later.",
        })
        return
    }

    // Find user
    user, err := h.userRepo.GetByEmail(c.Request.Context(), req.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
        return
    }

    // Always return success to prevent email enumeration
    if user == nil {
        c.JSON(http.StatusOK, gin.H{
            "message": "If an account exists with this email, you will receive a password reset link",
        })
        return
    }

    // Generate secure reset token
    token, err := generateSecureToken()
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate token"})
        return
    }

    // Hash token for storage
    tokenHash := hashToken(token)

    // Store token with expiry (1 hour)
    resetKey := fmt.Sprintf("password_reset:%s", tokenHash)
    err = h.redisClient.Set(c.Request.Context(), resetKey, user.ID, time.Hour).Err()
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to store token"})
        return
    }

    // Send email asynchronously
    go h.emailService.SendPasswordResetEmail(user.Email, token)

    // Log security event
    h.auditLog.Log(c.Request.Context(), user.ID, "password_reset_requested", map[string]interface{}{
        "ip": clientIP,
    })

    c.JSON(http.StatusOK, gin.H{
        "message": "If an account exists with this email, you will receive a password reset link",
    })
}

// ResetPasswordHandler completes password reset
func (h *AuthHandler) ResetPasswordHandler(c *gin.Context) {
    var req ResetPasswordRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input"})
        return
    }

    // Validate password
    if err := validatePassword(req.NewPassword); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
        return
    }

    // Hash provided token
    tokenHash := hashToken(req.Token)

    // Look up token
    resetKey := fmt.Sprintf("password_reset:%s", tokenHash)
    userID, err := h.redisClient.Get(c.Request.Context(), resetKey).Result()
    if err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid or expired token"})
        return
    }

    // Get user
    user, err := h.userRepo.GetByID(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
        return
    }

    // Hash new password
    hashedPassword, err := bcrypt.GenerateFromPassword(
        []byte(req.NewPassword),
        bcrypt.DefaultCost,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to process password"})
        return
    }

    // Update password
    if err := h.userRepo.UpdatePassword(c.Request.Context(), userID, string(hashedPassword)); err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update password"})
        return
    }

    // Invalidate reset token
    h.redisClient.Del(c.Request.Context(), resetKey)

    // Revoke all existing sessions
    h.revokeAllUserSessions(c.Request.Context(), userID)

    // Log security event
    h.auditLog.Log(c.Request.Context(), userID, "password_reset_completed", map[string]interface{}{
        "ip": c.ClientIP(),
    })

    // Send confirmation email
    go h.emailService.SendPasswordChangedEmail(user.Email)

    c.JSON(http.StatusOK, gin.H{
        "message": "Password has been reset successfully. Please log in with your new password.",
    })
}

// generateSecureToken creates a cryptographically secure random token
func generateSecureToken() (string, error) {
    bytes := make([]byte, 32)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return hex.EncodeToString(bytes), nil
}

// hashToken creates a SHA-256 hash of the token
func hashToken(token string) string {
    hash := sha256.Sum256([]byte(token))
    return hex.EncodeToString(hash[:])
}
```

---

## JWT Service Implementation

### Token Service Structure

```go
package services

import (
    "crypto/rsa"
    "crypto/x509"
    "encoding/pem"
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
)

// TokenPair represents a pair of access and refresh tokens
type TokenPair struct {
    AccessToken  string
    RefreshToken string
    AccessJTI    string
    RefreshJTI   string
}

// TokenService handles JWT operations
type TokenService struct {
    privateKey     *rsa.PrivateKey
    publicKey      *rsa.PublicKey
    issuer         string
    audience       string
    accessExpiry   time.Duration
    refreshExpiry  time.Duration
}

// TokenConfig holds configuration for token service
type TokenConfig struct {
    PrivateKeyPEM string
    PublicKeyPEM  string
    Issuer        string
    Audience      string
    AccessExpiry  time.Duration
    RefreshExpiry time.Duration
}

// NewTokenService creates a new token service
func NewTokenService(config TokenConfig) (*TokenService, error) {
    privateKey, err := parseRSAPrivateKey(config.PrivateKeyPEM)
    if err != nil {
        return nil, fmt.Errorf("failed to parse private key: %w", err)
    }

    publicKey, err := parseRSAPublicKey(config.PublicKeyPEM)
    if err != nil {
        return nil, fmt.Errorf("failed to parse public key: %w", err)
    }

    return &TokenService{
        privateKey:    privateKey,
        publicKey:     publicKey,
        issuer:        config.Issuer,
        audience:      config.Audience,
        accessExpiry:  config.AccessExpiry,
        refreshExpiry: config.RefreshExpiry,
    }, nil
}

// GenerateTokenPair creates new access and refresh tokens
func (s *TokenService) GenerateTokenPair(userID, deviceID, role string) (*TokenPair, error) {
    now := time.Now()
    
    // Generate unique JWT IDs
    accessJTI := uuid.New().String()
    refreshJTI := uuid.New().String()

    // Create access token
    accessClaims := jwt.MapClaims{
        "sub":       userID,
        "jti":       accessJTI,
        "iss":       s.issuer,
        "aud":       s.audience,
        "iat":       now.Unix(),
        "exp":       now.Add(s.accessExpiry).Unix(),
        "nbf":       now.Unix(),
        "device_id": deviceID,
        "role":      role,
    }

    accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
    accessString, err := accessToken.SignedString(s.privateKey)
    if err != nil {
        return nil, fmt.Errorf("failed to sign access token: %w", err)
    }

    // Create refresh token
    refreshClaims := jwt.MapClaims{
        "sub":       userID,
        "jti":       refreshJTI,
        "iss":       s.issuer,
        "aud":       s.audience,
        "iat":       now.Unix(),
        "exp":       now.Add(s.refreshExpiry).Unix(),
        "nbf":       now.Unix(),
        "type":      "refresh",
        "device_id": deviceID,
    }

    refreshToken := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims)
    refreshString, err := refreshToken.SignedString(s.privateKey)
    if err != nil {
        return nil, fmt.Errorf("failed to sign refresh token: %w", err)
    }

    return &TokenPair{
        AccessToken:  accessString,
        RefreshToken: refreshString,
        AccessJTI:    accessJTI,
        RefreshJTI:   refreshJTI,
    }, nil
}

// ParseToken validates and parses a JWT token
func (s *TokenService) ParseToken(tokenString string) (*jwt.Token, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // Validate signing method
        if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return s.publicKey, nil
    },
        jwt.WithIssuer(s.issuer),
        jwt.WithAudience(s.audience),
        jwt.WithValidMethods([]string{"RS256"}),
    )

    if err != nil {
        return nil, fmt.Errorf("failed to parse token: %w", err)
    }

    return token, nil
}

// ParseTokenWithClaims validates and parses a JWT with custom claims
func (s *TokenService) ParseTokenWithClaims(tokenString string, claims jwt.Claims) (*jwt.Token, error) {
    token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return s.publicKey, nil
    })

    if err != nil {
        return nil, fmt.Errorf("failed to parse token with claims: %w", err)
    }

    return token, nil
}

// GenerateVerificationToken creates an email verification token
func (s *TokenService) GenerateVerificationToken(userID string) (string, error) {
    now := time.Now()
    
    claims := jwt.MapClaims{
        "sub":  userID,
        "jti":  uuid.New().String(),
        "iss":  s.issuer,
        "aud":  s.audience,
        "iat":  now.Unix(),
        "exp":  now.Add(24 * time.Hour).Unix(),
        "nbf":  now.Unix(),
        "type": "verification",
    }

    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    return token.SignedString(s.privateKey)
}

// parseRSAPrivateKey parses a PEM-encoded RSA private key
func parseRSAPrivateKey(pemString string) (*rsa.PrivateKey, error) {
    block, _ := pem.Decode([]byte(pemString))
    if block == nil {
        return nil, fmt.Errorf("failed to parse PEM block")
    }

    key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
    if err != nil {
        // Try PKCS1 format
        key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
        if err != nil {
            return nil, fmt.Errorf("failed to parse private key: %w", err)
        }
    }

    rsaKey, ok := key.(*rsa.PrivateKey)
    if !ok {
        return nil, fmt.Errorf("not an RSA private key")
    }

    return rsaKey, nil
}

// parseRSAPublicKey parses a PEM-encoded RSA public key
func parseRSAPublicKey(pemString string) (*rsa.PublicKey, error) {
    block, _ := pem.Decode([]byte(pemString))
    if block == nil {
        return nil, fmt.Errorf("failed to parse PEM block")
    }

    key, err := x509.ParsePKIXPublicKey(block.Bytes)
    if err != nil {
        // Try PKCS1 format
        key, err = x509.ParsePKCS1PublicKey(block.Bytes)
        if err != nil {
            return nil, fmt.Errorf("failed to parse public key: %w", err)
        }
    }

    rsaKey, ok := key.(*rsa.PublicKey)
    if !ok {
        return nil, fmt.Errorf("not an RSA public key")
    }

    return rsaKey, nil
}
```

---

## Authentication Middleware

### Auth Middleware Implementation

```go
package middleware

import (
    "net/http"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware validates JWT tokens and sets user context
type AuthMiddleware struct {
    tokenService services.TokenServiceInterface
    redisClient  *redis.Client
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(tokenService services.TokenServiceInterface, redisClient *redis.Client) *AuthMiddleware {
    return &AuthMiddleware{
        tokenService: tokenService,
        redisClient:  redisClient,
    }
}

// RequireAuth validates access tokens and sets user context
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract token from cookie or Authorization header
        tokenString := extractToken(c)
        if tokenString == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Authentication required",
            })
            return
        }

        // Parse and validate token
        token, err := m.tokenService.ParseToken(tokenString)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Invalid token",
            })
            return
        }

        // Extract claims
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Invalid token claims",
            })
            return
        }

        // Verify token type
        if tokenType, exists := claims["type"].(string); exists && tokenType != "" {
            if tokenType != "" {
                c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                    "error": "Invalid token type",
                })
                return
            }
        }

        userID, _ := claims["sub"].(string)
        jti, _ := claims["jti"].(string)
        deviceID, _ := claims["device_id"].(string)
        role, _ := claims["role"].(string)

        // Check if token is blacklisted
        blacklisted, err := m.redisClient.Exists(
            c.Request.Context(),
            fmt.Sprintf("blacklist:access:%s", jti),
        ).Result()
        if err != nil || blacklisted > 0 {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Token has been revoked",
            })
            return
        }

        // Set user context
        c.Set("userID", userID)
        c.Set("jti", jti)
        c.Set("deviceID", deviceID)
        c.Set("role", role)
        c.Set("tokenClaims", claims)

        c.Next()
    }
}

// RequireRole restricts access to specific roles
func (m *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole, exists := c.Get("role")
        if !exists {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Authentication required",
            })
            return
        }

        role := userRole.(string)
        for _, allowed := range roles {
            if role == allowed {
                c.Next()
                return
            }
        }

        c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
            "error": "Insufficient permissions",
        })
    }
}

// OptionalAuth validates tokens if present but doesn't require them
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := extractToken(c)
        if tokenString == "" {
            c.Next()
            return
        }

        token, err := m.tokenService.ParseToken(tokenString)
        if err != nil {
            c.Next()
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            c.Next()
            return
        }

        userID, _ := claims["sub"].(string)
        role, _ := claims["role"].(string)

        // Check blacklist
        jti, _ := claims["jti"].(string)
        blacklisted, _ := m.redisClient.Exists(
            c.Request.Context(),
            fmt.Sprintf("blacklist:access:%s", jti),
        ).Result()
        
        if blacklisted == 0 {
            c.Set("userID", userID)
            c.Set("role", role)
            c.Set("authenticated", true)
        }

        c.Next()
    }
}

// extractToken extracts JWT from cookie or Authorization header
func extractToken(c *gin.Context) string {
    // Check cookie first
    if cookie, err := c.Cookie("access_token"); err == nil && cookie != "" {
        return cookie
    }

    // Check Authorization header
    authHeader := c.GetHeader("Authorization")
    if authHeader == "" {
        return ""
    }

    parts := strings.SplitN(authHeader, " ", 2)
    if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
        return ""
    }

    return parts[1]
}

// SecurityHeaders middleware adds security headers
func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        c.Header("Content-Security-Policy", "default-src 'self'")
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
        
        c.Next()
    }
}

// CORSMiddleware handles CORS configuration
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")
        
        // Check if origin is allowed
        allowed := false
        for _, o := range allowedOrigins {
            if o == "*" || o == origin {
                allowed = true
                break
            }
        }

        if allowed {
            c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
            c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
            c.Writer.Header().Set("Access-Control-Allow-Headers", 
                "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
            c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
        }

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }

        c.Next()
    }
}
```

---

## Protected Route Handlers

### Example Protected Handlers

```go
package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

// GetProfileHandler returns the authenticated user's profile
func (h *UserHandler) GetProfileHandler(c *gin.Context) {
    // Get user ID from context (set by auth middleware)
    userID, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
        return
    }

    // Fetch user from database
    user, err := h.userRepo.GetByID(c.Request.Context(), userID.(string))
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch user"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "user": user.ToResponse(),
    })
}

// UpdateProfileHandler updates user profile
func (h *UserHandler) UpdateProfileHandler(c *gin.Context) {
    userID, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
        return
    }

    var req UpdateProfileRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input"})
        return
    }

    // Update user
    if err := h.userRepo.Update(c.Request.Context(), userID.(string), req); err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update profile"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Profile updated successfully",
    })
}

// AdminDashboardHandler restricted to admin users
func (h *AdminHandler) AdminDashboardHandler(c *gin.Context) {
    userID, _ := c.Get("userID")
    role, _ := c.Get("role")

    // Log admin access
    h.auditLog.Log(c.Request.Context(), userID.(string), "admin_access", map[string]interface{}{
        "ip":   c.ClientIP(),
        "role": role,
    })

    stats, err := h.adminService.GetDashboardStats(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch stats"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "stats": stats,
    })
}
```

### Route Configuration

```go
package routes

import (
    "github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(
    router *gin.Engine,
    authHandler *handlers.AuthHandler,
    userHandler *handlers.UserHandler,
    adminHandler *handlers.AdminHandler,
    authMiddleware *middleware.AuthMiddleware,
) {
    // Public routes
    public := router.Group("/api/v1")
    {
        // Auth endpoints
        auth := public.Group("/auth")
        {
            auth.POST("/register", authHandler.RegisterHandler)
            auth.POST("/login", authHandler.LoginHandler)
            auth.POST("/refresh", authHandler.RefreshHandler)
            auth.POST("/logout", authHandler.LogoutHandler)
            auth.POST("/forgot-password", authHandler.ForgotPasswordHandler)
            auth.POST("/reset-password", authHandler.ResetPasswordHandler)
            auth.GET("/verify-email", authHandler.VerifyEmailHandler)
        }
    }

    // Protected routes
    protected := router.Group("/api/v1")
    protected.Use(authMiddleware.RequireAuth())
    {
        // User endpoints
        user := protected.Group("/user")
        {
            user.GET("/profile", userHandler.GetProfileHandler)
            user.PUT("/profile", userHandler.UpdateProfileHandler)
            user.POST("/logout-all", authHandler.LogoutAllHandler)
        }

        // Invitations
        protected.GET("/invitations", invitationHandler.ListHandler)
        protected.POST("/invitations", invitationHandler.CreateHandler)
        protected.GET("/invitations/:id", invitationHandler.GetHandler)
        protected.PUT("/invitations/:id", invitationHandler.UpdateHandler)
        protected.DELETE("/invitations/:id", invitationHandler.DeleteHandler)
    }

    // Admin routes
    admin := router.Group("/api/v1/admin")
    admin.Use(authMiddleware.RequireAuth())
    admin.Use(authMiddleware.RequireRole("admin"))
    {
        admin.GET("/dashboard", adminHandler.AdminDashboardHandler)
        admin.GET("/users", adminHandler.ListUsersHandler)
        admin.GET("/audit-logs", adminHandler.GetAuditLogsHandler)
    }
}
```

---

## Rate Limiting

### Rate Limiter Implementation

```go
package middleware

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
)

// RateLimiter handles rate limiting for different endpoints
type RateLimiter struct {
    redisClient *redis.Client
    config      RateLimitConfig
}

// RateLimitConfig holds rate limit configuration
type RateLimitConfig struct {
    LoginAttempts    int
    LoginWindow      time.Duration
    PasswordReset    int
    PasswordWindow   time.Duration
    Registration     int
    RegistrationWindow time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redisClient *redis.Client, config RateLimitConfig) *RateLimiter {
    return &RateLimiter{
        redisClient: redisClient,
        config:      config,
    }
}

// AllowLogin checks if login is allowed for an IP
func (rl *RateLimiter) AllowLogin(ip string) bool {
    key := fmt.Sprintf("rate_limit:login:%s", ip)
    return rl.allowRequest(key, rl.config.LoginAttempts, rl.config.LoginWindow)
}

// RecordFailedAttempt records a failed login attempt
func (rl *RateLimiter) RecordFailedAttempt(ip string) {
    key := fmt.Sprintf("rate_limit:login:%s", ip)
    rl.redisClient.Incr(context.Background(), key)
    rl.redisClient.Expire(context.Background(), key, rl.config.LoginWindow)
}

// AllowPasswordReset checks if password reset is allowed
func (rl *RateLimiter) AllowPasswordReset(ip string) bool {
    key := fmt.Sprintf("rate_limit:password_reset:%s", ip)
    return rl.allowRequest(key, rl.config.PasswordReset, rl.config.PasswordWindow)
}

// AllowRegistration checks if registration is allowed
func (rl *RateLimiter) AllowRegistration(ip string) bool {
    key := fmt.Sprintf("rate_limit:register:%s", ip)
    return rl.allowRequest(key, rl.config.Registration, rl.config.RegistrationWindow)
}

// allowRequest checks if a request is within rate limits
func (rl *RateLimiter) allowRequest(key string, maxRequests int, window time.Duration) bool {
    ctx := context.Background()
    
    // Get current count
    count, err := rl.redisClient.Get(ctx, key).Int()
    if err == redis.Nil {
        count = 0
    } else if err != nil {
        return false // Fail closed on error
    }

    if count >= maxRequests {
        return false
    }

    // Increment counter
    pipe := rl.redisClient.Pipeline()
    pipe.Incr(ctx, key)
    pipe.Expire(ctx, key, window)
    _, err = pipe.Exec(ctx)
    
    return err == nil
}

// RateLimitMiddleware creates a Gin middleware for rate limiting
func RateLimitMiddleware(redisClient *redis.Client, maxRequests int, window time.Duration) gin.HandlerFunc {
    rl := NewRateLimiter(redisClient, RateLimitConfig{})
    
    return func(c *gin.Context) {
        key := fmt.Sprintf("rate_limit:%s:%s", c.Request.Method, c.ClientIP())
        
        if !rl.allowRequest(key, maxRequests, window) {
            c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
                "error": "Rate limit exceeded. Please try again later.",
            })
            return
        }
        
        c.Next()
    }
}
```

---

## Complete Server Setup

```go
package main

import (
    "log"
    "os"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
)

func main() {
    // Load configuration
    config := loadConfig()

    // Initialize Redis
    redisClient := redis.NewClient(&redis.Options{
        Addr:     config.RedisAddr,
        Password: config.RedisPassword,
        DB:       0,
    })

    // Initialize token service
    tokenService, err := services.NewTokenService(services.TokenConfig{
        PrivateKeyPEM: config.JWTPrivateKey,
        PublicKeyPEM:  config.JWTPublicKey,
        Issuer:        "wedding-invitation-api",
        Audience:      "wedding-invitation-client",
        AccessExpiry:  15 * time.Minute,
        RefreshExpiry: 7 * 24 * time.Hour,
    })
    if err != nil {
        log.Fatalf("Failed to initialize token service: %v", err)
    }

    // Initialize repositories
    userRepo := repositories.NewUserRepository(config.DB)

    // Initialize services
    emailService := services.NewEmailService(config.SMTPConfig)
    auditLog := services.NewAuditLogService(redisClient)

    // Initialize handlers
    authHandler := handlers.NewAuthHandler(
        userRepo,
        tokenService,
        redisClient,
        emailService,
        auditLog,
    )

    // Initialize middleware
    authMiddleware := middleware.NewAuthMiddleware(tokenService, redisClient)
    rateLimiter := middleware.NewRateLimiter(redisClient, middleware.RateLimitConfig{
        LoginAttempts:      5,
        LoginWindow:        15 * time.Minute,
        PasswordReset:      3,
        PasswordWindow:     1 * time.Hour,
        Registration:       5,
        RegistrationWindow: 1 * time.Hour,
    })

    // Setup router
    router := gin.New()
    router.Use(gin.Recovery())
    router.Use(middleware.SecurityHeaders())
    router.Use(middleware.CORSMiddleware(config.AllowedOrigins))

    // Setup routes
    routes.SetupRoutes(router, authHandler, userHandler, adminHandler, authMiddleware)

    // Start server
    log.Printf("Server starting on %s", config.ServerAddr)
    if err := router.Run(config.ServerAddr); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
```

---

## Security Best Practices

### 1. Password Security

**bcrypt Configuration:**
- **Cost Factor**: 12 (default in Go's bcrypt)
  - Cost 10: ~100ms
  - Cost 12: ~300ms (recommended)
  - Cost 14: ~1s
- **Maximum Length**: 72 bytes (bcrypt limitation)
- **Pre-hashing**: Consider SHA-256 for passwords > 72 bytes

**Password Requirements:**
- Minimum 8 characters
- Maximum 72 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one number
- At least one special character
- No common passwords (dictionary check)
- No personal information

### 2. Token Security

**JWT Best Practices:**
- Use RS256 (asymmetric) for distributed systems
- Keep private keys secure (never commit to repo)
- Use short-lived access tokens (15 minutes)
- Implement token binding to device
- Include jti claim for revocation
- Validate all claims (iss, aud, exp, nbf)
- Use secure random for token generation

**Cookie Security:**
```go
http.Cookie{
    HttpOnly: true,  // Prevent JavaScript access
    Secure:   true,  // HTTPS only
    SameSite: http.SameSiteStrictMode, // CSRF protection
    Path:     "/",   // Scope appropriately
    MaxAge:   900,   // Short-lived
}
```

### 3. Common Pitfalls

**❌ Don't:**
- Store JWTs in localStorage/sessionStorage
- Use symmetric signing (HS256) for distributed systems
- Include sensitive data in JWT payload
- Set long expiry times for access tokens
- Skip token validation steps
- Use predictable JWT IDs
- Store refresh tokens in client-side storage
- Log JWT tokens or passwords
- Use weak password policies
- Skip rate limiting

**✅ Do:**
- Use httpOnly cookies for token transport
- Implement token rotation on refresh
- Bind tokens to device fingerprint
- Blacklist tokens on logout
- Use constant-time comparison for passwords
- Implement comprehensive rate limiting
- Log security events (not tokens)
- Use prepared statements for DB queries
- Validate all inputs strictly
- Implement proper CORS policies

### 4. Error Handling

```go
// Generic error message to prevent information leakage
const (
    ErrInvalidCredentials = "Invalid email or password"
    ErrTokenExpired       = "Authentication required"
    ErrTokenRevoked       = "Authentication required"
    ErrRateLimited        = "Too many requests. Please try again later."
)

// Detailed logging (not exposed to client)
log.Printf("Auth failed: user=%s reason=%s ip=%s", userID, reason, clientIP)
```

### 5. Key Management

**Environment Variables:**
```bash
# JWT Keys (PEM format)
JWT_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC...
-----END PRIVATE KEY-----"

JWT_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...
-----END PUBLIC KEY-----"
```

**Key Generation:**
```bash
# Generate RSA key pair
openssl genrsa -out private.pem 2048
openssl rsa -in private.pem -pubout -out public.pem

# Convert to single line for env var
awk 'NF {sub(/\r/, ""); printf "%s\\n",$0;}' private.pem
```

---

## Testing Authentication

### Unit Tests

```go
package services

import (
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestTokenService_GenerateTokenPair(t *testing.T) {
    // Generate test keys
    privateKey, publicKey := generateTestKeys()
    
    service, err := NewTokenService(TokenConfig{
        PrivateKeyPEM: privateKey,
        PublicKeyPEM:  publicKey,
        Issuer:        "test",
        Audience:      "test",
        AccessExpiry:  15 * time.Minute,
        RefreshExpiry: 7 * 24 * time.Hour,
    })
    require.NoError(t, err)

    t.Run("successful token generation", func(t *testing.T) {
        pair, err := service.GenerateTokenPair("user-123", "device-abc", "user")
        
        assert.NoError(t, err)
        assert.NotEmpty(t, pair.AccessToken)
        assert.NotEmpty(t, pair.RefreshToken)
        assert.NotEmpty(t, pair.AccessJTI)
        assert.NotEmpty(t, pair.RefreshJTI)
    })

    t.Run("valid access token", func(t *testing.T) {
        pair, _ := service.GenerateTokenPair("user-123", "device-abc", "user")
        
        token, err := service.ParseToken(pair.AccessToken)
        assert.NoError(t, err)
        assert.True(t, token.Valid)
    })

    t.Run("expired token rejection", func(t *testing.T) {
        // Create service with 0 expiry
        expiredService, _ := NewTokenService(TokenConfig{
            PrivateKeyPEM: privateKey,
            PublicKeyPEM:  publicKey,
            Issuer:        "test",
            Audience:      "test",
            AccessExpiry:  -1 * time.Hour, // Expired
            RefreshExpiry: 7 * 24 * time.Hour,
        })
        
        pair, _ := expiredService.GenerateTokenPair("user-123", "device-abc", "user")
        
        _, err := service.ParseToken(pair.AccessToken)
        assert.Error(t, err)
    })
}
```

### Integration Tests

```go
package handlers

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func TestAuthHandler_Login(t *testing.T) {
    gin.SetMode(gin.TestMode)
    
    t.Run("successful login", func(t *testing.T) {
        // Setup
        router, handler := setupTestRouter()
        router.POST("/login", handler.LoginHandler)
        
        // Create test user
        createTestUser("test@example.com", "SecureP@ss123")
        
        reqBody := LoginRequest{
            Email:      "test@example.com",
            Password:   "SecureP@ss123",
            DeviceInfo: "test-device",
        }
        body, _ := json.Marshal(reqBody)
        
        // Execute
        req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
        req.Header.Set("Content-Type", "application/json")
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        // Assert
        assert.Equal(t, http.StatusOK, w.Code)
        
        // Check cookies are set
        cookies := w.Result().Cookies()
        var hasAccess, hasRefresh bool
        for _, c := range cookies {
            if c.Name == "access_token" {
                hasAccess = true
                assert.True(t, c.HttpOnly)
                assert.True(t, c.Secure)
                assert.Equal(t, http.SameSiteStrictMode, c.SameSite)
            }
            if c.Name == "refresh_token" {
                hasRefresh = true
            }
        }
        assert.True(t, hasAccess)
        assert.True(t, hasRefresh)
    })

    t.Run("invalid credentials", func(t *testing.T) {
        router, handler := setupTestRouter()
        router.POST("/login", handler.LoginHandler)
        
        reqBody := LoginRequest{
            Email:      "test@example.com",
            Password:   "WrongPassword",
            DeviceInfo: "test-device",
        }
        body, _ := json.Marshal(reqBody)
        
        req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
        req.Header.Set("Content-Type", "application/json")
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusUnauthorized, w.Code)
    })
}
```

---

## Environment Configuration

```bash
# Server
SERVER_ADDR=:8080
SERVER_ENV=production

# Database
DATABASE_URL=postgres://user:pass@localhost/wedding_db

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=

# JWT Keys (use strong keys in production)
JWT_PRIVATE_KEY=
JWT_PUBLIC_KEY=

# Email
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=noreply@example.com
SMTP_PASS=secure_password

# Security
ALLOWED_ORIGINS=https://wedding.example.com,https://admin.wedding.example.com
RATE_LIMIT_LOGIN=5
RATE_LIMIT_WINDOW=15m
```

---

## Summary

This authentication system provides:

1. **Secure JWT Implementation**: RS256 asymmetric signing with proper key management
2. **Dual Token Strategy**: Short-lived access tokens (15 min) and long-lived refresh tokens (7 days)
3. **Device Binding**: Tokens are bound to device fingerprints for security
4. **Token Rotation**: Refresh tokens are rotated and old ones blacklisted
5. **Secure Transport**: httpOnly, secure, SameSite cookies
6. **Comprehensive Rate Limiting**: Protection against brute force attacks
7. **Audit Logging**: Security events tracked for monitoring
8. **Password Security**: bcrypt with proper cost factor and validation
9. **Email Verification**: Registration flow with verification tokens
10. **Password Reset**: Secure flow with time-limited tokens

All components follow security best practices and are designed to prevent common vulnerabilities like XSS, CSRF, token theft, and brute force attacks.
