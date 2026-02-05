package services

import (
	"crypto/rand"
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
	privateKey    *rsa.PrivateKey
	publicKey     *rsa.PublicKey
	issuer        string
	audience      string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
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
		return key.(*rsa.PrivateKey), nil
	}

	return key.(*rsa.PrivateKey), nil
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
		return key.(*rsa.PublicKey), nil
	}

	return key.(*rsa.PublicKey), nil
}

// GenerateRSAKeyPair generates a new RSA key pair for JWT signing
func GenerateRSAKeyPair(bits int) (privateKeyPEM, publicKeyPEM string, err error) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}

	// Encode private key to PEM
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal private key: %w", err)
	}

	privateKeyPEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	}))

	// Encode public key to PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}))

	return privateKeyPEM, publicKeyPEM, nil
}
