package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		setup    func()
		validate func(t *testing.T, cfg *Config)
		wantErr  bool
	}{
		{
			name: "default configuration",
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "8080", cfg.Server.Port)
				assert.Equal(t, "development", cfg.Server.Environment)
				assert.Equal(t, []string{"*"}, cfg.Server.AllowedOrigins)
				assert.Equal(t, "mongodb://localhost:27017", cfg.Database.URI)
				assert.Equal(t, "wedding_invitations", cfg.Database.Database)
				assert.Equal(t, 10, cfg.Database.Timeout)
				assert.Equal(t, 15*time.Minute, cfg.Auth.AccessTokenTTL)
				assert.Equal(t, 12, cfg.Auth.BcryptCost)
			},
			wantErr: false,
		},
		{
			name: "environment variables override",
			envVars: map[string]string{
				"PORT":            "9000",
				"APP_ENV":         "production",
				"MONGODB_URI":     "mongodb://prod:27017",
				"JWT_SECRET":      "prod-secret",
				"JWT_ACCESS_TTL":  "1h",
				"BCRYPT_COST":     "14",
				"ALLOWED_ORIGINS": "https://example.com,https://app.example.com",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "9000", cfg.Server.Port)
				assert.Equal(t, "production", cfg.Server.Environment)
				assert.Equal(t, []string{"https://example.com", "https://app.example.com"}, cfg.Server.AllowedOrigins)
				assert.Equal(t, "mongodb://prod:27017", cfg.Database.URI)
				assert.Equal(t, "prod-secret", cfg.Auth.JWTSecret)
				assert.Equal(t, 1*time.Hour, cfg.Auth.AccessTokenTTL)
				assert.Equal(t, 14, cfg.Auth.BcryptCost)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				// Clean up environment variables
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			cfg, err := Load()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestConfigIsDevelopment(t *testing.T) {
	cfg := &Config{}
	cfg.Server.Environment = "development"
	assert.True(t, cfg.IsDevelopment())
	assert.False(t, cfg.IsProduction())
}

func TestConfigIsProduction(t *testing.T) {
	cfg := &Config{}
	cfg.Server.Environment = "production"
	assert.True(t, cfg.IsProduction())
	assert.False(t, cfg.IsDevelopment())
}

func TestConfigStaging(t *testing.T) {
	cfg := &Config{}
	cfg.Server.Environment = "staging"
	assert.False(t, cfg.IsDevelopment())
	assert.False(t, cfg.IsProduction())
}
