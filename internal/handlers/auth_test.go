package handlers

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/services"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	args := m.Called(ctx, userID, passwordHash)
	return args.Error(0)
}

// MockRedisClient is a mock implementation of RedisClient
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	cmd := redis.NewStatusCmd(ctx)
	if args.Error(0) != nil {
		cmd.SetErr(args.Error(0))
	}
	return cmd
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	cmd := redis.NewStringCmd(ctx)
	if args.Error(0) != nil {
		cmd.SetErr(args.Error(0))
	} else {
		cmd.SetVal(args.String(0))
	}
	return cmd
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	cmd := redis.NewIntCmd(ctx)
	if args.Error(0) != nil {
		cmd.SetErr(args.Error(0))
	}
	return cmd
}

func (m *MockRedisClient) SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	args := m.Called(ctx, key, members)
	cmd := redis.NewIntCmd(ctx)
	if args.Error(0) != nil {
		cmd.SetErr(args.Error(0))
	}
	return cmd
}

func (m *MockRedisClient) SIsMember(ctx context.Context, key string, member interface{}) *redis.BoolCmd {
	args := m.Called(ctx, key, member)
	cmd := redis.NewBoolCmd(ctx)
	if args.Error(0) != nil {
		cmd.SetErr(args.Error(0))
	} else {
		cmd.SetVal(args.Bool(0))
	}
	return cmd
}

func (m *MockRedisClient) Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
	m.Called(ctx, cursor, match, count)
	return nil
}

func (m *MockRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, expiration)
	cmd := redis.NewBoolCmd(ctx)
	if args.Error(0) != nil {
		cmd.SetErr(args.Error(0))
	}
	return cmd
}

// MockAuditLogger is a mock implementation of AuditLogger
type MockAuditLogger struct {
	mock.Mock
}

func (m *MockAuditLogger) Log(ctx context.Context, userID, action string, metadata map[string]interface{}) {
	m.Called(ctx, userID, action, metadata)
}

// MockRateLimiter is a mock implementation of RateLimiter
type MockRateLimiter struct {
	mock.Mock
}

func (m *MockRateLimiter) AllowLogin(clientIP string) bool {
	args := m.Called(clientIP)
	return args.Bool(0)
}

func (m *MockRateLimiter) AllowPasswordReset(clientIP string) bool {
	args := m.Called(clientIP)
	return args.Bool(0)
}

func (m *MockRateLimiter) RecordFailedAttempt(clientIP string) {
	m.Called(clientIP)
}

// MockEmailService is a mock implementation of EmailService
type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendVerificationEmail(email, token string) {
	m.Called(email, token)
}

func (m *MockEmailService) SendPasswordResetEmail(email, token string) {
	m.Called(email, token)
}

func (m *MockEmailService) SendPasswordChangedEmail(email string) {
	m.Called(email)
}

func TestValidatePassword(t *testing.T) {
	// Test valid passwords
	assert.NoError(t, validatePassword("Password123!"))
	assert.NoError(t, validatePassword("StrongPass1@"))

	// Test invalid passwords
	assert.Error(t, validatePassword("short"))                 // Too short
	assert.Error(t, validatePassword("alllowercase123!"))      // No uppercase
	assert.Error(t, validatePassword("ALLUPPERCASE123!"))      // No lowercase
	assert.Error(t, validatePassword("NoNumbers!"))            // No numbers
	assert.Error(t, validatePassword("NoSpecials123"))         // No special characters
	assert.Error(t, validatePassword(strings.Repeat("a", 73))) // Too long
}

func TestGenerateDeviceID(t *testing.T) {
	deviceID1 := generateDeviceID("device-info", "user-agent")
	deviceID2 := generateDeviceID("device-info", "user-agent")
	deviceID3 := generateDeviceID("different-device", "user-agent")

	// Same input should generate same device ID
	assert.Equal(t, deviceID1, deviceID2)

	// Different input should generate different device ID
	assert.NotEqual(t, deviceID1, deviceID3)

	// Device ID should be a valid SHA-256 hash (64 hex characters)
	assert.Len(t, deviceID1, 64)
}

func TestGenerateSecureToken(t *testing.T) {
	token1, err := generateSecureToken()
	require.NoError(t, err)
	assert.Len(t, token1, 64) // 32 bytes = 64 hex characters

	token2, err := generateSecureToken()
	require.NoError(t, err)
	assert.Len(t, token2, 64)

	// Tokens should be different
	assert.NotEqual(t, token1, token2)
}

func TestHashToken(t *testing.T) {
	token := "test-token"
	hash1 := hashToken(token)
	hash2 := hashToken(token)

	// Same token should generate same hash
	assert.Equal(t, hash1, hash2)

	// Hash should be a valid SHA-256 hash (64 hex characters)
	assert.Len(t, hash1, 64)
}

// Helper function to create a test token service
func createTestTokenService(t *testing.T) *services.TokenService {
	privateKeyPEM, publicKeyPEM, err := services.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	config := services.TokenConfig{
		PrivateKeyPEM: privateKeyPEM,
		PublicKeyPEM:  publicKeyPEM,
		Issuer:        "test-issuer",
		Audience:      "test-audience",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	tokenService, err := services.NewTokenService(config)
	require.NoError(t, err)
	return tokenService
}
