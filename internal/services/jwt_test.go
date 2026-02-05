package services

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRSAKeyPair(t *testing.T) {
	privateKeyPEM, publicKeyPEM, err := GenerateRSAKeyPair(2048)
	require.NoError(t, err)
	assert.NotEmpty(t, privateKeyPEM)
	assert.NotEmpty(t, publicKeyPEM)
	assert.Contains(t, privateKeyPEM, "PRIVATE KEY")
	assert.Contains(t, publicKeyPEM, "PUBLIC KEY")
}

func TestNewTokenService(t *testing.T) {
	privateKeyPEM, publicKeyPEM, err := GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	config := TokenConfig{
		PrivateKeyPEM: privateKeyPEM,
		PublicKeyPEM:  publicKeyPEM,
		Issuer:        "test-issuer",
		Audience:      "test-audience",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	tokenService, err := NewTokenService(config)
	require.NoError(t, err)
	assert.NotNil(t, tokenService)
}

func TestTokenService_GenerateTokenPair(t *testing.T) {
	tokenService := createTestTokenService(t)

	userID := "test-user-id"
	deviceID := "test-device-id"
	role := "user"

	tokenPair, err := tokenService.GenerateTokenPair(userID, deviceID, role)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.NotEmpty(t, tokenPair.AccessJTI)
	assert.NotEmpty(t, tokenPair.RefreshJTI)
	assert.NotEqual(t, tokenPair.AccessJTI, tokenPair.RefreshJTI)
}

func TestTokenService_ParseToken(t *testing.T) {
	tokenService := createTestTokenService(t)

	userID := "test-user-id"
	deviceID := "test-device-id"
	role := "user"

	// Generate tokens
	tokenPair, err := tokenService.GenerateTokenPair(userID, deviceID, role)
	require.NoError(t, err)

	// Parse access token
	token, err := tokenService.ParseToken(tokenPair.AccessToken)
	require.NoError(t, err)
	assert.True(t, token.Valid)

	// Extract and verify claims
	claims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, userID, claims["sub"])
	assert.Equal(t, tokenPair.AccessJTI, claims["jti"])
	assert.Equal(t, "test-issuer", claims["iss"])
	assert.Equal(t, "test-audience", claims["aud"])
	assert.Equal(t, deviceID, claims["device_id"])
	assert.Equal(t, role, claims["role"])

	// Parse refresh token
	refreshToken, err := tokenService.ParseToken(tokenPair.RefreshToken)
	require.NoError(t, err)
	assert.True(t, refreshToken.Valid)

	refreshClaims, ok := refreshToken.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, userID, refreshClaims["sub"])
	assert.Equal(t, tokenPair.RefreshJTI, refreshClaims["jti"])
	assert.Equal(t, "refresh", refreshClaims["type"])
	assert.Equal(t, deviceID, refreshClaims["device_id"])
}

func TestTokenService_ParseTokenWithClaims(t *testing.T) {
	tokenService := createTestTokenService(t)

	userID := "test-user-id"
	deviceID := "test-device-id"
	role := "user"

	tokenPair, err := tokenService.GenerateTokenPair(userID, deviceID, role)
	require.NoError(t, err)

	// Parse with custom claims
	customClaims := jwt.MapClaims{}
	token, err := tokenService.ParseTokenWithClaims(tokenPair.AccessToken, &customClaims)
	require.NoError(t, err)
	assert.True(t, token.Valid)

	assert.Equal(t, userID, customClaims["sub"])
	assert.Equal(t, role, customClaims["role"])
}

func TestTokenService_GenerateVerificationToken(t *testing.T) {
	tokenService := createTestTokenService(t)

	userID := "test-user-id"

	token, err := tokenService.GenerateVerificationToken(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Parse and verify the token
	parsedToken, err := tokenService.ParseToken(token)
	require.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, userID, claims["sub"])
	assert.Equal(t, "verification", claims["type"])
}

func TestTokenService_InvalidToken(t *testing.T) {
	tokenService := createTestTokenService(t)

	// Test with invalid token
	_, err := tokenService.ParseToken("invalid.token.here")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse token")

	// Test with malformed token
	_, err = tokenService.ParseToken("not.a.jwt")
	assert.Error(t, err)
}

func TestTokenService_ExpiredToken(t *testing.T) {
	// Create token service with very short expiry
	privateKeyPEM, publicKeyPEM, err := GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	config := TokenConfig{
		PrivateKeyPEM: privateKeyPEM,
		PublicKeyPEM:  publicKeyPEM,
		Issuer:        "test-issuer",
		Audience:      "test-audience",
		AccessExpiry:  1 * time.Millisecond, // Very short expiry
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	tokenService, err := NewTokenService(config)
	require.NoError(t, err)

	// Generate token
	tokenPair, err := tokenService.GenerateTokenPair("test-user", "test-device", "user")
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Try to parse expired token
	_, err = tokenService.ParseToken(tokenPair.AccessToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse token")
}

func TestTokenService_WrongIssuer(t *testing.T) {
	tokenService := createTestTokenService(t)

	// Create a token with different issuer
	privateKeyPEM, publicKeyPEM, err := GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	wrongConfig := TokenConfig{
		PrivateKeyPEM: privateKeyPEM,
		PublicKeyPEM:  publicKeyPEM,
		Issuer:        "wrong-issuer",
		Audience:      "test-audience",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	wrongTokenService, err := NewTokenService(wrongConfig)
	require.NoError(t, err)

	// Generate token with wrong issuer
	tokenPair, err := wrongTokenService.GenerateTokenPair("test-user", "test-device", "user")
	require.NoError(t, err)

	// Try to parse with correct token service (should fail due to wrong issuer)
	_, err = tokenService.ParseToken(tokenPair.AccessToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse token")
}

// Helper function to create a test token service
func createTestTokenService(t *testing.T) *TokenService {
	privateKeyPEM, publicKeyPEM, err := GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	config := TokenConfig{
		PrivateKeyPEM: privateKeyPEM,
		PublicKeyPEM:  publicKeyPEM,
		Issuer:        "test-issuer",
		Audience:      "test-audience",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	tokenService, err := NewTokenService(config)
	require.NoError(t, err)
	return tokenService
}
