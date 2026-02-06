package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/services"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	userRepo     UserRepository
	tokenService *services.TokenService
	redisClient  RedisClient
	auditLog     AuditLogger
	rateLimiter  RateLimiter
	emailService EmailService
}

// UserRepository defines the user repository interface
type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
}

// RedisClient defines the Redis client interface
type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	SIsMember(ctx context.Context, key string, member interface{}) *redis.BoolCmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
}

// AuditLogger defines the audit logging interface
type AuditLogger interface {
	Log(ctx context.Context, userID, action string, metadata map[string]interface{})
}

// RateLimiter defines the rate limiting interface
type RateLimiter interface {
	AllowLogin(clientIP string) bool
	AllowPasswordReset(clientIP string) bool
	RecordFailedAttempt(clientIP string)
}

// EmailService defines the email service interface
type EmailService interface {
	SendVerificationEmail(email, token string)
	SendPasswordResetEmail(email, token string)
	SendPasswordChangedEmail(email string)
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	userRepo UserRepository,
	tokenService *services.TokenService,
	redisClient RedisClient,
	auditLog AuditLogger,
	rateLimiter RateLimiter,
	emailService EmailService,
) *AuthHandler {
	return &AuthHandler{
		userRepo:     userRepo,
		tokenService: tokenService,
		redisClient:  redisClient,
		auditLog:     auditLog,
		rateLimiter:  rateLimiter,
		emailService: emailService,
	}
}

// RegistrationRequest represents the registration payload
type RegistrationRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=8,max=72"`
	Name       string `json:"name" binding:"required,min=2,max=100"`
	DeviceInfo string `json:"device_info" binding:"required"`
}

// LoginRequest represents the login payload
type LoginRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	DeviceInfo string `json:"device_info" binding:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	User *models.User `json:"user"`
}

// ForgotPasswordRequest represents the forgot password payload
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents the reset password payload
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=72"`
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
		ID:           primitive.NewObjectID(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.Name,
		Name:         req.Name, // For compatibility
		Status:       models.UserStatusUnverified,
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create user"})
		return
	}

	// Generate verification token
	verificationToken, err := h.tokenService.GenerateVerificationToken(user.ID.Hex())
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
	tokenPair, err := h.tokenService.GenerateTokenPair(user.ID.Hex(), deviceID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate tokens"})
		return
	}

	// Store refresh token in Redis with device binding
	refreshKey := fmt.Sprintf("refresh:%s:%s", user.ID.Hex(), tokenPair.RefreshJTI)
	err = h.redisClient.Set(c.Request.Context(), refreshKey, deviceID, 7*24*time.Hour).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to store session"})
		return
	}

	// Set HTTP-only cookies
	setAuthCookies(c, tokenPair)

	// Log successful login
	h.auditLog.Log(c.Request.Context(), user.ID.Hex(), "login", map[string]interface{}{
		"ip":         clientIP,
		"device":     deviceID,
		"user_agent": c.Request.UserAgent(),
	})

	c.JSON(http.StatusOK, LoginResponse{User: user})
}

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
	err = h.redisClient.Set(c.Request.Context(), newRefreshKey, deviceID, 7*24*time.Hour).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to store session"})
		return
	}

	// Set new cookies
	setAuthCookies(c, tokenPair)

	c.JSON(http.StatusOK, gin.H{"message": "Token refreshed successfully"})
}

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
	err = h.redisClient.Set(c.Request.Context(), resetKey, user.ID.Hex(), time.Hour).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to store token"})
		return
	}

	// Send email asynchronously
	go h.emailService.SendPasswordResetEmail(user.Email, token)

	// Log security event
	h.auditLog.Log(c.Request.Context(), user.ID.Hex(), "password_reset_requested", map[string]interface{}{
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

// Helper functions

// validatePassword validates password strength
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

// generateDeviceID creates a device fingerprint
func generateDeviceID(deviceInfo, userAgent string) string {
	data := deviceInfo + userAgent
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// setAuthCookies sets authentication cookies
func setAuthCookies(c *gin.Context, tokens *services.TokenPair) {
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

// generateUUID generates a UUID v4
func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	// Set version (4) and variant bits
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
