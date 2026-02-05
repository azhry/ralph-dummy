package services

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// BlacklistService handles token blacklisting
type BlacklistService struct {
	redisClient *redis.Client
}

// NewBlacklistService creates a new blacklist service
func NewBlacklistService(redisClient *redis.Client) *BlacklistService {
	return &BlacklistService{
		redisClient: redisClient,
	}
}

// IsBlacklisted checks if a token is blacklisted
func (b *BlacklistService) IsBlacklisted(c *gin.Context, jti string) (bool, error) {
	// Check access token blacklist
	accessKey := "blacklist:access:" + jti
	exists, err := b.redisClient.Exists(c.Request.Context(), accessKey).Result()
	if err != nil {
		return false, err
	}
	if exists > 0 {
		return true, nil
	}

	// Check refresh token blacklist
	refreshKey := "blacklist:refresh"
	isMember, err := b.redisClient.SIsMember(c.Request.Context(), refreshKey, jti).Result()
	if err != nil {
		return false, err
	}

	return isMember, nil
}

// BlacklistAccessToken adds an access token to the blacklist
func (b *BlacklistService) BlacklistAccessToken(ctx context.Context, jti string, ttl int64) error {
	key := "blacklist:access:" + jti
	return b.redisClient.Set(ctx, key, "1", 0).Err()
}

// BlacklistRefreshToken adds a refresh token to the blacklist
func (b *BlacklistService) BlacklistRefreshToken(ctx context.Context, jti string) error {
	key := "blacklist:refresh"
	return b.redisClient.SAdd(ctx, key, jti).Err()
}

// RevokeAllUserTokens revokes all tokens for a user
func (b *BlacklistService) RevokeAllUserTokens(ctx context.Context, userID string) error {
	// Find all refresh tokens for user
	pattern := "refresh:" + userID + ":*"
	iter := b.redisClient.Scan(ctx, 0, pattern, 0).Iterator()

	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	// Blacklist all refresh tokens
	if len(keys) > 0 {
		// Extract JTI from keys and add to blacklist
		for _, key := range keys {
			parts := strings.Split(key, ":")
			if len(parts) > 2 {
				jti := parts[len(parts)-1]
				b.BlacklistRefreshToken(ctx, jti)
			}
		}

		// Delete all sessions
		b.redisClient.Del(ctx, keys...)
	}

	return nil
}
