package middleware

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// RedisBlacklistChecker implements BlacklistChecker using Redis
type RedisBlacklistChecker struct {
	client *redis.Client
}

// NewRedisBlacklistChecker creates a new Redis-based blacklist checker
func NewRedisBlacklistChecker(client *redis.Client) BlacklistChecker {
	return &RedisBlacklistChecker{
		client: client,
	}
}

// IsBlacklisted checks if a token JTI is blacklisted
func (r *RedisBlacklistChecker) IsBlacklisted(c *gin.Context, jti string) (bool, error) {
	ctx := c.Request.Context()

	// Check access token blacklist
	accessKey := fmt.Sprintf("blacklist:access:%s", jti)
	exists, err := r.client.Exists(ctx, accessKey).Result()
	if err != nil {
		return false, err
	}

	if exists > 0 {
		return true, nil
	}

	// Check refresh token blacklist
	refreshKey := fmt.Sprintf("blacklist:refresh:%s", jti)
	exists, err = r.client.Exists(ctx, refreshKey).Result()
	if err != nil {
		return false, err
	}

	return exists > 0, nil
}

// BlacklistToken adds a token to the blacklist
func (r *RedisBlacklistChecker) BlacklistToken(ctx context.Context, jti string, tokenType string, ttl int) error {
	key := fmt.Sprintf("blacklist:%s:%s", tokenType, jti)
	return r.client.Set(ctx, key, "1", 0).Err()
}

// MemoryBlacklistChecker implements BlacklistChecker using in-memory storage
// This is useful for testing or small-scale deployments
type MemoryBlacklistChecker struct {
	blacklist map[string]bool
}

// NewMemoryBlacklistChecker creates a new memory-based blacklist checker
func NewMemoryBlacklistChecker() BlacklistChecker {
	return &MemoryBlacklistChecker{
		blacklist: make(map[string]bool),
	}
}

// IsBlacklisted checks if a token JTI is blacklisted
func (m *MemoryBlacklistChecker) IsBlacklisted(c *gin.Context, jti string) (bool, error) {
	accessKey := "blacklist:access:" + jti
	refreshKey := "blacklist:refresh:" + jti

	return m.blacklist[accessKey] || m.blacklist[refreshKey], nil
}

// BlacklistToken adds a token to the blacklist
func (m *MemoryBlacklistChecker) BlacklistToken(ctx context.Context, jti string, tokenType string, ttl int) error {
	key := fmt.Sprintf("blacklist:%s:%s", tokenType, jti)
	m.blacklist[key] = true
	return nil
}

// ClearBlacklist clears all blacklisted tokens
func (m *MemoryBlacklistChecker) ClearBlacklist() {
	m.blacklist = make(map[string]bool)
}
