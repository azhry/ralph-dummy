package config

import (
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	Server   ServerConfig   `mapstructure:",squash"`
	Database DatabaseConfig `mapstructure:",squash"`
	Auth     AuthConfig     `mapstructure:",squash"`
	Storage  StorageConfig  `mapstructure:",squash"`
	Email    EmailConfig    `mapstructure:",squash"`
}

type ServerConfig struct {
	Port           string        `mapstructure:"PORT"`
	Environment    string        `mapstructure:"APP_ENV"`
	AllowedOrigins []string      `mapstructure:"ALLOWED_ORIGINS"`
	ReadTimeout    time.Duration `mapstructure:"SERVER_READ_TIMEOUT"`
	WriteTimeout   time.Duration `mapstructure:"SERVER_WRITE_TIMEOUT"`
}

type DatabaseConfig struct {
	URI      string `mapstructure:"MONGODB_URI"`
	Database string `mapstructure:"MONGODB_DATABASE"`
	Timeout  int    `mapstructure:"MONGODB_TIMEOUT_SECONDS"`
}

type AuthConfig struct {
	JWTSecret        string        `mapstructure:"JWT_SECRET"`
	JWTRefreshSecret string        `mapstructure:"JWT_REFRESH_SECRET"`
	AccessTokenTTL   time.Duration `mapstructure:"JWT_ACCESS_TTL"`
	RefreshTokenTTL  time.Duration `mapstructure:"JWT_REFRESH_TTL"`
	BcryptCost       int           `mapstructure:"BCRYPT_COST"`
}

type StorageConfig struct {
	Provider  string `mapstructure:"STORAGE_PROVIDER"`
	Region    string `mapstructure:"AWS_REGION"`
	Bucket    string `mapstructure:"S3_BUCKET_NAME"`
	AccessKey string `mapstructure:"AWS_ACCESS_KEY_ID"`
	SecretKey string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
	CDNURL    string `mapstructure:"CDN_URL"`
}

type EmailConfig struct {
	Provider string `mapstructure:"EMAIL_PROVIDER"`
	APIKey   string `mapstructure:"SENDGRID_API_KEY"`
	From     string `mapstructure:"EMAIL_FROM"`
}

func Load() (*Config, error) {
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("MONGODB_URI", "mongodb://localhost:27017")
	viper.SetDefault("MONGODB_DATABASE", "wedding_invitations")
	viper.SetDefault("MONGODB_TIMEOUT_SECONDS", 10)
	viper.SetDefault("JWT_SECRET", "")
	viper.SetDefault("JWT_REFRESH_SECRET", "")
	viper.SetDefault("JWT_ACCESS_TTL", "15m")
	viper.SetDefault("JWT_REFRESH_TTL", "168h")
	viper.SetDefault("BCRYPT_COST", 12)
	viper.SetDefault("ALLOWED_ORIGINS", []string{"*"})

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// Bind environment variables to keys
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		// Config file not required - env vars can be used
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}
