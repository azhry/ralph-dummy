package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidTokenType = errors.New("invalid token type")
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Claims struct {
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	TokenType   TokenType `json:"token_type"`
	Permissions []string  `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type JWTManager struct {
	secretKey        []byte
	refreshSecretKey []byte
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
	issuer           string
}

func NewJWTManager(secretKey, refreshSecretKey string, accessTokenTTL, refreshTokenTTL time.Duration, issuer string) *JWTManager {
	return &JWTManager{
		secretKey:        []byte(secretKey),
		refreshSecretKey: []byte(refreshSecretKey),
		accessTokenTTL:   accessTokenTTL,
		refreshTokenTTL:  refreshTokenTTL,
		issuer:           issuer,
	}
}

func (j *JWTManager) GenerateTokenPair(userID primitive.ObjectID, email string, permissions []string) (*TokenPair, error) {
	now := time.Now()

	// Generate access token
	accessClaims := &Claims{
		UserID:      userID.Hex(),
		Email:       email,
		TokenType:   AccessToken,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        primitive.NewObjectID().Hex(),
			Issuer:    j.issuer,
			Subject:   userID.Hex(),
			Audience:  []string{"wedding-invitation-api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTokenTTL)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(j.secretKey)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshClaims := &Claims{
		UserID:      userID.Hex(),
		Email:       email,
		TokenType:   RefreshToken,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        primitive.NewObjectID().Hex(),
			Issuer:    j.issuer,
			Subject:   userID.Hex(),
			Audience:  []string{"wedding-invitation-api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshTokenTTL)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(j.refreshSecretKey)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    now.Add(j.accessTokenTTL),
	}, nil
}

func (j *JWTManager) ValidateToken(tokenString string, tokenType TokenType) (*Claims, error) {
	// First, parse the token without signature verification to check the token type
	unsafeToken, _, err := jwt.NewParser().ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, ErrInvalidToken
	}

	unsafeClaims, ok := unsafeToken.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Check if the token type matches before verifying the signature
	if unsafeClaims.TokenType != tokenType {
		return nil, ErrInvalidTokenType
	}

	// Now validate with the correct key
	var secretKey []byte
	switch tokenType {
	case AccessToken:
		secretKey = j.secretKey
	case RefreshToken:
		secretKey = j.refreshSecretKey
	default:
		return nil, ErrInvalidTokenType
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Double check token type (should match from above)
	if claims.TokenType != tokenType {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

func (j *JWTManager) RefreshAccessToken(refreshTokenString string) (*TokenPair, error) {
	claims, err := j.ValidateToken(refreshTokenString, RefreshToken)
	if err != nil {
		return nil, err
	}

	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	return j.GenerateTokenPair(userID, claims.Email, claims.Permissions)
}

func (j *JWTManager) ExtractUserIDFromToken(tokenString string) (primitive.ObjectID, error) {
	claims, err := j.ValidateToken(tokenString, AccessToken)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return primitive.ObjectIDFromHex(claims.UserID)
}

func (j *JWTManager) ExtractEmailFromToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString, AccessToken)
	if err != nil {
		return "", err
	}

	return claims.Email, nil
}

func (j *JWTManager) HasPermission(tokenString string, permission string) bool {
	claims, err := j.ValidateToken(tokenString, AccessToken)
	if err != nil {
		return false
	}

	for _, p := range claims.Permissions {
		if p == permission {
			return true
		}
	}

	return false
}
