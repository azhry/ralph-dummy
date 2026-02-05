package utils

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestJWTManager(t *testing.T) {
	jwtManager := NewJWTManager(
		"test-secret-key",
		"test-refresh-secret-key",
		15*time.Minute,
		7*24*time.Hour,
		"test-issuer",
	)

	// Test token generation
	userID := primitive.NewObjectID()
	email := "test@example.com"
	permissions := []string{"user"}

	tokenPair, err := jwtManager.GenerateTokenPair(userID, email, permissions)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	if tokenPair.AccessToken == "" {
		t.Error("Access token is empty")
	}

	if tokenPair.RefreshToken == "" {
		t.Error("Refresh token is empty")
	}

	if tokenPair.ExpiresAt.Before(time.Now()) {
		t.Error("Token expiration is in the past")
	}

	// Test access token validation
	claims, err := jwtManager.ValidateToken(tokenPair.AccessToken, AccessToken)
	if err != nil {
		t.Fatalf("Failed to validate access token: %v", err)
	}

	if claims.UserID != userID.Hex() {
		t.Errorf("Expected user ID %s, got %s", userID.Hex(), claims.UserID)
	}

	if claims.Email != email {
		t.Errorf("Expected email %s, got %s", email, claims.Email)
	}

	if claims.TokenType != AccessToken {
		t.Errorf("Expected token type %s, got %s", AccessToken, claims.TokenType)
	}

	// Test refresh token validation
	refreshClaims, err := jwtManager.ValidateToken(tokenPair.RefreshToken, RefreshToken)
	if err != nil {
		t.Fatalf("Failed to validate refresh token: %v", err)
	}

	if refreshClaims.TokenType != RefreshToken {
		t.Errorf("Expected token type %s, got %s", RefreshToken, refreshClaims.TokenType)
	}

	// Test invalid token
	_, err = jwtManager.ValidateToken("invalid-token", AccessToken)
	if err == nil {
		t.Error("Expected error for invalid token")
	}

	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}

	// Test token refresh
	newTokenPair, err := jwtManager.RefreshAccessToken(tokenPair.RefreshToken)
	if err != nil {
		t.Fatalf("Failed to refresh token: %v", err)
	}

	if newTokenPair.AccessToken == tokenPair.AccessToken {
		t.Error("New access token should be different from old one")
	}

	if newTokenPair.RefreshToken == tokenPair.RefreshToken {
		t.Error("New refresh token should be different from old one")
	}

	// Test user ID extraction
	extractedUserID, err := jwtManager.ExtractUserIDFromToken(newTokenPair.AccessToken)
	if err != nil {
		t.Fatalf("Failed to extract user ID: %v", err)
	}

	if extractedUserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, extractedUserID)
	}

	// Test email extraction
	extractedEmail, err := jwtManager.ExtractEmailFromToken(newTokenPair.AccessToken)
	if err != nil {
		t.Fatalf("Failed to extract email: %v", err)
	}

	if extractedEmail != email {
		t.Errorf("Expected email %s, got %s", email, extractedEmail)
	}

	// Test permission check
	if !jwtManager.HasPermission(newTokenPair.AccessToken, "user") {
		t.Error("Expected user to have 'user' permission")
	}

	if jwtManager.HasPermission(newTokenPair.AccessToken, "admin") {
		t.Error("Expected user not to have 'admin' permission")
	}
}

func TestJWTManager_ExpiredToken(t *testing.T) {
	jwtManager := NewJWTManager(
		"test-secret-key",
		"test-refresh-secret-key",
		1*time.Millisecond, // Very short expiration
		7*24*time.Hour,
		"test-issuer",
	)

	userID := primitive.NewObjectID()
	tokenPair, err := jwtManager.GenerateTokenPair(userID, "test@example.com", []string{"user"})
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Test expired token validation
	_, err = jwtManager.ValidateToken(tokenPair.AccessToken, AccessToken)
	if err == nil {
		t.Error("Expected error for expired token")
	}

	if err != ErrExpiredToken {
		t.Errorf("Expected ErrExpiredToken, got %v", err)
	}
}

func TestJWTManager_WrongTokenType(t *testing.T) {
	jwtManager := NewJWTManager(
		"test-secret-key",
		"test-refresh-secret-key",
		15*time.Minute,
		7*24*time.Hour,
		"test-issuer",
	)

	userID := primitive.NewObjectID()
	tokenPair, err := jwtManager.GenerateTokenPair(userID, "test@example.com", []string{"user"})
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	// Try to validate refresh token as access token
	_, err = jwtManager.ValidateToken(tokenPair.RefreshToken, AccessToken)
	if err == nil {
		t.Error("Expected error for wrong token type")
	}

	if err != ErrInvalidTokenType {
		t.Errorf("Expected ErrInvalidTokenType, got %v", err)
	}

	// Try to validate access token as refresh token
	_, err = jwtManager.ValidateToken(tokenPair.AccessToken, RefreshToken)
	if err == nil {
		t.Error("Expected error for wrong token type")
	}

	if err != ErrInvalidTokenType {
		t.Errorf("Expected ErrInvalidTokenType, got %v", err)
	}
}

func TestPasswordUtils(t *testing.T) {
	// Test password hashing
	password := "TestPassword123!"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hashedPassword == password {
		t.Error("Hashed password should not equal original password")
	}

	// Test password checking
	if !CheckPassword(hashedPassword, password) {
		t.Error("Password check failed for correct password")
	}

	if CheckPassword(hashedPassword, "wrong-password") {
		t.Error("Password check should fail for wrong password")
	}

	// Test password strength
	validator := NewPasswordValidator()

	// Test weak passwords
	weakPasswords := []string{
		"123",
		"password",
		"12345678",
		"abcdefgh",
		"ABCDEFGH",
		"Password",
	}

	for _, weakPassword := range weakPasswords {
		if validator.Validate(weakPassword) == nil {
			t.Errorf("Expected validation to fail for weak password: %s", weakPassword)
		}
	}

	// Test strong password
	strongPassword := "StrongP@ssw0rd!"
	if validator.Validate(strongPassword) != nil {
		t.Errorf("Expected validation to pass for strong password: %s", strongPassword)
	}

	// Test secure token generation
	token1, err := GenerateSecureToken(32)
	if err != nil {
		t.Fatalf("Failed to generate secure token: %v", err)
	}

	token2, err := GenerateSecureToken(32)
	if err != nil {
		t.Fatalf("Failed to generate secure token: %v", err)
	}

	if token1 == token2 {
		t.Error("Generated tokens should be unique")
	}

	if len(token1) != 32 {
		t.Errorf("Expected token length 32, got %d", len(token1))
	}

	// Test password comparison
	if !ComparePasswords("password", "password") {
		t.Error("Password comparison should succeed for identical passwords")
	}

	if ComparePasswords("password", "different") {
		t.Error("Password comparison should fail for different passwords")
	}
}
